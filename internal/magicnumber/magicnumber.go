// Package magicnumber contains the magic number matchers for identifying file types that
// are expected to be handled by the Defacto2 server application. Magic numbers are not
// always accurate and should be used as hints combined with other checks such as
// file extension matching.
//
// Usually, the magic number is the first few bytes of a file that uniquely identify the file type.
// But a number of document formats also check the final few bytes of a file.
//
// At a later stage, the magic number matchers will be used to extract metadata from files
// and support for module tracking music files will be added.
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
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"slices"
	"strings"
)

// Signature represents a file type signature.
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
	PKWAREZipShrink
	PKWAREZipReduce
	PKWAREZipImplode
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
	X7zCompressArchive
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
	CDISO9660
	CDNero
	CDPowerISO
	CDAlcohol120
	JavaARchive
	WindowsHelpFile
	PortableDocumentFormat
	RichTextFormat
	UTF8Text
	UTF16Text
	UTF32Text
	ANSIEscapeText
	PlainText
)

func (sign Signature) String() string { //nolint:funlen
	if sign < 0 || sign > PlainText {
		return "binary data"
	}
	return [...]string{
		"IFF image",
		"AV1 image",
		"JPEG image",
		"JPEG 2000 image",
		"PNG image",
		"GIF image",
		"WebP image",
		"TIFF image",
		"BMP image",
		"PCX image",
		"BMP image",
		"Microsoft icon",
		"MPEG-4 video",
		"QuickTime video",
		"QuickTime video",
		"AVI video",
		"Windows Media video",
		"MPEG-4 video",
		"Flash video",
		"RealPlayer video",
		"MIDI audio",
		"MP3 audio",
		"Ogg audio",
		"FLAC audio",
		"Wave audio",
		"pkzip shrunk archive",
		"pkzip reduced archive",
		"pkzip imploded archive",
		"zip64 archive",
		"zip archive",
		"multivolume zip",
		"pklite compressed",
		"self-extracting zip",
		"Tape archive",
		"RAR archive",
		"RAR v5+ archive",
		"Gzip archive",
		"Bzip2 archive",
		"7z archive",
		"XZ archive",
		"ZST archive",
		"FreeARC",
		"ARC by SEA",
		"LHA by Yoshi",
		"Zoo archive",
		"ARJ archive",
		"Microsoft cabinet",
		"MS-DOS KWAJ",
		"MS-DOS SZDD",
		"MS-DOS executable",
		"Microsoft compound fFile",
		"CD, ISO 9660",
		"CD, Nero",
		"CD, PowerISO",
		"CD, Alcohol 120",
		"Java archive",
		"Windows help",
		"PDF document",
		"rich text",
		"UTF-8 text",
		"UTF-16 text",
		"UTF-32 text",
		"ANSI text",
		"plain text",
	}[sign]
}

func (sign Signature) Title() string { //nolint:funlen
	if sign < 0 || sign > PlainText {
		return "Binary data"
	}
	return [...]string{
		"Electronic Arts IFF",
		"AV1 Image File",
		"JPEG File Interchange Format",
		"JPEG 2000",
		"Portable Network Graphics",
		"Graphics Interchange Format",
		"Google WebP",
		"Tagged Image File Format",
		"BMP File Format",
		"Personal Computer eXchange",
		"Interleaved Bitmap",
		"Microsoft Icon",
		"MPEG-4 video",
		"QuickTime Movie",
		"QuickTime M4V",
		"Microsoft Audio Video Interleave",
		"Microsoft Windows Media",
		"MPEG-4 video",
		"Flash Video",
		"RealPlayer",
		"Musical Instrument Digital Interface",
		"MPEG-1 Audio Layer 3",
		"Ogg Vorbis Codec",
		"Free Lossless Audio Codec",
		"Wave Audio for Windows",
		"Shrunked pkzip archive",
		"Reduced pkzip archive",
		"Imploded pkzip archive",
		"PKWARE zip64 archive",
		"Zip archive",
		"Zip multi-Volume archive",
		"PKLITE compressed executable",
		"PKSFX self-extracting archive",
		"Tape Archive",
		"Roshal Archive",
		"Roshal Archive v5",
		"Gzip compress archive",
		"Bzip2 compress archive",
		"7z compress archive",
		"XZ compress archive",
		"ZStandard archive",
		"FreeArc",
		"Archive by SEA",
		"Yoshi LHA",
		"Zoo Archive",
		"Archive by Robert Jung",
		"Microsoft Cabinet",
		"Microsoft DOS KWAJ",
		"Microsoft DOS SZDD",
		"Microsoft executable",
		"Microsoft compound file",
		"CD ISO 9660",
		"CD Nero",
		"CD PowerISO",
		"CD Alcohol 120",
		"Java archive",
		"Windows Help File",
		"Portable Document Format",
		"Rich Text Format",
		"UTF-8 text",
		"UTF-16 text",
		"UTF-32 text",
		"ANSI escaped text",
		"Plain text",
	}[sign]
}

