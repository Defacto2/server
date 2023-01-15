package models

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/Defacto2/server/helpers"
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

// FmtPublish takes optional year, month and values and formats them to dd-mmm-yyyy.
// Depending on the context, any missing time values will be left blank or replaced with ?? question marks.
func FmtPublish(y, m, d null.Int16) string {
	const (
		yx       = "????"
		mx       = "???"
		dx       = "??"
		sp       = " "
		yPadding = 7
		dPadding = 3
	)
	ys, ms, ds := yx, mx, dx
	if y.Valid {
		if i := int(y.Int16); IsYear(i) {
			ys = strconv.Itoa(i)
		}
	}
	if m.Valid {
		if s := helpers.ShortMonth(int(m.Int16)); s != "" {
			ms = s
		}
	}
	if d.Valid {
		if i := int(d.Int16); IsDay(i) {
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

// FmtTime formats the time to use dd-mmm-yyyy.
func FmtTime(t null.Time) string {
	if !t.Valid {
		return ""
	}
	d := t.Time.Day()
	m := helpers.ShortMonth(int(t.Time.Month()))
	y := t.Time.Year()
	return fmt.Sprintf("%02d-%s-%d", d, m, y)
}

// IsDay returns true if the i value can be used as a day time value.
func IsDay(i int) bool {
	const maxDay = 31
	if i > 0 && i <= maxDay {
		return true
	}
	return false
}

// IsYear returns true  if the i value is greater than 1969
// or equal to the current year.
func IsYear(i int) bool {
	const unix = 1970
	now := time.Now().Year()
	if i >= unix && i <= now {
		return true
	}
	return false
}

// Icon returns the extensionless name of a .gif image file to use as an icon
// for the named file.
// The icons are found in /public/images/html3/.
func Icon(name null.String) string {
	const (
		app   = "comp2"
		doc   = "doc"
		error = "unknown"
		htm   = "generic"
		pic   = "image2"
		sfx   = "sound2"
		vid   = "movie"
		zip   = "compressed"
	)
	if !name.Valid {
		return error
	}
	n := strings.ToLower(filepath.Ext(name.String))
	switch {
	case IsArchive(n):
		return zip
	case IsApp(n):
		return app
	case IsAudio(n):
		return sfx
	case IsDocument(n):
		return doc
	case IsHTML(n):
		return htm
	case IsImage(n):
		return pic
	case IsTune(n):
		return sfx
	case IsVideo(n):
		return vid
	}
	return error
}

// IsApp returns true if the named file uses a Windows application filename.
func IsApp(name string) bool {
	s := []string{".exe", ".com"}
	return IsExt(name, s...)
}

// IsApp returns true if the named file uses a common compressed or archived filename.
func IsArchive(name string) bool {
	s := []string{".7z", ".arc", ".ark", ".arj", ".cab", ".gz", ".lha", ".lzh", ".rar", ".tar", ".tar.gz", ".zip"}
	return IsExt(name, s...)
}

// IsDocument returns true if the named file uses a common document or text filename.
func IsDocument(name string) bool {
	s := []string{".1st", ".asc", ".ans", ".cap", ".diz", ".doc", ".dox", ".me", ".nfo", ".pcb", ".pdf", ".txt", ".unp"}
	return IsExt(name, s...)
}

// IsImage returns true if the named file uses a common image or photo filename.
func IsImage(name string) bool {
	s := []string{".bmp", ".gif", ".ico", ".iff", ".jpg", ".jpeg", ".lbm", ".png", ".pcx"}
	return IsExt(name, s...)
}

// IsHTML returns true if the named file uses a HTML markup filename.
func IsHTML(name string) bool {
	s := []string{".htm", ".html"}
	return IsExt(name, s...)
}

// IsImage returns true if the named file uses a common digital audio filename.
func IsAudio(name string) bool {
	s := []string{".au", ".flac", ".m1a", ".m2a", ".mid", ".midi", ".mp1", ".mp2", ".mp3",
		".mpa", ".mpga", ".mpeg", ".ogg", ".snd", ".wav", ".wave", ".wma"}
	return IsExt(name, s...)
}

// IsImage returns true if the named file uses a common tracker music filename.
func IsTune(name string) bool {
	s := []string{".it", ".mod", ".s3m", ".xm"}
	return IsExt(name, s...)
}

// IsImage returns true if the named file uses a common video filename.
func IsVideo(name string) bool {
	s := []string{".avi", ".divx", ".flv", ".gt", ".mov", ".m4a", ".m4v", ".mp4", ".swf", ".rm", ".ram", ".wmv", ".xvid"}
	return IsExt(name, s...)
}

// IsExt returns true if the file extension of the named file is found in the collection of extensions.
func IsExt(name string, extensions ...string) bool {
	ext := strings.ToLower(filepath.Ext(name))
	return helpers.Finds(ext, extensions...)
}
