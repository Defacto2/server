package config

import (
	"errors"
	"log/slog"
	"os"
)

type File string // File contains an absolute path to a file.

func (f File) Check() error {
	st, err := os.Stat(string(f))
	if err != nil {
		return err
	}
	if st.IsDir() {
		return ErrNotFile
	}
	return nil
}

func (f File) Issue() string {
	if f == "" {
		return ""
	}
	err := f.Check()
	if errors.Is(err, os.ErrNotExist) {
		return "File does not exist"
	}
	if errors.Is(err, ErrNotDir) {
		return "File path points to a file and cannot be used"
	}
	return ""
}

func (f File) LogValue() slog.Value {
	if f == "" {
		return slog.StringValue("")
	}
	return slog.StringValue(string(f))
}

func (f File) String() string {
	return string(f)
}
