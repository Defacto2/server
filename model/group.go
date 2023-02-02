package model

import (
	"context"
	"database/sql"

	"github.com/Defacto2/server/pkg/helpers"
	"github.com/Defacto2/server/pkg/postgres"
	"github.com/Defacto2/server/pkg/postgres/models"
	"github.com/volatiletech/sqlboiler/v4/queries"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

// GroupStats are statistics of the unique groups.
type GroupStats struct {
	Count int `boil:"counter"`
}

// Stat counts the total number of unique groups.
func (g *GroupStats) Stat(ctx context.Context, db *sql.DB) error {
	if g.Count > 0 {
		return nil
	}
	r, err := models.Files(qm.SQL(postgres.SQLGroupStat())).All(ctx, db)
	if err != nil {
		return err
	}
	g.Count = len(r)
	return nil
}

type Group struct {
	Name  string `boil:"group_brand"`
	URI   string // URI slug for the scener.
	Bytes int    `boil:"size_sum"`
	Count int    `boil:"count"`
}

type Groups []*struct {
	Group Group `boil:",bind"`
}

// All the names and statistics of the unique groups.
func (g *Groups) All(offset, limit int, o Order, ctx context.Context, db *sql.DB) error {
	if len(*g) > 0 {
		return nil
	}
	if err := queries.Raw(postgres.SQLGroupAll()).Bind(ctx, db, g); err != nil {
		return err
	}
	g.Slugs()
	return nil
}

// Slugs saves URL friendly strings to the Group names.
func (g *Groups) Slugs() {
	for _, group := range *g {
		group.Group.URI = helpers.Slug(group.Group.Name)
	}
}
