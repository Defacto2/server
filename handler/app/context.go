package app

// Package file context.go contains the router handlers for the Defacto2 website.

import (
	"context"
	"crypto/sha512"
	"errors"
	"fmt"
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
	"github.com/Defacto2/server/internal/config"
	"github.com/Defacto2/server/internal/demozoo"
	"github.com/Defacto2/server/internal/helper"
	"github.com/Defacto2/server/internal/postgres"
	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/Defacto2/server/internal/pouet"
	"github.com/Defacto2/server/internal/site"
	"github.com/Defacto2/server/internal/sixteen"
	"github.com/Defacto2/server/internal/tags"
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
	az      = ", a-z"
	byyear  = ", by year"
	alpha   = "alphabetically"
	year    = "by year"
)

// Artifacts is the handler for the list and preview of the files page.
// The uri is the category or collection of files to display.
// The page is the page number of the results to display.
func Artifacts(c echo.Context, uri, page string) error {
	if !Valid(uri) {
		return Artifacts404(c, uri)
	}
	if page == "" {
		return artifacts(c, uri, 1)
	}
	p, err := strconv.Atoi(page)
	if err != nil {
		return Page404(c, uri, page)
	}
	return artifacts(c, uri, p)
}

// Artifacts404 renders the files error page for the Artifacts menu and categories.
// It provides different error messages to the standard error page.
func Artifacts404(c echo.Context, uri string) error {
	const name = "status"
	errs := fmt.Sprint("artifact page not found for,", uri)
	if c == nil {
		return InternalErr(c, errs, ErrCxt)
	}
	data := empty(c)
	data["title"] = fmt.Sprintf("%d error, files page not found", http.StatusNotFound)
	data["description"] = fmt.Sprintf("HTTP status %d error", http.StatusNotFound)
	data["code"] = http.StatusNotFound
	data["logo"] = "Artifacts not found"
	data["alert"] = "Artifacts page cannot be found"
	data["probl"] = "The files category or menu option does not exist, there is probably a typo with the URL."
	data["uriOkay"] = "files/"
	data["uriErr"] = uri
	err := c.Render(http.StatusNotFound, name, data)
	if err != nil {
		return InternalErr(c, errs, err)
	}
	return nil
}

// Artist is the handler for the Artist sceners page.
func Artist(c echo.Context) error {
	data := empty(c)
	title := "Pixel artists and graphic designers"
	data["title"] = title
	data["logo"] = title
	data["h1"] = title
	data["description"] = demo
	return scener(c, postgres.Artist, data)
}

// BBS is the handler for the BBS page ordered by the most files.
func BBS(c echo.Context) error {
	return bbsHandler(c, model.Prolific)
}

// BBSAZ is the handler for the BBS page ordered alphabetically.
func BBSAZ(c echo.Context) error {
	return bbsHandler(c, model.Alphabetical)
}

// BBSYear is the handler for the BBS page ordered by the year.
func BBSYear(c echo.Context) error {
	return bbsHandler(c, model.Oldest)
}

// Checksum is the handler for the Checksum file record page.
func Checksum(c echo.Context, id string) error {
	const uri = "sum"
	if err := download.Checksum(c, id); err != nil {
		if errors.Is(err, download.ErrStat) {
			return FileMissingErr(c, uri, err)
		}
		return DownloadErr(c, uri, err)
	}
	return nil
}

// Code is the handler for the Coder sceners page.
func Coder(c echo.Context) error {
	data := empty(c)
	title := "Coder and programmers"
	data["title"] = title
	data["logo"] = title
	data["h1"] = title
	data["description"] = demo
	return scener(c, postgres.Writer, data)
}

// Configurations is the handler for the Configuration page.
func Configurations(cx echo.Context, conf config.Config) error {
	const name = "configs"
	data := empty(cx)
	data["description"] = "Defacto2 configurations."
	data["h1"] = "Configurations"
	data["lead"] = "The web application configurations, tools and links to special records."
	data["title"] = "Configs"
	data["configurations"] = conf

	data["countArtifacts"] = 0
	data["countPublic"] = 0
	data["countNewUpload"] = 0
	data["countHidden"] = 0
	ctx := context.Background()
	db, err := postgres.ConnectDB()
	if err == nil {
		ca, cp, cnu, err := model.Counts(ctx, db)
		if err == nil {
			data["countArtifacts"] = ca
			data["countPublic"] = cp
			data["countNewUpload"] = cnu
			data["countHidden"] = ca - cp - cnu
		}
	}
	defer db.Close()
	data = configurations(data, conf)
	conns, max, err := postgres.Connections()
	if err != nil {
		data["dbConnections"] = err.Error()
	}
	data["dbConnections"] = fmt.Sprintf("%d of %d", conns, max)
	err = cx.Render(http.StatusOK, name, data)
	if err != nil {
		return InternalErr(cx, name, err)
	}
	return nil
}

func configurations(data map[string]interface{}, conf config.Config) map[string]interface{} {
	check := config.CheckDir(conf.AbsDownload, "downloads")
	data["checkDownloads"] = check
	data["countDownloads"] = 0
	data["extsDownloads"] = []helper.Extension{}
	if check == nil {
		data["countDownloads"], _ = helper.Count(conf.AbsDownload)
		exts, _ := helper.CountExts(conf.AbsDownload)
		data["extsDownloads"] = exts
	}
	check = config.CheckDir(conf.AbsPreview, "previews")
	data["checkPreviews"] = check
	data["countPreviews"] = 0
	data["extsPreviews"] = []helper.Extension{}
	if check == nil {
		data["countPreviews"], _ = helper.Count(conf.AbsPreview)
		exts, _ := helper.CountExts(conf.AbsPreview)
		data["extsPreviews"] = exts
	}
	check = config.CheckDir(conf.AbsThumbnail, "thumbnails")
	data["checkThumbnails"] = check
	data["countThumbnails"] = 0
	data["extsThumbnails"] = []helper.Extension{}
	if check == nil {
		data["countThumbnails"], _ = helper.Count(conf.AbsThumbnail)
		exts, _ := helper.CountExts(conf.AbsThumbnail)
		data["extsThumbnails"] = exts
	}
	check = config.CheckDir(conf.AbsExtra, "extra")
	data["checkExtras"] = check
	data["countExtras"] = 0
	data["extsExtras"] = []helper.Extension{}
	if check == nil {
		data["countExtras"], _ = helper.Count(conf.AbsExtra)
		exts, _ := helper.CountExts(conf.AbsExtra)
		data["extsExtras"] = exts
	}
	check = config.CheckDir(conf.AbsOrphaned, "orphaned")
	data["checkOrphaned"] = check
	data["countOrphaned"] = 0
	data["extsOrphaned"] = []helper.Extension{}
	if check == nil {
		data["countOrphaned"], _ = helper.Count(conf.AbsOrphaned)
		exts, _ := helper.CountExts(conf.AbsOrphaned)
		data["extsOrphaned"] = exts
	}
	return data
}

// Download is the handler for the Download file record page.
func Download(c echo.Context, logger *zap.SugaredLogger, path string) error {
	d := download.Download{
		Inline: false,
		Path:   path,
	}
	const uri = "d"
	if err := d.HTTPSend(c, logger); err != nil {
		if errors.Is(err, download.ErrStat) {
			return FileMissingErr(c, uri, err)
		}
		return DownloadErr(c, uri, err)
	}
	return nil
}

