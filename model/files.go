package model

// Package file files.go contains the database queries for the listing of sorted files.

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Defacto2/server/internal/postgres"
	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

// Files contain statistics for every release.
type Files struct {
	Bytes   int `boil:"size_total"`
	Count   int `boil:"count_total"`
	MinYear int `boil:"min_year"`
	MaxYear int `boil:"max_year"`
}

// Stat returns the total number of files and the total size of all files that are not soft deleted.
func (f *Files) Stat(ctx context.Context, db *sql.DB) error {
	if db == nil {
		return ErrDB
	}
	if f.Bytes > 0 && f.Count > 0 {
		return nil
	}
	return models.NewQuery(
		qm.Select(postgres.Columns()...),
		qm.Where(ClauseNoSoftDel),
		qm.From(From)).Bind(ctx, db, f)
}

// SearchFilename returns a list of files that match the search terms.
// The search terms are matched against the filename column.
// The results are ordered by the filename column in ascending order.
func (f *Files) SearchFilename(ctx context.Context, db *sql.DB, terms []string) (models.FileSlice, error) {
	if db == nil {
		return nil, ErrDB
	}
	if terms == nil {
		return models.FileSlice{}, nil
	}
	mods := []qm.QueryMod{}
	for i, term := range terms {
		if i == 0 {
			mods = append(mods, qm.Where("filename ~ ? OR filename ILIKE ? OR filename ILIKE ? OR filename ILIKE ?",
				term, term+"%", "%"+term, "%"+term+"%"))
			continue
		}
		mods = append(mods, qm.Or("filename ~ ? OR filename ILIKE ? OR filename ILIKE ? OR filename ILIKE ?",
			term, term+"%", "%"+term, "%"+term+"%"))
	}
	mods = append(mods, qm.OrderBy("filename ASC"), qm.Limit(Maximum))
	return models.Files(mods...).All(ctx, db)
}

// SearchDescription returns a list of files that match the search terms.
// The search terms are matched against the record_title column.
// The results are ordered by the filename column in ascending order.
func (f *Files) SearchDescription(ctx context.Context, db *sql.DB, terms []string) (models.FileSlice, error) {
	if db == nil {
		return nil, ErrDB
	}
	if terms == nil {
		return models.FileSlice{}, nil
	}
	mods := []qm.QueryMod{}
	const clauseT = "to_tsvector(record_title) @@ to_tsquery(?)"
	const clauseC = "to_tsvector(comment) @@ to_tsquery(?)"
	for i, term := range terms {
		term = fmt.Sprintf("'%s'", term) // the single quotes are required for terms containing spaces
		if i == 0 {
			mods = append(mods, qm.Where(clauseT, term))
			mods = append(mods, qm.Or(clauseC, term))
			continue
		}
		mods = append(mods, qm.Or(clauseT, term))
		mods = append(mods, qm.Or(clauseC, term))
	}
	mods = append(mods, qm.Limit(Maximum))
	return models.Files(mods...).All(ctx, db)
}

// List returns a list of files reversed ordered by the ID column.
func (f *Files) List(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	if db == nil {
		return nil, ErrDB
	}
	if err := f.Stat(ctx, db); err != nil {
		return nil, fmt.Errorf("f.Stat: %w", err)
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
		return nil, fmt.Errorf("f.Stat: %w", err)
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
		return nil, fmt.Errorf("f.Stat: %w", err)
	}
	const clause = "date_issued_year DESC NULLS LAST, " +
		"date_issued_month DESC NULLS LAST, " +
		"date_issued_day DESC NULLS LAST"
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
		return nil, fmt.Errorf("f.Stat: %w", err)
	}
	const clause = "updatedat DESC"
	return models.Files(
		qm.OrderBy(clause),
		qm.Offset(calc(offset, limit)),
		qm.Limit(limit)).All(ctx, db)
}

func (f *Files) ListDeletions(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	if db == nil {
		return nil, ErrDB
	}
	if err := f.StatDeletions(ctx, db); err != nil {
		return nil, fmt.Errorf("f.Stat: %w", err)
	}
	boil.DebugMode = true
	const clause = "deletedat DESC"
	return models.Files(
		models.FileWhere.Deletedat.IsNotNull(),
		models.FileWhere.Deletedby.IsNotNull(),
		qm.WithDeleted(),
		qm.OrderBy(clause),
		qm.Offset(calc(offset, limit)),
		qm.Limit(limit)).All(ctx, db)
}

func (f *Files) ListUnwanted(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	if db == nil {
		return nil, ErrDB
	}
	if err := f.StatUnwanted(ctx, db); err != nil {
		return nil, fmt.Errorf("f.StatUnwanted: %w", err)
	}
	// boil.DebugMode = true
	const clause = "id DESC"
	return models.Files(
		models.FileWhere.FileSecurityAlertURL.IsNotNull(),
		qm.WithDeleted(),
		qm.OrderBy(clause),
		qm.Offset(calc(offset, limit)),
		qm.Limit(limit)).All(ctx, db)
}

func (f *Files) ListForApproval(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	if db == nil {
		return nil, ErrDB
	}
	if err := f.StatForApproval(ctx, db); err != nil {
		return nil, fmt.Errorf("f.StatForApproval: %w", err)
	}
	// boil.DebugMode = true
	const clause = "id DESC"
	return models.Files(
		models.FileWhere.Deletedat.IsNotNull(),
		qm.WithDeleted(),
		qm.OrderBy(clause),
		qm.Offset(calc(offset, limit)),
		qm.Limit(limit)).All(ctx, db)
}

func (f *Files) StatForApproval(ctx context.Context, db *sql.DB) error {
	if db == nil {
		return ErrDB
	}
	if f.Bytes > 0 && f.Count > 0 {
		return nil
	}
	// boil.DebugMode = true
	return models.NewQuery(
		models.FileWhere.Deletedat.IsNotNull(),
		qm.WithDeleted(),
		qm.Select(postgres.Columns()...),
		qm.From(From)).Bind(ctx, db, f)
}

func (f *Files) StatDeletions(ctx context.Context, db *sql.DB) error {
	if db == nil {
		return ErrDB
	}
	if f.Bytes > 0 && f.Count > 0 {
		return nil
	}
	// boil.DebugMode = true
	return models.NewQuery(
		models.FileWhere.Deletedat.IsNotNull(),
		models.FileWhere.Deletedby.IsNotNull(),
		qm.WithDeleted(),
		qm.Select(postgres.Columns()...),
		qm.From(From)).Bind(ctx, db, f)
}

func (f *Files) StatUnwanted(ctx context.Context, db *sql.DB) error {
	if db == nil {
		return ErrDB
	}
	if f.Bytes > 0 && f.Count > 0 {
		return nil
	}
	// boil.DebugMode = true
	return models.NewQuery(
		models.FileWhere.FileSecurityAlertURL.IsNotNull(),
		qm.WithDeleted(),
		qm.Select(postgres.Columns()...),
		qm.From(From)).Bind(ctx, db, f)
}
