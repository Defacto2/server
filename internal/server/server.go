// Package server contains internal functions for the main server application.
package server

import (
	"fmt"
	"strconv"
	"strings"
)

// ParsePsVersion returns the database server name and version
// from the PosgreSQL result of the "SELECT version();" SQL statement.
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
