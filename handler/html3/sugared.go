package html3

// Package file sugared.go contains the HTML3 website route functions.

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Defacto2/helper"
	"github.com/Defacto2/server/internal/panics"
	"github.com/Defacto2/server/internal/tags"
	"github.com/Defacto2/server/model"
	"github.com/Defacto2/server/model/html3"
	"github.com/labstack/echo/v4"
)

// All method lists every release.
func All(c echo.Context, db *sql.DB, sl *slog.Logger) error {
	return List(c, db, sl, Everything)
}

// Art lists the file records described as art are digital + pixel art files.
func Art(c echo.Context, db *sql.DB, sl *slog.Logger) error {
	return List(c, db, sl, AsArt)
}

// Categories lists the names, descriptions and sums of the category (section) tags.
func Categories(c echo.Context, sl *slog.Logger) error {
	const msg = "htm3 categories"
	if err := panics.EchoContextS(c, sl); err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}
	start := helper.Latency()
	err := c.Render(http.StatusOK, string(tag), map[string]any{
		"title":       title + "/categories",
		"description": "Artifact categories and classification tags.",
		"latency":     time.Since(*start).String() + ".",
		"path":        "category",
		"tagFirst":    tags.FirstCategory,
		"tagEnd":      tags.LastCategory,
		"tags":        tags.Names(),
	})
	if err != nil {
		sl.Error(msg, slog.String("render", ErrTmpl.Error()), slog.Any("error", err))
		return echo.NewHTTPError(http.StatusInternalServerError, ErrTmpl)
	}
	return nil
}

// Category lists the file records associated with the category tag that is provided by the ID param in the URL.
func Category(c echo.Context, db *sql.DB, sl *slog.Logger) error {
	return List(c, db, sl, BySection)
}

// Documents lists the file records described as document + text art files.
func Documents(c echo.Context, db *sql.DB, sl *slog.Logger) error {
	return List(c, db, sl, AsDocument)
}

// Group lists the file records associated with the group that is provided by the ID param in the URL.
func Group(c echo.Context, db *sql.DB, sl *slog.Logger) error {
	return List(c, db, sl, ByGroup)
}

// Groups lists the names and sums of all the distinct scene groups.
func Groups(c echo.Context, db *sql.DB, sl *slog.Logger) error {
	const msg = "html3 groups listings"
	if err := panics.EchoContextS(c, sl); err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}
	start := helper.Latency()
	ctx := context.Background()
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
	if err := unique.DistinctGroups(ctx, db); err != nil {
		sl.Error(msg, slog.String("distinct", ErrSQL.Error()), slog.Any("error", err))
		return echo.NewHTTPError(http.StatusNotFound, ErrSQL)
	}
	count := len(unique)
	maxPage := 0
	limit := model.Maximum
	if limit > 0 {
		maxPage = helper.PageCount(count, limit)
		if page > maxPage {
			return echo.NewHTTPError(http.StatusNotFound,
				fmt.Sprintf("Page %d of %d for %s doesn't exist", page, maxPage, " groups"))
		}
	}
	navi := Navi(limit, page, maxPage, "groups", qs(c.QueryString()))
	navi.Link1, navi.Link2, navi.Link3 = Pagi(page, maxPage)
	// releasers are the distinct groups from the file table.
	releasers := model.Releasers{}
	if err := releasers.Limit(ctx, db, model.Alphabetical, model.Maximum, page); err != nil {
		sl.Error(msg, slog.String("alphabetical", ErrSQL.Error()), slog.Any("error", err))
		return echo.NewHTTPError(http.StatusNotFound, ErrSQL)
	}
	err := c.Render(http.StatusOK, "html3_groups", map[string]any{
		"title": title + "/groups",
		"description": "Listed is an exhaustive, distinct collection of scene groups and site brands." +
			" Do note that Defacto2 is a file-serving site, so the list doesn't distinguish between" +
			" different groups with the same name or brand.",
		"latency":   time.Since(*start).String() + ".",
		"path":      "group",
		"releasers": releasers, // model.Grps.List
		"navigate":  navi,
	})
	if err != nil {
		sl.Error(msg, slog.String("template", ErrTmpl.Error()), slog.Any("error", err))
		return echo.NewHTTPError(http.StatusInternalServerError, ErrTmpl)
	}
	return nil
}

