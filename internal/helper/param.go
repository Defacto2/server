package helper

// Package file param.go contains functions specific for the unique, Defacto2 ID URLs.

import (
	"fmt"
	"math"
	"net/url"
	"path"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

const (
	hexadecimal  = 16
	obfuscateXOR = 461
	obfuscateSum = 154
)

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
	for i := 0; i < l; i++ {
		f := baseNum[l-i:][:1]
		value += f
	}
	// create checks
	l = len(value)
	chksumTest := 0
	for i := 0; i < l; i++ {
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

// Deobfuscate an obfuscated ID to return the primary key of the record.
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

// Slug returns a URL friendly string of the named group.
func Slug(name string) string {
	s := name
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

// SearchTerm returns a list of search terms from the input string.
// The input string is split by double quotes.
func SearchTerm(input string) []string {
	if input == "" {
		return []string{}
	}
	// split the input by double quotes
	q := strings.Split(input, "\"")
	fmt.Println("P1 ->", q)
	q = slices.DeleteFunc(q, func(s string) bool {
		if s == "" {
			return true
		}
		fmt.Println("S", s)
		if []rune(s)[0] == '"' && []rune(s)[len(s)-1] == '"' {
			return false
		}
		return true
	})
	fmt.Println("P2 ->", q)
	// join the two slices
	terms := []string{}
	terms = append(terms, q...)

	t := strings.Split(input, "\"")
	t = slices.DeleteFunc(t, func(s string) bool {
		if s == "" {
			return true
		}
		if []rune(s)[0] == '"' && []rune(s)[len(s)-1] == '"' {
			return true
		}
		return false
	})
	fmt.Println("T ->", t)
	for _, v := range t {
		// split the input by spaces
		x := strings.Split(v, " ")
		for _, y := range x {
			// delete any empty strings
			if len(y) == 0 {
				continue
			}
			terms = append(terms, y)
		}
	}

	// terms = append(terms, t...)

	// for i, v := range q {
	// 	q[i] = strings.TrimSpace(v)
	// }
	// slices.Sort(q)
	// q = slices.Compact(q)
	// // delete any empty strings
	// q = slices.DeleteFunc(q, func(s string) bool {
	// 	return len(s) == 0
	// })
	// terms := []string{}
	// for _, v := range q {
	// 	// split the input by spaces
	// 	x := strings.Split(v, " ")
	// 	for _, y := range x {
	// 		// delete any empty strings
	// 		if len(y) == 0 {
	// 			continue
	// 		}
	// 		terms = append(terms, y)
	// 	}
	// }
	return terms
}