// FTP is the handler for the FTP page.
func FTP(c echo.Context) error {
	const title, name = "FTP", "ftp"
	ctx := context.Background()
	db, err := postgres.ConnectDB()
	if err != nil {
		return InternalErr(c, name, err)
	}
	defer db.Close()
	data := empty(c)
	const lead = "FTP sites are historical, internet-based file servers for uploading " +
		"and downloading \"elite\" scene releases."
	const key = "releasers"
	data["title"] = title
	data["description"] = lead
	data["logo"] = "FTP sites, A-Z"
	data["h1"] = title
	data["lead"] = lead
	// releaser.html specific data items
	data["itemName"] = name
	data[key] = model.Releasers{}
	data["stats"] = map[string]string{}
	r := model.Releasers{}
	if err := r.FTP(ctx, db); err != nil {
		return DatabaseErr(c, name, err)
	}
	data[key] = r
	data["stats"] = map[string]string{
		"pubs":    fmt.Sprintf("%d sites", len(r)),
		"orderBy": alpha,
	}
	err = c.Render(http.StatusOK, name, data)
	if err != nil {
		return InternalErr(c, name, err)
	}
	return nil
}

// Categories is the handler for the artifact categories page.
func Categories(c echo.Context, logger *zap.SugaredLogger, stats bool) error {
	const title, name = "Artifact categories", "categories"
	if logger == nil {
		return InternalErr(c, "name", ErrZap)
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
		logger.Warn(err)
		data["dbError"] = true
	}
	err = c.Render(http.StatusOK, name, data)
	if err != nil {
		return InternalErr(c, name, err)
	}
	return nil
}

func Deletions(c echo.Context, page string) error {
	uri := deletions.String()
	if !Valid(uri) {
		return Artifacts404(c, uri)
	}
	if page == "" {
		return artifacts(c, uri, 1)
	}
	p, err := strconv.Atoi(page)
	if err != nil {
		return Page404(c, uri, page)
	}
	return artifacts(c, uri, p)
}

func Unwanted(c echo.Context, page string) error {
	uri := unwanted.String()
	if !Valid(uri) {
		return Artifacts404(c, uri)
	}
	if page == "" {
		return artifacts(c, uri, 1)
	}
	p, err := strconv.Atoi(page)
	if err != nil {
		return Page404(c, uri, page)
	}
	return artifacts(c, uri, p)
}

// ForApproval is the handler for the list and preview of the files page.
// The uri is the category or collection of files to display.
// The page is the page number of the results to display.
func ForApproval(c echo.Context, page string) error {
	uri := forApproval.String()
	if !Valid(uri) {
		return Artifacts404(c, uri)
	}
	if page == "" {
		return artifacts(c, uri, 1)
	}
	p, err := strconv.Atoi(page)
	if err != nil {
		return Page404(c, uri, page)
	}
	return artifacts(c, uri, p)
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
func GetDemozooLink(c echo.Context, downloadDir string) error {
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

// GoogleCallback is the handler for the Google OAuth2 callback page to verify
// the [Google ID token].
//
// [Google ID token]: https://developers.google.com/identity/gsi/web/guides/verify-google-id-token
func GoogleCallback(c echo.Context, clientID string, maxAge int, accounts ...[48]byte) error {
	const name = "google/callback"

	// Cross-Site Request Forgery cookie token
	const csrf = "g_csrf_token"
	cookie, err := c.Cookie(csrf)
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			return c.Redirect(http.StatusForbidden, "/signin")
		}
		return BadRequestErr(c, name, err)
	}
	token := cookie.Value

	// Cross-Site Request Forgery post token
	bodyToken := c.FormValue(csrf)
	if token != bodyToken {
		return BadRequestErr(c, name, ErrMisMatch)
	}

	// Create a new token verifier.
	// https://pkg.go.dev/google.golang.org/api/idtoken
	ctx := context.Background()
	validator, err := idtoken.NewValidator(ctx)
	if err != nil {
		return BadRequestErr(c, name, err)
	}

	// Verify the ID token and using the client ID from the Google API.
	credential := c.FormValue("credential")
	playload, err := validator.Validate(ctx, credential, clientID)
	if err != nil {
		return BadRequestErr(c, name, err)
	}

	// Verify the sub value against the list of allowed accounts.
	check := false
	if sub, subExists := playload.Claims["sub"]; subExists {
		for _, account := range accounts {
			if id, subString := sub.(string); subString && sha512.Sum384([]byte(id)) == account {
				check = true
				break
			}
		}
	}
	if !check {
		fullname := playload.Claims["name"]
		sub := playload.Claims["sub"]
		return ForbiddenErr(c, name,
			fmt.Errorf("%w %s. "+
				"If this is a mistake, contact Defacto2 admin and give them this Google account ID: %s",
				ErrUser, fullname, sub))
	}

	if err = sessionHandler(c, maxAge, playload.Claims); err != nil {
		return BadRequestErr(c, name, err)
	}
	return c.Redirect(http.StatusFound, "/")
}

// History is the handler for the History page.
func History(c echo.Context) error {
	const name = "history"
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
		return InternalErr(c, name, err)
	}
	return nil
}

// Index is the handler for the Home page.
func Index(c echo.Context) error {
	const name = "index"
	const lead = "the website preserving the historic PC cracking scene subculture. " +
		"It covers digital artifacts including text files, demos, music, art, " +
		"magazines, and other projects."
	const desc = "Defacto2 is " + lead
	data := empty(c)
	data["title"] = "Milestones"
	data["description"] = desc
	data["h1"] = "Welcome,"
	data["milestones"] = Collection()
	{
		// get the signed in given name
		sess, err := session.Get(sess.Name, c)
		if err == nil {
			if givenName, givenExists := sess.Values["givenName"]; givenExists {
				if name, nameStr := givenName.(string); nameStr && name != "" {
					data["h1"] = "Welcome, " + name
				}
			}
		}
	}
	err := c.Render(http.StatusOK, name, data)
	if err != nil {
		return InternalErr(c, name, err)
	}
	return nil
}

// Inline is the handler for the Download file record page.
func Inline(c echo.Context, logger *zap.SugaredLogger, path string) error {
	d := download.Download{
		Inline: true,
		Path:   path,
	}
	const uri = "v"
	if err := d.HTTPSend(c, logger); err != nil {
		if errors.Is(err, download.ErrStat) {
			return FileMissingErr(c, uri, err)
		}
		return DownloadErr(c, uri, err)
	}
	return nil
}

// Interview is the handler for the People Interviews page.
func Interview(c echo.Context) error {
	const title, name = "Interviews with sceners", "interview"
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
		return InternalErr(c, name, err)
	}
	return nil
}

// Magazine is the handler for the Magazine page.
func Magazine(c echo.Context) error {
	return magazines(c, true)
}

// MagazineAZ is the handler for the Magazine page ordered chronologically.
func MagazineAZ(c echo.Context) error {
	return magazines(c, false)
}

// Musician is the handler for the Musiciansceners page.
func Musician(c echo.Context) error {
	data := empty(c)
	title := "Musicians and composers"
	data["title"] = title
	data["logo"] = title
	data["h1"] = title
	data["description"] = demo
	return scener(c, postgres.Musician, data)
}

