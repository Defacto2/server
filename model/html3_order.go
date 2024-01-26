package model

// Package html3_order.go contains the database queries the HTML3 order and sorting statements.

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	namer "github.com/Defacto2/releaser/name"
	"github.com/Defacto2/server/internal/postgres/models"
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

	all = 0 // all returns all the records.
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
	m := make(map[Order]string, DescDes+1)
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

// AllFiles returns all of the file records.
func (o Order) AllFiles(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	if db == nil {
		return nil, ErrDB
	}
	return models.Files(
		qm.OrderBy(o.String()),
		qm.Offset(calc(offset, limit)),
		qm.Limit(limit)).All(ctx, db)
}

// FilesByCategory returns all the files that match the named category.
func (o Order) FilesByCategory(
	ctx context.Context, db *sql.DB, offset, limit int, name string) (
	models.FileSlice, error,
) {
	if db == nil {
		return nil, ErrDB
	}
	mods := models.FileWhere.Section.EQ(null.StringFrom(name))
	if limit == all {
		return models.Files(mods,
			qm.OrderBy(o.String())).All(ctx, db)
	}
	return models.Files(mods,
		qm.OrderBy(o.String()),
		qm.Offset(calc(offset, limit)),
		qm.Limit(limit)).All(ctx, db)
}

// FilesByPlatform returns all the files that match the named platform.
func (o Order) FilesByPlatform(ctx context.Context, db *sql.DB, name string) (models.FileSlice, error) {
	if db == nil {
		return nil, ErrDB
	}
	mods := models.FileWhere.Platform.EQ(null.StringFrom(name))
	return models.Files(mods,
		qm.OrderBy(o.String())).All(ctx, db)
}

// FilesByGroup returns all the files that match an exact named group.
func (o Order) FilesByGroup(ctx context.Context, db *sql.DB, name string) (models.FileSlice, error) {
	if db == nil {
		return nil, ErrDB
	}
	s, err := namer.Humanize(namer.Path(name))
	if err != nil {
		return nil, err
	}
	n := strings.ToUpper(s)
	mods := models.FileWhere.GroupBrandFor.EQ(null.StringFrom(n))
	return models.Files(mods,
		qm.OrderBy(o.String())).All(ctx, db)
}

// ArtFiles returns all the files that could be considered as digital or pixel art.
func (o Order) ArtFiles(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	if db == nil {
		return nil, ErrDB
	}
	if limit == all {
		return models.Files(
			SelectHTML3(),
			ArtExpr(),
			qm.OrderBy(o.String())).All(ctx, db)
	}
	return models.Files(
		SelectHTML3(),
		ArtExpr(),
		qm.OrderBy(o.String()),
		qm.Offset(calc(offset, limit)),
		qm.Limit(limit)).All(ctx, db)
}

// DocumentFiles returns all the files that  are considered to be documents.
func (o Order) DocumentFiles(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	if db == nil {
		return nil, ErrDB
	}
	if limit == all {
		return models.Files(
			DocumentExpr(),
			qm.OrderBy(o.String())).All(ctx, db)
	}
	return models.Files(
		DocumentExpr(),
		qm.OrderBy(o.String()),
		qm.Offset(calc(offset, limit)),
		qm.Limit(limit)).All(ctx, db)
}

// SoftwareFiles returns all the files that  are considered to be software.
func (o Order) SoftwareFiles(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	if db == nil {
		return nil, ErrDB
	}
	if limit == all {
		return models.Files(
			SoftwareExpr(),
			qm.OrderBy(o.String())).All(ctx, db)
	}
	return models.Files(
		SoftwareExpr(),
		qm.OrderBy(o.String()),
		qm.Offset(calc(offset, limit)), qm.Limit(limit)).All(ctx, db)
}

func calc(o, l int) int {
	if o < 1 {
		o = 1
	}
	return (o - 1) * l
}
