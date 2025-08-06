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
	"github.com/Defacto2/releaser"
	"github.com/Defacto2/server/handler/pouet"
	"github.com/Defacto2/server/internal/panics"
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
func UpdateEmulateUMB(db *sql.DB, id int64, val bool) error {
	return UpdateBoolFrom(db, emulateUMB, id, val)
}

// UpdateEmulateEMS updates the column dosee_no_ems with val.
func UpdateEmulateEMS(db *sql.DB, id int64, val bool) error {
	return UpdateBoolFrom(db, emulateEMS, id, val)
}

// UpdateEmulateXMS updates the column dosee_no_xms with val.
func UpdateEmulateXMS(db *sql.DB, id int64, val bool) error {
	return UpdateBoolFrom(db, emulateXMS, id, val)
}

// UpdateEmulateBroken updates the column dosee_broken with val.
func UpdateEmulateBroken(db *sql.DB, id int64, val bool) error {
	return UpdateBoolFrom(db, emulateBroken, id, val)
}

// UpdateReadmeDisable updates the column retrotxt_no_readme with val.
func UpdateReadmeDisable(db *sql.DB, id int64, val bool) error {
	return UpdateBoolFrom(db, readmeDisable, id, val)
}

// UpdateBoolFrom updates the column bool from value with val.
// The boolFrom columns are table columns that can either be null, empty, or have a smallint value.
func UpdateBoolFrom(db *sql.DB, column boolFrom, id int64, val bool) error {
	const msg = "update bool from"
	if db == nil {
		return fmt.Errorf("%s: %w", msg, panics.ErrNoDB)
	}
	ctx := context.Background()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}
	f, err := OneFile(ctx, tx, id)
	if err != nil {
		return fmt.Errorf("%s find file for %q: %w", msg, column, err)
	}
	const yes, no = int16(1), int16(0)
	i := yes
	if val {
		i = no
	}
	switch column {
	case emulateUMB:
		f.DoseeNoUmb = null.NewInt16(i, true)
	case emulateEMS:
		f.DoseeNoEms = null.NewInt16(i, true)
	case emulateXMS:
		f.DoseeNoXMS = null.NewInt16(i, true)
	case emulateBroken:
		f.DoseeIncompatible = null.NewInt16(i, true)
	case readmeDisable:
		f.RetrotxtNoReadme = null.NewInt16(i, true)
	}
	if _, err = f.Update(ctx, tx, boil.Infer()); err != nil {
		return fmt.Errorf("%s %q %v: %w", msg, column, val, err)
	}
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}
	return nil
}

func UpdateEmulateRunProgram(db *sql.DB, id int64, val string) error {
	const msg = "update emulate run program"
	if db == nil {
		return fmt.Errorf("%s: %w", msg, panics.ErrNoDB)
	}
	s := strings.TrimSpace(strings.ToUpper(val))
	ctx := context.Background()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}
	f, err := OneFile(ctx, tx, id)
	if err != nil {
		return fmt.Errorf("%s find file for: %w", msg, err)
	}
	f.DoseeRunProgram = null.StringFrom(s)
	if _, err = f.Update(ctx, tx, boil.Infer()); err != nil {
		return fmt.Errorf("%s %s: %w", msg, s, err)
	}
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}
	return nil
}

func UpdateEmulateMachine(db *sql.DB, id int64, val string) error {
	const msg = "update emulate machine"
	if db == nil {
		return fmt.Errorf("%s: %w", msg, panics.ErrNoDB)
	}
	validate := strings.TrimSpace(strings.ToLower(val))
	switch validate {
	case "cga", "ega", "vga", "tandy", "nolfb", "et3000", "paradise", "et4000", "oldvbe":
		// success
	case auto:
		validate = emulateAuto
	default:
		return fmt.Errorf("%s %s: %w", msg, val, ErrMachine)
	}
	ctx := context.Background()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}
	f, err := OneFile(ctx, tx, id)
	if err != nil {
		return fmt.Errorf("%s find file for: %w", msg, err)
	}
	f.DoseeHardwareGraphic = null.StringFrom(validate)
	if _, err = f.Update(ctx, tx, boil.Infer()); err != nil {
		return fmt.Errorf("%s %s: %w", msg, validate, err)
	}
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}
	return nil
}

