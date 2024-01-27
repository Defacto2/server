package model

// Package file helper.go contains helper functions for the model package.

import (
	"fmt"
	"strconv"
	"time"

	"github.com/Defacto2/server/internal/helper"
	"github.com/Defacto2/server/internal/postgres/models"
)

var ErrModel = fmt.Errorf("error, no file model")

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

func calc(o, l int) int {
	if o < 1 {
		o = 1
	}
	return (o - 1) * l
}
