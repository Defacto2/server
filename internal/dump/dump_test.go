package dump_test

import (
	"encoding/binary"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/Defacto2/server/internal/dump"
	"github.com/nalgeon/be"
)

func TestExtractRaw(t *testing.T) {
	t.Parallel()

	raw := make([]byte, 16384)
	raw[0] = 0xAA
	raw[8192] = 0x55

	got, err := dump.ExtractCGA(raw)
	be.Err(t, err, nil)
	be.Equal(t, len(got), 16384)
	be.Equal(t, got[0], byte(0xAA))
	be.Equal(t, got[8192], byte(0x55))

	bsave := make([]byte, 7+16384)
	bsave[0] = 0xFD
	binary.LittleEndian.PutUint16(bsave[1:3], 0xB800)
	binary.LittleEndian.PutUint16(bsave[5:7], 16384)
	copy(bsave[7:], raw)

	got, err = dump.ExtractCGA(bsave)
	be.Err(t, err, nil)
	be.Equal(t, got[0], byte(0xAA))
	be.Equal(t, got[8192], byte(0x55))
}

func TestExtractLinear(t *testing.T) {
	t.Parallel()

	linear := make([]byte, 16384)
	binary.LittleEndian.PutUint16(linear[0:2], 2)
	binary.LittleEndian.PutUint16(linear[2:4], 320)
	binary.LittleEndian.PutUint16(linear[4:6], 200)
	// Bottom-up linear rows: row 199 -> top screen row 0 (even bank), row 198 -> row 1 (odd bank).
	linear[6+199*80] = 0xAA
	linear[6+198*80] = 0x55

	got, err := dump.ExtractCGA(linear)
	be.Err(t, err, nil)
	be.Equal(t, len(got), 16384)
	be.Equal(t, got[0], byte(0xAA))
	be.Equal(t, got[8192], byte(0x55))
}

func TestExtractSPI(t *testing.T) {
	t.Parallel()

	t.Run("full", func(t *testing.T) {
		t.Parallel()
		// 128 * 128 = 16384 bytes encoded as 128 pairs of (0x5A, 127).
		const pairs = 128
		spi := make([]byte, 0, 10+5+pairs*2)
		spi = append(spi, 0xE9, 0x05, 0x00, 'H', 'I')
		spi = append(spi, 'S', 'P', 'I')
		spi = binary.LittleEndian.AppendUint16(spi, pairs)
		for range pairs {
			spi = append(spi, 0x5A, 127)
		}

		got, err := dump.ExtractCGA(spi)
		be.Err(t, err, nil)
		be.Equal(t, len(got), 16384)
		be.Equal(t, got[0], byte(0x5A))
		be.Equal(t, got[16383], byte(0x5A))
	})

	t.Run("padded", func(t *testing.T) {
		t.Parallel()
		// False "SPI" occurrence in prefix followed by a real 16,000-byte SPI stream (padded to 16,384).
		const pairs = 125 // 125 * 128 = 16000 bytes
		spi := []byte{0xEB, 0x05, 'S', 'P', 'I', 0xFF, 0xFF, 'S', 'P', 'I'}
		spi = binary.LittleEndian.AppendUint16(spi, pairs)
		for range pairs {
			spi = append(spi, 0x5A, 127)
		}

		got, err := dump.ExtractCGA(spi)
		be.Err(t, err, nil)
		be.Equal(t, len(got), 16384)
		be.Equal(t, got[0], byte(0x5A))
		be.Equal(t, got[15999], byte(0x5A))
		be.Equal(t, got[16000], byte(0x00))
	})
}

func TestExtractCOM(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		stub []byte
		off  int
	}{
		{
			name: "direct",
			// mov ax, 0xB800; mov es, ax; mov si, 0x0110 (offset 0x10)
			stub: []byte{0xB8, 0x00, 0xB8, 0x8E, 0xC0, 0xBE, 0x10, 0x01},
			off:  0x10,
		},
		{
			name: "via ax",
			// mov ax, 0xB800; mov es, ax; mov ax, 0x0110; mov si, ax
			stub: []byte{0xB8, 0x00, 0xB8, 0x8E, 0xC0, 0xB8, 0x10, 0x01, 0x8B, 0xF0},
			off:  0x10,
		},
		{
			name: "viewer ax",
			// mov ax, 0x0004; int 0x10; mov ax, 0xB800; mov si, 0x0120
			stub: []byte{0xB8, 0x04, 0x00, 0xCD, 0x10, 0xB8, 0x00, 0xB8, 0xBE, 0x20, 0x01},
			off:  0x20,
		},
		{
			name: "viewer al",
			// mov al, 0x04; int 0x10; mov ax, 0xB800; mov si, 0x0120
			stub: []byte{0xB0, 0x04, 0xCD, 0x10, 0xB8, 0x00, 0xB8, 0xBE, 0x20, 0x01},
			off:  0x20,
		},
		{
			name: "jmp",
			stub: []byte{0xEB, 0x10},
			off:  0x20,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			com := make([]byte, tt.off+16384)
			copy(com, tt.stub)
			com[tt.off] = 0xAA
			com[tt.off+8192] = 0x55

			got, err := dump.ExtractCGA(com)
			be.Err(t, err, nil)
			be.Equal(t, got[0], byte(0xAA))
			be.Equal(t, got[8192], byte(0x55))
		})
	}
}

