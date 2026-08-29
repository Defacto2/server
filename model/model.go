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

// checked in Aug 26, test coverage was good at around 50%

var (
	ErrModel      = errors.New("model: file model is not provided")
	ErrMethod     = errors.New("model: method value is invalid")
	ErrName       = errors.New("model: name value is empty")
	ErrSearch     = errors.New("model: search term(s) is invalid")
	ErrSha384     = errors.New("model: sha384 value is invalid")
	ErrTime       = errors.New("model: time value is invalid")
	ErrURI        = errors.New("model: uri value is invalid")
	ErrUUID       = errors.New("model: could not create uuid")
	ErrBadDDay    = errors.New("model: invalid day")
	ErrBadDMonth  = errors.New("model: invalid month")
	ErrBadDYear   = errors.New("model: invalid year")
	ErrBadID      = errors.New("model: artifact key id cannot be found")
	ErrBadIDInt   = errors.New("model: artifact key id is zero or negative")
	ErrBadOS      = errors.New("model: operating system or platform tag is invalid")
	ErrBadTag     = errors.New("model: category or section tag is invalid")
	ErrBadFname   = errors.New("model: filename is missing")
	ErrBadRel     = errors.New("model: cannot update releasers, for must be set before by")
	ErrBadMag     = errors.New("model: magazine requires an issue number, volume, or a title")
	ErrBadYT      = errors.New("model: invalid youtube id")
	ErrBadCPU     = errors.New("model: emulate-cpu value must be one of auto, 8086, 386, 486")
	ErrBadMachine = errors.New("model: emulate-machine value must be one of auto, " +
		"cga, ega, vga, tandy, nolfb, et3000, paradise, et4000, oldvbe")
	ErrBadSFX = errors.New("model: emulate-sfx value must be one of " +
		"auto, covox, sb1, sb16, gus, pcspeaker, none")
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

func calc(page, limit int) (offset int) { //nolint:nonamedreturns
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
