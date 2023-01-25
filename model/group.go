package model

import (
	"context"
	"database/sql"

	"github.com/Defacto2/server/pkg/helpers"
	"github.com/Defacto2/server/pkg/postgres/models"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

// Groups contain statistics for releases that could be considered as digital or pixel art.
type Groups struct {
	Bytes int `boil:"size_sum"` // Unused.
	Count int `boil:"counter"`
}

// Stat counts the total number of unique groups.
func (g *Groups) Stat(ctx context.Context, db *sql.DB) error {
	if g.Count > 0 {
		return nil
	}
	r, err := models.Files(qm.SQL("SELECT DISTINCT group_brand, COUNT(*) as count "+
		"FROM files "+
		"CROSS JOIN LATERAL (values(group_brand_for),(group_brand_by)) AS T(group_brand) "+
		"WHERE NULLIF(group_brand, '') IS NOT NULL "+ // handle empty and null values
		"GROUP BY group_brand")).All(ctx, db)
	if err != nil {
		return err
	}
	g.Count = len(r)
	return nil
}

type Group struct {
	Name  string `boil:"group_brand"` // todo rename to group_brand
	URI   string // URI slug for the scener.
	Bytes int    `boil:"size_sum"`
	Count int    `boil:"count"`
}

type GroupS []*struct {
	Group Group `boil:",bind"`
}

// All the names and statistics of the unique groups.
func (g *GroupS) All(offset, limit int, o Order, ctx context.Context, db *sql.DB) error {
	if len(*g) > 0 {
		return nil
	}
	err := models.Files(
		qm.SQL("SELECT DISTINCT group_brand, "+
			"COUNT(group_brand) AS count, "+
			"SUM(files.filesize) AS size_sum "+
			"FROM files "+
			"CROSS JOIN LATERAL (values(group_brand_for),(group_brand_by)) AS T(group_brand) "+
			"WHERE NULLIF(group_brand, '') IS NOT NULL "+ // handle empty and null values
			"GROUP BY group_brand "+
			"ORDER BY count DESC"), //"+"LIMIT 500"
	).Bind(ctx, db, g)
	if err != nil {
		return err
	}
	g.Slugs()
	return nil
}

// Slugs saves URL friendly strings to the Group names.
func (g *GroupS) Slugs() {
	for _, group := range *g {
		group.Group.URI = helpers.Slug(group.Group.Name)
	}
}
