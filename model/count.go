package model

// Package file count.go contains the database queries for the counting of records and summing of column values.

import (
	"context"
	"fmt"
	"strings"

	namer "github.com/Defacto2/releaser/name"
	"github.com/Defacto2/server/internal/postgres"
	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
)

// Counts returns the total numbers of artifact records.
// The first result is the total number of public,
// the second is the number of non-public records.
// The final number is the number of new uploads waiting for approval.
func Counts(ctx context.Context, exec boil.ContextExecutor) (int64, int64, int64, error) {
	if invalidExec(exec) {
		return 0, 0, 0, ErrDB
	}
	all, err := models.Files(qm.WithDeleted()).Count(ctx, exec)
	if err != nil {
		return 0, 0, 0, err
	}
	public, err := models.Files(qm.Where(ClauseNoSoftDel)).Count(ctx, exec)
	if err != nil {
		return 0, 0, 0, err
	}
	uploads, err := models.Files(
		models.FileWhere.Deletedat.IsNotNull(),
		models.FileWhere.Deletedby.IsNull(),
		qm.WithDeleted()).Count(ctx, exec)
	return all, public, uploads, err
}

// CategoryCount counts the files that match the named category.
func CategoryCount(ctx context.Context, exec boil.ContextExecutor, name string) (int64, error) {
	if invalidExec(exec) {
		return 0, ErrDB
	}
	if name == "" {
		return 0, ErrName
	}
	mods := models.FileWhere.Section.EQ(null.StringFrom(name))
	i, err := models.Files(mods).Count(ctx, exec)
	if err != nil {
		return 0, fmt.Errorf("count by category %q: %w", name, err)
	}
	return i, nil
}

// CategoryByteSum sums the byte file sizes for all the files that match the named category.
func CategoryByteSum(ctx context.Context, exec boil.ContextExecutor, name string) (int64, error) {
	if invalidExec(exec) {
		return 0, ErrDB
	}
	if name == "" {
		return 0, ErrName
	}
	mods := qm.SQL(string(postgres.SumSection()), null.StringFrom(name))
	i, err := models.Files(mods).Count(ctx, exec)
	if err != nil {
		return 0, fmt.Errorf("bytecount by category %q: %w", name, err)
	}
	return i, nil
}

// ClassificationCount counts the files that match the named category and platform.
func ClassificationCount(ctx context.Context, exec boil.ContextExecutor, section, platform string) (int64, error) {
	if invalidExec(exec) {
		return 0, ErrDB
	}
	if section == "" || platform == "" {
		return 0, ErrName
	}
	sect := models.FileWhere.Section.EQ(null.StringFrom(section))
	plat := models.FileWhere.Platform.EQ(null.StringFrom(platform))
	i, err := models.Files(sect, plat).Count(ctx, exec)
	if err != nil {
		return 0, fmt.Errorf("count by classification %q %q: %w", section, platform, err)
	}
	return i, nil
}

// PlatformCount counts the files that match the named platform.
func PlatformCount(ctx context.Context, exec boil.ContextExecutor, name string) (int64, error) {
	if invalidExec(exec) {
		return 0, ErrDB
	}
	if name == "" {
		return 0, ErrName
	}
	mods := models.FileWhere.Platform.EQ(null.StringFrom(name))
	i, err := models.Files(mods).Count(ctx, exec)
	if err != nil {
		return 0, fmt.Errorf("count by platform %q: %w", name, err)
	}
	return i, nil
}

// PlatformByteSum sums the byte filesizes for all the files that match the category name.
func PlatformByteSum(ctx context.Context, exec boil.ContextExecutor, name string) (int64, error) {
	if invalidExec(exec) {
		return 0, ErrDB
	}
	if name == "" {
		return 0, ErrName
	}
	mods := qm.SQL(string(postgres.SumPlatform()), null.StringFrom(name))
	i, err := models.Files(mods).Count(ctx, exec)
	if err != nil {
		return 0, fmt.Errorf("bytecount by platform %q: %w", name, err)
	}
	return i, nil
}

// ReleaserByteSum sums the byte file sizes for all the files that match the group name.
func ReleaserByteSum(ctx context.Context, exec boil.ContextExecutor, name string) (int64, error) {
	if invalidExec(exec) {
		return 0, ErrDB
	}
	if name == "" {
		return 0, ErrName
	}
	s, err := namer.Humanize(namer.Path(name))
	if err != nil {
		return 0, fmt.Errorf("releaser byte sum namer humanize: %w", err)
	}
	n := strings.ToUpper(s)
	mods := qm.SQL(string(postgres.SumGroup()), null.StringFrom(n))
	i, err := models.Files(mods).Count(ctx, exec)
	if err != nil {
		return 0, fmt.Errorf("bytecount by releaser %q: %w", name, err)
	}
	return i, nil
}
