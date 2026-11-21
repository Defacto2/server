// Package xbin decodes eXtended BINary (XBIN / .XB) files into styled HTML spans
// using the embedded palette (or standard CGA) and CP437 character grid.
package xbin

import (
	"bytes"
	"encoding/binary"
	"errors"

	"github.com/Defacto2/server/internal/vga"
	"github.com/bengarrett/sauce"
)

// ErrNotXBIN is returned when data is not a valid XBIN file.
var ErrNotXBIN = errors.New("not an xbin file")

const (
	headerMagic   = "XBIN\x1a"
	headerLen     = 11
	magicLen      = 5
	widthOff      = 5
	heightOff     = 7
	fontHeightOff = 9
	flagsOff      = 10
	cellB         = 2
	maxCols       = 640
	maxRows       = 2000
	flagPalette   = 0x01
	flagFont      = 0x02
	flagCompress  = 0x04
	flagNonBlink  = 0x08
	flag512Chars  = 0x10
	glyphs256     = 256
	glyphs512     = 512
	rleModeShift  = 6
	rleModeMask   = 0x03
	rleCountMask  = 0x3F
	modeRaw       = 0
	modeCharRep   = 1
	modeAttrRep   = 2
	modePairRep   = 3
)

// HasFont reports whether data is an XBIN file with an embedded custom font (flag 0x02).
func HasFont(data []byte) bool {
	return len(data) >= headerLen &&
		string(data[:magicLen]) == headerMagic &&
		(data[flagsOff]&flagFont) != 0
}

// Render decodes an XBIN (.XB) byte slice into an HTML fragment compatible with
// the binary text viewer (<div>...</div> with inline RGB colors).
func Render(data []byte) (*bytes.Buffer, error) {
	payload := sauce.Trim(data)
	if len(payload) < headerLen || string(payload[:magicLen]) != headerMagic {
		return nil, ErrNotXBIN
	}

	cols := int(binary.LittleEndian.Uint16(payload[widthOff:heightOff]))
	rows := int(binary.LittleEndian.Uint16(payload[heightOff:fontHeightOff]))
	fontHeight := int(payload[fontHeightOff])
	flags := payload[flagsOff]
	if cols <= 0 || cols > maxCols || rows <= 0 || rows > maxRows {
		return nil, ErrNotXBIN
	}

	off := headerLen
	pal := vga.CGA()
	if (flags & flagPalette) != 0 {
		if off+vga.PaletteBytes > len(payload) {
			return nil, ErrNotXBIN
		}
		var ok bool
		pal, ok = vga.ParsePalette(payload[off : off+vga.PaletteBytes])
		if !ok {
			return nil, ErrNotXBIN
		}
		off += vga.PaletteBytes
	}

	if (flags & flagFont) != 0 {
		glyphs := glyphs256
		if (flags & flag512Chars) != 0 {
			glyphs = glyphs512
		}
		fontBytes := fontHeight * glyphs
		if fontBytes <= 0 || off+fontBytes > len(payload) {
			return nil, ErrNotXBIN
		}
		off += fontBytes
	}

	stream := payload[off:]
	need := cols * rows
	var cells []vga.Cell
	if (flags & flagCompress) != 0 {
		cells = decodeRLE(stream, need)
	} else {
		cells = decodeRaw(stream, need)
	}

	vr := vga.Rows(cells, cols)
	opts := vga.Options{
		ICEColor: (flags & flagNonBlink) != 0,
		Chars512: (flags & flag512Chars) != 0,
		SkipTrim: false,
	}
	return vga.HTML(vr, pal, opts), nil
}

func decodeRaw(stream []byte, need int) []vga.Cell {
	count := min(len(stream)/cellB, need)
	cells := make([]vga.Cell, 0, count)
	for i := 0; i+1 < count*cellB; i += cellB {
		cells = append(cells, vga.Cell{Ch: stream[i], Attr: stream[i+1]})
	}
	return cells
}

func decodeRLE(stream []byte, need int) []vga.Cell {
	cells := make([]vga.Cell, 0, need)
	for i := 0; i < len(stream) && len(cells) < need; {
		ctrl := stream[i]
		i++
		mode := (ctrl >> rleModeShift) & rleModeMask
		count := min(int(ctrl&rleCountMask)+1, need-len(cells))

		var ok bool
		cells, i, ok = stepRLE(stream, i, mode, count, cells)
		if !ok {
			break
		}
	}
	return cells
}

func stepRLE(stream []byte, i int, mode byte, count int, cells []vga.Cell) ([]vga.Cell, int, bool) {
	switch mode {
	case modeRaw:
		if i+count*cellB > len(stream) {
			return cells, i, false
		}
		for range count {
			cells = append(cells, vga.Cell{Ch: stream[i], Attr: stream[i+1]})
			i += cellB
		}
	case modeCharRep:
		if i+1+count > len(stream) {
			return cells, i, false
		}
		ch := stream[i]
		i++
		for range count {
			cells = append(cells, vga.Cell{Ch: ch, Attr: stream[i]})
			i++
		}
	case modeAttrRep:
		if i+1+count > len(stream) {
			return cells, i, false
		}
		attr := stream[i]
		i++
		for range count {
			cells = append(cells, vga.Cell{Ch: stream[i], Attr: attr})
			i++
		}
	case modePairRep:
		if i+cellB > len(stream) {
			return cells, i, false
		}
		c := vga.Cell{Ch: stream[i], Attr: stream[i+1]}
		i += cellB
		for range count {
			cells = append(cells, c)
		}
	}
	return cells, i, true
}
