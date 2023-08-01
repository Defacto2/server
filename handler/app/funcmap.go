package app

// Helper functions for the TemplateFuncMap var.

import (
	"crypto/sha512"
	"embed"
	"encoding/base64"
	"fmt"
	"html/template"
	"net/url"
	"reflect"
	"strings"
	"time"

	"github.com/Defacto2/server/pkg/helpers"
)

const (
	// Welcome is the default logo monospace text,
	// each side contains 20 whitespace characters.
	// The welcome to defacto2 text is 19 characters long.
	// The letter 'O' of TO is the center of the text.
	Welcome = ":                    ·· WELCOME TO DEFACTO2 ··                    ·"

	// wiki and link are SVG icons.
	wiki  = `<svg class="bi" aria-hidden="true"><use xlink:href="bootstrap-icons.svg#arrow-right-short"></use></svg>`
	link  = `<svg class="bi" aria-hidden="true"><use xlink:href="bootstrap-icons.svg#link"></use></svg>`
	merge = `<svg class="bi" aria-hidden="true" fill="currentColor"><use xlink:href="bootstrap-icons.svg#forward"></use></svg>`

	typeErr = "error: received an invalid type to "
)

// ByteFormat returns a human readable string of the byte count.
func ByteFormat(b any) string {
	switch val := b.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		i := reflect.ValueOf(val).Int()
		return helpers.ByteCount(i)
	default:
		return fmt.Sprintf("%sByteFmt: %s", typeErr, reflect.TypeOf(b).String())
	}
}

// ExternalLink returns a HTML link with an embedded SVG icon to an external website.
func ExternalLink(href, name string) template.HTML {
	if href == "" {
		return "error: href is empty"
	}
	if name == "" {
		return "error: name is empty"
	}

	return template.HTML(fmt.Sprintf(`<a class="dropdown-item icon-link icon-link-hover" href="%s">%s %s</a>`, href, name, link))
}

// FmtDate returns a string of the date in the format YYYY-MM-DD.
func FmtDate(d any) string {
	switch val := d.(type) {
	case time.Time:
		return val.Format("2006-01-02")
	default:
		return fmt.Sprintf("%sFmtDate: %s", typeErr, reflect.TypeOf(d).String())
	}
}

// FmtDateTime returns a string of the date and time in the format YYYY-MM-DD HH:MM:SS.
func FmtDateTime(d any) string {
	switch val := d.(type) {
	case time.Time:
		return val.Format("2006-01-02 15:04:05")
	default:
		return fmt.Sprintf("%sFmtDateTime: %s", typeErr, reflect.TypeOf(d).String())
	}
}

// FmtDay returns a string of the day number from the day d number between 1 and 31.
func FmtDay(d any) string {
	switch val := d.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		i := reflect.ValueOf(val).Int()
		if i == 0 {
			return ""
		}
		if i < 0 || i > 31 {
			return fmt.Sprintf(" error: day out of range %d", i)
		}
		return fmt.Sprintf(" %d", i)
	default:
		return fmt.Sprintf("%sFmtDay: %s", typeErr, reflect.TypeOf(d).String())
	}
}

// FmtMonth returns a string of the month name from the month m number between 1 and 12.
func FmtMonth(m any) string {
	switch val := m.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		i := reflect.ValueOf(val).Int()
		if i == 0 {
			return ""
		}
		if i < 0 || i > 12 {
			return fmt.Sprintf(" error: month out of range %d", i)
		}
		return " " + time.Month(i).String()
	default:
		return fmt.Sprintf("%sFmtMonth: %s", typeErr, reflect.TypeOf(m).String())
	}
}

// FmtPrefix returns a string prefixed with a space.
func FmtPrefix(s string) string {
	if s == "" {
		return ""
	}
	return fmt.Sprintf("%s ", s)
}

// Integrity returns the sha384 hash of the named embed file.
// This is intended to be used for Subresource Integrity (SRI)
// verification with integrity attributes in HTML script and link tags.
func Integrity(name string, fs embed.FS) (string, error) {
	b, err := fs.ReadFile(name)
	if err != nil {
		return "", err
	}
	return IntegrityBytes(b), nil
}

// IntegrityBytes returns the sha384 hash of the given byte slice.
func IntegrityBytes(b []byte) string {
	sum := sha512.Sum384(b)
	b64 := base64.StdEncoding.EncodeToString(sum[:])
	return fmt.Sprintf("sha384-%s", b64)
}

// LogoText returns a string of text padded with spaces to center it in the logo.
func LogoText(s string) string {
	indent := strings.Repeat(" ", 6)
	if s == "" {
		return indent + Welcome
	}

	// odd returns true if the given integer is odd.
	odd := func(i int) bool {
		return i%2 != 0
	}

	s = strings.ToUpper(s)

	const padder = " ·· "
	const wl, pl = len(Welcome), len(padder)
	const limit = wl - (pl + pl) - 3

	// Truncate the string to the limit.
	if len(s) > limit {
		return fmt.Sprintf("%s:%s%s%s·",
			indent, padder, s[:limit], padder)
	}

	styled := fmt.Sprintf("%s%s%s", padder, s, padder)
	if !odd(len(s)) {
		styled = fmt.Sprintf(" %s%s%s", padder, s, padder)
	}

	// Pad the string with spaces to center it.
	count := (wl / 2) - (len(styled) / 2) - 2

	text := fmt.Sprintf(":%s%s%s·",
		strings.Repeat(" ", count),
		styled,
		strings.Repeat(" ", count))
	return indent + text
}

// Mod returns true if the given integer is a multiple of the given max integer.
func Mod(i any, max int) bool {
	switch val := i.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		v := reflect.ValueOf(val).Int()
		return v%int64(max) == 0
	default:
		return false
	}
}

// Mod3 returns true if the given integer is a multiple of 3.
func Mod3(i any) bool {
	return Mod(i, 3)
}

// WikiLink returns a HTML link with an embedded SVG icon to the Defacto2 wiki on GitHub.
func WikiLink(uri, name string) template.HTML {
	if uri == "" {
		return "error: href is empty"
	}
	if name == "" {
		return "error: name is empty"
	}
	href, err := url.JoinPath("https://github.com/Defacto2/defacto2.net/wiki/", uri)
	if err != nil {
		return template.HTML(err.Error())
	}
	return template.HTML(fmt.Sprintf(`<a class="dropdown-item icon-link icon-link-hover" href="%s">%s %s</a>`, href, name, wiki))
}
