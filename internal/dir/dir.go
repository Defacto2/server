package dir

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
)

var (
	ErrFile = fmt.Errorf("file error")
	ErrSave = fmt.Errorf("save error")
)

// Directory is a string type that represents an internal server directory path.
type Directory string

// Join combines the directory path with the given file or directory name.
func (d Directory) Join(name string) string {
	return filepath.Clean(filepath.Join(d.Path(), name))
}

func (d Directory) Path() string {
	return string(d)
}

// Check confirms that the directory exists and is writable.
func (d Directory) Check() error {
	name := d.Path()
	st, err := os.Stat(name)
	if err != nil {
		return err
	}
	if !st.IsDir() {
		return ErrFile
	}
	f, err := os.CreateTemp(name, "uploader-*.zip")
	if err != nil {
		return ErrSave
	}
	defer f.Close()
	defer os.Remove(f.Name())
	return nil
}

// Paths converts the slice of Directory types to a slice of strings
// representing the directory paths.
func Paths(d ...Directory) []string {
	var paths []string
	for val := range slices.Values(d) {
		paths = append(paths, val.Path())
	}
	return paths
}
