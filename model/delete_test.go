package model_test

import (
	"testing"

	"github.com/Defacto2/server/model"
	"github.com/nalgeon/be"
)

func TestDeleteOne(t *testing.T) {
	t.Parallel()

	db := openDB(t)
	if db == nil {
		return
	}
	tx0, err := db.BeginTx(t.Context(), nil)
	be.Err(t, err, nil)
	tx1, err := db.BeginTx(t.Context(), nil)
	be.Err(t, err, nil)

	const newProd = 1000
	key, _, got := model.InsertDemozoo(t.Context(), tx0, newProd)
	be.Err(t, got, nil)
	be.True(t, key > 0)

	got = model.DeleteOne(t.Context(), tx1, key)
	be.Err(t, got, nil)

	t.Cleanup(func() {
		_ = tx0.Rollback()
	})
}
