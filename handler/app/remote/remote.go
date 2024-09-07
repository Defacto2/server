// Package remote provides the remote download and update of artifact data from third-party sources such as API's.
package remote

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/Defacto2/archive"
	"github.com/Defacto2/helper"
	"github.com/Defacto2/server/handler/app/internal/simple"
	"github.com/Defacto2/server/handler/demozoo"
	"github.com/Defacto2/server/handler/pouet"
	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/Defacto2/server/model"
	"github.com/labstack/echo/v4"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

var (
	ErrDB    = errors.New("database connection is nil")
	ErrExist = errors.New("file already exists")
)

// DemozooLink is the response from the task of GetDemozooFile.
//
//nolint:tagliatelle
type DemozooLink struct {
	ID          int      `json:"id"`            // ID is the Demozoo production ID.
	UUID        string   `json:"uuid"`          // UUID is the file production UUID.
	Github      string   `json:"github_repo"`   // GitHub is the GitHub repository URI.
	YouTube     string   `json:"youtube_video"` // YouTube is the YouTube watch video URI.
	Pouet       int      `json:"pouet_prod"`    // Pouet is the Pouet production ID.
	Releaser1   string   `json:"releaser1"`     // Releaser1 is the first releaser of the file.
	Releaser2   string   `json:"releaser2"`     // Releaser2 is the second releaser of the file.
	Title       string   `json:"title"`         // Title is the file title.
	IssuedYear  int16    `json:"issued_year"`   // Year is the year the file was issued.
	IssuedMonth int16    `json:"issued_month"`  // Month is the month the file was issued.
	IssuedDay   int16    `json:"issued_day"`    // Day is the day the file was issued.
	CreditText  []string `json:"credit_text"`   // credit_text, writer
	CreditCode  []string `json:"credit_code"`   // credit_program, programmer/coder
	CreditArt   []string `json:"credit_art"`    // credit_illustration, artist/graphics
	CreditAudio []string `json:"credit_audio"`  // credit_audio, musician/sound
	Filename    string   `json:"filename"`      // Filename is the file name of the download.
	FileSize    int      `json:"file_size"`     // Size is the file size in bytes.
	Content     string   `json:"content"`       // Content is the file archive content.
	FileType    string   `json:"file_type"`     // Type is the file type.
	FileHash    string   `json:"file_hash"`     // Hash is the file integrity hash.
	Platform    string   `json:"platform"`      // Platform is the file platform.
	Section     string   `json:"section"`       // Section is the file section.
	Error       string   `json:"error"`         // Error is the error message if the download or record update failed.
}

// Download fetches the download link from Demozoo and saves it to the download directory.
// It then runs Update to modify the database record with various metadata from the file and Demozoo record API data.
func (got *DemozooLink) Download(c echo.Context, db *sql.DB, downloadDir string) error {
	var prod demozoo.Production
	if _, err := prod.Get(got.ID); err != nil {
		return fmt.Errorf("could not get record %d from demozoo api: %w", got.ID, err)
	}
	for _, link := range prod.DownloadLinks {
		if link.URL == "" {
			continue
		}
		df, err := GetFile(link.URL, 0)
		if tryNextLink := err != nil || df.Path == ""; tryNextLink {
			continue
		}
		base := filepath.Base(link.URL)
		dst := filepath.Join(downloadDir, got.UUID)
		got.Filename = base
		if err := helper.RenameFileOW(df.Path, dst); err != nil {
			sameFiles, err := helper.FileMatch(df.Path, dst)
			if err != nil {
				return fmt.Errorf("could not rename file, %s: %w", dst, err)
			}
			if !sameFiles {
				return fmt.Errorf("%w, will not overwrite, %s", ErrExist, dst)
			}
		}
		size, err := strconv.Atoi(df.ContentLength)
		if err == nil {
			got.FileSize = size
		}
		got.Filename = base
		got.Error = ""
		got.Github = prod.GithubRepo()
		got.Pouet = prod.PouetProd()
		got.YouTube = prod.YouTubeVideo()
		y, m, d := prod.Released()
		got.IssuedYear = y
		got.IssuedMonth = m
		got.IssuedDay = d
		r1, r2 := prod.Groups()
		got.Releaser1 = r1
		got.Releaser2 = r2
		got.Title = prod.Title
		ctext, ccode, cart, caudio := prod.Releasers()
		got.CreditText = ctext
		got.CreditCode = ccode
		got.CreditArt = cart
		got.CreditAudio = caudio
		plat, sect := prod.SuperType()
		got.Platform = plat.String()
		got.Section = sect.String()
		return got.Stat(c, db, downloadDir)
	}
	got.Error = "no usable download links found, they returned 404 or were empty"
	return c.JSON(http.StatusNotModified, got)
}

