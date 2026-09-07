package fulltext_test

import (
	"embed"
	"testing"

	"github.com/Defacto2/server/handler/fulltext"
	"github.com/Defacto2/server/internal/testutil"
	"github.com/nalgeon/be"
)

// checked in Sep 26, test coverage was poor at under 15%
// the coverage is low due to issues with the blaze indexing

func TestAdd(t *testing.T) {
	t.Parallel()

	ts := fulltext.Tidbits{}
	err := ts.Add("", "")
	be.Err(t, err)

	err = ts.Add("abc", "xyz")
	be.Err(t, err)
}

func TestNewIndex(t *testing.T) {
	t.Parallel()

	ts := fulltext.Tidbits{}
	var efs embed.FS
	err := ts.NewIndex(efs, "")
	be.Err(t, err)

	fsys := testutil.ProjectFS(t)
	err = ts.NewIndex(fsys, "/")
	be.Err(t, err)
}

func TestSearch(t *testing.T) {
	t.Parallel()

	ts := fulltext.Tidbits{}
	r := ts.Search("", 0)
	be.Equal(t, len(r), 0)
}
