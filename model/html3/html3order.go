package html3

// Package html3Order.go contains the database queries the HTML3 order and sorting statements.

import (
	"context"
	"fmt"
	"strings"

	namer "github.com/Defacto2/releaser/name"
	"github.com/Defacto2/server/internal/panics"
	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
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

const all = 0 // all returns all the records.

// Art returns all the files that could be considered as digital or pixel art.
func (o Order) Art(ctx context.Context, exec boil.ContextExecutor, offset, limit int) (models.FileSlice, error) {
	const msg = "html3 all art"
	if panics.BoilExec(exec) {
		return nil, fmt.Errorf("%s: %w", msg, panics.ErrNoBoil)
	}
	if limit == all {
		return models.Files(
			SelectHTML3(),
			ArtExpr(),
			qm.Where(ClauseNoSoftDel),
			qm.OrderBy(o.String())).All(ctx, exec)
	}
	return models.Files(
		SelectHTML3(),
		ArtExpr(),
		qm.Where(ClauseNoSoftDel),
		qm.OrderBy(o.String()),
		qm.Offset(calc(offset, limit)),
		qm.Limit(limit)).All(ctx, exec)
}

// ByCategory returns all the files that match the named category.
func (o Order) ByCategory(
	ctx context.Context, exec boil.ContextExecutor, offset, limit int, name string) (
	models.FileSlice, error,
) {
	const msg = "html3 all by category"
	if panics.BoilExec(exec) {
		return nil, fmt.Errorf("%s: %w", msg, panics.ErrNoBoil)
	}
	mods := models.FileWhere.Section.EQ(null.StringFrom(name))
	if limit == all {
		return models.Files(mods,
			qm.Where(ClauseNoSoftDel),
			qm.OrderBy(o.String())).All(ctx, exec)
	}
	return models.Files(mods,
		qm.Where(ClauseNoSoftDel),
		qm.OrderBy(o.String()),
		qm.Offset(calc(offset, limit)),
		qm.Limit(limit)).All(ctx, exec)
}

// ByGroup returns all the files that match an exact named group.
func (o Order) ByGroup(ctx context.Context, exec boil.ContextExecutor, name string) (models.FileSlice, error) {
	const msg = "html3 all by group"
	if panics.BoilExec(exec) {
		return nil, fmt.Errorf("%s: %w", msg, panics.ErrNoBoil)
	}
	s, err := namer.Humanize(namer.Path(name))
	if err != nil {
		return nil, fmt.Errorf("order by group namer humanize: %w", err)
	}
	n := strings.ToUpper(s)
	mods := models.FileWhere.GroupBrandFor.EQ(null.StringFrom(n))
	return models.Files(mods,
		qm.Where(ClauseNoSoftDel),
		qm.OrderBy(o.String())).All(ctx, exec)
}

// ByPlatform returns all the files that match the named platform.
func (o Order) ByPlatform(
	ctx context.Context, exec boil.ContextExecutor, offset, limit int, name string) (
	models.FileSlice, error,
) {
	const msg = "html3 all by platform"
	if panics.BoilExec(exec) {
		return nil, fmt.Errorf("%s: %w", msg, panics.ErrNoBoil)
	}
	mods := models.FileWhere.Platform.EQ(null.StringFrom(name))
	if limit == all {
		return models.Files(mods,
			qm.Where(ClauseNoSoftDel),
			qm.OrderBy(o.String())).All(ctx, exec)
	}
	return models.Files(mods,
		qm.Where(ClauseNoSoftDel),
		qm.OrderBy(o.String()),
		qm.Offset(calc(offset, limit)),
		qm.Limit(limit)).All(ctx, exec)
}

// Document returns all the files that  are considered to be documents.
func (o Order) Document(ctx context.Context, exec boil.ContextExecutor, offset, limit int) (models.FileSlice, error) {
	const msg = "html3 all documents"
	if panics.BoilExec(exec) {
		return nil, fmt.Errorf("%s: %w", msg, panics.ErrNoBoil)
	}
	if limit == all {
		return models.Files(
			DocumentExpr(),
			qm.Where(ClauseNoSoftDel),
			qm.OrderBy(o.String())).All(ctx, exec)
	}
	return models.Files(
		DocumentExpr(),
		qm.Where(ClauseNoSoftDel),
		qm.OrderBy(o.String()),
		qm.Offset(calc(offset, limit)),
		qm.Limit(limit)).All(ctx, exec)
}

// Everything returns all of the file records.
func (o Order) Everything(ctx context.Context, exec boil.ContextExecutor, offset, limit int) (models.FileSlice, error) {
	const msg = "html3 everything"
	if panics.BoilExec(exec) {
		return nil, fmt.Errorf("%s: %w", msg, panics.ErrNoBoil)
	}
	return models.Files(
		qm.Where(ClauseNoSoftDel),
		qm.OrderBy(o.String()),
		qm.Offset(calc(offset, limit)),
		qm.Limit(limit)).All(ctx, exec)
}

// Software returns all the files that  are considered to be software.
func (o Order) Software(ctx context.Context, exec boil.ContextExecutor, offset, limit int) (models.FileSlice, error) {
	const msg = "html3 all software"
	if panics.BoilExec(exec) {
		return nil, fmt.Errorf("%s: %w", msg, panics.ErrNoBoil)
	}
	if limit == all {
		return models.Files(
			SoftwareExpr(),
			qm.Where(ClauseNoSoftDel),
			qm.OrderBy(o.String())).All(ctx, exec)
	}
	return models.Files(
		SoftwareExpr(),
		qm.Where(ClauseNoSoftDel),
		qm.OrderBy(o.String()),
		qm.Offset(calc(offset, limit)), qm.Limit(limit)).All(ctx, exec)
}

func (o Order) String() string {
	return orderClauses()[o]
}

// calc returns the offset value.
func calc(o, l int) int {
	if o < 1 {
		o = 1
	}
	return (o - 1) * l
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
	m[NameAsc] = fn + " " + a
	m[NameDes] = fn + " " + d
	m[PublAsc] = fmt.Sprintf("%s %s, %s %s, %s %s", dy, a, dm, a, dd, a)
	m[PublDes] = fmt.Sprintf("%s %s, %s %s, %s %s", dy, d, dm, d, dd, d)
	m[PostAsc] = ca + " " + a
	m[PostDes] = ca + " " + d
	m[SizeAsc] = fs + " " + a
	m[SizeDes] = fs + " " + d
	m[DescAsc] = rt + " " + a
	m[DescDes] = rt + " " + d
	return m
}
