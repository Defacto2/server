package app_test

import (
	"testing"

	"github.com/Defacto2/server/handler/app"
	"github.com/stretchr/testify/assert"
)

func TestValid(t *testing.T) {
	t.Parallel()
	assert.False(t, app.Valid("not-a-valid-uri"))
	assert.False(t, app.Valid("/files/newest"))
	assert.True(t, app.Valid("newest"))
	assert.True(t, app.Valid("windows-pack"))
	assert.True(t, app.Valid("advert"))
}

func TestMatch(t *testing.T) {
	t.Parallel()
	assert.Equal(t, app.URI(-1), app.Match("not-a-valid-uri"))
	assert.Equal(t, app.URI(35), app.Match("newest"))
	assert.Equal(t, app.URI(57), app.Match("windows-pack"))
	assert.Equal(t, app.URI(1), app.Match("advert"))
}
