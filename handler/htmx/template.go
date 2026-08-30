package htmx

// Package file template.go provides functions for rendering HTML templates.

import (
	"html/template"
	"io/fs"
	"reflect"
	"strconv"
	"strings"

	"github.com/Defacto2/helper"
	"github.com/Defacto2/server/handler/app"
	"github.com/Defacto2/server/handler/releaser"
	"github.com/Defacto2/server/handler/releaser/initialism"
	"github.com/Defacto2/server/handler/releaser/name"
	"github.com/Defacto2/server/handler/tidbit"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

// GlobTo returns the path to the template file.
func GlobTo(name string) string {
	const sep = "/"
	return strings.Join([]string{"view", "htmx", name}, sep)
}

// Templates returns a map of the templates.
func Templates(fsys fs.FS) map[string]*template.Template {
	t := make(map[string]*template.Template)
	t["searchids"] = ids(fsys)
	t["searchreleasers"] = releasers(fsys)
	t["datalistreleasers"] = datalistReleasers(fsys)

	return t
}

func emptyFS(fsys fs.FS) bool {
	if fsys == nil {
		return true
	}

	entries, err := fs.ReadDir(fsys, ".")
	return err != nil || len(entries) == 0
}

func ids(fsys fs.FS) *template.Template {
	if emptyFS(fsys) {
		return nil
	}

	patterns := []string{GlobTo("layout.tmpl"), GlobTo("searchids.tmpl")}
	return template.Must(template.New("").Funcs(TemplateFuncMap()).ParseFS(fsys, patterns...))
}

func releasers(fsys fs.FS) *template.Template {
	if emptyFS(fsys) {
		return nil
	}

	patterns := []string{GlobTo("layout.tmpl"), GlobTo("searchreleasers.tmpl")}
	return template.Must(template.New("").Funcs(TemplateFuncMap()).ParseFS(fsys, patterns...))
}

func datalistReleasers(fsys fs.FS) *template.Template {
	if emptyFS(fsys) {
		return nil
	}

	patterns := []string{GlobTo("layout.tmpl"), GlobTo("datalistreleasers.tmpl")}
	return template.Must(template.New("").Funcs(TemplateFuncMap()).ParseFS(fsys, patterns...))
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
		"mark": func(highlight, s string) template.HTML {
			return template.HTML(app.MarkAll(highlight, s))
		},
		"linkRelrs":     app.LinkReleasers,
		"obfuscateID":   helper.ObfuscateID,
		"releaserIndex": releaser.Index,
		"safeHTML": func(s string) template.HTML {
			return template.HTML(s)
		},
		"searchResult": SearchResult,
		"state": func(isNotDeleted, noDeleterInfo bool) template.HTML {
			if !isNotDeleted && !noDeleterInfo {
				const span = `<span title="Removed from public">🚫</span>`
				return span
			}
			if !isNotDeleted && noDeleterInfo {
				const span = `<span title="Not approved">⛔</span>`
				return span
			}
			return ""
		},
		"suggestion": Suggestion,
		"tidbits": func(id int) template.HTML {
			return tidbit.ID(id).URL("")
		},
	}
}

func SearchResult(index any) template.HTML {
	ind := 0 //nolint:wastedassign
	switch v := index.(type) {
	case int:
		ind = v
	default:
		return ""
	}
	const limit = 13
	if ind < 0 || ind > limit {
		return ""
	}
	const i9, i10, i11, i12, i13 = 9, 10, 11, 12, 13
	switch ind {
	case i9:
		return "0"
	case i10:
		return "-"
	case i11:
		return "="
	case i12:
		return "["
	case i13:
		return "]"
	default:
		return template.HTML(strconv.Itoa(ind + 1))
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
