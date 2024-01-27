package html3

// Package html3_document.go contains the database queries the HTML3 document category.

import (
	"context"
	"database/sql"

	"github.com/Defacto2/server/internal/postgres"
	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/Defacto2/server/model/expr"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

// Documents contain statistics for releases that could be considered documents.
type Documents struct {
	Bytes int `boil:"size_total"`
	Count int `boil:"count_total"`
}

// Stat returns the total bytes and count of releases that could be considered documents.
func (d *Documents) Stat(ctx context.Context, db *sql.DB) error {
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

// DocumentExpr returns a query modifier for the document category.
func DocumentExpr() qm.QueryMod {
	return qm.Expr(
		models.FileWhere.Platform.EQ(expr.PAnsi()),
		qm.Or2(models.FileWhere.Platform.EQ(expr.PText())),
		qm.Or2(models.FileWhere.Platform.EQ(expr.PTextAmiga())),
		qm.Or2(models.FileWhere.Platform.EQ(expr.PPdf())),
	)
}
