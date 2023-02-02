package helpers

// params.go helpers are funcs specific for the unique Defacto2 URL IDs.

import (
	"fmt"
	"math"
	"net/url"
	"path"
	"strconv"
)

const (
	hexadecimal  = 16
	obfuscateXOR = 461
	obfuscateSum = 154
)

// Deobfuscate an obfuscated ID to return the primary key of the record.
// Returns a 0 if the id is not valid.
func Deobfuscate(id string) int {
	key, _ := strconv.Atoi(deobfuscate(id))
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
	return Deobfuscate(path.Base(u.Path))
}

// Deobfuscate the obfuscated string, or return the original string.
// This is a port of a CFWheels framework function programmed in Coldfusion (CFML).
// See: https://github.com/cfwheels/cfwheels/blob/cf8e6da4b9a216b642862e7205345dd5fca34b54/wheels/global/misc.cfm#L508
func deobfuscate(s string) string {
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
	num ^= 461
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
	chksumY := strconv.FormatInt(int64(chksumTest+obfuscateXOR), decimal)
	if err := chksumX != chksumY; err {
		return s
	}
	return value
}

// Obfuscates the primary key of a record as a string that is used as a URL param or path.
func Obfuscate(key int64) string {
	return obfuscate(uint(key))
}

// obfuscate returns the value of i as an obfuscatated string.
// This is a port of a CFWheels framework function programmed in Coldfusion (CFML).
// See: https://github.com/cfwheels/cfwheels/blob/cf8e6da4b9a216b642862e7205345dd5fca34b54/wheels/global/misc.cfm#L483
func obfuscate(i uint) string {
	s := strconv.Itoa(int(i))
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

// ReverseInt swaps the direction of the i value, 12345 would return 54321.
func ReverseInt(i uint) (uint, error) {
	var (
		n int
		s string
	)
	v := strconv.Itoa(int(i))
	for x := len(v); x > 0; x-- {
		s += string(v[x-1])
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return i, fmt.Errorf("reverse int %q: %w", s, err)
	}
	return uint(n), nil
}

// PageCount returns the maximum pages possible for the sum of records with a record limit per-page.
func PageCount(sum, limit int) uint {
	if sum <= 0 || limit <= 0 {
		return 0
	}
	x := math.Ceil(float64(sum) / float64(limit))
	return uint(x)
}
