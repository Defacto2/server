// Package model_test requires an active database connection.
package model_test

import (
	"context"
	"testing"

	"github.com/Defacto2/server/internal/postgres"
	"github.com/Defacto2/server/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOne(t *testing.T) {
	t.Parallel()
	ctx := context.TODO()
	one, err := model.One(ctx, nil, false, -1)
	require.Error(t, err)
	assert.Nil(t, one)

	one, err = model.One(ctx, nil, false, -1)
	require.Error(t, err)
	assert.Nil(t, one)

	db, err := postgres.ConnectDB()
	require.NoError(t, err)
	defer db.Close()

	one, err = model.One(ctx, db, false, -1)
	require.Error(t, err)
	assert.Nil(t, one)

	one, err = model.One(ctx, db, false, 1)
	// there's no db password so an error will be returned.
	require.Error(t, err)
	assert.Nil(t, one)
}

func TestByteCountByCategory(t *testing.T) {
	t.Parallel()
	ctx := context.TODO()
	i, err := model.ByteCountByCategory(ctx, nil, "")
	require.Error(t, err)
	assert.Zero(t, i)

	i, err = model.ByteCountByCategory(ctx, nil, "")
	require.Error(t, err)
	assert.Zero(t, i)

	db, err := postgres.ConnectDB()
	require.NoError(t, err)
	defer db.Close()

	i, err = model.ByteCountByCategory(ctx, db, "")
	require.Error(t, err)
	assert.Zero(t, i)

	i, err = model.ByteCountByCategory(ctx, db, "bbs")
	require.Error(t, err)
	assert.Zero(t, i)
}

func TestByteCountByReleaser(t *testing.T) {
	t.Parallel()
	ctx := context.TODO()
	i, err := model.ByteCountByReleaser(ctx, nil, "")
	require.Error(t, err)
	assert.Zero(t, i)

	i, err = model.ByteCountByReleaser(ctx, nil, "")
	require.Error(t, err)
	assert.Zero(t, i)

	db, err := postgres.ConnectDB()
	require.NoError(t, err)
	defer db.Close()

	i, err = model.ByteCountByReleaser(ctx, db, "")
	require.Error(t, err)
	assert.Zero(t, i)

	i, err = model.ByteCountByReleaser(ctx, db, "bbs")
	require.Error(t, err)
	assert.Zero(t, i)
}

func TestByteCountByPlatform(t *testing.T) {
	t.Parallel()
	ctx := context.TODO()
	i, err := model.ByteCountByPlatform(ctx, nil, "")
	require.Error(t, err)
	assert.Zero(t, i)

	i, err = model.ByteCountByPlatform(ctx, nil, "")
	require.Error(t, err)
	assert.Zero(t, i)

	db, err := postgres.ConnectDB()
	require.NoError(t, err)
	defer db.Close()

	i, err = model.ByteCountByPlatform(ctx, db, "")
	require.Error(t, err)
	assert.Zero(t, i)

	i, err = model.ByteCountByPlatform(ctx, db, "bbs")
	require.Error(t, err)
	assert.Zero(t, i)
}

func TestCountByCategory(t *testing.T) {
	t.Parallel()
	ctx := context.TODO()
	i, err := model.CountByCategory(ctx, nil, "")
	require.Error(t, err)
	assert.Zero(t, i)

	i, err = model.CountByCategory(ctx, nil, "")
	require.Error(t, err)
	assert.Zero(t, i)

	db, err := postgres.ConnectDB()
	require.NoError(t, err)
	defer db.Close()

	i, err = model.CountByCategory(ctx, db, "")
	require.Error(t, err)
	assert.Zero(t, i)

	i, err = model.CountByCategory(ctx, db, "bbs")
	require.Error(t, err)
	assert.Zero(t, i)
}

func TestCountByPlatform(t *testing.T) {
	t.Parallel()
	ctx := context.TODO()
	i, err := model.CountByPlatform(ctx, nil, "")
	require.Error(t, err)
	assert.Zero(t, i)

	i, err = model.CountByPlatform(ctx, nil, "")
	require.Error(t, err)
	assert.Zero(t, i)

	db, err := postgres.ConnectDB()
	require.NoError(t, err)
	defer db.Close()

	i, err = model.CountByPlatform(ctx, db, "")
	require.Error(t, err)
	assert.Zero(t, i)

	i, err = model.CountByPlatform(ctx, db, "bbs")
	require.Error(t, err)
	assert.Zero(t, i)
}
