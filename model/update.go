package model

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Defacto2/helper"
	"github.com/Defacto2/server/handler/pouet"
	"github.com/Defacto2/server/handler/releaser"
	"github.com/Defacto2/server/internal/nils"
	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/Defacto2/server/internal/tags"
	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
)

var (
	ErrCPU     = errors.New("emulate-cpu value must be one of auto, 8086, 386, 486")
	ErrMachine = errors.New("emulate-machine value must be one of auto, " +
		"cga, ega, vga, tandy, nolfb, et3000, paradise, et4000, oldvbe")
	ErrSfx = errors.New("emulate-sfx value must be one of auto, covox, sb1, sb16, gus, pcspeaker, none")
)

const (
	auto        = "auto" // the auto value for the dosbox emulator
	emulateAuto = ""     // the dosbox emulator value to use for automatic configuration
)

// boolFrom is a type for the bool columns that can be updated.
type boolFrom int

const (
	emulateUMB boolFrom = iota
	emulateEMS
	emulateXMS
	emulateBroken
	readmeDisable
)

// UpdateEmulateUMB updates the column dosee_no_umb with val.
func UpdateEmulateUMB(ctx context.Context, db *sql.DB, id int64, val bool) error {
	const format = "update emulate ums %s: %w"
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf(format, "begin tx", err)
	}
	return UpdateBoolFrom(ctx, db, tx, emulateUMB, id, val)
}

// UpdateEmulateEMS updates the column dosee_no_ems with val.
func UpdateEmulateEMS(ctx context.Context, db *sql.DB, id int64, val bool) error {
	const format = "update emulate ems %s: %w"
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf(format, "begin tx", err)
	}
	return UpdateBoolFrom(ctx, db, tx, emulateEMS, id, val)
}

// UpdateEmulateXMS updates the column dosee_no_xms with val.
func UpdateEmulateXMS(ctx context.Context, db *sql.DB, id int64, val bool) error {
	const format = "update emulate xms %s: %w"
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf(format, "begin tx", err)
	}
	return UpdateBoolFrom(ctx, db, tx, emulateXMS, id, val)
}

// UpdateEmulateBroken updates the column dosee_broken with val.
func UpdateEmulateBroken(ctx context.Context, db *sql.DB, id int64, val bool) error {
	const format = "update emulate broken %s: %w"
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf(format, "begin tx", err)
	}
	return UpdateBoolFrom(ctx, db, tx, emulateBroken, id, val)
}

// UpdateReadmeDisable updates the column retrotxt_no_readme with val.
func UpdateReadmeDisable(ctx context.Context, db *sql.DB, id int64, val bool) error {
	const format = "update readme disable %s: %w"
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf(format, "begin tx", err)
	}
	return UpdateBoolFrom(ctx, db, tx, readmeDisable, id, val)
}

// UpdateBoolFrom updates the column bool from value with val.
// The boolFrom columns are table columns that can either be null, empty, or have a smallint value.
func UpdateBoolFrom(ctx context.Context, db *sql.DB, tx *sql.Tx, column boolFrom, id int64, val bool) error {
	const format = "bool from %v: %w"
	if err := nils.Check(ctx, db, tx); err != nil {
		return fmt.Errorf(format, "check", err)
	}

	f, err := OneFile(ctx, tx, id)
	if err != nil {
		return fmt.Errorf(format, "one file", err)
	}

	const yes, no = int16(1), int16(0)
	i := yes
	if val {
		i = no
	}
	switch column {
	case emulateUMB:
		f.DoseeNoUmb = null.Int16From(i)
	case emulateEMS:
		f.DoseeNoEms = null.Int16From(i)
	case emulateXMS:
		f.DoseeNoXMS = null.Int16From(i)
	case emulateBroken:
		f.DoseeIncompatible = null.Int16From(i)
	case readmeDisable:
		f.RetrotxtNoReadme = null.Int16From(i)
	}

	if _, err = f.Update(ctx, tx, boil.Infer()); err != nil {
		return fmt.Errorf(format, column, err)
	}
	if err = tx.Commit(); err != nil {
		return fmt.Errorf(format, "tx commit", err)
	}

	return nil
}