// Stat sets the file size, hash, type, and archive content of the file.
// The UUID is used to locate the file in the download directory.
func (got *DemozooLink) Stat(c echo.Context, db *sql.DB, downloadDir string) error {
	name := filepath.Join(downloadDir, got.UUID)
	if got.FileSize == 0 {
		stat, err := os.Stat(name)
		if err != nil {
			return fmt.Errorf("could not stat file, %s: %w", name, err)
		}
		got.FileSize = int(stat.Size())
	}
	strong, err := helper.StrongIntegrity(name)
	if err != nil {
		return fmt.Errorf("could not get strong integrity hash, %s: %w", name, err)
	}
	got.FileHash = strong
	if got.FileType == "" {
		got.FileType = simple.MagicAsTitle(name)
	}
	return got.ArchiveContent(c, db, name)
}

// ArchiveContent sets the archive content and readme text of the source file.
func (got *DemozooLink) ArchiveContent(c echo.Context, db *sql.DB, src string) error {
	files, err := archive.List(src, got.Filename)
	if err != nil {
		fmt.Fprint(io.Discard, err)
		return nil
	}
	got.Content = strings.Join(files, "\n")
	return got.Update(c, db)
}

// Update modifies the database record using data provided by the DemozooLink struct.
// A JSON response is returned with the success status of the update.
func (got DemozooLink) Update(c echo.Context, db *sql.DB) error {
	if db == nil {
		return ErrDB
	}
	uid := got.UUID
	ctx := context.Background()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("demozoolink update begin tx %w: %s", err, uid)
	}
	f, err := model.OneByUUID(ctx, tx, true, uid)
	if err != nil {
		return fmt.Errorf("demozoolink update by uuid %w: %s", err, uid)
	}
	got.updates(f)
	if _, err = f.Update(ctx, tx, boil.Infer()); err != nil {
		return fmt.Errorf("demozoolink update infer %w: %s", err, uid)
	}
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("demozoolink update commit %w: %s", err, uid)
	}
	return c.HTML(http.StatusOK, `<p class="text-success">Successful Demozoo update</p>`)
}

func (got DemozooLink) updates(f *models.File) { //nolint:cyclop
	if s := strings.TrimSpace(got.Github); s != "" {
		f.WebIDGithub = null.StringFrom(s)
	}
	if s := strings.TrimSpace(got.YouTube); s != "" {
		f.WebIDYoutube = null.StringFrom(s)
	}
	if i := int64(got.Pouet); i > 0 {
		f.WebIDPouet = null.Int64From(i)
	}
	if s := strings.TrimSpace(got.Releaser1); s != "" {
		f.GroupBrandFor = null.StringFrom(s)
	}
	if s := strings.TrimSpace(got.Releaser2); s != "" {
		f.GroupBrandBy = null.StringFrom(s)
	}
	if s := strings.TrimSpace(got.Title); s != "" {
		f.RecordTitle = null.StringFrom(s)
	}
	if i := (got.IssuedDay); i > 0 {
		f.DateIssuedDay = null.Int16From(i)
	}
	if i := (got.IssuedMonth); i > 0 {
		f.DateIssuedMonth = null.Int16From(i)
	}
	if i := (got.IssuedYear); i > 0 {
		f.DateIssuedYear = null.Int16From(i)
	}
	if s := strings.Join(got.CreditAudio, ","); s != "" {
		f.CreditAudio = null.StringFrom(s)
	}
	if s := strings.Join(got.CreditArt, ","); s != "" {
		f.CreditIllustration = null.StringFrom(s)
	}
	if s := strings.Join(got.CreditCode, ","); s != "" {
		f.CreditProgram = null.StringFrom(s)
	}
	if s := strings.Join(got.CreditText, ","); s != "" {
		f.CreditText = null.StringFrom(s)
	}
	if s := strings.TrimSpace(got.Filename); s != "" {
		f.Filename = null.StringFrom(s)
	}
	if i := int64(got.FileSize); i > 0 {
		f.Filesize = null.Int64From(i)
	}
	if s := strings.TrimSpace(got.FileType); s != "" {
		f.FileMagicType = null.StringFrom(s)
	}
	if s := strings.TrimSpace(got.FileHash); s != "" {
		f.FileIntegrityStrong = null.StringFrom(s)
	}
	if s := strings.TrimSpace(got.Content); s != "" {
		f.FileZipContent = null.StringFrom(s)
	}
	if s := strings.TrimSpace(got.Platform); s != "" {
		f.Platform = null.StringFrom(s)
	}
	if s := strings.TrimSpace(got.Section); s != "" {
		f.Section = null.StringFrom(s)
	}
}

