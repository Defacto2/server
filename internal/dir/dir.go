// Package dir handles the directories in use by the internal server.
// Such as the download, extra, and preview directories.
package dir

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Defacto2/server/internal/nils"
)

const Pattern = "df2app"

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

// Check to confirm the directory is writable.
func (d Directory) Check(sl *slog.Logger) error {
	const format = "check directory %s: %w"
	if err := nils.Check(sl); err != nil {
		return fmt.Errorf(format, "arguments", err)
	}

	if err := d.IsDir(); err != nil {
		return fmt.Errorf(format, "isdir", err)
	}

	tmp, err := os.CreateTemp(d.Path(), "uploader-*.zip")
	if err != nil {
		return fmt.Errorf(format, "is not writable", err)
	}

	defer func() {
		_ = tmp.Close()
		if rErr := os.Remove(tmp.Name()); rErr != nil && sl != nil {
			sl.Error("directory check", slog.String("name", tmp.Name()), slog.Any("error", rErr))
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

// MkdirTemp creates a temporary directory at [os.TempDir] using the provided pattern.
// The optional pattern should not include a randomizer string "-*".
//
// Example usage on Linux:
//
//	path, _ := MkdirTemp("abc")
//	fmt.Println(path)
//
//	Output:
//	/tmp/df2app-abc-d83gsddb34
func MkdirTemp(pattern string) (string, error) {
	if pattern != "" {
		pattern = Pattern + "-" + pattern + "-*"
	} else {
		pattern = Pattern + "-*"
	}
	s, err := os.MkdirTemp(os.TempDir(), pattern)
	if err != nil {
		return "", fmt.Errorf("dir mkdir temp: %w", err)
	}
	return s, nil
}

// MkdirStale creates a temporary subdirectory in [os.TempDir] using the
// provided path as the basis of the directory name.
// Unlike [MkdirTemp], this func does not apply randomization to the
// directory name and the result will not always be unique.
// Allowing the directory to be used as a short-term file cache.
//
// The returned string is the path to an existing
// or the newly created temporary directory.
func MkdirStale(path string) (string, error) {
	const random = "-*"
	var pattern string
	base := filepath.Base(path)
	local, err := filepath.Localize(base)
	if err != nil {
		pattern = Pattern + random // for edge cases use a randomization
	} else {
		pattern = Pattern + "-" + local // no tailing randomization
	}

	newpath := filepath.Join(os.TempDir(), pattern)

	st, err := os.Stat(newpath)
	switch {
	case err == nil && st.IsDir():
		// return the exist temporary directory based on this path
		return newpath, nil
	case err == nil:
		// when an unexpected entry exists in the temporary directory,
		// append randomness to the pattern to make and use a new subdirectory.
		pattern += random
	}

	if strings.HasSuffix(pattern, random) {
		s, err := os.MkdirTemp(os.TempDir(), pattern)
		if err != nil {
			return "", fmt.Errorf("dir mkdir stale: %w", err)
		}
		return s, nil
	}

	err = os.MkdirAll(newpath, 0o700)
	if err != nil {
		return "", fmt.Errorf("dir mkdir stale: %w", err)
	}
	return newpath, nil
}

// CreateTemp creates a new temporary file using [os.CreateTemp].
// However, no path is required and the new file will be located
// in a subdirectory created using [MkdirTemp].
func CreateTemp(pattern string) (*os.File, error) {
	const format = "dir create temp %s: %w"
	if pattern == "" {
		pattern = Pattern + "-*"
	}

	path, err := MkdirTemp("createtemp")
	if err != nil {
		return nil, fmt.Errorf(format, "mkdir", err)
	}

	r, err := os.CreateTemp(path, pattern)
	if err != nil {
		return nil, fmt.Errorf(format, "create", err)
	}

	return r, nil
}

// IsTemp returns true when the dir entry is likely a temporary directory or
// file created using [MkdirTemp].
func IsTemp(d os.DirEntry) bool {
	if d == nil {
		return false
	}

	name := d.Name()
	if !d.IsDir() || name == "" {
		return false
	}

	return strings.HasPrefix(name, Pattern+"-")
}

// CleanTemp removes the temporary directories likely created using [MkdirTemp],
// that are older than the stale time value. An error is returned if a directory
// cannot be removed.
func CleanTemp(stale time.Duration) error {
	const format = "dir clean temp (%s) %s: %w"

	tempDir := os.TempDir()
	entries, err := os.ReadDir(tempDir)
	if err != nil {
		return fmt.Errorf(format, tempDir, "read dir", err)
	}

	for _, d := range entries {
		if !d.IsDir() || !IsTemp(d) {
			continue
		}

		info, err := d.Info()
		if err != nil {
			continue
		}

		if time.Since(info.ModTime()) < stale {
			continue
		}

		path := filepath.Join(tempDir, d.Name())
		if err := os.RemoveAll(path); err != nil {
			return fmt.Errorf(format, path, "remove all", err)
		}
	}

	return nil
}

// RemoveAll removes the named directory.
// For safety, the directory removal is locked using [os.OpenRoot]
// to [os.TempDir].
func RemoveAll(name string) error {
	const format = "dir remove all %s: %w"
	basePath := os.TempDir()

	root, err := os.OpenRoot(basePath)
	if err != nil {
		return fmt.Errorf(format, "open root", err)
	}
	defer root.Close()

	relPath, err := filepath.Rel(basePath, name)
	if err != nil {
		return fmt.Errorf(format, "resolve relative path", err)
	}

	if err := root.RemoveAll(relPath); err != nil {
		return fmt.Errorf(format, "root-scoped removal", err)
	}

	return nil
}
