package fulltext_test

import (
	"embed"
	"testing"

	"github.com/Defacto2/server/handler/fulltext"
	"github.com/nalgeon/be"
)

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
	var fsys embed.FS
	err := ts.NewIndex(fsys, "")
	be.Err(t, err)
}

func TestSearch(t *testing.T) {
	t.Parallel()
	ts := fulltext.Tidbits{}
	r := ts.Search("", 0)
	be.Equal(t, len(r), 0)
}
