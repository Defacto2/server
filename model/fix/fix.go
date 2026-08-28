// Package fix contains functions for repairing the database data.
package fix

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"strings"

	"github.com/Defacto2/helper"
	"github.com/Defacto2/server/internal/nils"
	"github.com/Defacto2/server/internal/postgres"
	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
)

var ErrRepair = errors.New("fix: invalid repair option")

// Repair a column or type of data within the database.
type Repair int

const (
	// None does nothing.
	None Repair = iota - 1
	// Artifacts repairs all the artifact data.
	Artifacts
	// Releaser focuses on the releaser data using the group_brand_by and group_brand_for columns.
	Releaser
)

func (r Repair) String() string {
	switch r {
	case None:
		return "skip"
	case Artifacts:
		return "on all artifacts"
	case Releaser:
		return "on the releasers"
	default:
		return "error, unknown"
	}
}

const (
	groupBrandBy = "group_brand_by"
	updateSet    = "UPDATE files SET "
)

// In the future we may want to add a Debug or TestRun func.

// Run the database repair based on the repair option.
//
// The exec boil context executor is required.
// The db sql DB pointer is optional for a VACUUM statement that cannot be used in a transaction block.
func (r Repair) Run(ctx context.Context, sl *slog.Logger, db *sql.DB, exec boil.ContextExecutor) error {
	const format = "repair database %s: %w"
	if err := nils.Check(ctx, sl, exec); err != nil {
		return fmt.Errorf(format, "check", err)
	}

	sl.Info("Database: Check records with invalid UUIDs",
		slog.String("task", "run a cleanup of the database"))
	// run the syntax checks before sanity checks
	if r < None || r > Releaser {
		return fmt.Errorf("%w: %d", ErrRepair, r)
	}
	if r == None {
		return nil
	}

	if err := invalidUUIDs(ctx, sl, exec); err != nil {
		return fmt.Errorf(format, "invalid uuids", err)
	}
	if err := coldfusionIDs(ctx, sl, exec); err != nil {
		return fmt.Errorf(format, "coldfusion ids", err)
	}
	switch r { //nolint:exhaustive
	case Artifacts:
		sl.Info("Database: Clean records of whitespace and null values")
		if err := contentWhiteSpace(exec); err != nil {
			return fmt.Errorf(format, "content white space", err)
		}
		if err := nullifyEmpty(exec); err != nil {
			return fmt.Errorf(format, "nullify empty", err)
		}
		if err := nullifyZero(exec); err != nil {
			return fmt.Errorf(format, "nullify zero", err)
		}
		if err := trimFwdSlash(exec); err != nil {
			return fmt.Errorf(format, "trim forward slash", err)
		}
		if err := trainers(ctx, sl, exec); err != nil {
			return fmt.Errorf(format, "trainers", err)
		}
		fallthrough
	case Releaser:
		if err := releasers(ctx, sl, exec); err != nil {
			return fmt.Errorf(format, "releasers", err)
		}
	}
	if db != nil {
		if err := optimize(db); err != nil {
			return fmt.Errorf(format, "optimize", err)
		}
	}
	if err := SyncFilesIDSeq(exec); err != nil {
		return fmt.Errorf(format, "", err)
	}

	return nil
}

// SyncFilesIDSeq will synchronize the files ID sequence with the current maximum ID.
//
// This will only work with the correct database account permissions.
func SyncFilesIDSeq(exec boil.ContextExecutor) error {
	const format = "fix synchronize id sequence %s: %w"
	if err := nils.Check(exec); err != nil {
		return fmt.Errorf(format, "check", err)
	}

	query := `SELECT MAX(id) FROM files;` +
		`SELECT nextVal('"files_id_seq"');` +
		`SELECT setval('"files_id_seq"', (SELECT MAX(id) FROM files)+1);`
	_, err := queries.Raw(query).Exec(exec)
	if err != nil {
		return fmt.Errorf(format, "execute", err)
	}

	return nil
}

