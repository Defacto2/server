package model

// Package file count.go contains the database queries for the counting of records and summing of column values.

import (
	"context"
	"fmt"
	"slices"
	"strings"

	namer "github.com/Defacto2/releaser/name"
	"github.com/Defacto2/server/internal/panics"
	"github.com/Defacto2/server/internal/postgres"
	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	"github.com/google/uuid"
)

// Count returns the total numbers of public artifact records.
func Count(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	panics.BoilExecCrash(exec)
	public, err := models.Files(qm.Where(ClauseNoSoftDel)).Count(ctx, exec)
	if err != nil {
		return 0, err
	}
	return public, nil
}

// Counts returns the total numbers of artifact records.
// The first result is the total number of public,
// the second is the number of non-public records.
// The final number is the number of new uploads waiting for approval.
func Counts(ctx context.Context, exec boil.ContextExecutor) (int64, int64, int64, error) {
	panics.BoilExecCrash(exec)
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
		qm.WithDeleted(),
	).Count(ctx, exec)
	return all, public, uploads, err
}

// CategoryCount counts the files that match the named category.
func CategoryCount(ctx context.Context, exec boil.ContextExecutor, name string) (int64, error) {
	panics.BoilExecCrash(exec)
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
	panics.BoilExecCrash(exec)
	if name == "" {
		return 0, ErrName
	}
	var s Summary
	mods := qm.SQL(string(postgres.SumSection()), null.StringFrom(name))
	err := models.NewQuery(mods, qm.Select(postgres.SumSize)).Bind(ctx, exec, &s)
	if err != nil {
		return 0, fmt.Errorf("bytecount by category %q: %w", name, err)
	}
	return s.SumBytes.Int64, nil
}

// ClassificationCount counts the files that match the named category and platform.
func ClassificationCount(ctx context.Context, exec boil.ContextExecutor, section, platform string) (int64, error) {
	panics.BoilExecCrash(exec)
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
	panics.BoilExecCrash(exec)
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
	panics.BoilExecCrash(exec)
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
	panics.BoilExecCrash(exec)
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

// UUIDVers contains the UUID version usage statistics.
type UUIDVers struct {
	V1      int
	V2      int
	V3      int
	V4      int
	V5      int
	V6      int
	V7      int
	V8      int
	Count   int // Count is the total number of UUIDs parsed.
	Error   int // Error are the UUIDs parsed that returned an error.
	Unknown int // Unknown are the UUIDs parsed that returned an out of range version.
}

func (u UUIDVers) String() string {
	s := []string{}
	if n := u.V1; n > 0 {
		s = append(s, fmt.Sprintf("V1: %d", n))
	}
	if n := u.V2; n > 0 {
		s = append(s, fmt.Sprintf("V2: %d", n))
	}
	if n := u.V3; n > 0 {
		s = append(s, fmt.Sprintf("V3: %d", n))
	}
	if n := u.V4; n > 0 {
		s = append(s, fmt.Sprintf("V4: %d", n))
	}
	if n := u.V5; n > 0 {
		s = append(s, fmt.Sprintf("V5: %d", n))
	}
	if n := u.V6; n > 0 {
		s = append(s, fmt.Sprintf("V6: %d", n))
	}
	if n := u.V7; n > 0 {
		s = append(s, fmt.Sprintf("V7: %d", n))
	}
	if n := u.V8; n > 0 {
		s = append(s, fmt.Sprintf("V8: %d", n))
	}
	if n := u.Error; n > 0 {
		s = append(s, fmt.Sprintf("errors: %d", n))
	}
	if n := u.Unknown; n > 0 {
		s = append(s, fmt.Sprintf("unknown: %d", n))
	}
	return strings.Join(s, ", ")
}

// UUIDs returns the counts of the UUID versions in use, ranging from V1 to V8.
func UUIDs(ctx context.Context, exec boil.ContextExecutor) (UUIDVers, error) {
	const msg = "count uuids"
	vers := UUIDVers{
		V1: 0, V2: 0, V3: 0, V4: 0, V5: 0, V6: 0, V7: 0, V8: 0,
		Count: 0, Error: 0, Unknown: 0,
	}
	uuids, err := UUID(ctx, exec)
	if err != nil {
		return vers, fmt.Errorf("%s: %w", msg, err)
	}

	const v1, v2, v3, v4, v5, v6, v7, v8 = 1, 2, 3, 4, 5, 6, 7, 8
	for val := range slices.Values(uuids) {
		vers.Count++
		s := val.UUID.String
		id, err := uuid.Parse(s)
		if err != nil {
			vers.Error++
			continue
		}
		switch id.Version() {
		case v1:
			vers.V1++
		case v2:
			vers.V2++
		case v3:
			vers.V3++
		case v4:
			vers.V4++
		case v5:
			vers.V5++
		case v6:
			vers.V6++
		case v7:
			vers.V7++
		case v8:
			vers.V8++
		default:
			vers.Unknown++
		}
	}
	return vers, nil
}
