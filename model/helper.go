package model

// Package file helper.go contains helper functions for the model package.

import (
	"errors"
	"fmt"
	"path/filepath"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/Defacto2/server/internal/helper"
	"github.com/Defacto2/server/internal/postgres/models"
)

var ErrModel = errors.New("error, no file model")

func DosPaths(zipContent string) []string {
	if zipContent == "" {
		return []string{}
	}
	const delimiter = ":" // the colon is an illegal character as a DOS filename
	archive := zipContent
	archive = strings.ReplaceAll(archive, "\r\n", delimiter) // replace Microsoft-style CRLF with delimiter
	archive = strings.ReplaceAll(archive, "\n", delimiter)   // replace Unix LF with delimiter
	archive = strings.ReplaceAll(archive, "\r", delimiter)   // replace 8-bit microcomputer era CR with delimiter
	paths := strings.Split(archive, delimiter)
	// TODO convert into DOS 8.3 filename format?
	return paths
}

func DosBins(paths ...string) []string {
	if len(paths) == 0 {
		return []string{}
	}
	programs := []string{".bat", ".com", ".exe"}
	executables := []string{}
	for _, path := range paths {
		p := strings.ToLower(path)
		if slices.Contains(programs, filepath.Ext(p)) {
			executables = append(executables, path)
		}
	}
	return executables
}

func DosMatch(filename string, paths ...string) string {
	if filename == "" || len(paths) == 0 {
		return ""
	}
	// sort by the number of directories in the path, to prioritise binaries in the root of the archive
	sort.Slice(paths, func(i, j int) bool {
		return len(filepath.SplitList(paths[i])) < len(filepath.SplitList(paths[j]))
	})
	// prioritise the most likely executable that matches the archive name
	// e.g. if the archive is 'myapp.zip' then the most likely executable order are
	// 'myapp.exe', 'myapp.com', 'myapp.bat'
	root := paths
	sort.Slice(root, func(i, j int) bool {
		// only consider executables in the root of the archive
		return len(filepath.SplitList(root[i])) == 0
	})
	if len(root) == 0 {
		return ""
	}
	base := filepath.Base(filename)
	name := strings.TrimSuffix(base, filepath.Ext(base))
	priority := []string{name + ".exe", name + ".com", name + ".bat"}
	for _, name := range priority {
		for _, path := range root {
			if strings.ToLower(path) == name {
				return path
			}
		}
	}
	return ""
}

func DosBin(paths ...string) string {
	if len(paths) == 0 {
		return ""
	}
	// sort by the number of directories in the path, to prioritise binaries in the root of the archive
	sort.Slice(paths, func(i, j int) bool {
		return len(filepath.SplitList(paths[i])) < len(filepath.SplitList(paths[j]))
	})
	// in the future we could limit the directory depth of the search

	for _, path := range paths {
		if strings.ToLower(filepath.Ext(path)) == ".bat" {
			return path
		}
	}
	for _, path := range paths {
		if strings.ToLower(filepath.Ext(path)) == ".com" {
			return path
		}
	}
	for _, path := range paths {
		if strings.ToLower(filepath.Ext(path)) == ".exe" {
			return path
		}
	}
	return ""
}

func DosBinary(filename, zipContent string) string {
	if filename == "" {
		return ""
	}
	if zipContent == "" {
		return filename
	}

	archives := []string{".zip"} // js-dos only supports ZIP archives
	ext := strings.ToLower(filepath.Ext(filename))
	if !slices.Contains(archives, ext) {
		return filename
	}
	paths := DosPaths(zipContent)
	bins := DosBins(paths...)
	switch len(bins) {
	case 0:
		return ""
	case 1:
		return bins[0]
	}
	if s := DosMatch(filename, paths...); s != "" {
		return s
	}
	return DosBin(paths...)
}

func JsDosBinary(f *models.File) string {
	if f == nil {
		return ErrModel.Error()
	}
	if f.DoseeRunProgram.Valid && f.DoseeRunProgram.String != "" {
		return f.DoseeRunProgram.String
	}
	if !f.Filename.Valid || f.Filename.IsZero() || f.Filename.String == "" {
		return ""
	}
	name := strings.ToLower(f.Filename.String)
	switch filepath.Ext(name) {
	case ".com", ".exe":
		return f.Filename.String
	}
	if !f.FileZipContent.Valid || f.FileZipContent.IsZero() || f.FileZipContent.String == "" {
		return ""
	}
	return DosBinary(f.Filename.String, f.FileZipContent.String)
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
			ds = strconv.Itoa(i)
		}
	}
	if isYearOnly := ys != "" && ms == "" && ds == ""; isYearOnly {
		return ys
	}
	if isInvalidDay := ys != "" && ms != "" && ds == ""; isInvalidDay {
		return ys + " " + ms
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
