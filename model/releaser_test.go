package model_test

import (
	"testing"

	"github.com/Defacto2/server/internal/testutil"
	"github.com/Defacto2/server/model"
	"github.com/nalgeon/be"
)

func TestReleaserNames(t *testing.T) {
	t.Parallel()

	var all model.ReleaserNames
	got := all.Distinct(nil, nil)
	be.Err(t, got)

	db := testutil.DB(t)
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

	db := testutil.DB(t)
	fs, got := model.ReleasersWhere(t.Context(), db, "")
	be.Err(t, got, nil)
	be.Equal(t, len(fs), 0)

	fs, got = model.ReleasersWhere(t.Context(), db, "defacto2")
	be.Err(t, got, nil)
	be.True(t, len(fs) > 0)
}

func TestReleasersLimit(t *testing.T) {
	t.Parallel()

	var obj model.Releasers
	_, got := model.ReleasersWhere(nil, nil, "")
	be.Err(t, got)

	db := testutil.DB(t)
	got = model.Alphabetical.Limit(t.Context(), db, nil, 0, 0)
	be.Err(t, got)

	got = model.Alphabetical.Limit(t.Context(), db, &obj, 1, 1)
	be.Err(t, got, nil)
	be.Equal(t, len(obj), 1)
	rec1 := obj[0].Unique.Name
	be.True(t, rec1 != "")

	got = model.Oldest.Limit(t.Context(), db, &obj, 1, 2)
	be.Err(t, got, nil)
	be.Equal(t, len(obj), 1)
	rec2 := obj[0].Unique.Name
	be.True(t, rec2 != "")
	be.Equal(t, rec1, rec2)
}

func TestReleasersSimilar(t *testing.T) {
	t.Parallel()

	_, got := model.ReleasersWhere(nil, nil, "")
	be.Err(t, got)

	db := testutil.DB(t)
	var obj model.Releasers
	got = obj.Similar(t.Context(), db, 0, "")
	be.Err(t, got, nil)
	be.True(t, len(obj) > 0)

	obj = model.Releasers{}
	got = obj.Similar(t.Context(), db, 0, "thereisnosuchgroup")
	be.Err(t, got, nil)
	be.True(t, len(obj) == 0)

	obj = model.Releasers{}
	got = obj.Similar(t.Context(), db, 0, "razor", "the", "at")
	be.Err(t, got, nil)
	be.True(t, len(obj) == model.PageSet)

	obj = model.Releasers{}
	got = obj.Similar(t.Context(), db, 1, "razor")
	be.Err(t, got, nil)
	be.True(t, len(obj) == 1)

	obj = model.Releasers{}
	got = obj.Similar(t.Context(), db, 1, "razor")
	be.Err(t, got, nil)
	be.True(t, len(obj) == 1)
}

func TestReleasersInit(t *testing.T) {
	t.Parallel()

	db := testutil.DB(t)
	var obj model.Releasers
	got := obj.Initialism(t.Context(), db, 1, "df")
	be.Err(t, got, nil)
	be.True(t, len(obj) > 0)
}

func TestReleasersMagazine(t *testing.T) {
	t.Parallel()

	db := testutil.DB(t)
	var obj model.Releasers
	got := obj.SimilarMagazine(t.Context(), db, 1, "defacto2")
	be.Err(t, got, nil)
	be.True(t, len(obj) > 0)
}

func TestMagazine(t *testing.T) {
	t.Parallel()

	db := testutil.DB(t)
	var obj model.Releasers
	got := obj.Magazine(t.Context(), db)
	be.Err(t, got, nil)
	be.True(t, len(obj) > 0)
	first := obj[0].Unique

	obj = model.Releasers{}
	got = obj.MagazineAZ(t.Context(), db)
	be.Err(t, got, nil)
	be.True(t, len(obj) > 0)
	firstAZ := obj[0].Unique

	be.True(t, first.Name != firstAZ.Name)
}

func TestSites(t *testing.T) {
	t.Parallel()

	db := testutil.DB(t)
	var obj model.Releasers
	got := obj.FTP(t.Context(), db)
	be.Err(t, got, nil)
	be.True(t, len(obj) > 0)
}
