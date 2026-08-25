// Package remote provides the remote download and update of artifact data from third-party sources such as API's.
package remote

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Defacto2/archive"
	"github.com/Defacto2/helper"
	"github.com/Defacto2/server/handler/app/internal/simple"
	"github.com/Defacto2/server/handler/demozoo"
	"github.com/Defacto2/server/handler/pouet"
	"github.com/Defacto2/server/internal/dir"
	"github.com/Defacto2/server/internal/nils"
	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/Defacto2/server/model"
	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/labstack/echo/v5"
)

var (
	ErrExist    = errors.New("file already exists")
	ErrNoRecord = errors.New("could not the get record from demozoo api")
)

// DemozooLink is the response from the task of GetDemozooFile.
//
//nolint:tagliatelle
type DemozooLink struct {
	UUID        string   `json:"uuid"`          // UUID is the file production UUID.
	Github      string   `json:"github_repo"`   // GitHub is the GitHub repository URI.
	YouTube     string   `json:"youtube_video"` // YouTube is the YouTube watch video URI.
	Releaser1   string   `json:"releaser1"`     // Releaser1 is the first releaser of the file.
	Releaser2   string   `json:"releaser2"`     // Releaser2 is the second releaser of the file.
	Title       string   `json:"title"`         // Title is the file title.
	Filename    string   `json:"filename"`      // Filename is the file name of the download.
	Content     string   `json:"content"`       // Content is the file archive content.
	FileType    string   `json:"file_type"`     // Type is the file type.
	FileHash    string   `json:"file_hash"`     // Hash is the file integrity hash.
	Platform    string   `json:"platform"`      // Platform is the file platform.
	Section     string   `json:"section"`       // Section is the file section.
	Error       string   `json:"error"`         // Error is the error message if the download or record update failed.
	CreditText  []string `json:"credit_text"`   // credit_text, writer
	CreditCode  []string `json:"credit_code"`   // credit_program, programmer/coder
	CreditArt   []string `json:"credit_art"`    // credit_illustration, artist/graphics
	CreditAudio []string `json:"credit_audio"`  // credit_audio, musician/sound
	ID          int      `json:"id"`            // ID is the Demozoo production ID.
	Pouet       int      `json:"pouet_prod"`    // Pouet is the Pouet production ID.
	FileSize    int      `json:"file_size"`     // Size is the file size in bytes.
	IssuedYear  int16    `json:"issued_year"`   // Year is the year the file was issued.
	IssuedMonth int16    `json:"issued_month"`  // Month is the month the file was issued.
	IssuedDay   int16    `json:"issued_day"`    // Day is the day the file was issued.
}

// Download fetches the download link from Demozoo and saves it to the download directory.
// It then runs Update to modify the database record with various metadata from the file and Demozoo record API data.
func (got *DemozooLink) Download( //nolint:funlen
	ctx context.Context, sl *slog.Logger, c *echo.Context, db *sql.DB, download dir.Directory,
) error {
	const msg = "demozoo link download"
	const format = "%s for id %d: %w"
	if err := nils.Check(ctx, sl, c, db); err != nil {
		return fmt.Errorf(format, "check", 0, err)
	}
	id := func() slog.Attr {
		return slog.Int("id", got.ID)
	}
	var prod demozoo.Production
	errStatus, err := prod.Get(ctx, got.ID)
	if err != nil {
		const s = " could not get record from api"
		sl.Info(msg+s, id(), slog.Any("error", err))
		return fmt.Errorf(format, s, got.ID, err)
	}
	if errStatus > 0 {
		s := " did not return an okay status"
		sl.Info(msg+s, id(), slog.Int("status", errStatus), slog.Any("error", err))
		return fmt.Errorf(format, s+", got "+strconv.Itoa(errStatus), got.ID, ErrNoRecord)
	}
	// Originally we would return an error and abort the database record update if the download link
	// could not be fetched. However, this happens too frequently with some of the more popular Scene
	// websites. Sites such as scene.org frequently time out requests. So, as of 20-Jul-25, the
	// behavior now updates the record even when the download fails.
	for i, link := range prod.DownloadLinks {
		if link.URL == "" {
			sl.Info(msg+" link url is empty", id(), slog.Any("link", link))
			continue
		}
		// append demozoo record metadata
		base := filepath.Base(link.URL)
		dst := filepath.Join(download.Path(), got.UUID)
		got.Filename = base
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
		// attempt to download the remote file
		// if this task fails, update the record with the demozoo metadata
		response, err := getRemoteFile(ctx, sl, prod, i, link.URL)
		if err != nil {
			sl.Info(msg+" download remote file",
				id(), slog.String("link", link.URL), slog.Any("error", err))
			if err1 := got.Update(ctx, c, db); err1 != nil {
				sl.Info(msg+" download remote file error but update", id(), slog.Any("error", err1))
				return err1
			}
			return err
		} else if response == (Response{ContentLength: "", ContentType: "", LastModified: "", Path: ""}) {
			sl.Info(msg+" download remote file but empty response", id())
			continue
		}
		// assuming the download link was successful in being fetched,
		// we now obtain the file's metadata and incorporate those into the database record.
		if err := renameOW(response.Path, dst); err != nil {
			sl.Info(msg, id(), slog.String("source", response.Path), slog.Any("error", err))
			return err
		}
		cl := response.ContentLength
		if size, err := strconv.Atoi(cl); err != nil {
			sl.Info(msg+" atoi error for content length", id(), slog.String("content_length", cl))
		} else {
			got.FileSize = size
		}
		got.Error = ""
		if err := got.Stat(ctx, sl, c, db, download); err != nil {
			sl.Info(msg, id(), slog.Any("error", err))
		}
		return nil
	}
	got.Error = "no usable download links found, they all returned a 404 error or were empty"
	return c.JSON(http.StatusNotModified, got)
}

