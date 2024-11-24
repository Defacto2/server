package model

// Package exists.go contains the database queries for checking if a record exists.

import (
	"context"
	"crypto/sha512"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/Defacto2/helper"
	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

// DemozooExists returns true if the file record exists in the database using a Demozoo production ID.
// This function will also return true for records that have been marked as deleted.
func DemozooExists(ctx context.Context, exec boil.ContextExecutor, id int64) (bool, error) {
	if invalidExec(exec) {
		return false, ErrDB
	}
	ok, err := models.Files(models.FileWhere.WebIDDemozoo.EQ(null.Int64From(id)),
		qm.WithDeleted()).Exists(ctx, exec)
	if err != nil {
		return false, fmt.Errorf("exist demozoo file %d: %w", id, err)
	}
	return ok, nil
}

// PouetExists returns true if the file record exists in the database using a Pouet production ID.
// This function will also return true for records that have been marked as deleted.
func PouetExists(ctx context.Context, exec boil.ContextExecutor, id int64) (bool, error) {
	if invalidExec(exec) {
		return false, ErrDB
	}
	ok, err := models.Files(models.FileWhere.WebIDPouet.EQ(null.Int64From(id)),
		qm.WithDeleted()).Exists(ctx, exec)
	if err != nil {
		return false, fmt.Errorf("exist pouet file %d: %w", id, err)
	}
	return ok, nil
}

// SHA384Exists returns true if the file record exists in the database using a SHA-384 hash.
func SHA384Exists(ctx context.Context, exec boil.ContextExecutor, sha384 []byte) (bool, error) {
	if invalidExec(exec) {
		return false, ErrDB
	}
	hash := hex.EncodeToString(sha384)
	return HashExists(ctx, exec, hash)
}

// HashExists returns true if the file record exists in the database using a SHA-384 hexadecimal hash.
func HashExists(ctx context.Context, exec boil.ContextExecutor, hash string) (bool, error) {
	if invalidExec(exec) {
		return false, ErrDB
	}
	if len(hash) != sha512.Size384*2 {
		return false, fmt.Errorf("%w: %d characters", ErrSha384, len(hash))
	}
	ok, err := models.Files(models.FileWhere.FileIntegrityStrong.EQ(null.StringFrom(hash)),
		qm.WithDeleted()).Exists(ctx, exec)
	if err != nil {
		return false, fmt.Errorf("models file hash %s: %w", hash, err)
	}
	return ok, nil
}

// HashFind returns the obfuscated ID of the file record in the database that was matched
// using a SHA-384 hexadecimal hash. If the hash does not exist in the database then an empty string is returned.
func HashFind(ctx context.Context, exec boil.ContextExecutor, hash string) (string, error) {
	if invalidExec(exec) {
		return "", ErrDB
	}
	if len(hash) != sha512.Size384*2 {
		return "", fmt.Errorf("%w: %d characters", ErrSha384, len(hash))
	}
	file, err := models.Files(models.FileWhere.FileIntegrityStrong.EQ(null.StringFrom(hash)),
		qm.WithDeleted()).One(ctx, exec)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	} else if err != nil {
		return "", fmt.Errorf("models file hash %s: %w", hash, err)
	}
	return helper.ObfuscateID(file.ID), nil
}

// UUIDExists returns true if the file record exists in the database using a UUID.
func UUIDExists(ctx context.Context, exec boil.ContextExecutor, uuid string) (bool, error) {
	if invalidExec(exec) {
		return false, ErrDB
	}
	ok, err := models.Files(models.FileWhere.UUID.EQ(null.StringFrom(uuid)),
		qm.WithDeleted()).Exists(ctx, exec)
	if err != nil {
		return false, fmt.Errorf("exist file uuid %s: %w", uuid, err)
	}
	return ok, nil
}
