package model_test

import (
	"log/slog"
	"testing"

	"github.com/Defacto2/server/model"
	"github.com/google/uuid"
	"github.com/nalgeon/be"
)

func TestArtifacts(t *testing.T) {
	t.Parallel()

	db := openDB(t)
	if db == nil {
		return
	}

	art := model.Artifacts{}
	err := art.Public(nil, nil)
	be.Err(t, err)

	err = art.Public(t.Context(), db)
	be.Err(t, err, nil)
	be.True(t, art.Bytes > 0)
	be.True(t, art.Count > 0)
	be.True(t, art.MinYear >= model.EpochYear)
	be.True(t, art.MaxYear >= model.EpochYear)

	// the remaining queries should return identical stats,
	// but use different ordering of results

	const limit = 3

	got, err := art.ByKey(t.Context(), db, 0, limit)
	be.Err(t, err, nil)
	be.True(t, len(got) == limit)
	first := got[0].ID
	second := got[1].ID
	third := got[2].ID

	// confirm the keys are in decending order
	be.True(t, first > second)
	be.True(t, second > third)

	// confirm the page / offset logic is working correctly
	got, err = art.ByKey(t.Context(), db, 10, limit)
	be.Err(t, err, nil)
	be.True(t, len(got) == limit)
	tenth := got[0].ID
	be.True(t, third > tenth)

	// confirm newest to oldest record orderings
	got, err = art.ByOldest(t.Context(), db, 0, limit)
	be.Err(t, err, nil)
	oldest := got[0].DateIssuedYear.Int16

	got, err = art.ByNewest(t.Context(), db, 0, limit)
	be.Err(t, err, nil)
	newest := got[0].DateIssuedYear.Int16

	be.True(t, newest > oldest)
}

func TestOnlys(t *testing.T) {
	t.Parallel()

	db := openDB(t)
	if db == nil {
		return
	}

	const limit = 3

	got, err := model.OnlyHidden(t.Context(), db, 0, limit)
	be.Err(t, err, nil)
	be.True(t, len(got) == limit)

	art := model.Artifacts{}
	got, err = art.OnlyUnwanted(t.Context(), db, 0, limit)
	be.Err(t, err, nil)
	be.True(t, len(got) == limit)
	be.True(t, art.Bytes > 0)
	be.True(t, art.Count > 0)
	be.True(t, art.MinYear >= model.EpochYear)
	be.True(t, art.MaxYear >= model.EpochYear)

	got, err = model.OnlyApproval(t.Context(), db, 0, limit)
	be.Err(t, err, nil)
	be.True(t, len(got) == limit)

	sl := slog.Default()
	got, err = model.OnlyDescriptions(t.Context(), sl, db, []string{})
	be.Err(t, err, nil)
	be.True(t, len(got) == 0)

	// note, common determiners and verbs ("the" etc) are ignored by postgres
	got, err = model.OnlyDescriptions(t.Context(), sl, db, []string{"bbs"})
	be.Err(t, err, nil)
	be.True(t, len(got) > 0)

	got, err = model.OnlyFilenames(t.Context(), db, []string{})
	be.Err(t, err, nil)
	be.True(t, len(got) == 0)

	got, err = model.OnlyFilenames(t.Context(), db, []string{".pdf"})
	be.Err(t, err, nil)
	be.True(t, len(got) > 0)
}

func TestOnlyUniqueIDs(t *testing.T) {
	t.Parallel()

	db := openDB(t)
	if db == nil {
		return
	}

	got, err := model.OnlyUniqueIDs(t.Context(), db, nil, uuid.UUID{})
	be.Err(t, err, nil)
	be.True(t, len(got) == 0)

	got, err = model.OnlyUniqueIDs(t.Context(), db, []int{1}, uuid.UUID{})
	be.Err(t, err, nil)
	be.True(t, len(got) == 1)

	got, err = model.OnlyUniqueIDs(t.Context(), db, []int{1, 2, 3, 4, 5}, uuid.UUID{})
	be.Err(t, err, nil)
	be.True(t, len(got) == 5)

	const r1 = "c8cd0b4c-2f54-11e0-8827-cc1607e15609"
	u1, err := uuid.Parse(r1)
	be.Err(t, err, nil)

	got, err = model.OnlyUniqueIDs(t.Context(), db, []int{}, u1)
	be.Err(t, err, nil)
	be.True(t, len(got) == 1)

	// request the uuid and id of the same single record
	got, err = model.OnlyUniqueIDs(t.Context(), db, []int{1}, u1)
	be.Err(t, err, nil)
	be.True(t, len(got) == 1)

	const r2 = "c8cd0f9e-2f54-11e0-8827-cc1607e15609"
	u2, err := uuid.Parse(r2)
	be.Err(t, err, nil)
	// get records with keys 1, 2, 3 using a combo of id types
	got, err = model.OnlyUniqueIDs(t.Context(), db, []int{1, 3}, u2)
	be.Err(t, err, nil)
	be.True(t, len(got) == 3)
}

func TestOnlyTexts(t *testing.T) {
	t.Parallel()

	db := openDB(t)
	if db == nil {
		return
	}

	got, err := model.OnlyTexts(t.Context(), db)
	be.Err(t, err, nil)
	be.True(t, len(got) > 2)
}

func TestOnlyMagicErrs(t *testing.T) {
	t.Parallel()

	db := openDB(t)
	if db == nil {
		return
	}

	// depending on the database, this could return zero records if all magic numbers are updated
	_, err := model.OnlyMagicErrs(t.Context(), db, false)
	be.Err(t, err, nil)
}
