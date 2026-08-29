package model

import (
	"crypto/sha512"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"mime"
	"strconv"
	"strings"
	"time"

	"github.com/Defacto2/server/handler/releaser"
	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/Defacto2/server/internal/tags"
	"github.com/aarondl/null/v8"
)

const (
	ShortLimit   = 100 // ShortLimit is the maximum length of a short string.
	LongFilename = 255 // LongFilename is the maximum length of a filename.
)

// Validate checks the artifact record for any missing or invalid values
// that should prevent it from being published and public.
//
// All found issues will be returned as joined error messages.
func Validate(art *models.File) error {
	const format = "validate models: %w"
	if art == nil {
		return fmt.Errorf(format, ErrModel)
	}

	var err error
	if !art.Section.Valid || art.Section.String == "" {
		err = errors.Join(err, fmt.Errorf("%w,", ErrBadTag))
	} else if !tags.IsCategory(art.Section.String) {
		err = errors.Join(err, fmt.Errorf("%w: %q,", ErrBadTag, art.Section.String))
	}

	if !art.Platform.Valid || art.Platform.String == "" {
		err = errors.Join(err, fmt.Errorf("%w,", ErrBadOS))
	} else if !tags.IsPlatform(art.Platform.String) {
		err = errors.Join(err, fmt.Errorf("%w: %q,", ErrBadOS, art.Platform.String))
	}

	if !art.Filename.Valid || art.Filename.String == "" {
		err = errors.Join(err, fmt.Errorf("%w,", ErrBadFname))
	}

	if (!art.GroupBrandBy.Valid && !art.GroupBrandFor.Valid) ||
		(art.GroupBrandBy.String == "" && art.GroupBrandFor.String == "") {
		err = errors.Join(err, fmt.Errorf("%w,", ErrBadRel))
	}

	if art.Section.String == tags.Mag.String() &&
		(!art.RecordTitle.Valid || art.RecordTitle.String == "") {
		err = errors.Join(err, fmt.Errorf("%w,", ErrBadMag))
	}

	return fmt.Errorf(format, err)
}

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
		return null.Int16{}
	}

	return null.Int16From(d)
}

// ValidM returns a valid month or a null value.
func ValidM(m int16) null.Int16 {
	const jan, dec = 1, 12
	if m < jan || m > dec {
		return null.Int16{}
	}

	return null.Int16From(m)
}

// ValidY returns a valid year or a null value.
func ValidY(y int16) null.Int16 {
	current := int16(math.Abs(float64(time.Now().Year())))
	if y < EpochYear || y > current {
		return null.Int16{}
	}

	return null.Int16From(y)
}

// ValidFilename returns a valid filename or a null value.
// The filename is trimmed and shortened to the long filename limit.
func ValidFilename(s string) null.String {
	t := trimName(s)
	if t == "" {
		return null.String{}
	}

	return null.StringFrom(t)
}

// ValidFilesize returns a valid file size or an error.
// The file size is parsed as an unsigned integer.
// An error is returned if the string cannot be parsed as an integer.
func ValidFilesize(size string) (null.Int64, error) {
	const format = "valid filesize %s: %w"
	s := strings.TrimSpace(size)
	if s == "" {
		return null.Int64{}, nil
	}

	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return null.Int64{}, fmt.Errorf(format, size, err)
	}

	return null.Int64From(i), nil
}

// ValidIntegrity confirms the integrity as a valid SHA-384 hexadecimal hash
// or returns a null value.
func ValidIntegrity(integrity string) null.String {
	if len(integrity) != sha512.Size384*2 {
		return null.String{}
	}

	_, err := hex.DecodeString(integrity)
	if err != nil {
		return null.String{}
	}

	return null.StringFrom(integrity)
}

// ValidLastMod returns a valid last modified time or a null value.
// The lastmod time is parsed as a Unix time in milliseconds.
// An error is returned if the string cannot be parsed as an integer.
// The lastmod time is validated to be within the current year and the epoch year of 1980.
func ValidLastMod(lastmod string) null.Time {
	if lastmod == "" {
		return null.Time{}
	}

	i, err := strconv.ParseInt(lastmod, 10, 64)
	if err != nil {
		return null.Time{}
	}

	t := time.UnixMilli(i)
	now := time.Now()
	if t.After(now) {
		return null.Time{}
	}

	epoch := time.Date(EpochYear, time.January, 1, 0, 0, 0, 0, time.UTC)
	if t.Before(epoch) {
		return null.Time{}
	}

	return null.TimeFrom(t)
}

