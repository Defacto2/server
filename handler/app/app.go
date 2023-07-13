package app

import (
	"embed"
	"errors"
	"html/template"
	"os"
	"path/filepath"
	"strings"

	"go.uber.org/zap"
)

const (
	layout = "layout.html" // layout is a partial template.
	modal  = "modal.html"  // modal is a partial template.

	bootCSS     = "public/css/bootstrap.min.css"      // bootCSS is the path to the minified Bootstrap 5 CSS file.
	bootJS      = "public/js/bootstrap.bundle.min.js" // bootJS is the path to the minified Bootstrap 5 JS file.
	layoutCSS   = "public/css/layout.min.css"         // layoutCSS is the path to the minified layout CSS file.
	fontawesome = "public/js/fontawesome.min.js"      // fontawesome is the path to the minified Font Awesome JS file.

	viewElem = "app" // viewElem is the name of the view element in the template.
)

var ErrTmpl = errors.New("the server could not render the HTML template for this page")

// SRI are the Subresource Integrity hashes for the layout.
type SRI struct {
	BootstrapCSS string // Bootstrap CSS verification hash.
	BootstrapJS  string // Bootstrap JS verification hash.
	FontAwesome  string // Font Awesome verification hash.
	LayoutCSS    string // Layout CSS verification hash.
}

// Verify checks the integrity of the embeded CSS and JS files.
// These are required for Subresource Integrity (SRI) verification in modern browsers.
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

// Tmpl returns a map of the templates used by the route.
func Tmpl(log *zap.SugaredLogger, public, view embed.FS) map[string]*template.Template {
	var sri SRI
	if err := sri.Verify(public); err != nil {
		panic(err)
	}
	templates := make(map[string]*template.Template)
	templates["index"] = tmpl(log, sri, view, "index.html")
	templates["file"] = tmpl(log, sri, view, "file.html")
	templates["history"] = tmpl(log, sri, view, "history.html")
	templates["thanks"] = tmpl(log, sri, view, "thanks.html")
	templates["thescene"] = tmpl(log, sri, view, "thescene.html")
	templates["websites"] = tmpl(log, sri, view, "websites.html")
	templates["error"] = httpErr(log, sri, view)
	return templates
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

// GlobTo returns the path to the template file.
func GlobTo(name string) string {
	// The path is relative to the embed.FS root and must not use the OS path separator.
	return strings.Join([]string{"view", viewElem, name}, "/")
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
