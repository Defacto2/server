//nolint:gochecknoglobals
package readme

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"regexp"
)

var (
	byteCR   = []byte("\r")
	byteCRLF = []byte("\r\n")
	// byteCRLFx2 = []byte("\n\n").
	byteEOF    = []byte("\x1a")
	byteLF     = []byte("\n")
	byteNull   = []byte{0x00}
	byteSpace  = []byte(" ")
	emptyBytes = []byte{}
	trimCutset = " \x1a" // Space + SUB character
)

// addPrefix injects content before the existing byte content.
// This is usually to inject FILE_ID.DIZ text files.
func addPrefix(p, prefix []byte) []byte {
	if len(bytes.TrimSpace(prefix)) == 0 {
		return p
	}
	if len(bytes.TrimSpace(p)) == 0 {
		return prefix
	}

	sep := []byte("\n\n")
	size := len(prefix) + len(sep) + len(p)

	// Single heap allocation with zero wrapper struct overhead
	buf := make([]byte, 0, size)
	buf = append(buf, prefix...)
	buf = append(buf, sep...)
	buf = append(buf, p...)

	return buf
}

// addSuffix injects content after the existing byte content.
// This is usually to inject an additional helper text file.
func addSuffix(p, suffix []byte) []byte {
	if len(bytes.TrimSpace(suffix)) == 0 {
		return p
	}
	if len(bytes.TrimSpace(p)) == 0 {
		return suffix
	}

	sep := []byte("\n\n")
	size := len(suffix) + len(sep) + len(p)

	// Single heap allocation with zero wrapper struct overhead
	buf := make([]byte, 0, size)
	buf = append(buf, p...)
	buf = append(buf, sep...)
	buf = append(buf, suffix...)

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
		e   = 0x8a // è
		eof = 0x1a // msdos end-of-file marker
	)
	match := bytes.LastIndexByte(s, e)
	if none := match == -1; none {
		return s
	}
	for i := match + 1; i < len(s); i++ {
		if s[i] != eof {
			return s
		}
	}
	return s[:match]
}

const (
	eof          = "\x1a"
	trimBytesSet = "\t\n\v\f\r " + eof
)

// trimBytes removes ending standard space characters and MS-DOS EOF marker (0x1a).
func trimBytes(b []byte) []byte {
	return bytes.TrimRight(b, trimBytesSet)
}

const (
	reSGR   = `\x1b\[`
	reAnsi  = `\x1b\[[0-9;]*[a-zA-Z]`
	reAmiga = `\x1b\[[0-9;]* p`
	reDEC   = `\x1b\[\?[0-9]+\w`
	reSauce = `SAUCE00`

	// reMovePos returns a regular expression for ANSI cursor position escape codes.
	//   - match "1B" (Escape)
	//   - match "[" (Left Bracket)
	//   - match the digits for line number
	//   - match ";" (semicolon)
	//   - match the digits for column number
	//   - match "H" cursor position or "f" cursor position
	reMovePos = `\x1b\[\d+;\d+[Hf]`

	// reMove returns a regular expression for ANSI cursor movement escape codes.
	//   - match "1B" (Escape)
	//   - match "[" (Left Bracket)
	//   - match optional digits or if no digits, then the cursor moves 1 position
	//   - match "A", "B", "C", "D", "E", "F", "G" for cursor movement up, down, left, right, etc.
	reMove = `\x1b\[\d*?[ABCDEFG]`
)

var reControlCodes = regexp.MustCompile(reAnsi + `|` + reDEC + `|` + reAmiga + `|` + reSauce)

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

	b = reControlCodes.ReplaceAll(b, emptyBytes)
	b = bytes.ReplaceAll(b, byteCRLF, byteLF)
	b = bytes.ReplaceAll(b, byteNull, byteSpace)
	b = bytes.TrimRight(b, trimCutset)

	return b
}

var reANSI = regexp.MustCompile(fmt.Sprintf("(?:%s|%s|%s)", reMove, reMovePos, reSGR))

// hasANSI scans the reader and returns true if ANSI escape codes are found.
// If the reader is too large to render in a HTML template, an error is returned.
func hasANSI(r io.Reader) (bool, error) {
	if r == nil {
		return false, nil
	}

	const (
		format    = "match ansi reader %s: %w"
		oneKB     = 1024
		maxBytes  = oneKB * 1024
		chunkSize = oneKB * 32
	)

	buf := make([]byte, chunkSize)
	var read int

	for {
		n, err := r.Read(buf)
		if n > 0 {
			read += n
			if read > maxBytes {
				return false, fmt.Errorf(format, "reader is larger than 1MB", bufio.ErrTooLong)
			}

			if reANSI.Match(buf[:n]) {
				return true, nil
			}
		}

		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return false, fmt.Errorf(format, "chunk read", err)
		}
	}

	return false, nil
}
