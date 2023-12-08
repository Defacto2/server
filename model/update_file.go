package model

import (
	"context"

	"github.com/Defacto2/server/internal/postgres"
	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/labstack/echo/v4"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"go.uber.org/zap"
)

func UpdateNoReadme(z *zap.SugaredLogger, c echo.Context, id int64, val bool) error {
	if z == nil {
		return ErrZap
	}
	if c == nil {
		return ErrCtx
	}

	ctx := context.Background()
	db, err := postgres.ConnectDB()
	if err != nil {
		return ErrDB
	}
	defer db.Close()

	f, err := models.FindFile(ctx, db, id)
	if err != nil {
		return err
	}
	i := int16(0)
	if val {
		i = 1
	}
	f.RetrotxtNoReadme = null.NewInt16(i, true)
	_, err = f.Update(ctx, db, boil.Infer())
	if err != nil {
		return err
	}
	return nil
}
