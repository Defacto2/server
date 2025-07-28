package config

// Package file check.go contains the sanity check functions for the configuration values.

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/Defacto2/helper"
	"github.com/Defacto2/server/internal/dir"
	"github.com/Defacto2/server/internal/out"
	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/Defacto2/server/model"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
)

const (
	PortMax = 65534 // PortMax is the highest valid port number.
	PortSys = 1024  // PortSys is the lowest valid port number that does not require system access.

	DirWriteWriteBlock = 0o770 // Directory permissions.
)

var (
	ErrPortMax    = fmt.Errorf("http port value must be between 1-%d", PortMax)
	ErrPortSys    = fmt.Errorf("http port values between 1-%d require system access", PortSys)
	ErrDir        = errors.New("the directory path is not set")
	ErrDir404     = errors.New("the directory path does not exist")
	ErrDirIs      = errors.New("the directory path points to the file")
	ErrDirRead    = errors.New("the directory path could not be read")
	ErrDirFew     = errors.New("the directory path contains only a few items")
	ErrNoOAuth2   = errors.New("the production server requires a google, oauth2 client id to allow admin logins")
	ErrNoAccounts = errors.New("the production server has no google oauth2 user accounts to allow admin logins")
	ErrZap        = errors.New("the zap logger instance is nil")
	ErrSlog       = errors.New("the slog logger instance is nil")
)

// Checks runs a number of sanity checks for the environment variable configurations.
func (c *Config) Checks(logger *slog.Logger) error {
	if logger == nil {
		return ErrSlog
	}

	c.checkHTTP(logger)
	c.checkHTTPS(logger)
	c.production(logger)

	msg, key := "directory", "check"
	// Check the download, preview and thumbnail directories.
	if err := CheckDir(dir.Directory(c.AbsDownload), "downloads"); err != nil {
		s := helper.Capitalize(err.Error())
		logger.Error(msg, slog.String(key, s))
	}
	if err := CheckDir(dir.Directory(c.AbsPreview), "previews"); err != nil {
		s := helper.Capitalize(err.Error())
		logger.Error(msg, slog.String(key, s))
	}
	if err := CheckDir(dir.Directory(c.AbsThumbnail), "thumbnails"); err != nil {
		s := helper.Capitalize(err.Error())
		logger.Error(msg, slog.String(key, s))
	}
	if err := CheckDir(dir.Directory(c.AbsOrphaned), "orphaned"); err != nil {
		s := helper.Capitalize(err.Error())
		logger.Error(msg, slog.String(key, s))
	}
	if err := CheckDir(dir.Directory(c.AbsExtra), "extra"); err != nil {
		s := helper.Capitalize(err.Error())
		logger.Error(msg, slog.String(key, s))
	}
	msg = "information"
	// Reminds for the optional configuration values.
	if c.NoCrawl {
		s := "Disallow search engine crawling is enabled"
		logger.Warn(msg, slog.String(key, s))
	}
	if c.ReadOnly {
		s := "The server is running in read-only mode, edits to the database are not allowed"
		logger.Warn(msg, slog.String(key, s))
	}
	return c.SetupLogDir(logger)
}

// checkHTTP logs a fatal error if the HTTP port is invalid.
func (c *Config) checkHTTP(l *slog.Logger) {
	if c.HTTPPort == 0 {
		return
	}
	const msg, key = "http port", "port"
	if err := c.HTTPPort.Check(); err != nil {
		c.fatalPort(l, msg, key, err)
	}
}

// checkHTTPS logs a fatal error if the HTTPS port is invalid.
func (c *Config) checkHTTPS(l *slog.Logger) {
	if c.TLSPort == 0 {
		return
	}
	const msg, key = "https port", "port"
	if err := c.TLSPort.Check(); err != nil {
		c.fatalPort(l, msg, key, err)
	}
}

func (c *Config) fatalPort(l *slog.Logger, msg, key string, err error) {
	inf := "HTTP"
	if msg == "https port" {
		inf = "HTTPS"
	}
	switch {
	case errors.Is(err, ErrPortMax):
		out.Fatal(l, msg,
			slog.String("issue", "The server cannot use the "+inf+" port"),
			slog.Int(key, int(c.HTTPPort)),
			slog.String("error", err.Error()))
	case errors.Is(err, ErrPortSys):
		out.Fatal(l, msg,
			slog.String("issue", "The server cannot use the system port"),
			slog.Int(key, int(c.HTTPPort)),
			slog.String("error", err.Error()))
	}
}

