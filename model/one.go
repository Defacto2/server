package model

// Package one.go contains the database queries for retrieving a single record.

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/Defacto2/helper"
	"github.com/Defacto2/server/internal/panics"
	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	"github.com/google/uuid"
)

// One retrieves a single file record from the database using the record key.
// This function can return records that have been marked as deleted.
func One(ctx context.Context, exec boil.ContextExecutor, deleted bool, key int) (*models.File, error) {
	panics.BoilExecCrash(exec)
	if key < -1 {
		return nil, fmt.Errorf("key value %d: %w", key, ErrKey)
	}
	mods := models.FileWhere.ID.EQ(int64(key))
	var file *models.File
	var err error
	if deleted {
		file, err = models.Files(mods, qm.WithDeleted()).One(ctx, exec)
	} else {
		file, err = models.Files(mods).One(ctx, exec)
	}
	if err != nil {
		return nil, fmt.Errorf("one record %d: %w", key, err)
	}
	return file, nil
}

// OneEditByKey retrieves a single file record from the database using the obfuscated record key.
// This function will also return records that have been marked as deleted.
func OneEditByKey(ctx context.Context, exec boil.ContextExecutor, key string) (*models.File, error) {
	return recordObf(ctx, exec, true, key)
}

// OneByUUID returns the record associated with the UUID key.
// Generally this method of retrieval is less efficient than using the numeric, record key ID.
func OneByUUID(ctx context.Context, exec boil.ContextExecutor, deleted bool, uid string) (*models.File, error) {
	panics.BoilExecCrash(exec)
	val, err := uuid.Parse(uid)
	if err != nil {
		return nil, fmt.Errorf("uuid validation %s: %w", uid, err)
	}
	mods := models.FileWhere.UUID.EQ(null.NewString(val.String(), true))
	var file *models.File
	if deleted {
		file, err = models.Files(mods, qm.WithDeleted()).One(ctx, exec)
	} else {
		file, err = models.Files(mods).One(ctx, exec)
	}
	if err != nil {
		return nil, fmt.Errorf("one record %s: %w", uid, err)
	}
	return file, nil
}

// OneFile retrieves a single file record from the database using the record key.
// This function will also return records that have been marked as deleted.
func OneFile(ctx context.Context, exec boil.ContextExecutor, id int64) (*models.File, error) {
	panics.BoilExecCrash(exec)
	f, err := models.Files(models.FileWhere.ID.EQ(id), qm.WithDeleted()).One(ctx, exec)
	if err != nil {
		return nil, fmt.Errorf("models file one %d: %w", id, err)
	}
	return f, nil
}

// OneFileByKey retrieves a single file record from the database using the obfuscated record key.
func OneFileByKey(ctx context.Context, exec boil.ContextExecutor, key string) (*models.File, error) {
	return recordObf(ctx, exec, false, key)
}

// OneDemozoo retrieves the ID or key of a single file record from the database using a Demozoo production ID.
// This function will also return records that have been marked as deleted and flag those with the boolean.
// If the record is not found then the function will return an ID of 0 but without an error.
func OneDemozoo(ctx context.Context, exec boil.ContextExecutor, id int64) (bool, int64, error) {
	panics.BoilExecCrash(exec)
	f, err := models.Files(
		qm.Select("id", "deletedat"),
		models.FileWhere.WebIDDemozoo.EQ(null.Int64From(id)),
		qm.WithDeleted()).One(ctx, exec)
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
func OnePouet(ctx context.Context, exec boil.ContextExecutor, id int64) (bool, int64, error) {
	panics.BoilExecCrash(exec)
	f, err := models.Files(
		qm.Select("id", "deletedat"),
		models.FileWhere.WebIDPouet.EQ(null.Int64From(id)),
		qm.WithDeleted()).One(ctx, exec)
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
func recordObf(ctx context.Context, exec boil.ContextExecutor, deleted bool, key string) (*models.File, error) {
	panics.BoilExecCrash(exec)
	id := helper.DeobfuscateID(key)
	if id < startID {
		return nil, fmt.Errorf("%w: %d ~ %s", ErrID, id, key)
	}
	// get record id, filename, uuid
	art, err := One(ctx, exec, deleted, id)
	if err != nil {
		return nil, fmt.Errorf("%w, %w: %s", ErrID, err, key)
	}
	if art.ID != int64(id) {
		return nil, fmt.Errorf("%w: %d ~ %s", ErrID, id, key)
	}
	return art, nil
}
