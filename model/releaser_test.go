package model_test

import (
	"context"
	"log"
	"testing"

	"github.com/Defacto2/server/internal/postgres"
	"github.com/Defacto2/server/model"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
)

func TestReleaserNames_List(t *testing.T) {
	ctx := context.Background()
	db, err := postgres.ConnectDB()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	var g model.ReleaserNames
	err = g.List(ctx, db)
	assert.Error(t, err)
	assert.Empty(t, g)
}

func TestReleasers_List(t *testing.T) {
	ctx := context.Background()
	db, err := postgres.ConnectDB()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	var g model.Releasers
	list, err := g.List(ctx, db, "")
	assert.Error(t, err)
	assert.Empty(t, list)

	list, err = g.List(ctx, db, "defacto2")
	assert.Error(t, err)
	assert.Empty(t, list)
}

func TestReleasers_Magazine(t *testing.T) {
	ctx := context.Background()
	db, err := postgres.ConnectDB()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	var r model.Releasers
	err = r.Magazine(ctx, db)
	assert.Error(t, err)
}

func TestReleasers_MagazineAZ(t *testing.T) {
	ctx := context.Background()
	db, err := postgres.ConnectDB()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	var r model.Releasers
	err = r.MagazineAZ(ctx, db)
	assert.Error(t, err)
}

func TestReleasers_BBS(t *testing.T) {
	ctx := context.Background()
	db, err := postgres.ConnectDB()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	var r model.Releasers
	err = r.BBS(ctx, db, false)
	assert.Error(t, err)
}

func TestReleasers_FTP(t *testing.T) {
	ctx := context.Background()
	db, err := postgres.ConnectDB()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	var r model.Releasers
	err = r.FTP(ctx, db)
	assert.Error(t, err)
}
