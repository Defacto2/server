package app

import (
	"context"
	"crypto/sha512"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Defacto2/releaser"
	"github.com/Defacto2/releaser/initialism"
	"github.com/Defacto2/server/handler/download"
	"github.com/Defacto2/server/handler/sess"
	"github.com/Defacto2/server/internal/archive"
	"github.com/Defacto2/server/internal/cache"
	"github.com/Defacto2/server/internal/command"
	"github.com/Defacto2/server/internal/helper"
	"github.com/Defacto2/server/internal/postgres"
	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/Defacto2/server/internal/pouet"
	"github.com/Defacto2/server/internal/sixteen"
	"github.com/Defacto2/server/internal/web"
	"github.com/Defacto2/server/internal/zoo"
	"github.com/Defacto2/server/model"
	"github.com/gabriel-vasile/mimetype"
	"github.com/google/uuid"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"go.uber.org/zap"
	"google.golang.org/api/idtoken"
)

var ErrExist = errors.New("file already exists")

const (
	demo    = "demo"
	limit   = 198 // per-page record limit
	page    = 1   // default page number
	records = "records"
	sep     = ";"
	txt     = ".txt" // txt file extension
)

// ArtifactErr renders the error page for the artifact links.
func ArtifactErr(logr *zap.SugaredLogger, c echo.Context, id string) error {
	const name = "status"
	if logr == nil {
		return InternalErr(logr, c, name, ErrZap)
	}
	if c == nil {
		return InternalErr(logr, c, name, ErrCxt)
	}
	data := empty(c)
	data["title"] = fmt.Sprintf("%d error, artifact page not found", http.StatusNotFound)
	data["description"] = fmt.Sprintf("HTTP status %d error", http.StatusNotFound)
	data["code"] = http.StatusNotFound
	data["logo"] = "Artifact not found"
	data["alert"] = fmt.Sprintf("Artifact %q cannot be found", strings.ToLower(id))
	data["probl"] = "The artifact page does not exist, there is probably a typo with the URL."
	data["uriOkay"] = "f/"
	data["uriErr"] = id
	err := c.Render(http.StatusNotFound, name, data)
	if err != nil {
		return InternalErr(logr, c, name, err)
	}
	return nil
}

// Artist is the handler for the Artist sceners page.
func Artist(logr *zap.SugaredLogger, c echo.Context) error {
	data := empty(c)
	title := "Pixel artists and graphic designers"
	data["title"] = title
	data["logo"] = title
	data["h1"] = title
	data["description"] = demo
	return scener(logr, c, postgres.Artist, data)
}

// BadRequestErr is the handler for handling Bad Request Errors, caused by invalid user input
// or a malformed client requests.
func BadRequestErr(logr *zap.SugaredLogger, c echo.Context, uri string, err error) error {
	const code = http.StatusBadRequest
	if logr == nil {
		zapNil(err)
	} else if err != nil {
		logr.Errorf("%d error for %q: %s", code, uri, err)
	}
	// render the fallback, text only error page
	if c == nil {
		if logr == nil {
			zapNil(fmt.Errorf("%w: internalerr", ErrCxt))
		} else {
			logr.Errorf("%s: %s", ErrTmpl, ErrCxt)
		}
		return echo.NewHTTPError(http.StatusInternalServerError,
			fmt.Errorf("%w: handler app status", ErrCxt))
	}
	// render a user friendly error page
	data := empty(c)
	data["description"] = fmt.Sprintf("HTTP status %d error", code)
	data["title"] = "400 error, there is a complication"
	data["code"] = code
	data["logo"] = "Client error"
	data["alert"] = "Something went wrong, " + err.Error()
	data["probl"] = "It might be a settings or configuration problem or a legacy browser issue."
	data["uriErr"] = uri
	if err := c.Render(code, "status", data); err != nil {
		if logr != nil {
			logr.Errorf("%s: %s", ErrTmpl, err)
		}
		return echo.NewHTTPError(http.StatusInternalServerError, ErrTmpl)
	}
	return nil
}

// BBS is the handler for the BBS page ordered by the most files.
func BBS(logr *zap.SugaredLogger, c echo.Context) error {
	return bbsHandler(logr, c, model.Prolific)
}

// BBSAZ is the handler for the BBS page ordered alphabetically.
func BBSAZ(logr *zap.SugaredLogger, c echo.Context) error {
	return bbsHandler(logr, c, model.Alphabetical)
}

// BBSYear is the handler for the BBS page ordered by the year.
func BBSYear(logr *zap.SugaredLogger, c echo.Context) error {
	return bbsHandler(logr, c, model.Oldest)
}

// Checksum is the handler for the Checksum file record page.
func Checksum(logr *zap.SugaredLogger, c echo.Context, id string) error {
	if err := download.Checksum(c, id); err != nil {
		return DownloadErr(logr, c, "sum", err)
	}
	return nil
}

// Code is the handler for the Coder sceners page.
func Coder(logr *zap.SugaredLogger, c echo.Context) error {
	data := empty(c)
	title := "Coder and programmers"
	data["title"] = title
	data["logo"] = title
	data["h1"] = title
	data["description"] = demo
	return scener(logr, c, postgres.Writer, data)
}

// DatabaseErr is the handler for database connection issues.
// A HTTP 503 Service Unavailable error is returned, to reflect the database
// connection issue but where the server is still running and usable for the client.
func DatabaseErr(logr *zap.SugaredLogger, c echo.Context, uri string, err error) error {
	const unavailable = http.StatusServiceUnavailable
	if logr == nil {
		zapNil(err)
	} else if err != nil {
		logr.Errorf("%d error for %q: %s", unavailable, uri, err)
	}
	// render the fallback, text only error page
	if c == nil {
		if logr == nil {
			zapNil(fmt.Errorf("%w: databaserr", ErrCxt))
		} else {
			logr.Warnf("%s: %s", ErrTmpl, ErrCxt)
		}
		return echo.NewHTTPError(http.StatusInternalServerError,
			fmt.Errorf("%w: handler app status", ErrCxt))
	}
	// render a user friendly error page
	data := empty(c)
	data["description"] = fmt.Sprintf("HTTP status %d error", unavailable)
	data["title"] = fmt.Sprintf("%d error, there is a complication", unavailable)
	data["code"] = fmt.Sprintf("%d service unavailable", unavailable)
	data["logo"] = "Database error"
	data["alert"] = "Cannot connect to the database!"
	data["uriErr"] = ""
	data["probl"] = "This is not your fault, but the server cannot communicate with the database to display this page."
	if err := c.Render(unavailable, "status", data); err != nil {
		if logr != nil {
			logr.Errorf("%s: %s", ErrTmpl, err)
		}
		return echo.NewHTTPError(unavailable, ErrTmpl)
	}
	return nil
}

// Download is the handler for the Download file record page.
func Download(logr *zap.SugaredLogger, c echo.Context, path string) error {
	d := download.Download{
		Inline: false,
		Path:   path,
	}
	err := d.HTTPSend(logr, c)
	if err != nil {
		return DownloadErr(logr, c, "d", err)
	}
	return nil
}

// DownloadErr is the handler for missing download files and database ID errors.
func DownloadErr(logr *zap.SugaredLogger, c echo.Context, uri string, err error) error {
	const code = http.StatusNotFound
	id := c.Param("id")
	if logr == nil {
		zapNil(err)
	} else if err != nil {
		logr.Errorf("%d error for %q: %s", code, id, err)
	}
	// render the fallback, text only error page
	if c == nil {
		if logr == nil {
			zapNil(fmt.Errorf("%w: downloaderr", ErrCxt))
		} else {
			logr.Errorf("%s: %s", ErrTmpl, ErrCxt)
		}
		return echo.NewHTTPError(http.StatusInternalServerError,
			fmt.Errorf("%w: handler app status", ErrCxt))
	}
	// render a user friendly error page
	data := empty(c)
	data["description"] = fmt.Sprintf("HTTP status %d error", code)
	data["title"] = "404 download error"
	data["code"] = code
	data["logo"] = "Download problem"
	data["alert"] = "Cannot send you this download"
	data["probl"] = "The download you are looking for might have been removed, " +
		"had its filename changed, or is temporarily unavailable. " +
		"Is the URL correct?"
	data["uriErr"] = strings.Join([]string{uri, id}, "/")
	if err := c.Render(code, "status", data); err != nil {
		if logr != nil {
			logr.Errorf("%s: %s", ErrTmpl, err)
		}
		return echo.NewHTTPError(http.StatusInternalServerError, ErrTmpl)
	}
	return nil
}

// FTP is the handler for the FTP page.
func FTP(logr *zap.SugaredLogger, c echo.Context) error {
	const title, name = "FTP", "ftp"
	if logr == nil {
		return InternalErr(logr, c, name, ErrZap)
	}
	data := empty(c)
	const lead = "FTP sites are historical, internet-based file servers for uploading and downloading \"elite\" scene releases."
	const key = "releasers"
	data["title"] = title
	data["description"] = lead
	data["logo"] = "FTP sites"
	data["h1"] = title
	data["lead"] = lead
	// releaser.html specific data items
	data["itemName"] = name
	data[key] = model.Releasers{}
	data["stats"] = map[string]string{}
	ctx := context.Background()
	db, err := postgres.ConnectDB()
	if err != nil {
		return InternalErr(logr, c, name, err)
	}
	defer db.Close()
	r := model.Releasers{}
	if err := r.FTP(ctx, db); err != nil {
		return DatabaseErr(logr, c, name, err)
	}
	m := model.Summary{}
	if err := m.FTP(ctx, db); err != nil {
		return DatabaseErr(logr, c, name, err)
	}
	data[key] = r
	data["stats"] = map[string]string{
		"pubs":   fmt.Sprintf("%d sites", len(r)),
		"issues": string(ByteFileS("file artifact", m.SumCount.Int64, m.SumBytes.Int64)),
		"years":  fmt.Sprintf("%d - %d", m.MinYear.Int16, m.MaxYear.Int16),
	}
	err = c.Render(http.StatusOK, name, data)
	if err != nil {
		return InternalErr(logr, c, name, err)
	}
	return nil
}

