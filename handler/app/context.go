package app

// Package file context.go contains the router handlers for the Defacto2 website.

import (
	"context"
	"crypto/sha512"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"html/template"
	"log/slog"
	"math"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"sync"

	"github.com/Defacto2/helper"
	"github.com/Defacto2/releaser"
	"github.com/Defacto2/releaser/initialism"
	"github.com/Defacto2/server/handler/app/internal/fileslice"
	"github.com/Defacto2/server/handler/app/remote"
	"github.com/Defacto2/server/handler/areacode"
	"github.com/Defacto2/server/handler/cache"
	"github.com/Defacto2/server/handler/csdb"
	"github.com/Defacto2/server/handler/demozoo"
	"github.com/Defacto2/server/handler/download"
	"github.com/Defacto2/server/handler/janeway"
	"github.com/Defacto2/server/handler/pouet"
	"github.com/Defacto2/server/handler/sess"
	"github.com/Defacto2/server/handler/site"
	"github.com/Defacto2/server/handler/sixteen"
	"github.com/Defacto2/server/handler/tidbit"
	"github.com/Defacto2/server/internal/config"
	"github.com/Defacto2/server/internal/dir"
	"github.com/Defacto2/server/internal/panics"
	"github.com/Defacto2/server/internal/postgres"
	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/Defacto2/server/internal/tags"
	"github.com/Defacto2/server/model"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/google/uuid"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"google.golang.org/api/idtoken"
)

// FileSearch is the type of search to perform.
type FileSearch int

const (
	Filenames    FileSearch = iota // Filenames is the search for filenames.
	Descriptions                   // Descriptions is the search for file descriptions and titles.
)

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

const (
	demo    = "demo"
	limit   = 198 // per-page record limit
	records = "records"
	sep     = ";"
	az      = ", a-z"
	byyear  = " by year"
	alpha   = "alphabetically"
	year    = "by year"
)

// Empty is a map of default values for an app template that are used by the layout template,
// that is the base template for all pages.
//
// All keys are optional except for the "description" and "title".
//   - The "description" is used by the meta description element.
//   - The "title" is used by the title element.
//
// The following optional keys are recommended:
//   - "h1" is the H1 heading of the page.
//   - "lead" is the lead or introduction paragraph of the page.
//   - "logo" is the brief text inserted into the ASCII art logo.
//
// Other optional keys are also available:
//   - "canonical" is the canonical URL of the best representative page from a group of duplicate pages.
//   - "carousel" is the ID of the carousel to display.
//   - "databaseErr" is true if the database is not available.
//   - "subheading" is the H1 sub-heading of the page.
//   - "jsdos6" is true if the js-dos v6.22 emulator files are to be loaded.
//   - "readonlymode" is true if the application is in read-only mode.
//
// These keys are autofilled:
//   - "cachefiles" is the total number of records and used by the defacto2:file-count meta element.
//   - "editor" is true if the editor mode is enabled for the browser session.
func empty(c echo.Context) map[string]any {
	return map[string]any{
		"cachefiles":   Caching.RecordCount, // autofilled
		"canonical":    "",
		"carousel":     "",
		"databaseErr":  false,
		"description":  "",
		"editor":       sess.Editor(c), // autofilled
		"h1":           "",
		"subheading":   "",
		"jsdos6":       false,
		"lead":         "",
		"logo":         "",
		"readonlymode": true,
		"title":        "",
		"ogtitle":      "",    // opengraph title
		"noindex":      false, // flag the layout to include robots=noindex metatag
	}
}

// NOTE:
// - The Artifacts handler below is for multiple artifacts with pagination.
// - The releasers handler are for the releasers and groups without pagination.
// - The individual artifact hander is Dirs.Artifact and is found in the dirs.go file.

// EmptyTester is a map of defaults for the app template tests.
func EmptyTester(c echo.Context) map[string]any {
	return empty(c)
}

// Artifacts is the handler for the list and preview of the files page.
// The uri is the category or collection of files to display.
// The page is the page number of the results to display.
func Artifacts(c echo.Context, db *sql.DB, sl *slog.Logger, uri, page string) error {
	const msg = "artifacts context"
	if err := panics.EchoContextDS(c, db, sl); err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}
	if !fileslice.Valid(uri) {
		return Artifacts404(c, sl, uri)
	}
	if page == "" {
		return artifacts(c, db, sl, uri, 1)
	}
	p, err := strconv.Atoi(page)
	if err != nil {
		return Page404(c, sl, uri, page)
	}
	return artifacts(c, db, sl, uri, p)
}

// artifacts is a helper function for Artifacts that returns the data map for the files page.
func artifacts(c echo.Context, db *sql.DB, sl *slog.Logger, uri string, page int) error {
	const title, name = "Artifacts", "artifacts"
	logo, subhead, lead := fileslice.FileInfo(uri)
	data := emptyFiles(c)
	data["title"] = title
	data["canonical"] = strings.Join([]string{"files", uri}, "/")
	data["description"] = "Table of contents for the collection of artifacts."
	data["logo"] = logo
	data["h1"] = title
	data["subheading"] = subhead
	data["lead"] = lead
	data[records] = []models.FileSlice{}
	data["unknownYears"] = true
	data["forApproval"] = false
	errs := fmt.Sprintf("artifacts page %d for %q", page, uri)
	ctx := context.Background()
	r, err := fileslice.Records(ctx, db, uri, page, limit)
	if err != nil {
		return DatabaseErr(c, sl, errs, err)
	}
	data[records] = r
	d, sum, err := stats(ctx, db, uri)
	if err != nil {
		return DatabaseErr(c, sl, errs, err)
	}
	data = artifactsDesc(uri, d["years"], sum, data)
	data["stats"] = d
	lastPage := math.Ceil(float64(sum) / float64(limit))
	if len(r) == 0 {
		if err = c.Render(http.StatusOK, name, data); err != nil {
			return InternalErr(c, sl, errs, err)
		}
		return nil
	}
	if page > int(lastPage) {
		i := strconv.Itoa(page)
		return Page404(c, sl, uri, i)
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
	if err = c.Render(http.StatusOK, name, data); err != nil {
		return InternalErr(c, sl, errs, err)
	}
	return nil
}

func artifactsDesc(uri, years string, sum int, data map[string]any) map[string]any {
	data["noindex"] = true
	switch fileslice.Match(uri) {
	case fileslice.NewUploads:
		data["description"] = "These are the most recent additions of scene history to the site."
		data["title"] = "New additions"
		data["unknownYears"] = false
	case fileslice.NewUpdates:
		data["description"] = "Artifacts that have been recently updated or revised."
		data["title"] = "Artifact revisions and changes"
		data["unknownYears"] = false
	case fileslice.Deletions:
		data["title"] = "Deleted artifacts"
		data["unknownYears"] = false
	case fileslice.Unwanted:
		data["title"] = "Dangerous files artifacts"
		data["unknownYears"] = false
	case fileslice.ForApproval:
		data["title"] = "For Approval artifacts"
		data["forApproval"] = true
	case fileslice.Oldest:
		data["description"] = "These are the oldest known artifacts held by the site."
		data["title"] = "Oldest artifacts"
		data["noindex"] = false
	case fileslice.Newest:
		data["description"] = "These are more recent artifacts held by the site."
		data["title"] = "Recent artifacts"
	case -1:
		return data
	default:
		// catch all other matches
		s := strings.TrimSpace(fileslice.RecordsSub(uri))
		data["title"] = helper.Capitalize(s) + " artifacts"
		desc := "The collection of " +
			strconv.Itoa(sum) + " " + s + " artifacts"
		if years != "" {
			desc += " from " + years
		}
		data["description"] = desc + "."
	}
	return data
}

// Artifacts404 renders the files error page for the Artifacts menu and categories.
// It provides different error messages to the standard error page.
func Artifacts404(c echo.Context, sl *slog.Logger, uri string) error {
	const msg = "artifacts 404 context"
	if err := panics.EchoContextS(c, sl); err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}
	const name = "status"
	errs := fmt.Sprint("artifact page not found for,", uri)
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
		return InternalErr(c, sl, errs, err)
	}
	return nil
}

// Apps is the handler for the modern applications and tools page.
func Apps(c echo.Context, sl *slog.Logger) error {
	const msg = "apps context"
	if err := panics.EchoContextS(c, sl); err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}
	const name = "apps"
	const h1 = "Applications"
	const lead = "Here are some software suggestions and Windows, Linux, macOS tools for running out-of-date programs " +
		"and using legacy media formats."
	data := empty(c)
	data["title"] = "Use current apps"
	data["description"] = "Software and application suggestions for using the historic artifacts and " +
		"file downloads on modern systems."
	data["logo"] = "Software suggestions"
	data["h1"] = h1
	data["lead"] = lead
	err := c.Render(http.StatusOK, name, data)
	if err != nil {
		return InternalErr(c, sl, name, err)
	}
	return nil
}

