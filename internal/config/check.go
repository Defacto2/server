package config

// Package file check.go contains the sanity check functions for the configuration values.

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Defacto2/helper"
	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/Defacto2/server/model"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"go.uber.org/zap"
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
)

// Checks runs a number of sanity checks for the environment variable configurations.
func (c *Config) Checks(logger *zap.SugaredLogger) error {
	if logger == nil {
		return ErrZap
	}

	c.httpPort(logger)
	c.tlsPort(logger)
	c.production(logger)

	// Check the download, preview and thumbnail directories.
	if err := CheckDir(c.AbsDownload, "downloads"); err != nil {
		s := helper.Capitalize(err.Error())
		logger.Error(s)
	}
	if err := CheckDir(c.AbsPreview, "previews"); err != nil {
		s := helper.Capitalize(err.Error())
		logger.Error(s)
	}
	if err := CheckDir(c.AbsThumbnail, "thumbnails"); err != nil {
		s := helper.Capitalize(err.Error())
		logger.Error(s)
	}
	if err := CheckDir(c.AbsOrphaned, "orphaned"); err != nil {
		s := helper.Capitalize(err.Error())
		logger.Error(s)
	}
	if err := CheckDir(c.AbsExtra, "extra"); err != nil {
		s := helper.Capitalize(err.Error())
		logger.Error(s)
	}

	// Reminds for the optional configuration values.
	if c.NoCrawl {
		logger.Warn("Disallow search engine crawling is enabled")
	}
	if c.ReadOnly {
		logger.Warn("The server is running in read-only mode, edits to the database are not allowed")
	}

	return c.SetupLogDir(logger)
}

// httpPort returns an error if the HTTP port is invalid.
func (c Config) httpPort(logger *zap.SugaredLogger) {
	if c.HTTPPort == 0 {
		return
	}
	if err := Validate(c.HTTPPort); err != nil {
		switch {
		case errors.Is(err, ErrPortMax):
			logger.Fatalf("The server could not use the HTTP port %d, %s.",
				c.HTTPPort, err)
		case errors.Is(err, ErrPortSys):
			logger.Infof("The server HTTP port %d, %s.",
				c.HTTPPort, err)
		}
	}
}

// tlsPort returns an error if the TLS port is invalid.
func (c Config) tlsPort(logger *zap.SugaredLogger) {
	if c.TLSPort == 0 {
		return
	}
	if err := Validate(c.TLSPort); err != nil {
		switch {
		case errors.Is(err, ErrPortMax):
			logger.Fatalf("The server could not use the HTTPS port %d, %s.",
				c.TLSPort, err)
		case errors.Is(err, ErrPortSys):
			logger.Infof("The server HTTPS port %d, %s.",
				c.TLSPort, err)
		}
	}
}

// The production mode checks when not in read-only mode. It
// expects the server to be configured with OAuth2 and Google IDs.
// The server should be running over HTTPS and not unencrypted HTTP.
func (c Config) production(logger *zap.SugaredLogger) {
	if !c.ProdMode || c.ReadOnly {
		return
	}
	if c.GoogleClientID == "" {
		s := helper.Capitalize(ErrNoOAuth2.Error())
		logger.Warn(s)
	}
	if c.GoogleIDs == "" && len(c.GoogleAccounts) == 0 {
		s := helper.Capitalize(ErrNoAccounts.Error())
		logger.Warn(s)
	}
	if c.SessionMaxAge == 0 {
		logger.Warn("A signed in client session lasts forever, this is a security risk")
	}
}

// LogStore determines the local storage path for all log files created by this web application.
func (c *Config) LogStore() error {
	logs := c.AbsLog
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
	c.AbsLog = logs
	return nil
}

// SetupLogDir runs checks against the configured log directory.
// If no log directory is configured, a default directory is used.
// Problems will either log warnings or fatal errors.
func (c *Config) SetupLogDir(logger *zap.SugaredLogger) error {
	if logger == nil {
		return ErrZap
	}
	if c.AbsLog == "" {
		if err := c.LogStore(); err != nil {
			return fmt.Errorf("%w: %w", ErrLog, err)
		}
	}
	dir, err := os.Stat(c.AbsLog)
	if os.IsNotExist(err) {
		return fmt.Errorf("log directory %w: %s", ErrDirNotExist, c.AbsLog)
	}
	if !dir.IsDir() {
		return fmt.Errorf("log directory %w: %s", ErrNotDir, dir.Name())
	}
	empty := filepath.Join(c.AbsLog, ".defacto2_touch_test")
	if _, err := os.Stat(empty); os.IsNotExist(err) {
		f, err := os.Create(empty)
		if err != nil {
			return fmt.Errorf("log directory %w: %w", ErrTouch, err)
		}
		defer func(f *os.File) {
			f.Close()
			if err := os.Remove(empty); err != nil {
				logger.Warnf("Could not remove the empty test file in the log directory path: %s: %s", err, empty)
				return
			}
		}(f)
		return nil
	}
	if err := os.Remove(empty); err != nil {
		logger.Warnf("Could not remove the empty test file in the log directory path: %s: %s", err, empty)
	}
	return nil
}

// CheckDir runs checks against the named directory,
// including whether it exists, is a directory, and contains a minimum number of files.
// Problems will either log warnings or fatal errors.
func CheckDir(name, desc string) error {
	if name == "" {
		return fmt.Errorf("%w: %s", ErrDir, desc)
	}
	dir, err := os.Stat(name)
	if os.IsNotExist(err) {
		return fmt.Errorf("%w, %s: %s", ErrDir404, desc, name)
	}
	if !dir.IsDir() {
		return fmt.Errorf("%w, %s: %s", ErrDirIs, desc, dir.Name())
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
	fmt.Fprintf(os.Stdout, "Temporary directory using, %s: %s\n", hdu, tmpdir)
}

// Validate returns an error if the HTTP or TLS port is invalid.
func Validate(port uint) error {
	const disabled = 0
	if port == disabled {
		return nil
	}
	if port > PortMax {
		return ErrPortMax
	}
	if port <= PortSys {
		return ErrPortSys
	}
	return nil
}
