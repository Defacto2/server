package model

import (
	"context"
	"database/sql"

	"github.com/Defacto2/server/pkg/postgres"
	"github.com/Defacto2/server/pkg/postgres/models"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

// Files contain statistics for every release.
type Files struct {
	Bytes int `boil:"size_sum"`
	Count int `boil:"counter"`
	Year0 int `boil:"min_year"`
	YearX int `boil:"max_year"`
}

// Stat counts the total number and total byte size of releases that could be considered as digital or pixel art.
func (f *Files) Stat(ctx context.Context, db *sql.DB) error {
	if f.Bytes > 0 && f.Count > 0 {
		return nil
	}
	return models.NewQuery(
		qm.Select(postgres.SumSize, postgres.Counter, postgres.MinYear, postgres.MaxYear),
		qm.From(From)).Bind(ctx, db, f)
}

// List returns all of the file records.
func (f *Files) List(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	if err := f.Stat(ctx, db); err != nil {
		return nil, err
	}
	return models.Files(qm.OrderBy("id DESC"),
		qm.Offset(calc(offset, limit)), qm.Limit(limit)).All(ctx, db)
}

// ListOldest returns all of the file records sorted by the date issued.
func (f *Files) ListOldest(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	if err := f.Stat(ctx, db); err != nil {
		return nil, err
	}
	return models.Files(qm.OrderBy("date_issued_year ASC NULLS LAST, date_issued_month ASC NULLS LAST, date_issued_day ASC NULLS LAST"),
		qm.Offset(calc(offset, limit)), qm.Limit(limit)).All(ctx, db)
}

// ListNewest returns all of the file records sorted by the date issued.
func (f *Files) ListNewest(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	if err := f.Stat(ctx, db); err != nil {
		return nil, err
	}
	return models.Files(qm.OrderBy("date_issued_year DESC NULLS LAST, date_issued_month DESC NULLS LAST, date_issued_day DESC NULLS LAST"),
		qm.Offset(calc(offset, limit)), qm.Limit(limit)).All(ctx, db)
}

// ListUpdates returns all of the file records sorted by the date updated.
func (f *Files) ListUpdates(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	if err := f.Stat(ctx, db); err != nil {
		return nil, err
	}
	// TODO: rename psql column from `updated_at` to `date_updated`
	return models.Files(qm.OrderBy("updatedat DESC"),
		qm.Offset(calc(offset, limit)), qm.Limit(limit)).All(ctx, db)
}
