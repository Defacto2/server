// Package dump extracts and decodes retro CGA graphics dumps into modern images.
package dump

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
)

const (
	// Width is the native horizontal resolution of CGA 320x200 4-color graphics.
	Width = 320
	// Height is the native vertical resolution of CGA 320x200 4-color graphics.
	Height = 200
	// Scale is the integer nearest-neighbor multiplier applied to decoded CGA images.
	Scale = 4

	cgaSize       = 16384
	bankSize      = 8192
	bankActive    = 8000
	rowBytesCount = 80
	pixelsPerByte = 4

	minSPIDecoded = 16000
	spiHeaderLen  = 5

	bsaveHeaderLen  = 7
	bsaveMagic      = 0xFD
	cgaSegment      = 0xB800
	linearHeaderLen = 6
	linearCGALen    = Height * rowBytesCount // 16000 bytes
	linearBPP       = 2

	comOrg  = 0x0100
	jmpNear = 0xE9
	jmpShrt = 0xEB
)

var (
	// ErrUnsupported indicates the input does not contain a recognized CGA graphics dump.
	ErrUnsupported = errors.New("unsupported CGA graphics dump")
)

// Palette selects one of the standard 4-color CGA graphics palettes.
type Palette string

const (
	// PaletteCyan is CGA mode 4 palette 1 (black, light cyan, light magenta, white).
	PaletteCyan Palette = "cyan"
	// PaletteGreen is CGA mode 4 palette 0 (black, light green, light red, yellow).
	PaletteGreen Palette = "green"
)

func colorPalette(p Palette) color.Palette {
	if p == PaletteGreen {
		return color.Palette{
			color.NRGBA{R: 0, G: 0, B: 0, A: 255},
			color.NRGBA{R: 85, G: 255, B: 85, A: 255},
			color.NRGBA{R: 255, G: 85, B: 85, A: 255},
			color.NRGBA{R: 255, G: 255, B: 85, A: 255},
		}
	}
	return color.Palette{
		color.NRGBA{R: 0, G: 0, B: 0, A: 255},
		color.NRGBA{R: 85, G: 255, B: 255, A: 255},
		color.NRGBA{R: 255, G: 85, B: 255, A: 255},
		color.NRGBA{R: 255, G: 255, B: 255, A: 255},
	}
}

// ExtractCGA extracts or decompresses a 16KB CGA graphics VRAM buffer from raw 16KB dumps,
// BSAVE files, DOS .COM embedded dumps, or "SPI" RLE-compressed .COM dumps.
func ExtractCGA(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, ErrUnsupported
	}

	if spi, ok := extractSPI(data); ok {
		return spi, nil
	}
	if bsave, ok := extractBSAVE(data); ok {
		return bsave, nil
	}
	if linear, ok := extractLinear(data); ok {
		return linear, nil
	}
	if len(data) == cgaSize {
		out := make([]byte, cgaSize)
		copy(out, data)
		return out, nil
	}
	if com, ok := extractCOM(data); ok {
		return com, nil
	}
	return nil, ErrUnsupported
}

// DecodeCGA renders a 16KB interleaved CGA graphics VRAM buffer into a pixel-exact
// 4x-scaled (1280x800) paletted image.
func DecodeCGA(cga []byte, pal Palette) *image.Paletted {
	const outW = Width * Scale
	img := image.NewPaletted(image.Rect(0, 0, outW, Height*Scale), colorPalette(pal))
	for y := range Height {
		bankOffset := (y / 2) * rowBytesCount
		if y%2 == 1 {
			bankOffset += bankSize
		}
		if bankOffset+rowBytesCount > len(cga) {
			break
		}
		row := cga[bankOffset : bankOffset+rowBytesCount]
		dstRow := y * Scale * img.Stride
		firstLine := img.Pix[dstRow : dstRow+outW]
		for bx := range rowBytesCount {
			b := row[bx]
			baseX := bx * pixelsPerByte * Scale
			for p := range pixelsPerByte {
				idx := (b >> (6 - 2*p)) & 3
				px := baseX + p*Scale
				firstLine[px+0] = idx
				firstLine[px+1] = idx
				firstLine[px+2] = idx
				firstLine[px+3] = idx
			}
		}
		for sy := 1; sy < Scale; sy++ {
			nextRow := dstRow + sy*img.Stride
			copy(img.Pix[nextRow:nextRow+outW], firstLine)
		}
	}
	return img
}

// ConvertToPNG reads a CGA graphics dump from srcPath, decodes it, and writes a PNG to dstPath.
func ConvertToPNG(srcPath, dstPath string) error {
	src, err := os.ReadFile(filepath.Clean(srcPath))
	if err != nil {
		return fmt.Errorf("read cga dump: %w", err)
	}
	cga, err := ExtractCGA(src)
	if err != nil {
		return fmt.Errorf("extract cga dump: %w", err)
	}
	img := DecodeCGA(cga, PaletteCyan)

	f, err := os.Create(filepath.Clean(dstPath))
	if err != nil {
		return fmt.Errorf("create png: %w", err)
	}
	defer f.Close()

	if err := png.Encode(f, img); err != nil {
		return fmt.Errorf("encode png: %w", err)
	}
	return nil
}

func extractSPI(data []byte) ([]byte, bool) {
	sig := []byte("SPI")
	if len(data) < spiHeaderLen {
		return nil, false
	}
	if data[0] != jmpNear && data[0] != jmpShrt && !bytes.HasPrefix(data, sig) {
		return nil, false
	}

	for searchFrom := 0; ; {
		rel := bytes.Index(data[searchFrom:], sig)
		if rel < 0 {
			return nil, false
		}
		idx := searchFrom + rel
		if out, ok := decodeSPI(data[idx:]); ok {
			return out, true
		}
		searchFrom = idx + 1
	}
}

