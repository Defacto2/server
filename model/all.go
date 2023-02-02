package model

import (
	"context"
	"database/sql"

	"github.com/Defacto2/server/pkg/postgres/models"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

// All contain statistics for every release.
type All struct {
	Bytes int `boil:"size_sum"`
	Count int `boil:"counter"`
}

// Stat counts the total number and total byte size of releases that could be considered as digital or pixel art.
func (a *All) Stat(ctx context.Context, db *sql.DB) error {
	if a.Bytes > 0 && a.Count > 0 {
		return nil
	}
	return models.NewQuery(
		qm.Select(SumSize, Counter),
		qm.From(From)).Bind(ctx, db, a)
}