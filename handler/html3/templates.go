package html3

// HTML templates for the /html3 router group.

import (
	"embed"
	"errors"
	"fmt"
	"html/template"
	"io"
	"path/filepath"

	"github.com/Defacto2/server/model"
	"github.com/Defacto2/server/pkg/tags"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
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
func TemplateFuncMap(log *zap.SugaredLogger) template.FuncMap {
	return template.FuncMap{
		"descript": Description,
		"linkHref": func(id int64) string {
			return FileHref(id, log)
		},
		"linkPad":    FileLinkPad,
		"linkFile":   Filename,
		"leading":    Leading,
		"byteFmt":    LeadFS,
		"byteInt":    LeadFSInt,
		"leadInt":    LeadInt,
		"leadStr":    LeadStr,
		"metaByName": tagByName,
		"icon":       model.Icon,
		"publish":    model.PublishedFW,
		"posted":     model.Created,
		"safeHTML": func(s string) template.HTML {
			return template.HTML(s)
		},
	}
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
func TmplHTML3(log *zap.SugaredLogger, fs embed.FS) map[string]*template.Template {
	templates := make(map[string]*template.Template)
	templates["index"] = index(log, fs)
	templates["all"] = list(log, fs)
	templates["art"] = list(log, fs)
	templates["documents"] = list(log, fs)
	templates["software"] = list(log, fs)
	templates["groups"] = listGroups(log, fs)
	templates["group"] = list(log, fs)
	templates["tag"] = listTags(log, fs)
	templates["platform"] = list(log, fs)
	templates["category"] = list(log, fs)
	templates["error"] = httpErr(log, fs)
	return templates
}

func GlobTo(name string) string {
	return filepath.Join("view", "html3", name)
}

// Index template.
func index(log *zap.SugaredLogger, fs embed.FS) *template.Template {
	return template.Must(template.New("").Funcs(TemplateFuncMap(log)).ParseFS(fs,
		GlobTo(layout), GlobTo(dirs), GlobTo("index.html")))
}

// List file records template.
func list(log *zap.SugaredLogger, fs embed.FS) *template.Template {
	return template.Must(template.New("").Funcs(TemplateFuncMap(log)).ParseFS(fs,
		GlobTo(layout), GlobTo(files), GlobTo(pagination), GlobTo(files)))
}

// List and filter the tags template.
func listTags(log *zap.SugaredLogger, fs embed.FS) *template.Template {
	return template.Must(template.New("").Funcs(TemplateFuncMap(log)).ParseFS(fs,
		GlobTo(layout), GlobTo(dirs), GlobTo("tags.html")))
}

// List the distinct groups template.
func listGroups(log *zap.SugaredLogger, fs embed.FS) *template.Template {
	return template.Must(template.New("").Funcs(TemplateFuncMap(log)).ParseFS(fs,
		GlobTo(layout), GlobTo(dirs), GlobTo("groups.html")))
}

// Template for displaying HTTP error codes and feedback.
func httpErr(log *zap.SugaredLogger, fs embed.FS) *template.Template {
	return template.Must(template.New("").Funcs(TemplateFuncMap(log)).ParseFS(fs,
		GlobTo(layout)))
}
