package htmx

import (
	"context"
	"database/sql"
	"encoding/binary"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Defacto2/server/internal/testutil"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/labstack/echo/v5"
	"github.com/nalgeon/be"
)

var (
	errBadReqBoom = errors.New("bad input")
	errLinkFail   = errors.New("db write failed")
)

func internalRecorder(t *testing.T, c *echo.Context) *httptest.ResponseRecorder {
	t.Helper()

	resp, err := echo.UnwrapResponse(c.Response())
	be.Err(t, err, nil)

	rec, ok := resp.Unwrap().(*httptest.ResponseRecorder)
	be.True(t, ok)

	return rec
}

func TestBadRequest(t *testing.T) {
	t.Parallel()

	be.Err(t, badRequest(nil, nil))
	be.Err(t, badRequest(nil, errBadReqBoom), errBadReqBoom)

	c := testutil.NewContext(t, "")
	be.Err(t, badRequest(c, nil), nil)
	rec := internalRecorder(t, c)
	be.Equal(t, rec.Code, http.StatusBadRequest)
	be.Equal(t, rec.Body.String(), "something went wrong")

	c = testutil.NewContext(t, "")
	be.Err(t, badRequest(c, errBadReqBoom), nil)
	rec = internalRecorder(t, c)
	be.Equal(t, rec.Code, http.StatusBadRequest)
	be.Equal(t, rec.Body.String(), errBadReqBoom.Error())
}

func TestCGADump(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	be.Equal(t, cgaDump(filepath.Join(tmp, "missing.bin")), false)

	txt := filepath.Join(tmp, "plain.txt")
	be.Err(t, os.WriteFile(txt, []byte("plain text file\n"), 0o600), nil)
	be.Equal(t, cgaDump(txt), false)

	dump := make([]byte, 4096)
	for i := 0; i+1 < len(dump); i += 2 {
		dump[i] = ' '
		dump[i+1] = 0x07
	}
	copy(dump[0:], []byte{'A', 0x0B, 'B', 0x0B})
	cga := filepath.Join(tmp, "screen.tpg")
	be.Err(t, os.WriteFile(cga, dump, 0o600), nil)
	be.Equal(t, cgaDump(cga), true)
}

func TestDuplicate(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()

	// Missing source returns error.
	be.Err(t, duplicate(filepath.Join(tmp, "missing"), filepath.Join(tmp, "out")))

	// Regular file is copied directly.
	src := filepath.Join(tmp, "plain.txt")
	dst := filepath.Join(tmp, "plain.out")
	be.Err(t, os.WriteFile(src, []byte("hello scene"), 0o600), nil)
	be.Err(t, duplicate(src, dst), nil)
	got, err := os.ReadFile(dst)
	be.Err(t, err, nil)
	be.Equal(t, string(got), "hello scene")

	// TheDraw COM screen save is decoded to ANSI bytes.
	const stubLen = 0xF0
	payload := []byte{0x0F, 'H', 'I', 0x18}
	td := make([]byte, stubLen+len(payload))
	copy(td[0:3], []byte{0xEB, 0x3D, 0x90})
	binary.LittleEndian.PutUint16(td[4:6], uint16(len(payload)))
	copy(td[9:], []byte("TheDraw COM file Screen Save\x1A"))
	copy(td[stubLen:], payload)

	tdSrc := filepath.Join(tmp, "screen.com")
	tdDst := filepath.Join(tmp, "screen.txt")
	be.Err(t, os.WriteFile(tdSrc, td, 0o600), nil)
	be.Err(t, duplicate(tdSrc, tdDst), nil)
	decoded, err := os.ReadFile(tdDst)
	be.Err(t, err, nil)
	be.True(t, len(decoded) > 0)
}

func TestCommitLink(t *testing.T) {
	t.Parallel()

	dummyTx := &sql.Tx{}
	failFn := func(context.Context, boil.ContextExecutor, int64, string) error {
		return errLinkFail
	}

	// Nil context returns an error without panicking.
	be.Err(t, commitLink(nil, dummyTx, "artifact-editor-demozoo", nil, failFn))

	// Nil tx returns a single 400 response without appending link previews.
	c := testutil.NewInputs(t, "", testutil.Input{
		EditorKey:                 "1",
		"artifact-editor-youtube": testutil.YT,
	})
	be.Err(t, commitLink(c, nil, "artifact-editor-demozoo", nil, failFn), nil)
	rec := internalRecorder(t, c)
	be.Equal(t, rec.Code, http.StatusBadRequest)
	be.True(t, !strings.Contains(rec.Body.String(), "Link to"))

	// Invalid key returns 400.
	c = testutil.NewInput(t, "", EditorKey, "bad")
	be.Err(t, commitLink(c, dummyTx, "artifact-editor-demozoo", nil, failFn), nil)
	be.Equal(t, internalRecorder(t, c).Code, http.StatusBadRequest)

	// Callback error returns a single 400 without appending link previews.
	c = testutil.NewInputs(t, "", testutil.Input{
		EditorKey:                 "1",
		"artifact-editor-youtube": testutil.YT,
	})
	be.Err(t, commitLink(c, dummyTx, "artifact-editor-demozoo", nil, failFn), nil)
	rec = internalRecorder(t, c)
	be.Equal(t, rec.Code, http.StatusBadRequest)
	be.True(t, !strings.Contains(rec.Body.String(), "Link to"))

	// Success writes a single 200 text/html response with the green checkmark and link previews.
	var gotKey int64
	var gotVal string
	okFn := func(_ context.Context, _ boil.ContextExecutor, k int64, v string) error {
		gotKey = k
		gotVal = v
		return nil
	}
	c = testutil.NewInputs(t, "", testutil.Input{
		EditorKey:                 "7",
		"artifact-editor-github":  "  Defacto2/server  ",
		"artifact-editor-youtube": testutil.YT,
	})
	be.Err(t, commitLink(c, dummyTx, "artifact-editor-github", strings.TrimSpace, okFn), nil)
	rec = internalRecorder(t, c)
	be.Equal(t, rec.Code, http.StatusOK)
	be.True(t, strings.HasPrefix(rec.Header().Get(echo.HeaderContentType), echo.MIMETextHTML))
	be.True(t, strings.HasPrefix(rec.Body.String(), successSpan))
	be.True(t, strings.Contains(rec.Body.String(), "Link to"))
	be.Equal(t, gotKey, int64(7))
	be.Equal(t, gotVal, "Defacto2/server")
}
