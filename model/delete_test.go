package model_test

import (
	"testing"

	"github.com/Defacto2/server/internal/testutil"
	"github.com/Defacto2/server/model"
	"github.com/nalgeon/be"
)

func TestDeleteOne(t *testing.T) {
	t.Parallel()

	tx0 := testutil.Tx(t)
	tx1 := testutil.Tx(t)

	const newProd = 1000
	key, _, got := model.InsertDemozoo(t.Context(), tx0, newProd)
	be.Err(t, got, nil)
	be.True(t, key > 0)

	got = model.DeleteOne(t.Context(), tx1, key)
	be.Err(t, got, nil)
}
