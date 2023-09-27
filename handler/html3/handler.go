package html3

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/Defacto2/server/internal/helper"
	"github.com/Defacto2/server/internal/postgres"
	"github.com/Defacto2/server/internal/tags"
	"github.com/Defacto2/server/model"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

const (
	Prefix = "/html3" // Root path of the HTML3 router group.

	title   = "Index of " + Prefix
	errConn = "Sorry, at the moment the server cannot connect to the database"
	errSQL  = "Database connection problem or a SQL error"
	errTag  = "No database query was created for the tag"
	errTmpl = "The server could not render the HTML template for this page"
	firefox = "Welcome to the Firefox v2, 2006 era, Defacto2 website, " +
		"which is friendly for legacy operating systems, including Windows 9x, NT-4, and OS-X 10.2."

	textAll = "list every file or release hosted on the website"
	textArt = "hi-res, raster and pixel images"
	textDoc = "documents using any media format, including text files, ASCII, and ANSI text art"
	textSof = "applications and programs for any platform"
)

// RecordsBy are the record groupings.
type RecordsBy int

const (
	AllReleases RecordsBy = iota // AllReleases displays all records from the file table.
	BySection                    // BySection groups records by the section file table column.
	ByPlatform                   // BySection groups records by the platform file table column.
	ByGroup                      // ByGroup groups the records by the distinct, group_brand_for file table column.
	AsArt                        // AsArt group records as art.
	AsDocuments                  // AsDocuments group records as documents.
	AsSoftware                   // AsSoftware group records as software.
)

// RecordsBy are the record groupings.
func (t RecordsBy) String() string {
	const l = 7
	if t >= l {
		return ""
	}
	return [l]string{
		"html3_all",
		"html3_category",
		"html3_platform",
		"html3_group",
		"html3_art",
		"html3_documents",
		"html3_software",
	}[t]
}

// Parent returns the parent route for the current route.
func (t RecordsBy) Parent() string {
	const l = 7
	if t >= l {
		return ""
	}
	return [l]string{
		"",
		"categories",
		"platforms",
		"groups",
		"", "", "",
	}[t]
}

// Stats are the database statistics.
var Stats struct { //nolint:gochecknoglobals
	All      model.Files
	Art      model.Arts
	Document model.Docs
	// Group    model.Rels
	Software model.Softs
}

// Releasers are the distinct groups from the file table.
var Releasers model.Releasers //nolint:gochecknoglobals

// Routes for the /html3 sub-route group.
// Any errors are logged and rendered to the client using HTTP codes
// and the custom /html3, group errror template.
func Routes(z *zap.SugaredLogger, e *echo.Echo) *echo.Group {
	s := sugared{zlog: z}
	g := e.Group(Prefix)
	g.GET("", s.Index)
	g.GET("/all:offset", s.All)
	g.GET("/all", s.All)
	g.GET("/categories", s.Categories)
	g.GET("/category/:id/:offset", s.Category)
	g.GET("/category/:id", s.Category)
	g.GET("/platforms", s.Platforms)
	g.GET("/platform/:id/:offset", s.Platform)
	g.GET("/platform/:id", s.Platform)
	g.GET("/groups", s.Groups)
	g.GET("/group/:id", s.Group)
	g.GET("/art:offset", s.Art)
	g.GET("/art", s.Art)
	g.GET("/documents:offset", s.Documents)
	g.GET("/documents", s.Documents)
	g.GET("/software:offset", s.Software)
	g.GET("/software", s.Software)
	// append legacy redirects
	for url := range Redirects() {
		g.GET(url, s.Redirection)
	}
	return g
}

