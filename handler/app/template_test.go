package app_test

import (
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

func TestLinkRels(t *testing.T) {
	t.Parallel()
	x := app.LinkRels(false, nil, nil)
	assert.NotNil(t, x)
	x = app.LinkRelsPerformant(nil, nil)
	assert.NotNil(t, x)
	x = app.LinkReleasers(false, false, nil, nil)
	assert.NotNil(t, x)
}
