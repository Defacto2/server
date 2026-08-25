package config

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/Defacto2/helper"
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
	const welcome = "Defacto2 web application"
	const msg = "Fixing and repairs"

	db, err := postgres.Open()
	if err != nil {
		s := psl + " fix could not initialize the database data"
		sl.Error(s, slog.Any("error", err))
	}
	defer func() {
		err := db.Close()
		sl.Info(msg+" failed to close the database", slog.Any("error", err))
	}()

	var database postgres.Version
	if err := database.Query(db); err != nil {
		s := psl + " version query problem"
		sl.Error(s, slog.Any("error", err))
	}

	count := RecordCount(ctx, db)
	switch {
	case count == 0:
		s := psl + " " + msg + " with no database records"
		sl.Error(s, slog.Any("error", err))
	case MinimumFiles > count:
		s := welcome + " too few database records"
		sl.Warn(s, slog.Int64("record_count", count))
	default:
		sl.Info(welcome, slog.Int64("records", count))
	}

	c.fix(ctx, sl, db)
	if err := nils.Check(ctx, sl); err != nil {
		sl.Error(msg, slog.Any("error", err))
	}
	TempInfo(sl)

	sl.Info(msg, slog.String("task", "Time taken"),
		slog.Duration("time", time.Since(d).Round(time.Millisecond)))

	return nil
}

// TempInfo is used to print the temporary directory and its disk usage.
func TempInfo(sl *slog.Logger) {
	const format = "tmp info check %s: %w"
	if err := nils.Check(sl); err != nil {
		panic(fmt.Errorf(format, "check", err))
	}

	const msg = "Temporary directory"
	entries, err := os.ReadDir(os.TempDir())
	if err != nil {
		sl.Error(msg+" cannot read temp dir", slog.Any("error", err))
		return
	}
	var du int64
	for _, d := range entries {
		path := d.Name()
		if !d.IsDir() || !dir.IsTemp(path) {
			continue
		}
		n, err := helper.DiskUsage(path)
		if err != nil {
			slog.Debug("du cannot read path", slog.String("path", path), slog.Any("error", err))
			continue
		}
		du += n
	}

	hdu := helper.ByteCountFloat(du)
	tmpdir := filepath.Join(os.TempDir(), dir.Pattern+"-*")
	sl.Info(msg, slog.String("Path", tmpdir), slog.String("Usage", hdu))
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
func RecordCount(ctx context.Context, db *sql.DB) int64 {
	const msg = "record count"
	if err := nils.Check(ctx, db); err != nil {
		panic(fmt.Errorf("%s: %w", msg, err))
	}

	n, err := models.Files(qm.Where(model.ClauseNoSoftDel)).Count(ctx, db)
	if err != nil {
		return 0
	}

	return n
}

// fix is used to repair any known issues with the file assets and the database entries.
// These are skipped if the Production mode environment variable is set to false.
func (c *Config) fix(ctx context.Context, sl *slog.Logger, db *sql.DB) {
	const format = "config fix %s: %w"
	if err := nils.Check(ctx, sl, db); err != nil {
		panic(fmt.Errorf(format, "check", err))
	}

	const msg = "Fixing"
	if err := databaseFix(ctx, sl, db); err != nil {
		if errors.Is(err, ErrPSVersion) {
			value := fmt.Sprintf("a %s, is the database server down?", ErrPSVersion)
			sl.Warn(msg, slog.String("database", value))
		}
		sl.Error(msg, slog.String("database", "could not initialize the database data"),
			slog.Any("error", err))
	}

	// repair assets should be run after the database has been repaired, as it may rely on database data.
	if err := c.assets(ctx, sl, db); err != nil {
		sl.Error(msg, slog.Any("error", err))
	}
}

// databaseFix on startup checks the database connection and make any data corrections.
func databaseFix(ctx context.Context, sl *slog.Logger, db *sql.DB) error {
	const format = "database fix %s: %w"
	if err := nils.Check(ctx, sl, db); err != nil {
		panic(fmt.Errorf(format, "check", err))
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf(format, "begin transaction", err)
	}

	if err := fix.Artifacts.Run(ctx, sl, db, tx); err != nil {
		if err := tx.Rollback(); err != nil {
			const msg = "Cannot rollback database transaction during a repair"
			sl.Error(msg, slog.Any("error", err))
		}
		return fmt.Errorf(format, "cannot fix all artifacts", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf(format, "cannot commit transaction", err)
	}

	return nil
}
