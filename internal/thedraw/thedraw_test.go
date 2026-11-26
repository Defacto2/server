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

// Layout of the TheDraw COM file header, which mirrors the format in thedraw.go.
const (
	rleStubLen   = 0xF0 // 240 byte stub with the RLE decoder
	rawStubLen   = 0xB0 // 176 byte stub with the raw VRAM copy
	rowsOff      = 4    // raw: row count, 1 byte
	payloadOff   = 4    // rle: payload length, 2 bytes
	colsOff      = 5    // raw: column count, 2 bytes
	startOff     = 7    // rle: start offset into the 80 column screen, 2 bytes
	sigOff       = 9
	leaOff       = 0x79 // location of the raw stub's load instruction
	wordBytes    = 2
	signature    = "TheDraw COM file Screen Save\x1A"
	rawLEA       = "\x8D\x36\xB0\x01" // LEA SI, [01B0h]
	jumpOverStub = "\xEB\x3D\x90"
)

func buildTheDrawCOM(t *testing.T, payload []byte) []byte {
	t.Helper()
	n := len(payload)
	if n > math.MaxUint16 {
		t.Fatalf("payload too large: %d", n)
	}
	out := make([]byte, rleStubLen+n)
	copy(out, jumpOverStub)
	binary.LittleEndian.PutUint16(out[payloadOff:payloadOff+wordBytes], uint16(n)) //nolint:gosec // G115: n is checked above
	copy(out[sigOff:], signature)
	copy(out[rleStubLen:], payload)
	return out
}

// buildTheDrawRaw returns an uncompressed (176 byte stub) TheDraw COM file.
// The cells are character and attribute byte pairs, which are not checked against rows and cols.
func buildTheDrawRaw(rows byte, cols uint16, cells []byte) []byte {
	out := make([]byte, rawStubLen+len(cells))
	copy(out, jumpOverStub)
	out[rowsOff] = rows
	binary.LittleEndian.PutUint16(out[colsOff:colsOff+wordBytes], cols)
	copy(out[sigOff:], signature)
	copy(out[leaOff:], rawLEA)
	copy(out[rawStubLen:], cells)
	return out
}

func TestIsCOM(t *testing.T) {
	t.Parallel()

	rle := func() []byte { return buildTheDrawCOM(t, []byte{0x0F, 'H', 'i', 0x18}) }
	badJump := rle()
	badJump[0] = 0x00
	badSig := rle()
	badSig[9] = 'X'
	truncated := rle()
	truncated = truncated[:len(truncated)-1]

	const row = 80 * 2
	rows2 := make([]byte, 2*row)
	tests := []struct {
		name string
		data []byte
		want bool
	}{
		{"valid rle", rle(), true},
		{"valid raw", buildTheDrawRaw(2, 80, rows2), true},
		{"nil", nil, false},
		{"text", []byte("Hello world\r\n"), false},
		{"truncated rle", truncated, false},
		{"wrong jump", badJump, false},
		{"wrong signature", badSig, false},
		{"zero payload", buildTheDrawCOM(t, nil), false},
		{"raw zero rows", buildTheDrawRaw(0, 80, rows2), false},
		{"raw too many columns", buildTheDrawRaw(1, 161, make([]byte, 161*2)), false},
		{"raw truncated", buildTheDrawRaw(2, 80, rows2[:row]), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			be.Equal(t, thedraw.IsCOM(tt.data), tt.want)
		})
	}
}

func TestExtractRLE(t *testing.T) {
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

	// Row 0: bright cyan "HI", red background "XXX", then a reset before the line break.
	// Row 1: bold, blinking cyan on red, four spaces and "!". The blank trailing row is trimmed.
	want := "\x1b[0;1;36mHI\x1b[41mXXX\x1b[0m\r\n" +
		"\x1b[1;5;36;41m    !\x1b[0m"
	be.Equal(t, string(got), want)

	blink, err := ansimate.DetectBlink(bytes.NewReader(got))
	be.Err(t, err, nil)
	be.True(t, blink)
}

func TestExtractFullRow(t *testing.T) {
	t.Parallel()

	// 80 '#' chars + newline + "ROW2" -> should not insert an extra blank line between row 0 and row 1.
	payload := []byte{0x0F, 0x1A, 0x4F, '#', 0x18, 'R', 'O', 'W', '2'}
	raw := buildTheDrawCOM(t, payload)

	got, err := thedraw.Extract(raw)
	be.Err(t, err, nil)

	// A full 80 column row wraps by itself, so no CRLF is added.
	want := "\x1b[0;1m" + strings.Repeat("#", 80) + "ROW2\x1b[0m"
	be.Equal(t, string(got), want)
}