// Areacodes is the handler for the BBS and telephone area codes page.
func Areacodes(c echo.Context, sl *slog.Logger) error {
	const msg = "areacodes context"
	if err := panics.EchoContextS(c, sl); err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}
	data := empty(c)
	data["title"] = "BBS and telephone area codes"
	data["description"] = "Lookup and list the North American Numbering Plan (NANP) area codes in common use until 1994."
	data["logo"] = "BBS area codes"
	data["h1"] = "BBS area codes"
	data["lead"] = "North American Numbering Plan (+1-XXX) telephone area codes until 1994."
	data["telephonecodes"] = areacode.AreaCodes()
	data["territories"] = areacode.Territories()
	data["abbreviations"] = areacode.Abbreviations()
	err := c.Render(http.StatusOK, "areacodes", data)
	if err != nil {
		return InternalErr(c, sl, "areacodes", err)
	}
	return nil
}

// Artist is the handler for the Artist sceners page.
func Artist(c echo.Context, db *sql.DB, sl *slog.Logger) error {
	const msg = "artist context"
	if err := panics.EchoContextDS(c, db, sl); err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}
	data := empty(c)
	title := "Pixel artists and graphic designers"
	data["title"] = title
	data["logo"] = title
	data["h1"] = title
	data["description"] = demo
	data["noindex"] = true
	return scener(c, db, sl, postgres.Artist, data)
}

// scener is the handler for the scener pages.
func scener(c echo.Context, db *sql.DB, sl *slog.Logger, r postgres.Role,
	data map[string]any,
) error {
	const name = "scener"
	s := model.Sceners{}
	ctx := context.Background()
	var err error
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
		return DatabaseErr(c, sl, name, err)
	}
	data["sceners"] = s.Sort()
	data["description"] = "This is a massive but incomplete list of aliases and pseudonyms used " +
		"in the Scene and offers links to individual profiles."
	data["lead"] = "This page shows the sceners and people credited for their work in The Scene." +
		`<br><small class="fw-lighter">` +
		"The list will never be complete or accurate due to the amount of data and the lack of a" +
		" common format for crediting people. " +
		" Sceners often used different names or spellings on their work, including character" +
		" swaps, aliases, initials, and even single-letter signatures." +
		"</small>"
	err = c.Render(http.StatusOK, name, data)
	if err != nil {
		return InternalErr(c, sl, name, err)
	}
	return nil
}

// BBS is the handler for the BBS page ordered by the most files.
func BBS(c echo.Context, db *sql.DB, sl *slog.Logger) error {
	return bbsHandler(c, db, sl, model.Prolific)
}

// BBSAZ is the handler for the BBS page ordered alphabetically.
func BBSAZ(c echo.Context, db *sql.DB, sl *slog.Logger) error {
	return bbsHandler(c, db, sl, model.Alphabetical)
}

// BBSYear is the handler for the BBS page ordered by the year.
func BBSYear(c echo.Context, db *sql.DB, sl *slog.Logger) error {
	return bbsHandler(c, db, sl, model.Oldest)
}

// bbsHandler is the handler for the BBS page.
func bbsHandler(c echo.Context, db *sql.DB, sl *slog.Logger, orderBy model.OrderBy) error {
	const msg = "bbs handler context"
	if err := panics.EchoContextDS(c, db, sl); err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}
	const title, name = "BBS", "bbs"
	// FTP sites were Internet-based file exchange servers that would host and share Scene releases.
	const lead = "Bulletin Board Systems were personal computers networked using the copper telephone network " +
		"and provided communication services, file hosting and exchanges."
	const logo = "Bulletin Board Systems"
	const key = "releasers"
	data := empty(c)
	data["noindex"] = true
	data["title"] = "Former " + title
	data["description"] = lead
	data["logo"] = logo
	data["h1"] = title
	data["lead"] = lead
	data["itemName"] = name
	data[key] = model.Releasers{}
	data["stats"] = map[string]string{}

	ctx := context.Background()
	r := model.Releasers{}
	if err := r.BBS(ctx, db, orderBy); err != nil {
		return DatabaseErr(c, sl, name, err)
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
		data["noindex"] = false
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
	err := c.Render(http.StatusOK, tmpl, data)
	if err != nil {
		return InternalErr(c, sl, name, err)
	}
	return nil
}

// BrokenTexts is the handler for the Broken texts page.
func BrokenTexts(c echo.Context, sl *slog.Logger) error {
	const msg = "broken texts context"
	if err := panics.EchoContextS(c, sl); err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}
	const name = "brokentexts"
	const h1 = "Broken text files"
	const lead = "Unfortunately, there are large numbers of incomplete, inaccurate," +
		" or corrupted information texts (NFOs). " +
		"While we'd prefer to offer a pristine copy of a Scene text, " +
		"hosting a broken text is more useful than offering nothing."
	data := empty(c)
	data["title"] = "Broken texts!?"
	data["description"] = "Learn why there are broken encodings, unreadable texts and corrupted nfo files."
	data["logo"] = "Broken text files and NFOs"
	data["h1"] = h1
	data["lead"] = lead
	err := c.Render(http.StatusOK, name, data)
	if err != nil {
		return InternalErr(c, sl, name, err)
	}
	return nil
}

// Checksum is the handler for the Checksum file record page.
func Checksum(c echo.Context, db *sql.DB, sl *slog.Logger, id string) error {
	const msg = "checksum context"
	if err := panics.EchoContextDS(c, db, sl); err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}
	const uri = "sum"
	if err := download.Checksum(c, db, id); err != nil {
		if errors.Is(err, download.ErrStat) {
			return FileMissingErr(c, sl, uri, err)
		}
		return DownloadErr(c, sl, uri, err)
	}
	return nil
}

// Coder is the handler for the Coder sceners page.
func Coder(c echo.Context, db *sql.DB, sl *slog.Logger) error {
	const msg = "coder context"
	if err := panics.EchoContextDS(c, db, sl); err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}
	data := empty(c)
	title := "Coder and programmers"
	data["title"] = title
	data["logo"] = title
	data["h1"] = title
	data["description"] = demo
	data["noindex"] = true
	return scener(c, db, sl, postgres.Writer, data)
}

// Compression is the handler for information on historic compression tools page.
func Compression(c echo.Context, sl *slog.Logger) error {
	const msg = "compression context"
	if err := panics.EchoContextS(c, sl); err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}
	const name = "compression"
	const h1 = "Archiving"
	const lead = "Compression and archiving formats of the 1980s were evolving by the month, and today are hard to parse."
	data := empty(c)
	data["title"] = "Compression and archiving formats"
	data["description"] = "Old file archives and compression methods used in the 1980s on the PC."
	data["logo"] = "Old file archives"
	data["h1"] = h1
	data["lead"] = lead
	err := c.Render(http.StatusOK, name, data)
	if err != nil {
		return InternalErr(c, sl, name, err)
	}
	return nil
}

// Configurations is the handler for the Configuration page.
func Configurations(c echo.Context, db *sql.DB, sl *slog.Logger, conf config.Config) error {
	const msg = "configurations context"
	if err := panics.EchoContextDS(c, db, sl); err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}
	const name = "configs"
	data := empty(c)
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
	// As we are collecting stats of both the file system and database, we may as well do it cocurrently
	var wg sync.WaitGroup
	var mu sync.Mutex
	wg.Go(func() {
		mu.Lock()
		ca, cp, cnu, err := model.Counts(ctx, db)
		if err == nil {
			data["countArtifacts"] = ca
			data["countPublic"] = cp
			data["countNewUpload"] = cnu
			data["countHidden"] = ca - cp - cnu
		}
		mu.Unlock()
	})
	wg.Go(func() {
		mu.Lock()
		data = configurations(data, conf)
		mu.Unlock()
	})
	wg.Wait()
	if db == nil {
		data["dbConnections"] = "database not set"
		err := c.Render(http.StatusOK, name, data)
		if err != nil {
			return InternalErr(c, sl, name, err)
		}
		return nil
	}
	conns, maxConn, err := postgres.Connections(db)
	if err != nil {
		data["dbConnections"] = err.Error()
	}
	data["dbConnections"] = fmt.Sprintf("%d of %d", conns, maxConn)
	err = c.Render(http.StatusOK, name, data)
	if err != nil {
		return InternalErr(c, sl, name, err)
	}
	return nil
}

