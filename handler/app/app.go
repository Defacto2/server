package app

import (
	"embed"
	"errors"
	"html/template"
	"os"
	"path/filepath"
	"strings"

	"github.com/labstack/gommon/log"
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

// Configuration of the app.
type Configuration struct {
	Brand       *[]byte            // Brand points to the Defacto2 ASCII logo.
	Log         *zap.SugaredLogger // Log is a sugared logger.
	Subresource SRI                // SRI are the Subresource Integrity hashes for the layout.
	Public      embed.FS           // Public facing files.
	Views       embed.FS           // Views are Go templates.
}

// Tmpl returns a map of the templates used by the route.
func (c *Configuration) Tmpl() map[string]*template.Template {
	if err := c.Subresource.Verify(c.Public); err != nil {
		panic(err)
	}
	templates := make(map[string]*template.Template)
	templates["index"] = c.tmpl("index.html")
	templates["file"] = c.tmpl("file.html")
	templates["history"] = c.tmpl("history.html")
	templates["thanks"] = c.tmpl("thanks.html")
	templates["thescene"] = c.tmpl("thescene.html")
	templates["websites"] = c.tmpl("websites.html")
	templates["error"] = c.httpErr()
	return templates
}

// tmpl returns a layout template for the given named view.
// Note that the name is relative to the view/defaults directory
func (c Configuration) tmpl(name string) *template.Template {
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
	return template.Must(template.New("").Funcs(c.TemplateFuncMap()).ParseFS(c.Views, files...))
}

// httpErr is the template for displaying HTTP error codes and feedback.
func (c Configuration) httpErr() *template.Template {
	return template.Must(template.New("").Funcs(c.TemplateFuncMap()).ParseFS(c.Views,
		GlobTo(layout), GlobTo("error.html"), GlobTo(modal)))
}

// GlobTo returns the path to the template file.
func GlobTo(name string) string {
	// The path is relative to the embed.FS root and must not use the OS path separator.
	return strings.Join([]string{"view", viewElem, name}, "/")
}

// TemplateFuncMap are a collection of mapped functions that can be used in a template.
func (c Configuration) TemplateFuncMap() template.FuncMap {
	return template.FuncMap{
		"externalLink": ExternalLink,
		"logo": func() string {
			return string(*c.Brand)
		},
		"logoText": LogoText,
		"wikiLink": WikiLink,
		"sriBootstrapCSS": func() string {
			return c.Subresource.BootstrapCSS
		},
		"sriBootstrapJS": func() string {
			return c.Subresource.BootstrapJS
		},
		"sriFontAwesome": func() string {
			return c.Subresource.FontAwesome
		},
		"sriLayoutCSS": func() string {
			return c.Subresource.LayoutCSS
		},
		"safeHTML": func(s string) template.HTML {
			return template.HTML(s)
		},
	}
}
