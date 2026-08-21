package config

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/Defacto2/helper"
)

const DirWriteWriteBlock = 0o770 // Directory permissions.

type Abslog Directory

func (a Abslog) Help() string {
	if a == "" {
		return "No logs will be saved"
	}
	return ""
}

func (a Abslog) Issue() string {
	return Directory(a).Issue()
}

func (a Abslog) LogValue() slog.Value {
	return Directory(a).LogValue()
}

func (a Abslog) String() string {
	return Directory(a).String()
}

// LogStore determines the local storage path for all log files created by this web application.
func (c *Config) LogStore() error {
	const format = "config: log store %s: %w"

	lp := c.AbsLog.String()
	if lp == "" {
		dir, err := os.UserConfigDir()
		if err != nil {
			return fmt.Errorf(format, "os user config dir", err)
		}
		lp = filepath.Join(dir, ConfigDir)
	}

	path, err := filepath.Abs(lp)
	if err != nil {
		return fmt.Errorf(format, lp, err)
	}

	if logsExists := helper.Stat(path); !logsExists {
		if err := os.MkdirAll(path, DirWriteWriteBlock); err != nil {
			return fmt.Errorf(format, path, err)
		}
	}

	c.AbsLog = Abslog(lp)
	return nil
}
