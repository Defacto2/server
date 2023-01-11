package models

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/volatiletech/null/v8"
)

// TODOs: https://github.com/bengarrett/Defacto2-2020/blob/78665d391f40bba14f44b6f9220dfe620c60650d/ROOT/views/html3/helpers.cfm
// https://github.com/bengarrett/Defacto2-2020/blob/78665d391f40bba14f44b6f9220dfe620c60650d/ROOT/views/html3/_pubedition.cfm
// https://github.com/bengarrett/Defacto2-2020/blob/78665d391f40bba14f44b6f9220dfe620c60650d/ROOT/views/html3/listfile.cfm
//
// <cfset fileText = Trim("#truncate(LCase(filenameLessExtension(records.fileName)),19,'.')#.#Left(fileExtension(records.fileName),3)#")>
// <cfset spacer = Val(23-Len(fileText))>

// todo: move to helpers/helpers.go
// https://yourbasic.org/golang/formatting-byte-size-to-human-readable-format/
func ByteCount(i null.Int64) string {
	if !i.Valid {
		return ""
	}
	b := i.Int64
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d%s", b, strings.Repeat(" ", 2))
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %c",
		float64(b)/float64(div), "KMGTPE"[exp])
}

func DateFmt(t null.Time) string {
	if !t.Valid {
		return ""
	}
	d := t.Time.Day()
	m := AbbrMonth(int(t.Time.Month()))
	y := t.Time.Year()
	return fmt.Sprintf("%02d-%s-%d", d, m, y)
}

func DatePub(y, m, d null.Int16) string {
	const (
		yx = "????"
		mx = "???"
		dx = "??"
		sp = " "
	)
	ys, ms, ds := yx, mx, dx
	if y.Valid {
		if i := int(y.Int16); IsYear(i) {
			ys = strconv.Itoa(i)
		}
	}
	if m.Valid {
		if s := AbbrMonth(int(m.Int16)); s != "" {
			ms = s
		}
	}
	if d.Valid {
		if i := int(d.Int16); IsDay(i) {
			ds = strconv.Itoa(i)
		}
	}
	if isYearOnly := ys != yx && ms == mx && ds == dx; isYearOnly {
		return fmt.Sprintf("%s%s", strings.Repeat(sp, 8), ys)
	}
	if isInvalidDay := ys != yx && ms != mx && ds == dx; isInvalidDay {
		return fmt.Sprintf("%s%s-%s", strings.Repeat(sp, 3), ms, ys)
	}
	return fmt.Sprintf("%s-%s-%s", ds, ms, ys)
}

func AbbrMonth(i int) string {
	const abbreviated = 3
	s := fmt.Sprint(time.Month(i))
	if len(s) >= abbreviated {
		return s[0:abbreviated]
	}
	return ""
}

func IsDay(i int) bool {
	const maxDay = 31
	if i > 0 && i <= maxDay {
		return true
	}
	return false
}

func IsYear(i int) bool {
	const unix = 1970
	now := time.Now().Year()
	if i >= unix && i <= now {
		return true
	}
	return false
}
