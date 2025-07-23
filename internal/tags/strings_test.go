package tags_test

import (
	"testing"

	"github.com/Defacto2/server/internal/tags"
	"github.com/nalgeon/be"
)

const expectedCount = 42

func TestTagStrings(t *testing.T) {
	uris := tags.URIs()
	names := tags.Names()
	determiner := tags.Determiner()
	infos := tags.Infos()

	be.True(t, len(uris) == expectedCount)
	be.True(t, len(names) == expectedCount)
	be.True(t, len(determiner) == expectedCount)
	be.True(t, len(infos) == expectedCount)

	for i := range expectedCount {
		if i == 0 {
			continue
		}
		x := tags.Tag(i)
		be.True(t, uris[x] != "")
		be.True(t, names[x] != "")
		be.True(t, determiner[x] != "")
		be.True(t, infos[x] != "")
	}
}
