package helper

// Package file bool.go contains functions that return a boolean.

import (
	"os"
	"time"
)

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

// IsDir returns true if the named directory exists on the system.
func IsDir(name string) bool {
	if s, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
		if err != nil {
			return false
		}
		if s.IsDir() {
			return true
		}
	}
	return false
}

// IsFile returns true if the named file exists on the system.
func IsFile(name string) bool {
	if s, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
		if err != nil {
			return false
		}
		if s.IsDir() {
			return false
		}
	}
	return true
}

// IsStat stats the named file or directory to confirm it exists on the system.
func IsStat(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
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
