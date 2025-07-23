package extensions_test

import (
	"slices"
	"testing"

	"github.com/Defacto2/server/internal/extensions"
	"github.com/nalgeon/be"
)

const (
	archives  = 10
	documents = 15
	images    = 10
	medias    = 7
)

func TestArchive(t *testing.T) {
	t.Parallel()
	a := extensions.Archive()
	be.True(t, len(a) == archives)
	for v := range slices.Values(a) {
		be.True(t, v != "")
	}
}

func TestDocument(t *testing.T) {
	t.Parallel()
	a := extensions.Document()
	be.True(t, len(a) == documents)
	for v := range slices.Values(a) {
		be.True(t, v != "")
	}
}

func TestImage(t *testing.T) {
	t.Parallel()
	a := extensions.Image()
	be.True(t, len(a) == images)
	for v := range slices.Values(a) {
		be.True(t, v != "")
	}
}

func TestMedia(t *testing.T) {
	t.Parallel()
	a := extensions.Media()
	be.True(t, len(a) == medias)
	for v := range slices.Values(a) {
		be.True(t, v != "")
	}
}