// coldfusionIDs will fix the invalid [ColdFusion language syntax] UUIDs in the database
// and rename the file assets using the newid UUIDs.
// ColdFusion uses an invalid 35 character UUID, which is a 32 character UUID with 3 hyphens,
// while the standard UUID is 36 characters with 4 hyphens.
//
// A blank UUID is "00000000-0000-0000-0000-000000000000".
//
// A blank CFID is "00000000-0000-0000-0000000000000000".
//
// [ColdFusion language syntax]: https://cfdocs.org/createuuid
func coldfusionIDs(ctx context.Context, sl *slog.Logger, exec boil.ContextExecutor) error {
	const format = "database repair : coldfusion %s: %w"
	const key = "task"
	sl.Info("Database: Check for invalid UUIDs using the ColdFusion syntax")
	mods := qm.SQL("SELECT uuid FROM files WHERE length(uuid)=35")
	fs, err := models.Files(mods).All(ctx, exec)
	if err != nil {
		return fmt.Errorf(format, "models files", err)
	}
	i := len(fs)
	if i == 0 {
		return nil
	}
	sl.Info("Database: Found records using the retired ColdFusion UUID syntax", slog.Int("finds", i))
	for _, f := range fs {
		if !f.UUID.Valid {
			continue
		}
		// 35 character UUIDs in a 36 character fixed length string will have a tailing space.
		old := strings.TrimSpace(f.UUID.String)
		newid, err := helper.CfUUID(old)
		if err != nil {
			sl.Warn("Convert ID failure", slog.String(key, old), slog.Any("error", err))
			continue
		}
		file, err := models.Files(qm.Where("uuid = ?", old)).One(ctx, exec)
		if err != nil {
			sl.Warn("Failed to find a record using the uuid",
				slog.String("uuid", old), slog.Any("error", err))
			continue
		}
		file.UUID = null.StringFrom(newid)
		_, err = file.Update(ctx, exec, boil.Infer())
		if err != nil {
			sl.Warn("Could not update the record", slog.String("uuid", old), slog.Any("error", err))
			continue
		}
	}
	return nil
}

const Trainer = "gamehack"

func trainers(ctx context.Context, sl *slog.Logger, exec boil.ContextExecutor) error {
	const msg = "database repair : " + Trainer
	const format = msg + " %s: %w"
	sl.Info(msg, slog.String("task", "Check for trainers that are incorrectly categorized"))

	const size = 5
	mods := make([]qm.QueryMod, 0, size)
	mods = append(mods, qm.Select("id"))
	mods = append(mods, qm.Where("section != ?", Trainer))
	mods = append(mods, qm.Where("section != 'magazine'"))
	mods = append(mods, qm.Where("record_title ILIKE '%trainer%'"))
	mods = append(mods, qm.Where("platform = ? OR platform = ?", "dos", "windows"))
	fs, err := models.Files(mods...).All(ctx, exec)
	if err != nil {
		return fmt.Errorf(format, "models files select", err)
	}

	l := len(fs)
	if l == 0 {
		return nil
	}

	mods = mods[:0]
	for i, f := range fs {
		if i == 0 {
			mods = append(mods, qm.Where("id = ?", f.ID))
			continue
		}
		mods = append(mods, qm.Or("id = ?", f.ID))
	}
	rowsAff, err := models.Files(mods...).UpdateAll(ctx, exec, models.M{"section": Trainer})
	if err != nil {
		return fmt.Errorf(format, "models files update all", err)
	}
	sl.Info(msg, slog.Int64("records_fixed", rowsAff))

	return nil
}

type Fix string

type List map[string]Fix

// Fix bad imported names, such as those from Demozoo data imports.
// Each one of these fixes also need an echo.redirect in router.go.
const (
	acidfix   = "ACID PRODUCTIONS"
	icefix    = "INSANE CREATORS ENTERPRISE"
	pwafix    = "pirates with attitudes"
	trsifix   = "TRISTAR & RED SECTOR INC"
	xpressfix = "X-PRESSION DESIGN"
	damnfix   = "DAMN EXCELLENT ANSI DESIGN"
	ofgfix    = "ORIGINALLY FUNNY GUYS"
	dsifix    = "DARKSIDE INCORPORATED"
	rssfix    = "renaissance"
	coop0fix  = "PE, TRSI, TDT"
	coop1fix  = "COOP"
)

