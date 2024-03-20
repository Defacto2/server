package htmx

import (
	"context"
	"embed"
	"html/template"
	"net/http"
	"strings"

	"github.com/Defacto2/server/internal/helper"
	"github.com/Defacto2/server/internal/postgres"
	"github.com/Defacto2/server/model"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// Routes for the /html3 sub-route group.
// Any errors are logged and rendered to the client using HTTP codes
// and the custom /html3, group errror template.
func Routes(logr *zap.SugaredLogger, e *echo.Echo) *echo.Echo {
	e.POST("/search/releaser", func(x echo.Context) error {
		return PostReleaser(logr, x)
	})
	return e
}

func GlobTo(name string) string {
	// note: the path is relative to the embed.FS root and must not use the OS path separator.
	return strings.Join([]string{"view", "htmx", name}, "/")
}
func PostReleaser(logr *zap.SugaredLogger, c echo.Context) error {
	const name = "postReleaser"
	ctx := context.Background()
	db, err := postgres.ConnectDB()
	if err != nil {
		return err
		//InternalErr(logr, c, name, err)
	}
	defer db.Close()
	var r model.Releasers
	//const prolificOrder = true
	// if err := r.All(ctx, db, prolificOrder, 0, 0); err != nil {
	// 	return DatabaseErr(logr, c, name, err)
	// }

	input := c.FormValue("releaser-data-list")
	val := helper.TrimRoundBraket(input)
	slug := helper.Slug(val)

	if err := r.Find(ctx, db, slug, 10); err != nil {
		return err
		//DatabaseErr(logr, c, name, err)
	}

	return c.Render(http.StatusOK, "hello", map[string]interface{}{
		"name": "Dolly!",
	})

	//err = c.HTML(http.StatusOK, fmt.Sprintf("%s", r))
	// if err != nil {
	// 	return InternalErr(logr, c, name, err)
	// }
	// return nil

	// err = c.Render(http.StatusOK, "html3_groups", map[string]interface{}{
	// 	"title": title + "/groups",
	// 	"description": "Listed is an exhaustive, distinct collection of scene groups and site brands." +
	// 		" Do note that Defacto2 is a file-serving site, so the list doesn't distinguish between" +
	// 		" different groups with the same name or brand.",
	// 	"latency":   time.Since(*start).String() + ".",
	// 	"path":      "group",
	// 	"releasers": releasers, // model.Grps.List
	// 	"navigate":  navi,
	// })
}

func hello(logr *zap.SugaredLogger, fs embed.FS) *template.Template {
	return template.Must(template.New("").Funcs(TemplateFuncMap(logr)).ParseFS(fs,
		GlobTo("layout.tmpl"), GlobTo("hello.tmpl")))
}

// Templates returns a map of the templates used by the HTML3 sub-group route.
func Templates(logr *zap.SugaredLogger, fs embed.FS) map[string]*template.Template {
	t := make(map[string]*template.Template)
	t["hello"] = hello(logr, fs)
	return t
}

// TemplateFuncMap are a collection of mapped functions that can be used in a template.
func TemplateFuncMap(logr *zap.SugaredLogger) template.FuncMap {
	return template.FuncMap{
		"safeHTML": func(s string) template.HTML {
			return template.HTML(s) //nolint:gosec
		},
	}
}
