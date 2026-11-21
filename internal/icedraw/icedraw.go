// Package icedraw decodes iCEDraw v1.4 (.IDF) files into styled HTML spans
// using the embedded 6-bit VGA palette and CP437 character grid.
package icedraw

import (
	"bytes"
	"encoding/binary"
	"errors"

	"github.com/Defacto2/server/internal/vga"
	"github.com/bengarrett/sauce"
)

// ErrNotIDF is returned when data is not a valid iCEDraw v1.4 file.
var ErrNotIDF = errors.New("not an icedraw v1.4 file")

const (
	headerMagic  = "\x041.4"
	headerLen    = 12
	fontLen      = 4096
	minLen       = headerLen + fontLen + vga.PaletteBytes
	maxCols      = 320
	maxRows      = 2000
	rleEscape    = 0x0001
	rleRecordLen = 4
	cellB        = 2
)

// HTML decodes an iCEDraw v1.4 (.IDF) byte slice into an HTML fragment
// compatible with the binary text viewer (<div>...</div> with inline RGB colors).
func HTML(data []byte) (*bytes.Buffer, error) {
	payload := sauce.Trim(data)
	if len(payload) < minLen || string(payload[:4]) != headerMagic {
		return nil, ErrNotIDF
	}

	x1 := binary.LittleEndian.Uint16(payload[4:6])
	y1 := binary.LittleEndian.Uint16(payload[6:8])
	x2 := binary.LittleEndian.Uint16(payload[8:10])
	y2 := binary.LittleEndian.Uint16(payload[10:12])
	if x1 > x2 || y1 > y2 {
		return nil, ErrNotIDF
	}

	cols := int(x2) - int(x1) + 1
	if cols <= 0 || cols > maxCols {
		return nil, ErrNotIDF
	}

	pal, ok := vga.ParsePalette(payload[len(payload)-vga.PaletteBytes:])
	if !ok {
		return nil, ErrNotIDF
	}

	body := payload[headerLen : len(payload)-fontLen-vga.PaletteBytes]
	cells := decodeRLE(body, cols*maxRows)
	rows := vga.Rows(cells, cols)
	return vga.HTML(rows, pal, vga.Options{
		ICEColor: false, Chars512: false, SkipTrim: false,
	}), nil
}

func decodeRLE(body []byte, maxCells int) []vga.Cell {
	cells := make([]vga.Cell, 0, min(len(body)/cellB, maxCells))
	for i := 0; i+1 < len(body) && len(cells) < maxCells; {
		ch, attr := body[i], body[i+1]
		word := binary.LittleEndian.Uint16(body[i : i+cellB])
		i += cellB
		if word == rleEscape {
			if i+rleRecordLen > len(body) {
				break
			}
			count := int(binary.LittleEndian.Uint16(body[i : i+cellB]))
			c := vga.Cell{Ch: body[i+cellB], Attr: body[i+cellB+1]}
			i += rleRecordLen
			for range min(count, maxCells-len(cells)) {
				cells = append(cells, c)
			}
			continue
		}
		cells = append(cells, vga.Cell{Ch: ch, Attr: attr})
	}
	return cells
}
