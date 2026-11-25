package model_test

import (
	"testing"
	"uuid"

	"github.com/Defacto2/server/internal/logs"
	"github.com/Defacto2/server/internal/testutil"
	"github.com/Defacto2/server/model"
	"github.com/nalgeon/be"
)

func TestArtifacts(t *testing.T) {
	t.Parallel()

	art := model.Artifacts{}
	be.Err(t, art.Public(t.Context(), nil))

	db := testutil.DB(t)

	be.Err(t, art.Public(t.Context(), db), nil)
	be.True(t, art.Bytes > 0)
	be.True(t, art.Count > 0)
	be.True(t, art.MinYear >= model.EpochYear)
	be.True(t, art.MaxYear >= model.EpochYear)

	const limit = 3

	got, err := art.ByKey(t.Context(), db, 0, limit)
	be.Err(t, err, nil)
	be.Equal(t, len(got), limit)
	first, second, third := got[0].ID, got[1].ID, got[2].ID

	// confirm the keys are in descending order
	be.True(t, first > second)
	be.True(t, second > third)

	// confirm the page / offset logic is working correctly
	got, err = art.ByKey(t.Context(), db, 10, limit)
	be.Err(t, err, nil)
	be.Equal(t, len(got), limit)
	be.True(t, third > got[0].ID)

	// confirm newest to oldest record orderings
	got, err = art.ByOldest(t.Context(), db, 0, limit)
	be.Err(t, err, nil)
	be.Equal(t, len(got), limit)
	oldest := got[0].DateIssuedYear.Int16

	got, err = art.ByNewest(t.Context(), db, 0, limit)
	be.Err(t, err, nil)
	be.Equal(t, len(got), limit)
	newest := got[0].DateIssuedYear.Int16

	be.True(t, newest > oldest)
}

func TestOnly(t *testing.T) {
	t.Parallel()

	sl := logs.Discard()
	_, err := model.OnlyHidden(t.Context(), nil, 0, 1)
	be.Err(t, err)
	_, err = model.OnlyApproval(t.Context(), nil, 0, 1)
	be.Err(t, err)
	_, err = model.OnlyDescriptions(t.Context(), sl, nil, []string{"bbs"})
	be.Err(t, err)
	_, err = model.OnlyFilenames(t.Context(), nil, []string{".pdf"})
	be.Err(t, err)

	db := testutil.DB(t)
	const limit = 3

	got, err := model.OnlyHidden(t.Context(), db, 0, limit)
	be.Err(t, err, nil)
	be.Equal(t, len(got), limit)

	art := model.Artifacts{}
	got, err = art.OnlyUnwanted(t.Context(), db, 0, limit)
	be.Err(t, err, nil)
	be.Equal(t, len(got), limit)
	be.True(t, art.Bytes > 0)
	be.True(t, art.Count > 0)
	be.True(t, art.MinYear >= model.EpochYear)
	be.True(t, art.MaxYear >= model.EpochYear)

	got, err = model.OnlyApproval(t.Context(), db, 0, limit)
	be.Err(t, err, nil)
	be.True(t, len(got) <= limit)

	got, err = model.OnlyDescriptions(t.Context(), sl, db, []string{})
	be.Err(t, err, nil)
	be.Equal(t, len(got), 0)

	// note, common determiners and verbs ("the" etc) are ignored by postgres
	got, err = model.OnlyDescriptions(t.Context(), sl, db, []string{"bbs"})
	be.Err(t, err, nil)
	be.True(t, len(got) > 0)

	got, err = model.OnlyFilenames(t.Context(), db, []string{})
	be.Err(t, err, nil)
	be.Equal(t, len(got), 0)

	got, err = model.OnlyFilenames(t.Context(), db, []string{".pdf"})
	be.Err(t, err, nil)
	be.True(t, len(got) > 0)
}

func TestOnlyUniqueIDs(t *testing.T) {
	t.Parallel()

	_, err := model.OnlyUniqueIDs(t.Context(), nil, []int{1}, uuid.UUID{})
	be.Err(t, err)

	db := testutil.DB(t)

	got, err := model.OnlyUniqueIDs(t.Context(), db, nil, uuid.UUID{})
	be.Err(t, err, nil)
	be.Equal(t, len(got), 0)

	// Fetch 3 real records so the test is not coupled to hardcoded IDs or UUIDs
	art := model.Artifacts{}
	sample, err := art.ByKey(t.Context(), db, 0, 3)
	be.Err(t, err, nil)
	be.Equal(t, len(sample), 3)

	id0, id1, id2 := int(sample[0].ID), int(sample[1].ID), int(sample[2].ID)
	u0, err := uuid.Parse(sample[0].UUID.String)
	be.Err(t, err, nil)
	u1, err := uuid.Parse(sample[1].UUID.String)
	be.Err(t, err, nil)

	got, err = model.OnlyUniqueIDs(t.Context(), db, []int{id0}, uuid.UUID{})
	be.Err(t, err, nil)
	be.Equal(t, len(got), 1)

	got, err = model.OnlyUniqueIDs(t.Context(), db, []int{id0, id1, id2}, uuid.UUID{})
	be.Err(t, err, nil)
	be.Equal(t, len(got), 3)

	got, err = model.OnlyUniqueIDs(t.Context(), db, nil, u0)
	be.Err(t, err, nil)
	be.Equal(t, len(got), 1)

	// request the uuid and id of the same single record
	got, err = model.OnlyUniqueIDs(t.Context(), db, []int{id0}, u0)
	be.Err(t, err, nil)
	be.Equal(t, len(got), 1)

	// combine 2 IDs with the UUID of a 3rd record
	got, err = model.OnlyUniqueIDs(t.Context(), db, []int{id0, id2}, u1)
	be.Err(t, err, nil)
	be.Equal(t, len(got), 3)
}

func TestTexts(t *testing.T) {
	t.Parallel()

	_, err := model.OnlyTexts(t.Context(), nil)
	be.Err(t, err)

	db := testutil.DB(t)
	got, err := model.OnlyTexts(t.Context(), db)
	be.Err(t, err, nil)
	be.True(t, len(got) > 2)
}

func TestMagicErrs(t *testing.T) {
	t.Parallel()

	_, err := model.OnlyMagicErrs(t.Context(), nil, false)
	be.Err(t, err)

	db := testutil.DB(t)
	// depending on the database, this could return zero records if all magic numbers are updated
	_, err = model.OnlyMagicErrs(t.Context(), db, false)
	be.Err(t, err, nil)
}
