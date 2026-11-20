package thedraw_test

import (
	"bytes"
	"encoding/binary"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Defacto2/server/internal/thedraw"
	"github.com/bengarrett/ansimate"
	"github.com/nalgeon/be"
)

func buildTheDrawCOM(t *testing.T, payload []byte) []byte {
	t.Helper()
	const (
		stubLen    = 0xF0
		payloadOff = 4
		sigOff     = 9
		wordBytes  = 2
	)
	n := len(payload)
	if n < 0 || n > math.MaxUint16 {
		t.Fatalf("payload too large: %d", n)
		return nil
	}
	out := make([]byte, stubLen+n)
	copy(out[0:3], []byte{0xEB, 0x3D, 0x90})
	binary.LittleEndian.PutUint16(out[payloadOff:payloadOff+wordBytes], uint16(n))
	copy(out[sigOff:], []byte("TheDraw COM file Screen Save\x1A"))
	copy(out[stubLen:], payload)
	return out
}

func TestIsTheDrawCOM(t *testing.T) {
	t.Parallel()

	valid := buildTheDrawCOM(t, []byte{0x0F, 'H', 'i', 0x18})
	be.True(t, thedraw.IsCOM(valid))

	be.True(t, !thedraw.IsCOM(nil))
	be.True(t, !thedraw.IsCOM([]byte("Hello world\r\n")))

	truncated := valid[:len(valid)-1]
	be.True(t, !thedraw.IsCOM(truncated))
}

func TestExtractTheDrawANS(t *testing.T) {
	t.Parallel()

	// FG bright cyan (0x0B), "HI", BG red (0x14 -> bg=4), repeat 'X' 3 times (0x1A 0x02 'X'),
	// newline (0x18), toggle blink (0x1B), repeat space 4 times (0x19 0x03), '!', newline, trailing blank row.
	// Trailing invisible row: toggle blink off (0x1B), BG black (0x10), then 80 spaces.
	payload := []byte{
		0x0B, 'H', 'I',
		0x14, 0x1A, 0x02, 'X',
		0x18,
		0x1B, 0x19, 0x03, '!',
		0x18,
		0x1B, 0x10, 0x19, 0x4F, // full 80-space invisible trailing row
	}
	raw := buildTheDrawCOM(t, payload)

	got, err := thedraw.Extract(raw)
	be.Err(t, err, nil)

	s := string(got)
	be.True(t, strings.Contains(s, "HI"))
	be.True(t, strings.Contains(s, "XXX"))
	be.True(t, strings.Contains(s, "    !"))
	// CGA fg 0x0B -> bright cyan (1;36), bg 0x04 -> red (41), blink -> 5.
	be.True(t, strings.Contains(s, "\x1b[0;1;36m"))
	be.True(t, strings.Contains(s, "\x1b[41m"))
	be.True(t, strings.Contains(s, ";5;"))

	// Trailing blank 80-space row should be trimmed (only 1 CRLF between line 0 and line 1).
	be.Equal(t, strings.Count(s, "\r\n"), 1)

	blink, err := ansimate.DetectBlink(bytes.NewReader(got))
	be.Err(t, err, nil)
	be.True(t, blink)
}

func TestExtractTheDrawANSFullRowWrap(t *testing.T) {
	t.Parallel()

	// 80 '#' chars + newline + "ROW2" -> should not insert an extra blank line between row 0 and row 1.
	payload := []byte{0x0F, 0x1A, 0x4F, '#', 0x18, 'R', 'O', 'W', '2'}
	raw := buildTheDrawCOM(t, payload)

	got, err := thedraw.Extract(raw)
	be.Err(t, err, nil)

	s := string(got)
	be.True(t, strings.Contains(s, strings.Repeat("#", 80)+"ROW2"))
	be.Equal(t, strings.Count(s, "\r\n"), 0)
}

func TestDecodeTheDrawFile(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "TEST.COM")
	raw := buildTheDrawCOM(t, []byte{0x0E, 'O', 'K'})
	be.Err(t, os.WriteFile(path, raw, 0o600), nil)

	got, err := thedraw.Decode(path)
	be.Err(t, err, nil)
	be.True(t, bytes.Contains(got, []byte("OK")))
}

func TestExtractTheDrawANSUncompressed176(t *testing.T) {
	t.Parallel()

	const (
		stubLen    = 0xB0
		rows       = 2
		cols       = 80
		cellBytes  = 2
		attrWhite  = 0x0F
		attrCyan   = 0x0B
		attrBlueBG = 0x1F
	)
	raw := make([]byte, stubLen+rows*cols*cellBytes)
	copy(raw[0:3], []byte{0xEB, 0x3D, 0x90})
	raw[4] = rows
	binary.LittleEndian.PutUint16(raw[5:7], cols)
	copy(raw[9:], []byte("TheDraw COM file Screen Save\x1A"))
	// Mimic the LEA SI, [01B0h] instruction inside the 176-byte stub.
	copy(raw[0x79:], []byte{0x8D, 0x36, 0xB0, 0x01})

	for i := stubLen; i+1 < len(raw); i += cellBytes {
		raw[i] = ' '
		raw[i+1] = attrWhite
	}
	copy(raw[stubLen:], []byte{'S', attrCyan, 'U', attrCyan})
	row1 := stubLen + cols*cellBytes
	copy(raw[row1:], []byte{'R', attrBlueBG, 'G', attrBlueBG})

	be.True(t, thedraw.IsCOM(raw))
	got, err := thedraw.Extract(raw)
	be.Err(t, err, nil)

	s := string(got)
	be.True(t, strings.Contains(s, "SU"))
	be.True(t, strings.Contains(s, "RG"))
	be.True(t, strings.Contains(s, "\x1b[0;1;36m"))
	be.True(t, strings.Contains(s, ";44m"))
}
