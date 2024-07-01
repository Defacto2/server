package magicnumber_test

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/Defacto2/server/internal/magicnumber"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func td(name string) string {
	_, file, _, usable := runtime.Caller(0)
	if !usable {
		panic("runtime.Caller failed")
	}
	d := filepath.Join(filepath.Dir(file), "../..")
	x := filepath.Join(d, "assets", "testdata", name)
	return x
}

func tduncompress(name string) string {
	_, file, _, usable := runtime.Caller(0)
	if !usable {
		panic("runtime.Caller failed")
	}
	d := filepath.Join(filepath.Dir(file), "../..")
	x := filepath.Join(d, "assets", "testdata", "uncompress", name)
	return x
}

func TestFinds(t *testing.T) {
	t.Parallel()
	f, err := os.Open(td("PKZ204EX.TXT"))
	require.NoError(t, err)
	defer f.Close()
	sign, err := magicnumber.Find(f)
	require.NoError(t, err)
	assert.Equal(t, magicnumber.PlainText, sign)

	f, err = os.Open(tduncompress("TEST.PNG"))
	require.NoError(t, err)
	defer f.Close()
	sign, err = magicnumber.Find(f)
	require.NoError(t, err)
	assert.Equal(t, magicnumber.PortableNetworkGraphics, sign)

	f, err = os.Open(tduncompress("TEST.GIF"))
	require.NoError(t, err)
	defer f.Close()
	sign, err = magicnumber.Find(f)
	require.NoError(t, err)
	assert.Equal(t, magicnumber.GraphicsInterchangeFormat, sign)

	f, err = os.Open(tduncompress("TEST.PCX"))
	require.NoError(t, err)
	defer f.Close()
	sign, err = magicnumber.Find(f)
	require.NoError(t, err)
	assert.Equal(t, magicnumber.PersonalComputereXchange, sign)

	fmt.Println(td("TAR135.TAR"))
	f, err = os.Open(td("TAR135.TAR"))
	require.NoError(t, err)
	defer f.Close()
	sign, err = magicnumber.Find(f)
	require.NoError(t, err)
	assert.Equal(t, magicnumber.TapeARchive, sign)
}

func TestANSIMatch(t *testing.T) {
	t.Parallel()
	b, err := os.ReadFile(td("PKZ204EX.TXT"))
	require.NoError(t, err)
	assert.False(t, magicnumber.Ansi(b))
	b, err = os.ReadFile(tduncompress("TEST.ANS"))
	require.NoError(t, err)
	assert.True(t, magicnumber.Ansi(b))
}

func TestAscii(t *testing.T) {
	t.Parallel()
	p := []byte("Hello, World!")
	assert.True(t, magicnumber.Ascii(p))
	p = []byte("Hello, World!\x00")
	assert.True(t, magicnumber.Ascii(p))
	p = []byte("Hello, World!\x01")
	assert.False(t, magicnumber.Ascii(p))
	const esc = "\x1b"
	p = []byte("Hello, World!" + esc)
	assert.True(t, magicnumber.Ascii(p))

	b, err := os.ReadFile(td("PKZ204EX.TXT"))
	require.NoError(t, err)
	assert.True(t, magicnumber.Txt(b))

	b, err = os.ReadFile(td("PKZ204EX.TXT"))
	require.NoError(t, err)
	assert.True(t, magicnumber.TxtWindows(b))

	b, err = os.ReadFile(td("PKZ204EX.ZIP"))
	require.NoError(t, err)
	assert.False(t, magicnumber.Txt(b))
}

func TestTextLatin1(t *testing.T) {
	t.Parallel()
	p := []byte("Hello, World!")
	assert.True(t, magicnumber.TxtLatin1(p))
	p = []byte("Hello, World! \x92")
	assert.False(t, magicnumber.TxtLatin1(p))
	p = []byte("Hello, World! \x03")
	assert.False(t, magicnumber.TxtLatin1(p))
	b, err := os.ReadFile(td("PKZ204EX.TXT"))
	require.NoError(t, err)
	assert.True(t, magicnumber.TxtLatin1(b))
}

