// Package form provides functions for providing data for form and input elements.
package form

import (
	"context"
	"fmt"
	"html/template"
	"strconv"
	"strings"
	"time"

	"github.com/Defacto2/server/internal/postgres"
	"github.com/Defacto2/server/internal/tags"
	"github.com/Defacto2/server/model"
)

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

func ValidDate(y, m, d string) (bool, bool, bool) {
	yOk, mOk, dOk := true, true, true

	year, err := strconv.Atoi(y)
	if err != nil {
		yOk = false
	}
	currentYear := time.Now().Year()
	useYear := year != 0 && y != ""
	validYear := year >= model.EpochYear && year <= currentYear
	if useYear && !validYear {
		yOk = false
	}

	month, err := strconv.Atoi(m)
	if err != nil {
		mOk = false
	}
	useMonth := month != 0 && m != ""
	validMonth := month >= 1 && month <= 12
	if useMonth && !validMonth {
		mOk = false
	}

	day, err := strconv.Atoi(d)
	if err != nil {
		dOk = false
	}
	useDay := day != 0 && d != ""
	validDay := day >= 1 && day <= 31
	if useDay && !validDay {
		dOk = false
	}

	if !useYear && (validMonth || validDay) {
		yOk = false
	}
	if !useMonth && validDay {
		mOk = false
	}
	return yOk, mOk, dOk
}
