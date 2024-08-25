// Package remote provides the remote download and update of artifact data from third-party sources such as API's.
package remote

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Defacto2/server/handler/app/internal/str"
	"github.com/Defacto2/server/internal/archive"
	"github.com/Defacto2/server/internal/demozoo"
	"github.com/Defacto2/server/internal/helper"
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
	//Readme    string `json:"readme"`     // Readme is the file readme, text or NFO file.
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
		got.Error = fmt.Errorf("could not get record %d from demozoo api: %w", got.ID, err).Error()
		return c.JSON(http.StatusInternalServerError, got)
	}
	for _, link := range prod.DownloadLinks {
		if link.URL == "" {
			continue
		}
		df, err := helper.GetFile(link.URL)
		tryNextLink := err != nil || df.Path == ""
		if tryNextLink {
			continue
		}
		base := filepath.Base(link.URL)
		dst := filepath.Join(downloadDir, got.UUID)
		got.Filename = base
		if err := helper.RenameFileOW(df.Path, dst); err != nil {
			sameFiles, err := helper.FileMatch(df.Path, dst)
			if err != nil {
				got.Error = fmt.Errorf("could not rename file, %s: %w", dst, err).Error()
				return c.JSON(http.StatusInternalServerError, got)
			}
			if !sameFiles {
				got.Error = fmt.Errorf("%w, will not overwrite, %s", ErrExist, dst).Error()
				return c.JSON(http.StatusConflict, got)
			}
		}
		size, err := strconv.Atoi(df.ContentLength)
		if err == nil {
			got.FileSize = size
		}

		got.Filename = base
		got.Error = ""

		got.Github = prod.GithubRepo()
		fmt.Printf("github: %q %q\n", got.Github, strings.TrimSpace(got.Github))
		got.Pouet = prod.PouetProd()
		got.YouTube = prod.YouTubeVideo()

		y, m, d := prod.Released()
		got.IssuedYear = int16(y)
		got.IssuedMonth = int16(m)
		got.IssuedDay = int16(d)

		r1, r2 := prod.Groups()
		got.Releaser1 = r1
		got.Releaser2 = r2
		got.Title = prod.Title

		ctext, ccode, cart, caudio := prod.Releasers() // TODO: rename to be more descriptive.
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
			got.Error = fmt.Errorf("could not stat file, %s: %w", name, err).Error()
			return c.JSON(http.StatusInternalServerError, got)
		}
		got.FileSize = int(stat.Size())
	}
	strong, err := helper.StrongIntegrity(name)
	if err != nil {
		got.Error = fmt.Errorf("could not get strong integrity hash, %s: %w", name, err).Error()
		return c.JSON(http.StatusInternalServerError, got)
	}
	got.FileHash = strong
	if got.FileType == "" {
		got.FileType = str.MagicAsTitle(name)
	}
	return got.ArchiveContent(c, db, name)
}

// ArchiveContent sets the archive content and readme text of the source file.
func (got *DemozooLink) ArchiveContent(c echo.Context, db *sql.DB, src string) error {
	files, err := archive.List(src, got.Filename)
	if err != nil {
		return c.JSON(http.StatusOK, got)
	}
	//got.Readme = archive.Readme(got.Filename, files...)
	got.Content = strings.Join(files, "\n")
	return got.Update(c, db)
}

// Update modifies the database record using data provided by the DemozooLink struct.
// A JSON response is returned with the success status of the update.
func (got DemozooLink) Update(c echo.Context, db *sql.DB) error {
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

	// https://pkg.go.dev/github.com/volatiletech/null/v8@v8.1.2#StringFrom

	f.Filename = null.StringFrom(got.Filename)
	f.Filesize = null.Int64From(int64(got.FileSize))
	f.FileMagicType = null.StringFrom(got.FileType)
	f.FileIntegrityStrong = null.StringFrom(got.FileHash)
	f.FileZipContent = null.StringFrom(got.Content)
	// rm := strings.TrimSpace(got.Readme)
	// f.RetrotxtReadme = null.StringFrom(rm)
	gt := strings.TrimSpace(got.Github)
	f.WebIDGithub = null.StringFrom(gt)
	f.WebIDPouet = null.Int64From(int64(got.Pouet))
	yt := strings.TrimSpace(got.YouTube)
	f.WebIDYoutube = null.StringFrom(yt)

	f.DateIssuedDay = null.Int16From(got.IssuedDay)
	f.DateIssuedMonth = null.Int16From(got.IssuedMonth)
	f.DateIssuedYear = null.Int16From(got.IssuedYear)

	f.GroupBrandFor = null.StringFrom(got.Releaser1)
	f.GroupBrandBy = null.StringFrom(got.Releaser2)

	f.RecordTitle = null.StringFrom(got.Title)

	f.CreditAudio = null.StringFrom(strings.Join(got.CreditAudio, ","))
	f.CreditIllustration = null.StringFrom(strings.Join(got.CreditArt, ","))
	f.CreditProgram = null.StringFrom(strings.Join(got.CreditCode, ","))
	f.CreditText = null.StringFrom(strings.Join(got.CreditText, ","))

	f.Platform = null.StringFrom(got.Platform)
	f.Section = null.StringFrom(got.Section)

	if _, err = f.Update(ctx, tx, boil.Infer()); err != nil {
		return fmt.Errorf("demozoolink update infer %w: %s", err, uid)
	}
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("demozoolink update commit %w: %s", err, uid)
	}
	return c.JSON(http.StatusOK, got)
}
