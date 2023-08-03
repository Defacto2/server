package model

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/Defacto2/server/pkg/postgres"
	"github.com/Defacto2/server/pkg/postgres/models"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

// From is the name of the table containing records of files.
const From = "files"

// Cache returns true if the statistics are considered to be valid.
func Cache(b, c int, t time.Time) bool {
	fmt.Println(t.Before(time.Now().Add(-time.Hour * 1)))
	return b > 0 && c > 0 && t.Before(time.Now().Add(-time.Hour*1))
}

// One returns the record associated with the key ID.
func One(ctx context.Context, db *sql.DB, key int) (*models.File, error) {
	file, err := models.Files(models.FileWhere.ID.EQ(int64(key))).One(ctx, db)
	if err != nil {
		return nil, err
	}
	return file, err
}

// ByteCountByCategory sums the byte filesizes for all the files that match the category name.
func ByteCountByCategory(ctx context.Context, db *sql.DB, name string) (int64, error) {
	i, err := models.Files(
		qm.SQL(postgres.SQLSumSection(), null.StringFrom(name))).Count(ctx, db)
	if err != nil {
		return 0, fmt.Errorf("bytecount by section %q: %w", name, err)
	}
	return i, nil
}

// ByteCountByGroup sums the byte filesizes for all the files that match the group name.
func ByteCountByGroup(ctx context.Context, db *sql.DB, name string) (int64, error) {
	x := null.StringFrom(name)
	i, err := models.Files(qm.SQL(postgres.SQLSumGroup(), x)).Count(ctx, db)
	if err != nil {
		return 0, fmt.Errorf("bytecount by group %q: %w", name, err)
	}
	return i, nil
}

// ByteCountByPlatform sums the byte filesizes for all the files that match the category name.
func ByteCountByPlatform(ctx context.Context, db *sql.DB, name string) (int64, error) {
	i, err := models.Files(qm.SQL(postgres.SQLSumPlatform(), null.StringFrom(name))).Count(ctx, db)
	if err != nil {
		return 0, fmt.Errorf("bytecount by platform %q: %w", name, err)
	}
	return i, nil
}

// CountByCategory counts the files that match the named category.
func CountByCategory(ctx context.Context, db *sql.DB, name string) (int64, error) {
	x := null.StringFrom(name)
	return models.Files(models.FileWhere.Section.EQ(x)).Count(ctx, db)
}

// CountByPlatform counts the files that match the named category.
func CountByPlatform(ctx context.Context, db *sql.DB, name string) (int64, error) {
	x := null.StringFrom(name)
	return models.Files(models.FileWhere.Platform.EQ(x)).Count(ctx, db)
}

// SelectHTML3 selects only the columns required by the HTML3 template.
func SelectHTML3() qm.QueryMod {
	return qm.Select(
		models.FileColumns.ID,
		models.FileColumns.Filename,
		models.FileColumns.DateIssuedDay,
		models.FileColumns.DateIssuedMonth,
		models.FileColumns.DateIssuedYear,
		models.FileColumns.Createdat,
		models.FileColumns.Filesize,
		models.FileColumns.Platform,
		models.FileColumns.Section,
		models.FileColumns.GroupBrandFor,
		models.FileColumns.RecordTitle,
	)
}
