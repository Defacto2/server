// Package html3 is a sub-package of the model package that should only be used by the html3 handler.
// It contains the database queries for the HTML3 templates used to display the file lists in a table format.
//
//nolint:ireturn
package html3

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/Defacto2/helper"
	"github.com/Defacto2/server/handler/html3/ext"
	"github.com/Defacto2/server/internal/nils"
	"github.com/Defacto2/server/internal/postgres"
	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/Defacto2/server/model/querymod"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
)

const (
	// from is the name of the table containing records of files.
	from = "files"
	// clause is the clause to exclude soft deleted records.
	clause = "deletedat IS NULL"

	padding = " "
	pad3cnt = 3
	pad7cnt = 7
)

var ErrModel = errors.New("html3: no file model")

func pad3() string {
	return strings.Repeat(padding, pad3cnt)
}

func pad7() string {
	return strings.Repeat(padding, pad7cnt)
}

// ArtExpr returns a query modifier for the digital or pixel art category.
func ArtExpr() qm.QueryMod {
	return qm.Expr(
		qm.Where(clause),
		models.FileWhere.Section.NEQ(querymod.SBbs()),
		models.FileWhere.Platform.EQ(querymod.PImage()),
	)
}

// DocumentExpr returns a query modifier for the document category.
func DocumentExpr() qm.QueryMod { //nolint:ireturn
	return qm.Expr(
		qm.Where(clause),
		models.FileWhere.Platform.EQ(querymod.PAnsi()),
		qm.Or2(models.FileWhere.Platform.EQ(querymod.PText())),
		qm.Or2(models.FileWhere.Platform.EQ(querymod.PTextAmiga())),
		qm.Or2(models.FileWhere.Platform.EQ(querymod.PPdf())),
	)
}

// SelectExpr selects only the columns required by the HTML3 template.
func SelectExpr() qm.QueryMod { //nolint:ireturn
	return qm.Select(
		models.FileColumns.ID,
		models.FileColumns.Filename,
		models.FileColumns.DateIssuedDay,
		models.FileColumns.DateIssuedMonth,
		models.FileColumns.DateIssuedYear,
		models.FileColumns.Createdat,
		models.FileColumns.Filesize,
		models.FileColumns.Platform,
		models.FileColumns.Section,
		models.FileColumns.GroupBrandFor,
		models.FileColumns.RecordTitle,
	)
}

// SoftwareExpr returns a query modifier for the software category.
func SoftwareExpr() qm.QueryMod { //nolint:ireturn
	return qm.Expr(
		qm.Where(clause),
		models.FileWhere.Platform.EQ(querymod.PJava()),
		qm.Or2(models.FileWhere.Platform.EQ(querymod.PLinux())),
		qm.Or2(models.FileWhere.Platform.EQ(querymod.PDos())),
		qm.Or2(models.FileWhere.Platform.EQ(querymod.PScript())),
		qm.Or2(models.FileWhere.Platform.EQ(querymod.PWindows())),
	)
}

// Created returns the Createdat time to use a dd-mmm-yyyy format.
func Created(f *models.File) string {
	if f == nil {
		return fmt.Sprint(ErrModel)
	}

	const none = "-- --- ----"

	if !f.Createdat.Valid {
		return none
	}

	y := f.Createdat.Time.Year()
	if !helper.Year(y) {
		return none
	}
	d := f.Createdat.Time.Day()
	m := helper.ShortMonth(int(f.Createdat.Time.Month()))
	return fmt.Sprintf("%02d-%s-%d", d, m, y)
}

// Icon returns the extensionless name of a .gif image file to use as an icon
// for the filename. The icons are found in `public/image/html3/`.
func Icon(f *models.File) string {
	if f == nil {
		return fmt.Sprint(ErrModel)
	}

	const unknown = "unknown"
	if !f.Filename.Valid {
		return unknown
	}

	name := ext.IconName(f.Filename.String)
	if name == "" {
		return unknown
	}

	return name
}

// LeadStr takes a string and returns the leading whitespace padding, characters wide.
func LeadStr(width int, s string) string {
	n := utf8.RuneCountInString(s)
	if n >= width {
		return ""
	}

	const (
		use3 = 3
		use7 = 7
	)
	count := width - n
	switch count {
	case use3:
		return pad3()
	case use7:
		return pad7()
	default:
		return strings.Repeat(padding, count)
	}
}

// PublishedFW formats the publication year, month and day to a fixed-width length w value.
func PublishedFW(width int, f *models.File) string {
	s := Published(f)
	if utf8.RuneCountInString(s) < width {
		return LeadStr(width, s) + s
	}
	return s
}