// Extension is a map of file type signatures to file extensions.
type Extension map[Signature][]string

// Ext returns a map of file type signatures to common file extensions.
func Ext() Extension { //nolint:funlen
	return Extension{
		ElectronicArtsIFF:                 []string{".iff"},
		AV1ImageFile:                      []string{".avif"},
		JPEGFileInterchangeFormat:         []string{".jpg", ".jpeg"},
		JPEG2000:                          []string{".jp2", ".j2k", ".jpf", ".jpx", ".jpm", ".mj2"},
		PortableNetworkGraphics:           []string{".png"},
		GraphicsInterchangeFormat:         []string{".gif"},
		GoogleWebP:                        []string{".webp"},
		TaggedImageFileFormat:             []string{".tif", ".tiff"},
		BMPFileFormat:                     []string{".bmp"},
		PersonalComputereXchange:          []string{".pcx"},
		InterleavedBitmap:                 []string{".ilbm"},
		MicrosoftIcon:                     []string{".ico"},
		MPEG4:                             []string{".mp4"},
		QuickTimeMovie:                    []string{".mov"},
		QuickTimeM4V:                      []string{".m4v"},
		MicrosoftAudioVideoInterleave:     []string{".avi"},
		MicrosoftWindowsMedia:             []string{".wmv"},
		MPEG:                              []string{".mpg", ".mpeg"},
		FlashVideo:                        []string{".flv"},
		RealPlayer:                        []string{".rv", ".rm", ".rmvb"},
		MusicalInstrumentDigitalInterface: []string{".mid", ".midi"},
		MPEG1AudioLayer3:                  []string{".mp3"},
		OggVorbisCodec:                    []string{".ogg"},
		FreeLosslessAudioCodec:            []string{".flac"},
		WaveAudioForWindows:               []string{".wav"},
		PKWAREZipShrink:                   []string{".zip"},
		PKWAREZipReduce:                   []string{".zip"},
		PKWAREZipImplode:                  []string{".zip"},
		PKWAREZip64:                       []string{".zip"},
		PKWAREZip:                         []string{".zip"},
		PKWAREMultiVolume:                 []string{".zip"},
		PKLITE:                            []string{".zip"},
		PKSFX:                             []string{".zip"},
		TapeARchive:                       []string{".tar"},
		RoshalARchive:                     []string{".rar"},
		RoshalARchivev5:                   []string{".rar"},
		GzipCompressArchive:               []string{".gz"},
		Bzip2CompressArchive:              []string{".bz2"},
		X7zCompressArchive:                []string{".7z"},
		XZCompressArchive:                 []string{".xz"},
		ZStandardArchive:                  []string{".zst"},
		FreeArc:                           []string{".arc"},
		ARChiveSEA:                        []string{".arc"},
		YoshiLHA:                          []string{".lzh", ".lha"},
		ZooArchive:                        []string{".zoo"},
		ArchiveRobertJung:                 []string{".arj"},
		MicrosoftCABinet:                  []string{".cab"},
		MicrosoftDOSKWAJ:                  []string{".com"},
		MicrosoftDOSSZDD:                  []string{".exe"},
		MicrosoftExecutable:               []string{".exe"},
		MicrosoftCompoundFile:             []string{".exe"},
		CDISO9660:                         []string{".iso"},
		CDNero:                            []string{".nri"},
		CDPowerISO:                        []string{".daa"},
		CDAlcohol120:                      []string{".mdf"},
		JavaARchive:                       []string{".jar"},
		WindowsHelpFile:                   []string{".hlp"},
		PortableDocumentFormat:            []string{".pdf"},
		RichTextFormat:                    []string{".rtf"},
		UTF8Text:                          []string{".txt"},
		UTF16Text:                         []string{".txt"},
		UTF32Text:                         []string{".txt"},
		ANSIEscapeText:                    []string{".ans"},
		PlainText:                         []string{".txt"},
	}
}