func UpdateEmulateCPU(db *sql.DB, id int64, val string) error {
	const msg = "update emulate cpu"
	if db == nil {
		return fmt.Errorf("%s: %w", msg, panics.ErrNoDB)
	}
	validate := strings.TrimSpace(strings.ToLower(val))
	switch validate {
	case "8086", "386", "486":
		// success
	case auto:
		validate = emulateAuto
	default:
		return fmt.Errorf("%s %s: %w", msg, val, ErrCPU)
	}
	ctx := context.Background()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}
	f, err := OneFile(ctx, tx, id)
	if err != nil {
		return fmt.Errorf("%s find file for: %w", msg, err)
	}
	f.DoseeHardwareCPU = null.StringFrom(validate)
	if _, err = f.Update(ctx, tx, boil.Infer()); err != nil {
		return fmt.Errorf("%s %s: %w", msg, validate, err)
	}
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}
	return nil
}

func UpdateEmulateSfx(db *sql.DB, id int64, val string) error {
	const msg = "update emulate sfx"
	if db == nil {
		return fmt.Errorf("%s: %w", msg, panics.ErrNoDB)
	}
	validate := strings.TrimSpace(strings.ToLower(val))
	switch validate {
	case "covox", "sb1", "sb16", "gus", "pcspeaker", "none":
		// success
	case auto:
		validate = emulateAuto
	default:
		return fmt.Errorf("%s %s: %w", msg, val, ErrSfx)
	}
	ctx := context.Background()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}
	f, err := OneFile(ctx, tx, id)
	if err != nil {
		return fmt.Errorf("%s find file for: %w", msg, err)
	}
	f.DoseeHardwareAudio = null.StringFrom(validate)
	if _, err = f.Update(ctx, tx, boil.Infer()); err != nil {
		return fmt.Errorf("%s %s: %w", msg, validate, err)
	}
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("%s: %w", msg, err)
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
func Update16Colors(db *sql.DB, id int64, val string) error {
	return UpdateStringFrom(db, colors16, id, val)
}

// UpdateComment updates the Comment column value with val.
func UpdateComment(db *sql.DB, id int64, val string) error {
	return UpdateStringFrom(db, comment, id, val)
}

// UpdateCreatorAudio updates the CreditAudio column with val.
func UpdateCreatorAudio(db *sql.DB, id int64, val string) error {
	return UpdateStringFrom(db, credAudio, id, val)
}

// UpdateCreatorIll updates the CreditIllustration column with val.
func UpdateCreatorIll(db *sql.DB, id int64, val string) error {
	return UpdateStringFrom(db, credIll, id, val)
}

// UpdateCreatorProg updates the CreditProgram column with val.
func UpdateCreatorProg(db *sql.DB, id int64, val string) error {
	return UpdateStringFrom(db, credProg, id, val)
}

// UpdateCreatorText updates the CreditText column with val.
func UpdateCreatorText(db *sql.DB, id int64, val string) error {
	return UpdateStringFrom(db, creText, id, val)
}

// UpdateDemozoo updates the WebIDDemozoo column with val.
func UpdateDemozoo(db *sql.DB, id int64, val string) error {
	return UpdateInt64From(db, demozooProd, id, val)
}

// UpdateFilename updates the Filename column with val.
func UpdateFilename(db *sql.DB, id int64, val string) error {
	return UpdateStringFrom(db, filename, id, val)
}

// UpdateGitHub updates the WebIDGithub column with val.
func UpdateGitHub(db *sql.DB, id int64, val string) error {
	return UpdateStringFrom(db, github, id, val)
}

// UpdatePlatform updates the Platform column value with val.
func UpdatePlatform(db *sql.DB, id int64, val string) error {
	return UpdateStringFrom(db, platform, id, val)
}

// UpdatePouet updates the WebIDPouet column with val.
func UpdatePouet(db *sql.DB, id int64, val string) error {
	return UpdateInt64From(db, pouetProd, id, val)
}

// UpdateRelations updates the ListRelations column value with val.
func UpdateRelations(db *sql.DB, id int64, val string) error {
	return UpdateStringFrom(db, relations, id, val)
}

// UpdateSites updates the ListLinks column with val.
func UpdateSites(db *sql.DB, id int64, val string) error {
	return UpdateStringFrom(db, sites, id, val)
}

// UpdateTag updates the Section column with val.
func UpdateTag(db *sql.DB, id int64, val string) error {
	return UpdateStringFrom(db, section, id, val)
}

// UpdateTitle updates the RecordTitle column with val.
func UpdateTitle(db *sql.DB, id int64, val string) error {
	return UpdateStringFrom(db, title, id, val)
}

// UpdateVirusTotal updates the FileSecurityAlertURL value with val.
func UpdateVirusTotal(db *sql.DB, id int64, val string) error {
	return UpdateStringFrom(db, virusTotal, id, val)
}

