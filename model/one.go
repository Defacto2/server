//nolint:nonamedreturns
package model

// Package one.go contains the database queries for retrieving a single record.

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"

	"github.com/Defacto2/helper"
	"github.com/Defacto2/server/handler/demozoo"
	"github.com/Defacto2/server/handler/pouet"
	"github.com/Defacto2/server/internal/nils"
	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	"github.com/google/uuid"
)

// One retrieves a single file record from the database using the record key.
// This function can return records that have been marked as deleted.
func One(ctx context.Context, exec boil.ContextExecutor, withDeleted bool, key int) (*models.File, error) {
	const format = "one record %s: %w"
	if err := nils.Check(ctx, exec); err != nil {
		return nil, fmt.Errorf(format, "check", err)
	}
	if key < 0 {
		return nil, fmt.Errorf(format, strconv.Itoa(key), ErrBadIDInt)
	}

	mods := models.FileWhere.ID.EQ(int64(key))

	var file *models.File
	var err error
	if withDeleted {
		file, err = models.Files(mods, qm.WithDeleted()).One(ctx, exec)
	} else {
		file, err = models.Files(mods).One(ctx, exec)
	}
	if err != nil {
		return nil, fmt.Errorf(format, strconv.Itoa(key), err)
	}

	return file, nil
}

// OneByUUID returns the record associated with the UUID key.
// Generally this method of retrieval is less efficient than using the numeric, record key ID.
func OneByUUID(ctx context.Context, exec boil.ContextExecutor, withDeleted bool, uid string) (*models.File, error) {
	const format = "one record by uuid %s: %w"
	if err := nils.Check(ctx, exec); err != nil {
		return nil, fmt.Errorf(format, "check", err)
	}

	val, err := uuid.Parse(uid)
	if err != nil {
		return nil, fmt.Errorf(format, uid, err)
	}

	mods := models.FileWhere.UUID.EQ(null.NewString(val.String(), true))

	var file *models.File
	if withDeleted {
		file, err = models.Files(mods, qm.WithDeleted()).One(ctx, exec)
	} else {
		file, err = models.Files(mods).One(ctx, exec)
	}
	if err != nil {
		return nil, fmt.Errorf(format, uid, err)
	}

	return file, nil
}

// OneFile retrieves a single file record from the database using the record key.
// This function will also return records that have been marked as deleted.
func OneFile(ctx context.Context, exec boil.ContextExecutor, key int64) (*models.File, error) {
	const format = "one file inc deleted record %s: %w"
	if err := nils.Check(ctx, exec); err != nil {
		return nil, fmt.Errorf(format, "check", err)
	}

	f, err := models.Files(
		models.FileWhere.ID.EQ(key), qm.WithDeleted()).One(ctx, exec)
	if err != nil {
		return nil, fmt.Errorf(format, strconv.Itoa(int(key)), err)
	}

	return f, nil
}

// OneDemozoo retrieves the ID or key of a single file record from the database using a Demozoo production ID.
// This function will also return records that have been marked as deleted and flag those with the boolean.
// If the record is not found then the function will return an ID of 0 but without an error.
func OneDemozoo(ctx context.Context, exec boil.ContextExecutor, prodID int64) (
	deleted bool, key int64, err error,
) {
	const format = "one record by demozoo %s: %w"
	if err := nils.Check(ctx, exec); err != nil {
		return false, 0, fmt.Errorf(format, "check", err)
	}

	if prodID < 1 || prodID > demozoo.Sanity {
		return false, 0, nil
	}

	f, err := models.Files(
		qm.Select("id", "deletedat"),
		models.FileWhere.WebIDDemozoo.EQ(null.Int64From(prodID)),
		qm.WithDeleted()).One(ctx, exec)
	if errors.Is(err, sql.ErrNoRows) {
		return false, 0, nil
	}
	if err != nil {
		return false, 0, fmt.Errorf(format, "get record", err)
	}

	deleted = !f.Deletedat.IsZero()
	key = f.ID

	return deleted, key, nil
}

// OnePouet retrieves the ID or key of a single file record from the database using a Pouet production ID.
// This function will also return records that have been marked as deleted and flag those with the boolean.
// If the record is not found then the function will return an ID of 0 but without an error.
func OnePouet(ctx context.Context, exec boil.ContextExecutor, prodID int64) (
	deleted bool, key int64, err error,
) {
	const format = "one record by pouet %s: %w"
	if err := nils.Check(ctx, exec); err != nil {
		return false, 0, fmt.Errorf(format, "check", err)
	}

	if prodID < 1 || prodID > pouet.Sanity {
		return false, 0, nil
	}

	f, err := models.Files(
		qm.Select("id", "deletedat"),
		models.FileWhere.WebIDPouet.EQ(null.Int64From(prodID)),
		qm.WithDeleted()).One(ctx, exec)
	if errors.Is(err, sql.ErrNoRows) {
		return false, 0, nil
	}
	if err != nil {
		return false, 0, fmt.Errorf(format, "get record", err)
	}

	deleted = !f.Deletedat.IsZero()
	key = f.ID

	return deleted, key, nil
}

// OneEditByKey retrieves a single file record from the database using the obfuscated record key.
// This function will also return records that have been marked as deleted.
func OneEditByKey(ctx context.Context, exec boil.ContextExecutor, obfsKey string) (*models.File, error) {
	return recordObf(ctx, exec, true, obfsKey)
}

// OneFileByKey retrieves a single file record from the database using the obfuscated record key.
func OneFileByKey(ctx context.Context, exec boil.ContextExecutor, obfsKey string) (*models.File, error) {
	return recordObf(ctx, exec, false, obfsKey)
}

// recordObf retrieves a single file record from the database using the uid URL ID.
func recordObf(ctx context.Context, exec boil.ContextExecutor, withDeleted bool, obfsKey string) (*models.File, error) {
	const format = "one record by obfuscated key %s: %w"
	if err := nils.Check(ctx, exec); err != nil {
		return nil, fmt.Errorf(format, "check", err)
	}

	id := helper.DeobfuscateID(obfsKey)
	if id < startID {
		return nil, fmt.Errorf("%w: %d ~ %s", ErrBadID, id, obfsKey)
	}

	// get record id, filename, uuid
	art, err := One(ctx, exec, withDeleted, id)
	if err != nil {
		return nil, fmt.Errorf(format, obfsKey, err)
	}

	if art.ID != int64(id) {
		return nil, fmt.Errorf(format, obfsKey, ErrBadID)
	}

	return art, nil
}
