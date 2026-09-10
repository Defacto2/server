//nolint:exhaustive,exhaustruct_v5,wrapcheck
package app

// Package file context.go contains the router handlers for the Defacto2 website.

import (
	"context"
	"crypto/sha512"
	"database/sql"
	"errors"
	"fmt"
	"html/template"
	"io/fs"
	"log/slog"
	"math"
	"net/http"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"sync"

	"github.com/Defacto2/helper"
	"github.com/Defacto2/server/handler/app/remote"
	"github.com/Defacto2/server/handler/areacode"
	"github.com/Defacto2/server/handler/cache"
	"github.com/Defacto2/server/handler/csdb"
	"github.com/Defacto2/server/handler/demozoo"
	"github.com/Defacto2/server/handler/download"
	"github.com/Defacto2/server/handler/internal/fileslice"
	"github.com/Defacto2/server/handler/janeway"
	"github.com/Defacto2/server/handler/pouet"
	"github.com/Defacto2/server/handler/releaser"
	"github.com/Defacto2/server/handler/releaser/lism"
	"github.com/Defacto2/server/handler/sess"
	"github.com/Defacto2/server/handler/site"
	"github.com/Defacto2/server/handler/sixteen"
	"github.com/Defacto2/server/handler/tidbit"
	"github.com/Defacto2/server/internal/config"
	"github.com/Defacto2/server/internal/dir"
	"github.com/Defacto2/server/internal/logs"
	"github.com/Defacto2/server/internal/nils"
	"github.com/Defacto2/server/internal/postgres"
	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/Defacto2/server/internal/tags"
	"github.com/Defacto2/server/model"
	"github.com/Defacto2/server/model/fix"
	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/dustin/go-humanize"
	"github.com/google/uuid"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/v5/session"
	"github.com/labstack/echo/v5"
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
	limit       = 198 // per-page record limit
	sep         = ";"
	az          = ", a-z"
	byyear      = " by year"
	apps        = "apps"
	areacodes   = "areacodes"
	artifact    = "artifact"
	artifacts   = "artifacts"
	bbsx        = "bbs"
	brokentexts = "brokentexts"
	callback    = "google_callback"
	categories  = "categories"
	compression = "compression"
	configs     = "configs"
	dx          = "d"
	fixes       = "fixes"
	fixers      = "fixers"
	ftp         = "ftp"
	history     = "history"
	index       = "index"
	interview   = "interview"
	jsdos       = "jsdos"
	mag         = "magazine"
	newx        = "new"
	pouetx      = "pouet"
	releaserx   = "releaser"
	scener      = "scener"
	searchhtmx  = "searchhtmx"
	searchpost  = "searchpost"
	signedout   = "signedout"
	signin      = "signin"
	signout     = "signout"
	status      = "status"
	routes      = "routes"
	terms       = "terms"
	titles      = "titles"
	thanks      = "thanks"
	thescene    = "thescene"
	vx          = "v"
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
func empty(c *echo.Context) map[string]any {
	const minSize = 20
	m := make(map[string]any, minSize)
	m["cachefiles"] = Caching.RecordCount
	m[canonical] = ""
	m["carousel"] = ""
	m["databaseErr"] = false
	m["description"] = ""
	m["editor"] = sess.Editor(c)
	m["h1"] = ""
	m["subheading"] = ""
	m["jsdos6"] = false
	m["lead"] = ""
	m["logo"] = ""
	m["readonlymode"] = true
	m["title"] = ""
	m["ogtitle"] = ""
	m["noindex"] = false
	m["uploader"] = true
	return m
}

// NOTE:
// - The Artifacts handler below is for multiple artifacts with pagination.
// - The releasers handler are for the releasers and groups without pagination.
// - The individual artifact hander is Dirs.Artifact and is found in the dirs.go file.

// EmptyTester is a map of defaults for the app template tests.
func EmptyTester(c *echo.Context) map[string]any {
	return empty(c)
}

// APIInfo is the handler for the APIInfo end-user helper page.
func APIInfo(sl *slog.Logger, c *echo.Context) error {
	if err := nils.Check(sl, c); err != nil {
		return fmt.Errorf("context api info: %w", err)
	}

	const title = "API Information"
	const descr = "A special thanks to the hundreds of contributors and the thousands of contributions."
	const leadr = "Basic information on how to use the Defacto2 API."
	const name = "api-info"

	data := empty(c)
	data["description"] = descr
	data["h1"] = "RESTful API"
	data["logo"] = "application programming interface"
	data["lead"] = leadr
	data["title"] = title

	err := c.Render(http.StatusOK, name, data)
	if err != nil {
		return InternalErr(sl, c, name, err)
	}
	return nil
}

// Artifacts is the handler for the list and preview of the files page.
// The uri is the category or collection of files to display.
// The page is the page number of the results to display.
func Artifacts(sl *slog.Logger, c *echo.Context, db *sql.DB, uri, page string) error {
	const format = "artifacts context: %w"
	if err := nils.Check(sl, c, db); err != nil {
		return fmt.Errorf(format, err)
	}

	ctx := c.Request().Context()
	switch {
	case !fileslice.Valid(uri):
		return ArtifactsErr(sl, c, uri)

	case page == "":
		return artifactsTable(ctx, sl, c, db, uri, 1)

	default:
		p, err := strconv.Atoi(page)
		if err != nil {
			return PageErr(sl, c, uri, page)
		}

		return artifactsTable(ctx, sl, c, db, uri, p)
	}
}

// artifactsTable is a helper function for Artifacts that returns the data map for the files page.
func artifactsTable(ctx context.Context, sl *slog.Logger, c *echo.Context, db *sql.DB, uri string, page int) error {
	const format = "sub-artifacts context: %w"
	if err := nils.Check(ctx, sl, c, db); err != nil {
		return fmt.Errorf(format, err)
	}

	const title = "Artifacts"
	const descr = "Table of contents for the collection of artifacts."
	logo, subhead, lead := fileslice.FileInfo(uri)

	data := emptyFiles(c)
	data["title"] = title
	data[canonical] = strings.Join([]string{files, uri}, "/")
	data["description"] = descr
	data["logo"] = logo
	data["h1"] = title
	data["subheading"] = subhead
	data["lead"] = lead
	data["unknownYears"] = true
	data["forApproval"] = false

	spage := strconv.Itoa(page)
	errURI := "artifacts page " + spage + " for '" + uri + "'"

	d, sum, err := stats(ctx, db, uri)
	if err != nil {
		return DatabaseErr(sl, c, errURI, err)
	}
	lastPage := math.Ceil(float64(sum) / float64(limit))
	if page > int(lastPage) {
		return PageErr(sl, c, uri, spage)
	}

	data = artifactsDesc(uri, d[years], sum, data)
	data["stats"] = d

	data[records] = []models.FileSlice{}
	r, err := fileslice.Records(ctx, db, uri, page, limit)
	if err != nil {
		return DatabaseErr(sl, c, errURI, err)
	}
	if len(r) == 0 {
		if err = c.Render(http.StatusOK, artifacts, data); err != nil {
			return InternalErr(sl, c, errURI, err)
		}
		return nil
	}
	data[records] = r

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

	if err = c.Render(http.StatusOK, artifacts, data); err != nil {
		return InternalErr(sl, c, errURI, err)
	}

	return nil
}

func artifactsDesc(uri, years string, sum int, data map[string]any) map[string]any {
	match := fileslice.Match(uri)
	if match == -1 {
		return data
	}

	data["noindex"] = true

	switch match {
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
	case fileslice.Sensenstahl:
		data["title"] = "Sensenstahl artifacts"
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

// Apps is the handler for the modern applications and tools page.
func Apps(sl *slog.Logger, c *echo.Context) error {
	const format = "apps handler context: %w"
	if err := nils.Check(sl, c); err != nil {
		return fmt.Errorf(format, err)
	}

	const title = "Use current apps"
	const descr = "Software and application suggestions for using the historic artifacts and " +
		"file downloads on modern systems."
	const leadr = "Here are some software suggestions and Windows, Linux, macOS tools for running out-of-date programs " +
		"and using legacy media formats."

	data := empty(c)
	data["title"] = title
	data["description"] = descr
	data["logo"] = "Software suggestions"
	data["h1"] = "Modern Applications and Tools"
	data["lead"] = leadr

	err := c.Render(http.StatusOK, apps, data)
	if err != nil {
		return InternalErr(sl, c, apps, err)
	}
	return nil
}

// Areacodes is the handler for the BBS and telephone area codes page.
func Areacodes(sl *slog.Logger, c *echo.Context) error {
	const format = "areacodes context: %w"
	if err := nils.Check(sl, c); err != nil {
		return fmt.Errorf(format, err)
	}

	const title = "BBS and telephone area codes"
	const descr = "Lookup and list the North American Numbering Plan (NANP) area codes in common use until 1994."
	const leadr = "North American Numbering Plan (+1-XXX) telephone area codes until 1994."
	const logo = "BBS area codes"

	data := empty(c)
	data["title"] = title
	data["description"] = descr
	data["logo"] = logo
	data["h1"] = logo
	data["lead"] = leadr
	data["telephonecodes"] = areacode.AreaCodes()
	data["territories"] = areacode.Regions()
	data["abbreviations"] = areacode.Abbreviations()

	err := c.Render(http.StatusOK, areacodes, data)
	if err != nil {
		return InternalErr(sl, c, areacodes, err)
	}
	return nil
}

// Artist is the handler for the Artist sceners page.
func Artist(sl *slog.Logger, c *echo.Context, db *sql.DB) error {
	const format = "artist context: %w"
	if err := nils.Check(sl, c, db); err != nil {
		return fmt.Errorf(format, err)
	}

	const title = "Pixel artists and graphic designers"
	data := empty(c)
	data["title"] = title
	data["logo"] = title
	data["h1"] = title
	data["noindex"] = true

	ctx := c.Request().Context()
	return scenerPage(ctx, sl, c, db, postgres.Artist, data)
}

// scenerPage is the handler for the scenerPage pages.
func scenerPage(ctx context.Context, sl *slog.Logger, c *echo.Context, db *sql.DB, r postgres.Role,
	data map[string]any,
) error {
	if err := nils.Check(ctx, sl, c, db); err != nil {
		return fmt.Errorf("scener context: %w", err)
	}

	const descr = "This is a massive but incomplete list of aliases and pseudonyms used " +
		"in the Scene and offers links to individual profiles."

	s := model.Sceners{}
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
	default:
		// Handle unknown roles gracefully
		err = fmt.Errorf("unknown role %s: %w", r, model.ErrModel)
	}
	if err != nil {
		return DatabaseErr(sl, c, scener, err)
	}

	data["sceners"] = s.Sort()
	data["description"] = descr
	data["lead"] = "This page shows the sceners and people credited for their work in The Scene." +
		`<br><small class="fw-lighter">` +
		"The list will never be complete or accurate due to the amount of data and the lack of a" +
		" common format for crediting people. " +
		" Sceners often used different names or spellings on their work, including character" +
		" swaps, aliases, initials, and even single-letter signatures." +
		"</small>"

	err = c.Render(http.StatusOK, scener, data)
	if err != nil {
		return InternalErr(sl, c, scener, err)
	}
	return nil
}

// BBS is the handler for the BBS page ordered by the most files.
func BBS(sl *slog.Logger, c *echo.Context, db *sql.DB) error {
	ctx := c.Request().Context()
	return bbsHandler(ctx, sl, c, db, model.Prolific)
}

// BBSAZ is the handler for the BBS page ordered alphabetically.
func BBSAZ(sl *slog.Logger, c *echo.Context, db *sql.DB) error {
	ctx := c.Request().Context()
	return bbsHandler(ctx, sl, c, db, model.Alphabetical)
}

// BBSYear is the handler for the BBS page ordered by the year.
func BBSYear(sl *slog.Logger, c *echo.Context, db *sql.DB) error {
	ctx := c.Request().Context()
	return bbsHandler(ctx, sl, c, db, model.Oldest)
}

// bbsHandler is the handler for the BBS page.
func bbsHandler(ctx context.Context, sl *slog.Logger, c *echo.Context, db *sql.DB, orderBy model.OrderBy) error {
	const format = "bbs handler context: %w"
	if err := nils.Check(ctx, sl, c, db); err != nil {
		return fmt.Errorf(format, err)
	}

	const title = "BBS"
	const leadr = "Bulletin Board Systems were personal computers networked using the copper telephone network " +
		"and provided communication services, file hosting and exchanges."
	const logo = "Bulletin Board Systems"
	const key = "releasers"

	data := empty(c)
	data["noindex"] = true
	data["title"] = "Former " + title
	data["description"] = leadr
	data["logo"] = logo
	data["h1"] = title
	data["lead"] = leadr
	data["itemName"] = bbsx
	data[key] = model.Releasers{}
	data["stats"] = map[string]string{}

	r := model.Releasers{}
	if err := orderBy.BBS(ctx, db, &r); err != nil {
		return DatabaseErr(sl, c, bbsx, err)
	}

	data[key] = r
	tmpl := bbsx
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
		order = byYear
	default:
		// Handle unknown order types
		order = alpha
	}

	data["stats"] = map[string]string{
		pubs:   fmt.Sprintf("%d boards", len(r)),
		ordrby: order,
	}

	err := c.Render(http.StatusOK, tmpl, data)
	if err != nil {
		return InternalErr(sl, c, bbsx, err)
	}
	return nil
}

// BrokenTexts is the handler for the Broken texts page.
func BrokenTexts(sl *slog.Logger, c *echo.Context) error {
	const format = "broken texts context: %w"
	if err := nils.Check(c, sl); err != nil {
		return fmt.Errorf(format, err)
	}

	const title = "Broken texts!?"
	const descr = "Learn why there are broken encodings, unreadable texts and corrupted nfo files."
	const leadr = "Unfortunately, there are large numbers of incomplete, inaccurate, " +
		"or corrupted information texts (NFOs). While we'd prefer to offer a pristine copy of a Scene text, " +
		"hosting a broken text is more useful than offering nothing."

	data := empty(c)
	data["title"] = title
	data["description"] = descr
	data["logo"] = "Broken text files and NFOs"
	data["h1"] = "Broken text files"
	data["lead"] = leadr

	err := c.Render(http.StatusOK, brokentexts, data)
	if err != nil {
		return InternalErr(sl, c, brokentexts, err)
	}
	return nil
}

// Checksum is the handler for the Checksum file record page.
func Checksum(sl *slog.Logger, c *echo.Context, db *sql.DB, id string) error {
	const format = "checksum context: %w"
	if err := nils.Check(sl, c, db); err != nil {
		return fmt.Errorf(format, err)
	}

	const uri = "sum"

	ctx := c.Request().Context()

	if err := download.Checksum(ctx, c, db, id); err != nil {
		if errors.Is(err, download.ErrStat) {
			return FileMissingErr(sl, c, uri, err)
		}
		return DownloadErr(sl, c, uri, err)
	}
	return nil
}

// Coder is the handler for the Coder sceners page.
func Coder(sl *slog.Logger, c *echo.Context, db *sql.DB) error {
	const format = "coder context: %w"
	if err := nils.Check(sl, c, db); err != nil {
		return fmt.Errorf(format, err)
	}

	const title = "Coder and programmers"
	data := empty(c)
	data["title"] = title
	data["logo"] = title
	data["h1"] = title
	data["noindex"] = true

	ctx := c.Request().Context()
	return scenerPage(ctx, sl, c, db, postgres.Writer, data)
}

// Compression is the handler for information on historic compression tools page.
func Compression(sl *slog.Logger, c *echo.Context) error {
	const format = "compression context: %w"
	if err := nils.Check(c, sl); err != nil {
		return fmt.Errorf(format, err)
	}

	const title = "Compression and archiving formats"
	const descr = "Old file archives and compression methods used in the 1980s on the PC."
	const leadr = "Compression and archiving formats of the 1980s were evolving by the month, " +
		"and today, are hard to parse."

	data := empty(c)
	data["title"] = title
	data["description"] = descr
	data["logo"] = "Old archives"
	data["h1"] = "File compression formats"
	data["lead"] = leadr

	err := c.Render(http.StatusOK, compression, data)
	if err != nil {
		return InternalErr(sl, c, compression, err)
	}
	return nil
}

// Configurations is the handler for the Configuration page.
func Configurations(sl *slog.Logger, c *echo.Context, db *sql.DB, conf config.Config) error {
	const format = "configurations context: %w"
	if err := nils.Check(sl, c, db); err != nil {
		return fmt.Errorf(format, err)
	}

	const title = "Configs"
	const descr = "Defacto2 configurations."
	const leadr = "The web application configurations, tools and links to special records."

	data := empty(c)
	data["description"] = descr
	data["h1"] = "Configurations"
	data["lead"] = leadr
	data["title"] = title
	data["configurations"] = conf
	data["countArtifacts"] = 0
	data["countPublic"] = 0
	data["countNewUpload"] = 0
	data["countHidden"] = 0
	data["uuidVersions"] = ""

	ctx := c.Request().Context()
	// As we are collecting stats of both the file system and database, we may as well do it cocurrently
	var wg sync.WaitGroup
	var mu sync.Mutex

	wg.Go(func() {
		mu.Lock()
		ca, cp, cnu, err := model.Counts(ctx, db)
		if err != nil {
			return
		}
		data["countArtifacts"] = ca
		data["countPublic"] = cp
		data["countNewUpload"] = cnu
		data["countHidden"] = ca - cp - cnu
		mu.Unlock()
	})

	wg.Go(func() {
		mu.Lock()
		data = configurations(data, conf)
		mu.Unlock()
	})

	wg.Go(func() {
		mu.Lock()
		vers, err := model.UUIDs(ctx, db)
		if err != nil {
			return
		}
		data["uuidVersions"] = vers
		mu.Unlock()
	})

	wg.Wait()
	if db == nil {
		data["dbConnections"] = "database not set"
		err := c.Render(http.StatusOK, configs, data)
		if err != nil {
			return InternalErr(sl, c, configs, err)
		}
		return nil
	}

	count, maximum, err := postgres.Connections(ctx, db)
	if err != nil {
		data["dbConnections"] = err.Error()
	} else {
		data["dbConnections"] = fmt.Sprintf("%d of %d", count, maximum)
	}

	err = c.Render(http.StatusOK, configs, data)
	if err != nil {
		return InternalErr(sl, c, configs, err)
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
		wdt, wdu, _, wdp, _ := helper.DiskStat(conf.AbsDownload.String())
		data["wdTotal"] = helper.ByteCount(int64(wdt))
		data["wdFree"] = helper.ByteCount(int64(wdu))
		data["wdPercent"] = wdp
		mu.Unlock()
	})

	wg.Wait()

	return data
}