func UpdateEmulateRunProgram(ctx context.Context, db *sql.DB, id int64, val string) error {
	const format = "update emulate run program %s: %w"
	if err := nils.Check(ctx, db); err != nil {
		return fmt.Errorf(format, "check", err)
	}

	s := strings.TrimSpace(strings.ToUpper(val))

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf(format, "begin tx", err)
	}
	f, err := OneFile(ctx, tx, id)
	if err != nil {
		return fmt.Errorf(format, "one file", err)
	}

	f.DoseeRunProgram = null.StringFrom(s)

	if _, err = f.Update(ctx, tx, boil.Infer()); err != nil {
		return fmt.Errorf(format, s, err)
	}
	if err = tx.Commit(); err != nil {
		return fmt.Errorf(format, "tx commit", err)
	}

	return nil
}

func UpdateEmulateMachine(ctx context.Context, db *sql.DB, id int64, val string) error {
	const format = "update emulate machine %s: %w"
	if err := nils.Check(ctx, db); err != nil {
		return fmt.Errorf(format, "check", err)
	}

	validate := strings.TrimSpace(strings.ToLower(val))
	switch validate {
	case "cga", "ega", "vga", "tandy", "nolfb", "et3000", "paradise", "et4000", "oldvbe":
		// success
	case auto:
		validate = emulateAuto
	default:
		return fmt.Errorf(format, val, ErrMachine)
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf(format, "begin tx", err)
	}
	f, err := OneFile(ctx, tx, id)
	if err != nil {
		return fmt.Errorf(format, "one file", err)
	}

	f.DoseeHardwareGraphic = null.StringFrom(validate)

	if _, err = f.Update(ctx, tx, boil.Infer()); err != nil {
		return fmt.Errorf(format, validate, err)
	}
	if err = tx.Commit(); err != nil {
		return fmt.Errorf(format, "tx commit", err)
	}

	return nil
}

func UpdateEmulateCPU(ctx context.Context, db *sql.DB, id int64, val string) error {
	const format = "update emulate cpu %s: %w"
	if err := nils.Check(ctx, db); err != nil {
		return fmt.Errorf(format, "check", err)
	}

	validate := strings.TrimSpace(strings.ToLower(val))
	switch validate {
	case "8086", "386", "486":
		// success
	case auto:
		validate = emulateAuto
	default:
		return fmt.Errorf(format, val, ErrCPU)
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf(format, "begin tx", err)
	}
	f, err := OneFile(ctx, tx, id)
	if err != nil {
		return fmt.Errorf(format, "one file", err)
	}

	f.DoseeHardwareCPU = null.StringFrom(validate)

	if _, err = f.Update(ctx, tx, boil.Infer()); err != nil {
		return fmt.Errorf(format, validate, err)
	}
	if err = tx.Commit(); err != nil {
		return fmt.Errorf(format, "tx commit", err)
	}

	return nil
}

func UpdateEmulateSfx(ctx context.Context, db *sql.DB, id int64, val string) error {
	const format = "update emulate sfx %s: %w"
	if err := nils.Check(ctx, db); err != nil {
		return fmt.Errorf(format, "check", err)
	}

	validate := strings.TrimSpace(strings.ToLower(val))
	switch validate {
	case "covox", "sb1", "sb16", "gus", "pcspeaker", "none":
		// success
	case auto:
		validate = emulateAuto
	default:
		return fmt.Errorf(format, val, ErrSfx)
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf(format, "begin tx", err)
	}
	f, err := OneFile(ctx, tx, id)
	if err != nil {
		return fmt.Errorf(format, "one file", err)
	}

	f.DoseeHardwareAudio = null.StringFrom(validate)

	if _, err = f.Update(ctx, tx, boil.Infer()); err != nil {
		return fmt.Errorf(format, validate, err)
	}
	if err = tx.Commit(); err != nil {
		return fmt.Errorf(format, "tx commit", err)
	}

	return nil
}

// int64From is a type for the int64 columns that can be updated.
type int64From int

const (
	demozooProd int64From = iota
	pouetProd
)

// Update16Colors updates the WebID16colors column value with val.
func Update16Colors(ctx context.Context, db *sql.DB, id int64, val string) error {
	const format = "update 16colors: %w"
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf(format, err)
	}

	return UpdateStringFrom(ctx, db, tx, colors16, id, val)
}