// configurations handles the host system drive queries for free and used space.
func configurations(data map[string]any, conf config.Config) map[string]any {
	var wg sync.WaitGroup
	var mu sync.Mutex
	wg.Go(func() {
		mu.Lock()
		data = downloader(data, conf)
		mu.Unlock()
	})
	wg.Go(func() {
		mu.Lock()
		data = previewer(data, conf)
		mu.Unlock()
	})
	wg.Go(func() {
		mu.Lock()
		data = thumbnailer(data, conf)
		mu.Unlock()
	})
	wg.Go(func() {
		mu.Lock()
		data = extraer(data, conf)
		mu.Unlock()
	})
	wg.Go(func() {
		mu.Lock()
		data = orphaneder(data, conf)
		mu.Unlock()
	})
	wg.Go(func() {
		mu.Lock()
		wdu, wdt, _, wdp, _ := helper.DiskStat(conf.AbsDownload.String())
		data["wdUsage"] = helper.ByteCount(int64(wdu))
		data["wdTotal"] = helper.ByteCount(int64(wdt))
		data["wdPercent"] = wdp
		mu.Unlock()
	})
	wg.Wait()
	return data
}

func downloader(data map[string]any, conf config.Config) map[string]any {
	download := dir.Directory(string(conf.AbsDownload))
	check := config.CheckDir(download, "downloads")
	data["checkDownloads"] = check
	data["countDownloads"] = 0
	data["usageDownloads"] = 0
	data["extsDownloads"] = []helper.Extension{}
	if check == nil {
		data["countDownloads"], _ = helper.Count(string(conf.AbsDownload))
		b, _ := helper.DiskUsage(string(conf.AbsDownload))
		data["usageDownloads"] = helper.ByteCount(b)
		exts, _ := helper.CountExts(string(conf.AbsDownload))
		data["extsDownloads"] = exts
	}
	return data
}

func previewer(data map[string]any, conf config.Config) map[string]any {
	preview := dir.Directory(conf.AbsPreview)
	check := config.CheckDir(preview, "previews")
	data["checkPreviews"] = check
	data["countPreviews"] = 0
	data["usagePreviews"] = 0
	data["extsPreviews"] = []helper.Extension{}
	if check == nil {
		data["countPreviews"], _ = helper.Count(conf.AbsPreview.String())
		b, _ := helper.DiskUsage(string(conf.AbsPreview))
		data["usagePreviews"] = helper.ByteCount(b)
		exts, _ := helper.CountExts(conf.AbsPreview.String())
		data["extsPreviews"] = exts
	}
	return data
}

func thumbnailer(data map[string]any, conf config.Config) map[string]any {
	thumbnail := dir.Directory(conf.AbsThumbnail)
	check := config.CheckDir(thumbnail, "thumbnails")
	data["checkThumbnails"] = check
	data["countThumbnails"] = 0
	data["usageThumbnails"] = 0
	data["extsThumbnails"] = []helper.Extension{}
	if check == nil {
		data["countThumbnails"], _ = helper.Count(conf.AbsThumbnail.String())
		b, _ := helper.DiskUsage(string(conf.AbsThumbnail))
		data["usageThumbnails"] = helper.ByteCount(b)
		exts, _ := helper.CountExts(conf.AbsThumbnail.String())
		data["extsThumbnails"] = exts
	}
	return data
}

func extraer(data map[string]any, conf config.Config) map[string]any {
	extra := dir.Directory(conf.AbsExtra)
	check := config.CheckDir(extra, "extra")
	data["checkExtras"] = check
	data["countExtras"] = 0
	data["usageExtras"] = 0
	data["extsExtras"] = []helper.Extension{}
	if check == nil {
		data["countExtras"], _ = helper.Count(conf.AbsExtra.String())
		b, _ := helper.DiskUsage(string(conf.AbsExtra))
		data["usageExtras"] = helper.ByteCount(b)
		exts, _ := helper.CountExts(conf.AbsExtra.String())
		data["extsExtras"] = exts
	}
	return data
}

func orphaneder(data map[string]any, conf config.Config) map[string]any {
	orphaned := dir.Directory(conf.AbsOrphaned.String())
	check := config.CheckDir(orphaned, "orphaned")
	data["checkOrphaned"] = check
	data["countOrphaned"] = 0
	data["usageOrphaned"] = 0
	data["extsOrphaned"] = []helper.Extension{}
	if check == nil {
		data["countOrphaned"], _ = helper.Count(conf.AbsOrphaned.String())
		b, _ := helper.DiskUsage(string(conf.AbsOrphaned))
		data["usageOrphaned"] = helper.ByteCount(b)
		exts, _ := helper.CountExts(conf.AbsOrphaned.String())
		data["extsOrphaned"] = exts
	}
	return data
}

// DownloadJsDos is the handler for the js-dos emulator to download zip files that are then
// mounted as a C: hard drive in the emulation. js-dos only supports common zip compression methods,
// so this func first attempts to offer a re-archived zip file found in the extra directory, and
// only if that fails does it offer the original download file.
func DownloadJsDos(c echo.Context, db *sql.DB, sl *slog.Logger, extra, downl dir.Directory) error {
	const msg = "download jsdos context"
	if err := panics.EchoContextDS(c, db, sl); err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}
	e := download.ExtraZip{
		Extra:    extra,
		Download: downl,
	}
	const uri = "jsdos"
	if err := e.HTTPSend(c, db); err != nil {
		if errors.Is(err, download.ErrStat) {
			return FileMissingErr(c, sl, uri, err)
		}
		return DownloadErr(c, sl, uri, err)
	}
	return nil
}

// Download is the handler for the Download file record page.
func Download(c echo.Context, db *sql.DB, sl *slog.Logger, downl dir.Directory) error {
	const msg = "download context"
	if err := panics.EchoContextDS(c, db, sl); err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}
	d := download.Download{
		Inline: false,
		Dir:    downl,
	}
	const uri = "d"
	if id := c.Param("id"); id != "" {
		r := c.Response()
		r.Header().Set("Link", fmt.Sprintf(`<https://defacto2.net/%s/%s; rel="canonical">`, uri, id))
	}
	if err := d.HTTPSend(c, db, sl); err != nil {
		if errors.Is(err, download.ErrStat) {
			return FileMissingErr(c, sl, uri, err)
		}
		return DownloadErr(c, sl, uri, err)
	}
	return nil
}

// FTP is the handler for the FTP page.
func FTP(c echo.Context, db *sql.DB, sl *slog.Logger) error {
	const msg = "ftp context"
	if err := panics.EchoContextDS(c, db, sl); err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}
	const title, name = "FTP", "ftp"
	ctx := context.Background()
	data := empty(c)
	const lead = "FTP sites were Internet-based file exchange servers that would host and share Scene releases."
	const key = "releasers"
	data["title"] = "Former " + title + " sites"
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
		return DatabaseErr(c, sl, name, err)
	}
	data[key] = r
	data["stats"] = map[string]string{
		"pubs":    fmt.Sprintf("%d sites", len(r)),
		"orderBy": alpha,
	}
	err := c.Render(http.StatusOK, name, data)
	if err != nil {
		return InternalErr(c, sl, name, err)
	}
	return nil
}

// Categories is the handler for the artifact categories page.
func Categories(c echo.Context, db *sql.DB, sl *slog.Logger, stats bool) error {
	const msg = "categories context"
	if err := panics.EchoContextDS(c, db, sl); err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}
	const title, name = "Artifact categories", "categories"
	data := empty(c)
	data["noindex"] = true
	data["title"] = title
	data["description"] = "A table of contents for the collection."
	data["logo"] = title
	data["h1"] = title
	data["lead"] = "This page shows the categories and platforms in the collection of file artifacts."
	data["stats"] = stats
	data["counter"] = fileslice.Statistics()
	data, err := fileWStats(db, data, stats)
	if err != nil {
		sl.Warn("context categories", slog.Any("error", err))
		data["databaseErr"] = true
	}
	err = c.Render(http.StatusOK, name, data)
	if err != nil {
		return InternalErr(c, sl, name, err)
	}
	return nil
}

