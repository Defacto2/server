package extensions_test

import (
	"slices"
	"testing"

	"github.com/Defacto2/server/handler/app/internal/extensions"
	"github.com/stretchr/testify/assert"
)

const (
	archives  = 10
	documents = 15
	images    = 10
	medias    = 7
)

func TestArchive(t *testing.T) {
	a := extensions.Archive()
	assert.Len(t, a, archives)
	for v := range slices.Values(a) {
		assert.NotEmpty(t, v)
	}
}

func TestDocument(t *testing.T) {
	a := extensions.Document()
	assert.Len(t, a, documents)
	for v := range slices.Values(a) {
		assert.NotEmpty(t, v)
	}
}

func TestImage(t *testing.T) {
	a := extensions.Image()
	assert.Len(t, a, images)
	for v := range slices.Values(a) {
		assert.NotEmpty(t, v)
	}
}

func TestMedia(t *testing.T) {
	a := extensions.Media()
	assert.Len(t, a, medias)
	for v := range slices.Values(a) {
		assert.NotEmpty(t, v)
	}
}
