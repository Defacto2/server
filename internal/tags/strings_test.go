package tags_test

import (
	"fmt"
	"testing"

	"github.com/Defacto2/server/internal/tags"
	"github.com/stretchr/testify/assert"
)

const expectedCount = 42

func TestTagStrings(t *testing.T) {
	uris := tags.URIs()
	names := tags.Names()
	determiner := tags.Determiner()
	infos := tags.Infos()

	assert.Len(t, uris, expectedCount)
	assert.Len(t, names, expectedCount)
	assert.Len(t, determiner, expectedCount)
	assert.Len(t, infos, expectedCount)

	for i := range expectedCount {
		if i == 0 {
			continue
		}
		x := tags.Tag(i)
		msg := fmt.Sprintf("tag %d, '%s' should not be empty", x, x)
		assert.NotEmpty(t, uris[x], msg)
		assert.NotEmpty(t, names[x], msg)
		assert.NotEmpty(t, determiner[x], msg)
		assert.NotEmpty(t, infos[x], msg)
	}
}