// Matcher is a function that matches a byte slice to a file type.
type Matcher func([]byte) bool

// Finder is a map of file type signatures to matchers.
type Finder map[Signature]Matcher

// New returns a new Finder with all the matchers.
//
// ANSIEscapeText and PlainText are not included as they need to be
// checked separately and in a specific order.
func New() Finder { //nolint:funlen
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
		PKWAREZipShrink:                   PkShrink,
		PKWAREZipReduce:                   PkReduce,
		PKWAREZipImplode:                  PkImplode,
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
		X7zCompressArchive:                X7z,
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
		CDISO9660:                         ISO,
		CDNero:                            Nri,
		CDPowerISO:                        Daa,
		CDAlcohol120:                      Mdf,
		JavaARchive:                       Jar,
		WindowsHelpFile:                   Hlp,
		PortableDocumentFormat:            Pdf,
		RichTextFormat:                    Rtf,
		UTF8Text:                          Utf8,
		UTF16Text:                         Utf16,
		UTF32Text:                         Utf32,
	}
}

// Archives returns all the archive file type signatures.
func Archives() []Signature {
	return []Signature{
		PKWAREZipShrink,
		PKWAREZipReduce,
		PKWAREZipImplode,
		PKWAREZip64,
		PKWAREZip,
		PKWAREMultiVolume,
		PKLITE,
		PKSFX,
		TapeARchive,
		RoshalARchive,
		RoshalARchivev5,
		GzipCompressArchive,
		Bzip2CompressArchive,
		X7zCompressArchive,
		XZCompressArchive,
		ZStandardArchive,
		FreeArc,
		ARChiveSEA,
		YoshiLHA,
		ZooArchive,
		ArchiveRobertJung,
		MicrosoftCABinet,
	}
}

// Archives returns all the archive file type signatures that were
// commonly used in the BBS online era of the 1980s and early 1990s.
// Eventually these were replaced by the universal ZIP format using
// the Deflate and Store compression methods.
func ArchivesBBS() []Signature {
	return []Signature{
		PKWAREZipShrink,
		PKWAREZipReduce,
		PKWAREZipImplode,
		ARChiveSEA,
		YoshiLHA,
		ZooArchive,
		ArchiveRobertJung,
	}
}

// DiscImages returns all the CD disk image file type signatures.
func DiscImages() []Signature {
	return []Signature{
		CDISO9660,
		CDNero,
		CDPowerISO,
		CDAlcohol120,
	}
}

func Documents() []Signature {
	return []Signature{
		WindowsHelpFile,
		PortableDocumentFormat,
		RichTextFormat,
		UTF8Text,
		UTF16Text,
		UTF32Text,
	}
}

// Images returns all the image file type signatures.
func Images() []Signature {
	return []Signature{
		AV1ImageFile,
		JPEGFileInterchangeFormat,
		JPEG2000,
		PortableNetworkGraphics,
		GraphicsInterchangeFormat,
		GoogleWebP,
		TaggedImageFileFormat,
		BMPFileFormat,
		PersonalComputereXchange,
		InterleavedBitmap,
		MicrosoftIcon,
	}
}

// Programs returns all the program file type signatures for
// Microsoft operating systems, DOS and Windows.
func Programs() []Signature {
	return []Signature{
		MicrosoftExecutable,
		MicrosoftDOSKWAJ,
		MicrosoftDOSSZDD,
		MicrosoftCompoundFile,
	}
}

// Texts returns all the text file type signatures.
func Texts() []Signature {
	return []Signature{
		UTF8Text,
		UTF16Text,
		UTF32Text,
		ANSIEscapeText,
		PlainText,
	}
}

