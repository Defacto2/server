package model

import (
	"context"
	"crypto/sha512"
	"database/sql"
	"encoding/hex"
	"fmt"
	"mime"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/Defacto2/releaser"
	"github.com/Defacto2/server/internal/demozoo"
	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/Defacto2/server/internal/tags"
	"github.com/google/uuid"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

const (
	shortLimit   = 100
	longFilename = 255
)

// uuidV7 generates a new UUID version 7, if that fails then it will fallback to version 1.
// It also returns the current time.
func uuidV7() (time.Time, uuid.UUID, error) {
	now := time.Now()
	uid, err := uuid.NewV7()
	if err == nil {
		return now, uid, nil
	}
	uid, err = uuid.NewUUID()
	if err != nil {
		return now, uuid.Nil, fmt.Errorf("%w: %s", ErrUUID, err)
	}
	return now, uid, nil
}

// dateIssue returns a valid year, month and day or a null value.
func dateIssue(y, m, d string) (null.Int16, null.Int16, null.Int16) {
	const base, bitSize = 10, 16
	i, _ := strconv.ParseInt(y, base, bitSize)
	year := ValidY(int16(i))

	i, _ = strconv.ParseInt(m, base, bitSize)
	month := ValidM(int16(i))

	i, _ = strconv.ParseInt(d, base, bitSize)
	day := ValidD(int16(i))

	return year, month, day
}

// ValidD returns a valid day or a null value.
func ValidD(d int16) null.Int16 {
	const first, last = 1, 31
	if d < first || d > last {
		return null.Int16{Int16: 0, Valid: false}
	}
	return null.Int16{Int16: d, Valid: true}
}

// ValidM returns a valid month or a null value.
func ValidM(m int16) null.Int16 {
	const jan, dec = 1, 12
	if m < jan || m > dec {
		return null.Int16{Int16: 0, Valid: false}
	}
	return null.Int16{Int16: m, Valid: true}
}

// ValidY returns a valid year or a null value.
func ValidY(y int16) null.Int16 {
	current := int16(time.Now().Year())
	if y < EpochYear || y > current {
		return null.Int16{Int16: 0, Valid: false}
	}
	return null.Int16{Int16: y, Valid: true}
}

// trimShort returns a string that is no longer than the short limit.
// It will also remove any leading or trailing white space.
func trimShort(s string) string {
	s = strings.TrimSpace(s)
	if len(s) > shortLimit {
		return s[:shortLimit]
	}
	return s
}

// trimName returns a string that is no longer than the long filename limit.
// It will also remove any leading or trailing white space.
func trimName(s string) string {
	s = strings.TrimSpace(s)
	if len(s) > longFilename {
		return s[:longFilename]
	}
	return s
}

// ValidReleasers returns two valid releaser group strings or null values.
func ValidReleasers(s1, s2 string) (null.String, null.String) {
	invalid := null.String{String: "", Valid: false}
	t1, t2 := trimShort(s1), trimShort(s2)
	t1, t2 = releaser.Clean(t1), releaser.Clean(t2)
	t1, t2 = strings.ToUpper(t1), strings.ToUpper(t2)
	x1, x2 := invalid, invalid
	if len(t1) > 0 {
		x1 = null.StringFrom(s1)
	}
	if len(t2) > 0 {
		x2 = null.StringFrom(s2)
	}
	return x1, x2
}

// ValidTitle returns a valid title or a null value.
// The title is trimmed and shortened to the short limit.
func ValidTitle(s string) null.String {
	invalid := null.String{String: "", Valid: false}
	t := trimShort(s)
	if len(t) == 0 {
		return invalid
	}
	return null.StringFrom(t)
}

// ValidYouTube returns true if the string is a valid YouTube video ID.
// An error is only returned if the regular expression match cannot compile.
func ValidYouTube(s string) (null.String, error) {
	const fixLen = 11
	invalid := null.String{String: "", Valid: false}
	if len(s) != fixLen {
		return invalid, nil
	}
	match, err := regexp.MatchString("^[a-zA-Z0-9_-]{11}$", s)
	if err != nil {
		return invalid, err
	}
	if !match {
		return invalid, nil
	}
	return null.String{String: s, Valid: true}, nil

}

// ValidFilename returns a valid filename or a null value.
// The filename is trimmed and shortened to the long filename limit.
func ValidFilename(s string) null.String {
	invalid := null.String{String: "", Valid: false}
	t := trimName(s)
	if len(t) == 0 {
		return invalid
	}
	return null.StringFrom(t)
}

// ValidMagic returns a valid media type or a null value.
// It is validated using the mime package.
// The media type is trimmed and validated using the mime package.
func ValidMagic(mediatype string) null.String {
	invalid := null.String{String: "", Valid: false}
	mtype := strings.TrimSpace(mediatype)
	if len(mtype) == 0 {
		return invalid
	}
	params := map[string]string{}
	result := mime.FormatMediaType(mtype, params)
	if result != "" {
		return invalid
	}
	return null.StringFrom(mtype)
}

// ValidFilesize returns a valid file size or an error.
// The file size is parsed as an unsigned integer.
// An error is returned if the string cannot be parsed as an integer.
func ValidFilesize(size string) (int64, error) {
	size = strings.TrimSpace(size)
	if len(size) == 0 {
		return 0, nil
	}
	s, err := strconv.ParseUint(size, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("%w: %q, %s", ErrSize, size, err)
	}
	return int64(s), nil
}

// ValidIntegrity confirms the integrity as a valid SHA-384 hexadecimal hash
// or returns a null value.
func ValidIntegrity(integrity string) null.String {
	invalid := null.String{String: "", Valid: false}
	if len(integrity) == 0 {
		return invalid
	}
	if len(integrity) != sha512.Size384*2 {
		return invalid
	}
	_, err := hex.DecodeString(integrity)
	if err != nil {
		return invalid
	}
	return null.StringFrom(integrity)
}

// ValidLastMod returns a valid last modified time or a null value.
// The lastmod time is parsed as a Unix time in milliseconds.
// An error is returned if the string cannot be parsed as an integer.
// The lastmod time is validated to be within the current year and the epoch year of 1980.
func ValidLastMod(lastmod string) null.Time {
	invalid := null.Time{Time: time.Time{}, Valid: false}
	if len(lastmod) == 0 {
		return invalid
	}
	i, err := strconv.ParseInt(lastmod, 10, 64)
	if err != nil {
		return invalid
	}
	val := time.UnixMilli(i)
	now := time.Now()
	if val.After(now) {
		return invalid
	}
	eposh := time.Date(EpochYear, time.January, 1, 0, 0, 0, 0, time.UTC)
	if val.Before(eposh) {
		return invalid
	}
	return null.TimeFrom(val)
}

// InsertUpload inserts a new file record into the database using a URL values map.
// This will not check if the file already exists in the database.
// Invalid values will be ignored, but will not prevent the record from being inserted.
// When successful the function will return the new record ID.
func InsertUpload(ctx context.Context, db *sql.DB, values url.Values) (int64, error) {
	if db == nil {
		return 0, ErrDB
	}

	// handle required table fields
	now, uid, err := uuidV7()
	if err != nil {
		return 0, err
	}
	uniqueID := null.StringFrom(uid.String())

	delTime := null.TimeFromPtr(&now)
	if !delTime.Valid || delTime.Time.IsZero() {
		return 0, fmt.Errorf("%w: %v", ErrTime, delTime.Time)
	}

	makeTime := null.TimeFromPtr(&now)
	if !makeTime.Valid || makeTime.Time.IsZero() {
		return 0, fmt.Errorf("%w: %v", ErrTime, makeTime.Time)
	}

	fname := ValidFilename(values.Get("filename"))
	if !fname.Valid || fname.IsZero() {
		return 0, fmt.Errorf("%w: %v", ErrName, "filename is required")
	}

	s := tags.Intro.String()
	var section null.String
	if tags.IsCategory(s) {
		section = null.StringFrom(s)
	}

	p := values.Get("platform")
	var platform null.String
	if tags.IsPlatform(p) {
		platform = null.StringFrom(p)
	}

	// handle optional table fields
	year, month, _ := dateIssue(values.Get("year"), values.Get("month"), "0")
	tube, err := ValidYouTube(values.Get("youtube"))
	if err != nil {
		return 0, err
	}
	rel1, rel2 := ValidReleasers(values.Get("group"), values.Get("brand"))
	title := ValidTitle(values.Get("title"))

	magic := ValidMagic(values.Get("magic"))

	size, err := ValidFilesize(values.Get("size"))
	if err != nil {
		return 0, err
	}

	integrity := ValidIntegrity(values.Get("integrity"))

	lastMod := ValidLastMod(values.Get("lastmod"))

	f := models.File{
		UUID:                uniqueID,
		Deletedat:           delTime,
		Createdat:           makeTime,
		WebIDYoutube:        tube,
		GroupBrandFor:       rel1,
		GroupBrandBy:        rel2,
		RecordTitle:         title,
		DateIssuedYear:      year,
		DateIssuedMonth:     month,
		Filename:            fname,
		Filesize:            size,
		FileMagicType:       magic,
		FileIntegrityStrong: integrity,
		FileLastModified:    lastMod,
		Platform:            platform,
		Section:             section,
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	if err = f.Insert(ctx, db, boil.Infer()); err != nil {
		return 0, err
	}
	if err = tx.Rollback(); err != nil {
		return 0, err
	}
	return f.ID, nil
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

	now, uid, err := uuidV7()
	if err != nil {
		return 0, err
	}

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

	now, uid, err := uuidV7()
	if err != nil {
		return 0, err
	}

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
