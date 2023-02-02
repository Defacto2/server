package model

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Defacto2/server/pkg/postgres"
	"github.com/Defacto2/server/pkg/postgres/models"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

// From is the name of the table containing records of files.
const From = "files"

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
		qm.SQL(postgres.SQLSumSection(), null.StringFrom(name))).Count(ctx, db)
	if err != nil {
		return 0, fmt.Errorf("bytecount by section %q: %w", name, err)
	}
	return i, nil
}

// ByteCountByGroup sums the byte filesizes for all the files that match the group name.
func ByteCountByGroup(name string, ctx context.Context, db *sql.DB) (int64, error) {
	x := null.StringFrom(name)
	i, err := models.Files(qm.SQL(postgres.SQLSumGroup(), x)).Count(ctx, db)
	if err != nil {
		return 0, fmt.Errorf("bytecount by group %q: %w", name, err)
	}
	return i, nil
}

// ByteCountByPlatform sums the byte filesizes for all the files that match the category name.
func ByteCountByPlatform(name string, ctx context.Context, db *sql.DB) (int64, error) {
	i, err := models.Files(qm.SQL(postgres.SQLSumPlatform(), null.StringFrom(name))).Count(ctx, db)
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
