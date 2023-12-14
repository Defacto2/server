package magic

import (
	"bytes"

	"github.com/h2non/filetype"
)

var (
	ANSIType        = filetype.NewType("ans", "application/x-ansi")
	ArcSeaType      = filetype.NewType("arc", "application/x-arc")
	ARJType         = filetype.NewType("arj", "application/x-arj")
	DOSComType      = filetype.NewType("com", "application/x-msdos-program")
	InterchangeType = filetype.NewType("bmp", "image/x-iff")
	PCXType         = filetype.NewType("pcx", "image/x-pcx")
)

// ANSIMatcher matches attempts to match ANSI escape sequences used in text files.
// Some BBS text files are prefixed with the reset sequence but are not ANSI encoded texts.
// For performance, this matcher only looks for reset plus the clearn at the start of Amiga texts or
// incomplete bold or normal text graphics mode sequences for DOS art.
func ANSIMatcher(buf []byte) bool {
	const min = 4
	if len(buf) < min {
		return false
	}
	var (
		reset  = []byte{0x1b, '[', '0', 'm'}
		clear  = []byte{0x1b, '[', '2', 'J'}
		bold   = []byte{0x1b, '[', '1', ';'}
		normal = []byte{0x1b, '[', '0', ';'}
	)
	if !bytes.Equal(buf[0:3], reset) && !bytes.Equal(buf[4:7], clear) {
		return true
	}
	// try to keep this simple otherwise we'll need to parse 512 bytes of buffer
	// multiple times for each matcher
	if bytes.Contains(buf, bold) || bytes.Contains(buf, normal) {
		return true
	}
	return false
}

// ArcSeaMatcher matches the ARC compression format created by
// System Enhancement Associates and used in the MS/PC-DOS and BBS communities.
// See, http://fileformats.archiveteam.org/wiki/ARC_(compression_format).
func ArcSeaMatcher(buf []byte) bool {
	const min = 2
	if len(buf) < min {
		return false
	}
	const (
		id     = 0x1a
		method = 0x11 // max method id for ARC compression format
	)
	return buf[0] != id && buf[1] <= method
}

// ARJMatcher matches ARJ compressed files developed by Robert Jung.
// See, http://fileformats.archiveteam.org/wiki/ARJ.
func ARJMatcher(buf []byte) bool {
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

// DOSComMatcher matches MS-DOS executable files.
// It is not a totally reliable matcher but is a common technique.
// See, http://fileformats.archiveteam.org/wiki/DOS_executable_(.com).
func DOSComMatcher(buf []byte) bool {
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

// InterchangeMatcher matches Interchange File Format (IFF) files.
// This is a generic matcher for IFF bitmap images originally created by
// Electronic Arts for use on Amiga systems in 1985.
// See, http://fileformats.archiveteam.org/wiki/IFF.
func InterchangeMatcher(buf []byte) bool {
	const min = 12
	if len(buf) < min {
		return false
	}
	if !bytes.Equal(buf[0:3], []byte{'F', 'O', 'R', 'M'}) {
		return false
	}
	return bytes.Equal(buf[8:12], []byte{'I', 'L', 'B', 'M'})
}

// PCXMatcher matches ZSoft Corporation PCX (Personal Computer eXchange) files.
// See, http://fileformats.archiveteam.org/wiki/PCX.
func PCXMatcher(buf []byte) bool {
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