func renameOW(src, dst string) error {
	const format = "cannot rename dst file %s %s: %w"
	if err := helper.RenameFileOW(src, dst); err != nil {
		sameFiles, err := helper.FileMatch(src, dst)
		if err != nil {
			return fmt.Errorf(format, "file match error", dst, err)
		}
		if !sameFiles {
			return fmt.Errorf(format, "as existing files will be overwritten", dst, err)
		}
	}
	return nil
}

// getRemoteFile fetches the download link from Demozoo and saves it to the download directory.
// If the DownloadResponse is empty due to a production without a download link or a timeout,
// then it should be handled as a continue in the calling function.
func getRemoteFile(
	ctx context.Context, sl *slog.Logger, prod demozoo.Production, i int, linkURL string,
) (Response, error) {
	const format = "cannot get the remote file from %s: %w"
	if err := nils.Check(ctx, sl); err != nil {
		return Response{}, fmt.Errorf("get remove file check: %w", err)
	}
	timeout := TimeoutShort
	if len(prod.DownloadLinks) == 1 {
		timeout = TimeoutLong
	}
	resp, err := GetFile(ctx, sl, timeout, linkURL)
	if skip := err != nil || resp.Path == ""; skip {
		sl.Info("get remote file returned an error or is invalid",
			slog.String("resp_path", resp.Path),
			slog.Any("error", err))
		// If the last link failed then return the error, otherwise this will fail silently.
		if lastLink := i+1 >= len(prod.DownloadLinks); lastLink {
			return Response{},
				fmt.Errorf(format, "any linked url "+linkURL, err)
		}
		return Response{ContentLength: "", ContentType: "", LastModified: "", Path: ""}, nil
	}
	return resp, nil
}

// Stat sets the file size, hash, type, and archive content of the file.
// The UUID is used to locate the file in the download directory.
//
//nolint:dupl // intentional similarity
func (got *DemozooLink) Stat(
	ctx context.Context, sl *slog.Logger, c *echo.Context, db *sql.DB, download dir.Directory,
) error {
	const format = "demozoo link stat file and integrity %s: %w"
	if err := nils.Check(ctx, sl, c, db); err != nil {
		return fmt.Errorf(format, "check", err)
	}
	name := filepath.Join(download.Path(), got.UUID)
	if got.FileSize == 0 {
		stat, err := os.Stat(name)
		if err != nil {
			return fmt.Errorf(format, "but could not stat file "+name, err)
		}
		got.FileSize = int(stat.Size())
	}
	strong, err := helper.StrongIntegrity(name)
	if err != nil {
		return fmt.Errorf(format, "but could not get the strong integrity hash "+name, err)
	}
	got.FileHash = strong
	if got.FileType == "" {
		got.FileType = simple.MagicAsTitle(sl, name)
	}
	return got.ArchiveContent(ctx, sl, c, db, name)
}

