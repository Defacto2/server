// Package dump extracts and decodes retro CGA graphics dumps into modern images.
// These can be found from a number of sources in programs for PC/MS-DOS:
//
//   - saved raw 16KB dumps.
//   - saved BSAVE files.
//   - DOS .COM files with embedded dumps.
//   - DOS .COM files with "SPI" RLE-compressed dumps.
//
// These forms of hidden art were commonly used by PC game developers of
// the 1980s, especially with titles that only supported CGA graphics.
// And briefly by software pirates on the PC from possibly 1987-1990.
// Who would use tools such as PC Paint to create draw their own
// crack-screens, then ship them in their own .COM files to act as
// static crack-tros.
//
//nolint:gochecknoglobals
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
	rows          = 80
	pixelsPerByte = 4

	minSPIDecoded = 16000
	spiHeaderLen  = 5

	bsaveHeaderLen  = 7
	bsaveMagic      = 0xFD
	cgaSegment      = 0xB800
	linearHeaderLen = 6
	linearCGALen    = Height * rows // 16000 bytes
	linearBPP       = 2

	comOrg  = 0x0100
	jmpNear = 0xE9
	jmpShrt = 0xEB
)

// ErrUnsupported indicates the input does not contain a recognized CGA graphics dump.
var ErrUnsupported = errors.New("unsupported CGA graphics dump")

// Standard IBM CGA RGBI hardware colors.
//
//nolint:mnd
var (
	cgaBlack        = color.NRGBA{R: 0, G: 0, B: 0, A: 255}       // #000000
	cgaGreen        = color.NRGBA{R: 0, G: 170, B: 0, A: 255}     // #00AA00
	cgaCyan         = color.NRGBA{R: 0, G: 170, B: 170, A: 255}   // #00AAAA
	cgaRed          = color.NRGBA{R: 170, G: 0, B: 0, A: 255}     // #AA0000
	cgaMagenta      = color.NRGBA{R: 170, G: 0, B: 170, A: 255}   // #AA00AA
	cgaBrown        = color.NRGBA{R: 170, G: 85, B: 0, A: 255}    // #AA5500
	cgaLightGray    = color.NRGBA{R: 170, G: 170, B: 170, A: 255} // #AAAAAA
	cgaLightGreen   = color.NRGBA{R: 85, G: 255, B: 85, A: 255}   // #55FF55
	cgaLightCyan    = color.NRGBA{R: 85, G: 255, B: 255, A: 255}  // #55FFFF
	cgaLightRed     = color.NRGBA{R: 255, G: 85, B: 85, A: 255}   // #FF5555
	cgaLightMagenta = color.NRGBA{R: 255, G: 85, B: 255, A: 255}  // #FF55FF
	cgaYellow       = color.NRGBA{R: 255, G: 255, B: 85, A: 255}  // #FFFF55
	cgaWhite        = color.NRGBA{R: 255, G: 255, B: 255, A: 255} // #FFFFFF
)

// Standard 4-color CGA graphics palettes (Mode 4 Palettes 0 & 1 and Mode 5, high and low intensity).
// For a visualization of the palettes, see https://en.wikipedia.org/wiki/Color_Graphics_Adapter#320_%C3%97_200
var (
	// PaletteCyan is CGA mode 4, palette 1, high intensity.
	PaletteCyan = color.Palette{cgaBlack, cgaLightCyan, cgaLightMagenta, cgaWhite}
	// PaletteCyanLow is CGA mode 4, palette 1, low intensity.
	PaletteCyanLow = color.Palette{cgaBlack, cgaCyan, cgaMagenta, cgaLightGray}
	// PaletteGreen is CGA mode 4, palette 0, high intensity.
	PaletteGreen = color.Palette{cgaBlack, cgaLightGreen, cgaLightRed, cgaYellow}
	// PaletteGreenLow is CGA mode 4, palette 0, low intensity.
	PaletteGreenLow = color.Palette{cgaBlack, cgaGreen, cgaRed, cgaBrown}
	// PaletteRed is CGA mode 5, high intensity.
	PaletteRed = color.Palette{cgaBlack, cgaLightCyan, cgaLightRed, cgaWhite}
	// PaletteRedLow is CGA mode 5, low intensity.
	PaletteRedLow = color.Palette{cgaBlack, cgaCyan, cgaRed, cgaLightGray}
)