// PouetLink is the response from the task of GetDemozooFile.
//
//nolint:tagliatelle
type PouetLink struct {
	ID          int    `json:"id"`           // ID is the Demozoo production ID.
	UUID        string `json:"uuid"`         // UUID is the file production UUID.
	Demozoo     int    `json:"demozoo_prod"` // Demozoo production ID.
	Releaser1   string `json:"releaser1"`    // Releaser1 is the first releaser of the file.
	Releaser2   string `json:"releaser2"`    // Releaser2 is the second releaser of the file.
	Title       string `json:"title"`        // Title is the file title.
	IssuedYear  int16  `json:"issued_year"`  // Year is the year the file was issued.
	IssuedMonth int16  `json:"issued_month"` // Month is the month the file was issued.
	IssuedDay   int16  `json:"issued_day"`   // Day is the day the file was issued.
	Filename    string `json:"filename"`     // Filename is the file name of the download.
	FileSize    int    `json:"file_size"`    // Size is the file size in bytes.
	Content     string `json:"content"`      // Content is the file archive content.
	FileType    string `json:"file_type"`    // Type is the file type.
	FileHash    string `json:"file_hash"`    // Hash is the file integrity hash.
	Platform    string `json:"platform"`     // Platform is the file platform.
	Section     string `json:"section"`      // Section is the file section.
	Error       string `json:"error"`        // Error is the error message if the download or record update failed.
}

func (got *PouetLink) Download(c echo.Context, db *sql.DB, downloadDir string) error {
	var prod pouet.Production
	if _, err := prod.Get(got.ID); err != nil {
		return fmt.Errorf("could not get record %d from demozoo api: %w", got.ID, err)
	}
	downloadURL := prod.Download
	if downloadURL == "" {
		return nil
	}
	df, err := GetFile(downloadURL, 10*time.Second)
	if err != nil {
		return fmt.Errorf("could not get file, %s: %w", downloadURL, err)
	}
	base := filepath.Base(downloadURL)
	dst := filepath.Join(downloadDir, got.UUID)
	got.Filename = base
	if err := helper.RenameFileOW(df.Path, dst); err != nil {
		sameFiles, err := helper.FileMatch(df.Path, dst)
		if err != nil {
			return fmt.Errorf("could not rename file, %s: %w", dst, err)
		}
		if !sameFiles {
			return fmt.Errorf("%w, will not overwrite, %s", ErrExist, dst)
		}
	}
	got.Filename = base
	got.Error = ""
	if i, err := strconv.Atoi(prod.Demozoo); err == nil && i > 0 {
		got.Demozoo = i
	}
	y, m, d := prod.Released()
	got.IssuedYear = y
	got.IssuedMonth = m
	got.IssuedDay = d
	r1, r2 := prod.Releasers()
	got.Releaser1 = r1
	got.Releaser2 = r2
	got.Title = prod.Title
	plat, sect := prod.PlatformType()
	got.Platform = plat.String()
	got.Section = sect.String()
	return got.Stat(c, db, downloadDir)
}

