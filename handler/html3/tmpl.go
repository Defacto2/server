package html3

// HTML templates for the /html3 router group.

import (
	"embed"
	"html/template"
	"strings"

	"github.com/Defacto2/server/internal/tags"
	"go.uber.org/zap"
)

type Templ string // Template name

const (
	layout           = "layout.html"
	dirs             = "dirs.html"
	files            = "files.html"
	pagination       = "pagination.html"
	subDirs          = "dirs_sub.html"
	tag        Templ = "html3_tag"
)

// Index template.
func index(logger *zap.SugaredLogger, fs embed.FS) *template.Template {
	return template.Must(template.New("").Funcs(TemplateFuncMap(logger)).ParseFS(fs,
		GlobTo(layout), GlobTo(dirs), GlobTo("index.html")))
}

// List file records template.
func list(logger *zap.SugaredLogger, fs embed.FS) *template.Template {
	return template.Must(template.New("").Funcs(TemplateFuncMap(logger)).ParseFS(fs,
		GlobTo(layout), GlobTo(files), GlobTo(pagination), GlobTo(files)))
}

// List and filter the tags template.
func listTags(logger *zap.SugaredLogger, fs embed.FS) *template.Template {
	return template.Must(template.New("").Funcs(TemplateFuncMap(logger)).ParseFS(fs,
		GlobTo(layout), GlobTo(subDirs), GlobTo("tags.html")))
}

// List the distinct groups template.
func listGroups(logger *zap.SugaredLogger, fs embed.FS) *template.Template {
	return template.Must(template.New("").Funcs(TemplateFuncMap(logger)).ParseFS(fs,
		GlobTo(layout), GlobTo(dirs), GlobTo(pagination), GlobTo("groups.html")))
}

// Template for displaying HTTP error codes and feedback.
func httpErr(logger *zap.SugaredLogger, fs embed.FS) *template.Template {
	return template.Must(template.New("").Funcs(TemplateFuncMap(logger)).ParseFS(fs,
		GlobTo(layout)))
}

func tagByName(name string) (tags.TagData, error) {
	t, err := tags.Tags.ByName(name)
	if err != nil {
		return t, err
	}
	s := strings.TrimSpace(t.Info)
	const tooSmall = 2
	if len(s) < tooSmall {
		return t, nil
	}
	t.Info = strings.ToUpper(string(s[0])) + s[1:]
	return t, nil
}
