//nolint:gochecknoglobals
package fix

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/Defacto2/server/internal/nils"
)

type indexDef struct {
	name string
	ddl  string
}

const (
	trgmIndexName = "idx_files_filename_trgm"
	trgmIndexDDL  = "CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_files_filename_trgm " +
		"ON files USING GIN (filename gin_trgm_ops) " +
		"WHERE deletedat IS NULL"
)

var indexDefs = []indexDef{
	{
		name: "idx_files_desc_fts",
		ddl: "CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_files_desc_fts " +
			"ON files USING GIN (to_tsvector('english', COALESCE(record_title, '') || ' ' || COALESCE(comment, ''))) " +
			"WHERE deletedat IS NULL",
	},
	{
		name: "idx_files_section_date",
		ddl: "CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_files_section_date " +
			"ON files (section, date_issued_year, date_issued_month, date_issued_day) " +
			"WHERE deletedat IS NULL",
	},
	{
		name: "idx_files_platform_date",
		ddl: "CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_files_platform_date " +
			"ON files (platform, date_issued_year, date_issued_month, date_issued_day) " +
			"WHERE deletedat IS NULL",
	},
	{
		name: "idx_files_upper_group_for",
		ddl: "CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_files_upper_group_for " +
			"ON files (upper(group_brand_for)) " +
			"WHERE deletedat IS NULL",
	},
	{
		name: "idx_files_upper_group_by",
		ddl: "CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_files_upper_group_by " +
			"ON files (upper(group_brand_by)) " +
			"WHERE deletedat IS NULL AND group_brand_by IS NOT NULL",
	},
	{
		name: "idx_files_uuid_lower_pattern",
		ddl: "CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_files_uuid_lower_pattern " +
			"ON files (lower(uuid) text_pattern_ops)",
	},
	{
		name: "idx_files_integrity",
		ddl: "CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_files_integrity " +
			"ON files (file_integrity_strong) " +
			"WHERE file_integrity_strong IS NOT NULL",
	},
	{
		name: "idx_files_demozoo",
		ddl: "CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_files_demozoo " +
			"ON files (web_id_demozoo) " +
			"WHERE web_id_demozoo IS NOT NULL",
	},
	{
		name: "idx_files_pouet",
		ddl: "CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_files_pouet " +
			"ON files (web_id_pouet) " +
			"WHERE web_id_pouet IS NOT NULL",
	},
}

var obsoleteIndexes = []string{
	"idx_16386_pubfor_pubby_pubedition_filename_comment",
	"idx_16386_browsing",
	"idx_files_uuid",
	"idx_files_section_platform_date",
}

// IndexPlan describes the current PostgreSQL index state on files and what Indexes will change.
type IndexPlan struct {
	Existing   []string // recommended indexes already present and valid
	Invalid    []string // indexes on files marked indisvalid = false
	Create     []string // recommended indexes missing (excluding trgm)
	CreateTrgm bool     // true if idx_files_filename_trgm is missing
	NeedTrgm   bool     // true if the pg_trgm extension is not installed yet
	Drop       []string // obsolete indexes currently present on files
}

// PlanIndexes inspects PostgreSQL's system catalogs (read-only) to determine which
// indexes need to be created, dropped, or repaired on the files table.
func PlanIndexes(ctx context.Context, db *sql.DB) (IndexPlan, error) {
	const format = "plan indexes %s: %w"
	if err := nils.Check(ctx, db); err != nil {
		return IndexPlan{}, fmt.Errorf(format, "check", err)
	}

	const idxQuery = `
		SELECT c.relname, i.indisvalid
		FROM pg_index i
		JOIN pg_class c ON c.oid = i.indexrelid
		JOIN pg_class t ON t.oid = i.indrelid
		JOIN pg_namespace n ON n.oid = t.relnamespace
		WHERE n.nspname = current_schema() AND t.relname = 'files'
		ORDER BY c.relname`

	rows, err := db.QueryContext(ctx, idxQuery)
	if err != nil {
		return IndexPlan{}, fmt.Errorf(format, "query pg_index", err)
	}
	defer rows.Close()

	present := make(map[string]bool)
	var plan IndexPlan
	for rows.Next() {
		var name string
		var valid bool
		if err := rows.Scan(&name, &valid); err != nil {
			return IndexPlan{}, fmt.Errorf(format, "scan pg_index", err)
		}
		present[name] = valid
		if !valid {
			plan.Invalid = append(plan.Invalid, name)
		}
	}
	if err := rows.Err(); err != nil {
		return IndexPlan{}, fmt.Errorf(format, "rows pg_index", err)
	}

	var hasTrgm bool
	if err := db.QueryRowContext(ctx,
		"SELECT EXISTS(SELECT 1 FROM pg_extension WHERE extname = 'pg_trgm')",
	).Scan(&hasTrgm); err != nil {
		return IndexPlan{}, fmt.Errorf(format, "check pg_trgm", err)
	}
	plan.NeedTrgm = !hasTrgm

	for _, idx := range indexDefs {
		if valid, ok := present[idx.name]; ok && valid {
			plan.Existing = append(plan.Existing, idx.name)
		} else {
			plan.Create = append(plan.Create, idx.name)
		}
	}

	if valid, ok := present[trgmIndexName]; ok && valid {
		plan.Existing = append(plan.Existing, trgmIndexName)
	} else {
		plan.CreateTrgm = true
	}

	for _, name := range obsoleteIndexes {
		if _, ok := present[name]; ok {
			plan.Drop = append(plan.Drop, name)
		}
	}

	return plan, nil
}

