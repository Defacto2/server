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

func TestXXX(t *testing.T) {
	t.Parallel()
	// test the test data paths
	p, err := os.ReadFile("TGDEMO.EXE")
	require.NoError(t, err)
	b, maj, min := magicnumber.NE(p)
	assert.Equal(t, b, magicnumber.Windows286Exe)
	assert.Equal(t, 3, maj)
	assert.Equal(t, 0, min)

	p, err = os.ReadFile("XXX.exe")
	require.NoError(t, err)
	pe, _, _ := magicnumber.PE(p)
	assert.Equal(t, pe, magicnumber.Intel386PE)

	p, err = os.ReadFile("7z.exe")
	require.NoError(t, err)
	pe, _, _ = magicnumber.PE(p)
	assert.Equal(t, pe, magicnumber.Intel386PE)

	p, err = os.ReadFile("7za.exe")
	require.NoError(t, err)
	pe, _, _ = magicnumber.PE(p)
	assert.Equal(t, pe, magicnumber.AMD64PE)

	var x = uint8(2)
	for _, v := range magicnumber.Indexes(x) {
		fmt.Println(v)
	}
	fmt.Printf("%08b\n", x)

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
	assert.True(t, magicnumber.ASCII(p))
	p = []byte("Hello, World!\x00")
	assert.True(t, magicnumber.ASCII(p))
	p = []byte("Hello, World!\x01")
	assert.False(t, magicnumber.ASCII(p))
	const esc = "\x1b"
	p = []byte("Hello, World!" + esc)
	assert.True(t, magicnumber.ASCII(p))

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
	sign, err := magicnumber.Find512B(f)
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
		skip := []string{"SAMPLE.DAT", "uncompress.bin"}
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
		case ".BAT", ".INI", ".CUE":
			assert.Equal(t, magicnumber.PlainText, sign, prob(ext, path))
		case ".BMP":
			assert.Equal(t, magicnumber.BMPFileFormat, sign, prob(ext, path))
		case ".CHM", ".HLP":
			assert.Equal(t, magicnumber.WindowsHelpFile, sign, prob(ext, path))
		case ".DAA":
			assert.Equal(t, magicnumber.CDPowerISO, sign, prob(ext, path))
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
		case ".ISO":
			assert.Equal(t, magicnumber.CDISO9660, sign, prob(ext, path))
		case ".LZH":
			assert.Equal(t, magicnumber.YoshiLHA, sign, prob(ext, path))
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

func TestMatchExt(t *testing.T) {
	t.Parallel()
	base := "TEST.JPEG"
	f, err := os.Open(tduncompress(base))
	require.NoError(t, err)
	defer f.Close()
	match, sign, err := magicnumber.MatchExt(base, f)
	require.NoError(t, err)
	assert.True(t, match)
	assert.Equal(t, magicnumber.JPEGFileInterchangeFormat, sign)

	// test a mismatch extension
	base = "TEST.PNG"
	f, err = os.Open(tduncompress(base))
	require.NoError(t, err)
	defer f.Close()
	match, sign, err = magicnumber.MatchExt("TEST.JPG", f)
	require.NoError(t, err)
	assert.False(t, match)
	assert.Equal(t, magicnumber.PortableNetworkGraphics, sign)
}

func TestArchive(t *testing.T) {
	t.Parallel()
	f, err := os.Open(td("ARC521P.ARC"))
	require.NoError(t, err)
	defer f.Close()
	sign, err := magicnumber.Archive(f)
	require.NoError(t, err)
	assert.Equal(t, magicnumber.ARChiveSEA, sign)

	f, err = os.Open(tduncompress("TEST.PNG"))
	require.NoError(t, err)
	defer f.Close()
	sign, err = magicnumber.Archive(f)
	require.NoError(t, err)
	assert.Equal(t, magicnumber.Unknown, sign)
}

func TestDiscs(t *testing.T) {
	t.Parallel()
	f, err := os.Open(td("discimages/uncompress.iso"))
	require.NoError(t, err)
	defer f.Close()
	sign, err := magicnumber.DiscImage(f)
	require.NoError(t, err)
	assert.Equal(t, magicnumber.CDISO9660, sign)

	f, err = os.Open(tduncompress("TEST.PNG"))
	require.NoError(t, err)
	defer f.Close()
	sign, err = magicnumber.DiscImage(f)
	require.NoError(t, err)
	assert.Equal(t, magicnumber.Unknown, sign)
}

func TestDocument(t *testing.T) {
	t.Parallel()
	f, err := os.Open(td("PKZ204EX.TXT"))
	require.NoError(t, err)
	defer f.Close()
	sign, err := magicnumber.Document(f)
	require.NoError(t, err)
	assert.Equal(t, magicnumber.PlainText, sign)

	f, err = os.Open(tduncompress("TEST.PNG"))
	require.NoError(t, err)
	defer f.Close()
	sign, err = magicnumber.Document(f)
	require.NoError(t, err)
	assert.Equal(t, magicnumber.Unknown, sign)
}

func TestImage(t *testing.T) {
	t.Parallel()

	f, err := os.Open(tduncompress("TEST.PNG"))
	require.NoError(t, err)
	defer f.Close()
	sign, err := magicnumber.Image(f)
	require.NoError(t, err)
	assert.Equal(t, magicnumber.PortableNetworkGraphics, sign)

	f, err = os.Open(td("TEST.7z"))
	require.NoError(t, err)
	defer f.Close()
	sign, err = magicnumber.Image(f)
	require.NoError(t, err)
	assert.Equal(t, magicnumber.Unknown, sign)
}

func TestProgram(t *testing.T) {
	t.Parallel()

	f, err := os.Open(td("binaries/freedos/press/PRESS.EXE"))
	require.NoError(t, err)
	defer f.Close()
	sign, err := magicnumber.Program(f)
	require.NoError(t, err)
	assert.Equal(t, magicnumber.MicrosoftExecutable, sign)

	f, err = os.Open(tduncompress("TEST.PNG"))
	require.NoError(t, err)
	defer f.Close()
	sign, err = magicnumber.Program(f)
	require.NoError(t, err)
	assert.Equal(t, magicnumber.Unknown, sign)
}
