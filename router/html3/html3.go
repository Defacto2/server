// Package html3 handles the routes and views for the retro,
// mini-website that is rendered in HTML 3 syntax.
package html3

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/Defacto2/sceners"
	"github.com/Defacto2/server/helpers"
	"github.com/Defacto2/server/models"
	"github.com/Defacto2/server/tags"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"

	"github.com/Defacto2/server/postgres"
	pgm "github.com/Defacto2/server/postgres/models"
)

// HTTP status codes in Go
// https://go.dev/src/net/http/status.go

var ErrByTag = errors.New("unknown bytag record group")

// RecordsBy are the record groupings.
type RecordsBy int

const (
	BySection   RecordsBy = iota // BySection groups records by the section file table column.
	ByPlatform                   // BySection groups records by the platform file table column.
	ByGroup                      // ByGroup groups the records by the distinct, group_brand_for file table column.
	AsArt                        // AsArt group records as art.
	AsDocuments                  // AsDocuments group records as documents.
	AsSoftware                   // AsSoftware group records as software.

	asc  = "A" // asc is order by ascending.
	desc = "D" // desc is order by descending.
)

func (t RecordsBy) String() string {
	const l = 6
	if t >= l {
		return ""
	}
	return [l]string{"category", "platform", "group", "art", "document", "software"}[t]
}

func (t RecordsBy) Parent() string {
	const l = 6
	if t >= l {
		return ""
	}
	return [l]string{"categories", "platforms", "groups", "", "", ""}[t]
}

const (
	NameAsc = "C=N&O=A" // Name ascending order.
	NameDes = "C=N&O=D" // Name descending order.
	PublAsc = "C=D&O=A" // Date published ascending order.
	PublDes = "C=D&O=D" // Date published descending order.
	PostAsc = "C=P&O=A" // Posted ascending order.
	PostDes = "C=P&O=D" // Posted descending order.
	SizeAsc = "C=S&O=A" // Size ascending order.
	SizeDes = "C=S&O=D" // Size descending order.
	DescAsc = "C=I&O=A" // Description ascending order.
	DescDes = "C=I&O=D" // Description descending order.
)

const (
	Prefix = "/html3" // Root path of the HTML3 router group.
	title  = "Index of " + Prefix
)

var Counts = models.Counts

type sugared struct {
	log *zap.SugaredLogger
}

// Routes for the /html3 sub-route group.
// Any errors are logged and rendered to the client using HTTP codes
// and the custom /html3, group errror template.
func Routes(e *echo.Echo, log *zap.SugaredLogger) {
	s := sugared{log: log}
	g := e.Group(Prefix)
	g.GET("", s.Index)
	g.GET("/categories", Categories)
	g.GET("/category/:id", s.Category)
	g.GET("/platforms", Platforms)
	g.GET("/platform/:id", s.Platform)
	g.GET("/groups", Groups)
	g.GET("/group/:id", s.Group)
	g.GET("/art", s.Art)
	g.GET("/documents", s.Documents)
	g.GET("/software", s.Software)
	// append legacy redirects
	for url := range LegacyURLs {
		g.GET(url, Redirection)
	}
}

// LegacyURLs are partial URL routers that are to be redirected with a HTTP 308
// permanent redirect status code. These are for retired URL syntaxes that are still
// found on websites online, so their links to Defacto2 do not break with 404, not found errors.
var LegacyURLs = map[string]string{
	"/index":            "",
	"/categories/index": "/categories",
	"/platforms/index":  "/platforms",
}

// Sort is the display name of column that can be used to sort and order the records.
type Sort string

const (
	Name    Sort = "Name"        // Sort records by the filename.
	Publish Sort = "Publish"     // Sort records by the published year, month and day.
	Posted  Sort = "Posted"      // Sort records by the record creation dated.
	Size    Sort = "Size"        // Sort records by the file size in byte units.
	Desc    Sort = "Description" // Sort the records by the title.
)

// Sortings are the name and order of columns that the records can be ordered by.
var Sortings = map[Sort]string{
	Name:    asc,
	Publish: asc,
	Posted:  asc,
	Size:    asc,
	Desc:    asc,
}