// CGAImage reads a CGA graphics dump from the src path, decodes it, and writes the dst PNG path.
// It uses the default cyan palette and scales the graphics by x4 to create a 1280x800 pixel image.
func CGAImage(src, dst string) error {
	return CGACustom(src, dst, PaletteCyan, Scale)
}

// CGACustom reads a CGA graphics dump from the src path, decodes it, and writes the dst PNG path.
func CGACustom(src, dst string, pal color.Palette, scale int) error {
	data, err := os.ReadFile(filepath.Clean(src))
	if err != nil {
		return fmt.Errorf("read cga dump: %w", err)
	}
	cga, err := ExtractCGA(data)
	if err != nil {
		return fmt.Errorf("extract cga dump: %w", err)
	}
	img := DecodeCGA(cga, pal, scale)

	f, err := os.Create(filepath.Clean(dst))
	if err != nil {
		return fmt.Errorf("create png: %w", err)
	}
	defer f.Close()

	if err := png.Encode(f, img); err != nil {
		return fmt.Errorf("encode png: %w", err)
	}
	return nil
}

// ExtractCGA extracts or decompresses a 16KB CGA graphics VRAM buffer.
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

// DecodeCGA to a pixel-exact x-scaled (1280x800) paletted image.
// By default it uses a 4x scale value, meaning a 320x200 dump
// will be resized into a 1280x800 image.
func DecodeCGA(data []byte, pal color.Palette, scale int) *image.Paletted {
	if len(pal) == 0 {
		pal = PaletteCyan
	}
	if scale < 1 || scale > Scale {
		scale = Scale
	}

	x1 := Width * scale
	y1 := Height * scale
	img := image.NewPaletted(image.Rect(0, 0, x1, y1), pal)
	const banks = 2
	const mask = 3 // 2-bit mask
	const bpp = 2  // bits per pixel
	const shift = 6
	for y := range Height {
		bankOffset := (y / banks) * rows
		if y%banks == 1 {
			bankOffset += bankSize
		}
		if bankOffset+rows > len(data) {
			break
		}
		row := data[bankOffset : bankOffset+rows]
		dstRow := y * scale * img.Stride
		firstLine := img.Pix[dstRow : dstRow+x1]
		for bx := range rows {
			b := row[bx]
			baseX := bx * pixelsPerByte * scale
			for p := range pixelsPerByte {
				idx := (b >> (shift - bpp*p)) & mask
				px := baseX + p*scale
				for sx := range scale {
					firstLine[px+sx] = idx
				}
			}
		}
		for sy := 1; sy < scale; sy++ {
			nextRow := dstRow + sy*img.Stride
			copy(img.Pix[nextRow:nextRow+x1], firstLine)
		}
	}
	return img
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
	const banks = 2
	for y := range Height {
		srcRow := Height - 1 - y
		dstOff := (y / banks) * rows
		if y%banks == 1 {
			dstOff += bankSize
		}
		copy(out[dstOff:dstOff+rows], linear[srcRow*rows:(srcRow+1)*rows])
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

	const minBytes = 8
	n := min(len(data), scanMax)
	if n < minBytes {
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
		headerScan = 24
		sanity     = 16
		opMovSi    = 0xBE
	)
	if len(data) < sanity {
		return 0, false
	}

	hdr := data[:min(len(data), headerScan)]
	rest, ok := trimMode4Prefix(hdr)
	if !ok || !bytes.Contains(rest, []byte{0xB8, 0x00, 0xB8}) {
		return 0, false
	}

	for i := 0; i+2 < len(rest); i++ {
		if rest[i] != opMovSi {
			continue
		}
		si := int(binary.LittleEndian.Uint16(rest[i+1 : i+3]))
		if si >= comOrg {
			off := si - comOrg
			return off, off+bankSize+bankActive <= len(data)
		}
	}
	return 0, false
}

func trimMode4Prefix(hdr []byte) ([]byte, bool) {
	movAxMode4 := []byte{0xB8, 0x04, 0x00, 0xCD, 0x10}
	if bytes.HasPrefix(hdr, movAxMode4) {
		return hdr[len(movAxMode4):], true
	}
	movAlMode4 := []byte{0xB0, 0x04, 0xCD, 0x10}
	if bytes.HasPrefix(hdr, movAlMode4) {
		return hdr[len(movAlMode4):], true
	}
	return nil, false
}