func downloader(data map[string]any, conf config.Config) map[string]any {
	data["countDownloads"] = 0
	data["usageDownloads"] = 0
	data["extsDownloads"] = []helper.Extension{}

	download := dir.Directory(string(conf.AbsDownload))
	check := config.CheckDir(download, "downloads")
	data["checkDownloads"] = check
	if check == nil {
		c, _ := helper.Count(string(conf.AbsDownload))
		data["countDownloads"] = humanize.Comma(int64(c))
		b, _ := helper.DiskUsage(string(conf.AbsDownload))
		data["usageDownloads"] = helper.ByteCountFloat(b)
		exts, _ := helper.CountExts(string(conf.AbsDownload))
		data["extsDownloads"] = exts
	}

	return data
}

func previewer(data map[string]any, conf config.Config) map[string]any {
	data["countPreviews"] = 0
	data["usagePreviews"] = 0
	data["extsPreviews"] = []helper.Extension{}

	preview := dir.Directory(conf.AbsPreview)
	check := config.CheckDir(preview, "previews")
	data["checkPreviews"] = check
	if check == nil {
		c, _ := helper.Count(conf.AbsPreview.String())
		data["countPreviews"] = humanize.Comma(int64(c))
		b, _ := helper.DiskUsage(string(conf.AbsPreview))
		data["usagePreviews"] = helper.ByteCountFloat(b)
		exts, _ := helper.CountExts(conf.AbsPreview.String())
		data["extsPreviews"] = exts
	}

	return data
}

