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

func TestFindExecutable(t *testing.T) {
	t.Parallel()

	w := magicnumber.FindExecutable([]byte{})
	assert.Equal(t, magicnumber.UnknownPE, w.PE)
	assert.Equal(t, magicnumber.NoneNE, w.NE)

	freedos := []string{"/exe/EXE.EXE", "/exemenu/exemenu.exe", "/press/PRESS.EXE", "/rread/rread.exe"}
	for _, v := range freedos {
		p, err := os.ReadFile(td("binaries/freedos" + v))
		require.NoError(t, err)
		w = magicnumber.FindExecutable(p)
		assert.Equal(t, magicnumber.UnknownPE, w.PE)
		assert.Equal(t, magicnumber.NoneNE, w.NE)
	}

	vista := []string{"/hello.com", "/hellojs.com", "/life.com"}
	for _, v := range vista {
		p, err := os.ReadFile(td("binaries/windows" + v))
		require.NoError(t, err)
		w = magicnumber.FindExecutable(p)
		assert.Equal(t, magicnumber.AMD64PE, w.PE)
		assert.Equal(t, 6, w.Major)
		assert.Equal(t, 0, w.Minor)
		assert.Equal(t, 2019, w.TimeDateStamp.Year())
		assert.Equal(t, "Windows Vista 64-bit", fmt.Sprint(w))
		assert.Equal(t, magicnumber.NoneNE, w.NE)
	}

	winv3 := []string{"/calmir10/CALMIRA.EXE", "/calmir10/TASKBAR.EXE", "/dskutl21/DISKUTIL.EXE"}
	for _, v := range winv3 {
		p, err := os.ReadFile(td("binaries/windows3x" + v))
		require.NoError(t, err)
		w = magicnumber.FindExecutable(p)
		assert.Equal(t, magicnumber.UnknownPE, w.PE)
		assert.Equal(t, magicnumber.Windows286Exe, w.NE)
		assert.Equal(t, 3, w.Major)
		assert.Equal(t, 10, w.Minor)
		assert.Equal(t, "Windows v3.10 for 286", fmt.Sprint(w))
	}

	p, err := os.ReadFile(td("binaries/windowsXP/CoreTempv13/32bit/Core Temp.exe"))
	require.NoError(t, err)
	w = magicnumber.FindExecutable(p)
	assert.Equal(t, magicnumber.Intel386PE, w.PE)
	assert.Equal(t, magicnumber.NoneNE, w.NE)
	assert.Equal(t, 5, w.Major)
	assert.Equal(t, 0, w.Minor)
	assert.Equal(t, "Windows 2000 32-bit", fmt.Sprint(w))

	p, err = os.ReadFile(td("binaries/windowsXP/CoreTempv13/64bit/Core Temp.exe"))
	require.NoError(t, err)
	w = magicnumber.FindExecutable(p)
	assert.Equal(t, magicnumber.AMD64PE, w.PE)
	assert.Equal(t, magicnumber.NoneNE, w.NE)
	assert.Equal(t, 5, w.Major)
	assert.Equal(t, 2, w.Minor)
	assert.Equal(t, "Windows XP Professional x64 Edition 64-bit", fmt.Sprint(w))
}

func TestFindExecutableWinNT(t *testing.T) {
	win9x := []string{
		"/rlowe-encrypt/DEMOCD.EXE",
		"/rlowe-encrypt/DISKDVR.EXE",
		"/rlowe-cdrools/DEMOCD.EXE",
		"/7za920/7za.exe",
		"/7z1604-extra/7za.exe",
	}
	for _, v := range win9x {
		p, err := os.ReadFile(td("binaries/windows9x" + v))
		require.NoError(t, err)
		w := magicnumber.FindExecutable(p)
		assert.Equal(t, magicnumber.Intel386PE, w.PE)
		assert.Equal(t, 4, w.Major)
		assert.Equal(t, 0, w.Minor)
		assert.Greater(t, w.TimeDateStamp.Year(), 2000)
		assert.Equal(t, "Windows NT v4.0", fmt.Sprint(w))
		assert.Equal(t, magicnumber.NoneNE, w.NE)
	}
	unknown := []string{
		"/rlowe-rformat/RFORMATD.EXE",
		"/rlowe-encrypt/DFMINST.COM",
		"/rlowe-encrypt/UNINST.COM",
	}
	for _, v := range unknown {
		p, err := os.ReadFile(td("binaries/windows9x" + v))
		require.NoError(t, err)
		w := magicnumber.FindExecutable(p)
		assert.Equal(t, magicnumber.UnknownPE, w.PE)
		assert.Equal(t, 0, w.Major)
		assert.Equal(t, 0, w.Minor)
		assert.Equal(t, w.TimeDateStamp.Year(), 1)
		assert.Equal(t, "Unknown PE executable", fmt.Sprint(w))
		assert.Equal(t, magicnumber.NoneNE, w.NE)
	}

	p, err := os.ReadFile(td("binaries/windows9x/7z1604-extra/x64/7za.exe"))
	require.NoError(t, err)
	w := magicnumber.FindExecutable(p)
	assert.Equal(t, magicnumber.AMD64PE, w.PE)
	assert.Equal(t, 4, w.Major)
	assert.Equal(t, 0, w.Minor)
	assert.Equal(t, w.TimeDateStamp.Year(), 2016)
	assert.Equal(t, "Windows NT v4.0 64-bit", fmt.Sprint(w))
	assert.Equal(t, magicnumber.NoneNE, w.NE)
}

func TestXXX(t *testing.T) {
	t.Parallel()
	// test the test data paths
	p, err := os.ReadFile("TGDEMO.EXE")
	require.NoError(t, err)
	w := magicnumber.NE(p)
	assert.Equal(t, w.NE, magicnumber.Windows286Exe)
	assert.Equal(t, w.Major, 3)
	assert.Equal(t, w.Minor, 0)
	fmt.Printf("NE: %+v\n---\n", w)

	p, err = os.ReadFile("XXX.exe")
	require.NoError(t, err)
	w = magicnumber.PE(p)
	assert.Equal(t, magicnumber.Intel386PE, w.PE)
	fmt.Printf("PE: %+v\n", w)

	p, err = os.ReadFile("7z.exe")
	require.NoError(t, err)
	w = magicnumber.PE(p)
	assert.Equal(t, magicnumber.Intel386PE, w.PE)
	fmt.Printf("PE: %+v\n", w)

	p, err = os.ReadFile("7za.exe")
	require.NoError(t, err)
	w = magicnumber.PE(p)
	assert.NotEqual(t, magicnumber.AMD64PE, w.PE)
	fmt.Printf("PE: %+v\n", w)

	p, err = os.ReadFile("life.com")
	require.NoError(t, err)
	w = magicnumber.FindExecutable(p)
	fmt.Printf(">>%+v\n", w)

	p, err = os.ReadFile("hello.com")
	require.NoError(t, err)
	w = magicnumber.FindExecutable(p)
	fmt.Printf(">>%+v\n", w)

	p, err = os.ReadFile("hellojs.com")
	require.NoError(t, err)
	w = magicnumber.FindExecutable(p)
	fmt.Printf(">>%+v\n", w)

	x := uint8(2)
	for _, v := range magicnumber.Flags(x) {
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