// Videos returns all the video file type signatures.
func Videos() []Signature {
	return []Signature{
		MPEG4,
		QuickTimeMovie,
		QuickTimeM4V,
		MicrosoftAudioVideoInterleave,
		MicrosoftWindowsMedia,
		MPEG,
		FlashVideo,
		RealPlayer,
	}
}

// MatchExt determines if the reader matches the file type signature expected
// from the extension of the filename. It returns true if the file type matches and
// a found signature is always returned.
//
// A PNG encoded image using the filename TEST.PNG will return true
// and the PortableNetworkGraphics signature.
// A PNG encoded image using the filename TEST.JPG will return false
// and the PortableNetworkGraphics signature.
func MatchExt(filename string, r io.Reader) (bool, Signature, error) {
	ext := strings.ToLower(filepath.Ext(filename))
	buf, err := io.ReadAll(r)
	if err != nil {
		return false, Unknown, fmt.Errorf("magic number match extension: %w", err)
	}
	finds := New()
	for extSign, exts := range Ext() {
		if slices.Contains(exts, ext) {
			for findSign, matcher := range finds {
				if matcher(buf) {
					if findSign == extSign {
						return true, findSign, nil
					}
					return false, findSign, nil
				}
			}
		}
	}
	sig := FindBytes(buf)
	return false, sig, nil
}

// Find reads all the bytes from the reader and returns the file type signature.
// Generally, magic numbers are the first few bytes of a file that uniquely identify the file type.
// But a number of document formats also check the body content or the final few bytes of a file.
func Find(r io.Reader) (Signature, error) {
	buf, err := io.ReadAll(r)
	if err != nil {
		return Unknown, fmt.Errorf("magic number find: %w", err)
	}
	return FindBytes(buf), nil
}

// Find512B reads the first 512 bytes from the reader and returns the file type signature.
// This is a less accurate method than Find but should be faster.
func Find512B(r io.Reader) (Signature, error) {
	buf := make([]byte, 512)
	_, err := io.ReadFull(r, buf)
	if err != nil && !errors.Is(err, io.ErrUnexpectedEOF) {
		return Unknown, fmt.Errorf("magic number find first 512 bytes: %w", err)
	}
	return FindBytes512B(buf), nil
}

// Archive reads all the bytes from the reader and returns the file type signature if
// the file is a known archive of files or Unknown if the file is not an archive.
func Archive(r io.Reader) (Signature, error) {
	buf, err := io.ReadAll(r)
	if err != nil {
		return Unknown, fmt.Errorf("magic number archive: %w", err)
	}
	archives := Archives()
	find := New()
	for _, archive := range archives {
		if finder, exists := find[archive]; exists {
			if finder(buf) {
				return archive, nil
			}
		}
	}
	return Unknown, nil
}

// DiscImage reads all the bytes from the reader and returns the file type signature if
// the file is a known CD disk image or Unknown if the file is not a disk image.
func DiscImage(r io.Reader) (Signature, error) {
	buf, err := io.ReadAll(r)
	if err != nil {
		return Unknown, fmt.Errorf("magic number disc image: %w", err)
	}
	discs := DiscImages()
	find := New()
	for _, disc := range discs {
		if finder, exists := find[disc]; exists {
			if finder(buf) {
				return disc, nil
			}
		}
	}
	return Unknown, nil
}

// Document reads all the bytes from the reader and returns the file type signature if
// the file is a known document or Unknown if the file is not a document.
func Document(r io.Reader) (Signature, error) {
	buf, err := io.ReadAll(r)
	if err != nil {
		return Unknown, fmt.Errorf("magic number document: %w", err)
	}
	docs := Documents()
	find := New()
	for _, doc := range docs {
		if finder, exists := find[doc]; exists {
			if finder(buf) {
				return doc, nil
			}
		}
	}
	switch {
	case Ansi(buf):
		return ANSIEscapeText, nil
	case Txt(buf):
		return PlainText, nil
	default:
		return Unknown, nil
	}
}

// Image reads all the bytes from the reader and returns the file type signature if
// the file is a known image or Unknown if the file is not an image.
func Image(r io.Reader) (Signature, error) {
	buf, err := io.ReadAll(r)
	if err != nil {
		return Unknown, fmt.Errorf("magic number image: %w", err)
	}
	imgs := Images()
	find := New()
	for _, image := range imgs {
		if finder, exists := find[image]; exists {
			if finder(buf) {
				return image, nil
			}
		}
	}
	return Unknown, nil
}

