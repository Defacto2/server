// Package magicnumber contains the magic number matchers for identifying file types that
// are expected to be handled by the Defacto2 server application. Magic numbers are not
// always accurate and should be used as hints combined with other checks such as
// file extension matching.
//
// Usually, the magic number is the first few bytes of a file that uniquely identify the file type.
// But a number of document formats also check the final few bytes of a file.
//
// The sources for the magic numbers byte values are from the following:
//   - [Gary Kessler's File Signatures Table]
//   - [Just Solve the File Format Problem]
//   - [OSDev Wiki]
//   - [Wikipedia]
//
// [Gary Kessler's File Signatures Table]: https://www.garykessler.net/library/file_sigs.html
// [Just Solve the File Format Problem]: http://fileformats.archiveteam.org/wiki/Electronic_File_Formats
// [OSDev Wiki]: https://wiki.osdev.org]
// [Wikipedia]: https://en.wikipedia.org/wiki/List_of_file_signatures
package magicnumber

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"slices"
	"strings"
)

type Signature int

const (
	Unknown Signature = iota - 1
	ElectronicArtsIFF
	AV1ImageFile
	JPEGFileInterchangeFormat
	JPEG2000
	PortableNetworkGraphics
	GraphicsInterchangeFormat
	GoogleWebP
	TaggedImageFileFormat
	BMPFileFormat
	PersonalComputereXchange
	InterleavedBitmap
	MicrosoftIcon
	MPEG4
	QuickTimeMovie
	QuickTimeM4V
	MicrosoftAudioVideoInterleave
	MicrosoftWindowsMedia
	MPEG
	FlashVideo
	RealPlayer
	MusicalInstrumentDigitalInterface
	MPEG1AudioLayer3
	OggVorbisCodec
	FreeLosslessAudioCodec
	WaveAudioForWindows
	PKWAREZip64
	PKWAREZip
	PKWAREMultiVolume
	PKLITE
	PKSFX
	TapeARchive
	RoshalARchive
	RoshalARchivev5
	GzipCompressArchive
	Bzip2CompressArchive
	x7zCompressArchive
	XZCompressArchive
	ZStandardArchive
	FreeArc
	ARChiveSEA
	YoshiLHA
	ZooArchive
	ArchiveRobertJung
	MicrosoftCABinet
	MicrosoftDOSKWAJ
	MicrosoftDOSSZDD
	MicrosoftExecutable
	MicrosoftCompoundFile
	ISO9660
	ISONeroCD
	ISOPowerISO
	CDAlcohol120
	JavaARchive
	PortableDocumentFormat
	RichTextFromat
	UTF8Text
	UTF16Text
	UTF32Text
	ANSIEscapeText
	PlainText
)

type Matcher func([]byte) bool

type Finder map[Signature]Matcher

func New() Finder {
	return Finder{
		ElectronicArtsIFF:                 Iff,
		AV1ImageFile:                      Avif,
		JPEGFileInterchangeFormat:         Jpeg,
		JPEG2000:                          Jpeg2000,
		PortableNetworkGraphics:           Png,
		GraphicsInterchangeFormat:         Gif,
		GoogleWebP:                        Webp,
		TaggedImageFileFormat:             Tiff,
		BMPFileFormat:                     Bmp,
		PersonalComputereXchange:          Pcx,
		InterleavedBitmap:                 Ilbm,
		MicrosoftIcon:                     Ico,
		MPEG4:                             Mp4,
		QuickTimeMovie:                    QTMov,
		QuickTimeM4V:                      M4v,
		MicrosoftAudioVideoInterleave:     Avi,
		MicrosoftWindowsMedia:             Wmv,
		MPEG:                              Mpeg,
		FlashVideo:                        Flv,
		RealPlayer:                        Ivr,
		MusicalInstrumentDigitalInterface: Midi,
		MPEG1AudioLayer3:                  Mp3,
		OggVorbisCodec:                    Ogg,
		FreeLosslessAudioCodec:            Flac,
		WaveAudioForWindows:               Wave,
		PKWAREZip64:                       Zip64,
		PKWAREZip:                         Pkzip,
		PKWAREMultiVolume:                 PkzipMulti,
		PKLITE:                            Pklite,
		PKSFX:                             Pksfx,
		TapeARchive:                       Tar,
		RoshalARchive:                     Rar,
		RoshalARchivev5:                   Rarv5,
		GzipCompressArchive:               Gzip,
		Bzip2CompressArchive:              Bzip2,
		x7zCompressArchive:                X7z,
		XZCompressArchive:                 XZ,
		ZStandardArchive:                  ZStd,
		FreeArc:                           Arc,
		ARChiveSEA:                        ArcArk,
		YoshiLHA:                          LzhLha,
		ZooArchive:                        Zoo,
		ArchiveRobertJung:                 Arj,
		MicrosoftCABinet:                  Cab,
		MicrosoftDOSKWAJ:                  DosKWAJ,
		MicrosoftDOSSZDD:                  DosSZDD,
		MicrosoftExecutable:               MSExe,
		MicrosoftCompoundFile:             MSComp,
		ISO9660:                           ISO,
		ISONeroCD:                         Nri,
		ISOPowerISO:                       Daa,
		CDAlcohol120:                      Mdf,
		JavaARchive:                       Jar,
		PortableDocumentFormat:            Pdf,
		RichTextFromat:                    Rtf,
		UTF8Text:                          Utf8,
		UTF16Text:                         Utf16,
		UTF32Text:                         Utf32,
		ANSIEscapeText:                    Ansi,
		PlainText:                         Txt,
	}
}

