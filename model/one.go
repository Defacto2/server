package model

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/Defacto2/server/internal/helper"
	"github.com/Defacto2/server/internal/postgres"
	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

// OneEditByKey retrieves a single file record from the database using the obfuscated record key.
// This function will also return records that have been marked as deleted.
func OneEditByKey(key string) (*models.File, error) {
	return recordObf(true, key)
}

// OneFile retrieves a single file record from the database using the record key.
// This function will also return records that have been marked as deleted.
func OneFile(ctx context.Context, db *sql.DB, id int64) (*models.File, error) {
	if db == nil {
		return nil, ErrDB
	}
	f, err := models.Files(models.FileWhere.ID.EQ(id), qm.WithDeleted()).One(ctx, db)
	if err != nil {
		return nil, fmt.Errorf("models file one %d: %w", id, err)
	}
	return f, nil
}

// OneFileByKey retrieves a single file record from the database using the obfuscated record key.
func OneFileByKey(key string) (*models.File, error) {
	return recordObf(false, key)
}

// OneDemozoo retrieves the ID or key of a single file record from the database using a Demozoo production ID.
// This function will also return records that have been marked as deleted and flag those with the boolean.
// If the record is not found then the function will return an ID of 0 but without an error.
func OneDemozoo(ctx context.Context, db *sql.DB, id int64) (bool, int64, error) {
	if db == nil {
		return false, 0, ErrDB
	}
	f, err := models.Files(
		qm.Select("id", "deletedat"),
		models.FileWhere.WebIDDemozoo.EQ(null.Int64From(id)),
		qm.WithDeleted()).One(ctx, db)
	if errors.Is(err, sql.ErrNoRows) {
		return false, 0, nil
	}
	if err != nil {
		return false, 0, fmt.Errorf("find demozoo file: %w", err)
	}
	deleted := !f.Deletedat.IsZero()
	return deleted, f.ID, nil
}

// OnePouet retrieves the ID or key of a single file record from the database using a Pouet production ID.
// This function will also return records that have been marked as deleted and flag those with the boolean.
// If the record is not found then the function will return an ID of 0 but without an error.
func OnePouet(ctx context.Context, db *sql.DB, id int64) (bool, int64, error) {
	if db == nil {
		return false, 0, ErrDB
	}
	f, err := models.Files(
		qm.Select("id", "deletedat"),
		models.FileWhere.WebIDPouet.EQ(null.Int64From(id)),
		qm.WithDeleted()).One(ctx, db)
	if errors.Is(err, sql.ErrNoRows) {
		return false, 0, nil
	}
	if err != nil {
		return false, 0, fmt.Errorf("find pouet file: %w", err)
	}
	deleted := !f.Deletedat.IsZero()
	return deleted, f.ID, nil
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
	art, err := One(ctx, db, deleted, id)
	if err != nil {
		return nil, fmt.Errorf("%w, %w: %s", ErrID, err, key)
	}
	if art.ID != int64(id) {
		return nil, fmt.Errorf("%w: %d ~ %s", ErrID, id, key)
	}
	return art, nil
}