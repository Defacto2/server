package model

import (
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"mime"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/Defacto2/releaser"
	"github.com/Defacto2/server/internal/tags"
	"github.com/volatiletech/null/v8"
)

const (
	ShortLimit   = 100
	LongFilename = 255
)

// ValidDateIssue returns a valid year, month and day or a null value.
func ValidDateIssue(y, m, d string) (null.Int16, null.Int16, null.Int16) {
	const base, bitSize = 10, 16
	i, _ := strconv.ParseInt(y, base, bitSize)
	year := ValidY(int16(i))

	i, _ = strconv.ParseInt(m, base, bitSize)
	month := ValidM(int16(i))

	i, _ = strconv.ParseInt(d, base, bitSize)
	day := ValidD(int16(i))

	return year, month, day
}

// ValidD returns a valid day or a null value.
func ValidD(d int16) null.Int16 {
	const first, last = 1, 31
	if d < first || d > last {
		return null.Int16{Int16: 0, Valid: false}
	}
	return null.Int16{Int16: d, Valid: true}
}

// ValidM returns a valid month or a null value.
func ValidM(m int16) null.Int16 {
	const jan, dec = 1, 12
	if m < jan || m > dec {
		return null.Int16{Int16: 0, Valid: false}
	}
	return null.Int16{Int16: m, Valid: true}
}

// ValidY returns a valid year or a null value.
func ValidY(y int16) null.Int16 {
	current := int16(time.Now().Year())
	if y < EpochYear || y > current {
		return null.Int16{Int16: 0, Valid: false}
	}
	return null.Int16{Int16: y, Valid: true}
}

// ValidFilename returns a valid filename or a null value.
// The filename is trimmed and shortened to the long filename limit.
func ValidFilename(s string) null.String {
	invalid := null.String{String: "", Valid: false}
	t := trimName(s)
	if len(t) == 0 {
		return invalid
	}
	return null.StringFrom(t)
}

// ValidFilesize returns a valid file size or an error.
// The file size is parsed as an unsigned integer.
// An error is returned if the string cannot be parsed as an integer.
func ValidFilesize(size string) (int64, error) {
	size = strings.TrimSpace(size)
	if len(size) == 0 {
		return 0, nil
	}
	s, err := strconv.ParseUint(size, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("%w: %q, %w", ErrSize, size, err)
	}
	return int64(s), nil
}

// ValidIntegrity confirms the integrity as a valid SHA-384 hexadecimal hash
// or returns a null value.
func ValidIntegrity(integrity string) null.String {
	invalid := null.String{String: "", Valid: false}
	if len(integrity) == 0 {
		return invalid
	}
	if len(integrity) != sha512.Size384*2 {
		return invalid
	}
	_, err := hex.DecodeString(integrity)
	if err != nil {
		return invalid
	}
	return null.StringFrom(integrity)
}

// ValidLastMod returns a valid last modified time or a null value.
// The lastmod time is parsed as a Unix time in milliseconds.
// An error is returned if the string cannot be parsed as an integer.
// The lastmod time is validated to be within the current year and the epoch year of 1980.
func ValidLastMod(lastmod string) null.Time {
	invalid := null.Time{Time: time.Time{}, Valid: false}
	if len(lastmod) == 0 {
		return invalid
	}
	i, err := strconv.ParseInt(lastmod, 10, 64)
	if err != nil {
		return invalid
	}
	val := time.UnixMilli(i)
	now := time.Now()
	if val.After(now) {
		return invalid
	}
	eposh := time.Date(EpochYear, time.January, 1, 0, 0, 0, 0, time.UTC)
	if val.Before(eposh) {
		return invalid
	}
	return null.TimeFrom(val)
}

