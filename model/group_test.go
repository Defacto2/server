package model_test

import (
	"context"
	"testing"

	"github.com/Defacto2/sceners"
	"github.com/Defacto2/server/model"
	"github.com/Defacto2/server/pkg/postgres"
	"github.com/labstack/gommon/log"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
)

func TestSlug(t *testing.T) {
	tests := []struct {
		name      string
		expect    string
		assertion assert.ComparisonAssertionFunc
	}{
		{"the-group", "the_group", assert.Equal},
		{"group1, group2", "group1*group2", assert.Equal},
		{"group1 & group2", "group1-ampersand-group2", assert.Equal},
		{"group 1, group 2", "group-1*group-2", assert.Equal},
		{"GROUP 👾", "group", assert.Equal},
	}
	for _, tt := range tests {
		t.Run(tt.expect, func(t *testing.T) {
			tt.assertion(t, tt.expect, model.Slug(tt.name))
		})
	}
}

/* TODO:
Error:      	Not equal:
            	expected: "Mooñpeople"
            	actual  : "Moopeople"

				use utf8 lib to detect extended chars?
*/

func TestAllSlugs(t *testing.T) {
	ctx := context.Background()
	db, err := postgres.ConnectDB()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	fs, err := model.GroupList(ctx, db)
	if err != nil {
		log.Fatal(err)
	}
	for _, x := range fs {
		og := sceners.CleanURL(x.GroupBrandFor.String)
		y := model.Slug(og)
		z := sceners.CleanURL(y)
		assert.Equal(t, og, z, "slug is "+y)
	}
}
