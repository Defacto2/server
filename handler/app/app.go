// Package app handles the routes and views for the Defacto2 website.
package app

import (
	"embed"
	"errors"
	"html/template"
	"os"
	"path/filepath"
	"strings"

	"github.com/Defacto2/server/pkg/config"
	"github.com/Defacto2/server/pkg/helper"
	"github.com/labstack/gommon/log"
	"go.uber.org/zap"
)

const (
	app    = "app" // app is the name of the view element in the template.
	public = "public"

	BootCSS  = "/css/bootstrap.min.css" // BootCSS is the path to the minified Bootstrap 5 CSS file.
	BootCPub = public + BootCSS

	BootJS   = "/js/bootstrap.bundle.min.js" // BootJS is the path to the minified Bootstrap 5 JS file.
	BootJPub = public + BootJS

	LayoutCSS = "/css/layout.min.css" // LayoutCSS is the path to the minified layout CSS file.
	LayoutPub = public + LayoutCSS

	FAJS  = "/js/fontawesome.min.js" // FAJS is the path to the minified Font Awesome JS file.
	FAPub = public + FAJS

	PouetJS  = "/js/pouet.min.js" // PouetJS is the path to the minified Pouet JS file.
	PouetPub = public + PouetJS
)

var (
	ErrCode = errors.New("the http status code is not valid")
	ErrCxt  = errors.New("the server could not create a context")
	ErrDB   = errors.New("database connection is nil")
	ErrTmpl = errors.New("the server could not render the html template for this page")
	ErrZap  = errors.New("the zap logger cannot be nil")
)

// GlobTo returns the path to the template file.
func GlobTo(name string) string {
	// The path is relative to the embed.FS root and must not use the OS path separator.
	return strings.Join([]string{"view", app, name}, "/")
}

// Web is the configuration and status of the web app.
type Web struct {
	Brand       *[]byte            // Brand points to the Defacto2 ASCII logo.
	Import      *config.Config     // Import configurations from the host system environment.
	Logger      *zap.SugaredLogger // Logger is the zap sugared logger.
	Subresource SRI                // SRI are the Subresource Integrity hashes for the layout.
	Public      embed.FS           // Public facing files.
	View        embed.FS           // Views are Go templates.
}

type (
	filename string // filename is the name of the template file in the view directory.
)

// Tmpl returns a map of the templates used by the route.
func (web *Web) Tmpl() (map[string]*template.Template, error) {
	if err := web.Subresource.Verify(web.Public); err != nil {
		return nil, err
	}
	const r, s = "releaser.tmpl", "scener.tmpl"
	list := map[string]filename{
		"index":       "index.tmpl",
		"about":       "about.tmpl",
		"bbs":         r,
		"coder":       s,
		"file":        "file.tmpl",
		"files":       "files.tmpl",
		"ftp":         r,
		"history":     "history.tmpl",
		"interview":   "interview.tmpl",
		"magazine":    "releaser_year.tmpl",
		"magazine-az": r,
		"reader":      "reader.tmpl",
		"releaser":    r,
		"scener":      s,
		"searchList":  "searchList.tmpl",
		"searchPost":  "searchPost.tmpl",
		"status":      "status.tmpl",
		"thanks":      "thanks.tmpl",
		"thescene":    "the_scene.tmpl",
		"websites":    "websites.tmpl",
	}
	tmpls := make(map[string]*template.Template)
	for k, name := range list {
		tmpl, err := web.tmpl(name)
		if err != nil {
			return nil, err
		}
		tmpls[k] = tmpl
	}
	return tmpls, nil
}

// Web tmpl returns a layout template for the given named view.
// Note that the name is relative to the view/defaults directory.
func (web Web) tmpl(name filename) (*template.Template, error) {
	const (
		fileExp    = "file_expand.tmpl"
		layout     = "layout.tmpl"
		modal      = "modal.tmpl"
		pagination = "pagination.tmpl"
		website    = "website.tmpl"
	)
	fp := filepath.Join("view", app, string(name))
	if _, err := os.Stat(fp); os.IsNotExist(err) {
		log.Errorf("tmpl template not found, %s: %q", err, fp)
		return nil, err
	} else if err != nil {
		log.Errorf("tmpl template has a problem: %s", err)
		return nil, err
	}
	files := []string{GlobTo(layout), GlobTo(pagination), GlobTo(string(name)), GlobTo(modal)}
	// append any additional templates
	switch name {
	case "about.tmpl":
		files = append(files, GlobTo("about_table.tmpl"))
	case "file.tmpl":
		files = append(files, GlobTo(fileExp))
	case "websites.tmpl":
		files = append(files, GlobTo(website))
	}
	return template.Must(
		template.New("").Funcs(web.TemplateFuncMap()).ParseFS(web.View, files...)), nil
}

// SRI are the Subresource Integrity hashes for the layout.
type SRI struct {
	BootstrapCSS string // Bootstrap CSS verification hash.
	BootstrapJS  string // Bootstrap JS verification hash.
	FontAwesome  string // Font Awesome verification hash.
	LayoutCSS    string // Layout CSS verification hash.
	PouetJS      string // Pouet JS verification hash.
}

// Verify checks the integrity of the embedded CSS and JS files.
// These are required for Subresource Integrity (SRI) verification in modern browsers.
func (s *SRI) Verify(fs embed.FS) error {
	var err error
	s.BootstrapCSS, err = helper.Integrity(BootCPub, fs)
	if err != nil {
		return err
	}
	s.BootstrapJS, err = helper.Integrity(BootJPub, fs)
	if err != nil {
		return err
	}
	s.FontAwesome, err = helper.Integrity(FAPub, fs)
	if err != nil {
		return err
	}
	s.LayoutCSS, err = helper.Integrity(LayoutPub, fs)
	if err != nil {
		return err
	}
	s.PouetJS, err = helper.Integrity(PouetPub, fs)
	if err != nil {
		return err
	}
	return nil
}
