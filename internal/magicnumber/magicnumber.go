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
	"errors"
	"fmt"
	"io"
	"math/bits"
	"path/filepath"
	"slices"
	"strings"
)

// Signature represents a file type signature.
type Signature int

// Signature aliases for common file type signatures.
const (
	IFF  = ElectronicArtsIFF
	JPG  = JPEGFileInterchangeFormat
	PNG  = PortableNetworkGraphics
	GIF  = GraphicsInterchangeFormat
	WebP = GoogleWebP
	TIFF = TaggedImageFileFormat
	BMP  = BMPFileFormat
	PCX  = PersonalComputereXchange
	AVI  = MicrosoftAudioVideoInterleave
)

const (
	ZeroByte Signature = iota - 2
	Unknown
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
	RIPscrip
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
	MusicModule
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
	switch {
	case sign <= ZeroByte:
		return "0-byte data"
	case sign == Unknown:
		return "binary data"
	case sign > PlainText:
		return "error"
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
		"RIPscrip",
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
		"Tracker music",
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
	switch {
	case sign <= ZeroByte:
		return "Zero-byte data"
	case sign == Unknown:
		return "Binary data"
	case sign > PlainText:
		return "Error"
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
		"RIPscrip vector graphic",
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
		"Tracker music",
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
		RIPscrip:                          []string{".rip"},
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
		MusicModule:                       []string{".mod", ".s3m", ".xm", ".it", ".mtm", ".mo3"},
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
		RIPscrip:                          Ripscrip,
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
		MusicModule:                       Mod,
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
		RIPscrip,
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

// Find2K reads the first 2048 bytes from the reader and returns the file type signature.
// This is a less accurate method than Find but should be faster.
func Find2K(r io.Reader) (Signature, error) {
	buf := make([]byte, MusicTrackerSize)
	_, err := io.ReadFull(r, buf)
	if err != nil && !errors.Is(err, io.ErrUnexpectedEOF) {
		return Unknown, fmt.Errorf("magic number find first %d bytes: %w", MusicTrackerSize, err)
	}
	return FindBytes2K(buf), nil
}

// FindBytes returns the file type signature from the byte slice.
func FindBytes(p []byte) Signature {
	if len(p) == 0 {
		return ZeroByte
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

// FindBytes2K returns the file type signature and skips the magic number checks
// that require the entire file to be read.
func FindBytes2K(p []byte) Signature {
	if len(p) == 0 {
		return ZeroByte
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

// Flags returns the indexes of the set bits in the byte.
func Flags(x uint8) []int {
	ones := make([]int, bits.OnesCount8(x))
	i := 0
	for x != 0 {
		ones[i] = bits.TrailingZeros8(x)
		x &= x - 1
		i++
	}
	return ones
}

// TODO Seek end?

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
