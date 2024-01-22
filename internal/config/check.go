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
	ErrPortMax = fmt.Errorf("http port value must be between 0-%d", PortMax)
	ErrPortSys = fmt.Errorf("http port values between 0-%d require system access", PortSys)
)

// Checks runs a number of sanity checks for the environment variable configurations.
func (c *Config) Checks(z *zap.SugaredLogger) {
	if z == nil {
		fmt.Fprintf(os.Stderr, "Cannot run config checks as the logger instance is nil.")
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

	if c.NoRobots {
		z.Warn("NoRobots is on, most web crawlers will ignore this site.")
	}
	if c.HTTPSRedirect {
		z.Warn("HTTPSRedirect is on, all HTTP requests will be redirected to HTTPS.")
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

// HTTPPort returns an error if the HTTP port is invalid.
func HTTPPort(port uint) error {
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
	s := ""
	switch desc {
	case "download":
		s = "file downloading will not work"
	case "log":
		s = "the server cannot log to files"
	case "preview":
		s = "previews images will not show"
	case "thumbnail":
		s = "thumbnail images will be blank"
	}
	if name == "" {
		return fmt.Errorf("no %s directory path, %s", desc, s)
	}
	dir, err := os.Stat(name)
	if os.IsNotExist(err) {
		return fmt.Errorf("the %s directory path does not exist, %s: %s", desc, s, name)
	}
	if !dir.IsDir() {
		return fmt.Errorf("the %s directory path points to the file, %s: %s", desc, s, dir.Name())
	}
	files, err := os.ReadDir(name)
	if err != nil {
		return fmt.Errorf("the %s directory path could not be read, %s: %s", desc, s, err)
	}
	if len(files) < toFewFiles {
		return fmt.Errorf("the %s directory path contains only a few items, is the directory correct:  %s",
			desc, dir.Name())
	}
	return nil
}