// Program reads all the bytes from the reader and returns the file type signature if
// the file is a known DOS or Windows program or Unknown if the file is not a program.
func Program(r io.Reader) (Signature, error) {
	buf, err := io.ReadAll(r)
	if err != nil {
		return Unknown, fmt.Errorf("magic number program: %w", err)
	}
	progs := Programs()
	find := New()
	for _, prog := range progs {
		if finder, exists := find[prog]; exists {
			if finder(buf) {
				return prog, nil
			}
		}
	}
	return Unknown, nil
}

// Text reads the first 512 bytes from the reader and returns the file type signature if
// the file is a known plain text file or Unknown if the file is not a text file.
func Text(r io.Reader) (Signature, error) {
	buf := make([]byte, 512)
	_, err := io.ReadFull(r, buf)
	if err != nil && !errors.Is(err, io.ErrUnexpectedEOF) && !errors.Is(err, io.EOF) {
		return Unknown, fmt.Errorf("magic number text: %w", err)
	}
	txts := Texts()
	find := New()
	for _, txt := range txts {
		if finder, exists := find[txt]; exists {
			if finder(buf) {
				return txt, nil
			}
		}
	}
	switch {
	case Ansi(buf):
		return ANSIEscapeText, nil
	case Txt(buf):
		return PlainText, nil
	default:
		return Unknown, nil
	}
}

// Video reads all the bytes from the reader and returns the file type signature if
// the file is a known video or Unknown if the file is not a video.
func Video(r io.Reader) (Signature, error) {
	buf, err := io.ReadAll(r)
	if err != nil {
		return Unknown, fmt.Errorf("magic number video: %w", err)
	}
	vids := Videos()
	find := New()
	for _, video := range vids {
		if finder, exists := find[video]; exists {
			if finder(buf) {
				return video, nil
			}
		}
	}
	return Unknown, nil
}

// FindBytes returns the file type signature from the byte slice.
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
	switch {
	case Ansi(p):
		return ANSIEscapeText
	case Txt(p):
		return PlainText
	default:
		return Unknown
	}
}

// FindBytes512B returns the file type signature and skips the magic number checks
// that require the entire file to be read.
func FindBytes512B(p []byte) Signature {
	if p == nil {
		return Unknown
	}
	find := New()
	for sig, matcher := range find {
		switch sig {
		case RichTextFormat:
			matcher = RtfNoSuffix
		case PortableDocumentFormat:
			matcher = PdfNoSuffix
		case JPEGFileInterchangeFormat:
			matcher = JpegNoSuffix
		}
		if matcher(p) {
			return sig
		}
	}
	switch {
	case Ansi(p):
		return ANSIEscapeText
	case Txt(p):
		return PlainText
	default:
		return Unknown
	}
}

// Iff matches the Interchange File Format image in the byte slice.
// This is a generic wrapper format originally created by Electronic Arts
// for storing data in chunks.
func Iff(p []byte) bool {
	const min = 4
	if len(p) < min {
		return false
	}
	return bytes.Equal(p[:4], []byte{'C', 'A', 'T', 0x20})
}

// Avif matches the AV1 Image File image format in the byte slice, also known as AVIF.
// This is a new image format based on the AV1 video codec from the Alliance for Open Media.
// But the detection method is not accurate and should be used as a hint.
func Avif(p []byte) bool {
	const min, offset = 8, 4
	if len(p) < min+offset {
		return false
	}
	// Gary Kessler's File Signatures suggests the AVIF image format is 0x0A 0x00 0x00
	// but this maybe out dated and definitely causes false positives.
	// According to the AV1 Image File Format specification there is no magic number.
	// https://aomediacodec.github.io/av1-avif/v1.0.0.html
	//
	// As a workaround, we detect the AVIF image format by checking for the 'ftypavif' string.
	// 'ftyp' matches the HEIF container and 'avif' is the brand.
	return bytes.Equal(p[offset:min+offset], []byte{'f', 't', 'y', 'p', 'a', 'v', 'i', 'f'})
}

// Jpeg matches the JPEG File Interchange Format v1 image in the byte slice.
func Jpeg(p []byte) bool {
	return jpeg(p, true)
}

