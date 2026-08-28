package model_test

import (
	"testing"

	"github.com/Defacto2/server/model"
	"github.com/nalgeon/be"
)

func TestReleaserNames(t *testing.T) {
	t.Parallel()

	var all model.ReleaserNames
	got := all.Distinct(nil, nil)
	be.Err(t, got)

	db := openDB(t)
	if db == nil {
		return
	}

	got = all.Distinct(t.Context(), db)
	be.Err(t, got, nil)

	total := len(all)
	be.True(t, total > 0)

	var grp model.ReleaserNames
	got = grp.DistinctGroups(t.Context(), db)
	be.Err(t, got, nil)
	be.True(t, len(grp) > 0)
	be.True(t, total > len(grp))

	var mag model.ReleaserNames
	got = mag.DistinctMagazines(t.Context(), db)
	be.Err(t, got, nil)
	be.True(t, len(mag) > 0)
	be.True(t, total > len(mag))

	var bbs model.ReleaserNames
	got = bbs.DistinctBBS(t.Context(), db)
	be.Err(t, got, nil)
	be.True(t, len(bbs) > 0)
	be.True(t, total > len(bbs))
}

func TestReleasersWhere(t *testing.T) {
	t.Parallel()

	_, got := model.ReleasersWhere(nil, nil, "")
	be.Err(t, got)

	db := openDB(t)
	if db == nil {
		return
	}

	fs, got := model.ReleasersWhere(t.Context(), db, "")
	be.Err(t, got, nil)
	be.Equal(t, len(fs), 0)

	fs, got = model.ReleasersWhere(t.Context(), db, "defacto2")
	be.Err(t, got, nil)
	be.True(t, len(fs) > 0)
}

func TestReleasersLimit(t *testing.T) {
	t.Parallel()

	var rels model.Releasers
	_, got := model.ReleasersWhere(nil, nil, "")
	be.Err(t, got)

	db := openDB(t)
	if db == nil {
		return
	}

	got = rels.Limit(t.Context(), db, 9999, 0, 0)
	be.Err(t, got)

	got = rels.Limit(t.Context(), db, model.Alphabetical, 1, 1)
	be.Err(t, got, nil)
	be.Equal(t, len(rels), 1)
	rec1 := rels[0].Unique.Name
	be.True(t, rec1 != "")

	got = rels.Limit(t.Context(), db, model.Oldest, 1, 2)
	be.Err(t, got, nil)
	be.Equal(t, len(rels), 1)
	rec2 := rels[0].Unique.Name
	be.True(t, rec2 != "")
	be.Equal(t, rec1, rec2)
}

func TestReleasersSimilar(t *testing.T) {
	t.Parallel()

	_, got := model.ReleasersWhere(nil, nil, "")
	be.Err(t, got)

	db := openDB(t)
	if db == nil {
		return
	}

	var sim model.Releasers
	got = sim.Similar(t.Context(), db, 0, "")
	be.Err(t, got, nil)
	be.True(t, len(sim) > 0)

	sim = model.Releasers{}
	got = sim.Similar(t.Context(), db, 0, "thereisnosuchgroup")
	be.Err(t, got, nil)
	be.True(t, len(sim) == 0)

	sim = model.Releasers{}
	got = sim.Similar(t.Context(), db, 0, "razor", "the", "at")
	be.Err(t, got, nil)
	be.True(t, len(sim) == model.PageSet)
	t.Log(len(sim))

	sim = model.Releasers{}
	got = sim.Similar(t.Context(), db, 1, "razor")
	be.Err(t, got, nil)
	be.True(t, len(sim) == 1)
	t.Log(len(sim), sim[0])
}
