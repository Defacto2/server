package defaults

import (
	"embed"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
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

	viewElem = "defaults"
)

var ErrTmpl = errors.New("the server could not render the HTML template for this page")

// SRI are the Subresource Integrity hashes for the layout.
type SRI struct {
	BootstrapCSS string // Bootstrap CSS verification hash.
	BootstrapJS  string // Bootstrap JS verification hash.
	FontAwesome  string // Font Awesome verification hash.
	LayoutCSS    string // Layout CSS verification hash.
}

func (s *SRI) Verify(fs embed.FS) error {
	var err error
	s.BootstrapCSS, err = Integrity(bootCSS, fs)
	if err != nil {
		return err
	}
	s.BootstrapJS, err = Integrity(bootJS, fs)
	if err != nil {
		return err
	}
	s.FontAwesome, err = Integrity(fontawesome, fs)
	if err != nil {
		return err
	}
	s.LayoutCSS, err = Integrity(layoutCSS, fs)
	if err != nil {
		return err
	}
	return nil
}

func initData() map[string]interface{} {
	return map[string]interface{}{
		"canonical":   "",
		"carousel":    "",
		"description": "",
		"h1":          "",
		"lead":        "",
		"logo":        "",
		"title":       "",
	}
}

// Index method is the homepage of the / sub-route.
func Index(s *zap.SugaredLogger, ctx echo.Context) error {
	data := initData()
	data["description"] = "demo"
	data["title"] = "demo"
	err := ctx.Render(http.StatusOK, "index", data)
	if err != nil {
		s.Errorf("%s: %s", ErrTmpl, err)
		return echo.NewHTTPError(http.StatusInternalServerError, ErrTmpl)
	}
	return nil
}

func History(s *zap.SugaredLogger, ctx echo.Context) error {
	const lead = "Defacto founded in late February or early March of 1996, as an electronic magazine that wrote about The Scene subculture."
	data := initData()
	data["carousel"] = "#carouselDf2Artpacks"
	data["description"] = lead
	data["h1"] = "Our history"
	data["lead"] = lead
	data["title"] = "Our history"
	err := ctx.Render(http.StatusOK, "history", data)
	if err != nil {
		s.Errorf("%s: %s", ErrTmpl, err)
		return echo.NewHTTPError(http.StatusInternalServerError, ErrTmpl)
	}
	return nil
}

func Thanks(s *zap.SugaredLogger, ctx echo.Context) error {
	data := initData()
	data["description"] = "Defacto2 thankyous."
	data["h1"] = "Thank you!"
	data["lead"] = "Thanks to the hundreds of people who have contributed to Defacto2 over the decades with file submissions, hard drive donations, interviews, corrections, artwork and monetiary donations!"
	data["title"] = "Thanks!"
	err := ctx.Render(http.StatusOK, "thanks", data)
	if err != nil {
		s.Errorf("%s: %s", ErrTmpl, err)
		return echo.NewHTTPError(http.StatusInternalServerError, ErrTmpl)
	}
	return nil
}

func TheScene(s *zap.SugaredLogger, ctx echo.Context) error {
	const h1 = "What is the scene?"
	const lead = "Collectively referred to as The Scene, it is a subculture of different computer activities where participants actively share ideas and creations."
	data := initData()
	data["description"] = fmt.Sprint(h1, " ", lead)
	data["h1"] = h1
	data["lead"] = lead
	data["title"] = "The Scene"
	err := ctx.Render(http.StatusOK, "thescene", data)
	if err != nil {
		s.Errorf("%s: %s", ErrTmpl, err)
		return echo.NewHTTPError(http.StatusInternalServerError, ErrTmpl)
	}
	return nil
}

// Tmpl returns a map of the templates used by the route.
func Tmpl(log *zap.SugaredLogger, public, view embed.FS) map[string]*template.Template {
	var sri SRI
	if err := sri.Verify(public); err != nil {
		panic(err)
	}
	templates := make(map[string]*template.Template)
	templates["index"] = tmpl(log, sri, view, "index.html")
	templates["history"] = tmpl(log, sri, view, "history.html")
	templates["thanks"] = tmpl(log, sri, view, "thanks.html")
	templates["thescene"] = tmpl(log, sri, view, "thescene.html")
	templates["websites"] = tmpl(log, sri, view, "websites.html")
	templates["error"] = httpErr(log, sri, view)
	return templates
}

func GlobTo(name string) string {
	// note: the path is relative to the embed.FS root and must not use the OS path separator.
	return strings.Join([]string{"view", viewElem, name}, "/")
}

// tmpl returns a layout template for the given named view.
// Note that the name is relative to the view/defaults directory
func tmpl(log *zap.SugaredLogger, sri SRI, view embed.FS, name string) *template.Template {
	if _, err := os.Stat(filepath.Join("view", viewElem, name)); os.IsNotExist(err) {
		log.Errorf("tmpl template not found: %s", err)
		panic(err)
	} else if err != nil {
		log.Errorf("tmpl template has a problem: %s", err)
		panic(err)
	}
	files := []string{GlobTo(layout), GlobTo(name), GlobTo(modal)}
	// append any additional templates
	if name == "websites.html" {
		files = append(files, GlobTo("website.html"))
	}
	return template.Must(template.New("").Funcs(TemplateFuncMap(log, sri)).ParseFS(view, files...))
}

// httpErr is the template for displaying HTTP error codes and feedback.
func httpErr(log *zap.SugaredLogger, sri SRI, view embed.FS) *template.Template {
	return template.Must(template.New("").Funcs(TemplateFuncMap(log, sri)).ParseFS(view,
		GlobTo(layout), GlobTo("error.html"), GlobTo(modal)))
}

// TemplateFuncMap are a collection of mapped functions that can be used in a template.
func TemplateFuncMap(log *zap.SugaredLogger, sri SRI) template.FuncMap {
	return template.FuncMap{
		"externalLink": ExternalLink,
		"logoText":     LogoText,
		"wikiLink":     WikiLink,
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