func Find(r io.Reader) (Signature, error) {
	buf, err := io.ReadAll(r)
	if err != nil {
		return Unknown, err
	}
	return FindBytes(buf), nil
}

func Find1K(r io.Reader) (Signature, error) {
	buf := make([]byte, 1024) // 1KB buffer
	_, err := io.ReadFull(r, buf)
	if err != nil && !errors.Is(err, io.ErrUnexpectedEOF) {
		return Unknown, err
	}
	return FindBytes(buf), nil // create FindBytes1K, where suffix matchers are not used nor is ansi escape lookup
}

func FindBytes(p []byte) Signature {
	if p == nil {
		return Unknown
	}
	find := New()
	for sig, matcher := range find {
		if matcher(p) {
			return sig
		}
	}
	return Unknown
}

func Iff(p []byte) bool {
	const min = 4
	if len(p) < min {
		return false
	}
	return bytes.Equal(p[:4], []byte{'C', 'A', 'T', 0x20})
}

func Avif(p []byte) bool {
	const min = 3
	if len(p) < min {
		return false
	}
	return bytes.Equal(p[:min], []byte{0x0, 0x0, 0x0})
}

func Jpeg(p []byte) bool {
	const min = 4
	if len(p) < min {
		return false
	}
	if !bytes.Equal(p[:3], []byte{0xff, 0xd8, 0xff}) {
		return false
	}
	if p[4] != 0xe0 && p[4] != 0xe1 {
		return false
	}
	if !bytes.Equal(p[6:11], []byte{'J', 'F', 'I', 'F', 0x0}) &&
		!bytes.Equal(p[6:11], []byte{'E', 'x', 'i', 'f', 0x0}) {
		return false
	}
	return bytes.HasSuffix(p, []byte{0xff, 0xd9})
}

func Jpeg2000(p []byte) bool {
	const min = 10
	if len(p) < min {
		return false
	}
	return bytes.Equal(p[:min], []byte{0x0, 0x0, 0x0, 0xc, 0x6a, 0x50, 0x20, 0x20, 0xd, 0xa})
}

func Png(p []byte) bool {
	const min = 8
	if len(p) < min {
		return false
	}
	return bytes.Equal(p[:min], []byte{0x89, 0x50, 0x4E, 0x47, 0x0d, 0x0a, 0x1a, 0x0a})

}

func Gif(p []byte) bool {
	const min = 6
	if len(p) < min {
		return false
	}
	gif87a := []byte{0x47, 0x49, 0x46, 0x38, 0x37, 0x61}
	gif89a := []byte{0x47, 0x49, 0x46, 0x38, 0x39, 0x61}
	return bytes.Equal(p[:min], gif87a) ||
		bytes.Equal(p[:min], gif89a)
}

func Webp(p []byte) bool {
	const min = 12
	if len(p) < min {
		return false
	}
	return bytes.Equal(p[:4], []byte{'R', 'I', 'F', 'F'}) &&
		bytes.Equal(p[8:12], []byte{'W', 'E', 'B', 'P'})
}

func Tiff(p []byte) bool {
	const min = 4
	if len(p) < min {
		return false
	}
	le := []byte{0x49, 0x49, 0x2a, 0x0}
	be := []byte{0x4d, 0x4d, 0x0, 0x2a}
	return bytes.Equal(p[:min], le) ||
		bytes.Equal(p[:min], be)
}

