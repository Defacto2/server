package magicnumberr_test

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/Defacto2/server/internal/magicnumberr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	emptyFile  = "EMPTY"
	avifFile   = "TEST.avif"
	bmpFile    = "TEST.BMP"
	gifFile    = "TEST.GIF"
	gif2File   = "TEST2.gif"
	ilbmFile   = "TEST.IFF"
	jpegFile   = "TEST.JPEG"
	jpgFile    = "TEST.JPG"
	icoFile    = "favicon.ico"
	pcxFile    = "TEST.PCX"
	pngFile    = "TEST.PNG"
	rtfFile    = "TEST.rtf"
	webpFile   = "TEST.webp"
	asciiFile  = "TEST.ASC"
	ansiFile   = "TEST.ANS"
	txtFile    = "TEST.TXT"
	badFile    = "τεχτƒιℓε.τχτ"
	manualFile = "PKZ204EX.TXT"
	pdfFile    = "TEST.pdf"
	utf16File  = "TEST-U16.txt"
	iso7File   = "TEST-8859-7.txt"
	modFile    = "TEST.mod"
	xmFile     = "TEST.xm"
)

func uncompress(name string) string {
	_, file, _, usable := runtime.Caller(0)
	if !usable {
		panic("runtime.Caller failed")
	}
	d := filepath.Join(filepath.Dir(file), "../..")
	x := filepath.Join(d, "assets", "testdata", "uncompress", name)
	return x
}

func mp3file(name string) string {
	_, file, _, usable := runtime.Caller(0)
	if !usable {
		panic("runtime.Caller failed")
	}
	d := filepath.Join(filepath.Dir(file), "../..")
	x := filepath.Join(d, "assets", "testdata", "mp3", name)
	return x
}

func imgfile(name string) string {
	_, file, _, usable := runtime.Caller(0)
	if !usable {
		panic("runtime.Caller failed")
	}
	d := filepath.Join(filepath.Dir(file), "../..")
	x := filepath.Join(d, "assets", "testdata", "discimages", name)
	return x
}

func td(name string) string {
	_, file, _, usable := runtime.Caller(0)
	if !usable {
		panic("runtime.Caller failed")
	}
	d := filepath.Join(filepath.Dir(file), "../..")
	x := filepath.Join(d, "assets", "testdata", name)
	return x
}

func TestUnknowns(t *testing.T) {
	t.Parallel()

	data := "some binary data"
	nr := strings.NewReader(data)
	sign, err := magicnumberr.Archive(nr)
	require.NoError(t, err)
	assert.Equal(t, magicnumberr.Unknown, sign)
	assert.Equal(t, "binary data", sign.String())
	assert.Equal(t, "Binary data", sign.Title())

	b, sign, err := magicnumberr.MatchExt(emptyFile, nr)
	require.NoError(t, err)
	assert.False(t, b)
	assert.Equal(t, magicnumberr.PlainText, sign)

	r, err := os.Open(uncompress(emptyFile))
	require.NoError(t, err)
	defer r.Close()
	sign = magicnumberr.Find(r)
	assert.Equal(t, magicnumberr.ZeroByte, sign)
}

