// Package form provides functions for providing data for form and input elements.
package form

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"html/template"
	"net/url"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/Defacto2/server/handler/releaser"
	"github.com/Defacto2/server/internal/nils"
	"github.com/Defacto2/server/internal/tags"
	"github.com/Defacto2/server/model"
	"github.com/aarondl/sqlboiler/v4/boil"
)

var ErrFilename = errors.New("form: invalid filename")

// Checkname returns an error if the named file has directory traversal characters.
func Checkname(name string) error {
	if !filepath.IsLocal(name) {
		return ErrFilename
	}

	return nil
}

// HumanizeCount returns the human readable name of the platform and section tags combined
// and the number of existing artifacts. The number of existing artifacts is colored based on
// the count. If the count is 0, the text is red. If the count is 1, the text is blue. If the
// count is greater than 1, the text is unmodified.
func HumanizeCount(ctx context.Context, exec boil.ContextExecutor, section, platform string) (template.HTML, error) {
	const format = "form humanize count: %w"
	if err := nils.Check(ctx, exec); err != nil {
		return "", fmt.Errorf(format, err)
	}

	count, tag, err := humanizeCount(ctx, exec, section, platform)
	if err != nil {
		return "", err
	}

	var html string
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

// HumanizeCountStr returns the human readable name of the platform and section tags combined
// and the number of existing artifacts. Any errors are returned as a string.
func HumanizeCountStr(ctx context.Context, db *sql.DB, section, platform string) string {
	count, tag, err := humanizeCount(ctx, db, section, platform)
	if err != nil {
		return err.Error()
	}

	return fmt.Sprintf("%s, %d existing artifacts", tag, count)
}

func humanizeCount(ctx context.Context, exec boil.ContextExecutor, section, platform string) (int64, string, error) {
	s := tags.TagByURI(section)
	p := tags.TagByURI(platform)
	tag := tags.Humanize(p, s)
	if strings.HasPrefix(tag, "unknown") {
		switch {
		case p.String() == "" && s.String() == "":
			return 0, "please choose both classifications", nil
		case s.String() == "":
			return 0, "please choose a tag as category", nil
		case p.String() == "":
			return 0, "please choose an operating system", nil
		default:
			return 0, "unknown classification", nil
		}
	}

	count, err := model.CountTags(ctx, exec, s, p)
	if err != nil {
		return 0, "cannot count the classification",
			fmt.Errorf("form humanize and count classification %w", err)
	}

	return count, tag, nil
}

func SanitizeCredit(s string) string {
	if s == "" {
		return ""
	}

	creators := strings.Split(s, ",")
	for i, credit := range creators {
		creators[i] = releaser.Clean(credit)
	}

	return strings.Join(creators, ",")
}

// SanitizeFilename returns a sanitized version of the filename.
// The filename is trimmed of any leading or trailing whitespace,
// and any parent directory references are removed. Any Linux or
// Windows directory separators are replaced with a "-" hyphen.
func SanitizeFilename(name string) string {
	s := strings.TrimSpace(name)
	if s == "" {
		return ""
	}

	s = strings.ReplaceAll(s, "../", "")
	if s == "." || s == ".." {
		return ""
	}
	return strings.Map(mapper, s)
}

func mapper(r rune) rune {
	if r == '/' || r == '\\' {
		return '-'
	}
	return r
}

// SanitizeSeparators returns a sanitized version of the URL path.
// The path is trimmed of any URL scheme, host or query parameters, as well
// as any invalid path separators.
func SanitizeSeparators(rawPath string) string {
	const separator = "/"

	s := strings.TrimSpace(rawPath)
	s = strings.ReplaceAll(s, separator+separator, separator)
	s = strings.Trim(s, separator)

	u, err := url.Parse(s)
	if err != nil {
		return "sanitize separators url parse error: " + err.Error()
	}

	return u.Path
}

const sanitizePath = "[^a-zA-Z0-9-._/]+" // Regular expression to sanitize the URL path.
var reSanitizePath = regexp.MustCompile(sanitizePath)

// SanitizeURLPath returns a sanitized version of the URL path.
// Invalid characters are removed as are as incorrect path separators.
func SanitizeURLPath(rawPath string) string {
	if strings.Contains(rawPath, "://") {
		return ""
	}

	return SanitizeSeparators(
		reSanitizePath.ReplaceAllString(rawPath, ""))
}

// SanitizeGitHub returns a sanitized version of the GitHub repository.
// The repo is trimmed of any invalid characters listed in the GitHub documentation.
func SanitizeGitHub(repo string) string {
	return strings.TrimPrefix(SanitizeURLPath(repo), "refs/")
}

// ValidDate returns three boolean values that indicate if the year, month, and day are valid.
// If any of the bool values are false, the date syntax is invalid and should not be used.
//
// The year must be between 1980 and the current year.
// If the year is not in use, the month and day must not be in use.
// And if the month is not in use, the day must not in use.
//
// A not in use value is either "0" or an empty string.
func ValidDate(y, m, d string) (bool, bool, bool) { //nolint:cyclop
	current := time.Now().Year()

	// NOTE: about the logic, by default no value is considered valid.
	yok, mok, dok := true, true, true

	year, err := strconv.Atoi(y)
	if err != nil {
		yok = false
	}
	month, err := strconv.Atoi(m)
	if err != nil {
		mok = false
	}
	day, err := strconv.Atoi(d)
	if err != nil {
		dok = false
	}

	useYear := year != 0 && y != ""
	useMonth := month != 0 && m != ""
	useDay := day != 0 && d != ""

	const jan, dec = 1, 12
	const first, last = 1, 31

	validYear := year >= model.EpochYear && year <= current
	if useYear && !validYear {
		yok = false
	}
	validMonth := month >= jan && month <= dec
	if useMonth && !validMonth {
		mok = false
	}
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
	const prefix = "https://www.virustotal.com/"

	link = strings.TrimSpace(link)
	return len(link) > 0 && strings.HasPrefix(link, prefix)
}

// ValidYouTube returns true when the string is either
// blank or is 11 characters.
func ValidYouTube(videoID string) bool {
	const require = 11
	return len(videoID) == 0 || len(videoID) == require
}