// Stat sets the file size, hash, type, and archive content of the file.
// The UUID is used to locate the file in the download directory.
func (got *PouetLink) Stat(c echo.Context, db *sql.DB, downloadDir string) error {
	name := filepath.Join(downloadDir, got.UUID)
	if got.FileSize == 0 {
		stat, err := os.Stat(name)
		if err != nil {
			return fmt.Errorf("could not stat file, %s: %w", name, err)
		}
		got.FileSize = int(stat.Size())
	}
	strong, err := helper.StrongIntegrity(name)
	if err != nil {
		return fmt.Errorf("could not get strong integrity hash, %s: %w", name, err)
	}
	got.FileHash = strong
	if got.FileType == "" {
		got.FileType = simple.MagicAsTitle(name)
	}
	return got.ArchiveContent(c, db, name)
}

// ArchiveContent sets the archive content and readme text of the source file.
func (got *PouetLink) ArchiveContent(c echo.Context, db *sql.DB, src string) error {
	files, err := archive.List(src, got.Filename)
	if err != nil {
		return c.JSON(http.StatusOK, got)
	}
	got.Content = strings.Join(files, "\n")
	return got.Update(c, db)
}

// Update modifies the database record using data provided by the DemozooLink struct.
// A JSON response is returned with the success status of the update.
func (got PouetLink) Update(c echo.Context, db *sql.DB) error {
	if db == nil {
		return ErrDB
	}
	uid := got.UUID
	ctx := context.Background()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("demozoolink update begin tx %w: %s", err, uid)
	}
	f, err := model.OneByUUID(ctx, tx, true, uid)
	if err != nil {
		return fmt.Errorf("demozoolink update by uuid %w: %s", err, uid)
	}
	got.updates(f)
	if _, err = f.Update(ctx, tx, boil.Infer()); err != nil {
		return fmt.Errorf("demozoolink update infer %w: %s", err, uid)
	}
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("demozoolink update commit %w: %s", err, uid)
	}
	return c.HTML(http.StatusOK, `<p class="text-success">Successful Pouet update</p>`)
}

func (got PouetLink) updates(f *models.File) {
	if i := got.Demozoo; i > 0 {
		f.WebIDDemozoo = null.Int64From(int64(i))
	}
	if s := strings.TrimSpace(got.Releaser1); s != "" {
		f.GroupBrandFor = null.StringFrom(s)
	}
	if s := strings.TrimSpace(got.Releaser2); s != "" {
		f.GroupBrandBy = null.StringFrom(s)
	}
	if s := strings.TrimSpace(got.Title); s != "" {
		f.RecordTitle = null.StringFrom(s)
	}
	if i := (got.IssuedDay); i > 0 {
		f.DateIssuedDay = null.Int16From(i)
	}
	if i := (got.IssuedMonth); i > 0 {
		f.DateIssuedMonth = null.Int16From(i)
	}
	if i := (got.IssuedYear); i > 0 {
		f.DateIssuedYear = null.Int16From(i)
	}
	if s := strings.TrimSpace(got.Filename); s != "" {
		f.Filename = null.StringFrom(s)
	}
	if i := int64(got.FileSize); i > 0 {
		f.Filesize = null.Int64From(i)
	}
	if s := strings.TrimSpace(got.FileType); s != "" {
		f.FileMagicType = null.StringFrom(s)
	}
	if s := strings.TrimSpace(got.FileHash); s != "" {
		f.FileIntegrityStrong = null.StringFrom(s)
	}
	if s := strings.TrimSpace(got.Content); s != "" {
		f.FileZipContent = null.StringFrom(s)
	}
	if s := strings.TrimSpace(got.Platform); s != "" {
		f.Platform = null.StringFrom(s)
	}
	if s := strings.TrimSpace(got.Section); s != "" {
		f.Section = null.StringFrom(s)
	}
}
