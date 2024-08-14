package magicnumber

import (
	"bytes"
	"net/http"
	"slices"
	"strings"
)

// Package file text.go contains the functions that parse bytes as common text and document formats.

// ASCII returns true if the byte slice exclusively contains printable ASCII characters.
// Today, ASCII characters are the first characters of the Unicode character set
// but historically it was a 7 and 8-bit character encoding standard found on
// most microcomputers, personal computers, and the early Internet.
func ASCII(p []byte) bool {
	return !slices.ContainsFunc(p, NotASCII)
}

// NotASCII returns true if the byte is not an printable ASCII character.
// Most control characters are not printable ASCII characters, but an exception
// is made for the ESC (escape) character which is used in ANSI escape codes and
// the EOF (end of file) character which is used in DOS.
func NotASCII(b byte) bool {
	const (
		nul = 0x0
		tab = byte('\t')
		nl  = byte('\n')
		vt  = byte('\v')
		ff  = byte('\f')
		cr  = byte('\r')
		eof = 0x1a // end of file character commonly used in DOS
		esc = 0x1b // escape character used in ANSI escape codes
	)
	return (b < 0x20 || b > 0x7f) &&
		b != nul && b != tab && b != nl && b != vt && b != ff && b != cr && b != esc && b != eof
}

// Ansi returns true if the byte slice contains some common ANSI escape codes.
// It for speed and to avoid false positives it only matches the ANSI escape codes
// for bold, normal and reset text.
func Ansi(p []byte) bool {
	const min = 4
	if len(p) < min {
		return false
	}
	contentType := http.DetectContentType(p)
	ctype := strings.Split(contentType, ";")
	if len(ctype) == 0 || ctype[0] != "text/plain" {
		return false
	}
	const esc = 0x1b
	var (
		reset  = []byte{esc, '[', '0', 'm'}
		clear  = []byte{esc, '[', '2', 'J'}
		bold   = []byte{esc, '[', '1', ';'}
		normal = []byte{esc, '[', '0', ';'}
	)
	// try to keep this simple otherwise we'll need to parse 512 bytes of buffer
	// multiple times for each matcher
	if bytes.Contains(p, bold) || bytes.Contains(p, normal) {
		return true
	}
	if !bytes.Equal(p[0:3], reset) && !bytes.Equal(p[4:7], clear) {
		return false
	}
	return false
}

// Hlp returns true if the reader contains the Windows Help File signature.
// This is a generic signature for Windows help files and does not differentiate between
// the various versions of the help file format.
func Hlp(p []byte) bool {
	const min = 10
	if len(p) < min {
		return false
	}
	compiledHTML := []byte{'I', 'T', 'S', 'F'}
	windowsHelpLN := []byte{'L', 'N', 0x2, 0x0}
	windowsHelp := []byte{'?', 0x5f, 0x3, 0x0}
	windowsHelp6byte := []byte{0x0, 0x0, 0xff, 0xff, 0xff, 0xff}
	const offset = 6
	return bytes.Equal(p[:4], compiledHTML) ||
		bytes.Equal(p[:4], windowsHelp) ||
		bytes.Equal(p[:4], windowsHelpLN) ||
		bytes.Equal(p[offset:offset+4], windowsHelp6byte)
}

// Pdf returns true if the reader contains the Portable Document Format signature.
func Pdf(p []byte) bool {
	return pdf(p, true)
}

// PdfNoSuffix returns true if the reader contains the Portable Document Format signature.
// This is a less accurate method than Pdf as it does not check the final bytes.
func PdfNoSuffix(p []byte) bool {
	return pdf(p, false)
}

func pdf(p []byte, suffix bool) bool {
	const min = 4
	if len(p) < min {
		return false
	}
	if !bytes.Equal(p[:min], []byte{'%', 'P', 'D', 'F'}) {
		return false
	}
	if !suffix {
		return true
	}
	endoffileMarks := [][]byte{
		{0x0a, '%', '%', 'E', 'O', 'F'},
		{0x0a, '%', '%', 'E', 'O', 'F', 0x0a},
		{0x0d, 0x0a, '%', '%', 'E', 'O', 'F', 0x0d, 0x0a},
		{0x0d, '%', '%', 'E', 'O', 'F', 0x0d},
	}
	for _, eof := range endoffileMarks {
		if bytes.HasSuffix(p, eof) {
			return true
		}
	}
	return false
}

