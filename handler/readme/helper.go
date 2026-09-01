//nolint:gochecknoglobals
package readme

import (
	"bytes"
	"fmt"
	"io"
	"regexp"
)

var (
	byteCR     = []byte("\r")
	byteCRLF   = []byte("\r\n")
	byteEOF    = []byte("\x1a")
	byteLF     = []byte("\n")
	byteNull   = []byte{0x00}
	byteSpace  = []byte(" ")
	emptyBytes = []byte{}
	trimCutset = " \x1a" // Space + SUB character
)

const (
	eof          = "\x1a"
	trimBytesSet = "\t\n\v\f\r " + eof
)

// AddPrefix injects content before the existing byte content.
// This is usually to inject FILE_ID.DIZ text files.
func AddPrefix(p, prefix []byte) []byte {
	if len(bytes.TrimSpace(prefix)) == 0 {
		return p
	}
	if len(bytes.TrimSpace(p)) == 0 {
		return prefix
	}

	if len(prefix) == 0 {
		return bytes.Clone(prefix)
	}

	const nl = 2
	size := len(prefix) + nl + len(p)
	buf := make([]byte, size)
	n := copy(buf, prefix)
	buf[n] = '\n'
	buf[n+1] = '\n'
	copy(buf[n+nl:], p)

	return buf
}

// AddSuffix injects content after the existing byte content.
// This is usually to inject an additional helper text file.
func AddSuffix(p, suffix []byte) []byte {
	if len(bytes.TrimSpace(suffix)) == 0 {
		return p
	}
	if len(bytes.TrimSpace(p)) == 0 {
		return suffix
	}

	const nl = 2
	size := len(suffix) + nl + len(p)
	buf := make([]byte, size)

	n := copy(buf, p)
	buf[n] = '\n'
	buf[n+1] = '\n'
	copy(buf[n+nl:], suffix)

	return buf
}

// trimEOF to handle some edge cases whereby an è followed
// by a number of DOS-era end-of-file markers tail the text.
//
// Maybe these introduced by some specific bbs software in the day.
//
// Some examples from 1985:
//   - https://defacto2.net/f/b22621c
//   - https://defacto2.net/f/b328b2c
func trimEOF(s []byte) []byte {
	const (
		e   = 0x8a // CP437: è
		eof = 0x1a // MSDOS: end-of-file mark
	)
	n := len(s)
	if n == 0 {
		return s
	}

	i := n - 1
	for i >= 0 && s[i] == eof {
		i--
	}

	if i >= 0 && s[i] == e {
		i--
	}

	return s[:i+1]
}

// trimBytes removes ending standard space characters and MS-DOS EOF marker (0x1a).
func trimBytes(b []byte) []byte {
	return bytes.TrimRight(b, trimBytesSet)
}

var reControlCodes = regexp.MustCompile(`\x1b\[(?:\?[0-9]+\w|[0-9;]*[a-zA-Z]|[0-9;]* p)|SAUCE00`)

// removeControls removes known problematic characters and controls, including:
//   - ASCII control codes, NULL, SUB
//   - Amiga control operators
//   - DEC control codes
//   - ANSI escape codes
//   - Windows and PC DOS newlines
func removeControls(b []byte) []byte {
	if len(b) == 0 {
		return b
	}

	b = reControlCodes.ReplaceAll(b, nil)

	s := make([]byte, 0, len(b))
	for i := 0; i < len(b); i++ {
		switch {
		case b[i] == '\x00':
			s = append(s, ' ')
		case b[i] == '\r' && i+1 < len(b) && b[i+1] == '\n':
			s = append(s, '\n')
			i++ // skip the newline '\n' in CRLF pairs
		default:
			s = append(s, b[i])
		}
	}

	return bytes.TrimRight(s, trimCutset)
}

var ansiSequence = []byte("\x1b[")

// hasANSI scans the reader and returns true if ANSI escape codes are found.
// If the reader is too large to render in a HTML template, an error is returned.
func hasANSI(r io.Reader) (bool, error) {
	if r == nil {
		return false, nil
	}

	const (
		format    = "match ansi reader %s: %w"
		maxBytes  = 1024 * 1024 // 1MB
		chunkSize = 8 * 1024    // 8KB stack buffer with 0 allocs
	)

	var buf [chunkSize]byte
	var read, carry int

	for {
		n, err := r.Read(buf[carry:])
		if n > 0 {
			total := carry + n
			read += n

			if read > maxBytes {
				return false, fmt.Errorf(format, "read bytes", ErrTooLong)
			}

			if bytes.Contains(buf[:total], ansiSequence) {
				return true, nil
			}

			if buf[total-1] == '\x1b' {
				buf[0] = '\x1b'
				carry = 1
			} else {
				carry = 0
			}
		}

		if err != nil {
			if err == io.EOF {
				break
			}
			return false, fmt.Errorf(format, "chunk read", err)
		}
	}

	return false, nil
}
