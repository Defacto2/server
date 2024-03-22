package model

// Package file helper.go contains helper functions for the model package.

import (
	"errors"
	"fmt"
	"html/template"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/Defacto2/server/internal/helper"
	"github.com/Defacto2/server/internal/jsdos"
	"github.com/Defacto2/server/internal/postgres/models"
)

var ErrModel = errors.New("error, no file model")

// JsDosBinary returns the program executable to run in the JS-DOS emulator.
// If the dosee_run_program is set then it is the preferred executable.
// If the filename is a .com or .exe then it will return the filename.
// Otherwise, it will attempt to find the most likely executable in the archive.
func JsDosBinary(f *models.File) string {
	if f == nil {
		return ErrModel.Error()
	}
	// if set, the dosee_run_program is the preferred executable to run
	if f.DoseeRunProgram.Valid && f.DoseeRunProgram.String != "" {
		return f.DoseeRunProgram.String
	}
	if !f.Filename.Valid || f.Filename.IsZero() || f.Filename.String == "" {
		return ""
	}
	name := strings.ToLower(f.Filename.String)
	switch filepath.Ext(name) {
	case ".com", ".exe":
		return jsdos.Fmt8dot3(f.Filename.String)
	}
	if !f.FileZipContent.Valid || f.FileZipContent.IsZero() || f.FileZipContent.String == "" {
		return ""
	}
	return jsdos.Fmt8dot3(jsdos.FindBinary(f.Filename.String, f.FileZipContent.String))
}

// TODO JsDosConfig returns the JS-DOS configuration for the emulator.

func PublishedFmt(f *models.File) template.HTML {
	if f == nil {
		return template.HTML(ErrModel.Error())
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
			ds = strconv.Itoa(i)
		}
	}
	strong := func(s string) template.HTML {
		return template.HTML("<strong>" + s + "</strong>")
	}
	if isYearOnly := ys != "" && ms == "" && ds == ""; isYearOnly {
		return template.HTML(strong(ys))
	}
	if isInvalidDay := ys != "" && ms != "" && ds == ""; isInvalidDay {
		return strong(ys) + template.HTML(" "+ms)
	}
	if isInvalid := ys == "" && ms == "" && ds == ""; isInvalid {
		return "unknown date"
	}
	return strong(ys) + template.HTML(fmt.Sprintf(" %s %s", ms, ds))
}

func calc(o, l int) int {
	if o < 1 {
		o = 1
	}
	return (o - 1) * l
}
