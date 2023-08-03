package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

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
func (c *Config) Checks(log *zap.SugaredLogger) {
	if log == nil {
		fmt.Fprintf(os.Stderr, "cannot run config checks as the logger instance is nil.")
		return
	}
	if err := HTTPPort(c.HTTPPort); err != nil {
		switch {
		case errors.Is(err, ErrPortMax):
			log.Fatalf("The server could not use the HTTP port %d, %s.",
				c.HTTPPort, err)
		case errors.Is(err, ErrPortSys):
			log.Infof("The server HTTP port %d, %s.",
				c.HTTPPort, err)
		}
	}

	DownloadDir(c.DownloadDir, log)
	ScreenshotsDir(c.ScreenshotsDir, log)
	ThumbnailDir(c.ThumbnailDir, log)
	c.SetupLogDir(log)
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

// DownloadDir runs checks against the named directory containing the UUID record downloads.
// Problems will either log warnings or fatal errors.
func DownloadDir(name string, log *zap.SugaredLogger) {
	CheckDir(name, "download", log)
}

// ScreenshotsDir runs checks against the named directory containing the screenshot images.
// Problems will either log warnings or fatal errors.
func ScreenshotsDir(name string, log *zap.SugaredLogger) {
	CheckDir(name, "screenshot", log)
}

// ThumbnailDir runs checks against the named directory containing the thumbnail images.
// Problems will either log warnings or fatal errors.
func ThumbnailDir(name string, log *zap.SugaredLogger) {
	CheckDir(name, "thumbnail", log)
}

// CheckDir runs checks against the named directory,
// including whether it exists, is a directory, and contains a minimum number of files.
// Problems will either log warnings or fatal errors.
func CheckDir(name, desc string, log *zap.SugaredLogger) {
	if log == nil {
		fmt.Fprintf(os.Stderr, "The logger instance for the config dir check is nil.")
	}
	s := ""
	switch desc {
	case "download":
		s = "the server cannot send file downloads"
	case "log":
		s = "the server cannot log to files"
	case "screenshot":
		s = "the server cannot show screenshot images"
	case "thumbnail":
		s = "the server cannot show thumbnail images"
	}
	if name == "" {
		log.Warnf("The %s directory path was not provided, %s.", desc, s)
		return
	}
	dir, err := os.Stat(name)
	if os.IsNotExist(err) {
		log.Warnf("The %s directory path does not exist, %s: %s", desc, s, name)
		return
	}
	if !dir.IsDir() {
		log.Fatalf("The %s directory path points to the file, %s: %s", desc, s, dir.Name())
	}
	files, err := os.ReadDir(name)
	if err != nil {
		log.Fatalf("The %s directory path could not be read, %s: %s.", desc, s, err)
	}
	if len(files) < toFewFiles {
		log.Warnf("The %s directory path contains only a few items, is the directory correct:  %s",
			desc, dir.Name())
		return
	}
}

// SetupLogDir runs checks against the configured log directory.
// Problems will either log warnings or fatal errors.
func (c *Config) SetupLogDir(log *zap.SugaredLogger) {
	if log == nil {
		fmt.Fprintf(os.Stderr, "The logger instance for the config log dir is nil.")
	}
	if c.LogDir == "" {
		if err := c.LogStorage(); err != nil {
			log.Fatalf("The server cannot log to files: %s", err)
		}
	}
	dir, err := os.Stat(c.LogDir)
	if os.IsNotExist(err) {
		log.Fatalf("The log directory path does not exist, the server cannot log to files: %s", c.LogDir)
	}
	if !dir.IsDir() {
		log.Fatalf("The log directory path points to the file: %s", dir.Name())
	}
	empty := filepath.Join(c.LogDir, ".defacto2_touch_test")
	if _, err := os.Stat(empty); os.IsNotExist(err) {
		f, err := os.Create(empty)
		if err != nil {
			log.Fatalf("Could not create a file in the log directory path: %s.", err)
		}
		defer func(f *os.File) {
			f.Close()
			if err := os.Remove(empty); err != nil {
				log.Warnf("Could not remove the empty test file in the log directory path: %s: %s", err, empty)
				return
			}
		}(f)
		return
	}
	if err := os.Remove(empty); err != nil {
		log.Warnf("Could not remove the empty test file in the log directory path: %s: %s", err, empty)
		return
	}
}
