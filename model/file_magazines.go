package model

import (
	"context"
	"database/sql"

	"github.com/Defacto2/server/model/modext"
	"github.com/Defacto2/server/pkg/postgres"
	"github.com/Defacto2/server/pkg/postgres/models"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type Mag struct {
	Bytes int `boil:"size_sum"`
	Count int `boil:"counter"`
	Year0 int `boil:"min_year"`
	YearX int `boil:"max_year"`
}

// Stat counts the total number and total byte size of releases magazine files.
func (m *Mag) Stat(ctx context.Context, db *sql.DB) error {
	return models.NewQuery(
		qm.Select(postgres.SumSize, postgres.Counter, postgres.MinYear, postgres.MaxYear),
		modext.MagExpr(),
		qm.From(From)).Bind(ctx, db, m)
}

// List returns a list of magazine files.
func (m *Mag) List(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	return models.Files(modext.MagExpr(),
		qm.Offset(calc(offset, limit)), qm.Limit(limit)).All(ctx, db)
}
