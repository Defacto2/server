// Package magicnumber contains the magic number matchers for identifying file types.
package magicnumber

import (
	"bytes"
	"net/http"
	"strings"
)

// ANSI matches attempts to match ANSI escape sequences used in text files.
// Some BBS text files are prefixed with the reset sequence but are not ANSI encoded texts.
// For performance, this matcher only looks for reset plus the clean at the start of Amiga
// texts or incomplete bold or normal text graphics mode sequences for DOS art.
func ANSI(buf []byte) bool {
	const min = 4
	if len(buf) < min {
		return false
	}
	contentType := http.DetectContentType(buf)
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
	if bytes.Contains(buf, bold) || bytes.Contains(buf, normal) {
		return true
	}
	if !bytes.Equal(buf[0:3], reset) && !bytes.Equal(buf[4:7], clear) {
		return false
	}
	return false
}

// ArcSea matches the ARC compression format created by
// System Enhancement Associates and used in the MS/PC-DOS and BBS communities.
// See, http://fileformats.archiveteam.org/wiki/ARC_(compression_format).
func ArcSea(buf []byte) bool {
	const min = 2
	if len(buf) < min {
		return false
	}
	const (
		id     = 0x1a
		method = 0x11 // max method id for ARC compression format
	)
	return buf[0] == id && buf[1] <= method
}

// ARJ matches ARJ compressed files developed by Robert Jung.
// See, http://fileformats.archiveteam.org/wiki/ARJ.
func ARJ(buf []byte) bool {
	const min = 11
	if len(buf) < min {
		return false
	}
	const (
		id        = 0x60
		signature = 0xea
		offset    = 0x02
	)
	return buf[0] == id && buf[1] == signature && buf[10] == offset
}

// DOSCom matches MS-DOS executable files.
// It is not a totally reliable matcher but is a common technique.
// See, http://fileformats.archiveteam.org/wiki/DOS_executable_(.com).
func DOSCom(buf []byte) bool {
	const min = 2
	if len(buf) < min {
		return false
	}
	const (
		shortJumpE9 = 0xe9
		shortJumpEB = 0xeb
	)
	return buf[0] == shortJumpE9 || buf[0] == shortJumpEB
}

// InterchangeFF matches Interchange File Format (IFF) files.
// This is a generic matcher for IFF bitmap images originally created by
// Electronic Arts for use on Amiga systems in 1985.
// See, http://fileformats.archiveteam.org/wiki/IFF.
func InterchangeFF(buf []byte) bool {
	const min = 12
	if len(buf) < min {
		return false
	}
	if !bytes.Equal(buf[0:4], []byte{'F', 'O', 'R', 'M'}) {
		return false
	}
	return bytes.Equal(buf[8:12], []byte{'I', 'L', 'B', 'M'})
}

// PCX matches ZSoft Corporation PCX (Personal Computer eXchange) files.
// See, http://fileformats.archiveteam.org/wiki/PCX.
func PCX(buf []byte) bool {
	if len(buf) < 1 {
		return false
	}
	id := buf[0]  // idenfitier
	ver := buf[1] // version of pcx
	enc := buf[2] // encoding (0 = uncompressed, 1 = run-length encoding compressed)

	const pcx = 0x0a
	if id != pcx {
		return false
	}
	if ver != 0x00 && ver != 0x02 && ver != 0x03 && ver != 0x04 && ver != 0x05 {
		return false
	}
	if enc != 0x00 && enc != 0x01 {
		return false
	}
	return true
}

// PNG returns true if the byte slice has a PNG file signature.
func PNG(buf []byte) bool {
	fileSignature := []byte{137, 80, 78, 71, 13, 10, 26, 10}
	if len(buf) < len(fileSignature) {
		return false
	}
	return bytes.EqualFold(buf[:8], fileSignature)
}