// Index method is the homepage of the /html3 sub-route.
func Index(c echo.Context, db *sql.DB, sl *slog.Logger) error {
	const msg = "html3 index"
	if err := panics.EchoContextDS(c, db, sl); err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}
	start := helper.Latency()
	const desc = firefox
	// Stats are the database statistics.
	var stats struct {
		All      model.Artifacts
		Art      html3.Arts
		Document html3.Documents
		Software html3.Softwares
	}
	ctx := context.Background()
	if err := stats.All.Public(ctx, db); err != nil {
		sl.Warn(msg, slog.String("statistics", "results for all"), slog.Any("error", err))
	}
	if err := stats.Art.Stat(ctx, db); err != nil {
		sl.Warn(msg, slog.String("statistics", "results for art"), slog.Any("error", err))
	}
	if err := stats.Document.Stat(ctx, db); err != nil {
		sl.Warn(msg, slog.String("statistics", "results for documents"), slog.Any("error", err))
	}
	if err := stats.Software.Stat(ctx, db); err != nil {
		sl.Warn(msg, slog.String("statistics", "results for software"), slog.Any("error", err))
	}
	descs := [4]string{
		helper.Capitalize(textArt),
		helper.Capitalize(textDoc),
		helper.Capitalize(textSof),
		helper.Capitalize(textAll),
	}
	err := c.Render(http.StatusOK, "html3_index", map[string]any{
		"title":       title,
		"description": desc,
		"descs":       descs,
		"relstats":    stats,
		"cat":         tags.CategoryCount,
		"plat":        tags.PlatformCount,
		"latency":     time.Since(*start).String() + ".",
	})
	if err != nil {
		sl.Error(msg, "template", slog.String("info", ErrTmpl.Error()), slog.Any("error", err))
		return echo.NewHTTPError(http.StatusInternalServerError, ErrTmpl)
	}
	return nil
}

// List all the records associated with the RecordsBy grouping.
func List(c echo.Context, db *sql.DB, sl *slog.Logger, tt RecordsBy) error { //nolint:funlen
	const msg = "htm3 list records by"
	if err := panics.EchoContextDS(c, db, sl); err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}
	start := helper.Latency()
	var id string
	switch tt {
	case BySection, ByPlatform:
		id = ID(c)
	default:
		id = c.Param("id")
	}
	// pagination offset and page number
	page := 1
	offset := strings.TrimPrefix(c.Param("offset"), "/")
	if offset != "" {
		// this permits blank offsets param but returns 404 for a /0 value
		page, _ = strconv.Atoi(offset)
		if page < 1 {
			return echo.NewHTTPError(http.StatusNotFound,
				fmt.Sprintf("Page %d of %s doesn't exist", page, tt))
		}
	}
	// query database to return records and statistics
	limit, count, byteSum, records, err := Query(c, db, tt, page)
	if err != nil {
		sl.Error(msg, slog.String("database", "record and statistics query problem"), slog.Any("error", err))
		return echo.NewHTTPError(http.StatusServiceUnavailable, ErrConn)
	}
	if limit > 0 && count == 0 {
		return echo.NewHTTPError(http.StatusNotFound,
			fmt.Sprintf("The %s %q doesn't exist", tt, id))
	}
	// pagination maximum page number
	maxPage := 0
	if limit > 0 {
		maxPage = helper.PageCount(count, limit)
		if page > maxPage {
			return echo.NewHTTPError(http.StatusNotFound,
				fmt.Sprintf("Page %d of %d for %s doesn't exist", page, maxPage, tt))
		}
	}
	// pagination values
	current := strings.TrimPrefix(tt.String(), "html3_")
	switch tt {
	case BySection:
		current = "category/" + id
	case ByPlatform:
		current = "platform/" + id
	}
	navi := Navi(limit, page, maxPage, current, qs(c.QueryString()))
	navi.Link1, navi.Link2, navi.Link3 = Pagi(page, maxPage)
	// string based values for use in templates
	stat := fmt.Sprintf("%d files, %s", count, helper.ByteCountFloat(byteSum))
	title, desc := ListInfo(tt, current, id)
	err = c.Render(http.StatusOK, tt.String(), map[string]any{
		"title":       title,
		"home":        "",
		"description": desc,
		"parent":      tt.Parent(),
		"stats":       stat,
		"sort":        Sorter(c.QueryString()),
		"records":     records,
		"latency":     time.Since(*start).String() + ".",
		"navigate":    navi,
	})
	if err != nil {
		sl.Error(msg, slog.String("template", ErrTmpl.Error()), slog.Any("error", err))
		return echo.NewHTTPError(http.StatusInternalServerError, ErrTmpl)
	}
	return nil
}

// Platform lists the file records associated with the platform tag that is provided by the ID param in the URL.
func Platform(c echo.Context, db *sql.DB, sl *slog.Logger) error {
	return List(c, db, sl, ByPlatform)
}

// Platforms lists the names, descriptions and sums of the platform tags.
func Platforms(c echo.Context, sl *slog.Logger) error {
	const msg = "htm3 platforms"
	if err := panics.EchoContextS(c, sl); err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}
	start := helper.Latency()
	err := c.Render(http.StatusOK, string(tag), map[string]any{
		"title":       title + "/platforms",
		"description": "File platforms, operating systems and media types.",
		"latency":     time.Since(*start).String() + ".",
		"path":        "platform",
		"tagFirst":    tags.FirstPlatform,
		"tagEnd":      tags.LastPlatform,
		"tags":        tags.Names(),
	})
	if err != nil {
		sl.Error(msg, slog.String("template", ErrTmpl.Error()), slog.Any("error", err))
		return echo.NewHTTPError(http.StatusInternalServerError, ErrTmpl)
	}
	return nil
}

// Software lists the file records described as software files.
func Software(c echo.Context, db *sql.DB, sl *slog.Logger) error {
	return List(c, db, sl, AsSoftware)
}
