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
	"github.com/Defacto2/server/internal/panics"
	"github.com/Defacto2/server/internal/postgres"
	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
)

var ErrRepair = errors.New("invalid repair option")

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
	UpdateSet = "UPDATE files SET "
	msg       = "database repair"
)

// In the future we may want to add a Debug or TestRun func.

// Run the database repair based on the repair option.
func (r Repair) Run(ctx context.Context, db *sql.DB, tx *sql.Tx, sl *slog.Logger) error {
	const msg = "repair database runner"
	if err := panics.ContextDTS(ctx, db, tx, sl); err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}
	sl.Info(msg,
		slog.String("startup", "check for records with invalid uuid values"),
		slog.String("task", "run a cleanup of the database"))
	if r < None || r > Releaser {
		return fmt.Errorf("%w: %d", ErrRepair, r)
	}
	if r == None {
		return nil
	}
	if err := invalidUUIDs(ctx, db, sl); err != nil {
		return fmt.Errorf("%s invalid uuids: %w", msg, err)
	}
	if err := coldfusionIDs(ctx, db, sl); err != nil {
		return fmt.Errorf("%s coldfusion ids: %w", msg, err)
	}
	switch r {
	case Artifacts:
		sl.Info(msg, slog.String("task", "clean the artifacts whitespace and null values"))
		if err := contentWhiteSpace(tx); err != nil {
			return fmt.Errorf("%s content white space: %w", msg, err)
		}
		if err := nullifyEmpty(tx); err != nil {
			return fmt.Errorf("%s nullify empty: %w", msg, err)
		}
		if err := nullifyZero(tx); err != nil {
			return fmt.Errorf("%s nullify zero: %w", msg, err)
		}
		if err := trimFwdSlash(tx); err != nil {
			return fmt.Errorf("%s trim forward slash: %w", msg, err)
		}
		if err := trainers(ctx, tx, sl); err != nil {
			return fmt.Errorf("%s trainers: %w", msg, err)
		}
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("%s transaction commit %w", msg, err)
		}
		fallthrough
	case Releaser:
		if err := releasers(ctx, db, sl); err != nil {
			return fmt.Errorf("%s releasers: %w", msg, err)
		}
	}
	if err := optimize(db); err != nil {
		return fmt.Errorf("%s optimize: %w", msg, err)
	}
	if err := SyncFilesIDSeq(db); err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}
	return nil
}

