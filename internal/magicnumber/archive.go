package magicnumber

import "bytes"

// Package file archive.go contains the functions that parse bytes as common file archive, compression and disk image formats.

// Arc matches the FreeArc compression format in the byte slice.
func Arc(p []byte) bool {
	const min = 4
	if len(p) < min {
		return false
	}
	return bytes.Equal(p[:min], []byte{'A', 'r', 'C', 0x1})
}

// ArcArk matches the ARChive SEA compression format in the byte slice.
func ArcArk(p []byte) bool {
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

// Arj matches ARJ compression format in the byte slice.
func Arj(p []byte) bool {
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

// Bzip2 matches the Bzip2 Compress archive format in the byte slice.
func Bzip2(p []byte) bool {
	const min = 3
	if len(p) < min {
		return false
	}
	return bytes.Equal(p[:min], []byte{'B', 'Z', 'h'})
}

// Cab matches the Microsoft CABinet archive format in the byte slice.
func Cab(p []byte) bool {
	const min = 4
	if len(p) < min {
		return false
	}
	return bytes.Equal(p[:min], []byte{'M', 'S', 'C', 'F'})
}

// Gzip matches the Gzip Compress archive format in the byte slice.
func Gzip(p []byte) bool {
	const min = 3
	if len(p) < min {
		return false
	}
	return bytes.Equal(p[:min], []byte{0x1f, 0x8b, 0x8})
}

// Jar returns true if the reader contains the Java ARchive signature.
func Jar(p []byte) bool {
	const min = 10
	if len(p) < min {
		return false
	}
	return bytes.Equal(p[:min], []byte{0x50, 0x4b, 0x3, 0x4, 0x14, 0x0, 0x8, 0x0, 0x8, 0x0})
}

// LzhLha matches the LHA and LZH compression formats in the byte slice.
func LzhLha(p []byte) bool {
	const min = 5
	if len(p) < min {
		return false
	}
	const offset = 2
	return bytes.Equal(p[offset:3+offset], []byte{'-', 'l', 'h'})
}

//
// Pksfx and Pklite functions can be found in internal/magicnumber/executable.go
//

func Pkzip(p []byte) bool {
	return pkzip(p) == pkZip
}

func PkImplode(p []byte) bool {
	return pkzip(p) == pkImplode
}

func PkReduce(p []byte) bool {
	return pkzip(p) == pkReduce
}

func PkShrink(p []byte) bool {
	return pkzip(p) == pkSkrink
}

// PkzipMulti matches the PKWARE Multi-Volume Zip archive format in the byte slice.
func PkzipMulti(p []byte) bool {
	const min = 4
	if len(p) < min {
		return false
	}
	return bytes.Equal(p[:min], []byte{'P', 'K', 0x7, 0x8})
}

type pkComp int

const (
	pkNone pkComp = iota
	pkZip
	pkSkrink
	pkReduce
	pkImplode
)

// pkzip matches the PKWARE Zip archive format in the byte slice.
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
func pkzip(p []byte) pkComp {
	const min = 30
	if len(p) < min {
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

// Rar matches the Roshal ARchive format in the byte slice.
func Rar(p []byte) bool {
	const min = 7
	if len(p) < min {
		return false
	}
	return bytes.Equal(p[:min], []byte{'R', 'a', 'r', 0x21, 0x1a, 0x7, 0x0})
}

// Rarv5 matches the Roshal ARchive v5 format in the byte slice.
func Rarv5(p []byte) bool {
	const min = 8
	if len(p) < min {
		return false
	}
	return bytes.Equal(p[:min], []byte{'R', 'a', 'r', 0x21, 0x1a, 0x7, 0x1, 0x0})
}

// Tar matches the Tape ARchive format in the byte slice.
func Tar(p []byte) bool {
	const min = 5
	const offset = 257
	if len(p) < min+offset {
		return false
	}
	return bytes.Equal(p[offset:min+offset], []byte{'u', 's', 't', 'a', 'r'})
}

// X7z matches the 7z Compress archive format in the byte slice.
func X7z(p []byte) bool {
	const min = 6
	if len(p) < min {
		return false
	}
	return bytes.Equal(p[:min], []byte{'7', 'z', 0xbc, 0xaf, 0x27, 0x1c})
}

// XZ matches the XZ Compress archive format in the byte slice.
func XZ(p []byte) bool {
	const min = 6
	if len(p) < min {
		return false
	}
	return bytes.Equal(p[:min], []byte{0xfd, '7', 'z', 'X', 'Z', 0x0})
}

// Zip64 matches the PKWARE Zip64 archive format in the byte slice.
// This is an extension to the original ZIP format that allows for larger files.
// But it is not widely supported.
func Zip64(p []byte) bool {
	const min = 30
	if len(p) < min {
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

// Zoo matches the Zoo compression format in the byte slice.
func Zoo(p []byte) bool {
	const min = 4
	if len(p) < min {
		return false
	}
	return bytes.Equal(p[:min], []byte{'Z', 'O', 'O', 0x20})
}

// ZStd matches the ZStandard archive format in the byte slice.
func ZStd(p []byte) bool {
	const min = 4
	if len(p) < min {
		return false
	}
	return bytes.Equal(p[:min], []byte{0x28, 0xb5, 0x2f, 0xfd})
}

//
// Phyiscal media disk image formats
//

// Daa returns true if the reader contains the PowerISO DAA CD image signature.
func Daa(p []byte) bool {
	const min = 8
	if len(p) < min {
		return false
	}
	return bytes.Equal(p[:min], []byte{'D', 'A', 'A', 0x0, 0x0, 0x0, 0x0, 0x0})
}

// ISO returns true if the reader contains the ISO 9660 CD-ROM filesystem signature.
// To be accurate, it requires at least 36KB of data to be read.
func ISO(p []byte) bool {
	const min = 5
	if len(p) < min {
		return false
	}
	offsets := []int{0, 32769, 34817, 36865}
	for _, offset := range offsets {
		if len(p) < min+offset {
			return false
		}
		if bytes.Equal(p[offset:min+offset], []byte{0x43, 0x44, 0x30, 0x30, 0x31}) {
			return true
		}
	}
	return false
}

// Mdf returns true if the reader contains the Alcohol 120% MDF CD image signature.
func Mdf(p []byte) bool {
	const min = 16
	if len(p) < min {
		return false
	}
	return bytes.Equal(p[:min],
		[]byte{
			0x0, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
			0xff, 0xff, 0xff, 0x0, 0x0, 0x2, 0x0, 0x1,
		})
}

// Nri returns true if the reader contains the Nero CD image signature.
func Nri(p []byte) bool {
	const min = 8
	if len(p) < min {
		return false
	}
	return bytes.Equal(p[:min], []byte{0x0e, 'N', 'e', 'r', 'o', 'I', 'S', 'O'})
}
