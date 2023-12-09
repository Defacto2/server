// Package command provides interfaces for shell commands and applications on the host system.
package command

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

type Dir struct {
	Download string
}

const (
	pattern = "defacto2-" // prefix for temporary directories
)

var (
	ErrEmpty = errors.New("file is empty")
	ErrIsDir = errors.New("file is a directory")
)

//Complementary assets

func RemoveImgs(preview, thumb, uuid string) error {
	exts := []string{".jpg", ".png", ".gif", ".webp"}
	// remove previews
	for _, ext := range exts {
		name := filepath.Join(preview, uuid+ext)
		st, err := os.Stat(name)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if st.IsDir() {
			return ErrIsDir
		}
		if err = os.Remove(name); err != nil {
			return err
		}
	}
	// remove thumbnails
	for _, ext := range exts {
		name := filepath.Join(thumb, uuid+ext)
		st, err := os.Stat(name)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if st.IsDir() {
			return ErrIsDir
		}
		if err = os.Remove(name); err != nil {
			return err
		}
	}
	return nil
}

// RemoveMe removes the file with the uuid name combined with a ".txt" extension
// from the download directory. It returns nil if the file does not exist.
func RemoveMe(download, uuid string) error {
	name := filepath.Join(download, uuid+".txt")
	st, err := os.Stat(name)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	if st.IsDir() {
		return ErrIsDir
	}
	return os.Remove(name)
}

// UnzipOne extracts a single file from a zip archive.
// The extracted file is copied to the src with the ext extension appended.
// It requires the [unzip] command to be available on the host system.
// This allows for better compatibility with retro zip archives,
// such as those that use the [compression methods] prior to zip deflate.
//
// The src argument is the path to the zip archive.
// The ext argument is the destination extension and should include a leading dot, eg. ".txt".
// The name argument is the name of the one file to unzip and copy.
//
// [unzip]: https://sourceforge.net/projects/infozip
// [compression methods]: https://www.hanshq.net/zip.html
func UnzipOne(src, ext, name string) error {

	st, err := os.Stat(src)
	if err != nil {
		return err
	}
	if st.IsDir() {
		return ErrIsDir
	}
	if st.Size() == 0 {
		return ErrEmpty
	}

	_, err = exec.LookPath("unzip")
	if errors.Is(err, exec.ErrDot) {
		err = nil
	}
	if err != nil {
		return err
	}

	tmp, err := os.MkdirTemp(os.TempDir(), pattern)
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)

	const exdir = "-d" // Directory to which to extract files.
	out, err := exec.Command("unzip", src, name, exdir, tmp).Output()
	if errors.Is(err, exec.ErrDot) {
		err = nil
	}
	if err != nil {
		return err
	}
	fmt.Println("out", string(out)) // TODO: print to terminal?

	extracted := filepath.Join(tmp, name)
	st, err = os.Stat(extracted)
	if err != nil {
		return err
	}
	if st.IsDir() {
		return ErrIsDir
	}
	if st.Size() == 0 {
		return ErrEmpty
	}

	srcFile, err := os.Open(extracted)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dst := fmt.Sprintf("%s%s", src, ext)
	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return err
	}

	err = dstFile.Sync()
	if err != nil {
		return err
	}

	return nil
}
