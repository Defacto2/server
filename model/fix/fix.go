// Package fix contains functions for repairing the database data.
package fix

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/Defacto2/server/internal/postgres"
	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/google/uuid"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"go.uber.org/zap"
)

var (
	ErrCtxLog = errors.New("context logger is invalid")
	ErrDB     = errors.New("database connection is nil")
	ErrRepair = errors.New("invalid repair option")
)

// Repair a column or type of data within the database.
type Repair int

const (
	None     Repair = iota - 1 // None does nothing.
	All                        // All repairs all the repairable data.
	Releaser                   // Releaser focuses on the releaser data using the group_brand_by and group_brand_for columns.
)

const (
	UpdateSet = "UPDATE files SET "
)

// In the future we may want to add a Debug or TestRun func.

// Run the database repair based on the repair option.
func (r Repair) Run(ctx context.Context, logr *zap.SugaredLogger, db *sql.DB) error {
	if db == nil {
		return ErrDB
	}
	if r < None || r > Releaser {
		return fmt.Errorf("%w: %d", ErrRepair, r)
	}
	if r == None {
		return nil
	}
	if err := invalidUUIDs(ctx, db); err != nil {
		return err
	}
	if err := coldfusionIDs(ctx, db); err != nil {
		return err
	}
	switch r {
	case All:
		if err := contentWhiteSpace(db); err != nil {
			return err
		}
		if err := nullifyEmpty(db); err != nil {
			return err
		}
		if err := nullifyZero(db); err != nil {
			return err
		}
		if err := trimFwdSlash(db); err != nil {
			return err
		}
		fallthrough
	case Releaser:
		if err := releasers(ctx, logr, db); err != nil {
			return err
		}
	}
	return optimize(db)
}

