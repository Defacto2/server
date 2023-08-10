// Package app handles the routes and views for the Defacto2 website.
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
	app         = "app"                               // app is the name of the view element in the template.
	layout      = "layout.html"                       // layout is a partial template.
	modal       = "modal.html"                        // modal is a partial template.
	bootCSS     = "public/css/bootstrap.min.css"      // bootCSS is the path to the minified Bootstrap 5 CSS file.
	bootJS      = "public/js/bootstrap.bundle.min.js" // bootJS is the path to the minified Bootstrap 5 JS file.
	layoutCSS   = "public/css/layout.min.css"         // layoutCSS is the path to the minified layout CSS file.
	fontawesome = "public/js/fontawesome.min.js"      // fontawesome is the path to the minified Font Awesome JS file.

	errConn = "Sorry, at the moment the server cannot connect to the database"
)

var (
	ErrContext = errors.New("the server could not create a context")
	ErrDB      = errors.New("database connection is nil")
	ErrLogger  = errors.New("the server could not create a logger")
	ErrTmpl    = errors.New("the server could not render the HTML template for this page")
)

// GlobTo returns the path to the template file.
func GlobTo(name string) string {
	// The path is relative to the embed.FS root and must not use the OS path separator.
	return strings.Join([]string{"view", app, name}, "/")
}

// Configuration of the app.
type Configuration struct {
	Brand       *[]byte            // Brand points to the Defacto2 ASCII logo.
	DatbaseErr  bool               // DBErr is true if the database connection failed.
	ZLog        *zap.SugaredLogger // Log is a sugared logger.
	Subresource SRI                // SRI are the Subresource Integrity hashes for the layout.
	Public      embed.FS           // Public facing files.
	Views       embed.FS           // Views are Go templates.
}

// Tmpl returns a map of the templates used by the route.
func (c *Configuration) Tmpl() map[string]*template.Template {
	if err := c.Subresource.Verify(c.Public); err != nil {
		panic(err)
	}
	const relTmpl = "releaser.html"
	templates := make(map[string]*template.Template)
	templates["index"] = c.tmpl("index.html")
	templates["artist"] = c.tmpl("artist.html")
	templates["bbs"] = c.tmpl(relTmpl)
	templates["coder"] = c.tmpl("coder.html")
	templates["file"] = c.tmpl("file.html")
	templates["files"] = c.tmpl("files.html")
	templates["ftp"] = c.tmpl(relTmpl)
	templates["history"] = c.tmpl("history.html")
	templates["interview"] = c.tmpl("interview.html")
	templates["magazine"] = c.tmpl(relTmpl)
	templates["musician"] = c.tmpl("musician.html")
	templates["releaser"] = c.tmpl(relTmpl)
	templates["scener"] = c.tmpl("scener.html")
	templates["status"] = c.tmpl("status.html")
	templates["thanks"] = c.tmpl("thanks.html")
	templates["thescene"] = c.tmpl("the_scene.html")
	templates["websites"] = c.tmpl("websites.html")
	templates["writer"] = c.tmpl("writer.html")
	return templates
}

// Configuration tmpl returns a layout template for the given named view.
// Note that the name is relative to the view/defaults directory.
func (c Configuration) tmpl(name string) *template.Template {
	if _, err := os.Stat(filepath.Join("view", app, name)); os.IsNotExist(err) {
		log.Errorf("tmpl template not found: %s", err)
		panic(err)
	} else if err != nil {
		log.Errorf("tmpl template has a problem: %s", err)
		panic(err)
	}
	files := []string{GlobTo(layout), GlobTo(name), GlobTo(modal)}
	// append any additional templates
	switch name {
	case "file.html":
		files = append(files, GlobTo("file_expand.html"))
	case "websites.html":
		files = append(files, GlobTo("website.html"))
	}
	return template.Must(template.New("").Funcs(c.TemplateFuncMap()).ParseFS(c.Views, files...))
}

// SRI are the Subresource Integrity hashes for the layout.
type SRI struct {
	BootstrapCSS string // Bootstrap CSS verification hash.
	BootstrapJS  string // Bootstrap JS verification hash.
	FontAwesome  string // Font Awesome verification hash.
	LayoutCSS    string // Layout CSS verification hash.
}

// Verify checks the integrity of the embedded CSS and JS files.
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
