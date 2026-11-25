//nolint:exhaustruct_v5
package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Defacto2/server/internal/nils"
)

const (
	filesTableName = "files"
	timeFmt        = "2006-01-02 15:04"
	neverLabel     = "never"
	dashLabel      = "—"
)

// IndexStat holds live PostgreSQL catalog and usage statistics for a single index.
type IndexStat struct {
	Name       string
	Method     string
	Primary    bool
	Valid      bool
	Partial    bool
	SizeBytes  int64
	Size       string
	Scans      int64
	TupRead    int64
	TupFetch   int64
	BlocksHit  int64
	BlocksRead int64
	LastScan   *time.Time
	Definition string
}

// CacheHitPct returns the shared-buffer cache hit percentage for this index.
func (s IndexStat) CacheHitPct() string {
	total := s.BlocksHit + s.BlocksRead
	if total <= 0 {
		return dashLabel
	}
	return fmt.Sprintf("%.1f%%", float64(s.BlocksHit)*100/float64(total))
}

// LastScanString returns the formatted timestamp of the last index scan, or "never".
func (s IndexStat) LastScanString() string {
	if s.LastScan == nil {
		return neverLabel
	}
	return s.LastScan.Format(timeFmt)
}

// ExprSummary strips the CREATE INDEX ... USING <method> prefix and noisy PostgreSQL
// internal type casts from pg_get_indexdef for compact display in admin tables.
func (s IndexStat) ExprSummary() string {
	expr := s.Definition
	marker := " USING " + s.Method + " "
	if _, after, ok := strings.Cut(expr, marker); ok {
		expr = after
	}
	r := strings.NewReplacer(
		"::character varying", "",
		"::regconfig", "",
		"::text", "",
	)
	return r.Replace(expr)
}

// ScansFormatted returns Scans formatted with thousand separators.
func (s IndexStat) ScansFormatted() string { return formatInt64(s.Scans) }

// TupReadFormatted returns TupRead formatted with thousand separators.
func (s IndexStat) TupReadFormatted() string { return formatInt64(s.TupRead) }

// TupFetchFormatted returns TupFetch formatted with thousand separators.
func (s IndexStat) TupFetchFormatted() string { return formatInt64(s.TupFetch) }

// TableIndexStats holds table-level storage/scan/HOT/vacuum summary alongside all active indexes.
type TableIndexStats struct {
	TableName   string
	TableSize   string
	IndexesSize string
	TotalSize   string
	SeqScan     int64
	IdxScan     int64
	LiveTuples  int64
	DeadTuples  int64
	Updates     int64
	HotUpdates  int64
	LastVacuum  *time.Time
	LastAnalyze *time.Time
	StatsReset  *time.Time
	Indexes     []IndexStat
}

// IdxScanPct returns the percentage of table scans served by indexes.
func (t TableIndexStats) IdxScanPct() string {
	total := t.SeqScan + t.IdxScan
	if total <= 0 {
		return dashLabel
	}
	return fmt.Sprintf("%.1f%%", float64(t.IdxScan)*100/float64(total))
}

// HotUpdatePct returns the percentage of UPDATEs handled as Heap-Only Tuple (HOT) updates.
func (t TableIndexStats) HotUpdatePct() string {
	if t.Updates <= 0 {
		return dashLabel
	}
	return fmt.Sprintf("%.1f%%", float64(t.HotUpdates)*100/float64(t.Updates))
}

// LastVacuumString returns the formatted timestamp of the last vacuum/autovacuum.
func (t TableIndexStats) LastVacuumString() string {
	if t.LastVacuum == nil {
		return neverLabel
	}
	return t.LastVacuum.Format(timeFmt)
}

// LastAnalyzeString returns the formatted timestamp of the last analyze/autoanalyze.
func (t TableIndexStats) LastAnalyzeString() string {
	if t.LastAnalyze == nil {
		return neverLabel
	}
	return t.LastAnalyze.Format(timeFmt)
}

// StatsResetString returns the formatted timestamp of the last database stats reset.
func (t TableIndexStats) StatsResetString() string {
	if t.StatsReset == nil {
		return ""
	}
	return t.StatsReset.Format(timeFmt)
}

func (t TableIndexStats) SeqScanFormatted() string    { return formatInt64(t.SeqScan) }
func (t TableIndexStats) IdxScanFormatted() string    { return formatInt64(t.IdxScan) }
func (t TableIndexStats) LiveTuplesFormatted() string { return formatInt64(t.LiveTuples) }
func (t TableIndexStats) DeadTuplesFormatted() string { return formatInt64(t.DeadTuples) }
func (t TableIndexStats) UpdatesFormatted() string    { return formatInt64(t.Updates) }
func (t TableIndexStats) HotUpdatesFormatted() string { return formatInt64(t.HotUpdates) }

