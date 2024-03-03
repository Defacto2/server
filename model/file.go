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

// FindUUID retrieves a single file record from the database using the Unique Universal ID.
func FindUUID(uid string) (*models.File, error) {
	return recordUID(false, uid)
}

// EditUUID retrieves a single file record from the database using the Unique Universal ID.
// This function will also return records that have been marked as deleted.
func EditUUID(uid string) (*models.File, error) {
	return recordUID(true, uid)
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

// recordUID retrieves a single file record from the database using the uid URL ID.
func recordUID(deleted bool, uid string) (*models.File, error) {
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
