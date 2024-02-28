package helper

// Package file bool.go contains functions that return a boolean.

import (
	"fmt"
	"io"
	"os"
	"time"
)

// FileMatch returns true if the two named files are the same.
// It returns false if the files are of different lengths or
// if an error occurs while reading the files.
// The read buffer size is 4096 bytes.
func FileMatch(file1, file2 string) (bool, error) {
	f1, err := os.Open(file1)
	if err != nil {
		return false, err
	}
	defer f1.Close()

	f2, err := os.Open(file2)
	if err != nil {
		return false, err
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
			return false, fmt.Errorf("%w: %s, %s", ErrRead, file1, file2)
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

// IsDay returns true if the i value can be used as a day time value.
func IsDay(i int) bool {
	const maxDay = 31
	if i > 0 && i <= maxDay {
		return true
	}
	return false
}

// IsFile returns true if the named file exists on the system.
func IsFile(name string) bool {
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

// IsStat stats the named file or directory to confirm it exists on the system.
func IsStat(name string) bool {
	if _, err := os.Stat(name); err != nil {
		return false
	}
	return true
}

// IsYear returns true  if the i value is greater than 1969
// or equal to the current year.
func IsYear(i int) bool {
	const unix = 1970
	now := time.Now().Year()
	if i >= unix && i <= now {
		return true
	}
	return false
}
