package readme_test

import (
	"bytes"
	"log/slog"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Defacto2/magicnumber"
	"github.com/Defacto2/server/handler/readme"
	"github.com/Defacto2/server/internal/dir"
	"github.com/Defacto2/server/internal/logs"
	"github.com/Defacto2/server/internal/testutil"
	"github.com/nalgeon/be"
)

func TestNormalize(t *testing.T) {
	t.Parallel()

	nulls := bytes.Repeat([]byte{0x00}, 50)
	text := []byte("hello world.")
	crlf := []byte("\r\n")
	eof := []byte{0x1a}
	serial := []byte(" Serial!! B014-56789A-BCDEFGH-G45X example << ")

	s := [][]byte{text, serial, text, crlf, eof, nulls}
	got := readme.Normalize(bytes.Join(s, []byte("")))

	const want = 122
	be.Equal(t, len(got), want)

	wants := `hello world. Sxxxxx!! B014-5$$89A-BCDEFGH-G45X example << hello world.` +
		"\n\x1a" +
		strings.Repeat(` `, 50)
	be.Equal(t, string(got), wants)
}

func TestBuffers0(t *testing.T) {
	t.Parallel()

	sl := logs.Discard()
	txt := readme.Text{}
	got0, got1, err := txt.Buffers(sl)
	be.Err(t, err, nil)
	be.Equal(t, got0, nil)
	be.Equal(t, got1, nil)
}

func TestBuffers1(t *testing.T) {
	t.Parallel()

	const unid = "d9c025c8-44be-4ec5-a9ec-40463a39a722"
	testdata, err := filepath.Abs("testdata")
	be.Err(t, err, nil)

	sl := slog.Default()
	txt := readme.Text{
		Download: dir.Directory(testdata),
		Extra:    dir.Directory(testdata),
		Filename: "testbuffers1.zip",
		MaxSize:  1024 * 100,
		UUID:     unid,
	}

	got0, got1, err := txt.Buffers(sl)
	be.Err(t, err, nil)

	const diz = ":DIZ BEGIN:\n\n  <- TAB\n  A placeholder FILE_ID.DIZ\n\n: DIZ END :\n\n\n" // TODO: research the 3 * newline
	const body = ":TXT BEGIN:\n\nHELLO WORLD!\n\n: TXT END :\n"
	const help = "\n:HLP BEGIN:\n\nhelper text...\n\n: HLP END :\n"
	const long = "skipped, text is too long\n"

	wants := diz + body + help
	got := got0.String()
	be.Equal(t, got, wants)

	got = got1.String()
	be.Equal(t, got, body)

	// test the too long feedback
	//
	txt.MaxSize = 1 // 1 byte
	got0, got1, err = txt.Buffers(sl)
	be.Err(t, err, nil)

	wants = diz + long + help
	got = got0.String()
	be.Equal(t, got, wants)

	got = got1.String()
	be.Equal(t, got, "") // TODO: expected?
}

func TestBuffers2(t *testing.T) {
	t.Parallel()

	const unid = "3e54b770-e71c-4e95-aa8c-f898e6b18f77"
	testdata, err := filepath.Abs("testdata")
	be.Err(t, err, nil)

	sl := slog.Default()
	txt := readme.Text{
		Download: dir.Directory(testdata),
		Extra:    dir.Directory(testdata),
		Filename: "testbuffers1.zip",
		MaxSize:  1024 * 100,
		UUID:     unid,
	}
	got0, got1, err := txt.Buffers(sl)
	be.Err(t, err, nil)

	const wants = `<div style="color:#aaa;background-color:#000;"><span style="color:#aaa;">:TXT CSI BEGIN:</span>` +
		"\n\n" +
		`<span style="color:#fff;">hello world</span>` +
		"\n\n" +
		`<span style="color:#aaa;">: TXT CSI END :</span>` +
		"\n\n" +
		`</div>`
	got := got0.String()
	be.Equal(t, got, wants)

	got = got1.String()
	be.Equal(t, got, "")
}

func TestBuffers3(t *testing.T) {
	t.Parallel()

	const unid = "3e54b770-e71c-4e95-aa8c-f898e6b18f77"
	const body = "hello world!!!"
	dest, err := testutil.MkDOSBIN(t, unid+".txt", body)
	be.Err(t, err, nil)
	temp := filepath.Dir(dest)

	sl := slog.Default()
	txt := readme.Text{
		Download: dir.Directory(temp),
		Extra:    dir.Directory(temp),
		Filename: "hello.bin",
		MaxSize:  1024 * 100,
		Sign:     magicnumber.Unknown,
		UUID:     unid,
	}
	got0, _, err := txt.Buffers(sl)
	be.Err(t, err, nil)
	got := got0.String()
	be.True(t, len(got) > 0)

	const pre = `<div><span style="color:#fff;background-color:#000;">`
	const charsPerLine = 80
	padding := bytes.Repeat([]byte{0x20}, charsPerLine-len(body))
	const suf = "</span>\n</div>"

	wants := pre + body + string(padding) + suf
	be.Equal(t, got, wants)
}
