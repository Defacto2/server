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

// DateIssue returns a valid year, month and day or a null value.
func DateIssue(y, m, d string) (null.Int16, null.Int16, null.Int16) {
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

// TrimShort returns a string that is no longer than the short limit.
// It will also remove any leading or trailing white space.
func TrimShort(s string) string {
	x := strings.TrimSpace(s)
	if len(x) > shortLimit {
		return x[:shortLimit]
	}
	return x
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
	t1, t2 := TrimShort(s1), TrimShort(s2)
	t1, t2 = releaser.Clean(t1), releaser.Clean(t2)
	t1, t2 = strings.ToUpper(t1), strings.ToUpper(t2)
	x1, x2 := invalid, invalid
	if len(t1) > 0 {
		x1 = null.StringFrom(t1)
	}
	if len(t2) > 0 {
		x2 = null.StringFrom(t2)
	}
	return x1, x2
}

// ValidSceners returns a valid sceners string or a null value.
func ValidSceners(s string) null.String {
	invalid := null.String{String: "", Valid: false}
	t := TrimShort(s)
	if len(t) == 0 {
		return invalid
	}
	const sep = ","
	ts := strings.Split(t, sep)
	for i, v := range ts {
		ts[i] = releaser.Clean(strings.TrimSpace(v))
	}
	t = strings.Join(ts, sep)
	return null.StringFrom(t)
}

// ValidTitle returns a valid title or a null value.
// The title is trimmed and shortened to the short limit.
func ValidTitle(s string) null.String {
	invalid := null.String{String: "", Valid: false}
	t := TrimShort(s)
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
	if result == "" {
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

// ValidPlatform returns a valid platform or a null value.
func ValidPlatform(platform string) null.String {
	invalid := null.String{String: "", Valid: false}
	p := strings.TrimSpace(platform)
	if tags.IsPlatform(p) {
		return null.StringFrom(p)
	}
	return invalid
}

// ValidSection returns a valid section or a null value.
func ValidSection(section string) null.String {
	invalid := null.String{String: "", Valid: false}
	s := strings.TrimSpace(section)
	if tags.IsCategory(s) {
		return null.StringFrom(s)
	}
	return invalid

}

// ValidString returns a valid string or a null value.
func ValidString(s string) null.String {
	invalid := null.String{String: "", Valid: false}
	x := strings.TrimSpace(s)
	if len(x) == 0 {
		return invalid
	}
	return null.StringFrom(x)
}

// InsertUpload inserts a new file record into the database using a URL values map.
// This will not check if the file already exists in the database.
// Invalid values will be ignored, but will not prevent the record from being inserted.
// When successful the function will return the new record ID.
func InsertUpload(ctx context.Context, tx *sql.Tx, values url.Values, key string) (int64, error) {
	if tx == nil {
		return 0, ErrDB
	}
	now, uid, err := uuidV7()
	if err != nil {
		return 0, err
	}
	unique := null.StringFrom(uid.String())
	deleteT := null.TimeFromPtr(&now)
	if !deleteT.Valid || deleteT.Time.IsZero() {
		return 0, fmt.Errorf("%w: %v", ErrTime, deleteT.Time)
	}
	createT := null.TimeFromPtr(&now)
	if !createT.Valid || createT.Time.IsZero() {
		return 0, fmt.Errorf("%w: %v", ErrTime, createT.Time)
	}
	youtube, err := ValidYouTube(values.Get(key + "-youtube"))
	if err != nil {
		return 0, err
	}
	releaser1, releaser2 := ValidReleasers(
		values.Get(key+"-releaser1"),
		values.Get(key+"-releaser2"),
	)
	title := ValidTitle(values.Get(key + "-title"))
	year, month, day := DateIssue(
		values.Get(key+"-year"),
		values.Get(key+"-month"),
		values.Get(key+"-day"),
	)
	fname := values.Get(key + "-filename")
	filename := ValidFilename(fname)
	if !filename.Valid || filename.IsZero() {
		return 0, fmt.Errorf("%w: %v", ErrName, key+"-filename is required")
	}
	filesize, err := ValidFilesize(values.Get(key + "-size"))
	if err != nil {
		return 0, err
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

	f := models.File{
		UUID:                unique,
		Deletedat:           deleteT,
		Createdat:           createT,
		WebIDYoutube:        youtube,
		GroupBrandFor:       releaser1,
		GroupBrandBy:        releaser2,
		RecordTitle:         title,
		DateIssuedYear:      year,
		DateIssuedMonth:     month,
		DateIssuedDay:       day,
		Filename:            filename,
		Filesize:            filesize,
		FileZipContent:      content,
		RetrotxtReadme:      readme,
		FileMagicType:       filemagic,
		FileIntegrityStrong: integrity,
		FileLastModified:    lastMod,
		Platform:            platform,
		Section:             section,
		CreditText:          creditT,
		CreditIllustration:  creditI,
		CreditProgram:       creditP,
		CreditAudio:         creditA,
	}
	if err = f.Insert(ctx, tx, boil.Infer()); err != nil {
		return 0, err
	}
	if err = tx.Commit(); err != nil {
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