// ValidMagic returns a valid media type or a null value.
// It is validated using the mime package.
// The media type is trimmed and validated using the mime package.
func ValidMagic(mediatype string) null.String {
	invalid := null.String{String: "", Valid: false}
	mtype := strings.TrimSpace(mediatype)
	if len(mtype) == 0 {
		return invalid
	}
	r, err := mime.ExtensionsByType(mtype)
	if err != nil || len(r) == 0 {
		return invalid
	}
	param := map[string]string{}
	result := mime.FormatMediaType(mediatype, param)
	return null.StringFrom(result)
}

// ValidPlatform returns a valid platform or a null value.
func ValidPlatform(platform string) null.String {
	invalid := null.String{String: "", Valid: false}
	p := strings.TrimSpace(platform)
	if tags.IsPlatform(p) {
		s := tags.TagByURI(p).String()
		return null.StringFrom(s)
	}
	return invalid
}

// ValidReleasers returns two valid releaser group strings or null values.
func ValidReleasers(s1, s2 string) (null.String, null.String) {
	invalid := null.String{String: "", Valid: false}
	t1, t2 := trimShort(s1), trimShort(s2)
	t1, t2 = releaser.Clean(t1), releaser.Clean(t2)
	t1, t2 = strings.ToUpper(t1), strings.ToUpper(t2)
	x1, x2 := invalid, invalid
	if len(t1) > 0 {
		x1 = null.StringFrom(t1)
	}
	if len(t2) > 0 {
		x2 = null.StringFrom(t2)
	}
	if len(t1) == 0 && len(t2) > 0 {
		x1 = null.StringFrom(t2)
		x2 = invalid
	}
	return x1, x2
}

// ValidSceners returns a valid sceners string or a null value.
func ValidSceners(s string) null.String {
	invalid := null.String{String: "", Valid: false}
	t := trimShort(s)
	if len(t) == 0 {
		return invalid
	}
	const sep = ","
	ts := strings.Split(t, sep)
	for i, v := range ts {
		ts[i] = releaser.Clean(strings.TrimSpace(v))
	}
	t = strings.Join(ts, sep)
	return null.StringFrom(t)
}

// ValidSection returns a valid section or a null value.
func ValidSection(section string) null.String {
	invalid := null.String{String: "", Valid: false}
	tag := strings.TrimSpace(section)
	if tags.IsCategory(tag) {
		s := tags.TagByURI(tag).String()
		return null.StringFrom(s)
	}
	return invalid
}

// ValidString returns a valid string or a null value.
func ValidString(s string) null.String {
	invalid := null.String{String: "", Valid: false}
	x := strings.TrimSpace(s)
	if len(x) == 0 {
		return invalid
	}
	return null.StringFrom(x)
}

// ValidTitle returns a valid title or a null value.
// The title is trimmed and shortened to the short limit.
func ValidTitle(s string) null.String {
	invalid := null.String{String: "", Valid: false}
	t := trimShort(s)
	if len(t) == 0 {
		return invalid
	}
	return null.StringFrom(t)
}

// ValidYouTube returns true if the string is a valid YouTube video ID.
// An error is only returned if the regular expression match cannot compile.
func ValidYouTube(s string) (null.String, error) {
	const fixLen = 11
	invalid := null.String{String: "", Valid: false}
	if len(s) != fixLen {
		return invalid, nil
	}
	match, err := regexp.MatchString("^[a-zA-Z0-9_-]{11}$", s)
	if err != nil {
		return invalid, fmt.Errorf("regexp.MatchString: %w", err)
	}
	if !match {
		return invalid, nil
	}
	return null.String{String: s, Valid: true}, nil
}

// trimShort returns a string that is no longer than the short limit.
// It will also remove any leading or trailing white space.
func trimShort(s string) string {
	x := strings.TrimSpace(s)
	if len(x) > ShortLimit {
		return x[:ShortLimit]
	}
	return x
}

// trimName returns a string that is no longer than the long filename limit.
// It will also remove any leading or trailing white space.
func trimName(s string) string {
	s = strings.TrimSpace(s)
	if len(s) > LongFilename {
		return s[:LongFilename]
	}
	return s
}
