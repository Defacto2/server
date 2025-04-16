package app_test

import (
	"path/filepath"
	"testing"

	"github.com/Defacto2/server/handler/app"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTemplTemplates(t *testing.T) {
	t.Parallel()
	tpl := app.Templ{}
	x, err := tpl.Templates(nil)
	require.Error(t, err)
	assert.Nil(t, x)
}

func TestFuncClosures(t *testing.T) {
	t.Parallel()
	tpl := app.Templ{}
	x := tpl.FuncClosures(nil)
	assert.NotNil(t, x)
}

func TestLinkRelrs(t *testing.T) {
	t.Parallel()
	x := app.LinkRelrs(false, nil, nil)
	assert.NotNil(t, x)
	x = app.LinkRelsPerf(nil, nil)
	assert.NotNil(t, x)
	x = app.LinkReleasers(false, false, nil, nil)
	assert.NotNil(t, x)
}

func TestTempls(t *testing.T) {
	t.Parallel()
	x := app.Templ{}
	pages := x.Pages()

	p := filepath.Join("../", "../", "view", "app")
	view, err := filepath.Abs(p)
	require.NoError(t, err)

	for _, page := range *pages {
		assert.NotNil(t, page)
		ext := filepath.Ext(string(page))
		assert.Equal(t, ".tmpl", ext)
		assert.FileExists(t, filepath.Join(view, string(page)))
	}
}
