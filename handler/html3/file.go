package html3

// Package file html3.go contains the file record detail functions.

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/Defacto2/server/internal/helper"
	"github.com/Defacto2/server/internal/tags"
	"github.com/Defacto2/server/model/html3"
)

const (
	maxPad  = 80
	padding = " "
	noValue = "-"
)

// File record details.
type File struct {
	Filename string // Filename of the file.
	Title    string // Title of the file.
	GroupBy  string // Group name that's is credited with the file.
	Section  string // Section is a tag categorization.
	Platform string // Platform or operating system of the file.
	Size     int64  // Size of the file in bytes.
}

// Description returns a HTML3 friendly file description.
func (f File) Description() string {
	if f.GroupBy == "" {
		return ""
	}
	var desc string
	category := strings.TrimSpace(f.Section)
	if category == tags.Mag.String() {
		desc = fmt.Sprintf("%s issue %s", f.GroupBy, f.Title)
		if f.IsOS() {
			desc += f.OS()
		}
		return desc + "."
	}
	if t := helper.TrimPunct(f.Title); t == "" {
		desc = "A release from "
	} else {
		desc = t + " from "
	}
	desc += f.GroupBy
	if f.IsOS() {
		desc += f.OS()
	}
	return desc + "."
}

// FileLinkPad adds whitespace padding after the hyperlinked filename.
func (f File) FileLinkPad(width int) string {
	s := helper.TruncFilename(width, f.Filename)
	if utf8.RuneCountInString(s) < width {
		return html3.LeadStr(width, s)
	}
	return ""
}

// IsOS returns true if the platform matches Windows, macOS, Linux, MS-DOS or Java.
func (f File) IsOS() bool {
	s := tags.OSTags()
	apps := s[:]
	plat := strings.TrimSpace(strings.ToLower(f.Platform))
	return helper.Finds(plat, apps...)
}

// LeadFS formats the file size to the fixed-width length w value.
func (f File) LeadFS(width int) string {
	s := helper.ByteCount(f.Size)
	l := utf8.RuneCountInString(s)
	return Leading(width-l) + s
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
		return " for " + tags.Names()[p]
	default:
		return ""
	}
}
