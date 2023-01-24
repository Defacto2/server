package model

import (
	"context"
	"database/sql"

	"github.com/Defacto2/server/pkg/postgres/models"
	"github.com/Defacto2/server/pkg/tags"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

// Docs contain statistics for releases that could be considered documents.
type Docs struct {
	Bytes int `boil:"size_sum"`
	Count int `boil:"counter"`
}

// Stat counts the total number and total byte size of releases that could be considered documents.
func (d *Docs) Stat(ctx context.Context, db *sql.DB) error {
	if d.Bytes > 0 && d.Count > 0 {
		return nil
	}
	return models.NewQuery(
		qm.Select(SumSize, Counter),
		DocumentExpr(),
		qm.From(From)).Bind(ctx, db, d)
}

// DocumentExpr is a the query mod expression for document files.
func DocumentExpr() qm.QueryMod {
	ansi := null.String{String: tags.URIs[tags.ANSI], Valid: true}
	text := null.String{String: tags.URIs[tags.Text], Valid: true}
	amiga := null.String{String: tags.URIs[tags.TextAmiga], Valid: true}
	pdf := null.String{String: tags.URIs[tags.PDF], Valid: true}
	return qm.Expr(
		models.FileWhere.Platform.EQ(ansi),
		qm.Or2(models.FileWhere.Platform.EQ(text)),
		qm.Or2(models.FileWhere.Platform.EQ(amiga)),
		qm.Or2(models.FileWhere.Platform.EQ(pdf)),
	)
}
