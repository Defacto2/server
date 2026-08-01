// Package dir handles the directories in use by the internal server.
// Such as the download, extra, and preview directories.
package dir

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/Defacto2/server/internal/nils"
)

var (
	ErrFile   = errors.New("file error")
	ErrSave   = errors.New("save error")
	ErrNoPath = errors.New("the directory path is not set")
	ErrNoDir  = errors.New("the directory path does not exist")
)

// Directory is a string type that represents an internal server directory path.
type Directory string

// Join combines the directory path with the given file or directory name.
func (d Directory) Join(name string) string {
	return filepath.Join(d.Path(), name)
}

// Path returns the directory path as a string.
func (d Directory) Path() string {
	return string(d)
}

// Check confirms that the directory exists and is writable.
func (d Directory) Check(sl *slog.Logger) error {
	const msg = "directory check"
	const format = msg + ": %w"
	if err := nils.Check(sl); err != nil {
		return fmt.Errorf(format, err)
	}
	if err := d.IsDir(); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(d.Path(), "uploader-*.zip")
	if err != nil {
		return ErrSave
	}
	defer func() {
		_ = tmp.Close()
		if err := os.Remove(tmp.Name()); err != nil {
			sl.Error(msg, slog.String("name", tmp.Name()), slog.Any("error", err))
		}
	}()
	return nil
}

// IsDir returns an error if the path does not exists or is not a directory.
func (d Directory) IsDir() error {
	const format = "directory isdir: %w"
	path := d.Path()
	if path == "" {
		return ErrNoPath
	}
	st, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return ErrNoDir
		}
		return fmt.Errorf(format, err)
	}
	if !st.IsDir() {
		return ErrFile
	}
	return nil
}

// Paths converts the slice of Directory types to a slice of strings
// representing the directory paths.
func Paths(d ...Directory) []string {
	paths := make([]string, len(d))
	for i, dir := range d {
		paths[i] = dir.Path()
	}
	return paths
}
