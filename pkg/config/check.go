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
)

var (
	ErrPortMax = fmt.Errorf("http port value must be between 0-%d", PortMax)
	ErrPortSys = fmt.Errorf("http port values between 0-%d require system access", PortSys)
)

// Checks runs a number of sanity checks for the environment variable configurations.
func (c Config) Checks(log *zap.SugaredLogger) {
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
	LogDir(c.LogDir, log)

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
	if name == "" {
		log.Warn("The download directory path is empty, the server cannot send record downloads.")
		return
	}
	dir, err := os.Stat(name)
	if os.IsNotExist(err) {
		log.Warnf("The download directory path does not exist, the server cannot send record downloads: %s", name)
		return
	}
	if !dir.IsDir() {
		log.Fatalf("The download directory path points to the file: %s", dir.Name())
	}
	files, err := os.ReadDir(name)
	if err != nil {
		log.Fatalf("The download directory path could not be read: %s.", err)
	}
	if len(files) < 10 {
		log.Warnf("The download directory path contains only a few items, is the directory correct:  %s",
			dir.Name())
		return
	}
}

// LogDir runs checks against the named log directory.
// Problems will either log warnings or fatal errors.
func LogDir(name string, log *zap.SugaredLogger) {
	if name == "" {
		// recommended
		return
	}
	dir, err := os.Stat(name)
	if os.IsNotExist(err) {
		log.Fatalf("The log directory path does not exist, the server cannot log to files: %s", name)
	}
	if !dir.IsDir() {
		log.Fatalf("The log directory path points to the file: %s", dir.Name())
	}
	empty := filepath.Join(name, ".defacto2_touch_test")
	f, err := os.Create(empty)
	if err != nil {
		log.Fatalf("Could not create a file in the log directory path: %s.", err)
	}
	defer f.Close()
	if err := os.Remove(empty); err != nil {
		log.Warnf("Could not remove the empty test file in the log directory path: %s: %s", err, empty)
		return
	}
}
