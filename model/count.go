//nolint:nonamedreturns
package model

// Package file count.go contains the database queries for the counting of records and summing of column values.

import (
	"context"
	"fmt"
	"strings"

	namer "github.com/Defacto2/server/handler/releaser/name"
	"github.com/Defacto2/server/internal/nils"
	"github.com/Defacto2/server/internal/postgres"
	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/Defacto2/server/internal/tags"
	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
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

const countsSQL = `SELECT
	COUNT(*),
	COUNT(*) FILTER (WHERE deletedat IS NULL),
	COUNT(*) FILTER (WHERE deletedat IS NOT NULL AND deletedby IS NULL)
FROM files`

// Counts returns the total numbers of artifact records in a single table scan.
//
//   - Returned are the total number of artifacts, both public and private.
//   - The number of public artifacts.
//   - The number of artifacts waiting for approval.
func Counts(ctx context.Context, exec boil.ContextExecutor) (total, public, waiting int64, err error) {
	const format = "counts %s: %w"
	if err := nils.Check(ctx, exec); err != nil {
		return 0, 0, 0, fmt.Errorf(format, "check", err)
	}

	if err := exec.QueryRowContext(ctx, countsSQL).Scan(&total, &public, &waiting); err != nil {
		return 0, 0, 0, fmt.Errorf(format, "scan", err)
	}

	return total, public, waiting, nil
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

// TagStat holds the artifact count and total byte size for a single tag.
type TagStat struct {
	Count int64
	Bytes int64
}

const (
	sectionStatsSQL = "SELECT LOWER(BTRIM(section)), COUNT(*), COALESCE(SUM(filesize), 0) " +
		"FROM files WHERE deletedat IS NULL AND section IS NOT NULL GROUP BY LOWER(BTRIM(section))"
	platformStatsSQL = "SELECT LOWER(BTRIM(platform)), COUNT(*), COALESCE(SUM(filesize), 0) " +
		"FROM files WHERE deletedat IS NULL AND platform IS NOT NULL GROUP BY LOWER(BTRIM(platform))"
)

// PlatformStats returns the file count and total byte size grouped by platform tag.
func PlatformStats(ctx context.Context, exec boil.ContextExecutor) (map[string]TagStat, error) {
	const format = "platform stats: %w"
	stats, err := plat.stats(ctx, exec)
	if err != nil {
		return nil, fmt.Errorf(format, err)
	}
	return stats, nil
}

// SectionStats returns the file count and total byte size grouped by section tag.
func SectionStats(ctx context.Context, exec boil.ContextExecutor) (map[string]TagStat, error) {
	const format = "section stats: %w"
	stats, err := sect.stats(ctx, exec)
	if err != nil {
		return nil, fmt.Errorf(format, err)
	}
	return stats, nil
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

func (cat category) stats(ctx context.Context, exec boil.ContextExecutor) (map[string]TagStat, error) {
	const format = "tag stats %s: %w"
	if err := nils.Check(ctx, exec); err != nil {
		return nil, fmt.Errorf(format, "check", err)
	}

	query := sectionStatsSQL
	if cat == plat {
		query = platformStatsSQL
	}

	rows, err := exec.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf(format, "query", err)
	}
	defer func() {
		_ = rows.Close()
	}()

	stats := make(map[string]TagStat)
	for rows.Next() {
		var key string
		var stat TagStat
		if err := rows.Scan(&key, &stat.Count, &stat.Bytes); err != nil {
			return nil, fmt.Errorf(format, "scan", err)
		}
		stats[key] = stat
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(format, "rows", err)
	}

	return stats, nil
}

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

// UUIDs counts the different UUID versions in use in the database, ranging from V1 to V8,
// aggregated directly in PostgreSQL.
func UUIDs(ctx context.Context, exec boil.ContextExecutor) (UUIDVers, error) {
	const format = "count uuids %s: %w"
	if err := nils.Check(ctx, exec); err != nil {
		return UUIDVers{}, fmt.Errorf(format, "check", err)
	}

	const query = `
		SELECT
			CASE
				WHEN uuid IS NULL THEN 'err'
				WHEN pg_input_is_valid(BTRIM(uuid), 'uuid')
					THEN substr((BTRIM(uuid)::uuid)::text, 15, 1)
				ELSE 'err'
			END AS ver,
			COUNT(*) AS cnt
		FROM files
		WHERE deletedat IS NULL
		GROUP BY 1;`

	rows, err := exec.QueryContext(ctx, query)
	if err != nil {
		return UUIDVers{}, fmt.Errorf(format, "query", err)
	}
	defer func() { _ = rows.Close() }()

	uuids := UUIDVers{} //nolint:exhaustruct_v5
	for rows.Next() {
		var ver string
		var cnt int
		if err := rows.Scan(&ver, &cnt); err != nil {
			return UUIDVers{}, fmt.Errorf(format, "scan", err)
		}
		uuids.Count += cnt
		switch ver {
		case "1":
			uuids.V1 += cnt
		case "2":
			uuids.V2 += cnt
		case "3":
			uuids.V3 += cnt
		case "4":
			uuids.V4 += cnt
		case "5":
			uuids.V5 += cnt
		case "6":
			uuids.V6 += cnt
		case "7":
			uuids.V7 += cnt
		case "8":
			uuids.V8 += cnt
		case "err":
			uuids.Error += cnt
		default:
			uuids.Unknown += cnt
		}
	}
	if err := rows.Err(); err != nil {
		return UUIDVers{}, fmt.Errorf(format, "rows", err)
	}

	return uuids, nil
}
