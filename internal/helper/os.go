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
	"os/user"
	"path/filepath"
	"sort"
	"strings"
	"unicode/utf8"
)

const (
	// WriteWriteRead is the file mode for read and write access.
	// The file owner and group has read and write access, and others have read access.
	WriteWriteRead fs.FileMode = 0o664
	DSStore                    = ".DS_Store" // DSStore is the macOS directory service store file.
)

// Extension is a file extension with a count of files.
type Extension struct {
	Name  string // Name is the file extension.
	Count int64  // Count is the number of files with the extension.
}

func CountExts(dir string) ([]Extension, error) {
	exts := make(map[string]int64)
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("count extensions read directory: %w", err)
	}
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		if file.Name() == DSStore {
			continue
		}
		ext := strings.ToLower(filepath.Ext(file.Name()))
		exts[ext]++
	}
	extensions := make([]Extension, 0, len(exts))
	for k, v := range exts {
		if k == "" {
			k = "uuid"
		}
		extensions = append(extensions, Extension{Name: k, Count: v})
	}
	sort.Slice(extensions, func(i, j int) bool {
		return extensions[i].Count > extensions[j].Count
	})
	return extensions, nil
}

// Count returns the number of files in the given directory.
func Count(dir string) (int, error) {
	i := 0
	st, err := os.Stat(dir)
	if err != nil {
		return 0, fmt.Errorf("count os.stat %w", err)
	}
	if !st.IsDir() {
		return 0, fmt.Errorf("%w: %s", ErrDirPath, dir)
	}
	files, err := os.ReadDir(dir)
	if err != nil {
		return 0, fmt.Errorf("count os.readdir %w", err)
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
// It returns the number of bytes written to the new file.
// The function returns an error if the newpath already exists.
func Duplicate(oldpath, newpath string) (int64, error) {
	const createNoTruncate = os.O_CREATE | os.O_WRONLY | os.O_EXCL
	return duplicate(oldpath, newpath, createNoTruncate)
}

// DuplicateOW is a workaround for renaming files across different devices.
// A cross device can also be a different file system such as a Docker volume.
// It returns the number of bytes written to the new file.
// The function will truncate and overwrite the newpath if it already exists.
func DuplicateOW(oldpath, newpath string) (int64, error) {
	const createTruncate = os.O_CREATE | os.O_WRONLY | os.O_TRUNC
	return duplicate(oldpath, newpath, createTruncate)
}

func duplicate(oldpath, newpath string, flag int) (int64, error) {
	src, err := os.Open(oldpath)
	if err != nil {
		return 0, fmt.Errorf("duplicate os.open %w", err)
	}
	defer src.Close()

	dst, err := os.OpenFile(newpath, flag, WriteWriteRead)
	if err != nil {
		return 0, fmt.Errorf("duplicate os.create %w", err)
	}
	defer dst.Close()

	written, err := io.Copy(dst, src)
	if err != nil {
		return 0, fmt.Errorf("duplicate io.copy %w", err)
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
		return nil, fmt.Errorf("files os.stat %w", err)
	}
	if !st.IsDir() {
		return nil, fmt.Errorf("%w: %s", ErrDirPath, dir)
	}
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("files os.readdir: %w", err)
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
		return false, fmt.Errorf("file match os.open %s: %w", name1, err)
	}
	defer f1.Close()

	f2, err := os.Open(name2)
	if err != nil {
		return false, fmt.Errorf("file match os.open %s: %w", name2, err)
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
			return false, fmt.Errorf("file match %w: %s, %s", ErrRead, name1, name2)
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
		return "", fmt.Errorf("integrity fs.readfile %w", err)
	}
	return IntegrityBytes(b), nil
}

// IntegrityFile returns the sha384 hash of the named file.
// This can be used as a link cache buster.
func IntegrityFile(name string) (string, error) {
	b, err := os.ReadFile(name)
	if err != nil {
		return "", fmt.Errorf("integrity os.readfile %w", err)
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
		return 0, fmt.Errorf("integrity os.open %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lines := 0
	for scanner.Scan() {
		lines++
	}

	if err := scanner.Err(); err != nil {
		return 0, fmt.Errorf("integrity scanner.scan %w", err)
	}

	return lines, nil
}

// Owner returns the running user and group of the web application.
// The function returns the group names and the username of the owner.
func Owner() ([]string, string, error) {
	curr, err := user.Current()
	if err != nil {
		return nil, "", fmt.Errorf("owner user current %w", err)
	}
	grps, err := curr.GroupIds()
	if err != nil {
		return nil, "", fmt.Errorf("owner user group ids %w", err)
	}
	groups := make([]string, len(grps))
	for i, id := range grps {
		group, err := user.LookupId(id)
		if err != nil {
			continue
		}
		groups[i] = group.Name
	}
	return groups, curr.Username, nil
}

// RenameFile renames a file from oldpath to newpath.
// It returns an error if the oldpath does not exist or is a directory,
// newpath already exists, or the rename fails.
func RenameFile(oldpath, newpath string) error {
	st, err := os.Stat(oldpath)
	if err != nil {
		return fmt.Errorf("rename file os.stat %w", err)
	}
	if st.IsDir() {
		return fmt.Errorf("rename file oldpath %w: %s", ErrFilePath, oldpath)
	}
	if _, err = os.Stat(newpath); err == nil {
		return fmt.Errorf("rename file newpath %w: %s", ErrExistPath, newpath)
	}
	if err := os.Rename(oldpath, newpath); err != nil {
		var linkErr *os.LinkError
		if errors.As(err, &linkErr) && linkErr.Err.Error() == "invalid cross-device link" {
			return RenameCrossDevice(oldpath, newpath)
		}
		return fmt.Errorf("rename file os.rename %w", err)
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
		return fmt.Errorf("rename cross device open source, %w", err)
	}
	defer src.Close()
	dst, err := os.Create(newpath)
	if err != nil {
		return fmt.Errorf("rename cross device create new, %w", err)
	}
	defer dst.Close()

	if _, err = io.Copy(dst, src); err != nil {
		return fmt.Errorf("rename cross device copy %w", err)
	}
	if fi, err := os.Stat(oldpath); err != nil {
		defer os.Remove(newpath)
		return fmt.Errorf("rename cross device stat %w", err)
	} else if fi.Size() == 0 {
		defer os.Remove(newpath)
		defer os.Remove(oldpath)
		return fmt.Errorf("rename cross device empty file, %w", os.ErrNotExist)
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
		return "", fmt.Errorf("strong integrity open %w: %s", err, name)
	}
	defer f.Close()
	strong, err := Sum386(f)
	if err != nil {
		return "", fmt.Errorf("strong integrity %w", err)
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
		return "", fmt.Errorf("sha386 checksum %s: %w", f.Name(), err)
	}
	s := hex.EncodeToString(strong.Sum(nil))
	return s, nil
}

// Touch creates a new, empty named file.
// If the file already exists, an error is returned.
func Touch(name string) error {
	file, err := os.OpenFile(name, os.O_CREATE|os.O_EXCL, WriteWriteRead)
	if err != nil {
		return fmt.Errorf("touch open file %w", err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("touch file close %w", err)
	}
	return nil
}

// TouchW creates a new named file with the given data.
// If the file already exists, an error is returned.
func TouchW(name string, data ...byte) (int, error) {
	file, err := os.OpenFile(name, os.O_CREATE|os.O_WRONLY, WriteWriteRead)
	if err != nil {
		return 0, fmt.Errorf("touch write open file %w", err)
	}
	if len(data) == 0 {
		if err := file.Close(); err != nil {
			return 0, fmt.Errorf("touch write open file close %w", err)
		}
		return 0, nil
	}
	i, err := file.Write(data)
	if err != nil {
		return 0, fmt.Errorf("touch write file write %w", err)
	}
	if err := file.Close(); err != nil {
		return 0, fmt.Errorf("touch write file write close %w", err)
	}
	return i, nil
}

// UTF8 returns true if the named file is a valid UTF-8 encoded file.
// The function reads the first 512 bytes of the file to determine the encoding.
func UTF8(name string) (bool, error) {
	f, err := os.Open(name)
	if err != nil {
		return false, fmt.Errorf("utf8 open %w", err)
	}
	defer f.Close()
	const sample = 512
	buf := make([]byte, sample)
	_, err = f.Read(buf)
	if err != nil {
		return false, fmt.Errorf("utf8 read %w", err)
	}
	return utf8.Valid(buf), nil
}