// UpdateComment updates the Comment column value with val.
func UpdateComment(ctx context.Context, db *sql.DB, id int64, val string) error {
	const format = "update comment: %w"
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf(format, err)
	}

	return UpdateStringFrom(ctx, db, tx, comment, id, val)
}

// UpdateCreatorAudio updates the CreditAudio column with val.
func UpdateCreatorAudio(ctx context.Context, db *sql.DB, id int64, val string) error {
	const format = "update creator audio: %w"
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf(format, err)
	}

	return UpdateStringFrom(ctx, db, tx, credAudio, id, val)
}

// UpdateCreatorIll updates the CreditIllustration column with val.
func UpdateCreatorIll(ctx context.Context, db *sql.DB, id int64, val string) error {
	const format = "update creator ill: %w"
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf(format, err)
	}

	return UpdateStringFrom(ctx, db, tx, credIll, id, val)
}

// UpdateCreatorProg updates the CreditProgram column with val.
func UpdateCreatorProg(ctx context.Context, db *sql.DB, id int64, val string) error {
	const format = "update creator prog: %w"
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf(format, err)
	}

	return UpdateStringFrom(ctx, db, tx, credProg, id, val)
}

// UpdateCreatorText updates the CreditText column with val.
func UpdateCreatorText(ctx context.Context, db *sql.DB, id int64, val string) error {
	const format = "update creator text: %w"
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf(format, err)
	}

	return UpdateStringFrom(ctx, db, tx, creText, id, val)
}

// UpdateDemozoo updates the WebIDDemozoo column with val.
func UpdateDemozoo(ctx context.Context, db *sql.DB, id int64, val string) error {
	const format = "update demozoo: %w"
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf(format, err)
	}

	return UpdateInt64From(ctx, db, tx, demozooProd, id, val)
}

// UpdateFilename updates the Filename column with val.
func UpdateFilename(ctx context.Context, db *sql.DB, id int64, val string) error {
	const format = "update filename: %w"
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf(format, err)
	}

	return UpdateStringFrom(ctx, db, tx, filename, id, val)
}

// UpdateGitHub updates the WebIDGithub column with val.
func UpdateGitHub(ctx context.Context, db *sql.DB, id int64, val string) error {
	const format = "update github: %w"
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf(format, err)
	}

	return UpdateStringFrom(ctx, db, tx, github, id, val)
}

// UpdatePlatform updates the Platform column value with val.
func UpdatePlatform(ctx context.Context, db *sql.DB, id int64, val string) error {
	const format = "update platform: %w"
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf(format, err)
	}

	return UpdateStringFrom(ctx, db, tx, platform, id, val)
}

// UpdatePouet updates the WebIDPouet column with val.
func UpdatePouet(ctx context.Context, db *sql.DB, id int64, val string) error {
	const format = "update pouet: %w"
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf(format, err)
	}

	return UpdateInt64From(ctx, db, tx, pouetProd, id, val)
}

// UpdateRelations updates the ListRelations column value with val.
func UpdateRelations(ctx context.Context, db *sql.DB, id int64, val string) error {
	const format = "update relations: %w"
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf(format, err)
	}

	return UpdateStringFrom(ctx, db, tx, relations, id, val)
}

// UpdateSites updates the ListLinks column with val.
func UpdateSites(ctx context.Context, db *sql.DB, id int64, val string) error {
	const format = "update sites: %w"
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf(format, err)
	}

	return UpdateStringFrom(ctx, db, tx, sites, id, val)
}

// UpdateTag updates the Section column with val.
func UpdateTag(ctx context.Context, db *sql.DB, id int64, val string) error {
	const format = "update tag: %w"
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf(format, err)
	}

	return UpdateStringFrom(ctx, db, tx, section, id, val)
}

// UpdateTitle updates the RecordTitle column with val.
func UpdateTitle(ctx context.Context, db *sql.DB, id int64, val string) error {
	const format = "update title: %w"
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf(format, err)
	}

	return UpdateStringFrom(ctx, db, tx, title, id, val)
}

// UpdateVirusTotal updates the FileSecurityAlertURL value with val.

func UpdateVirusTotal(ctx context.Context, db *sql.DB, id int64, val string) error {
	const format = "update virus total: %w"
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf(format, err)
	}

	return UpdateStringFrom(ctx, db, tx, virusTotal, id, val)
}

