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

const (
	SessionName = "d2_op" // SessionName is the name given to the session cookie.
)

var (
	ErrCode = errors.New("the http status code is not valid")
	ErrCxt  = errors.New("the server could not create a context")
	ErrDB   = errors.New("database connection is nil")
	ErrTmpl = errors.New("the server could not render the html template for this page")
	ErrZap  = errors.New("the zap logger cannot be nil")
)

// Cache contains database values that are used throughout the app or layouts.
type Cache struct {
	RecordCount int // The total number of file records in the database.
}

// Asset is a relative path to a public facing CSS, JS or WASM file.
type Asset int

const (
	Bootstrap   Asset = iota // Bootstrap is the path to the minified Bootstrap 5.3 CSS file.
	BootstrapJS              // BootstrapJS is the path to the minified Bootstrap 5.3 JS file.
	Editor                   // Editor is the path to the minified Editor JS file.
	EditAssets               // EditAssets is the path to the minified Editor assets JS file.
	EditArchive              // EditArchive is the path to the minified Editor archive JS file.
	FontAwesome              // FontAwesome is the path to the minified Font Awesome 3 JS file.
	JSDosUI                  // JSDosUI is the path to the minified JS DOS user-interface JS file.
	JSDosW                   // JSDosW is the JS DOS default variant compiled with emscripten.
	JSDosWasm                // JSDOSWasm is the JS DOS WASM binary file.
	Layout                   // Layout is the path to the minified layout CSS file.
	Pouet                    // Pouet is the path to the minified Pouet JS file.
	Readme                   // Readme is the path to the minified Readme JS file.
	RESTPouet                // RESTPouet is the path to the minified Pouet REST JS file.
	RESTZoo                  // RESTZoo is the path to the minified Demozoo REST JS file.
	Uploader                 // Uploader is the path to the minified Uploader JS file.
)

// Paths are a map of the public facing CSS, JS and WASM files.
type Paths map[Asset]string

// Hrefs returns the relative path of the public facing CSS, JS and WASM files.
// The strings are intended for href attributes in HTML link elements and
// the src attribute in HTML script elements.
func Hrefs() Paths {
	// note, the js-dos (JS DOS v6) are minified files,
	// help: https://js-dos.com/6.22/examples/?arkanoid
	return Paths{
		Bootstrap:   "/css/bootstrap.min.css",
		BootstrapJS: "/js/bootstrap.bundle.min.js",
		Editor:      "/js/editor.min.js",
		EditAssets:  "/js/editor-assets.min.js",
		EditArchive: "/js/editor-archive.min.js",
		FontAwesome: "/js/fontawesome.min.js",
		JSDosW:      "/js/wdosbox.js",
		JSDosWasm:   "/js/wdosbox.wasm",
		JSDosUI:     "/js/js-dos.js",
		Layout:      "/css/layout.min.css",
		Pouet:       "/js/pouet.min.js",
		Readme:      "/js/readme.min.js",
		RESTPouet:   "/js/rest-pouet.min.js",
		RESTZoo:     "/js/rest-zoo.min.js",
		Uploader:    "/js/uploader.min.js",
	}
}

// Names returns the absolute path of the public facing CSS, JS and WASM files
// relative to the embed.FS root.
func Names() Paths {
	const public = "public"
	href := Hrefs()
	// iterate and return HRefs() with the public prefix
	paths := make(Paths, len(href))
	for k, v := range href {
		paths[k] = public + v
	}
	return paths
}

// Font is a relative path to a public facing font file.
type Font int

const (
	VGA8      Font = iota // VGA8 is the path to the IBM VGA 8px font file.
	VGA8Woff              // VGA8Woff is the path to the IBM VGA 8px legacy WOFF format font file.
	VGA8TT                // VGA8TT is the path to the IBM VGA 8px legacy TrueType format font file.
	A1200                 // A1200 is the path to the Topaz Plus font file.
	A1200Woff             // A1200Woff is the path to the Topaz Plus legacy WOFF format font file.
	A1200TT               // A1200TT is the path to the Topaz Plus legacy TrueType format font file.
)

// Fonts are a map of the public facing font files.
type Fonts map[Font]string

// FontRefs returns the relative path of the public facing font files.
// The strings are intended for href attributes in HTML link elements and
// the src attribute in HTML script elements.
func FontRefs() Fonts {
	return Fonts{
		VGA8:      "/pxplus_ibm_vga8.woff2",
		VGA8Woff:  "/pxplus_ibm_vga8.woff",
		VGA8TT:    "/pxplus_ibm_vga8.ttf",
		A1200:     "/topazplus_a1200.woff2",
		A1200Woff: "/topazplus_a1200.woff",
		A1200TT:   "/topazplus_a1200.ttf",
	}
}

// Names returns the absolute path of the public facing font files
// relative to the embed.FS root.
func FontNames() Fonts {
	const public = "public/font"
	href := FontRefs()
	paths := make(Fonts, len(href))
	for k, v := range href {
		paths[k] = public + v
	}
	return paths
}

// Caching are values that are used throughout the app or layouts.
var Caching = Cache{} //nolint:gochecknoglobals

