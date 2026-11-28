package filerecord

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Defacto2/magicnumber"
	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/Defacto2/server/internal/tags"
	"github.com/Defacto2/server/internal/testutil"
	"github.com/aarondl/null/v8"
	"github.com/nalgeon/be"
)

func TestFileIntegrity(t *testing.T) {
	t.Parallel()

	be.Equal(t, fileIntegrity(nil), "")
	be.Equal(t, fileIntegrity(&models.File{}), "")
	be.Equal(t, fileIntegrity(&models.File{FileIntegrityStrong: null.StringFrom("abc123")}), "abc123")
}

func TestProgr(t *testing.T) {
	t.Parallel()

	built := time.Date(1998, time.May, 4, 0, 0, 0, 0, time.UTC)

	// PE executable with valid TimeDateStamp.
	peExec := magicnumber.Windows{
		PE:            magicnumber.Intel386PE,
		NE:            magicnumber.NoneNE,
		TimeDateStamp: built,
	}
	got := progr(peExec, ".exe", 1024)
	be.True(t, strings.Contains(got, "executable"))
	be.True(t, strings.Contains(got, ", built 1998-05-04"))

	// DOS <= 64KB -> Dos command; > 64KB -> Dos executable.
	dosExec := magicnumber.Windows{PE: magicnumber.UnknownPE, NE: magicnumber.UnknownNE}
	be.True(t, strings.Contains(progr(dosExec, ".com", 64*1024), "Dos command"))
	be.True(t, strings.Contains(progr(dosExec, ".exe", 64*1024+1), "Dos executable"))

	// NE executable (Windows 3.x).
	neExec := magicnumber.Windows{
		PE:    magicnumber.UnknownPE,
		NE:    magicnumber.NewExecutable(1),
		Major: 3,
		Minor: 1,
	}
	be.True(t, strings.Contains(progr(neExec, ".exe", 2048), "executable"))

	// Plain MS-DOS program (NoneNE, UnknownPE).
	noneExec := magicnumber.Windows{PE: magicnumber.UnknownPE, NE: magicnumber.NoneNE}
	be.True(t, strings.Contains(progr(noneExec, ".exe", 2048), "MS Dos program"))

	// Windows DLL.
	be.True(t, strings.Contains(progr(peExec, ".dll", 4096), "Windows dynamic-link library"))

	// Non-DOS extensions with NE / PE data.
	be.True(t, strings.Contains(progr(neExec, ".drv", 4096), "NE program data"))
	be.True(t, strings.Contains(progr(noneExec, ".sys", 4096), "PE program data"))
}

func TestCOMHelpers(t *testing.T) {
	t.Parallel()

	be.True(t, isCOMPrintable('A'))
	be.True(t, isCOMPrintable('\t'))
	be.True(t, isCOMPrintable('\n'))
	be.True(t, isCOMPrintable('\r'))
	be.True(t, !isCOMPrintable(0x01))

	be.True(t, isCOMUnexpectedCtrl(0x7F))
	be.True(t, !isCOMUnexpectedCtrl('A'))
	be.True(t, !isCOMUnexpectedCtrl('\t'))
	be.True(t, !isCOMUnexpectedCtrl(0x1A))
	be.True(t, !isCOMUnexpectedCtrl(0x10))
	be.True(t, !isCOMUnexpectedCtrl(0x1F))
	be.True(t, isCOMUnexpectedCtrl(0x01))

	tmp := t.TempDir()
	openTemp := func(name string, data []byte) *os.File {
		t.Helper()
		p := filepath.Join(tmp, name)
		be.Err(t, os.WriteFile(p, data, 0o600), nil)
		f, err := os.Open(p)
		be.Err(t, err, nil)
		t.Cleanup(func() { _ = f.Close() })
		return f
	}

	// Valid text with trailing DOS EOF (0x1A) and NUL padding -> not a binary COM stub.
	fText := openTemp("text.com", []byte("Hello scene world!\r\n\x1a\x00\x00"))
	data, ok := readCOM(fText)
	be.True(t, ok)
	be.Equal(t, string(data), "Hello scene world!\r\n")
	be.True(t, !simpleCOM(fText))

	// Empty file -> readCOM returns false, simpleCOM returns true.
	fEmpty := openTemp("empty.com", nil)
	_, ok = readCOM(fEmpty)
	be.True(t, !ok)
	be.True(t, simpleCOM(fEmpty))

	// Embedded NUL byte -> readCOM returns false.
	fNul := openTemp("nul.com", []byte("abc\x00def"))
	_, ok = readCOM(fNul)
	be.True(t, !ok)
	be.True(t, simpleCOM(fNul))

	// Unexpected control bytes exceeding threshold -> simpleCOM returns true.
	fCtrl := openTemp("ctrl.com", append(bytes.Repeat([]byte{'A'}, 10), 0x01, 0x02))
	be.True(t, simpleCOM(fCtrl))

	// Closed file -> Seek fails cleanly.
	closedPath := filepath.Join(tmp, "closed.com")
	be.Err(t, os.WriteFile(closedPath, []byte("abc"), 0o600), nil)
	fClosed, err := os.Open(closedPath)
	be.Err(t, err, nil)
	be.Err(t, fClosed.Close(), nil)
	_, ok = readCOM(fClosed)
	be.True(t, !ok)
}

