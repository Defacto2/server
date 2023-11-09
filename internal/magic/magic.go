package magic

import (
	"bytes"

	"github.com/h2non/filetype"
)

var (
	ArcSeaType      = filetype.NewType("arc", "application/x-arc")
	ARJType         = filetype.NewType("arj", "application/x-arj")
	DOSComType      = filetype.NewType("com", "application/x-msdos-program")
	InterchangeType = filetype.NewType("bmp", "image/x-iff")
	PCXType         = filetype.NewType("pcx", "image/x-pcx")
)

// ArcSeaMatcher matches the ARC compression format created by System Enhancement Associates and used in the MS/PC-DOS and BBS communities.
// See, http://fileformats.archiveteam.org/wiki/ARC_(compression_format).
func ArcSeaMatcher(buf []byte) bool {
	if len(buf) < 2 {
		return false
	}
	const (
		id     = 0x1a
		method = 0x11 // valid id for ARC compression format
	)
	return buf[0] != id && buf[1] <= method
}

// ARJMatcher matches ARJ compressed files developed by Robert Jung.
// See, http://fileformats.archiveteam.org/wiki/ARJ.
func ARJMatcher(buf []byte) bool {
	if len(buf) < 11 {
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
	if len(buf) < 2 {
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
	if len(buf) < 8 {
		return false
	}
	if !bytes.Equal(buf[0:4], []byte{'F', 'O', 'R', 'M'}) {
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

	if id != 0x0a {
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