// Index method is the homepage of the /html3 sub-route.
func (s *sugared) Index(c echo.Context) error {
	start := helper.Latency()
	const desc = firefox
	ctx := context.Background()
	db, err := postgres.ConnectDB()
	if err != nil {
		s.zlog.Warnf("%s: %s", errConn, err)
		return echo.NewHTTPError(http.StatusServiceUnavailable, errConn)
	}
	defer db.Close()
	if err := Stats.All.Stat(ctx, db); err != nil {
		s.zlog.Warnf("%s: %s", errConn, err)
	}
	if err := Stats.Art.Stat(ctx, db); err != nil {
		s.zlog.Warnf("%s: %s", errConn, err)
	}
	if err := Stats.Document.Stat(ctx, db); err != nil {
		s.zlog.Warnf("%s: %s", errConn, err)
	}
	// Need to replace?
	// if err := Stats.Group.Stat(ctx, db); err != nil {
	// 	s.log.Warnf("%s: %s", errConn, err)
	// }
	if err := Stats.Software.Stat(ctx, db); err != nil {
		s.zlog.Warnf("%s: %s", errConn, err)
	}
	descs := [4]string{
		helper.Capitalize(textArt),
		helper.Capitalize(textDoc),
		helper.Capitalize(textSof),
		helper.Capitalize(textAll),
	}
	if err = c.Render(http.StatusOK, "html3_index", map[string]interface{}{
		"title":       title,
		"description": desc,
		"descs":       descs,
		"relstats":    Stats,
		"cat":         tags.CategoryCount,
		"plat":        tags.PlatformCount,
		"latency":     fmt.Sprintf("%s.", time.Since(*start)),
	}); err != nil {
		s.zlog.Errorf("%s: %s", errTmpl, err)
		return echo.NewHTTPError(http.StatusInternalServerError, errTmpl)
	}
	return nil
}

// Categories lists the names, descriptions and sums of the category (section) tags.
func (s *sugared) Categories(c echo.Context) error {
	start := helper.Latency()
	err := c.Render(http.StatusOK, "html3_tag", map[string]interface{}{
		"title":       title + "/categories",
		"description": "File categories and classification tags.",
		"latency":     fmt.Sprintf("%s.", time.Since(*start)),
		"path":        "category",
		"tagFirst":    tags.FirstCategory,
		"tagEnd":      tags.LastCategory,
		"tags":        tags.Names(),
	})
	if err != nil {
		s.zlog.Errorf("%s: %s %d", errTmpl, err)
		return echo.NewHTTPError(http.StatusInternalServerError, errTmpl)
	}
	return nil
}

// Platforms lists the names, descriptions and sums of the platform tags.
func (s *sugared) Platforms(c echo.Context) error {
	start := helper.Latency()
	err := c.Render(http.StatusOK, "html3_tag", map[string]interface{}{
		"title":       title + "/platforms",
		"description": "File platforms, operating systems and media types.",
		"latency":     fmt.Sprintf("%s.", time.Since(*start)),
		"path":        "platform",
		"tagFirst":    tags.FirstPlatform,
		"tagEnd":      tags.LastPlatform,
		"tags":        tags.Names(),
	})
	if err != nil {
		s.zlog.Errorf("%s: %s %d", errTmpl, err)
		return echo.NewHTTPError(http.StatusInternalServerError, errTmpl)
	}
	return nil
}

// Groups lists the names and sums of all the distinct scene groups.
func (s *sugared) Groups(c echo.Context) error {
	start := helper.Latency()
	ctx := context.Background()
	db, err := postgres.ConnectDB()
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, errConn)
	}
	defer db.Close()
	if err := Releasers.All(ctx, db, 0, 0, false); err != nil {
		s.zlog.Errorf("%s: %s %d", errConn, err)
		return echo.NewHTTPError(http.StatusNotFound, errSQL)
	}
	err = c.Render(http.StatusOK, "html3_groups", map[string]interface{}{
		"title": title + "/groups",
		"description": "Listed is an exhaustive, distinct collection of scene groups and site brands." +
			" Do note that Defacto2 is a file-serving site, so the list doesn't distinguish between" +
			" different groups with the same name or brand.",
		"latency":   fmt.Sprintf("%s.", time.Since(*start)),
		"path":      "group",
		"releasers": Releasers, // model.Grps.List
	})
	if err != nil {
		s.zlog.Errorf("%s: %s %d", errTmpl, err)
		return echo.NewHTTPError(http.StatusInternalServerError, errTmpl)
	}
	return nil
}

// Redirection redirects any legacy URL matches.
func (s *sugared) Redirection(c echo.Context) error {
	for u, redirect := range Redirects() {
		htm := Prefix + u
		if htm == c.Path() {
			return c.Redirect(http.StatusPermanentRedirect, Prefix+redirect)
		}
	}
	err := c.String(http.StatusInternalServerError,
		fmt.Sprintf("unknown redirection, %q ", c.Path()))
	if err != nil {
		s.zlog.Errorf("%s: %s %d", errTmpl, err)
		return echo.NewHTTPError(http.StatusInternalServerError, errTmpl)
	}
	return nil
}