func thumbnailer(data map[string]any, conf config.Config) map[string]any {
	data["countThumbnails"] = 0
	data["usageThumbnails"] = 0
	data["extsThumbnails"] = []helper.Extension{}

	thumbnail := dir.Directory(conf.AbsThumbnail)
	check := config.CheckDir(thumbnail, "thumbnails")
	data["checkThumbnails"] = check
	if check == nil {
		c, _ := helper.Count(conf.AbsThumbnail.String())
		data["countThumbnails"] = humanize.Comma(int64(c))
		b, _ := helper.DiskUsage(string(conf.AbsThumbnail))
		data["usageThumbnails"] = helper.ByteCountFloat(b)
		exts, _ := helper.CountExts(conf.AbsThumbnail.String())
		data["extsThumbnails"] = exts
	}

	return data
}

func extraer(data map[string]any, conf config.Config) map[string]any {
	data["countExtras"] = 0
	data["usageExtras"] = 0
	data["extsExtras"] = []helper.Extension{}

	extra := dir.Directory(conf.AbsExtra)
	check := config.CheckDir(extra, "extra")
	data["checkExtras"] = check
	if check == nil {
		c, _ := helper.Count(conf.AbsExtra.String())
		data["countExtras"] = humanize.Comma(int64(c))
		b, _ := helper.DiskUsage(string(conf.AbsExtra))
		data["usageExtras"] = helper.ByteCountFloat(b)
		exts, _ := helper.CountExts(conf.AbsExtra.String())
		data["extsExtras"] = exts
	}

	return data
}

func orphaneder(data map[string]any, conf config.Config) map[string]any {
	data["countOrphaned"] = 0
	data["usageOrphaned"] = 0
	data["extsOrphaned"] = []helper.Extension{}

	orphaned := dir.Directory(conf.AbsOrphaned.String())
	check := config.CheckDir(orphaned, "orphaned")
	data["checkOrphaned"] = check
	if check == nil {
		c, _ := helper.Count(conf.AbsOrphaned.String())
		data["countOrphaned"] = humanize.Comma(int64(c))
		b, _ := helper.DiskUsage(string(conf.AbsOrphaned))
		data["usageOrphaned"] = helper.ByteCountFloat(b)
		exts, _ := helper.CountExts(conf.AbsOrphaned.String())
		data["extsOrphaned"] = exts
	}

	return data
}

// DownloadJsDos is the handler for the js-dos emulator to download zip files that are then
// mounted as a C: hard drive in the emulation. js-dos only supports common zip compression methods,
// so this func first attempts to offer a re-archived zip file found in the extra directory, and
// only if that fails does it offer the original download file.
func DownloadJsDos(sl *slog.Logger, c *echo.Context, db *sql.DB, extra, downl dir.Directory,
) error {
	const format = "download jsdos context: %w"
	if err := nils.Check(sl, c, db); err != nil {
		return fmt.Errorf(format, err)
	}

	e := download.ExtraZip{
		Extra:    extra,
		Download: downl,
	}

	ctx := c.Request().Context()
	if err := e.HTTPSend(ctx, c, db); err != nil {
		if errors.Is(err, download.ErrStat) {
			return FileMissingErr(sl, c, jsdos, err)
		}
		return DownloadErr(sl, c, jsdos, err)
	}
	return nil
}

// Download is the handler for the Download file record page.
func Download(sl *slog.Logger, c *echo.Context, db *sql.DB, path dir.Directory) error {
	const format = "download context: %w"
	if err := nils.Check(sl, c, db); err != nil {
		return fmt.Errorf(format, err)
	}

	d := download.Download{
		Inline: false,
		Dir:    path,
	}

	if id := c.Param("id"); id != "" {
		r := c.Response()
		r.Header().Set("Link", `<https://defacto2.net/`+dx+`/`+id+`; rel=canon>`)
	}

	if err := d.HTTPSend(sl, c, db); err != nil {
		if errors.Is(err, download.ErrStat) {
			return FileMissingErr(sl, c, dx, err)
		}
		return DownloadErr(sl, c, dx, err)
	}
	return nil
}

// FTP is the handler for the FTP page.
func FTP(sl *slog.Logger, c *echo.Context, db *sql.DB) error {
	const format = "ftp context: %w"
	if err := nils.Check(sl, c, db); err != nil {
		return fmt.Errorf(format, err)
	}

	const title = "FTP"
	const descr = "FTP sites were Internet-based file exchange servers that would host and share Scene releases."
	data := empty(c)
	data["title"] = "Former " + title + " sites"
	data["description"] = descr
	data["logo"] = "FTP sites, A-Z"
	data["h1"] = title
	data["lead"] = descr

	// releaser.html specific data items
	data["itemName"] = ftp
	const key = "releasers"
	data[key] = model.Releasers{}
	data["stats"] = map[string]string{}

	r := model.Releasers{}
	ctx := c.Request().Context()
	if err := r.FTP(ctx, db); err != nil {
		return DatabaseErr(sl, c, ftp, err)
	}
	data[key] = r

	data["stats"] = map[string]string{
		pubs:   fmt.Sprintf("%d sites", len(r)),
		ordrby: alpha,
	}

	err := c.Render(http.StatusOK, ftp, data)
	if err != nil {
		return InternalErr(sl, c, ftp, err)
	}
	return nil
}

// Categories is the handler for the artifact categories page.
func Categories(sl *slog.Logger, c *echo.Context, db *sql.DB, stats bool) error {
	const format = "categories context: %w"
	if err := nils.Check(sl, c, db); err != nil {
		return fmt.Errorf(format, err)
	}

	const title = "Artifact categories"
	const descr = "A table of contents for the collection."
	const leadr = "This page shows the categories and platforms in the collection of file artifacts."

	data := empty(c)
	data["noindex"] = true
	data["title"] = title
	data["description"] = descr
	data["logo"] = title
	data["h1"] = title
	data["lead"] = leadr
	data["stats"] = stats
	data["counter"] = fileslice.Statistics()

	ctx := c.Request().Context()
	data, err := fileWStats(ctx, db, data, stats)
	if err != nil {
		sl.Warn("context_categories", slog.Any("error", err))
		data["databaseErr"] = true
	}

	err = c.Render(http.StatusOK, categories, data)
	if err != nil {
		return InternalErr(sl, c, categories, err)
	}
	return nil
}

// fileWStats is a helper function for File that adds the statistics to the data map.
func fileWStats(ctx context.Context, db *sql.DB, data map[string]any, stats bool) (map[string]any, error) {
	empty := make(map[string]any)
	if err := nils.Check(ctx, db); err != nil {
		return empty, fmt.Errorf("file with stats: %w", err)
	}

	const title = "Artifact category statistics"
	const descr = "This page shows the artifacts categories with selected statistics, " +
		"such as the number of files in the category or platform."
	if data == nil {
		data = empty
	}
	if !stats {
		return data, nil
	}

	c, err := fileslice.Counter(ctx, db)
	if err != nil {
		return data, fmt.Errorf("counter: %w", err)
	}

	data["counter"] = c
	data["orderByBytes"] = c.SortByte()
	data["orderByCount"] = c.SortCount()
	data["orderByName"] = c.SortName()
	data["orderByYear"] = c.SortYear()
	data["logo"] = title
	data["lead"] = descr

	return data, nil
}

// Deletions is the handler to list the files that have been marked for deletion.
func Deletions(sl *slog.Logger, c *echo.Context, db *sql.DB, page string) error {
	const format = "deletions context: %w"
	if err := nils.Check(sl, c, db); err != nil {
		return fmt.Errorf(format, err)
	}

	uri := fileslice.Deletions.String()
	if !fileslice.Valid(uri) {
		return ArtifactsErr(sl, c, uri)
	}

	ctx := c.Request().Context()
	if page == "" {
		return artifactsTable(ctx, sl, c, db, uri, 1)
	}

	p, err := strconv.Atoi(page)
	if err != nil {
		return PageErr(sl, c, uri, page)
	}

	return artifactsTable(ctx, sl, c, db, uri, p)
}

// Fixes is the handler for the problems and fixes page.
func Fixes(sl *slog.Logger, c *echo.Context) error {
	const format = "fixes for programs context: %w"
	if err := nils.Check(c, sl); err != nil {
		return fmt.Errorf(format, err)
	}

	const title = "Problems and fixes"
	const descr = "Fix common errors found in applications and tools authored by the Scene for Windows and DOS."
	const leadr = "Shrinker dispatcher or runtime 200 errors, or missing NPMOD32.DLL or D3DRM.DLL files?"

	data := empty(c)
	data["title"] = title
	data["description"] = descr
	data["logo"] = title
	data["h1"] = "Common problems and fixes"
	data["lead"] = leadr

	err := c.Render(http.StatusOK, fixes, data)
	if err != nil {
		return InternalErr(sl, c, fixes, err)
	}
	return nil
}

// Routes is the handler for the listing of all the routes page.
func Routes(sl *slog.Logger, c *echo.Context, r echo.Routes) error {
	const format = "routes context: %w"
	if err := nils.Check(sl, c, r); err != nil {
		return fmt.Errorf(format, err)
	}

	const title = "List of routes"
	const descr = "Lists the web browser routes and parameters registered by the web application."

	data := empty(c)
	data["title"] = title
	data["description"] = descr
	data["logo"] = "Routes"
	data["routesList"] = r
	data["routesCount"] = len(r)

	err := c.Render(http.StatusOK, routes, data)
	if err != nil {
		return InternalErr(sl, c, routes, err)
	}
	return nil
}