// UpdateYouTube updates the WebIDYoutube column value with val.
func UpdateYouTube(db *sql.DB, id int64, val string) error {
	return UpdateStringFrom(db, youtube, id, val)
}

// UpdateInt64From updates the column int64 from value with val.
// The int64From columns are table columns that can either be null, empty, or have an int64 value.
// The demozooProd and pouetProd values are validated to be within a sane range
// and a zero value will set their column's to null.
func UpdateInt64From(db *sql.DB, column int64From, id int64, val string) error {
	const msg = "update int64 from"
	if db == nil {
		return fmt.Errorf("%s: %w", msg, panics.ErrNoDB)
	}
	ctx := context.Background()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}
	f, err := OneFile(ctx, tx, id)
	if err != nil {
		return fmt.Errorf("%s find file for %q: %w", msg, column, err)
	}
	if strings.TrimSpace(val) == "" {
		val = "0"
	}
	i64, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return fmt.Errorf("%s %s: %w", msg, val, err)
	}
	var invalid bool
	switch {
	case i64 == 0 && column == demozooProd:
		f.WebIDDemozoo = null.Int64FromPtr(nil)
	case i64 == 0 && column == pouetProd:
		f.WebIDPouet = null.Int64FromPtr(nil)
	case column == demozooProd:
		invalid = i64 < 1 || i64 > DemozooSanity
		f.WebIDDemozoo = null.Int64From(i64)
	case column == pouetProd:
		invalid = i64 < 1 || i64 > pouet.Sanity
		f.WebIDPouet = null.Int64From(i64)
	default:
		return fmt.Errorf("%s: %w", msg, ErrColumn)
	}
	if invalid {
		return fmt.Errorf("%s %d: %w", msg, i64, ErrID)
	}
	if _, err = f.Update(ctx, tx, boil.Infer()); err != nil {
		return fmt.Errorf("%s %q %s: %w", msg, column, val, err)
	}
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("%s: %w", msg, err)
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
func UpdateStringFrom(db *sql.DB, column stringFrom, id int64, val string) error {
	const msg = "update string from"
	if db == nil {
		return fmt.Errorf("%s: %w", msg, panics.ErrNoDB)
	}
	ctx := context.Background()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}
	f, err := OneFile(ctx, tx, id)
	if err != nil {
		return fmt.Errorf("%s find file for %q: %w", msg, column, err)
	}
	if err = updateStringCases(f, column, val); err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}
	if _, err = f.Update(ctx, tx, boil.Infer()); err != nil {
		return fmt.Errorf("%q %s: %w", column, val, err)
	}
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}
	return nil
}

func updateStringCases(f *models.File, column stringFrom, val string) error {
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

// UpdateCreators updates the text, illustration, program, and audio credit columns with the values provided.
func UpdateCreators(db *sql.DB, id int64, text, ill, prog, audio string) error {
	const msg = "update creators"
	if db == nil {
		return fmt.Errorf("%s: %w", msg, panics.ErrNoDB)
	}
	ctx := context.Background()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}
	f, err := OneFile(ctx, tx, id)
	if err != nil {
		return fmt.Errorf("%s find file, %d: %w", msg, id, err)
	}
	f.CreditText = null.StringFrom(text)
	f.CreditIllustration = null.StringFrom(ill)
	f.CreditProgram = null.StringFrom(prog)
	f.CreditAudio = null.StringFrom(audio)
	if _, err = f.Update(ctx, tx, boil.Infer()); err != nil {
		return fmt.Errorf("%s updater: %w", msg, err)
	}
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}
	return nil
}

// UpdateLinks updates the youtube, 16colors, relations, sites, demozoo, and pouet columns with the values provided.
func UpdateLinks(db *sql.DB, id int64,
	youtube, colors16, github, relations, sites string,
	demozoo, pouet int64,
) error {
	const msg = "update links"
	if db == nil {
		return fmt.Errorf("%s: %w", msg, panics.ErrNoDB)
	}
	ctx := context.Background()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}
	f, err := OneFile(ctx, tx, id)
	if err != nil {
		return fmt.Errorf("%s find file %d: %w", msg, id, err)
	}
	f.WebIDYoutube = null.StringFrom(youtube)
	f.WebID16colors = null.StringFrom(colors16)
	f.WebIDGithub = null.StringFrom(github)
	f.ListRelations = null.StringFrom(relations)
	f.ListLinks = null.StringFrom(sites)
	f.WebIDDemozoo = null.Int64From(demozoo)
	f.WebIDPouet = null.Int64From(pouet)
	if _, err = f.Update(ctx, tx, boil.Infer()); err != nil {
		return fmt.Errorf("%s updater: %w", msg, err)
	}
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}
	return nil
}

