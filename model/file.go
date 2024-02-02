package model

import (
	"context"
	"errors"
	"fmt"

	"github.com/Defacto2/server/internal/helper"
	"github.com/Defacto2/server/internal/postgres"
	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

const startID = 1 // startID is the default, first ID value.

var (
	ErrCtx = errors.New("echo context is nil")
	ErrID  = errors.New("file download database id cannot be found")
	ErrZap = errors.New("zap logger instance is nil")
)

// OneRecord retrieves a single file record from the database using the uid URL ID.
func OneRecord(z *zap.SugaredLogger, c echo.Context, deleted bool, uid string) (*models.File, error) {
	if z == nil {
		return nil, ErrZap
	}
	if c == nil {
		return nil, ErrCtx
	}
	id := helper.DeobfuscateID(uid)
	if id < startID {
		return nil, fmt.Errorf("%w: %d ~ %s", ErrID, id, uid)
	}
	// get record id, filename, uuid
	ctx := context.Background()
	db, err := postgres.ConnectDB()
	if err != nil {
		return nil, ErrDB
	}
	defer db.Close()
	res, err := One(ctx, db, deleted, id)
	if err != nil {
		return nil, ErrDB
	}
	if res.ID != int64(id) {
		return nil, fmt.Errorf("%w: %d ~ %s", ErrID, id, uid)
	}
	return res, nil
}

// Record retrieves a single file record from the database using the record key.
func Record(z *zap.SugaredLogger, c echo.Context, key int) (*models.File, error) {
	if z == nil {
		return nil, ErrZap
	}
	if c == nil {
		return nil, ErrCtx
	}
	// get record id, filename, uuid
	ctx := context.Background()
	db, err := postgres.ConnectDB()
	if err != nil {
		return nil, ErrDB
	}
	defer db.Close()
	res, err := One(ctx, db, false, key)
	if err != nil {
		return nil, ErrDB
	}
	return res, nil
}