// fileWStats is a helper function for File that adds the statistics to the data map.
func fileWStats(db *sql.DB, data map[string]any, stats bool) (map[string]any, error) {
	if data == nil {
		data = make(map[string]any) // avoid nil map
	}
	if !stats {
		return data, nil
	}
	c, err := fileslice.Counter(db)
	if err != nil {
		return data, fmt.Errorf("counter: %w", err)
	}
	data["counter"] = c
	data["logo"] = "Artifact category statistics"
	data["lead"] = "This page shows the artifacts categories with selected statistics, " +
		"such as the number of files in the category or platform."
	return data, nil
}

// Deletions is the handler to list the files that have been marked for deletion.
func Deletions(c echo.Context, db *sql.DB, sl *slog.Logger, page string) error {
	const msg = "deletions context"
	if err := panics.EchoContextDS(c, db, sl); err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}
	uri := fileslice.Deletions.String()
	if !fileslice.Valid(uri) {
		return Artifacts404(c, sl, uri)
	}
	if page == "" {
		return artifacts(c, db, sl, uri, 1)
	}
	p, err := strconv.Atoi(page)
	if err != nil {
		return Page404(c, sl, uri, page)
	}
	return artifacts(c, db, sl, uri, p)
}

// Fixes is the handler for the problems and fixes page.
func Fixes(c echo.Context, sl *slog.Logger) error {
	const msg = "apps context"
	if err := panics.EchoContextS(c, sl); err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}
	const name = "fixes"
	const h1 = "Problems and fixes"
	const lead = "Shrinker dispatcher or runtime 200 errors, or missing NPMOD32.DLL or D3DRM.DLL files?"
	data := empty(c)
	data["title"] = "Problems and fixes"
	data["description"] = "Fix common errors found in applications and tools authored by the Scene for Windows and DOS."
	data["logo"] = "Problems and fixes"
	data["h1"] = h1
	data["lead"] = lead
	err := c.Render(http.StatusOK, name, data)
	if err != nil {
		return InternalErr(c, sl, name, err)
	}
	return nil
}

// Terms is the handler for the problems and fixes page.
func Terms(c echo.Context, sl *slog.Logger) error {
	const msg = "apps context"
	if err := panics.EchoContextS(c, sl); err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}
	const name = "terms"
	const h1 = "Glossary of common terms"
	const lead = "A glossary of the unique and common terms used in the scene."
	data := empty(c)
	data["title"] = "Common terms"
	data["description"] = "This is a list of the unique and common terms used in the scene."
	data["logo"] = "Glossary of terms"
	data["h1"] = h1
	data["lead"] = lead
	err := c.Render(http.StatusOK, name, data)
	if err != nil {
		return InternalErr(c, sl, name, err)
	}
	return nil
}

// Unwanted is the handler to list the files that have been marked as unwanted.
func Unwanted(c echo.Context, db *sql.DB, sl *slog.Logger, page string) error {
	const msg = "unwanted context"
	if err := panics.EchoContextDS(c, db, sl); err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}
	uri := fileslice.Unwanted.String()
	if !fileslice.Valid(uri) {
		return Artifacts404(c, sl, uri)
	}
	if page == "" {
		return artifacts(c, db, sl, uri, 1)
	}
	p, err := strconv.Atoi(page)
	if err != nil {
		return Page404(c, sl, uri, page)
	}
	return artifacts(c, db, sl, uri, p)
}

// ForApproval is the handler for the list and preview of the files page.
// The uri is the category or collection of files to display.
// The page is the page number of the results to display.
func ForApproval(c echo.Context, db *sql.DB, sl *slog.Logger, page string) error {
	const msg = "for approval context"
	if err := panics.EchoContextDS(c, db, sl); err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}
	uri := fileslice.ForApproval.String()
	if !fileslice.Valid(uri) {
		return Artifacts404(c, sl, uri)
	}
	if page == "" {
		return artifacts(c, db, sl, uri, 1)
	}
	p, err := strconv.Atoi(page)
	if err != nil {
		return Page404(c, sl, uri, page)
	}
	return artifacts(c, db, sl, uri, p)
}

// GetDemozooParam fetches the multiple download_links values from the
// Demozoo production API and attempts to download and save one of the
// linked files. If multiple links are found, the first link is used as
// they should all point to the same asset.
//
// Both the Demozoo production ID param and the Defacto2 UUID query
// param values are required as params to fetch the production data and
// to save the file to the correct filename.
func GetDemozooParam(c echo.Context, db *sql.DB, download dir.Directory) error {
	const msg = "get demozoo param context"
	if err := panics.EchoContextD(c, db); err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}
	got := remote.DemozooLink{}
	sid := c.Param("id")
	id, err := strconv.Atoi(sid)
	if err != nil {
		got.Error = "demozoo id must be a numeric value, " + sid
		return c.JSON(http.StatusBadRequest, got)
	}
	got.ID = id
	unid := c.QueryParam("unid")
	if err = uuid.Validate(unid); err != nil {
		got.Error = "uuid syntax did not validate, " + unid
		return c.JSON(http.StatusBadRequest, got)
	}
	got.UUID = unid
	return got.Download(c, db, download)
}

// GetDemozoo fetches the download link from Demozoo and saves it to the download directory.
// It then runs Update to modify the database record with various metadata from the file and Demozoo record API data.
//
// This function is a wrapper for the remote.DemozooLink.Download method.
func GetDemozoo(c echo.Context, db *sql.DB, demozooID int, defacto2UNID string, download dir.Directory) error {
	got := remote.DemozooLink{
		ID:   demozooID,
		UUID: defacto2UNID,
	}
	return got.Download(c, db, download)
}

// GetPouet fetches the download link from Pouet and saves it to the download directory.
// It then runs Update to modify the database record with various metadata from the file and Pouet record API data.
//
// This function is a wrapper for the remote.PouetLink.Download method.
func GetPouet(c echo.Context, db *sql.DB, pouetID int, defacto2UNID string, download dir.Directory) error {
	got := remote.PouetLink{
		ID:   pouetID,
		UUID: defacto2UNID,
	}
	return got.Download(c, db, download)
}

// GoogleCallback is the handler for the Google OAuth2 callback page to verify
// the [Google ID token].
//
// [Google ID token]: https://developers.google.com/identity/gsi/web/guides/verify-google-id-token
func GoogleCallback(c echo.Context, sl *slog.Logger, clientID string, maxAge int, accounts ...[48]byte) error {
	const msg = "google callback context"
	if err := panics.EchoContextS(c, sl); err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}
	const name = "google/callback"
	// Cross-Site Request Forgery cookie token
	const csrf = "g_csrf_token"
	cookie, err := c.Cookie(csrf)
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			return c.Redirect(http.StatusForbidden, "/signin")
		}
		return BadRequestErr(c, sl, name, err)
	}
	token := cookie.Value

	// Cross-Site Request Forgery post token
	bodyToken := c.FormValue(csrf)
	if token != bodyToken {
		return BadRequestErr(c, sl, name, ErrMisMatch)
	}

	// Create a new token verifier.
	// https://pkg.go.dev/google.golang.org/api/idtoken
	ctx := context.Background()
	validator, err := idtoken.NewValidator(ctx)
	if err != nil {
		return BadRequestErr(c, sl, name, err)
	}

	// Verify the ID token and using the client ID from the Google API.
	credential := c.FormValue("credential")
	playload, err := validator.Validate(ctx, credential, clientID)
	if err != nil {
		return BadRequestErr(c, sl, name, err)
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
		sub := playload.Claims["sub"]
		return ForbiddenErr(c, sl, name,
			fmt.Errorf("%w. If this is a mistake, contact Defacto2 admin and give them this Google account ID: %s",
				ErrUser, sub))
	}

	if err = sessionHandler(c, maxAge, playload.Claims); err != nil {
		return BadRequestErr(c, sl, name, err)
	}
	return c.Redirect(http.StatusFound, "/")
}

// sessionHandler creates a [new session] and populates it with
// the claims data created by the [ID Tokens for Google HTTP APIs].
//
// [new session]: https://pkg.go.dev/github.com/gorilla/sessions
// [ID Tokens for Google HTTP APIs]: https://pkg.go.dev/google.golang.org/api/idtoken
func sessionHandler(c echo.Context, maxAge int, claims map[string]any,
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

// History is the handler for the History page.
func History(c echo.Context, sl *slog.Logger) error {
	const msg = "history context"
	if err := panics.EchoContextS(c, sl); err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}
	const name = "history"
	const lead = "In the past, alternative iterations of the name have included" +
		" De Facto, DF, DeFacto, Defacto II, Defacto 2, and the defacto2.com domain."
	const h1 = "The history of the brand"
	data := empty(c)
	data["carousel"] = "#carouselDf2Artpacks"
	data["description"] = "Learn about the many iterations of Defacto2 and the original DeFacto magazine from 1996."
	data["logo"] = "The history of Defacto"
	data["h1"] = h1
	data["lead"] = lead
	data["title"] = h1
	err := c.Render(http.StatusOK, name, data)
	if err != nil {
		return InternalErr(c, sl, name, err)
	}
	return nil
}

