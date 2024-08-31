package magicnumberr

import (
	"io"
)

// Archive reads all the bytes from the reader and returns the file type signature if
// the file is a known archive of files or Unknown if the file is not an archive.
func Archive(r io.ReaderAt) (Signature, error) {
	find := New()
	for _, archive := range Archives() {
		if finder, exists := find[archive]; exists {
			if finder(r) {
				return archive, nil
			}
		}
	}
	return Unknown, nil
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

// DiscImage reads all the bytes from the reader and returns the file type signature if
// the file is a known CD disk image or Unknown if the file is not a disk image.
func DiscImage(r io.ReaderAt) (Signature, error) {
	find := New()
	for _, img := range DiscImages() {
		if finder, exists := find[img]; exists {
			if finder(r) {
				return img, nil
			}
		}
	}
	return Unknown, nil
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

// Document reads all the bytes from the reader and returns the file type signature if
// the file is a known document or Unknown if the file is not a document.
func Document(r io.ReaderAt) (Signature, error) {
	find := New()
	for _, doc := range Documents() {
		if finder, exists := find[doc]; exists {
			if finder(r) {
				return doc, nil
			}
		}
	}
	switch {
	case Ansi(r):
		return ANSIEscapeText, nil
	case Txt(r):
		return PlainText, nil
	default:
		return Unknown, nil
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

// Image reads all the bytes from the reader and returns the file type signature if
// the file is a known image or Unknown if the file is not an image.
func Image(r io.ReaderAt) (Signature, error) {
	find := New()
	for _, img := range Images() {
		if finder, exists := find[img]; exists {
			if finder(r) {
				return img, nil
			}
		}
	}
	return Unknown, nil
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

// Program reads all the bytes from the reader and returns the file type signature if
// the file is a known DOS or Windows program or Unknown if the file is not a program.
func Program(r io.ReaderAt) (Signature, error) {
	find := New()
	for _, app := range Programs() {
		if finder, exists := find[app]; exists {
			if finder(r) {
				return app, nil
			}
		}
	}
	return Unknown, nil
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

// Text reads the first 512 bytes from the reader and returns the file type signature if
// the file is a known plain text file or Unknown if the file is not a text file.
func Text(r io.ReaderAt) (Signature, error) {
	find := New()
	for _, doc := range Texts() {
		if finder, exists := find[doc]; exists {
			if finder(r) {
				return doc, nil
			}
		}
	}
	switch {
	case Ansi(r):
		return ANSIEscapeText, nil
	case Txt(r):
		return PlainText, nil
	default:
		return Unknown, nil
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

// Video reads all the bytes from the reader and returns the file type signature if
// the file is a known video or Unknown if the file is not a video.
func Video(r io.ReaderAt) (Signature, error) {
	find := New()
	for _, vid := range Videos() {
		if finder, exists := find[vid]; exists {
			if finder(r) {
				return vid, nil
			}
		}
	}
	return Unknown, nil
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
