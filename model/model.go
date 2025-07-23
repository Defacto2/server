// Package model provides a database model for the Defacto2 website.
package model

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/Defacto2/server/handler/jsdos"
	"github.com/Defacto2/server/handler/jsdos/msdos"
	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/Defacto2/server/model/html3"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	"github.com/subpop/go-ini"
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
	ErrTx       = errors.New("transaction value is nil")
	ErrURI      = errors.New("uri value is invalid")
	ErrUUID     = errors.New("could not create a new universial unique identifier")
	ErrYear     = errors.New("invalid year")
)

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

const (
	DemozooSanity = 450000 // Sanity is to check the maximum permitted production ID.
)

func calc(o, l int) int {
	if o < 1 {
		o = 1
	}
	return (o - 1) * l
}

// JsDosCommand returns the program executable or commands to run in the js-dos emulator.
// If the dosee_run_program is set then it is the preferred executable.
// If the filename is a .com or .exe then it will return the filename.
// Otherwise, it will attempt to find the most likely executable in the archive.
func JsDosCommand(f *models.File) (string, error) {
	if f == nil {
		return "", ErrModel
	}
	if f.DoseeRunProgram.Valid && f.DoseeRunProgram.String != "" {
		return f.DoseeRunProgram.String, nil
	}
	return JsDosBinary(f)
}

// JsDosBinary returns the program executable to run in the js-dos emulator.
// If the filename is a .com or .exe then it will return the filename.
// Otherwise, it will attempt to find the most likely executable in the archive.
func JsDosBinary(f *models.File) (string, error) {
	if f == nil {
		return "", ErrModel
	}
	if !f.Filename.Valid || f.Filename.IsZero() || f.Filename.String == "" {
		return "", nil
	}
	name := strings.ToLower(f.Filename.String)
	switch filepath.Ext(name) {
	case ".com", ".exe", ".bat":
		break
	default:
		if !f.FileZipContent.Valid || f.FileZipContent.IsZero() || f.FileZipContent.String == "" {
			return "", nil
		}
	}
	const dosPathSeparator, winPathSeparator = "\\", "/"
	findname := jsdos.FindBinary(f.Filename.String, f.FileZipContent.String)
	if !strings.Contains(findname, dosPathSeparator) && !strings.Contains(findname, winPathSeparator) {
		return msdos.Truncate(findname), nil
	}
	dir := filepath.Dir(findname)
	// replace all windows path separators with dos path separators,
	// as often the FileZipContent paths use non-dos path separators
	// despite the zipfile being a DOS file.
	dir = strings.ReplaceAll(dir, winPathSeparator, dosPathSeparator)
	base := msdos.Truncate(filepath.Base(findname))
	return strings.Join([]string{dir, base}, dosPathSeparator), nil
}

// JsDosConfig creates a js-dos .ini configuration for the emulator.
func JsDosConfig(f *models.File) (string, error) {
	if f == nil {
		return "", ErrModel
	}
	j := jsdos.Jsdos{}
	cpu := f.DoseeHardwareCPU.String
	if f.DoseeHardwareCPU.Valid && cpu != "" {
		j.CPU(cpu)
	}
	hw := f.DoseeHardwareGraphic.String
	if f.DoseeHardwareGraphic.Valid && hw != "" {
		j.Machine(hw)
	}
	sfx := f.DoseeHardwareAudio.String
	if f.DoseeHardwareAudio.Valid && sfx != "" {
		j.Sound(sfx)
	}
	mem := f.DoseeNoEms.Int16
	if f.DoseeNoEms.Valid && mem == 1 {
		j.NoEMS(true)
	}
	mem = f.DoseeNoXMS.Int16
	if f.DoseeNoXMS.Valid && mem == 1 {
		j.NoXMS(true)
	}
	mem = f.DoseeNoUmb.Int16
	if f.DoseeNoUmb.Valid && mem == 1 {
		j.NoUMB(true)
	}
	b, err := ini.Marshal(j)
	if err != nil {
		return "", fmt.Errorf("ini.Marshal: %w", err)
	}
	return string(b), nil
}

// invalidExec returns true if the database context executor is invalid such as nil.
func invalidExec(exec boil.ContextExecutor) bool {
	return html3.InvalidExec(exec)
}

// UUID returns a slice of all the UUIDs in the database.
func UUID(ctx context.Context, exec boil.ContextExecutor) (models.FileSlice, error) {
	if invalidExec(exec) {
		return nil, ErrDB
	}
	return models.Files(qm.Select("uuid")).All(ctx, exec)
}
