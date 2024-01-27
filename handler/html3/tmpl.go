package html3

// HTML templates for the /html3 router group.

import (
	"embed"
	"html/template"
	"strings"

	"github.com/Defacto2/releaser"
	"github.com/Defacto2/server/internal/tags"
	"github.com/Defacto2/server/model"
	"go.uber.org/zap"
)

const (
	layout     = "layout.html"
	dirs       = "dirs.html"
	files      = "files.html"
	pagination = "pagination.html"
	subDirs    = "dirs_sub.html"
)

// TemplateFuncMap are a collection of mapped functions that can be used in a template.
func TemplateFuncMap(z *zap.SugaredLogger) template.FuncMap {
	return template.FuncMap{
		"descript": Description,
		"linkHref": func(id int64) string {
			return FileHref(z, id)
		},
		"linkPad":    FileLinkPad,
		"linkFile":   Filename,
		"leading":    Leading,
		"fmtByte":    LeadFS,
		"fmtURI":     releaser.Link,
		"byteInt":    LeadFSInt,
		"leadInt":    LeadInt,
		"leadStr":    LeadStr,
		"metaByName": tagByName,
		"icon":       model.Icon,
		"publish":    model.PublishedFW,
		"posted":     model.Created,
		"safeHTML": func(s string) template.HTML {
			return template.HTML(s) //nolint:gosec
		},
	}
}

func tagByName(name string) tags.TagData {
	return tags.Tags.ByName(nil, name)
}

// Templates returns a map of the templates used by the HTML3 sub-group route.
func Templates(z *zap.SugaredLogger, fs embed.FS) map[string]*template.Template {
	t := make(map[string]*template.Template)
	t["html3_index"] = index(z, fs)
	t["html3_all"] = list(z, fs)
	t["html3_art"] = list(z, fs)
	t["html3_documents"] = list(z, fs)
	t["html3_software"] = list(z, fs)
	t["html3_groups"] = listGroups(z, fs)
	t["html3_group"] = list(z, fs)
	t["html3_tag"] = listTags(z, fs)
	t["html3_platform"] = list(z, fs)
	t["html3_category"] = list(z, fs)
	t["html3_error"] = httpErr(z, fs)
	return t
}

func GlobTo(name string) string {
	// note: the path is relative to the embed.FS root and must not use the OS path separator.
	return strings.Join([]string{"view", "html3", name}, "/")
}

// Index template.
func index(z *zap.SugaredLogger, fs embed.FS) *template.Template {
	return template.Must(template.New("").Funcs(TemplateFuncMap(z)).ParseFS(fs,
		GlobTo(layout), GlobTo(dirs), GlobTo("index.html")))
}

// List file records template.
func list(z *zap.SugaredLogger, fs embed.FS) *template.Template {
	return template.Must(template.New("").Funcs(TemplateFuncMap(z)).ParseFS(fs,
		GlobTo(layout), GlobTo(files), GlobTo(pagination), GlobTo(files)))
}

// List and filter the tags template.
func listTags(z *zap.SugaredLogger, fs embed.FS) *template.Template {
	return template.Must(template.New("").Funcs(TemplateFuncMap(z)).ParseFS(fs,
		GlobTo(layout), GlobTo(subDirs), GlobTo("tags.html")))
}

// List the distinct groups template.
func listGroups(z *zap.SugaredLogger, fs embed.FS) *template.Template {
	return template.Must(template.New("").Funcs(TemplateFuncMap(z)).ParseFS(fs,
		GlobTo(layout), GlobTo(dirs), GlobTo(pagination), GlobTo("groups.html")))
}

// Template for displaying HTTP error codes and feedback.
func httpErr(z *zap.SugaredLogger, fs embed.FS) *template.Template {
	return template.Must(template.New("").Funcs(TemplateFuncMap(z)).ParseFS(fs,
		GlobTo(layout)))
}
