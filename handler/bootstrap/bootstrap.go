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
	layout     = "layout.html"
	modal      = "modal.html"
	pagination = "pagination.html"

	bootCSS     = "public/css/bootstrap.min.css"
	bootJS      = "public/js/bootstrap.bundle.min.js"
	layoutCSS   = "public/css/layout.min.css"
	fontawesome = "public/js/fontawesome.min.js"
)

// SRI are the Subresource Integrity hashes for the layout.
type SRI struct {
	BootstrapCSS string // Bootstrap CSS verification hash.
	BootstrapJS  string // Bootstrap JS verification hash.
	FontAwesome  string // Font Awesome verification hash.
	LayoutCSS    string // Layout CSS verification hash.
}

func (s *SRI) Verify(css, js embed.FS) error {
	var err error
	s.BootstrapCSS, err = Integrity(bootCSS, css)
	if err != nil {
		return err
	}
	s.BootstrapJS, err = Integrity(bootJS, js)
	if err != nil {
		return err
	}
	s.FontAwesome, err = Integrity(fontawesome, js)
	if err != nil {
		return err
	}
	s.LayoutCSS, err = Integrity(layoutCSS, css)
	if err != nil {
		return err
	}
	return nil
}

// Index method is the homepage of the / sub-route.
func Index(s *zap.SugaredLogger, ctx echo.Context) error {
	errTmpl := "The server could not render the HTML template for this page"
	err := ctx.Render(http.StatusOK, "index", map[string]interface{}{
		// "integrityCSS":    css,
		// "integrityLayout": cssLay,
		// "integrityJS":     js,
		// "integrityFA":     fa,
		"title": "demo",
	})
	if err != nil {
		s.Errorf("%s: %s", errTmpl, err)
		return echo.NewHTTPError(http.StatusInternalServerError, errTmpl)
	}

	return nil
}

// Tmpl returns a map of the templates used by the route.
func Tmpl(log *zap.SugaredLogger, css, js, view embed.FS) map[string]*template.Template {
	var sri SRI
	if err := sri.Verify(css, js); err != nil {
		panic(err)
	}
	templates := make(map[string]*template.Template)
	templates["index"] = index(log, sri, view)
	templates["error"] = httpErr(log, sri, view)
	return templates
}

func GlobTo(name string) string {
	// note: the path is relative to the embed.FS root and must not use the OS path separator.
	return strings.Join([]string{"view", "bootstrap", name}, "/")
}

// Index template.
func index(log *zap.SugaredLogger, sri SRI, view embed.FS) *template.Template {
	return template.Must(template.New("").Funcs(TemplateFuncMap(log, sri)).ParseFS(view,
		GlobTo(layout), GlobTo("index.html"), GlobTo(modal)))
}

// Template for displaying HTTP error codes and feedback.
func httpErr(log *zap.SugaredLogger, sri SRI, view embed.FS) *template.Template {
	return template.Must(template.New("").Funcs(TemplateFuncMap(log, sri)).ParseFS(view,
		GlobTo(layout)))
}

// TemplateFuncMap are a collection of mapped functions that can be used in a template.
func TemplateFuncMap(log *zap.SugaredLogger, sri SRI) template.FuncMap {
	return template.FuncMap{
		"sriBootstrapCSS": func() string {
			return sri.BootstrapCSS
		},
		"sriBootstrapJS": func() string {
			return sri.BootstrapJS
		},
		"sriFontAwesome": func() string {
			return sri.FontAwesome
		},
		"sriLayoutCSS": func() string {
			return sri.LayoutCSS
		},
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