//nolint:gochecknoglobals
var replacers = List{
	"ACID":                          acidfix,
	"ANSI Creators in Demand":       acidfix,
	"ICE":                           icefix,
	"pirates with attitude":         pwafix,
	"TRISTAR AND RED SECTOR INC":    trsifix,
	"X-PRESSION":                    xpressfix,
	"DAMN EXCELLENT ANSI DESIGNERS": damnfix,
	"THE ORIGINAL FUNNY GUYS":       ofgfix,
	"ORIGINAL FUNNY GUYS":           ofgfix,
	"DARKSIDE INC":                  dsifix,
	"RSS":                           rssfix,
	"Public Enemy, " +
		"Tristar & Red Sector Inc, " +
		"The Dream Team": coop0fix,
	"The Dream Team, " +
		"Tristar & Red Sector Inc": coop1fix,
}

var replacements List

func init() {
	replacements = make(List, len(replacers))
	for old, fix := range replacers {
		replacements[strings.ToUpper(old)] = Fix(strings.ToUpper(string(fix)))
	}
}

// Replacements returns a copy of the replacers List.
func Replacements() List {
	return replacements
}

// releasers will repair the group_brand_by and group_brand_for releasers data.
func releasers(ctx context.Context, sl *slog.Logger, exec boil.ContextExecutor) error {
	const msg = "database repair : releasers"
	sl.Info(msg,
		slog.String("task", "Clean up the named releasers in group_brand_by and group_brand_for"))
	f, err := models.Files(
		qm.Where("group_brand_for = group_brand_by"),
		qm.WithDeleted(),
	).All(ctx, exec)
	if err != nil {
		return fmt.Errorf("update group_brand_for = group_brand_by: %w", err)
	}
	if len(f) > 0 {
		empty := null.NewString("", true)
		rowsAff, err := f.UpdateAll(ctx, exec, models.M{groupBrandBy: empty})
		if err != nil {
			return fmt.Errorf("update all to null group_brand_by: %w", err)
		}
		if rowsAff > 0 {
			sl.Info(msg,
				slog.String("task", "Update the group_brand_by values to be null"),
				slog.Int64("updated", rowsAff))
		}
	}
	for bad, fix := range replacements {
		f, err = models.Files(
			qm.Where("group_brand_for = ?", bad),
			qm.WithDeleted(),
		).All(ctx, exec)
		if err != nil {
			return fmt.Errorf("where group_brand_for is bad: %w", err)
		}
		if len(f) > 0 {
			rowsAff, err := f.UpdateAll(ctx, exec, models.M{"group_brand_for": fix})
			if err != nil {
				return fmt.Errorf("update all group_brand_for fix: %w", err)
			}
			if rowsAff > 0 {
				sl.Info(msg,
					slog.String("task", "Fix the group_brand_for column"),
					slog.Int64("updated", rowsAff))
			}
		}
		f, err = models.Files(
			qm.Where("group_brand_by = ?", bad),
			qm.WithDeleted(),
		).All(ctx, exec)
		if err != nil {
			return fmt.Errorf("where group_brand_by is bad: %w", err)
		}
		if len(f) > 0 {
			rowsAff, err := f.UpdateAll(ctx, exec, models.M{groupBrandBy: fix})
			if err != nil {
				return fmt.Errorf("update all to null group_brand_by fix: %w", err)
			}
			if rowsAff > 0 {
				sl.Info(msg,
					slog.String("task", "Fix the group_brand_by column"),
					slog.Int64("updated", rowsAff))
			}
		}
	}
	if err := moreReleases(ctx, exec); err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}
	return nil
}

