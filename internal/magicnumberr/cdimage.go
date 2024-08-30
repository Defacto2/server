package magicnumberr

// File cdimage.go contains the file type signature for physical media disk image formats.

import (
	"bytes"
	"io"
)

// Daa returns true if the reader contains the PowerISO DAA CD image signature.
func Daa(r io.ReaderAt) bool {
	const size = 8
	p := make([]byte, size)
	sr := io.NewSectionReader(r, 0, size)
	if n, err := sr.Read(p); err != nil || n < size {
		return false
	}
	return bytes.Equal(p, []byte{'D', 'A', 'A', 0x0, 0x0, 0x0, 0x0, 0x0})
}

// ISO returns true if the reader contains the ISO 9660 CD-ROM filesystem signature.
// To be accurate, it requires at least 36KB of data to be read.
func ISO(r io.ReaderAt) bool {
	const size = 5
	p := make([]byte, size)
	offsets := []int64{0, 32769, 34817, 36865}
	for _, offset := range offsets {
		sr := io.NewSectionReader(r, offset, size)
		if n, err := sr.Read(p); err != nil || n < size {
			return false
		}
		if bytes.Equal(p, []byte{0x43, 0x44, 0x30, 0x30, 0x31}) {
			return true
		}
	}
	return false
}

// Mdf returns true if the reader contains the Alcohol 120% MDF CD image signature.
func Mdf(r io.ReaderAt) bool {
	const size = 16
	p := make([]byte, size)
	sr := io.NewSectionReader(r, 0, size)
	if n, err := sr.Read(p); err != nil || n < size {
		return false
	}
	return bytes.Equal(p,
		[]byte{
			0x00, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
			0xff, 0xff, 0xff, 0x00, 0x00, 0x02, 0x00, 0x01,
		})
}

// Nri returns true if the reader contains the Nero CD image signature.
// This method is untested.
func Nri(r io.ReaderAt) bool {
	const size = 8
	p := make([]byte, size)
	sr := io.NewSectionReader(r, 0, size)
	if n, err := sr.Read(p); err != nil || n < size {
		return false
	}
	return bytes.Equal(p, []byte{0x0e, 'N', 'e', 'r', 'o', 'I', 'S', 'O'})
}
