package html3

// Package file template.go contains the HTML3 website template functions.

import (
	"database/sql"
	"embed"
	"fmt"
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
	subDirs          = "dirssub.html"
	tag        Templ = "html3_tag"
)

func emptyFS(fs embed.FS) bool {
	entries, err := fs.ReadDir(".")
	result := err != nil || len(entries) == 0
	clear(entries)
	return result
}

// GlobTo returns the path to the template file.
func GlobTo(name string) string {
	const pathSeparator = "/"
	return strings.Join([]string{"view", "html3", name}, pathSeparator)
}

// Index template.
func index(db *sql.DB, logger *zap.SugaredLogger, fs embed.FS) *template.Template {
	if emptyFS(fs) {
		return nil
	}
	return template.Must(template.New("").Funcs(TemplateFuncMap(db, logger)).ParseFS(fs,
		GlobTo(layout), GlobTo(dirs), GlobTo("index.html")))
}

// List file records template.
func list(db *sql.DB, logger *zap.SugaredLogger, fs embed.FS) *template.Template {
	if emptyFS(fs) {
		return nil
	}
	return template.Must(template.New("").Funcs(TemplateFuncMap(db, logger)).ParseFS(fs,
		GlobTo(layout), GlobTo(files), GlobTo(pagination), GlobTo(files)))
}

// List and filter the tags template.
func listTags(db *sql.DB, logger *zap.SugaredLogger, fs embed.FS) *template.Template {
	if emptyFS(fs) {
		return nil
	}
	return template.Must(template.New("").Funcs(TemplateFuncMap(db, logger)).ParseFS(fs,
		GlobTo(layout), GlobTo(subDirs), GlobTo("tags.html")))
}

// List the distinct groups template.
func listGroups(db *sql.DB, logger *zap.SugaredLogger, fs embed.FS) *template.Template {
	if emptyFS(fs) {
		return nil
	}
	return template.Must(template.New("").Funcs(TemplateFuncMap(db, logger)).ParseFS(fs,
		GlobTo(layout), GlobTo(dirs), GlobTo(pagination), GlobTo("groups.html")))
}

// Template for displaying HTTP error codes and feedback.
func httpErr(db *sql.DB, logger *zap.SugaredLogger, fs embed.FS) *template.Template {
	if emptyFS(fs) {
		return nil
	}
	return template.Must(template.New("").Funcs(TemplateFuncMap(db, logger)).ParseFS(fs,
		GlobTo(layout)))
}

func tagByName(t *tags.T, name string) (tags.TagData, error) {
	if t == nil {
		return tags.TagData{}, fmt.Errorf("html3 template tag by name %w", tags.ErrT)
	}
	data, err := t.ByName(name)
	if err != nil {
		return data, fmt.Errorf("html3 tag by name: %w", err)
	}
	s := strings.TrimSpace(data.Info)
	const tooSmall = 2
	if len(s) < tooSmall {
		return data, nil
	}
	data.Info = strings.ToUpper(string(s[0])) + s[1:]
	return data, nil
}
