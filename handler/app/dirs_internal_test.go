package app

import (
	"bytes"
	"encoding/binary"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Defacto2/server/internal/dir"
	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/aarondl/null/v8"
	"github.com/nalgeon/be"
)

func TestAddTextBinaryTheDrawCOM(t *testing.T) {
	t.Parallel()

	const stubLen = 0xF0
	payload := []byte{0x0F, 'H', 'I', 0x18}
	td := make([]byte, stubLen+len(payload))
	copy(td[0:3], []byte{0xEB, 0x3D, 0x90})
	binary.LittleEndian.PutUint16(td[4:6], uint16(len(payload)))
	copy(td[9:], []byte("TheDraw COM file Screen Save\x1A"))
	copy(td[stubLen:], payload)

	dlDir := t.TempDir()
	const unid = "b13e2a00-0000-0000-0000-000000000001"
	be.Err(t, os.WriteFile(filepath.Join(dlDir, unid), td, 0o600), nil)

	ds := Dirs{Download: dir.Directory(dlDir)}
	art := &models.File{
		UUID:           null.StringFrom(unid),
		DateIssuedYear: null.Int16From(1992),
	}
	data := map[string]any{
		"magic":    "Binary data or binary text",
		"platform": "dos",
	}
	elems := []string{"reader-invert", "border"}
	buf := bytes.NewBufferString("<span>HI</span>")

	got := ds.addTextBinary(art, buf, elems, data)
	be.Equal(t, got["contentBinarySwappers"], true)

	classes, ok := got["preElementClass"].(string)
	be.True(t, ok)
	be.True(t, strings.Contains(classes, "font-dos"))
	be.True(t, strings.Contains(classes, "render"))
	be.True(t, !strings.Contains(classes, "text-bg-dark"))
}
