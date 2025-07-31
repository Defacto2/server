package config

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/Defacto2/helper"
	"github.com/Defacto2/server/internal/command"
	"github.com/Defacto2/server/internal/dir"
	"github.com/Defacto2/server/internal/out"
	"github.com/Defacto2/server/internal/panics"
	"github.com/Defacto2/server/internal/postgres"
	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/Defacto2/server/model"
	"github.com/Defacto2/server/model/fix"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
)

// Checks runs a number of sanity checks for the environment variable configurations.
func (c *Config) Checks(sl *slog.Logger) error {
	const msg, key = "config directory", "check"
	if sl == nil {
		return fmt.Errorf("%s: %w", msg, panics.ErrNoSlog)
	}
	c.checkHTTP(sl)
	c.checkHTTPS(sl)
	c.production(sl)
	// Check the download, preview and thumbnail directories.
	println("CHECK DIRECTORY")
	if err := CheckDir(dir.Directory(c.AbsDownload), "downloads"); err != nil {
		println(">> ", err.Error())
		s := helper.Capitalize(err.Error())
		sl.Error(msg, slog.String(key, s))
	}
	if err := CheckDir(dir.Directory(c.AbsPreview), "previews"); err != nil {
		s := helper.Capitalize(err.Error())
		sl.Error(msg, slog.String(key, s))
	}
	if err := CheckDir(dir.Directory(c.AbsThumbnail), "thumbnails"); err != nil {
		s := helper.Capitalize(err.Error())
		sl.Error(msg, slog.String(key, s))
	}
	if err := CheckDir(dir.Directory(c.AbsOrphaned), "orphaned"); err != nil {
		s := helper.Capitalize(err.Error())
		sl.Error(msg, slog.String(key, s))
	}
	if err := CheckDir(dir.Directory(c.AbsExtra), "extra"); err != nil {
		s := helper.Capitalize(err.Error())
		sl.Error(msg, slog.String(key, s))
	}
	msgi := "information"
	// Reminds for the optional configuration values.
	if c.NoCrawl {
		s := "Disallow search engine crawling is enabled"
		sl.Warn(msgi, slog.String(key, s))
	}
	if c.ReadOnly {
		s := "The server is running in read-only mode, edits to the database are not allowed"
		sl.Warn(msgi, slog.String(key, s))
	}
	return c.SetupLogDir(sl)
}

// SetupLogDir runs checks against the configured log directory.
// If no log directory is configured, a default directory is used.
// Problems will either log warnings or fatal errors.
func (c *Config) SetupLogDir(sl *slog.Logger) error {
	const msg = "setup log directory"
	if sl == nil {
		return fmt.Errorf("%s: %w", msg, panics.ErrNoSlog)
	}
	if c.AbsLog == "" {
		if err := c.LogStore(); err != nil {
			return fmt.Errorf("%s: %w", msg, err)
		}
	}
	logs := string(c.AbsLog)
	dir, err := os.Stat(logs)
	if os.IsNotExist(err) {
		return fmt.Errorf("log directory %w: %s", ErrNoDir, c.AbsLog)
	}
	if err != nil {
		return fmt.Errorf("log directory: %w", err)
	}
	if !dir.IsDir() {
		return fmt.Errorf("log directory %w: %s", ErrNotDir, dir.Name())
	}
	const issue = "could not remove the empty test file in the log directory path"
	empty := filepath.Join(logs, ".defacto2_touch_test")
	if _, err := os.Stat(empty); os.IsNotExist(err) {
		f, err := os.Create(empty)
		if err != nil {
			return fmt.Errorf("log directory %w: %w", ErrTouch, err)
		}
		defer func(f *os.File) {
			_ = f.Close()
			if err := os.Remove(empty); err != nil {
				sl.Warn(msg,
					slog.String("issue", issue),
					slog.String("error", err.Error()),
					slog.String("path", empty))
				return
			}
		}(f)
		return nil
	}
	if err := os.Remove(empty); err != nil {
		sl.Warn(msg,
			slog.String("issue", issue),
			slog.String("error", err.Error()),
			slog.String("path", empty))
	}
	return nil
}

// checkHTTP logs a fatal error if the HTTP port is invalid.
func (c *Config) checkHTTP(sl *slog.Logger) {
	const msg, key = "check http port", "port"
	if sl == nil {
		panic(fmt.Errorf("%s: %w", msg, panics.ErrNoSlog))
	}
	if c.HTTPPort == 0 {
		return
	}
	if err := c.HTTPPort.Check(); err != nil {
		c.fatalPort(sl, msg, key, err)
	}
}

// checkHTTPS logs a fatal error if the HTTPS port is invalid.
func (c *Config) checkHTTPS(sl *slog.Logger) {
	const msg, key = "check https port", "port"
	if sl == nil {
		panic(fmt.Errorf("%s: %w", msg, panics.ErrNoSlog))
	}
	if c.TLSPort == 0 {
		return
	}
	if err := c.TLSPort.Check(); err != nil {
		c.fatalPort(sl, msg, key, err)
	}
}

