package model

// Package file helper.go contains helper functions for the model package.

import (
	"fmt"
	"html/template"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/Defacto2/server/internal/helper"
	"github.com/Defacto2/server/internal/jsdos"
	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/subpop/go-ini"
)

// JsDosBinary returns the program executable to run in the js-dos emulator.
// If the dosee_run_program is set then it is the preferred executable.
// If the filename is a .com or .exe then it will return the filename.
// Otherwise, it will attempt to find the most likely executable in the archive.
func JsDosBinary(f *models.File) (string, error) {
	if f == nil {
		return "", ErrModel
	}
	// if set, the dosee_run_program is the preferred executable to run
	if f.DoseeRunProgram.Valid && f.DoseeRunProgram.String != "" {
		return f.DoseeRunProgram.String, nil
	}
	if !f.Filename.Valid || f.Filename.IsZero() || f.Filename.String == "" {
		return "", nil
	}
	name := strings.ToLower(f.Filename.String)
	switch filepath.Ext(name) {
	case ".com", ".exe":
		return jsdos.Fmt8dot3(f.Filename.String), nil
	}
	if !f.FileZipContent.Valid || f.FileZipContent.IsZero() || f.FileZipContent.String == "" {
		return "", nil
	}
	return jsdos.Fmt8dot3(jsdos.FindBinary(f.Filename.String, f.FileZipContent.String)), nil
}

// JsDosConfig creates a js-dos .ini configuration for the emulator.
func JsDosConfig(f *models.File) (string, error) {
	if f == nil {
		return "", ErrModel
	}
	j := jsdos.Jsdos{}
	cpu := f.DoseeHardwareCPU.String
	if f.DoseeHardwareCPU.Valid && cpu != "" {
		j.CPU(cpu)
	}
	hw := f.DoseeHardwareGraphic.String
	if f.DoseeHardwareGraphic.Valid && hw != "" {
		j.Machine(hw)
	}
	sfx := f.DoseeHardwareAudio.String
	if f.DoseeHardwareAudio.Valid && sfx != "" {
		j.Sound(sfx)
	}
	mem := f.DoseeNoEms.Int16
	if f.DoseeNoEms.Valid && mem == 1 {
		j.NoEMS(true)
	}
	mem = f.DoseeNoXMS.Int16
	if f.DoseeNoXMS.Valid && mem == 1 {
		j.NoXMS(true)
	}
	mem = f.DoseeNoUmb.Int16
	if f.DoseeNoUmb.Valid && mem == 1 {
		j.NoUMB(true)
	}
	b, err := ini.Marshal(j)
	if err != nil {
		return "", fmt.Errorf("ini.Marshal: %w", err)
	}
	return string(b), nil
}

// PublishedFmt returns a formatted date string for the artifact's published date.
func PublishedFmt(f *models.File) template.HTML {
	if f == nil {
		return template.HTML(ErrModel.Error()) //nolint: gosec
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
		return template.HTML("<strong>" + s + "</strong>") //nolint: gosec
	}
	if isYearOnly := ys != "" && ms == "" && ds == ""; isYearOnly {
		return strong(ys)
	}
	if isInvalidDay := ys != "" && ms != "" && ds == ""; isInvalidDay {
		return strong(ys) + template.HTML(" "+ms) //nolint: gosec
	}
	if isInvalid := ys == "" && ms == "" && ds == ""; isInvalid {
		return "unknown date"
	}
	return strong(ys) + template.HTML(fmt.Sprintf(" %s %s", ms, ds)) //nolint: gosec
}

func calc(o, l int) int {
	if o < 1 {
		o = 1
	}
	return (o - 1) * l
}
