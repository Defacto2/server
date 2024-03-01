package model_test

import (
	"context"
	"testing"

	"github.com/Defacto2/server/handler/app"
	"github.com/Defacto2/server/internal/postgres"
	"github.com/Defacto2/server/model"
	"github.com/stretchr/testify/assert"
)

func TestSummary_SearchDesc(t *testing.T) {
	t.Parallel()
	ctx := context.TODO()
	var s model.Summary
	err := s.SearchDesc(ctx, nil, nil)
	assert.Error(t, err)

	err = s.SearchDesc(ctx, nil, nil)
	assert.Error(t, err)

	db, err := postgres.ConnectDB()
	assert.NoError(t, err)
	defer db.Close()

	err = s.SearchDesc(ctx, db, nil)
	assert.Error(t, err)

	err = s.SearchDesc(ctx, db, []string{"search", "term"})
	assert.Error(t, err)
}

func TestSummary_SearchFilename(t *testing.T) {
	t.Parallel()
	ctx := context.TODO()
	var s model.Summary
	err := s.SearchFilename(ctx, nil, nil)
	assert.Error(t, err)
	err = s.SearchFilename(ctx, nil, nil)
	assert.Error(t, err)

	db, err := postgres.ConnectDB()
	assert.NoError(t, err)
	defer db.Close()

	err = s.SearchFilename(ctx, db, nil)
	assert.Error(t, err)

	err = s.SearchFilename(ctx, db, []string{"search.txt", "term.com"})
	assert.Error(t, err)
}

func TestSummary_All(t *testing.T) {
	t.Parallel()
	ctx := context.TODO()
	var s model.Summary
	err := s.All(ctx, nil)
	assert.Error(t, err)

	err = s.All(ctx, nil)
	assert.Error(t, err)

	db, err := postgres.ConnectDB()
	assert.NoError(t, err)
	defer db.Close()

	err = s.All(ctx, db)
	assert.Error(t, err)
}

func TestSummary_BBS(t *testing.T) {
	t.Parallel()
	ctx := context.TODO()
	var s model.Summary
	err := s.BBS(ctx, nil)
	assert.Error(t, err)

	err = s.BBS(ctx, nil)
	assert.Error(t, err)

	db, err := postgres.ConnectDB()
	assert.NoError(t, err)
	defer db.Close()

	err = s.BBS(ctx, db)
	assert.Error(t, err)
}

func TestSummary_Scener(t *testing.T) {
	t.Parallel()
	var s model.Summary
	ctx := context.TODO()
	err := s.Scener(ctx, nil, "")
	assert.Error(t, err)

	err = s.Scener(ctx, nil, "")
	assert.Error(t, err)

	db, err := postgres.ConnectDB()
	assert.NoError(t, err)
	defer db.Close()

	err = s.Scener(ctx, db, "")
	assert.Error(t, err)
	err = s.Scener(ctx, db, "006")
	assert.Error(t, err)
}

func TestSummary_Releaser(t *testing.T) {
	t.Parallel()
	var s model.Summary
	ctx := context.TODO()
	err := s.Releaser(ctx, nil, "")
	assert.Error(t, err)

	err = s.Releaser(ctx, nil, "")
	assert.Error(t, err)

	db, err := postgres.ConnectDB()
	assert.NoError(t, err)
	defer db.Close()

	err = s.Releaser(ctx, db, "")
	assert.Error(t, err)
	err = s.Releaser(ctx, db, "defacto2")
	assert.Error(t, err)
}

func TestSummary_URI(t *testing.T) {
	t.Parallel()
	var s model.Summary
	ctx := context.TODO()
	err := s.URI(ctx, nil, "")
	assert.Error(t, err)

	err = s.URI(ctx, nil, "")
	assert.Error(t, err)

	db, err := postgres.ConnectDB()
	assert.NoError(t, err)
	defer db.Close()

	for i := range 57 {
		uri := app.URI(i).String()
		err = s.URI(ctx, db, uri)
		assert.Error(t, err)
	}
}
