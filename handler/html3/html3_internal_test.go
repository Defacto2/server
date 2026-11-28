package html3

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/Defacto2/server/internal/tags"
	"github.com/Defacto2/server/internal/testutil"
	"github.com/Defacto2/server/model"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/nalgeon/be"
)

var (
	errFetchBoom = errors.New("fetch failed")
	errCountBoom = errors.New("count failed")
	errSumBoom   = errors.New("sum failed")
)

func TestNavigateHelpers(t *testing.T) {
	t.Parallel()

	be.Equal(t, qs(""), "")
	be.Equal(t, qs("C=N&O=A"), "?C=N&O=A")

	be.Equal(t, previous(0), 1)
	be.Equal(t, previous(-3), 1)
	be.Equal(t, previous(1), 1)
	be.Equal(t, previous(4), 3)

	be.Equal(t, next(5, 5), 5)
	be.Equal(t, next(6, 5), 5)
	be.Equal(t, next(2, 5), 3)
}

func TestQueryHelpers(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	c := testutil.NewContext(t, "/html3/category/bbs")
	c.SetPath("/html3/category/bbs")
	dummyDB := &sql.DB{}

	okFetch := func(context.Context, boil.ContextExecutor, int, int, string) (models.FileSlice, error) {
		return models.FileSlice{{ID: 1}}, nil
	}
	okCount := func(context.Context, boil.ContextExecutor, tags.Tag) (int64, error) {
		return 42, nil
	}
	okSum := func(context.Context, boil.ContextExecutor, tags.Tag) (int64, error) {
		return 1024, nil
	}

	// Nil exec returns argsNil error.
	_, _, _, _, err := execute{label: "test", fetch: okFetch, count: okCount, sum: okSum}.query(ctx, nil, c, 1)
	be.Err(t, err)

	// Fetch error wraps with queryErr.
	_, _, _, _, err = execute{
		label: "by category",
		fetch: func(context.Context, boil.ContextExecutor, int, int, string) (models.FileSlice, error) {
			return nil, errFetchBoom
		},
		count: okCount,
		sum:   okSum,
	}.query(ctx, dummyDB, c, 1)
	be.Err(t, err, errFetchBoom)

	// Count error wraps with statErr.
	_, _, _, _, err = execute{
		label: "by category",
		fetch: okFetch,
		count: func(context.Context, boil.ContextExecutor, tags.Tag) (int64, error) {
			return 0, errCountBoom
		},
		sum: okSum,
	}.query(ctx, dummyDB, c, 1)
	be.Err(t, err, errCountBoom)

	// Sum error wraps with statErr.
	_, _, _, _, err = execute{
		label: "by category",
		fetch: okFetch,
		count: okCount,
		sum: func(context.Context, boil.ContextExecutor, tags.Tag) (int64, error) {
			return 0, errSumBoom
		},
	}.query(ctx, dummyDB, c, 1)
	be.Err(t, err, errSumBoom)

	// Happy path returns limit, total, sum, and records.
	limit, total, sum, recs, err := execute{
		label: "by category",
		fetch: okFetch,
		count: okCount,
		sum:   okSum,
	}.query(ctx, dummyDB, c, 1)
	be.Err(t, err, nil)
	be.Equal(t, limit, model.Maximum)
	be.Equal(t, total, 42)
	be.Equal(t, sum, int64(1024))
	be.Equal(t, len(recs), 1)
}

func TestTagByName(t *testing.T) {
	t.Parallel()

	be.Equal(t, emptyFS(nil), true)

	_, err := tagByName(nil, "bbs")
	be.Err(t, err, tags.ErrNoTags)

	_, err = tagByName(&tags.T{}, "bbs")
	be.Err(t, err, tags.ErrNoTags)

	tg := &tags.T{
		List: []tags.TagData{
			{URI: "bbs", Name: "BBS", Info: "bulletin board systems", Count: 10},
			{URI: "intro", Name: "Intro", Info: "a", Count: 1},
			{URI: "demo", Name: "Demo", Info: "über demos", Count: 5},
		},
	}

	empty, err := tagByName(tg, "nonexistent")
	be.Err(t, err, nil)
	be.Equal(t, empty, tags.TagData{})

	got, err := tagByName(tg, "BBS")
	be.Err(t, err, nil)
	be.Equal(t, got.Info, "Bulletin board systems")

	short, err := tagByName(tg, "Intro")
	be.Err(t, err, nil)
	be.Equal(t, short.Info, "a")

	utf8Tag, err := tagByName(tg, "Demo")
	be.Err(t, err, nil)
	be.Equal(t, utf8Tag.Info, "Über demos")
}
