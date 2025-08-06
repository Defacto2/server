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
