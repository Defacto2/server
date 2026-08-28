package model_test

import (
	"testing"

	"github.com/Defacto2/server/model"
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

	got, err = art.ByHidden(t.Context(), db, 0, limit)
	be.Err(t, err, nil)
	be.True(t, len(got) > 0)
}
