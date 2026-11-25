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
	fs, err := s.Where(t.Context(), nil, "")
	be.Err(t, err)
	be.Equal(t, len(fs), 0)

	db := testutil.DB(t)
	fs, err = s.Where(t.Context(), db, "")
	be.Err(t, err, nil)
	be.Equal(t, len(fs), 0)

	fs1, err := s.Where(t.Context(), db, "ace")
	be.Err(t, err, nil)
	be.True(t, len(fs1) > 0)

	fs2, err := s.Where(t.Context(), db, "ACE ")
	be.Err(t, err, nil)
	be.True(t, len(fs2) > 0)

	be.Equal(t, len(fs1), len(fs2))
}

func TestSceners(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	var empty model.Sceners
	be.Err(t, empty.Distinct(ctx, nil))
	be.Err(t, empty.Writer(ctx, nil))
	be.Err(t, empty.Artist(ctx, nil))
	be.Err(t, empty.Coder(ctx, nil))
	be.Err(t, empty.Musician(ctx, nil))

	db := testutil.DB(t)
	var d1 model.Sceners
	be.Err(t, d1.Distinct(ctx, db), nil)
	be.True(t, len(d1) > 0)

	var w1 model.Sceners
	be.Err(t, w1.Writer(ctx, db), nil)
	be.True(t, len(w1) > 0)

	var a1 model.Sceners
	be.Err(t, a1.Artist(ctx, db), nil)
	be.True(t, len(a1) > 0)

	var c1 model.Sceners
	be.Err(t, c1.Coder(ctx, db), nil)
	be.True(t, len(c1) > 0)

	var m1 model.Sceners
	be.Err(t, m1.Musician(ctx, db), nil)
	be.True(t, len(m1) > 0)

	sorted := d1.Sort()
	be.True(t, len(sorted) > 0)
}

func TestScenersSort(t *testing.T) {
	t.Parallel()

	s := model.Sceners{
		{Name: "zeta, Beta"},
		{Name: "ace"},
		{Name: "ACE"},
		{Name: "007"},
	}
	got := s.Sort()
	want := []string{"007", "ACE", "Beta", "zeta"}
	be.Equal(t, got, want)
}
