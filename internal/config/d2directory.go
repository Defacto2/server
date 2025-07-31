package config

import (
	"errors"
	"log/slog"
	"os"
)

type Directory string // Directory contains an absolute path to a directory.

func (d Directory) Check() error {
	st, err := os.Stat(string(d))
	if err != nil {
		return err
	}
	if !st.IsDir() {
		return ErrNotDir
	}
	return nil
}

func (d Directory) Issue() string {
	if d == "" {
		return ""
	}
	err := d.Check()
	if errors.Is(err, os.ErrNotExist) {
		return "Directory does not exist"
	}
	if errors.Is(err, ErrNotDir) {
		return "Directory path points to a file and cannot be used"
	}
	return ""
}

func (d Directory) LogValue() slog.Value {
	if d == "" {
		return slog.StringValue("")
	}
	return slog.StringValue(string(d))
}

func (d Directory) String() string {
	return string(d)
}
