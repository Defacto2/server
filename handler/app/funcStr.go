package app

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/Defacto2/server/internal/helper"
	"github.com/volatiletech/null/v8"
)

// Prefix returns a string prefixed with a space.
func Prefix(s string) string {
	if s == "" {
		return ""
	}
	return fmt.Sprintf("%s ", s)
}

// TrimSiteSuffix returns a string with the last 4 characters removed if they are " FTP" or " BBS".
func TrimSiteSuffix(s string) string {
	n := strings.ToLower(strings.TrimSpace(s))
	const chrs = 4
	if len(s) < chrs {
		return s
	}
	switch n[len(s)-chrs:] {
	case " ftp", " bbs":
		return s[:len(s)-chrs]
	}
	return s
}

// TrimSpace returns a string with all leading and trailing whitespace removed.
func TrimSpace(s string) string {
	return strings.TrimSpace(s)
}

// Updated returns a string of the time since the given time t.
// The time is formatted as "Last updated 1 hour ago".
// If the time is not valid, an empty string is returned.
func Updated(t any, s string) string {
	if s == "" {
		s = "Time"
	}
	switch val := t.(type) {
	case null.Time:
		if !val.Valid {
			return ""
		}
		return fmt.Sprintf("%s %s ago", s, helper.TimeDistance(val.Time, time.Now(), true))
	case time.Time:
		return fmt.Sprintf("%s %s ago", s, helper.TimeDistance(val, time.Now(), true))
	default:
		return fmt.Sprintf("%supdated: %s", typeErr, reflect.TypeOf(t).String())
	}
}

// WebsiteIcon returns a Bootstrap icon name for the given website url.
func WebsiteIcon(url string) string {
	switch {
	case strings.Contains(url, "archive.org"):
		return "bank"
	case strings.Contains(url, "reddit.com"):
		return "reddit"
	case strings.Contains(url, "yalebooks.yale.edu"),
		strings.Contains(url, "explodingthephone.com"),
		strings.Contains(url, "punctumbooks"):
		return "book"
	case strings.Contains(url, "youtube.com"):
		return "youtube"
	case strings.Contains(url, "vimeo.com"):
		return "vimeo"
	case strings.Contains(url, "slashdot.org"):
		return "slash-circle"
	}
	return "arrow-right"
}
