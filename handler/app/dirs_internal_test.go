package app

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Defacto2/helper"
	"github.com/Defacto2/server/internal/dir"
	"github.com/labstack/echo/v5"
	"github.com/nalgeon/be"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"
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

func Test_artifact404(t *testing.T) {
	t.Parallel()

	got := artifact404(nil, echoCtx(t), "")
	be.Err(t, got)
}

func Test_dirs_editorContent(t *testing.T) {
	t.Parallel()

	d := Dirs{}
	x := d.addEditor(t.Context(), nil, echoCtx(t), nil, nil)
	be.True(t, len(x) == 0)
}

var (
	latin1 encoding.Encoding = charmap.ISO8859_1
	cp437  encoding.Encoding = charmap.CodePage437
)

func TestEncoder(t *testing.T) {
	t.Parallel()

	d := Dirs{}
	got := d.encoding(nil)
	be.Equal(t, got, cp437)
}

func TestEncoder_amiga(t *testing.T) {
	t.Parallel()

	d := Dirs{Platform: "textamiga"}
	got := d.encoding(nil)
	be.Equal(t, got, latin1)
}

func TestEncoder_section(t *testing.T) {
	t.Parallel()

	d := Dirs{Section: "appleii"}
	got := d.encoding(nil)
	be.Equal(t, got, latin1)

	d.Section = "atarist"
	got = d.encoding(nil)
	be.Equal(t, got, latin1)
}

func TestEncoder_textdos(t *testing.T) {
	t.Parallel()

	d := Dirs{Platform: "textdos"}
	r := strings.NewReader("Hello\nworld\nthis is some text.\n")
	got := d.encoding(r)
	be.Equal(t, got, latin1)
}

func TestEncoder_textutf8(t *testing.T) {
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