func decodeSPI(tail []byte) ([]byte, bool) {
	if len(tail) < spiHeaderLen {
		return nil, false
	}
	pairCount := int(binary.LittleEndian.Uint16(tail[3:5]))
	if spiHeaderLen+2*pairCount != len(tail) {
		return nil, false
	}
	out := make([]byte, 0, cgaSize)
	for pos := spiHeaderLen; pos < len(tail); pos += 2 {
		val := tail[pos]
		rep := int(tail[pos+1]) + 1
		if len(out)+rep > cgaSize {
			return nil, false
		}
		for range rep {
			out = append(out, val)
		}
	}
	if len(out) < minSPIDecoded {
		return nil, false
	}
	if len(out) < cgaSize {
		out = append(out, make([]byte, cgaSize-len(out))...)
	}
	return out, true
}

func extractBSAVE(data []byte) ([]byte, bool) {
	if len(data) < bsaveHeaderLen+cgaSize || data[0] != bsaveMagic {
		return nil, false
	}
	seg := binary.LittleEndian.Uint16(data[1:3])
	payloadLen := int(binary.LittleEndian.Uint16(data[5:7]))
	if seg != cgaSegment || payloadLen != cgaSize {
		return nil, false
	}
	out := make([]byte, cgaSize)
	copy(out, data[bsaveHeaderLen:bsaveHeaderLen+cgaSize])
	return out, true
}

func extractLinear(data []byte) ([]byte, bool) {
	if len(data) < linearHeaderLen+linearCGALen {
		return nil, false
	}
	bpp := binary.LittleEndian.Uint16(data[0:2])
	w := binary.LittleEndian.Uint16(data[2:4])
	h := binary.LittleEndian.Uint16(data[4:6])
	if bpp != linearBPP || w != Width || h != Height {
		return nil, false
	}
	linear := data[linearHeaderLen : linearHeaderLen+linearCGALen]
	out := make([]byte, cgaSize)
	for y := range Height {
		srcRow := Height - 1 - y
		dstOff := (y / 2) * rowBytesCount
		if y%2 == 1 {
			dstOff += bankSize
		}
		copy(out[dstOff:dstOff+rowBytesCount], linear[srcRow*rowBytesCount:(srcRow+1)*rowBytesCount])
	}
	return out, true
}

func extractCOM(data []byte) ([]byte, bool) {
	if len(data) < cgaSize {
		return nil, false
	}
	if off, ok := comB800Offset(data); ok {
		return copyPaddedCGA(data[off:]), true
	}
	if off, ok := comViewerOffset(data); ok {
		return copyPaddedCGA(data[off:]), true
	}
	if data[0] == jmpNear || data[0] == jmpShrt {
		return copyPaddedCGA(data[len(data)-cgaSize:]), true
	}
	return nil, false
}

func copyPaddedCGA(src []byte) []byte {
	out := make([]byte, cgaSize)
	copy(out, src)
	return out
}

func comB800Offset(data []byte) (int, bool) {
	const (
		scanMax     = 4096
		windowBack  = 12
		windowAhead = 10
		minMovSi    = 3
		opMovSi     = 0xBE
		opMovAx     = 0xB8
		opMovReg    = 0x8B
		modRmSiAx   = 0xF0
	)
	n := min(len(data), scanMax)
	if n < 8 {
		return 0, false
	}
	pointsToCGA := func(i int) bool {
		return data[i] == 0xB8 && data[i+1] == 0x00 && data[i+2] == 0xB8 &&
			data[i+3] == 0x8E && data[i+4] == 0xC0
	}
	readSourceAddr := func(j int) (int, bool) {
		isDirect := data[j] == opMovSi
		isViaAX := j+4 < n && data[j] == opMovAx && data[j+3] == opMovReg && data[j+4] == modRmSiAx
		if !isDirect && !isViaAX {
			return 0, false
		}
		return int(binary.LittleEndian.Uint16(data[j+1 : j+3])), true
	}

	for i := 0; i+4 < n; i++ {
		if !pointsToCGA(i) {
			continue
		}
		lo := max(0, i-windowBack)
		hi := min(n-minMovSi, i+windowAhead)
		for j := lo; j <= hi; j++ {
			if si, ok := readSourceAddr(j); ok && si >= comOrg && si-comOrg+bankSize+bankActive <= len(data) {
				return si - comOrg, true
			}
		}
	}
	return 0, false
}

func comViewerOffset(data []byte) (int, bool) {
	const (
		headerScan    = 24
		opMovSi       = 0xBE
		movAxInt10Len = 5
		movAlInt10Len = 4
	)
	if len(data) < 16 {
		return 0, false
	}
	hdr := data[:min(len(data), headerScan)]
	start := 0
	switch {
	case hdr[0] == 0xB8 && hdr[1] == 0x04 && hdr[2] == 0x00 && hdr[3] == 0xCD && hdr[4] == 0x10:
		start = movAxInt10Len
	case hdr[0] == 0xB0 && hdr[1] == 0x04 && hdr[2] == 0xCD && hdr[3] == 0x10:
		start = movAlInt10Len
	default:
		return 0, false
	}

	var hasB800 bool
	siOff := -1
	for i := start; i+2 < len(hdr); i++ {
		if hdr[i] == 0xB8 && hdr[i+1] == 0x00 && hdr[i+2] == 0xB8 {
			hasB800 = true
		}
		if siOff < 0 && hdr[i] == opMovSi {
			si := int(binary.LittleEndian.Uint16(hdr[i+1 : i+3]))
			if si >= comOrg {
				siOff = si - comOrg
			}
		}
	}
	if !hasB800 || siOff < 0 || siOff+bankSize+bankActive > len(data) {
		return 0, false
	}
	return siOff, true
}