// ValidMagic returns a valid media type or a null value.
// It is validated using the mime package.
// The media type is trimmed and validated using the mime package.
func ValidMagic(mediatype string) null.String {
	t := strings.TrimSpace(mediatype)
	if t == "" {
		return null.String{}
	}
	r, err := mime.ExtensionsByType(t)
	if err != nil || len(r) == 0 {
		return null.String{}
	}

	param := map[string]string{}
	result := mime.FormatMediaType(mediatype, param)

	return null.StringFrom(result)
}

// ValidPlatform returns a valid platform or a null value.
func ValidPlatform(platform string) null.String {
	p := strings.TrimSpace(platform)
	if !tags.IsPlatform(p) {
		return null.String{}
	}
	s := tags.TagByURI(p).String()
	return null.StringFrom(s)
}

// ValidReleasers returns two valid releaser group strings or null values.
func ValidReleasers(s1, s2 string) (null.String, null.String) {
	x := strings.ToUpper(releaser.Clean(trimShort(s1)))
	y := strings.ToUpper(releaser.Clean(trimShort(s2)))

	if x != "" && x == y {
		y = ""
	}

	if x == "" && y != "" {
		x, y = y, ""
	}

	var r1, r2 null.String
	if x != "" {
		r1 = null.StringFrom(x)
	}
	if y != "" {
		r2 = null.StringFrom(y)
	}

	return r1, r2
}

// ValidSceners returns a valid sceners string or a null value.
func ValidSceners(s string) null.String {
	t := trimShort(s)
	if t == "" {
		return null.String{}
	}

	const sep = ","
	x := strings.Split(t, sep)

	n := 0
	for _, elem := range x {
		cleaned := releaser.Clean(strings.TrimSpace(elem))
		if cleaned != "" {
			x[n] = cleaned
			n++
		}
	}
	x = x[:n]

	if len(x) == 0 {
		return null.String{}
	}

	return null.StringFrom(strings.Join(x, sep))
}

// ValidSection returns a valid section or a null value.
func ValidSection(section string) null.String {
	tag := strings.TrimSpace(section)
	if !tags.IsCategory(tag) {
		return null.String{}
	}
	s := tags.TagByURI(tag).String()
	return null.StringFrom(s)
}

// ValidString returns a valid string or a null value.
func ValidString(s string) null.String {
	t := strings.TrimSpace(s)
	if t == "" {
		return null.String{}
	}
	return null.StringFrom(t)
}

// ValidTitle returns a valid title or a null value.
// The title is trimmed and shortened to the short limit.
func ValidTitle(s string) null.String {
	t := trimShort(s)
	if t == "" {
		return null.String{}
	}
	return null.StringFrom(t)
}

// ValidYouTube returns true if the string is a valid YouTube video ID.
// An error is only returned if the regular expression match cannot compile.
func ValidYouTube(s string) null.String {
	if len(s) != 11 {
		return null.String{}
	}

	for i := 0; i < len(s); i++ {
		b := s[i]
		switch {
		case b >= 'a' && b <= 'z':
		case b >= 'A' && b <= 'Z':
		case b >= '0' && b <= '9':
		case b == '_' || b == '-':
		default:
			return null.String{}
		}
	}

	return null.StringFrom(s)
}

// trimShort returns a string that is no longer than the short limit.
// It will also remove any leading or trailing white space.
func trimShort(s string) string {
	x := strings.TrimSpace(s)

	count := 0
	for i := range x {
		if count == ShortLimit {
			return x[:i]
		}
		count++
	}

	return x
}

// trimName returns a string that is no longer than the long filename limit.
// It will also remove any leading or trailing white space.
func trimName(s string) string {
	x := strings.TrimSpace(s)

	count := 0
	for i := range x {
		if count == LongFilename {
			return x[:i]
		}
		count++
	}

	return x
}
