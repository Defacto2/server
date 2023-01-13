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

type Count int

const (
	Art int = iota
	Doc
	Soft
)

var Counts = map[int]Count{
	Art:  0,
	Doc:  0,
	Soft: 0,
}

func FilesByCategory(s string, ctx context.Context, db *sql.DB) (models.FileSlice, error) {
	x := null.StringFrom(s)
	y, err := models.Files(models.FileWhere.Section.EQ(x)).All(ctx, db)
	for i, z := range y {
		fmt.Println("->", i, "==>", z.Filename.String, z.DateIssuedYear, z.Createdat, z.Filesize.Int64, z.RecordTitle)
		fmt.Printf("%T", z.Createdat)
	}
	return y, err
}

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