// JpegNoSuffix matches the JPEG File Interchange Format v1 image in the byte slice.
// This is a less accurate method than Jpeg as it does not check the final bytes.
func JpegNoSuffix(p []byte) bool {
	return jpeg(p, false)
}

func jpeg(p []byte, suffix bool) bool {
	const min = 11
	if len(p) < min {
		return false
	}
	if !bytes.Equal(p[:3], []byte{0xff, 0xd8, 0xff}) {
		return false
	}
	if p[3] != 0xe0 && p[3] != 0xe1 {
		return false
	}
	if !bytes.Equal(p[6:11], []byte{'J', 'F', 'I', 'F', 0x0}) &&
		!bytes.Equal(p[6:11], []byte{'E', 'x', 'i', 'f', 0x0}) {
		return false
	}
	if !suffix {
		return true
	}
	return bytes.HasSuffix(p, []byte{0xff, 0xd9})
}

// Jpeg2000 matches the JPEG 2000 image format in the byte slice.
func Jpeg2000(p []byte) bool {
	const min = 10
	if len(p) < min {
		return false
	}
	return bytes.Equal(p[:min], []byte{0x0, 0x0, 0x0, 0xc, 0x6a, 0x50, 0x20, 0x20, 0xd, 0xa})
}

// Png matches the Portable Network Graphics image format in the byte slice.
func Png(p []byte) bool {
	const min = 8
	if len(p) < min {
		return false
	}
	return bytes.Equal(p[:min], []byte{0x89, 0x50, 0x4E, 0x47, 0x0d, 0x0a, 0x1a, 0x0a})
}

// Gif matches the image Graphics Interchange Format in the byte slice.
// There are two versions of the GIF format, GIF87a and GIF89a.
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

// Webp matches the Google WebP image format in the byte slice.
func Webp(p []byte) bool {
	const min = 12
	if len(p) < min {
		return false
	}
	return bytes.Equal(p[:4], []byte{'R', 'I', 'F', 'F'}) &&
		bytes.Equal(p[8:12], []byte{'W', 'E', 'B', 'P'})
}

// Tiff matches the Tagged Image File Format in the byte slice.
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

// Bmp matches the BMP image format in the byte slice.
func Bmp(p []byte) bool {
	const min = 2
	if len(p) < min {
		return false
	}
	return bytes.Equal(p[:min], []byte{'B', 'M'})
}

// Pcx matches the Personal Computer eXchange image format in the byte slice.
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

// Ico matches the Microsoft Icon image format in the byte slice.
func Ico(p []byte) bool {
	const min = 4
	if len(p) < min {
		return false
	}
	return bytes.Equal(p[:min], []byte{0x0, 0x0, 0x1, 0x0})
}

// Ilbm matches the InterLeaved Bitmap image format in the byte slice.
// Created by Electronic Arts it conforms to the IFF standard.
func Ilbm(p []byte) bool {
	const min = 12
	if len(p) < min {
		return false
	}
	return bytes.Equal(p[:4], []byte{'F', 'O', 'R', 'M'}) &&
		bytes.Equal(p[8:12], []byte{'I', 'L', 'B', 'M'})
}

// QTMov matches the QuickTime Movie video format in the byte slice.
func QTMov(p []byte) bool {
	const min = 8
	if len(p) < min {
		return false
	}
	const offset = 4
	return bytes.Equal(p[offset:8], []byte{'m', 'o', 'o', 'v'}) ||
		bytes.Equal(p[offset:10], []byte{'f', 't', 'y', 'p', 'q', 't'})
}

// Mp4 matches the MPEG-4 video format in the byte slice.
func Mp4(p []byte) bool {
	const min = 8
	if len(p) < min {
		return false
	}
	return bytes.Equal(p[:min], []byte{'f', 't', 'y', 'p', 'M', 'S', 'N', 'V'})
}

// M4v matches the QuickTime M4V video format in the byte slice.
func M4v(p []byte) bool {
	const min = 8
	if len(p) < min {
		return false
	}
	return bytes.Equal(p[:min], []byte{'f', 't', 'y', 'p', 'm', 'p', '4', '2'})
}

