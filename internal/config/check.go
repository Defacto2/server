package config

// Package file check.go contains the sanity check functions for the configuration values.

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Defacto2/server/internal/helper"
	"github.com/Defacto2/server/internal/postgres"
	"go.uber.org/zap"
)

const (
	PortMax = 65534 // PortMax is the highest valid port number.
	PortSys = 1024  // PortSys is the lowest valid port number that does not require system access.

	toFewFiles = 10 // toFewFiles is the minimum number of files required in a directory.
)

var (
	ErrPortMax     = fmt.Errorf("http port value must be between 1-%d", PortMax)
	ErrPortSys     = fmt.Errorf("http port values between 1-%d require system access", PortSys)
	ErrDir         = errors.New("the directory path is not set")
	ErrDir404      = errors.New("the directory path does not exist")
	ErrDirIs       = errors.New("the directory path points to the file")
	ErrDirRead     = errors.New("the directory path could not be read")
	ErrDirFew      = errors.New("the directory path contains only a few items")
	ErrUnencrypted = errors.New("the production server is configured to use unencrypted HTTP connections")
	ErrNoOAuth2    = errors.New("the production server requires a google, oauth2 client id to allow admin logins")
	ErrNoAccounts  = errors.New("the production server has no google oauth2 user accounts to allow admin logins")
	ErrSessionKey  = errors.New("the production server has a session, " +
		"encryption key set instead of using a randomized key")
	ErrZap = errors.New("the zap logger instance is nil")
)

// Checks runs a number of sanity checks for the environment variable configurations.
func (c *Config) Checks(logger *zap.SugaredLogger) error {
	if logger == nil {
		return ErrZap
	}

	if c.HTTPSRedirect && c.TLSPort == 0 {
		logger.Warn("HTTPSRedirect is on but the HTTPS port is not set," +
			" so the server will not redirect HTTP requests to HTTPS.")
	}

	c.httpPort(logger)
	c.tlsPort(logger)
	c.production(logger)

	// Check the download, preview and thumbnail directories.
	if err := DownloadDir(c.DownloadDir); err != nil {
		s := helper.Capitalize(err.Error()) + "."
		logger.Warn(s)
	}
	if err := PreviewDir(c.PreviewDir); err != nil {
		s := helper.Capitalize(err.Error()) + "."
		logger.Warn(s)
	}
	if err := ThumbnailDir(c.ThumbnailDir); err != nil {
		s := helper.Capitalize(err.Error()) + "."
		logger.Warn(s)
	}

	// Reminds for the optional configuration values.
	if c.NoCrawl {
		logger.Warn("NoCrawl is on, web crawlers should ignore this site.")
	}
	if c.HTTPSRedirect && c.TLSPort > 0 {
		logger.Info("HTTPSRedirect is on, all HTTP requests will be redirected to HTTPS.")
	}
	if c.HostName == postgres.DockerHost {
		logger.Info("The application is configured for use in a Docker container.")
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
	if !c.ProductionMode || c.ReadMode {
		return
	}
	if c.GoogleClientID == "" {
		s := helper.Capitalize(ErrNoOAuth2.Error()) + "."
		logger.Warn(s)
	}
	if c.GoogleIDs == "" && len(c.GoogleAccounts) == 0 {
		s := helper.Capitalize(ErrNoAccounts.Error()) + "."
		logger.Warn(s)
	}
	if c.HTTPPort > 0 {
		s := fmt.Sprintf("%s over port %d.",
			helper.Capitalize(ErrUnencrypted.Error()),
			c.HTTPPort)
		logger.Info(s)
	}
	if c.SessionKey != "" {
		s := helper.Capitalize(ErrSessionKey.Error()) + "."
		logger.Warn(s)
		logger.Warn("This means that all signed in clients will not be logged out on a server restart.")
	}
	if c.SessionMaxAge > 0 {
		logger.Infof("A signed in client session lasts for %d hour(s).", c.SessionMaxAge)
	} else {
		logger.Warn("A signed in client session lasts forever.")
	}
}

// LogStore determines the local storage path for all log files created by this web application.
func (c *Config) LogStore() error {
	const ownerGroupAll = 0o770
	logs := c.LogDir
	if logs == "" {
		dir, err := os.UserConfigDir()
		if err != nil {
			return fmt.Errorf("os.UserConfigDir: %w", err)
		}
		logs = filepath.Join(dir, ConfigDir)
	}
	if ok := helper.Stat(logs); !ok {
		if err := os.MkdirAll(logs, ownerGroupAll); err != nil {
			return fmt.Errorf("%w: %s", err, logs)
		}
	}
	c.LogDir = logs
	return nil
}

// SetupLogDir runs checks against the configured log directory.
// If no log directory is configured, a default directory is used.
// Problems will either log warnings or fatal errors.
func (c *Config) SetupLogDir(logger *zap.SugaredLogger) error {
	if logger == nil {
		return ErrZap
	}
	if c.LogDir == "" {
		if err := c.LogStore(); err != nil {
			return fmt.Errorf("%w: %w", ErrLog, err)
		}
	} else {
		logger.Info("The server logs are found in: ", c.LogDir)
	}
	dir, err := os.Stat(c.LogDir)
	if os.IsNotExist(err) {
		return fmt.Errorf("log directory %w: %s", ErrDirNotExist, c.LogDir)
	}
	if !dir.IsDir() {
		return fmt.Errorf("log directory %w: %s", ErrNotDir, dir.Name())
	}
	empty := filepath.Join(c.LogDir, ".defacto2_touch_test")
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
	files, err := os.ReadDir(name)
	if err != nil {
		return fmt.Errorf("%w, %s: %w", ErrDirRead, desc, err)
	}
	if len(files) < toFewFiles {
		return fmt.Errorf("%w, %s: %s", ErrDirFew, desc, dir.Name())
	}
	return nil
}

// DownloadDir runs checks against the named directory containing the UUID artifact downloads.
// Problems will either log warnings or fatal errors.
func DownloadDir(name string) error {
	return CheckDir(name, "download")
}

// PreviewDir runs checks against the named directory containing the preview and screenshot images.
// Problems will either log warnings or fatal errors.
func PreviewDir(name string) error {
	return CheckDir(name, "preview")
}

// ThumbnailDir runs checks against the named directory containing the thumbnail images.
// Problems will either log warnings or fatal errors.
func ThumbnailDir(name string) error {
	return CheckDir(name, "thumbnail")
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
