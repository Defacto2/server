//nolint:testpackage
package readme

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Defacto2/server/internal/dir"
	"github.com/Defacto2/server/internal/logs"
	"github.com/nalgeon/be"
)

func Test_hasANSI(t *testing.T) {
	t.Parallel()

	got, err := hasANSI(nil)
	be.Err(t, err, nil)
	be.True(t, !got)

	r := strings.NewReader("a\x1b[1;cabc")
	got, err = hasANSI(r)
	be.Err(t, err, nil)
	be.True(t, got)

	r = strings.NewReader("a\x1b[Acabc")
	got, err = hasANSI(r)
	be.Err(t, err, nil)
	be.True(t, got)
}

func Test_removeControls(t *testing.T) {
	t.Parallel()

	b := []byte("a\x1b[1;cabc")
	got := removeControls(b)
	be.Equal(t, got, []byte("aabc"))
}

func Test_trimEOF(t *testing.T) {
	t.Parallel()

	wants := []byte("hello world")
	got := trimE(wants)
	be.Equal(t, got, wants)

	wants = []byte("requires._____")
	s := wants
	s = append(s, []byte("\x8a\x1a")...)
	got = trimE(s)
	be.Equal(t, got, wants)
}

func Test_viewer(t *testing.T) {
	t.Parallel()

	text := Text{}
	got := text.useViewer()
	be.True(t, !got)

	text.Platform = "textamiga"
	be.True(t, text.useViewer())
}

func Test_decodeUTF16(t *testing.T) {
	t.Parallel()

	le := []byte{0xff, 0xfe, 'H', 0x00, 'i', 0x00}
	got, err := decodeUTF16(le)
	be.Err(t, err, nil)
	be.Equal(t, got, []byte("Hi"))

	beBytes := []byte{0xfe, 0xff, 0x00, 'H', 0x00, 'i'}
	got, err = decodeUTF16(beBytes)
	be.Err(t, err, nil)
	be.Equal(t, got, []byte("Hi"))

	plain := []byte("Hi")
	got, err = decodeUTF16(plain)
	be.Err(t, err, nil)
	be.Equal(t, got, plain)
}

func Test_buffersTheDrawCOM(t *testing.T) {
	t.Parallel()

	const stubLen = 0xF0
	payload := []byte{0x0B, 'T', 'D', 0x18, 0x0E, 'B', 'B', 'S'}
	td := make([]byte, stubLen+len(payload))
	copy(td[0:3], []byte{0xEB, 0x3D, 0x90})
	binary.LittleEndian.PutUint16(td[4:6], uint16(len(payload)))
	copy(td[9:], []byte("TheDraw COM file Screen Save\x1A"))
	copy(td[stubLen:], payload)

	dlDir := t.TempDir()
	extraDir := t.TempDir()
	const unid = "b13e2a00-0000-0000-0000-000000000001"
	be.Err(t, os.WriteFile(filepath.Join(dlDir, unid), td, 0o600), nil)

	text := Text{
		Download: dir.Directory(dlDir),
		Extra:    dir.Directory(extraDir),
		UUID:     unid,
		Filename: "!README!.COM",
		Platform: "dos",
		MaxSize:  1024 * 1024,
	}
	textBuf, runeBuf, err := text.Buffers(logs.Discard())
	be.Err(t, err, nil)
	be.True(t, textBuf.Len() > 0)
	be.Equal(t, runeBuf.Len(), 0)
	s := textBuf.String()
	be.True(t, strings.Contains(s, "TD"))
	be.True(t, strings.Contains(s, "BBS"))
}
