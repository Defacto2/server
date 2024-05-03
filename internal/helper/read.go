package helper

import (
	"bufio"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"io"
	"os"
)

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
		names = append(names, file.Name())
	}
	return names, nil
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

// Size returns the size of the named file.
// If the file does not exist, it returns -1.
func Size(name string) int64 {
	st, err := os.Stat(name)
	if err != nil {
		return -1
	}
	return st.Size()
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
