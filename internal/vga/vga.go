// Package vga provides shared helpers for rendering 16-color CP437 character/attribute
// grids and 6-bit VGA palettes into styled HTML fragments.
//
//nolint:gochecknoglobals
package vga

import (
	"bytes"
	"fmt"
	"html"
	"strings"

	"golang.org/x/text/encoding/charmap"
)

const (
	// PaletteColors is the number of colors in a 4-bit CGA/VGA text palette.
	PaletteColors = 16
	// PaletteBytes is the byte length of an RGB 6-bit VGA DAC palette (16 × 3).
	PaletteBytes = 48

	maxPaletteVal = 0x3F
	fgMask        = 0x0F
	fgBankMask    = 0x07
	bgShift       = 4
	bgMask        = 0x07
	bgICEMask     = 0x0F
	hiAttrMask    = 0xF0
)

// Cell represents a single DOS text-mode character and attribute pair.
type Cell struct {
	Ch   byte
	Attr byte
}

// Palette holds 16 CSS #rrggbb color strings.
type Palette [PaletteColors]string

// Options controls attribute decoding during HTML rendering.
type Options struct {
	ICEColor bool // Treat attribute bit 7 as bright background (16 backgrounds) instead of blink.
	Chars512 bool // Treat attribute bit 3 as glyph bank selector (limits foreground to 8 colors).
	SkipTrim bool // Preserve trailing blank rows; when false, trailing blank rows are trimmed.
}

// CGA returns the default IBM CGA/VGA 16-color text palette.
func CGA() Palette {
	return Palette{
		"#000000", "#0000aa", "#00aa00", "#00aaaa",
		"#aa0000", "#aa00aa", "#aa5500", "#aaaaaa",
		"#555555", "#5555ff", "#55ff55", "#55ffff",
		"#ff5555", "#ff55ff", "#ffff55", "#ffffff",
	}
}

// ParsePalette parses a 48-byte 6-bit VGA DAC palette (values 0..63) into CSS #rrggbb strings.
func ParsePalette(raw []byte) (Palette, bool) {
	var pal Palette
	if len(raw) < PaletteBytes {
		return pal, false
	}

	const lshift, rshift = 2, 4

	scale := func(v byte) byte {
		return (v << lshift) | (v >> rshift)
	}

	for i := range PaletteColors {
		r, g, b := raw[i*3], raw[i*3+1], raw[i*3+2]
		if r > maxPaletteVal || g > maxPaletteVal || b > maxPaletteVal {
			return Palette{}, false
		}
		pal[i] = fmt.Sprintf("#%02x%02x%02x", scale(r), scale(g), scale(b))
	}
	return pal, true
}

// Rows reshapes a flat slice of cells into rows of cols cells.
func Rows(cells []Cell, cols int) [][]Cell {
	if cols <= 0 || len(cells) == 0 {
		return nil
	}
	rowCount := (len(cells) + cols - 1) / cols
	buf := make([]Cell, rowCount*cols)
	copy(buf, cells)
	rows := make([][]Cell, rowCount)
	for i := range rowCount {
		off := i * cols
		rows[i] = buf[off : off+cols : off+cols]
	}
	return rows
}

// HTML renders rows into a <div>...</div> containing merged <span style="..."> runs.
func HTML(rows [][]Cell, pal Palette, opts Options) *bytes.Buffer {
	if !opts.SkipTrim {
		rows = trim(rows, opts)
	}
	var out bytes.Buffer
	out.WriteString("<div>")
	for _, row := range rows {
		writeRow(&out, row, pal, opts)
		out.WriteByte('\n')
	}
	out.WriteString("</div>")
	return &out
}

func writeRow(out *bytes.Buffer, row []Cell, pal Palette, opts Options) {
	if len(row) == 0 {
		return
	}
	var sb strings.Builder
	curAttr := attribute(row[0].Attr, opts)
	openSpan(out, curAttr, pal, opts)
	for _, c := range row {
		attr := attribute(c.Attr, opts)
		if attr != curAttr {
			out.WriteString(html.EscapeString(sb.String()))
			sb.Reset()
			out.WriteString("</span>")
			curAttr = attr
			openSpan(out, curAttr, pal, opts)
		}
		sb.WriteRune(cp437(c.Ch))
	}
	out.WriteString(html.EscapeString(sb.String()))
	out.WriteString("</span>")
}

// trim strips trailing rows containing only NUL/space characters and zero effective background bits.
func trim(rows [][]Cell, opts Options) [][]Cell {
	if len(rows) == 0 {
		return nil
	}
	end := len(rows)
	for end > 0 {
		blank := true
		for _, c := range rows[end-1] {
			if (c.Ch != 0 && c.Ch != ' ') || (attribute(c.Attr, opts)&hiAttrMask) != 0 {
				blank = false
				break
			}
		}
		if !blank {
			break
		}
		end--
	}
	if end == 0 {
		return nil
	}
	return rows[:end]
}

func attribute(b byte, opts Options) byte {
	if !opts.ICEColor {
		b &^= 0x80
	}
	if opts.Chars512 {
		b &^= 0x08
	}
	return b
}

func openSpan(out *bytes.Buffer, attr byte, pal Palette, opts Options) {
	fgIdx := attr & fgMask
	if opts.Chars512 {
		fgIdx &= fgBankMask
	}
	bgIdx := (attr >> bgShift) & bgMask
	if opts.ICEColor {
		bgIdx = (attr >> bgShift) & bgICEMask
	}
	out.WriteString(`<span style="color:`)
	out.WriteString(pal[fgIdx])
	out.WriteString(`;background-color:`)
	out.WriteString(pal[bgIdx])
	out.WriteString(`;">`)
}

// controls maps CP437 bytes 0x00..0x1F (with 0x00 as space) to printable Unicode glyphs.
var controls = [0x20]rune{
	' ', '☺', '☻', '♥', '♦', '♣', '♠', '•',
	'◘', '○', '◙', '♂', '♀', '♪', '♫', '☼',
	'►', '◄', '↕', '‼', '¶', '§', '▬', '↨',
	'↑', '↓', '→', '←', '∟', '↔', '▲', '▼',
}

// cp437 decodes a CP437 byte into a printable Unicode rune, turning 0x00 into space.
func cp437(b byte) rune {
	const space, del, nbsp = 0x20, 0x7F, 0xFF
	switch {
	case b < space:
		return controls[b]
	case b == del:
		return '⌂'
	case b == nbsp:
		return '\u00A0'
	default:
		return charmap.CodePage437.DecodeByte(b)
	}
}
