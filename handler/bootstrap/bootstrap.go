package bootstrap

import (
	"embed"
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
	"time"

	"github.com/Defacto2/server/pkg/helpers"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

const (
	layout     = "layout.html"
	pagination = "pagination.html"
	nameCSS    = "public/css/bootstrap.min.css"
	nameJS     = "public/js/bootstrap.bundle.min.js"
)

// Index method is the homepage of the / sub-route.
func Index(s *zap.SugaredLogger, ctx echo.Context, CSS, JS embed.FS) error {
	errTmpl := "The server could not render the HTML template for this page"

	css, err := Integrity(nameCSS, CSS)
	if err != nil {
		fmt.Println(err) // TODO: logger
		return err
	}
	js, err := Integrity(nameJS, JS)
	if err != nil {
		fmt.Println(err) // TODO: logger
		return err
	}

	err = ctx.Render(http.StatusOK, "index", map[string]interface{}{
		"integrityCSS": css,
		"integrityJS":  js,
	})
	if err != nil {
		s.Errorf("%s: %s", errTmpl, err)
		return echo.NewHTTPError(http.StatusInternalServerError, errTmpl)
	}

	return nil
}

// Tmpl returns a map of the templates used by the route.
func Tmpl(log *zap.SugaredLogger, fs embed.FS) map[string]*template.Template {
	templates := make(map[string]*template.Template)
	templates["index"] = index(log, fs)
	templates["error"] = httpErr(log, fs)
	return templates
}

func GlobTo(name string) string {
	return filepath.Join("view", "bootstrap", name)
}

// Index template.
func index(log *zap.SugaredLogger, fs embed.FS) *template.Template {
	return template.Must(template.New("").Funcs(TemplateFuncMap(log)).ParseFS(fs,
		GlobTo(layout), GlobTo("index.html")))
}

// Template for displaying HTTP error codes and feedback.
func httpErr(log *zap.SugaredLogger, fs embed.FS) *template.Template {
	return template.Must(template.New("").Funcs(TemplateFuncMap(log)).ParseFS(fs,
		GlobTo(layout)))
}

// TemplateFuncMap are a collection of mapped functions that can be used in a template.
func TemplateFuncMap(log *zap.SugaredLogger) template.FuncMap {
	return template.FuncMap{
		"safeHTML": func(s string) template.HTML {
			return template.HTML(s)
		},
	}
}

// Error renders a custom HTTP error page for the HTML3 sub-group.
func Error(err error, c echo.Context) error {
	// Echo custom error handling: https://echo.labstack.com/guide/error-handling/
	start := helpers.Latency()
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
