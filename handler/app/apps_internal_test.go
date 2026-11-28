package app

import (
	"archive/zip"
	"database/sql"
	"errors"
	"math"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Defacto2/helper"
	"github.com/Defacto2/magicnumber"
	"github.com/Defacto2/server/handler/internal/simple"
	"github.com/Defacto2/server/internal/dir"
	"github.com/Defacto2/server/internal/logs"
	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/Defacto2/server/internal/testutil"
	"github.com/aarondl/null/v8"
	"github.com/labstack/echo/v5"
	"github.com/nalgeon/be"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/encoding/unicode"
)

var (
	latin1 encoding.Encoding = charmap.ISO8859_1
	cp437  encoding.Encoding = charmap.CodePage437

	errSampleApp = errors.New("sample app error")
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

func TestAddEditorNil(t *testing.T) {
	t.Parallel()

	d := Dirs{}
	x := d.addEditor(t.Context(), nil, echoCtx(t), nil, nil)
	be.True(t, len(x) == 0)
}

func TestEncoding(t *testing.T) {
	t.Parallel()

	be.Equal(t, (&Dirs{}).encoding(nil), cp437)
	be.Equal(t, (&Dirs{Platform: "textamiga"}).encoding(nil), latin1)
	be.Equal(t, (&Dirs{Section: "appleii"}).encoding(nil), latin1)
	be.Equal(t, (&Dirs{Section: "atarist"}).encoding(nil), latin1)

	d := Dirs{Platform: "textdos"}
	be.Equal(t, d.encoding(strings.NewReader("Hello\nworld\nthis is some text.\n")), latin1)

	d = Dirs{}
	be.Equal(t, d.encoding(strings.NewReader("Hello\nworld\nthis is some text. 👾\n")), unicode.UTF8)
	be.Equal(t, d.encoding(strings.NewReader("Denuvo взяла на работу хакера\n")), unicode.UTF8)
	// Legacy CP437 box drawing and ISO-8859-1 accented bytes are not valid UTF-8.
	be.Equal(t, d.encoding(strings.NewReader("\xc4\xc4\xc4\n")), cp437)
	be.Equal(t, d.encoding(strings.NewReader("caf\xe9\n")), latin1)
}

func TestAddText8bitUTF8(t *testing.T) {
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

func TestScreenshot(t *testing.T) {
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
	be.Err(t, helper.Touch(name), nil)
	d.Preview = dir.Directory(temp)
	be.True(t, d.screenshot())
}

func TestDecode(t *testing.T) {
	t.Parallel()

	got, err := decode(nil)
	be.Err(t, err, nil)
	be.Equal(t, got, "")

	got, err = decode(strings.NewReader("hello world"))
	be.Err(t, err, nil)
	be.Equal(t, got, "hello world\n")
}

func TestFirstLead(t *testing.T) {
	t.Parallel()

	be.Equal(t, string(firstLead(nil)), "")

	art := testutil.NewModel(t)
	got := firstLead(art)
	be.Equal(t, string(got), `<span class="bg-body-tertiary border rounded px-2 py-0 font-monospace fs-6 text-body-secondary">`+
		`filename.txt</span>`)

	art.GroupBrandBy.String = "NOCHANGE BBS"
	art.GroupBrandBy.Valid = true
	got = firstLead(art)
	be.True(t, strings.Contains(string(got), `d-inline-flex flex-wrap align-items-center gap-2`))
	be.True(t, strings.Contains(string(got), `NOCHANGE BBS`))
	be.True(t, strings.Contains(string(got), `filename.txt`))
	be.True(t, !strings.Contains(string(got), `<br>`))

	// Releaser-only without filename
	onlyRel := &models.File{GroupBrandBy: null.StringFrom("Razor 1911")}
	be.True(t, strings.Contains(string(firstLead(onlyRel)), "Razor 1911"))
}

func TestLockWidth(t *testing.T) {
	t.Parallel()

	be.Equal(t, lockWidth(3, []byte("abcdef")), []byte("\nabc\ndef"))
	be.Equal(t, lockWidth(2, []byte("abcdef")), []byte("\nab\ncd\nef"))
	be.Equal(t, lockWidth(1, []byte("abcdef")), []byte("\na\nb\nc\nd\ne\nf"))
	be.Equal(t, lockWidth(0, []byte("abcdef")), []byte("abcdef"))
	// 3+ tabs return unmodified
	tabbed := []byte("\t\t\tabcdefghij")
	be.Equal(t, lockWidth(3, tabbed), tabbed)
}

func TestPlainText(t *testing.T) {
	t.Parallel()

	be.True(t, !plainText(123))
	be.True(t, !plainText("unknown"))
	be.True(t, plainText(magicnumber.PlainText.Title()))
	be.True(t, plainText(magicnumber.ANSIEscapeText.Title()))
	be.True(t, plainText(magicnumber.ANSIEscapeText.Title()+simple.ANSIBlink))
	be.True(t, plainText(simple.ANSImation))
}

func TestRadio(t *testing.T) {
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

func TestErrorWithIDAndBadRequest(t *testing.T) {
	t.Parallel()

	be.Err(t, errorWithID(nil, "k", 1), nil)
	be.Err(t, errorWithID(errSampleApp, "k1", int64(42)), errSampleApp)
	be.True(t, strings.Contains(errorWithID(errSampleApp, "k2", "str").Error(), "k2 (str)"))

	be.Err(t, badRequest(nil, errSampleApp))
	be.Err(t, badRequest(echoCtx(t), nil), nil)
	be.Err(t, badRequest(echoCtx(t), errSampleApp), nil)

	discard(errSampleApp)
	logID(logs.Discard(), "msg", "uri", 1, errSampleApp)
}

func TestRequireReplacementZip(t *testing.T) {
	t.Parallel()

	be.True(t, !requireReplacementZip(""))
	be.True(t, !requireReplacementZip(filepath.Join(t.TempDir(), "missing.zip")))

	// Standard Deflate zip does not require replacement.
	zipPath := filepath.Join(t.TempDir(), "ok.zip")
	f, err := os.Create(zipPath)
	be.Err(t, err, nil)
	zw := zip.NewWriter(f)
	w, err := zw.Create("a.txt")
	be.Err(t, err, nil)
	_, err = w.Write([]byte("hello"))
	be.Err(t, err, nil)
	be.Err(t, zw.Close(), nil)
	be.Err(t, f.Close(), nil)

	be.True(t, !requireReplacementZip(zipPath))
}

func TestParseVals(t *testing.T) {
	t.Parallel()

	// parseValI across all supported types
	for _, v := range []any{
		int(1), int8(2), int16(3), int32(4), int64(5),
		uint(6), uint8(7), uint16(8), uint32(9), uint64(10),
		null.IntFrom(11), null.Int16From(12), null.Int32From(13), null.Int64From(14), null.ByteFrom(15),
		sql.NullInt16{Int16: 16, Valid: true},
		sql.NullInt32{Int32: 17, Valid: true},
		sql.NullInt64{Int64: 18, Valid: true},
		sql.NullByte{Byte: 19, Valid: true},
	} {
		n, ok := parseValI(v)
		be.True(t, ok)
		be.True(t, n > 0)
	}
	for _, invalid := range []any{
		-1, uint64(math.MaxUint64), "str",
		null.Int{},
		null.Int16{},
		null.Int32{},
		null.Int64{},
		null.Byte{},
		sql.NullInt16{},
		sql.NullInt32{},
		sql.NullInt64{},
		sql.NullByte{},
	} {
		_, ok := parseValI(invalid)
		be.True(t, !ok)
	}

	// parseValI64 across all supported types
	for _, v := range []any{
		int(1), int8(2), int16(3), int32(4), int64(5),
		uint(6), uint8(7), uint16(8), uint32(9), uint64(10),
	} {
		n, ok := parseValI64(v)
		be.True(t, ok)
		be.True(t, n > 0)
	}
	_, ok := parseValI64(-1)
	be.True(t, !ok)
	_, ok = parseValI64(uint64(math.MaxUint64))
	be.True(t, !ok)
	_, ok = parseValI64("str")
	be.True(t, !ok)

	// Misc helpers
	be.Equal(t, releasersHrefs(nil), "")
	be.Equal(t, scenerBucketKey(""), "#")
	be.Equal(t, formatAreaCodes(nil), "n/a")
}
