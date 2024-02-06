package helper

import (
	"bufio"
	"crypto/sha512"
	"fmt"
	"io"
	"os"
)

// Files returns the filenames in the given directory.
func Files(dir string) ([]string, error) {
	st, err := os.Stat(dir)
	if err != nil {
		return nil, err
	}
	if !st.IsDir() {
		return nil, fmt.Errorf("%w: %s", ErrDirPath, dir)
	}
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var names []string
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
		return 0, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lines := 0
	for scanner.Scan() {
		lines++
	}

	if err := scanner.Err(); err != nil {
		return 0, err
	}

	return lines, nil
}

// StrongIntegrity returns the SHA-386 checksum value of the named file.
func StrongIntegrity(name string) (string, error) {
	// strong hashes require the named file to be reopened after being read.
	f, err := os.Open(name)
	if err != nil {
		return "", fmt.Errorf("%w: %s", err, name)
	}
	defer f.Close()
	strong, err := Sum386(f)
	if err != nil {
		return "", err
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
		return "", fmt.Errorf("%s: %w", f.Name(), err)
	}
	return fmt.Sprintf("%x", strong.Sum(nil)), nil
}