// Clauses for ordering file record queries.
func Clauses(query string) models.Order {
	switch strings.ToUpper(query) {
	case NameAsc:
		return models.NameAsc
	case NameDes:
		return models.NameDes
	case PublAsc:
		return models.PublAsc
	case PublDes:
		return models.PublDes
	case PostAsc:
		return models.PostAsc
	case PostDes:
		return models.PostDes
	case SizeAsc:
		return models.SizeAsc
	case SizeDes:
		return models.SizeDes
	case DescAsc:
		return models.DescAsc
	case DescDes:
		return models.DescDes
	default:
		return models.NameAsc
	}
}

const (
	errConn = "Sorry, at the moment the server cannot connect to the database"
	errTag  = "No database query was created for the tag"
	errTmpl = "The server could not render the HTML template for this page"
	firefox = "Welcome to the Firefox 2 era (October 2006) Defacto2 website, which is friendly for legacy operating systems, including Windows 9x, NT-4, and OS-X 10.2."
)

// GroupCache is a cached collection of important, expensive group data.
// The Mu mutex must always be locked before writing this varable.
var IndexCache IndexSums

// GroupCol is a cached collection of important, expensive group data.
// The Mu mutex must always be locked when writing to the Groups map.
type IndexSums struct {
	Mu   sync.Mutex
	Sums map[int]int
}

// Index method is the homepage of the /html3 sub-route.
func (s *sugared) Index(c echo.Context) error {
	start := latency()
	const desc = firefox
	ctx := context.Background()
	db, err := postgres.ConnectDB()
	if err != nil {
		s.log.Warnf("%s: %s", errConn, err)
		return echo.NewHTTPError(http.StatusServiceUnavailable, errConn)
	}
	defer db.Close()

	// Cache for the database counts.
	IndexCache.Mu.Lock()
	defer IndexCache.Mu.Unlock()
	// Get and store database counts.
	if IndexCache.Sums == nil {
		const loop = 4
		IndexCache.Sums = make(map[int]int, loop)
		for i := 0; i < loop; i++ {
			IndexCache.Sums[i] = 0
		}
	}
	for i, value := range IndexCache.Sums {
		if value > 0 {
			continue
		}
		var err error
		sum := 0
		switch i {
		case 0:
			sum, err = models.ArtCount(ctx, db)
		case 1:
			sum, err = models.DocumentCount(ctx, db)
		case 2:
			sum, err = models.SoftwareCount(ctx, db)
		case 3:
			sum, err = models.GroupsTotalCount(ctx, db)
		}
		if err != nil {
			s.log.Warnf("%s: %s", errConn, err)
			continue
		}
		IndexCache.Sums[i] = sum
	}
	return c.Render(http.StatusOK, "index", map[string]interface{}{
		"title":       title,
		"description": desc,
		"art":         IndexCache.Sums[0],
		"doc":         IndexCache.Sums[1],
		"sw":          IndexCache.Sums[2],
		"grp":         IndexCache.Sums[3],
		"cat":         tags.CategoryCount,
		"plat":        tags.PlatformCount,
		"latency":     fmt.Sprintf("%s.", time.Since(*start)),
	})
}

// Categories lists the names, descriptions and sums of the category (section) tags.
func Categories(c echo.Context) error {
	start := latency()
	return c.Render(http.StatusOK, "tag", map[string]interface{}{
		"title":    title + "/categories",
		"latency":  fmt.Sprintf("%s.", time.Since(*start)),
		"path":     "category",
		"tagFirst": tags.FirstCategory,
		"tagEnd":   tags.LastCategory,
		"tags":     tags.Names,
	})
}

// Platforms lists the names, descriptions and sums of the platform tags.
func Platforms(c echo.Context) error {
	start := latency()
	return c.Render(http.StatusOK, "tag", map[string]interface{}{
		"title":    title + "/platforms",
		"latency":  fmt.Sprintf("%s.", time.Since(*start)),
		"path":     "platform",
		"tagFirst": tags.FirstPlatform,
		"tagEnd":   tags.LastPlatform,
		"tags":     tags.Names,
	})
}

