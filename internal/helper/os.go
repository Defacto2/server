package helper

// Package file os.go contains the helper functions for file system operations.

import (
	"bufio"
	"crypto/sha512"
	"embed"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"unicode/utf8"
)

const (
	// WriteWriteRead is the file mode for read and write access.
	// The file owner and group has read and write access, and others have read access.
	WriteWriteRead fs.FileMode = 0o664
	// TODO: user and group chown of file?
	DSStore = ".DS_Store" // DSStore is the macOS directory service store file.
)

// Count returns the number of files in the given directory.
func Count(dir string) (int, error) {
	i := 0
	st, err := os.Stat(dir)
	if err != nil {
		return 0, fmt.Errorf("os.Stat: %w", err)
	}
	if !st.IsDir() {
		return 0, fmt.Errorf("%w: %s", ErrDirPath, dir)
	}
	files, err := os.ReadDir(dir)
	if err != nil {
		return 0, fmt.Errorf("os.ReadDir: %w", err)
	}
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		if file.Name() == DSStore {
			continue
		}
		i++
	}
	return i, nil
}

// Duplicate is a workaround for renaming files across different devices.
// A cross device can also be a different file system such as a Docker volume.
func Duplicate(oldpath, newpath string) (int64, error) {
	src, err := os.Open(oldpath)
	if err != nil {
		return 0, fmt.Errorf("os.Open: %w", err)
	}
	defer src.Close()
	dst, err := os.Create(newpath)
	if err != nil {
		return 0, fmt.Errorf("os.Create: %w", err)
	}
	defer dst.Close()

	written, err := io.Copy(dst, src)
	if err != nil {
		return 0, fmt.Errorf("io.Copy: %w", err)
	}
	if err = os.Chmod(newpath, WriteWriteRead); err != nil {
		defer os.Remove(newpath)
		return 0, fmt.Errorf("os.Chmod %d: %w", WriteWriteRead, err)
	}
	return written, nil
}

// File returns true if the named file exists on the system.
func File(name string) bool {
	s, err := os.Stat(name)
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}
		return false
	}
	if s.IsDir() {
		return false
	}
	return true
}

// Files returns the filenames in the given directory.
func Files(dir string) ([]string, error) {
	st, err := os.Stat(dir)
	if err != nil {
		return nil, fmt.Errorf("os.Stat: %w", err)
	}
	if !st.IsDir() {
		return nil, fmt.Errorf("%w: %s", ErrDirPath, dir)
	}
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("os.ReadDir: %w", err)
	}
	names := []string{}
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		if file.Name() == DSStore {
			continue
		}
		names = append(names, file.Name())
	}
	return names, nil
}

// FileMatch returns true if the two named files are the same.
// It returns false if the files are of different lengths or
// if an error occurs while reading the files.
// The read buffer size is 4096 bytes.
func FileMatch(name1, name2 string) (bool, error) {
	f1, err := os.Open(name1)
	if err != nil {
		return false, fmt.Errorf("%w: %w", ErrFileMatch, err)
	}
	defer f1.Close()

	f2, err := os.Open(name2)
	if err != nil {
		return false, fmt.Errorf("%w: %w", ErrFileMatch, err)
	}
	defer f2.Close()

	const maxSize = 4096
	buf1 := make([]byte, maxSize)
	buf2 := make([]byte, maxSize)

	for {
		n1, err1 := f1.Read(buf1)
		n2, err2 := f2.Read(buf2)
		if err1 != nil || err2 != nil {
			if err1 == io.EOF && err2 == io.EOF {
				break
			} else if err1 == io.EOF || err2 == io.EOF {
				return false, ErrDiffLength
			}
			return false, fmt.Errorf("%w: %s, %s", ErrRead, name1, name2)
		}

		if n1 != n2 || string(buf1[:n1]) != string(buf2[:n2]) {
			return false, nil
		}
	}
	return true, nil
}

// Finds returns true if the name is found in the collection of names.
func Finds(name string, names ...string) bool {
	for _, n := range names {
		if n == name {
			return true
		}
	}
	return false
}

// Integrity returns the sha384 hash of the named embed file.
// This is intended to be used for Subresource Integrity (SRI)
// verification with integrity attributes in HTML script and link tags.
func Integrity(name string, fs embed.FS) (string, error) {
	b, err := fs.ReadFile(name)
	if err != nil {
		return "", fmt.Errorf("fs.ReadFile: %w", err)
	}
	return IntegrityBytes(b), nil
}

// IntegrityFile returns the sha384 hash of the named file.
// This can be used as a link cache buster.
func IntegrityFile(name string) (string, error) {
	b, err := os.ReadFile(name)
	if err != nil {
		return "", fmt.Errorf("os.ReadFile: %w", err)
	}
	return IntegrityBytes(b), nil
}

// IntegrityBytes returns the sha384 hash of the given byte slice.
func IntegrityBytes(b []byte) string {
	sum := sha512.Sum384(b)
	b64 := base64.StdEncoding.EncodeToString(sum[:])
	return "sha384-" + b64
}

