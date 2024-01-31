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

var ErrDB = errors.New("database connection is nil")

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
	switch r {
	case None:
		return nil
	case Releaser, All:
		return releasers(context.Background(), w, db)
	}
	return fmt.Errorf("invalid repair option %d", r)
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
	_, err = f.UpdateAll(ctx, db, models.M{"group_brand_by": x})
	if err != nil {
		return err
	}
	// fix bad imported names, such as those from Demozoo data imports
	// each one of these fixes need a redirect
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
	)
	fixes := map[string]string{
		acidbad: acidfix,
		icebad:  icefix,
		pwabad:  pwafix,
		trsibad: trsifix,
		xpress:  xpressfix,
	}
	var rowsAff int64
	for bad, fix := range fixes {
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

	magics, err := models.Files(qm.Where("file_magic_type ILIKE ?", "ERROR: %")).All(ctx, db)
	if err != nil {
		return err
	}
	rowsAff, err = magics.UpdateAll(ctx, db, models.M{"file_magic_type": ""})
	if err != nil {
		return err
	}
	if rowsAff > 0 {
		fmt.Fprintln(w, "removed", rowsAff, "file magic types with errors")
	}
	return nil
}
