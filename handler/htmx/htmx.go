// Package htmx handles the routes and views for the AJAX responses using the htmx library.
package htmx

import (
	"context"
	"embed"
	"html/template"
	"net/http"
	"strings"

	"github.com/Defacto2/releaser"
	"github.com/Defacto2/releaser/initialism"
	"github.com/Defacto2/releaser/name"
	"github.com/Defacto2/server/handler/app"
	"github.com/Defacto2/server/internal/helper"
	"github.com/Defacto2/server/internal/postgres"
	"github.com/Defacto2/server/model"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// Routes for the /htmx sub-route group that returns HTML fragments
// using the htmx library for AJAX responses.
func Routes(logr *zap.SugaredLogger, e *echo.Echo, dlDir string) *echo.Echo {
	e.POST("/search/releaser", func(x echo.Context) error {
		return PostReleaser(logr, x)
	})
	e.POST("/demozoo/download", func(x echo.Context) error {
		return PostDemozooLink(logr, x, dlDir) // dir.Download
	})
	return e
}

// GlobTo returns the path to the template file.
func GlobTo(name string) string {
	// note: the path is relative to the embed.FS root and must not use the OS path separator.
	return strings.Join([]string{"view", "htmx", name}, "/")
}

// PostReleaser is a handler for the /search/releaser route.
func PostReleaser(logr *zap.SugaredLogger, c echo.Context) error {
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

func releasers(fs embed.FS) *template.Template {
	return template.Must(template.New("").Funcs(TemplateFuncMap()).ParseFS(fs,
		GlobTo("layout.tmpl"), GlobTo("releasers.tmpl")))
}

// Templates returns a map of the templates used by the HTML3 sub-group route.
func Templates(fs embed.FS) map[string]*template.Template {
	t := make(map[string]*template.Template)
	t["releasers"] = releasers(fs)
	return t
}

// TemplateFuncMap are a collection of mapped functions that can be used in a template.
func TemplateFuncMap() template.FuncMap {
	return template.FuncMap{
		"borderClass": func(name, path string) string {
			const mark = "border border-primary"
			if strings.EqualFold(name, path) {
				return mark
			}
			init := initialism.Join(initialism.Path(path))
			if strings.EqualFold(name, init) {
				return mark
			}
			return "border"
		},
		"byteFileS": app.ByteFileS,
		"fmtPath": func(path string) string {
			if val := name.Path(path); val.String() != "" {
				return val.String()
			}
			return releaser.Humanize(path)
		},
		"initialisms": func(s string) string {
			return initialism.Join(initialism.Path(s))
		},
		"safeHTML": func(s string) template.HTML {
			return template.HTML(s) //nolint:gosec
		},
	}
}
