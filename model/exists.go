package model

// Package exists.go contains the database queries for checking if a record exists.

import (
	"context"
	"crypto/sha512"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"strconv"

	"github.com/Defacto2/helper"
	"github.com/Defacto2/server/internal/nils"
	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
)

// ExistDemozoo returns true if a file record links to a Demozoo production ID.
// This function will also return true for records that have been marked as deleted.
func ExistDemozoo(ctx context.Context, exec boil.ContextExecutor, prodID int) (bool, error) {
	const format = "demozoo prod exists %s: %w"
	if err := nils.Check(ctx, exec); err != nil {
		return false, fmt.Errorf(format, "check", err)
	}

	id := null.Int64From(int64(math.Abs(float64(prodID))))
	ex := Exist{
		ProdID: prodID,
		Mod:    models.FileWhere.WebIDDemozoo.EQ(id),
		Format: format,
	}
	return ex.remote(ctx, exec)
}

// ExistPouet returns true if a file record links to a Pouet production ID.
// This function will also return true for records that have been marked as deleted.
func ExistPouet(ctx context.Context, exec boil.ContextExecutor, prodID int) (bool, error) {
	const format = "pouet prod exists %s: %w"
	if err := nils.Check(ctx, exec); err != nil {
		return false, fmt.Errorf(format, "check", err)
	}

	id := null.Int64From(int64(math.Abs(float64(prodID))))
	ex := Exist{
		ProdID: prodID,
		Mod:    models.FileWhere.WebIDPouet.EQ(id),
		Format: format,
	}
	return ex.remote(ctx, exec)
}

type Exist struct {
	ProdID int
	Mod    qm.QueryMod
	Format string
}

func (ex Exist) remote(ctx context.Context, exec boil.ContextExecutor) (bool, error) {
	format := ex.Format
	if err := nils.Check(ctx, exec); err != nil {
		return false, fmt.Errorf(format, "check", err)
	}

	ok, err := models.Files(ex.Mod, qm.WithDeleted()).Exists(ctx, exec)
	if err != nil {
		return false, fmt.Errorf(format, strconv.Itoa(ex.ProdID), err)
	}

	return ok, nil
}

// ExistSHA returns true if a file record uses the SHA-384 hash.
func ExistSHA(ctx context.Context, exec boil.ContextExecutor, sha384 []byte) (bool, error) {
	const format = "sha384 exists %s: %w"
	if err := nils.Check(ctx, exec); err != nil {
		return false, fmt.Errorf(format, "check", err)
	}

	if len(sha384) != sha512.Size384 {
		return false, nil
	}

	hexHash := hex.EncodeToString(sha384)
	return ExistHash(ctx, exec, hexHash)
}

// ExistHash returns true if a file record uses the hexadecimal representation of the SHA-384 hash.
func ExistHash(ctx context.Context, exec boil.ContextExecutor, hexHash string) (bool, error) {
	const format = "sha384 exists %s: %w"
	if err := nils.Check(ctx, exec); err != nil {
		return false, fmt.Errorf(format, "check", err)
	}

	if len(hexHash) != sha512.Size384*2 {
		return false, fmt.Errorf(format, strconv.Itoa(len(hexHash)), ErrSha384)
	}

	ok, err := models.Files(
		models.FileWhere.FileIntegrityStrong.EQ(null.StringFrom(hexHash)),
		qm.WithDeleted()).Exists(ctx, exec)
	if err != nil {
		return false, fmt.Errorf(format, hexHash, err)
	}

	return ok, nil
}

// OneByHash returns the obfuscated ID of the file record that uses the
// hexadecimal representation of the SHA-384 hash.
func OneByHash(ctx context.Context, exec boil.ContextExecutor, hexHash string) (string, error) {
	const format = "one by hash %s: %w"
	if err := nils.Check(ctx, exec); err != nil {
		return "", fmt.Errorf(format, "check", err)
	}

	if len(hexHash) != sha512.Size384*2 {
		return "", fmt.Errorf(format, strconv.Itoa(len(hexHash)), ErrSha384)
	}

	file, err := models.Files(
		models.FileWhere.FileIntegrityStrong.EQ(null.StringFrom(hexHash)),
		qm.WithDeleted()).One(ctx, exec)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf(format, hexHash, err)
	}

	key := file.ID
	if key <= 0 {
		return "", nil
	}

	return helper.ObfuscateID(key), nil
}

// ExistUUID returns true if a file record uses the Universal Unique ID.
func ExistUUID(ctx context.Context, exec boil.ContextExecutor, uuid string) (bool, error) {
	const format = "uuid exists %s: %w"
	if err := nils.Check(ctx, exec); err != nil {
		return false, fmt.Errorf(format, "check", err)
	}

	const expect = 36
	if len(uuid) < expect || len(uuid) > expect {
		return false, nil
	}

	ok, err := models.Files(models.FileWhere.UUID.EQ(null.StringFrom(uuid)),
		qm.WithDeleted()).Exists(ctx, exec)
	if err != nil {
		return false, fmt.Errorf("exists file uuid %s: %w", uuid, err)
	}

	return ok, nil
}
