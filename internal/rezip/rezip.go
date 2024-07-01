package rezip

import (
	"archive/zip"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/Defacto2/server/internal/magicnumber/pkzip"
)

const (
	unzip     = "unzip"
	unzipTest = "-t"

	createUnique  = os.O_RDWR | os.O_CREATE | os.O_EXCL
	filemodeWrite = 0o644
)

var ErrTest = errors.New("unzip test failed")

// Compress compresses the named file into the dest zip file using the
// Deflate method. The total number of bytes written to the zip file is returned.
//
// The dest must be a valid file path and should include the .zip extension.
// If the dest file already exists, an error is returned.
func Compress(name, dest string) (int, error) {
	zipfile, err := os.OpenFile(dest, createUnique, filemodeWrite)
	if err != nil {
		return 0, fmt.Errorf("unzip compress failed to open file: %w", err)
	}
	defer zipfile.Close()

	w := zip.NewWriter(zipfile)
	defer w.Close()

	zipWr, err := w.Create(filepath.Base(name))
	if err != nil {
		return 0, fmt.Errorf("unzip compress failed to create writer: %w", err)
	}
	b, err := os.ReadFile(name)
	if err != nil {
		return 0, fmt.Errorf("unzip compress failed to read file: %w", err)
	}
	n, err := zipWr.Write(b)
	if err != nil {
		return 0, fmt.Errorf("unzip compress failed to write bytes: %w", err)
	}
	return n, nil
}

// CompressDir compresses the named root directory into the dest zip file
// using both the Deflate method. The total number
// of bytes written to the zip file is returned.
//
// The dest must be a valid file path and should include the .zip extension.
// If the dest file already exists, an error is returned.
func CompressDir(root, dest string) (int64, error) {
	zipfile, err := os.OpenFile(dest, createUnique, filemodeWrite)
	if err != nil {
		return 0, fmt.Errorf("unzip compress dir failed to open file: %w", err)
	}
	defer zipfile.Close()

	w := zip.NewWriter(zipfile)
	defer w.Close()

	var written int64
	addFile := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("add file: %w", err)
		}
		if info.IsDir() {
			return nil
		}
		if self := path == root; self {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return fmt.Errorf("add file: %w", err)
		}
		zipWr, err := w.Create(rel)
		if err != nil {
			return fmt.Errorf("add file: %w", err)
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("add file: %w", err)
		}
		n, err := zipWr.Write(b)
		if err != nil {
			return fmt.Errorf("add file: %w", err)
		}
		written += int64(n)
		return nil
	}

	err = filepath.Walk(root, addFile)
	if err != nil {
		return 0, fmt.Errorf("unzip compress dir failed to add file: %w", err)
	}

	return written, nil
}

// Test runs the unzip test command on the named file. If the file is a directory
// or empty, an error is returned. If the test command fails, an error is returned.
func Test(name string) error {
	path, err := exec.LookPath(unzip)
	if err != nil {
		return fmt.Errorf("unzip test failed to find unzip executable: %w", err)
	}
	st, err := os.Stat(name)
	if err != nil {
		return fmt.Errorf("unzip test failed to stat file: %w", err)
	}
	if st.IsDir() {
		return fmt.Errorf("%w: %s is a directory", ErrTest, name)
	}
	if st.Size() == 0 {
		return fmt.Errorf("%w: %s is empty", ErrTest, name)
	}
	err = exec.Command(path, unzipTest, name).Run()
	if err != nil {
		diag := pkzip.ExitStatus(err)
		switch diag {
		case pkzip.Normal, pkzip.Warning:
			// normal or warnings are fine
			return nil
		}
		return fmt.Errorf("%w: %s", ErrTest, diag)
	}
	return nil
}

// Recompress placeholder, TODO.
func Recompress(name, dest string) error {
	// run tests on paths

	// magicnumber test the named file

	// switch then load a func for each match

	fmt.Fprintln(os.Stdout, "recompress", name, dest)

	return nil
}
