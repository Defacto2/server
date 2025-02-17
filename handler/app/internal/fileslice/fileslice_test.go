package fileslice_test

import (
	"context"
	"fmt"
	"slices"
	"testing"

	"github.com/Defacto2/server/handler/app/internal/fileslice"
	"github.com/Defacto2/server/model"
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
	assert.Equal(t, fileslice.Newest, fileslice.Match("newest"))
	assert.Equal(t, fileslice.WindowsPack, fileslice.Match("windows-pack"))
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

func Slices() []fileslice.URI {
	return []fileslice.URI{
		fileslice.NewUploads,
		fileslice.NewUpdates,
		fileslice.ForApproval,
		fileslice.Deletions,
		fileslice.Unwanted,
		fileslice.Oldest,
		fileslice.Newest,
		fileslice.Sensenstahl,
	}
}

func TestFileInfo(t *testing.T) {
	t.Parallel()
	a, b, c := fileslice.FileInfo("")
	assert.Equal(t, "unknown uri", a)
	assert.Equal(t, "unknown uri", b)
	assert.Equal(t, "", c)
	for uri := range slices.Values(Slices()) {
		a, b, c = fileslice.FileInfo(uri.String())
		assert.NotEmpty(t, a)
		assert.NotEmpty(t, b)
		assert.NotEmpty(t, c)
	}
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

	uris := []fileslice.URI{}
	for i := range fileslice.WindowsPack {
		uris = append(uris, i)
	}
	for uri := range slices.Values(uris) {
		if uri.String() == "" {
			continue
		}
		_, err = fileslice.Records(context.TODO(), nil, uri.String(), 1, 1)
		msg := fmt.Sprintf("this uri caused an issue: %q %d", uri, uri)
		assert.Equal(t, model.ErrDB.Error(), err.Error(), msg)
	}
}

func TestCounter(t *testing.T) {
	t.Parallel()
	x, err := fileslice.Counter(nil)
	require.Error(t, err)
	assert.Empty(t, x)
}
