package app

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Defacto2/helper"
	"github.com/Defacto2/magicnumber"
	"github.com/Defacto2/server/handler/internal/simple"
	"github.com/Defacto2/server/internal/dir"
	"github.com/Defacto2/server/internal/testutil"
	"github.com/labstack/echo/v5"
	"github.com/nalgeon/be"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/encoding/unicode"
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
	be.Equal(t, got, unicode.UTF8)

	r = strings.NewReader("Denuvo взяла на работу хакера\n")
	got = d.encoding(r)
	be.Equal(t, got, unicode.UTF8)

	// Legacy CP437 box drawing and ISO-8859-1 accented bytes are not valid UTF-8
	// and must continue to be handled as CP437 / ISO-8859-1.
	r = strings.NewReader("\xc4\xc4\xc4\n")
	got = d.encoding(r)
	be.Equal(t, got, cp437)

	r = strings.NewReader("caf\xe9\n")
	got = d.encoding(r)
	be.Equal(t, got, latin1)
}

func Test_dirs_addtext8bit_utf8_defaults_web(t *testing.T) {
	t.Parallel()

	d := Dirs{}
	data := map[string]any{
		"webCheck":       "",
		"topazCheck":     "",
		"vgaCheck":       "checked",
		"preClassUTF8":   "d-none ",
		"preClassLatin1": "d-none ",
		"preClassCP437":  "font-dos font-large",
	}
	got := d.addtext8bit(nil, []byte("Denuvo взяла на работу хакера\n"), data)
	be.Equal(t, got["webCheck"], "checked")
	be.Equal(t, got["topazCheck"], "")
	be.Equal(t, got["vgaCheck"], "")
	be.Equal(t, got["preClassUTF8"], "")
	be.Equal(t, got["preClassLatin1"], "d-none ")
	be.Equal(t, got["preClassCP437"], "d-none font-dos font-large")
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
	be.Equal(t, got, `<span class="bg-body-tertiary border rounded px-2 py-0 font-monospace fs-6 text-body-secondary">`+
		`filename.txt</span>`)

	art.GroupBrandBy.String = "NOCHANGE BBS"
	art.GroupBrandBy.Valid = true
	got = firstLead(art)
	be.True(t, strings.Contains(got, `d-inline-flex flex-wrap align-items-center gap-2`))
	be.True(t, strings.Contains(got, `NOCHANGE BBS`))
	be.True(t, strings.Contains(got, `filename.txt`))
	be.True(t, !strings.Contains(got, `<br>`))
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
	be.True(t, plainText(magicnumber.ANSIEscapeText.Title()))
	be.True(t, plainText(magicnumber.ANSIEscapeText.Title()+simple.ANSIBlink))
	be.True(t, plainText(simple.ANSImation))
}

func Test_radioPublicHidden(t *testing.T) {
	t.Parallel()

	pubOn := string(radioPublic(true))
	hidOn := string(radioHidden(true))
	be.True(t, strings.Contains(pubOn, "checked"))
	be.True(t, !strings.Contains(hidOn, "checked"))
	be.True(t, strings.Contains(pubOn, `hx-swap="none"`))
	be.True(t, strings.Contains(hidOn, `hx-swap="none"`))
	be.True(t, strings.Contains(pubOn, `hx-sync="closest .row:replace"`))
	be.True(t, strings.Contains(hidOn, `hx-sync="closest .row:replace"`))

	pubOff := string(radioPublic(false))
	hidOff := string(radioHidden(false))
	be.True(t, !strings.Contains(pubOff, "checked"))
	be.True(t, strings.Contains(hidOff, "checked"))
}
