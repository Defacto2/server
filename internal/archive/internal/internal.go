package internal

import (
	"strconv"
	"strings"
)

// ArjItem returns true if the string is a row from the [arj program] list command.
//
// [arj program]: https://arj.sourceforge.net/
func ARJItem(s string) bool {
	const minLen = 6
	if len(s) < minLen {
		return false
	}
	if s[3:4] != ")" {
		return false
	}
	x := s[:3]
	if _, err := strconv.Atoi(x); err != nil {
		return false
	}
	return true
}

// MagicLHA returns true if the LHA file type is matched in the magic string.
func MagicLHA(magic string) bool {
	s := strings.Split(magic, " ")
	const lha, lharc = "lha", "lharc"
	if s[0] == lharc {
		return true
	}
	if s[0] != lha {
		return false
	}
	if len(s) < len(lha) {
		return false
	}
	if strings.Join(s[0:3], " ") == "lha archive data" {
		return true
	}
	if strings.Join(s[2:4], " ") == "archive data" {
		return true
	}
	return false
}
