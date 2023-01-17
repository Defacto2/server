// Package models contain the custom queries for the database that are not available using the ORM,
// as well as methods to interact with the query data.
package models

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Defacto2/server/postgres/models"
	"github.com/Defacto2/server/tags"
	"github.com/volatiletech/null/v8"
	. "github.com/volatiletech/sqlboiler/v4/queries/qm"
)

// https://github.com/volatiletech/sqlboiler#constants

type Count int // Count is the number of found files.

// Scener contains the usable data for a group or person.
type Scener struct {
	URI   string // URI slug for the scener.
	Name  string // Name to display.
	Count int    // Count the records associated with the scene.
}

// Sceners is a collection of sceners.
type Sceners map[string]Scener

// Counts caches the number of found files fetched from SQL queries.
var Counts = map[int]Count{
	Art:  0,
	Doc:  0,
	Soft: 0,
}

const (
	Art  int = iota // Art are digital + pixel art files.
	Doc             // Doc are document + text art files.
	Soft            // Soft are software files.

	groupFor   = "group_brand_for = ?"
	section    = "section = ?"
	notSection = "section != ?"
	platform   = "platform = ?"
)

// Order the query using a table column.
type Order int

const (
	NameAsc Order = iota // NameAsc order the ascending query using the filename.
	NameDes              // NameDes order the descending query using the filename.
	PublAsc              // PublAsc order the ascending query using the date published.
	PublDes              // PublDes order the descending query using the date published.
	PostAsc              // PostAsc order the ascending query using the date posted.
	PostDes              // PostDes order the descending query using the date posted.
	SizeAsc              // SizeAsc order the ascending query using the file size.
	SizeDes              // SizeDes order the descending query using the file size.
	DescAsc              // DescAsc order the ascending query using the record title.
	DescDes              // DescDes order the descending query using the record title.
)

func (o Order) String() string {
	return OrderClauses()[o]
}

// FilesByCategory returns all the files that match the named category.
func (o Order) FilesByCategory(name string, ctx context.Context, db *sql.DB) (models.FileSlice, error) {
	x := null.StringFrom(name)
	return models.Files(Where(section, x), OrderBy(o.String())).All(ctx, db)
}

// FilesByPlatform returns all the files that match the named platform.
func (o Order) FilesByPlatform(name string, ctx context.Context, db *sql.DB) (models.FileSlice, error) {
	x := null.StringFrom(name)
	return models.Files(Where(platform, x), OrderBy(o.String())).All(ctx, db)
}

// FilesByGroup returns all the files that match an exact named group.
func (o Order) FilesByGroup(name string, ctx context.Context, db *sql.DB) (models.FileSlice, error) {
	x := null.StringFrom(name)
	return models.Files(Where(groupFor, x), OrderBy(o.String())).All(ctx, db)
}

// ArtImagesCount counts the number of files that could be classified as digital or pixel art.
func ArtImagesCount(ctx context.Context, db *sql.DB) (int, error) {
	if c := Counts[Art]; c > 0 {
		return int(c), nil
	}
	bbs := tags.URIs[tags.BBS]
	image := tags.URIs[tags.Image]
	c, err := models.Files(
		Where(platform, image),
		Where(notSection, bbs)).Count(ctx, db)
	if err != nil {
		return -1, err
	}
	Counts[Art] = Count(c)
	return int(c), nil
}

// ByteCountByCategory sums the byte filesizes for all the files that match the category name.
func ByteCountByCategory(name string, ctx context.Context, db *sql.DB) (int64, error) {
	i, err := models.Files(
		SQL("SELECT sum(filesize) FROM files WHERE section = $1",
			null.StringFrom(name)),
	).Count(ctx, db)
	if err != nil {
		return 0, err
	}
	return i, err
}

// ByteCountByPlatform sums the byte filesizes for all the files that match the category name.
func ByteCountByPlatform(name string, ctx context.Context, db *sql.DB) (int64, error) {
	i, err := models.Files(
		SQL("SELECT sum(filesize) FROM files WHERE platform = $1",
			null.StringFrom(name)),
	).Count(ctx, db)
	if err != nil {
		return 0, err
	}
	return i, err
}

// DocumentCount counts the number of files that could be classified as document or text art.
func DocumentCount(ctx context.Context, db *sql.DB) (int, error) {
	if c := Counts[Doc]; c > 0 {
		return int(c), nil
	}
	ansi := tags.URIs[tags.ANSI]
	text := tags.URIs[tags.Text]
	amiga := tags.URIs[tags.TextAmiga]
	pdf := tags.URIs[tags.PDF]
	c, err := models.Files(
		Where(platform, ansi), Or(platform, text), Or(platform, amiga), Or(platform, pdf)).Count(ctx, db)
	if err != nil {
		return -1, err
	}
	Counts[Doc] = Count(c)
	return int(c), nil
}

// OrderClauses returns a map of all the SQL, ORDER BY clauses.
func OrderClauses() map[Order]string {
	const a, d = "asc", "desc"
	ca := models.FileColumns.Createdat
	dy := models.FileColumns.DateIssuedYear
	dm := models.FileColumns.DateIssuedMonth
	dd := models.FileColumns.DateIssuedDay
	fn := models.FileColumns.Filename
	fs := models.FileColumns.Filesize
	rt := models.FileColumns.RecordTitle
	var m = make(map[Order]string, DescDes+1)
	m[NameAsc] = fmt.Sprintf("%s %s", fn, a)
	m[NameDes] = fmt.Sprintf("%s %s", fn, d)
	m[PublAsc] = fmt.Sprintf("%s %s, %s %s, %s %s", dy, a, dm, a, dd, a)
	m[PublDes] = fmt.Sprintf("%s %s, %s %s, %s %s", dy, d, dm, d, dd, d)
	m[PostAsc] = fmt.Sprintf("%s %s", ca, a)
	m[PostDes] = fmt.Sprintf("%s %s", ca, d)
	m[SizeAsc] = fmt.Sprintf("%s %s", fs, a)
	m[SizeDes] = fmt.Sprintf("%s %s", fs, d)
	m[DescAsc] = fmt.Sprintf("%s %s", rt, a)
	m[DescDes] = fmt.Sprintf("%s %s", rt, d)
	return m
}

// File returns the record associated with the key ID.
func File(key int, ctx context.Context, db *sql.DB) (*models.File, error) {
	file, err := models.Files(models.FileWhere.ID.EQ(int64(key))).One(ctx, db)
	if err != nil {
		return &models.File{}, err
	}
	return file, err
}

// SoftwareCount counts the number of files that could be classified as software.
func SoftwareCount(ctx context.Context, db *sql.DB) (int, error) {
	if c := Counts[Soft]; c > 0 {
		return int(c), nil
	}
	java := tags.URIs[tags.PDF]
	linux := tags.URIs[tags.PDF]
	dos := tags.URIs[tags.PDF]
	php := tags.URIs[tags.PDF]
	windows := tags.URIs[tags.PDF]
	c, err := models.Files(
		Where(platform, java),
		Or(platform, linux),
		Or(platform, dos),
		Or(platform, php),
		Or(platform, windows)).Count(ctx, db)
	if err != nil {
		return -1, err
	}
	Counts[Soft] = Count(c)
	return int(c), nil
}
