package model

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Defacto2/server/pkg/postgres/models"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
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

// All returns all the file records.
func (o Order) All(key int, ctx context.Context, db *sql.DB) (*models.FileSlice, error) {
	files, err := models.Files(qm.OrderBy(o.String())).All(ctx, db)
	if err != nil {
		return nil, err
	}
	return &files, err
}

// FilesByCategory returns all the files that match the named category.
func (o Order) FilesByCategory(name string, offset, limit int, ctx context.Context, db *sql.DB) (models.FileSlice, error) {
	x := null.StringFrom(name)
	if limit == 0 {
		return models.Files(models.FileWhere.Section.EQ(x), qm.OrderBy(o.String())).All(ctx, db)
	}
	return models.Files(models.FileWhere.Section.EQ(x), qm.OrderBy(o.String()),
		qm.Offset(calc(offset, limit)), qm.Limit(limit)).All(ctx, db)
}

// FilesByPlatform returns all the files that match the named platform.
func (o Order) FilesByPlatform(name string, ctx context.Context, db *sql.DB) (models.FileSlice, error) {
	x := null.StringFrom(name)
	return models.Files(models.FileWhere.Platform.EQ(x), qm.OrderBy(o.String())).All(ctx, db)
}

// FilesByGroup returns all the files that match an exact named group.
func (o Order) FilesByGroup(name string, ctx context.Context, db *sql.DB) (models.FileSlice, error) {
	x := null.StringFrom(name)
	//group_brand_for
	return models.Files(models.FileWhere.GroupBrandFor.EQ(x), qm.OrderBy(o.String())).All(ctx, db)
}

// ArtFiles returns all the files that could be considered as digital or pixel art.
func (o Order) ArtFiles(offset, limit int, ctx context.Context, db *sql.DB) (models.FileSlice, error) {
	if limit == 0 {
		return models.Files(SelectHTML3(), ArtExpr(), qm.OrderBy(o.String())).All(ctx, db)
	}
	return models.Files(SelectHTML3(), ArtExpr(), qm.OrderBy(o.String()),
		qm.Offset(calc(offset, limit)), qm.Limit(limit)).All(ctx, db)
}

// DocumentFiles returns all the files that that are considered to be documents.
func (o Order) DocumentFiles(offset, limit int, ctx context.Context, db *sql.DB) (models.FileSlice, error) {
	if limit == 0 {
		return models.Files(DocumentExpr(), qm.OrderBy(o.String())).All(ctx, db)
	}
	return models.Files(DocumentExpr(), qm.OrderBy(o.String()),
		qm.Offset(calc(offset, limit)), qm.Limit(limit)).All(ctx, db)
}

// SoftwareFiles returns all the files that that are considered to be software.
func (o Order) SoftwareFiles(offset, limit int, ctx context.Context, db *sql.DB) (models.FileSlice, error) {
	if limit == 0 {
		return models.Files(SoftwareExpr(), qm.OrderBy(o.String())).All(ctx, db)
	}
	return models.Files(SoftwareExpr(), qm.OrderBy(o.String()),
		qm.Offset(calc(offset, limit)), qm.Limit(limit)).All(ctx, db)
}

func calc(o, l int) int {
	if o < 1 {
		o = 1
	}
	return (o - 1) * l
}
