package htmx

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
func SearchReleaser(logr *zap.SugaredLogger, c echo.Context) error {
	const maxResults = 14
	ctx := context.Background()
	db, err := postgres.ConnectDB()
	if err != nil {
		logr.Error(err)
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
		logr.Error(err)
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

func DataListReleasers(logr *zap.SugaredLogger, c echo.Context) error {
	const maxResults = 14
	ctx := context.Background()
	db, err := postgres.ConnectDB()
	if err != nil {
		logr.Error(err)
		return c.String(http.StatusServiceUnavailable,
			"cannot connect to the database")
	}
	defer db.Close()

	input := c.FormValue("uploader-intro-releasers")

	// todo split input by +

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
		logr.Error(err)
		return c.String(http.StatusServiceUnavailable,
			"the search query failed")
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
	// s := `<option value="Edge">`
	// s += `<option value="Firefox">`
	// s += `<option value="Chrome">`
	// s += `<option value="Opera">`
	// s += `<option value="Safari">`
	// return c.HTML(http.StatusOK, s)
}
