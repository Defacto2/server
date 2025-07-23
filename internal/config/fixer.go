package config

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/Defacto2/server/internal/command"
	"github.com/Defacto2/server/internal/postgres"
	"github.com/Defacto2/server/internal/zaplog"
	"github.com/Defacto2/server/model/fix"
)

// Fixer is used to fix any known issues with the file assets and the database entries.
func (c *Config) Fixer(w io.Writer, d time.Time) error {
	if w == nil {
		w = io.Discard
	}
	logger := zaplog.Timestamp().Sugar()
	db, err := postgres.Open()
	if err != nil {
		logger.Errorf("fix could not initialize the database data: %s", err)
	}
	defer func() { _ = db.Close() }()
	var database postgres.Version
	if err := database.Query(db); err != nil {
		logger.Errorf("postgres version query: %w", err)
	}
	_, _ = fmt.Fprintf(w, "\n%+v\n", c)
	ctx := context.WithValue(context.Background(), zaplog.LoggerKey, logger)
	count := RecordCount(ctx, db)
	const welcome = "Defacto2 web application"
	switch {
	case count == 0:
		logger.Error(welcome + " with no database records")
	case MinimumFiles > count:
		logger.Warnf(welcome+" with only %d records, expecting at least %d+", count, MinimumFiles)
	default:
		logger.Infof(welcome+" using %d records", count)
	}
	c.repairer(ctx, db)
	c.sanityChecks(ctx)
	SanityTmpDir()
	logger.Infof("Fixer completed in %.1fs", time.Since(d).Seconds())
	return nil
}

// repairer is used to fix any known issues with the file assets and the database entries.
// These are skipped if the Production mode environment variable is set to false.
func (c *Config) repairer(ctx context.Context, db *sql.DB) {
	if db == nil {
		panic(fmt.Errorf("%w: repairer", ErrPointer))
	}
	logger := zaplog.Logger(ctx)
	if err := repairDatabase(ctx, db); err != nil {
		if errors.Is(err, ErrVer) {
			logger.Warnf("A %s, is the database server down?", ErrVer)
		}
		logger.Errorf("repair database could not initialize the database data: %s", err)
	}
	// repair assets should be run after the database has been repaired, as it may rely on database data.
	if err := c.RepairAssets(ctx, db); err != nil {
		logger.Errorf("asset repairs: %s", err)
	}
}

// repairDatabase on startup checks the database connection and make any data corrections.
func repairDatabase(ctx context.Context, db *sql.DB) error {
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
				logger := zaplog.Logger(ctx)
				logger.Error(err)
			}
		}()
		return fmt.Errorf("repair database could not fix all artifacts: %w", err)
	}
	return nil
}

// sanityChecks is used to perform a number of sanity checks on the file assets and database.
// These are skipped if the Production mode environment variable is set.to false.
func (c *Config) sanityChecks(ctx context.Context) {
	logger := zaplog.Logger(ctx)
	if err := c.Checks(logger); err != nil {
		logger.Errorf("sanity checks could not read the environment variable, "+
			"it probably contains an invalid value: %s", err)
	}
	cmdChecks(ctx)
	conn, err := postgres.New()
	if err != nil {
		logger.Errorf("sanity checks could not initialize the database data: %s", err)
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
