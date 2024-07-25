package mfs_test

import (
	"testing"

	"github.com/Defacto2/server/handler/app/internal/mfs"
	"github.com/stretchr/testify/assert"
)

func TestRecordsSubs(t *testing.T) {
	t.Parallel()
	s := mfs.RecordsSub("")
	assert.Equal(t, "unknown uri", s)
	s = mfs.RecordsSub("hack")
	assert.Equal(t, "game trainers or hacks", s)
}

func TestValid(t *testing.T) {
	t.Parallel()
	assert.False(t, mfs.Valid("not-a-valid-uri"))
	assert.False(t, mfs.Valid("/files/newest"))
	assert.True(t, mfs.Valid("newest"))
	assert.True(t, mfs.Valid("windows-pack"))
	assert.True(t, mfs.Valid("advert"))
}

func TestMatch(t *testing.T) {
	t.Parallel()
	assert.Equal(t, mfs.URI(-1), mfs.Match("not-a-valid-uri"))
	assert.Equal(t, mfs.URI(37), mfs.Match("newest"))
	assert.Equal(t, mfs.URI(60), mfs.Match("windows-pack"))
	assert.Equal(t, mfs.URI(1), mfs.Match("advert"))
}

func TestRecordsSub(t *testing.T) {
	t.Parallel()
	s := mfs.RecordsSub("")
	assert.Equal(t, "unknown uri", s)
	for i := range 57 {
		assert.NotEqual(t, "unknown uri", mfs.URI(i).String())
	}
}
