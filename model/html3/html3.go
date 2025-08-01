// Package html3 is a sub-package of the model package that should only be used by the html3 handler.
// It contains the database queries for the HTML3 templates used to display the file lists in a table format.
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
	"github.com/Defacto2/server/internal/panics"
	"github.com/Defacto2/server/internal/postgres"
	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/Defacto2/server/model/querymod"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
)

const (
	// From is the name of the table containing records of files.
	From = "files"
	// ClauseNoSoftDel is the clause to exclude soft deleted records.
	ClauseNoSoftDel = "deletedat IS NULL"

	padding = " "
)

var ErrModel = errors.New("error, no file model")

// ArtExpr returns a query modifier for the digital or pixel art category.
func ArtExpr() qm.QueryMod {
	return qm.Expr(
		qm.Where(ClauseNoSoftDel),
		models.FileWhere.Section.NEQ(querymod.SBbs()),
		models.FileWhere.Platform.EQ(querymod.PImage()),
	)
}

// Created returns the Createdat time to use a dd-mmm-yyyy format.
func Created(f *models.File) string {
	if f == nil {
		return fmt.Sprint(ErrModel)
	}
	if !f.Createdat.Valid {
		return "-- --- ----"
	}
	d := f.Createdat.Time.Day()
	m := helper.ShortMonth(int(f.Createdat.Time.Month()))
	y := f.Createdat.Time.Year()
	if !helper.Year(y) {
		return "-- --- ----"
	}
	return fmt.Sprintf("%02d-%s-%d", d, m, y)
}

// DocumentExpr returns a query modifier for the document category.
func DocumentExpr() qm.QueryMod {
	return qm.Expr(
		qm.Where(ClauseNoSoftDel),
		models.FileWhere.Platform.EQ(querymod.PAnsi()),
		qm.Or2(models.FileWhere.Platform.EQ(querymod.PText())),
		qm.Or2(models.FileWhere.Platform.EQ(querymod.PTextAmiga())),
		qm.Or2(models.FileWhere.Platform.EQ(querymod.PPdf())),
	)
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
	if n := ext.IconName(f.Filename.String); n != "" {
		return n
	}
	return unknown
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
		return fmt.Sprint(ErrModel)
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
		qm.Where(ClauseNoSoftDel),
		models.FileWhere.Platform.EQ(querymod.PJava()),
		qm.Or2(models.FileWhere.Platform.EQ(querymod.PLinux())),
		qm.Or2(models.FileWhere.Platform.EQ(querymod.PDos())),
		qm.Or2(models.FileWhere.Platform.EQ(querymod.PScript())),
		qm.Or2(models.FileWhere.Platform.EQ(querymod.PWindows())),
	)
}

// Arts statistics for releases that are digital or pixel art.
type Arts struct {
	Bytes int `boil:"size_total"`  // the total bytes of all the files
	Count int `boil:"count_total"` // the total number of files
}

// Stat sets the total bytes and total count.
func (a *Arts) Stat(ctx context.Context, exec boil.ContextExecutor) error {
	const msg = "html3 arts statistics"
	if panics.BoilExec(exec) {
		return fmt.Errorf("%s: %w", msg, panics.ErrNoBoil)
	}
	if a.Bytes > 0 && a.Count > 0 {
		return nil
	}
	return models.NewQuery(
		qm.Select(postgres.Stat()...),
		qm.Where(ClauseNoSoftDel),
		ArtExpr(),
		qm.From(From)).Bind(ctx, exec, a)
}

// Documents statistics for releases that are documents.
type Documents struct {
	Bytes int `boil:"size_total"`  // the total bytes of all the files
	Count int `boil:"count_total"` // the total number of files
}

// Stat sets the total bytes and total count.
func (d *Documents) Stat(ctx context.Context, exec boil.ContextExecutor) error {
	const msg = "html3 documents statistics"
	if panics.BoilExec(exec) {
		return fmt.Errorf("%s: %w", msg, panics.ErrNoBoil)
	}
	if d.Bytes > 0 && d.Count > 0 {
		return nil
	}
	return models.NewQuery(
		qm.Select(postgres.Stat()...),
		qm.Where(ClauseNoSoftDel),
		DocumentExpr(),
		qm.From(From)).Bind(ctx, exec, d)
}

// Softwares contain statistics for releases that are software.
type Softwares struct {
	Bytes int `boil:"size_total"`  // the total bytes of all the files
	Count int `boil:"count_total"` // the total number of files
}

// Stat sets the total bytes and total count.
func (s *Softwares) Stat(ctx context.Context, exec boil.ContextExecutor) error {
	const msg = "html3 software statistics"
	if panics.BoilExec(exec) {
		return fmt.Errorf("%s: %w", msg, panics.ErrNoBoil)
	}
	if s.Bytes > 0 && s.Count > 0 {
		return nil
	}
	return models.NewQuery(
		qm.Select(postgres.Stat()...),
		qm.Where(ClauseNoSoftDel),
		SoftwareExpr(),
		qm.From(From)).Bind(ctx, exec, s)
}