// Avi matches the Microsoft Audio Video Interleave video format in the byte slice.
func Avi(p []byte) bool {
	const min = 16
	if len(p) < min {
		return false
	}
	return bytes.Equal(p[:4], []byte{'R', 'I', 'F', 'F'}) &&
		bytes.Equal(p[8:16], []byte{'A', 'V', 'I', 0x20, 'L', 'I', 'S', 'T'})
}

// Wmv matches the Microsoft Windows Media video format in the byte slice.
func Wmv(p []byte) bool {
	const min = 16
	if len(p) < min {
		return false
	}
	return bytes.Equal(p[:min],
		[]byte{
			0x30, 0x26, 0xb2, 0x75, 0x8e, 0x66, 0xcf, 0x11,
			0xa6, 0xd9, 0x0, 0xaa, 0x0, 0x62, 0xce, 0x6c,
		})
}

// Mpeg matches the MPEG video format in the byte slice.
func Mpeg(p []byte) bool {
	const min = 4
	if len(p) < min {
		return false
	}
	return bytes.Equal(p[:3], []byte{0x0, 0x0, 0x1}) && p[4] >= 0xba && p[4] <= 0xbf
}

// Flv matches the Shockwave Flash Video format in the byte slice.
func Flv(p []byte) bool {
	const min = 4
	if len(p) < min {
		return false
	}
	return bytes.Equal(p[:min], []byte{'F', 'L', 'V', 0x1})
}

// Ivr matches the RealPlayer video format in the byte slice.
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

// Mp3 matches the MPEG-1 Audio Layer 3 audio format in the byte slice.
func Mp3(p []byte) bool {
	const min = 3
	if len(p) < min {
		return false
	}
	return bytes.Equal(p[:min], []byte{'I', 'D', '3'})
}

// Ogg matches the Ogg Vorbis audio format in the byte slice.
func Ogg(p []byte) bool {
	const min = 12
	if len(p) < min {
		return false
	}
	return bytes.Equal(p[:min], []byte{
		'O', 'g', 'g', 'S', 0x0, 0x2, 0x0, 0x0,
		0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	})
}

// Flac matches the Free Lossless Audio Codec audio format in the byte slice.
func Flac(p []byte) bool {
	const min = 8
	if len(p) < min {
		return false
	}
	return bytes.Equal(p[:min], []byte{'f', 'L', 'a', 'C', 0x0, 0x0, 0x0, 0x2})
}

