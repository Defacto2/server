package helpers

// params.go helpers are funcs specific for the unique Defacto2 URL IDs.

import (
	"fmt"
	"math"
	"path"
	"strconv"
	"strings"
)

// Deobfuscate a public facing, obfuscated file ID or file URL.
// A URL can point to a Defacto2 file download or detail page.
func Deobfuscate(s string) int {
	p := strings.Split(s, "?")
	d := deobfuscate(path.Base(p[0]))
	id, _ := strconv.Atoi(d)
	return id
}

// deObfuscate de-obfuscates a CFWheels obfuscateParam or Obfuscate() obfuscated string.
func deobfuscate(s string) string {
	const twoChrs, decimal, hexadecimal = 2, 10, 16
	// CFML source:
	// https://github.com/cfwheels/cfwheels/blob/cf8e6da4b9a216b642862e7205345dd5fca34b54/wheels/global/misc.cfm
	if _, err := strconv.Atoi(s); err == nil || len(s) < twoChrs {
		return s
	}
	// deobfuscate string.
	tail := s[twoChrs:]
	n, err := strconv.ParseInt(tail, hexadecimal, 0)
	if err != nil {
		return s
	}

	n ^= 461 // bitxor
	ns := strconv.Itoa(int(n))
	l := len(ns) - 1
	tail = ""

	for i := 0; i < l; i++ {
		f := ns[l-i:][:1]
		tail += f
	}
	// Create checks.
	ct := 0
	l = len(tail)

	for i := 0; i < l; i++ {
		chr := tail[i : i+1]
		n, err1 := strconv.Atoi(chr)

		if err1 != nil {
			return s
		}

		ct += n
	}
	// Run checks.
	ci, err := strconv.ParseInt(s[:2], hexadecimal, 0)
	if err != nil {
		return s
	}

	c2 := strconv.FormatInt(ci, decimal)

	const unknown = 154

	if strconv.FormatInt(int64(ct+unknown), decimal) != c2 {
		return s
	}

	return tail
}

// ObfuscateParam hides the param value using the method implemented in CFWheels obfuscateParam() helper.
func ObfuscateParam(param string) string {
	if param == "" {
		return ""
	}
	// check to make sure param doesn't begin with a 0 digit
	if param[0] == '0' {
		return param
	}
	pint, err := strconv.Atoi(param)
	if err != nil {
		return param
	}
	l := len(param)
	r, err := ReverseInt(uint(pint))
	if err != nil {
		return param
	}
	afloat64 := math.Pow10(l) + float64(r)
	// keep a and b as int type
	a, b := int(afloat64), 0
	for i := 1; i <= l; i++ {
		// slice individual digits from param and sum them
		s, err := strconv.Atoi(string(param[l-i]))
		if err != nil {
			return param
		}
		b += s
	}
	// base 64 conversion
	const hex, xor, sum = 16, 461, 154
	a ^= xor
	b += sum
	return strconv.FormatInt(int64(b), hex) + strconv.FormatInt(int64(a), hex)
}

// ReverseInt swaps the direction of the value, 12345 would return 54321.
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

func PageCount(sum, limit int) uint {
	if sum <= 0 || limit <= 0 {
		return 0
	}
	x := math.Ceil(float64(sum) / float64(limit))
	return uint(x)
}
