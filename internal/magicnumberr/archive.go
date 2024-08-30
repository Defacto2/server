package magicnumberr

import (
	"bytes"
	"io"
)

// Archive reads all the bytes from the reader and returns the file type signature if
// the file is a known archive of files or Unknown if the file is not an archive.
func Archive(r io.ReaderAt) (Signature, error) {
	archives := Archives()
	find := New()
	for _, archive := range archives {
		if finder, exists := find[archive]; exists {
			if finder(r) {
				return archive, nil
			}
		}
	}
	return Unknown, nil
}

//
// Pksfx and Pklite functions can be found in internal/magicnumber/executable.go
//

// Zip64 matches the PKWARE Zip64 archive format.
// This is an extension to the original ZIP format that allows for larger files.
// But it is not widely supported and this method is untested.
func Zip64(r io.ReaderAt) bool {
	const size = 30
	p := make([]byte, size)
	sr := io.NewSectionReader(r, 0, size)
	if n, err := sr.Read(p); err != nil || n < size {
		return false
	}
	localFileHeader := []byte{'P', 'K', 0x3, 0x4}
	if !bytes.Equal(p[:4], localFileHeader) {
		return false
	}
	centralDirectoryHeader := []byte{0x6, 0x6, 0x4b, 0x50}
	centralDirectoryEnd := []byte{0x7, 0x6, 0x4b, 0x50}
	if !bytes.Contains(p, centralDirectoryHeader) || !bytes.Contains(p, centralDirectoryEnd) {
		return false
	}
	return true
}

// Pkzip matches the zip archive format.
func Pkzip(r io.ReaderAt) bool {
	return pkzip(r) == pkZip
}

// PkImplode matches the PKWARE Implode method zip archive format.
// This is a legacy method and is generally not supported in modern ZIP tools and libraries.
func PkImplode(r io.ReaderAt) bool {
	return pkzip(r) == pkImplode
}

// PkReduce matches the PKWARE Reduce method zip archive format.
// This is a legacy method and is generally not supported in modern ZIP tools and libraries.
func PkReduce(r io.ReaderAt) bool {
	return pkzip(r) == pkReduce
}

// PkShrink matches the PKWARE Shrink method zip archive format.
// This is a legacy method and is generally not supported in modern ZIP tools and libraries.
func PkShrink(r io.ReaderAt) bool {
	return pkzip(r) == pkSkrink
}

// PkzipMulti matches the PKWARE Multi-Volume Zip archive format.
func PkzipMulti(r io.ReaderAt) bool {
	const size = 4
	p := make([]byte, size)
	sr := io.NewSectionReader(r, 0, size)
	if n, err := sr.Read(p); err != nil || n < size {
		return false
	}
	return bytes.Equal(p, []byte{'P', 'K', 0x7, 0x8})
}

type pkComp int

const (
	pkNone pkComp = iota
	pkZip
	pkSkrink
	pkReduce
	pkImplode
)

