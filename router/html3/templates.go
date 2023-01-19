package html3

// HTML templates for the /html3 router group.

import (
	"embed"
	"errors"
	"fmt"
	"html/template"
	"io"
	"path/filepath"

	"github.com/Defacto2/server/models"
	"github.com/Defacto2/server/tags"
	"github.com/labstack/echo/v4"
)

var (
	ErrNoTmpl = errors.New("no template name exists for recordsby type index")
	ErrTmpl   = errors.New("named template cannot be found")
)

// TemplateRegistry is template registry struct.
type TemplateRegistry struct {
	Templates map[string]*template.Template
}

const (
	layout     = "layout.html"
	dirs       = "dirs.html"
	files      = "files.html"
	pagination = "pagination.html"
)

// TemplateFuncMap are a collection of mapped functions that can be used in a template.
var TemplateFuncMap = template.FuncMap{
	"descript":   Description,
	"linkHref":   FileHref,
	"linkPad":    FileLinkPad,
	"linkFile":   Filename,
	"leading":    Leading,
	"byteFmt":    LeadFS,
	"leadInt":    LeadInt,
	"datePost":   LeadPost,
	"datePub":    LeadPub,
	"leadStr":    LeadStr,
	"iconFmt":    models.Icon,
	"metaByName": tagByName,
	"safeHTML": func(s string) template.HTML {
		return template.HTML(s)
	},
}

func tagByName(name string) tags.TagData {
	return tags.Tags.ByName(name, nil)
}

// Render the HTML3 layout template with the core HTML, META and BODY elements.
func (t *TemplateRegistry) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	if name == "" {
		return ErrNoTmpl
	}
	tmpl, ok := t.Templates[name]
	if !ok {
		return fmt.Errorf("%w: %s", ErrTmpl, name)
	}
	return tmpl.ExecuteTemplate(w, "layout", data)
}

// TmplHTML3 returns a map of the templates used by the HTML3 sub-group route.
func TmplHTML3(fs embed.FS) map[string]*template.Template {
	templates := make(map[string]*template.Template)
	templates["index"] = index(fs)
	templates["art"] = list(fs)
	templates["document"] = list(fs)
	templates["software"] = list(fs)
	templates["groups"] = listGroups(fs)
	templates["group"] = list(fs)
	templates["tag"] = listTags(fs)
	templates["platform"] = list(fs)
	templates["category"] = list(fs)
	templates["error"] = httpErr(fs)
	return templates
}

func GlobTo(name string) string {
	return filepath.Join("public", "views", "html3", name)
}

// Index template.
func index(fs embed.FS) *template.Template {
	return template.Must(template.New("").Funcs(TemplateFuncMap).ParseFS(fs,
		GlobTo(layout), GlobTo(dirs), GlobTo("index.html")))
}

// List file records template.
func list(fs embed.FS) *template.Template {
	return template.Must(template.New("").Funcs(TemplateFuncMap).ParseFS(fs,
		GlobTo(layout), GlobTo(files), GlobTo(pagination), GlobTo(files)))
}

// List and filter the tags template.
func listTags(fs embed.FS) *template.Template {
	return template.Must(template.New("").Funcs(TemplateFuncMap).ParseFS(fs,
		GlobTo(layout), GlobTo(dirs), GlobTo("tags.html")))
}

// List the distinct groups template.
func listGroups(fs embed.FS) *template.Template {
	return template.Must(template.New("").Funcs(TemplateFuncMap).ParseFS(fs,
		GlobTo(layout), GlobTo(dirs), GlobTo("groups.html")))
}

// Template for displaying HTTP error codes and feedback.
func httpErr(fs embed.FS) *template.Template {
	return template.Must(template.New("").Funcs(TemplateFuncMap).ParseFS(fs,
		GlobTo(layout)))
}
