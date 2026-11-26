package extensions_test

import (
	"strings"
	"testing"

	"github.com/Defacto2/server/internal/extensions"
	"github.com/nalgeon/be"
)

func TestExtensions(t *testing.T) {
	t.Parallel()

	lists := []struct {
		name string
		exts []string
	}{
		{"archive", extensions.Archive()},
		{"document", extensions.Document()},
		{"image", extensions.Image()},
		{"media", extensions.Media()},
	}

	seen := make(map[string]string)
	for _, list := range lists {
		t.Run(list.name, func(t *testing.T) {
			t.Parallel()
			be.True(t, len(list.exts) > 0)
			for _, ext := range list.exts {
				be.True(t, len(ext) > 1 && strings.HasPrefix(ext, "."))
				be.Equal(t, ext, strings.ToLower(ext))
			}
		})

		for _, ext := range list.exts {
			if prev, dup := seen[ext]; dup {
				t.Errorf("duplicate extension %q in %s and %s", ext, prev, list.name)
			}
			seen[ext] = list.name
		}
	}
}