// Published takes optional DateIssuedYear, DateIssuedMonth and DateIssuedDay values and
// formats them into dd-mmm-yyyy string format. Depending on the context, any missing time
// values will be left blank or replaced with "??" question marks.
func Published(f *models.File) string {
	if f == nil {
		return fmt.Sprint(ErrModel)
	}

	const (
		yx = "????"
		mx = "???"
		dx = "??"
	)
	ys, ms, ds := yx, mx, dx
	if f.DateIssuedYear.Valid {
		if i := int(f.DateIssuedYear.Int16); helper.Year(i) {
			ys = strconv.Itoa(i)
		}
	}
	if f.DateIssuedMonth.Valid {
		if s := helper.ShortMonth(int(f.DateIssuedMonth.Int16)); s != "" {
			ms = s
		}
	}
	if f.DateIssuedDay.Valid {
		if i := int(f.DateIssuedDay.Int16); helper.Day(i) {
			ds = fmt.Sprintf("%02d", i)
		}
	}

	yearOk := ys != yx
	monthOk := ms != mx
	dayOk := ds != dx

	switch {
	case yearOk && !monthOk && !dayOk:
		return pad7() + ys
	case yearOk && monthOk && !dayOk:
		return pad3() + ms + "-" + ys
	case !yearOk && !monthOk && !dayOk:
		return pad7() + yx
	default:
		return ds + "-" + ms + "-" + ys
	}
}

// Stat returns the SumSize and TotalCnt column selections.
func Stat() [2]string {
	return [...]string{postgres.SumSize, postgres.TotalCnt}
}

// StatContainer represents types that hold byte and count metrics.
type StatContainer interface {
	GetBytes() int
	SetBytes(n int)
	GetCount() int
	SetCount(n int)
}

// statQuery executes a statistics query using the provided expression.
func statQuery(ctx context.Context, exec boil.ContextExecutor, stats StatContainer, expr qm.QueryMod,
) error {
	if stats.GetBytes() > 0 && stats.GetCount() > 0 {
		return nil
	}

	c := Stat()
	columns := c[:]
	err := models.NewQuery(qm.Select(columns...), qm.Where(clause), expr, qm.From(from)).Bind(ctx, exec, stats)
	if err != nil {
		return fmt.Errorf("stat new query: %w", err)
	}

	return nil
}

// Arts statistics for releases that are digital or pixel art.
type Arts struct {
	Bytes int `boil:"size_total"`  // the total bytes of all the files
	Count int `boil:"count_total"` // the total number of files
}

// GetBytes returns the bytes count.
func (a *Arts) GetBytes() int { return a.Bytes }

// SetBytes sets the bytes count.
func (a *Arts) SetBytes(b int) { a.Bytes = b }

// GetCount returns the count.
func (a *Arts) GetCount() int { return a.Count }

// SetCount sets the count.
func (a *Arts) SetCount(c int) { a.Count = c }

// Stat sets the total bytes and total count.
func (a *Arts) Stat(ctx context.Context, exec boil.ContextExecutor) error {
	const format = "html3 arts statistics: %w"
	if err := nils.Check(ctx, exec); err != nil {
		return fmt.Errorf(format, err)
	}

	if err := statQuery(ctx, exec, a, ArtExpr()); err != nil {
		return fmt.Errorf(format, err)
	}

	return nil
}

// Documents statistics for releases that are documents.
type Documents struct {
	Bytes int `boil:"size_total"`  // the total bytes of all the files
	Count int `boil:"count_total"` // the total number of files
}

// GetBytes returns the bytes count.
func (d *Documents) GetBytes() int { return d.Bytes }

// SetBytes sets the bytes count.
func (d *Documents) SetBytes(b int) { d.Bytes = b }

// GetCount returns the count.
func (d *Documents) GetCount() int { return d.Count }

// SetCount sets the count.
func (d *Documents) SetCount(c int) { d.Count = c }

// Stat sets the total bytes and total count.
func (d *Documents) Stat(ctx context.Context, exec boil.ContextExecutor) error {
	const format = "html3 documents statistics: %w"
	if err := nils.Check(ctx, exec); err != nil {
		return fmt.Errorf(format, err)
	}

	if err := statQuery(ctx, exec, d, DocumentExpr()); err != nil {
		return fmt.Errorf(format, err)
	}

	return nil
}

// Softwares contain statistics for releases that are software.
type Softwares struct {
	Bytes int `boil:"size_total"`  // the total bytes of all the files
	Count int `boil:"count_total"` // the total number of files
}

// GetBytes returns the bytes count.
func (s *Softwares) GetBytes() int { return s.Bytes }

// SetBytes sets the bytes count.
func (s *Softwares) SetBytes(b int) { s.Bytes = b }

// GetCount returns the count.
func (s *Softwares) GetCount() int { return s.Count }

// SetCount sets the count.
func (s *Softwares) SetCount(c int) { s.Count = c }

// Stat sets the total bytes and total count.
func (s *Softwares) Stat(ctx context.Context, exec boil.ContextExecutor) error {
	const format = "html3 software statistics: %w"
	if err := nils.Check(ctx, exec); err != nil {
		return fmt.Errorf(format, err)
	}

	if err := statQuery(ctx, exec, s, SoftwareExpr()); err != nil {
		return fmt.Errorf(format, err)
	}

	return nil
}
