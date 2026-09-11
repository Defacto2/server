package app

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Defacto2/helper"
	"github.com/Defacto2/magicnumber"
	"github.com/Defacto2/server/internal/dir"
	"github.com/Defacto2/server/internal/testutil"
	"github.com/labstack/echo/v5"
	"github.com/nalgeon/be"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"
)

var (
	latin1 encoding.Encoding = charmap.ISO8859_1
	cp437  encoding.Encoding = charmap.CodePage437
)

func echoCtx(t *testing.T) *echo.Context {
	t.Helper()

	e := echo.New()
	r := httptest.NewRequestWithContext(
		t.Context(), http.MethodPost, "/", strings.NewReader("{}"))
	r.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	w := httptest.NewRecorder()

	return e.NewContext(r, w)
}

func Test_dirs_addEditor(t *testing.T) {
	t.Parallel()

	d := Dirs{}
	x := d.addEditor(t.Context(), nil, echoCtx(t), nil, nil)
	be.True(t, len(x) == 0)
}

func Test_dirs_encoding(t *testing.T) {
	t.Parallel()

	d := Dirs{}
	got := d.encoding(nil)
	be.Equal(t, got, cp437)
}

func Test_dirs_encoding_amiga(t *testing.T) {
	t.Parallel()

	d := Dirs{Platform: "textamiga"}
	got := d.encoding(nil)
	be.Equal(t, got, latin1)
}

func Test_dirs_encoding_section(t *testing.T) {
	t.Parallel()

	d := Dirs{Section: "appleii"}
	got := d.encoding(nil)
	be.Equal(t, got, latin1)

	d.Section = "atarist"
	got = d.encoding(nil)
	be.Equal(t, got, latin1)
}

func Test_dirs_encoding_textdos(t *testing.T) {
	t.Parallel()

	d := Dirs{Platform: "textdos"}
	r := strings.NewReader("Hello\nworld\nthis is some text.\n")
	got := d.encoding(r)
	be.Equal(t, got, latin1)
}

func Test_dirs_encoding_textutf8(t *testing.T) {
	t.Parallel()

	d := Dirs{}
	r := strings.NewReader("Hello\nworld\nthis is some text. 👾\n")
	got := d.encoding(r)
	// without a byte-order-mark we cannot reliably determine 8-bit CP-437 over UTF-8,
	// as both rely on 8-bit character sets. so the priority is to render legacy text.
	be.Equal(t, got, cp437)
}

func Test_screenshot(t *testing.T) {
	t.Parallel()

	var d Dirs
	be.True(t, !d.screenshot())

	d.Platform = "textamiga"
	be.True(t, !d.screenshot())

	const unid = "5b4c5f6e-8a1e-11e9-9f0e-000000000000"
	d.UUID = unid
	d.Platform = ""
	temp := t.TempDir()
	name := filepath.Join(temp, unid) + ".webp"
	if err := helper.Touch(name); err != nil {
		be.Err(t, err, nil)
		return
	}
	d.Preview = dir.Directory(temp)
	be.True(t, d.screenshot())
}

func Test_decode(t *testing.T) {
	t.Parallel()

	input := "hello world"
	r := strings.NewReader(input)
	got, err := decode(r)
	be.Err(t, err, nil)
	be.Equal(t, got, "hello world\n")
}

func Test_firstLead(t *testing.T) {
	t.Parallel()

	got := firstLead(nil)
	be.Equal(t, got, "")

	art := testutil.NewModel(t)
	got = firstLead(art)
	be.Equal(t, got, `<br><span class="font-monospace fs-5 fw-light">`+
		`filename.txt</span>`)
}

func Test_legacyArchiving(t *testing.T) {
	t.Parallel()

	got := legacyArchiving(nil)
	be.True(t, !got)

	got = legacyArchiving(magicnumber.PKWAREZipImplode.Title())
	be.True(t, got)
}

func Test_lockWidth(t *testing.T) {
	t.Parallel()

	got := lockWidth(3, []byte("abcdef"))
	be.Equal(t, got, []byte("\nabc\ndef"))

	got = lockWidth(2, []byte("abcdef"))
	be.Equal(t, got, []byte("\nab\ncd\nef"))

	got = lockWidth(1, []byte("abcdef"))
	be.Equal(t, got, []byte("\na\nb\nc\nd\ne\nf"))

	got = lockWidth(0, []byte("abcdef"))
	be.Equal(t, got, []byte("abcdef"))
}

func Test_plainText(t *testing.T) {
	t.Parallel()

	got := plainText(magicnumber.PlainText.Title())
	be.True(t, got)
}
