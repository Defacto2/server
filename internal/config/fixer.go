package config

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"time"

	"github.com/Defacto2/server/internal/command"
	"github.com/Defacto2/server/internal/postgres"
	"github.com/Defacto2/server/model/fix"
)

// Fixer is used to fix any known issues with the file assets and the database entries.
func (c *Config) Fixer(w io.Writer, sl *slog.Logger, d time.Time) error {
	if sl == nil {
		return ErrNoSlog
	}
	if w == nil {
		w = io.Discard
	}
	msg := "postgres"
	db, err := postgres.Open()
	if err != nil {
		s := "fix could not initialize the database data"
		sl.Error(msg,
			slog.String("issue", s),
			slog.Any("error", err))
	}
	defer func() { _ = db.Close() }()
	var database postgres.Version
	if err := database.Query(db); err != nil {
		s := "version query problem"
		sl.Error(msg,
			slog.String("issue", s),
			slog.Any("error", err))
	}
	_, _ = fmt.Fprintf(w, "\n%+v\n", c)
	ctx := context.Background()
	count := RecordCount(ctx, db)
	const welcome = "Defacto2 web application"
	switch {
	case count == 0:
		s := welcome + " with no database records"
		sl.Error(msg,
			slog.String("issue", s),
			slog.Any("error", err))
	case MinimumFiles > count:
		s := welcome + " too few database records"
		sl.Warn(msg,
			slog.String("issue", s),
			slog.Int("record count", count))
	default:
		s := fmt.Sprintf("%s using %d records", welcome, count)
		sl.Info("fixer", slog.String("info", s))
	}
	c.repairer(ctx, db, sl)
	c.sanityChecks(sl)
	SanityTmpDir()
	sl.Info("fixer", slog.Float64("time to completed", time.Since(d).Seconds()))
	return nil
}

// repairer is used to fix any known issues with the file assets and the database entries.
// These are skipped if the Production mode environment variable is set to false.
func (c *Config) repairer(ctx context.Context, db *sql.DB, sl *slog.Logger) {
	if db == nil {
		panic(fmt.Errorf("%w: repairer", ErrPointer))
	}
	if err := repairDatabase(ctx, db, sl); err != nil {
		if errors.Is(err, ErrVer) {
			sl.Warn("repair",
				slog.String("database", fmt.Sprintf("a %s, is the database server down?", ErrVer)))
		}
		sl.Error("repair",
			slog.String("database", "could not initialize the database data"),
			slog.Any("error", err))
	}
	// repair assets should be run after the database has been repaired, as it may rely on database data.
	if err := c.RepairAssets(ctx, db, sl); err != nil {
		sl.Error("repair", slog.Any("error", err))
	}
}

// repairDatabase on startup checks the database connection and make any data corrections.
func repairDatabase(ctx context.Context, db *sql.DB, sl *slog.Logger) error {
	if db == nil {
		panic(fmt.Errorf("%w: repair database", ErrPointer))
	}
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("repair database could not begin a transaction: %w", err)
	}
	if err := fix.Artifacts.Run(ctx, db, tx, sl); err != nil {
		defer func() {
			if err := tx.Rollback(); err != nil {
				sl.Error("repair database", slog.Any("error", err))
			}
		}()
		return fmt.Errorf("repair database could not fix all artifacts: %w", err)
	}
	return nil
}

// sanityChecks is used to perform a number of sanity checks on the file assets and database.
// These are skipped if the Production mode environment variable is set.to false.
func (c *Config) sanityChecks(sl *slog.Logger) {
	if err := c.Checks(sl); err != nil {
		sl.Error("check",
			slog.String("issue", "sanity checks could not read the environment variable, "+
				"it probably contains an invalid value"),
			slog.Any("error", err))
	}
	cmdChecks(sl)
	conn, err := postgres.New()
	if err != nil {
		sl.Error("check",
			slog.String("issue", "sanity checks could not initialize the database data"),
			slog.Any("error", err))
		return
	}
	if err := conn.Validate(sl); err != nil {
		panic(fmt.Errorf("sanity check conn validate: %w", err))
	}
}

// checks is used to confirm the required commands are available.
// These are skipped if readonly is true.
func cmdChecks(sl *slog.Logger) {
	var attrs []slog.Attr
	for i, name := range command.Lookups() {
		if err := command.LookCmd(name); err != nil {
			attrs = append(attrs, slog.String(name, command.Infos()[i]))
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
		if errors.Is(err, command.ErrVers) {
			sl.Warn("command unrar",
				slog.String("invalid", "Found unrar but it is not authored by Alexander Roshal"),
				slog.String("incorrect application", "Is unrar-free mistakenly installed?"))
			return
		}
		sl.Warn("command unrar", slog.Any("error", err))
	}
}