// UpdateYouTube updates the WebIDYoutube column value with val.
func UpdateYouTube(ctx context.Context, db *sql.DB, id int64, val string) error {
	const format = "update youtube: %w"
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf(format, err)
	}

	return UpdateStringFrom(ctx, db, tx, youtube, id, val)
}

// UpdateInt64From updates the column int64 from value with val.
// The int64From columns are table columns that can either be null, empty, or have an int64 value.
// The demozooProd and pouetProd values are validated to be within a sane range
// and a zero value will set their column's to null.
func UpdateInt64From(ctx context.Context, db *sql.DB, tx *sql.Tx, column int64From, id int64, val string) error {
	const format = "update int64 from: %w"
	const fmtVal = "update int64 value %v: %w"
	if err := nils.Check(ctx, db, tx); err != nil {
		return fmt.Errorf(format, err)
	}

	f, err := OneFile(ctx, tx, id)
	if err != nil {
		return fmt.Errorf(fmtVal, column, err)
	}

	if strings.TrimSpace(val) == "" {
		val = "0"
	}
	n, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return fmt.Errorf(fmtVal, val, err)
	}

	var outOfRange bool
	switch {
	case n == 0 && column == demozooProd:
		// remove demozoo entry
		f.WebIDDemozoo = null.Int64FromPtr(nil)
	case n == 0 && column == pouetProd:
		// remove pouet entry
		f.WebIDPouet = null.Int64FromPtr(nil)
	case column == demozooProd:
		// update demozoo entry
		outOfRange = n < 1 || n > DemozooSanity
		f.WebIDDemozoo = null.Int64From(n)
	case column == pouetProd:
		// update pouet entry
		outOfRange = n < 1 || n > pouet.Sanity
		f.WebIDPouet = null.Int64From(n)
	default:
		return fmt.Errorf(format, ErrColumn)
	}
	if outOfRange {
		return fmt.Errorf(fmtVal, n, ErrID)
	}

	if _, err = f.Update(ctx, tx, boil.Infer()); err != nil {
		return fmt.Errorf(fmtVal, n, err)
	}
	if err = tx.Commit(); err != nil {
		return fmt.Errorf(format, err)
	}

	return nil
}

// stringFrom is a type for the string columns that can be updated.
type stringFrom int

const (
	colors16 stringFrom = iota
	comment
	credAudio
	credIll
	credProg
	creText
	filename
	github
	integrity
	platform
	magic
	relations
	section
	sites
	title
	virusTotal
	youtube
	zipContent
)

// UpdateStringFrom updates the column string from value with val.
// The stringFrom columns are table columns that can either be null, empty, or have a string value.
func UpdateStringFrom(ctx context.Context, db *sql.DB, tx *sql.Tx, column stringFrom, id int64, val string) error {
	const format = "update string from: %w"
	if err := nils.Check(ctx, db, tx); err != nil {
		return fmt.Errorf(format, err)
	}

	const fmtVal = "update string value %v: %w"
	f, err := OneFile(ctx, tx, id)
	if err != nil {
		return fmt.Errorf(fmtVal, column, err)
	}

	if err = updateStringCases(f, column, val); err != nil {
		return fmt.Errorf(format, err)
	}

	if _, err = f.Update(ctx, tx, boil.Infer()); err != nil {
		return fmt.Errorf(fmtVal, val, err)
	}
	if err = tx.Commit(); err != nil {
		return fmt.Errorf(format, err)
	}

	return nil
}

func updateStringCases(f *models.File, column stringFrom, val string) error { //nolint:cyclop
	if err := nils.Check(f); err != nil {
		return fmt.Errorf("update string cases: %w", err)
	}

	// strings must be sanitized otherwise there is a possibility of invalid
	// characters being injected via FileZipContent, which will error the SQL update
	val = strings.ToValidUTF8(val, "?")
	s := null.StringFrom(strings.TrimSpace(val))
	switch column {
	case colors16:
		f.WebID16colors = s
	case comment:
		f.Comment = s
	case credAudio:
		f.CreditAudio = s
	case credIll:
		f.CreditIllustration = s
	case credProg:
		f.CreditProgram = s
	case creText:
		f.CreditText = s
	case filename:
		f.Filename = s
	case github:
		f.WebIDGithub = s
	case integrity:
		f.FileIntegrityStrong = s
	case magic:
		f.FileMagicType = s
	case platform:
		f.Platform = s
	case relations:
		f.ListRelations = s
	case section:
		f.Section = s
	case sites:
		f.ListLinks = s
	case title:
		f.RecordTitle = s
	case virusTotal:
		f.FileSecurityAlertURL = s
	case youtube:
		f.WebIDYoutube = s
	case zipContent:
		f.FileZipContent = s
	default:
		return ErrColumn
	}

	return nil
}

