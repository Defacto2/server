package bootstrap

import (
	"embed"
	"fmt"
	"html/template"
	"net/http"
	"strings"
	"time"

	"github.com/Defacto2/server/pkg/helpers"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

const (
	layout      = "layout.html"
	modal       = "modal.html"
	pagination  = "pagination.html"
	layoutCSS   = "public/css/layout.min.css"
	nameCSS     = "public/css/bootstrap.min.css"
	nameJS      = "public/js/bootstrap.bundle.min.js"
	fontawesome = "public/js/fontawesome.min.js"
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
	fa, err := Integrity(fontawesome, JS)
	if err != nil {
		fmt.Println(err) // TODO: logger
		return err
	}
	cssLay, err := Integrity(layoutCSS, CSS)
	if err != nil {
		fmt.Println(err) // TODO: logger
		return err
	}

	err = ctx.Render(http.StatusOK, "index", map[string]interface{}{
		"integrityCSS":    css,
		"integrityLayout": cssLay,
		"integrityJS":     js,
		"integrityFA":     fa,
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
	// note: the path is relative to the embed.FS root and must not use the OS path separator.
	return strings.Join([]string{"view", "bootstrap", name}, "/")
}

// Index template.
func index(log *zap.SugaredLogger, fs embed.FS) *template.Template {
	return template.Must(template.New("").Funcs(TemplateFuncMap(log)).ParseFS(fs,
		GlobTo(layout), GlobTo("index.html"), GlobTo(modal)))
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
