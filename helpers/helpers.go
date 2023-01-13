package helpers

import (
	"fmt"
	"path/filepath"
	"strings"
	"unicode"
	"unicode/utf8"
)

func ByteCount(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d%s", b, strings.Repeat(" ", 1))
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.0f%c",
		float64(b)/float64(div), "KMGTPE"[exp])
}

func LastChr(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	r, _ := utf8.DecodeLastRuneInString(s)
	return string(r)
}

func TrimPunct(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	rs := []rune(s)
	for i := len(rs) - 1; i >= 0; i-- {
		r := rs[i]
		// https://www.compart.com/en/unicode/category/Po
		if !unicode.Is(unicode.Po, r) {
			return string(rs[0 : i+1])
		}
	}
	return s
}

func TruncFilename(w int, name string) string {
	const trunc = "."
	if w == 0 {
		return ""
	}
	l := len(name)
	if w >= l {
		return name
	}
	ext := filepath.Ext(name)
	if w <= len(ext) {
		return ext
	}
	s := name[0 : w-len(ext)-len(trunc)]
	return fmt.Sprintf("%s%s%s", s, trunc, ext)
}

func IsValid(name string, valid ...string) bool {
	for _, n := range valid {
		if n == name {
			return true
		}
	}
	return false
}
