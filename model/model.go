// Package model provides a database model for the Defacto2 website.
package model

import (
	"context"
	"errors"
	"fmt"

	"github.com/Defacto2/server/internal/nils"
	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
)

var (
	ErrColumn   = errors.New("column not implemented")
	ErrDay      = errors.New("invalid day")
	ErrDB       = errors.New("database value is nil")
	ErrID       = errors.New("file download database id cannot be found")
	ErrKey      = errors.New("key value is zero or negative")
	ErrModel    = errors.New("error, no file model")
	ErrMonth    = errors.New("invalid month")
	ErrName     = errors.New("name value is empty")
	ErrOrderBy  = errors.New("order by value is invalid")
	ErrSize     = errors.New("size value is invalid")
	ErrRels     = errors.New("too many releasers, only two are allowed")
	ErrPlatform = errors.New("invalid platform")
	ErrSha384   = errors.New("sha384 value is invalid")
	ErrTime     = errors.New("time value is invalid")
	ErrURI      = errors.New("uri value is invalid")
	ErrUUID     = errors.New("could not create a new universal unique identifier")
	ErrYear     = errors.New("invalid year")
	ErrYouTube  = errors.New("invalid youtube id")
)

const (
	startID        = 1                                      // startID is the default, first ID value.
	uidPlaceholder = `ADB7C2BF-7221-467B-B813-3636FE4AE16B` // UID of the user who deleted the file.
)

// EpochYear is the epoch year for the website,
// for example, using it as year zero for the era of MS-DOS.
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

const (
	DemozooSanity = 450000 // Sanity is to check the maximum permitted production ID.
)

func calc(page, limit int) (offset int) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 0 // fallback to use no limit
	}

	offset = (page - 1) * limit
	if offset < 0 {
		return 0
	}
	return offset
}

// UUID returns a slice of all the UUIDs in the database.
func UUID(ctx context.Context, exec boil.ContextExecutor) (models.FileSlice, error) {
	const format = "model uuid: %w"

	if err := nils.Check(ctx, exec); err != nil {
		return models.FileSlice{}, fmt.Errorf(format, err)
	}

	return models.Files(qm.Select("uuid")).All(ctx, exec)
}
