package model

// Package html3_document.go contains the database queries the HTML3 document category.

import (
	"context"
	"database/sql"

	"github.com/Defacto2/server/model/expr"
	"github.com/Defacto2/server/pkg/postgres"
	"github.com/Defacto2/server/pkg/postgres/models"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

// Docs contain statistics for releases that could be considered documents.
type Docs struct {
	Bytes int `boil:"size_sum"`
	Count int `boil:"counter"`
}

func (d *Docs) Stat(ctx context.Context, db *sql.DB) error {
	if db == nil {
		return ErrDB
	}
	if d.Bytes > 0 && d.Count > 0 {
		return nil
	}
	return models.NewQuery(
		qm.Select(postgres.Stat()...),
		DocumentExpr(),
		qm.From(From)).Bind(ctx, db, d)
}

func DocumentExpr() qm.QueryMod {
	return qm.Expr(
		models.FileWhere.Platform.EQ(expr.PAnsi()),
		qm.Or2(models.FileWhere.Platform.EQ(expr.PText())),
		qm.Or2(models.FileWhere.Platform.EQ(expr.PTextAmiga())),
		qm.Or2(models.FileWhere.Platform.EQ(expr.PPdf())),
	)
}