// Page404 renders the files page error page for the Artifacts menu and categories.
// It provides different error messages to the standard error page.
func Page404(c echo.Context, uri, page string) error {
	const name = "status"
	errs := fmt.Sprintf("page not found for %q at %q", page, uri)
	if c == nil {
		return InternalErr(c, errs, ErrCxt)
	}
	data := empty(c)
	data["title"] = fmt.Sprintf("%d error, files page not found", http.StatusNotFound)
	data["description"] = fmt.Sprintf("HTTP status %d error", http.StatusNotFound)
	data["code"] = http.StatusNotFound
	data["logo"] = "Page not found"
	data["alert"] = fmt.Sprintf("Artifacts %s page does not exist", uri)
	data["probl"] = "The files page does not exist, there is probably a typo with the URL."
	data["uriOkay"] = fmt.Sprintf("files/%s/", uri)
	data["uriErr"] = page
	err := c.Render(http.StatusNotFound, name, data)
	if err != nil {
		return InternalErr(c, errs, err)
	}
	return nil
}

// PlatformEdit handles the post submission for the Platform selection field.
func PlatformEdit(c echo.Context) error {
	var f Form
	if err := c.Bind(&f); err != nil {
		return badRequest(c, err)
	}
	ctx := context.Background()
	db, err := postgres.ConnectDB()
	if err != nil {
		return badRequest(c, err)
	}
	defer db.Close()
	r, err := model.One(ctx, db, true, f.ID)
	if err != nil {
		return fmt.Errorf("platform edit %w: %d", err, f.ID)
	}
	if err = model.UpdatePlatform(int64(f.ID), f.Value); err != nil {
		return badRequest(c, err)
	}
	return c.JSON(http.StatusOK, r)
}

// PlatformTagInfo handles the POST submission for the platform and tag info.
func PlatformTagInfo(c echo.Context) error {
	var f Form
	if err := c.Bind(&f); err != nil {
		return badRequest(c, err)
	}
	info, err := tags.Platform(f.Platform, f.Tag)
	if err != nil {
		return badRequest(c, err)
	}
	return c.String(http.StatusOK, info)
}

// PostIntro handles the POST request for the intro upload form.
func PostIntro(c echo.Context) error {
	const name = "post intro"
	x, err := c.FormParams()
	if err != nil {
		return InternalErr(c, name, err)
	}
	return c.JSON(http.StatusOK, x)
}

// PostDesc is the handler for the Search for file descriptions form post page.
func PostDesc(c echo.Context, input string) error {
	const name = "artifacts"
	errs := fmt.Sprint("post desc search for,", input)
	ctx := context.Background()
	db, err := postgres.ConnectDB()
	if err != nil {
		return DatabaseErr(c, errs, err)
	}
	defer db.Close()
	terms := helper.SearchTerm(input)
	rel := model.Artifacts{}
	fs, _ := rel.Description(ctx, db, terms)
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
	err = c.Render(http.StatusOK, name, data)
	if err != nil {
		return InternalErr(c, errs, err)
	}
	return nil
}

// PostFilename is the handler for the Search for filenames form post page.
func PostFilename(c echo.Context) error {
	return PostName(c, Filenames)
}

