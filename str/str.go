package str

import (
	"fmt"
	"strconv"
	"strings"
)

// FWInt takes an int and returns it as a string, fw (fixed width) characters wide, with whitespace padding.
func FWInt(i, fw int) string {
	const pad = " "
	s := strconv.Itoa(i)
	l := len(s)
	if l >= fw {
		return s
	}
	return fmt.Sprintf("%s%s", strings.Repeat(pad, fw-l), s)
}
