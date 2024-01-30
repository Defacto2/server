package model_test

import (
	"context"
	"log"
	"testing"

	"github.com/Defacto2/server/internal/postgres"
	"github.com/Defacto2/server/model"
	"github.com/stretchr/testify/assert"
)

func TestScenerSQL(t *testing.T) {
	t.Parallel()

	s := model.ScenerSQL("")
	assert.NotEmpty(t, s, "")

	s = model.ScenerSQL("defacto2")
	assert.Contains(t, s, "DEFACTO2")
}

func TestSceners(t *testing.T) {
	t.Parallel()

	var s model.Sceners
	ctx := context.Background()
	err := s.All(ctx, nil)
	assert.Error(t, err)

	db, err := postgres.ConnectDB()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	err = s.All(ctx, db)
	assert.Error(t, err)

	err = s.Writer(ctx, db)
	assert.Error(t, err)
	err = s.Artist(ctx, db)
	assert.Error(t, err)
	err = s.Coder(ctx, db)
	assert.Error(t, err)
	err = s.Musician(ctx, db)
	assert.Error(t, err)

	x := s.Sort()
	assert.Empty(t, x)
}
