// Package form provides functions for providing data for form and input elements.
package form

import (
	"context"
	"fmt"
	"html/template"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/Defacto2/server/internal/postgres"
	"github.com/Defacto2/server/internal/tags"
	"github.com/Defacto2/server/model"
)

const ReSanitizePath = "[^a-zA-Z0-9-._/]+" // Regular expression to sanitize the URL path.

// HumanizeAndCount returns the human readable name of the platform and section tags combined
// and the number of existing artifacts. The number of existing artifacts is colored based on
// the count. If the count is 0, the text is red. If the count is 1, the text is blue. If the
// count is greater than 1, the text is unmodified.
func HumanizeAndCount(section, platform string) (template.HTML, error) {
	ctx := context.Background()
	db, err := postgres.ConnectDB()
	if err != nil {
		return "cannot connect to the database",
			fmt.Errorf("postgres.ConnectDB: %w", err)
	}
	defer db.Close()
	s := tags.TagByURI(section)
	p := tags.TagByURI(platform)
	tag := tags.Humanize(p, s)
	if strings.HasPrefix(tag, "unknown") {
		return "unknown classification", nil
	}
	count, err := model.CountByClassification(ctx, db, section, platform)
	if err != nil {
		return "cannot count the classification",
			fmt.Errorf("model.CountByClassification: %w", err)
	}
	html := ""
	switch count {
	case 0:
		html = fmt.Sprintf("%s, %d existing artifacts", tag, count)
		html = `<span class="text-danger-emphasis">` + html + `</span>`
	case 1:
		html = fmt.Sprintf("%s, %d existing artifacts", tag, count)
		html = `<span class="text-info-emphasis">` + html + `</span>`
	default:
		html = fmt.Sprintf("%s, %d existing artifacts", tag, count)
	}
	return template.HTML(html), nil
}

// SanitizeFilename returns a sanitized version of the filename.
// The filename is trimmed of any leading or trailing whitespace,
// and any parent directory references are removed. Any Linux or
// Windows directory separators are replaced with a "-" hyphen.
func SanitizeFilename(name string) string {
	const hyphen = "-"
	s := strings.TrimSpace(name)
	const parentDir = "../"
	s = strings.ReplaceAll(s, parentDir, "")
	const linuxDir = "/"
	s = strings.ReplaceAll(s, linuxDir, hyphen)
	const windowsDir = "\\"
	s = strings.ReplaceAll(s, windowsDir, hyphen)
	return s
}

// SanitizeSeparators returns a sanitized version of the URL path.
// The path is trimmed of any URL scheme, host or query parameters, as well
// as any invalid path separators.
func SanitizeSeparators(rawPath string) string {
	const separator = "/"
	raw := strings.TrimSpace(rawPath)
	raw = strings.ReplaceAll(raw, separator+separator, separator)
	raw = strings.Trim(raw, separator)
	u, err := url.Parse(raw)
	if err != nil {
		return "url.Parse error: " + err.Error()
	}
	return u.Path
}

// SanitizeURLPath returns a sanitized version of the URL path.
// Invalid characters are removed as are as incorrect path separators.
func SanitizeURLPath(rawPath string) string {
	s := SanitizeSeparators(rawPath)
	if strings.Contains(s, "://") {
		return ""
	}
	re, err := regexp.Compile(ReSanitizePath)
	if err != nil {
		return "regexp.Compile error: " + err.Error()
	}
	s = re.ReplaceAllString(s, "")
	return s
}

// SanitizeGitHub returns a sanitized version of the GitHub repository.
// The repo is trimmed of any invalid characters listed in the [GitHub documentation].
//
// [GitHub documentation]: https://docs.github.com/en/get-started/using-git/dealing-with-special-characters-in-branch-and-tag-names#naming-branches-and-tags
func SanitizeGitHub(repo string) string {
	s := SanitizeURLPath(repo)
	s = strings.TrimPrefix(s, "refs/")
	return s
}

// ValidDate returns three boolean values that indicate if the year, month, and day are valid.
// If any of the bool values are false, the date syntax is invalid and should not be used.
//
// The year must be between 1980 and the current year.
// If the year is not in use, the month and day must not be in use.
// And if the month is not in use, the day must not in use.
//
// A not in use value is either "0" or an empty string.
func ValidDate(y, m, d string) (bool, bool, bool) {
	yok, mok, dok := true, true, true
	current := time.Now().Year()

	year, err := strconv.Atoi(y)
	if err != nil {
		yok = false
	}
	useYear := year != 0 && y != ""
	validYear := year >= model.EpochYear && year <= current
	if useYear && !validYear {
		yok = false
	}

	month, err := strconv.Atoi(m)
	if err != nil {
		mok = false
	}
	useMonth := month != 0 && m != ""
	const jan, dec = 1, 12
	validMonth := month >= jan && month <= dec
	if useMonth && !validMonth {
		mok = false
	}

	day, err := strconv.Atoi(d)
	if err != nil {
		dok = false
	}
	useDay := day != 0 && d != ""
	const first, last = 1, 31
	validDay := day >= first && day <= last
	if useDay && !validDay {
		dok = false
	}

	if !useYear && (validMonth || validDay) {
		yok = false
	}
	if !useMonth && validDay {
		mok = false
	}
	return yok, mok, dok
}

// ValidVT returns true if the link is a valid VirusTotal URL
// or if it is an empty string.
func ValidVT(link string) bool {
	const expect = "https://www.virustotal.com/"
	if len(link) > 0 && !strings.HasPrefix(link, expect) {
		return false
	}
	return true
}