// Terms is the handler for the problems and fixes page.
func Terms(sl *slog.Logger, c *echo.Context) error {
	const format = "glossary of terms context: %w"
	if err := nils.Check(c, sl); err != nil {
		return fmt.Errorf(format, err)
	}

	const title = "Common terms"
	const descr = "This is a list of the unique and common terms used in the scene."
	const leadr = "A glossary of unique and common terms used in The Scene."

	data := empty(c)
	data["title"] = title
	data["description"] = descr
	data["logo"] = "Glossary of terms"
	data["h1"] = "Glossary of common terms"
	data["lead"] = leadr

	err := c.Render(http.StatusOK, terms, data)
	if err != nil {
		return InternalErr(sl, c, terms, err)
	}
	return nil
}

// Unwanted is the handler to list the files that have been marked as unwanted.
func Unwanted(sl *slog.Logger, c *echo.Context, db *sql.DB, page string) error {
	const format = "unwanted context: %w"
	if err := nils.Check(sl, c, db); err != nil {
		return fmt.Errorf(format, err)
	}

	uri := fileslice.Unwanted.String()
	if !fileslice.Valid(uri) {
		return ArtifactsErr(sl, c, uri)
	}

	ctx := c.Request().Context()
	if page == "" {
		return artifactsTable(ctx, sl, c, db, uri, 1)
	}

	p, err := strconv.Atoi(page)
	if err != nil {
		return PageErr(sl, c, uri, page)
	}

	return artifactsTable(ctx, sl, c, db, uri, p)
}

// ForApproval is the handler for the list and preview of the files page.
// The uri is the category or collection of files to display.
// The page is the page number of the results to display.
func ForApproval(sl *slog.Logger, c *echo.Context, db *sql.DB, page string) error {
	const format = "for approval context: %w"
	if err := nils.Check(sl, c, db); err != nil {
		return fmt.Errorf(format, err)
	}

	uri := fileslice.ForApproval.String()
	if !fileslice.Valid(uri) {
		return ArtifactsErr(sl, c, uri)
	}

	ctx := c.Request().Context()
	if page == "" {
		return artifactsTable(ctx, sl, c, db, uri, 1)
	}

	p, err := strconv.Atoi(page)
	if err != nil {
		return PageErr(sl, c, uri, page)
	}

	return artifactsTable(ctx, sl, c, db, uri, p)
}

// GetDemozooParam fetches the multiple download_links values from the
// Demozoo production API and attempts to download and save one of the
// linked files. If multiple links are found, the first link is used as
// they should all point to the same asset.
//
// Both the Demozoo production ID param and the Defacto2 UUID query
// param values are required as params to fetch the production data and
// to save the file to the correct filename.
func GetDemozooParam(sl *slog.Logger, c *echo.Context, tx *sql.Tx, download dir.Directory) error {
	const format = "get demozoo param context: %w"
	if err := nils.Check(sl, c, tx); err != nil {
		return fmt.Errorf(format, err)
	}

	got := remote.Demozoo(0, "", download, 0)
	id, err := echo.PathParam[int](c, "id")
	if err != nil {
		got.Error = "demozoo id must be a numeric value"
		return c.JSON(http.StatusBadRequest, got)
	}
	got.ID = id

	unid := c.QueryParam("unid")
	if err = uuid.Validate(unid); err != nil {
		got.Error = "uuid syntax did not validate, " + unid
		return c.JSON(http.StatusBadRequest, got)
	}
	got.UUID = unid

	ctx := c.Request().Context()
	return got.Download(ctx, sl, c, tx)
}

// GetDemozoo fetches the download link from Demozoo and saves it to the download directory.
// It then runs Update to modify the database record with various metadata from the file and Demozoo record API data.
//
// This function is a wrapper for the remote.DemozooLink.Download method.
func GetDemozoo(
	ctx context.Context, sl *slog.Logger, c *echo.Context, tx *sql.Tx,
	prodID int, unid string, download dir.Directory,
) error {
	got := remote.Demozoo(prodID, unid, download, 0)
	return got.Download(ctx, sl, c, tx)
}

// GetPouet fetches the download link from Pouet and saves it to the download directory.
// It then runs Update to modify the database record with various metadata from the file and Pouet record API data.
//
// This function is a wrapper for the remote.PouetLink.Download method.
func GetPouet(
	ctx context.Context, sl *slog.Logger, c *echo.Context, tx *sql.Tx,
	prodID int, unid string, download dir.Directory,
) error {
	got := remote.Pouet(prodID, unid, download, 0)
	return got.Download(ctx, sl, c, tx)
}

// GoogleCallback is the handler for the Google OAuth2 callback page to verify
// the [Google ID token].
//
// [Google ID token]: https://developers.google.com/identity/gsi/web/guides/verify-google-id-token
func GoogleCallback(sl *slog.Logger, c *echo.Context, clientID string, maxAge int, accounts ...[48]byte,
) error {
	const format = "google callback context: %w"
	if err := nils.Check(sl, c); err != nil {
		return fmt.Errorf(format, err)
	}

	// Cross-Site Request Forgery cookie token
	const csrf = "g_csrf_token"
	cookie, err := c.Cookie(csrf)
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			return c.Redirect(http.StatusForbidden, "/signin")
		}
		return BadRequestErr(sl, c, callback, err)
	}
	token := cookie.Value

	// Cross-Site Request Forgery post token
	bodyToken := c.FormValue(csrf)
	if token != bodyToken {
		return BadRequestErr(sl, c, callback, ErrMisMatch)
	}

	// Create a new token verifier.
	// https://pkg.go.dev/google.golang.org/api/idtoken
	ctx := c.Request().Context()
	validator, err := idtoken.NewValidator(ctx)
	if err != nil {
		return BadRequestErr(sl, c, callback, err)
	}

	// Verify the ID token and using the client ID from the Google API.
	credential := c.FormValue("credential")
	playload, err := validator.Validate(ctx, credential, clientID)
	if err != nil {
		return BadRequestErr(sl, c, callback, err)
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
		const format = "%w. If this is a mistake, contact Defacto2 admin and give them this Google account ID: %s"
		return ForbiddenErr(sl, c, callback, fmt.Errorf(format, ErrUser, sub))
	}

	if err = sessionHandler(c, maxAge, playload.Claims); err != nil {
		return BadRequestErr(sl, c, callback, err)
	}
	return c.Redirect(http.StatusFound, "/")
}