// coldfusionIDs will fix the invalid [ColdFusion language syntax] UUIDs in the database
// and rename the file assets using the new UUIDs.
// ColdFusion uses an invalid 35 character UUID, which is a 32 character UUID with 3 hyphens,
// while the standard UUID is 36 characters with 4 hyphens.
//
// A blank UUID is "00000000-0000-0000-0000-000000000000".
//
// A blank CFID is "00000000-0000-0000-0000000000000000".
//
// [ColdFusion language syntax]: https://cfdocs.org/createuuid
func coldfusionIDs(ctx context.Context, db *sql.DB) error {
	mods := qm.SQL("SELECT uuid FROM files WHERE length(uuid)=35")
	fs, err := models.Files(mods).All(ctx, db)
	if err != nil {
		return err
	}
	i := len(fs)
	if i == 0 {
		return nil
	}
	logr, ok := ctx.Value("logger").(*zap.SugaredLogger)
	if !ok {
		return ErrCtxLog
	}
	logr.Infoln(i, "invalid UUIDs found using the ColdFusion syntax")
	for i, f := range fs {
		if !f.UUID.Valid {
			continue
		}

		const pos = 23
		// 35 character UUIDs in a 36 character fixed length string will have a tailing space.
		oldUUID := strings.TrimSpace(f.UUID.String)
		r := []rune(oldUUID)
		r = append(r[:pos], append([]rune{'-'}, r[pos:]...)...)
		newUUID := string(r)

		err := uuid.Validate(newUUID)
		if err != nil {
			logr.Warnln("%d. %q is invalid, %s\n", i, newUUID, err)
			continue
		}

		file, err := models.Files(qm.Where("uuid = ?", oldUUID)).One(ctx, db)
		if err != nil {
			logr.Warnln("%d. %q failed to find, %s\n", i, oldUUID, err)
			continue
		}
		file.UUID = null.StringFrom(newUUID)
		_, err = file.Update(ctx, db, boil.Infer())
		if err != nil {
			logr.Warnln("%d. %q failed to update, %s\n", i, oldUUID, err)
			continue
		}
	}
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
func releasers(ctx context.Context, logr *zap.SugaredLogger, db *sql.DB) error {
	x := null.NewString("", true)
	f, err := models.Files(
		qm.Where("group_brand_for = group_brand_by"),
		qm.WithDeleted()).All(ctx, db)
	if err != nil {
		return err
	}
	if _, err = f.UpdateAll(ctx, db, models.M{"group_brand_by": x}); err != nil {
		return err
	}
	var rowsAff int64
	for bad, fix := range fixes() {
		bad = strings.ToUpper(bad)
		fix = strings.ToUpper(fix)
		f, err = models.Files(
			qm.Where("group_brand_for = ?", bad),
			qm.WithDeleted()).All(ctx, db)
		if err != nil {
			return err
		}
		rowsAff, err = f.UpdateAll(ctx, db, models.M{"group_brand_for": fix})
		if err != nil {
			return err
		}
		if rowsAff > 0 {
			logr.Infoln("updated", rowsAff, "groups for to", fix)
		}
		f, err = models.Files(
			qm.Where("group_brand_by = ?", bad),
			qm.WithDeleted()).All(ctx, db)
		if err != nil {
			return err
		}
		rowsAff, err = f.UpdateAll(ctx, db, models.M{"group_brand_by": fix})
		if err != nil {
			return err
		}
		if rowsAff > 0 {
			logr.Infoln("updated", rowsAff, "groups by to", fix)
		}
	}
	_, err = queries.Raw(postgres.SetUpper("group_brand_for")).Exec(db)
	if err != nil {
		return err
	}
	_, err = queries.Raw(postgres.SetUpper("group_brand_by")).Exec(db)
	if err != nil {
		return err
	}
	_, err = queries.Raw(postgres.SetFilesize0()).Exec(db)
	if err != nil {
		return err
	}
	return magics(ctx, db)
}

func magics(ctx context.Context, db *sql.DB) error {
	magics, err := models.Files(qm.Where("file_magic_type ILIKE ?", "ERROR: %")).All(ctx, db)
	if err != nil {
		return err
	}
	rowsAff, err := magics.UpdateAll(ctx, db, models.M{"file_magic_type": ""})
	if err != nil {
		return err
	}
	if rowsAff > 0 {
		logr, ok := ctx.Value("logger").(*zap.SugaredLogger)
		if ok {
			logr.Infoln("removed", rowsAff, "file magic types with errors")
		}
	}
	return nil
}

// contentWhiteSpace will remove any duplicate newline white space from file_zip_content.
func contentWhiteSpace(db *sql.DB) error {
	_, err := queries.Raw("UPDATE files SET file_zip_content = RTRIM(regexp_replace(file_zip_content, '\n+', '\n', 'g'), '\r');").Exec(db)
	if err != nil {
		return err
	}
	return nil
}

// optimize reclaims storage occupied by dead tuples in the database and
// also analyzes the most efficient execution plans for queries.
func optimize(db *sql.DB) error {
	_, err := queries.Raw("VACUUM ANALYZE files").Exec(db)
	if err != nil {
		return err
	}
	return nil
}

// invalidUUIDs will count the number of invalid UUIDs in the database.
// This should be part of a future function to repair the UUIDs and rename the file assets.
func invalidUUIDs(ctx context.Context, db *sql.DB) error {
	mods := qm.SQL("SELECT COUNT(*) FROM files WHERE files.uuid" +
		" !~ '^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[1-5][0-9a-fA-F]{3}-[89abAB][0-9a-fA-F]{3}-[0-9a-fA-F]{12}';")
	i, err := models.Files(mods).Count(ctx, db)
	if err != nil {
		return err
	}
	if i == 0 {
		return nil
	}
	logr, ok := ctx.Value("logger").(*zap.SugaredLogger)
	if ok {
		logr.Warnf("%d invalid UUIDs found", i)
	}
	return nil
}

func nullifyEmpty(db *sql.DB) error {
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
	if _, err := queries.Raw(query).Exec(db); err != nil {
		return err
	}
	return nil
}

func nullifyZero(db *sql.DB) error {
	query := ""
	columns := []string{
		"web_id_pouet", "web_id_demozoo",
		"date_issued_year", "date_issued_month", "date_issued_day",
	}
	for _, column := range columns {
		query += UpdateSet + column + " = NULL WHERE " + column + " = 0; "
	}
	if _, err := queries.Raw(query).Exec(db); err != nil {
		return err
	}
	return nil
}

func trimFwdSlash(db *sql.DB) error {
	query := ""
	columns := []string{"web_id_16colors"}
	for _, column := range columns {
		query += UpdateSet + column + " = LTRIM(web_id_16colors, '/') WHERE web_id_16colors LIKE '/%'; "
	}
	if _, err := queries.Raw(query).Exec(db); err != nil {
		return err
	}
	return nil
}
