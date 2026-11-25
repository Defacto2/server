package model_test

import (
	"testing"

	"github.com/Defacto2/server/internal/tags"
	"github.com/Defacto2/server/internal/testutil"
	"github.com/Defacto2/server/model"
	"github.com/nalgeon/be"
)

func TestCount(t *testing.T) {
	t.Parallel()

	got, err := model.Count(t.Context(), nil)
	be.Err(t, err)
	be.Equal(t, got, 0)
	_, err = model.SectionStats(t.Context(), nil)
	be.Err(t, err)
	_, err = model.PlatformStats(t.Context(), nil)
	be.Err(t, err)

	db := testutil.DB(t)

	got, err = model.Count(t.Context(), db)
	be.Err(t, err, nil)
	be.True(t, got > 2)

	gotx, goty, gotz, err := model.Counts(t.Context(), db)
	be.Err(t, err, nil)
	be.True(t, gotx > 2)                  // total
	be.True(t, goty > 2 && goty <= gotx)  // public
	be.True(t, gotz >= 0 && gotz < goty)  // waiting approve

	got, err = model.CountSection(t.Context(), db, -1)
	be.Err(t, err)
	be.Equal(t, got, 0)

	count, err := model.CountSection(t.Context(), db, tags.Nfo)
	be.Err(t, err, nil)
	be.True(t, count > 2)

	total, err := model.SumSection(t.Context(), db, tags.Nfo)
	be.Err(t, err, nil)
	be.True(t, total > 2)
	be.True(t, count != total)

	secStats, err := model.SectionStats(t.Context(), db)
	be.Err(t, err, nil)
	be.Equal(t, secStats[tags.Nfo.String()].Count, count)
	be.Equal(t, secStats[tags.Nfo.String()].Bytes, total)

	platStats, err := model.PlatformStats(t.Context(), db)
	be.Err(t, err, nil)
	textCount, err := model.CountPlatform(t.Context(), db, tags.Text)
	be.Err(t, err, nil)
	be.Equal(t, platStats[tags.Text.String()].Count, textCount)
	textSum, err := model.SumPlatform(t.Context(), db, tags.Text)
	be.Err(t, err, nil)
	be.Equal(t, platStats[tags.Text.String()].Bytes, textSum)

	count, err = model.CountTags(t.Context(), db, -1, -1)
	be.Err(t, err)
	be.Equal(t, count, 0)

	count, err = model.CountTags(t.Context(), db, -1, tags.Nfo)
	be.Err(t, err)
	be.Equal(t, count, 0)

	count, err = model.CountTags(t.Context(), db, tags.Nfo, tags.Nfo)
	be.Err(t, err)
	be.Equal(t, count, 0)

	count, err = model.CountTags(t.Context(), db, tags.Text, tags.Nfo)
	be.Err(t, err, nil)
	be.Equal(t, count, int64(0))

	count, err = model.CountTags(t.Context(), db, tags.Nfo, tags.Text)
	be.Err(t, err, nil)
	be.True(t, count > 2)

	count, err = model.SumReleaser(t.Context(), db, "")
	be.Err(t, err)
	be.Equal(t, count, 0)

	count, err = model.SumReleaser(t.Context(), db, "defacto")
	be.Err(t, err, nil)
	be.True(t, count > 2)

	// Non-existent releaser should return 0 instead of a NULL scan error
	count, err = model.SumReleaser(t.Context(), db, "nonexistentgroupname12345")
	be.Err(t, err, nil)
	be.Equal(t, count, int64(0))
}

func TestUUIDs(t *testing.T) {
	t.Parallel()

	got, err := model.UUIDs(t.Context(), nil)
	be.Err(t, err)
	be.Equal(t, got, model.UUIDVers{})

	db := testutil.DB(t)

	got, err = model.UUIDs(t.Context(), db)
	be.Err(t, err, nil)
	be.True(t, got.Count > 2)
	sum := got.V1 + got.V2 + got.V3 + got.V4 + got.V5 + got.V6 + got.V7 + got.V8 + got.Error + got.Unknown
	be.Equal(t, got.Count, sum)
	pubCount, err := model.Count(t.Context(), db)
	be.Err(t, err, nil)
	be.Equal(t, int64(got.Count), pubCount)
}

func BenchmarkCounts(b *testing.B) {
	db := testutil.DB(b)
	ctx := b.Context()
	for b.Loop() {
		_, _, _, err := model.Counts(ctx, db)
		if err != nil {
			b.Fatal(err)
		}
	}
}
