package html3

// Helper functions for the TemplateFuncMap var.

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/Defacto2/server/pkg/helpers"
	"github.com/Defacto2/server/pkg/tags"
	"github.com/volatiletech/null/v8"
	"go.uber.org/zap"
)

const (
	maxPad  = 80
	padding = " "
	noValue = "-"
)

// File record details.
type File struct {
	Filename string // Filename of the file.
	Size     int64  // Size of the file in bytes.
	Title    string // Title of the file.
	GroupBy  string // Group name that's is credited with the file.
	Section  string // Section is a tag categorization.
	Platform string // Platform or operating system of the file.
}

// Description returns a HTML3 friendly file description.
func (f File) Description() string {
	if f.GroupBy == "" {
		return ""
	}
	desc := ""
	category := strings.TrimSpace(f.Section)
	if category == tags.Mag.String() {
		desc = fmt.Sprintf("%s issue %s", f.GroupBy, f.Title)
		if f.IsOS() {
			desc += f.OS()
		}
		return fmt.Sprintf("%s.", desc)
	}
	if t := helpers.TrimPunct(f.Title); t == "" {
		desc = "A release from "
	} else {
		desc = fmt.Sprintf("%s from ", t)
	}
	desc += f.GroupBy
	if f.IsOS() {
		desc += f.OS()
	}
	return fmt.Sprintf("%s.", desc)
}

// Description returns a HTML3 friendly file description.
func Description(section, platform, brand, title null.String) string {
	return File{
		Section:  section.String,
		Platform: platform.String,
		GroupBy:  brand.String,
		Title:    title.String,
	}.Description()
}

// FileHref creates a URL to link to the file download of the ID.
func FileHref(id int64, log *zap.SugaredLogger) string {
	href, err := url.JoinPath("/", "html3", "d",
		helpers.ObfuscateParam(strconv.Itoa(int(id))))
	if err != nil {
		log.Error("FileHref ID %d could not be made into a valid URL: %s", err)
		return ""
	}
	return href
}

// FileLinkPad adds whitespace padding after the hyperlinked filename.
func FileLinkPad(width int, name null.String) string {
	if !name.Valid {
		return Leading(width)
	}
	return File{Filename: name.String}.FileLinkPad(width)
}

// FileLinkPad adds whitespace padding after the hyperlinked filename.
func (f File) FileLinkPad(width int) string {
	s := helpers.TruncFilename(width, f.Filename)
	fmt.Println(s, utf8.RuneCountInString(s))
	if utf8.RuneCountInString(s) < width {
		return LeadStr(width, s)
	}
	return ""
}

// Filename returns a truncated filename with to the w maximum width.
func Filename(width int, name null.String) string {
	return helpers.TruncFilename(width, name.String)
}

// Leading repeats the number of space characters.
func Leading(count int) string {
	if count < 1 {
		return ""
	}
	return strings.Repeat(padding, count)
}

// LeadFS formats the file size to the fixed-width length w value.
func LeadFS(width int, size null.Int64) string {
	if !size.Valid {
		return Leading(width)
	}
	return File{Size: size.Int64}.LeadFS(width)
}

// LeadFSInt formats the file size to the fixed-width length w value.
func LeadFSInt(width, size int) string {
	return File{Size: int64(size)}.LeadFS(width)
}

// LeadFS formats the file size to the fixed-width length w value.
func (f File) LeadFS(width int) string {
	s := helpers.ByteCount(f.Size)
	l := utf8.RuneCountInString(s)
	return Leading(width-l) + s
}

// LeadInt takes an int and returns it as a string, w characters wide with whitespace padding.
func LeadInt(width, i int) string {
	s := noValue
	if i > 0 {
		s = strconv.Itoa(i)
	}
	l := utf8.RuneCountInString(s)
	if l >= width {
		return s
	}
	count := width - l
	if count > maxPad {
		count = maxPad
	}
	return fmt.Sprintf("%s%s", strings.Repeat(padding, count), s)
}

// LeadStr takes a string and returns the leading whitespace padding, characters wide.
// the value of string is note returned.
func LeadStr(width int, s string) string {
	l := utf8.RuneCountInString(s)
	if l >= width {
		return ""
	}
	return strings.Repeat(padding, width-l)
}

// IsOS returns true if the platform matches Windows, macOS, Linux, MS-DOS or Java.
func (f File) IsOS() bool {
	s := tags.OSTags()
	apps := s[:]
	plat := strings.TrimSpace(strings.ToLower(f.Platform))
	return helpers.Finds(plat, apps...)
}

// OS returns the platform operating system description
// or an empty string for generic platforms and media.
func (f File) OS() string {
	p := tags.TagByURI(strings.TrimSpace(f.Platform))
	switch p {
	case tags.DOS,
		tags.Java,
		tags.Linux,
		tags.Windows,
		tags.Mac:
		return fmt.Sprintf(" for %s", tags.Names[p])
	default:
		return ""
	}
}
