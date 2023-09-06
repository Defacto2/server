package helper

import (
	"unicode/utf8"

	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"
)

// DetermineEncoding returns the encoding of the plain text byte slice.
// If the byte slice contains Unicode multi-byte characters then nil is returned.
// Otherwise a charmap.ISO8859_1 or charmap.CodePage437 encoding is returned.
func DetermineEncoding(p []byte) encoding.Encoding {
	if utf8.RuneCount(p) < len(p) {
		// detected multi-byte characters
		return nil
	}
	const (
		controlStart   = 0x00
		controlEnd     = 0x1f
		undefinedStart = 0x7f
		undefinedEnd   = 0x9f
		newline        = '\n'
		carriageReturn = '\r'
		tab            = '\t'
		escape         = 0x1b
		lastChar       = 0xff
	)
	for i := range p {
		switch {
		case p[i] == byte(newline), p[i] == byte(carriageReturn), p[i] == byte(tab):
			continue
		case p[i] == escape:
			continue
		case p[i] >= undefinedStart && p[i] <= undefinedEnd:
			return charmap.CodePage437
		case p[i] >= controlStart && p[i] <= controlEnd:
			return charmap.CodePage437
			// case p[i] > lastChar:
			// 	return nil
		}
	}
	return charmap.ISO8859_1
}