func TestSkippedEmpty(t *testing.T) {
	t.Parallel()

	be.Equal(t, skippedEmpty(0), "")
	be.True(t, strings.Contains(skippedEmpty(3), "skipped 3 empty (0 B) files"))
}

func TestEntryHelpers(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	missing := filepath.Join(tmp, "does-not-exist")

	// isImage excludes RIPscrip.
	e := Entry{sign: magicnumber.RIPscrip}
	be.True(t, !e.isImage())
	e.sign = magicnumber.PortableNetworkGraphics
	be.True(t, e.isImage())

	// isProgram skips when platform is DOS, matches for other platforms.
	e.sign = magicnumber.Programs()[0]
	be.True(t, !e.isProgram(tags.DOS.String()))
	be.True(t, e.isProgram(tags.Windows.String()))

	// isBinaryText branches.
	e.sign = magicnumber.XBinaryText
	be.True(t, !e.isBinaryText(tags.Text.String()))
	e.sign = magicnumber.CGAVideoDump
	be.True(t, e.isBinaryText(tags.DOS.String()))
	e.sign = magicnumber.Unknown
	be.True(t, e.isBinaryText(tags.Text.String()))
	be.True(t, e.isBinaryText(tags.ANSI.String()))
	be.True(t, !e.isBinaryText(tags.DOS.String()))

	// isText special cases.
	e.sign = magicnumber.ICEDraw
	be.True(t, e.isText())
	e.sign = magicnumber.XBinaryText
	be.True(t, e.isText())

	// skipImage: missing, PNG, ILBM fallback, and generic fallback.
	be.True(t, e.skipImage(missing))

	pngEntry := Entry{sign: magicnumber.PortableNetworkGraphics}
	be.True(t, !pngEntry.skipImage(testutil.FileData("testpng").Abs()))
	be.True(t, strings.HasPrefix(pngEntry.format, "png image, "))

	rawImg := filepath.Join(tmp, "fake.lbm")
	be.Err(t, os.WriteFile(rawImg, []byte("FORM\x00\x00\x00\x20ILBMBMHD\x00\x00\x00\x14\x01\x40\x00\xc8"), 0o600), nil)
	ilbmEntry := Entry{sign: magicnumber.InterleavedBitmap}
	be.True(t, !ilbmEntry.skipImage(rawImg))
	be.True(t, strings.Contains(ilbmEntry.format, "ILBM image, "))

	fallbackImg := Entry{sign: magicnumber.PortableNetworkGraphics}
	be.True(t, !fallbackImg.skipImage(rawImg))
	be.Equal(t, fallbackImg.format, magicnumber.PortableNetworkGraphics.Title()+" image")

	// skipProgram: missing and valid file.
	be.True(t, e.skipProgram(missing))
	be.True(t, !e.skipProgram(testutil.FileData("readmetxt").Abs()))

	// skipMusicMOD: missing and synthetic XM header.
	be.True(t, e.skipMusicMOD(missing))
	xmPath := filepath.Join(tmp, "song.xm")
	xm := make([]byte, 80)
	copy(xm[0:], "Extended Module: ")
	copy(xm[17:], "Demo Tune")
	be.Err(t, os.WriteFile(xmPath, xm, 0o600), nil)
	modEntry := Entry{}
	be.True(t, !modEntry.skipMusicMOD(xmPath))

	// skipMusic: missing, ID3v1 tag, and non-tagged audio file.
	be.True(t, e.skipMusic(missing))
	mp3Path := filepath.Join(tmp, "song.mp3")
	mp3 := make([]byte, 256)
	copy(mp3[128:], "TAG")
	copy(mp3[131:], "Track Title")
	copy(mp3[161:], "Scene Artist")
	copy(mp3[221:], "1999")
	be.Err(t, os.WriteFile(mp3Path, mp3, 0o600), nil)
	audioEntry := Entry{}
	be.True(t, !audioEntry.skipMusic(mp3Path))
	be.True(t, strings.Contains(audioEntry.module, "Track Title"))

	// Plain file without ID3 tags still returns false.
	plainAudio := filepath.Join(tmp, "plain.mp3")
	be.Err(t, os.WriteFile(plainAudio, []byte("not an id3 tag"), 0o600), nil)
	plainEntry := Entry{}
	be.True(t, !plainEntry.skipMusic(plainAudio))
	be.Equal(t, plainEntry.module, "")

	// listErr and listEntry populate fields cleanly.
	leErr := listErr(pngEntry)
	be.Equal(t, leErr.RelativeName, "")
	be.Equal(t, leErr.ImageConfig, pngEntry.format)

	le := listEntry(pngEntry, "test.png", "uid-1")
	be.Equal(t, le.RelativeName, "test.png")
	be.Equal(t, le.UniqueID, "uid-1")
}
