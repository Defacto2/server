package model_test

import (
	"testing"

	"github.com/Defacto2/server/internal/testutil"
	"github.com/Defacto2/server/model"
	"github.com/nalgeon/be"
)

func assertSummaryPositive(t *testing.T, s model.Summary) {
	t.Helper()
	be.True(t, s.SumBytes.Int64 > 0)
	be.True(t, s.SumCount.Int64 > 0)
	be.True(t, s.MinYear.Int16 > 0)
	be.True(t, s.MaxYear.Int16 > 0)
}

func assertSummaryZero(t *testing.T, s model.Summary) {
	t.Helper()
	be.Equal(t, s.SumBytes.Int64, int64(0))
	be.Equal(t, s.SumCount.Int64, int64(0))
	be.Equal(t, s.MinYear.Int16, int16(0))
	be.Equal(t, s.MaxYear.Int16, int16(0))
}

func assertSummaryEqual(t *testing.T, got, want model.Summary) {
	t.Helper()
	be.Equal(t, got.SumBytes.Int64, want.SumBytes.Int64)
	be.Equal(t, got.SumCount.Int64, want.SumCount.Int64)
	be.Equal(t, got.MinYear.Int16, want.MinYear.Int16)
	be.Equal(t, got.MaxYear.Int16, want.MaxYear.Int16)
}

func TestByDescription(t *testing.T) {
	t.Parallel()

	s1 := model.Summary{}
	be.Err(t, s1.ByDescription(t.Context(), nil, []string{}))

	db := testutil.DB(t)
	be.Err(t, s1.ByDescription(t.Context(), db, []string{}))

	const term0 = "defacto"
	be.Err(t, s1.ByDescription(t.Context(), db, []string{term0}), nil)
	assertSummaryPositive(t, s1)

	const term1 = "razor 1911"
	s2 := model.Summary{}
	be.Err(t, s2.ByDescription(t.Context(), db, []string{term1, term0}), nil)
	assertSummaryPositive(t, s2)
	be.True(t, s2.SumBytes.Int64 > s1.SumBytes.Int64)
	be.True(t, s2.SumCount.Int64 > s1.SumCount.Int64)

	s3 := model.Summary{}
	be.Err(t, s3.ByDescription(t.Context(), db, []string{"", "  ", term0}), nil)
	be.Equal(t, s3.SumCount.Int64, s1.SumCount.Int64)
	be.Equal(t, s3.SumBytes.Int64, s1.SumBytes.Int64)
}

func TestByFilename(t *testing.T) {
	t.Parallel()

	s1 := model.Summary{}
	be.Err(t, s1.ByFilename(t.Context(), nil, []string{}))

	db := testutil.DB(t)
	be.Err(t, s1.ByFilename(t.Context(), db, []string{}))

	be.Err(t, s1.ByFilename(t.Context(), db, []string{".GIF"}), nil)
	assertSummaryPositive(t, s1)

	s2 := model.Summary{}
	be.Err(t, s2.ByFilename(t.Context(), db, []string{".gif"}), nil)
	assertSummaryEqual(t, s2, s1)
}

func TestByScener(t *testing.T) {
	t.Parallel()

	s1 := model.Summary{}
	be.Err(t, s1.ByScener(t.Context(), nil, "ace"))

	db := testutil.DB(t)
	be.Err(t, s1.ByScener(t.Context(), db, ""), nil)
	assertSummaryZero(t, s1)

	be.Err(t, s1.ByScener(t.Context(), db, "ace"), nil)
	assertSummaryPositive(t, s1)

	s2 := model.Summary{}
	be.Err(t, s2.ByScener(t.Context(), db, "ACE"), nil)
	assertSummaryEqual(t, s2, s1)
}

func TestByReleaser(t *testing.T) {
	t.Parallel()

	s1 := model.Summary{}
	be.Err(t, s1.ByReleaser(t.Context(), nil, "razor-1911"))

	db := testutil.DB(t)
	be.Err(t, s1.ByReleaser(t.Context(), db, ""))
	be.Err(t, s1.ByReleaser(t.Context(), db, "Razor 1911"))

	be.Err(t, s1.ByReleaser(t.Context(), db, "razor9111"), nil)
	assertSummaryZero(t, s1)

	be.Err(t, s1.ByReleaser(t.Context(), db, "razor-1911"), nil)
	assertSummaryPositive(t, s1)
}

func TestByMatch(t *testing.T) {
	t.Parallel()

	s1 := model.Summary{}
	be.Err(t, s1.ByMatch(t.Context(), nil, string(model.KeyScript)))

	db := testutil.DB(t)
	be.Err(t, s1.ByMatch(t.Context(), db, ""))
	be.Err(t, s1.ByMatch(t.Context(), db, "abcdefghi"))

	be.Err(t, s1.ByMatch(t.Context(), db, string(model.KeyScript)), nil)
	assertSummaryPositive(t, s1)
}

func TestByMatches(t *testing.T) {
	t.Parallel()

	db := testutil.DB(t)
	for _, key := range model.Keys() {
		uri := string(key)
		t.Run(uri, func(t *testing.T) {
			t.Parallel()
			s1 := model.Summary{}
			be.Err(t, s1.ByMatch(t.Context(), db, uri), nil)
			assertSummaryPositive(t, s1)
		})
	}
}
