package app_test

import (
	"bytes"
	"database/sql"
	"html/template"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/Defacto2/server/handler/app"
	"github.com/Defacto2/server/handler/areacode"
	"github.com/nalgeon/be"
)

func TestTemplates(t *testing.T) {
	t.Parallel()

	w := app.WebApp{}
	_, err := w.Templates(t.Context(), nil)
	be.Err(t, err)
}

func TestFuncMap(t *testing.T) {
	t.Parallel()

	w := app.WebApp{}
	db := &sql.DB{}
	m := w.FuncMap(t.Context(), db)
	keys := slices.Sorted(maps.Keys(m))
	be.True(t, slices.Contains(keys, "add"))
	be.True(t, slices.Contains(keys, "version"))
	be.True(t, slices.Contains(keys, "az"))
	be.True(t, slices.Contains(keys, "msdos"))
}

func TestLinkSamples(t *testing.T) {
	t.Parallel()

	x := app.LinkPreviews("1", "2", "3", "4", "5", "6", "7")
	be.True(t, len(x) == 7)
	be.True(t, strings.Contains(x[0], "youtube.com/watch?v=1"))
	be.True(t, strings.Contains(x[1], "demozoo.org/productions/2"))
}

func TestLinkRelsPerf(t *testing.T) {
	t.Parallel()

	s := app.LinkRelsPerf("", "")
	be.Equal(t, s, "")

	s = app.LinkRelsPerf("Group 1", "Group 2")
	be.True(t, strings.Contains(string(s), "Group 1"))
	be.True(t, strings.Contains(string(s), "Group 2"))
	be.True(t, strings.Contains(string(s), `href="/g/group-1"`))
	be.True(t, strings.Contains(string(s), `href="/g/group-2"`))
}

func TestTemplTemplates(t *testing.T) {
	t.Parallel()

	tpl := app.WebApp{}
	x, err := tpl.Templates(t.Context(), nil)
	be.Err(t, err)
	be.True(t, x == nil)
}

func TestFuncClosures(t *testing.T) {
	t.Parallel()

	// tpl := app.WebApp{}
	// x := tpl.FuncClosure()
	// be.True(t, x == template.FuncMap{})
}

func TestLinkRelrs(t *testing.T) {
	t.Parallel()

	x := string(app.LinkRelrs(false, nil, nil))
	be.True(t, x == "")

	x = string(app.LinkRelsPerf(nil, nil))
	be.True(t, x == "")

	x = string(app.LinkReleasers(false, false, nil, nil))
	be.True(t, x == "")
}

func TestTempls(t *testing.T) {
	t.Parallel()

	x := app.WebApp{}
	pages := x.Pages()

	p := filepath.Join("../", "../", "view", "app")
	view, err := filepath.Abs(p)
	be.Err(t, err, nil)

	for _, page := range pages {
		be.True(t, page != "")

		ext := filepath.Ext(string(page))
		be.Equal(t, ".tmpl", ext)

		stat, err := os.Stat(filepath.Join(view, string(page)))
		be.Err(t, err, nil)
		be.True(t, stat.Size() > 0)
	}
}

func TestScenerTemplate(t *testing.T) {
	t.Parallel()

	wa := app.WebApp{}
	db := &sql.DB{}
	tmpl, err := template.New("scener.tmpl").Funcs(wa.FuncMap(t.Context(), db)).
		ParseFiles(filepath.Join("../", "../", "view", "app", "scener.tmpl"))
	be.Err(t, err, nil)

	names := []string{"007", "ACE"}
	groups, letters := app.BuildScenerIndex(names)
	var buf bytes.Buffer
	err = tmpl.ExecuteTemplate(&buf, "content", map[string]any{
		"sceners":       names,
		"scenerRole":    "coder",
		"scenerGroups":  groups,
		"scenerLetters": letters,
	})
	be.Err(t, err, nil)

	out := buf.String()
	be.True(t, strings.Contains(out, `href="/coder" rel="nofollow" aria-current="page"`))
	be.True(t, strings.Contains(out, `id="scener-sec-num"`))
	be.True(t, strings.Contains(out, `id="scener-sec-A"`))
	be.True(t, strings.Contains(out, `id="scener-filter-clear" class="btn btn-outline-secondary"`))
	be.True(t, strings.Contains(out, `2 sceners`))
}