// Groups lists the names and sums of all the distinct scene groups.
func Groups(c echo.Context) error {
	start := latency()
	ctx := context.Background()
	db, err := postgres.ConnectDB()
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, errConn)
	}
	defer db.Close()
	total, err := models.GroupsTotalCount(ctx, db)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, errConn)
	}
	// if there is an out of date cache, it will get updated in the background
	// but the client will probably be rendered with an incomplete, stale cache.
	feedback := ""
	models.GroupCache.Mu.RLock()
	l := len(models.GroupCache.Groups)
	models.GroupCache.Mu.RUnlock()
	if l != total {
		go func(err error) error {
			return models.GroupCache.Update()
		}(err)
		if err != nil {
			return echo.NewHTTPError(http.StatusNotFound, errConn)
		}
		feedback = refreshInfo(l, total)
	}
	models.GroupCache.Mu.RLock()
	defer models.GroupCache.Mu.RUnlock()
	return c.Render(http.StatusOK, "groups", map[string]interface{}{
		"feedback": feedback,
		"title":    title + "/groups",
		"latency":  fmt.Sprintf("%s.", time.Since(*start)),
		"path":     "group",
		"sceners":  models.GroupCache.Groups,
	})
}

func refreshInfo(l, total int) string {
	if l == 0 {
		// pause for a second so the client can display some records
		time.Sleep(1 * time.Second)
		return fmt.Sprintf("The list of %d groups is stale and is being updated, please refresh for an updated list.", total)
	}
	return fmt.Sprintf("The list of groups is stale and is being updated."+
		" Only showing %d of %d groups, please refresh for an updated list.", l, total)
}

// Category lists the file records associated with the category tag that is provided by the ID param in the URL.
func (s *sugared) Category(c echo.Context) error {
	return s.Tag(BySection, c)
}

// Platform lists the file records associated with the platform tag that is provided by the ID param in the URL.
func (s *sugared) Platform(c echo.Context) error {
	return s.Tag(ByPlatform, c)
}

// Group lists the file records associated with the group that is provided by the ID param in the URL.
func (s *sugared) Group(c echo.Context) error {
	return s.Tag(ByGroup, c)
}

func (s *sugared) Art(c echo.Context) error {
	return s.Tag(AsArt, c)
}

func (s *sugared) Documents(c echo.Context) error {
	return s.Tag(AsDocuments, c)
}

func (s *sugared) Software(c echo.Context) error {
	return s.Tag(AsSoftware, c)
}

// Tag fetches all the records associated with the RecordsBy grouping.
func (s *sugared) Tag(tt RecordsBy, c echo.Context) error {
	start := latency()
	id := c.Param("id")
	name := sceners.CleanURL(id)
	ctx := context.Background()
	db, err := postgres.ConnectDB()
	if err != nil {
		s.log.Warnf("%s: %s", errConn, err)
		return echo.NewHTTPError(http.StatusServiceUnavailable, errConn)
	}
	defer db.Close()
	var records pgm.FileSlice
	order := Clauses(c.QueryString())
	switch tt {
	case BySection:
		records, err = order.FilesByCategory(id, ctx, db)
	case ByPlatform:
		records, err = order.FilesByPlatform(id, ctx, db)
	case ByGroup:
		records, err = order.FilesByGroup(name, ctx, db)
	case AsArt:
		records, err = order.ArtFiles(ctx, db)
	case AsDocuments:
		records, err = order.DocumentFiles(ctx, db)
	case AsSoftware:
		records, err = order.SoftwareFiles(ctx, db)
	default:
		s.log.Warnf("%s: %s", errTag, tt)
		return echo.NewHTTPError(http.StatusServiceUnavailable, errTag)
	}
	if err != nil {
		s.log.Warnf("%s: %s", errConn, err)
		return echo.NewHTTPError(http.StatusServiceUnavailable, errConn)
	}
	count := len(records)
	if count == 0 {
		return echo.NewHTTPError(http.StatusNotFound,
			fmt.Sprintf("The %s %q doesn't exist", tt, id))
	}
	var byteSum int64
	switch tt {
	case BySection:
		byteSum, err = models.ByteCountByCategory(id, ctx, db)
	case ByPlatform:
		byteSum, err = models.ByteCountByPlatform(id, ctx, db)
	case ByGroup:
		byteSum, err = models.ByteCountByGroup(name, ctx, db)
	case AsArt:
		byteSum, err = models.ArtByteCount(ctx, db)
	case AsDocuments:
		byteSum, err = models.DocumentByteCount(ctx, db)
	case AsSoftware:
		byteSum, err = models.SoftwareByteCount(ctx, db)
	default:
		s.log.Warnf("%s: %s", errTag, tt)
		return echo.NewHTTPError(http.StatusServiceUnavailable, errTag)
	}
	if err != nil {
		s.log.Warnf("%s %s", errConn, err)
		return echo.NewHTTPError(http.StatusServiceUnavailable, errConn)
	}
	desc := ""
	switch tt {
	case BySection, ByPlatform:
		key := tags.TagByURI(id)
		info := tags.Infos[key]
		name := tags.Names[key]
		desc = fmt.Sprintf("%s - %s.", name, info)
	case AsArt:
		desc = "Digital + pixel art, hi-res, raster and pixel images."
	}
	stat := fmt.Sprintf("%d files, %s", count, helpers.ByteCountFloat(byteSum))
	sorter := sorter(c.QueryString())
	err = c.Render(http.StatusOK, tt.String(), map[string]interface{}{
		"title":       fmt.Sprintf("%s%s%s", title, fmt.Sprintf("/%s/", tt), id),
		"home":        "",
		"description": desc,
		"parent":      tt.Parent(),
		"stats":       stat,
		"sort":        sorter,
		"records":     records,
		"latency":     fmt.Sprintf("%s.", time.Since(*start)),
	})
	if err != nil {
		s.log.Errorf("%s: %s %d", errTmpl, err, tt)
		return echo.NewHTTPError(http.StatusInternalServerError, errTmpl)
	}
	return nil
}