func TestTxtWindows(t *testing.T) {
	t.Parallel()
	p := []byte("Hello, World!")
	assert.True(t, magicnumber.TxtWindows(p))

	p = []byte("Hello, World! \x92")
	assert.True(t, magicnumber.TxtWindows(p))

	p = []byte("Hello, World! \x8f")
	assert.False(t, magicnumber.TxtWindows(p))

	p = []byte("Hello, World! \x03")
	assert.False(t, magicnumber.TxtWindows(p))

	b, err := os.ReadFile(td("PKZ204EX.TXT"))
	require.NoError(t, err)
	assert.True(t, magicnumber.TxtWindows(b))
}

func TestFind512B(t *testing.T) {
	f, err := os.Open(tduncompress("TEST.JPEG"))
	require.NoError(t, err)
	defer f.Close()
	sign, err := magicnumber.Find(f)
	require.NoError(t, err)
	assert.Equal(t, magicnumber.JPEGFileInterchangeFormat, sign)
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
		skip := []string{"SAMPLE.DAT"}
		for _, s := range skip {
			if s == base {
				return nil
			}
		}
		f, err := os.Open(path)
		require.NoError(t, err)
		defer f.Close()
		sign, err := magicnumber.Find(f)
		require.NoError(t, err)

		if base == "τεχτƒιℓε.τχτ" {
			assert.Equal(t, magicnumber.PlainText, sign, prob(ext, path))
			return nil
		}

		switch ext {
		case ".COM": // binary files with no magic numbers
			assert.Equal(t, magicnumber.Unknown, sign, prob(ext, path))
		case ".7Z":
			assert.Equal(t, magicnumber.X7zCompressArchive, sign, prob(ext, path))
		case ".ANS":
			assert.Equal(t, magicnumber.ANSIEscapeText, sign, prob(ext, path))
		case ".ARC":
			assert.Equal(t, magicnumber.ARChiveSEA, sign, prob(ext, path))
		case ".ARJ":
			assert.Equal(t, magicnumber.ArchiveRobertJung, sign, prob(ext, path))
		case ".AVIF":
			assert.Equal(t, magicnumber.AV1ImageFile, sign, prob(ext, path))
		case ".BAT", ".INI":
			assert.Equal(t, magicnumber.PlainText, sign, prob(ext, path))
		case ".BMP":
			assert.Equal(t, magicnumber.BMPFileFormat, sign, prob(ext, path))
		case ".CHM", ".HLP":
			assert.Equal(t, magicnumber.WindowsHelpFile, sign, prob(ext, path))
		case ".EXE", ".DLL":
			assert.Equal(t, magicnumber.MicrosoftExecutable, sign, prob(ext, path))
		case ".GIF":
			assert.Equal(t, magicnumber.GraphicsInterchangeFormat, sign, prob(ext, path))
		case ".GZ":
			assert.Equal(t, magicnumber.GzipCompressArchive, sign, prob(ext, path))
		case ".JPG", ".JPEG":
			assert.Equal(t, magicnumber.JPEGFileInterchangeFormat, sign, prob(ext, path))
		case ".ICO":
			assert.Equal(t, magicnumber.MicrosoftIcon, sign, prob(ext, path))
		case ".IFF":
			assert.Equal(t, magicnumber.InterleavedBitmap, sign, prob(ext, path))
		case ".LZH":
			assert.Equal(t, magicnumber.YoshiLHA, sign, prob(ext, path)) // not working
		case ".PCX":
			assert.Equal(t, magicnumber.PersonalComputereXchange, sign, prob(ext, path))
		case ".PNG":
			assert.Equal(t, magicnumber.PortableNetworkGraphics, sign, prob(ext, path))
		case ".RAR":
			assert.Equal(t, magicnumber.RoshalARchivev5, sign, prob(ext, path))
		case ".TAR":
			assert.Equal(t, magicnumber.TapeARchive, sign, prob(ext, path))
		case ".TXT", ".MD", ".NFO", ".ME", ".DIZ", ".ASC", ".CAP", ".DOC":
			assert.Equal(t, magicnumber.PlainText, sign, prob(ext, path))
		case ".WEBP":
			assert.Equal(t, magicnumber.GoogleWebP, sign, prob(ext, path))
		case ".XZ":
			assert.Equal(t, magicnumber.XZCompressArchive, sign, prob(ext, path))
		case ".ZIP":
			assert.Equal(t, magicnumber.PKWAREZip, sign, prob(ext, path))
		default:
			assert.NotEqual(t, magicnumber.Unknown, sign, prob(ext, path))
			fmt.Fprintln(os.Stderr, ext, filepath.Base(path), fmt.Sprint(sign))
		}

		return nil
	})
	require.NoError(t, err)
}
