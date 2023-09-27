package model

// Package file files.go contains the database queries for the listing of sorted files.

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"

	"github.com/Defacto2/server/internal/postgres"
	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

// Maximum number of files to return per query.
const Maximum = 998

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
	// boil.DebugMode = true
	return models.NewQuery(
		qm.Select(postgres.Columns()...),
		qm.Where(ClauseNoSoftDel),
		qm.From(From)).Bind(ctx, db, f)
	// return models.Files(qm.Limit(Maximum)).Bind(ctx, db, f)
}

// Search returns a list of files that match ....
func (f *Files) Search(ctx context.Context, db *sql.DB, terms []string) (models.FileSlice, error) {
	if db == nil {
		return nil, ErrDB
	}
	if terms == nil {
		return models.FileSlice{}, nil
	}
	// if err := f.Stat(ctx, db); err != nil { // TODO: remove this
	// 	return nil, err
	// }

	// match years
	// names := terms
	// names = slices.DeleteFunc(names, func(s string) bool {
	// 	i, err := strconv.Atoi(s)
	// 	if err != nil {
	// 		fmt.Println("x", s)
	// 		return false
	// 	}
	// 	const minYear = 1980
	// 	if i < minYear || i > time.Now().Year() {
	// 		fmt.Println("xx")
	// 		return false
	// 	}
	// 	return true
	// })

	// fmt.Println(names, "<<")

	// match file extensions
	// otherwise match string

	// const clause = "id DESC"
	// qm.Where("upper(group_brand_for) = ? OR upper(group_brand_by) = ?", x, x),
	// qm.Or2(models.FileWhere.Platform.EQ(expr.PText())
	mods := []qm.QueryMod{}
	for i, term := range terms {
		fmt.Println("-->", term, filepath.Ext(term) == term)
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
	// boil.DebugMode = true
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
	const clause = "updatedat DESC"
	return models.Files(
		qm.OrderBy(clause),
		qm.Offset(calc(offset, limit)),
		qm.Limit(limit)).All(ctx, db)
}
