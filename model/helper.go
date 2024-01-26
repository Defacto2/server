package model

// Package file helper.go contains helper functions for the model package.

import (
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/Defacto2/server/internal/exts"
	"github.com/Defacto2/server/internal/helper"
	"github.com/Defacto2/server/internal/postgres/models"
)

const (
	padding = " "
)

var ErrModel = fmt.Errorf("error, no file model")

// LeadStr takes a string and returns the leading whitespace padding, characters wide.
// the value of string is note returned.
func LeadStr(width int, s string) string {
	l := utf8.RuneCountInString(s)
	if l >= width {
		return ""
	}
	return strings.Repeat(padding, width-l)
}

// PublishedFW formats the publication year, month and day to a fixed-width length w value.
func PublishedFW(width int, f *models.File) string {
	s := Published(f)
	if utf8.RuneCountInString(s) < width {
		return LeadStr(width, s) + s
	}
	return s
}

func PublishedFmt(f *models.File) string {
	if f == nil {
		return ErrModel.Error()
	}
	ys, ms, ds := "", "", ""
	if f.DateIssuedYear.Valid {
		if i := int(f.DateIssuedYear.Int16); helper.IsYear(i) {
			ys = strconv.Itoa(i)
		}
	}
	if f.DateIssuedMonth.Valid {
		if s := time.Month(f.DateIssuedMonth.Int16); s.String() != "" {
			ms = s.String()
		}
	}
	if f.DateIssuedDay.Valid {
		if i := int(f.DateIssuedDay.Int16); helper.IsDay(i) {
			ds = fmt.Sprintf("%d", i)
		}
	}
	if isYearOnly := ys != "" && ms == "" && ds == ""; isYearOnly {
		return ys
	}
	if isInvalidDay := ys != "" && ms != "" && ds == ""; isInvalidDay {
		return fmt.Sprintf("%s %s", ys, ms)
	}
	if isInvalid := ys == "" && ms == "" && ds == ""; isInvalid {
		return "unknown date"
	}
	return fmt.Sprintf("%s %s %s", ys, ms, ds)
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

// Icon returns the extensionless name of a .gif image file to use as an icon
// for the filename. The icons are found in /public/image/html3/.
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
