package tidbit

import (
	"testing"

	"github.com/nalgeon/be"
)

func TestMatch(t *testing.T) {
	t.Parallel()

	for _, uris := range groups {
		for _, uri := range uris {
			key := string(uri)
			got := Missing(key)
			be.True(t, !got)

			finds := Find(key)
			be.True(t, len(finds) > 0)
		}
	}
}