// Index is the handler for the Home page.
func Index(c echo.Context, sl *slog.Logger) error {
	const msg = "index context"
	if err := panics.EchoContextS(c, sl); err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}
	const name = "index"
	data := empty(c)
	data["title"] = "Introduction and milestones"
	data["canonical"] = "/"
	data["h1"] = "The subcultures of obsolete microcomputers"
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
		return InternalErr(c, sl, name, err)
	}
	return nil
}

// Inline is the handler for the Download file record page.
func Inline(c echo.Context, db *sql.DB, sl *slog.Logger, downl dir.Directory) error {
	const msg = "inline context"
	if err := panics.EchoContextDS(c, db, sl); err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}
	d := download.Download{
		Inline: true,
		Dir:    downl,
	}
	const uri = "v"
	if err := d.HTTPSend(c, db, sl); err != nil {
		if errors.Is(err, download.ErrStat) {
			return FileMissingErr(c, sl, uri, err)
		}
		return DownloadErr(c, sl, uri, err)
	}
	return nil
}

// Interview is the handler for the People Interviews page.
func Interview(c echo.Context, sl *slog.Logger) error {
	const msg = "interview context"
	if err := panics.EchoContextS(c, sl); err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}
	const title, name = "Interviews with former Sceners", "interview"
	data := empty(c)
	data["title"] = title
	data["description"] = "A collection of historical interviews and discussions with former members of the Scene."
	data["logo"] = title
	data["h1"] = title
	data["lead"] = "Here is a centralized page for the discussions and unedited" +
		" interviews with former sceners, crackers, and demo makers."
	data["interviews"] = Interviewees()
	err := c.Render(http.StatusOK, name, data)
	if err != nil {
		return InternalErr(c, sl, name, err)
	}
	return nil
}

// Magazine is the handler for the Magazine page.
func Magazine(c echo.Context, db *sql.DB, sl *slog.Logger) error {
	return magazines(c, db, sl, true)
}

// MagazineAZ is the handler for the Magazine page ordered chronologically.
func MagazineAZ(c echo.Context, db *sql.DB, sl *slog.Logger) error {
	return magazines(c, db, sl, false)
}

// magazines is the handler for the magazine page.
func magazines(c echo.Context, db *sql.DB, sl *slog.Logger, chronological bool) error {
	const msg = "magazines context"
	if err := panics.EchoContextDS(c, db, sl); err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}
	const title, name = "Magazines", "magazine"
	data := empty(c)
	const lead = "Scene magazines are the newsletters, reports, " +
		"and publications on the activities of the subculture community."
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
	var order string
	r := model.Releasers{}
	switch chronological {
	case true:
		if err := r.Magazine(ctx, db); err != nil {
			return DatabaseErr(c, sl, name, err)
		}
		s := title + byyear
		data["logo"] = s
		data["title"] = title + byyear
		order = year
	case false:
		if err := r.MagazineAZ(ctx, db); err != nil {
			return DatabaseErr(c, sl, name, err)
		}
		data["noindex"] = true
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
	err := c.Render(http.StatusOK, name, data)
	if err != nil {
		return InternalErr(c, sl, name, err)
	}
	return nil
}

// Musician is the handler for the Musiciansceners page.
func Musician(c echo.Context, db *sql.DB, sl *slog.Logger) error {
	const msg = "musician context"
	if err := panics.EchoContextDS(c, db, sl); err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}
	data := empty(c)
	title := "Musicians and composers"
	data["title"] = title
	data["logo"] = title
	data["h1"] = title
	data["description"] = demo
	data["noindex"] = true
	return scener(c, db, sl, postgres.Musician, data)
}

// New is the handler for the what is new page.
func New(c echo.Context, sl *slog.Logger) error {
	const name = "new"
	data := empty(c)
	data["noindex"] = true // apply noindex to what's new, so we don't have to worry using about <a href rel="nofollow">
	data["description"] = "What is new on the Defacto2 website?"
	data["logo"] = "New stuff"
	data["h1"] = "What is new?"
	data["lead"] = "This quaint page does not appeal to algorithms, so who will see it?"
	data["title"] = "New stuff"
	data["carousel"] = "#carouselWhatsNew"
	err := c.Render(http.StatusOK, name, data)
	if err != nil {
		return InternalErr(c, sl, name, err)
	}
	return nil
}

// Page404 renders the files page error page for the Artifacts menu and categories.
// It provides different error messages to the standard error page.
func Page404(c echo.Context, sl *slog.Logger, uri, page string) error {
	const msg = "page 404 context"
	if err := panics.EchoContextS(c, sl); err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}
	const name = "status"
	errs := fmt.Sprintf("page not found for %q at %q", page, uri)
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
		return InternalErr(c, sl, errs, err)
	}
	return nil
}

// PlatformEdit handles the post submission for the Platform selection field.
func PlatformEdit(c echo.Context, db *sql.DB, sl *slog.Logger) error {
	const msg = "platform edit context"
	if err := panics.EchoContextDS(c, db, sl); err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}
	var f Form
	if err := c.Bind(&f); err != nil {
		return badRequest(c, err)
	}
	ctx := context.Background()
	r, err := model.One(ctx, db, true, f.ID)
	if err != nil {
		return fmt.Errorf("platform edit %w: %d", err, f.ID)
	}
	if err = model.UpdatePlatform(db, int64(f.ID), f.Value); err != nil {
		return badRequest(c, err)
	}
	return c.JSON(http.StatusOK, r)
}

// PlatformTagInfo handles the POST submission for the platform and tag info.
func PlatformTagInfo(c echo.Context) error {
	const msg = "platform tag info"
	if c == nil {
		return fmt.Errorf("%s: %w", msg, panics.ErrNoEchoC)
	}
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

// PostDesc is the handler for the Search for file descriptions form post page.
func PostDesc(c echo.Context, db *sql.DB, sl *slog.Logger, input string) error {
	const msg = "post desc context"
	if err := panics.EchoContextDS(c, db, sl); err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}
	const name = "artifacts"
	errs := fmt.Sprint("post desc search for,", input)
	ctx := context.Background()
	terms := helper.SearchTerm(input)
	rel := model.Artifacts{}
	fs, _ := rel.Description(ctx, db, sl, terms)
	d := Descriptions.postStats(ctx, db, terms)
	s := strings.Join(terms, ", ")
	data := emptyFiles(c)
	const brief = "Game and program title"
	data["title"] = brief + " results"
	data["h1"] = brief + " search"
	data["lead"] = "Results for " + s
	data["logo"] = s + " results"
	data["description"] = brief + " search results for " + s + "."
	data["unknownYears"] = false
	data[records] = fs
	data["stats"] = d
	err := c.Render(http.StatusOK, name, data)
	if err != nil {
		return InternalErr(c, sl, errs, err)
	}
	return nil
}

// PostFilename is the handler for the Search for filenames form post page.
func PostFilename(c echo.Context, db *sql.DB, sl *slog.Logger) error {
	return PostName(c, db, sl, Filenames)
}

// PostName is the handler for the Search for filenames form post page.
func PostName(c echo.Context, db *sql.DB, sl *slog.Logger, mode FileSearch) error {
	const msg = "post name context"
	if err := panics.EchoContextDS(c, db, sl); err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}
	const name = "artifacts"
	errs := fmt.Sprint("post name search for,", mode)
	ctx := context.Background()
	input := c.FormValue("search-term-query")
	terms := helper.SearchTerm(input)
	rel := model.Artifacts{}
	fs, _ := rel.Filename(ctx, db, terms)
	d := mode.postStats(ctx, db, terms)
	s := strings.Join(terms, ", ")
	data := emptyFiles(c)
	data["title"] = "Filename results"
	data["h1"] = "Filename search"
	data["lead"] = "Results for " + s
	data["logo"] = s + " results"
	data["description"] = "Filename search results for " + s + "."
	data["unknownYears"] = false
	data[records] = fs
	data["stats"] = d
	err := c.Render(http.StatusOK, name, data)
	if err != nil {
		return InternalErr(c, sl, errs, err)
	}
	return nil
}

