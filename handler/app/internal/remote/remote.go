// Package remote provides the remote download and update of artifact data from third-party sources such as API's.
package remote

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Defacto2/server/internal/archive"
	"github.com/Defacto2/server/internal/demozoo"
	"github.com/Defacto2/server/internal/helper"
	"github.com/Defacto2/server/internal/postgres"
	"github.com/Defacto2/server/model"
	"github.com/gabriel-vasile/mimetype"
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
	UUID      string `json:"uuid"`       // UUID is the file production UUID.
	Filename  string `json:"filename"`   // Filename is the file name of the download.
	FileType  string `json:"file_type"`  // Type is the file type.
	FileHash  string `json:"file_hash"`  // Hash is the file integrity hash.
	Content   string `json:"content"`    // Content is the file archive content.
	Readme    string `json:"readme"`     // Readme is the file readme, text or NFO file.
	LinkURL   string `json:"link_url"`   // LinkURL is the download file link used to fetch the file.
	LinkClass string `json:"link_class"` // LinkClass is the download link class provided by Demozoo.
	Error     string `json:"error"`      // Error is the error message if the download or record update failed.
	Github    string `json:"github_repo"`
	YouTube   string `json:"youtube_video"`
	ID        int    `json:"id"`        // ID is the Demozoo production ID.
	FileSize  int    `json:"file_size"` // Size is the file size in bytes.
	Pouet     int    `json:"pouet_prod"`
	Success   bool   `json:"success"` // Success is the success status of the download and record update.
}

func (got *DemozooLink) Download(c echo.Context, downloadDir string) error {
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
		got.LinkClass = link.LinkClass
		got.LinkURL = link.URL
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
		if df.ContentType != "" {
			got.FileType = df.ContentType
		}
		got.Filename = base
		got.LinkURL = link.URL
		got.LinkClass = link.LinkClass
		got.Success = true
		got.Error = ""
		got.Github = prod.GithubRepo()
		got.Pouet = prod.PouetProd()
		got.YouTube = prod.YouTubeVideo()
		return got.Stat(c, downloadDir)
	}
	got.Error = "no usable download links found, they returned 404 or were empty"
	return c.JSON(http.StatusNotModified, got)
}

func (got *DemozooLink) Stat(c echo.Context, downloadDir string) error {
	path := filepath.Join(downloadDir, got.UUID)
	if got.FileSize == 0 {
		stat, err := os.Stat(path)
		if err != nil {
			got.Error = fmt.Errorf("could not stat file, %s: %w", path, err).Error()
			return c.JSON(http.StatusInternalServerError, got)
		}
		got.FileSize = int(stat.Size())
	}
	strong, err := helper.StrongIntegrity(path)
	if err != nil {
		got.Error = fmt.Errorf("could not get strong integrity hash, %s: %w", path, err).Error()
		return c.JSON(http.StatusInternalServerError, got)
	}
	got.FileHash = strong
	if got.FileType == "" {
		m, err := mimetype.DetectFile(path)
		if err != nil {
			return fmt.Errorf("demozoo stat content filemime failure on %q: %w", path, err)
		}
		got.FileType = m.String()
	}
	return got.ArchiveContent(c, path)
}

func (got *DemozooLink) ArchiveContent(c echo.Context, path string) error {
	files, err := archive.List(path, got.Filename)
	if err != nil {
		return c.JSON(http.StatusOK, got)
	}
	got.Readme = archive.Readme(got.Filename, files...)
	got.Content = strings.Join(files, "\n")
	return got.Update(c)
}

func (got DemozooLink) Update(c echo.Context) error {
	uid := got.UUID
	ctx := context.Background()
	db, tx, err := postgres.ConnectTx()
	if err != nil {
		return ErrDB
	}
	defer db.Close()
	f, err := model.OneByUUID(ctx, tx, true, uid)
	if err != nil {
		return fmt.Errorf("demozoolink update by uuid %w: %s", err, uid)
	}
	f.Filename = null.StringFrom(got.Filename)
	f.Filesize = null.Int64From(int64(got.FileSize))
	f.FileMagicType = null.StringFrom(got.FileType)
	f.FileIntegrityStrong = null.StringFrom(got.FileHash)
	f.FileZipContent = null.StringFrom(got.Content)
	rm := strings.TrimSpace(got.Readme)
	f.RetrotxtReadme = null.StringFrom(rm)
	gt := strings.TrimSpace(got.Github)
	f.WebIDGithub = null.StringFrom(gt)
	f.WebIDPouet = null.Int64From(int64(got.Pouet))
	yt := strings.TrimSpace(got.YouTube)
	f.WebIDYoutube = null.StringFrom(yt)
	if _, err = f.Update(ctx, tx, boil.Infer()); err != nil {
		return fmt.Errorf("demozoolink update infer %w: %s", err, uid)
	}
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("demozoolink update commit %w: %s", err, uid)
	}
	return c.JSON(http.StatusOK, got)
}
