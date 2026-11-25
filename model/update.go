//nolint:exhaustruct_v5
package model

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Defacto2/helper"
	"github.com/Defacto2/server/handler/demozoo"
	"github.com/Defacto2/server/handler/pouet"
	"github.com/Defacto2/server/handler/releaser"
	"github.com/Defacto2/server/internal/nils"
	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/Defacto2/server/internal/tags"
	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
)

// updateCols updates only the specified columns (plus updatedat) for f.ID in a single UPDATE query.
func updateCols(ctx context.Context, exec boil.ContextExecutor, f *models.File, cols ...string) error {
	if f.ID < startID {
		return ErrBadIDInt
	}
	wl := make([]string, 0, len(cols)+1)
	wl = append(wl, models.FileColumns.Updatedat)
	wl = append(wl, cols...)
	rowsAff, err := f.Update(ctx, exec, boil.Whitelist(wl...))
	if err != nil {
		return err
	}
	if rowsAff == 0 {
		return sql.ErrNoRows
	}
	return nil
}

const (
	auto        = "auto" // the auto value for the dosbox emulator
	emulateAuto = ""     // the dosbox emulator value to use for automatic configuration
)

// BoolFrom is a type for the bool columns that can be updated.
type BoolFrom int

const (
	EmulateUMB    BoolFrom = iota // DoseeNoUmb with value
	EmulateEMS                    // DoseeNoEms with value
	EmulateXMS                    // DoseeNoXMS with value
	EmulateBroken                 // DoseeIncompatible with value
	ReadmeDisable                 // RetrotxtNoReadme with value
)

// Update updates the column bool from value with val.
// The boolFrom columns are table columns that can either be null, empty, or have a smallint value.
func (col BoolFrom) Update(ctx context.Context, exec boil.ContextExecutor, key int64, val bool) error {
	const format = "bool from %v: %w"
	if err := nils.Check(ctx, exec); err != nil {
		return fmt.Errorf(format, "check", err)
	}

	const yes, no = int16(1), int16(0)
	i := yes
	if val {
		i = no
	}

	f := models.File{ID: key}
	var colName string
	switch col {
	case EmulateUMB:
		f.DoseeNoUmb = null.Int16From(i)
		colName = models.FileColumns.DoseeNoUmb
	case EmulateEMS:
		f.DoseeNoEms = null.Int16From(i)
		colName = models.FileColumns.DoseeNoEms
	case EmulateXMS:
		f.DoseeNoXMS = null.Int16From(i)
		colName = models.FileColumns.DoseeNoXMS
	case EmulateBroken:
		f.DoseeIncompatible = null.Int16From(i)
		colName = models.FileColumns.DoseeIncompatible
	case ReadmeDisable:
		f.RetrotxtNoReadme = null.Int16From(i)
		colName = models.FileColumns.RetrotxtNoReadme
	default:
		return fmt.Errorf(format, col, ErrMethod)
	}

	if err := updateCols(ctx, exec, &f, colName); err != nil {
		return fmt.Errorf(format, "update", err)
	}

	return nil
}

func UpdateEmulateRunProgram(ctx context.Context, exec boil.ContextExecutor, key int64, val string) error {
	const format = "update emulate run program %s: %w"
	if err := nils.Check(ctx, exec); err != nil {
		return fmt.Errorf(format, "check", err)
	}

	s := strings.TrimSpace(strings.ToUpper(val))
	f := models.File{
		ID:              key,
		DoseeRunProgram: null.StringFrom(s),
	}

	if err := updateCols(ctx, exec, &f, models.FileColumns.DoseeRunProgram); err != nil {
		return fmt.Errorf(format, s, err)
	}

	return nil
}