func moreReleases(_ context.Context, exec boil.ContextExecutor) error {
	_, err := queries.Raw(postgres.SetUpper("group_brand_for")).Exec(exec)
	if err != nil {
		return fmt.Errorf("set upper group_brand_for: %w", err)
	}
	_, err = queries.Raw(postgres.SetUpper(groupBrandBy)).Exec(exec)
	if err != nil {
		return fmt.Errorf("set upper group_brand_by: %w", err)
	}
	_, err = queries.Raw(postgres.SetFilesize0()).Exec(exec)
	if err != nil {
		return fmt.Errorf("set filesize 0: %w", err)
	}
	if err := Magics(exec); err != nil {
		return fmt.Errorf("magics: %w", err)
	}
	return demozooTitles(exec)
}

// demozooTitles fixes the redundant titles from Demozoo data imports
// where the title matches the name of the group, for example:
//
//	"Awesome Cool BBS (1) for Awesome Cool BBS"
func demozooTitles(exec boil.ContextExecutor) error {
	// cleanup the XXX (?) titles
	// UPDATE files
	// SET record_title = NULL
	// WHERE record_title ILIKE group_brand_for || ' (%)';
	_, err := queries.Raw(`UPDATE files SET record_title = NULL ` +
		`WHERE record_title ILIKE group_brand_for || ' (%)';`).Exec(exec)
	if err != nil {
		return fmt.Errorf("set title redundant group for: %w", err)
	}
	_, err = queries.Raw(`UPDATE files SET record_title = NULL ` +
		`WHERE record_title ILIKE group_brand_by || ' (%)';`).Exec(exec)
	if err != nil {
		return fmt.Errorf("set title redundant group by: %w", err)
	}
	// cleanup the XXX == titles
	// UPDATE files
	// SET record_title = NULL
	// WHERE record_title ILIKE group_brand_for;
	_, err = queries.Raw(`UPDATE files SET record_title = NULL ` +
		`WHERE record_title ILIKE group_brand_for;`).Exec(exec)
	if err != nil {
		return fmt.Errorf("set title redundant title = group for: %w", err)
	}
	_, err = queries.Raw(`UPDATE files SET record_title = NULL ` +
		`WHERE record_title ILIKE group_brand_by;`).Exec(exec)
	if err != nil {
		return fmt.Errorf("set title redundant title = group by: %w", err)
	}
	// cleanup the The XXX (?) titles
	// UPDATE files
	// SET record_title = NULL
	// WHERE record_title ILIKE 'the ' || group_brand_for || ' (%)';
	_, err = queries.Raw(`UPDATE files SET record_title = NULL ` +
		`WHERE record_title ILIKE 'the ' || group_brand_for || ' (%)';`).Exec(exec)
	if err != nil {
		return fmt.Errorf("set title redundant the group for: %w", err)
	}
	_, err = queries.Raw(`UPDATE files SET record_title = NULL ` +
		`WHERE record_title ILIKE 'the ' || group_brand_by || ' (%)';`).Exec(exec)
	if err != nil {
		return fmt.Errorf("set title redundant the group by: %w", err)
	}
	return nil
}

// Magics will set invalid file_magic_type to NULL.
// Invalid file_magic_type values are those that start with "ERROR: " or contain a "/"
// such as a mime-type.
func Magics(exec boil.ContextExecutor) error {
	_, err := queries.Raw(`UPDATE files SET file_magic_type = NULL ` +
		`WHERE file_magic_type ILIKE ANY(ARRAY['ERROR: %', '%/%']);`).Exec(exec)
	if err != nil {
		return fmt.Errorf("set invalid file_magic_type to \"\": %w", err)
	}
	return nil
}

// contentWhiteSpace will remove any duplicate newline white space from file_zip_content.
func contentWhiteSpace(exec boil.ContextExecutor) error {
	_, err := queries.Raw("UPDATE files SET file_zip_content = " +
		"RTRIM(regexp_replace(file_zip_content, '\n+', '\n', 'g'), '\r');").Exec(exec)
	if err != nil {
		return fmt.Errorf("queries raw %w", err)
	}
	return nil
}

// Optimize reclaims storage occupied by dead tuples in the database and
// also analyzes the most efficient execution plans for queries.
//
// Optimize runs the VACUUM query which cannot be used in a transaction.
func optimize(db *sql.DB) error {
	if err := nils.Check(db); err != nil {
		return fmt.Errorf("fix optimize check: %w", err)
	}

	_, err := queries.Raw("VACUUM ANALYZE files").Exec(db)
	if err != nil {
		return fmt.Errorf("execute vacuum and analyze: %w", err)
	}

	return nil
}

