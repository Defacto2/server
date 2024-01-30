package app_test

import (
	"testing"

	"github.com/Defacto2/server/handler/app"
	"github.com/stretchr/testify/assert"
)

func TestIsFiles(t *testing.T) {
	t.Parallel()
	assert.False(t, app.IsFiles("not-a-valid-uri"))
	assert.False(t, app.IsFiles("/files/newest"))
	assert.True(t, app.IsFiles("newest"))
	assert.True(t, app.IsFiles("windows-pack"))
	assert.True(t, app.IsFiles("advert"))
}

func TestMatch(t *testing.T) {
	t.Parallel()
	assert.Equal(t, app.URI(-1), app.Match("not-a-valid-uri"))
	assert.Equal(t, app.URI(35), app.Match("newest"))
	assert.Equal(t, app.URI(57), app.Match("windows-pack"))
	assert.Equal(t, app.URI(1), app.Match("advert"))
}
