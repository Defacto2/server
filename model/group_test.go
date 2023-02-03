package model_test

import (
	"context"
	"log"
	"testing"

	"github.com/Defacto2/sceners"
	"github.com/Defacto2/server/model"
	"github.com/Defacto2/server/pkg/helpers"
	"github.com/Defacto2/server/pkg/postgres"
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

	var g model.Groups
	if err := g.All(ctx, db, 0, 0, model.NameAsc); err != nil {
		log.Fatal(err)
	}
	for _, x := range g {
		og := sceners.CleanURL(x.Group.Name)
		y := helpers.Slug(og)
		z := sceners.CleanURL(y)
		assert.Equal(t, og, z, "slug is "+y)
	}
}
