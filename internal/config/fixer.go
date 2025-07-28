package config

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"time"

	"github.com/Defacto2/server/internal/command"
	"github.com/Defacto2/server/internal/out"
	"github.com/Defacto2/server/internal/postgres"
	"github.com/Defacto2/server/internal/zaplog"
	"github.com/Defacto2/server/model/fix"
)

// Fixer is used to fix any known issues with the file assets and the database entries.
func (c *Config) Fixer(w io.Writer, l *slog.Logger, d time.Time) error {
	if w == nil {
		w = io.Discard
	}
	msg := "postgres"
	db, err := postgres.Open()
	if err != nil {
		s := "fix could not initialize the database data"
		l.Error(msg,
			slog.String("issue", s),
			slog.String("error", err.Error()))
	}
	defer func() { _ = db.Close() }()
	var database postgres.Version
	if err := database.Query(db); err != nil {
		s := "version query problem"
		l.Error(msg,
			slog.String("issue", s),
			slog.String("error", err.Error()))
	}
	_, _ = fmt.Fprintf(w, "\n%+v\n", c)
	ctx := context.Background()
	count := RecordCount(ctx, db)
	const welcome = "Defacto2 web application"
	switch {
	case count == 0:
		s := welcome + " with no database records"
		l.Error(msg,
			slog.String("issue", s),
			slog.String("error", fmt.Sprint(err)))
	case MinimumFiles > count:
		s := welcome + " too few database records"
		l.Warn(msg,
			slog.String("issue", s),
			slog.Int("record count", count))
	default:
		s := fmt.Sprintf("%s using %d records", welcome, count)
		l.Info("fixer", slog.String("info", s))
	}
	c.repairer(ctx, db, l)
	c.sanityChecks(ctx)
	SanityTmpDir()
	l.Info("fixer", slog.Float64("time to completed", time.Since(d).Seconds()))
	return nil
}

// repairer is used to fix any known issues with the file assets and the database entries.
// These are skipped if the Production mode environment variable is set to false.
func (c *Config) repairer(ctx context.Context, db *sql.DB, l *slog.Logger) {
	if db == nil {
		panic(fmt.Errorf("%w: repairer", ErrPointer))
	}
	// logger := zaplog.Logger(ctx)
	if err := repairDatabase(ctx, db, l); err != nil {
		if errors.Is(err, ErrVer) {
			l.Warn("repair",
				slog.String("database", fmt.Sprintf("a %s, is the database server down?", ErrVer)))
		}
		l.Error("repair",
			slog.String("database", "could not initialize the database data"),
			slog.String("error", err.Error()))
	}
	// repair assets should be run after the database has been repaired, as it may rely on database data.
	if err := c.RepairAssets(ctx, db); err != nil {
		l.Error("repair assets",
			slog.String("error", err.Error()))
	}
}

// repairDatabase on startup checks the database connection and make any data corrections.
func repairDatabase(ctx context.Context, db *sql.DB, l *slog.Logger) error {
	if db == nil {
		panic(fmt.Errorf("%w: repair database", ErrPointer))
	}
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("repair database could not begin a transaction: %w", err)
	}
	if err := fix.Artifacts.Run(ctx, db, tx); err != nil {
		defer func() {
			if err := tx.Rollback(); err != nil {
				l.Error("repair database", slog.String("error", err.Error()))
			}
		}()
		return fmt.Errorf("repair database could not fix all artifacts: %w", err)
	}
	return nil
}

// sanityChecks is used to perform a number of sanity checks on the file assets and database.
// These are skipped if the Production mode environment variable is set.to false.
func (c *Config) sanityChecks(ctx context.Context) {
	logger := out.Devel()
	if err := c.Checks(logger); err != nil {
		logger.Error("check",
			slog.String("issue", "sanity checks could not read the environment variable, "+
				"it probably contains an invalid value"),
			slog.String("error", err.Error()))
	}
	cmdChecks(ctx)
	conn, err := postgres.New()
	if err != nil {
		logger.Error("check",
			slog.String("issue", "sanity checks could not initialize the database data"),
			slog.String("error", err.Error()))
		return
	}
	if err := conn.Validate(logger); err != nil {
		panic(fmt.Errorf("sanity check conn validate: %w", err))
	}
}

// checks is used to confirm the required commands are available.
// These are skipped if readonly is true.
func cmdChecks(ctx context.Context) {
	logger := zaplog.Logger(ctx)
	var buf strings.Builder
	for i, name := range command.Lookups() {
		if err := command.LookCmd(name); err != nil {
			buf.WriteString("\n\t\t\tmissing: " + name)
			buf.WriteString("\t" + command.Infos()[i])
		}
	}
	if buf.Len() > 0 {
		logger.Warnln("The following commands are required for the server to run in WRITE MODE",
			"\n\t\t\tThese need to be installed and accessible on the system path:"+
				"\t\t\t"+buf.String())
	}
	if err := command.LookupUnrar(); err != nil {
		if errors.Is(err, command.ErrVers) {
			logger.Warnf("Found unrar but " +
				"could not find unrar by Alexander Roshal, " +
				"is unrar-free mistakenly installed?")
			return
		}
		logger.Warnf("lookup unrar check: %s", err)
	}
}
