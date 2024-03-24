package fix_test

import (
	"context"
	"testing"

	"github.com/Defacto2/server/internal/postgres"
	"github.com/Defacto2/server/model/fix"
	"github.com/stretchr/testify/require"
)

func TestRepair(t *testing.T) {
	t.Parallel()
	var r fix.Repair

	ctx := context.TODO()
	err := r.Run(ctx, nil, nil)
	require.ErrorIs(t, err, fix.ErrDB)

	db, err := postgres.ConnectDB()
	require.NoError(t, err)

	err = r.Run(ctx, nil, db)
	require.Error(t, err)
}
