package app

import (
	"embed"
	"errors"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"
	"time"

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

// GlobTo returns the path to the template file.
func GlobTo(name string) string {
	// The path is relative to the embed.FS root and must not use the OS path separator.
	return strings.Join([]string{"view", viewElem, name}, "/")
}

// Configuration of the app.
type Configuration struct {
	Brand       *[]byte            // Brand points to the Defacto2 ASCII logo.
	DatbaseErr  bool               // DBErr is true if the database connection failed.
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
	templates["artist"] = c.tmpl("artist.html")
	templates["bbs"] = c.tmpl("bbs.html")
	templates["coder"] = c.tmpl("coder.html")
	templates["file"] = c.tmpl("file.html")
	templates["ftp"] = c.tmpl("ftp.html")
	templates["history"] = c.tmpl("history.html")
	templates["interview"] = c.tmpl("interview.html")
	templates["magazine"] = c.tmpl("magazine.html")
	templates["musician"] = c.tmpl("musician.html")
	templates["releaser"] = c.tmpl("releaser.html")
	templates["scener"] = c.tmpl("scener.html")
	templates["status"] = c.tmpl("status.html")
	templates["thanks"] = c.tmpl("thanks.html")
	templates["thescene"] = c.tmpl("the_scene.html")
	templates["websites"] = c.tmpl("websites.html")
	templates["writer"] = c.tmpl("writer.html")
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
	switch name {
	case "file.html":
		files = append(files, GlobTo("file_expand.html"))
	case "websites.html":
		files = append(files, GlobTo("website.html"))
	}
	return template.Must(template.New("").Funcs(c.TemplateFuncMap()).ParseFS(c.Views, files...))
}

// TemplateFuncMap are a collection of mapped functions that can be used in a template.
func (c Configuration) TemplateFuncMap() template.FuncMap {
	return template.FuncMap{
		"databaseDown": func() bool {
			return c.DatbaseErr
		},
		"mergeIcon": func() string {
			return merge
		},
		"externalLink": ExternalLink,
		"logo": func() string {
			return string(*c.Brand)
		},
		"logoText": LogoText,
		"mod3": func(i int) bool {
			const x = 3
			fmt.Println(i, x, i%x == 0)
			return i%x == 0
		},
		"mod3end": func(i int) bool {
			const x = 3
			fmt.Println("->", i, x, i%x == (x-1))
			return i%x == x-1
		},
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
		"fmtPrefix": func(s string) string {
			if s == "" {
				return ""
			}
			return fmt.Sprintf("%s ", s)
		},
		"fmtMonth": func(m int) string {
			if m == 0 {
				return ""
			}
			if m < 0 || m > 12 {
				return " ERR MONTH"
			}
			return " " + time.Month(m).String()
		},
		"fmtDay": func(d int) string {
			if d == 0 {
				return ""
			}
			if d < 0 || d > 31 {
				return " ERR DAY"
			}
			return fmt.Sprintf(" %d", d)
		},
	}
}

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
