//nolint:testpackage
package readme

import (
	"bytes"
	"encoding/binary"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Defacto2/server/internal/dir"
	"github.com/Defacto2/server/internal/logs"
	"github.com/nalgeon/be"
)

type errReader struct {
	err error
}

func (e errReader) Read([]byte) (int, error) {
	return 0, e.err
}

func TestHasANSI(t *testing.T) {
	t.Parallel()

	t.Run("nil reader", func(t *testing.T) {
		t.Parallel()

		got, err := hasANSI(nil)
		be.Err(t, err, nil)
		be.True(t, !got)
	})

	t.Run("plain text", func(t *testing.T) {
		t.Parallel()

		got, err := hasANSI(strings.NewReader("plain text without escape codes"))
		be.Err(t, err, nil)
		be.True(t, !got)
	})

	t.Run("csi sequence", func(t *testing.T) {
		t.Parallel()

		for _, s := range []string{"a\x1b[1;cabc", "a\x1b[Acabc"} {
			got, err := hasANSI(strings.NewReader(s))
			be.Err(t, err, nil)
			be.True(t, got)
		}
	})

	t.Run("chunk boundary carry", func(t *testing.T) {
		t.Parallel()

		const chunkSize = 8 * 1024
		buf := make([]byte, chunkSize+3)
		for i := range chunkSize - 1 {
			buf[i] = 'a'
		}
		buf[chunkSize-1] = '\x1b'
		buf[chunkSize] = '['
		buf[chunkSize+1] = '1'
		buf[chunkSize+2] = 'm'

		got, err := hasANSI(bytes.NewReader(buf))
		be.Err(t, err, nil)
		be.True(t, got)
	})

	t.Run("too long", func(t *testing.T) {
		t.Parallel()

		buf := bytes.Repeat([]byte{'a'}, 1024*1024+1)
		got, err := hasANSI(bytes.NewReader(buf))
		be.Err(t, err, ErrTooLong)
		be.True(t, !got)
	})

	t.Run("read error", func(t *testing.T) {
		t.Parallel()

		sentinel := errors.New("read failed")
		got, err := hasANSI(errReader{err: sentinel})
		be.Err(t, err, sentinel)
		be.True(t, !got)
	})
}

func TestRemoveControls(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input []byte
		want  []byte
	}{
		{"empty", []byte{}, []byte{}},
		{"ansi escape", []byte("a\x1b[1;cabc"), []byte("aabc")},
		{"dec and amiga", []byte("x\x1b[?7hy\x1b[1 pz"), []byte("xyz")},
		{"sauce marker", []byte("textSAUCE00"), []byte("text")},
		{"nul and crlf", []byte("a\x00b\r\nc\r\n"), []byte("a b\nc\n")},
		{"trailing space and eof", []byte("hello  \x1a\x1a"), []byte("hello")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := removeControls(tt.input)
			be.Equal(t, got, tt.want)
		})
	}
}

func TestTrimE(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input []byte
		want  []byte
	}{
		{"plain", []byte("hello world"), []byte("hello world")},
		{"egrave and eof", []byte("requires._____\x8a\x1a"), []byte("requires._____")},
		{"multiple eof", []byte("abc\x8a\x1a\x1a\x1a"), []byte("abc")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := trimE(tt.input)
			be.Equal(t, got, tt.want)
		})
	}
}

func TestUseViewer(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		text Text
		want bool
	}{
		{"empty", Text{}, false},
		{"file_id.diz", Text{Filename: " FILE_ID.DIZ "}, true},
		{"textamiga", Text{Platform: "textamiga"}, true},
		{"pcb bbs", Text{Platform: "pcb", Section: "bbs"}, true},
		{"pcb tool", Text{Platform: "pcb", Section: "tool"}, false},
		{"pcb package", Text{Platform: "pcb", Section: "package"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			be.Equal(t, tt.text.useViewer(), tt.want)
		})
	}
}

func TestDecodeUTF16(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input []byte
		want  []byte
	}{
		{"empty", []byte{}, []byte{}},
		{"single byte", []byte{'A'}, []byte{'A'}},
		{"utf16 le", []byte{0xff, 0xfe, 'H', 0x00, 'i', 0x00}, []byte("Hi")},
		{"utf16 be", []byte{0xfe, 0xff, 0x00, 'H', 0x00, 'i'}, []byte("Hi")},
		{"plain", []byte("Hi"), []byte("Hi")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := decodeUTF16(tt.input)
			be.Err(t, err, nil)
			be.Equal(t, got, tt.want)
		})
	}
}

func TestTrimBSAVE(t *testing.T) {
	t.Parallel()

	makeHeader := func(magic byte, seg, length uint16, payload []byte) []byte {
		b := make([]byte, 7+len(payload))
		b[0] = magic
		binary.LittleEndian.PutUint16(b[1:3], seg)
		binary.LittleEndian.PutUint16(b[5:7], length)
		copy(b[7:], payload)
		return b
	}

	payload := []byte{'A', 0x07, 'B', 0x0F}
	tests := []struct {
		name  string
		input []byte
		want  []byte
	}{
		{"short", []byte{0xFD, 0x00}, []byte{0xFD, 0x00}},
		{"wrong magic", makeHeader(0x00, 0xB800, 4, payload), makeHeader(0x00, 0xB800, 4, payload)},
		{"wrong segment", makeHeader(0xFD, 0xA000, 4, payload), makeHeader(0xFD, 0xA000, 4, payload)},
		{"zero length", makeHeader(0xFD, 0xB800, 0, payload), makeHeader(0xFD, 0xB800, 0, payload)},
		{"truncated", makeHeader(0xFD, 0xB800, 10, payload), makeHeader(0xFD, 0xB800, 10, payload)},
		{"valid cga", makeHeader(0xFD, 0xB800, 4, payload), payload},
		{"valid mda", makeHeader(0xFD, 0xB000, 4, payload), payload},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := trimBSAVE(tt.input)
			be.Equal(t, got, tt.want)
		})
	}
}

func TestBuffersTheDrawCOM(t *testing.T) {
	t.Parallel()

	const stubLen = 0xF0
	const payloadLen = 8
	payload := []byte{0x0B, 'T', 'D', 0x18, 0x0E, 'B', 'B', 'S'}
	be.Equal(t, len(payload), payloadLen)
	td := make([]byte, stubLen+payloadLen)
	copy(td[0:3], []byte{0xEB, 0x3D, 0x90})
	binary.LittleEndian.PutUint16(td[4:6], payloadLen)
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
