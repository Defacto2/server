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
	return orderClauses()[o]
}

// orderClauses returns a map of all the SQL, ORDER BY clauses.
func orderClauses() map[Order]string {
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

// One returns the record associated with the key ID.
func One(key int, ctx context.Context, db *sql.DB) (*models.File, error) {
	file, err := models.Files(models.FileWhere.ID.EQ(int64(key))).One(ctx, db)
	if err != nil {
		return nil, err
	}
	return file, err
}

// All returns all the file records.
func (o Order) All(key int, ctx context.Context, db *sql.DB) (*models.FileSlice, error) {
	files, err := models.Files(OrderBy(o.String())).All(ctx, db)
	if err != nil {
		return nil, err
	}
	return &files, err
}

// FilesByCategory returns all the files that match the named category.
func (o Order) FilesByCategory(name string, ctx context.Context, db *sql.DB) (models.FileSlice, error) {
	x := null.StringFrom(name)
	return models.Files(Where(section, x), OrderBy(o.String())).All(ctx, db)
}

// ByteCountByCategory sums the byte filesizes for all the files that match the category name.
func ByteCountByCategory(name string, ctx context.Context, db *sql.DB) (int64, error) {
	i, err := models.Files(
		SQL("SELECT sum(filesize) FROM files WHERE section = $1",
			null.StringFrom(name)),
	).Count(ctx, db)
	if err != nil {
		return 0, fmt.Errorf("bytecount by section %q: %w", name, err)
	}
	return i, nil
}

// FilesByPlatform returns all the files that match the named platform.
func (o Order) FilesByPlatform(name string, ctx context.Context, db *sql.DB) (models.FileSlice, error) {
	x := null.StringFrom(name)
	return models.Files(Where(platform, x), OrderBy(o.String())).All(ctx, db)
}

// ByteCountByPlatform sums the byte filesizes for all the files that match the category name.
func ByteCountByPlatform(name string, ctx context.Context, db *sql.DB) (int64, error) {
	i, err := models.Files(
		SQL("SELECT sum(filesize) FROM files WHERE platform = $1",
			null.StringFrom(name)),
	).Count(ctx, db)
	if err != nil {
		return 0, fmt.Errorf("bytecount by platform %q: %w", name, err)
	}
	return i, nil
}

// FilesByGroup returns all the files that match an exact named group.
func (o Order) FilesByGroup(name string, ctx context.Context, db *sql.DB) (models.FileSlice, error) {
	x := null.StringFrom(name)
	return models.Files(Where(groupFor, x), OrderBy(o.String())).All(ctx, db)
}

// ByteCountByGroup sums the byte filesizes for all the files that match the group name.
func ByteCountByGroup(name string, ctx context.Context, db *sql.DB) (int64, error) {
	x := null.StringFrom(name)
	i, err := models.Files(SQL("SELECT SUM(filesize) as size_sum FROM files WHERE group_brand_for = $1", x)).Count(ctx, db)
	if err != nil {
		return 0, fmt.Errorf("bytecount by group %q: %w", name, err)
	}
	return i, nil
}

// ArtFiles returns all the files that could be considered as digital or pixel art.
func (o Order) ArtFiles(ctx context.Context, db *sql.DB) (models.FileSlice, error) {
	return models.Files(ArtExpr(), OrderBy(o.String())).All(ctx, db)
}

// ArtCount counts the files that could be considered as digital or pixel art.
func ArtCount(ctx context.Context, db *sql.DB) (int, error) {
	if c := Counts[Art]; c > 0 {
		return int(c), nil
	}
	c, err := models.Files(ArtExpr()).Count(ctx, db)
	if err != nil {
		return -1, err
	}
	Counts[Art] = Count(c)
	return int(c), nil
}

// ArtByteCount sums the byte filesizes for all the files that is considered as digital or pixel art.
func ArtByteCount(ctx context.Context, db *sql.DB) (int64, error) {
	stmt := "SELECT SUM(files.filesize) AS size_sum FROM files WHERE" +
		fmt.Sprintf(" files.section != '%s'", tags.BBS) +
		fmt.Sprintf(" AND files.platform = '%s';", tags.Image)
	return models.Files(SQL(stmt)).Count(ctx, db)
}

// ArtExpr is a the query mod expression for art files.
func ArtExpr() QueryMod {
	bbs := null.String{String: tags.URIs[tags.BBS], Valid: true}
	image := null.String{String: tags.URIs[tags.Image], Valid: true}
	return Expr(
		models.FileWhere.Section.NEQ(bbs),
		models.FileWhere.Platform.EQ(image),
	)
}

// DocumentFiles returns all the files that that are considered to be documents.
func (o Order) DocumentFiles(ctx context.Context, db *sql.DB) (models.FileSlice, error) {
	return models.Files(DocumentExpr(), OrderBy(o.String())).All(ctx, db)
}

// DocumentByteCount sums the byte filesizes for all the files that are considered to be documents.
func DocumentByteCount(ctx context.Context, db *sql.DB) (int64, error) {
	stmt := "SELECT SUM(files.filesize) AS size_sum FROM files WHERE " +
		fmt.Sprintf("platform = '%s'", tags.ANSI) +
		fmt.Sprintf("OR platform = '%s'", tags.Text) +
		fmt.Sprintf("OR platform = '%s'", tags.TextAmiga) +
		fmt.Sprintf("OR platform = '%s'", tags.PDF)
	return models.Files(SQL(stmt)).Count(ctx, db)
}

// DocumentCount counts the number of files that are considered to be documents.
func DocumentCount(ctx context.Context, db *sql.DB) (int, error) {
	if c := Counts[Doc]; c > 0 {
		return int(c), nil
	}
	c, err := models.Files(DocumentExpr()).Count(ctx, db)
	if err != nil {
		return -1, err
	}
	Counts[Doc] = Count(c)
	return int(c), nil
}

// DocumentExpr is a the query mod expression for document files.
func DocumentExpr() QueryMod {
	ansi := null.String{String: tags.URIs[tags.ANSI], Valid: true}
	text := null.String{String: tags.URIs[tags.Text], Valid: true}
	amiga := null.String{String: tags.URIs[tags.TextAmiga], Valid: true}
	pdf := null.String{String: tags.URIs[tags.PDF], Valid: true}
	return Expr(
		models.FileWhere.Platform.EQ(ansi),
		Or2(models.FileWhere.Platform.EQ(text)),
		Or2(models.FileWhere.Platform.EQ(amiga)),
		Or2(models.FileWhere.Platform.EQ(pdf)),
	)
}

// SoftwareFiles returns all the files that that are considered to be software.
func (o Order) SoftwareFiles(ctx context.Context, db *sql.DB) (models.FileSlice, error) {
	return models.Files(SoftwareExpr(), OrderBy(o.String())).All(ctx, db)
}

// SoftwareCount counts the number of files that are considered to be software.
func SoftwareCount(ctx context.Context, db *sql.DB) (int, error) {
	if c := Counts[Soft]; c > 0 {
		return int(c), nil
	}
	c, err := models.Files(SoftwareExpr()).Count(ctx, db)
	if err != nil {
		return -1, err
	}
	Counts[Soft] = Count(c)
	return int(c), nil
}

// SoftwareByteCount sums the byte filesizes for all the files that are considered to be software.
func SoftwareByteCount(ctx context.Context, db *sql.DB) (int64, error) {
	stmt := "SELECT SUM(files.filesize) AS size_sum FROM files WHERE " +
		fmt.Sprintf("platform = '%s'", tags.Java) +
		fmt.Sprintf("OR platform = '%s'", tags.Linux) +
		fmt.Sprintf("OR platform = '%s'", tags.DOS) +
		fmt.Sprintf("OR platform = '%s'", tags.PHP) +
		fmt.Sprintf("OR platform = '%s'", tags.Windows)
	return models.Files(SQL(stmt)).Count(ctx, db)
}

// SoftwareExpr is a the query mod expression for software files.
func SoftwareExpr() QueryMod {
	java := null.String{String: tags.URIs[tags.Java], Valid: true}
	linux := null.String{String: tags.URIs[tags.Linux], Valid: true}
	dos := null.String{String: tags.URIs[tags.DOS], Valid: true}
	php := null.String{String: tags.URIs[tags.PHP], Valid: true}
	windows := null.String{String: tags.URIs[tags.Windows], Valid: true}
	return Expr(
		models.FileWhere.Platform.EQ(java),
		Or2(models.FileWhere.Platform.EQ(linux)),
		Or2(models.FileWhere.Platform.EQ(dos)),
		Or2(models.FileWhere.Platform.EQ(php)),
		Or2(models.FileWhere.Platform.EQ(windows)),
	)
}