// postStats is a helper function for PostName that returns the statistics for the files page.
func (mode FileSearch) postStats(ctx context.Context, db *sql.DB, terms []string) map[string]string {
	none := func() map[string]string {
		return map[string]string{
			"files": "no files found",
			"years": "",
		}
	}
	m := model.Summary{}
	switch mode {
	case Filenames:
		if err := m.ByFilename(ctx, db, terms); err != nil {
			return none()
		}
	case Descriptions:
		if err := m.ByDescription(ctx, db, terms); err != nil {
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

// PouetCache parses the cached data for the Pouet production votes.
// If the cache is valid it is returned as JSON response.
// If the cache is invalid or corrupt an error will be returned
// and a API request should be made to Pouet.
func PouetCache(c echo.Context, data string) error {
	const msg = "pouet cache context"
	if c == nil {
		return fmt.Errorf("%s: %w", msg, panics.ErrNoEchoC)
	}
	if data == "" {
		return nil
	}
	pv := pouet.Votes{}
	x := strings.Split(data, sep)
	const req = 4
	if len(x) < req {
		return fmt.Errorf("%s %w: %d, want %d", msg, ErrCorrupt, len(x), req)
	}
	stars, err := strconv.ParseFloat(x[0], 64)
	if err != nil {
		return fmt.Errorf("%s %w: %s", msg, err, x[0])
	}
	vd, err := strconv.Atoi(x[1])
	if err != nil {
		return fmt.Errorf("%s %w: %s", msg, err, x[1])
	}
	vu, err := strconv.Atoi(x[2])
	if err != nil {
		return fmt.Errorf("%s %w: %s", msg, err, x[2])
	}
	vm, err := strconv.Atoi(x[3])
	if err != nil {
		return fmt.Errorf("%s %w: %s", msg, err, x[3])
	}
	pv.Stars = stars
	pv.VotesDown = uint64(math.Abs(float64(vd)))
	pv.VotesUp = uint64(math.Abs(float64(vu)))
	pv.VotesMeh = uint64(math.Abs(float64(vm)))
	if err = c.JSON(http.StatusOK, pv); err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return nil
}

// ProdPouet is the handler for the Pouet prod JSON page.
func ProdPouet(c echo.Context, id string) error {
	const msg = "prod pouet context"
	if c == nil {
		return fmt.Errorf("%s: %w", msg, panics.ErrNoEchoC)
	}
	p := pouet.Production{}
	i, err := strconv.Atoi(id)
	if err != nil {
		return c.String(http.StatusNotFound, err.Error())
	}
	if _, err = p.Get(i); err != nil {
		return c.String(http.StatusNotFound, err.Error())
	}
	if err = c.JSON(http.StatusOK, p); err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return nil
}

// ProdZoo is the handler for the Demozoo production JSON page.
func ProdZoo(c echo.Context, id string) error {
	const msg = "prod zoo context"
	if c == nil {
		return fmt.Errorf("%s: %w", msg, panics.ErrNoEchoC)
	}
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

// Releaser is the handler for the releaser page ordered by the most files.
func Releaser(c echo.Context, db *sql.DB, sl *slog.Logger) error {
	return releasers(c, db, sl, model.Prolific)
}

// ReleaserAZ is the handler for the releaser page ordered alphabetically.
func ReleaserAZ(c echo.Context, db *sql.DB, sl *slog.Logger) error {
	return releasers(c, db, sl, model.Alphabetical)
}

// ReleaserYear is the handler for the releaser page ordered by year of the first release.
func ReleaserYear(c echo.Context, db *sql.DB, sl *slog.Logger) error {
	return releasers(c, db, sl, model.Oldest)
}

// releasers is the handler for the Releaser page.
func releasers(c echo.Context, db *sql.DB, sl *slog.Logger, orderBy model.OrderBy) error {
	const msg = "releaser context"
	if err := panics.EchoContextDS(c, db, sl); err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}
	const title, name = "Releaser", "releaser"
	data := empty(c)
	const lead = "A releaser is a brand or a collective group of " +
		"sceners responsible for releasing or distributing products."
	const logo = "Former groups and releasers"
	const key = "releasers"
	data["noindex"] = true
	data["title"] = title + "s and groups"
	data["description"] = "A linked list of the former Scene releasers and groups, that were collectives of people " +
		"who would work together and operate under a common brand."
	data["logo"] = logo
	data["h1"] = title
	data["lead"] = lead
	data["itemName"] = "file"
	data[key] = model.Releasers{}
	data["stats"] = map[string]string{}

	ctx := context.Background()
	var r model.Releasers
	if err := r.Limit(ctx, db, orderBy, 0, 0); err != nil {
		return DatabaseErr(c, sl, name, err)
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
		data["noindex"] = false
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
	err := c.Render(http.StatusOK, tmpl, data)
	if err != nil {
		return InternalErr(c, sl, tmpl, err)
	}
	return nil
}

// Releaser404 renders the files error page for the Groups menu and invalid releasers.
func Releaser404(c echo.Context, sl *slog.Logger, invalidID string) error {
	const msg = "releaser 404 context"
	if err := panics.EchoContextS(c, sl); err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}
	const name = "status"
	errs := fmt.Sprint("releaser page not found for,", invalidID)
	data := empty(c)
	data["title"] = fmt.Sprintf("%d error, releaser page not found", http.StatusNotFound)
	data["description"] = fmt.Sprintf("HTTP status %d error", http.StatusNotFound)
	data["code"] = http.StatusNotFound
	data["logo"] = "Releaser not found"
	data["alert"] = fmt.Sprintf("Releaser %q cannot be found", invalidID)
	data["probl"] = "The releaser page does not exist, there is probably a typo with the URL."
	data["uriOkay"] = "g/"
	data["uriErr"] = invalidID
	err := c.Render(http.StatusNotFound, name, data)
	if err != nil {
		return InternalErr(c, sl, errs, err)
	}
	return nil
}

// Releasers is the handler for the list and preview of files credited to a releaser.
func Releasers(c echo.Context, db *sql.DB, sl *slog.Logger, uri string, public embed.FS) error {
	const msg = "releasers context handler"
	if err := panics.EchoContextDS(c, db, sl); err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}
	const name = "artifacts"
	errs := fmt.Sprint("releasers page for, ", uri)
	ctx := context.Background()
	relname := releaser.Link(uri)
	rel := model.Releasers{}
	fs, err := rel.Where(ctx, db, uri)
	if err != nil {
		sl.Error(msg, slog.String("database", "releasers lookup problem"),
			slog.String("uri", uri), slog.Any("error", err))
		return Releaser404(c, sl, uri)
	}
	if len(fs) == 0 {
		return Releaser404(c, sl, uri)
	}
	data := emptyFiles(c)
	data["title"] = relname + " artifacts"
	data["canonical"] = strings.Join([]string{"g", uri}, "/")
	data["h1"] = relname
	altnames := initialism.Join(initialism.Path(uri))
	data["lead"] = altnames
	data["logo"] = relname
	data["demozoo"] = strconv.Itoa(int(demozoo.Find(uri)))
	data["sixteen"] = sixteen.Find(uri)
	data["csdb"] = csdb.Find(uri)
	data["janeway"] = janeway.Find(uri)
	data["website"] = site.Find(uri)
	tidbits := tidbit.Find(uri)
	slices.Sort(tidbits)
	htm := tibits(sl, uri, public)
	data["tidbits"] = template.HTML(htm)
	if strings.HasSuffix(uri, "-bbs") {
		data["bbs"] = true
		relname = "the " + relname
	}
	if strings.HasSuffix(uri, "-ftp") {
		relname = "the " + relname
	}
	data["uploader-releaser-index"] = releaser.Index(uri)
	data[records] = fs
	data = releaserLead(uri, data)
	d, err := releaserSum(ctx, db, uri)
	if err != nil {
		sl.Error(msg, slog.String("database", "releasers statistics problem"),
			slog.String("uri", uri), slog.Any("error", err))
		return Releaser404(c, sl, uri)
	}
	data = releasersDesc(relname, altnames, d, data)
	err = c.Render(http.StatusOK, name, data)
	if err != nil {
		return InternalErr(c, sl, errs, err)
	}
	return nil
}

// releasersDesc appends stats to the description.
func releasersDesc(relname, altnames string,
	d map[string]string, data map[string]any,
) map[string]any {
	data["stats"] = d
	// append stats to the description
	dfiles := d["sum"]
	dyears := d["years"]
	var desc string
	switch {
	case strings.EqualFold(relname, "independent"):
		desc = "The collection of " + dfiles + " unaffiliated artifacts"
	case strings.EqualFold(relname, "none"):
		desc = "The collection of the " + dfiles + " artifacts that were not authored by the Scene,"
		data["title"] = "Non-Scene artifacts"
		data["h1"] = "Non-Scene artifacts"
	case dfiles == "1":
		desc = "The single artifact for " + relname
	case dfiles != "":
		desc = "The collection of " + dfiles + " artifacts for " + relname
	default:
		desc = "The collection of artifacts for " + relname
	}
	if altnames != "" {
		desc += " (" + altnames + ")"
	}
	if dyears != "" {
		desc += " from " + dyears
	}
	data["description"] = desc + "."
	return data
}

func tibits(sl *slog.Logger, uri string, public embed.FS) string {
	var htm strings.Builder
	tidbits := tidbit.Find(uri)
	slices.Sort(tidbits)
	for value := range slices.Values(tidbits) {
		s := value.String(sl, public)
		if strings.HasSuffix(strings.TrimSpace(s), "</p>") {
			fmt.Fprintf(&htm, `<li class="list-group-item">%s%s</li>`, s, value.URL(uri))
			continue
		}
		fmt.Fprintf(&htm, `<li class="list-group-item">%s<br>%s</li>`, s, value.URL(uri))
	}
	return htm.String()
}

func releaserLead(uri string, data map[string]any) map[string]any {
	switch uri {
	case "independent":
		data["lead"] = initialism.Join(initialism.Path(uri)) +
			", independent releases are files with no group or releaser affiliation." +
			`<br><small class="fw-lighter">In the scene's early years,` +
			` releasing documents or software cracks under a personal alias or a` +
			` real-name attribution was commonplace.</small>`
	case "none":
		data["lead"] = "None, are files which were never intended for the scene." +
			`<br><small class="fw-lighter">These can include commercial or free software` +
			` applications, articles for the general public, and are often credited to a real name author.</small>`
	default:
		// placeholder to handle other releaser types
	}
	return data
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
		"sum":   strconv.Itoa(int(m.SumCount.Int64)),
	}
	return d, nil
}

// Scener is the handler for the page to list all the sceners.
func Scener(c echo.Context, db *sql.DB, sl *slog.Logger) error {
	const msg = "scener context"
	if err := panics.EchoContextDS(c, db, sl); err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}
	data := empty(c)
	title := "Sceners, people who were apart of the Scene"
	data["title"] = title
	data["logo"] = title
	data["h1"] = title
	data["description"] = demo
	return scener(c, db, sl, postgres.Roles(), data)
}

