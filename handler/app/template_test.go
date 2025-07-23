package app_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Defacto2/server/handler/app"
	"github.com/nalgeon/be"
)

func TestTemplTemplates(t *testing.T) {
	t.Parallel()
	tpl := app.Templ{}
	x, err := tpl.Templates(nil)
	be.Err(t, err)
	be.True(t, x == nil)
}

func TestFuncClosures(t *testing.T) {
	t.Parallel()
	tpl := app.Templ{}
	x := tpl.FuncClosures(nil)
	be.True(t, x != nil)
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
	x := app.Templ{}
	pages := x.Pages()

	p := filepath.Join("../", "../", "view", "app")
	view, err := filepath.Abs(p)
	be.Err(t, err, nil)

	for _, page := range *pages {
		be.True(t, page != "")
		ext := filepath.Ext(string(page))
		be.Equal(t, ".tmpl", ext)
		stat, err := os.Stat(filepath.Join(view, string(page)))
		be.Err(t, err, nil)
		be.True(t, stat.Size() > 0)
	}
}
