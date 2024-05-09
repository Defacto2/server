package model

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	namer "github.com/Defacto2/releaser/name"
	"github.com/Defacto2/server/internal/postgres"
	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

// CategoryCount counts the files that match the named category.
func CategoryCount(ctx context.Context, db *sql.DB, name string) (int64, error) {
	if db == nil {
		return 0, ErrDB
	}
	if name == "" {
		return 0, ErrName
	}
	mods := models.FileWhere.Section.EQ(null.StringFrom(name))
	i, err := models.Files(mods).Count(ctx, db)
	if err != nil {
		return 0, fmt.Errorf("count by category %q: %w", name, err)
	}
	return i, nil
}

// CategoryByteSum sums the byte file sizes for all the files that match the named category.
func CategoryByteSum(ctx context.Context, db *sql.DB, name string) (int64, error) {
	if db == nil {
		return 0, ErrDB
	}
	if name == "" {
		return 0, ErrName
	}
	mods := qm.SQL(string(postgres.SumSection()), null.StringFrom(name))
	i, err := models.Files(mods).Count(ctx, db)
	if err != nil {
		return 0, fmt.Errorf("bytecount by category %q: %w", name, err)
	}
	return i, nil
}

// ClassificationCount counts the files that match the named category and platform.
func ClassificationCount(ctx context.Context, db *sql.DB, section, platform string) (int64, error) {
	if db == nil {
		return 0, ErrDB
	}
	if section == "" || platform == "" {
		return 0, ErrName
	}
	sect := models.FileWhere.Section.EQ(null.StringFrom(section))
	plat := models.FileWhere.Platform.EQ(null.StringFrom(platform))
	i, err := models.Files(sect, plat).Count(ctx, db)
	if err != nil {
		return 0, fmt.Errorf("count by classification %q %q: %w", section, platform, err)
	}
	return i, nil
}

// PlatformCount counts the files that match the named platform.
func PlatformCount(ctx context.Context, db *sql.DB, name string) (int64, error) {
	if db == nil {
		return 0, ErrDB
	}
	if name == "" {
		return 0, ErrName
	}
	mods := models.FileWhere.Platform.EQ(null.StringFrom(name))
	i, err := models.Files(mods).Count(ctx, db)
	if err != nil {
		return 0, fmt.Errorf("count by platform %q: %w", name, err)
	}
	return i, nil
}

// PlatformByteSum sums the byte filesizes for all the files that match the category name.
func PlatformByteSum(ctx context.Context, db *sql.DB, name string) (int64, error) {
	if db == nil {
		return 0, ErrDB
	}
	if name == "" {
		return 0, ErrName
	}
	mods := qm.SQL(string(postgres.SumPlatform()), null.StringFrom(name))
	i, err := models.Files(mods).Count(ctx, db)
	if err != nil {
		return 0, fmt.Errorf("bytecount by platform %q: %w", name, err)
	}
	return i, nil
}

// ReleaserByteSum sums the byte file sizes for all the files that match the group name.
func ReleaserByteSum(ctx context.Context, db *sql.DB, name string) (int64, error) {
	if db == nil {
		return 0, ErrDB
	}
	if name == "" {
		return 0, ErrName
	}
	s, err := namer.Humanize(namer.Path(name))
	if err != nil {
		return 0, fmt.Errorf("namer.Humanize: %w", err)
	}
	n := strings.ToUpper(s)
	mods := qm.SQL(string(postgres.SumGroup()), null.StringFrom(n))
	i, err := models.Files(mods).Count(ctx, db)
	if err != nil {
		return 0, fmt.Errorf("bytecount by releaser %q: %w", name, err)
	}
	return i, nil
}
