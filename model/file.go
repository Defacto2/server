package model

import (
	"context"
	"crypto/sha512"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"strconv"
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

// ExistSumHash returns true if the file record exists in the database using a SHA-384 hash.
func ExistSumHash(ctx context.Context, db *sql.DB, sha384 []byte) (bool, error) {
	if db == nil {
		return false, ErrDB
	}
	hash := hex.EncodeToString(sha384)
	return ExistHash(ctx, db, hash)
}

// ExistHash returns true if the file record exists in the database using a SHA-384 hexadecimal hash.
func ExistHash(ctx context.Context, db *sql.DB, hash string) (bool, error) {
	if db == nil {
		return false, ErrDB
	}
	if len(hash) != sha512.Size384*2 {
		return false, fmt.Errorf("%w: %d characters", ErrSha384, len(hash))
	}
	return models.Files(models.FileWhere.FileIntegrityStrong.EQ(null.StringFrom(hash)), qm.WithDeleted()).Exists(ctx, db)
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

func InsertUpload(ctx context.Context, db *sql.DB, values url.Values) (int64, error) {
	if db == nil {
		return 0, ErrDB
	}
	uid, err := uuid.NewV7()
	if err != nil {
		return 0, err
	}
	now := time.Now()

	y, _ := strconv.ParseInt(values.Get("year"), 10, 16)
	year := int16(y)
	m, _ := strconv.ParseInt(values.Get("month"), 10, 16)
	month := int16(m)
	s, _ := strconv.ParseInt(values.Get("size"), 10, 64)
	size := int64(s)

	f := models.File{
		UUID:                null.StringFrom(uid.String()),
		Deletedat:           null.TimeFromPtr(&now),
		Createdat:           null.TimeFromPtr(&now),
		WebIDYoutube:        null.StringFrom(values.Get("youtube")), // validate
		GroupBrandFor:       null.StringFrom(values.Get("group")),   // validate and format
		GroupBrandBy:        null.StringFrom(values.Get("brand")),   // validate and format
		RecordTitle:         null.StringFrom(values.Get("title")),   // validate and format
		DateIssuedYear:      null.Int16From(year),
		DateIssuedMonth:     null.Int16From(month),
		Filename:            null.StringFrom(values.Get("filename")), // validate
		Filesize:            size,
		FileMagicType:       null.StringFrom(values.Get("magic")),     // validate
		FileIntegrityStrong: null.StringFrom(values.Get("integrity")), // validate
		FileLastModified:    null.TimeFromPtr(&now),                   // collect from form and validate
		Platform:            null.StringFrom(values.Get("platform")),  // validate
		Section:             null.StringFrom(values.Get("section")),   // hardcode value and validate
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
