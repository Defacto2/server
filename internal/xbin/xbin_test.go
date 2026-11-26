package xbin_test

import (
	"encoding/binary"
	"testing"

	"github.com/Defacto2/server/internal/xbin"
	"github.com/nalgeon/be"
)

// testxbin returns an XBIN file. The pal and font data are only valid when
// their flags (0x01 palette, 0x02 font) are set, otherwise the test fails.
func testxbin(t *testing.T, w, h, fontH, flags byte, pal, font, body []byte) []byte {
	t.Helper()
	if len(pal) > 0 && (flags&0x01) == 0 {
		t.Fatal("palette data given without the palette flag")
	}
	if len(font) > 0 && (flags&0x02) == 0 {
		t.Fatal("font data given without the font flag")
	}
	hdr := []byte{
		'X', 'B', 'I', 'N', 0x1A,
		w, 0x00,
		h, 0x00,
		fontH,
		flags,
	}
	out := append([]byte{}, hdr...)
	out = append(out, pal...)
	out = append(out, font...)
	return append(out, body...)
}

// span returns the expected HTML for a run of text with the given colors.
func span(fg, bg, text string) string {
	return `<span style="color:` + fg + `;background-color:` + bg + `;">` + text + `</span>`
}

func TestRenderNotXBIN(t *testing.T) {
	t.Parallel()

	_, err := xbin.Render([]byte("invalid"))
	be.Err(t, err, xbin.ErrNotXBIN)
}

func TestRenderUncompressed(t *testing.T) {
	t.Parallel()

	// 4x2 uncompressed with the default CGA palette. Attribute 0x9F has the high
	// bit set, which is either a bright background (iCE colors, flag 0x08)
	// or a blinking foreground that keeps the dark background.
	body := []byte{
		'X', 0x9F, 'B', 0x0F, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // trailing blank row trimmed
	}
	tests := []struct {
		name  string
		flags byte
		bgX   string
	}{
		{"ice colors", 0x08, "#5555ff"}, // bright blue
		{"blink", 0x00, "#0000aa"},      // blue, blink bit ignored
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			xb := testxbin(t, 4, 2, 16, tt.flags, nil, nil, body)
			buf, err := xbin.Render(xb)
			be.Err(t, err, nil)

			// The blank second row is trimmed, leaving one row of four cells.
			want := "<div>" +
				span("#ffffff", tt.bgX, "X") +
				span("#ffffff", "#000000", "B") +
				span("#000000", "#000000", "  ") + "\n</div>"
			be.Equal(t, buf.String(), want)
		})
	}
}

func TestRenderCompressed(t *testing.T) {
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
	xb := testxbin(t, 4, 2, 16, 0x0F, pal, font, body)

	buf, err := xbin.Render(xb)
	be.Err(t, err, nil)
	want := "<div>" +
		span("#864d04", "#000000", "A") +
		span("#ffffff", "#000000", "B") +
		span("#864d04", "#000000", "C") +
		span("#ffffff", "#000000", "C") + "\n" +
		span("#864d04", "#000000", "DE") +
		span("#ffffff", "#000000", "**") + "\n</div>"
	be.Equal(t, buf.String(), want)
}

func TestRenderSAUCE(t *testing.T) {
	t.Parallel()

	// 4x2 uncompressed with only 2 cells of data. Without trimming, the appended
	// EOF marker and SAUCE record would be decoded as extra cells.
	body := []byte{'X', 0x0F, 'Y', 0x0F}
	xb := testxbin(t, 4, 2, 16, 0x08, nil, nil, body)
	want, err := xbin.Render(xb)
	be.Err(t, err, nil)

	sauce := make([]byte, 128)
	copy(sauce[:7], "SAUCE00")
	copy(sauce[82:90], "20260101") // date
	sauce[94], sauce[95] = 6, 0    // DataType XBin, FileType 0
	size := uint32(len(xb))        //nolint:gosec // G115: the fixture is a few bytes
	binary.LittleEndian.PutUint32(sauce[90:94], size)
	withSAUCE := append(append(append([]byte{}, xb...), 0x1A), sauce...)

	got, err := xbin.Render(withSAUCE)
	be.Err(t, err, nil)
	be.Equal(t, got.String(), want.String())
}
