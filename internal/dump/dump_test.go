package dump_test

import (
	"encoding/binary"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/Defacto2/server/internal/dump"
	"github.com/nalgeon/be"
)

func TestExtractCGARawAndBSAVE(t *testing.T) {
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

func TestExtractCGALinear(t *testing.T) {
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

func TestExtractCGASPI(t *testing.T) {
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
}

func TestExtractCGACom(t *testing.T) {
	t.Parallel()

	// Stub: mov ax, 0xB800; mov es, ax; mov si, 0x0110 (offset 0x10)
	stub := []byte{
		0xB8, 0x00, 0xB8,
		0x8E, 0xC0,
		0xBE, 0x10, 0x01,
	}
	com := make([]byte, 0x10+16384)
	copy(com, stub)
	com[0x10] = 0xCC
	com[0x10+8192] = 0x33

	got, err := dump.ExtractCGA(com)
	be.Err(t, err, nil)
	be.Equal(t, got[0], byte(0xCC))
	be.Equal(t, got[8192], byte(0x33))
}

func TestConvertToPNG(t *testing.T) {
	t.Parallel()

	raw := make([]byte, 16384)
	raw[0] = 0xE4    // pixels: 3, 2, 1, 0
	raw[8192] = 0x1B // row 1 pixels: 0, 1, 2, 3

	dir := t.TempDir()
	src := filepath.Join(dir, "test.cga")
	dst := filepath.Join(dir, "test.png")
	be.Err(t, os.WriteFile(src, raw, 0o600), nil)

	be.Err(t, dump.ConvertToPNG(src, dst), nil)

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