// GlobTo returns the path to the template file.
func GlobTo(name string) string {
	// The path is relative to the embed.FS root and must not use the OS path separator.
	return strings.Join([]string{"view", "app", name}, "/")
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

func templates() map[string]filename {
	const releaser, scener = "releaser.tmpl", "scener.tmpl"
	return map[string]filename{
		"index":       "index.tmpl",
		"about":       "about.tmpl",
		"bbs":         releaser,
		"coder":       scener,
		"file":        "file.tmpl",
		"files":       "files.tmpl",
		"ftp":         releaser,
		"history":     "history.tmpl",
		"interview":   "interview.tmpl",
		"magazine":    "releaser_year.tmpl",
		"magazine-az": releaser,
		"reader":      "reader.tmpl",
		"releaser":    releaser,
		"scener":      scener,
		"searchList":  "searchList.tmpl",
		"searchPost":  "searchPost.tmpl",
		"signin":      "signin.tmpl",
		"signout":     "signout.tmpl",
		"status":      "status.tmpl",
		"thanks":      "thanks.tmpl",
		"thescene":    "the_scene.tmpl",
		"websites":    "websites.tmpl",
	}
}

// Templates returns a map of the templates used by the route.
func (web *Web) Templates() (map[string]*template.Template, error) {
	if err := web.Subresource.Verify(web.Public); err != nil {
		return nil, err
	}
	tmpls := make(map[string]*template.Template)
	for k, name := range templates() {
		tmpl := web.tmpl(name)
		tmpls[k] = tmpl
	}
	return tmpls, nil
}

// Web tmpl returns a layout template for the given named view.
// Note that the name is relative to the view/defaults directory.
func (web Web) tmpl(name filename) *template.Template {
	files := []string{
		GlobTo("layout.tmpl"),
		GlobTo("modal.tmpl"),
		GlobTo("option_os.tmpl"),
		GlobTo("option_tag.tmpl"),
		GlobTo(string(name)),
		GlobTo("pagination.tmpl"),
	}
	config := web.Import
	files = uploaderTmpls(config.ReadMode, files...)
	// append any additional and embedded templates
	switch name {
	case "about.tmpl":
		files = aboutTmpls(config.ReadMode, files...)
	case "file.tmpl":
		files = append(files, GlobTo("file_expand.tmpl"))
	case "websites.tmpl":
		const individualWebsite = "website.tmpl"
		files = append(files, GlobTo(individualWebsite))
	}
	return template.Must(template.New("").Funcs(
		web.TemplateFuncMap()).ParseFS(web.View, files...))
}

func uploaderTmpls(lock bool, files ...string) []string {
	if lock {
		return append(files,
			GlobTo("layout_editor_null.tmpl"),
			GlobTo("layout_uploader_null.tmpl"),
			GlobTo("uploader_null.tmpl"))
	}
	return append(files,
		GlobTo("layout_editor.tmpl"),
		GlobTo("layout_uploader.tmpl"),
		GlobTo("uploader.tmpl"))
}

func aboutTmpls(lock bool, files ...string) []string {
	files = append(files,
		GlobTo("about_table.tmpl"),
		GlobTo("about_jsdos.tmpl"),
		GlobTo("about_editor_archive.tmpl"))
	if lock {
		return append(files,
			GlobTo("about_editor_null.tmpl"),
			GlobTo("about_editor_table_null.tmpl"),
			GlobTo("about_table_switch_null.tmpl"))
	}
	return append(files,
		GlobTo("about_editor.tmpl"),
		GlobTo("about_editor_table.tmpl"),
		GlobTo("about_table_switch.tmpl"))
}

// SRI are the Subresource Integrity hashes for the layout.
type SRI struct {
	Bootstrap   string // Bootstrap CSS verification hash.
	BootstrapJS string // Bootstrap JS verification hash.
	Editor      string // Editor JS verification hash.
	EditAssets  string // Editor Assets JS verification hash.
	EditArchive string // Editor Archive JS verification hash.
	FontAwesome string // Font Awesome verification hash.
	JSDosUI     string // JS DOS verification hash.
	JSDosW      string // JS DOS emscripten verification hash.
	Layout      string // Layout CSS verification hash.
	Pouet       string // Pouet JS verification hash.
	Readme      string // Readme JS verification hash.
	RESTPouet   string // Pouet REST JS verification hash.
	RESTZoo     string // Demozoo REST JS verification hash.
	Uploader    string // Uploader JS verification hash.
}

// Verify checks the integrity of the embedded CSS and JS files.
// These are required for Subresource Integrity (SRI) verification in modern browsers.
func (s *SRI) Verify(fs embed.FS) error {
	names := Names()
	var err error
	s.Bootstrap, err = helper.Integrity(names[Bootstrap], fs)
	if err != nil {
		return err
	}
	s.BootstrapJS, err = helper.Integrity(names[BootstrapJS], fs)
	if err != nil {
		return err
	}
	s.Editor, err = helper.Integrity(names[Editor], fs)
	if err != nil {
		return err
	}
	s.EditAssets, err = helper.Integrity(names[EditAssets], fs)
	if err != nil {
		return err
	}
	s.EditArchive, err = helper.Integrity(names[EditArchive], fs)
	if err != nil {
		return err
	}
	s.FontAwesome, err = helper.Integrity(names[FontAwesome], fs)
	if err != nil {
		return err
	}
	s.JSDosUI, err = helper.Integrity(names[JSDosUI], fs)
	if err != nil {
		return err
	}
	s.JSDosW, err = helper.Integrity(names[JSDosW], fs)
	if err != nil {
		return err
	}
	s.Layout, err = helper.Integrity(names[Layout], fs)
	if err != nil {
		return err
	}
	s.Pouet, err = helper.Integrity(names[Pouet], fs)
	if err != nil {
		return err
	}
	s.Readme, err = helper.Integrity(names[Readme], fs)
	if err != nil {
		return err
	}
	s.RESTPouet, err = helper.Integrity(names[RESTPouet], fs)
	if err != nil {
		return err
	}
	s.RESTZoo, err = helper.Integrity(names[RESTZoo], fs)
	if err != nil {
		return err
	}
	s.Uploader, err = helper.Integrity(names[Uploader], fs)
	if err != nil {
		return err
	}
	return nil
}