func TestSearchPostTemplate(t *testing.T) {
	t.Parallel()

	wa := app.WebApp{}
	db := &sql.DB{}
	tmpl, err := template.New("searchpost.tmpl").Funcs(wa.FuncMap(t.Context(), db)).
		ParseFiles(filepath.Join("../", "../", "view", "app", "searchpost.tmpl"))
	be.Err(t, err, nil)

	var buf bytes.Buffer
	err = tmpl.ExecuteTemplate(&buf, "content", map[string]any{
		"title":      "Game or app titles search",
		"info":       "search artifact titles, descriptions, and comments",
		"searchMode": "desc",
		"editor":     true,
	})
	be.Err(t, err, nil)

	out := buf.String()
	be.True(t, strings.Contains(out, `href="/search/desc" rel="nofollow" aria-current="page"`))
	be.True(t, strings.Contains(out, `href="/editor/search/id"`))
	be.True(t, strings.Contains(out, `name="search-term-query"`))
	be.True(t, strings.Contains(out, `ultima 7, ultima 8`))
}

func TestSearchHtmxTemplate(t *testing.T) {
	t.Parallel()

	wa := app.WebApp{}
	db := &sql.DB{}
	tmpl, err := template.New("searchhtmx.tmpl").Funcs(wa.FuncMap(t.Context(), db)).
		ParseFiles(filepath.Join("../", "../", "view", "app", "searchhtmx.tmpl"))
	be.Err(t, err, nil)

	var buf bytes.Buffer
	err = tmpl.ExecuteTemplate(&buf, "content", map[string]any{
		"title":            "Lookup releasers",
		"info":             "find groups, names, magazines, bbs boards, ftp sites",
		"helpText":         "acronyms only match exact finds: 'rc' and 'rcn' are treated different",
		"hxPost":           "/search/releaser",
		"searchMode":       "releaser",
		"inputPlaceholder": "Type to search for a releaser…",
	})
	be.Err(t, err, nil)

	out := buf.String()
	be.True(t, strings.Contains(out, `href="/search/releaser" rel="nofollow" aria-current="page"`))
	be.True(t, strings.Contains(out, `id="search-htmx-clear" class="btn btn-outline-secondary" disabled`))
	be.True(t, strings.Contains(out, `hx-post="/search/releaser"`))
}

func TestAreacodesTemplate(t *testing.T) {
	t.Parallel()

	wa := app.WebApp{}
	db := &sql.DB{}
	tmpl, err := template.New("areacodes.tmpl").Funcs(wa.FuncMap(t.Context(), db)).
		ParseFiles(filepath.Join("../", "../", "view", "app", "areacodes.tmpl"))
	be.Err(t, err, nil)

	var buf bytes.Buffer
	err = tmpl.ExecuteTemplate(&buf, "content", map[string]any{
		"telephonecodes":  areacode.AreaCodes(),
		"territories":     areacode.Regions(),
		"abbreviations":   areacode.Abbreviations(),
		"areaCodeEntries": app.BuildAreaCodeEntries(),
		"regionEntries":   app.BuildAreaRegionEntries(),
		"abbrEntries":     app.BuildAreaAbbrEntries(),
	})
	be.Err(t, err, nil)

	out := buf.String()
	be.True(t, strings.Contains(out, `hx-trigger="input changed delay:150ms, search"`))
	be.True(t, strings.Contains(out, `id="areacodes-clear" disabled`))
	be.True(t, strings.Contains(out, `href="#sec-codes"`))
	be.True(t, strings.Contains(out, `>212</span>`))
	be.True(t, strings.Contains(out, `>NY</span>`))
	be.True(t, strings.Contains(out, `New York (NY)`))
}