func UpdateEmulateMachine(ctx context.Context, exec boil.ContextExecutor, key int64, val string) error {
	const format = "update emulate machine %s: %w"
	if err := nils.Check(ctx, exec); err != nil {
		return fmt.Errorf(format, "check", err)
	}

	validate := strings.TrimSpace(strings.ToLower(val))
	switch validate {
	case "cga", "ega", "vga", "tandy", "nolfb", "et3000", "paradise", "et4000", "oldvbe":
		// success
	case auto:
		validate = emulateAuto
	default:
		return fmt.Errorf(format, val, ErrBadMachine)
	}

	f := models.File{
		ID:                   key,
		DoseeHardwareGraphic: null.StringFrom(validate),
	}

	if err := updateCols(ctx, exec, &f, models.FileColumns.DoseeHardwareGraphic); err != nil {
		return fmt.Errorf(format, validate, err)
	}

	return nil
}

func UpdateEmulateCPU(ctx context.Context, exec boil.ContextExecutor, key int64, val string) error {
	const format = "update emulate cpu %s: %w"
	if err := nils.Check(ctx, exec); err != nil {
		return fmt.Errorf(format, "check", err)
	}

	validate := strings.TrimSpace(strings.ToLower(val))
	switch validate {
	case "8086", "386", "486":
		// success
	case auto:
		validate = emulateAuto
	default:
		return fmt.Errorf(format, val, ErrBadCPU)
	}

	f := models.File{
		ID:               key,
		DoseeHardwareCPU: null.StringFrom(validate),
	}

	if err := updateCols(ctx, exec, &f, models.FileColumns.DoseeHardwareCPU); err != nil {
		return fmt.Errorf(format, validate, err)
	}

	return nil
}

func UpdateEmulateSfx(ctx context.Context, exec boil.ContextExecutor, key int64, val string) error {
	const format = "update emulate sfx %s: %w"
	if err := nils.Check(ctx, exec); err != nil {
		return fmt.Errorf(format, "check", err)
	}

	validate := strings.TrimSpace(strings.ToLower(val))
	switch validate {
	case "covox", "sb1", "sb16", "gus", "pcspeaker", "none":
		// success
	case auto:
		validate = emulateAuto
	default:
		return fmt.Errorf(format, val, ErrBadSFX)
	}

	f := models.File{
		ID:                 key,
		DoseeHardwareAudio: null.StringFrom(validate),
	}

	if err := updateCols(ctx, exec, &f, models.FileColumns.DoseeHardwareAudio); err != nil {
		return fmt.Errorf(format, validate, err)
	}

	return nil
}

// Int64From is a type for the int64 columns that can be updated.
type Int64From int

const (
	Demozoo Int64From = iota // WebIDPouet column with value
	Pouet                    // WebIDDemozoo column with value
)

// Update the column int64 from value with val.
// The int64From columns are table columns that can either be null, empty, or have an int64 value.
// The values for both demozoo and pouet are validated to be within a sane range
// and a zero value will set their column's to null.
func (col Int64From) Update(ctx context.Context, exec boil.ContextExecutor, key int64, val string) error {
	const format = "update int64 from %s: %w"
	if err := nils.Check(ctx, exec); err != nil {
		return fmt.Errorf(format, "check", err)
	}

	if strings.TrimSpace(val) == "" {
		val = "0"
	}
	n, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return fmt.Errorf(format, "parse int", err)
	}

	f := models.File{ID: key}
	var colName string
	var outOfRange bool
	switch {
	case n == 0 && col == Demozoo:
		// remove demozoo entry
		f.WebIDDemozoo = null.Int64FromPtr(nil)
		colName = models.FileColumns.WebIDDemozoo
	case n == 0 && col == Pouet:
		// remove pouet entry
		f.WebIDPouet = null.Int64FromPtr(nil)
		colName = models.FileColumns.WebIDPouet
	case col == Demozoo:
		// update demozoo entry
		outOfRange = n < 1 || n > demozoo.Sanity
		f.WebIDDemozoo = null.Int64From(n)
		colName = models.FileColumns.WebIDDemozoo
	case col == Pouet:
		// update pouet entry
		outOfRange = n < 1 || n > pouet.Sanity
		f.WebIDPouet = null.Int64From(n)
		colName = models.FileColumns.WebIDPouet
	default:
		return fmt.Errorf(format, "parse", ErrMethod)
	}
	if outOfRange {
		return fmt.Errorf(format, "out of range", ErrBadID)
	}

	if err := updateCols(ctx, exec, &f, colName); err != nil {
		return fmt.Errorf(format, "update", err)
	}

	return nil
}

