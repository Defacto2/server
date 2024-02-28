// Package fix contains functions for repairing the database data.
package fix

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/Defacto2/server/internal/postgres"
	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/queries"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

var (
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

// In the future we may want to add a Debug or TestRun func.

// Run the database repair based on the repair option.
func (r Repair) Run(ctx context.Context, w io.Writer, db *sql.DB) error {
	if w == nil {
		w = io.Discard
	}
	if db == nil {
		return ErrDB
	}
	if r < None || r > Releaser {
		return fmt.Errorf("%w: %d", ErrRepair, r)
	}
	if r == None {
		return nil
	}
	if err := invalidUUIDs(ctx, w, db); err != nil {
		return err
	}
	if r == (All | Releaser) {
		if err := releasers(ctx, w, db); err != nil {
			return err
		}
	}
	if r == All {
		if err := contentWhiteSpace(db); err != nil {
			return err
		}
	}
	return optimize(db)
}

// Fix bad imported names, such as those from Demozoo data imports.
// Each one of these fixes need a redirect.
const (
	acidbad   = "ACID"
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
)

func fixes() map[string]string {
	return map[string]string{
		acidbad: acidfix,
		icebad:  icefix,
		pwabad:  pwafix,
		trsibad: trsifix,
		xpress:  xpressfix,
		damn:    damnfix,
		ofg:     ofgfix,
		ofg1:    ofgfix,
	}
}

// releasers will repair the group_brand_by and group_brand_for releasers data.
func releasers(ctx context.Context, w io.Writer, db *sql.DB) error {
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
			fmt.Fprintln(w, "updated", rowsAff, "groups for to", fix)
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
			fmt.Fprintln(w, "updated", rowsAff, "groups by to", fix)
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
	return magics(ctx, w, db)
}

func magics(ctx context.Context, w io.Writer, db *sql.DB) error {
	magics, err := models.Files(qm.Where("file_magic_type ILIKE ?", "ERROR: %")).All(ctx, db)
	if err != nil {
		return err
	}
	rowsAff, err := magics.UpdateAll(ctx, db, models.M{"file_magic_type": ""})
	if err != nil {
		return err
	}
	if rowsAff > 0 {
		fmt.Fprintln(w, "removed", rowsAff, "file magic types with errors")
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
func invalidUUIDs(ctx context.Context, w io.Writer, db *sql.DB) error {
	// SELECT *
	mods := qm.SQL("SELECT COUNT(*) FROM files WHERE files.uuid" +
		" !~ '^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[1-5][0-9a-fA-F]{3}-[89abAB][0-9a-fA-F]{3}-[0-9a-fA-F]{12}';")
	i, err := models.Files(mods).Count(ctx, db)
	if err != nil {
		return err
	}
	fmt.Fprintln(w, i, "invalid UUIDs found")
	return nil
}
