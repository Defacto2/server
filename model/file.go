package model

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/Defacto2/server/internal/demozoo"
	"github.com/Defacto2/server/internal/helper"
	"github.com/Defacto2/server/internal/postgres"
	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/google/uuid"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

// FindFile returns true if the file record exists in the database.
// This function will also return true for records that have been marked as deleted.
func ExistFile(ctx context.Context, db *sql.DB, id int64) (bool, error) {
	if db == nil {
		return false, ErrDB
	}
	return models.Files(models.FileWhere.ID.EQ(id), qm.WithDeleted()).Exists(ctx, db)
}

// ExistsHash returns true if the file record exists in the database using a SHA-384 hash.
func ExistsHash(ctx context.Context, db *sql.DB, sha384 []byte) (bool, error) {
	if db == nil {
		return false, ErrDB
	}
	hash := fmt.Sprintf("%x", sha384)
	// todo validate sha384 is not empty, is valid
	strong := null.String{String: hash, Valid: true}
	return models.Files(models.FileWhere.FileIntegrityStrong.EQ(strong), qm.WithDeleted()).Exists(ctx, db)
}

// FindFile retrieves a single file record from the database using the record key.
// This function will also return records that have been marked as deleted.
func FindFile(ctx context.Context, db *sql.DB, id int64) (*models.File, error) {
	if db == nil {
		return nil, ErrDB
	}
	return models.Files(models.FileWhere.ID.EQ(id), qm.WithDeleted()).One(ctx, db)
}

// ExistDemozooFile returns true if the file record exists in the database using a Demozoo production ID.
// This function will also return true for records that have been marked as deleted.
func ExistDemozooFile(ctx context.Context, db *sql.DB, id int64) (bool, error) {
	if db == nil {
		return false, ErrDB
	}
	return models.Files(models.FileWhere.WebIDDemozoo.EQ(null.Int64From(id)), qm.WithDeleted()).Exists(ctx, db)
}

// FindDemozooFile retrieves the ID or key of a single file record from the database using a Demozoo production ID.
// This function will also return records that have been marked as deleted and flag those with the boolean.
// If the record is not found then the function will return an ID of 0 but without an error.
func FindDemozooFile(ctx context.Context, db *sql.DB, id int64) (bool, int64, error) {
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
		return false, 0, err
	}
	deleted := !f.Deletedat.IsZero()
	return deleted, f.ID, nil
}

// InsertDemozooFile inserts a new file record into the database using a Demozoo production ID.
// This will not check if the Demozoo production ID already exists in the database.
// When successful the function will return the new record ID.
func InsertDemozooFile(ctx context.Context, db *sql.DB, id int64) (int64, error) {
	if db == nil {
		return 0, ErrDB
	}
	if id < startID || id > demozoo.Sanity {
		return 0, fmt.Errorf("%w: %d", ErrID, id)
	}
	uid, err := uuid.NewV7()
	if err != nil {
		return 0, err
	}
	now := time.Now()
	f := models.File{
		UUID:         null.StringFrom(uid.String()),
		WebIDDemozoo: null.Int64From(id),
		Deletedat:    null.TimeFromPtr(&now),
	}
	if err = f.Insert(ctx, db, boil.Infer()); err != nil {
		return 0, err
	}
	return f.ID, nil
}

// ExistPouetFile returns true if the file record exists in the database using a Pouet production ID.
// This function will also return true for records that have been marked as deleted.
func ExistPouetFile(ctx context.Context, db *sql.DB, id int64) (bool, error) {
	if db == nil {
		return false, ErrDB
	}
	return models.Files(models.FileWhere.WebIDPouet.EQ(null.Int64From(id)), qm.WithDeleted()).Exists(ctx, db)
}

// FindPouetFile retrieves the ID or key of a single file record from the database using a Pouet production ID.
// This function will also return records that have been marked as deleted and flag those with the boolean.
// If the record is not found then the function will return an ID of 0 but without an error.
func FindPouetFile(ctx context.Context, db *sql.DB, id int64) (bool, int64, error) {
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
		return false, 0, err
	}
	deleted := !f.Deletedat.IsZero()
	return deleted, f.ID, nil
}

// InsertPouetFile inserts a new file record into the database using a Pouet production ID.
// This will not check if the Pouet production ID already exists in the database.
// When successful the function will return the new record ID.
func InsertPouetFile(ctx context.Context, db *sql.DB, id int64) (int64, error) {
	if db == nil {
		return 0, ErrDB
	}
	if id < startID || id > demozoo.Sanity {
		return 0, fmt.Errorf("%w: %d", ErrID, id)
	}
	uid, err := uuid.NewV7()
	if err != nil {
		return 0, err
	}
	now := time.Now()
	f := models.File{
		UUID:       null.StringFrom(uid.String()),
		WebIDPouet: null.Int64From(id),
		Deletedat:  null.TimeFromPtr(&now),
	}
	if err = f.Insert(ctx, db, boil.Infer()); err != nil {
		return 0, err
	}
	return f.ID, nil
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
	art, err := One(ctx, db, deleted, key)
	if err != nil {
		return nil, fmt.Errorf("%w, %w: %d", ErrID, err, key)
	}
	return art, nil
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