// PostName is the handler for the Search for filenames form post page.
func PostName(c echo.Context, mode FileSearch) error {
	const name = "artifacts"
	errs := fmt.Sprint("post name search for,", mode)
	ctx := context.Background()
	db, err := postgres.ConnectDB()
	if err != nil {
		return DatabaseErr(c, errs, err)
	}
	defer db.Close()
	input := c.FormValue("search-term-query")
	terms := helper.SearchTerm(input)
	rel := model.Artifacts{}
	fs, _ := rel.Filename(ctx, db, terms)
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
	err = c.Render(http.StatusOK, name, data)
	if err != nil {
		return InternalErr(c, errs, err)
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
		return fmt.Errorf("pouet cache %w: %d, want %d", ErrData, l, expect)
	}
	stars, err := strconv.ParseFloat(x[0], 64)
	if err != nil {
		return fmt.Errorf("pouet cache %w: %s", err, x[0])
	}
	vd, err := strconv.Atoi(x[1])
	if err != nil {
		return fmt.Errorf("pouet cache %w: %s", err, x[1])
	}
	vu, err := strconv.Atoi(x[2])
	if err != nil {
		return fmt.Errorf("pouet cache %w: %s", err, x[2])
	}
	vm, err := strconv.Atoi(x[3])
	if err != nil {
		return fmt.Errorf("pouet cache %w: %s", err, x[3])
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
func ProdPouet(c echo.Context, id string) error {
	p := pouet.Production{}
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
func ProdZoo(c echo.Context, id string) error {
	prod := demozoo.Production{}
	i, err := strconv.Atoi(id)
	if err != nil {
		return c.String(http.StatusNotFound, err.Error())
	}
	if code, err := prod.Get(i); err != nil {
		return c.String(code, err.Error())
	}
	if err = c.JSON(http.StatusOK, prod); err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return nil
}

// Interview is the handler for the People Interviews page.
func Reader(c echo.Context) error {
	const title, name = "Textfile reader", "reader"
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
		return InternalErr(c, name, err)
	}
	return nil
}

// ReadmeDel handles the post submission for the Delete readme asset button.
func ReadmeDel(c echo.Context, extraDir string) error {
	var f Form
	if err := c.Bind(&f); err != nil {
		return badRequest(c, err)
	}
	ctx := context.Background()
	db, err := postgres.ConnectDB()
	if err != nil {
		return ErrDB
	}
	defer db.Close()
	r, err := model.One(ctx, db, true, f.ID)
	if err != nil {
		return fmt.Errorf("readme delete %w: %d", err, f.ID)
	}
	if err = command.RemoveMe(r.UUID.String, extraDir); err != nil {
		return badRequest(c, err)
	}
	return c.JSON(http.StatusOK, r)
}

// ReadmePost handles the post submission for the Readme in archive.
func ReadmePost(c echo.Context, logger *zap.SugaredLogger, downloadDir, extraDir string) error {
	const name = "editor readme"
	if logger == nil {
		return InternalErr(c, name, ErrZap)
	}

	var f Form
	if err := c.Bind(&f); err != nil {
		return badRequest(c, err)
	}
	ctx := context.Background()
	db, err := postgres.ConnectDB()
	if err != nil {
		return badRequest(c, err)
	}
	defer db.Close()
	r, err := model.One(ctx, db, true, f.ID)
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
	dst := filepath.Join(extraDir, r.UUID.String+txt)
	ext := filepath.Ext(strings.ToLower(r.Filename.String))
	err = command.ExtractOne(logger, src, dst, ext, target)
	if err != nil {
		return badRequest(c, err)
	}
	return c.JSON(http.StatusOK, r)
}

// ReadmeToggle handles the post submission for the Hide readme from view toggle.
func ReadmeToggle(c echo.Context) error {
	var f Form
	if err := c.Bind(&f); err != nil {
		return badRequest(c, err)
	}
	if err := model.UpdateNoReadme(int64(f.ID), f.Readme); err != nil {
		return badRequest(c, err)
	}
	return c.JSON(http.StatusOK, f)
}

// RecordToggle handles the post submission for the File artifact is online and public toggle.
func RecordToggle(c echo.Context, state bool) error {
	var f Form
	if err := c.Bind(&f); err != nil {
		return badRequest(c, err)
	}
	if state {
		if err := model.UpdateOnline(int64(f.ID)); err != nil {
			return badRequest(c, err)
		}
		return c.JSON(http.StatusOK, f)
	}
	if err := model.UpdateOffline(int64(f.ID)); err != nil {
		return badRequest(c, err)
	}
	return c.JSON(http.StatusOK, f)
}

// Records returns the records for the artifacts category URI.
// Note that the record statistics and counts get cached.
func Records(ctx context.Context, exec boil.ContextExecutor, uri string, page, limit int) (models.FileSlice, error) {
	switch Match(uri) {
	// pulldown editor menu matches
	case forApproval:
		r := model.Artifacts{}
		return r.ByForApproval(ctx, exec, page, limit)
	case deletions:
		r := model.Artifacts{}
		return r.ByHidden(ctx, exec, page, limit)
	case unwanted:
		r := model.Artifacts{}
		return r.ByUnwanted(ctx, exec, page, limit)
	// pulldown menu matches
	case newUploads:
		r := model.Artifacts{}
		return r.ByKey(ctx, exec, page, limit)
	case newUpdates:
		r := model.Artifacts{}
		return r.ByUpdated(ctx, exec, page, limit)
	case oldest:
		r := model.Artifacts{}
		return r.ByOldest(ctx, exec, page, limit)
	case newest:
		r := model.Artifacts{}
		return r.ByNewest(ctx, exec, page, limit)
	}
	return recordsZ(ctx, exec, uri, page, limit)
}

func recordsZ(ctx context.Context, exec boil.ContextExecutor, uri string, page, limit int) (models.FileSlice, error) {
	switch Match(uri) {
	case advert:
		r := model.Advert{}
		return r.List(ctx, exec, page, limit)
	case announcement:
		r := model.Announcement{}
		return r.List(ctx, exec, page, limit)
	case ansi:
		r := model.Ansi{}
		return r.List(ctx, exec, page, limit)
	case ansiBrand:
		r := model.AnsiBrand{}
		return r.List(ctx, exec, page, limit)
	case ansiBBS:
		r := model.AnsiBBS{}
		return r.List(ctx, exec, page, limit)
	case ansiFTP:
		r := model.AnsiFTP{}
		return r.List(ctx, exec, page, limit)
	case ansiNfo:
		r := model.AnsiNfo{}
		return r.List(ctx, exec, page, limit)
	case ansiPack:
		r := model.AnsiPack{}
		return r.List(ctx, exec, page, limit)
	case bbs:
		r := model.BBS{}
		return r.List(ctx, exec, page, limit)
	case bbsImage:
		r := model.BBSImage{}
		return r.List(ctx, exec, page, limit)
	case bbstro:
		r := model.BBStro{}
		return r.List(ctx, exec, page, limit)
	case bbsText:
		r := model.BBSText{}
		return r.List(ctx, exec, page, limit)
	}
	return records0(ctx, exec, uri, page, limit)
}

func records0(ctx context.Context, exec boil.ContextExecutor, uri string, page, limit int) (models.FileSlice, error) {
	switch Match(uri) {
	case database:
		r := model.Database{}
		return r.List(ctx, exec, page, limit)
	case demoscene:
		r := model.Demoscene{}
		return r.List(ctx, exec, page, limit)
	case drama:
		r := model.Drama{}
		return r.List(ctx, exec, page, limit)
	case ftp:
		r := model.FTP{}
		return r.List(ctx, exec, page, limit)
	case hack:
		r := model.Hack{}
		return r.List(ctx, exec, page, limit)
	case htm:
		r := model.HTML{}
		return r.List(ctx, exec, page, limit)
	case howTo:
		r := model.HowTo{}
		return r.List(ctx, exec, page, limit)
	case imageFile:
		r := model.Image{}
		return r.List(ctx, exec, page, limit)
	case imagePack:
		r := model.ImagePack{}
		return r.List(ctx, exec, page, limit)
	case installer:
		r := model.Installer{}
		return r.List(ctx, exec, page, limit)
	case intro:
		r := model.Intro{}
		return r.List(ctx, exec, page, limit)
	case linux:
		r := model.Linux{}
		return r.List(ctx, exec, page, limit)
	case java:
		r := model.Java{}
		return r.List(ctx, exec, page, limit)
	case jobAdvert:
		r := model.JobAdvert{}
		return r.List(ctx, exec, page, limit)
	}
	return records1(ctx, exec, uri, page, limit)
}

func records1(ctx context.Context, exec boil.ContextExecutor, uri string, page, limit int) (models.FileSlice, error) {
	switch Match(uri) {
	case macos:
		r := model.Macos{}
		return r.List(ctx, exec, page, limit)
	case msdosPack:
		r := model.MsDosPack{}
		return r.List(ctx, exec, page, limit)
	case music:
		r := model.Music{}
		return r.List(ctx, exec, page, limit)
	case newsArticle:
		r := model.NewsArticle{}
		return r.List(ctx, exec, page, limit)
	case nfo:
		r := model.Nfo{}
		return r.List(ctx, exec, page, limit)
	case nfoTool:
		r := model.NfoTool{}
		return r.List(ctx, exec, page, limit)
	case standards:
		r := model.Standard{}
		return r.List(ctx, exec, page, limit)
	case script:
		r := model.Script{}
		return r.List(ctx, exec, page, limit)
	case introMsdos:
		r := model.IntroMsDos{}
		return r.List(ctx, exec, page, limit)
	case introWindows:
		r := model.IntroWindows{}
		return r.List(ctx, exec, page, limit)
	case magazine:
		r := model.Magazine{}
		return r.List(ctx, exec, page, limit)
	case msdos:
		r := model.MsDos{}
		return r.List(ctx, exec, page, limit)
	case pdf:
		r := model.PDF{}
		return r.List(ctx, exec, page, limit)
	case proof:
		r := model.Proof{}
		return r.List(ctx, exec, page, limit)
	}
	return records2(ctx, exec, uri, page, limit)
}

func records2(ctx context.Context, exec boil.ContextExecutor, uri string, page, limit int) (models.FileSlice, error) {
	switch Match(uri) {
	case restrict:
		r := model.Restrict{}
		return r.List(ctx, exec, page, limit)
	case takedown:
		r := model.Takedown{}
		return r.List(ctx, exec, page, limit)
	case text:
		r := model.Text{}
		return r.List(ctx, exec, page, limit)
	case textAmiga:
		r := model.TextAmiga{}
		return r.List(ctx, exec, page, limit)
	case textApple2:
		r := model.TextApple2{}
		return r.List(ctx, exec, page, limit)
	case textAtariST:
		r := model.TextAtariST{}
		return r.List(ctx, exec, page, limit)
	case textPack:
		r := model.TextPack{}
		return r.List(ctx, exec, page, limit)
	case tool:
		r := model.Tool{}
		return r.List(ctx, exec, page, limit)
	case trialCrackme:
		r := model.TrialCrackme{}
		return r.List(ctx, exec, page, limit)
	case video:
		r := model.Video{}
		return r.List(ctx, exec, page, limit)
	case windows:
		r := model.Windows{}
		return r.List(ctx, exec, page, limit)
	case windowsPack:
		r := model.WindowsPack{}
		return r.List(ctx, exec, page, limit)
	default:
		return nil, fmt.Errorf("artifacts category %w: %s", ErrCategory, uri)
	}
}

// Releaser is the handler for the releaser page ordered by the most files.
func Releaser(c echo.Context) error {
	return releasers(c, model.Prolific)
}

// ReleaserAZ is the handler for the releaser page ordered alphabetically.
func ReleaserAZ(c echo.Context) error {
	return releasers(c, model.Alphabetical)
}

// ReleaserYear is the handler for the releaser page ordered by year of the first release.
func ReleaserYear(c echo.Context) error {
	return releasers(c, model.Oldest)
}

// Releaser404 renders the files error page for the Groups menu and invalid releasers.
func Releaser404(c echo.Context, id string) error {
	const name = "status"
	errs := fmt.Sprint("releaser page not found for,", id)
	if c == nil {
		return InternalErr(c, errs, ErrCxt)
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
		return InternalErr(c, errs, err)
	}
	return nil
}

// ReleaserEdit handles the post submission for the Platform selection field.
func ReleaserEdit(c echo.Context) error {
	var f Form
	if err := c.Bind(&f); err != nil {
		return badRequest(c, err)
	}
	ctx := context.Background()
	db, err := postgres.ConnectDB()
	if err != nil {
		return badRequest(c, err)
	}
	defer db.Close()
	r, err := model.One(ctx, db, true, f.ID)
	if err != nil {
		return fmt.Errorf("releaser edit  %w: %d", err, f.ID)
	}
	if err = model.UpdateReleasers(int64(f.ID), f.Value); err != nil {
		return badRequest(c, err)
	}
	return c.JSON(http.StatusOK, r)
}

// Releasers is the handler for the list and preview of files credited to a releaser.
func Releasers(c echo.Context, uri string) error {
	const name = "artifacts"
	errs := fmt.Sprint("releasers page for,", uri)
	ctx := context.Background()
	db, err := postgres.ConnectDB()
	if err != nil {
		return InternalErr(c, errs, err)
	}
	defer db.Close()

	s := releaser.Link(uri)
	rel := model.Releasers{}
	fs, err := rel.Where(ctx, db, uri)
	if err != nil {
		return InternalErr(c, errs, err)
	}
	if len(fs) == 0 {
		return Releaser404(c, uri)
	}
	data := emptyFiles(c)
	data["title"] = "Artifacts for " + s
	data["h1"] = s
	data["lead"] = initialism.Join(initialism.Path(uri))
	data["logo"] = s
	data["description"] = "The collection of files for " + s + "."
	data["demozoo"] = strconv.Itoa(int(demozoo.Find(uri)))
	data["sixteen"] = sixteen.Find(uri)
	data["website"] = site.Find(uri)
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
		return InternalErr(c, errs, err)
	}
	data["stats"] = d
	err = c.Render(http.StatusOK, name, data)
	if err != nil {
		return InternalErr(c, errs, err)
	}
	return nil
}

// releaserSum is a helper function for Releasers that returns the statistics for the files page.
func releaserSum(ctx context.Context, exec boil.ContextExecutor, uri string) (map[string]string, error) {
	m := model.Summary{}
	if err := m.ByReleaser(ctx, exec, uri); err != nil {
		return nil, fmt.Errorf("releaser sum %w: %s", err, uri)
	}
	d := map[string]string{
		"files": string(ByteFileS("file", m.SumCount.Int64, m.SumBytes.Int64)),
		"years": helper.Years(m.MinYear.Int16, m.MaxYear.Int16),
	}
	return d, nil
}

// Scener is the handler for the page to list all the sceners.
func Scener(c echo.Context) error {
	data := empty(c)
	title := "Sceners, the people of The Scene"
	data["title"] = title
	data["logo"] = title
	data["h1"] = title
	data["description"] = demo
	return scener(c, postgres.Roles(), data)
}

// Scener404 renders the files error page for the People menu and invalid sceners.
func Scener404(c echo.Context, id string) error {
	const name = "status"
	errs := fmt.Sprint("scener page not found for,", id)
	if c == nil {
		return InternalErr(c, errs, ErrCxt)
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
		return InternalErr(c, errs, err)
	}
	return nil
}

// Sceners is the handler for the list and preview of files credited to a scener.
func Sceners(c echo.Context, uri string) error {
	const name = "artifacts"
	errs := fmt.Sprint("sceners page for,", uri)
	ctx := context.Background()
	db, err := postgres.ConnectDB()
	if err != nil {
		return InternalErr(c, errs, err)
	}
	defer db.Close()

	s := releaser.Link(uri)
	var ms model.Scener
	fs, err := ms.Where(ctx, db, uri)
	if err != nil {
		return InternalErr(c, errs, err)
	}
	if len(fs) == 0 {
		return Scener404(c, uri)
	}
	data := emptyFiles(c)
	data["title"] = s + attr
	data["h1"] = s
	data["lead"] = "Artifacts attributed to " + s + "."
	data["logo"] = s
	data["description"] = "The collection of files attributed to " + s + "."
	data["scener"] = s
	data[records] = fs
	d, err := scenerSum(ctx, db, uri)
	if err != nil {
		return InternalErr(c, errs, err)
	}
	data["stats"] = d
	err = c.Render(http.StatusOK, name, data)
	if err != nil {
		return InternalErr(c, errs, err)
	}
	return nil
}

// SearchDesc is the handler for the Search for file descriptions page.
func SearchDesc(c echo.Context) error {
	const title, name = "Search titles and descriptions", "searchpost"
	data := empty(c)
	data["description"] = "Search form to scan through file descriptions."
	data["logo"] = title
	data["title"] = title
	data["info"] = "search the metadata descriptions of file artifacts"
	err := c.Render(http.StatusOK, name, data)
	if err != nil {
		return InternalErr(c, name, err)
	}
	return nil
}

// SearchID is the handler for the Record by ID Search page.
func SearchID(c echo.Context) error {
	const title, name = "Search for artifacts", "searchhtmx"
	data := empty(c)
	data["description"] = "Search form to discover artifacts by ID."
	data["logo"] = title
	data["title"] = title
	data["info"] = "search for artifacts by their record id, uuid or URL key"
	data["hxPost"] = "/editor/search/id"
	data["inputPlaceholder"] = "Type to search for an artifact…"
	err := c.Render(http.StatusOK, name, data)
	if err != nil {
		return InternalErr(c, name, err)
	}
	return nil
}

// SearchFile is the handler for the Search for files page.
func SearchFile(c echo.Context) error {
	const title, name = "Search for filenames", "searchpost"
	data := empty(c)
	data["description"] = "Search form to discover files."
	data["logo"] = title
	data["title"] = title
	data["info"] = "search for filenames or extensions"
	err := c.Render(http.StatusOK, name, data)
	if err != nil {
		return InternalErr(c, name, err)
	}
	return nil
}

// SearchReleaser is the handler for the Releaser Search page.
func SearchReleaser(c echo.Context) error {
	const title, name = "Search for releasers", "searchhtmx"
	data := empty(c)
	data["description"] = "Search form to discover releasers."
	data["logo"] = title
	data["title"] = title
	data["info"] = "search for a group, initialism, magazine, board, or site"
	data["hxPost"] = "/search/releaser"
	data["inputPlaceholder"] = "Type to search for a releaser…"
	err := c.Render(http.StatusOK, name, data)
	if err != nil {
		return InternalErr(c, name, err)
	}
	return nil
}

// SignedOut is the handler to sign out and remove the current session.
func SignedOut(c echo.Context) error {
	const name = "signedout"
	{ // get any existing session
		sess, err := session.Get(sess.Name, c)
		if err != nil {
			return BadRequestErr(c, name, err)
		}
		id, subExists := sess.Values["sub"]
		if !subExists || id == "" {
			return ForbiddenErr(c, name, ErrSession)
		}
		const remove = -1
		sess.Options.MaxAge = remove
		err = sess.Save(c.Request(), c.Response())
		if err != nil {
			return InternalErr(c, name, err)
		}
	}
	return c.Redirect(http.StatusFound, "/")
}

// SignOut is the handler for the Sign out of Defacto2 page.
func SignOut(c echo.Context) error {
	const name = "signout"
	data := empty(c)
	data["title"] = "Sign out"
	data["description"] = "Sign out of Defacto2."
	data["h1"] = "Sign out"
	err := c.Render(http.StatusOK, name, data)
	if err != nil {
		return InternalErr(c, name, err)
	}
	return nil
}

// Signin is the handler for the Sign in session page.
func Signin(c echo.Context, clientID, nonce string) error {
	const name = "signin"
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
			return remove(c, name, data)
		}
		subID, subExists := sess.Values["sub"]
		if !subExists {
			return remove(c, name, data)
		}
		val, valExists := subID.(string)
		if valExists && val != "" {
			return SignOut(c)
		}
	}
	err := c.Render(http.StatusOK, name, data)
	if err != nil {
		return InternalErr(c, name, err)
	}
	return nil
}

