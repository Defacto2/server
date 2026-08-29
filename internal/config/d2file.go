package config

import (
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
)

type File string // File contains an absolute path to a file.

func (f File) Check() error {
	const format = "config file check: %w"

	st, err := os.Stat(string(f))
	if err != nil {
		return fmt.Errorf(format, err)
	}
	if st.IsDir() {
		return fmt.Errorf(format, ErrNotFile)
	}

	return nil
}

func (f File) Issue() string {
	if f == "" {
		return ""
	}

	err := f.Check()
	if err == nil {
		return ""
	}
	if errors.Is(err, os.ErrNotExist) {
		return "File does not exist"
	}
	if errors.Is(err, ErrNotFile) {
		return "File path points to a directory and cannot be used"
	}
	if errors.Is(err, fs.ErrPermission) {
		return "File cannot be accessed due to permission denied"
	}

	return ""
}

func (f File) LogValue() slog.Value {
	return slog.StringValue(string(f))
}

func (f File) String() string {
	return string(f)
}