// Error renders a custom HTTP error page for the HTML3 sub-group.
func Error(err error, c echo.Context) error {
	// Echo custom error handling: https://echo.labstack.com/guide/error-handling/
	start := latency()
	code := http.StatusInternalServerError
	msg := "This is a server problem"
	if he, ok := err.(*echo.HTTPError); ok {
		code = he.Code
		msg = fmt.Sprint(he.Message)
	}
	// TODO:
	// this function should log error values.

	return c.Render(code, "error", map[string]interface{}{
		"title":       fmt.Sprintf("%d error, there is a complication", code),
		"description": fmt.Sprintf("%s.", msg),
		"latency":     fmt.Sprintf("%s.", time.Since(*start)),
	})
}

// Redirection redirects any legacy URL matches.
func Redirection(c echo.Context) error {
	for u, redirect := range LegacyURLs {
		htm := Prefix + u
		if htm == c.Path() {
			return c.Redirect(http.StatusPermanentRedirect, Prefix+redirect)
		}
	}
	return c.String(http.StatusInternalServerError,
		fmt.Sprintf("unknown redirection, %q ", c.Path()))
}

// 	e.GET("/users/:name", func(c echo.Context) error {
// 		name := c.Param("name")
// 		return c.String(http.StatusOK, name)
//  })
//
//  curl http://localhost:1323/users/Joe

func latency() *time.Time {
	start := time.Now()
	r := new(big.Int)
	const n, k = 1000, 10
	r.Binomial(n, k)
	return &start
}

func sorter(query string) map[string]string {
	s := Sortings
	switch strings.ToUpper(query) {
	case NameAsc:
		s[Name] = desc
	case NameDes:
		s[Name] = asc
	case PublAsc:
		s[Publish] = desc
	case PublDes:
		s[Publish] = asc
	case PostAsc:
		s[Posted] = desc
	case PostDes:
		s[Posted] = asc
	case SizeAsc:
		s[Size] = desc
	case SizeDes:
		s[Size] = asc
	case DescAsc:
		s[Desc] = desc
	case DescDes:
		s[Desc] = asc
	default:
		// When no query is provided, it is assumed the records have been
		// ordered with Name ASC. So set DESC for the clickable Name link.
		s[Name] = desc
	}
	// to be usable in the template, convert the map keys into strings
	tmplSorts := make(map[string]string, len(s))
	for key, value := range Sortings {
		tmplSorts[string(key)] = value
	}
	return tmplSorts
}
