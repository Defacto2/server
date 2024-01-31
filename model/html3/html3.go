// Package htm3 is a sub-package of the model package that should only be used by the html3 handler.
// It contains the database queries for the HTML3 templates used to display the file lists in a table format.
package html3

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/Defacto2/server/internal/exts"
	"github.com/Defacto2/server/internal/helper"
	"github.com/Defacto2/server/internal/postgres"
	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/Defacto2/server/model/expr"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

const (
	From = "files" // the name of the table containing records of files
)

var (
	ErrDB    = fmt.Errorf("database value is nil")
	ErrModel = fmt.Errorf("error, no file model")
)

const padding = " "

// ArtExpr returns a query modifier for the digital or pixel art category.
func ArtExpr() qm.QueryMod {
	return qm.Expr(
		models.FileWhere.Section.NEQ(expr.SBbs()),
		models.FileWhere.Platform.EQ(expr.PImage()),
	)
}

// Created returns the Createdat time to use a dd-mmm-yyyy format.
func Created(f *models.File) string {
	if f == nil {
		return ErrModel.Error()
	}
	if !f.Createdat.Valid {
		return "-- --- ----"
	}
	d := f.Createdat.Time.Day()
	m := helper.ShortMonth(int(f.Createdat.Time.Month()))
	y := f.Createdat.Time.Year()
	if !helper.IsYear(y) {
		return "-- --- ----"
	}
	return fmt.Sprintf("%02d-%s-%d", d, m, y)
}

// DocumentExpr returns a query modifier for the document category.
func DocumentExpr() qm.QueryMod {
	return qm.Expr(
		models.FileWhere.Platform.EQ(expr.PAnsi()),
		qm.Or2(models.FileWhere.Platform.EQ(expr.PText())),
		qm.Or2(models.FileWhere.Platform.EQ(expr.PTextAmiga())),
		qm.Or2(models.FileWhere.Platform.EQ(expr.PPdf())),
	)
}

// Icon returns the extensionless name of a .gif image file to use as an icon
// for the filename. The icons are found in `public/image/html3/`.
func Icon(f *models.File) string {
	if f == nil {
		return ErrModel.Error()
	}
	const err = "unknown"
	if !f.Filename.Valid {
		return err
	}
	if n := exts.IconName(f.Filename.String); n != "" {
		return n
	}
	return err
}

// LeadStr takes a string and returns the leading whitespace padding, characters wide.
func LeadStr(width int, s string) string {
	l := utf8.RuneCountInString(s)
	if l >= width {
		return ""
	}
	return strings.Repeat(padding, width-l)
}

// Published takes optional DateIssuedYear, DateIssuedMonth and DateIssuedDay values and
// formats them into dd-mmm-yyyy string format. Depending on the context, any missing time
// values will be left blank or replaced with ?? question marks.
func Published(f *models.File) string {
	if f == nil {
		return ErrModel.Error()
	}
	const (
		yx       = "????"
		mx       = "???"
		dx       = "??"
		sp       = " "
		yPadding = 7
		dPadding = 3
	)
	ys, ms, ds := yx, mx, dx
	if f.DateIssuedYear.Valid {
		if i := int(f.DateIssuedYear.Int16); helper.IsYear(i) {
			ys = strconv.Itoa(i)
		}
	}
	if f.DateIssuedMonth.Valid {
		if s := helper.ShortMonth(int(f.DateIssuedMonth.Int16)); s != "" {
			ms = s
		}
	}
	if f.DateIssuedDay.Valid {
		if i := int(f.DateIssuedDay.Int16); helper.IsDay(i) {
			ds = fmt.Sprintf("%02d", i)
		}
	}
	if isYearOnly := ys != yx && ms == mx && ds == dx; isYearOnly {
		return fmt.Sprintf("%s%s", strings.Repeat(sp, yPadding), ys)
	}
	if isInvalidDay := ys != yx && ms != mx && ds == dx; isInvalidDay {
		return fmt.Sprintf("%s%s-%s", strings.Repeat(sp, dPadding), ms, ys)
	}
	if isInvalid := ys == yx && ms == mx && ds == dx; isInvalid {
		return fmt.Sprintf("%s%s", strings.Repeat(sp, yPadding), yx)
	}
	return fmt.Sprintf("%s-%s-%s", ds, ms, ys)
}

// PublishedFW formats the publication year, month and day to a fixed-width length w value.
func PublishedFW(width int, f *models.File) string {
	s := Published(f)
	if utf8.RuneCountInString(s) < width {
		return LeadStr(width, s) + s
	}
	return s
}

// SelectHTML3 selects only the columns required by the HTML3 template.
func SelectHTML3() qm.QueryMod {
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
func SoftwareExpr() qm.QueryMod {
	return qm.Expr(
		models.FileWhere.Platform.EQ(expr.PJava()),
		qm.Or2(models.FileWhere.Platform.EQ(expr.PLinux())),
		qm.Or2(models.FileWhere.Platform.EQ(expr.PDos())),
		qm.Or2(models.FileWhere.Platform.EQ(expr.PScript())),
		qm.Or2(models.FileWhere.Platform.EQ(expr.PWindows())),
	)
}

// Arts statistics for releases that are digital or pixel art.
type Arts struct {
	Bytes int `boil:"size_total"`  // the total bytes of all the files
	Count int `boil:"count_total"` // the total number of files
}

// Stat sets the total bytes and total count.
func (a *Arts) Stat(ctx context.Context, db *sql.DB) error {
	if db == nil {
		return ErrDB
	}
	if a.Bytes > 0 && a.Count > 0 {
		return nil
	}
	return models.NewQuery(
		qm.Select(postgres.Stat()...),
		ArtExpr(),
		qm.From(From)).Bind(ctx, db, a)
}

// Documents statistics for releases that are documents.
type Documents struct {
	Bytes int `boil:"size_total"`  // the total bytes of all the files
	Count int `boil:"count_total"` // the total number of files
}

// Stat sets the total bytes and total count.
func (d *Documents) Stat(ctx context.Context, db *sql.DB) error {
	if db == nil {
		return ErrDB
	}
	if d.Bytes > 0 && d.Count > 0 {
		return nil
	}
	return models.NewQuery(
		qm.Select(postgres.Stat()...),
		DocumentExpr(),
		qm.From(From)).Bind(ctx, db, d)
}

// Softwares contain statistics for releases that are software.
type Softwares struct {
	Bytes int `boil:"size_total"`  // the total bytes of all the files
	Count int `boil:"count_total"` // the total number of files
}

// Stat sets the total bytes and total count.
func (s *Softwares) Stat(ctx context.Context, db *sql.DB) error {
	if db == nil {
		return ErrDB
	}
	if s.Bytes > 0 && s.Count > 0 {
		return nil
	}
	return models.NewQuery(
		qm.Select(postgres.Stat()...),
		SoftwareExpr(),
		qm.From(From)).Bind(ctx, db, s)
}
