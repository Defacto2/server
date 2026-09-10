package app_test

import (
	"database/sql"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/Defacto2/server/handler/app"
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