// StringFrom is a type for the string columns that can be updated.
type StringFrom int

const (
	Colors16     StringFrom = iota // WebID16colors column value
	Comment                        // Comment column value
	CreatorAudio                   // CreditAudio column value
	CreatorIll                     // CreditIllustration column value
	CreatorProg                    // CreditProgram column with value
	CreatorText                    // CreditText column with value
	Filename                       // Filename column with value
	GitHub                         // WebIDGithub column with value
	Integrity                      // FileIntegrityStrong column with value
	Platform                       // Platform column value with value
	Magic                          // FileMagicType value with value
	Relations                      // ListRelations column with value
	Section                        // Section column with value
	Sites                          // ListLinks column with value
	Title                          // RecordTitle column with value
	VirusTotal                     // FileSecurityAlertURL with value
	YouTube                        // WebIDYoutube column with value
	ZipContent                     // FileZipContent column with value
)

// Update updates the StringFrom value with val.
// The StringFrom columns are table column cells that can either be null, empty, or have a string value.
func (sf StringFrom) Update(ctx context.Context, exec boil.ContextExecutor, key int64, val string) error {
	const format = "update string from %s: %w"
	if err := nils.Check(ctx, exec); err != nil {
		return fmt.Errorf(format, "check", err)
	}

	f := models.File{ID: key}
	colName, err := sf.cases(&f, val)
	if err != nil {
		return fmt.Errorf(format, "cases", err)
	}

	if err := updateCols(ctx, exec, &f, colName); err != nil {
		return fmt.Errorf(format, val, err)
	}

	return nil
}

type stringTarget struct {
	col   string
	field *null.String
}

func stringTargets(f *models.File) [ZipContent + 1]stringTarget {
	return [...]stringTarget{
		Colors16:     {col: models.FileColumns.WebID16colors, field: &f.WebID16colors},
		Comment:      {col: models.FileColumns.Comment, field: &f.Comment},
		CreatorAudio: {col: models.FileColumns.CreditAudio, field: &f.CreditAudio},
		CreatorIll:   {col: models.FileColumns.CreditIllustration, field: &f.CreditIllustration},
		CreatorProg:  {col: models.FileColumns.CreditProgram, field: &f.CreditProgram},
		CreatorText:  {col: models.FileColumns.CreditText, field: &f.CreditText},
		Filename:     {col: models.FileColumns.Filename, field: &f.Filename},
		GitHub:       {col: models.FileColumns.WebIDGithub, field: &f.WebIDGithub},
		Integrity:    {col: models.FileColumns.FileIntegrityStrong, field: &f.FileIntegrityStrong},
		Magic:        {col: models.FileColumns.FileMagicType, field: &f.FileMagicType},
		Platform:     {col: models.FileColumns.Platform, field: &f.Platform},
		Relations:    {col: models.FileColumns.ListRelations, field: &f.ListRelations},
		Section:      {col: models.FileColumns.Section, field: &f.Section},
		Sites:        {col: models.FileColumns.ListLinks, field: &f.ListLinks},
		Title:        {col: models.FileColumns.RecordTitle, field: &f.RecordTitle},
		VirusTotal:   {col: models.FileColumns.FileSecurityAlertURL, field: &f.FileSecurityAlertURL},
		YouTube:      {col: models.FileColumns.WebIDYoutube, field: &f.WebIDYoutube},
		ZipContent:   {col: models.FileColumns.FileZipContent, field: &f.FileZipContent},
	}
}

