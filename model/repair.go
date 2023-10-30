package model

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"strings"

	"github.com/Defacto2/server/internal/postgres"
	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/queries"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

// Package file repair.go contains functions for repairing the database data.

// RepairReleasers will repair the group_brand_by and group_brand_for releasers data.
func RepairReleasers(w io.Writer, ctx context.Context, db *sql.DB) error {
	if w == nil {
		w = io.Discard
	}
	if db == nil {
		return ErrDB
	}
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
	rowsAff := int64(0)
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