const (
	indexStatsErrFmt = "postgresql index stats %s: %w"

	tableStatsSQL = `
		SELECT
			pg_size_pretty(pg_table_size('files')),
			pg_size_pretty(pg_indexes_size('files')),
			pg_size_pretty(pg_total_relation_size('files')),
			COALESCE(t.seq_scan, 0),
			COALESCE(t.idx_scan, 0),
			COALESCE(t.n_live_tup, 0),
			COALESCE(t.n_dead_tup, 0),
			COALESCE(t.n_tup_upd, 0),
			COALESCE(t.n_tup_hot_upd, 0),
			GREATEST(t.last_vacuum, t.last_autovacuum),
			GREATEST(t.last_analyze, t.last_autoanalyze),
			d.stats_reset
		FROM pg_stat_user_tables t
		CROSS JOIN pg_stat_database d
		WHERE t.relname = $1
		  AND d.datname = current_database();`

	indexListSQL = `
		SELECT
			s.indexrelname,
			am.amname,
			i.indisprimary,
			i.indisvalid,
			(i.indpred IS NOT NULL) AS is_partial,
			pg_relation_size(s.indexrelid) AS size_bytes,
			pg_size_pretty(pg_relation_size(s.indexrelid)) AS size_pretty,
			COALESCE(s.idx_scan, 0),
			COALESCE(s.idx_tup_read, 0),
			COALESCE(s.idx_tup_fetch, 0),
			COALESCE(io.idx_blks_hit, 0),
			COALESCE(io.idx_blks_read, 0),
			s.last_idx_scan,
			pg_get_indexdef(s.indexrelid)
		FROM pg_stat_user_indexes s
		JOIN pg_index i ON i.indexrelid = s.indexrelid
		JOIN pg_class c ON c.oid = s.indexrelid
		JOIN pg_am am ON am.oid = c.relam
		LEFT JOIN pg_statio_user_indexes io ON io.indexrelid = s.indexrelid
		WHERE s.relname = $1
		ORDER BY pg_relation_size(s.indexrelid) DESC, s.indexrelname ASC;`
)

// Indexes queries PostgreSQL catalog and usage statistics for all active indexes on the files table.
func Indexes(ctx context.Context, db *sql.DB) (TableIndexStats, error) {
	if err := nils.Check(ctx, db); err != nil {
		return TableIndexStats{}, fmt.Errorf(indexStatsErrFmt, "check", err)
	}

	stats, err := tableSummary(ctx, db)
	if err != nil {
		return TableIndexStats{}, err
	}

	stats.Indexes, err = indexList(ctx, db)
	if err != nil {
		return TableIndexStats{}, err
	}

	return stats, nil
}

func tableSummary(ctx context.Context, db *sql.DB) (TableIndexStats, error) {
	stats := TableIndexStats{TableName: filesTableName}
	var lastVac, lastAnl, statsReset sql.NullTime
	err := db.QueryRowContext(ctx, tableStatsSQL, filesTableName).Scan(
		&stats.TableSize,
		&stats.IndexesSize,
		&stats.TotalSize,
		&stats.SeqScan,
		&stats.IdxScan,
		&stats.LiveTuples,
		&stats.DeadTuples,
		&stats.Updates,
		&stats.HotUpdates,
		&lastVac,
		&lastAnl,
		&statsReset,
	)
	if err != nil {
		return TableIndexStats{}, fmt.Errorf(indexStatsErrFmt, "table summary", err)
	}
	stats.LastVacuum = nullTimePtr(lastVac)
	stats.LastAnalyze = nullTimePtr(lastAnl)
	stats.StatsReset = nullTimePtr(statsReset)
	return stats, nil
}

func indexList(ctx context.Context, db *sql.DB) ([]IndexStat, error) {
	rows, err := db.QueryContext(ctx, indexListSQL, filesTableName)
	if err != nil {
		return nil, fmt.Errorf(indexStatsErrFmt, "index list", err)
	}
	defer func() { _ = rows.Close() }()

	var indexes []IndexStat
	for rows.Next() {
		var idx IndexStat
		var lastScan sql.NullTime
		err := rows.Scan(
			&idx.Name,
			&idx.Method,
			&idx.Primary,
			&idx.Valid,
			&idx.Partial,
			&idx.SizeBytes,
			&idx.Size,
			&idx.Scans,
			&idx.TupRead,
			&idx.TupFetch,
			&idx.BlocksHit,
			&idx.BlocksRead,
			&lastScan,
			&idx.Definition,
		)
		if err != nil {
			return nil, fmt.Errorf(indexStatsErrFmt, "scan row", err)
		}
		idx.LastScan = nullTimePtr(lastScan)
		indexes = append(indexes, idx)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(indexStatsErrFmt, "rows", err)
	}
	return indexes, nil
}

func nullTimePtr(nt sql.NullTime) *time.Time {
	if !nt.Valid {
		return nil
	}
	t := nt.Time
	return &t
}

func formatInt64(n int64) string {
	s := strconv.FormatInt(n, 10)
	if len(s) <= 3 {
		return s
	}
	neg := strings.HasPrefix(s, "-")
	if neg {
		s = s[1:]
	}
	var b strings.Builder
	if neg {
		b.WriteByte('-')
	}
	rem := len(s) % 3
	if rem > 0 {
		b.WriteString(s[:rem])
		if len(s) > rem {
			b.WriteByte(',')
		}
	}
	for i := rem; i < len(s); i += 3 {
		if i > rem {
			b.WriteByte(',')
		}
		b.WriteString(s[i : i+3])
	}
	return b.String()
}
