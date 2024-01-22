package model_test

import (
	"context"
	"log"
	"testing"

	"github.com/Defacto2/releaser"
	"github.com/Defacto2/server/internal/helper"
	"github.com/Defacto2/server/internal/postgres"
	"github.com/Defacto2/server/model"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
)

func TestAllSlugs(t *testing.T) {
	ctx := context.Background()
	db, err := postgres.ConnectDB()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	var g model.Releasers
	if err := g.All(ctx, db, 0, 0, false); err != nil {
		log.Println(err)
		return
	}
	for _, x := range g {
		og := releaser.Clean(x.Unique.Name)
		y := helper.Slug(og)
		z := releaser.Clean(y)
		assert.Equal(t, og, z, "slug is "+y)
	}
}
