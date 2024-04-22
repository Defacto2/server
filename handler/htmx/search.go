package htmx

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/Defacto2/releaser/initialism"
	"github.com/Defacto2/server/internal/helper"
	"github.com/Defacto2/server/internal/postgres"
	"github.com/Defacto2/server/internal/tags"
	"github.com/Defacto2/server/model"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// LookupSHA384 is a handler for the /uploader/sha384 route.
func LookupSHA384(c echo.Context, logger *zap.SugaredLogger) error {
	hash := c.Param("hash")
	if hash == "" {
		return c.String(http.StatusBadRequest, "empty hash error")
	}
	match, err := regexp.MatchString("^[a-fA-F0-9]{96}$", hash)
	if err != nil {
		logger.Error(err)
		return c.String(http.StatusBadRequest, "regex match error")
	}
	if !match {
		return c.String(http.StatusBadRequest, "invalid hash error: "+hash)
	}

	ctx := context.Background()
	db, err := postgres.ConnectDB()
	if err != nil {
		logger.Error(err)
		return c.String(http.StatusServiceUnavailable,
			"cannot connect to the database")
	}
	defer db.Close()

	exist, err := model.ExistHash(ctx, db, hash)
	if err != nil {
		logger.Error(err)
		return c.String(http.StatusServiceUnavailable,
			"cannot confirm the hash with the database")
	}
	switch exist {
	case true:
		return c.String(http.StatusOK, "true")
	case false:
		return c.String(http.StatusOK, "false")
	}
	return c.String(http.StatusServiceUnavailable,
		"unexpected boolean error occurred")
}

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

func DataListReleasers(c echo.Context, logger *zap.SugaredLogger, input string) error {
	return datalist(c, logger, input, false)
}

func DataListMagazines(c echo.Context, logger *zap.SugaredLogger, input string) error {
	return datalist(c, logger, input, true)
}

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

func Classification(c echo.Context, logger *zap.SugaredLogger) error {
	ctx := context.Background()
	db, err := postgres.ConnectDB()
	if err != nil {
		logger.Error(err)
		return c.String(http.StatusServiceUnavailable,
			"cannot connect to the database")
	}
	defer db.Close()

	v1 := c.FormValue("uploader-advanced-category")
	v2 := c.FormValue("uploader-advanced-operatingsystem")
	sec := tags.TagByURI(v1)
	pla := tags.TagByURI(v2)
	s := tags.Humanize(pla, sec)
	if strings.HasPrefix(s, "unknown") {
		return c.HTML(http.StatusOK, "<p>unknown classification</p>")
	}

	count, err := model.CountByClassification(ctx, db, sec.String(), pla.String())
	if err != nil {
		logger.Error(err)
		return c.String(http.StatusServiceUnavailable,
			"cannot count the classification")
	}
	s = fmt.Sprintf("%s, %d existing artifacts", s, count)
	return c.HTML(http.StatusOK, s)
}