type Creators struct {
	ID    int64
	Text  string
	Ill   string
	Prog  string
	Audio string
}

// Update updates the text, illustration, program, and audio credit columns with the values provided.
func (c Creators) Update(ctx context.Context, db *sql.DB) error {
	const format = "update creators %s: %w"
	if err := nils.Check(ctx, db); err != nil {
		return fmt.Errorf(format, "check", err)
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf(format, "begin tx", err)
	}

	f, err := OneFile(ctx, tx, c.ID)
	if err != nil {
		return fmt.Errorf(format, "find file", err)
	}

	f.CreditText = null.StringFrom(c.Text)
	f.CreditIllustration = null.StringFrom(c.Ill)
	f.CreditProgram = null.StringFrom(c.Prog)
	f.CreditAudio = null.StringFrom(c.Audio)

	if _, err = f.Update(ctx, tx, boil.Infer()); err != nil {
		return fmt.Errorf(format, "update", err)
	}
	if err = tx.Commit(); err != nil {
		return fmt.Errorf(format, "tx commit", err)
	}

	return nil
}

type Links struct {
	ID        int64
	Demozoo   int64
	Pouet     int64
	YouTube   string
	Colors16  string
	GitHub    string
	Relations string
	Sites     string
}

// UpdateLinks updates the youtube, 16colors, relations, sites, demozoo, and pouet columns with the values provided.
func (l Links) UpdateLinks(ctx context.Context, db *sql.DB) error {
	const format = "update links %s: %w"
	if err := nils.Check(ctx, db); err != nil {
		return fmt.Errorf(format, "check", err)
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf(format, "begin tx", err)
	}

	f, err := OneFile(ctx, tx, l.ID)
	if err != nil {
		return fmt.Errorf(format, "find file", err)
	}

	f.WebIDYoutube = null.StringFrom(l.YouTube)
	f.WebID16colors = null.StringFrom(l.Colors16)
	f.WebIDGithub = null.StringFrom(l.GitHub)
	f.ListRelations = null.StringFrom(l.Relations)
	f.ListLinks = null.StringFrom(l.Sites)
	f.WebIDDemozoo = null.Int64From(l.Demozoo)
	f.WebIDPouet = null.Int64From(l.Pouet)

	if _, err = f.Update(ctx, tx, boil.Infer()); err != nil {
		return fmt.Errorf(format, "update", err)
	}
	if err = tx.Commit(); err != nil {
		return fmt.Errorf(format, "tx commit", err)
	}

	return nil
}

type Classification struct {
	ID       int64
	Platform string
	Tag      string
}

// UpdateClassification updates the classification of a file in the database.
// It takes an ID, platform, and tag as parameters and returns an error if any.
// Both platform and tag must be valid values.
func (cl Classification) Update(ctx context.Context, db *sql.DB) error {
	const format = "update classification %s: %w"
	if err := nils.Check(ctx, db); err != nil {
		return fmt.Errorf(format, "check", err)
	}

	p, t := tags.TagByURI(cl.Platform), tags.TagByURI(cl.Tag)
	if p == -1 {
		return fmt.Errorf(format, cl.Platform, ErrPlatform)
	}
	if !tags.IsPlatform(cl.Platform) {
		return fmt.Errorf(format, cl.Platform, ErrPlatform)
	}

	if t == -1 {
		return fmt.Errorf(format, cl.Tag, tags.ErrTag)
	}
	if !tags.IsTag(cl.Tag) {
		return fmt.Errorf(format, cl.Tag, tags.ErrTag)
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf(format, "begin tx", err)
	}
	f, err := OneFile(ctx, tx, cl.ID)
	if err != nil {
		return fmt.Errorf(format, "find file", err)
	}

	f.Platform = null.StringFrom(p.String())
	f.Section = null.StringFrom(t.String())

	if _, err = f.Update(ctx, tx, boil.Infer()); err != nil {
		return fmt.Errorf(format, "update", err)
	}
	if err = tx.Commit(); err != nil {
		return fmt.Errorf(format, "tx commit", err)
	}

	return nil
}

