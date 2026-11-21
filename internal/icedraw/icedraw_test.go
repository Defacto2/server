package icedraw_test

import (
	"encoding/binary"
	"strings"
	"testing"

	"github.com/Defacto2/server/internal/icedraw"
	"github.com/nalgeon/be"
)

func buildTestIDF(x1, y1, x2, y2 uint16, body []byte, customColor [3]byte) []byte {
	const (
		hdrLen  = 12
		fontLen = 4096
		palLen  = 48
	)
	out := make([]byte, hdrLen+len(body)+fontLen+palLen)
	copy(out[:4], "\x041.4")
	binary.LittleEndian.PutUint16(out[4:6], x1)
	binary.LittleEndian.PutUint16(out[6:8], y1)
	binary.LittleEndian.PutUint16(out[8:10], x2)
	binary.LittleEndian.PutUint16(out[10:12], y2)
	copy(out[hdrLen:], body)
	palOff := hdrLen + len(body) + fontLen
	// Color 3 = customColor, color 15 = white (63,63,63).
	out[palOff+3*3] = customColor[0]
	out[palOff+3*3+1] = customColor[1]
	out[palOff+3*3+2] = customColor[2]
	out[palOff+15*3] = 63
	out[palOff+15*3+1] = 63
	out[palOff+15*3+2] = 63
	return out
}

func TestRender(t *testing.T) {
	t.Parallel()

	_, err := icedraw.HTML([]byte("invalid"))
	be.Err(t, err, icedraw.ErrNotIDF)

	// 4 columns wide, row 0 = "AB" + 2 spaces, row 1 = 4 repeated '*' via RLE,
	// rows 2..4 = trailing blank rows that should be trimmed.
	body := []byte{
		'A', 0x03, // color 3 on color 0
		'B', 0x0F, // color 15 on color 0
		0x00, 0x00, 0x00, 0x00, // 2 NUL blank cells
		// Row 1: RLE 4x '*' with attr 0x03
		0x01, 0x00, 0x04, 0x00, '*', 0x03,
		// Rows 2..4: RLE 12x NUL blank cells
		0x01, 0x00, 0x0C, 0x00, 0x00, 0x00,
	}

	// Color 3 = (33, 19, 1) -> 8-bit RGB (#864d04).
	idf := buildTestIDF(0, 0, 3, 0, body, [3]byte{33, 19, 1})
	buf, err := icedraw.HTML(idf)
	be.Err(t, err, nil)

	got := buf.String()
	be.True(t, !strings.Contains(got, "\x00"))
	be.True(t, strings.Contains(got, `<span style="color:#864d04;background-color:#000000;">A</span>`))
	be.True(t, strings.Contains(got, `<span style="color:#ffffff;background-color:#000000;">B</span>`))
	be.True(t, strings.Contains(got, `<span style="color:#864d04;background-color:#000000;">****</span>`))
	// Trimmed trailing blanks -> only 2 rows.
	be.Equal(t, strings.Count(got, "\n"), 2)
}

func TestRenderWithSAUCE(t *testing.T) {
	t.Parallel()

	body := []byte{'H', 0x0F, 'I', 0x0F}
	idf := buildTestIDF(0, 0, 1, 0, body, [3]byte{0, 42, 42})
	sauce := make([]byte, 128)
	copy(sauce[:7], "SAUCE00")
	idfSauce := append(append([]byte{}, idf...), 0x1A)
	idfSauce = append(idfSauce, sauce...)

	buf, err := icedraw.HTML(idfSauce)
	be.Err(t, err, nil)
	be.True(t, strings.Contains(buf.String(), `>HI</span>`))
}