// pkzip matches the PKWARE Zip archive format.
// This is the most common ZIP format and is widely supported and has been
// tested against many discountinued and legacy ZIP methods and packagers.
//
// Due to the complex history of the ZIP format, 4 possible return values
// maybe returned.
//   - pkNone is returned if the file is not a ZIP archive.
//   - pkOkay is returned if the file is a ZIP archive, except for the compression methods below.
//   - pkSkrink is returned if the ZIP archive uses the PKWARE shrink method, found in PKZIP v0.9.
//   - pkReduce is returned if the ZIP archive uses the PKWARE reduction method, found in PKZIP v0.8.
//   - pkImplode is returned if the ZIP archive uses the PKWARE implode method, found in PKZIP v1.01.
//
// Compression methods Shrink, Reduce and Implode are legacy and are generally
// not supported in modern ZIP tools and libraries.
func pkzip(r io.ReaderAt) pkComp {
	const size = 30
	p := make([]byte, size)
	sr := io.NewSectionReader(r, 0, size)
	if n, err := sr.Read(p); err != nil || n < size {
		return pkNone
	}
	if len(p) < size {
		return pkNone
	}
	// local file header signature     4 bytes  (0x04034b50)
	localFileHeader := []byte{'P', 'K', 0x3, 0x4}
	if !bytes.Equal(p[:4], localFileHeader) {
		return pkNone // 50 4b 03 04
	}
	// version needed to extract       2 bytes
	versionNeeded := p[4] + p[5]
	if versionNeeded == 0 {
		// legacy versions of PKZIP returned either 0x.0a (10) or 0x14 (20).
		return pkNone // 0a 00
	}
	// general purpose bit flag        2 bytes
	// skip this as there's too many reserved values that might cause false positive rejections
	//
	// compression method              2 bytes
	compresionMethod := p[8] + p[9]
	const (
		store       = 0x0
		shrink      = 0x1
		reduce1     = 0x2
		reduce2     = 0x3
		reduce3     = 0x4
		reduce4     = 0x5
		implode     = 0x6
		deflate     = 0x8
		deflate64   = 0x9
		ibmTerse    = 0xa
		bzip2       = 0xc
		lzma        = 0xe
		ibmCMPSC    = 0x10
		ibmTerseNew = 0x12
		ibmLZ77z    = 0x13
		zstd        = 0x5d
		mp3         = 0x5e
		xz          = 0x5f
		jpeg        = 0x60
		wavPack     = 0x61
		ppmd        = 0x62
		ae          = 0x63
	)
	switch compresionMethod {
	case store, deflate, deflate64:
		return pkZip
	case shrink:
		return pkSkrink
	case reduce1, reduce2, reduce3, reduce4:
		return pkReduce
	case implode:
		return pkImplode
	case ibmTerse, bzip2, lzma, ibmCMPSC, ibmTerseNew, ibmLZ77z, zstd, mp3, xz, jpeg, wavPack, ppmd, ae:
		return pkZip
	default:
		return pkNone
	}
}

// Tar matches the Tape ARchive format.
func Tar(r io.ReaderAt) bool {
	const offset = 257
	const size = 5
	p := make([]byte, size)
	sr := io.NewSectionReader(r, offset, size)
	if n, err := sr.Read(p); err != nil || n < size {
		return false
	}
	return bytes.Equal(p, []byte{'u', 's', 't', 'a', 'r'})
}

// Rar matches the Roshal ARchive format.
// This method is untested.
func Rar(r io.ReaderAt) bool {
	const size = 7
	p := make([]byte, size)
	sr := io.NewSectionReader(r, 0, size)
	if n, err := sr.Read(p); err != nil || n < size {
		return false
	}
	return bytes.Equal(p, []byte{'R', 'a', 'r', 0x21, 0x1a, 0x7, 0x0})
}

// Rarv5 matches the Roshal ARchive v5 format.
func Rarv5(r io.ReaderAt) bool {
	const size = 8
	p := make([]byte, size)
	sr := io.NewSectionReader(r, 0, size)
	if n, err := sr.Read(p); err != nil || n < size {
		return false
	}
	return bytes.Equal(p, []byte{'R', 'a', 'r', 0x21, 0x1a, 0x7, 0x1, 0x0})
}

// Gzip matches the Gzip Compress archive format.
func Gzip(r io.ReaderAt) bool {
	const size = 3
	p := make([]byte, size)
	sr := io.NewSectionReader(r, 0, size)
	if n, err := sr.Read(p); err != nil || n < size {
		return false
	}
	if bytes.Equal(p, []byte{0x1f, 0x8b, 0x08}) {
		return true
	}
	const offset = 512
	p = make([]byte, size)
	sr = io.NewSectionReader(r, offset, size)
	if n, err := sr.Read(p); err != nil || n < size {
		return false
	}
	return bytes.Equal(p, []byte{0x1f, 0x8b, 0x08})
}

