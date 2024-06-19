package htmx

// Package file template.go provides functions for rendering HTML templates.

import (
	"embed"
	"html/template"
	"reflect"
	"strings"

	"github.com/Defacto2/releaser"
	"github.com/Defacto2/releaser/initialism"
	"github.com/Defacto2/releaser/name"
	"github.com/Defacto2/server/handler/app"
	"github.com/Defacto2/server/internal/helper"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

// GlobTo returns the path to the template file.
func GlobTo(name string) string {
	const pathSeparator = "/"
	return strings.Join([]string{"view", "htmx", name}, pathSeparator)
}

// Templates returns a map of the templates.
func Templates(fs embed.FS) map[string]*template.Template {
	t := make(map[string]*template.Template)
	t["searchids"] = ids(fs)
	t["searchreleasers"] = releasers(fs)
	t["datalistreleasers"] = datalistReleasers(fs)
	return t
}

func ids(fs embed.FS) *template.Template {
	return template.Must(template.New("").Funcs(TemplateFuncMap()).ParseFS(fs,
		GlobTo("layout.tmpl"), GlobTo("searchids.tmpl")))
}

func releasers(fs embed.FS) *template.Template {
	return template.Must(template.New("").Funcs(TemplateFuncMap()).ParseFS(fs,
		GlobTo("layout.tmpl"), GlobTo("searchreleasers.tmpl")))
}

func datalistReleasers(fs embed.FS) *template.Template {
	return template.Must(template.New("").Funcs(TemplateFuncMap()).ParseFS(fs,
		GlobTo("layout.tmpl"), GlobTo("datalistreleasers.tmpl")))
}

// TemplateFuncMap are a collection of mapped functions that can be used in a template.
func TemplateFuncMap() template.FuncMap {
	return template.FuncMap{
		"borderClass": func(name, path string) string {
			const mark = "border border-primary"
			if strings.EqualFold(name, path) {
				return mark
			}
			init := initialism.Join(initialism.Path(path))
			if strings.EqualFold(name, init) {
				return mark
			}
			return "border"
		},
		"byteCount": helper.ByteCount,
		"byteFileS": app.ByteFileS,
		"describe":  app.Describe,
		"fmtPath": func(path string) string {
			if val := name.Path(path); val.String() != "" {
				return val.String()
			}
			return releaser.Humanize(path)
		},
		"initialisms": func(s string) string {
			return initialism.Join(initialism.Path(s))
		},
		"linkRelrs":   app.LinkRelrs,
		"obfuscateID": helper.ObfuscateID,
		"safeHTML": func(s string) template.HTML {
			return template.HTML(s)
		},
		"state": func(deleteat, deleteby bool) template.HTML {
			if !deleteat && deleteby {
				return "<span title=\"Not approved\">⛔</span>"
			}
			if !deleteat && !deleteby {
				return "<span title=\"Removed from public\">🚫</span>"
			}
			return ""
		},
		"suggestion": Suggestion,
	}
}

// Suggestion returns a human readable string of the byte count with a named description.
func Suggestion(name, initialism string, count any) string {
	s := name
	if initialism != "" {
		s += ", " + initialism
	}
	switch val := count.(type) {
	case int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64:
		i := reflect.ValueOf(val).Int()
		p := message.NewPrinter(language.English)
		s += p.Sprintf(" (%d item", i)
		if i > 1 {
			s += "s"
		}
		s += ")"
	default:
		s += "suggestion type error: " + reflect.TypeOf(count).String()
		return s
	}
	return s
}
