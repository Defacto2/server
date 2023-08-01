package model

import (
	"context"
	"database/sql"

	"github.com/Defacto2/server/pkg/postgres"
	"github.com/Defacto2/server/pkg/postgres/models"
	"github.com/Defacto2/server/pkg/tags"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

// Windows contain statistics for software releases that requires the Windows operating system.
type Windows struct {
	Bytes int `boil:"size_sum"`
	Count int `boil:"counter"`
}

// Stat counts the total number and total byte size of releases that could be considered as digital or pixel art.
func (w *Windows) Stat(ctx context.Context, db *sql.DB) error {
	if w.Bytes > 0 && w.Count > 0 {
		return nil
	}
	return models.NewQuery(
		qm.Select(postgres.SumSize, postgres.Counter),
		WinExpr(),
		qm.From(From)).Bind(ctx, db, w)
}

// WinExpr is a the query mod expression for windows releases.
func WinExpr() qm.QueryMod {
	windows := null.String{String: tags.URIs()[tags.Windows], Valid: true}
	return qm.Expr(
		models.FileWhere.Platform.EQ(windows),
	)
}