// Statistics returns the empty database statistics for the artifacts categories.
func Statistics() Stats {
	return Stats{}
}

// TagEdit handles the post submission for the Tag selection field.
func TagEdit(c echo.Context) error {
	var f Form
	if err := c.Bind(&f); err != nil {
		return badRequest(c, err)
	}
	ctx := context.Background()
	db, err := postgres.ConnectDB()
	if err != nil {
		return badRequest(c, err)
	}
	defer db.Close()
	r, err := model.One(ctx, db, true, f.ID)
	if err != nil {
		return fmt.Errorf("tag edit %w: %d", err, f.ID)
	}
	if err = model.UpdateTag(int64(f.ID), f.Value); err != nil {
		return badRequest(c, err)
	}
	return c.JSON(http.StatusOK, r)
}

// TagInfo handles the POST submission for the platform and tag info.
func TagInfo(c echo.Context) error {
	var f Form
	if err := c.Bind(&f); err != nil {
		return badRequest(c, err)
	}
	info, err := tags.Description(f.Tag)
	if err != nil {
		return badRequest(c, err)
	}
	return c.String(http.StatusOK, info)
}

// Thanks is the handler for the Thanks page.
func Thanks(c echo.Context) error {
	const name = "thanks"
	data := empty(c)
	data["description"] = "Defacto2 thankyous."
	data["h1"] = "Thank you!"
	data["lead"] = "Thanks to the hundreds of people who have contributed to" +
		" Defacto2 over the decades with file submissions, " +
		"hard drive donations, interviews, corrections, artwork, and monetary contributions!"
	data["title"] = "Thanks!"
	err := c.Render(http.StatusOK, name, data)
	if err != nil {
		return InternalErr(c, name, err)
	}
	return nil
}

