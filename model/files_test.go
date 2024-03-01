package model_test

import (
	"context"
	"testing"

	"github.com/Defacto2/server/internal/postgres"
	"github.com/Defacto2/server/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFiles(t *testing.T) {
	t.Parallel()
	files := model.Files{}
	ctx := context.TODO()
	err := files.Stat(ctx, nil)
	require.Error(t, err)

	db, err := postgres.ConnectDB()
	require.NoError(t, err)
	defer db.Close()
	err = files.Stat(ctx, db)
	require.Error(t, err)

	fs, err := files.SearchFilename(ctx, db, nil)
	require.NoError(t, err)
	assert.Empty(t, fs)
}