// sessionHandler creates a [new session] and populates it with
// the claims data created by the [ID Tokens for Google HTTP APIs].
//
// [new session]: https://pkg.go.dev/github.com/gorilla/sessions
// [ID Tokens for Google HTTP APIs]: https://pkg.go.dev/google.golang.org/api/idtoken
func sessionHandler(c *echo.Context, maxAge int, claims map[string]any,
) error {
	const format = "session handler get: %w"
	if err := nils.Check(c); err != nil {
		return fmt.Errorf(format, err)
	}

	session, err := session.Get(sess.Name, c)
	if err != nil {
		return fmt.Errorf(format, err)
	}

	// session Options are cookie options and are all optional
	// https://developer.mozilla.org/en-US/docs/Web/HTTP/Cookies
	const hour = 60 * 60
	session.Options = &sessions.Options{
		Path:        "/",                  // path that must exist in the requested URL to send the Cookie header
		Domain:      "",                   // which server can receive a cookie
		MaxAge:      hour * maxAge,        // maximum age for the cookie, in seconds
		Secure:      true,                 // cookie requires HTTPS except for localhost
		HttpOnly:    true,                 // stops the cookie being read by JS
		SameSite:    http.SameSiteLaxMode, // LaxMode (default) or StrictMode
		Partitioned: false,
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
func History(sl *slog.Logger, c *echo.Context) error {
	const format = "history context: %w"
	if err := nils.Check(c, sl); err != nil {
		return fmt.Errorf(format, err)
	}

	const title = "The history of the brand"
	const descr = "Learn about the many iterations of Defacto2 and the original DeFacto magazine from 1996."
	const leadr = "In the past, alternative iterations of the name have included" +
		" De Facto, DF, DeFacto, Defacto II, Defacto 2, and the defacto2.com domain."

	data := empty(c)
	data["carousel"] = `#carouselDf2Artpacks`
	data["description"] = descr
	data["logo"] = "The history of Defacto"
	data["h1"] = title
	data["lead"] = leadr
	data["title"] = title

	err := c.Render(http.StatusOK, history, data)
	if err != nil {
		return InternalErr(sl, c, history, err)
	}
	return nil
}

// Index is the handler for the Home page.
func Index(sl *slog.Logger, c *echo.Context) error {
	const format = "index context: %w"
	if err := nils.Check(c, sl); err != nil {
		return fmt.Errorf(format, err)
	}

	const title = "Introduction and milestones"
	const h1 = "The subcultures of obsolete microcomputers"

	data := empty(c)
	// NOTE: if the title get's changed, the indexJS conditional in layout.tmpl needs updating
	data["title"] = title
	data[canonical] = "/"
	data["h1"] = h1
	data["milestones"] = Collection()
	{
		// get the given name of the signed in session
		sess, err := session.Get(sess.Name, c)
		if err == nil {
			if givenName, givenExists := sess.Values["givenName"]; givenExists {
				if name, nameStr := givenName.(string); nameStr && name != "" {
					data["h1"] = "Welcome, " + name
				}
			}
		}
	}

	err := c.Render(http.StatusOK, index, data)
	if err != nil {
		return InternalErr(sl, c, index, err)
	}
	return nil
}

// Inline is the handler for the Download file record page.
func Inline(sl *slog.Logger, c *echo.Context, db *sql.DB, path dir.Directory) error {
	const format = "inline context: %w"
	if err := nils.Check(sl, c, db); err != nil {
		return fmt.Errorf(format, err)
	}

	d := download.Download{
		Inline: true,
		Dir:    path,
	}

	if err := d.HTTPSend(sl, c, db); err != nil {
		if errors.Is(err, download.ErrStat) {
			return FileMissingErr(sl, c, vx, err)
		}
		return DownloadErr(sl, c, vx, err)
	}
	return nil
}

// Interview is the handler for the People Interviews page.
func Interview(sl *slog.Logger, c *echo.Context) error {
	const format = "interview context: %w"
	if err := nils.Check(c, sl); err != nil {
		return fmt.Errorf(format, err)
	}

	const title = "Interviews with former Sceners"
	const descr = "A collection of historical interviews and discussions with former members of the Scene."
	const leadr = "Here is a centralized page for the discussions and unedited" +
		" interviews with former sceners, crackers, and demo makers."

	data := empty(c)
	data["title"] = title
	data["description"] = descr
	data["logo"] = title
	data["h1"] = title
	data["lead"] = leadr
	data["interviews"] = Interviewees()

	err := c.Render(http.StatusOK, interview, data)
	if err != nil {
		return InternalErr(sl, c, interview, err)
	}
	return nil
}

// Magazine is the handler for the Magazine page.
func Magazine(sl *slog.Logger, c *echo.Context, db *sql.DB) error {
	return magazines(c.Request().Context(), sl, c, db, true)
}

// MagazineAZ is the handler for the Magazine page ordered chronologically.
func MagazineAZ(sl *slog.Logger, c *echo.Context, db *sql.DB) error {
	return magazines(c.Request().Context(), sl, c, db, false)
}

// magazines is the handler for the magazine page.
func magazines(ctx context.Context, sl *slog.Logger, c *echo.Context, db *sql.DB, chronological bool) error {
	const format = "magazines context: %w"
	if err := nils.Check(ctx, sl, c, db); err != nil {
		return fmt.Errorf(format, err)
	}

	const title = "Magazines"
	const descr = "Scene magazines are the newsletters, reports, " +
		"and publications on the activities of the subculture community."
	const issue = "issue"
	const key = "releasers"

	data := empty(c)
	data["title"] = title
	data["description"] = descr
	data["logo"] = title
	data["h1"] = title
	data["lead"] = descr
	data["itemName"] = issue
	data[key] = model.Releasers{}
	data["stats"] = map[string]string{}

	var order string
	r := model.Releasers{}
	render := mag

	switch chronological {
	case true:
		if err := r.Magazine(ctx, db); err != nil {
			return DatabaseErr(sl, c, mag, err)
		}
		s := title + byyear
		data["logo"] = s
		data["title"] = title + byyear
		order = byYear
	case false:
		render = mag + "-az"
		if err := r.MagazineAZ(ctx, db); err != nil {
			return DatabaseErr(sl, c, mag, err)
		}
		data["noindex"] = true
		s := title + az
		data["logo"] = s
		data["title"] = title + az
		order = alpha
	default:
		// Handle unknown order types
		order = alpha
	}

	data[key] = r
	data["stats"] = map[string]string{
		pubs:   fmt.Sprintf("%d publications", len(r)),
		ordrby: order,
	}

	err := c.Render(http.StatusOK, render, data)
	if err != nil {
		return InternalErr(sl, c, mag, err)
	}
	return nil
}

// Musician is the handler for the Musiciansceners page.
func Musician(sl *slog.Logger, c *echo.Context, db *sql.DB) error {
	const format = "musician context: %w"
	if err := nils.Check(sl, c, db); err != nil {
		return fmt.Errorf(format, err)
	}

	const title = "Musicians and composers"

	data := empty(c)
	data["title"] = title
	data["logo"] = title
	data["h1"] = title
	data["noindex"] = true

	ctx := c.Request().Context()
	return scenerPage(ctx, sl, c, db, postgres.Musician, data)
}

// New is the handler for the what is new page.
func New(sl *slog.Logger, c *echo.Context) error {
	if err := nils.Check(sl, c); err != nil {
		return fmt.Errorf("new context: %w", err)
	}

	const title = "New stuff"
	const descr = `What is new on the Defacto2 website?`
	const leadr = `This quaint page does not appeal to algorithms, so who will see it?`

	data := empty(c)
	data["noindex"] = true // apply noindex to what's new, so we don't have to worry using about <a href rel="nofollow">
	data["description"] = descr
	data["logo"] = title
	data["h1"] = `What is new?`
	data["lead"] = leadr
	data["title"] = title
	data["carousel"] = "#carouselWhatsNew"

	err := c.Render(http.StatusOK, newx, data)
	if err != nil {
		return InternalErr(sl, c, newx, err)
	}
	return nil
}

func EditFn(sl *slog.Logger, c *echo.Context, db *sql.DB,
	fn func(context.Context, boil.ContextExecutor, int64, string) error,
) error {
	const format = "editfn context %s: %w"
	if err := nils.Check(sl, c, db); err != nil {
		return fmt.Errorf(format, "check", err)
	}

	var f Form
	if err := c.Bind(&f); err != nil {
		return badRequest(c, err)
	}

	ctx := c.Request().Context()
	r, err := model.One(ctx, db, true, f.ID)
	if err != nil {
		return fmt.Errorf(format, strconv.Itoa(f.ID), err)
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return badRequest(c, err)
	}
	if err = fn(ctx, tx, int64(f.ID), f.Value); err != nil {
		return badRequest(c, err)
	}
	if err = tx.Commit(); err != nil {
		return badRequest(c, err)
	}

	return c.JSON(http.StatusOK, r)
}

// PlatformEdit handles the post submission for the Platform selection field.
func PlatformEdit(sl *slog.Logger, c *echo.Context, db *sql.DB) error {
	return EditFn(sl, c, db, model.Platform.Update)
}

// TagEdit handles the post submission for the Tag selection field.
func TagEdit(sl *slog.Logger, c *echo.Context, db *sql.DB) error {
	return EditFn(sl, c, db, model.Section.Update)
}

// PlatformTagInfo handles the POST submission for the platform and tag info.
func PlatformTagInfo(c *echo.Context) error {
	const format = "platform tag info: %w"
	if err := nils.Check(c); err != nil {
		return fmt.Errorf(format, err)
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
func PostDesc(sl *slog.Logger, c *echo.Context, db *sql.DB, input string) error {
	const format = "post desc context: %w"
	if err := nils.Check(sl, c, db); err != nil {
		return fmt.Errorf(format, err)
	}

	if s := strings.TrimSpace(input); s == "" {
		// Redirect back to the search form
		return c.Redirect(http.StatusFound, "/search/desc")
	}

	errs := "post desc search for, " + input
	inputs := helper.SearchTerm(input)
	// Clean terms by removing empty strings from trailing commas
	terms := make([]string, 0, len(inputs))
	for _, term := range inputs {
		if trimmed := strings.TrimSpace(term); trimmed != "" {
			terms = append(terms, trimmed)
		}
	}

	ctx := c.Request().Context()
	fs, _ := model.OnlyDescriptions(ctx, sl, db, terms)
	d := Descriptions.postStats(ctx, db, terms)
	s := strings.Join(terms, ", ")

	data := emptyFiles(c)
	const brief = "Game or app titles"
	data["title"] = brief + " results"
	data["h1"] = brief + " search"
	data["lead"] = "Results for " + s
	data["logo"] = s + " results"
	data["description"] = brief + " search results for " + s + "."
	data["unknownYears"] = false
	data[records] = fs
	data["stats"] = d

	err := c.Render(http.StatusOK, artifacts, data)
	if err != nil {
		return InternalErr(sl, c, errs, err)
	}
	return nil
}

// PostFilename is the handler for the Search for filenames form post page.
func PostFilename(sl *slog.Logger, c *echo.Context, db *sql.DB) error {
	return PostName(sl, c, db, Filenames)
}

// PostName is the handler for the Search for filenames form post page.
func PostName(sl *slog.Logger, c *echo.Context, db *sql.DB, mode FileSearch) error {
	const format = "post name context: %w"
	if err := nils.Check(sl, c, db); err != nil {
		return fmt.Errorf(format, err)
	}

	const title = "Filename results"
	const h1 = "Filename search"
	input := c.FormValue("search-term-query")
	if s := strings.TrimSpace(input); s == "" {
		// Redirect back to the search form
		return c.Redirect(http.StatusFound, "/search/file")
	}
	errs := fmt.Sprint("post name search for,", mode)
	inputs := helper.SearchTerm(input)
	// Clean terms by removing empty strings from trailing commas
	terms := make([]string, 0, len(inputs))
	for _, term := range inputs {
		if trimmed := strings.TrimSpace(term); trimmed != "" {
			terms = append(terms, trimmed)
		}
	}

	ctx := c.Request().Context()
	fs, _ := model.OnlyFilenames(ctx, db, terms)
	d := mode.postStats(ctx, db, terms)
	s := strings.Join(terms, ", ")

	data := emptyFiles(c)
	data["title"] = title
	data["h1"] = h1
	data["lead"] = "Results for " + s
	data["logo"] = s + " results"
	data["description"] = "Filename search results for " + s + "."
	data["unknownYears"] = false
	data[records] = fs
	data["stats"] = d

	err := c.Render(http.StatusOK, artifacts, data)
	if err != nil {
		return InternalErr(sl, c, errs, err)
	}
	return nil
}

// postStats is a helper function for PostName that returns the statistics for the files page.
func (mode FileSearch) postStats(ctx context.Context, db *sql.DB, terms []string) map[string]string {
	none := func() map[string]string {
		// TODO: log errors
		return map[string]string{files: "no files found", years: ""}
	}
	if len(terms) == 0 {
		return none()
	}
	if err := nils.Check(ctx, db); err != nil {
		return none()
	}

	// trim whitespace and recheck
	empty := true
	for _, term := range terms {
		if strings.TrimSpace(term) != "" {
			empty = false
			break
		}
	}
	if empty {
		return none()
	}

	m := model.Summary{
		SumBytes: sql.NullInt64{Int64: 0, Valid: false},
		SumCount: sql.NullInt64{Int64: 0, Valid: false},
		MinYear:  sql.NullInt16{Int16: 0, Valid: false},
		MaxYear:  sql.NullInt16{Int16: 0, Valid: false},
	}
	switch mode {
	case Filenames:
		if err := m.ByFilename(ctx, db, terms); err != nil {
			return none()
		}
	case Descriptions:
		if err := m.ByDescription(ctx, db, terms); err != nil {
			return none()
		}
	default:
		// Handle unknown search modes
		return none()
	}
	if m.SumCount.Int64 == 0 {
		return none()
	}

	d := map[string]string{
		files: string(ByteFileS("file", m.SumCount.Int64, m.SumBytes.Int64)),
		years: helper.Years(m.MinYear.Int16, m.MaxYear.Int16),
	}
	return d
}

// PouetCache parses the cached data for the Pouet production votes.
// If the cache is valid it is returned as JSON response.
// If the cache is invalid or corrupt an error will be returned
// and a API request should be made to Pouet.
func PouetCache(c *echo.Context, data string) error {
	const msg = "pouet cache context"
	if err := nils.Check(c); err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}
	if data == "" {
		return nil
	}

	pv := pouet.Votes{ID: 0, Stars: 0.0, VotesAvg: 0.0, VotesUp: 0, VotesMeh: 0, VotesDown: 0}
	votes := strings.Split(data, sep)
	const req = 4
	if len(votes) < req {
		return fmt.Errorf("%s: %d, want %d: %w", msg, len(votes), req, ErrCorrupt)
	}

	const format = msg + "%s: %w"
	stars, err := strconv.ParseFloat(votes[0], 64)
	if err != nil {
		return fmt.Errorf(format, votes[0], err)
	}

	down, err := strconv.Atoi(votes[1])
	if err != nil {
		return fmt.Errorf(format, votes[1], err)
	}

	up, err := strconv.Atoi(votes[2])
	if err != nil {
		return fmt.Errorf(format, votes[2], err)
	}

	meh, err := strconv.Atoi(votes[3])
	if err != nil {
		return fmt.Errorf(format, votes[3], err)
	}

	pv.Stars = stars
	pv.VotesDown = uint64(math.Abs(float64(down)))
	pv.VotesUp = uint64(math.Abs(float64(up)))
	pv.VotesMeh = uint64(math.Abs(float64(meh)))

	if err = c.JSON(http.StatusOK, pv); err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return nil
}

// ProdPouet is the handler for the Pouet prod JSON page.
func ProdPouet(c *echo.Context, id string) error {
	const format = "prod pouet context: %w"
	if err := nils.Check(c); err != nil {
		return fmt.Errorf(format, err)
	}

	p := pouet.Production{}
	i, err := strconv.Atoi(id)
	if err != nil {
		return c.String(http.StatusNotFound, err.Error())
	}

	ctx := c.Request().Context()
	if _, err = p.Get(ctx, i); err != nil {
		return c.String(http.StatusNotFound, err.Error())
	}
	if err = c.JSON(http.StatusOK, p); err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	return nil
}

// ProdZoo is the handler for the Demozoo production JSON page.
func ProdZoo(c *echo.Context, id string) error {
	const format = "prod zoo context: %w"
	if err := nils.Check(c); err != nil {
		return fmt.Errorf(format, err)
	}

	prod := demozoo.Production{}
	i, err := strconv.Atoi(id)
	if err != nil {
		return c.String(http.StatusNotFound, err.Error())
	}

	ctx := c.Request().Context()
	if code, err := prod.Get(ctx, i); err != nil {
		return c.String(code, err.Error())
	}

	if err = c.JSON(http.StatusOK, prod); err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	return nil
}

// Releasers is the handler for the releaser page ordered by the most files.
func Releasers(sl *slog.Logger, c *echo.Context, db *sql.DB) error {
	return releasers(c.Request().Context(), sl, c, db, model.Prolific)
}

// ReleasersAZ is the handler for the releaser page ordered alphabetically.
func ReleasersAZ(sl *slog.Logger, c *echo.Context, db *sql.DB) error {
	return releasers(c.Request().Context(), sl, c, db, model.Alphabetical)
}

// ReleasersYear is the handler for the releaser page ordered by year of the first release.
func ReleasersYear(sl *slog.Logger, c *echo.Context, db *sql.DB) error {
	return releasers(c.Request().Context(), sl, c, db, model.Oldest)
}

// releasers is the handler for the Releaser page.
func releasers(ctx context.Context, sl *slog.Logger, c *echo.Context, db *sql.DB, orderBy model.OrderBy) error {
	const format = "releaser context: %w"
	if err := nils.Check(ctx, sl, c, db); err != nil {
		return fmt.Errorf(format, err)
	}

	const title = "Releaser"
	const descr = "A linked list of the former Scene releasers and groups, that were collectives of people " +
		"who would work together and operate under a common brand."
	const leadr = "A releaser is a brand or a collective group of " +
		"sceners responsible for releasing or distributing products."
	const logo = "Former groups and releasers"
	const key = "releasers"

	data := empty(c)
	data["noindex"] = true
	data["title"] = title + "s and groups"
	data["description"] = descr
	data["logo"] = logo
	data["h1"] = title
	data["lead"] = leadr
	data["itemName"] = "file"
	data[key] = model.Releasers{}
	data["stats"] = map[string]string{}

	var r model.Releasers
	if err := orderBy.Limit(ctx, db, &r, 0, 0); err != nil {
		return DatabaseErr(sl, c, releaserx, err)
	}

	data[key] = r
	tmpl := releaserx
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
		order = byYear
	default:
		// Handle unknown order types
		order = alpha
	}

	data["stats"] = map[string]string{
		pubs:   fmt.Sprintf("%d releasers and groups", len(r)),
		ordrby: order,
	}

	err := c.Render(http.StatusOK, tmpl, data)
	if err != nil {
		return InternalErr(sl, c, tmpl, err)
	}
	return nil
}

// Releaser is the handler for the list and preview of files credited to a releaser.
func Releaser(sl *slog.Logger, c *echo.Context, db *sql.DB, uri string, public fs.FS) error {
	const msg = "releasers context handler"
	const format = msg + ": %w"
	if err := nils.Check(sl, c, db); err != nil {
		return fmt.Errorf(format, err)
	}

	const code = http.StatusNotFound
	ctx := c.Request().Context()
	fs, err := model.ReleasersWhere(ctx, db, uri)
	if err != nil {
		logErr(sl, msg+" cannot lookup releaser", uri, code, err)
		return ReleaserErr(sl, c, uri)
	}
	if len(fs) == 0 {
		return ReleaserErr(sl, c, uri)
	}

	relname := releaser.Link(uri)
	data := emptyFiles(c)
	data["title"] = relname + " artifacts"
	data[canonical] = strings.Join([]string{"g", uri}, "/")
	data["h1"] = relname
	altnames := lism.String(lism.Path(uri))
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
		logErr(sl, msg+" cannot lookup releaser stats", uri, code, err)
		return ReleaserErr(sl, c, uri)
	}

	data = releasersDesc(relname, altnames, d, data)

	err = c.Render(http.StatusOK, artifacts, data)
	if err != nil {
		return InternalErr(sl, c, "releasers page for, "+uri, err)
	}
	return nil
}

// releasersDesc appends stats to the description.
func releasersDesc(relname, altnames string, d map[string]string, data map[string]any,
) map[string]any {
	data["stats"] = d
	// append stats to the description
	dfiles := d["sum"]
	dyears := d[years]

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

func tibits(sl *slog.Logger, uri string, public fs.FS) string {
	if sl == nil {
		sl = logs.Discard()
	}

	var htm strings.Builder

	tidbits := tidbit.Find(uri)
	slices.Sort(tidbits)
	for value := range slices.Values(tidbits) {
		s := value.String(sl, public)
		if strings.HasSuffix(strings.TrimSpace(s), "</p>") {
			htm.WriteString(`<li class="list-group-item">` + s + string(value.URL(uri)) + `</li>`)
			continue
		}
		htm.WriteString(`<li class="list-group-item">` + s + `<br>` + string(value.URL(uri)) + `</li>`)
	}

	return htm.String()
}

func releaserLead(uri string, data map[string]any) map[string]any {
	switch uri {
	case "independent":
		data["lead"] = lism.String(lism.Path(uri)) +
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
	const format = "releaser sum %q: %w"
	if err := nils.Check(ctx, exec); err != nil {
		return nil, fmt.Errorf(format, uri, err)
	}

	m := model.Summary{
		SumBytes: sql.NullInt64{Int64: 0, Valid: false},
		SumCount: sql.NullInt64{Int64: 0, Valid: false},
		MinYear:  sql.NullInt16{Int16: 0, Valid: false},
		MaxYear:  sql.NullInt16{Int16: 0, Valid: false},
	}
	if err := m.ByReleaser(ctx, exec, uri); err != nil {
		return nil, fmt.Errorf(format, uri, err)
	}

	d := map[string]string{
		files: string(ByteFileS("file", m.SumCount.Int64, m.SumBytes.Int64)),
		years: helper.Years(m.MinYear.Int16, m.MaxYear.Int16),
		"sum": strconv.Itoa(int(m.SumCount.Int64)),
	}
	return d, nil
}

// Scener is the handler for the page to list all the sceners.
func Scener(sl *slog.Logger, c *echo.Context, db *sql.DB) error {
	const format = "scener context: %w"
	if err := nils.Check(sl, c, db); err != nil {
		return fmt.Errorf(format, err)
	}

	const title = "Sceners, people who were apart of the Scene"
	data := empty(c)
	data["title"] = title
	data["logo"] = title
	data["h1"] = title

	ctx := c.Request().Context()
	return scenerPage(ctx, sl, c, db, postgres.Roles(), data)
}

// Sceners is the handler for the list and preview of files credited to a scener.
func Sceners(sl *slog.Logger, c *echo.Context, db *sql.DB, uri string) error {
	if err := nils.Check(sl, c, db); err != nil {
		return fmt.Errorf("sceners context: %w", err)
	}

	errs := "sceners page for, " + uri
	s := releaser.Link(uri)
	var ms model.Scener
	ctx := c.Request().Context()
	fs, err := ms.Where(ctx, db, uri)
	if err != nil {
		return InternalErr(sl, c, errs, err)
	}
	if len(fs) == 0 {
		return ScenerErr(sl, c, uri)
	}

	data := emptyFiles(c)
	data[canonical] = strings.Join([]string{"p", uri}, "/")
	data["title"] = s + attr
	data["h1"] = s
	data["lead"] = "Artifacts attributed to " + s + "."
	data["logo"] = s
	data["description"] = "These are the documented artifacts attributed to the person known as " + s + "."
	data["scener"] = s
	data[records] = fs

	d, err := scenerSum(ctx, db, uri)
	if err != nil {
		return InternalErr(sl, c, errs, err)
	}
	data["stats"] = d

	err = c.Render(http.StatusOK, artifacts, data)
	if err != nil {
		return InternalErr(sl, c, errs, err)
	}
	return nil
}

// scenerSum is a helper function for Sceners that returns the statistics for the files page.
func scenerSum(ctx context.Context, exec boil.ContextExecutor, uri string) (map[string]string, error) {
	const format = "scener sum %q: %w"
	if err := nils.Check(ctx, exec); err != nil {
		return nil, fmt.Errorf(format, uri, err)
	}

	m := model.Summary{
		SumBytes: sql.NullInt64{Int64: 0, Valid: false},
		SumCount: sql.NullInt64{Int64: 0, Valid: false},
		MinYear:  sql.NullInt16{Int16: 0, Valid: false},
		MaxYear:  sql.NullInt16{Int16: 0, Valid: false},
	}
	if err := m.ByScener(ctx, exec, uri); err != nil {
		return nil, fmt.Errorf(format, uri, err)
	}

	d := map[string]string{
		files: string(ByteFileS("file", m.SumCount.Int64, m.SumBytes.Int64)),
		years: helper.Years(m.MinYear.Int16, m.MaxYear.Int16),
	}
	return d, nil
}

// SearchDesc is the handler for the Search for file descriptions page.
func SearchDesc(sl *slog.Logger, c *echo.Context) error {
	if err := nils.Check(c, sl); err != nil {
		return fmt.Errorf("search desc context: %w", err)
	}

	data := empty(c)
	data["noindex"] = true
	data["description"] = "Use this search to uncover named applications, games, and descriptions of artifacts."
	data["logo"] = search
	data["title"] = "Game or app titles search"

	err := c.Render(http.StatusOK, searchpost, data)
	if err != nil {
		return InternalErr(sl, c, searchpost, err)
	}
	return nil
}

// SearchID is the handler for the Record by ID Search page.
func SearchID(sl *slog.Logger, c *echo.Context) error {
	if err := nils.Check(c, sl); err != nil {
		return fmt.Errorf("search id context: %w", err)
	}

	const title = "Search the IDs of artifacts"

	data := empty(c)
	data["noindex"] = true
	data["description"] = "Use this search to lookup artifacts by their database identities."
	data["logo"] = title
	data["title"] = title
	data["info"] = "search for artifacts by their record id, uuid or URL key"
	data["hxPost"] = "/editor/search/id"
	data["inputPlaceholder"] = "Type to search for an artifact…"

	err := c.Render(http.StatusOK, searchhtmx, data)
	if err != nil {
		return InternalErr(sl, c, searchhtmx, err)
	}
	return nil
}

// SearchFile is the handler for the Search for files page.
func SearchFile(sl *slog.Logger, c *echo.Context) error {
	if err := nils.Check(c, sl); err != nil {
		return fmt.Errorf("search file context: %w", err)
	}

	data := empty(c)
	data["noindex"] = true
	data["description"] = "Use this search to lookup artifacts by their filenames."
	data["logo"] = search
	data["title"] = "Filename or extensions search"

	err := c.Render(http.StatusOK, searchpost, data)
	if err != nil {
		return InternalErr(sl, c, searchpost, err)
	}
	return nil
}

// SearchReleaser is the handler for the Releaser Search page.
func SearchReleaser(sl *slog.Logger, c *echo.Context) error {
	if err := nils.Check(c, sl); err != nil {
		return fmt.Errorf("search releaser context: %w", err)
	}

	data := empty(c)
	data["noindex"] = true
	data["description"] = "Lookup groups, magazines, bbs boards, and ftp sites, by their names."
	data["logo"] = search
	data["title"] = "Lookup releasers"
	data["info"] = "find groups, names, magazines, bbs boards, ftp sites"
	data["helpText"] = "acronyms only match exact finds: 'rc' and 'rcn' are treated different"
	data["hxPost"] = "/search/releaser"
	data["inputPlaceholder"] = "Type to search for a releaser…"

	err := c.Render(http.StatusOK, searchhtmx, data)
	if err != nil {
		return InternalErr(sl, c, searchhtmx, err)
	}
	return nil
}

// SignedOut is the handler to sign out and remove the current session.
func SignedOut(sl *slog.Logger, c *echo.Context) error {
	if err := nils.Check(c, sl); err != nil {
		return fmt.Errorf("signed out context: %w", err)
	}
	{ // get any existing session
		sess, err := session.Get(sess.Name, c)
		if err != nil {
			return BadRequestErr(sl, c, signedout, err)
		}

		id, ok := sess.Values["sub"]
		if !ok || id == "" {
			return ForbiddenErr(sl, c, signedout, ErrSession)
		}

		const remove = -1
		sess.Options.MaxAge = remove
		err = sess.Save(c.Request(), c.Response())
		if err != nil {
			return InternalErr(sl, c, signedout, err)
		}
	}

	return c.Redirect(http.StatusFound, "/")
}

// SignOut is the handler for the Sign out of Defacto2 page.
func SignOut(sl *slog.Logger, c *echo.Context) error {
	if err := nils.Check(c, sl); err != nil {
		return fmt.Errorf("sign out context: %w", err)
	}

	const title = "Sign out"
	data := empty(c)
	data["noindex"] = true
	data["title"] = title
	data["description"] = "Sign out of Defacto2."
	data["h1"] = title

	err := c.Render(http.StatusOK, signout, data)
	if err != nil {
		return InternalErr(sl, c, signout, err)
	}
	return nil
}

// Signin is the handler for the Sign in session page.
func Signin(sl *slog.Logger, c *echo.Context, clientID string, nonce []byte) error {
	if err := nils.Check(c, sl); err != nil {
		return fmt.Errorf("signin context: %w", err)
	}

	const title = "Sign in"
	data := empty(c)
	data["noindex"] = true
	data["title"] = title
	data["description"] = "Sign in to Defacto2."
	data["h1"] = title
	data["lead"] = "This is not open to the general public."
	data["callback"] = "/google/callback"
	data["clientID"] = clientID
	data["nonce"] = string(nonce)
	{ // get any existing session
		sess, err := session.Get(sess.Name, c)
		if err != nil {
			return expireCookie(sl, c, signin, data)
		}

		id, ok := sess.Values["sub"]
		if !ok {
			return expireCookie(sl, c, signin, data)
		}
		val, find := id.(string)
		if find && val != "" {
			return SignOut(sl, c)
		}
	}

	err := c.Render(http.StatusOK, signin, data)
	if err != nil {
		return InternalErr(sl, c, signin, err)
	}
	return nil
}

// expireCookie is a helper function to remove the session cookie by setting the MaxAge to -1.
func expireCookie(sl *slog.Logger, c *echo.Context, name string, data map[string]any) error {
	if err := nils.Check(sl, c); err != nil {
		return fmt.Errorf("context remove cookie: %w", err)
	}

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
		return InternalErr(sl, c, name, err)
	}
	return nil
}

// TagInfo handles the POST submission for the platform and tag info.
func TagInfo(c *echo.Context) error {
	if err := nils.Check(c); err != nil {
		return fmt.Errorf("tag info context: %w", err)
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
func Titles(sl *slog.Logger, c *echo.Context) error {
	if err := nils.Check(c, sl); err != nil {
		return fmt.Errorf("titles context: %w", err)
	}

	const title = "Titles"
	data := empty(c)
	data["title"] = title
	data["description"] = "Titles are important."
	data["logo"] = title
	data["h1"] = "Titles and naming are important"

	err := c.Render(http.StatusOK, titles, data)
	if err != nil {
		return InternalErr(sl, c, titles, err)
	}
	return nil
}

// Fixers is the handler for the editor, batch-fixers page.
func Fixers(sl *slog.Logger, c *echo.Context, db *sql.DB) error {
	if err := nils.Check(sl, c, db); err != nil {
		return fmt.Errorf("fixers context: %w", err)
	}

	const title = "Fixers"
	data := empty(c)
	data["description"] = "Defacto2 fixers tool."
	data["h1"] = title
	data["lead"] = "Artifact fixes using batch-friendly tools."
	data["title"] = title

	// Get files with numeric suffixes
	ctx := c.Request().Context()
	filesFix, err := fix.NumSuffix(ctx, db)
	if err != nil {
		sl.Error("failed to get files with numeric suffixes", slog.String("error", err.Error()))
		// Don't return error, just continue without the data
	} else {
		data["numericSuffixCount"] = filesFix.Count
		// Add obfuscated IDs for the /f/ route
		for i := range filesFix.Files {
			filesFix.Files[i].ObfuscatedID = helper.ObfuscateID(filesFix.Files[i].ID)
		}
		// Pass the full file data (including ID, UUID, and obfuscated ID) to the template
		data["numericSuffixFiles"] = filesFix.Files
	}

	err = c.Render(http.StatusOK, fixers, data)
	if err != nil {
		return InternalErr(sl, c, fixers, err)
	}
	return nil
}

var numericSuffixRegexes = []*regexp.Regexp{
	regexp.MustCompile(` \([0-9]{1,3}\)`), // Pattern with space: " (123)"
	regexp.MustCompile(`\([0-9]{1,3}\)`),  // Pattern without space: "(123)"
}

// FixNumericSuffix handles the fixing of numeric suffixes in filenames.
func FixNumericSuffix(sl *slog.Logger, c *echo.Context, db *sql.DB) error {
	if err := nils.Check(sl, c, db); err != nil {
		return fmt.Errorf("fix numeric suffix: %w", err)
	}

	obfuscatedID := c.Param("id")
	if obfuscatedID == "" {
		return fmt.Errorf("missing obfuscated ID: %w", ErrMissingObfuscatedID)
	}
	fileID := int64(helper.DeobfuscateID(obfuscatedID))
	if fileID == 0 {
		return fmt.Errorf("invalid obfuscated ID %s: %w", obfuscatedID, ErrInvalidObfuscatedID)
	}

	// Get the file from the database using the standard One function
	// Use deleted=true to allow fixing soft-deleted files
	ctx := c.Request().Context()
	file, err := model.One(ctx, db, true, int(fileID))
	if err != nil {
		logID(sl, "failure to find file in fix handler", obfuscatedID, fileID, err)
		return fmt.Errorf("failed to find file %s: %w", obfuscatedID, err)
	}
	if file == nil {
		logID(sl, "failure to find file in fix handler", obfuscatedID, fileID, ErrFileNotFound)
		return fmt.Errorf("file not found: %w", ErrFileNotFound)
	}

	// Remove the numeric suffix from the filename
	originalFilename := file.Filename.String
	// Try both patterns: with space and without space before the parenthesis
	var matched bool
	baseFilename := originalFilename
	for _, regex := range numericSuffixRegexes {
		if regex.MatchString(originalFilename) {
			baseFilename = regex.ReplaceAllString(originalFilename, "")
			matched = true
			break
		}
	}

	if !matched {
		return fmt.Errorf("filename does not match numeric suffix pattern: %s: %w",
			originalFilename, ErrInvalidFilenamePattern)
	}

	// Update the filename in the database
	file.Filename = null.StringFrom(baseFilename)
	_, err = file.Update(ctx, db, boil.Infer())
	if err != nil {
		return fmt.Errorf("failed to update filename %s: %w", baseFilename, err)
	}

	// Return the updated file info as HTML to replace the list item
	obfuscatedID = helper.ObfuscateID(file.ID)
	id := strconv.Itoa(int(file.ID))
	return c.HTML(http.StatusOK, `
		<div class="list-group-item list-group-item-success">
			<div class="d-flex justify-content-between align-items-center">
				<div>
					<code>`+baseFilename+`</code>
					<small class="text-muted d-block">
						ID: `+id+` | UUID: `+file.UUID.String+`
					</small>
					<small class="text-success d-block">
						Fixed: `+originalFilename+` → `+baseFilename+`
					</small>
				</div>
				<div>
					<a href="/f/`+obfuscatedID+`" class="btn btn-sm btn-outline-secondary" target="_blank">View</a>
				</div>
			</div>
		</div>
		`)
}

// Thanks is the handler for the Thanks page.
func Thanks(sl *slog.Logger, c *echo.Context) error {
	if err := nils.Check(c, sl); err != nil {
		return fmt.Errorf("thanks context: %w", err)
	}

	const title = "Thank you!"
	data := empty(c)
	data["description"] = "A special thanks to the hundreds of contributors and the thousands of contributions."
	data["h1"] = title
	data["lead"] = "Thanks to the hundreds of people who have contributed to" +
		" Defacto2 over the decades with file submissions, " +
		"hard drive donations, interviews, corrections, artwork, and monetary contributions!"
	data["title"] = title

	err := c.Render(http.StatusOK, thanks, data)
	if err != nil {
		return InternalErr(sl, c, thanks, err)
	}
	return nil
}

// TheScene is the handler for the The Scene page.
func TheScene(sl *slog.Logger, c *echo.Context) error {
	if err := nils.Check(c, sl); err != nil {
		return fmt.Errorf("the scene context: %w", err)
	}

	const title = "The Scene"
	data := empty(c)
	data["description"] = "A short introduction on The Scene, the online subcultures and its underground origins."
	data["logo"] = "The underground"
	data["h1"] = title
	data["lead"] = "The Scene is broad church of people and online communities that is collectively grouped;" +
		" it is subculture of niche activities using personal computers, " +
		"where the participants share creations and exchange ideas."
	data["title"] = title

	err := c.Render(http.StatusOK, thescene, data)
	if err != nil {
		return InternalErr(sl, c, thescene, err)
	}
	return nil
}

// VotePouet is the handler for the Pouet production votes JSON page.
func VotePouet(sl *slog.Logger, c *echo.Context, id string) error {
	if err := nils.Check(c, sl); err != nil {
		return fmt.Errorf("vote pouet context: %w", err)
	}

	const (
		title = "Pouet"
		sep   = ";"
	)
	pv := pouet.Votes{ID: 0, Stars: 0.0, VotesAvg: 0.0, VotesUp: 0, VotesMeh: 0, VotesDown: 0}
	i, err := strconv.Atoi(id)
	if err != nil {
		return c.String(http.StatusNotFound, err.Error())
	}

	cp := cache.PouetVote
	if s, err := cp.Read(id); err == nil {
		if err := PouetCache(c, s); err == nil {
			sl.Debug("vote.pouet", slog.String("cache.hit.id", id))
			return nil
		}
	}
	sl.Debug("vote.pouet", slog.String("cache.miss.for.pouet.id", id))

	ctx := c.Request().Context()
	if err = pv.Votes(ctx, i); err != nil {
		return c.String(http.StatusNotFound, err.Error())
	}
	if err = c.JSON(http.StatusOK, pv); err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	value := fmt.Sprintf("%.1f%s%d%s%d%s%d", pv.Stars, sep, pv.VotesDown, sep, pv.VotesUp, sep, pv.VotesMeh)
	if err := cp.Write(id, value, cache.ExpiredAt); err != nil {
		logID(sl, "vote pouet cache failure", id, 0, err)
	}
	return nil
}

// Website is the handler for the websites page.
// Open is the ID of the accordion section to open.
func Website(sl *slog.Logger, c *echo.Context, open string) error {
	if err := nils.Check(c, sl); err != nil {
		return fmt.Errorf("website context: %w", err)
	}

	const title = "Websites"
	const logo = "Videos, Books, Films, Sites, Podcasts"
	data := empty(c)
	data["title"] = title
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
		data["title"] = title + " categories"
	}

	// If a section was requested but not found, return a 404.
	if open != "hide" && closeAll {
		return StatusErr(sl, c, http.StatusNotFound, open)
	}

	// Render the page.
	data["accordion"] = accordion
	err := c.Render(http.StatusOK, websites, data)
	if err != nil {
		return InternalErr(sl, c, "render open website, "+open, err)
	}
	return nil
}

// Writer is the handler for the Writer page.
func Writer(sl *slog.Logger, c *echo.Context, db *sql.DB) error {
	if err := nils.Check(sl, c, db); err != nil {
		return fmt.Errorf("writer context: %w", err)
	}

	const title = "Writers, editors, and authors"
	data := empty(c)
	data["title"] = title
	data["logo"] = title
	data["h1"] = title
	data["noindex"] = true

	ctx := c.Request().Context()
	return scenerPage(ctx, sl, c, db, postgres.Writer, data)
}

// stats is a helper function for Artifacts that returns the statistics for the files page.
func stats(ctx context.Context, exec boil.ContextExecutor, uri string) (map[string]string, int, error) {
	const format = `context artifacts stats %q: %w`
	if err := nils.Check(ctx, exec); err != nil {
		return nil, 0, fmt.Errorf(format, uri, err)
	}

	if !fileslice.Valid(uri) {
		return nil, 0, nil
	}

	m := model.Summary{
		SumBytes: sql.NullInt64{Int64: 0, Valid: false},
		SumCount: sql.NullInt64{Int64: 0, Valid: false},
		MinYear:  sql.NullInt16{Int16: 0, Valid: false},
		MaxYear:  sql.NullInt16{Int16: 0, Valid: false},
	}
	err := m.ByMatch(ctx, exec, uri)
	if err != nil && !errors.Is(err, model.ErrURI) {
		return nil, 0, fmt.Errorf(format, uri, err)
	}
	if errors.Is(err, model.ErrURI) {
		if err := statsByURI(ctx, exec, uri, &m); err != nil {
			return nil, 0, err
		}
	}

	d := map[string]string{
		files: string(ByteFileS("file", m.SumCount.Int64, m.SumBytes.Int64)),
		years: strconv.Itoa(int(m.MinYear.Int16)) + " - " + strconv.Itoa(int(m.MaxYear.Int16)),
	}
	return d, int(m.SumCount.Int64), nil
}

func steps(lastPage float64) int {
	const (
		one        = 1
		two        = 2
		four       = 4
		skip2Pages = 39
		skip4Pages = 99
	)
	switch {
	case lastPage > skip4Pages:
		return four
	case lastPage > skip2Pages:
		return two
	default:
		return one
	}
}

// statsByURI handles the different URI types for statistics calculation.
func statsByURI(ctx context.Context, exec boil.ContextExecutor, uri string, m *model.Summary) error {
	const format = "artifacts stats uri %s %q: %w"
	if err := nils.Check(ctx, exec, m); err != nil {
		return fmt.Errorf(format, "check", uri, err)
	}

	switch fileslice.Match(uri) {
	case fileslice.ForApproval:
		if err := m.ByForApproval(ctx, exec); err != nil {
			return fmt.Errorf(format, "by for approval", uri, err)
		}
	case fileslice.Deletions:
		if err := m.ByHidden(ctx, exec); err != nil {
			return fmt.Errorf(format, "by hidden", uri, err)
		}
	case fileslice.Unwanted:
		if err := m.ByUnwanted(ctx, exec); err != nil {
			return fmt.Errorf(format, "by unwanted", uri, err)
		}
	case
		fileslice.NewUploads,
		fileslice.NewUpdates,
		fileslice.Oldest,
		fileslice.Newest,
		fileslice.Sensenstahl:
		// For these cases, use the public artifacts method as fallback
		if err := m.ByPublic(ctx, exec); err != nil {
			return fmt.Errorf(format, "by public fallback", uri, err)
		}
	default:
		if err := m.ByPublic(ctx, exec); err != nil {
			return fmt.Errorf(format, "by public", uri, err)
		}
	}

	return nil
}

// emptyFiles is a map of default values specific to the files templates.
func emptyFiles(c *echo.Context) map[string]any {
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
