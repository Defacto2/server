package model_test

import (
	"testing"

	"github.com/Defacto2/server/internal/tags"
	"github.com/Defacto2/server/model"
	"github.com/nalgeon/be"
)

func TestCount(t *testing.T) {
	t.Parallel()

	db := openDB(t)
	if db == nil {
		return
	}

	got, err := model.Count(nil, nil)
	be.Err(t, err)
	be.Equal(t, got, 0)

	got, err = model.Count(t.Context(), db)
	be.Err(t, err, nil)
	be.True(t, got > 2)

	gotx, goty, gotz, err := model.Counts(t.Context(), db)
	be.Err(t, err, nil)
	be.True(t, gotx > 2)
	be.True(t, goty > 2)
	be.True(t, gotz > 2)

	got, err = model.CountSection(t.Context(), db, -1)
	be.Err(t, err)
	be.Equal(t, got, 0)

	count, err := model.CountSection(t.Context(), db, tags.Nfo)
	be.Err(t, err, nil)
	be.True(t, count > 2)

	total, err := model.SumSection(t.Context(), db, tags.Nfo)
	be.Err(t, err, nil)
	be.True(t, total > 2)
	be.True(t, count != total) // technically this could return a false positive

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
	be.True(t, count == 0)

	count, err = model.CountTags(t.Context(), db, tags.Nfo, tags.Text)
	be.Err(t, err, nil)
	be.True(t, count > 2)

	count, err = model.SumReleaser(t.Context(), db, "")
	be.Err(t, err)
	be.Equal(t, count, 0)

	count, err = model.SumReleaser(t.Context(), db, "defacto")
	be.Err(t, err, nil)
	be.True(t, count > 2)
}

func TestUUIDs(t *testing.T) {
	t.Parallel()

	db := openDB(t)
	if db == nil {
		return
	}

	got, err := model.UUIDs(nil, nil)
	be.Err(t, err)
	be.True(t, got == model.UUIDVers{})

	got, err = model.UUIDs(t.Context(), db)
	be.Err(t, err, nil)
	be.True(t, got.Count > 2)
}
