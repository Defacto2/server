package config

import (
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
)

type Directory string // Directory contains an absolute path to a directory.

func (d Directory) Check() error {
	const format = "config directory check: %w"

	st, err := os.Stat(string(d))
	if err != nil {
		return fmt.Errorf(format, err)
	}
	if !st.IsDir() {
		return fmt.Errorf(format, ErrNotDir)
	}

	return nil
}

func (d Directory) Issue() string {
	if d == "" {
		return ""
	}

	err := d.Check()
	if err == nil {
		return ""
	}
	if errors.Is(err, os.ErrNotExist) {
		return "Directory does not exist"
	}
	if errors.Is(err, ErrNotDir) {
		return "Directory path points to a file and cannot be used"
	}
	if errors.Is(err, fs.ErrPermission) {
		return "Directory cannot be accessed due to permission denied"
	}

	return ""
}

func (d Directory) LogValue() slog.Value {
	return slog.StringValue(string(d))
}

func (d Directory) String() string {
	return string(d)
}
