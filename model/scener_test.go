package model_test

import (
	"context"
	"log"
	"testing"

	"github.com/Defacto2/server/internal/postgres"
	"github.com/Defacto2/server/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSceners(t *testing.T) {
	t.Parallel()

	var s model.Sceners
	ctx := context.Background()
	err := s.All(ctx, nil)
	require.Error(t, err)

	db, err := postgres.ConnectDB()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	err = s.All(ctx, db)
	require.Error(t, err)

	err = s.Writer(ctx, db)
	require.Error(t, err)
	err = s.Artist(ctx, db)
	require.Error(t, err)
	err = s.Coder(ctx, db)
	require.Error(t, err)
	err = s.Musician(ctx, db)
	require.Error(t, err)

	x := s.Sort()
	assert.Empty(t, x)
}
