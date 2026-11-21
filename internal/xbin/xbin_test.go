package xbin_test

import (
	"encoding/binary"
	"math"
	"strings"
	"testing"

	"github.com/Defacto2/server/internal/xbin"
	"github.com/nalgeon/be"
)

func buildTestXBIN(w, h, fontH, flags byte, pal, font, body []byte) []byte {
	hdr := []byte{
		'X', 'B', 'I', 'N', 0x1A,
		w, 0x00,
		h, 0x00,
		fontH,
		flags,
	}
	out := append([]byte{}, hdr...)
	if (flags & 0x01) != 0 {
		out = append(out, pal...)
	}
	if (flags & 0x02) != 0 {
		out = append(out, font...)
	}
	return append(out, body...)
}

func TestRenderUncompressed(t *testing.T) {
	t.Parallel()

	_, err := xbin.Render([]byte("invalid"))
	be.Err(t, err, xbin.ErrNotXBIN)

	// 4x2 uncompressed, default CGA palette, iCE colors enabled (flag 0x08).
	// Background 0x90 -> bright blue (#5555ff) instead of blinking blue (#0000aa).
	body := []byte{
		'X', 0x9F, 'B', 0x0F, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // trailing blank row trimmed
	}
	xb := buildTestXBIN(4, 2, 16, 0x08, nil, nil, body)
	buf, err := xbin.Render(xb)
	be.Err(t, err, nil)

	got := buf.String()
	be.True(t, !strings.Contains(got, "\x00"))
	be.True(t, strings.Contains(got, `<span style="color:#ffffff;background-color:#5555ff;">X</span>`))
	be.True(t, strings.Contains(got, `<span style="color:#ffffff;background-color:#000000;">B</span>`))
	be.Equal(t, strings.Count(got, "\n"), 1)
}

func TestRenderCompressedWithPaletteFontSAUCE(t *testing.T) {
	t.Parallel()

	pal := make([]byte, 48)
	// Color 3 = (33, 19, 1) -> #864d04; Color 15 = (63, 63, 63) -> #ffffff
	pal[3*3], pal[3*3+1], pal[3*3+2] = 33, 19, 1
	pal[15*3], pal[15*3+1], pal[15*3+2] = 63, 63, 63

	font := make([]byte, 16*256)
	// Test all 4 RLE modes on a 4x2 grid (8 cells total):
	// Row 0: mode 0 (raw, count=2 -> 0x01) "AB", mode 1 (char rep 'C' x2 -> 0x41) with attrs 0x03, 0x0F
	// Row 1: mode 2 (attr rep 0x03 x2 -> 0x81) chars 'D','E', mode 3 (pair rep '*' 0x0F x2 -> 0xC1)
	body := []byte{
		0x01, 'A', 0x03, 'B', 0x0F,
		0x41, 'C', 0x03, 0x0F,
		0x81, 0x03, 'D', 'E',
		0xC1, '*', 0x0F,
	}
	// flags = Palette(1) | Font(2) | Compress(4) | NonBlink(8) = 0x0F
	xb := buildTestXBIN(4, 2, 16, 0x0F, pal, font, body)

	if len(xb) > math.MaxUint32 {
		t.Fatalf("xb length %d exceeds max uint32", len(xb))
	}
	sauce := make([]byte, 128)
	copy(sauce[:7], "SAUCE00")
	binary.LittleEndian.PutUint32(sauce[90:94], uint32(len(xb))) //nolint:gosec // G115: len(xb) is checked above
	xb = append(xb, 0x1A)
	xb = append(xb, sauce...)

	buf, err := xbin.Render(xb)
	be.Err(t, err, nil)
	got := buf.String()
	be.True(t, strings.Contains(got, `<span style="color:#864d04;background-color:#000000;">A</span>`))
	be.True(t, strings.Contains(got, `<span style="color:#ffffff;background-color:#000000;">B</span>`))
	be.True(t, strings.Contains(got, `<span style="color:#864d04;background-color:#000000;">C</span>`))
	be.True(t, strings.Contains(got, `<span style="color:#ffffff;background-color:#000000;">C</span>`))
	be.True(t, strings.Contains(got, `<span style="color:#864d04;background-color:#000000;">DE</span>`))
	be.True(t, strings.Contains(got, `<span style="color:#ffffff;background-color:#000000;">**</span>`))
	be.Equal(t, strings.Count(got, "\n"), 2)
}
