package model

import (
	"context"

	"github.com/Defacto2/server/internal/postgres"
	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/labstack/echo/v4"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

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
