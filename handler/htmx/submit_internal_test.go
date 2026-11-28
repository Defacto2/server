package htmx

import (
	"archive/zip"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Defacto2/server/handler/pouet"
	"github.com/Defacto2/server/internal/logs"
	"github.com/Defacto2/server/internal/testutil"
	"github.com/nalgeon/be"
)

var errSubBoom = errors.New("submit helper failure")

func TestSubmitErrHelpers(t *testing.T) {
	t.Parallel()

	sl := logs.Discard()

	logFormErr(nil, "msg", "form", "name", errSubBoom)
	logFormErr(sl, "msg", "form", "name", errSubBoom)

	be.Err(t, errUpload(nil, nil))
	be.Err(t, errUpload(nil, errSubBoom), errSubBoom)

	c := testutil.NewContext(t, "/")
	be.Err(t, errUpload(c, nil), nil)
	rec := internalRecorder(t, c)
	be.Equal(t, rec.Code, http.StatusInternalServerError)
	be.Equal(t, rec.Body.String(), "unexpected error")

	c = testutil.NewContext(t, "/")
	be.Err(t, errUpload(c, errSubBoom), nil)
	rec = internalRecorder(t, c)
	be.Equal(t, rec.Code, http.StatusInternalServerError)
	be.True(t, strings.Contains(rec.Body.String(), "cannot save"))

	be.Err(t, errFormHeader(nil, nil, "f", errSubBoom))
	c = testutil.NewContext(t, "/")
	be.Err(t, errFormHeader(sl, c, "f", errSubBoom), nil)
	be.Equal(t, internalRecorder(t, c).Code, http.StatusBadRequest)

	be.Err(t, errMultipartFile(nil, nil, "f", errSubBoom))
	c = testutil.NewContext(t, "/")
	be.Err(t, errMultipartFile(sl, c, "f", errSubBoom), nil)
	be.Equal(t, internalRecorder(t, c).Code, http.StatusBadRequest)

	be.Err(t, errSHA384(nil, nil, "f", errSubBoom))
	c = testutil.NewContext(t, "/")
	be.Err(t, errSHA384(sl, c, "f", errSubBoom), nil)
	be.Equal(t, internalRecorder(t, c).Code, http.StatusInternalServerError)

	be.Err(t, errExistSHA(nil, nil, errSubBoom))
	c = testutil.NewContext(t, "/")
	be.Err(t, errExistSHA(sl, c, errSubBoom), nil)
	be.Equal(t, internalRecorder(t, c).Code, http.StatusServiceUnavailable)
}

func TestTransferContent(t *testing.T) {
	t.Parallel()

	sl := logs.Discard()
	tr := Transfer{}

	// Non-archive returns nil/empty slice without failing.
	got := tr.content(t.Context(), sl, "test", "missing.zip", filepath.Join(t.TempDir(), "missing.zip"))
	be.Equal(t, len(got), 0)

	// Valid zip returns entry list.
	zipPath := filepath.Join(t.TempDir(), "sample.zip")
	f, err := os.Create(zipPath)
	be.Err(t, err, nil)
	zw := zip.NewWriter(f)
	w, err := zw.Create("README.NFO")
	be.Err(t, err, nil)
	_, err = w.Write([]byte("scene nfo"))
	be.Err(t, err, nil)
	be.Err(t, zw.Close(), nil)
	be.Err(t, f.Close(), nil)

	got = tr.content(t.Context(), sl, "test", "sample.zip", zipPath)
	be.Equal(t, len(got), 1)
	be.Equal(t, got[0], "README.NFO")
}

func TestProdInternal(t *testing.T) {
	t.Parallel()

	c := testutil.NewContext(t, "/")
	htm := Demozoo.submitHTML(c, 123, 1)
	be.True(t, strings.Contains(htm, "Demozoo production, 123"))

	htm = Pouet.submitHTML(c, 456, 1)
	be.True(t, strings.Contains(htm, "Pouet production, 456"))
	be.Equal(t, Prod(99).title(), "")

	// Unknown Prod returns zero values from exists/insert.
	ok, err := Prod(99).exists(t.Context(), nil, 1)
	be.Err(t, err, nil)
	be.Equal(t, ok, false)

	id, unid, err := Prod(99).insert(t.Context(), nil, 1)
	be.Err(t, err, nil)
	be.Equal(t, id, int64(0))
	be.Equal(t, unid, "")

	// prodLink direct download & empty links.
	var empty pouet.Response
	be.Equal(t, prodLink(empty), "")

	var direct pouet.Response
	direct.Prod.Download = "https://files.scene.org/demo.zip"
	be.Equal(t, prodLink(direct), "https://files.scene.org/demo.zip")

	// Skips empty, YouTube, SourceForge, and GitHub links and returns the first valid mirror.
	var fallback pouet.Response
	fallback.Prod.DownloadLinks = []struct {
		Type string `json:"type"`
		Link string `json:"link"`
	}{
		{Link: ""},
		{Link: "   "},
		{Link: "https://www.YouTube.com/watch?v=abc"},
		{Link: "https://sourceforge.net/projects/demo"},
		{Link: "https://github.com/org/demo"},
		{Link: "https://files.scene.org/mirror.zip"},
	}
	be.Equal(t, prodLink(fallback), "https://files.scene.org/mirror.zip")

	// All unwanted links return "".
	fallback.Prod.DownloadLinks = fallback.Prod.DownloadLinks[:5]
	be.Equal(t, prodLink(fallback), "")
}
