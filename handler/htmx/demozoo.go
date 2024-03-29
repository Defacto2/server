package htmx

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
	"go.uber.org/zap"
)

var (
	ErrDB    = errors.New("database connection is nil")
	ErrExist = errors.New("file already exists")
)

// DemozooProd fetches the multiple download_links values from the
// Demozoo production API and attempts to download and save one of the
// linked files. If multiple links are found, the first link is used as
// they should all point to the same asset.
//
// Both the Demozoo production ID param and the Defacto2 UUID query
// param values are required as params to fetch the production data and
// to save the file to the correct filename.
func DemozooProd(logr *zap.SugaredLogger, c echo.Context, downloadDir string) error {
	if logr == nil {
		return c.String(http.StatusInternalServerError, "Error, demozoo prod logger is nil")
	}
	sid := c.FormValue("demozoo-submission")
	id, err := strconv.Atoi(sid)
	if err != nil {
		return c.String(http.StatusNotAcceptable,
			"The Demozoo production ID must be a numeric value, "+sid)
	}

	db, err := postgres.ConnectDB()
	if err != nil {
		return ErrDB
	}
	defer db.Close()
	ctx := context.Background()

	key, err := model.FindDemozooFile(ctx, db, int64(id))
	if err != nil {
		return c.String(http.StatusServiceUnavailable, "Error, the database query failed")
	}
	if key != 0 {
		// ID to test: 198232
		html := fmt.Sprintf("This Demozoo production is already <a href=\"/f/%s\">in use</a>.", helper.ObfuscateID(key))
		return c.HTML(http.StatusOK, html)
	}

	info, err := DemozooValid(c, id)
	if err != nil {
		return err
	}
	if info == "" {
		return nil
	}
	// ID to test: 66654
	html := `<div class="d-grid gap-2 col-6 mx-auto">`
	html += fmt.Sprintf(`<button type="button" class="btn btn-outline-success">Submit ID %d</button>`, id)
	html += `</div>`
	html += fmt.Sprintf(`<p class="mt-3">%s</p>`, info)
	return c.HTML(http.StatusOK, html)
}

// DemozooValid fetches the first usable download link from the Demozoo API.
// The production ID is validated and the production is checked to see if it
// is suitable for Defacto2. If suitable, the production title and
// author groups are returned.
func DemozooValid(c echo.Context, id int) (string, error) {
	if id < 1 {
		return "", c.String(http.StatusNotAcceptable, fmt.Sprintf("invalid id: %d", id))
	}
	var prod demozoo.Production
	if code, err := prod.Get(id); err != nil {
		return "", c.String(code, err.Error())
	}
	plat, sect := prod.SuperType()
	if plat == -1 || sect == -1 {
		s := []string{}
		for _, p := range prod.Platforms {
			s = append(s, p.Name)
		}
		for _, t := range prod.Types {
			s = append(s, t.Name)
		}
		return "", c.HTML(http.StatusOK,
			fmt.Sprintf("Production %d is probably not suitable for Defacto2.<br>Types: %s",
				id, strings.Join(s, " - ")))
	}

	var ok string
	for _, link := range prod.DownloadLinks {
		if link.URL == "" {
			continue
		}
		ok = link.URL
		break
	}
	if ok == "" {
		return "", c.String(http.StatusOK, "This Demozoo production has no suitable download links.")
	}
	s := []string{fmt.Sprintf("%q", prod.Title)}
	for _, a := range prod.Authors {
		if a.Releaser.IsGroup {
			s = append(s, a.Releaser.Name)
		}
	}
	return strings.Join(s, " "), nil
}

// Production is the response from the task of GetDemozooFile.
//
//nolint:tagliatelle
type Production struct {
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

func (got *Production) Download(c echo.Context, downloadDir string) error {
	var rec demozoo.Production
	if _, err := rec.Get(got.ID); err != nil {
		got.Error = fmt.Errorf("could not get record %d from demozoo api: %w", got.ID, err).Error()
		return c.JSON(http.StatusInternalServerError, got)
	}
	for _, link := range rec.DownloadLinks {
		if link.URL == "" {
			continue
		}
		df, err := helper.DownloadFile(link.URL)
		if err != nil || df.Path == "" {
			// continue, to attempt the next download link
			continue
		}
		base := filepath.Base(link.URL)
		dst := filepath.Join(downloadDir, got.UUID)
		got.Filename = base
		got.LinkClass = link.LinkClass
		got.LinkURL = link.URL
		if err := helper.RenameFileOW(df.Path, dst); err != nil {
			// if the rename file fails, check if the uuid file asset already exists
			// and if it is the same as the downloaded file, if not then return an error.
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
		// get the file size
		size, err := strconv.Atoi(df.ContentLength)
		if err == nil {
			got.FileSize = size
		}
		// get the file type
		if df.ContentType != "" {
			got.FileType = df.ContentType
		}
		got.Filename = base
		got.LinkURL = link.URL
		got.LinkClass = link.LinkClass
		got.Success = true
		got.Error = ""
		// obtain data from the external links
		got.Github = rec.GithubRepo()
		got.Pouet = rec.PouetProd()
		got.YouTube = rec.YouTubeVideo()
		return got.Stat(c, downloadDir)
	}
	got.Error = "no usable download links found, they returned 404 or were empty"
	return c.JSON(http.StatusNotModified, got)
}

func (got *Production) Stat(c echo.Context, downloadDir string) error {
	path := filepath.Join(downloadDir, got.UUID)
	// get the file size if not already set
	if got.FileSize == 0 {
		stat, err := os.Stat(path)
		if err != nil {
			got.Error = fmt.Errorf("could not stat file, %s: %w", path, err).Error()
			return c.JSON(http.StatusInternalServerError, got)
		}
		got.FileSize = int(stat.Size())
	}
	// get the file integrity hash
	strong, err := helper.StrongIntegrity(path)
	if err != nil {
		got.Error = fmt.Errorf("could not get strong integrity hash, %s: %w", path, err).Error()
		return c.JSON(http.StatusInternalServerError, got)
	}
	got.FileHash = strong
	// get the file type if not already set
	if got.FileType == "" {
		m, err := mimetype.DetectFile(path)
		if err != nil {
			return fmt.Errorf("content filemime failure on %q: %w", path, err)
		}
		got.FileType = m.String()
	}
	return got.ArchiveContent(c, path)
}

func (got *Production) ArchiveContent(c echo.Context, path string) error {
	files, err := archive.List(path, got.Filename)
	if err != nil {
		return c.JSON(http.StatusOK, got)
	}
	got.Readme = archive.Readme(got.Filename, files...)
	got.Content = strings.Join(files, "\n")
	return got.Update(c)
}

func (got Production) Update(c echo.Context) error {
	uid := got.UUID
	db, err := postgres.ConnectDB()
	if err != nil {
		return ErrDB
	}
	defer db.Close()
	ctx := context.Background()
	f, err := model.OneByUUID(ctx, db, true, uid)
	if err != nil {
		return err
	}
	f.Filename = null.StringFrom(got.Filename)
	f.Filesize = int64(got.FileSize)
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
	if _, err = f.Update(ctx, db, boil.Infer()); err != nil {
		return err
	}
	return c.JSON(http.StatusOK, got)
}
