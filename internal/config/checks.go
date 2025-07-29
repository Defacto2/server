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

// Checks runs a number of sanity checks for the environment variable configurations.
func (c *Config) Checks(sl *slog.Logger) error {
	if sl == nil {
		return ErrNoSlog
	}
	c.checkHTTP(sl)
	c.checkHTTPS(sl)
	c.production(sl)
	msg, key := "directory", "check"
	// Check the download, preview and thumbnail directories.
	if err := CheckDir(dir.Directory(c.AbsDownload), "downloads"); err != nil {
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
	msg = "information"
	// Reminds for the optional configuration values.
	if c.NoCrawl {
		s := "Disallow search engine crawling is enabled"
		sl.Warn(msg, slog.String(key, s))
	}
	if c.ReadOnly {
		s := "The server is running in read-only mode, edits to the database are not allowed"
		sl.Warn(msg, slog.String(key, s))
	}
	return c.SetupLogDir(sl)
}

// checkHTTP logs a fatal error if the HTTP port is invalid.
func (c *Config) checkHTTP(sl *slog.Logger) {
	if c.HTTPPort == 0 {
		return
	}
	const msg, key = "http port", "port"
	if err := c.HTTPPort.Check(); err != nil {
		c.fatalPort(sl, msg, key, err)
	}
}

// checkHTTPS logs a fatal error if the HTTPS port is invalid.
func (c *Config) checkHTTPS(sl *slog.Logger) {
	if c.TLSPort == 0 {
		return
	}
	const msg, key = "https port", "port"
	if err := c.TLSPort.Check(); err != nil {
		c.fatalPort(sl, msg, key, err)
	}
}

func (c *Config) fatalPort(sl *slog.Logger, msg, key string, err error) {
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
	if !bool(c.ProdMode) || bool(c.ReadOnly) {
		return
	}
	const msg, key = "production mode", "check"
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
func (c *Config) SetupLogDir(ls *slog.Logger) error {
	if ls == nil {
		return ErrNoSlog
	}
	if c.AbsLog == "" {
		if err := c.LogStore(); err != nil {
			return fmt.Errorf("%w: %w", ErrLog, err)
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
				ls.Warn(msg,
					slog.String("issue", issue),
					slog.String("error", err.Error()),
					slog.String("path", empty))
				return
			}
		}(f)
		return nil
	}
	if err := os.Remove(empty); err != nil {
		ls.Warn(msg,
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