// TheScene is the handler for the The Scene page.
func TheScene(c echo.Context) error {
	const name = "thescene"
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
		return InternalErr(c, name, err)
	}
	return nil
}

// TitleEdit handles the post submission for the Delete readme asset button.
func TitleEdit(c echo.Context) error {
	var f Form
	if err := c.Bind(&f); err != nil {
		return badRequest(c, err)
	}
	ctx := context.Background()
	db, err := postgres.ConnectDB()
	if err != nil {
		return badRequest(c, err)
	}
	defer db.Close()
	r, err := model.One(ctx, db, true, f.ID)
	if err != nil {
		return fmt.Errorf("title edit %w: %d", err, f.ID)
	}
	if err = model.UpdateTitle(int64(f.ID), f.Value); err != nil {
		return badRequest(c, err)
	}
	return c.JSON(http.StatusOK, r)
}

// VotePouet is the handler for the Pouet production votes JSON page.
func VotePouet(c echo.Context, logger *zap.SugaredLogger, id string) error {
	const title, name, sep = "Pouet", "pouet", ";"
	if logger == nil {
		return InternalErr(c, name, ErrZap)
	}
	pv := pouet.Votes{}
	i, err := strconv.Atoi(id)
	if err != nil {
		return c.String(http.StatusNotFound, err.Error())
	}

	cp := cache.Pouet
	if s, err := cp.Read(id); err == nil {
		if err := PouetCache(c, s); err == nil {
			logger.Debugf("cache hit for pouet id %s", id)
			return nil
		}
	}
	logger.Debugf("cache miss for pouet id %s", id)

	if err = pv.Votes(i); err != nil {
		return c.String(http.StatusNotFound, err.Error())
	}

	if err = c.JSON(http.StatusOK, pv); err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	val := fmt.Sprintf("%.1f%s%d%s%d%s%d",
		pv.Stars, sep, pv.VotesDown, sep, pv.VotesUp, sep, pv.VotesMeh)
	if err := cp.Write(id, val, cache.ExpiredAt); err != nil {
		logger.Errorf("failed to write pouet id %s to cache db: %s", id, err)
	}
	return nil
}

// Website is the handler for the websites page.
// Open is the ID of the accordion section to open.
func Website(c echo.Context, open string) error {
	const name = "websites"
	data := empty(c)
	data["title"] = "Websites"
	const logo = "Videos, Books, Films, Sites, Podcasts"
	data["logo"] = logo
	data["description"] = "A collection of " + logo + " about the scene."
	acc := List()
	// Open the accordion section.
	closeAll := true
	for i, site := range acc {
		if site.ID == open || open == "" {
			site.Open = true
			closeAll = false
			acc[i] = site
			if open == "" {
				continue
			}
			break
		}
	}
	if closeAll {
		data["title"] = "Website categories"
	}
	// If a section was requested but not found, return a 404.
	if open != "hide" && closeAll {
		return StatusErr(c, http.StatusNotFound, open)
	}
	// Render the page.
	data["accordion"] = acc
	err := c.Render(http.StatusOK, name, data)
	if err != nil {
		return InternalErr(c, fmt.Sprint("render open website,", open), err)
	}
	return nil
}