// Scener404 renders the files error page for the People menu and invalid sceners.
func Scener404(c echo.Context, sl *slog.Logger, id string) error {
	const msg = "scene 404 context"
	if err := panics.EchoContextS(c, sl); err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}
	const name = "status"
	errs := fmt.Sprint("scener page not found for,", id)
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
		return InternalErr(c, sl, errs, err)
	}
	return nil
}

// Sceners is the handler for the list and preview of files credited to a scener.
func Sceners(c echo.Context, db *sql.DB, sl *slog.Logger, uri string) error {
	const msg = "sceners context"
	if err := panics.EchoContextDS(c, db, sl); err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}
	const name = "artifacts"
	errs := fmt.Sprint("sceners page for,", uri)
	ctx := context.Background()
	s := releaser.Link(uri)
	var ms model.Scener
	fs, err := ms.Where(ctx, db, uri)
	if err != nil {
		return InternalErr(c, sl, errs, err)
	}
	if len(fs) == 0 {
		return Scener404(c, sl, uri)
	}
	data := emptyFiles(c)
	data["canonical"] = strings.Join([]string{"p", uri}, "/")
	data["title"] = s + attr
	data["h1"] = s
	data["lead"] = "Artifacts attributed to " + s + "."
	data["logo"] = s
	data["description"] = "These are the documented artifacts attributed to the person known as " + s + "."
	data["scener"] = s
	data[records] = fs
	d, err := scenerSum(ctx, db, uri)
	if err != nil {
		return InternalErr(c, sl, errs, err)
	}
	data["stats"] = d
	err = c.Render(http.StatusOK, name, data)
	if err != nil {
		return InternalErr(c, sl, errs, err)
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

// SearchDesc is the handler for the Search for file descriptions page.
func SearchDesc(c echo.Context, sl *slog.Logger) error {
	const msg = "search desc context"
	if err := panics.EchoContextS(c, sl); err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}
	const title, name = "Game and program title search", "searchpost"
	data := empty(c)
	data["noindex"] = true
	data["description"] = "Use this search to uncover named applications, games, and descriptions of artifacts."
	data["logo"] = title
	data["title"] = title
	data["info"] = "search the names, titles, and comments of artifacts"
	err := c.Render(http.StatusOK, name, data)
	if err != nil {
		return InternalErr(c, sl, name, err)
	}
	return nil
}

// SearchID is the handler for the Record by ID Search page.
func SearchID(c echo.Context, sl *slog.Logger) error {
	const msg = "search id context"
	if err := panics.EchoContextS(c, sl); err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}
	const title, name = "Search the IDs of artifacts", "searchhtmx"
	data := empty(c)
	data["noindex"] = true
	data["description"] = "Use this search to lookup artifacts by their database identities."
	data["logo"] = title
	data["title"] = title
	data["info"] = "search for artifacts by their record id, uuid or URL key"
	data["hxPost"] = "/editor/search/id"
	data["inputPlaceholder"] = "Type to search for an artifact…"
	err := c.Render(http.StatusOK, name, data)
	if err != nil {
		return InternalErr(c, sl, name, err)
	}
	return nil
}

// SearchFile is the handler for the Search for files page.
func SearchFile(c echo.Context, sl *slog.Logger) error {
	const msg = "search file context"
	if err := panics.EchoContextS(c, sl); err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}
	const title, name = "Filename search", "searchpost"
	data := empty(c)
	data["noindex"] = true
	data["description"] = "Use this search to lookup artifacts by their filenames."
	data["logo"] = title
	data["title"] = title
	data["info"] = "search for filenames and filename extensions"
	err := c.Render(http.StatusOK, name, data)
	if err != nil {
		return InternalErr(c, sl, name, err)
	}
	return nil
}

// SearchReleaser is the handler for the Releaser Search page.
func SearchReleaser(c echo.Context, sl *slog.Logger) error {
	const msg = "search releaser context"
	if err := panics.EchoContextS(c, sl); err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}
	const title, name = "Search for releasers", "searchhtmx"
	data := empty(c)
	data["noindex"] = true
	data["description"] = "Use this search to lookup releasers, groups, magazines, boards and sites by their names."
	data["logo"] = title
	data["title"] = title
	data["info"] = "search for a group, initialism, magazine, board, or site"
	data["helpText"] = "searching for 4 or fewer characters triggers an initialism lookup, " +
		"if no results are found, a new search is made for the releaser names"
	data["hxPost"] = "/search/releaser"
	data["inputPlaceholder"] = "Type to search for a releaser…"
	err := c.Render(http.StatusOK, name, data)
	if err != nil {
		return InternalErr(c, sl, name, err)
	}
	return nil
}

// SignedOut is the handler to sign out and remove the current session.
func SignedOut(c echo.Context, sl *slog.Logger) error {
	const msg = "signed out context"
	if err := panics.EchoContextS(c, sl); err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}
	const name = "signedout"
	{ // get any existing session
		sess, err := session.Get(sess.Name, c)
		if err != nil {
			return BadRequestErr(c, sl, name, err)
		}
		id, subExists := sess.Values["sub"]
		if !subExists || id == "" {
			return ForbiddenErr(c, sl, name, ErrSession)
		}
		const remove = -1
		sess.Options.MaxAge = remove
		err = sess.Save(c.Request(), c.Response())
		if err != nil {
			return InternalErr(c, sl, name, err)
		}
	}
	return c.Redirect(http.StatusFound, "/")
}

// SignOut is the handler for the Sign out of Defacto2 page.
func SignOut(c echo.Context, sl *slog.Logger) error {
	const msg = "sign out context"
	if err := panics.EchoContextS(c, sl); err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}
	const name = "signout"
	data := empty(c)
	data["noindex"] = true
	data["title"] = "Sign out"
	data["description"] = "Sign out of Defacto2."
	data["h1"] = "Sign out"
	err := c.Render(http.StatusOK, name, data)
	if err != nil {
		return InternalErr(c, sl, name, err)
	}
	return nil
}

