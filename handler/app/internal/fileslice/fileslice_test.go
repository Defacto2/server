package fileslice_test

import (
	"context"
	"testing"

	"github.com/Defacto2/server/handler/app/internal/fileslice"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	assert.Equal(t, fileslice.URI(61), fileslice.Match("windows-pack"))
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

func TestFileInfo(t *testing.T) {
	t.Parallel()
	a, b, c := fileslice.FileInfo("")
	assert.Equal(t, "unknown uri", a)
	assert.Equal(t, "unknown uri", b)
	assert.Equal(t, "", c)

	a, b, c = fileslice.FileInfo("newest")
	assert.Equal(t, "newest releases", a)
	assert.Equal(t, "the newest releases", b)
	assert.NotEmpty(t, c)
}

func TestRecords(t *testing.T) {
	t.Parallel()
	x, err := fileslice.Records(context.TODO(), nil, "", 0, 0)
	require.Error(t, err)
	assert.Nil(t, x)

	proof := fileslice.URI(45).String()
	x, err = fileslice.Records(context.TODO(), nil, proof, 1, 1)
	require.Error(t, err)
	assert.Nil(t, x)
}

func TestCounter(t *testing.T) {
	t.Parallel()
	x, err := fileslice.Counter(nil)
	require.Error(t, err)
	assert.Empty(t, x)
}