// UpdateClassification updates the classification of a file in the database.
// It takes an ID, platform, and tag as parameters and returns an error if any.
// Both platform and tag must be valid values.
func UpdateClassification(db *sql.DB, id int64, platform, tag string) error {
	const msg = "update classification"
	if db == nil {
		return fmt.Errorf("%s: %w", msg, panics.ErrNoDB)
	}
	p, t := tags.TagByURI(platform), tags.TagByURI(tag)
	if p == -1 {
		return fmt.Errorf("%s %s: %w", msg, platform, ErrPlatform)
	}
	if !tags.IsPlatform(platform) {
		return fmt.Errorf("%s %s: %w", msg, platform, ErrPlatform)
	}
	if t == -1 {
		return fmt.Errorf("%s %s: %w", msg, tag, tags.ErrTag)
	}
	if !tags.IsTag(tag) {
		return fmt.Errorf("%s %s: %w", msg, tag, tags.ErrTag)
	}
	ctx := context.Background()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}
	f, err := OneFile(ctx, tx, id)
	if err != nil {
		return fmt.Errorf("%s find file: %w", msg, err)
	}
	f.Platform = null.StringFrom(p.String())
	f.Section = null.StringFrom(t.String())
	if _, err = f.Update(ctx, tx, boil.Infer()); err != nil {
		return fmt.Errorf("%s update: %w", msg, err)
	}
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("%s tx commit: %w", msg, err)
	}
	return nil
}

// UpdateDateIssued updates the date issued year, month and day columns with the values provided.
// Columns updated are DateIssuedYear, DateIssuedMonth, and DateIssuedDay.
func UpdateDateIssued(db *sql.DB, id int64, y, m, d string) error {
	const msg = "update date issued"
	if db == nil {
		return fmt.Errorf("%s: %w", msg, panics.ErrNoDB)
	}
	ctx := context.Background()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}
	f, err := OneFile(ctx, tx, id)
	if err != nil {
		return fmt.Errorf("%s find file: %w", msg, err)
	}
	year, month, day := ValidDateIssue(y, m, d)
	f.DateIssuedYear = year
	f.DateIssuedMonth = month
	f.DateIssuedDay = day
	if _, err = f.Update(ctx, tx, boil.Infer()); err != nil {
		return fmt.Errorf("%s %q %q %q: %w", msg, y, m, d, err)
	}
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("%s tx.commit: %w", msg, err)
	}
	return nil
}

// UpdateOffline updates the record to be offline and inaccessible to the public.
func UpdateOffline(db *sql.DB, id int64) error {
	const msg = "update offline"
	if db == nil {
		return fmt.Errorf("%s: %w", msg, panics.ErrNoDB)
	}
	ctx := context.Background()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("%s offline: %w", msg, err)
	}
	f, err := OneFile(ctx, tx, id)
	if err != nil {
		return fmt.Errorf("%s find file: %w", msg, err)
	}
	now := time.Now()
	f.Deletedat = null.TimeFromPtr(&now)
	f.Deletedby = null.StringFrom(strings.ToLower(uidPlaceholder))
	if _, err = f.Update(ctx, tx, boil.Infer()); err != nil {
		return fmt.Errorf("%s update: %w", msg, err)
	}
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("%s tx commit: %w", msg, err)
	}
	return nil
}

// UpdateOnline updates the record to be online and public.
func UpdateOnline(db *sql.DB, id int64) error {
	const msg = "update online"
	if db == nil {
		return fmt.Errorf("%s: %w", msg, panics.ErrNoDB)
	}
	ctx := context.Background()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("%s online: %w", msg, err)
	}
	f, err := OneFile(ctx, tx, id)
	if err != nil {
		return fmt.Errorf("%s find file: %w", msg, err)
	}
	f.Deletedat = null.TimeFromPtr(nil)
	f.Deletedby = null.String{}
	if _, err = f.Update(ctx, tx, boil.Infer()); err != nil {
		return fmt.Errorf("%s update: %w", msg, err)
	}
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("%s tx commit: %w", msg, err)
	}
	return nil
}