// Signin is the handler for the Sign in session page.
func Signin(c echo.Context, sl *slog.Logger, clientID, nonce string) error {
	const msg = "signin context"
	if err := panics.EchoContextS(c, sl); err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}
	const name = "signin"
	data := empty(c)
	data["noindex"] = true
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
			return remove(c, sl, name, data)
		}
		subID, subExists := sess.Values["sub"]
		if !subExists {
			return remove(c, sl, name, data)
		}
		val, valExists := subID.(string)
		if valExists && val != "" {
			return SignOut(c, sl)
		}
	}
	err := c.Render(http.StatusOK, name, data)
	if err != nil {
		return InternalErr(c, sl, name, err)
	}
	return nil
}

// remove is a helper function to remove the session cookie by setting the MaxAge to -1.
func remove(c echo.Context, sl *slog.Logger, name string, data map[string]any) error {
	sess, err := session.Get(sess.Name, c)
	if err == nil {
		const remove = -1
		if sess != nil {
			sess.Options.MaxAge = remove
			_ = sess.Save(c.Request(), c.Response())
		}
	}
	err = c.Render(http.StatusOK, name, data)
	if err != nil {
		return InternalErr(c, sl, name, err)
	}
	return nil
}

// TagEdit handles the post submission for the Tag selection field.
func TagEdit(c echo.Context, db *sql.DB, sl *slog.Logger) error {
	const msg = "tag edit context"
	if err := panics.EchoContextDS(c, db, sl); err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}
	var f Form
	if err := c.Bind(&f); err != nil {
		return badRequest(c, err)
	}
	ctx := context.Background()
	r, err := model.One(ctx, db, true, f.ID)
	if err != nil {
		return fmt.Errorf("tag edit %w: %d", err, f.ID)
	}
	if err = model.UpdateTag(db, int64(f.ID), f.Value); err != nil {
		return badRequest(c, err)
	}
	return c.JSON(http.StatusOK, r)
}

// TagInfo handles the POST submission for the platform and tag info.
func TagInfo(c echo.Context) error {
	const msg = "tag info context"
	if c == nil {
		return fmt.Errorf("%s: %w", msg, panics.ErrNoEchoC)
	}
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

// Titles is the handler for the Titles page.
func Titles(c echo.Context, sl *slog.Logger) error {
	const msg = "titles context"
	if err := panics.EchoContextS(c, sl); err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}
	const name = "titles"
	data := empty(c)
	data["title"] = "Titles"
	data["description"] = "Titles are important."
	data["logo"] = "Artifact Titles"
	data["h1"] = "Titles"
	err := c.Render(http.StatusOK, name, data)
	if err != nil {
		return InternalErr(c, sl, name, err)
	}
	return nil
}

// Thanks is the handler for the Thanks page.
func Thanks(c echo.Context, sl *slog.Logger) error {
	const msg = "thanks context"
	if err := panics.EchoContextS(c, sl); err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}
	const name = "thanks"
	data := empty(c)
	data["description"] = "A special thanks to the hundreds of contributors and the thousands of contributions."
	data["h1"] = "Thank you!"
	data["lead"] = "Thanks to the hundreds of people who have contributed to" +
		" Defacto2 over the decades with file submissions, " +
		"hard drive donations, interviews, corrections, artwork, and monetary contributions!"
	data["title"] = "Thanks!"
	err := c.Render(http.StatusOK, name, data)
	if err != nil {
		return InternalErr(c, sl, name, err)
	}
	return nil
}

// TheScene is the handler for the The Scene page.
func TheScene(c echo.Context, sl *slog.Logger) error {
	const msg = "the scene context"
	if err := panics.EchoContextS(c, sl); err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}
	const name = "thescene"
	const h1 = "The Scene"
	const lead = "The Scene is broad church of people and online communities that is collectively grouped;" +
		" it is subculture of niche activities using personal computers, " +
		"where the participants share creations and exchange ideas."
	data := empty(c)
	data["description"] = "A short introduction on The Scene, the online subcultures and its underground origins."
	data["logo"] = "The underground"
	data["h1"] = h1
	data["lead"] = lead
	data["title"] = h1
	err := c.Render(http.StatusOK, name, data)
	if err != nil {
		return InternalErr(c, sl, name, err)
	}
	return nil
}

// VotePouet is the handler for the Pouet production votes JSON page.
func VotePouet(c echo.Context, sl *slog.Logger, id string) error {
	const msg = "vote pouet context"
	if err := panics.EchoContextS(c, sl); err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}
	const title, name, sep = "Pouet", "pouet", ";"
	pv := pouet.Votes{}
	i, err := strconv.Atoi(id)
	if err != nil {
		return c.String(http.StatusNotFound, err.Error())
	}
	cp := cache.PouetVote
	if s, err := cp.Read(id); err == nil {
		if err := PouetCache(c, s); err == nil {
			if sl != nil {
				sl.Debug("vote pouet", slog.String("cache hit id", id))
			}
			return nil
		}
	}
	if sl != nil {
		sl.Debug("vote pouet", slog.String("cache miss for pouet id", id))
	}
	if err = pv.Votes(i); err != nil {
		return c.String(http.StatusNotFound, err.Error())
	}
	if err = c.JSON(http.StatusOK, pv); err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	val := fmt.Sprintf("%.1f%s%d%s%d%s%d",
		pv.Stars, sep, pv.VotesDown, sep, pv.VotesUp, sep, pv.VotesMeh)
	if err := cp.Write(id, val, cache.ExpiredAt); err != nil {
		if sl != nil {
			sl.Error("vote pouet",
				slog.String("failed", "could not write to the cache database"),
				slog.String("id", id), slog.Any("error", err))
		}
	}
	return nil
}

// Website is the handler for the websites page.
// Open is the ID of the accordion section to open.
func Website(c echo.Context, sl *slog.Logger, open string) error {
	const msg = "website context"
	if err := panics.EchoContextS(c, sl); err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}
	const name = "websites"
	data := empty(c)
	data["title"] = "Websites"
	const logo = "Videos, Books, Films, Sites, Podcasts"
	data["logo"] = logo
	data["description"] = "Our curated collection of " + strings.ToLower(logo) + ", with topics about the Scene."
	accordion := List()
	// Open the accordion section.
	closeAll := true
	for i, site := range accordion {
		if site.ID == open || open == "" {
			site.Open = true
			closeAll = false
			accordion[i] = site
			if open == "" {
				continue
			}
			break
		}
	}
	if closeAll {
		data["noindex"] = true
		data["title"] = "Website categories"
	}
	// If a section was requested but not found, return a 404.
	if open != "hide" && closeAll {
		return StatusErr(c, sl, http.StatusNotFound, open)
	}
	// Render the page.
	data["accordion"] = accordion
	err := c.Render(http.StatusOK, name, data)
	if err != nil {
		return InternalErr(c, sl, fmt.Sprint("render open website,", open), err)
	}
	return nil
}

// Writer is the handler for the Writer page.
func Writer(c echo.Context, db *sql.DB, sl *slog.Logger) error {
	const msg = "writer context"
	if err := panics.EchoContextDS(c, db, sl); err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}
	data := empty(c)
	title := "Writers, editors and authors"
	data["title"] = title
	data["logo"] = title
	data["h1"] = title
	data["description"] = demo
	data["noindex"] = true
	return scener(c, db, sl, postgres.Writer, data)
}

// stats is a helper function for Artifacts that returns the statistics for the files page.
func stats(ctx context.Context, exec boil.ContextExecutor, uri string) (map[string]string, int, error) {
	if !fileslice.Valid(uri) {
		return nil, 0, nil
	}
	m := model.Summary{}
	err := m.ByMatch(ctx, exec, uri)
	if err != nil && !errors.Is(err, model.ErrURI) {
		return nil, 0, fmt.Errorf("artifacts stats %w: %s", err, uri)
	}
	if errors.Is(err, model.ErrURI) {
		switch fileslice.Match(uri) {
		case fileslice.ForApproval:
			if err := m.ByForApproval(ctx, exec); err != nil {
				return nil, 0, fmt.Errorf("artifacts stats for approval %w: %s", err, uri)
			}
		case fileslice.Deletions:
			if err := m.ByHidden(ctx, exec); err != nil {
				return nil, 0, fmt.Errorf("artifacts stats by hidden %w: %s", err, uri)
			}
		case fileslice.Unwanted:
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

// emptyFiles is a map of default values specific to the files templates.
func emptyFiles(c echo.Context) map[string]any {
	data := empty(c)
	data["bbs"] = false
	data["demozoo"] = "0"
	data["csdb"] = 0
	data["janeway"] = 0
	data["scener"] = ""
	data["sixteen"] = ""
	data["tidbits"] = ""
	data["website"] = ""
	data["unknownYears"] = true
	return data
}
