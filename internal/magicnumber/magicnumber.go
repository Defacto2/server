// Package magicnumber contains the magic number matchers for identifying file types.
package magicnumber

/*
Magic number matchers are used to identify file types by examining the first few bytes of a file.
The matchers are not foolproof and may return false positives. The matchers are used to help
determine the file type before attempting to uncompress or decode the file.

Some resources used to create these matchers,
Wikipedia:
https://en.wikipedia.org/wiki/List_of_file_signatures
https://en.wikipedia.org/wiki/ZIP_(file_format)
The structure of a PKZip file by Florian Buchholz:
https://users.cs.jmu.edu/buchhofp/forensics/formats/pkzip.html
PKWARE ZIP APPNOTE.TXT:
https://pkware.cachefly.net/webdocs/APPNOTE/APPNOTE-2.0.txt
Shrink, Reduce, and Implode: The Legacy Zip Compression Methods:
https://www.hanshq.net/zip2.html
ZIP file tests:
https://github.com/jvilk/browserfs-zipfs-extras/tree/master/test/fixtures
*/

import (
	"archive/zip"
	"bytes"
	"io"
	"net/http"
	"strings"

	"github.com/Defacto2/server/internal/magicnumber/pkzip"
)

// ANSIB matches attempts to match ANSI escape sequences used in text files.
// Some BBS text files are prefixed with the reset sequence but are not ANSI encoded texts.
// For performance, this matcher only looks for reset plus the clean at the start of Amiga
// texts or incomplete bold or normal text graphics mode sequences for DOS art.
func ANSI(r io.Reader) bool {
	buf, err := io.ReadAll(r)
	if err != nil {
		return false
	}
	return ANSIB(buf)
}

