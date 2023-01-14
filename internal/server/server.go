// Package server contains internal functions for the main server application.
package server

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

var (
	ErrEmpty    = errors.New("empty directory input")
	ErrNoReader = errors.New("reader cannot be nil, it should be os.stdin")
)

func ParsePsVersion(s string) string {
	if x := strings.Split(s, " "); len(x) > 2 {
		_, err := strconv.ParseFloat(x[1], 32)
		if err != nil {
			return fmt.Sprintln(s)
		}
		return fmt.Sprintf(" with %s\n", strings.Join(x[0:2], " "))
	}
	return fmt.Sprintln(s)
}
