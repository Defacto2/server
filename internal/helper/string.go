package helper

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"golang.org/x/exp/slices"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// ByteCount formats b as in a compact, human-readable unit of measure.
func ByteCount(b int64) string {
	// source: https://yourbasic.org/golang/formatting-byte-size-to-human-readable-format/
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d%s", b, strings.Repeat("B", 1))
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.0f%c",
		float64(b)/float64(div), byteUnits[exp])
}

// ByteCountFloat formats b as in a human-readable unit of measure.
// Units measured in gigabytes or larger are returned with 1 decimal place.
func ByteCountFloat(b int64) string {
	const unit = 1000
	if b < unit {
		return fmt.Sprintf("%d bytes", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	const gigabyte = 2
	if exp < gigabyte {
		return fmt.Sprintf("%.0f %cB",
			float64(b)/float64(div), byteUnits[exp])
	}
	return fmt.Sprintf("%.1f %cB",
		float64(b)/float64(div), byteUnits[exp])
}

// Capitalize returns a string with the first letter of the first word capitalized.
// If the first word is an acronym, it is capitalized as a word.
func Capitalize(s string) string {
	if s == "" {
		return ""
	}
	const sep = " "
	caser := cases.Title(language.English)
	x := strings.Split(s, sep)
	if len(x) == 1 {
		return caser.String(s)
	}
	return caser.String(x[0]) + sep + strings.Join(x[1:], sep)
}

// DeleteDupe removes duplicate strings from a slice.
// The returned slice is sorted and compacted.
func DeleteDupe(s ...string) []string {
	slices.Sort(s)
	s = slices.Compact(s)
	x := make([]string, 0, len(s))
	for _, val := range s {
		if slices.Contains(x, val) {
			continue
		}
		x = append(x, val)
	}
	return slices.Compact(x)
}

// FmtSlice formats a comma separated string.
func FmtSlice(s string) string {
	x := []string{}
	y := strings.Split(s, ",")
	for _, z := range y {
		z = strings.TrimSpace(z)
		if z == "" {
			continue
		}
		x = append(x, Capitalize(z))
	}
	return strings.Join(x, ", ")
}

// LastChr returns the last character or rune of the string.
func LastChr(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	r, _ := utf8.DecodeLastRuneInString(s)
	return string(r)
}

// MaxLineLength counts the character length of the longest line in a string.
func MaxLineLength(s string) int {
	lines := strings.Split(s, "\n")
	max := 0
	for _, line := range lines {
		l := utf8.RuneCountInString(line)
		if l > max {
			max = l
		}
	}
	return max
}

// ReverseInt reverses an integer.
func ReverseInt(i int) (int, error) {
	// credit: Wade73
	// http://stackoverflow.com/questions/35972561/reverse-int-golang
	itoa, str := strconv.Itoa(i), ""
	for x := len(itoa); x > 0; x-- {
		str += string(itoa[x-1])
	}

	reverse, err := strconv.Atoi(str)
	if err != nil {
		return 0, fmt.Errorf("reverseInt %d: %w", i, err)
	}

	return reverse, nil
}

// ShortMonth takes a month integer and abbreviates it to a three letter English month.
func ShortMonth(month int) string {
	if month < 1 || month > 12 {
		return ""
	}
	const abbreviated = 3
	s := fmt.Sprint(time.Month(month))
	if len(s) >= abbreviated {
		return s[0:abbreviated]
	}
	return ""
}

// SplitAsSpaces splits a string at each capital letter.
func SplitAsSpaces(s string) string {
	var result strings.Builder
	for i, r := range s {
		if unicode.IsUpper(r) && i != 0 {
			result.WriteRune(' ')
		}
		result.WriteRune(r)
	}
	x := result.String()
	x = strings.ReplaceAll(x, "Dir", "Directory")
	x = strings.ReplaceAll(x, "H T T P", "HTTP") //nolint:dupword
	x = strings.ReplaceAll(x, "P S ", "PS ")
	x = strings.ReplaceAll(x, "I D", "ID")
	x = strings.ReplaceAll(x, "  ", " ")
	return x
}

// TruncFilename reduces a filename to the length of w characters.
// The file extension is always preserved with the truncation.
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

// TrimRoundBraket removes the tailing round brakets and any whitespace.
func TrimRoundBraket(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	l, r := strings.Index(s, "("), strings.Index(s, ")")
	if l < r {
		return strings.TrimSpace(s[:l-1])
	}
	return s
}

// TrimPunct removes any trailing, common punctuation characters from the string.
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
			punctless := string(rs[0 : i+1])
			return strings.TrimSpace(punctless)
		}
	}
	return s
}

// Years returns a string of the years if they are different.
// If they are the same, it returns a singular year.
func Years(a, b int16) string {
	if a == b {
		return fmt.Sprintf("the year %d", a)
	}
	if b-a == 1 {
		return fmt.Sprintf("the years %d and %d", a, b)
	}
	return fmt.Sprintf("the years %d - %d", a, b)
}
