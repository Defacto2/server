package model

// Package file inserts.go contains the database queries for inserting new file records.

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"time"

	"github.com/Defacto2/server/internal/demozoo"
	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/google/uuid"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

// InsertDemozoo inserts a new file record into the database using a Demozoo production ID.
// This will not check if the Demozoo production ID already exists in the database.
// When successful the function will return the new record ID.
func InsertDemozoo(ctx context.Context, exec boil.ContextExecutor, id int64) (int64, error) {
	if exec == nil {
		return 0, ErrDB
	}
	if id < startID || id > demozoo.Sanity {
		return 0, fmt.Errorf("%w: %d", ErrID, id)
	}

	now, uid, err := NewV7()
	if err != nil {
		return 0, fmt.Errorf("uuid.NewV7: %w", err)
	}

	f := models.File{
		UUID:         null.StringFrom(uid.String()),
		WebIDDemozoo: null.Int64From(id),
		Deletedat:    null.TimeFromPtr(&now),
	}
	if err = f.Insert(ctx, exec, boil.Infer()); err != nil {
		return 0, fmt.Errorf("f.Insert: %w", err)
	}
	return f.ID, nil
}

// InsertPouet inserts a new file record into the database using a Pouet production ID.
// This will not check if the Pouet production ID already exists in the database.
// When successful the function will return the new record ID.
func InsertPouet(ctx context.Context, exec boil.ContextExecutor, id int64) (int64, error) {
	if exec == nil {
		return 0, ErrDB
	}
	if id < startID || id > demozoo.Sanity {
		return 0, fmt.Errorf("%w: %d", ErrID, id)
	}

	now, uid, err := NewV7()
	if err != nil {
		return 0, fmt.Errorf("uuid.NewV7: %w", err)
	}

	f := models.File{
		UUID:       null.StringFrom(uid.String()),
		WebIDPouet: null.Int64From(id),
		Deletedat:  null.TimeFromPtr(&now),
	}
	if err = f.Insert(ctx, exec, boil.Infer()); err != nil {
		return 0, fmt.Errorf("f.Insert: %w", err)
	}
	return f.ID, nil
}

// InsertUpload inserts a new file record into the database using a URL values map.
// This will not check if the file already exists in the database.
// Invalid values will be ignored, but will not prevent the record from being inserted.
// When successful the function will return the new record ID key and the UUID.
func InsertUpload(ctx context.Context, tx *sql.Tx, values url.Values, key string) (int64, uuid.UUID, error) {
	noID := uuid.UUID{}
	if tx == nil {
		return 0, noID, ErrDB
	}
	now, uid, err := NewV7()
	if err != nil {
		return 0, noID, fmt.Errorf("uuid.NewV7: %w", err)
	}
	unique := null.StringFrom(uid.String())
	deleteT := null.TimeFromPtr(&now)
	if !deleteT.Valid || deleteT.Time.IsZero() {
		return 0, noID, fmt.Errorf("%w: %v", ErrTime, deleteT.Time)
	}
	createT := null.TimeFromPtr(&now)
	if !createT.Valid || createT.Time.IsZero() {
		return 0, noID, fmt.Errorf("%w: %v", ErrTime, createT.Time)
	}
	f := models.File{
		UUID:      unique,
		Deletedat: deleteT,
		Createdat: createT,
	}
	f, err = upload(f, values, key)
	if err != nil {
		return 0, noID, fmt.Errorf("upload: %w", err)
	}
	if err = f.Insert(ctx, tx, boil.Infer()); err != nil {
		return 0, noID, fmt.Errorf("f.Insert: %w", err)
	}
	if err = tx.Commit(); err != nil {
		return 0, noID, fmt.Errorf("tx.Commit: %w", err)
	}
	return f.ID, uid, nil
}

func upload(f models.File, values url.Values, key string) (models.File, error) {
	youtube, err := ValidYouTube(values.Get(key + "-youtube"))
	if err != nil {
		return f, fmt.Errorf("ValidYouTube: %w", err)
	}
	releaser1, releaser2 := ValidReleasers(
		values.Get(key+"-releaser1"),
		values.Get(key+"-releaser2"),
	)
	title := ValidTitle(values.Get(key + "-title"))
	year, month, day := ValidDateIssue(
		values.Get(key+"-year"),
		values.Get(key+"-month"),
		values.Get(key+"-day"),
	)
	fname := values.Get(key + "-filename")
	filename := ValidFilename(fname)
	if !filename.Valid || filename.IsZero() {
		return f, fmt.Errorf("%w: %v", ErrName, key+"-filename is required")
	}
	filesize, err := ValidFilesize(values.Get(key + "-size"))
	if err != nil {
		return f, fmt.Errorf("ValidFilesize: %w", err)
	}
	content := ValidString(values.Get(key + "-content"))
	readme := ValidFilename(values.Get(key + "-readme"))
	filemagic := ValidMagic(values.Get(key + "-magic"))
	integrity := ValidIntegrity(values.Get(key + "-integrity"))
	lastMod := ValidLastMod(values.Get(key + "-lastmodified"))
	platform := ValidPlatform(values.Get(key + "-operating-system"))
	section := ValidSection(values.Get(key + "-category"))
	creditT := ValidSceners(values.Get(key + "-credittext"))
	creditI := ValidSceners(values.Get(key + "-creditill"))
	creditP := ValidSceners(values.Get(key + "-creditprog"))
	creditA := ValidSceners(values.Get(key + "-creditaudio"))

	f.WebIDYoutube = youtube
	f.GroupBrandFor = releaser1
	f.GroupBrandBy = releaser2
	f.RecordTitle = title
	f.DateIssuedYear = year
	f.DateIssuedMonth = month
	f.DateIssuedDay = day
	f.Filename = filename
	f.Filesize = filesize
	f.FileZipContent = content
	f.RetrotxtReadme = readme
	f.FileMagicType = filemagic
	f.FileIntegrityStrong = integrity
	f.FileLastModified = lastMod
	f.Platform = platform
	f.Section = section
	f.CreditText = creditT
	f.CreditIllustration = creditI
	f.CreditProgram = creditP
	f.CreditAudio = creditA
	return f, nil
}

// NewV7 generates a new UUID version 7, if that fails then it will fallback to version 1.
// It also returns the current time.
func NewV7() (time.Time, uuid.UUID, error) {
	now := time.Now()
	uid, err := uuid.NewV7()
	if err == nil {
		return now, uid, nil
	}
	uid, err = uuid.NewUUID()
	if err != nil {
		return now, uuid.Nil, fmt.Errorf("%w: %w", ErrUUID, err)
	}
	return now, uid, nil
}