// Indexes ensures all recommended PostgreSQL indexes exist on the files table,
// drops legacy MySQL-migrated indexes, rebuilds all indexes concurrently, and runs VACUUM ANALYZE.
//
// Must be called on *sql.DB outside a transaction block (required for CONCURRENTLY).
func Indexes(ctx context.Context, sl *slog.Logger, db *sql.DB) error {
	return RunIndexes(ctx, sl, db, false)
}

// RunIndexes inspects the current PostgreSQL indexes on files and either logs the
// planned operations (when dryRun is true) or executes them concurrently.
func RunIndexes(ctx context.Context, sl *slog.Logger, db *sql.DB, dryRun bool) error {
	const format = "fix indexes %s: %w"
	if err := nils.Check(ctx, sl, db); err != nil {
		return fmt.Errorf(format, "check", err)
	}

	plan, err := PlanIndexes(ctx, db)
	if err != nil {
		return fmt.Errorf(format, "plan", err)
	}

	const msg = "Database indexes"
	for _, name := range plan.Invalid {
		sl.Warn(msg, slog.String("invalid_index", name),
			slog.String("detail", "index is marked INVALID (likely from an interrupted CONCURRENTLY build)"))
	}

	if dryRun {
		return logIndexDryRun(sl, plan)
	}

	createSet := make(map[string]struct{}, len(plan.Create))
	for _, name := range plan.Create {
		createSet[name] = struct{}{}
	}
	for _, idx := range indexDefs {
		if _, missing := createSet[idx.name]; !missing {
			continue
		}
		sl.Info(msg, slog.String("create", idx.name))
		if _, err := db.ExecContext(ctx, idx.ddl); err != nil {
			return fmt.Errorf(format, idx.name, err)
		}
	}

	if plan.CreateTrgm {
		if plan.NeedTrgm {
			if _, err := db.ExecContext(ctx, "CREATE EXTENSION IF NOT EXISTS pg_trgm"); err != nil {
				sl.Warn(msg, slog.String("skip", trgmIndexName),
					slog.String("reason", "could not enable pg_trgm extension"), slog.Any("error", err))
				plan.CreateTrgm = false
			}
		}
		if plan.CreateTrgm {
			sl.Info(msg, slog.String("create", trgmIndexName))
			if _, err := db.ExecContext(ctx, trgmIndexDDL); err != nil {
				return fmt.Errorf(format, trgmIndexName, err)
			}
		}
	}

	for _, name := range plan.Drop {
		sl.Info(msg, slog.String("drop_obsolete", name))
		if _, err := db.ExecContext(ctx, "DROP INDEX CONCURRENTLY IF EXISTS "+name); err != nil {
			return fmt.Errorf(format, "drop "+name, err)
		}
	}

	sl.Info(msg, slog.String("task", "REINDEX TABLE CONCURRENTLY files"))
	if _, err := db.ExecContext(ctx, "REINDEX TABLE CONCURRENTLY files"); err != nil {
		return fmt.Errorf(format, "reindex", err)
	}

	sl.Info(msg, slog.String("task", "VACUUM ANALYZE files"))
	if err := Optimize(ctx, db); err != nil {
		return fmt.Errorf(format, "optimize", err)
	}

	return nil
}

func logIndexDryRun(sl *slog.Logger, plan IndexPlan) error {
	const msg = "[dry-run] Database indexes"
	sl.Info(msg, slog.Int("existing_valid", len(plan.Existing)),
		slog.Int("to_create", len(plan.Create)+boolInt(plan.CreateTrgm)),
		slog.Int("to_drop", len(plan.Drop)),
		slog.Int("invalid", len(plan.Invalid)))

	for _, name := range plan.Create {
		sl.Info(msg, slog.String("would_create", name))
	}
	if plan.CreateTrgm {
		if plan.NeedTrgm {
			sl.Info(msg, slog.String("would_enable_extension", "pg_trgm"))
		}
		sl.Info(msg, slog.String("would_create", trgmIndexName))
	}
	for _, name := range plan.Drop {
		sl.Info(msg, slog.String("would_drop", name))
	}
	sl.Info(msg, slog.String("would_run", "REINDEX TABLE CONCURRENTLY files"))
	sl.Info(msg, slog.String("would_run", "VACUUM ANALYZE files"))
	return nil
}

func boolInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