// SyncFilesIDSeq will synchronize the files ID sequence with the current maximum ID.
//
// This will only work with the correct database account permissions.
func SyncFilesIDSeq(db *sql.DB) error {
	const msg = "fix synchronize id sequence"
	if db == nil {
		return fmt.Errorf("%s: %w", msg, panics.ErrNoDB)
	}
	query := `SELECT MAX(id) FROM files;` +
		`SELECT nextVal('"files_id_seq"');` +
		`SELECT setval('"files_id_seq"', (SELECT MAX(id) FROM files)+1);`
	_, err := queries.Raw(query).Exec(db)
	if err != nil {
		return fmt.Errorf("%s execute: %w", msg, err)
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
func coldfusionIDs(ctx context.Context, exec boil.ContextExecutor, sl *slog.Logger) error {
	const msg = "coldfusion id fixes"
	sl.Info(msg, slog.String("task", "check for invalid UUIDs using the ColdFusion syntax"))
	mods := qm.SQL("SELECT uuid FROM files WHERE length(uuid)=35")
	fs, err := models.Files(mods).All(ctx, exec)
	if err != nil {
		return fmt.Errorf("%s models files: %w", msg, err)
	}
	i := len(fs)
	if i == 0 {
		return nil
	}
	sl.Info(msg,
		slog.String("task", "found records using the retired ColdFusion UUID syntax"),
		slog.Int("finds", i))
	for _, f := range fs {
		if !f.UUID.Valid {
			continue
		}
		// 35 character UUIDs in a 36 character fixed length string will have a tailing space.
		old := strings.TrimSpace(f.UUID.String)
		newid, err := helper.CfUUID(old)
		if err != nil {
			sl.Warn(msg, slog.String("invalid id syntax", old), slog.Any("error", err))
			continue
		}
		file, err := models.Files(qm.Where("uuid = ?", old)).One(ctx, exec)
		if err != nil {
			sl.Warn(msg, slog.String("database", "failed to find a record using the uuid"),
				slog.String("uuid", old), slog.Any("error", err))
			continue
		}
		file.UUID = null.StringFrom(newid)
		_, err = file.Update(ctx, exec, boil.Infer())
		if err != nil {
			sl.Warn(msg, "database update", "could not update the record",
				slog.String("uuid", old), slog.Any("error", err))
			continue
		}
	}
	return nil
}

func trainers(ctx context.Context, tx *sql.Tx, sl *slog.Logger) error {
	const msg = "trainers not using gamehack fix"
	const trainer = "gamehack"
	sl.Info(msg,
		slog.String("task", "check trainers that are not correctly categorized"))
	mods := []qm.QueryMod{}
	mods = append(mods, qm.Select("id"))
	mods = append(mods, qm.Where(fmt.Sprintf("section != '%s'", trainer)))
	mods = append(mods, qm.Where("section != 'magazine'"))
	mods = append(mods, qm.Where("record_title ILIKE '%trainer%'"))
	mods = append(mods, qm.Where("platform = 'dos' OR platform = 'windows'"))
	fs, err := models.Files(mods...).All(ctx, tx)
	if err != nil {
		return fmt.Errorf("%s models files select: %w", msg, err)
	}
	l := len(fs)
	if l == 0 {
		return nil
	}
	mods = []qm.QueryMod{}
	for i, f := range fs {
		if i == 0 {
			mods = append(mods, qm.Where("id = ?", f.ID))
			continue
		}
		if i < l {
			mods = append(mods, qm.Or("id = ?", f.ID))
		}
	}
	rowsAff, err := models.Files(mods...).UpdateAll(ctx, tx, models.M{"section": trainer})
	if err != nil {
		return fmt.Errorf("%s models files update all: %w", msg, err)
	}
	sl.Info(msg, slog.Int64("records fixed", rowsAff))
	return nil
}

// Fix bad imported names, such as those from Demozoo data imports.
// Each one of these fixes also need an echo.redirect in router.go.
const (
	acidbad   = "ACID"
	ansibad   = "ANSI Creators in Demand"
	acidfix   = "ACID PRODUCTIONS"
	icebad    = "ICE"
	icefix    = "INSANE CREATORS ENTERPRISE"
	pwabad    = "pirates with attitude"
	pwafix    = "pirates with attitudes"
	trsibad   = "TRISTAR AND RED SECTOR INC"
	trsifix   = "TRISTAR & RED SECTOR INC"
	xpress    = "X-PRESSION"
	xpressfix = "X-PRESSION DESIGN"
	damn      = "DAMN EXCELLENT ANSI DESIGNERS"
	damnfix   = "DAMN EXCELLENT ANSI DESIGN"
	ofg       = "THE ORIGINAL FUNNY GUYS"
	ofg1      = "ORIGINAL FUNNY GUYS"
	ofgfix    = "ORIGINALLY FUNNY GUYS"
	dsi       = "DARKSIDE INC"
	dsifix    = "DARKSIDE INCORPORATED"
	rss       = "RSS"
	rssfix    = "renaissance"
	coop0     = "Public Enemy, Tristar & Red Sector Inc, The Dream Team"
	coop0fix  = "PE, TRSI, TDT"
	coop1     = "The Dream Team, Tristar & Red Sector Inc"
	coop1fix  = "COOP"
)

func fixes() map[string]string {
	return map[string]string{
		acidbad: acidfix,
		ansibad: acidfix,
		icebad:  icefix,
		pwabad:  pwafix,
		trsibad: trsifix,
		xpress:  xpressfix,
		damn:    damnfix,
		ofg:     ofgfix,
		ofg1:    ofgfix,
		dsi:     dsifix,
		rss:     rssfix,
		coop0:   coop0fix,
		coop1:   coop1fix,
	}
}

// releasers will repair the group_brand_by and group_brand_for releasers data.
func releasers(ctx context.Context, exec boil.ContextExecutor, sl *slog.Logger) error {
	const msg = "releaser database fixer"
	sl.Info(msg,
		slog.String("task", "clean up the releasers such as group_brand_by and .._for"))
	f, err := models.Files(
		qm.Where("group_brand_for = group_brand_by"),
		qm.WithDeleted()).All(ctx, exec)
	if err != nil {
		return fmt.Errorf("update group_brand_for = group_brand_by: %w", err)
	}
	if len(f) > 0 {
		empty := null.NewString("", true)
		rowsAff, err := f.UpdateAll(ctx, exec, models.M{"group_brand_by": empty})
		if err != nil {
			return fmt.Errorf("update all to null group_brand_by: %w", err)
		}
		if rowsAff > 0 {
			sl.Info(msg,
				slog.String("task", "update group_brand_by to null"),
				slog.Int64("updated", rowsAff))
		}
	}
	for bad, fix := range fixes() {
		bad = strings.ToUpper(bad)
		fix = strings.ToUpper(fix)
		f, err = models.Files(
			qm.Where("group_brand_for = ?", bad),
			qm.WithDeleted()).All(ctx, exec)
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
					slog.String("task", "update group_brand_for fixes"),
					slog.Int64("updated", rowsAff))
			}
		}
		f, err = models.Files(
			qm.Where("group_brand_by = ?", bad),
			qm.WithDeleted()).All(ctx, exec)
		if err != nil {
			return fmt.Errorf("where group_brand_by is bad: %w", err)
		}
		if len(f) > 0 {
			rowsAff, err := f.UpdateAll(ctx, exec, models.M{"group_brand_by": fix})
			if err != nil {
				return fmt.Errorf("update all to null group_brand_by fix: %w", err)
			}
			if rowsAff > 0 {
				sl.Info(msg,
					slog.String("task", "update group_brand_by fixes"),
					slog.Int64("updated", rowsAff))
			}
		}
	}
	if err := moreReleases(exec); err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}
	return nil
}

