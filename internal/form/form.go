// Package form provides functions for providing data for form and input elements.
package form

import (
	"context"
	"fmt"
	"strings"

	"github.com/Defacto2/server/internal/postgres"
	"github.com/Defacto2/server/internal/tags"
	"github.com/Defacto2/server/model"
)

// HumanizeAndCount returns the human readable name of the platform and section tags combined
// and the number of existing artifacts.
func HumanizeAndCount(section, platform string) (string, error) {
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
	return fmt.Sprintf("%s, %d existing artifacts", tag, count), nil
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
