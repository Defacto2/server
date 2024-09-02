package fileslice_test

import (
	"testing"

	"github.com/Defacto2/server/handler/app/internal/fileslice"
	"github.com/stretchr/testify/assert"
)

func TestRecordsSubs(t *testing.T) {
	t.Parallel()
	s := fileslice.RecordsSub("")
	assert.Equal(t, "unknown uri", s)
	s = fileslice.RecordsSub("hack")
	assert.Equal(t, "game trainers or hacks", s)
}

func TestValid(t *testing.T) {
	t.Parallel()
	assert.False(t, fileslice.Valid("not-a-valid-uri"))
	assert.False(t, fileslice.Valid("/files/newest"))
	assert.True(t, fileslice.Valid("newest"))
	assert.True(t, fileslice.Valid("windows-pack"))
	assert.True(t, fileslice.Valid("advert"))
}

func TestMatch(t *testing.T) {
	t.Parallel()
	assert.Equal(t, fileslice.URI(-1), fileslice.Match("not-a-valid-uri"))
	assert.Equal(t, fileslice.URI(37), fileslice.Match("newest"))
	assert.Equal(t, fileslice.URI(60), fileslice.Match("windows-pack"))
	assert.Equal(t, fileslice.URI(1), fileslice.Match("advert"))
}

func TestRecordsSub(t *testing.T) {
	t.Parallel()
	s := fileslice.RecordsSub("")
	assert.Equal(t, "unknown uri", s)
	for i := range 57 {
		assert.NotEqual(t, "unknown uri", fileslice.URI(i).String())
	}
}