func Bmp(p []byte) bool {
	const min = 2
	if len(p) < min {
		return false
	}
	return bytes.Equal(p[:min], []byte{'B', 'M'})
}

func Pcx(p []byte) bool {
	const min = 3
	if len(p) < min {
		return false
	}
	id := p[0]
	ver := p[1] // version of PCX v0 through to v5
	enc := p[2] // encoding (0 = uncompressed, 1 = run-length encoding compressed)
	return id == 0x0a && ver <= 0x5 && (enc == 0x0 || enc == 0x1)
}

func Ico(p []byte) bool {
	const min = 4
	if len(p) < min {
		return false
	}
	return bytes.Equal(p[:min], []byte{0x0, 0x0, 0x1, 0x0})
}

func Ilbm(p []byte) bool {
	const min = 12
	if len(p) < min {
		return false
	}
	return bytes.Equal(p[:4], []byte{'F', 'O', 'R', 'M'}) &&
		bytes.Equal(p[8:12], []byte{'I', 'L', 'B', 'M'})
}

func QTMov(p []byte) bool {
	const min = 8
	if len(p) < min {
		return false
	}
	const offset = 4
	return bytes.Equal(p[offset:], []byte{'m', 'o', 'o', 'v'}) ||
		bytes.Equal(p[offset:], []byte{'f', 't', 'y', 'p', 'q', 't'})
}

func Mp4(p []byte) bool {
	const min = 8
	if len(p) < min {
		return false
	}
	return bytes.Equal(p[:min], []byte{'f', 't', 'y', 'p', 'M', 'S', 'N', 'V'})
}

func M4v(p []byte) bool {
	const min = 8
	if len(p) < min {
		return false
	}
	return bytes.Equal(p[:min], []byte{'f', 't', 'y', 'p', 'm', 'p', '4', '2'})
}

func Avi(p []byte) bool {
	const min = 16
	if len(p) < min {
		return false
	}
	return bytes.Equal(p[:4], []byte{'R', 'I', 'F', 'F'}) &&
		bytes.Equal(p[8:16], []byte{'A', 'V', 'I', 0x20, 'L', 'I', 'S', 'T'})
}

func Wmv(p []byte) bool {
	const min = 16
	if len(p) < min {
		return false
	}
	return bytes.Equal(p[:min],
		[]byte{0x30, 0x26, 0xb2, 0x75, 0x8e, 0x66, 0xcf, 0x11,
			0xa6, 0xd9, 0x0, 0xaa, 0x0, 0x62, 0xce, 0x6c})
}

func Mpeg(p []byte) bool {
	const min = 4
	if len(p) < min {
		return false
	}
	return bytes.Equal(p[:3], []byte{0x0, 0x0, 0x1}) && p[4] >= 0xba && p[4] <= 0xbf
}

func Flv(p []byte) bool {
	const min = 4
	if len(p) < min {
		return false
	}
	return bytes.Equal(p[:min], []byte{'F', 'L', 'V', 0x1})
}

func Ivr(p []byte) bool {
	const min = 4
	if len(p) < min {
		return false
	}
	return bytes.Equal(p[:min], []byte{0x2e, 'R', 'E', 'C'}) ||
		bytes.Equal(p[:min], []byte{0x2e, 'R', 'M', 'F'})
}

func Midi(p []byte) bool {
	const min = 4
	if len(p) < min {
		return false
	}
	return bytes.Equal(p[:min], []byte{'M', 'T', 'h', 'd'})
}

func Mp3(p []byte) bool {
	const min = 3
	if len(p) < min {
		return false
	}
	return bytes.Equal(p[:min], []byte{'I', 'D', '3'})
}

func Ogg(p []byte) bool {
	const min = 12
	if len(p) < min {
		return false
	}
	return bytes.Equal(p[:min], []byte{'O', 'g', 'g', 'S', 0x0, 0x2, 0x0, 0x0,
		0x0, 0x0, 0x0, 0x0, 0x0, 0x0})
}

func Flac(p []byte) bool {
	const min = 8
	if len(p) < min {
		return false
	}
	return bytes.Equal(p[:min], []byte{'f', 'L', 'a', 'C', 0x0, 0x0, 0x0, 0x2})
}

func Wave(p []byte) bool {
	const min = 12
	if len(p) < min {
		return false
	}
	return bytes.Equal(p[:4], []byte{'R', 'I', 'F', 'F'}) &&
		bytes.Equal(p[8:16], []byte{'W', 'A', 'V', 'E', 'f', 'm', 't', 0x20})
}

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