// UpdateDateIssued updates the date issued year, month and day columns with the values provided.
// Columns updated are DateIssuedYear, DateIssuedMonth, and DateIssuedDay.
func UpdateDateIssued(ctx context.Context, db *sql.DB, id int64, y, m, d string) error {
	const format = "update date issued %s: %w"
	if err := nils.Check(ctx, db); err != nil {
		return fmt.Errorf(format, "check", err)
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf(format, "begin tx", err)
	}

	f, err := OneFile(ctx, tx, id)
	if err != nil {
		return fmt.Errorf(format, "find file", err)
	}

	year, month, day := ValidDateIssue(y, m, d)
	f.DateIssuedYear = year
	f.DateIssuedMonth = month
	f.DateIssuedDay = day

	if _, err = f.Update(ctx, tx, boil.Infer()); err != nil {
		return fmt.Errorf(y+"-"+m+"-"+d+" "+format, err)
	}
	if err = tx.Commit(); err != nil {
		return fmt.Errorf(format, "tx commit", err)
	}

	return nil
}

// UpdateOffline updates the record to be offline and inaccessible to the public.
func UpdateOffline(ctx context.Context, db *sql.DB, id int64) error {
	const format = "update offline %s: %w"
	if err := nils.Check(ctx, db); err != nil {
		return fmt.Errorf(format, "check", err)
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf(format, "offline", err)
	}

	f, err := OneFile(ctx, tx, id)
	if err != nil {
		return fmt.Errorf(format, "find file", err)
	}

	now := time.Now()
	f.Deletedat = null.TimeFromPtr(&now)
	f.Deletedby = null.StringFrom(strings.ToLower(uidPlaceholder))

	if _, err = f.Update(ctx, tx, boil.Infer()); err != nil {
		return fmt.Errorf(format, "update", err)
	}
	if err = tx.Commit(); err != nil {
		return fmt.Errorf(format, "tx commit", err)
	}

	return nil
}

// UpdateOnline updates the record to be online and public.
func UpdateOnline(ctx context.Context, db *sql.DB, id int64) error {
	const format = "update online %s: %w"
	if err := nils.Check(ctx, db); err != nil {
		return fmt.Errorf(format, " check", err)
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf(format, "online", err)
	}

	f, err := OneFile(ctx, tx, id)
	if err != nil {
		return fmt.Errorf(format, "find file", err)
	}

	f.Deletedat = null.TimeFromPtr(nil)
	f.Deletedby = null.String{String: "", Valid: false}

	if _, err = f.Update(ctx, tx, boil.Infer()); err != nil {
		return fmt.Errorf(format, "update", err)
	}
	if err = tx.Commit(); err != nil {
		return fmt.Errorf(format, "tx commit", err)
	}

	return nil
}

// UpdateReleasers updates the releasers values with val.
// Two releases can be separated by a + (plus) character.
// The columns updated are GroupBrandFor and GroupBrandBy.
func UpdateReleasers(ctx context.Context, db *sql.DB, id int64, val string) error {
	const format = "update releasers %s: %w"
	if err := nils.Check(ctx, db); err != nil {
		return fmt.Errorf(format, "check", err)
	}

	val = strings.TrimSpace(val)
	s := strings.Split(val, "+")
	count := len(s)

	const bothRelrs = 2
	if count > bothRelrs {
		return fmt.Errorf("%s: %w", s, ErrRels)
	}
	for i, v := range s {
		s[i] = releaser.Cell(v)
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf(format, "begin tx", err)
	}
	f, err := OneFile(ctx, tx, id)
	if err != nil {
		return fmt.Errorf(format, "find file", err)
	}

	switch count {
	case bothRelrs:
		f.GroupBrandFor = null.StringFrom(s[0])
		f.GroupBrandBy = null.StringFrom(s[1])
	case 1:
		f.GroupBrandFor = null.StringFrom(s[0])
		f.GroupBrandBy = null.StringFrom("")
	case 0:
		f.GroupBrandFor = null.StringFrom("")
		f.GroupBrandBy = null.StringFrom("")
	}

	if _, err = f.Update(ctx, tx, boil.Infer()); err != nil {
		return fmt.Errorf(format, val, err)
	}
	if err = tx.Commit(); err != nil {
		return fmt.Errorf(format, "tx commit", err)
	}

	return nil
}