func TestFind(t *testing.T) {
	t.Parallel()
	prob := func(ext, path string) string {
		return fmt.Sprintf("ext: %s, path: %s", ext, path)
	}
	// walk the assets directory
	err := filepath.Walk(td(""), func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		ext := strings.ToUpper(filepath.Ext(path))
		if info.IsDir() || ext == "" {
			return nil
		}
		base := filepath.Base(path)
		skip := []string{"SAMPLE.DAT", "uncompress.bin"}
		for _, s := range skip {
			if s == base {
				return nil
			}
		}
		f, err := os.Open(path)
		require.NoError(t, err)
		defer f.Close()
		sign := magicnumberr.Find(f)
		if base == "τεχτƒιℓε.τχτ" {
			assert.Equal(t, magicnumberr.PlainText, sign, prob(ext, path))
			return nil
		}

		switch ext {
		case ".COM":
			// do not test as it returns different results based on the file
			return nil
		case ".7Z":
			assert.Equal(t, magicnumberr.X7zCompressArchive, sign, prob(ext, path))
		case ".ANS":
			assert.Equal(t, magicnumberr.ANSIEscapeText, sign, prob(ext, path))
		case ".ARC":
			// two different signatures used for the same file extension
			assert.Contains(t, []magicnumberr.Signature{
				magicnumberr.FreeArc,
				magicnumberr.ARChiveSEA,
			}, sign, prob(ext, path))
		case ".ARJ":
			assert.Equal(t, magicnumberr.ArchiveRobertJung, sign, prob(ext, path))
		case ".AVIF":
			assert.Equal(t, magicnumberr.AV1ImageFile, sign, prob(ext, path))
		case ".BAT", ".INI", ".CUE":
			assert.Equal(t, magicnumberr.PlainText, sign, prob(ext, path))
		case ".BMP":
			assert.Equal(t, magicnumberr.BMPFileFormat, sign, prob(ext, path))
		case ".BZ2":
			assert.Equal(t, magicnumberr.Bzip2CompressArchive, sign, prob(ext, path))
		case ".CHM", ".HLP":
			assert.Equal(t, magicnumberr.WindowsHelpFile, sign, prob(ext, path))
		case ".DAA":
			assert.Equal(t, magicnumberr.CDPowerISO, sign, prob(ext, path))
		case ".EXE", ".DLL":
			assert.Equal(t, magicnumberr.MicrosoftExecutable, sign, prob(ext, path))
		case ".GIF":
			assert.Equal(t, magicnumberr.GraphicsInterchangeFormat, sign, prob(ext, path))
		case ".GZ":
			assert.Equal(t, magicnumberr.GzipCompressArchive, sign, prob(ext, path))
		case ".JPG", ".JPEG":
			assert.Equal(t, magicnumberr.JPEGFileInterchangeFormat, sign, prob(ext, path))
		case ".ICO":
			assert.Equal(t, magicnumberr.MicrosoftIcon, sign, prob(ext, path))
		case ".IFF":
			assert.Equal(t, magicnumberr.InterleavedBitmap, sign, prob(ext, path))
		case ".ISO":
			assert.Equal(t, magicnumberr.CDISO9660, sign, prob(ext, path))
		case ".LZH":
			assert.Equal(t, magicnumberr.YoshiLHA, sign, prob(ext, path))
		case ".MP3":
			// do not test as it returns different results based on the file's ID3 tag
			return nil
		case ".PCX":
			assert.Equal(t, magicnumberr.PersonalComputereXchange, sign, prob(ext, path))
		case ".PNG":
			assert.Equal(t, magicnumberr.PortableNetworkGraphics, sign, prob(ext, path))
		case ".RAR":
			assert.Equal(t, magicnumberr.RoshalARchivev5, sign, prob(ext, path))
		case ".TAR":
			assert.Equal(t, magicnumberr.TapeARchive, sign, prob(ext, path))
		case ".TXT", ".MD", ".NFO", ".ME", ".DIZ", ".ASC", ".CAP", ".DOC":
			assert.Contains(t, []magicnumberr.Signature{
				magicnumberr.PlainText,
				magicnumberr.UTF16Text,
			}, sign, prob(ext, path))
		case ".WEBP":
			assert.Equal(t, magicnumberr.GoogleWebP, sign, prob(ext, path))
		case ".XZ":
			assert.Equal(t, magicnumberr.XZCompressArchive, sign, prob(ext, path))
		case ".ZIP":
			if base == "EMPTY.ZIP" {
				assert.Equal(t, magicnumberr.ZeroByte, sign, prob(ext, path))
				return nil
			}
			zips := []magicnumberr.Signature{
				magicnumberr.PKWAREZip,
				magicnumberr.PKWAREZip64,
				magicnumberr.PKWAREZipImplode,
				magicnumberr.PKWAREZipReduce,
				magicnumberr.PKWAREZipShrink,
			}
			assert.Contains(t, zips, sign, prob(ext, path))
		default:
			assert.NotEqual(t, magicnumberr.Unknown, sign, prob(ext, path))
			fmt.Fprintln(os.Stderr, ext, filepath.Base(path), fmt.Sprint(sign))
		}

		return nil
	})
	require.NoError(t, err)
}