func Pkzip(p []byte) bool {
	const min = 30
	if len(p) < min {
		return false
	}
	localFileHeader := []byte{'P', 'K', 0x3, 0x4}
	if !bytes.Equal(p[:4], localFileHeader) {
		return false
	}
	centralDirectoryHeader := []byte{0x2, 0x1, 0x4b, 0x50}
	centralDirectoryEnd := []byte{0x6, 0x5, 0x4b, 0x50}
	if !bytes.Contains(p, centralDirectoryHeader) || !bytes.Contains(p, centralDirectoryEnd) {
		return false
	}
	return true
}

func PkzipMulti(p []byte) bool {
	const min = 4
	if len(p) < min {
		return false
	}
	return bytes.Equal(p[:min], []byte{'P', 'K', 0x7, 0x8})
}

func Pklite(p []byte) bool {
	const min = 6
	const offset = 30
	if len(p) < min+offset {
		return false
	}
	return bytes.Equal(p[offset:min+offset], []byte{0x50, 0x4b, 0x4c, 0x49, 0x54, 0x45})
}

func Pksfx(p []byte) bool {
	const min = 5
	const offset = 526
	if len(p) < min+offset {
		return false
	}
	return bytes.Equal(p[offset:min+offset], []byte{0x50, 0x4b, 0x53, 0x70, 0x58})
}

func Tar(p []byte) bool {
	const min = 5
	const offset = 257
	if len(p) < min+offset {
		return false
	}
	return bytes.Equal(p[offset:min+offset], []byte{'u', 's', 't', 'a', 'r'})
}

func Rar(p []byte) bool {
	const min = 7
	if len(p) < min {
		return false
	}
	return bytes.Equal(p[:min], []byte{'R', 'a', 'r', 0x21, 0x1a, 0x7, 0x0})
}

func Rarv5(p []byte) bool {
	const min = 8
	if len(p) < min {
		return false
	}
	return bytes.Equal(p[:min], []byte{'R', 'a', 'r', 0x21, 0x1a, 0x7, 0x1, 0x0})
}

func Gzip(p []byte) bool {
	const min = 3
	if len(p) < min {
		return false
	}
	return bytes.Equal(p[:min], []byte{0x1f, 0x8b, 0x8})
}

func Bzip2(p []byte) bool {
	const min = 3
	if len(p) < min {
		return false
	}
	return bytes.Equal(p[:min], []byte{'B', 'Z', 'h'})
}

func X7z(p []byte) bool {
	const min = 6
	if len(p) < min {
		return false
	}
	return bytes.Equal(p[:min], []byte{'7', 'z', 0xbc, 0xaf, 0x27, 0x1c})
}

func XZ(p []byte) bool {
	const min = 6
	if len(p) < min {
		return false
	}
	return bytes.Equal(p[:min], []byte{0xfd, '7', 'z', 'X', 'Z', 0x0})
}

func Cab(p []byte) bool {
	const min = 4
	if len(p) < min {
		return false
	}
	return bytes.Equal(p[:min], []byte{'M', 'S', 'C', 'F'})
}

func ZStd(p []byte) bool {
	const min = 4
	if len(p) < min {
		return false
	}
	return bytes.Equal(p[:min], []byte{0x28, 0xb5, 0x2f, 0xfd})
}

func Arc(p []byte) bool {
	const min = 4
	if len(p) < min {
		return false
	}
	return bytes.Equal(p[:min], []byte{'A', 'r', 'C', 0x1})
}

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

func LzhLha(p []byte) bool {
	const min = 5
	if len(p) < min {
		return false
	}
	const offset = 2
	return bytes.Equal(p[offset:3], []byte{'-', 'l', 'h'})
}