func TestDecode(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "TEST.COM")
	raw := buildTheDrawCOM(t, []byte{0x0E, 'O', 'K'})
	be.Err(t, os.WriteFile(path, raw, 0o600), nil)

	got, err := thedraw.Decode(path)
	be.Err(t, err, nil)
	want, err := thedraw.Extract(raw)
	be.Err(t, err, nil)
	be.Equal(t, string(got), string(want))
	be.Equal(t, string(got), "\x1b[0;1;33mOK\x1b[0m")
}

func TestExtractRaw(t *testing.T) {
	t.Parallel()

	const (
		rows       = 2
		cols       = 80
		attrWhite  = 0x0F // bright white on black
		attrCyan   = 0x0B
		attrBlueBG = 0x1F
	)
	cells := make([]byte, 0, rows*cols*2)
	for range rows * cols {
		cells = append(cells, ' ', attrWhite)
	}
	copy(cells, []byte{'S', attrCyan, 'U', attrCyan})
	copy(cells[cols*2:], []byte{'R', attrBlueBG, 'G', attrBlueBG})
	raw := buildTheDrawRaw(rows, cols, cells)

	be.True(t, thedraw.IsCOM(raw))
	got, err := thedraw.Extract(raw)
	be.Err(t, err, nil)

	// Row 0: bold cyan "SU". Row 1: white on blue "RG". Trailing blanks are trimmed.
	want := "\x1b[0;1;36mSU\r\n" +
		"\x1b[37;44mRG\x1b[0m"
	be.Equal(t, string(got), want)
}

func TestExtractSanitizes(t *testing.T) {
	t.Parallel()

	const (
		attr = 0x07
		cols = 6
	)
	// Control bytes that could break lines or inject terminal escape sequences.
	chars := [cols]byte{0x00, '\r', '\n', 0x1A, 0x1B, 'A'}
	cells := make([]byte, 0, len(chars)*2)
	for _, ch := range chars {
		cells = append(cells, ch, attr)
	}
	raw := buildTheDrawRaw(1, cols, cells)

	got, err := thedraw.Extract(raw)
	be.Err(t, err, nil)
	be.Equal(t, string(got), "\x1b[0m     A")
	be.Equal(t, bytes.Count(got, []byte{0x1B}), 1) // only the SGR reset
}

func TestDecodeErrors(t *testing.T) {
	t.Parallel()

	t.Run("missing file", func(t *testing.T) {
		t.Parallel()

		_, err := thedraw.Decode(filepath.Join(t.TempDir(), "MISSING.COM"))
		be.Err(t, err, os.ErrNotExist)
	})

	t.Run("too large", func(t *testing.T) {
		t.Parallel()

		// A valid COM that is only rejected for being larger than the 64KB limit.
		const payloadLen = 65537 - rleStubLen // file is exactly one byte over the limit, so truncation keeps it valid
		raw := buildTheDrawCOM(t, bytes.Repeat([]byte{'A'}, payloadLen))
		be.True(t, thedraw.IsCOM(raw))
		be.Equal(t, len(raw), 65537)

		path := filepath.Join(t.TempDir(), "BIG.COM")
		be.Err(t, os.WriteFile(path, raw, 0o600), nil)
		_, err := thedraw.Decode(path)
		be.Err(t, err, thedraw.ErrNotCOM)
	})

	t.Run("not a COM", func(t *testing.T) {
		t.Parallel()

		path := filepath.Join(t.TempDir(), "TEXT.COM")
		be.Err(t, os.WriteFile(path, []byte("Hello world\r\n"), 0o600), nil)
		_, err := thedraw.Decode(path)
		be.Err(t, err, thedraw.ErrNotCOM)
	})
}

func TestExtractStartOffset(t *testing.T) {
	t.Parallel()

	const rowBytes = 80 * 2 // bytes per row of 80 cells
	tests := []struct {
		name    string
		start   uint16
		payload []byte
		want    string
	}{
		// Offset 6 is row 0, column 3. Cells before the cursor have no attribute set,
		// so they are drawn as black.
		{"column", 6, []byte{0x07, 'A', 'B'}, "\x1b[0;30m   \x1b[0mAB"},
		// Rows before the cursor are empty lines.
		{"row", 2*rowBytes + 4, []byte{0x07, 'X'}, "\r\n\r\n\x1b[0;30m  \x1b[0mX"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			raw := buildTheDrawCOM(t, tt.payload)
			binary.LittleEndian.PutUint16(raw[startOff:startOff+wordBytes], tt.start)
			got, err := thedraw.Extract(raw)
			be.Err(t, err, nil)
			be.Equal(t, string(got), tt.want)
		})
	}
}