func TestCGAImage(t *testing.T) {
	t.Parallel()

	raw := make([]byte, 16384)
	raw[0] = 0xE4    // pixels: 3, 2, 1, 0
	raw[8192] = 0x1B // row 1 pixels: 0, 1, 2, 3

	dir := t.TempDir()
	src := filepath.Join(dir, "test.cga")
	dst := filepath.Join(dir, "test.png")
	be.Err(t, os.WriteFile(src, raw, 0o600), nil)

	be.Err(t, dump.CGAImage(src, dst), nil)

	f, err := os.Open(dst)
	be.Err(t, err, nil)
	defer f.Close()

	img, err := png.Decode(f)
	be.Err(t, err, nil)
	be.Equal(t, img.Bounds().Dx(), dump.Width*dump.Scale)
	be.Equal(t, img.Bounds().Dy(), dump.Height*dump.Scale)

	pal, ok := img.(*image.Paletted)
	be.True(t, ok)
	// First byte 0xE4 => first pixel index 3, replicated across a 4x4 block.
	be.Equal(t, pal.ColorIndexAt(0, 0), uint8(3))
	be.Equal(t, pal.ColorIndexAt(3, 3), uint8(3))
	be.Equal(t, pal.ColorIndexAt(4, 0), uint8(2))
	// Odd bank byte 0x1B => row 1 (y=4..7) first pixel index 0, second index 1.
	be.Equal(t, pal.ColorIndexAt(0, 4), uint8(0))
	be.Equal(t, pal.ColorIndexAt(4, 7), uint8(1))
}

func TestPalettes(t *testing.T) {
	t.Parallel()

	raw := make([]byte, 16384)
	palettes := []color.Palette{
		dump.PaletteCyan,
		dump.PaletteCyanLow,
		dump.PaletteGreen,
		dump.PaletteGreenLow,
		dump.PaletteRed,
		dump.PaletteRedLow,
	}
	for _, pal := range palettes {
		be.Equal(t, len(pal), 4)
		img := dump.DecodeCGA(raw, pal, dump.Scale)
		be.Equal(t, len(img.Palette), 4)
		be.Equal(t, img.Palette[1], pal[1])
	}

	def := dump.DecodeCGA(raw, nil, dump.Scale)
	be.Equal(t, len(def.Palette), 4)
	be.Equal(t, def.Palette[1], dump.PaletteCyan[1])
}

func TestScale(t *testing.T) {
	t.Parallel()

	raw := make([]byte, 16384)
	raw[0] = 0xE4    // pixels: 3, 2, 1, 0
	raw[8192] = 0x1B // row 1 pixels: 0, 1, 2, 3

	tests := []struct {
		name      string
		scale     int
		wantScale int
	}{
		{"1x", 1, 1},
		{"2x", 2, 2},
		{"3x", 3, 3},
		{"4x", 4, 4},
		{"zero", 0, dump.Scale},
		{"over", 5, dump.Scale},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			img := dump.DecodeCGA(raw, dump.PaletteCyan, tt.scale)
			be.Equal(t, img.Bounds().Dx(), dump.Width*tt.wantScale)
			be.Equal(t, img.Bounds().Dy(), dump.Height*tt.wantScale)
			be.Equal(t, img.ColorIndexAt(0, 0), uint8(3))
			be.Equal(t, img.ColorIndexAt(tt.wantScale-1, tt.wantScale-1), uint8(3))
			be.Equal(t, img.ColorIndexAt(tt.wantScale, 0), uint8(2))
			be.Equal(t, img.ColorIndexAt(tt.wantScale, tt.wantScale), uint8(1))
		})
	}
}

func TestExtractErrors(t *testing.T) {
	t.Parallel()

	_, err := dump.ExtractCGA(nil)
	be.Err(t, err, dump.ErrUnsupported)

	_, err = dump.ExtractCGA([]byte("invalid"))
	be.Err(t, err, dump.ErrUnsupported)

	dir := t.TempDir()
	be.Err(t, dump.CGAImage(filepath.Join(dir, "missing.cga"), filepath.Join(dir, "out.png")))

	badSrc := filepath.Join(dir, "bad.cga")
	be.Err(t, os.WriteFile(badSrc, []byte("not cga"), 0o600), nil)
	be.Err(t, dump.CGAImage(badSrc, filepath.Join(dir, "out.png")), dump.ErrUnsupported)
}
