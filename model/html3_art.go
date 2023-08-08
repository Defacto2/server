package model

// This file is the custom art category for the HTML3 template.

import (
	"context"
	"database/sql"

	"github.com/Defacto2/server/model/modext"
	"github.com/Defacto2/server/pkg/postgres"
	"github.com/Defacto2/server/pkg/postgres/models"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

// Arts contain statistics for releases that could be considered as digital or pixel art.
type Arts struct {
	Bytes int `boil:"size_sum"`
	Count int `boil:"counter"`
}

// Stat counts the total number and total byte size of releases that could be considered as digital or pixel art.
func (a *Arts) Stat(ctx context.Context, db *sql.DB) error {
	if db == nil {
		return ErrDB
	}
	if a.Bytes > 0 && a.Count > 0 {
		return nil
	}
	return models.NewQuery(
		qm.Select(postgres.SumSize, postgres.Counter),
		ArtExpr(),
		qm.From(From)).Bind(ctx, db, a)
}

// ArtExpr is a the query mod expression for art releases.
func ArtExpr() qm.QueryMod {
	return qm.Expr(
		models.FileWhere.Section.NEQ(modext.SBbs()),
		models.FileWhere.Platform.EQ(modext.PImage()),
	)
}