func moreReleases(exec boil.ContextExecutor) error {
	_, err := queries.Raw(postgres.SetUpper("group_brand_for")).Exec(exec)
	if err != nil {
		return fmt.Errorf("set upper group_brand_for: %w", err)
	}
	_, err = queries.Raw(postgres.SetUpper("group_brand_by")).Exec(exec)
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

// optimize reclaims storage occupied by dead tuples in the database and
// also analyzes the most efficient execution plans for queries.
func optimize(db *sql.DB) error {
	if db == nil {
		return fmt.Errorf("fix optimize: %w", panics.ErrNoDB)
	}
	_, err := queries.Raw("VACUUM ANALYZE files").Exec(db)
	if err != nil {
		return fmt.Errorf("execute vacuum and analyze: %w", err)
	}
	return nil
}

// invalidUUIDs will count the number of invalid UUIDs in the database.
// This should be part of a future function to repair the UUIDs and rename the file assets.
func invalidUUIDs(ctx context.Context, exec boil.ContextExecutor, sl *slog.Logger) error {
	const msg = "invalid uuid"
	mods := qm.SQL("SELECT COUNT(*) FROM files WHERE files.uuid" +
		" !~ '^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[1-5][0-9a-fA-F]{3}-[89abAB][0-9a-fA-F]{3}-[0-9a-fA-F]{12}';")
	i, err := models.Files(mods).Count(ctx, exec)
	if err != nil {
		return fmt.Errorf("query couunt: %w", err)
	}
	if i == 0 {
		return nil
	}
	sl.Warn(msg,
		slog.String("task", "invalid uuids found"),
		slog.Int64("finds", i))
	return nil
}

func nullifyEmpty(exec boil.ContextExecutor) error {
	query := ""
	columns := []string{
		"list_relations", "web_id_github", "web_id_youtube",
		"group_brand_for", "group_brand_by", "record_title",
		"credit_text", "credit_program", "credit_illustration", "credit_audio", "comment",
		"dosee_hardware_cpu", "dosee_hardware_graphic", "dosee_hardware_audio",
	}
	for column := range slices.Values(columns) {
		query += UpdateSet + column + " = NULL WHERE " + column + " = ''; "
	}
	if _, err := queries.Raw(query).Exec(exec); err != nil {
		return fmt.Errorf("query execute: %w", err)
	}
	return nil
}

func nullifyZero(exec boil.ContextExecutor) error {
	query := ""
	columns := []string{
		"web_id_pouet", "web_id_demozoo",
		"date_issued_year", "date_issued_month", "date_issued_day",
	}
	for column := range slices.Values(columns) {
		query += UpdateSet + column + " = NULL WHERE " + column + " = 0; "
	}
	if _, err := queries.Raw(query).Exec(exec); err != nil {
		return fmt.Errorf("query execute: %w", err)
	}
	return nil
}

func trimFwdSlash(exec boil.ContextExecutor) error {
	query := ""
	columns := []string{"web_id_16colors"}
	for column := range slices.Values(columns) {
		query += UpdateSet + column + " = LTRIM(web_id_16colors, '/') WHERE web_id_16colors LIKE '/%'; "
	}
	if _, err := queries.Raw(query).Exec(exec); err != nil {
		return fmt.Errorf("query execute: %w", err)
	}
	return nil
}
