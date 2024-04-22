package helper

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"io"

	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"
)

// DetermineEncoding returns the encoding of the plain text byte slice.
// If the byte slice contains Unicode multi-byte characters then nil is returned.
// Otherwise a charmap.ISO8859_1 or charmap.CodePage437 encoding is returned.
func DetermineEncoding(reader io.Reader) encoding.Encoding {
	if reader == nil {
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
		multiByte      = 0x100
		unknownChr     = 65533
	)
	p, err := io.ReadAll(reader)
	if err != nil {
		return nil
	}
	s := string(p)
	for i, r := range s {
		switch {
		case // common whitespace control characters
			r == rune(newline),
			r == rune(carriageReturn),
			r == rune(tab):
			continue
		case r == rune(escape):
			// escape control character commonly used for ANSI
			continue
		case p[i] >= undefinedStart && p[i] <= undefinedEnd:
			// unused ASCII, which we can probably assumed to be CP-437
			return charmap.CodePage437
		case p[i] >= controlStart && p[i] <= controlEnd:
			// ASCII control characters, which we can probably assumed to be CP-437 glyphs
			return charmap.CodePage437
		case r == unknownChr:
			// when an unknown extended-ASCII character (128-255) is encountered, it is probably CP-437
			return charmap.CodePage437
		case r > unknownChr:
			// The maximum value of an 8-bit character is 255 (0xff),
			// so rune valud above that, 256+ (0x100) is a Unicode multi-byte character,
			// which we can probably assumed to be UTF-8.
			return nil
		}
	}
	return patternCheck(p)
}

// patternCheck returns the encoding based on the presence of common CP-437 or ISO-8859-1 patterns.
// Such patterns as full block, medium shade, horizontal bars and half blocks are sequences of
// characters that are unique to the CP-437 encoding.
func patternCheck(p []byte) encoding.Encoding {
	const (
		lowerHalfBlock = 0xdc
		upperHalfBlock = 0xdf
		doubleHorizBar = 0xcd
		singleHorizBar = 0xc4
		mediumShade    = 0xb1
		fullBlock      = 0xdb
	)
	patterns := []byte{
		lowerHalfBlock,
		upperHalfBlock,
		doubleHorizBar,
		singleHorizBar,
		mediumShade,
		fullBlock,
	}
	for _, pattern := range patterns {
		const count = 4
		subslice := bytes.Repeat([]byte{pattern}, count)
		if bytes.Contains(p, subslice) {
			return charmap.CodePage437
		}
	}
	return charmap.ISO8859_1
}

// CookieStore generates a key for use with the sessions cookie store middleware.
// envKey is the value of an imported environment session key. But if it is empty,
// a 32-bit randomized value is generated that changes on every restart.
//
// The effect of using a randomized key will invalidate all existing sessions on every restart.
func CookieStore(envKey string) ([]byte, error) {
	if envKey != "" {
		key := []byte(envKey)
		return key, nil
	}
	const length = 32
	key := make([]byte, length)
	n, err := rand.Read(key)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrKey, err.Error())
	}
	if n != length {
		return nil, ErrKey
	}
	return key, nil
}
