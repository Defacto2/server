package config

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/Defacto2/helper"
	"github.com/Defacto2/server/internal/command"
	"github.com/Defacto2/server/internal/dir"
	"github.com/Defacto2/server/internal/nils"
	"github.com/Defacto2/server/internal/postgres"
	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/Defacto2/server/model"
	"github.com/Defacto2/server/model/fix"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
)

// Fixer is used to fix any known issues with the file assets and the database entries.
func (c *Config) Fixer(ctx context.Context, sl *slog.Logger, d time.Time) error {
	psl := "PostgreSQL"
	if err := nils.Check(ctx, sl); err != nil {
		return fmt.Errorf("%s: %w", psl, err)
	}
	db, err := postgres.Open()
	if err != nil {
		s := "fix could not initialize the database data"
		sl.Error(psl,
			slog.String("issue", s),
			slog.Any("error", err))
	}
	defer func() { _ = db.Close() }()
	var database postgres.Version
	if err := database.Query(db); err != nil {
		s := "version query problem"
		sl.Error(psl,
			slog.String("issue", s),
			slog.Any("error", err))
	}
	count := RecordCount(ctx, db)
	const welcome = "Defacto2 web application"
	const msg = "Fixing and repairs"
	switch {
	case count == 0:
		s := welcome + " with no database records"
		sl.Error(psl,
			slog.String("issue", s),
			slog.Any("error", err))
	case MinimumFiles > count:
		s := welcome + " too few database records"
		sl.Warn(psl,
			slog.String("issue", s),
			slog.Int("record_count", count))
	default:
		sl.Info(msg, slog.String("info", welcome),
			slog.Int("records", count))
	}
	c.repairer(ctx, sl, db)
	c.sanityChecks(ctx, sl)
	TmpInfo(sl)
	sl.Info(msg, slog.String("task", "Time taken"),
		slog.Duration("time", time.Since(d).Round(time.Millisecond)))
	return nil
}

// TmpInfo is used to print the temporary directory and its disk usage.
func TmpInfo(sl *slog.Logger) {
	const msg = "tmp info check"
	if err := nils.Check(sl); err != nil {
		panic(fmt.Errorf("%s: %w", msg, err))
	}
	tmpdir := helper.TmpDir()
	du, err := helper.DiskUsage(tmpdir)
	if err != nil {
		sl.Error(msg, slog.String("disk usage", "could not obtain the tmp directory"),
			slog.String("tmp_directory", tmpdir), slog.Any("error", err))
		return
	}
	hdu := helper.ByteCountFloat(du)
	sl.Info("Temporary directory", slog.String("Path", tmpdir), slog.String("Usage", hdu))
}

// CheckDir runs checks against the named directory,
// including whether it exists, is a directory, and contains a minimum number of files.
// Problems will either log warnings or fatal errors.
func CheckDir(name dir.Directory, desc string) error {
	if name == "" {
		return fmt.Errorf("%s: %w", desc, ErrNoPath)
	}
	if err := name.IsDir(); err != nil {
		return fmt.Errorf("%q %q: %w", name, desc, err)
	}
	return nil
}

// RecordCount returns the number of records in the database.
func RecordCount(ctx context.Context, db *sql.DB) int {
	const msg = "record count"
	if err := nils.Check(ctx, db); err != nil {
		panic(fmt.Errorf("%s: %w", msg, err))
	}
	fs, err := models.Files(qm.Where(model.ClauseNoSoftDel)).Count(ctx, db)
	if err != nil {
		return 0
	}
	return int(fs)
}

// repairer is used to fix any known issues with the file assets and the database entries.
// These are skipped if the Production mode environment variable is set to false.
func (c *Config) repairer(ctx context.Context, sl *slog.Logger, db *sql.DB) {
	const msg = "Repairing"
	if err := nils.Check(ctx, sl, db); err != nil {
		panic(fmt.Errorf("%s: %w", msg, err))
	}
	if err := repairDatabase(ctx, sl, db); err != nil {
		if errors.Is(err, ErrPSVersion) {
			sl.Warn(msg,
				slog.String("database", fmt.Sprintf("a %s, is the database server down?", ErrPSVersion)))
		}
		sl.Error(msg,
			slog.String("database", "could not initialize the database data"),
			slog.Any("error", err))
	}
	// repair assets should be run after the database has been repaired, as it may rely on database data.
	if err := c.RepairAssets(ctx, sl, db); err != nil {
		sl.Error(msg, slog.Any("error", err))
	}
}

// repairDatabase on startup checks the database connection and make any data corrections.
func repairDatabase(ctx context.Context, sl *slog.Logger, db *sql.DB) error {
	const msg = "repair database"
	if err := nils.Check(ctx, sl, db); err != nil {
		panic(fmt.Errorf("%s: %w", msg, err))
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("%s could not begin a transaction: %w", msg, err)
	}
	if err := fix.Artifacts.Run(ctx, sl, db, tx); err != nil {
		if err := tx.Rollback(); err != nil {
			sl.Error(msg, slog.Any("error", err))
		}
		return fmt.Errorf("%s could not fix all artifacts: %w", msg, err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("%s could not commit the transaction: %w", msg, err)
	}
	return nil
}

// sanityChecks is used to perform a number of sanity checks on the file assets and database.
// These are skipped if the Production mode environment variable is set to false.
func (c *Config) sanityChecks(ctx context.Context, sl *slog.Logger) {
	const msg = "sanity check"
	if err := nils.Check(ctx, sl); err != nil {
		panic(fmt.Errorf("%s: %w", msg, err))
	}
	if err := c.Checks(ctx, sl); err != nil {
		sl.Error(msg,
			slog.String("issue", "sanity checks could not read the environment variable, "+
				"it probably contains an invalid value"),
			slog.Any("error", err))
	}
	cmdChecks(sl)
	conn, err := postgres.New()
	if err != nil {
		sl.Error(msg,
			slog.String("issue", "sanity checks could not initialize the database data"),
			slog.Any("error", err))
		return
	}
	if err := conn.Validate(sl); err != nil {
		panic(fmt.Errorf("%s conn validate: %w", msg, err))
	}
}

// checks is used to confirm the required commands are available.
// These are skipped if readonly is true.
func cmdChecks(sl *slog.Logger) {
	const msg = "command checks"
	if err := nils.Check(sl); err != nil {
		panic(fmt.Errorf("%s: %w", msg, err))
	}
	lookups := command.Lookups()
	infos := command.Infos()
	var attrs []slog.Attr
	for i, name := range lookups {
		if err := command.LookCmd(name); err != nil {
			attrs = append(attrs, slog.String(name, infos[i]))
		}
	}
	if len(attrs) > 0 {
		s := "The following commands are required for the server to run in WRITE MODE. " +
			"These need to be installed and accessible on the system path."
		sl.Warn("command lookups", slog.String("issue", s))
		for _, attr := range attrs {
			sl.Warn("missing command", slog.String(attr.Key, attr.Value.String()))
		}
	}
	if err := command.LookupUnrar(); err != nil {
		if errors.Is(err, command.ErrVersion) {
			sl.Warn("command unrar",
				slog.String("invalid", "Found unrar but it is not authored by Alexander Roshal"),
				slog.String("incorrect_application", "Is unrar-free mistakenly installed?"))
			return
		}
		sl.Warn("command unrar", slog.Any("error", err))
	}
}