// UpdateReleasers updates the releasers values with val.
// Two releases can be separated by a + (plus) character.
// The columns updated are GroupBrandFor and GroupBrandBy.
func UpdateReleasers(db *sql.DB, id int64, val string) error {
	const msg = "update releasers"
	if db == nil {
		return fmt.Errorf("%s: %w", msg, panics.ErrNoDB)
	}
	const maxReleasers = 2
	val = strings.TrimSpace(val)
	s := strings.Split(val, "+")
	if len(s) > maxReleasers {
		return fmt.Errorf("%s: %w", s, ErrRels)
	}
	for i, v := range s {
		s[i] = releaser.Cell(v)
	}
	ctx := context.Background()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}
	f, err := OneFile(ctx, tx, id)
	if err != nil {
		return fmt.Errorf("%s find file: %w", msg, err)
	}
	switch len(s) {
	case maxReleasers:
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
		return fmt.Errorf("%s %q: %w", msg, val, err)
	}
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("%s tx commit: %w", msg, err)
	}
	return nil
}

// UpdateYMD updates the date issued year, month and day columns with the values provided.
func UpdateYMD(ctx context.Context, exec boil.ContextExecutor, id int64, y, m, d null.Int16) error {
	const msg = "update ymd"
	if panics.BoilExec(exec) {
		return fmt.Errorf("%s: %w", msg, panics.ErrNoBoil)
	}
	if id <= 0 {
		return fmt.Errorf("%s id %d: %w", msg, id, ErrKey)
	}
	if !y.IsZero() && !helper.Year(int(y.Int16)) {
		return fmt.Errorf("%s %w: %d", msg, ErrYear, y.Int16)
	}
	if !m.IsZero() && helper.ShortMonth(int(m.Int16)) == "" {
		return fmt.Errorf("%s %w: %d", msg, ErrMonth, m.Int16)
	}
	if !d.IsZero() && !helper.Day(int(d.Int16)) {
		return fmt.Errorf("%s %w: %d", msg, ErrDay, d.Int16)
	}
	f, err := OneFile(ctx, exec, id)
	if err != nil {
		return fmt.Errorf("%s one file %w: %d", msg, err, id)
	}
	f.DateIssuedYear = y
	f.DateIssuedMonth = m
	f.DateIssuedDay = d
	if _, err = f.Update(ctx, exec, boil.Infer()); err != nil {
		return fmt.Errorf("%s update %w: %d", msg, err, id)
	}
	return nil
}

// UpdateMagic updates the file magictype (magic number) column with the magic value provided.
func UpdateMagic(ctx context.Context, exec boil.ContextExecutor, id int64, magic string) error {
	const msg = "update magic"
	if panics.BoilExec(exec) {
		return fmt.Errorf("%s: %w", msg, panics.ErrNoBoil)
	}
	if id <= 0 {
		return fmt.Errorf("%s id %d: %w", msg, id, ErrKey)
	}
	f, err := OneFile(ctx, exec, id)
	if err != nil {
		return fmt.Errorf("%s find file id %d: %w", msg, id, err)
	}
	f.FileMagicType = null.StringFrom(magic)
	if _, err = f.Update(ctx, exec, boil.Infer()); err != nil {
		return fmt.Errorf("%s update id %d: %w", msg, id, err)
	}
	return nil
}

// FileUpload is a struct that contains the values needed to update an existing file record
// after a new file has been uploaded to the server.
type FileUpload struct {
	LastMod     time.Time
	Filename    string
	Integrity   string
	MagicNumber string
	Content     string
	Filesize    int64
}

// Update the file record with the values provided in the FileUpload struct.
// The id is the database id key of the record.
func (fu FileUpload) Update(ctx context.Context, exec boil.ContextExecutor, id int64) error {
	const msg = "file upload update"
	if panics.BoilExec(exec) {
		return fmt.Errorf("%s: %w", msg, panics.ErrNoBoil)
	}
	if id <= 0 {
		return fmt.Errorf("%s id value %w: %d", msg, ErrKey, id)
	}
	f, err := OneFile(ctx, exec, id)
	if err != nil {
		return fmt.Errorf("%s one file %w: %d", msg, err, id)
	}
	if err = updateStringCases(f, filename, fu.Filename); err != nil {
		return fmt.Errorf("%s filename: %w", msg, err)
	}
	if err = updateStringCases(f, integrity, fu.Integrity); err != nil {
		return fmt.Errorf("%s integrity: %w", msg, err)
	}
	if err = updateStringCases(f, magic, fu.MagicNumber); err != nil {
		return fmt.Errorf("%s magic number: %w", msg, err)
	}
	if err = updateStringCases(f, zipContent, fu.Content); err != nil {
		return fmt.Errorf("%s zip content: %w", msg, err)
	}
	f.Filesize = null.Int64From(fu.Filesize)
	f.FileLastModified = null.TimeFrom(fu.LastMod)
	if _, err = f.Update(ctx, exec, boil.Infer()); err != nil {
		return fmt.Errorf("%s update record %d: %w", msg, id, err)
	}
	return nil
}