// invalidUUIDs will count the number of invalid UUIDs in the database.
// This should be part of a future function to repair the UUIDs and rename the file assets.
func invalidUUIDs(ctx context.Context, sl *slog.Logger, exec boil.ContextExecutor) error {
	const msg = "Database repair"
	mods := qm.SQL("SELECT COUNT(*) FROM files WHERE files.uuid" +
		" !~ '^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[1-5][0-9a-fA-F]{3}-[89abAB][0-9a-fA-F]{3}-[0-9a-fA-F]{12}';")
	i, err := models.Files(mods).Count(ctx, exec)
	if err != nil {
		return fmt.Errorf("query count: %w", err)
	}
	if i == 0 {
		return nil
	}
	sl.Warn(msg,
		slog.String("task", "Invalid UUID(s) found"),
		slog.Int64("finds", i))
	return nil
}

func nullifyEmpty(exec boil.ContextExecutor) error {
	var query strings.Builder
	columns := []string{
		"list_relations", "web_id_github", "web_id_youtube",
		"group_brand_for", groupBrandBy, "record_title",
		"credit_text", "credit_program", "credit_illustration", "credit_audio", "comment",
		"dosee_hardware_cpu", "dosee_hardware_graphic", "dosee_hardware_audio",
	}
	for column := range slices.Values(columns) {
		fmt.Fprintf(&query, "%s%s = NULL WHERE %s = ''; ", updateSet, column, column)
	}
	if _, err := queries.Raw(query.String()).Exec(exec); err != nil {
		return fmt.Errorf("query execute: %w", err)
	}
	return nil
}

func nullifyZero(exec boil.ContextExecutor) error {
	var query strings.Builder
	columns := []string{
		"web_id_pouet", "web_id_demozoo",
		"date_issued_year", "date_issued_month", "date_issued_day",
	}
	for column := range slices.Values(columns) {
		fmt.Fprintf(&query, "%s%s = NULL WHERE %s = 0; ", updateSet, column, column)
	}
	if _, err := queries.Raw(query.String()).Exec(exec); err != nil {
		return fmt.Errorf("query execute: %w", err)
	}
	return nil
}

func trimFwdSlash(exec boil.ContextExecutor) error {
	var query strings.Builder
	columns := []string{"web_id_16colors"}
	for column := range slices.Values(columns) {
		s := updateSet + column + " = LTRIM(web_id_16colors, '/') WHERE web_id_16colors LIKE '/%'; "
		query.WriteString(s)
	}
	if _, err := queries.Raw(query.String()).Exec(exec); err != nil {
		return fmt.Errorf("query execute: %w", err)
	}
	return nil
}

// NumFile represents a file with a numeric suffix.
type NumFile struct {
	ID           int64  `boil:"id"                        json:"id"`
	UUID         string `boil:"uuid"                      json:"uuid"`
	Filename     string `boil:"suffix"                    json:"filename"`
	ObfuscatedID string `boil:"-"                         json:"obfuscatedId"`
}

// NumFiles represents files that have numeric suffixes that need fixing.
type NumFiles struct {
	Count int64     `boil:"-" json:"count"`
	Files []NumFile `boil:"-" json:"files"`
}

// NumSuffix retrieves files with numeric suffixes from the database.
func NumSuffix(ctx context.Context, exec boil.ContextExecutor) (*NumFiles, error) {
	const format = "files with numeric suffix %s: %w"
	none := &NumFiles{}
	if err := nils.Check(ctx, exec); err != nil {
		return none, fmt.Errorf(format, "check", err)
	}

	var files []NumFile
	if err := queries.Raw(postgres.NumSuffixes()).Bind(ctx, exec, &files); err != nil {
		return none, fmt.Errorf(format, "list", err)
	}
	count := int64(len(files))

	return &NumFiles{
		Count: count,
		Files: files,
	}, nil
}
