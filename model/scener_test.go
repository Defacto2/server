package model_test

import (
	"testing"

	"github.com/Defacto2/server/internal/testutil"
	"github.com/Defacto2/server/model"
	"github.com/nalgeon/be"
)

func TestScener(t *testing.T) {
	t.Parallel()

	var s model.Scener
	fs, got := s.Where(nil, nil, "")
	be.Err(t, got)
	be.True(t, len(fs) == 0)

	db := testutil.DB(t)
	fs, got = s.Where(t.Context(), db, "")
	be.Err(t, got, nil)
	be.True(t, len(fs) == 0)

	fs1, got := s.Where(t.Context(), db, "ace")
	be.Err(t, got, nil)
	be.True(t, len(fs1) > 0)

	fs2, got := s.Where(t.Context(), db, "ACE ")
	be.Err(t, got, nil)
	be.True(t, len(fs2) > 0)

	be.Equal(t, len(fs1), len(fs2))
}

func TestSceners(t *testing.T) {
	t.Parallel()

	db := testutil.DB(t)
	var d1 model.Sceners
	got := d1.Distinct(t.Context(), db)
	be.Err(t, got, nil)
	be.True(t, len(d1) > 0)

	var w1 model.Sceners
	got = w1.Writer(t.Context(), db)
	be.Err(t, got, nil)
	be.True(t, len(w1) > 0)

	var a1 model.Sceners
	got = a1.Artist(t.Context(), db)
	be.Err(t, got, nil)
	be.True(t, len(a1) > 0)

	var c1 model.Sceners
	got = c1.Coder(t.Context(), db)
	be.Err(t, got, nil)
	be.True(t, len(c1) > 0)

	var m1 model.Sceners
	got = m1.Musician(t.Context(), db)
	be.Err(t, got, nil)
	be.True(t, len(m1) > 0)

	if len(d1) == 0 {
		t.Fatal("d2 == 0")
	}

	preSort := d1[0]
	sorted := d1.Sort()
	be.True(t, len(sorted) > 0)
	first := sorted[0]
	be.True(t, string(preSort.Name) != first)
}