func (sf StringFrom) cases(f *models.File, val string) (string, error) {
	if err := nils.Check(f); err != nil {
		return "", fmt.Errorf("check: %w", err)
	}

	targets := stringTargets(f)
	if sf < 0 || int(sf) >= len(targets) || targets[sf].field == nil {
		return "", ErrMethod
	}

	// strings must be sanitized otherwise there is a possibility of invalid
	// characters being injected via FileZipContent, which will error the SQL update
	val = strings.ToValidUTF8(val, "?")
	*targets[sf].field = null.StringFrom(strings.TrimSpace(val))
	return targets[sf].col, nil
}

type Creators struct {
	ID    int64
	Text  string
	Ill   string
	Prog  string
	Audio string
}

// Update updates the text, illustration, program, and audio credit columns with the values provided.
func (c Creators) Update(ctx context.Context, exec boil.ContextExecutor) error {
	const format = "update creators %s: %w"
	if err := nils.Check(ctx, exec); err != nil {
		return fmt.Errorf(format, "check", err)
	}

	f := models.File{
		ID:                 c.ID,
		CreditText:         null.StringFrom(c.Text),
		CreditIllustration: null.StringFrom(c.Ill),
		CreditProgram:      null.StringFrom(c.Prog),
		CreditAudio:        null.StringFrom(c.Audio),
	}

	err := updateCols(ctx, exec, &f,
		models.FileColumns.CreditText,
		models.FileColumns.CreditIllustration,
		models.FileColumns.CreditProgram,
		models.FileColumns.CreditAudio,
	)
	if err != nil {
		return fmt.Errorf(format, "update", err)
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

// Update updates the youtube, 16colors, relations, sites, demozoo, and pouet columns with the values provided.
func (link Links) Update(ctx context.Context, exec boil.ContextExecutor) error {
	const format = "update links %s: %w"
	if err := nils.Check(ctx, exec); err != nil {
		return fmt.Errorf(format, "check", err)
	}

	f := models.File{
		ID:            link.ID,
		WebIDYoutube:  null.StringFrom(link.YouTube),
		WebID16colors: null.StringFrom(link.Colors16),
		WebIDGithub:   null.StringFrom(link.GitHub),
		ListRelations: null.StringFrom(link.Relations),
		ListLinks:     null.StringFrom(link.Sites),
		WebIDDemozoo:  null.Int64FromPtr(nil),
		WebIDPouet:    null.Int64FromPtr(nil),
	}
	if link.Demozoo > 0 {
		f.WebIDDemozoo = null.Int64From(link.Demozoo)
	}
	if link.Pouet > 0 {
		f.WebIDPouet = null.Int64From(link.Pouet)
	}

	err := updateCols(ctx, exec, &f,
		models.FileColumns.WebIDYoutube,
		models.FileColumns.WebID16colors,
		models.FileColumns.WebIDGithub,
		models.FileColumns.ListRelations,
		models.FileColumns.ListLinks,
		models.FileColumns.WebIDDemozoo,
		models.FileColumns.WebIDPouet,
	)
	if err != nil {
		return fmt.Errorf(format, "update", err)
	}

	return nil
}

type Classification struct {
	ID       int64
	Platform string
	Tag      string
}

// Update the classification of a file in the database.
// It takes an ID, platform, and tag as parameters and returns an error if any.
// Both platform and tag must be valid values.
func (cl Classification) Update(ctx context.Context, exec boil.ContextExecutor) error {
	const format = "update classification %s: %w"
	if err := nils.Check(ctx, exec); err != nil {
		return fmt.Errorf(format, "check", err)
	}

	p, t := tags.TagByURI(cl.Platform), tags.TagByURI(cl.Tag)
	if p == -1 {
		return fmt.Errorf(format, "platform, "+cl.Platform, tags.ErrTag)
	}
	if !tags.IsPlatform(cl.Platform) {
		return fmt.Errorf(format, "platform, "+cl.Platform, tags.ErrTag)
	}

	if t == -1 {
		return fmt.Errorf(format, cl.Tag, tags.ErrTag)
	}
	if !tags.IsTag(cl.Tag) {
		return fmt.Errorf(format, cl.Tag, tags.ErrTag)
	}

	f := models.File{
		ID:       cl.ID,
		Platform: null.StringFrom(p.String()),
		Section:  null.StringFrom(t.String()),
	}

	err := updateCols(ctx, exec, &f,
		models.FileColumns.Platform,
		models.FileColumns.Section,
	)
	if err != nil {
		return fmt.Errorf(format, "update", err)
	}

	return nil
}

// UpdateOnline updates the record to be online and public.
func UpdateOnline(ctx context.Context, exec boil.ContextExecutor, status bool, key int64) error {
	const format = "update online %s: %w"
	if err := nils.Check(ctx, exec); err != nil {
		return fmt.Errorf(format, " check", err)
	}

	f := models.File{ID: key}
	switch status {
	case true:
		f.Deletedat = null.TimeFromPtr(nil)
		f.Deletedby = null.String{String: "", Valid: false}
	case false:
		now := time.Now()
		f.Deletedat = null.TimeFromPtr(&now)
		f.Deletedby = null.StringFrom(strings.ToLower(uidPlaceholder))
	}

	err := updateCols(ctx, exec, &f,
		models.FileColumns.Deletedat,
		models.FileColumns.Deletedby,
	)
	if err != nil {
		return fmt.Errorf(format, "update", err)
	}

	return nil
}

// UpdateReleasers updates the releasers values with relFor and relBy.
// It always overwrites the existing releases for the artifact id.
// If relFor or relBy are left empty, they will delete the existing releasers.
//
// When providing single releasers, it must be placed in relFor.
// Using relBy, while leaving relFor blank will return an error.
//
// The table columns updated are GroupBrandFor and GroupBrandBy.
func UpdateReleasers(ctx context.Context, exec boil.ContextExecutor, key int64, relFor, relBy string) error {
	const format = "update releasers %s: %w"
	if err := nils.Check(ctx, exec); err != nil {
		return fmt.Errorf(format, "check", err)
	}

	if key < 1 {
		return fmt.Errorf(format, "key check", ErrBadIDInt)
	}

	bx := strings.TrimSpace(relFor)
	cx := releaser.Cell(bx)

	by := strings.TrimSpace(relBy)
	cy := releaser.Cell(by)

	f := models.File{ID: key}
	switch {
	case cx != "" && cy != "": // set both
		f.GroupBrandFor = null.StringFrom(cx)
		f.GroupBrandBy = null.StringFrom(cy)
	case cx != "" && cy == "": // set for, remove by
		f.GroupBrandFor = null.StringFrom(cx)
		f.GroupBrandBy = null.StringFrom("")
	case cx == "" && cy == "": // remove both
		f.GroupBrandFor = null.StringFrom("")
		f.GroupBrandBy = null.StringFrom("")
	default:
		return fmt.Errorf(format, "", ErrBadRel)
	}

	err := updateCols(ctx, exec, &f,
		models.FileColumns.GroupBrandFor,
		models.FileColumns.GroupBrandBy,
	)
	if err != nil {
		return fmt.Errorf(format, strings.ToLower("for, "+cx+"by, "+cy), err)
	}

	return nil
}

// UpdateYMD updates the date issued year, month and day columns with the int values provided.
//   - y int16 represents a year
//   - m int16 represents a numeric month
//   - d int16 represents a numeric day of the month
func UpdateYMD(ctx context.Context, exec boil.ContextExecutor, key int64, y, m, d null.Int16) error {
	const format = "update ymd %s: %w"
	if err := nils.Check(ctx, exec); err != nil {
		return fmt.Errorf(format, "check", err)
	}

	id := strconv.Itoa(int(key))
	if key <= 0 {
		return fmt.Errorf(format, id, ErrBadIDInt)
	}

	const formatd = "update ymd (%s=%d): %w"
	if !y.IsZero() && !helper.Year(int(y.Int16)) {
		return fmt.Errorf(formatd, "y", y.Int16, ErrBadDYear)
	}
	if !m.IsZero() && helper.ShortMonth(int(m.Int16)) == "" {
		return fmt.Errorf(formatd, "m", m.Int16, ErrBadDMonth)
	}
	if !d.IsZero() && !helper.Day(int(d.Int16)) {
		return fmt.Errorf(formatd, "d", d.Int16, ErrBadDDay)
	}

	f := models.File{
		ID:              key,
		DateIssuedYear:  y,
		DateIssuedMonth: m,
		DateIssuedDay:   d,
	}

	err := updateCols(ctx, exec, &f,
		models.FileColumns.DateIssuedYear,
		models.FileColumns.DateIssuedMonth,
		models.FileColumns.DateIssuedDay,
	)
	if err != nil {
		return fmt.Errorf(format, "update "+id, err)
	}

	return nil
}

// UpdateYMDS updates the date issued year, month and day columns with the string values provided.
//   - y is a year
//   - m is a numeric month, 1 - 12
//   - d is a numeric day of the month, 1 - 31
func UpdateYMDS(ctx context.Context, exec boil.ContextExecutor, key int64, y, m, d string) error {
	const format = "update date issued %s: %w"
	if err := nils.Check(ctx, exec); err != nil {
		return fmt.Errorf(format, "check", err)
	}

	year, month, day := ValidDateIssue(y, m, d)
	f := models.File{
		ID:              key,
		DateIssuedYear:  year,
		DateIssuedMonth: month,
		DateIssuedDay:   day,
	}

	err := updateCols(ctx, exec, &f,
		models.FileColumns.DateIssuedYear,
		models.FileColumns.DateIssuedMonth,
		models.FileColumns.DateIssuedDay,
	)
	if err != nil {
		const format = "update date issued (%s-%s-%s): %w"
		return fmt.Errorf(format, y, m, d, err)
	}

	return nil
}

// UpdateMagic updates the file magictype (magic number) column with the magic value provided.
func UpdateMagic(ctx context.Context, exec boil.ContextExecutor, key int64, magic string) error {
	const format = "update magic %s id %d: %w"
	if err := nils.Check(ctx, exec); err != nil {
		return fmt.Errorf("update magic %w", err)
	}

	if key <= 0 {
		return fmt.Errorf(format, "", key, ErrBadIDInt)
	}

	f := models.File{
		ID:            key,
		FileMagicType: null.StringFrom(magic),
	}

	if err := updateCols(ctx, exec, &f, models.FileColumns.FileMagicType); err != nil {
		return fmt.Errorf(format, "update", key, err)
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
	if err := nils.Check(ctx, exec); err != nil {
		return fmt.Errorf(format, "check", err)
	}

	id := strconv.Itoa(int(fu.ID))
	if fu.ID <= 0 {
		return fmt.Errorf(format, "id value "+id, ErrBadIDInt)
	}

	f := models.File{
		ID:               fu.ID,
		Filesize:         null.Int64From(fu.Filesize),
		FileLastModified: null.TimeFrom(fu.LastMod),
	}

	if _, err := Filename.cases(&f, fu.Filename); err != nil {
		return fmt.Errorf(format, "filename", err)
	}
	if _, err := Integrity.cases(&f, fu.Integrity); err != nil {
		return fmt.Errorf(format, "integrity", err)
	}
	if _, err := Magic.cases(&f, fu.MagicNumber); err != nil {
		return fmt.Errorf(format, "magic number", err)
	}
	if _, err := ZipContent.cases(&f, fu.Content); err != nil {
		return fmt.Errorf(format, "zip content", err)
	}

	err := updateCols(ctx, exec, &f,
		models.FileColumns.Filename,
		models.FileColumns.FileIntegrityStrong,
		models.FileColumns.FileMagicType,
		models.FileColumns.FileZipContent,
		models.FileColumns.Filesize,
		models.FileColumns.FileLastModified,
	)
	if err != nil {
		return fmt.Errorf(format, "update record "+id, err)
	}

	return nil
}