func Zoo(p []byte) bool {
	const min = 4
	if len(p) < min {
		return false
	}
	return bytes.Equal(p[:min], []byte{'Z', 'O', 'O', 0x20})
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

func MSExe(p []byte) bool {
	const min = 2
	if len(p) < min {
		return false
	}
	return p[0] == 'M' && p[1] == 'Z' || p[0] == 'Z' && p[1] == 'M'
}

// DosKWAJ returns true if the reader begins with the KWAJ compression signature,
// found in some DOS executables.
func DosKWAJ(p []byte) bool {
	const min = 8
	if len(p) < min {
		return false
	}
	return bytes.Equal(p[:min], []byte{'K', 'W', 'A', 'J', 0x88, 0xf0, 0x27, 0xd1})
}

func DosSZDD(p []byte) bool {
	const min = 8
	if len(p) < min {
		return false
	}
	return bytes.Equal(p[:min], []byte{'S', 'Z', 'D', 'D', 0x88, 0xf0, 0x27, 0x33})
}

func MSComp(p []byte) bool {
	const min = 8
	if len(p) < min {
		return false
	}
	return bytes.Equal(p[:min], []byte{0xd0, 0xcf, 0x11, 0xe0, 0xa1, 0xb1, 0x1a, 0xe1})
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

func Nri(p []byte) bool {
	const min = 8
	if len(p) < min {
		return false
	}
	return bytes.Equal(p[:min], []byte{0x0e, 'N', 'e', 'r', 'o', 'I', 'S', 'O'})
}

func Daa(p []byte) bool {
	const min = 8
	if len(p) < min {
		return false
	}
	return bytes.Equal(p[:min], []byte{'D', 'A', 'A', 0x0, 0x0, 0x0, 0x0, 0x0})
}

func Mdf(p []byte) bool {
	const min = 16
	if len(p) < min {
		return false
	}
	return bytes.Equal(p[:min],
		[]byte{0x0, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
			0xff, 0xff, 0xff, 0x0, 0x0, 0x2, 0x0, 0x1})
}

func Jar(p []byte) bool {
	const min = 10
	if len(p) < min {
		return false
	}
	return bytes.Equal(p[:min], []byte{0x50, 0x4b, 0x3, 0x4, 0x14, 0x0, 0x8, 0x0, 0x8, 0x0})
}

func Pdf(p []byte) bool {
	const min = 4
	if len(p) < min {
		return false
	}
	if !bytes.Equal(p[:min], []byte{'%', 'P', 'D', 'F'}) {
		return false
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

func Rtf(p []byte) bool {
	const min = 5
	if len(p) < min {
		return false
	}
	if !bytes.Equal(p[:min], []byte{'{', 0x5c, 'r', 't', 'f'}) {
		return false
	}
	return bytes.HasSuffix(p, []byte{'}'})
}

func Utf8(p []byte) bool {
	const min = 3
	if len(p) < min {
		return false
	}
	return bytes.Equal(p[:min], []byte{0xef, 0xbb, 0xbf})
}

func Utf16(p []byte) bool {
	const min = 2
	if len(p) < min {
		return false
	}
	return bytes.Equal(p[:min], []byte{0xff, 0xfe}) || bytes.Equal(p[:min], []byte{0xfe, 0xff})
}

func Utf32(p []byte) bool {
	const min = 4
	if len(p) < min {
		return false
	}
	return bytes.Equal(p[:min], []byte{0xff, 0xfe, 0x0, 0x0}) || bytes.Equal(p[:min], []byte{0x0, 0x0, 0xfe, 0xff})
}

func Txt(p []byte) bool {
	return !slices.ContainsFunc(p, NotPlainText)
}

func TxtLatin1(p []byte) bool {
	return !slices.ContainsFunc(p, NonISO88951)
}

func TxtWindows(p []byte) bool {
	return !slices.ContainsFunc(p, NonWindows1252)
}

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

func Ascii(p []byte) bool {
	return !slices.ContainsFunc(p, NotAscii)
}

func NotAscii(b byte) bool {
	const (
		nul = 0x0
		tab = byte('\t')
		nl  = byte('\n')
		vt  = byte('\v')
		ff  = byte('\f')
		cr  = byte('\r')
		esc = 0x1b
	)
	return (b < 0x20 || b > 0x7f) &&
		b != nul && b != tab && b != nl &&
		b != vt && b != ff && b != cr && b != esc
}

func NonISO88951(b byte) bool {
	if !NotAscii(b) {
		return false
	}
	ExtendedAscii := b >= 0xa0 && b <= 0xff
	return !ExtendedAscii
}

func NonWindows1252(b byte) bool {
	if !NonISO88951(b) {
		return false
	}
	ExtraTypography := b != 0x81 && b != 0x8d && b != 0x8f && b != 0x90 && b != 0x9d
	return !(b >= 0x80 && b <= 0xff && ExtraTypography)
}

func NotPlainText(b byte) bool {
	if !NotAscii(b) {
		return false
	}
	ExtendedAscii := b >= 0x80 && b <= 0xff
	return !ExtendedAscii
}