// The production mode checks when not in read-only mode. It
// expects the server to be configured with OAuth2 and Google IDs.
// The server should be running over HTTPS and not unencrypted HTTP.
func (c *Config) production(l *slog.Logger) {
	if !bool(c.ProdMode) || bool(c.ReadOnly) {
		return
	}
	const msg, key = "production mode", "check"
	if c.GoogleClientID == "" {
		s := helper.Capitalize(ErrNoOAuth2.Error())
		l.Warn(msg, slog.String(key, s))
	}
	if c.GoogleIDs == "" && len(c.GoogleAccounts) == 0 {
		s := helper.Capitalize(ErrNoAccounts.Error())
		l.Warn(msg, slog.String(key, s))
	}
	if c.SessionMaxAge == 0 {
		s := "Sign-in client sessions last indefinately, this is a security risk"
		l.Warn(msg, slog.String(key, s))
	}
}

// LogStore determines the local storage path for all log files created by this web application.
func (c *Config) LogStore() error {
	logs := c.AbsLog.String()
	if logs == "" {
		dir, err := os.UserConfigDir()
		if err != nil {
			return fmt.Errorf("os.UserConfigDir: %w", err)
		}
		logs = filepath.Join(dir, ConfigDir)
	}
	if logsExists := helper.Stat(logs); !logsExists {
		if err := os.MkdirAll(logs, DirWriteWriteBlock); err != nil {
			return fmt.Errorf("%w: %s", err, logs)
		}
	}
	c.AbsLog = Abslog(logs)
	return nil
}

// SetupLogDir runs checks against the configured log directory.
// If no log directory is configured, a default directory is used.
// Problems will either log warnings or fatal errors.
func (c *Config) SetupLogDir(logger *slog.Logger) error {
	if logger == nil {
		return ErrSlog
	}
	if c.AbsLog == "" {
		if err := c.LogStore(); err != nil {
			return fmt.Errorf("%w: %w", ErrLog, err)
		}
	}
	logs := string(c.AbsLog)
	dir, err := os.Stat(logs)
	if os.IsNotExist(err) {
		return fmt.Errorf("log directory %w: %s", ErrDirNotExist, c.AbsLog)
	}
	if err != nil {
		return fmt.Errorf("log directory: %w", err)
	}
	if !dir.IsDir() {
		return fmt.Errorf("log directory %w: %s", ErrNotDir, dir.Name())
	}
	const msg, issue = "touch test", "Could not remove the empty test file in the log directory path"
	empty := filepath.Join(logs, ".defacto2_touch_test")
	if _, err := os.Stat(empty); os.IsNotExist(err) {
		f, err := os.Create(empty)
		if err != nil {
			return fmt.Errorf("log directory %w: %w", ErrTouch, err)
		}
		defer func(f *os.File) {
			_ = f.Close()
			if err := os.Remove(empty); err != nil {
				logger.Warn(msg,
					slog.String("issue", issue),
					slog.String("error", err.Error()),
					slog.String("path", empty))
				return
			}
		}(f)
		return nil
	}
	if err := os.Remove(empty); err != nil {
		logger.Warn(msg,
			slog.String("issue", issue),
			slog.String("error", err.Error()),
			slog.String("path", empty))
	}
	return nil
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
	if db == nil {
		return 0
	}
	fs, err := models.Files(qm.Where(model.ClauseNoSoftDel)).Count(ctx, db)
	if err != nil {
		return 0
	}
	return int(fs)
}

// SanityTmpDir is used to print the temporary directory and its disk usage.
func SanityTmpDir() {
	tmpdir := helper.TmpDir()
	du, err := helper.DiskUsage(tmpdir)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	hdu := helper.ByteCountFloat(du)
	_, _ = fmt.Fprintf(os.Stdout, "Temporary directory using, %s: %s\n", hdu, tmpdir)
}