// ANSIB matches attempts to match ANSI escape sequences in the byte slice.
func ANSIB(p []byte) bool {
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

// ArcSea returns true if the reader uses the ARC compression format created by
// System Enhancement Associates and used in the MS/PC-DOS and BBS communities.
// See, http://fileformats.archiveteam.org/wiki/ARC_(compression_format).
func ArcSea(r io.Reader) bool {
	buf := make([]byte, 2)
	if _, err := io.ReadFull(r, buf); err != nil {
		return false
	}
	return ArcSeaB(buf)
}

// ArcSeaB matches the ARC compression format in the byte slice.
func ArcSeaB(p []byte) bool {
	const min = 2
	if len(p) < min {
		return false
	}
	const (
		id     = 0x1a
		method = 0x11 // max method id for ARC compression format
	)
	return p[0] == id && p[1] <= method
}

// ARJ returns true if the reader uses the ARJ compressed format developed by Robert Jung.
// See, http://fileformats.archiveteam.org/wiki/ARJ.
func ARJ(r io.Reader) bool {
	buf := make([]byte, 11)
	if _, err := io.ReadFull(r, buf); err != nil {
		return false
	}
	return ARJB(buf)
}

// ARJB matches ARJ compression format in the byte slice.
func ARJB(p []byte) bool {
	const min = 11
	if len(p) < min {
		return false
	}
	const (
		id        = 0x60
		signature = 0xea
		offset    = 0x02
	)
	return p[0] == id && p[1] == signature && p[10] == offset
}

// DOSCom returns true if the reader matches a MS-DOS command executable.
// It is not a reliable matcher but is a common discovery technique.
// See, http://fileformats.archiveteam.org/wiki/DOS_executable_(.com).
func DOSCom(r io.Reader) bool {
	buf := make([]byte, 2)
	if _, err := io.ReadFull(r, buf); err != nil {
		return false
	}
	return DOSComB(buf)
}

// DOSComB matches MS-DOS executable files in the byte slice.
func DOSComB(p []byte) bool {
	const min = 2
	if len(p) < min {
		return false
	}
	const (
		shortJumpE9 = 0xe9
		shortJumpEB = 0xeb
	)
	return p[0] == shortJumpE9 || p[0] == shortJumpEB
}

// InterchangeFF returns true if the reader contains a
// Interchange File Format (IFF) signature.
// This is a generic matcher for IFF bitmap images originally created by
// Electronic Arts for use on Amiga systems in 1985.
// See, http://fileformats.archiveteam.org/wiki/IFF.
func InterchangeFF(r io.Reader) bool {
	buf := make([]byte, 12)
	if _, err := io.ReadFull(r, buf); err != nil {
		return false
	}
	return InterchangeFFB(buf)
}

// InterchangeFFB matches Interchange File Format (IFF) files in the byte slice.
func InterchangeFFB(p []byte) bool {
	const min = 12
	if len(p) < min {
		return false
	}
	if !bytes.Equal(p[0:4], []byte{'F', 'O', 'R', 'M'}) {
		return false
	}
	return bytes.Equal(p[8:12], []byte{'I', 'L', 'B', 'M'})
}

// PCX returns true if the reader begins with a
// ZSoft Corporation PCX (Personal Computer eXchange) signature.
// See, http://fileformats.archiveteam.org/wiki/PCX.
func PCX(r io.Reader) bool {
	buf := make([]byte, 3)
	if _, err := io.ReadFull(r, buf); err != nil {
		return false
	}
	return PCXB(buf)
}

// PCXB ,matches the ZSoft Corporation PCX (Personal Computer eXchange) signature in the byte slice.
func PCXB(p []byte) bool {
	if len(p) < 1 {
		return false
	}
	id := p[0]  // idenfitier
	ver := p[1] // version of pcx
	enc := p[2] // encoding (0 = uncompressed, 1 = run-length encoding compressed)

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

func pngFileSignature() []byte {
	return []byte{137, 80, 78, 71, 13, 10, 26, 10}
}

// PNG returns true if the reader beings with a PNG image file signature.
func PNG(r io.Reader) bool {
	buf := make([]byte, len(pngFileSignature()))
	if _, err := io.ReadFull(r, buf); err != nil {
		return false
	}
	return bytes.EqualFold(buf, pngFileSignature())
}

// PNG returns true if the byte slice has a PNG file signature.
func PNGB(p []byte) bool {
	if len(p) < len(pngFileSignature()) {
		return false
	}
	return bytes.EqualFold(p[:8], pngFileSignature())
}

// Pkzip returns true if the reader begins with a PKZip file signature.
func Pkzip(r io.Reader) bool {
	buf := make([]byte, 4)
	if _, err := io.ReadFull(r, buf); err != nil {
		return false
	}
	return PkzipB(buf)
}

// PkzipB matches the PKZip file signature in the byte slice.
func PkzipB(p []byte) bool {
	const min = 4
	if len(p) < min {
		return false
	}
	const (
		id1 = 0x50
		id2 = 0x4b
		id3 = 0x03
		id4 = 0x04
	)
	return p[0] == id1 && p[1] == id2 && p[2] == id3 && p[3] == id4
}

// PkzipComp returns the PKZip compression methods used in the named file.
func PkzipComp(name string) ([]pkzip.Compression, error) {
	r, err := zip.OpenReader(name)
	if err != nil {
		return nil, err
	}
	defer r.Close()
	methods := []pkzip.Compression{}
	for _, file := range r.File {
		fh := file.FileHeader
		methods = append(methods, pkzip.Compression(fh.Method))
	}
	return methods, nil
}

// Zip returns true if the named file is a PKZip file that only uses the
// Deflated or Stored compression methods. These are  the only methods
// supported by the Go standard library's archive/zip package.
func Zip(name string) (bool, error) {
	methods, err := PkzipComp(name)
	if err != nil {
		return false, err
	}
	for _, m := range methods {
		if m == pkzip.Deflated || m == pkzip.Stored {
			continue
		}
		return false, nil
	}
	return true, nil
}
