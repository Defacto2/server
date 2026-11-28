package janeway

import (
	"testing"

	"github.com/Defacto2/server/handler/releaser/name"
)

func TestGroups(t *testing.T) {
	t.Parallel()

	seen := make(map[GroupID]URI, len(groups))
	for uri, id := range groups {
		if !name.Valid(name.Path(uri)) {
			t.Errorf("%q: invalid releaser URI", uri)
		}
		if id <= 0 {
			t.Errorf("%q: GroupID %d must be positive", uri, id)
		}
		if prev, dup := seen[id]; dup {
			t.Errorf("%q: duplicate GroupID %d (already used by %q)", uri, id, prev)
		}
		seen[id] = uri
	}
}
