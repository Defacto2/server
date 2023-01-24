package model

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Defacto2/server/pkg/postgres/models"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type Count int // Count is the number of found files.

// Counts caches the number of found files fetched from SQL queries.
var Counts = map[int]Count{
	Art:  0,
	Doc:  0,
	Soft: 0,
}

const (
	Art  int = iota // Art are digital + pixel art files.
	Doc             // Doc are document + text art files.
	Soft            // Soft are software files.
)

// One returns the record associated with the key ID.
func One(key int, ctx context.Context, db *sql.DB) (*models.File, error) {
	file, err := models.Files(models.FileWhere.ID.EQ(int64(key))).One(ctx, db)
	if err != nil {
		return nil, err
	}
	return file, err
}

// ByteCountByCategory sums the byte filesizes for all the files that match the category name.
func ByteCountByCategory(name string, ctx context.Context, db *sql.DB) (int64, error) {
	i, err := models.Files(
		qm.SQL("SELECT sum(files.filesize) FROM files WHERE section = $1",
			null.StringFrom(name)),
	).Count(ctx, db)
	if err != nil {
		return 0, fmt.Errorf("bytecount by section %q: %w", name, err)
	}
	return i, nil
}

// ByteCountByGroup sums the byte filesizes for all the files that match the group name.
func ByteCountByGroup(name string, ctx context.Context, db *sql.DB) (int64, error) {
	x := null.StringFrom(name)
	i, err := models.Files(qm.SQL("SELECT SUM(filesize) as size_sum FROM files WHERE group_brand_for = $1", x)).Count(ctx, db)
	if err != nil {
		return 0, fmt.Errorf("bytecount by group %q: %w", name, err)
	}
	return i, nil
}

// ByteCountByPlatform sums the byte filesizes for all the files that match the category name.
func ByteCountByPlatform(name string, ctx context.Context, db *sql.DB) (int64, error) {
	i, err := models.Files(
		qm.SQL("SELECT sum(filesize) FROM files WHERE platform = $1",
			null.StringFrom(name)),
	).Count(ctx, db)
	if err != nil {
		return 0, fmt.Errorf("bytecount by platform %q: %w", name, err)
	}
	return i, nil
}

// CountByCategory counts the files that match the named category.
func CountByCategory(name string, ctx context.Context, db *sql.DB) (int64, error) {
	x := null.StringFrom(name)
	return models.Files(models.FileWhere.Section.EQ(x)).Count(ctx, db)
}

// CountByPlatform counts the files that match the named category.
func CountByPlatform(name string, ctx context.Context, db *sql.DB) (int64, error) {
	x := null.StringFrom(name)
	return models.Files(models.FileWhere.Platform.EQ(x)).Count(ctx, db)
}