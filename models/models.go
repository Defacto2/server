// Package models contain the custom queries for the database that are not available using the ORM,
// as well as methods to interact with the query data.
package models

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/Defacto2/server/postgres/models"
	"github.com/volatiletech/null/v8"
	. "github.com/volatiletech/sqlboiler/v4/queries/qm"
)

// https://github.com/volatiletech/sqlboiler#constants

type Count int // Count is the number of found files.

const (
	Art  int = iota // Art are digital + pixel art files.
	Doc             // Doc are document + text art files.
	Soft            // Soft are software files.
)

const (
	// Name
	NameAsc = "C=N&O=A"
	NameDes = "C=N&O=D"
	// Date published
	PublAsc = "C=D&O=A"
	PublDes = "C=D&O=D"
	// Posted
	PostAsc = "C=P&O=A"
	PostDes = "C=P&O=D"
	// Size
	SizeAsc = "C=S&O=A"
	SizeDes = "C=S&O=D"
	// Description
	DescAsc = "C=I&O=A"
	DescDes = "C=I&O=D"
)

// Counts caches the number of found files fetched from SQL queries.
var Counts = map[int]Count{
	Art:  0,
	Doc:  0,
	Soft: 0,
}

// ArtImagesCount counts the number of files that could be classified as digital or pixel art.
func ArtImagesCount(ctx context.Context, db *sql.DB) (int, error) {
	if c := Counts[Art]; c > 0 {
		return int(c), nil
	}
	c, err := models.Files(
		Where("platform = ?", "image"),
		Where("section != ?", "bbs")).Count(ctx, db)
	if err != nil {
		return -1, err
	}
	Counts[Art] = Count(c)
	return int(c), nil
}

// ByteCountByCategory sums the byte filesizes for all the files that match a category.
func ByteCountByCategory(s string, ctx context.Context, db *sql.DB) (int64, error) {
	i, err := models.Files(
		SQL("SELECT sum(filesize) FROM files WHERE section = $1",
			null.StringFrom(s)),
	).Count(ctx, db)
	if err != nil {
		return 0, err
	}
	return i, err
}

// ByteCountByPlatform sums the byte filesizes for all the files that match a category.
func ByteCountByPlatform(s string, ctx context.Context, db *sql.DB) (int64, error) {
	i, err := models.Files(
		SQL("SELECT sum(filesize) FROM files WHERE platform = $1",
			null.StringFrom(s)),
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
	c, err := models.Files(
		Where("platform = ?", "ansi"),
		Or("platform = ?", "text"),
		Or("platform = ?", "textamiga"),
		Or("platform = ?", "pdf")).Count(ctx, db)
	if err != nil {
		return -1, err
	}
	Counts[Doc] = Count(c)
	return int(c), nil
}

func Download(id int, ctx context.Context, db *sql.DB) (*models.File, error) {
	file, err := models.Files(models.FileWhere.ID.EQ(int64(id))).One(ctx, db)
	if err != nil {
		return &models.File{}, err
	}
	return file, err
}

// FilesByCategory returns all the files that match a category tag.
func FilesByCategory(s, query string, ctx context.Context, db *sql.DB) (models.FileSlice, error) {
	x := null.StringFrom(s)
	return models.Files(Where("section = ?", x), OrderBy(Clauses(query))).All(ctx, db)
}

// FilesByPlatform returns all the files that match a platform tag.
func FilesByPlatform(s, query string, ctx context.Context, db *sql.DB) (models.FileSlice, error) {
	x := null.StringFrom(s)
	return models.Files(Where("platform = ?", x), OrderBy(Clauses(query))).All(ctx, db)
}

func Clauses(query string) string {
	const a, d = "asc", "desc"
	ca := models.FileColumns.Createdat
	dy := models.FileColumns.DateIssuedYear
	dm := models.FileColumns.DateIssuedMonth
	dd := models.FileColumns.DateIssuedDay
	fn := models.FileColumns.Filename
	fs := models.FileColumns.Filesize
	rt := models.FileColumns.RecordTitle
	switch strings.ToUpper(query) {
	case NameAsc:
		return fmt.Sprintf("%s %s", fn, a)
	case NameDes:
		return fmt.Sprintf("%s %s", fn, d)
	case PublAsc:
		return fmt.Sprintf("%s %s, %s %s, %s %s", dy, a, dm, a, dd, a)
	case PublDes:
		return fmt.Sprintf("%s %s, %s %s, %s %s", dy, d, dm, d, dd, d)
	case PostAsc:
		return fmt.Sprintf("%s %s", ca, a)
	case PostDes:
		return fmt.Sprintf("%s %s", ca, d)
	case SizeAsc:
		return fmt.Sprintf("%s %s", fs, a)
	case SizeDes:
		return fmt.Sprintf("%s %s", fs, d)
	case DescAsc:
		return fmt.Sprintf("%s %s", rt, a)
	case DescDes:
		return fmt.Sprintf("%s %s", rt, d)
	default:
		return fmt.Sprintf("%s %s", fn, a)
	}
}

// SoftwareCount counts the number of files that could be classified as software.
func SoftwareCount(ctx context.Context, db *sql.DB) (int, error) {
	if c := Counts[Soft]; c > 0 {
		return int(c), nil
	}
	c, err := models.Files(
		Where("platform = ?", "java"),
		Or("platform = ?", "linux"),
		Or("platform = ?", "dos"),
		Or("platform = ?", "php"),
		Or("platform = ?", "windows")).Count(ctx, db)
	if err != nil {
		return -1, err
	}
	Counts[Soft] = Count(c)
	return int(c), nil
}
