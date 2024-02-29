package fix_test

import (
	"context"
	"testing"

	"github.com/Defacto2/server/internal/postgres"
	"github.com/Defacto2/server/model/fix"
	"github.com/stretchr/testify/assert"
)

func TestRepair(t *testing.T) {
	t.Parallel()
	var r fix.Repair

	ctx := context.TODO()
	err := r.Run(ctx, nil)
	assert.ErrorIs(t, err, fix.ErrDB)

	db, err := postgres.ConnectDB()
	assert.NoError(t, err)

	err = r.Run(ctx, db)
	assert.Error(t, err)
}
