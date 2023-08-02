package model

import (
	"context"
	"database/sql"

	"github.com/Defacto2/server/pkg/postgres"
	"github.com/Defacto2/server/pkg/postgres/models"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

// Windows contain statistics for software releases that requires the Windows operating system.
type Windows struct {
	Bytes int `boil:"size_sum"`
	Count int `boil:"counter"`
	Year0 int `boil:"min_year"`
	YearX int `boil:"max_year"`
}

// Stat counts the total number and total byte size of releases that could be considered as digital or pixel art.
func (w *Windows) Stat(ctx context.Context, db *sql.DB) error {
	if w.Bytes > 0 && w.Count > 0 {
		return nil
	}
	return models.NewQuery(
		qm.Select(postgres.SumSize, postgres.Counter, postgres.MinYear, postgres.MaxYear),
		models.FileWhere.Platform.EQ(windows()),
		qm.From(From)).Bind(ctx, db, w)
}
