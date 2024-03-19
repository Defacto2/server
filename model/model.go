// Package model provides a database model for the Defacto2 website.
package model

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	namer "github.com/Defacto2/releaser/name"
	"github.com/Defacto2/server/internal/postgres"
	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/google/uuid"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

var (
	ErrDB   = errors.New("database value is nil")
	ErrKey  = errors.New("key value is zero or negative")
	ErrName = errors.New("name value is empty")
	ErrURI  = errors.New("uri value is invalid")
)

type Pagination struct {
	BaseURL   string // BaseURL is the base URL for the pagination links.
	CurrPage  int    // CurrPage is the current page number.
	SumPages  int    // SumPages is the total number of pages.
	PrevPage  int    // PrevPage is the previous page number.
	NextPage  int    // NextPage is the next page number.
	TwoBelow  int    // TwoBelow is the page number two below the current page.
	TwoAfter  int    // TwoAfter is the page number two after the current page.
	RangeStep int    // RangeStep is the number of pages to skip in the pagination range.
}

// From is the name of the table containing records of files.
const From = "files"

// ClauseOldDate orders the records by oldest date first.
const ClauseOldDate = "date_issued_year ASC NULLS LAST, " +
	"date_issued_month ASC NULLS LAST, " +
	"date_issued_day ASC NULLS LAST"

const ClauseNoSoftDel = "deletedat IS NULL"

// Cache returns true if the statistics are considered to be valid.
func Cache(b, c int, t time.Time) bool {
	return b > 0 && c > 0 && t.Before(time.Now().Add(-time.Hour*1))
}

// One returns the record associated with the key ID.
func One(ctx context.Context, db *sql.DB, deleted bool, key int) (*models.File, error) {
	if db == nil {
		return nil, ErrDB
	}
	if key <= 0 {
		return nil, fmt.Errorf("key value %d: %w", key, ErrKey)
	}
	mods := models.FileWhere.ID.EQ(int64(key))
	var file *models.File
	var err error
	if deleted {
		file, err = models.Files(mods, qm.WithDeleted()).One(ctx, db)
	} else {
		file, err = models.Files(mods).One(ctx, db)
	}
	if err != nil {
		return nil, fmt.Errorf("one record %d: %w", key, err)
	}
	return file, err
}

// OneByUUID returns the record associated with the key UUID.
func OneByUUID(ctx context.Context, db *sql.DB, deleted bool, uid string) (*models.File, error) {
	if db == nil {
		return nil, ErrDB
	}
	val, err := uuid.Parse(uid)
	if err != nil {
		return nil, fmt.Errorf("uuid validation %s: %w", uid, err)
	}
	mods := models.FileWhere.UUID.EQ(null.NewString(val.String(), true))
	var file *models.File
	if deleted {
		file, err = models.Files(mods, qm.WithDeleted()).One(ctx, db)
	} else {
		file, err = models.Files(mods).One(ctx, db)
	}
	if err != nil {
		return nil, fmt.Errorf("one record %s: %w", uid, err)
	}
	return file, err
}

// ByteCountByCategory sums the byte file sizes for all the files that match the category name.
func ByteCountByCategory(ctx context.Context, db *sql.DB, name string) (int64, error) {
	if db == nil {
		return 0, ErrDB
	}
	if name == "" {
		return 0, ErrName
	}
	mods := qm.SQL(string(postgres.SumSection()), null.StringFrom(name))
	i, err := models.Files(mods).Count(ctx, db)
	if err != nil {
		return 0, fmt.Errorf("bytecount by category %q: %w", name, err)
	}
	return i, nil
}

// ByteCountByReleaser sums the byte file sizes for all the files that match the group name.
func ByteCountByReleaser(ctx context.Context, db *sql.DB, name string) (int64, error) {
	if db == nil {
		return 0, ErrDB
	}
	if name == "" {
		return 0, ErrName
	}
	s, err := namer.Humanize(namer.Path(name))
	if err != nil {
		return 0, err
	}
	n := strings.ToUpper(s)
	mods := qm.SQL(string(postgres.SumGroup()), null.StringFrom(n))
	i, err := models.Files(mods).Count(ctx, db)
	if err != nil {
		return 0, fmt.Errorf("bytecount by releaser %q: %w", name, err)
	}
	return i, nil
}

// ByteCountByPlatform sums the byte filesizes for all the files that match the category name.
func ByteCountByPlatform(ctx context.Context, db *sql.DB, name string) (int64, error) {
	if db == nil {
		return 0, ErrDB
	}
	if name == "" {
		return 0, ErrName
	}
	mods := qm.SQL(string(postgres.SumPlatform()), null.StringFrom(name))
	i, err := models.Files(mods).Count(ctx, db)
	if err != nil {
		return 0, fmt.Errorf("bytecount by platform %q: %w", name, err)
	}
	return i, nil
}

// CountByCategory counts the files that match the named category.
func CountByCategory(ctx context.Context, db *sql.DB, name string) (int64, error) {
	if db == nil {
		return 0, ErrDB
	}
	if name == "" {
		return 0, ErrName
	}
	mods := models.FileWhere.Section.EQ(null.StringFrom(name))
	i, err := models.Files(mods).Count(ctx, db)
	if err != nil {
		return 0, fmt.Errorf("count by category %q: %w", name, err)
	}
	return i, nil
}

// CountByPlatform counts the files that match the named category.
func CountByPlatform(ctx context.Context, db *sql.DB, name string) (int64, error) {
	if db == nil {
		return 0, ErrDB
	}
	if name == "" {
		return 0, ErrName
	}
	mods := models.FileWhere.Platform.EQ(null.StringFrom(name))
	i, err := models.Files(mods).Count(ctx, db)
	if err != nil {
		return 0, fmt.Errorf("count by platform %q: %w", name, err)
	}
	return i, nil
}
