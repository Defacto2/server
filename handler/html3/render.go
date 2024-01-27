package html3

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Defacto2/server/internal/helper"
	"github.com/Defacto2/server/internal/postgres"
	"github.com/Defacto2/server/internal/tags"
	"github.com/Defacto2/server/model"
	"github.com/labstack/echo/v4"
)

const (
	Prefix = "/html3" // Root path of the HTML3 router group.

	title   = "Index of " + Prefix
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
	const blank = ""
	return [l]string{
		blank,
		"categories",
		"platforms",
		"groups",
		blank,
		blank,
		blank,
	}[t]
}

// Index method is the homepage of the /html3 sub-route.
func (s *sugared) Index(c echo.Context) error {
	start := helper.Latency()
	const desc = firefox
	// Stats are the database statistics.
	var stats struct {
		All      model.Files
		Art      model.Arts
		Document model.Docs
		// Group    model.Rels
		Software model.Softs
	}
	ctx := context.Background()
	db, err := postgres.ConnectDB()
	if err != nil {
		s.zlog.Warnf("%s: %s", errConn, err)
		return echo.NewHTTPError(http.StatusServiceUnavailable, errConn)
	}
	defer db.Close()
	if err := stats.All.Stat(ctx, db); err != nil {
		s.zlog.Warnf("%s: %s", errConn, err)
	}
	if err := stats.Art.Stat(ctx, db); err != nil {
		s.zlog.Warnf("%s: %s", errConn, err)
	}
	if err := stats.Document.Stat(ctx, db); err != nil {
		s.zlog.Warnf("%s: %s", errConn, err)
	}
	if err := stats.Software.Stat(ctx, db); err != nil {
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
		"relstats":    stats,
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

	page := 1
	offset := strings.TrimPrefix(c.Param("offset"), "/")
	if offset != "" {
		// this permits blank offsets param but returns 404 for a /0 value
		page, _ = strconv.Atoi(offset)
		if page < 1 {
			return echo.NewHTTPError(http.StatusNotFound,
				fmt.Sprintf("Page %d of %s doesn't exist", page, "/groups"))
		}
	}

	// releasers are the distinct groups from the file table.

	var unique model.ReleaserNames
	if err := unique.List(ctx, db); err != nil {
		s.zlog.Errorf("%s: %s %d", errConn, err)
		return echo.NewHTTPError(http.StatusNotFound, errSQL)
	}
	count := len(unique)

	maxPage := uint(0)
	limit := model.Maximum
	if limit > 0 {
		maxPage = helper.PageCount(count, limit)
		if page > int(maxPage) {
			return echo.NewHTTPError(http.StatusNotFound,
				fmt.Sprintf("Page %d of %d for %s doesn't exist", page, maxPage, " TODO"))
		}
	}

	navi := Navi(limit, page, maxPage, "groups", qs(c.QueryString()))
	navi.Link1, navi.Link2, navi.Link3 = Pagi(page, maxPage)

	// releasers are the distinct groups from the file table.
	releasers := model.Releasers{}
	if err := releasers.All(ctx, db, false, model.Maximum, page); err != nil {
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
		"releasers": releasers, // model.Grps.List
		"navigate":  navi,
	})
	if err != nil {
		s.zlog.Errorf("%s: %s %d", errTmpl, err)
		return echo.NewHTTPError(http.StatusInternalServerError, errTmpl)
	}
	return nil
}
