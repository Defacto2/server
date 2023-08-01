package model

import (
	"context"
	"database/sql"

	"github.com/Defacto2/server/pkg/postgres"
	"github.com/Defacto2/server/pkg/postgres/models"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type Linux struct {
	Bytes int `boil:"size_sum"`
	Count int `boil:"counter"`
	Year0 int `boil:"min_year"`
	YearX int `boil:"max_year"`
}

func (l *Linux) Stat(ctx context.Context, db *sql.DB) error {
	// if i.Bytes > 0 && i.Count > 0 {
	// 	return nil
	// }
	return models.NewQuery(
		qm.Select(postgres.SumSize, postgres.Counter, postgres.MinYear, postgres.MaxYear),
		qm.Expr(
			// TODO: handle incorrect EQ value in sqlboiler, ie Section.EQ(linux()) causing a panic.
			models.FileWhere.Platform.EQ(linux()),
		),
		qm.From(From)).Bind(ctx, db, l)
}