// Bzip2 matches the Bzip2 Compress archive format.
func Bzip2(r io.ReaderAt) bool {
	const size = 3
	p := make([]byte, size)
	sr := io.NewSectionReader(r, 0, size)
	if n, err := sr.Read(p); err != nil || n < size {
		return false
	}
	return bytes.Equal(p, []byte{'B', 'Z', 'h'})
}

// X7z matches the 7z Compress archive format.
func X7z(r io.ReaderAt) bool {
	const size = 6
	p := make([]byte, size)
	sr := io.NewSectionReader(r, 0, size)
	if n, err := sr.Read(p); err != nil || n < size {
		return false
	}
	return bytes.Equal(p, []byte{'7', 'z', 0xbc, 0xaf, 0x27, 0x1c})
}

// XZ matches the XZ Compress archive format.
func XZ(r io.ReaderAt) bool {
	const size = 6
	p := make([]byte, size)
	sr := io.NewSectionReader(r, 0, size)
	if n, err := sr.Read(p); err != nil || n < size {
		return false
	}
	return bytes.Equal(p, []byte{0xfd, '7', 'z', 'X', 'Z', 0x0})
}

// ZStd matches the ZStandard archive format.
func ZStd(r io.ReaderAt) bool {
	const size = 4
	p := make([]byte, size)
	sr := io.NewSectionReader(r, 0, size)
	if n, err := sr.Read(p); err != nil || n < size {
		return false
	}
	return bytes.Equal(p, []byte{0x28, 0xb5, 0x2f, 0xfd})
}

// ArcFree matches the FreeArc compression format.
func ArcFree(r io.ReaderAt) bool {
	const size = 4
	p := make([]byte, size)
	sr := io.NewSectionReader(r, 0, size)
	if n, err := sr.Read(p); err != nil || n < size {
		return false
	}
	return bytes.Equal(p, []byte{'A', 'r', 'C', 0x1})
}

// ArcSEA matches the ARChive SEA compression format.
func ArcSEA(r io.ReaderAt) bool {
	const size = 2
	p := make([]byte, size)
	sr := io.NewSectionReader(r, 0, size)
	if n, err := sr.Read(p); err != nil || n < size {
		return false
	}
	const (
		id     = 0x1a
		method = 0x11 // max method id for ARC compression format
	)
	return p[0] == id && p[1] <= method
}

// LzhLha matches the LHA and LZH compression formats.
func LzhLha(r io.ReaderAt) bool {
	const offset = 2
	const size = 3
	p := make([]byte, size)
	sr := io.NewSectionReader(r, offset, size)
	if n, err := sr.Read(p); err != nil || n < size {
		return false
	}
	return bytes.Equal(p, []byte{'-', 'l', 'h'})
}

// Zoo matches the Zoo compression format.
// This method is untested.
func Zoo(r io.ReaderAt) bool {
	const size = 4
	p := make([]byte, size)
	sr := io.NewSectionReader(r, 0, size)
	if n, err := sr.Read(p); err != nil || n < size {
		return false
	}
	return bytes.Equal(p, []byte{'Z', 'O', 'O', 0x20})
}

// Arj matches ARJ compression format.
func Arj(r io.ReaderAt) bool {
	const size = 11
	p := make([]byte, size)
	sr := io.NewSectionReader(r, 0, size)
	if n, err := sr.Read(p); err != nil || n < size {
		return false
	}
	const (
		id        = 0x60
		signature = 0xea
		offset    = 0x02
	)
	return p[0] == id && p[1] == signature && p[10] == offset
}

// Cab matches the Microsoft CABinet archive format.
// This method is untested.
func Cab(r io.ReaderAt) bool {
	const size = 4
	p := make([]byte, size)
	sr := io.NewSectionReader(r, 0, size)
	if n, err := sr.Read(p); err != nil || n < size {
		return false
	}
	return bytes.Equal(p, []byte{'M', 'S', 'C', 'F'})
}
