//nolint:exhaustruct_v5,nonamedreturns
package thedraw

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const header = "TheDraw COM file Screen Save\x1A"

const (
	rawStub      = 0xB0 // 176-byte header + raw VRAM copy stub
	rleStub      = 0xF0 // 240-byte header + 8086 RLE decoder stub
	stubEntryOff = 0x3F // offset of entry code after JMP 013Fh and error string
	sigOff       = 9
	payloadOff   = 4
	rowsOff      = 4
	colsOff      = 5
	startOff     = 7
	wordBytes    = 2
	cellBytes    = 2 // bytes per VGA text cell (character + attribute)
	Cols         = 80
	maxCols      = 160
	maxRows      = 2000
	maxBytes     = 65536
)

const (
	attrBGShift   = 4
	attrMaskFG    = 0x07
	attrMaskBold  = 0x08
	attrMaskHi    = 0xF0 // background + blink nibble mask
	attrMaskNoBG  = 0x8F // foreground + blink mask (clears background)
	attrMaskBlink = 0x80
	defaultAttr   = 0x07
	defaultFG     = 7
	defaultBG     = 0
	sgrFGBase     = 30
	sgrBGBase     = 40
	charESC       = 0x1B
)

// ErrNotCOM is returned when a file does not match the TheDraw .COM screen save format.
var ErrNotCOM = errors.New("not a thedraw com file screen save")

// IsCOM reports whether data has an uncompressed TheDraw .COM screen save header.
func IsCOM(data []byte) bool {
	if !hasHeader(data) {
		return false
	}
	if _, _, _, ok := rawSpan(data); ok {
		return true
	}
	payloadLen := int(binary.LittleEndian.Uint16(data[payloadOff : payloadOff+wordBytes]))
	return len(data) >= rleStub && payloadLen > 0 && rleStub+payloadLen <= len(data)
}

func hasHeader(data []byte) bool {
	if len(data) < rawStub {
		return false
	}
	jump := []byte{0xEB, 0x3D, 0x90}
	prefix := []byte(header)
	return bytes.Equal(data[:len(jump)], jump) &&
		bytes.HasPrefix(data[sigOff:], prefix)
}

func rawSpan(data []byte) (rows, cols, need int, ok bool) {
	if len(data) < rawStub {
		return 0, 0, 0, false
	}

	// The 176-byte uncompressed stub loads SI from CS:01B0h (file offset 0xB0).
	leaRaw := []byte{0x8D, 0x36, 0xB0, 0x01}
	movRaw := []byte{0xBE, 0xB0, 0x01}
	stub := data[stubEntryOff:rawStub]
	if !bytes.Contains(stub, leaRaw) && !bytes.Contains(stub, movRaw) {
		return 0, 0, 0, false
	}

	rows = int(data[rowsOff])
	cols = int(binary.LittleEndian.Uint16(data[colsOff : colsOff+wordBytes]))
	if rows <= 0 || rows > maxRows || cols <= 0 || cols > maxCols {
		return 0, 0, 0, false
	}

	need = rows * cols * cellBytes
	if rawStub+need > len(data) {
		return 0, 0, 0, false
	}
	return rows, cols, need, true
}

// Decode reads up to 64KB of the file at path and decodes any uncompressed
// "TheDraw .COM screen save" into ANSI text.
func Decode(path string) ([]byte, error) {
	const format = "thedraw file %s: %w"
	f, err := os.Open(filepath.Clean(path))
	if err != nil {
		return nil, fmt.Errorf(format, "decode", err)
	}
	defer f.Close()

	data, err := io.ReadAll(io.LimitReader(f, maxBytes+1))
	if err != nil {
		return nil, fmt.Errorf(format, "read", err)
	}
	if len(data) > maxBytes {
		return nil, ErrNotCOM
	}
	return Extract(data)
}

// Extract decodes an uncompressed "TheDraw .COM screen save".
func Extract(data []byte) ([]byte, error) {
	if !IsCOM(data) {
		return nil, ErrNotCOM
	}

	if rows, cols, need, ok := rawSpan(data); ok {
		grid := decodeRaw(data[rawStub:rawStub+need], rows, cols)
		return makeANSI(grid, cols), nil
	}

	pl := int(binary.LittleEndian.Uint16(data[payloadOff : payloadOff+wordBytes]))
	payload := data[rleStub : rleStub+pl]
	startOff := int(binary.LittleEndian.Uint16(data[startOff : startOff+wordBytes]))

	rows := decodeGrid(payload, startOff)
	return makeANSI(rows, Cols), nil
}

func decodeRaw(payload []byte, rowCount, cols int) [][]theDrawCell {
	rows := make([][]theDrawCell, rowCount)
	pos := 0
	for r := range rowCount {
		row := make([]theDrawCell, cols)
		for c := range cols {
			row[c] = theDrawCell{ch: payload[pos], attr: payload[pos+1]}
			pos += cellBytes
		}
		rows[r] = row
	}
	return rows
}

type theDrawCell struct {
	ch   byte
	attr byte
}

type rleGrid struct {
	baseRow int
	curRow  int
	curCol  int
	attr    byte
	rows    [][]theDrawCell
}

func decodeGrid(payload []byte, startOff int) [][]theDrawCell {
	const rowStride = Cols * cellBytes
	baseRow := min(startOff/rowStride, maxRows-1)
	g := &rleGrid{
		baseRow: baseRow,
		curRow:  baseRow,
		curCol:  (startOff % rowStride) / cellBytes,
	}
	for i := 0; i < len(payload); {
		var ok bool
		if i, ok = g.step(payload, i); !ok {
			return g.rows
		}
	}
	return g.rows
}

