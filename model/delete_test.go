package model_test

import (
	"database/sql"
	"testing"

	"github.com/Defacto2/server/internal/testutil"
	"github.com/Defacto2/server/model"
	"github.com/nalgeon/be"
)

func TestDeleteOne(t *testing.T) {
	t.Parallel()

	be.Err(t, model.DeleteOne(t.Context(), nil, 1))

	tx := testutil.Tx(t)

	be.Err(t, model.DeleteOne(t.Context(), tx, 0), model.ErrBadIDInt)
	be.Err(t, model.DeleteOne(t.Context(), tx, -1), model.ErrBadIDInt)

	const newProd = 1000
	key, _, err := model.InsertDemozoo(t.Context(), tx, newProd)
	be.Err(t, err, nil)
	be.True(t, key > 0)

	err = model.DeleteOne(t.Context(), tx, key)
	be.Err(t, err, nil)

	_, err = model.One(t.Context(), tx, true, int(key))
	be.Err(t, err, sql.ErrNoRows)
}
