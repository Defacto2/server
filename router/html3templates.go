package router

// HTML templates for the /html3 router group.

import (
	"errors"
	"fmt"
	"html/template"
	"io"

	"github.com/Defacto2/server/models"
	"github.com/Defacto2/server/tags"
	"github.com/labstack/echo/v4"
)

var ErrTmpl = errors.New("named template cannot be found")

// TemplateRegistry is template registry struct.
type TemplateRegistry struct {
	Templates map[string]*template.Template
}

const (
	layout = "public/views/html3/layout.html"
	dirs   = "public/views/html3/dirs.html"
	files  = "public/views/html3/files.html"
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
	"metaByName": tags.TagByName,
}

// Render the HTML3 layout template with the core HTML, META and BODY elements.
func (t *TemplateRegistry) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	tmpl, ok := t.Templates[name]
	if !ok {
		return fmt.Errorf("%w: %s", ErrTmpl, name)
	}
	return tmpl.ExecuteTemplate(w, "layout", data)
}

// TmplHTML3 returns a map of the templates used by the HTML3 sub-group route.
func TmplHTML3() map[string]*template.Template {
	templates := make(map[string]*template.Template)
	templates["index"] = index()
	templates["category"] = list()
	templates["platform"] = list()
	templates["error"] = httpErr()
	templates["tag"] = tag()
	templates["group"] = groups()
	return templates
}

// Index template.
func index() *template.Template {
	return template.Must(template.New("").Funcs(TemplateFuncMap).ParseFiles(
		layout, dirs, "public/views/html3/index.html"))
}

// List files template.
func list() *template.Template {
	return template.Must(template.New("").Funcs(TemplateFuncMap).ParseFiles(
		layout, files, "public/views/html3/files.html"))
}

// Tag lists template.
func tag() *template.Template {
	return template.Must(template.New("").Funcs(TemplateFuncMap).ParseFiles(
		layout, dirs, "public/views/html3/tag.html"))
}

// Groups list template.
func groups() *template.Template {
	return template.Must(template.New("").Funcs(TemplateFuncMap).ParseFiles(
		layout, dirs, "public/views/html3/groups.html"))
}

// Template for displaying HTTP error codes and feedback.
func httpErr() *template.Template {
	return template.Must(template.New("").Funcs(TemplateFuncMap).ParseFiles(
		layout))
}
