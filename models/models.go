package models

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/bengarrett/df2023/postgres/models"
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

// Counts caches the number of found files fetched from SQL queries.
var Counts = map[int]Count{
	Art:  0,
	Doc:  0,
	Soft: 0,
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

// FilesByCategory returns all the files that match a category.
func FilesByCategory(s string, ctx context.Context, db *sql.DB) (models.FileSlice, error) {
	x := null.StringFrom(s)
	y, err := models.Files(models.FileWhere.Section.EQ(x)).All(ctx, db)
	for i, z := range y {
		fmt.Println("->", i, "==>", z.Filename.String, z.DateIssuedYear, z.Createdat, z.Filesize.Int64, z.RecordTitle)
		fmt.Printf("%T", z.Createdat)
	}
	return y, err
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