// Writer is the handler for the Writer page.
func Writer(c echo.Context) error {
	data := empty(c)
	title := "Writers, editors and authors"
	data["title"] = title
	data["logo"] = title
	data["h1"] = title
	data["description"] = demo
	return scener(c, postgres.Writer, data)
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
	Record    model.Artifacts
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
func (s *Stats) Get(ctx context.Context, exec boil.ContextExecutor) error {
	if err := s.Record.Public(ctx, exec); err != nil {
		return fmt.Errorf("category get record stat: %w", err)
	}
	if err := s.Ansi.Stat(ctx, exec); err != nil {
		return fmt.Errorf("category get ansi stat: %w", err)
	}
	if err := s.AnsiBBS.Stat(ctx, exec); err != nil {
		return fmt.Errorf("category get ansiBBS stat: %w", err)
	}
	if err := s.BBS.Stat(ctx, exec); err != nil {
		return fmt.Errorf("category get bbs stat: %w", err)
	}
	if err := s.BBSText.Stat(ctx, exec); err != nil {
		return fmt.Errorf("category get bbs trext stat: %w", err)
	}
	if err := s.BBStro.Stat(ctx, exec); err != nil {
		return fmt.Errorf("category get bbstro stat: %w", err)
	}
	if err := s.MsDos.Stat(ctx, exec); err != nil {
		return fmt.Errorf("category get msdos stat: %w", err)
	}
	if err := s.Intro.Stat(ctx, exec); err != nil {
		return fmt.Errorf("category get intro stat: %w", err)
	}
	if err := s.IntroD.Stat(ctx, exec); err != nil {
		return fmt.Errorf("category get introd stat: %w", err)
	}
	if err := s.IntroW.Stat(ctx, exec); err != nil {
		return fmt.Errorf("category get introw stat: %w", err)
	}
	if err := s.Installer.Stat(ctx, exec); err != nil {
		return fmt.Errorf("category get installer stat: %w", err)
	}
	if err := s.Java.Stat(ctx, exec); err != nil {
		return fmt.Errorf("category get java stat: %w", err)
	}
	if err := s.Linux.Stat(ctx, exec); err != nil {
		return fmt.Errorf("category get linux stat: %w", err)
	}
	if err := s.Demoscene.Stat(ctx, exec); err != nil {
		return fmt.Errorf("category get demoscene stat: %w", err)
	}
	return s.get(ctx, exec)
}

func (s *Stats) get(ctx context.Context, exec boil.ContextExecutor) error {
	if err := s.Macos.Stat(ctx, exec); err != nil {
		return fmt.Errorf("category get macos stat: %w", err)
	}
	if err := s.Magazine.Stat(ctx, exec); err != nil {
		return fmt.Errorf("category get magazine stat: %w", err)
	}
	if err := s.Nfo.Stat(ctx, exec); err != nil {
		return fmt.Errorf("category get nfo stat: %w", err)
	}
	if err := s.NfoTool.Stat(ctx, exec); err != nil {
		return fmt.Errorf("category get nfoTool stat: %w", err)
	}
	if err := s.Proof.Stat(ctx, exec); err != nil {
		return fmt.Errorf("category get proof stat: %w", err)
	}
	if err := s.Script.Stat(ctx, exec); err != nil {
		return fmt.Errorf("category get script stat: %w", err)
	}
	if err := s.Text.Stat(ctx, exec); err != nil {
		return fmt.Errorf("category get text stat: %w", err)
	}
	return s.Windows.Stat(ctx, exec)
}

type Pagination struct {
	BaseURL   string // BaseURL is the base URL for the pagination links.
	CurrPage  int    // CurrPage is the current page number.
	SumPages  int    // SumPages is the total number of pages.
	PrevPage  int    // PrevPage is the previous page number.
	NextPage  int    // NextPage is the next page number.
	TwoBelow  int    // TwoBelow is the page number two below the current page.
	TwoAfter  int    // TwoAfter is the page number two after the current page.
	RangeStep int    // RangeStep is the number of pages to skip in the pagination range.
}

// artifacts is a helper function for Artifacts that returns the data map for the files page.
func artifacts(c echo.Context, uri string, page int) error {
	const title, name = "Artifacts", "artifacts"
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
	data["forApproval"] = false
	switch uri {
	case
		newUploads.String(),
		newUpdates.String(),
		deletions.String(),
		unwanted.String():
		data["unknownYears"] = false
	case forApproval.String():
		data["forApproval"] = true
	}
	errs := fmt.Sprintf("artifacts page %d for %q", page, uri)
	ctx := context.Background()
	db, err := postgres.ConnectDB()
	if err != nil {
		return InternalErr(c, errs, err)
	}
	defer db.Close()
	r, err := Records(ctx, db, uri, page, limit)
	if err != nil {
		return DatabaseErr(c, errs, err)
	}
	data[records] = r
	d, sum, err := stats(ctx, db, uri)
	if err != nil {
		return DatabaseErr(c, errs, err)
	}
	data["stats"] = d
	lastPage := math.Ceil(float64(sum) / float64(limit))
	if page > int(lastPage) {
		i := strconv.Itoa(page)
		return Page404(c, uri, i)
	}
	const pages = 2
	data["Pagination"] = Pagination{
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
		return InternalErr(c, errs, err)
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

// bbsHandler is the handler for the BBS page.
func bbsHandler(c echo.Context, orderBy model.OrderBy) error {
	const title, name = "BBS", "bbs"
	const lead = "Bulletin Board Systems are historical, " +
		"networked personal computer servers connected using the landline telephone network and provide forums, " +
		"real-time chat, mail, and file sharing for The Scene \"elites.\""
	const logo = "Bulletin Board Systems"
	const key = "releasers"
	data := empty(c)
	data["title"] = title
	data["description"] = lead
	data["logo"] = logo
	data["h1"] = title
	data["lead"] = lead
	data["itemName"] = name
	data[key] = model.Releasers{}
	data["stats"] = map[string]string{}

	ctx := context.Background()
	db, err := postgres.ConnectDB()
	if err != nil {
		return InternalErr(c, name, err)
	}
	defer db.Close()
	r := model.Releasers{}
	if err := r.BBS(ctx, db, orderBy); err != nil {
		return DatabaseErr(c, name, err)
	}
	data[key] = r
	tmpl := name
	var order string
	switch orderBy {
	case model.Alphabetical:
		s := logo + az
		data["logo"] = s
		data["title"] = title + az
		order = alpha
	case model.Prolific:
		s := logo + ", by count"
		data["logo"] = s
		order = "by file artifact count"
	case model.Oldest:
		tmpl = "bbs-year"
		s := logo + byyear
		data["title"] = title + byyear
		data["logo"] = s
		order = year
	}
	data["stats"] = map[string]string{
		"pubs":    fmt.Sprintf("%d boards", len(r)),
		"orderBy": order,
	}

	err = c.Render(http.StatusOK, tmpl, data)
	if err != nil {
		return InternalErr(c, name, err)
	}
	return nil
}

// counter returns the statistics for the artifacts categories.
func counter() (Stats, error) {
	ctx := context.Background()
	db, err := postgres.ConnectDB()
	if err != nil {
		return Stats{}, fmt.Errorf("cartifacts categories counter tx %w", err)
	}
	defer db.Close()
	counter := Stats{}
	if err := counter.Get(ctx, db); err != nil {
		return Stats{}, fmt.Errorf("cartifacts categories counter get %w", err)
	}
	return counter, nil
}

// empty is a map of default values for the app templates.
func empty(c echo.Context) map[string]interface{} {
	// the keys are listed in order of appearance in the templates.
	// * marked keys are required.
	// ! marked keys are suggested.
	return map[string]interface{}{
		// The number of records of files in the database.
		"cacheFiles": Caching.RecordCount,
		// A canonical URL is the URL of the best representative page from a group of duplicate pages.
		"canonical": "",
		// The ID of the carousel to display.
		"carousel": "",
		// Empty database counts for files and categories.
		"counter": Statistics(),
		// If true, the database is not available.
		"dbError": false,
		// * A short description of the page that get inserted into the description meta element.
		"description": "",
		// If true, the editor mode is enabled.
		"editor": sess.Editor(c),
		// ! The H1 heading of the page.
		"h1": "",
		// The H1 sub-heading of the page.
		"h1Sub": "",
		// If true, the large, js-dos v6.22 emulator files will be loaded.
		"jsdos6": false,
		// ! The enlarged, lead paragraph of the page.
		"lead": "",
		// ! Text to insert into the monospaced, ASCII art logo.
		"logo": "",
		// If true, the application is in read-only mode.
		"readOnly": true,
		// * The title of the page that get inserted into the title meta element.
		"title": "",
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
		return data, fmt.Errorf("counter: %w", err)
	}
	data["counter"] = c
	data["logo"] = "Artifact category statistics"
	data["lead"] = "This page shows the artifacts categories with selected statistics, " +
		"such as the number of files in the category or platform."
	return data, nil
}

// magazines is the handler for the magazine page.
func magazines(c echo.Context, chronological bool) error {
	const title, name = "Magazines", "magazine"
	data := empty(c)
	const lead = "The magazines are newsletters, reports, " +
		"and publications about activities within The Scene subculture."
	const issue = "issue"
	const key = "releasers"
	data["title"] = title
	data["description"] = lead
	data["logo"] = title
	data["h1"] = title
	data["lead"] = lead
	data["itemName"] = issue
	data[key] = model.Releasers{}
	data["stats"] = map[string]string{}

	ctx := context.Background()
	db, err := postgres.ConnectDB()
	if err != nil {
		return InternalErr(c, name, err)
	}
	defer db.Close()
	var order string
	r := model.Releasers{}
	switch chronological {
	case true:
		if err := r.Magazine(ctx, db); err != nil {
			return DatabaseErr(c, name, err)
		}
		s := title + byyear
		data["logo"] = s
		data["title"] = title + byyear
		order = year
	case false:
		if err := r.MagazineAZ(ctx, db); err != nil {
			return DatabaseErr(c, name, err)
		}
		s := title + az
		data["logo"] = s
		data["title"] = title + az
		order = alpha
	}
	data[key] = r
	data["stats"] = map[string]string{
		"pubs":    fmt.Sprintf("%d publications", len(r)),
		"orderBy": order,
	}
	err = c.Render(http.StatusOK, name, data)
	if err != nil {
		return InternalErr(c, name, err)
	}
	return nil
}

// postStats is a helper function for PostName that returns the statistics for the files page.
func (mode FileSearch) postStats(ctx context.Context, exec boil.ContextExecutor, terms []string) map[string]string {
	none := func() map[string]string {
		return map[string]string{
			"files": "no files found",
			"years": "",
		}
	}
	m := model.Summary{}
	switch mode {
	case Filenames:
		if err := m.ByFilename(ctx, exec, terms); err != nil {
			return none()
		}
	case Descriptions:
		if err := m.ByDescription(ctx, exec, terms); err != nil {
			return none()
		}
	}
	if m.SumCount.Int64 == 0 {
		return none()
	}
	d := map[string]string{
		"files": string(ByteFileS("file", m.SumCount.Int64, m.SumBytes.Int64)),
		"years": helper.Years(m.MinYear.Int16, m.MaxYear.Int16),
	}
	return d
}

// remove is a helper function to remove the session cookie by setting the MaxAge to -1.
func remove(c echo.Context, name string, data map[string]interface{}) error {
	sess, err := session.Get(sess.Name, c)
	if err != nil {
		const remove = -1
		sess.Options.MaxAge = remove
		_ = sess.Save(c.Request(), c.Response())
	}
	err = c.Render(http.StatusOK, name, data)
	if err != nil {
		return InternalErr(c, name, err)
	}
	return nil
}

// scenerSum is a helper function for Sceners that returns the statistics for the files page.
func scenerSum(ctx context.Context, exec boil.ContextExecutor, uri string) (map[string]string, error) {
	m := model.Summary{}
	if err := m.ByScener(ctx, exec, uri); err != nil {
		return nil, fmt.Errorf("scener sum %w: %s", err, uri)
	}
	d := map[string]string{
		"files": string(ByteFileS("file", m.SumCount.Int64, m.SumBytes.Int64)),
		"years": helper.Years(m.MinYear.Int16, m.MaxYear.Int16),
	}
	return d, nil
}

// scener is the handler for the scener pages.
func scener(c echo.Context, r postgres.Role,
	data map[string]interface{},
) error {
	const name = "scener"
	s := model.Sceners{}
	ctx := context.Background()
	db, err := postgres.ConnectDB()
	if err != nil {
		return InternalErr(c, name, err)
	}
	defer db.Close()
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
		err = s.Distinct(ctx, db)
	}
	if err != nil {
		return DatabaseErr(c, name, err)
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
		return InternalErr(c, name, err)
	}
	return nil
}

// sessionHandler creates a [new session] and populates it with
// the claims data created by the [ID Tokens for Google HTTP APIs].
//
// [new session]: https://pkg.go.dev/github.com/gorilla/sessions
// [ID Tokens for Google HTTP APIs]: https://pkg.go.dev/google.golang.org/api/idtoken
func sessionHandler(
	c echo.Context, maxAge int,
	claims map[string]interface{},
) error {
	session, err := session.Get(sess.Name, c)
	if err != nil {
		return fmt.Errorf("app session get: %w", err)
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

	const uniqueGoogleID = "sub"
	val, valExists := claims[uniqueGoogleID]
	if !valExists {
		return ErrClaims
	}
	session.Values[uniqueGoogleID] = val
	session.Values["givenName"] = claims["given_name"]
	session.Values["email"] = claims["email"]
	session.Values["emailVerified"] = claims["email_verified"]

	// save the session
	return session.Save(c.Request(), c.Response())
}

// stats is a helper function for Artifacts that returns the statistics for the files page.
func stats(ctx context.Context, exec boil.ContextExecutor, uri string) (map[string]string, int, error) {
	if !Valid(uri) {
		return nil, 0, nil
	}
	m := model.Summary{}
	err := m.ByMatch(ctx, exec, uri)
	if err != nil && !errors.Is(err, model.ErrURI) {
		return nil, 0, fmt.Errorf("artifacts stats %w: %s", err, uri)
	}
	if errors.Is(err, model.ErrURI) {
		switch uri {
		case "for-approval":
			if err := m.ByForApproval(ctx, exec); err != nil {
				return nil, 0, fmt.Errorf("artifacts stats for approval %w: %s", err, uri)
			}
		case "deletions":
			if err := m.ByHidden(ctx, exec); err != nil {
				return nil, 0, fmt.Errorf("artifacts stats by hidden %w: %s", err, uri)
			}
		case "unwanted":
			if err := m.ByUnwanted(ctx, exec); err != nil {
				return nil, 0, fmt.Errorf("artifacts stats unwanted %w: %s", err, uri)
			}
		default:
			if err := m.ByPublic(ctx, exec); err != nil {
				return nil, 0, fmt.Errorf("artifacts stats by public %w: %s", err, uri)
			}
		}
	}
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

// releasers is the handler for the Releaser page.
func releasers(c echo.Context, orderBy model.OrderBy) error {
	const title, name = "Releaser", "releaser"
	data := empty(c)
	const lead = "A releaser is a brand or a collective group of " +
		"sceners responsible for releasing or distributing products."
	const logo = "Groups and releasers"
	const key = "releasers"
	data["title"] = title
	data["description"] = fmt.Sprint(title, " ", lead)
	data["logo"] = logo
	data["h1"] = title
	data["lead"] = lead
	data["itemName"] = "file"
	data[key] = model.Releasers{}
	data["stats"] = map[string]string{}

	ctx := context.Background()
	db, err := postgres.ConnectDB()
	if err != nil {
		return InternalErr(c, name, err)
	}
	defer db.Close()
	var r model.Releasers
	if err := r.Limit(ctx, db, orderBy, 0, 0); err != nil {
		return DatabaseErr(c, name, err)
	}
	data[key] = r
	tmpl := name
	var order string
	switch orderBy {
	case model.Alphabetical:
		s := logo + az
		data["logo"] = s
		data["title"] = title + az
		order = alpha
	case model.Prolific:
		s := logo + ", by count"
		data["logo"] = s
		order = "by file artifact count"
	case model.Oldest:
		tmpl = "releaser-year"
		s := logo + byyear
		data["logo"] = s
		data["title"] = title + byyear
		order = year
	}
	data["stats"] = map[string]string{
		"pubs":    fmt.Sprintf("%d releasers and groups", len(r)),
		"orderBy": order,
	}
	err = c.Render(http.StatusOK, tmpl, data)
	if err != nil {
		return InternalErr(c, tmpl, err)
	}
	return nil
}
