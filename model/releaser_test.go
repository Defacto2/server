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
	be.Err(t, all.Distinct(t.Context(), nil))

	db := testutil.DB(t)
	be.Err(t, all.Distinct(t.Context(), db), nil)

	total := len(all)
	be.True(t, total > 0)

	var grp model.ReleaserNames
	be.Err(t, grp.DistinctGroups(t.Context(), db), nil)
	be.True(t, len(grp) > 0)
	be.True(t, total > len(grp))

	var mag model.ReleaserNames
	be.Err(t, mag.DistinctMagazines(t.Context(), db), nil)
	be.True(t, len(mag) > 0)
	be.True(t, total > len(mag))

	var bbs model.ReleaserNames
	be.Err(t, bbs.DistinctBBS(t.Context(), db), nil)
	be.True(t, len(bbs) > 0)
	be.True(t, total > len(bbs))
}

func TestReleasersWhere(t *testing.T) {
	t.Parallel()

	_, err := model.ReleasersWhere(t.Context(), nil, "")
	be.Err(t, err)

	db := testutil.DB(t)
	fs, err := model.ReleasersWhere(t.Context(), db, "")
	be.Err(t, err, nil)
	be.Equal(t, len(fs), 0)

	fs, err = model.ReleasersWhere(t.Context(), db, "defacto2")
	be.Err(t, err, nil)
	be.True(t, len(fs) > 0)
}

func TestReleasersLimit(t *testing.T) {
	t.Parallel()

	var obj model.Releasers
	be.Err(t, model.Alphabetical.Limit(t.Context(), nil, &obj, 1, 1))

	db := testutil.DB(t)
	be.Err(t, model.Alphabetical.Limit(t.Context(), db, nil, 0, 0))

	be.Err(t, model.Alphabetical.Limit(t.Context(), db, &obj, 1, 1), nil)
	be.Equal(t, len(obj), 1)
	rec1 := obj[0].Unique.Name
	be.True(t, rec1 != "")

	be.Err(t, model.Oldest.Limit(t.Context(), db, &obj, 1, 2), nil)
	be.Equal(t, len(obj), 1)
	rec2 := obj[0].Unique.Name
	be.True(t, rec2 != "")
	be.Equal(t, rec1, rec2)
}

func TestReleasersSimilar(t *testing.T) {
	t.Parallel()

	var obj model.Releasers
	be.Err(t, obj.Similar(t.Context(), nil, 0, ""))

	db := testutil.DB(t)
	be.Err(t, obj.Similar(t.Context(), db, 0, ""), nil)
	be.True(t, len(obj) > 0)

	obj = model.Releasers{}
	be.Err(t, obj.Similar(t.Context(), db, 0, "thereisnosuchgroup"), nil)
	be.Equal(t, len(obj), 0)

	obj = model.Releasers{}
	be.Err(t, obj.Similar(t.Context(), db, 0, "razor", "the", "at"), nil)
	be.Equal(t, len(obj), model.PageSet)

	obj = model.Releasers{}
	be.Err(t, obj.Similar(t.Context(), db, 1, "razor"), nil)
	be.Equal(t, len(obj), 1)

	// Regex metacharacters and apostrophes must not error
	obj = model.Releasers{}
	be.Err(t, obj.Similar(t.Context(), db, 5, "test[", "group (", "o'brien"), nil)
}

func TestReleasersInit(t *testing.T) {
	t.Parallel()

	var obj model.Releasers
	be.Err(t, obj.Initialism(t.Context(), nil, 1, "df"))

	db := testutil.DB(t)
	be.Err(t, obj.Initialism(t.Context(), db, 1, "df"), nil)
	be.True(t, len(obj) > 0)
}

func TestReleasersMagazine(t *testing.T) {
	t.Parallel()

	var obj model.Releasers
	be.Err(t, obj.SimilarMagazine(t.Context(), nil, 1, "defacto2"))

	db := testutil.DB(t)
	be.Err(t, obj.SimilarMagazine(t.Context(), db, 1, "defacto2"), nil)
	be.True(t, len(obj) > 0)
}

func TestMagazine(t *testing.T) {
	t.Parallel()

	var obj model.Releasers
	be.Err(t, obj.Magazine(t.Context(), nil))
	be.Err(t, obj.MagazineAZ(t.Context(), nil))

	db := testutil.DB(t)
	be.Err(t, obj.Magazine(t.Context(), db), nil)
	be.True(t, len(obj) > 0)
	first := obj[0].Unique

	obj = model.Releasers{}
	be.Err(t, obj.MagazineAZ(t.Context(), db), nil)
	be.True(t, len(obj) > 0)
	firstAZ := obj[0].Unique

	be.True(t, first.Name != firstAZ.Name)
}

func TestSites(t *testing.T) {
	t.Parallel()

	var obj model.Releasers
	be.Err(t, obj.FTP(t.Context(), nil))

	db := testutil.DB(t)
	be.Err(t, obj.FTP(t.Context(), db), nil)
	be.True(t, len(obj) > 0)
}
