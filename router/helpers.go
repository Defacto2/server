package router

import (
	"fmt"
	"html/template"
	"log"
	"net/url"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/bengarrett/df2023/helpers"
	"github.com/bengarrett/df2023/models"
	"github.com/volatiletech/null/v8"
)

const (
	maxPad  = 80
	padding = " "
	noValue = "-"
)

// TemplateFuncMap are a collection of mapped functions that can be used in a template.
var TemplateFuncMap = template.FuncMap{
	"leading":  Leading,
	"leadInt":  LeadInt,
	"leadStr":  LeadStr,
	"iconFmt":  models.Icon,
	"linkFile": FmtFilename,
	"linkHref": FileHref,
	"linkPad":  FileLinkPad,
	"datePub":  LeadPub,
	"datePost": LeadPost,
	"byteFmt":  LeadFS,
	"descript": Description,
}

// Description returns a HTML3 friendly file description.
func Description(w int, section, platform, brand, title null.String) string {
	category := strings.TrimSpace(section.String)
	if category == "magazine" {
		return fmt.Sprintf("%s issue %s.", brand.String, title.String)
	}
	desc := ""
	if t := helpers.TrimPunct(title.String); t == "" {
		desc = "From "
	} else {
		desc = fmt.Sprintf("%s from ", t)
	}
	desc += brand.String
	if IsApp(platform) {
		desc += FmtApp(platform)
	}
	return fmt.Sprintf("%s.", desc)
}

// IsApp returns true if the platform matches Windows, macOS, Linux, MS-DOS or Java.
func IsApp(platform null.String) bool {
	s := []string{"dos", "java", "linux", "windows", "mac10"}
	p := strings.TrimSpace(strings.ToLower(platform.String))
	return helpers.Find(p, s...)
}

// FmtApp returns the application platform as a string.
func FmtApp(platform null.String) string {
	s := ""
	p := strings.TrimSpace(strings.ToLower(platform.String))
	switch p {
	case "dos":
		s = "DOS"
	case "java":
		s = "Java"
	case "linux":
		s = "Linux"
	case "windows":
		s = "Windows"
	case "mac10":
		s = "macOS"
	}
	if s == "" {
		return ""
	}
	return fmt.Sprintf(" for %s", s)
}

// FmtFilename returns a truncated filename with to the w maximum width.
func FmtFilename(width int, name null.String) string {
	return helpers.TruncFilename(width, name.String)
}

// FileLinkPad adds whitespace padding after the hyperlinked filename.
func FileLinkPad(width int, name null.String) string {
	if !name.Valid {
		return Leading(width)
	}
	s := helpers.TruncFilename(width, name.String)

	if utf8.RuneCountInString(s) < width {
		return LeadStr(width, s)
	}
	return ""
}

// FileHref creates a URL to link to the file download of the ID.
func FileHref(id int64) string {
	href, err := url.JoinPath("/", "d", helpers.ObfuscateParam(strconv.Itoa(int(id))))
	if err != nil {
		log.Println(err) //TODO: log to file
	}
	return href
}

// LeadPub formats the publication year, month and day to a fixed-width length w value.
func LeadPub(width int, y, m, d null.Int16) string {
	s := models.FmtPublish(y, m, d)
	if utf8.RuneCountInString(s) < width {
		return LeadStr(width, s) + s
	}
	return s
}

// LeadPost formats the date published to the fixed-width length w value.
func LeadPost(width int, t null.Time) string {
	s := models.FmtTime(t)
	if utf8.RuneCountInString(s) < width {
		return LeadStr(width, s) + s
	}
	return s
}

// LeadFS formats the file size to the fixed-width length w value.
func LeadFS(width int, size null.Int64) string {
	if !size.Valid {
		return Leading(width)
	}
	s := helpers.ByteCount(size.Int64)
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

// Leading repeats the number of space characters.
func Leading(count int) string {
	if count < 1 {
		return ""
	}
	return strings.Repeat(padding, count)
}
