// Package html3 handles the routes and views for the retro,
// mini-website that is rendered in HTML 3 syntax.
package html3

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"sync"
	"time"

	"github.com/Defacto2/server/models"
	"github.com/Defacto2/server/pkg/helpers"
	"github.com/Defacto2/server/pkg/postgres"
	"github.com/Defacto2/server/router/dl"
	"github.com/Defacto2/server/tags"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
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
	Prefix = "/html3" // Root path of the HTML3 router group.

	title   = "Index of " + Prefix
	errConn = "Sorry, at the moment the server cannot connect to the database"
	errTag  = "No database query was created for the tag"
	errTmpl = "The server could not render the HTML template for this page"
	firefox = "Welcome to the Firefox 2 era (October 2006) Defacto2 website, which is friendly for legacy operating systems, including Windows 9x, NT-4, and OS-X 10.2."

	textArt = "hi-res, raster and pixel images"
	textDoc = "documents using any media format, including text files, ASCII, and ANSI text art"
	textSof = "applications and programs for any platform"
)

// LegacyURLs are partial URL routers that are to be redirected with a HTTP 308
// permanent redirect status code. These are for retired URL syntaxes that are still
// found on websites online, so their links to Defacto2 do not break with 404, not found errors.
var LegacyURLs = map[string]string{
	"/index":            "",
	"/categories/index": "/categories",
	"/platforms/index":  "/platforms",
}

// Routes for the /html3 sub-route group.
// Any errors are logged and rendered to the client using HTTP codes
// and the custom /html3, group errror template.
func Routes(e *echo.Echo, log *zap.SugaredLogger) {
	s := sugared{log: log}
	g := e.Group(Prefix)
	g.GET("", s.Index)
	g.GET("/categories", s.Categories)
	g.GET("/category/:id", s.Category)
	g.GET("/platforms", s.Platforms)
	g.GET("/platform/:id", s.Platform)
	g.GET("/groups", s.Groups)
	g.GET("/group/:id", s.Group)
	g.GET("/art", s.Art)
	g.GET("/documents", s.Documents)
	g.GET("/software:offset", s.Software)
	g.GET("/d/:id", s.Download)
	// append legacy redirects
	for url := range LegacyURLs {
		g.GET(url, s.Redirection)
	}
}

// GroupCache is a cached collection of important, expensive group data.
// The Mu mutex must always be locked before writing this varable.
var IndexCache IndexSums

// GroupCol is a cached collection of important, expensive group data.
// The Mu mutex must always be locked when writing to the Groups map.
type IndexSums struct {
	Mu   sync.Mutex
	Sums map[int]int
}

// Download serves the file download of a record to the user and prompts for a save location.
// The record key id is provided by a URL param.
func (s *sugared) Download(c echo.Context) error {
	return dl.Download(s.log, c)
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
			sum, err = models.GroupCount(ctx, db)
		}
		if err != nil {
			s.log.Warnf("%s: %s", errConn, err)
			continue
		}
		IndexCache.Sums[i] = sum
	}
	descs := [3]string{helpers.Sentence(textArt), helpers.Sentence(textDoc), helpers.Sentence(textSof)}
	err = c.Render(http.StatusOK, "index", map[string]interface{}{
		"title":       title,
		"description": desc,
		"descs":       descs,
		"sums":        IndexCache.Sums,
		"cat":         tags.CategoryCount,
		"plat":        tags.PlatformCount,
		"latency":     fmt.Sprintf("%s.", time.Since(*start)),
	})
	if err != nil {
		s.log.Errorf("%s: %s", errTmpl, err)
		return echo.NewHTTPError(http.StatusInternalServerError, errTmpl)
	}
	return nil
}

// Categories lists the names, descriptions and sums of the category (section) tags.
func (s *sugared) Categories(c echo.Context) error {
	start := latency()
	err := c.Render(http.StatusOK, "tag", map[string]interface{}{
		"title":       title + "/categories",
		"description": "File categories and classification tags.",
		"latency":     fmt.Sprintf("%s.", time.Since(*start)),
		"path":        "category",
		"tagFirst":    tags.FirstCategory,
		"tagEnd":      tags.LastCategory,
		"tags":        tags.Names,
	})
	if err != nil {
		s.log.Errorf("%s: %s %d", errTmpl, err)
		return echo.NewHTTPError(http.StatusInternalServerError, errTmpl)
	}
	return nil
}

// Platforms lists the names, descriptions and sums of the platform tags.
func (s *sugared) Platforms(c echo.Context) error {
	start := latency()
	err := c.Render(http.StatusOK, "tag", map[string]interface{}{
		"title":       title + "/platforms",
		"description": "File platforms, operating systems and media types.",
		"latency":     fmt.Sprintf("%s.", time.Since(*start)),
		"path":        "platform",
		"tagFirst":    tags.FirstPlatform,
		"tagEnd":      tags.LastPlatform,
		"tags":        tags.Names,
	})
	if err != nil {
		s.log.Errorf("%s: %s %d", errTmpl, err)
		return echo.NewHTTPError(http.StatusInternalServerError, errTmpl)
	}
	return nil
}

// Groups lists the names and sums of all the distinct scene groups.
func (s *sugared) Groups(c echo.Context) error {
	start := latency()
	ctx := context.Background()
	db, err := postgres.ConnectDB()
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, errConn)
	}
	defer db.Close()
	total, err := models.GroupCount(ctx, db)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, errConn)
	}
	// if there is an out of date cache, it will get updated in the background
	// but the client will probably be rendered with an incomplete, stale cache.
	feedback := ""
	models.Groups.Mu.RLock()
	l := len(models.Groups.List)
	models.Groups.Mu.RUnlock()
	if l != total {
		go func(err error) error {
			return models.Groups.Update()
		}(err)
		if err != nil {
			return echo.NewHTTPError(http.StatusNotFound, errConn)
		}
		feedback = refreshInfo(l, total)
	}
	models.Groups.Mu.RLock()
	defer models.Groups.Mu.RUnlock()
	err = c.Render(http.StatusOK, "groups", map[string]interface{}{
		"feedback": feedback,
		"title":    title + "/groups",
		"description": "Listed is an exhaustive, distinct collection of scene groups and site brands." +
			" Do note that Defacto2 is a file-serving site, so the list doesn't distinguish between different groups with the same name or brand.",
		"latency": fmt.Sprintf("%s.", time.Since(*start)),
		"path":    "group",
		"sceners": models.Groups.List,
	})
	if err != nil {
		s.log.Errorf("%s: %s %d", errTmpl, err)
		return echo.NewHTTPError(http.StatusInternalServerError, errTmpl)
	}
	return nil
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

// Redirection redirects any legacy URL matches.
func (s *sugared) Redirection(c echo.Context) error {
	for u, redirect := range LegacyURLs {
		htm := Prefix + u
		if htm == c.Path() {
			return c.Redirect(http.StatusPermanentRedirect, Prefix+redirect)
		}
	}
	err := c.String(http.StatusInternalServerError,
		fmt.Sprintf("unknown redirection, %q ", c.Path()))
	if err != nil {
		s.log.Errorf("%s: %s %d", errTmpl, err)
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
	return c.Render(code, "error", map[string]interface{}{
		"title":       fmt.Sprintf("%d error, there is a complication", code),
		"description": fmt.Sprintf("%s.", msg),
		"latency":     fmt.Sprintf("%s.", time.Since(*start)),
	})
}

func latency() *time.Time {
	start := time.Now()
	r := new(big.Int)
	const n, k = 1000, 10
	r.Binomial(n, k)
	return &start
}
