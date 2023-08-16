package model

// Package file files.go contains the database queries for the listing of sorted files.

import (
	"context"
	"database/sql"

	"github.com/Defacto2/server/pkg/postgres"
	"github.com/Defacto2/server/pkg/postgres/models"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

// Files contain statistics for every release.
type Files struct {
	Bytes   int `boil:"size_sum"`
	Count   int `boil:"counter"`
	MinYear int `boil:"min_year"`
	MaxYear int `boil:"max_year"`
}

func (f *Files) Stat(ctx context.Context, db *sql.DB) error {
	if db == nil {
		return ErrDB
	}
	if f.Bytes > 0 && f.Count > 0 {
		return nil
	}
	//boil.DebugMode = true
	return models.NewQuery(
		qm.Select(postgres.Columns()...),
		qm.Where(ClauseNoSoftDel),
		qm.From(From)).Bind(ctx, db, f)
}

// List returns a list of files reversed ordered by the ID column.
func (f *Files) List(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	if db == nil {
		return nil, ErrDB
	}
	if err := f.Stat(ctx, db); err != nil {
		return nil, err
	}
	const clause = "id DESC"
	return models.Files(
		qm.OrderBy(clause),
		qm.Offset(calc(offset, limit)),
		qm.Limit(limit)).All(ctx, db)
}

// ListOldest returns all of the file records sorted by the date issued.
func (f *Files) ListOldest(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	if db == nil {
		return nil, ErrDB
	}
	if err := f.Stat(ctx, db); err != nil {
		return nil, err
	}
	const clause = "date_issued_year ASC NULLS LAST, " +
		"date_issued_month ASC NULLS LAST, " +
		"date_issued_day ASC NULLS LAST"
	return models.Files(
		qm.OrderBy(clause),
		qm.Offset(calc(offset, limit)),
		qm.Limit(limit)).All(ctx, db)
}

// ListNewest returns all of the file records sorted by the date issued.
func (f *Files) ListNewest(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	if db == nil {
		return nil, ErrDB
	}
	if err := f.Stat(ctx, db); err != nil {
		return nil, err
	}
	const clause = "date_issued_year DESC NULLS LAST, " +
		"date_issued_month DESC NULLS LAST, " +
		"date_issued_day DESC NULLS LAST"
	//boil.DebugMode = true
	return models.Files(
		qm.OrderBy(clause),
		qm.Offset(calc(offset, limit)),
		qm.Limit(limit)).All(ctx, db)
}

// ListUpdates returns all of the file records sorted by the date updated.
func (f *Files) ListUpdates(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	if db == nil {
		return nil, ErrDB
	}
	if err := f.Stat(ctx, db); err != nil {
		return nil, err
	}
	// TODO: rename PSQL column from `updated_at` to `date_updated`
	const clause = "updatedat DESC"
	return models.Files(
		qm.OrderBy(clause),
		qm.Offset(calc(offset, limit)),
		qm.Limit(limit)).All(ctx, db)
}
