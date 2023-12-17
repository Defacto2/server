package model

import (
	"context"
	"strings"
	"time"

	"github.com/Defacto2/server/internal/postgres"
	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/labstack/echo/v4"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

const (
	uidPlaceholder = `ADB7C2BF-7221-467B-B813-3636FE4AE16B`
)

// UpdateOnline updates the record to be online and public.
func UpdateOnline(c echo.Context, id int64) error {
	if c == nil {
		return ErrCtx
	}
	db, err := postgres.ConnectDB()
	if err != nil {
		return ErrDB
	}
	defer db.Close()
	ctx := context.Background()
	f, err := models.FindFile(ctx, db, id)
	if err != nil {
		return err
	}
	f.Deletedat = null.TimeFromPtr(nil)
	f.Deletedby = null.String{}
	if _, err = f.Update(ctx, db, boil.Infer()); err != nil {
		return err
	}
	return nil
}

// UpdateOffline updates the record to be offline and inaccessible to the public.
func UpdateOffline(c echo.Context, id int64) error {
	if c == nil {
		return ErrCtx
	}
	db, err := postgres.ConnectDB()
	if err != nil {
		return ErrDB
	}
	defer db.Close()
	ctx := context.Background()
	f, err := models.FindFile(ctx, db, id)
	if err != nil {
		return err
	}
	now := time.Now()
	f.Deletedat = null.TimeFromPtr(&now)
	f.Deletedby = null.StringFrom(strings.ToLower(uidPlaceholder))
	if _, err = f.Update(ctx, db, boil.Infer()); err != nil {
		return err
	}
	return nil
}

// UpdateNoReadme updates the retrotxt_no_readme column value with val.
// It returns nil if the update was successful.
// Id is the database id of the record.
func UpdateNoReadme(c echo.Context, id int64, val bool) error {
	if c == nil {
		return ErrCtx
	}

	db, err := postgres.ConnectDB()
	if err != nil {
		return ErrDB
	}
	defer db.Close()

	ctx := context.Background()
	f, err := models.FindFile(ctx, db, id)
	if err != nil {
		return err
	}

	i := int16(0)
	if val {
		i = 1
	}
	f.RetrotxtNoReadme = null.NewInt16(i, true)

	if _, err = f.Update(ctx, db, boil.Infer()); err != nil {
		return err
	}
	return nil
}
