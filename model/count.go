//nolint:nonamedreturns
package model

// Package file count.go contains the database queries for the counting of records and summing of column values.

import (
	"context"
	"fmt"
	"slices"
	"strings"

	namer "github.com/Defacto2/server/handler/releaser/name"
	"github.com/Defacto2/server/internal/nils"
	"github.com/Defacto2/server/internal/postgres"
	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/Defacto2/server/internal/tags"
	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	"github.com/google/uuid"
)

// Count returns the total numbers of public artifact records.
func Count(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	const format = "count %s: %w"
	if err := nils.Check(ctx, exec); err != nil {
		return 0, fmt.Errorf(format, "check", err)
	}

	n, err := models.Files(qm.Where(ClauseNoSoftDel)).Count(ctx, exec)
	if err != nil {
		return 0, err
	}

	return n, nil
}

// Counts returns the total numbers of artifact records.
//
//   - Returned are the total number of artifacts, both public and private.
//   - The number of public artifacts.
//   - The number of artifacts waiting for approval.
func Counts(ctx context.Context, exec boil.ContextExecutor) (total, public, waiting int64, err error) {
	const format = "counts %s: %w"
	if err := nils.Check(ctx, exec); err != nil {
		return 0, 0, 0, fmt.Errorf(format, "check", err)
	}

	// todo: waitgroup?

	total, err = models.Files(qm.WithDeleted()).Count(ctx, exec)
	if err != nil {
		return total, public, waiting, err
	}

	public, err = models.Files(qm.Where(ClauseNoSoftDel)).Count(ctx, exec)
	if err != nil {
		return total, public, waiting, err
	}

	waiting, err = models.Files(
		models.FileWhere.Deletedat.IsNotNull(),
		models.FileWhere.Deletedby.IsNull(),
		qm.WithDeleted(),
	).Count(ctx, exec)

	return total, public, waiting, err
}

// CountTags counts the files that match the named category and platform.
func CountTags(ctx context.Context, exec boil.ContextExecutor, section, platform tags.Tag) (int64, error) {
	const format = "count tags %s: %w"
	if err := nils.Check(ctx, exec); err != nil {
		return 0, fmt.Errorf(format, "check", err)
	}

	ss := section.String()
	ps := platform.String()
	if ss == "" {
		return 0, fmt.Errorf(format, section, ErrName)
	}
	if ps == "" {
		return 0, fmt.Errorf(format, platform, ErrName)
	}
	if section == platform {
		return 0, fmt.Errorf(format, ss+" = "+ps, ErrName)
	}

	cmod := models.FileWhere.Section.EQ(null.StringFrom(ss))
	pmod := models.FileWhere.Platform.EQ(null.StringFrom(ps))
	n, err := models.Files(cmod, pmod).Count(ctx, exec)
	if err != nil {
		return 0, fmt.Errorf(format, ss+" + "+ps, err)
	}

	return n, nil
}

// CountPlatform counts the files that match the platform tag.
func CountPlatform(ctx context.Context, exec boil.ContextExecutor, tag tags.Tag) (int64, error) {
	const format = "platform: %w"
	n, err := plat.count(ctx, exec, tag)
	if err != nil {
		return 0, fmt.Errorf(format, err)
	}

	return n, nil
}

// CountSection counts the files that match the section tag.
func CountSection(ctx context.Context, exec boil.ContextExecutor, tag tags.Tag) (int64, error) {
	const format = "section: %w"
	n, err := sect.count(ctx, exec, tag)
	if err != nil {
		return 0, fmt.Errorf(format, err)
	}

	return n, nil
}

// SumReleaser sums the byte file sizes for all the files that match the named release or group.
func SumReleaser(ctx context.Context, exec boil.ContextExecutor, name string) (int64, error) {
	const format = "sum releaser %s: %w"
	if err := nils.Check(ctx, exec); err != nil {
		return 0, fmt.Errorf(format, "check", err)
	}

	if name == "" {
		return 0, fmt.Errorf(format, "no name", ErrName)
	}

	hum, err := namer.Humanize(namer.Path(name))
	if err != nil {
		return 0, fmt.Errorf(format, name, err)
	}

	s := strings.ToUpper(hum)
	mods := qm.SQL(string(postgres.SumGroup()), null.StringFrom(s))

	n, err := models.Files(mods).Count(ctx, exec)
	if err != nil {
		return 0, fmt.Errorf(format, name, err)
	}

	return n, nil
}

