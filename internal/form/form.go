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
		return "cannot connect to the database", err
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
		return "cannot count the classification", err
	}
	return fmt.Sprintf("%s, %d existing artifacts", tag, count), nil
}