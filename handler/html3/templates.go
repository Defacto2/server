package html3

// HTML templates for the /html3 router group.

import (
	"embed"
	"html/template"
	"path/filepath"

	"github.com/Defacto2/server/model"
	"github.com/Defacto2/server/pkg/tags"
	"go.uber.org/zap"
)

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

// TmplHTML3 returns a map of the templates used by the HTML3 sub-group route.
func TmplHTML3(log *zap.SugaredLogger, fs embed.FS) map[string]*template.Template {
	templates := make(map[string]*template.Template)
	templates["html3_index"] = index(log, fs)
	templates["html3_all"] = list(log, fs)
	templates["html3_art"] = list(log, fs)
	templates["html3_documents"] = list(log, fs)
	templates["html3_software"] = list(log, fs)
	templates["html3_groups"] = listGroups(log, fs)
	templates["html3_group"] = list(log, fs)
	templates["html3_tag"] = listTags(log, fs)
	templates["html3_platform"] = list(log, fs)
	templates["html3_category"] = list(log, fs)
	templates["html3_error"] = httpErr(log, fs)
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
