package archive

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/Defacto2/magicnumber"
	"github.com/Defacto2/server/internal/helper"
)

var (
	errIsDir   = errors.New("error, directory")
	errTooMany = errors.New("will not decompress this archive as it is very large")
)

// ExtractSource extracts the source file into a temporary directory.
// The named file is used as part of the extracted directory path.
// The src is the source file to extract.
func ExtractSource(src, name string) (string, error) {
	const mb150 = 150 * 1024 * 1024
	if st, err := os.Stat(src); err != nil {
		return "", fmt.Errorf("cannot stat file: %w", err)
	} else if st.IsDir() {
		return "", errIsDir
	} else if st.Size() > mb150 {
		return "", errTooMany
	}
	dst, err := helper.MkContent(src)
	if err != nil {
		return "", fmt.Errorf("cannot create content directory: %w", err)
	}
	entries, _ := os.ReadDir(dst)
	const extracted = 2
	if len(entries) >= extracted {
		return dst, nil
	}
	switch filearchive(src) {
	case false:
		newpath := filepath.Join(dst, name)
		if _, err := helper.DuplicateOW(src, newpath); err != nil {
			defer os.RemoveAll(dst)
			return "", fmt.Errorf("cannot duplicate file: %w", err)
		}
	case true:
		if err := ExtractAll(src, dst); err != nil {
			defer os.RemoveAll(dst)
			return "", fmt.Errorf("cannot read extracted archive: %w", err)
		}
	}
	return dst, nil
}

func filearchive(src string) bool {
	r, err := os.Open(src)
	if err != nil {
		return false
	}
	sign, err := magicnumber.Archive(r)
	if err != nil {
		return false
	}
	return sign != magicnumber.Unknown
}

// List returns the files within an rar, tar, lha, or zip archive.
// This filename extension is used to determine the archive format.
func List(src, filename string) ([]string, error) {
	st, err := os.Stat(src)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, fmt.Errorf("archive list %w: %s", ErrMissing, filepath.Base(src))
	}
	if st.IsDir() {
		return nil, fmt.Errorf("archive list %w: %s", ErrFile, filepath.Base(src))
	}
	path, err := ExtractSource(src, filename)
	if err != nil {
		return commander(src, filename)
	}
	var files []string
	err = filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			rel, err := filepath.Rel(path, filePath)
			if err != nil {
				fmt.Fprint(io.Discard, err)
				files = append(files, filePath)
				return nil
			}
			files = append(files, rel)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("archive list %w", err)
	}
	return files, nil
}

// commander uses system archiver and decompression programs to read the src archive file.
func commander(src, filename string) ([]string, error) {
	c := Content{}
	if err := c.Read(src); err != nil {
		return nil, fmt.Errorf("commander failed with %s (%q): %w", filename, c.Ext, err)
	}
	// remove empty entries
	files := c.Files
	files = slices.DeleteFunc(files, func(s string) bool {
		return strings.TrimSpace(s) == ""
	})
	return files, nil
}