// ArchiveContent sets the archive content and readme text of the source file.
func (got *DemozooLink) ArchiveContent(
	ctx context.Context, sl *slog.Logger, c *echo.Context, db *sql.DB, src string,
) error {
	const msg = "demozoo link archive content"
	const format = msg + ": %w"
	if err := nils.Check(ctx, sl, c, db); err != nil {
		return fmt.Errorf(format, err)
	}
	files, err := archive.Lists(ctx, src)
	if err != nil {
		sl.Info(msg+" caused an error",
			slog.String("source", src), slog.String("filename", got.Filename), slog.Any("error", err))
		return nil
	}
	got.Content = strings.Join(files, "\n")
	if err := got.Update(ctx, c, db); err != nil {
		sl.Info(msg+" update caused an error",
			slog.String("source", src), slog.String("filename", got.Filename), slog.Any("error", err))
		return err
	}
	const html = `<p class="text-success">Successful Demozoo update</p>`
	return c.HTML(http.StatusOK, html)
}

// Update modifies the database record using data provided by the DemozooLink struct.
// A JSON response is returned with the success status of the update.
func (got *DemozooLink) Update(ctx context.Context, c *echo.Context, db *sql.DB) error {
	const format = "demozoo link update %s uuid %s: %w"
	if err := nils.Check(ctx, c, db); err != nil {
		return fmt.Errorf(format, "check", "n/a", err)
	}
	uid := got.UUID
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf(format, "begin tx", uid, err)
	}
	f, err := model.OneByUUID(ctx, tx, true, uid)
	if err != nil {
		return fmt.Errorf(format, "one record by", uid, err)
	}
	got.updateValues(f)
	if _, err = f.Update(ctx, tx, boil.Infer()); err != nil {
		return fmt.Errorf(format, "infer", uid, err)
	}
	if err = tx.Commit(); err != nil {
		return fmt.Errorf(format, "tx commit", uid, err)
	}
	return nil
}