func (c *Config) fatalPort(sl *slog.Logger, msg, key string, err error) {
	if sl == nil {
		panic(fmt.Errorf("config fatal port: %w", panics.ErrNoSlog))
	}
	inf := "HTTP"
	if msg == "https port" {
		inf = "HTTPS"
	}
	switch {
	case errors.Is(err, ErrPortMax):
		out.Fatal(sl, msg,
			slog.String("issue", "The server cannot use the "+inf+" port"),
			slog.Int(key, int(c.HTTPPort)),
			slog.String("error", err.Error()))
	case errors.Is(err, ErrPortSys):
		out.Fatal(sl, msg,
			slog.String("issue", "The server cannot use the system port"),
			slog.Int(key, int(c.HTTPPort)),
			slog.String("error", err.Error()))
	}
}

// The production mode checks when not in read-only mode. It
// expects the server to be configured with OAuth2 and Google IDs.
// The server should be running over HTTPS and not unencrypted HTTP.
func (c *Config) production(sl *slog.Logger) {
	const msg, key = "production mode", "check"
	if sl == nil {
		panic(fmt.Errorf("%s: %w", msg, panics.ErrNoSlog))
	}
	if !bool(c.ProdMode) || bool(c.ReadOnly) {
		return
	}
	if c.GoogleClientID == "" {
		s := helper.Capitalize(ErrNoOAuth2.Error())
		sl.Warn(msg, slog.String(key, s))
	}
	if c.GoogleIDs == "" && len(c.GoogleAccounts) == 0 {
		s := helper.Capitalize(ErrNoAccounts.Error())
		sl.Warn(msg, slog.String(key, s))
	}
	if c.SessionMaxAge == 0 {
		s := "Sign-in client sessions last indefinately, this is a security risk"
		sl.Warn(msg, slog.String(key, s))
	}
}

// Fixer is used to fix any known issues with the file assets and the database entries.
func (c *Config) Fixer(w io.Writer, sl *slog.Logger, d time.Time) error {
	msg := "postgres"
	if sl == nil {
		return fmt.Errorf("%s: %w", msg, panics.ErrNoSlog)
	}
	if w == nil {
		w = io.Discard
	}
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
	TmpInfo(sl)
	sl.Info("fixer", slog.Float64("time to completed", time.Since(d).Seconds()))
	return nil
}

// TmpInfo is used to print the temporary directory and its disk usage.
func TmpInfo(sl *slog.Logger) {
	const msg = "tmp info check"
	if sl == nil {
		panic(fmt.Errorf("%s: %w", msg, panics.ErrNoSlog))
	}
	tmpdir := helper.TmpDir()
	du, err := helper.DiskUsage(tmpdir)
	if err != nil {
		sl.Error(msg, slog.String("disk usage", "could not obtain the tmp directory"),
			slog.String("tmp directory", tmpdir), slog.Any("error", err))
		return
	}
	hdu := helper.ByteCountFloat(du)
	sl.Info("Temporary directory", slog.String("Path", tmpdir), slog.String("Usage", hdu))
}

// CheckDir runs checks against the named directory,
// including whether it exists, is a directory, and contains a minimum number of files.
// Problems will either log warnings or fatal errors.
func CheckDir(name dir.Directory, desc string) error {
	if err := name.IsDir(); err != nil {
		return fmt.Errorf("%w, %s: %s", err, desc, name)
	}
	return nil
}

// RecordCount returns the number of records in the database.
func RecordCount(ctx context.Context, db *sql.DB) int {
	const msg = "record count"
	if err := panics.CD(ctx, db); err != nil {
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
func (c *Config) repairer(ctx context.Context, db *sql.DB, sl *slog.Logger) {
	const msg = "repairer"
	if err := panics.CDS(ctx, db, sl); err != nil {
		panic(fmt.Errorf("%s: %w", msg, err))
	}
	if err := repairDatabase(ctx, db, sl); err != nil {
		if errors.Is(err, ErrPSqlVer) {
			sl.Warn("repairer",
				slog.String("database", fmt.Sprintf("a %s, is the database server down?", ErrPSqlVer)))
		}
		sl.Error("repairer",
			slog.String("database", "could not initialize the database data"),
			slog.Any("error", err))
	}
	// repair assets should be run after the database has been repaired, as it may rely on database data.
	if err := c.RepairAssets(ctx, db, sl); err != nil {
		sl.Error("repairer", slog.Any("error", err))
	}
}

// repairDatabase on startup checks the database connection and make any data corrections.
func repairDatabase(ctx context.Context, db *sql.DB, sl *slog.Logger) error {
	const msg = "repair database"
	if err := panics.CDS(ctx, db, sl); err != nil {
		panic(fmt.Errorf("%s: %w", msg, err))
	}
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("%s could not begin a transaction: %w", msg, err)
	}
	if err := fix.Artifacts.Run(ctx, db, tx, sl); err != nil {
		defer func() {
			if err := tx.Rollback(); err != nil {
				sl.Error(msg, slog.Any("error", err))
			}
		}()
		return fmt.Errorf("%s could not fix all artifacts: %w", msg, err)
	}
	return nil
}

// sanityChecks is used to perform a number of sanity checks on the file assets and database.
// These are skipped if the Production mode environment variable is set.to false.
func (c *Config) sanityChecks(sl *slog.Logger) {
	const msg = "sanity check"
	if sl == nil {
		panic(fmt.Errorf("%s: %w", msg, panics.ErrNoSlog))
	}
	if err := c.Checks(sl); err != nil {
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
	if sl == nil {
		panic(fmt.Errorf("%s: %w", msg, panics.ErrNoSlog))
	}
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
