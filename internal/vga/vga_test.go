package vga_test

import (
	"testing"

	"github.com/Defacto2/server/internal/vga"
	"github.com/nalgeon/be"
)

func TestParsePalette(t *testing.T) {
	t.Parallel()

	t.Run("valid", func(t *testing.T) {
		t.Parallel()
		raw := make([]byte, vga.PaletteBytes)
		raw[0], raw[1], raw[2] = 0, 0, 0
		raw[3], raw[4], raw[5] = 63, 33, 1
		pal, ok := vga.ParsePalette(raw)
		be.True(t, ok)
		be.Equal(t, pal[0], "#000000")
		be.Equal(t, pal[1], "#ff8604")
	})

	t.Run("short", func(t *testing.T) {
		t.Parallel()
		pal, ok := vga.ParsePalette(make([]byte, vga.PaletteBytes-1))
		be.True(t, !ok)
		be.Equal(t, pal, vga.Palette{})
	})

	t.Run("out of range", func(t *testing.T) {
		t.Parallel()
		raw := make([]byte, vga.PaletteBytes)
		raw[0], raw[1], raw[2] = 63, 63, 63
		raw[9] = 64
		pal, ok := vga.ParsePalette(raw)
		be.True(t, !ok)
		be.Equal(t, pal, vga.Palette{})
	})
}

func TestRows(t *testing.T) {
	t.Parallel()

	be.Equal(t, len(vga.Rows(nil, 4)), 0)
	be.Equal(t, len(vga.Rows([]vga.Cell{{Ch: 'A'}}, 0)), 0)
	be.Equal(t, len(vga.Rows([]vga.Cell{{Ch: 'A'}}, -1)), 0)

	cells := []vga.Cell{{Ch: 'A', Attr: 0x0F}, {Ch: 'B', Attr: 0x0F}, {Ch: 'C', Attr: 0x0F}}
	rows := vga.Rows(cells, 2)
	be.Equal(t, len(rows), 2)
	be.Equal(t, len(rows[0]), 2)
	be.Equal(t, cap(rows[0]), 2)
	be.Equal(t, rows[1][0], vga.Cell{Ch: 'C', Attr: 0x0F})
	be.Equal(t, rows[1][1], vga.Cell{})
}

func TestHTML(t *testing.T) {
	t.Parallel()

	pal := vga.CGA()

	t.Run("Chars512", func(t *testing.T) {
		t.Parallel()
		// Attr 0x0F (bright white) with Chars512 clears bit 3 -> 0x07 (#aaaaaa).
		rows := [][]vga.Cell{{{Ch: 'A', Attr: 0x0F}}}
		got := vga.HTML(rows, pal, vga.Options{Chars512: true}).String()
		be.Equal(t, got, `<div><span style="color:#aaaaaa;background-color:#000000;">A</span>`+"\n</div>")
	})

	t.Run("SkipTrim", func(t *testing.T) {
		t.Parallel()
		rows := [][]vga.Cell{
			{{Ch: 'A', Attr: 0x0F}},
			{{Ch: ' ', Attr: 0x00}},
		}
		trimmed := vga.HTML(rows, pal, vga.Options{}).String()
		kept := vga.HTML(rows, pal, vga.Options{SkipTrim: true}).String()
		be.Equal(t, trimmed, `<div><span style="color:#ffffff;background-color:#000000;">A</span>`+"\n</div>")
		be.Equal(t, kept, `<div><span style="color:#ffffff;background-color:#000000;">A</span>`+"\n"+
			`<span style="color:#000000;background-color:#000000;"> </span>`+"\n</div>")
	})

	t.Run("trim blink", func(t *testing.T) {
		t.Parallel()
		// Trailing space with bit 7 set (0x80): invisible in blink mode, bright dark-gray bg in iCE mode.
		rows := [][]vga.Cell{
			{{Ch: 'A', Attr: 0x0F}},
			{{Ch: ' ', Attr: 0x80}},
		}
		blink := vga.HTML(rows, pal, vga.Options{ICEColor: false}).String()
		be.Equal(t, blink, `<div><span style="color:#ffffff;background-color:#000000;">A</span>`+"\n</div>")

		ice := vga.HTML(rows, pal, vga.Options{ICEColor: true}).String()
		be.Equal(t, ice, `<div><span style="color:#ffffff;background-color:#000000;">A</span>`+"\n"+
			`<span style="color:#000000;background-color:#555555;"> </span>`+"\n</div>")
	})

	t.Run("escape", func(t *testing.T) {
		t.Parallel()
		rows := [][]vga.Cell{{
			{Ch: '<', Attr: 0x07},
			{Ch: '&', Attr: 0x07},
			{Ch: '>', Attr: 0x07},
		}}
		got := vga.HTML(rows, pal, vga.Options{}).String()
		be.Equal(t, got, `<div><span style="color:#aaaaaa;background-color:#000000;">&lt;&amp;&gt;</span>`+"\n</div>")
	})

	t.Run("controls", func(t *testing.T) {
		t.Parallel()
		rows := [][]vga.Cell{{
			{Ch: 0x00, Attr: 0x07},
			{Ch: 0x01, Attr: 0x07},
			{Ch: 0x7F, Attr: 0x07},
			{Ch: 0xFF, Attr: 0x07},
		}}
		got := vga.HTML(rows, pal, vga.Options{}).String()
		be.Equal(t, got, `<div><span style="color:#aaaaaa;background-color:#000000;"> ☺⌂`+"\u00A0"+`</span>`+"\n</div>")
	})
}
