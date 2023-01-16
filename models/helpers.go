package models

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Defacto2/server/helpers"
	"github.com/volatiletech/null/v8"
)

// TODO: Magazines https://github.com/bengarrett/Defacto2-2020/blob/78665d391f40bba14f44b6f9220dfe620c60650d/ROOT/views/html3/_pubedition.cfm
//       https://github.com/bengarrett/Defacto2-2020/blob/78665d391f40bba14f44b6f9220dfe620c60650d/ROOT/views/html3/listfile.cfm

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
		if i := int(y.Int16); helpers.IsYear(i) {
			ys = strconv.Itoa(i)
		}
	}
	if m.Valid {
		if s := helpers.ShortMonth(int(m.Int16)); s != "" {
			ms = s
		}
	}
	if d.Valid {
		if i := int(d.Int16); helpers.IsDay(i) {
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
	case helpers.IsArchive(n):
		return zip
	case helpers.IsApp(n):
		return app
	case helpers.IsAudio(n):
		return sfx
	case helpers.IsDocument(n):
		return doc
	case helpers.IsHTML(n):
		return htm
	case helpers.IsImage(n):
		return pic
	case helpers.IsTune(n):
		return sfx
	case helpers.IsVideo(n):
		return vid
	}
	return error
}
