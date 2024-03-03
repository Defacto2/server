package model

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/Defacto2/server/internal/helper"
	"github.com/Defacto2/server/internal/postgres"
	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

const startID = 1 // startID is the default, first ID value.

var (
	ErrCtx = errors.New("echo context is nil")
	ErrID  = errors.New("file download database id cannot be found")
	ErrZap = errors.New("zap logger instance is nil")
)

// FindFile retrieves a single file record from the database using the record key.
// This function will also return records that have been marked as deleted.
func FindFile(ctx context.Context, db *sql.DB, id int64) (*models.File, error) {
	if db == nil {
		return nil, ErrDB
	}
	return models.Files(models.FileWhere.ID.EQ(id), qm.WithDeleted()).One(ctx, db)
}

// FindObf retrieves a single file record from the database using the obfuscated record key.
func FindObf(key string) (*models.File, error) {
	return recordObf(false, key)
}

// EditObf retrieves a single file record from the database using the obfuscated record key.
// This function will also return records that have been marked as deleted.
func EditObf(key string) (*models.File, error) {
	return recordObf(true, key)
}

// Find retrieves a single file record from the database using the record key.
func Find(key int) (*models.File, error) {
	return record(false, key)
}

// EditFind retrieves a single file record from the database using the record key.
// This function will also return records that have been marked as deleted.
func EditFind(key int) (*models.File, error) {
	return record(true, key)
}

// Record retrieves a single file record from the database using the record key.
func record(deleted bool, key int) (*models.File, error) {
	// get record id, filename, uuid
	ctx := context.Background()
	db, err := postgres.ConnectDB()
	if err != nil {
		return nil, ErrDB
	}
	defer db.Close()
	res, err := One(ctx, db, deleted, key)
	if err != nil {
		return nil, ErrDB
	}
	return res, nil
}

// recordObf retrieves a single file record from the database using the uid URL ID.
func recordObf(deleted bool, key string) (*models.File, error) {
	id := helper.DeobfuscateID(key)
	if id < startID {
		return nil, fmt.Errorf("%w: %d ~ %s", ErrID, id, key)
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
		return nil, fmt.Errorf("%w: %d ~ %s", ErrID, id, key)
	}
	return res, nil
}
