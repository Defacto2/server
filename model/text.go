package model

import (
	"context"
	"database/sql"

	"github.com/Defacto2/server/pkg/postgres"
	"github.com/Defacto2/server/pkg/postgres/models"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type Text struct {
	Bytes int `boil:"size_sum"`
	Count int `boil:"counter"`
	Year0 int `boil:"min_year"`
	YearX int `boil:"max_year"`
}

func (t *Text) Stat(ctx context.Context, db *sql.DB) error {
	// if a.Bytes > 0 && a.Count > 0 {
	// 	return nil
	// }
	return models.NewQuery(
		qm.Select(postgres.SumSize, postgres.Counter, postgres.MinYear, postgres.MaxYear),
		qm.Expr(
			models.FileWhere.Platform.EQ(text()),
		),
		qm.From(From)).Bind(ctx, db, t)
}