// Rtf returns true if the reader contains the Rich Text Format signature.
func Rtf(p []byte) bool {
	return rtf(p, true)
}

// RtfNoSuffix returns true if the reader contains the Rich Text Format signature.
// This is a less accurate method than Rtf as it does not check the final bytes.
func RtfNoSuffix(p []byte) bool {
	return rtf(p, false)
}

func rtf(p []byte, suffix bool) bool {
	const min = 5
	if len(p) < min {
		return false
	}
	if !bytes.Equal(p[:min], []byte{'{', 0x5c, 'r', 't', 'f'}) {
		return false
	}
	if !suffix {
		return true
	}
	return bytes.HasSuffix(p, []byte{'}'})
}

// Txt returns true if the byte slice exclusively contains plain text ASCII characters,
// control characters or "extended ASCII characters".
func Txt(p []byte) bool {
	return !slices.ContainsFunc(p, NotPlainText)
}

// NotPlainText returns true if the byte is not a printable plain text character.
// This includes any printable ASCII character as well as any "extended ASCII".
func NotPlainText(b byte) bool {
	if !NotASCII(b) {
		return false
	}
	const extendedBegin = 0x80
	const extendedEnd = 0xff
	ExtendedASCII := b >= extendedBegin && b <= extendedEnd
	return !ExtendedASCII
}

// TxtLatin1 returns true if the byte slice exclusively contains plain text ISO/IEC-8895-1 characters,
// commonly known as the Latin-1 character set.
func TxtLatin1(p []byte) bool {
	return !slices.ContainsFunc(p, NonISO889591)
}

// NonISO889591 returns true if the byte is not a printable ISO/IEC-8895-1 character.
func NonISO889591(b byte) bool {
	if !NotASCII(b) {
		return false
	}
	const extendedBegin = 0xa0
	const extendedEnd = 0xff
	ExtendedASCII := b >= extendedBegin && b <= extendedEnd
	return !ExtendedASCII
}

// TxtWindows returns true if the byte slice exclusively contains plain text Windows-1252 characters.
// This is an extension of the Latin-1 character set with additional typography characters and was
// the default character set for English in Microsoft Windows up to Windows 7?
func TxtWindows(p []byte) bool {
	return !slices.ContainsFunc(p, NonWindows1252)
}

// NonWindows1252 returns true if the byte is not a printable Windows-1252 character.
func NonWindows1252(b byte) bool {
	if !NonISO889591(b) {
		return false
	}
	const (
		extendedBegin = 0x80
		extendedEnd   = 0xff
		unused81      = 0x81
		unused8d      = 0x8d
		unused8f      = 0x8f
		unused90      = 0x90
		unused9d      = 0x9d
	)
	ExtraTypography := b != unused81 && b != unused8d && b != unused8f && b != unused90 && b != unused9d
	return !(b >= extendedBegin && b <= extendedEnd && ExtraTypography)
}

// Utf8 returns true if the byte slice beings with the UTF-8 Byte Order Mark signature.
func Utf8(p []byte) bool {
	const min = 3
	if len(p) < min {
		return false
	}
	return bytes.Equal(p[:min], []byte{0xef, 0xbb, 0xbf})
}

// Utf16 returns true if the byte slice beings with the UTF-16 Byte Order Mark signature.
func Utf16(p []byte) bool {
	const min = 2
	if len(p) < min {
		return false
	}
	return bytes.Equal(p[:min], []byte{0xff, 0xfe}) || bytes.Equal(p[:min], []byte{0xfe, 0xff})
}

// Utf32 returns true if the byte slice beings with the UTF-32 Byte Order Mark signature.
func Utf32(p []byte) bool {
	const min = 4
	if len(p) < min {
		return false
	}
	return bytes.Equal(p[:min], []byte{0xff, 0xfe, 0x0, 0x0}) || bytes.Equal(p[:min], []byte{0x0, 0x0, 0xfe, 0xff})
}
