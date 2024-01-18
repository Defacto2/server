// Package app handles the routes and views for the Defacto2 website.
package app

import (
	"embed"
	"errors"
	"html/template"
	"strings"

	"github.com/Defacto2/server/internal/config"
	"github.com/Defacto2/server/internal/helper"
	"go.uber.org/zap"
)

// Cache contains database values that are used throughout the app or layouts.
type Cache struct {
	RecordCount int // The total number of file records in the database.
}

const (
	app    = "app" // app is the name of the view element in the template.
	public = "public"

	BootCSS  = "/css/bootstrap.min.css" // BootCSS is the path to the minified Bootstrap 5 CSS file.
	BootCPub = public + BootCSS

	BootJS   = "/js/bootstrap.bundle.min.js" // BootJS is the path to the minified Bootstrap 5 JS file.
	BootJPub = public + BootJS

	EditorJS           = "/js/editor.min.js" // EditorJS is the path to the minified Editor JS file.
	EditorJSPub        = public + EditorJS
	EditorAssetsJS     = "/js/editor-assets.min.js" // EditorAssetsJS is the path to the minified Editor assets JS file.
	EditorAssetsJSPub  = public + EditorAssetsJS
	EditorArchiveJS    = "/js/editor-archive.min.js" // EditorArchiveJS is the path to the minified Editor archive JS file.
	EditorArchiveJSPub = public + EditorArchiveJS
	FAJS               = "/js/fontawesome.min.js" // FAJS is the path to the minified Font Awesome JS file.
	FAPub              = public + FAJS

	// JS DOS v6 are minified files.
	// https://js-dos.com/6.22/examples/?arkanoid
	JSDos     = "/js/js-dos.js"
	JSDosPub  = public + JSDos
	JSWDos    = "/js/wdosbox.js"
	JSWDosPub = public + JSWDos

	LayoutCSS = "/css/layout.min.css" // LayoutCSS is the path to the minified layout CSS file.
	LayoutPub = public + LayoutCSS

	PouetJS  = "/js/pouet.min.js" // PouetJS is the path to the minified Pouet JS file.
	PouetPub = public + PouetJS

	ReadmeJS  = "/js/readme.min.js" // ReadmeJS is the path to the minified Readme JS file.
	ReadmePub = public + ReadmeJS

	RestPouetJS  = "/js/rest-pouet.min.js" // RestPouetJS is the path to the Pouet REST JS file.
	RestPouetPub = public + RestPouetJS
	RestZooJS    = "/js/rest-zoo.min.js" // RestZooJS is the path to the Demozoo REST JS file.
	RestZooPub   = public + RestZooJS

	UploaderJS  = "/js/uploader.min.js" // UploaderJS is the path to the minified Uploader JS file.
	UploaderPub = public + UploaderJS
)

// Caching are values that are used throughout the app or layouts.
var Caching = Cache{}

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
	Version     string             // Version is the current version of the app.
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
		"signin":      "signin.tmpl",
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
		optionOS   = "option_os.tmpl"
		optionTag  = "option_tag.tmpl"
		pagination = "pagination.tmpl"
		website    = "website.tmpl"
		uploader   = "uploader.tmpl"
	)
	files := []string{
		GlobTo(layout),
		GlobTo(modal),
		GlobTo(optionOS),
		GlobTo(optionTag),
		GlobTo(string(name)),
		GlobTo(pagination),
		GlobTo(uploader),
	}
	if web.Import.IsReadOnly {
		files = append(files, GlobTo("layout_editor_null.tmpl"))
	} else {
		files = append(files, GlobTo("layout_editor.tmpl"))
	}
	// append any additional templates
	switch name {
	case "about.tmpl":
		files = append(files, GlobTo("about_table.tmpl"), GlobTo("about_jsdos.tmpl"))
		files = append(files, GlobTo("about_editor_archive.tmpl"))
		if web.Import.IsReadOnly {
			files = append(files, GlobTo("about_editor_null.tmpl"))
			files = append(files, GlobTo("about_editor_table_null.tmpl"))
			files = append(files, GlobTo("about_table_switch_null.tmpl"))
		} else {
			files = append(files, GlobTo("about_editor.tmpl"))
			files = append(files, GlobTo("about_editor_table.tmpl"))
			files = append(files, GlobTo("about_table_switch.tmpl"))
		}
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
	BootstrapCSS    string // Bootstrap CSS verification hash.
	BootstrapJS     string // Bootstrap JS verification hash.
	EditorJS        string // Editor JS verification hash.
	EditorAssetsJS  string // Editor Assets JS verification hash.
	EditorArchiveJS string // Editor Archive JS verification hash.
	FontAwesome     string // Font Awesome verification hash.
	JSDos           string // JS DOS verification hash.
	JSWDos          string // JS wasm verification hash.
	LayoutCSS       string // Layout CSS verification hash.
	PouetJS         string // Pouet JS verification hash.
	ReadmeJS        string // Readme JS verification hash.
	RestPouetJS     string // Pouet REST JS verification hash.
	RestZooJS       string // Demozoo REST JS verification hash.
	UploaderJS      string // Uploader JS verification hash.
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
	s.EditorJS, err = helper.Integrity(EditorJSPub, fs)
	if err != nil {
		return err
	}
	s.EditorAssetsJS, err = helper.Integrity(EditorAssetsJSPub, fs)
	if err != nil {
		return err
	}
	s.EditorArchiveJS, err = helper.Integrity(EditorArchiveJSPub, fs)
	if err != nil {
		return err
	}
	s.FontAwesome, err = helper.Integrity(FAPub, fs)
	if err != nil {
		return err
	}
	s.JSDos, err = helper.Integrity(JSDosPub, fs)
	if err != nil {
		return err
	}
	s.JSWDos, err = helper.Integrity(JSWDosPub, fs)
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
	s.ReadmeJS, err = helper.Integrity(ReadmePub, fs)
	if err != nil {
		return err
	}
	s.RestPouetJS, err = helper.Integrity(RestPouetPub, fs)
	if err != nil {
		return err
	}
	s.RestZooJS, err = helper.Integrity(RestZooPub, fs)
	if err != nil {
		return err
	}
	s.UploaderJS, err = helper.Integrity(UploaderPub, fs)
	if err != nil {
		return err
	}
	return nil
}
