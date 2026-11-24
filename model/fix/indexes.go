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

// Indexes ensures all recommended PostgreSQL indexes exist on the files table,
// drops legacy MySQL-migrated indexes, rebuilds all indexes concurrently, and runs VACUUM ANALYZE.
//
// Must be called on *sql.DB outside a transaction block (required for CONCURRENTLY).
func Indexes(ctx context.Context, sl *slog.Logger, db *sql.DB) error {
	const format = "fix indexes %s: %w"
	if err := nils.Check(ctx, sl, db); err != nil {
		return fmt.Errorf(format, "check", err)
	}

	const msg = "Database indexes"
	for _, idx := range indexDefs {
		sl.Info(msg, slog.String("create", idx.name))
		if _, err := db.ExecContext(ctx, idx.ddl); err != nil {
			return fmt.Errorf(format, idx.name, err)
		}
	}

	// Filename trigram index requires pg_trgm (a trusted PostgreSQL extension).
	// If unavailable or lacking privileges, log a warning and continue.
	if _, err := db.ExecContext(ctx, "CREATE EXTENSION IF NOT EXISTS pg_trgm"); err != nil {
		sl.Warn(msg, slog.String("skip", "idx_files_filename_trgm"),
			slog.String("reason", "could not enable pg_trgm extension"), slog.Any("error", err))
	} else {
		const trgmDDL = "CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_files_filename_trgm " +
			"ON files USING GIN (filename gin_trgm_ops) " +
			"WHERE deletedat IS NULL"
		sl.Info(msg, slog.String("create", "idx_files_filename_trgm"))
		if _, err := db.ExecContext(ctx, trgmDDL); err != nil {
			return fmt.Errorf(format, "idx_files_filename_trgm", err)
		}
	}

	for _, name := range obsoleteIndexes {
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