// UpdateYMD updates the date issued year, month and day columns with the values provided.
func UpdateYMD(ctx context.Context, exec boil.ContextExecutor, id int64, y, m, d null.Int16) error {
	const format = "update ymd %w: %d"
	const fmtid = "update ymd %s id %d: %w"
	if err := nils.Check(ctx, exec); err != nil {
		return fmt.Errorf(format, err, 0)
	}
	if id <= 0 {
		return fmt.Errorf(fmtid, "", id, ErrKey)
	}

	if !y.IsZero() && !helper.Year(int(y.Int16)) {
		return fmt.Errorf(format, ErrYear, y.Int16)
	}
	if !m.IsZero() && helper.ShortMonth(int(m.Int16)) == "" {
		return fmt.Errorf(format, ErrMonth, m.Int16)
	}
	if !d.IsZero() && !helper.Day(int(d.Int16)) {
		return fmt.Errorf(format, ErrDay, d.Int16)
	}

	f, err := OneFile(ctx, exec, id)
	if err != nil {
		return fmt.Errorf(fmtid, "one file", id, err)
	}

	f.DateIssuedYear = y
	f.DateIssuedMonth = m
	f.DateIssuedDay = d

	if _, err = f.Update(ctx, exec, boil.Infer()); err != nil {
		return fmt.Errorf(fmtid, "update", id, err)
	}

	return nil
}

// UpdateMagic updates the file magictype (magic number) column with the magic value provided.
func UpdateMagic(ctx context.Context, exec boil.ContextExecutor, id int64, magic string) error {
	const format = "update magic %s id %d: %w"
	if err := nils.Check(ctx, exec); err != nil {
		return fmt.Errorf("update magic %w", err)
	}

	if id <= 0 {
		return fmt.Errorf(format, "", id, ErrKey)
	}

	f, err := OneFile(ctx, exec, id)
	if err != nil {
		return fmt.Errorf(format, "find file", id, err)
	}

	f.FileMagicType = null.StringFrom(magic)

	if _, err = f.Update(ctx, exec, boil.Infer()); err != nil {
		return fmt.Errorf(format, "update", id, err)
	}

	return nil
}

// FileUpload contains the values needed to update an existing file record
// after a new file has been uploaded to the server.
type FileUpload struct {
	ID          int64
	LastMod     time.Time
	Filename    string
	Integrity   string
	MagicNumber string
	Content     string
	Filesize    int64
}

// Update the file record with the values provided in [FileUpload].
// The id is the database id key of the record.
func (fu FileUpload) Update(ctx context.Context, exec boil.ContextExecutor) error {
	const format = "file upload update %s: %w"
	const fmtID = format + ": %d"
	if err := nils.Check(ctx, exec); err != nil {
		return fmt.Errorf(format, "check", err)
	}

	if fu.ID <= 0 {
		return fmt.Errorf(fmtID, "id value", ErrKey, fu.ID)
	}

	f, err := OneFile(ctx, exec, fu.ID)
	if err != nil {
		return fmt.Errorf(fmtID, "one file", err, fu.ID)
	}

	if err = updateStringCases(f, filename, fu.Filename); err != nil {
		return fmt.Errorf(format, "filename", err)
	}
	if err = updateStringCases(f, integrity, fu.Integrity); err != nil {
		return fmt.Errorf(format, "integrity", err)
	}
	if err = updateStringCases(f, magic, fu.MagicNumber); err != nil {
		return fmt.Errorf(format, "magic number", err)
	}
	if err = updateStringCases(f, zipContent, fu.Content); err != nil {
		return fmt.Errorf(format, "zip content", err)
	}

	f.Filesize = null.Int64From(fu.Filesize)
	f.FileLastModified = null.TimeFrom(fu.LastMod)

	if _, err = f.Update(ctx, exec, boil.Infer()); err != nil {
		return fmt.Errorf(fmtID, "infer update record", err, fu.ID)
	}

	return nil
}
