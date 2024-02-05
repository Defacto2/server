package archive

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"unicode/utf8"

	"github.com/Defacto2/server/internal/archive/internal/command"
	"github.com/Defacto2/server/internal/archive/internal/mholter"
	"github.com/mholt/archiver"
	"github.com/nwaples/rardecode"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

var (
	ErrArchive = errors.New("format specified by source filename is not an archive format")
	ErrDir     = errors.New("is a directory")
	ErrFile    = errors.New("no such file")
	ErrWriter  = errors.New("writer must be a file object")
)

// Content returns both a list of files within an rar, tar, or zip archive;
// as-well as a suitable filename string for the archive. This filename is
// useful when the original archive filename has been given an invalid file
// extension.
//
// An absolute path is required by src that points to the archive file named as a unique id.
//
// The original archive filename with extension is required to determine text compression format.
func Content(src, filename string) ([]string, string, error) {
	st, err := os.Stat(src)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, "", fmt.Errorf("read %s: %w", filepath.Base(src), ErrFile)
	}
	if st.IsDir() {
		return nil, "", fmt.Errorf("read %s: %w", filepath.Base(src), ErrDir)
	}
	files, fname, err := Readr(src, filename)
	if err != nil {
		return nil, "", fmt.Errorf("read uuid/filename: %w", err)
	}
	return files, fname, nil
}

// Readr returns both a list of files within an rar, tar or zip archive,
// and a suitable archive filename string.
// If there are problems reading the archive due to an incorrect filename
// extension, the returned filename string will be corrected.
func Readr(src, filename string) ([]string, string, error) {
	if files, err := readr(src, filename); err == nil {
		return files, filename, nil
	}
	files, ext, err := command.Readr(src, filename)
	if errors.Is(err, command.ErrWrongExt) {
		newname := command.Rename(ext, filename)
		files, err = readr(src, newname)
		if err != nil {
			return nil, "", fmt.Errorf("readr fix: %w", err)
		}
		return files, newname, nil
	}
	if err != nil {
		return nil, "", fmt.Errorf("readr: %w", err)
	}
	// remove empty entries
	files = slices.DeleteFunc(files, func(s string) bool {
		return strings.TrimSpace(s) == ""
	})
	return files, filename, nil
}

func readr(src, filename string) ([]string, error) {
	files := []string{}
	return files, mholter.Walkr(src, filename, func(f archiver.File) error {
		if f.IsDir() {
			return nil
		}
		var fn string
		switch h := f.Header.(type) {
		case zip.FileHeader:
			fn = h.Name
		case *tar.Header:
			fn = h.Name
		case *rardecode.FileHeader:
			fn = h.Name
		default:
			fn = f.Name()
		}
		b := []byte(fn)
		if utf8.Valid(b) {
			files = append(files, fn)
			return nil
		}
		// handle cheeky DOS era filenames with CP437 extended characters.
		r := transform.NewReader(bytes.NewReader(b), charmap.CodePage437.NewDecoder())
		result, err := io.ReadAll(r)
		if err != nil {
			return err
		}
		files = append(files, string(result))
		return nil
	})
}
