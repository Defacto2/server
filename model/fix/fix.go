// Package fix contains functions for repairing the database data.
package fix

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/Defacto2/server/internal/helper"
	"github.com/Defacto2/server/internal/postgres"
	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"go.uber.org/zap"
)

var (
	ErrCtxLog = errors.New("context logger is invalid")
	ErrDB     = errors.New("database connection is nil")
	ErrLog    = errors.New("the server cannot save any logs")
	ErrRepair = errors.New("invalid repair option")
)

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
)

// In the future we may want to add a Debug or TestRun func.

// Run the database repair based on the repair option.
func (r Repair) Run(ctx context.Context, db *sql.DB, tx *sql.Tx) error {
	logger := helper.Logger(ctx)
	logger.Infof("Checking for records with invalid UUID values")
	logger.Infoln("Running a cleanup of the database", r)
	if r < None || r > Releaser {
		return fmt.Errorf("%w: %d", ErrRepair, r)
	}
	if r == None {
		return nil
	}
	if err := invalidUUIDs(ctx, db); err != nil {
		return fmt.Errorf("invalid UUIDs: %w", err)
	}
	if err := coldfusionIDs(ctx, db); err != nil {
		return fmt.Errorf("coldfusion IDs: %w", err)
	}
	switch r {
	case Artifacts:
		logger.Infoln("Cleaning up the artifacts whitespace and null values")
		if err := contentWhiteSpace(tx); err != nil {
			return fmt.Errorf("content white space: %w", err)
		}
		if err := nullifyEmpty(tx); err != nil {
			return fmt.Errorf("nullify empty: %w", err)
		}
		if err := nullifyZero(tx); err != nil {
			return fmt.Errorf("nullify zero: %w", err)
		}
		if err := trimFwdSlash(tx); err != nil {
			return fmt.Errorf("trim forward slash: %w", err)
		}
		if err := trainers(ctx, tx); err != nil {
			return fmt.Errorf("trainers: %w", err)
		}
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("artifacts transaction commit %w", err)
		}
		fallthrough
	case Releaser:
		if err := releasers(ctx, db); err != nil {
			return fmt.Errorf("releasers: %w", err)
		}
	}
	if err := optimize(db); err != nil {
		return fmt.Errorf("optimize: %w", err)
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
func coldfusionIDs(ctx context.Context, exec boil.ContextExecutor) error {
	logger := helper.Logger(ctx)
	logger.Infoln("Checking for invalid UUIDs using the ColdFusion syntax")
	mods := qm.SQL("SELECT uuid FROM files WHERE length(uuid)=35")
	fs, err := models.Files(mods).All(ctx, exec)
	if err != nil {
		return fmt.Errorf("models.Files: %w", err)
	}
	i := len(fs)
	if i == 0 {
		return nil
	}
	logger.Infoln(i, "invalid UUIDs found using the ColdFusion syntax")
	for i, f := range fs {
		if !f.UUID.Valid {
			continue
		}
		// 35 character UUIDs in a 36 character fixed length string will have a tailing space.
		old := strings.TrimSpace(f.UUID.String)
		newid, err := helper.CfUUID(old)
		if err != nil {
			logger.Warnf("%d. %q is invalid, %s", i, newid, err)
			continue
		}
		file, err := models.Files(qm.Where("uuid = ?", old)).One(ctx, exec)
		if err != nil {
			logger.Warnf("%d. %q failed to find, %s", i, old, err)
			continue
		}
		file.UUID = null.StringFrom(newid)
		_, err = file.Update(ctx, exec, boil.Infer())
		if err != nil {
			logger.Warnf("%d. %q failed to update, %s", i, old, err)
			continue
		}
	}
	return nil
}

func trainers(ctx context.Context, tx *sql.Tx) error {
	logger := helper.Logger(ctx)
	const trainer = "gamehack"
	logger.Infof("Checking for trainers that are not categorized as %q", trainer)
	mods := []qm.QueryMod{}
	mods = append(mods, qm.Select("id"))
	mods = append(mods, qm.Where(fmt.Sprintf("section != '%s'", trainer)))
	mods = append(mods, qm.Where("section != 'magazine'"))
	mods = append(mods, qm.Where("record_title ILIKE '%trainer%'"))
	mods = append(mods, qm.Where("platform = 'dos' OR platform = 'windows'"))
	fs, err := models.Files(mods...).All(ctx, tx)
	if err != nil {
		return fmt.Errorf("models.Files: %w", err)
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
		return fmt.Errorf("models.Files: %w", err)
	}
	logger.Infof("Updated %d trainers", rowsAff)
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
	}
}

// releasers will repair the group_brand_by and group_brand_for releasers data.
func releasers(ctx context.Context, exec boil.ContextExecutor) error {
	logger := helper.Logger(ctx)
	logger.Infoln("Cleaning up the releasers group_brand_by and group_brand_for")
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
			logger.Infof("Updated %d group_brand_by to NULL", rowsAff)
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
				logger.Infof("Updated %d groups for to %q", rowsAff, fix)
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
				logger.Infof("Updated %d groups by to %q", rowsAff, fix)
			}
		}
	}
	return moreReleases(ctx, exec)
}

func moreReleases(ctx context.Context, exec boil.ContextExecutor) error {
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
	if err := magics(ctx, exec); err != nil {
		return fmt.Errorf("magics: %w", err)
	}
	return nil
}

func magics(ctx context.Context, exec boil.ContextExecutor) error {
	magics, err := models.Files(qm.Where("file_magic_type ILIKE ?", "ERROR: %")).All(ctx, exec)
	if err != nil {
		return fmt.Errorf("where ilike file_magic_type: %w", err)
	}
	rowsAff, err := magics.UpdateAll(ctx, exec, models.M{"file_magic_type": ""})
	if err != nil {
		return fmt.Errorf("update all file_magic_type: %w", err)
	}
	if rowsAff > 0 {
		logger, loggerExists := ctx.Value("logger").(*zap.SugaredLogger)
		if loggerExists {
			logger.Infof("Removed %d file magic types with errors", rowsAff)
		}
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
	_, err := queries.Raw("VACUUM ANALYZE files").Exec(db)
	if err != nil {
		return fmt.Errorf("execute vacuum and analyze: %w", err)
	}
	return nil
}

// invalidUUIDs will count the number of invalid UUIDs in the database.
// This should be part of a future function to repair the UUIDs and rename the file assets.
func invalidUUIDs(ctx context.Context, exec boil.ContextExecutor) error {
	logger := helper.Logger(ctx)
	mods := qm.SQL("SELECT COUNT(*) FROM files WHERE files.uuid" +
		" !~ '^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[1-5][0-9a-fA-F]{3}-[89abAB][0-9a-fA-F]{3}-[0-9a-fA-F]{12}';")
	i, err := models.Files(mods).Count(ctx, exec)
	if err != nil {
		return fmt.Errorf("query couunt: %w", err)
	}
	if i == 0 {
		return nil
	}
	logger.Warnf("%d invalid UUIDs found", i)
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
	for _, column := range columns {
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
	for _, column := range columns {
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
	for _, column := range columns {
		query += UpdateSet + column + " = LTRIM(web_id_16colors, '/') WHERE web_id_16colors LIKE '/%'; "
	}
	if _, err := queries.Raw(query).Exec(exec); err != nil {
		return fmt.Errorf("query execute: %w", err)
	}
	return nil
}