// File is the handler for the artifact categories page.
func File(logr *zap.SugaredLogger, c echo.Context, stats bool) error {
	const title, name = "Artifact categories", "file"
	if logr == nil {
		return InternalErr(logr, c, name, ErrZap)
	}
	data := empty(c)
	data["title"] = title
	data["description"] = "A table of contents for the collection."
	data["logo"] = title
	data["h1"] = title
	data["lead"] = "This page shows the categories and platforms in the collection of file artifacts."
	data["stats"] = stats
	data["counter"] = Stats{}

	data, err := fileWStats(data, stats)
	if err != nil {
		logr.Warn(err)
		data["dbError"] = true
	}
	err = c.Render(http.StatusOK, name, data)
	if err != nil {
		return InternalErr(logr, c, name, err)
	}
	return nil
}

// Files is the handler for the list and preview of the files page.
// The uri is the category or collection of files to display.
// The page is the page number of the results to display.
func Files(logr *zap.SugaredLogger, c echo.Context, uri, page string) error {
	if logr == nil {
		return InternalErr(logr, c, "files", ErrZap)
	}
	// check the uri is valid
	if !Valid(uri) {
		return FilesErr(logr, c, uri)
	}
	// check the page is valid
	if page == "" {
		return files(logr, c, uri, 1)
	}
	p, err := strconv.Atoi(page)
	if err != nil {
		return PageErr(logr, c, uri, page)
	}
	return files(logr, c, uri, p)
}

// FilesErr renders the files error page for the Files menu and categories.
// It provides different error messages to the standard error page.
func FilesErr(logr *zap.SugaredLogger, c echo.Context, uri string) error {
	const name = "status"
	if logr == nil {
		return InternalErr(logr, c, name, ErrZap)
	}
	if c == nil {
		return InternalErr(logr, c, name, ErrCxt)
	}
	data := empty(c)
	data["title"] = fmt.Sprintf("%d error, files page not found", http.StatusNotFound)
	data["description"] = fmt.Sprintf("HTTP status %d error", http.StatusNotFound)
	data["code"] = http.StatusNotFound
	data["logo"] = "Files not found"
	data["alert"] = "Files page cannot be found"
	data["probl"] = "The files category or menu option does not exist, there is probably a typo with the URL."
	data["uriOkay"] = "files/"
	data["uriErr"] = uri
	err := c.Render(http.StatusNotFound, name, data)
	if err != nil {
		return InternalErr(logr, c, name, err)
	}
	return nil
}

func FilesDeletions(logr *zap.SugaredLogger, c echo.Context, page string) error {
	if logr == nil {
		return InternalErr(logr, c, "filesDeletions", ErrZap)
	}
	uri := deletions.String()
	// check the uri is valid
	if !Valid(uri) {
		return FilesErr(logr, c, uri)
	}
	// check the page is valid
	if page == "" {
		return files(logr, c, uri, 1)
	}
	p, err := strconv.Atoi(page)
	if err != nil {
		return PageErr(logr, c, uri, page)
	}
	return files(logr, c, uri, p)
}

func FilesUnwanted(logr *zap.SugaredLogger, c echo.Context, page string) error {
	if logr == nil {
		return InternalErr(logr, c, "filesUnwanted", ErrZap)
	}
	uri := unwanted.String()
	// check the uri is valid
	if !Valid(uri) {
		return FilesErr(logr, c, uri)
	}
	// check the page is valid
	if page == "" {
		return files(logr, c, uri, 1)
	}
	p, err := strconv.Atoi(page)
	if err != nil {
		return PageErr(logr, c, uri, page)
	}
	return files(logr, c, uri, p)
}

// Files is the handler for the list and preview of the files page.
// The uri is the category or collection of files to display.
// The page is the page number of the results to display.
func FilesWaiting(logr *zap.SugaredLogger, c echo.Context, page string) error {
	if logr == nil {
		return InternalErr(logr, c, "filesWaiting", ErrZap)
	}
	uri := forApproval.String()
	// check the uri is valid
	if !Valid(uri) {
		return FilesErr(logr, c, uri)
	}
	// check the page is valid
	if page == "" {
		return files(logr, c, uri, 1)
	}
	p, err := strconv.Atoi(page)
	if err != nil {
		return PageErr(logr, c, uri, page)
	}
	return files(logr, c, uri, p)
}

// ForbiddenErr is the handler for handling Forbidden Errors, caused by clients requesting
// pages that they do not have permission to access.
func ForbiddenErr(logr *zap.SugaredLogger, c echo.Context, uri string, err error) error {
	const code = http.StatusForbidden
	if logr == nil {
		zapNil(err)
	} else if err != nil {
		logr.Errorf("%d error for %q: %s", code, uri, err)
	}
	// render the fallback, text only error page
	if c == nil {
		if logr == nil {
			zapNil(fmt.Errorf("%w: internalerr", ErrCxt))
		} else {
			logr.Errorf("%s: %s", ErrTmpl, ErrCxt)
		}
		return echo.NewHTTPError(http.StatusInternalServerError,
			fmt.Errorf("%w: handler app status", ErrCxt))
	}
	// render a user friendly error page
	data := empty(c)
	data["description"] = fmt.Sprintf("HTTP status %d error", code)
	data["title"] = "403, forbidden"
	data["code"] = code
	data["logo"] = "Forbidden"
	data["alert"] = "This page is locked"
	data["probl"] = fmt.Sprintf("This page is not intended for the general public, %s.", err.Error())
	data["uriErr"] = uri
	if err := c.Render(code, "status", data); err != nil {
		if logr != nil {
			logr.Errorf("%s: %s", ErrTmpl, err)
		}
		return echo.NewHTTPError(http.StatusInternalServerError, ErrTmpl)
	}
	return nil
}

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

// GetDemozooLink fetches the multiple download_links values from the
// Demozoo production API and attempts to download and save one of the
// linked files. If multiple links are found, the first link is used as
// they should all point to the same asset.
//
// Both the Demozoo production ID param and the Defacto2 UUID query
// param values are required as params to fetch the production data and
// to save the file to the correct filename.
func GetDemozooLink(logr *zap.SugaredLogger, c echo.Context, downloadDir string) error {
	const name = "get/demozoo/download"
	if logr == nil {
		return InternalErr(logr, c, name, ErrZap)
	}
	got := DemozooLink{
		Filename:  "",
		FileSize:  0,
		FileType:  "",
		FileHash:  "",
		Content:   "",
		Readme:    "",
		LinkURL:   "",
		LinkClass: "",
		Success:   false,
		Error:     "",
	}
	sid := c.Param("id")
	id, err := strconv.Atoi(sid)
	if err != nil {
		got.Error = "demozoo id must be a numeric value, " + sid
		return c.JSON(http.StatusBadRequest, got)
	}
	got.ID = id
	sid = c.QueryParam("uuid")
	if err = uuid.Validate(sid); err != nil {
		got.Error = "uuid syntax did not validate, " + sid
		return c.JSON(http.StatusBadRequest, got)
	}
	got.UUID = sid
	return got.Download(c, downloadDir)
}