func (got *DemozooLink) updateValues(f *models.File) { //nolint:cyclop
	if f == nil {
		return
	}
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
	if i := got.IssuedDay; i > 0 {
		f.DateIssuedDay = null.Int16From(i)
	}
	if i := got.IssuedMonth; i > 0 {
		f.DateIssuedMonth = null.Int16From(i)
	}
	if i := got.IssuedYear; i > 0 {
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
	UUID        string `json:"uuid"`         // UUID is the file production UUID.
	Releaser1   string `json:"releaser1"`    // Releaser1 is the first releaser of the file.
	Releaser2   string `json:"releaser2"`    // Releaser2 is the second releaser of the file.
	Title       string `json:"title"`        // Title is the file title.
	Filename    string `json:"filename"`     // Filename is the file name of the download.
	Content     string `json:"content"`      // Content is the file archive content.
	FileType    string `json:"file_type"`    // Type is the file type.
	FileHash    string `json:"file_hash"`    // Hash is the file integrity hash.
	Platform    string `json:"platform"`     // Platform is the file platform.
	Section     string `json:"section"`      // Section is the file section.
	Error       string `json:"error"`        // Error is the error message if the download or record update failed.
	PouetID     int    `json:"id"`           // PouetID is the Pouet prod which ID.
	DemozooID   int    `json:"demozoo_prod"` // DemozooID is the production ID.
	FileSize    int    `json:"file_size"`    // Size is the file size in bytes.
	IssuedYear  int16  `json:"issued_year"`  // Year is the year the file was issued.
	IssuedMonth int16  `json:"issued_month"` // Month is the month the file was issued.
	IssuedDay   int16  `json:"issued_day"`   // Day is the day the file was issued.
}

// Download fetches the download link from Pouet and saves it to the download directory.
// It then runs Update to modify the database record with various metadata from the file and Pouet record API data.
func (got *PouetLink) Download(
	ctx context.Context, sl *slog.Logger, c *echo.Context, db *sql.DB, download dir.Directory,
) error {
	const msg = "pouet link download"
	const format = "%s for id %d: %w"
	id := got.PouetID
	if err := nils.Check(ctx, sl, c, db); err != nil {
		return fmt.Errorf(format, "check", id, err)
	}
	var prod pouet.Production
	if _, err := prod.Get(ctx, id); err != nil {
		sl.Info(msg+" get production error", slog.Int("id", id), slog.Any("error", err))
		return fmt.Errorf(format, "but could not get record", id, err)
	}
	downloadURL := prod.Download
	if downloadURL == "" {
		sl.Info(msg+" offers no download url", slog.Int("id", id))
		return nil
	}
	resp, err := GetFile(ctx, sl, TimeoutLong, downloadURL)
	if err != nil {
		sl.Info(msg+" get file download error",
			slog.Int("id", id), slog.String("url", downloadURL), slog.Any("error", err))
		return fmt.Errorf(format, "but could not get the file download "+downloadURL, id, err)
	}
	base := filepath.Base(downloadURL)
	dst := filepath.Join(download.Path(), got.UUID)
	got.Filename = base
	if err := helper.RenameFileOW(resp.Path, dst); err != nil {
		sameFiles, err := helper.FileMatch(resp.Path, dst)
		if err != nil {
			sl.Info(msg+" got file but cannot rename error", slog.Int("id", id),
				slog.String("dst", dst), slog.Any("error", err))
			return fmt.Errorf(format, "but could not rename the file download to "+dst, id, err)
		}
		if !sameFiles {
			const s = "was successful but will not overwrite the existing file"
			sl.Info(msg+" "+s, slog.Int("id", id), slog.String("dst", dst))
			return fmt.Errorf(format, s+" "+dst, id, ErrExist)
		}
	}
	got.Filename = base
	got.Error = ""
	if i, err := strconv.Atoi(prod.Demozoo); err == nil && i > 0 {
		got.DemozooID = i
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
	if err := got.Stat(ctx, sl, c, db, download); err != nil {
		sl.Info(msg, slog.Int("id", id), slog.Any("error", err))
	}
	return nil
}

// Stat sets the file size, hash, type, and archive content of the file.
// The UUID is used to locate the file in the download directory.
//
//nolint:dupl // intentional similarity
func (got *PouetLink) Stat(
	ctx context.Context, sl *slog.Logger, c *echo.Context, db *sql.DB, download dir.Directory,
) error {
	const format = "pouet link stat %s: %w"
	if err := nils.Check(ctx, sl, c, db); err != nil {
		return fmt.Errorf(format, "check", err)
	}
	name := filepath.Join(download.Path(), got.UUID)
	if got.FileSize == 0 {
		stat, err := os.Stat(name)
		if err != nil {
			return fmt.Errorf(format, "file download "+name, err)
		}
		got.FileSize = int(stat.Size())
	}
	strong, err := helper.StrongIntegrity(name)
	if err != nil {
		return fmt.Errorf(format, "file download strong integrity hash "+name, err)
	}
	got.FileHash = strong
	if got.FileType == "" {
		got.FileType = simple.MagicAsTitle(sl, name)
	}
	return got.ArchiveContent(ctx, sl, c, db, name)
}

// ArchiveContent sets the archive content and readme text of the source file.
func (got *PouetLink) ArchiveContent(
	ctx context.Context, sl *slog.Logger, c *echo.Context, db *sql.DB, src string,
) error {
	const msg = "pouet link archive content"
	const format = msg + " %s: %w"
	if err := nils.Check(ctx, sl, c, db); err != nil {
		return fmt.Errorf(format, "check", err)
	}
	files, err := archive.Lists(ctx, src)
	if err != nil {
		sl.Info(msg+" list caused an error",
			slog.String("src", src), slog.String("filename", got.Filename), slog.Any("error", err))
		return c.JSON(http.StatusOK, got)
	}
	got.Content = strings.Join(files, "\n")
	if err := got.Update(ctx, c, db); err != nil {
		sl.Info(msg + " update caused an error")
	}
	const html = `<p class="text-success">Successful Pouet update</p>`
	return c.HTML(http.StatusOK, html)
}

// Update modifies the database record using data provided by the DemozooLink struct.
// A JSON response is returned with the success status of the update.
func (got *PouetLink) Update(ctx context.Context, c *echo.Context, db *sql.DB) error {
	const format = "pouet link update %s uuid %s: %w"
	uid := got.UUID
	if err := nils.Check(ctx, c, db); err != nil {
		return fmt.Errorf(format, "check", uid, err)
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf(format, "begin tx", uid, err)
	}
	f, err := model.OneByUUID(ctx, tx, true, uid)
	if err != nil {
		return fmt.Errorf(format, "one record by", uid, err)
	}
	got.updateValues(f)
	if _, err = f.Update(ctx, tx, boil.Infer()); err != nil {
		return fmt.Errorf(format, "infer", uid, err)
	}
	if err = tx.Commit(); err != nil {
		return fmt.Errorf(format, "tx commit", uid, err)
	}
	return nil
}

func (got *PouetLink) updateValues(f *models.File) {
	if f == nil {
		return
	}
	if i := got.DemozooID; i > 0 {
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
	if i := got.IssuedDay; i > 0 {
		f.DateIssuedDay = null.Int16From(i)
	}
	if i := got.IssuedMonth; i > 0 {
		f.DateIssuedMonth = null.Int16From(i)
	}
	if i := got.IssuedYear; i > 0 {
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