// SumPlatform sums the file byte sizes of the files that match the platform tag.
func SumPlatform(ctx context.Context, exec boil.ContextExecutor, tag tags.Tag) (int64, error) {
	const format = "platform: %w"
	n, err := plat.sum(ctx, exec, tag)
	if err != nil {
		return 0, fmt.Errorf(format, err)
	}

	return n, nil
}

// SumSection sums the file byte sizes of the files that match the section tag.
func SumSection(ctx context.Context, exec boil.ContextExecutor, tag tags.Tag) (int64, error) {
	const format = "section: %w"
	n, err := sect.sum(ctx, exec, tag)
	if err != nil {
		return 0, fmt.Errorf(format, err)
	}

	return n, nil
}

type category int

const (
	plat category = iota
	sect
)

func (cat category) count(ctx context.Context, exec boil.ContextExecutor, tag tags.Tag) (int64, error) {
	const format = "count tag category %s: %w"
	if err := nils.Check(ctx, exec); err != nil {
		return 0, fmt.Errorf(format, "check", err)
	}

	s := tag.String()
	if s == "" {
		return 0, fmt.Errorf(format, tag, ErrName)
	}

	var mod qm.QueryMod
	switch cat {
	case plat:
		mod = models.FileWhere.Platform.EQ(null.StringFrom(s))
	case sect:
		mod = models.FileWhere.Section.EQ(null.StringFrom(s))
	}

	n, err := models.Files(mod).Count(ctx, exec)
	if err != nil {
		return 0, fmt.Errorf(format, tag, err)
	}

	return n, nil
}

func (cat category) sum(ctx context.Context, exec boil.ContextExecutor, tag tags.Tag) (int64, error) {
	const format = "sum tag category %s: %w"
	if err := nils.Check(ctx, exec); err != nil {
		return 0, fmt.Errorf(format, "check", err)
	}

	s := tag.String()
	if s == "" {
		return 0, fmt.Errorf(format, tag, ErrName)
	}

	var mod qm.QueryMod
	switch cat {
	case plat:
		mod = qm.SQL(string(postgres.SumPlatform()), null.StringFrom(s))
	case sect:
		mod = qm.SQL(string(postgres.SumSection()), null.StringFrom(s))
	}

	var obj Summary
	err := models.NewQuery(mod, qm.Select(postgres.SumSize)).Bind(ctx, exec, &obj)
	if err != nil {
		return 0, fmt.Errorf(format, s, err)
	}

	n := obj.SumBytes.Int64

	return n, nil
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

// UUIDs counts the different UUID versions in use in the database, ranging from V1 to V8.
func UUIDs(ctx context.Context, exec boil.ContextExecutor) (UUIDVers, error) {
	const format = "count uuids %s: %w"
	if err := nils.Check(ctx, exec); err != nil {
		return UUIDVers{}, fmt.Errorf(format, "check", err)
	}

	fs, err := UUID(ctx, exec)
	if err != nil {
		return UUIDVers{}, fmt.Errorf(format, "all uuids", err)
	}
	if len(fs) == 0 {
		return UUIDVers{}, nil //nolint:exhaustruct
	}

	uuids := UUIDVers{
		V1: 0, V2: 0, V3: 0, V4: 0, V5: 0, V6: 0, V7: 0, V8: 0, // counts
		Count: 0, Error: 0, Unknown: 0, // stats
	}

	const v1, v2, v3, v4, v5, v6, v7, v8 = 1, 2, 3, 4, 5, 6, 7, 8
	for record := range slices.Values(fs) {
		uuids.Count++

		s := record.UUID.String
		id, err := uuid.Parse(s)
		if err != nil {
			uuids.Error++
			continue
		}

		switch id.Version() {
		case v1:
			uuids.V1++
		case v2:
			uuids.V2++
		case v3:
			uuids.V3++
		case v4:
			uuids.V4++
		case v5:
			uuids.V5++
		case v6:
			uuids.V6++
		case v7:
			uuids.V7++
		case v8:
			uuids.V8++
		default:
			uuids.Unknown++
		}
	}

	return uuids, nil
}
