package model

// This file is the custom document category for the HTML3 template.

import (
	"context"
	"database/sql"

	"github.com/Defacto2/server/model/modext"
	"github.com/Defacto2/server/pkg/postgres"
	"github.com/Defacto2/server/pkg/postgres/models"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

// Docs contain statistics for releases that could be considered documents.
type Docs struct {
	Bytes int `boil:"size_sum"`
	Count int `boil:"counter"`
}

// Stat counts the total number and total byte size of releases that could be considered documents.
func (d *Docs) Stat(ctx context.Context, db *sql.DB) error {
	if db == nil {
		return ErrDB
	}
	if d.Bytes > 0 && d.Count > 0 {
		return nil
	}
	return models.NewQuery(
		qm.Select(postgres.SumSize, postgres.Counter),
		DocumentExpr(),
		qm.From(From)).Bind(ctx, db, d)
}

// DocumentExpr is a the query mod expression for document files.
func DocumentExpr() qm.QueryMod {
	return qm.Expr(
		models.FileWhere.Platform.EQ(modext.PAnsi()),
		qm.Or2(models.FileWhere.Platform.EQ(modext.PText())),
		qm.Or2(models.FileWhere.Platform.EQ(modext.PTextAmiga())),
		qm.Or2(models.FileWhere.Platform.EQ(modext.PPdf())),
	)
}