func (g *rleGrid) putCell(ch byte) {
	if g.curRow >= maxRows {
		return
	}
	if g.curCol >= Cols {
		g.curRow += g.curCol / Cols
		g.curCol %= Cols
		if g.curRow >= maxRows {
			return
		}
	}
	for len(g.rows) <= g.curRow {
		g.rows = append(g.rows, make([]theDrawCell, Cols))
	}
	g.rows[g.curRow][g.curCol] = theDrawCell{ch: ch, attr: g.attr}
	g.curCol++
}

func (g *rleGrid) putRepeat(ch byte, count int) {
	for range count {
		g.putCell(ch)
	}
}

func (g *rleGrid) step(payload []byte, i int) (int, bool) {
	const (
		opFGMax   = 0x10
		opBGMax   = 0x18
		opNewline = 0x18
		opSpace   = 0x19
		opRepeat  = 0x1A
		opBlink   = 0x1B
	)

	b := payload[i]
	i++
	switch {
	case b == opBlink:
		g.attr ^= attrMaskBlink
	case b < opFGMax:
		g.attr = (g.attr & attrMaskHi) | b
	case b == opNewline:
		g.baseRow++
		g.curRow = g.baseRow
		g.curCol = 0
	case b < opBGMax:
		g.attr = (g.attr & attrMaskNoBG) | ((b - opFGMax) << attrBGShift)
	case b == opSpace:
		if i >= len(payload) {
			return i, false
		}
		g.putRepeat(' ', int(payload[i])+1)
		i++
	case b == opRepeat:
		if i+1 >= len(payload) {
			return i, false
		}
		g.putRepeat(payload[i+1], int(payload[i])+1)
		i += wordBytes
	default:
		g.putCell(b)
	}
	return i, true
}

func isBlankCell(c theDrawCell) bool {
	return (c.ch == 0 || c.ch == ' ') && (c.attr&attrMaskHi) == 0
}

func trimBlank(rows [][]theDrawCell) [][]theDrawCell {
	end := len(rows)
	for end > 0 {
		blank := true
		for _, c := range rows[end-1] {
			if !isBlankCell(c) {
				blank = false
				break
			}
		}
		if !blank {
			break
		}
		end--
	}
	return rows[:end]
}

func usedLen(row []theDrawCell) int {
	end := len(row)
	for end > 0 && isBlankCell(row[end-1]) {
		end--
	}
	return end
}

func sanitize(b byte) byte {
	const dosEOF = 0x1A
	switch b {
	case 0, '\r', '\n', dosEOF, charESC:
		return ' '
	default:
		return b
	}
}

type attrState struct {
	bold  bool
	blink bool
	fg    int
	bg    int
}

func parseAttr(a byte) attrState {
	return attrState{
		bold:  (a & attrMaskBold) != 0,
		blink: (a & attrMaskBlink) != 0,
		fg:    int(a & attrMaskFG),
		bg:    int((a >> attrBGShift) & attrMaskFG),
	}
}

func appendSGR(buf []byte, prev byte, hasPrev bool, next byte) ([]byte, byte) {
	// CGA attribute bits: 0..2 FG (BGR), 3 bold/bright, 4..6 BG (BGR), 7 blink.
	cgaToANSI := [8]int{0, 4, 2, 6, 1, 5, 3, 7}

	if hasPrev && prev == next {
		return buf, prev
	}

	nxt := parseAttr(next)
	needReset := !hasPrev
	cur := attrState{bold: false, blink: false, fg: defaultFG, bg: defaultBG}
	if hasPrev {
		cur = parseAttr(prev)
		if (cur.bold && !nxt.bold) || (cur.blink && !nxt.blink) || next == defaultAttr {
			needReset = true
		}
	}

	var params []string
	if needReset {
		params = append(params, "0")
		cur = attrState{bold: false, blink: false, fg: defaultFG, bg: defaultBG}
	}
	if nxt.bold && !cur.bold {
		params = append(params, "1")
	}
	if nxt.blink && !cur.blink {
		params = append(params, "5")
	}
	if nxt.fg != cur.fg {
		params = append(params, strconv.Itoa(sgrFGBase+cgaToANSI[nxt.fg]))
	}
	if nxt.bg != cur.bg {
		params = append(params, strconv.Itoa(sgrBGBase+cgaToANSI[nxt.bg]))
	}

	if len(params) > 0 {
		buf = append(buf, "\x1b["...)
		buf = append(buf, strings.Join(params, ";")...)
		buf = append(buf, 'm')
	}
	return buf, next
}

func makeANSI(rows [][]theDrawCell, cols int) []byte {
	rows = trimBlank(rows)
	if len(rows) == 0 {
		return nil
	}

	var out []byte
	var curAttr byte
	hasAttr := false

	for i, row := range rows {
		used := usedLen(row)
		for _, cell := range row[:used] {
			out, curAttr = appendSGR(out, curAttr, hasAttr, cell.attr)
			hasAttr = true
			out = append(out, sanitize(cell.ch))
		}
		if i < len(rows)-1 && used < cols {
			if hasAttr && (curAttr&attrMaskHi) != 0 {
				out = append(out, "\x1b[0m"...)
				curAttr = defaultAttr
			}
			out = append(out, "\r\n"...)
		}
	}
	if hasAttr && curAttr != defaultAttr {
		out = append(out, "\x1b[0m"...)
	}
	return out
}
