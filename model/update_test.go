package model_test

import (
	"testing"

	"github.com/Defacto2/server/model"
	"github.com/nalgeon/be"
)

func TestUpdateBoolFrom(t *testing.T) {
	t.Parallel()

	got := model.UpdateBoolFrom(t.Context(), nil, nil, 0, 0, false)
	be.Err(t, got)

	db := openDB(t)
	if db == nil {
		return
	}

	tx0, err := db.BeginTx(t.Context(), nil)
	be.Err(t, err, nil)

	got = model.UpdateBoolFrom(t.Context(), db, tx0, 0, 0, false)
	be.Err(t, got)

	got = model.UpdateBoolFrom(t.Context(), db, tx0, 0, 1, false)
	be.Err(t, got, nil)
	_ = tx0.Rollback()

	tx1, err := db.BeginTx(t.Context(), nil)
	be.Err(t, err, nil)
	got = model.UpdateBoolFrom(t.Context(), db, tx1, 4, 1, true)
	be.Err(t, got, nil)
	_ = tx1.Rollback()
}

func TestUpdateInt64From(t *testing.T) {
	t.Parallel()

	got := model.UpdateInt64From(t.Context(), nil, nil, 0, 0, "")
	be.Err(t, got)

	db := openDB(t)
	if db == nil {
		return
	}

	tx0, err := db.BeginTx(t.Context(), nil)
	be.Err(t, err, nil)

	got = model.UpdateInt64From(t.Context(), db, tx0, 0, 0, "false")
	be.Err(t, got)

	got = model.UpdateInt64From(t.Context(), db, tx0, 10, 1, "1000")
	be.Err(t, got)

	got = model.UpdateInt64From(t.Context(), db, tx0, 0, 1, "1000")
	be.Err(t, got, nil)
	_ = tx0.Rollback()

	tx1, err := db.BeginTx(t.Context(), nil)
	be.Err(t, err, nil)
	got = model.UpdateInt64From(t.Context(), db, tx1, 1, 1, "999")
	be.Err(t, got, nil)
	_ = tx1.Rollback()
}

func TestUpdateStringFrom(t *testing.T) {
	t.Parallel()

	got := model.UpdateStringFrom(t.Context(), nil, nil, 0, 0, "")
	be.Err(t, got)

	db := openDB(t)
	if db == nil {
		return
	}

	tx0, err := db.BeginTx(t.Context(), nil)
	be.Err(t, err, nil)

	got = model.UpdateStringFrom(t.Context(), db, tx0, 0, 0, "false")
	be.Err(t, got)

	got = model.UpdateStringFrom(t.Context(), db, tx0, 99, 1, "1000")
	be.Err(t, got)

	got = model.UpdateStringFrom(t.Context(), db, tx0, 0, 1, "1000")
	be.Err(t, got, nil)
	_ = tx0.Rollback()

	tx1, err := db.BeginTx(t.Context(), nil)
	be.Err(t, err, nil)
	got = model.UpdateStringFrom(t.Context(), db, tx1, 1, 1, "999")
	be.Err(t, got, nil)
	_ = tx1.Rollback()
}
