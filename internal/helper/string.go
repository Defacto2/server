package helper

import (
	"fmt"
	"math"
	"net/url"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/google/uuid"
	"golang.org/x/exp/slices"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

const (
	hexadecimal  = 16
	obfuscateXOR = 461
	obfuscateSum = 154
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

// CFToUUID formats a 35 character, Coldfusion Universally Unique Identifier
// to a standard, 36 character, Universally Unique Identifier.
func CFToUUID(cfid string) (string, error) {
	if err := uuid.Validate(cfid); err == nil {
		return cfid, nil
	}
	const pos = 23
	const hyphen = '-'
	old := strings.TrimSpace(cfid)
	r := []rune(old)
	r = append(r[:pos], append([]rune{hyphen}, r[pos:]...)...)
	new := string(r)
	err := uuid.Validate(new)
	if err != nil {
		return "", fmt.Errorf("uuid.Validate: %w", err)
	}
	return new, nil
}

// CFToUUID formats a 31 character, Coldfusion Universally Unique Identifier
// to a standard, 32 character, Universally Unique Identifier.
// func CFToUUID(cfid string) string {
// 	const require = 35
// 	if len(cfid) != require {
// 		return cfid
// 	}
// 	const index = 23
// 	return cfid[:index] + "-" + cfid[index:]
// }

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

// Deobfuscate the obfuscated string, or return the original string.
func DeObfuscate(s string) string {
	// This function is a port of a CFWheels framework function programmed in ColdFusion (CFML).
	// See: https://github.com/cfwheels/cfwheels/blob/cf8e6da4b9a216b642862e7205345dd5fca34b54/wheels/global/misc.cfm#L508
	const checksum, decimal = 2, 10
	if len(s) < checksum {
		return s
	}
	if i, _ := strconv.Atoi(s); i > 0 {
		return s
	}
	// deobfuscate string
	num, err := strconv.ParseInt(s[checksum:], hexadecimal, 0)
	if err != nil {
		return s
	}
	num ^= obfuscateXOR
	baseNum := strconv.Itoa(int(num))
	l := len(baseNum) - 1
	value := ""
	for i := range l {
		f := baseNum[l-i:][:1]
		value += f
	}
	// create checks
	l = len(value)
	chksumTest := 0
	for i := range l {
		chr := value[i : i+1]
		n, err1 := strconv.Atoi(chr)
		if err1 != nil {
			return s
		}
		chksumTest += n
	}
	// run checks
	chksum, err := strconv.ParseInt(s[:2], hexadecimal, 0)
	if err != nil {
		return s
	}
	chksumX := strconv.FormatInt(chksum, decimal)
	chksumY := strconv.FormatInt(int64(chksumTest+obfuscateSum), decimal)
	if err := chksumX != chksumY; err {
		return s
	}

	return value
}

// DeobfuscateID an obfuscated ID to return the primary key of the record.
// Returns a 0 if the id is not valid.
func DeobfuscateID(id string) int {
	key, _ := strconv.Atoi(DeObfuscate(id))
	return key
}

// Deobfuscate an obfuscated record URL to return a record's primary key.
// A URL can point to a Defacto2 record download or detail page.
// Returns a 0 if the URL is not valid.
func DeobfuscateURL(rawURL string) int {
	u, err := url.Parse(rawURL)
	if err != nil {
		return 0
	}
	return DeobfuscateID(path.Base(u.Path))
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

// ObfuscateID the primary key of a record as a string that is used as a URL param or path.
func ObfuscateID(key int64) string {
	return Obfuscate(strconv.Itoa(int(key)))
}

// Obfuscate a numeric string to insecurely hide database primary key values when passed along a URL.
// This function is a port of a CFWheels framework function programmed in ColdFusion (CFML).
// https://github.com/cfwheels/cfwheels/blob/cf8e6da4b9a216b642862e7205345dd5fca34b54/wheels/global/misc.cfm#L483
func Obfuscate(s string) string {
	i, err := strconv.Atoi(s)
	if err != nil {
		return s
	}
	// confirm the first digit of i isn't a zero
	if s[0] == '0' {
		return s
	}
	reverse, err := ReverseInt(i)
	if err != nil {
		return s
	}
	l := len(s)
	a := int(math.Pow10(l) + float64(reverse))
	b := 0
	for i := 1; i <= l; i++ {
		// slice and sum the individual digits
		digit, err := strconv.Atoi(string(s[l-i]))
		if err != nil {
			return s
		}
		b += digit
	}
	// base64 conversion
	a ^= obfuscateXOR
	b += obfuscateSum

	return fmt.Sprintf("%s%s",
		strconv.FormatInt(int64(b), hexadecimal),
		strconv.FormatInt(int64(a), hexadecimal),
	)
}

// PageCount returns the maximum pages possible for the sum of records with a record limit per-page.
func PageCount(sum, limit int) uint {
	if sum <= 0 || limit <= 0 {
		return 0
	}
	x := math.Ceil(float64(sum) / float64(limit))
	return uint(x)
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

// SearchTerm returns a list of search terms from the input string.
// The input string is split by commas.
func SearchTerm(input string) []string {
	if input == "" {
		return []string{}
	}
	// split the input by double quotes
	q := strings.Split(input, ",")
	// join the two slices
	s := make([]string, 0, len(q))
	for _, v := range q {
		s = append(s, strings.TrimSpace(v))
	}
	return s
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

// Slug returns a URL friendly string of the named group.
func Slug(name string) string {
	s := name
	// remove diacritics
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	s, _, _ = transform.String(t, s)
	// hyphen to underscore
	re := regexp.MustCompile(`\-`)
	s = re.ReplaceAllString(s, "_")
	// multiple groups get separated with asterisk
	re = regexp.MustCompile(`\, `)
	s = re.ReplaceAllString(s, "*")
	// any & characters need replacement due to HTML escaping
	re = regexp.MustCompile(` \& `)
	s = re.ReplaceAllString(s, " ampersand ")
	// numbers receive a leading hyphen
	re = regexp.MustCompile(` ([0-9])`)
	s = re.ReplaceAllString(s, "-$1")
	// delete all other characters
	const deleteAllExcept = `[^A-Za-z0-9 \-\+\.\_\*]`
	re = regexp.MustCompile(deleteAllExcept)
	s = re.ReplaceAllString(s, "")
	// trim whitespace and replace any space separators with hyphens
	s = strings.TrimSpace(strings.ToLower(s))
	re = regexp.MustCompile(` `)
	s = re.ReplaceAllString(s, "-")
	return s
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
	x = strings.ReplaceAll(x, "T L S", "TLS")
	x = strings.ReplaceAll(x, "P S ", "PS ")
	x = strings.ReplaceAll(x, "I D", "ID")
	x = strings.ReplaceAll(x, "  ", " ")
	return x
}

// Titleize returns a string with the first letter each word capitalized.
// If a word is an acronym, it is capitalized as a word.
func Titleize(s string) string {
	if s == "" {
		return ""
	}
	const sep = " "
	caser := cases.Title(language.English)
	x := strings.Split(s, sep)
	if len(x) == 1 {
		return caser.String(s)
	}
	for i, word := range x {
		if word == "" {
			continue
		}
		x[i] = caser.String(word)
	}
	return strings.Join(x, sep)
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
