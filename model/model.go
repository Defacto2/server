// Package model provides a database model for the Defacto2 website.
package model

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/Defacto2/server/internal/postgres"
	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/google/uuid"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

var (
	ErrCtx      = errors.New("echo context is nil")
	ErrDay      = errors.New("invalid day")
	ErrDB       = errors.New("database value is nil")
	ErrID       = errors.New("file download database id cannot be found")
	ErrKey      = errors.New("key value is zero or negative")
	ErrModel    = errors.New("error, no file model")
	ErrMonth    = errors.New("invalid month")
	ErrName     = errors.New("name value is empty")
	ErrOrderBy  = errors.New("order by value is invalid")
	ErrSection  = errors.New("section tag value is empty")
	ErrSize     = errors.New("size value is invalid")
	ErrRels     = errors.New("too many releasers, only two are allowed")
	ErrPlatform = errors.New("invalid platform")
	ErrSha384   = errors.New("sha384 value is invalid")
	ErrTime     = errors.New("time value is invalid")
	ErrURI      = errors.New("uri value is invalid")
	ErrUUID     = errors.New("could not create a new universial unique identifier")
	ErrYear     = errors.New("invalid year")
	ErrZap      = errors.New("zap logger instance is nil")
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

const (
	startID        = 1                                      // startID is the default, first ID value.
	uidPlaceholder = `ADB7C2BF-7221-467B-B813-3636FE4AE16B` // UID of the user who deleted the file.
)

// EpochYear is the epoch year for the website,
// ie. the year 0 of the MS-DOS era.
const EpochYear = 1980

// Maximum number of files to return per query.
const Maximum = 998

// From is the name of the table containing records of files.
const From = "files"

// ClauseOldDate orders the records by oldest date first.
const ClauseOldDate = "date_issued_year ASC NULLS LAST, " +
	"date_issued_month ASC NULLS LAST, " +
	"date_issued_day ASC NULLS LAST"

// ClauseNoSoftDel is the clause to exclude soft deleted records.
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
	return file, nil
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
	return file, nil
}

// Find retrieves a single file record from the database using the record key.
func Find(key int) (*models.File, error) {
	return record(false, key)
}

// EditFind retrieves a single file record from the database using the record key.
// This function will also return records that have been marked as deleted.
func EditFind(key int) (*models.File, error) {
	return record(true, key)
}

// Record retrieves a single file record from the database using the record key.
func record(deleted bool, key int) (*models.File, error) {
	// get record id, filename, uuid
	ctx := context.Background()
	db, err := postgres.ConnectDB()
	if err != nil {
		return nil, ErrDB
	}
	defer db.Close()
	art, err := One(ctx, db, deleted, key)
	if err != nil {
		return nil, fmt.Errorf("%w, %w: %d", ErrID, err, key)
	}
	return art, nil
}
