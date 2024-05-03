package htmx

// Package file search.go provides functions for handling the HTMX requests for searching releasers and magazines.

import (
	"context"
	"net/http"
	"strings"

	"github.com/Defacto2/releaser/initialism"
	"github.com/Defacto2/server/internal/helper"
	"github.com/Defacto2/server/internal/postgres"
	"github.com/Defacto2/server/model"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// SearchReleaser is a handler for the /search/releaser route.
func SearchReleaser(c echo.Context, logger *zap.SugaredLogger) error {
	const maxResults = 14
	ctx := context.Background()
	db, err := postgres.ConnectDB()
	if err != nil {
		logger.Error(err)
		return c.String(http.StatusServiceUnavailable,
			"cannot connect to the database")
	}
	defer db.Close()

	input := c.FormValue("releaser-search")
	slug := helper.Slug(helper.TrimRoundBraket(input))
	if slug == "" {
		return c.HTML(http.StatusOK, "<!-- empty search query -->")
	}

	lookup := []string{}
	for key, values := range initialism.Initialisms() {
		for _, value := range values {
			if strings.Contains(strings.ToLower(value), strings.ToLower(slug)) {
				lookup = append(lookup, string(key))
			}
		}
	}
	lookup = append(lookup, slug)
	var r model.Releasers
	if err := r.Similar(ctx, db, maxResults, lookup...); err != nil {
		logger.Error(err)
		return c.String(http.StatusServiceUnavailable,
			"the search query failed")
	}
	if len(r) == 0 {
		return c.HTML(http.StatusOK, "No releasers found.")
	}
	err = c.Render(http.StatusOK, "releasers", map[string]interface{}{
		"maximum": maxResults,
		"name":    slug,
		"result":  r,
	})
	if err != nil {
		return c.String(http.StatusInternalServerError,
			"cannot render the htmx template")
	}
	return nil
}

// DataListReleasers is a handler for the /datalist/releasers route.
func DataListReleasers(c echo.Context, logger *zap.SugaredLogger, input string) error {
	return datalist(c, logger, input, false)
}

// DataListMagazines is a handler for the /datalist/magazines route.
func DataListMagazines(c echo.Context, logger *zap.SugaredLogger, input string) error {
	return datalist(c, logger, input, true)
}

// datalist is a shared handler for the /datalist/releasers and /datalist/magazines routes.
func datalist(c echo.Context, logger *zap.SugaredLogger, input string, magazine bool) error {
	const maxResults = 14
	ctx := context.Background()
	db, err := postgres.ConnectDB()
	if err != nil {
		logger.Error(err)
		return c.String(http.StatusServiceUnavailable,
			"cannot connect to the database")
	}
	defer db.Close()

	slug := helper.Slug(helper.TrimRoundBraket(input))
	if slug == "" {
		return c.HTML(http.StatusOK, "")
	}

	lookup := []string{}
	for key, values := range initialism.Initialisms() {
		for _, value := range values {
			if strings.Contains(strings.ToLower(value), strings.ToLower(slug)) {
				lookup = append(lookup, string(key))
			}
		}
	}
	lookup = append(lookup, slug)
	var r model.Releasers
	if magazine {
		err = r.SimilarMagazine(ctx, db, maxResults, lookup...)
	} else {
		err = r.Similar(ctx, db, maxResults, lookup...)
	}
	if err != nil {
		logger.Error(err)
		return c.String(http.StatusServiceUnavailable,
			"cannot connect to the database")
	}
	if len(r) == 0 {
		return c.HTML(http.StatusOK, "")
	}
	err = c.Render(http.StatusOK, "datalist-releasers", map[string]interface{}{
		"maximum": maxResults,
		"name":    slug,
		"result":  r,
	})
	if err != nil {
		return c.String(http.StatusInternalServerError,
			"cannot render the htmx template")
	}
	return nil
}