// Lines returns the number of lines in the named file.
func Lines(name string) (int, error) {
	file, err := os.Open(name)
	if err != nil {
		return 0, fmt.Errorf("os.Open: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lines := 0
	for scanner.Scan() {
		lines++
	}

	if err := scanner.Err(); err != nil {
		return 0, fmt.Errorf("scanner.Scan: %w", err)
	}

	return lines, nil
}

// RenameFile renames a file from oldpath to newpath.
// It returns an error if the oldpath does not exist or is a directory,
// newpath already exists, or the rename fails.
func RenameFile(oldpath, newpath string) error {
	st, err := os.Stat(oldpath)
	if err != nil {
		return fmt.Errorf("os.Stat: %w", err)
	}
	if st.IsDir() {
		return fmt.Errorf("oldpath %w: %s", ErrFilePath, oldpath)
	}
	if _, err = os.Stat(newpath); err == nil {
		return fmt.Errorf("newpath %w: %s", ErrExistPath, newpath)
	}
	if err := os.Rename(oldpath, newpath); err != nil {
		var linkErr *os.LinkError
		if errors.As(err, &linkErr) && linkErr.Err.Error() == "invalid cross-device link" {
			return RenameCrossDevice(oldpath, newpath)
		}
		return fmt.Errorf("os.Rename: %w", err)
	}
	return nil
}

// RenameFileOW renames a file from oldpath to newpath.
// It returns an error if the oldpath does not exist or is a directory
// or the rename fails.
func RenameFileOW(oldpath, newpath string) error {
	st, err := os.Stat(newpath)
	if err == nil && st.IsDir() {
		_ = os.Remove(newpath)
	}
	return RenameFile(oldpath, newpath)
}

// RenameCrossDevice is a workaround for renaming files across different devices.
// A cross device can also be a different file system such as a Docker volume.
func RenameCrossDevice(oldpath, newpath string) error {
	src, err := os.Open(oldpath)
	if err != nil {
		return fmt.Errorf("os.Open: %w", err)
	}
	defer src.Close()
	dst, err := os.Create(newpath)
	if err != nil {
		return fmt.Errorf("os.Create: %w", err)
	}
	defer dst.Close()

	if _, err = io.Copy(dst, src); err != nil {
		return fmt.Errorf("io.Copy: %w", err)
	}
	fi, err := os.Stat(oldpath)
	if err != nil {
		defer os.Remove(newpath)
		return fmt.Errorf("os.Stat: %w", err)
	}
	if err = os.Chmod(newpath, fi.Mode()); err != nil {
		defer os.Remove(newpath)
		return fmt.Errorf("os.Chmod: %w", err)
	}
	defer os.Remove(oldpath)
	return nil
}

// Size returns the size of the named file.
// If the file does not exist, it returns -1.
func Size(name string) int64 {
	st, err := os.Stat(name)
	if err != nil {
		return -1
	}
	return st.Size()
}

// Stat stats the named file or directory to confirm it exists on the system.
func Stat(name string) bool {
	if _, err := os.Stat(name); err != nil {
		return false
	}
	return true
}

// StrongIntegrity returns the SHA-386 checksum value of the named file.
func StrongIntegrity(name string) (string, error) {
	// strong hashes require the named file to be reopened after being read.
	f, err := os.Open(name)
	if err != nil {
		return "", fmt.Errorf("os.Open: %w: %s", err, name)
	}
	defer f.Close()
	strong, err := Sum386(f)
	if err != nil {
		return "", fmt.Errorf("Sum386: %w", err)
	}
	return strong, nil
}

// Sum386 returns the SHA-386 checksum value of the open file.
func Sum386(f *os.File) (string, error) {
	if f == nil {
		return "", ErrOSFile
	}
	strong := sha512.New384()
	if _, err := io.Copy(strong, f); err != nil {
		return "", fmt.Errorf("io.Copy %s: %w", f.Name(), err)
	}
	s := hex.EncodeToString(strong.Sum(nil))
	return s, nil
}

// Touch creates a new, empty named file.
// If the file already exists, an error is returned.
func Touch(name string) error {
	file, err := os.OpenFile(name, os.O_CREATE|os.O_EXCL, ReadWrite)
	if err != nil {
		return fmt.Errorf("os.OpenFile: %w", err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("file.Close: %w", err)
	}
	return nil
}

// TouchW creates a new named file with the given data.
// If the file already exists, an error is returned.
func TouchW(name string, data ...byte) (int, error) {
	file, err := os.OpenFile(name, os.O_CREATE|os.O_WRONLY, ReadWrite)
	if err != nil {
		return 0, fmt.Errorf("os.OpenFile: %w", err)
	}
	if len(data) == 0 {
		if err := file.Close(); err != nil {
			return 0, fmt.Errorf("file.Close: %w", err)
		}
		return 0, nil
	}
	i, err := file.Write(data)
	if err != nil {
		return 0, fmt.Errorf("file.Write: %w", err)
	}
	if err := file.Close(); err != nil {
		return 0, fmt.Errorf("file.Close: %w", err)
	}
	return i, nil
}

// UTF8 returns true if the named file is a valid UTF-8 encoded file.
// The function reads the first 512 bytes of the file to determine the encoding.
func UTF8(name string) (bool, error) {
	f, err := os.Open(name)
	if err != nil {
		return false, fmt.Errorf("os.Open: %w", err)
	}
	defer f.Close()

	buf := make([]byte, 512)
	_, err = f.Read(buf)
	if err != nil {
		return false, fmt.Errorf("f.Read: %w", err)
	}
	return utf8.Valid(buf), nil
}
