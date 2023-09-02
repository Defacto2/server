package model

import (
	"context"
	"database/sql"
	"strings"

	"github.com/Defacto2/server/pkg/postgres"
	"github.com/Defacto2/server/pkg/postgres/models"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/queries"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

// Package file repair.go contains functions for repairing the database data.

// RepairReleasers will repair the group_brand_by and group_brand_for releasers data.
func RepairReleasers(ctx context.Context, db *sql.DB) error {
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
	const (
		trsibad = "TRISTAR AND RED SECTOR INC"
		trsifix = "TRISTAR & RED SECTOR INC"
		acidbad = "ACID"
		acidfix = "ACID PRODUCTIONS"
		icebad  = "ICE"
		icefix  = "INSANE CREATORS ENTERPRISE"
		pwabad  = "pirates with attitude"
		pwafix  = "pirates with attitudes"
	)
	// TODO: globalize this map and create redirects for the old names?
	fixes := map[string]string{
		trsibad: trsifix,
		acidbad: acidfix,
		icebad:  icefix,
	}
	for bad, fix := range fixes {
		bad = strings.ToUpper(bad)
		fix = strings.ToUpper(fix)
		f, err = models.Files(
			qm.Where("group_brand_for = ?", bad),
			qm.WithDeleted()).All(ctx, db)
		if err != nil {
			return err
		}
		_, err = f.UpdateAll(ctx, db, models.M{"group_brand_for": fix})
		if err != nil {
			return err
		}
		f, err = models.Files(
			qm.Where("group_brand_by = ?", bad),
			qm.WithDeleted()).All(ctx, db)
		if err != nil {
			return err
		}
		_, err = f.UpdateAll(ctx, db, models.M{"group_brand_by": fix})
		if err != nil {
			return err
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
	return nil
}
