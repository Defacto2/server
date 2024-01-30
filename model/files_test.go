package model_test

import (
	"context"
	"testing"

	"github.com/Defacto2/server/internal/postgres"
	"github.com/Defacto2/server/model"
	"github.com/stretchr/testify/assert"
)

func FilesFiles(t *testing.T) {
	t.Parallel()
	files := model.Files{}
	ctx := context.TODO()
	err := files.Stat(ctx, nil)
	assert.Error(t, err)

	db, err := postgres.ConnectDB()
	assert.NoError(t, err)
	defer db.Close()
	err = files.Stat(ctx, db)
	assert.Error(t, err)

	fs, err := files.SearchFilename(ctx, db, nil)
	assert.NoError(t, err)
	assert.Empty(t, fs)
}