// Wave matches the IBM / Microsoft Waveform audio format in the byte slice.
func Wave(p []byte) bool {
	const min = 12
	if len(p) < min {
		return false
	}
	return bytes.Equal(p[:4], []byte{'R', 'I', 'F', 'F'}) &&
		bytes.Equal(p[8:16], []byte{'W', 'A', 'V', 'E', 'f', 'm', 't', 0x20})
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

// PkzipMulti matches the PKWARE Multi-Volume Zip archive format in the byte slice.
func PkzipMulti(p []byte) bool {
	const min = 4
	if len(p) < min {
		return false
	}
	return bytes.Equal(p[:min], []byte{'P', 'K', 0x7, 0x8})
}

// Pklite matches the PKLITE archive format in the byte slice which is a
// compressed executable format for DOS and 16-bit Windows.
func Pklite(p []byte) bool {
	const min = 6
	const offset = 30
	if len(p) < min+offset {
		return false
	}
	return bytes.Equal(p[offset:min+offset], []byte{0x50, 0x4b, 0x4c, 0x49, 0x54, 0x45})
}

// Pksfx matches the PKSFX archive format in the byte slice which is a
// self-extracting archive format.
func Pksfx(p []byte) bool {
	const min = 5
	const offset = 526
	if len(p) < min+offset {
		return false
	}
	return bytes.Equal(p[offset:min+offset], []byte{0x50, 0x4b, 0x53, 0x70, 0x58})
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

// Gzip matches the Gzip Compress archive format in the byte slice.
func Gzip(p []byte) bool {
	const min = 3
	if len(p) < min {
		return false
	}
	return bytes.Equal(p[:min], []byte{0x1f, 0x8b, 0x8})
}

// Bzip2 matches the Bzip2 Compress archive format in the byte slice.
func Bzip2(p []byte) bool {
	const min = 3
	if len(p) < min {
		return false
	}
	return bytes.Equal(p[:min], []byte{'B', 'Z', 'h'})
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

// Cab matches the Microsoft CABinet archive format in the byte slice.
func Cab(p []byte) bool {
	const min = 4
	if len(p) < min {
		return false
	}
	return bytes.Equal(p[:min], []byte{'M', 'S', 'C', 'F'})
}

// ZStd matches the ZStandard archive format in the byte slice.
func ZStd(p []byte) bool {
	const min = 4
	if len(p) < min {
		return false
	}
	return bytes.Equal(p[:min], []byte{0x28, 0xb5, 0x2f, 0xfd})
}

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

// LzhLha matches the LHA and LZH compression formats in the byte slice.
func LzhLha(p []byte) bool {
	const min = 5
	if len(p) < min {
		return false
	}
	const offset = 2
	return bytes.Equal(p[offset:3+offset], []byte{'-', 'l', 'h'})
}

// Zoo matches the Zoo compression format in the byte slice.
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

// MSExe returns true if the reader begins with the Microsoft executable signature.
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

// DosSZDD returns true if the reader begins with the SZDD compression signature.
func DosSZDD(p []byte) bool {
	const min = 8
	if len(p) < min {
		return false
	}
	return bytes.Equal(p[:min], []byte{'S', 'Z', 'D', 'D', 0x88, 0xf0, 0x27, 0x33})
}

// MSComp returns true if the reader contains the Microsoft Compound File signature.
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

// Nri returns true if the reader contains the Nero CD image signature.
func Nri(p []byte) bool {
	const min = 8
	if len(p) < min {
		return false
	}
	return bytes.Equal(p[:min], []byte{0x0e, 'N', 'e', 'r', 'o', 'I', 'S', 'O'})
}

// Daa returns true if the reader contains the PowerISO DAA CD image signature.
func Daa(p []byte) bool {
	const min = 8
	if len(p) < min {
		return false
	}
	return bytes.Equal(p[:min], []byte{'D', 'A', 'A', 0x0, 0x0, 0x0, 0x0, 0x0})
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

// Jar returns true if the reader contains the Java ARchive signature.
func Jar(p []byte) bool {
	const min = 10
	if len(p) < min {
		return false
	}
	return bytes.Equal(p[:min], []byte{0x50, 0x4b, 0x3, 0x4, 0x14, 0x0, 0x8, 0x0, 0x8, 0x0})
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

// Txt returns true if the byte slice exclusively contains plain text ASCII characters,
// control characters or "extended ASCII characters".
func Txt(p []byte) bool {
	return !slices.ContainsFunc(p, NotPlainText)
}

// TxtLatin1 returns true if the byte slice exclusively contains plain text ISO/IEC-8895-1 characters,
// commonly known as the Latin-1 character set.
func TxtLatin1(p []byte) bool {
	return !slices.ContainsFunc(p, NonISO88951)
}

// TxtWindows returns true if the byte slice exclusively contains plain text Windows-1252 characters.
// This is an extension of the Latin-1 character set with additional typography characters and was
// the default character set for English in Microsoft Windows up to Windows 7?
func TxtWindows(p []byte) bool {
	return !slices.ContainsFunc(p, NonWindows1252)
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

// NonISO88951 returns true if the byte is not a printable ISO/IEC-8895-1 character.
func NonISO88951(b byte) bool {
	if !NotASCII(b) {
		return false
	}
	ExtendedASCII := b >= 0xa0 && b <= 0xff
	return !ExtendedASCII
}

// NonWindows1252 returns true if the byte is not a printable Windows-1252 character.
func NonWindows1252(b byte) bool {
	if !NonISO88951(b) {
		return false
	}
	ExtraTypography := b != 0x81 && b != 0x8d && b != 0x8f && b != 0x90 && b != 0x9d
	return !(b >= 0x80 && b <= 0xff && ExtraTypography)
}

// NotPlainText returns true if the byte is not a printable plain text character.
// This includes any printable ASCII character as well as any "extended ASCII".
func NotPlainText(b byte) bool {
	if !NotASCII(b) {
		return false
	}
	ExtendedASCII := b >= 0x80 && b <= 0xff
	return !ExtendedASCII
}
