package model_test

import (
	"testing"

	"github.com/Defacto2/server/internal/testutil"
	"github.com/Defacto2/server/model"
	"github.com/nalgeon/be"
)

func TestByDescription(t *testing.T) {
	t.Parallel()

	s1 := model.Summary{}
	got := s1.ByDescription(nil, nil, []string{})
	be.Err(t, got)

	got = s1.ByDescription(t.Context(), nil, []string{})
	be.Err(t, got)

	db := testutil.DB(t)
	got = s1.ByDescription(t.Context(), db, []string{})
	be.Err(t, got)

	const term0 = "defacto"
	got = s1.ByDescription(t.Context(), db, []string{term0})
	be.Err(t, got, nil)
	be.True(t, s1.SumBytes.Int64 > 0)
	be.True(t, s1.SumCount.Int64 > 0)
	be.True(t, s1.MinYear.Int16 > 0)
	be.True(t, s1.MaxYear.Int16 > 0)

	const term1 = "razor 1911"
	s2 := model.Summary{}
	got = s2.ByDescription(t.Context(), db, []string{term1, term0})
	be.Err(t, got, nil)
	be.True(t, s2.SumBytes.Int64 > 0)
	be.True(t, s2.SumCount.Int64 > 0)
	be.True(t, s2.MinYear.Int16 > 0)
	be.True(t, s2.MaxYear.Int16 > 0)

	be.True(t, s2.SumBytes.Int64 > s1.SumBytes.Int64)
	be.True(t, s2.SumCount.Int64 > s1.SumCount.Int64)
}

func TestByFilename(t *testing.T) {
	t.Parallel()

	s1 := model.Summary{}
	got := s1.ByFilename(nil, nil, []string{})
	be.Err(t, got)

	got = s1.ByFilename(t.Context(), nil, []string{})
	be.Err(t, got)

	db := testutil.DB(t)
	got = s1.ByFilename(t.Context(), db, []string{})
	be.Err(t, got)

	got = s1.ByFilename(t.Context(), db, []string{".GIF"})
	be.Err(t, got, nil)
	be.True(t, s1.SumBytes.Int64 > 0)
	be.True(t, s1.SumCount.Int64 > 0)
	be.True(t, s1.MinYear.Int16 > 0)
	be.True(t, s1.MaxYear.Int16 > 0)

	s2 := model.Summary{}
	got = s2.ByFilename(t.Context(), db, []string{".gif"})
	be.Err(t, got, nil)
	t.Log(s1, s2)
	be.True(t, s2.SumBytes.Int64 == s1.SumBytes.Int64)
	be.True(t, s2.SumCount.Int64 == s1.SumCount.Int64)
	be.True(t, s2.MinYear.Int16 == s1.MinYear.Int16)
	be.True(t, s2.MaxYear.Int16 == s1.MaxYear.Int16)
}

func TestByScener(t *testing.T) {
	t.Parallel()

	db := testutil.DB(t)
	s1 := model.Summary{}
	got := s1.ByScener(t.Context(), db, "")
	be.Err(t, got, nil)
	be.True(t, s1.SumBytes.Int64 == 0)
	be.True(t, s1.SumCount.Int64 == 0)
	be.True(t, s1.MinYear.Int16 == 0)
	be.True(t, s1.MaxYear.Int16 == 0)

	got = s1.ByScener(t.Context(), db, "ace")
	be.Err(t, got, nil)
	be.True(t, s1.SumBytes.Int64 > 0)
	be.True(t, s1.SumCount.Int64 > 0)
	be.True(t, s1.MinYear.Int16 > 0)
	be.True(t, s1.MaxYear.Int16 > 0)

	s2 := model.Summary{}
	got = s2.ByScener(t.Context(), db, "ACE")
	be.Err(t, got, nil)
	t.Log(s1, s2)
	be.True(t, s2.SumBytes.Int64 == s1.SumBytes.Int64)
	be.True(t, s2.SumCount.Int64 == s1.SumCount.Int64)
	be.True(t, s2.MinYear.Int16 == s1.MinYear.Int16)
	be.True(t, s2.MaxYear.Int16 == s1.MaxYear.Int16)
}

func TestByReleaser(t *testing.T) {
	t.Parallel()

	db := testutil.DB(t)
	s1 := model.Summary{}
	got := s1.ByReleaser(t.Context(), db, "")
	be.Err(t, got)

	got = s1.ByReleaser(t.Context(), db, "Razor 1911")
	be.Err(t, got)

	got = s1.ByReleaser(t.Context(), db, "razor9111")
	be.Err(t, got, nil)
	t.Log(s1)
	be.True(t, s1.SumBytes.Int64 == 0)
	be.True(t, s1.SumCount.Int64 == 0)
	be.True(t, s1.MinYear.Int16 == 0)
	be.True(t, s1.MaxYear.Int16 == 0)

	got = s1.ByReleaser(t.Context(), db, "razor-1911")
	be.Err(t, got, nil)
	t.Log(s1)
	be.True(t, s1.SumBytes.Int64 > 0)
	be.True(t, s1.SumCount.Int64 > 0)
	be.True(t, s1.MinYear.Int16 > 0)
	be.True(t, s1.MaxYear.Int16 > 0)
}

func TestByMatch(t *testing.T) {
	t.Parallel()

	db := testutil.DB(t)
	s1 := model.Summary{}
	got := s1.ByMatch(t.Context(), db, "")
	be.Err(t, got)

	const invalid = "abcdefghi"
	got = s1.ByMatch(t.Context(), db, invalid)
	be.Err(t, got)

	const valid = model.KeyScript
	got = s1.ByMatch(t.Context(), db, string(valid))
	be.Err(t, got, nil)
	be.True(t, s1.SumBytes.Int64 > 0)
	be.True(t, s1.SumCount.Int64 > 0)
	be.True(t, s1.MinYear.Int16 > 0)
	be.True(t, s1.MaxYear.Int16 > 0)
}

func TestByMatches(t *testing.T) {
	t.Parallel()

	db := testutil.DB(t)
	for n, key := range model.Keys() {
		uri := string(key)
		t.Run("summary of "+uri, func(t *testing.T) {
			t.Parallel()
			t.Log(n, key, string(key))

			s1 := model.Summary{}
			got := s1.ByMatch(t.Context(), db, uri)

			be.Err(t, got, nil)
			be.True(t, s1.SumBytes.Int64 > 0)
			be.True(t, s1.SumCount.Int64 > 0)
			be.True(t, s1.MinYear.Int16 > 0)
			be.True(t, s1.MaxYear.Int16 > 0)
		})
	}
}