func (got *DemozooLink) Download(c echo.Context, downloadDir string) error {
	var rec zoo.Demozoo
	if err := rec.Get(got.ID); err != nil {
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

func (got *DemozooLink) Stat(c echo.Context, downloadDir string) error {
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

// GoogleCallback is the handler for the Google OAuth2 callback page to verify
// the [Google ID token].
//
// [Google ID token]: https://developers.google.com/identity/gsi/web/guides/verify-google-id-token
func GoogleCallback(logr *zap.SugaredLogger, c echo.Context, clientID string, maxAge int, accounts ...[48]byte) error {
	const name = "google/callback"
	if logr == nil {
		return InternalErr(logr, c, name, ErrZap)
	}

	// Cross-Site Request Forgery cookie token
	const csrf = "g_csrf_token"
	cookie, err := c.Cookie(csrf)
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			return c.Redirect(http.StatusForbidden, "/signin")
		}
		return BadRequestErr(logr, c, name, err)
	}
	token := cookie.Value

	// Cross-Site Request Forgery post token
	bodyToken := c.FormValue(csrf)
	if token != bodyToken {
		return BadRequestErr(logr, c, name, ErrMisMatch)
	}

	// Create a new token verifier.
	// https://pkg.go.dev/google.golang.org/api/idtoken
	ctx := context.Background()
	validator, err := idtoken.NewValidator(ctx)
	if err != nil {
		return BadRequestErr(logr, c, name, err)
	}

	// Verify the ID token and using the client ID from the Google API.
	credential := c.FormValue("credential")
	playload, err := validator.Validate(ctx, credential, clientID)
	if err != nil {
		return BadRequestErr(logr, c, name, err)
	}

	// Verify the sub value against the list of allowed accounts.
	check := false
	if sub, ok := playload.Claims["sub"]; ok {
		for _, account := range accounts {
			if id, ok := sub.(string); ok && sha512.Sum384([]byte(id)) == account {
				check = true
				break
			}
		}
	}
	if !check {
		fullname := playload.Claims["name"]
		sub := playload.Claims["sub"]
		return ForbiddenErr(logr, c, name,
			fmt.Errorf("%w %s. "+
				"If this is a mistake, contact Defacto2 admin and give them this Google account ID: %s",
				ErrUser, fullname, sub))
	}

	if err = sessionHandler(logr, c, maxAge, playload.Claims); err != nil {
		return BadRequestErr(logr, c, name, err)
	}
	return c.Redirect(http.StatusFound, "/")
}

// History is the handler for the History page.
func History(logr *zap.SugaredLogger, c echo.Context) error {
	const name = "history"
	if logr == nil {
		return InternalErr(logr, c, name, ErrZap)
	}
	const lead = "In the past, alternative iterations of the name have included" +
		" De Facto, DF, DeFacto, Defacto II, Defacto 2, and the defacto2.com domain."
	const h1 = "The history of the brand"
	data := empty(c)
	data["carousel"] = "#carouselDf2Artpacks"
	data["description"] = lead
	data["logo"] = "The history of Defacto"
	data["h1"] = h1
	data["lead"] = lead
	data["title"] = h1
	err := c.Render(http.StatusOK, name, data)
	if err != nil {
		return InternalErr(logr, c, name, err)
	}
	return nil
}

// Index is the handler for the Home page.
func Index(logr *zap.SugaredLogger, c echo.Context) error {
	const name = "index"
	if logr == nil {
		return InternalErr(logr, c, name, ErrZap)
	}
	const lead = "the website preserving the historic PC cracking scene subculture. " +
		"It covers digital artifacts including text files, demos, music, art, " +
		"magazines, and other projects."
	const desc = "Defacto2 is " + lead
	data := empty(c)
	data["title"] = "Home"
	data["description"] = desc
	data["h1"] = "Welcome,"
	data["milestones"] = Collection()
	{
		// get the signed in given name
		sess, err := session.Get(sess.Name, c)
		if err == nil {
			if name, ok := sess.Values["givenName"]; ok {
				if nameStr, ok := name.(string); ok && nameStr != "" {
					data["h1"] = "Welcome, " + nameStr
				}
			}
		}
	}
	err := c.Render(http.StatusOK, name, data)
	if err != nil {
		return InternalErr(logr, c, name, err)
	}
	return nil
}

// Inline is the handler for the Download file record page.
func Inline(logr *zap.SugaredLogger, c echo.Context, path string) error {
	d := download.Download{
		Inline: true,
		Path:   path,
	}
	err := d.HTTPSend(logr, c)
	if err != nil {
		return DownloadErr(logr, c, "v", err)
	}
	return nil
}

// InternalErr is the handler for handling Internal Server Errors, caused by programming bugs or crashes.
// The uri string is the part of the URL that caused the error.
// The optional error value is logged using the zap sugared logger.
// If the zap logger is nil then the error page is returned but no error is logged.
// If the echo context is nil then a user hostile, fallback error in raw text is returned.
func InternalErr(logr *zap.SugaredLogger, c echo.Context, uri string, err error) error {
	const code = http.StatusInternalServerError
	if logr == nil {
		zapNil(err)
	} else if err != nil {
		logr.Errorf("%d error for %q: %s", code, uri, err)
	}
	// render the fallback, text only error page
	if c == nil {
		if logr == nil {
			zapNil(fmt.Errorf("%w: internalerr", ErrCxt))
		} else {
			logr.Errorf("%s: %s", ErrTmpl, ErrCxt)
		}
		return echo.NewHTTPError(http.StatusInternalServerError,
			fmt.Errorf("%w: handler app status", ErrCxt))
	}
	// render a user friendly error page
	data := empty(c)
	data["description"] = fmt.Sprintf("HTTP status %d error", code)
	data["title"] = "500 error, there is a complication"
	data["code"] = code
	data["logo"] = "Server error"
	data["alert"] = "Something crashed!"
	data["probl"] = "This is not your fault," +
		" but the server encountered an internal error or misconfiguration and cannot display this page."
	data["uriErr"] = uri
	if err := c.Render(code, "status", data); err != nil {
		if logr != nil {
			logr.Errorf("%s: %s", ErrTmpl, err)
		}
		return echo.NewHTTPError(http.StatusInternalServerError, ErrTmpl)
	}
	return nil
}

// Interview is the handler for the People Interviews page.
func Interview(logr *zap.SugaredLogger, c echo.Context) error {
	const title, name = "Interviews with sceners", "interview"
	if logr == nil {
		return InternalErr(logr, c, name, ErrZap)
	}
	data := empty(c)
	data["title"] = title
	data["description"] = "Discussions with scene members."
	data["logo"] = title
	data["h1"] = title
	data["lead"] = "Here is a centralized page for the site's discussions and unedited" +
		" interviews with sceners, crackers, and demo makers. Currently, incomplete."
	data["interviews"] = Interviewees()
	err := c.Render(http.StatusOK, name, data)
	if err != nil {
		return InternalErr(logr, c, name, err)
	}
	return nil
}

// Magazine is the handler for the Magazine page.
func Magazine(logr *zap.SugaredLogger, c echo.Context) error {
	return mag(logr, c, true)
}

// MagazineAZ is the handler for the Magazine page ordered chronologically.
func MagazineAZ(logr *zap.SugaredLogger, c echo.Context) error {
	return mag(logr, c, false)
}

// Musician is the handler for the Musiciansceners page.
func Musician(logr *zap.SugaredLogger, c echo.Context) error {
	data := empty(c)
	title := "Musicians and composers"
	data["title"] = title
	data["logo"] = title
	data["h1"] = title
	data["description"] = demo
	return scener(logr, c, postgres.Musician, data)
}

// PageErr renders the files page error page for the Files menu and categories.
// It provides different error messages to the standard error page.
func PageErr(logr *zap.SugaredLogger, c echo.Context, uri, page string) error {
	const name = "status"
	if logr == nil {
		return InternalErr(logr, c, name, ErrZap)
	}
	if c == nil {
		return InternalErr(logr, c, name, ErrCxt)
	}
	data := empty(c)
	data["title"] = fmt.Sprintf("%d error, files page not found", http.StatusNotFound)
	data["description"] = fmt.Sprintf("HTTP status %d error", http.StatusNotFound)
	data["code"] = http.StatusNotFound
	data["logo"] = "Page not found"
	data["alert"] = fmt.Sprintf("Files %s page does not exist", uri)
	data["probl"] = "The files page does not exist, there is probably a typo with the URL."
	data["uriOkay"] = fmt.Sprintf("files/%s/", uri)
	data["uriErr"] = page
	err := c.Render(http.StatusNotFound, name, data)
	if err != nil {
		return InternalErr(logr, c, name, err)
	}
	return nil
}

// PlatformEdit handles the post submission for the Platform selection field.
func PlatformEdit(logr *zap.SugaredLogger, c echo.Context) error {
	const name = "editor platform"
	if logr == nil {
		return InternalErr(logr, c, name, ErrZap)
	}

	var f Form
	if err := c.Bind(&f); err != nil {
		return badRequest(c, err)
	}
	r, err := model.EditFind(f.ID)
	if err != nil {
		return err
	}
	if err = model.UpdatePlatform(c, int64(f.ID), f.Value); err != nil {
		return badRequest(c, err)
	}
	return c.JSON(http.StatusOK, r)
}

// PlatformTagInfo handles the POST submission for the platform and tag info.
func PlatformTagInfo(logr *zap.SugaredLogger, c echo.Context) error {
	const name = "editor platform tag info"
	if logr == nil {
		return InternalErr(logr, c, name, ErrZap)
	}
	var f Form
	if err := c.Bind(&f); err != nil {
		return badRequest(c, err)
	}
	info, err := model.GetPlatformTagInfo(c, f.Platform, f.Tag)
	if err != nil {
		return badRequest(c, err)
	}
	return c.String(http.StatusOK, info)
}

// PostIntro handles the POST request for the intro upload form.
func PostIntro(logr *zap.SugaredLogger, c echo.Context) error {
	const name = "post intro"
	if logr == nil {
		return InternalErr(logr, c, name, ErrZap)
	}
	x, err := c.FormParams()
	if err != nil {
		return InternalErr(logr, c, name, err)
	}
	return c.JSON(http.StatusOK, x)
}

// PostDesc is the handler for the Search for file descriptions form post page.
func PostDesc(logr *zap.SugaredLogger, c echo.Context, input string) error {
	const name = "files"
	ctx := context.Background()
	db, err := postgres.ConnectDB()
	if err != nil {
		return DatabaseErr(logr, c, name, err)
	}
	defer db.Close()

	// input := c.FormValue("search-term-query")
	terms := helper.SearchTerm(input)
	rel := model.Files{}

	fs, _ := rel.SearchDescription(ctx, db, terms)
	if err != nil {
		return DatabaseErr(logr, c, name, err)
	}
	d := Descriptions.postStats(ctx, db, terms)
	s := strings.Join(terms, ", ")
	data := emptyFiles(c)
	data["title"] = "Title and description results"
	data["h1"] = "Title and description search"
	data["lead"] = fmt.Sprintf("Results for %q", s)
	data["logo"] = s + " results"
	data["description"] = "Title and description search results for " + s + "."
	data["unknownYears"] = false
	data[records] = fs
	data["stats"] = d
	err = c.Render(http.StatusOK, "files", data)
	if err != nil {
		return InternalErr(logr, c, name, err)
	}
	return nil
}

// PostFilename is the handler for the Search for filenames form post page.
func PostFilename(logr *zap.SugaredLogger, c echo.Context) error {
	return PostName(logr, c, Filenames)
}

// PostName is the handler for the Search for filenames form post page.
func PostName(logr *zap.SugaredLogger, c echo.Context, mode FileSearch) error {
	const name = "files"
	ctx := context.Background()
	db, err := postgres.ConnectDB()
	if err != nil {
		return DatabaseErr(logr, c, name, err)
	}
	defer db.Close()

	input := c.FormValue("search-term-query")
	terms := helper.SearchTerm(input)
	rel := model.Files{}

	fs, _ := rel.SearchFilename(ctx, db, terms)
	if err != nil {
		return DatabaseErr(logr, c, name, err)
	}
	d := mode.postStats(ctx, db, terms)
	s := strings.Join(terms, ", ")
	data := emptyFiles(c)
	data["title"] = "Filename results"
	data["h1"] = "Filename search"
	data["lead"] = fmt.Sprintf("Results for %q", s)
	data["logo"] = s + " results"
	data["description"] = "Filename search results for " + s + "."
	data["unknownYears"] = false
	data[records] = fs
	data["stats"] = d
	err = c.Render(http.StatusOK, "files", data)
	if err != nil {
		return InternalErr(logr, c, name, err)
	}
	return nil
}

// PouetCache parses the cached data for the Pouet production votes.
// If the cache is valid it is returned as JSON response.
// If the cache is invalid or corrupt an error will be returned
// and a API request should be made to Pouet.
func PouetCache(c echo.Context, data string) error {
	if data == "" {
		return nil
	}
	pv := pouet.Votes{}
	x := strings.Split(data, sep)
	const expect = 4
	if l := len(x); l != expect {
		return fmt.Errorf("%w: %d, want %d", ErrData, l, expect)
	}
	stars, err := strconv.ParseFloat(x[0], 64)
	if err != nil {
		return fmt.Errorf("%w: %s", err, x[0])
	}
	vd, err := strconv.Atoi(x[1])
	if err != nil {
		return fmt.Errorf("%w: %s", err, x[1])
	}
	vu, err := strconv.Atoi(x[2])
	if err != nil {
		return fmt.Errorf("%w: %s", err, x[2])
	}
	vm, err := strconv.Atoi(x[3])
	if err != nil {
		return fmt.Errorf("%w: %s", err, x[3])
	}
	pv.Stars = stars
	pv.VotesDown = uint64(vd)
	pv.VotesUp = uint64(vu)
	pv.VotesMeh = uint64(vm)
	if err = c.JSON(http.StatusOK, pv); err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return nil
}

// ProdPouet is the handler for the Pouet prod JSON page.
func ProdPouet(logr *zap.SugaredLogger, c echo.Context, id string) error {
	const name = "pouet"
	if logr == nil {
		return InternalErr(logr, c, name, ErrZap)
	}
	p := pouet.Pouet{}
	i, err := strconv.Atoi(id)
	if err != nil {
		return c.String(http.StatusNotFound, err.Error())
	}
	if err = p.Uploader(i); err != nil {
		return c.String(http.StatusNotFound, err.Error())
	}
	if err = c.JSON(http.StatusOK, p); err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return nil
}

// ProdZoo is the handler for the Demozoo production JSON page.
func ProdZoo(logr *zap.SugaredLogger, c echo.Context, id string) error {
	const name = "demozoo"
	if logr == nil {
		return InternalErr(logr, c, name, ErrZap)
	}
	data := zoo.Demozoo{}
	i, err := strconv.Atoi(id)
	if err != nil {
		return c.String(http.StatusNotFound, err.Error())
	}
	if err = data.Get(i); err != nil {
		return c.String(http.StatusNotFound, err.Error())
	}
	if err = c.JSON(http.StatusOK, data); err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return nil
}

// Interview is the handler for the People Interviews page.
func Reader(logr *zap.SugaredLogger, c echo.Context) error {
	const title, name = "Textfile reader", "reader"
	if logr == nil {
		return InternalErr(logr, c, name, ErrZap)
	}
	data := empty(c)
	data["title"] = title
	data["description"] = "Discussions with scene members."
	data["logo"] = title
	data["h1"] = title
	data["lead"] = "An incomplete list of discussions and unedited interviews with sceners," +
		" crackers and demo makers."
	data["interviews"] = Interviewees()
	err := c.Render(http.StatusOK, name, data)
	if err != nil {
		return InternalErr(logr, c, name, err)
	}
	return nil
}

// ReadmeDel handles the post submission for the Delete readme asset button.
func ReadmeDel(logr *zap.SugaredLogger, c echo.Context, downloadDir string) error {
	const name = "editor readme delete"
	if logr == nil {
		return InternalErr(logr, c, name, ErrZap)
	}

	var f Form
	if err := c.Bind(&f); err != nil {
		return badRequest(c, err)
	}
	r, err := model.EditFind(f.ID)
	if err != nil {
		return err
	}
	if err = command.RemoveMe(r.UUID.String, downloadDir); err != nil {
		return badRequest(c, err)
	}
	return c.JSON(http.StatusOK, r)
}

// ReadmePost handles the post submission for the Readme in archive.
func ReadmePost(logr *zap.SugaredLogger, c echo.Context, downloadDir string) error {
	const name = "editor readme"
	if logr == nil {
		return InternalErr(logr, c, name, ErrZap)
	}

	var f Form
	if err := c.Bind(&f); err != nil {
		return badRequest(c, err)
	}
	r, err := model.EditFind(f.ID)
	if err != nil {
		return badRequest(c, err)
	}

	list := strings.Split(r.FileZipContent.String, "\n")
	target := ""
	for _, x := range list {
		s := strings.TrimSpace(x)
		if s == "" {
			continue
		}
		if strings.EqualFold(s, f.Target) {
			target = s
		}
	}
	if target == "" {
		return badRequest(c, ErrTarget)
	}

	src := filepath.Join(downloadDir, r.UUID.String)
	dst := filepath.Join(downloadDir, r.UUID.String+txt)
	ext := filepath.Ext(strings.ToLower(r.Filename.String))
	err = command.ExtractOne(logr, src, dst, ext, target)
	if err != nil {
		return badRequest(c, err)
	}
	return c.JSON(http.StatusOK, r)
}

// ReadmeToggle handles the post submission for the Hide readme from view toggle.
func ReadmeToggle(logr *zap.SugaredLogger, c echo.Context) error {
	const name = "editor readme toggle"
	if logr == nil {
		return InternalErr(logr, c, name, ErrZap)
	}

	var f Form
	if err := c.Bind(&f); err != nil {
		return badRequest(c, err)
	}
	if err := model.UpdateNoReadme(c, int64(f.ID), f.Readme); err != nil {
		return badRequest(c, err)
	}
	return c.JSON(http.StatusOK, f)
}

// RecordToggle handles the post submission for the File artifact is online and public toggle.
func RecordToggle(logr *zap.SugaredLogger, c echo.Context, state bool) error {
	const name = "editor record toggle"
	if logr == nil {
		return InternalErr(logr, c, name, ErrZap)
	}

	var f Form
	if err := c.Bind(&f); err != nil {
		return badRequest(c, err)
	}
	if state {
		if err := model.UpdateOnline(c, int64(f.ID)); err != nil {
			return badRequest(c, err)
		}
		return c.JSON(http.StatusOK, f)
	}
	if err := model.UpdateOffline(c, int64(f.ID)); err != nil {
		return badRequest(c, err)
	}
	return c.JSON(http.StatusOK, f)
}

// Records returns the records for the artifacts category URI.
func Records(ctx context.Context, db *sql.DB, uri string, page, limit int) (models.FileSlice, error) {
	if db == nil {
		return nil, ErrDB
	}
	switch Match(uri) {
	// pulldown editor menu matches
	case forApproval:
		r := model.Files{}
		return r.ListForApproval(ctx, db, page, limit)
	case deletions:
		r := model.Files{}
		return r.ListDeletions(ctx, db, page, limit)
	case unwanted:
		r := model.Files{}
		return r.ListUnwanted(ctx, db, page, limit)
	// pulldown menu matches
	case newUploads:
		r := model.Files{}
		return r.List(ctx, db, page, limit)
	case newUpdates:
		r := model.Files{}
		return r.ListUpdates(ctx, db, page, limit)
	case oldest:
		r := model.Files{}
		return r.ListOldest(ctx, db, page, limit)
	case newest:
		r := model.Files{}
		return r.ListNewest(ctx, db, page, limit)
	}
	return recordsZ(ctx, db, uri, page, limit)
}

func recordsZ(ctx context.Context, db *sql.DB, uri string, page, limit int) (models.FileSlice, error) {
	switch Match(uri) {
	// artifacts categories matches
	case advert:
		r := model.Advert{}
		return r.List(ctx, db, page, limit)
	case announcement:
		r := model.Announcement{}
		return r.List(ctx, db, page, limit)
	case ansi:
		r := model.Ansi{}
		return r.List(ctx, db, page, limit)
	case ansiBrand:
		r := model.AnsiBrand{}
		return r.List(ctx, db, page, limit)
	case ansiBBS:
		r := model.AnsiBBS{}
		return r.List(ctx, db, page, limit)
	case ansiFTP:
		r := model.AnsiFTP{}
		return r.List(ctx, db, page, limit)
	case ansiNfo:
		r := model.AnsiNfo{}
		return r.List(ctx, db, page, limit)
	case ansiPack:
		r := model.AnsiPack{}
		return r.List(ctx, db, page, limit)
	case bbs:
		r := model.BBS{}
		return r.List(ctx, db, page, limit)
	case bbsImage:
		r := model.BBSImage{}
		return r.List(ctx, db, page, limit)
	case bbstro:
		r := model.BBStro{}
		return r.List(ctx, db, page, limit)
	case bbsText:
		r := model.BBSText{}
		return r.List(ctx, db, page, limit)
	}
	return records0(ctx, db, uri, page, limit)
}

func records0(ctx context.Context, db *sql.DB, uri string, page, limit int) (models.FileSlice, error) {
	switch Match(uri) {
	case database:
		r := model.Database{}
		return r.List(ctx, db, page, limit)
	case demoscene:
		r := model.Demoscene{}
		return r.List(ctx, db, page, limit)
	case drama:
		r := model.Drama{}
		return r.List(ctx, db, page, limit)
	case ftp:
		r := model.FTP{}
		return r.List(ctx, db, page, limit)
	case hack:
		r := model.Hack{}
		return r.List(ctx, db, page, limit)
	case htm:
		r := model.HTML{}
		return r.List(ctx, db, page, limit)
	case howTo:
		r := model.HowTo{}
		return r.List(ctx, db, page, limit)
	case imageFile:
		r := model.Image{}
		return r.List(ctx, db, page, limit)
	case imagePack:
		r := model.ImagePack{}
		return r.List(ctx, db, page, limit)
	case installer:
		r := model.Installer{}
		return r.List(ctx, db, page, limit)
	case intro:
		r := model.Intro{}
		return r.List(ctx, db, page, limit)
	case linux:
		r := model.Linux{}
		return r.List(ctx, db, page, limit)
	case java:
		r := model.Java{}
		return r.List(ctx, db, page, limit)
	case jobAdvert:
		r := model.JobAdvert{}
		return r.List(ctx, db, page, limit)
	}
	return records1(ctx, db, uri, page, limit)
}

func records1(ctx context.Context, db *sql.DB, uri string, page, limit int) (models.FileSlice, error) {
	switch Match(uri) {
	case macos:
		r := model.Macos{}
		return r.List(ctx, db, page, limit)
	case msdosPack:
		r := model.MsDosPack{}
		return r.List(ctx, db, page, limit)
	case music:
		r := model.Music{}
		return r.List(ctx, db, page, limit)
	case newsArticle:
		r := model.NewsArticle{}
		return r.List(ctx, db, page, limit)
	case nfo:
		r := model.Nfo{}
		return r.List(ctx, db, page, limit)
	case nfoTool:
		r := model.NfoTool{}
		return r.List(ctx, db, page, limit)
	case standards:
		r := model.Standard{}
		return r.List(ctx, db, page, limit)
	case script:
		r := model.Script{}
		return r.List(ctx, db, page, limit)
	case introMsdos:
		r := model.IntroMsDos{}
		return r.List(ctx, db, page, limit)
	case introWindows:
		r := model.IntroWindows{}
		return r.List(ctx, db, page, limit)
	case magazine:
		r := model.Magazine{}
		return r.List(ctx, db, page, limit)
	case msdos:
		r := model.MsDos{}
		return r.List(ctx, db, page, limit)
	case pdf:
		r := model.PDF{}
		return r.List(ctx, db, page, limit)
	case proof:
		r := model.Proof{}
		return r.List(ctx, db, page, limit)
	}
	return records2(ctx, db, uri, page, limit)
}

func records2(ctx context.Context, db *sql.DB, uri string, page, limit int) (models.FileSlice, error) {
	switch Match(uri) {
	case restrict:
		r := model.Restrict{}
		return r.List(ctx, db, page, limit)
	case takedown:
		r := model.Takedown{}
		return r.List(ctx, db, page, limit)
	case text:
		r := model.Text{}
		return r.List(ctx, db, page, limit)
	case textAmiga:
		r := model.TextAmiga{}
		return r.List(ctx, db, page, limit)
	case textApple2:
		r := model.TextApple2{}
		return r.List(ctx, db, page, limit)
	case textAtariST:
		r := model.TextAtariST{}
		return r.List(ctx, db, page, limit)
	case textPack:
		r := model.TextPack{}
		return r.List(ctx, db, page, limit)
	case tool:
		r := model.Tool{}
		return r.List(ctx, db, page, limit)
	case trialCrackme:
		r := model.TrialCrackme{}
		return r.List(ctx, db, page, limit)
	case video:
		r := model.Video{}
		return r.List(ctx, db, page, limit)
	case windows:
		r := model.Windows{}
		return r.List(ctx, db, page, limit)
	case windowsPack:
		r := model.WindowsPack{}
		return r.List(ctx, db, page, limit)
	default:
		return nil, fmt.Errorf("%w: %s", ErrCategory, uri)
	}
}

// Releaser is the handler for the releaser page ordered by the most files.
func Releaser(logr *zap.SugaredLogger, c echo.Context) error {
	return rel(logr, c, model.Prolific)
}

// ReleaserAZ is the handler for the releaser page ordered alphabetically.
func ReleaserAZ(logr *zap.SugaredLogger, c echo.Context) error {
	return rel(logr, c, model.Alphabetical)
}

// ReleaserYear is the handler for the releaser page ordered by year of the first release.
func ReleaserYear(logr *zap.SugaredLogger, c echo.Context) error {
	return rel(logr, c, model.Oldest)
}

// ReleaserErr renders the files error page for the Groups menu and invalid releasers.
func ReleaserErr(logr *zap.SugaredLogger, c echo.Context, id string) error {
	const name = "status"
	if logr == nil {
		return InternalErr(logr, c, name, ErrZap)
	}
	if c == nil {
		return InternalErr(logr, c, name, ErrCxt)
	}
	data := empty(c)
	data["title"] = fmt.Sprintf("%d error, releaser page not found", http.StatusNotFound)
	data["description"] = fmt.Sprintf("HTTP status %d error", http.StatusNotFound)
	data["code"] = http.StatusNotFound
	data["logo"] = "Releaser not found"
	data["alert"] = fmt.Sprintf("Releaser %q cannot be found", releaser.Humanize(id))
	data["probl"] = "The releaser page does not exist, there is probably a typo with the URL."
	data["uriOkay"] = "g/"
	data["uriErr"] = id
	err := c.Render(http.StatusNotFound, name, data)
	if err != nil {
		return InternalErr(logr, c, name, err)
	}
	return nil
}

// ReleaserEdit handles the post submission for the Platform selection field.
func ReleaserEdit(logr *zap.SugaredLogger, c echo.Context) error {
	const name = "editor releasers"
	if logr == nil {
		return InternalErr(logr, c, name, ErrZap)
	}

	var f Form
	if err := c.Bind(&f); err != nil {
		return badRequest(c, err)
	}
	r, err := model.EditFind(f.ID)
	if err != nil {
		return err
	}
	if err = model.UpdateReleasers(c, int64(f.ID), f.Value); err != nil {
		return badRequest(c, err)
	}
	return c.JSON(http.StatusOK, r)
}

// Releasers is the handler for the list and preview of files credited to a releaser.
func Releasers(logr *zap.SugaredLogger, c echo.Context, uri string) error {
	const name = "files"
	if logr == nil {
		return InternalErr(logr, c, name, ErrZap)
	}
	ctx := context.Background()
	db, err := postgres.ConnectDB()
	if err != nil {
		return InternalErr(logr, c, name, err)
	}
	defer db.Close()

	s := releaser.Link(uri)
	rel := model.Releasers{}
	fs, err := rel.List(ctx, db, uri)
	if err != nil {
		return InternalErr(logr, c, name, err)
	}
	if len(fs) == 0 {
		return ReleaserErr(logr, c, uri)
	}
	data := emptyFiles(c)
	data["title"] = "Files for " + s
	data["h1"] = s
	data["lead"] = initialism.Join(initialism.Path(uri))
	data["logo"] = s
	data["description"] = "The collection of files for " + s + "."
	data["demozoo"] = strconv.Itoa(int(zoo.Find(uri)))
	data["sixteen"] = sixteen.Find(uri)
	data["website"] = web.Find(uri)
	data[records] = fs
	switch uri {
	case "independent":
		data["lead"] = initialism.Join(initialism.Path(uri)) +
			", independent releases are files with no group or releaser affiliation." +
			`<br><small class="fw-lighter">In the scene's early years,` +
			` releasing documents or software cracks under a personal alias or a` +
			` real-name attribution was commonplace.</small>`
	default:
		// placeholder to handle other releaser types
	}
	d, err := releaserSum(ctx, db, uri)
	if err != nil {
		return InternalErr(logr, c, name, err)
	}
	data["stats"] = d
	err = c.Render(http.StatusOK, "files", data)
	if err != nil {
		return InternalErr(logr, c, name, err)
	}
	return nil
}

// releaserSum is a helper function for Releasers that returns the statistics for the files page.
func releaserSum(ctx context.Context, db *sql.DB, uri string) (map[string]string, error) {
	if db == nil {
		return nil, ErrDB
	}
	// fetch the statistics of the category
	m := model.Summary{}
	if err := m.Releaser(ctx, db, uri); err != nil {
		return nil, err
	}
	// add the statistics to the data
	d := map[string]string{
		"files": string(ByteFileS("file", m.SumCount.Int64, m.SumBytes.Int64)),
		"years": helper.Years(m.MinYear.Int16, m.MaxYear.Int16),
	}
	return d, nil
}

// Scener is the handler for the page to list all the sceners.
func Scener(logr *zap.SugaredLogger, c echo.Context) error {
	data := empty(c)
	title := "Sceners, the people of The Scene"
	data["title"] = title
	data["logo"] = title
	data["h1"] = title
	data["description"] = demo
	return scener(logr, c, postgres.Roles(), data)
}

// ScenerErr renders the files error page for the People menu and invalid sceners.
func ScenerErr(logr *zap.SugaredLogger, c echo.Context, id string) error {
	const name = "status"
	if logr == nil {
		return InternalErr(logr, c, name, ErrZap)
	}
	if c == nil {
		return InternalErr(logr, c, name, ErrCxt)
	}
	data := empty(c)
	data["title"] = fmt.Sprintf("%d error, scener page not found", http.StatusNotFound)
	data["description"] = fmt.Sprintf("HTTP status %d error", http.StatusNotFound)
	data["code"] = http.StatusNotFound
	data["logo"] = "Scener not found"
	data["alert"] = fmt.Sprintf("Scener %q cannot be found", releaser.Humanize(id))
	data["probl"] = "The scener page does not exist, there is probably a typo with the URL."
	data["uriOkay"] = "p/"
	data["uriErr"] = id
	err := c.Render(http.StatusNotFound, name, data)
	if err != nil {
		return InternalErr(logr, c, name, err)
	}
	return nil
}

// Sceners is the handler for the list and preview of files credited to a scener.
func Sceners(logr *zap.SugaredLogger, c echo.Context, uri string) error {
	const name = "files"
	if logr == nil {
		return InternalErr(logr, c, name, ErrZap)
	}
	ctx := context.Background()
	db, err := postgres.ConnectDB()
	if err != nil {
		return InternalErr(logr, c, name, err)
	}
	defer db.Close()

	s := releaser.Link(uri)
	var rel model.Scener
	fs, err := rel.List(ctx, db, uri)
	if err != nil {
		return InternalErr(logr, c, name, err)
	}
	if len(fs) == 0 {
		return ScenerErr(logr, c, uri)
	}
	data := emptyFiles(c)
	data["title"] = s + attr
	data["h1"] = s
	data["lead"] = "Files attributed to " + s + "."
	data["logo"] = s
	data["description"] = "The collection of files attributed to " + s + "."
	data["scener"] = s
	data[records] = fs
	d, err := scenerSum(ctx, db, uri)
	if err != nil {
		return InternalErr(logr, c, name, err)
	}
	data["stats"] = d
	err = c.Render(http.StatusOK, "files", data)
	if err != nil {
		return InternalErr(logr, c, name, err)
	}
	return nil
}

// SearchDesc is the handler for the Search for file descriptions page.
func SearchDesc(logr *zap.SugaredLogger, c echo.Context) error {
	const title, name = "Search titles and descriptions", "searchPost"
	if logr == nil {
		return InternalErr(logr, c, name, ErrZap)
	}
	data := empty(c)
	data["description"] = "Search form to scan through file descriptions."
	data["logo"] = title
	data["title"] = title
	data["info"] = "search the metadata descriptions of file artifacts"
	err := c.Render(http.StatusOK, name, data)
	if err != nil {
		return InternalErr(logr, c, name, err)
	}
	return nil
}

// SearchFile is the handler for the Search for files page.
func SearchFile(logr *zap.SugaredLogger, c echo.Context) error {
	const title, name = "Search for filenames", "searchPost"
	if logr == nil {
		return InternalErr(logr, c, name, ErrZap)
	}
	data := empty(c)
	data["description"] = "Search form to discover files."
	data["logo"] = title
	data["title"] = title
	data["info"] = "search for filenames or extensions"
	err := c.Render(http.StatusOK, name, data)
	if err != nil {
		return InternalErr(logr, c, name, err)
	}
	return nil
}

// SearchReleaser is the handler for the Releaser Search page.
func SearchReleaser(logr *zap.SugaredLogger, c echo.Context) error {
	const title, name = "Search for releasers", "searchHtmx"
	if logr == nil {
		return InternalErr(logr, c, name, ErrZap)
	}
	data := empty(c)
	data["description"] = "Search form to discover releasers."
	data["logo"] = title
	data["title"] = title
	data["info"] = "search for a group, initialism, magazine, board, or site"
	err := c.Render(http.StatusOK, name, data)
	if err != nil {
		return InternalErr(logr, c, name, err)
	}
	return nil
}

// SignedOut is the handler to sign out and remove the current session.
func SignedOut(logr *zap.SugaredLogger, c echo.Context) error {
	const name = "signedout"
	if logr == nil {
		return InternalErr(logr, c, name, ErrZap)
	}
	{ // get any existing session
		sess, err := session.Get(sess.Name, c)
		if err != nil {
			return BadRequestErr(logr, c, name, err)
		}
		id, ok := sess.Values["sub"]
		if !ok || id == "" {
			return ForbiddenErr(logr, c, name, ErrSession)
		}
		const remove = -1
		sess.Options.MaxAge = remove
		err = sess.Save(c.Request(), c.Response())
		if err != nil {
			return InternalErr(logr, c, name, err)
		}
	}
	return c.Redirect(http.StatusFound, "/")
}

// SignOut is the handler for the Sign out of Defacto2 page.
func SignOut(logr *zap.SugaredLogger, c echo.Context) error {
	const name = "signout"
	if logr == nil {
		return InternalErr(logr, c, name, ErrZap)
	}
	data := empty(c)
	data["title"] = "Sign out"
	data["description"] = "Sign out of Defacto2."
	data["h1"] = "Sign out"
	err := c.Render(http.StatusOK, name, data)
	if err != nil {
		return InternalErr(logr, c, name, err)
	}
	return nil
}

// Signin is the handler for the Sign in session page.
func Signin(logr *zap.SugaredLogger, c echo.Context, clientID, nonce string) error {
	const name = "signin"
	if logr == nil {
		return InternalErr(logr, c, name, ErrZap)
	}
	data := empty(c)
	data["title"] = "Sign in"
	data["description"] = "Sign in to Defacto2."
	data["h1"] = "Sign in"
	data["lead"] = "This sign-in is not open to the general public, and no registration is available."
	data["callback"] = "/google/callback"
	data["clientID"] = clientID
	data["nonce"] = nonce
	{ // get any existing session
		sess, err := session.Get(sess.Name, c)
		if err != nil {
			return remove(logr, c, name, data)
		}
		id, ok := sess.Values["sub"]
		if !ok {
			return remove(logr, c, name, data)
		}
		idStr, ok := id.(string)
		if ok && idStr != "" {
			return SignOut(logr, c)
		}
	}
	err := c.Render(http.StatusOK, name, data)
	if err != nil {
		return InternalErr(logr, c, name, err)
	}
	return nil
}

// Statistics returns the empty database statistics for the artifacts categories.
func Statistics() Stats {
	return Stats{}
}

// StatusErr is the handler for the HTTP status pages such as the 404 - not found.
// If the zap logger is nil then the error page is returned but no error is logged.
// If the echo context is nil then a user hostile, fallback error in raw text is returned.
func StatusErr(logr *zap.SugaredLogger, c echo.Context, code int, uri string) error {
	if logr == nil {
		zapNil(fmt.Errorf("%w: statuserr", ErrZap))
	}
	// render the fallback, text only error page
	if c == nil {
		if logr == nil {
			zapNil(fmt.Errorf("%w: statuserr", ErrCxt))
		} else {
			logr.Errorf("%s: %s", ErrTmpl, ErrCxt)
		}
		return echo.NewHTTPError(http.StatusInternalServerError,
			fmt.Errorf("%w: handler app status", ErrCxt))
	}
	// render a user friendly error page
	data := empty(c)
	data["description"] = fmt.Sprintf("HTTP status %d error", code)
	var title, alert, logo, probl string
	switch code {
	case http.StatusNotFound:
		title = "404 error, page not found"
		logo = "Page not found"
		alert = "The page cannot be found"
		probl = "The page you are looking for might have been removed, had its name changed, or is temporarily unavailable."
	case http.StatusForbidden:
		title = "403 error, forbidden"
		logo = "Forbidden"
		alert = "The page is locked"
		probl = "You don't have permission to access this resource."
	case http.StatusInternalServerError:
		return InternalErr(logr, c, uri, nil)
	default:
		s := http.StatusText(code)
		if s == "" {
			err := fmt.Errorf("%w: %d", ErrCode, code)
			if logr != nil {
				logr.Error(err)
			}
			return echo.NewHTTPError(http.StatusInternalServerError, err)
		}
		title = fmt.Sprintf("%d error, %s", code, s)
		logo = s
		alert = s
		probl = fmt.Sprintf("%d error, %s", code, s)
	}
	data["title"] = title
	data["code"] = code
	data["logo"] = logo
	data["alert"] = alert
	data["probl"] = probl
	data["uriErr"] = uri
	if err := c.Render(code, "status", data); err != nil {
		if logr != nil {
			logr.Errorf("%s: %s", ErrTmpl, err)
		}
		return echo.NewHTTPError(http.StatusInternalServerError, ErrTmpl)
	}
	return nil
}

// TagEdit handles the post submission for the Tag selection field.
func TagEdit(logr *zap.SugaredLogger, c echo.Context) error {
	const name = "editor tag"
	if logr == nil {
		return InternalErr(logr, c, name, ErrZap)
	}

	var f Form
	if err := c.Bind(&f); err != nil {
		return badRequest(c, err)
	}
	r, err := model.EditFind(f.ID)
	if err != nil {
		return err
	}
	if err = model.UpdateTag(c, int64(f.ID), f.Value); err != nil {
		return badRequest(c, err)
	}
	return c.JSON(http.StatusOK, r)
}

// TagInfo handles the POST submission for the platform and tag info.
func TagInfo(logr *zap.SugaredLogger, c echo.Context) error {
	const name = "editor tag info"
	if logr == nil {
		return InternalErr(logr, c, name, ErrZap)
	}
	var f Form
	if err := c.Bind(&f); err != nil {
		return badRequest(c, err)
	}
	info, err := model.GetTagInfo(c, f.Tag)
	if err != nil {
		return badRequest(c, err)
	}
	return c.String(http.StatusOK, info)
}

// Thanks is the handler for the Thanks page.
func Thanks(logr *zap.SugaredLogger, c echo.Context) error {
	const name = "thanks"
	if logr == nil {
		return InternalErr(logr, c, name, ErrZap)
	}
	data := empty(c)
	data["description"] = "Defacto2 thankyous."
	data["h1"] = "Thank you!"
	data["lead"] = "Thanks to the hundreds of people who have contributed to" +
		" Defacto2 over the decades with file submissions, " +
		"hard drive donations, interviews, corrections, artwork, and monetary contributions!"
	data["title"] = "Thanks!"
	err := c.Render(http.StatusOK, name, data)
	if err != nil {
		return InternalErr(logr, c, name, err)
	}
	return nil
}

// TheScene is the handler for the The Scene page.
func TheScene(logr *zap.SugaredLogger, c echo.Context) error {
	const name = "thescene"
	if logr == nil {
		return InternalErr(logr, c, name, ErrZap)
	}
	const h1 = "The Scene?"
	const lead = "Collectively referred to as The Scene," +
		" it is a subculture of different computer activities where participants" +
		" actively share ideas and creations."
	data := empty(c)
	data["description"] = fmt.Sprint(h1, " ", lead)
	data["logo"] = "The underground"
	data["h1"] = h1
	data["lead"] = lead
	data["title"] = h1
	err := c.Render(http.StatusOK, name, data)
	if err != nil {
		return InternalErr(logr, c, name, err)
	}
	return nil
}

// TitleEdit handles the post submission for the Delete readme asset button.
func TitleEdit(logr *zap.SugaredLogger, c echo.Context) error {
	const name = "editor title"
	if logr == nil {
		return InternalErr(logr, c, name, ErrZap)
	}

	var f Form
	if err := c.Bind(&f); err != nil {
		return badRequest(c, err)
	}
	r, err := model.EditFind(f.ID)
	if err != nil {
		return err
	}
	if err = model.UpdateTitle(c, int64(f.ID), f.Value); err != nil {
		return badRequest(c, err)
	}
	return c.JSON(http.StatusOK, r)
}

// VotePouet is the handler for the Pouet production votes JSON page.
func VotePouet(logr *zap.SugaredLogger, c echo.Context, id string) error {
	const title, name, sep = "Pouet", "pouet", ";"
	if logr == nil {
		return InternalErr(logr, c, name, ErrZap)
	}
	pv := pouet.Votes{}
	i, err := strconv.Atoi(id)
	if err != nil {
		return c.String(http.StatusNotFound, err.Error())
	}

	cp := cache.Pouet
	if s, err := cp.Read(id); err == nil {
		if err := PouetCache(c, s); err == nil {
			logr.Debugf("cache hit for pouet id %s", id)
			return nil
		}
	}
	logr.Debugf("cache miss for pouet id %s", id)

	if err = pv.Votes(i); err != nil {
		return c.String(http.StatusNotFound, err.Error())
	}

	if err = c.JSON(http.StatusOK, pv); err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	val := fmt.Sprintf("%.1f%s%d%s%d%s%d",
		pv.Stars, sep, pv.VotesDown, sep, pv.VotesUp, sep, pv.VotesMeh)
	if err := cp.Write(id, val, cache.ExpiredAt); err != nil {
		logr.Errorf("failed to write pouet id %s to cache db: %s", id, err)
	}
	return nil
}

// Website is the handler for the websites page.
// Open is the ID of the accordion section to open.
func Website(logr *zap.SugaredLogger, c echo.Context, open string) error {
	const name = "websites"
	if logr == nil {
		return InternalErr(logr, c, name, ErrZap)
	}
	data := empty(c)
	data["title"] = "Websites"
	data["logo"] = "Websites, podcasts, videos, books and films"
	data["description"] = "A collection of websites, podcasts, videos, books and films about the scene."
	acc := List()
	// Open the accordion section.
	closeAll := true
	for i, site := range acc {
		if site.ID == open || open == "" {
			site.Open = true
			data["title"] = site.Title
			closeAll = false
			acc[i] = site
			if open == "" {
				continue
			}
			break
		}
	}
	// If a section was requested but not found, return a 404.
	if open != "hide" && closeAll {
		return StatusErr(logr, c, http.StatusNotFound, open)
	}
	// Render the page.
	data["accordion"] = acc
	err := c.Render(http.StatusOK, name, data)
	if err != nil {
		return InternalErr(logr, c, name, err)
	}
	return nil
}

// Writer is the handler for the Writer page.
func Writer(logr *zap.SugaredLogger, c echo.Context) error {
	data := empty(c)
	title := "Writers, editors and authors"
	data["title"] = title
	data["logo"] = title
	data["h1"] = title
	data["description"] = demo
	return scener(logr, c, postgres.Writer, data)
}

// FileSearch is the type of search to perform.
type FileSearch int

const (
	Filenames    FileSearch = iota // Filenames is the search for filenames.
	Descriptions                   // Descriptions is the search for file descriptions and titles.
)

// Stats are the database statistics for the artifacts categories.
type Stats struct {
	IntroW    model.IntroWindows
	Record    model.Files
	Ansi      model.Ansi
	AnsiBBS   model.AnsiBBS
	BBS       model.BBS
	BBSText   model.BBSText
	BBStro    model.BBStro
	Demoscene model.Demoscene
	MsDos     model.MsDos
	Intro     model.Intro
	IntroD    model.IntroMsDos
	Installer model.Installer
	Java      model.Java
	Linux     model.Linux
	Magazine  model.Magazine
	Macos     model.Macos
	Nfo       model.Nfo
	NfoTool   model.NfoTool
	Proof     model.Proof
	Script    model.Script
	Text      model.Text
	Windows   model.Windows
}

// Get and store the database statistics for the artifacts categories.
func (s *Stats) Get(ctx context.Context, db *sql.DB) error {
	if db == nil {
		return ErrDB
	}
	if err := s.Record.Stat(ctx, db); err != nil {
		return err
	}
	if err := s.Ansi.Stat(ctx, db); err != nil {
		return err
	}
	if err := s.AnsiBBS.Stat(ctx, db); err != nil {
		return err
	}
	if err := s.BBS.Stat(ctx, db); err != nil {
		return err
	}
	if err := s.BBSText.Stat(ctx, db); err != nil {
		return err
	}
	if err := s.BBStro.Stat(ctx, db); err != nil {
		return err
	}
	if err := s.MsDos.Stat(ctx, db); err != nil {
		return err
	}
	if err := s.Intro.Stat(ctx, db); err != nil {
		return err
	}
	if err := s.IntroD.Stat(ctx, db); err != nil {
		return err
	}
	if err := s.IntroW.Stat(ctx, db); err != nil {
		return err
	}
	if err := s.Installer.Stat(ctx, db); err != nil {
		return err
	}
	if err := s.Java.Stat(ctx, db); err != nil {
		return err
	}
	if err := s.Linux.Stat(ctx, db); err != nil {
		return err
	}
	if err := s.Demoscene.Stat(ctx, db); err != nil {
		return err
	}
	return s.get(ctx, db)
}

func (s *Stats) get(ctx context.Context, db *sql.DB) error {
	if err := s.Macos.Stat(ctx, db); err != nil {
		return err
	}
	if err := s.Magazine.Stat(ctx, db); err != nil {
		return err
	}
	if err := s.Nfo.Stat(ctx, db); err != nil {
		return err
	}
	if err := s.NfoTool.Stat(ctx, db); err != nil {
		return err
	}
	if err := s.Proof.Stat(ctx, db); err != nil {
		return err
	}
	if err := s.Script.Stat(ctx, db); err != nil {
		return err
	}
	if err := s.Text.Stat(ctx, db); err != nil {
		return err
	}
	return s.Windows.Stat(ctx, db)
}

// bbsHandler is the handler for the BBS page.
func bbsHandler(logr *zap.SugaredLogger, c echo.Context, orderBy model.OrderBy) error {
	const title, name = "BBS", "bbs"
	if logr == nil {
		return InternalErr(logr, c, name, ErrZap)
	}
	const lead = "Bulletin Board Systems are historical, " +
		"networked personal computer servers connected using the landline telephone network and provide forums, " +
		"real-time chat, mail, and file sharing for The Scene \"elites.\""
	const key = "releasers"
	data := empty(c)
	data["title"] = title
	data["description"] = lead
	data["logo"] = "Bulletin Board Systems"
	data["h1"] = title
	data["lead"] = lead
	// releaser.html specific data items
	data["itemName"] = name
	data[key] = model.Releasers{}
	data["stats"] = map[string]string{}

	ctx := context.Background()
	db, err := postgres.ConnectDB()
	if err != nil {
		return InternalErr(logr, c, name, err)
	}
	defer db.Close()
	r := model.Releasers{}
	if err := r.BBS(ctx, db, orderBy); err != nil {
		return DatabaseErr(logr, c, name, err)
	}
	m := model.Summary{}
	if err := m.BBS(ctx, db); err != nil {
		return DatabaseErr(logr, c, name, err)
	}
	data[key] = r
	data["stats"] = map[string]string{
		"pubs":   fmt.Sprintf("%d boards", len(r)),
		"issues": string(ByteFileS("file artifact", m.SumCount.Int64, m.SumBytes.Int64)),
		"years":  fmt.Sprintf("%d - %d", m.MinYear.Int16, m.MaxYear.Int16),
	}
	tmpl := name
	if orderBy == model.Oldest {
		tmpl = "bbs-year"
		s := lead + " The list is sorted by the year of the first release and " +
			"for performance, boards from 1993 require at least 2 file artifacts."
		data["lead"] = s
	}
	err = c.Render(http.StatusOK, tmpl, data)
	if err != nil {
		return InternalErr(logr, c, name, err)
	}
	return nil
}

// counter returns the statistics for the artifacts categories.
func counter() (Stats, error) {
	ctx := context.Background()
	db, err := postgres.ConnectDB()
	if err != nil {
		return Stats{}, err
	}
	defer db.Close()
	counter := Stats{}
	if err := counter.Get(ctx, db); err != nil {
		return Stats{}, err
	}
	return counter, nil
}

// empty is a map of default values for the app templates.
func empty(c echo.Context) map[string]interface{} {
	// the keys are listed in order of appearance in the templates.
	// * marked keys are required.
	// ! marked keys are suggested.
	return map[string]interface{}{
		"cacheFiles":  Caching.RecordCount, //   The number of records of files in the database.
		"canonical":   "",                  //   A canonical URL is the URL of the best representative page from a group of duplicate pages.
		"carousel":    "",                  //   The ID of the carousel to display.
		"counter":     Statistics(),        //   Empty database counts for files and categories.
		"dbError":     false,               //   If true, the database is not available.
		"description": "",                  // * A short description of the page that get inserted into the description meta element.
		"editor":      sess.Editor(c),      //   If true, the editor mode is enabled.
		"h1":          "",                  // ! The H1 heading of the page.
		"h1Sub":       "",                  //   The H1 sub-heading of the page.
		"jsdos6":      false,               //   If true, the large, js-dos v6.22 emulator files will be loaded.
		"lead":        "",                  // ! The enlarged, lead paragraph of the page.
		"logo":        "",                  // ! Text to insert into the monospaced, ASCII art logo.
		"readOnly":    true,                //   If true, the application is in read-only mode.
		"title":       "",                  // * The title of the page that get inserted into the title meta element.
	}
}

// emptyFiles is a map of default values specific to the files templates.
func emptyFiles(c echo.Context) map[string]interface{} {
	data := empty(c)
	data["demozoo"] = "0"
	data["sixteen"] = ""
	data["scener"] = ""
	data["website"] = ""
	data["unknownYears"] = true
	return data
}

// fileWStats is a helper function for File that adds the statistics to the data map.
func fileWStats(data map[string]interface{}, stats bool) (map[string]interface{}, error) {
	if !stats {
		return data, nil
	}
	c, err := counter()
	if err != nil {
		return data, err
	}
	data["counter"] = c
	data["logo"] = "Artifact category statistics"
	data["lead"] = "This page shows the artifacts categories with selected statistics, " +
		"such as the number of files in the category or platform." +
		fmt.Sprintf(" The total number of files in the database is %d.", c.Record.Count) +
		fmt.Sprintf(" The total size of all file artifacts are %s.", helper.ByteCount(int64(c.Record.Bytes)))
	return data, nil
}

// files is a helper function for Files that returns the data map for the files page.
func files(logr *zap.SugaredLogger, c echo.Context, uri string, page int) error {
	const title, name = "Files", "files"
	logo, h1sub, lead := fileInfo(uri)
	data := emptyFiles(c)
	data["title"] = title
	data["description"] = "Table of contents for the files."
	data["logo"] = logo
	data["h1"] = title
	data["h1Sub"] = h1sub
	data["lead"] = lead
	data[records] = []models.FileSlice{}
	data["unknownYears"] = true
	switch uri {
	case
		newUploads.String(),
		newUpdates.String(),
		forApproval.String(),
		deletions.String(),
		unwanted.String():
		data["unknownYears"] = false
	}

	ctx := context.Background()
	db, err := postgres.ConnectDB()
	if err != nil {
		return InternalErr(logr, c, name, err)
	}
	defer db.Close()
	// fetch the records by category
	r, err := Records(ctx, db, uri, page, limit)
	if err != nil {
		return DatabaseErr(logr, c, name, err)
	}
	data[records] = r
	d, sum, err := stats(ctx, db, uri)
	if err != nil {
		return DatabaseErr(logr, c, name, err)
	}
	data["stats"] = d
	lastPage := math.Ceil(float64(sum) / float64(limit))
	if page > int(lastPage) {
		i := strconv.Itoa(page)
		return PageErr(logr, c, uri, i)
	}
	const pages = 2
	data["Pagination"] = model.Pagination{
		TwoAfter:  page + pages,
		NextPage:  page + 1,
		CurrPage:  page,
		PrevPage:  page - 1,
		TwoBelow:  page - pages,
		SumPages:  int(lastPage),
		BaseURL:   "/files/" + uri,
		RangeStep: steps(lastPage),
	}
	err = c.Render(http.StatusOK, name, data)
	if err != nil {
		return InternalErr(logr, c, name, err)
	}
	return nil
}

func steps(lastPage float64) int {
	const one, two, four = 1, 2, 4
	const skip2Pages, skip4Pages = 39, 99
	switch {
	case lastPage > skip4Pages:
		return four
	case lastPage > skip2Pages:
		return two
	default:
		return one
	}
}

// mag is the handler for the magazine page.
func mag(logr *zap.SugaredLogger, c echo.Context, chronological bool) error {
	const title, name = "Magazine", "magazine"
	if logr == nil {
		return InternalErr(logr, c, name, ErrZap)
	}
	data := empty(c)
	const lead = "The magazines are newsletters, reports, and publications about activities within The Scene subculture."
	const issue = "issue"
	const key = "releasers"
	data["title"] = title
	data["description"] = lead
	data["logo"] = title
	data["h1"] = title
	data["lead"] = lead
	// releaser.html specific data items
	data["itemName"] = issue
	data[key] = model.Releasers{}
	data["stats"] = map[string]string{}

	ctx := context.Background()
	db, err := postgres.ConnectDB()
	if err != nil {
		return InternalErr(logr, c, name, err)
	}
	defer db.Close()
	r := model.Releasers{}
	if chronological {
		if err := r.Magazine(ctx, db); err != nil {
			return DatabaseErr(logr, c, name, err)
		}
	} else {
		if err := r.MagazineAZ(ctx, db); err != nil {
			return DatabaseErr(logr, c, name, err)
		}
	}
	m := model.Summary{}
	if err := m.Magazine(ctx, db); err != nil {
		return DatabaseErr(logr, c, name, err)
	}
	data[key] = r
	data["stats"] = map[string]string{
		"pubs":   fmt.Sprintf("%d publications", len(r)),
		"issues": string(ByteFileS(issue, m.SumCount.Int64, m.SumBytes.Int64)),
		"years":  fmt.Sprintf("%d - %d", m.MinYear.Int16, m.MaxYear.Int16),
	}
	err = c.Render(http.StatusOK, name, data)
	if err != nil {
		return InternalErr(logr, c, name, err)
	}
	return nil
}

// postStats is a helper function for PostName that returns the statistics for the files page.
func (mode FileSearch) postStats(ctx context.Context, db *sql.DB, terms []string) map[string]string {
	if db == nil {
		return nil
	}
	none := func() map[string]string {
		return map[string]string{
			"files": "no files found",
			"years": "",
		}
	}
	// fetch the statistics of the category
	m := model.Summary{}
	switch mode {
	case Filenames:
		if err := m.SearchFilename(ctx, db, terms); err != nil {
			return none()
		}
	case Descriptions:
		if err := m.SearchDesc(ctx, db, terms); err != nil {
			return none()
		}
	}
	if m.SumCount.Int64 == 0 {
		return none()
	}
	// add the statistics to the data
	d := map[string]string{
		"files": string(ByteFileS("file", m.SumCount.Int64, m.SumBytes.Int64)),
		"years": helper.Years(m.MinYear.Int16, m.MaxYear.Int16),
	}
	return d
}

// remove is a helper function to remove the session cookie by setting the MaxAge to -1.
func remove(logr *zap.SugaredLogger, c echo.Context, name string, data map[string]interface{}) error {
	sess, err := session.Get(sess.Name, c)
	if err != nil {
		const remove = -1
		sess.Options.MaxAge = remove
		_ = sess.Save(c.Request(), c.Response())
	}
	err = c.Render(http.StatusOK, name, data)
	if err != nil {
		return InternalErr(logr, c, name, err)
	}
	return nil
}

// scenerSum is a helper function for Sceners that returns the statistics for the files page.
func scenerSum(ctx context.Context, db *sql.DB, uri string) (map[string]string, error) {
	if db == nil {
		return nil, ErrDB
	}
	// fetch the statistics of the category
	m := model.Summary{}
	if err := m.Scener(ctx, db, uri); err != nil {
		return nil, err
	}
	// add the statistics to the data
	d := map[string]string{
		"files": string(ByteFileS("file", m.SumCount.Int64, m.SumBytes.Int64)),
		"years": helper.Years(m.MinYear.Int16, m.MaxYear.Int16),
	}
	return d, nil
}

// scener is the handler for the scener pages.
func scener(logr *zap.SugaredLogger, c echo.Context, r postgres.Role,
	data map[string]interface{},
) error {
	const name = "scener"
	if logr == nil {
		return InternalErr(logr, c, name, ErrZap)
	}
	s := model.Sceners{}
	ctx := context.Background()
	db, err := postgres.ConnectDB()
	if err != nil {
		return InternalErr(logr, c, name, err)
	}
	switch r {
	case postgres.Writer:
		err = s.Writer(ctx, db)
	case postgres.Artist:
		err = s.Artist(ctx, db)
	case postgres.Musician:
		err = s.Musician(ctx, db)
	case postgres.Coder:
		err = s.Coder(ctx, db)
	case postgres.Roles():
		err = s.All(ctx, db)
	}
	if err != nil {
		return DatabaseErr(logr, c, name, err)
	}
	data["sceners"] = s.Sort()
	data["description"] = "Sceners and people who have been credited for their work in The Scene."
	data["lead"] = "This page shows the sceners and people credited for their work in The Scene." +
		`<br><small class="fw-lighter">` +
		"The list will never be complete or accurate due to the amount of data and the lack of a" +
		" common format for crediting people. " +
		" Sceners often used different names or spellings on their work, including character" +
		" swaps, aliases, initials, and even single-letter signatures." +
		"</small>"
	err = c.Render(http.StatusOK, name, data)
	if err != nil {
		return InternalErr(logr, c, name, err)
	}
	return nil
}

// sessionHandler creates a [new session] and populates it with
// the claims data created by the [ID Tokens for Google HTTP APIs].
//
// [new session]: https://pkg.go.dev/github.com/gorilla/sessions
// [ID Tokens for Google HTTP APIs]: https://pkg.go.dev/google.golang.org/api/idtoken
func sessionHandler(
	logr *zap.SugaredLogger,
	c echo.Context, maxAge int,
	claims map[string]interface{},
) error {
	const name = "google/callback"
	if logr == nil {
		return InternalErr(logr, c, name, ErrZap)
	}

	// get always returns a session, even if empty.
	session, err := session.Get(sess.Name, c)
	if err != nil {
		return err
	}

	// session Options are cookie options and are all optional
	// https://developer.mozilla.org/en-US/docs/Web/HTTP/Cookies
	const hour = 60 * 60
	session.Options = &sessions.Options{
		Path:     "/",                  // path that must exist in the requested URL to send the Cookie header
		Domain:   "",                   // which server can receive a cookie
		MaxAge:   hour * maxAge,        // maximum age for the cookie, in seconds
		Secure:   true,                 // cookie requires HTTPS except for localhost
		HttpOnly: true,                 // stops the cookie being read by JS
		SameSite: http.SameSiteLaxMode, // LaxMode (default) or StrictMode
	}

	// sub is the unique user id from google
	val, ok := claims["sub"]
	if !ok {
		return ErrClaims
	}
	session.Values["sub"] = val

	// optionals
	session.Values["givenName"] = claims["given_name"]
	session.Values["email"] = claims["email"]
	session.Values["emailVerified"] = claims["email_verified"]

	// save the session
	return session.Save(c.Request(), c.Response())
}

// stats is a helper function for Files that returns the statistics for the files page.
func stats(ctx context.Context, db *sql.DB, uri string) (map[string]string, int, error) {
	if db == nil {
		return nil, 0, ErrDB
	}
	if !Valid(uri) {
		return nil, 0, nil
	}
	// fetch the statistics of the uri
	m := model.Summary{}
	err := m.URI(ctx, db, uri)
	if err != nil && !errors.Is(err, model.ErrURI) {
		return nil, 0, err
	}
	if errors.Is(err, model.ErrURI) {
		switch uri {
		case "for-approval":
			if err := m.StatForApproval(ctx, db); err != nil {
				return nil, 0, err
			}
		case "deletions":
			if err := m.StatDeletions(ctx, db); err != nil {
				return nil, 0, err
			}
		case "unwanted":
			if err := m.StatUnwanted(ctx, db); err != nil {
				return nil, 0, err
			}
		default:
			if err := m.All(ctx, db); err != nil {
				return nil, 0, err
			}
		}
	}
	// add the statistics to the data
	d := map[string]string{
		"files": string(ByteFileS("file", m.SumCount.Int64, m.SumBytes.Int64)),
		"years": fmt.Sprintf("%d - %d", m.MinYear.Int16, m.MaxYear.Int16),
	}
	switch uri {
	case "new-updates", "new-uploads", "newest", "for-approval":
		d["years"] = fmt.Sprintf("%d - %d", m.MaxYear.Int16, m.MinYear.Int16)
	}
	return d, int(m.SumCount.Int64), nil
}

// rel is the handler for the Releaser page.
func rel(logr *zap.SugaredLogger, c echo.Context, orderBy model.OrderBy) error {
	const title, name = "Releaser", "releaser"
	if logr == nil {
		return InternalErr(logr, c, name, ErrZap)
	}
	data := empty(c)
	const lead = "A releaser is a brand or a collective group of sceners responsible for releasing or distributing products."
	const key = "releasers"
	data["title"] = title
	data["description"] = fmt.Sprint(title, " ", lead)
	data["logo"] = "Groups, organizations and publications"
	data["h1"] = title
	data["lead"] = lead
	// releaser.html specific data items
	data["itemName"] = "file"
	data[key] = model.Releasers{}
	data["stats"] = map[string]string{}

	ctx := context.Background()
	db, err := postgres.ConnectDB()
	if err != nil {
		return InternalErr(logr, c, name, err)
	}
	defer db.Close()
	var r model.Releasers
	if err := r.All(ctx, db, orderBy, 0, 0); err != nil {
		return DatabaseErr(logr, c, name, err)
	}
	var m model.Summary
	if err := m.All(ctx, db); err != nil {
		return DatabaseErr(logr, c, name, err)
	}
	data[key] = r
	data["stats"] = map[string]string{
		"pubs":   fmt.Sprintf("%d releasers and groups", len(r)),
		"issues": string(ByteFileS("file artifact", m.SumCount.Int64, m.SumBytes.Int64)),
		"years":  helper.Years(m.MinYear.Int16, m.MaxYear.Int16),
	}
	tmpl := name
	if orderBy == model.Oldest {
		tmpl = "releaser-year"
		s := lead + " The list is sorted by the year of the first release and " +
			"only lists releasers from before 2000. For performance, releasers from 1991-99 require at least 3 file artifacts."
		data["lead"] = s
	}
	err = c.Render(http.StatusOK, tmpl, data)
	if err != nil {
		return InternalErr(logr, c, tmpl, err)
	}
	return nil
}

// zapNil is a helper function to log an error when the zap logger is nil.
func zapNil(err error) {
	log.Println(fmt.Errorf("cannot log the following error: %w", err))
}
