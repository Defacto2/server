package config

// Package file check.go contains the sanity check functions for the configuration values.

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Defacto2/server/internal/helper"
	"go.uber.org/zap"
)

const (
	PortMax = 65534 // PortMax is the highest valid port number.
	PortSys = 1024  // PortSys is the lowest valid port number that does not require system access.

	toFewFiles = 10 // toFewFiles is the minimum number of files required in a directory.
)

var (
	ErrPortMax = fmt.Errorf("http port value must be between 1-%d", PortMax)
	ErrPortSys = fmt.Errorf("http port values between 1-%d require system access", PortSys)

	ErrDir     = fmt.Errorf("the directory path is not set")
	ErrDir404  = fmt.Errorf("the directory path does not exist")
	ErrDirIs   = fmt.Errorf("the directory path points to the file")
	ErrDirRead = fmt.Errorf("the directory path could not be read")
	ErrDirFew  = fmt.Errorf("the directory path contains only a few items")

	ErrUnencrypted = fmt.Errorf("the production server is configured to use unencrypted HTTP connections")
	ErrNoOAuth2    = fmt.Errorf("the production server requires a google, oauth2 client id to allow admin logins")
	ErrNoAccounts  = fmt.Errorf("the production server has no google oauth2 user accounts to allow admin logins")
	ErrSessionKey  = fmt.Errorf("the production server has a session, " +
		"encryption key set instead of using a randomized key")
)

// httpPort returns an error if the HTTP port is invalid.
func (c Config) httpPort(z *zap.SugaredLogger) {
	if c.HTTPPort == 0 {
		return
	}
	if err := HTTPPort(c.HTTPPort); err != nil {
		switch {
		case errors.Is(err, ErrPortMax):
			z.Fatalf("The server could not use the HTTP port %d, %s.",
				c.HTTPPort, err)
		case errors.Is(err, ErrPortSys):
			z.Infof("The server HTTP port %d, %s.",
				c.HTTPPort, err)
		}
	}
}

// httpsPort returns an error if the HTTPS port is invalid.
func (c Config) httpsPort(z *zap.SugaredLogger) {
	if c.HTTPSPort == 0 {
		return
	}
	if err := HTTPPort(c.HTTPSPort); err != nil {
		switch {
		case errors.Is(err, ErrPortMax):
			z.Fatalf("The server could not use the HTTPS port %d, %s.",
				c.HTTPSPort, err)
		case errors.Is(err, ErrPortSys):
			z.Infof("The server HTTPS port %d, %s.",
				c.HTTPSPort, err)
		}
	}
}

// The production mode checks when not in read-only mode. It
// expects the server to be configured with OAuth2 and Google IDs.
// The server should be running over HTTPS and not unencrypted HTTP.
func (c Config) production(z *zap.SugaredLogger) {
	if !c.ProductionMode || c.ReadMode {
		return
	}
	if c.GoogleClientID == "" {
		s := helper.Capitalize(ErrNoOAuth2.Error()) + "."
		z.Warn(s)
	}
	if c.GoogleIDs == "" && len(c.GoogleAccounts) == 0 {
		s := helper.Capitalize(ErrNoAccounts.Error()) + "."
		z.Warn(s)
	}
	if c.HTTPPort > 0 {
		s := fmt.Sprintf("%s over port %d.",
			helper.Capitalize(ErrUnencrypted.Error()),
			c.HTTPPort)
		z.Info(s)
	}
	if c.SessionKey != "" {
		s := helper.Capitalize(ErrSessionKey.Error()) + "."
		z.Warn(s)
		z.Warn("This means that all signed in clients will not be logged out on a server restart.")
	}
	if c.SessionMaxAge > 0 {
		z.Infof("A signed in client session lasts for %d hour(s).", c.SessionMaxAge)
	} else {
		z.Warn("A signed in client session lasts forever.")
	}
}

// Checks runs a number of sanity checks for the environment variable configurations.
func (c *Config) Checks(z *zap.SugaredLogger) {
	if z == nil {
		fmt.Fprintf(os.Stderr, "Cannot run config checks as the logger instance is nil.")
		return
	}

	if c.HTTPSRedirect && c.HTTPSPort == 0 {
		z.Warn("HTTPSRedirect is on but the HTTPS port is not set, so the server will not redirect HTTP requests to HTTPS.")
	}

	c.httpPort(z)
	c.httpsPort(z)
	c.production(z)

	// Check the download, preview and thumbnail directories.
	if err := DownloadDir(c.DownloadDir); err != nil {
		s := helper.Capitalize(err.Error()) + "."
		z.Warn(s)
	}
	if err := PreviewDir(c.PreviewDir); err != nil {
		s := helper.Capitalize(err.Error()) + "."
		z.Warn(s)
	}
	if err := ThumbnailDir(c.ThumbnailDir); err != nil {
		s := helper.Capitalize(err.Error()) + "."
		z.Warn(s)
	}

	// Reminds for the optional configuration values.
	if c.NoCrawl {
		z.Warn("NoCrawl is on, web crawlers should ignore this site.")
	}
	if c.HTTPSRedirect && c.HTTPSPort > 0 {
		z.Info("HTTPSRedirect is on, all HTTP requests will be redirected to HTTPS.")
	}

	c.SetupLogDir(z)
}

// SetupLogDir runs checks against the configured log directory.
// If no log directory is configured, a default directory is used.
// Problems will either log warnings or fatal errors.
func (c *Config) SetupLogDir(z *zap.SugaredLogger) {
	if z == nil {
		fmt.Fprintf(os.Stderr, "The logger instance for the config log dir is nil.")
	}
	if c.LogDir == "" {
		if err := c.LogStorage(); err != nil {
			z.Fatalf("The server cannot log to files: %s", err)
		}
	} else {
		z.Info("The server logs are found in: ", c.LogDir)
	}
	dir, err := os.Stat(c.LogDir)
	if os.IsNotExist(err) {
		z.Fatalf("The log directory path does not exist, the server cannot log to files: %s", c.LogDir)
	}
	if !dir.IsDir() {
		z.Fatalf("The log directory path points to the file: %s", dir.Name())
	}
	empty := filepath.Join(c.LogDir, ".defacto2_touch_test")
	if _, err := os.Stat(empty); os.IsNotExist(err) {
		f, err := os.Create(empty)
		if err != nil {
			z.Fatalf("Could not create a file in the log directory path: %s.", err)
		}
		defer func(f *os.File) {
			f.Close()
			if err := os.Remove(empty); err != nil {
				z.Warnf("Could not remove the empty test file in the log directory path: %s: %s", err, empty)
				return
			}
		}(f)
		return
	}
	if err := os.Remove(empty); err != nil {
		z.Warnf("Could not remove the empty test file in the log directory path: %s: %s", err, empty)
		return
	}
}

// HTTPPort returns an error if the HTTP/HTTPS port is invalid.
func HTTPPort(port uint) error {
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
