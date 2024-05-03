package helper

import (
	"fmt"
	"os"
	"reflect"
)

// Add1 returns the value of a + 1.
// The type of a must be an integer type or the result is 0.
func Add1(a any) int64 {
	switch val := a.(type) {
	case int, int8, int16, int32, int64:
		i := reflect.ValueOf(val).Int()
		return i + 1
	default:
		return 0
	}
}

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
		i++
	}
	return i, nil
}
