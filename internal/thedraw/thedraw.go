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
	rawStub    = 0xB0 // 176-byte header + raw VRAM copy stub
	rleStub    = 0xF0 // 240-byte header + 8086 RLE decoder stub
	sigOff     = 9
	payloadOff = 4
	rowsOff    = 4
	colsOff    = 5
	startOff   = 7
	Cols       = 80
	maxCols    = 160
	maxRows    = 2000
	maxBytes   = 65536
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
	payloadLen := int(binary.LittleEndian.Uint16(data[payloadOff : payloadOff+2]))
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
	stub := data[0x3F:rawStub]
	if !bytes.Contains(stub, leaRaw) && !bytes.Contains(stub, movRaw) {
		return 0, 0, 0, false
	}
	rows = int(data[rowsOff])
	cols = int(binary.LittleEndian.Uint16(data[colsOff : colsOff+2]))
	if rows <= 0 || rows > maxRows || cols <= 0 || cols > maxCols {
		return 0, 0, 0, false
	}
	need = rows * cols * 2
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
	pl := int(binary.LittleEndian.Uint16(data[payloadOff : payloadOff+2]))
	payload := data[rleStub : rleStub+pl]
	startOff := int(binary.LittleEndian.Uint16(data[startOff : startOff+2]))

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
			pos += 2
		}
		rows[r] = row
	}
	return rows
}

type theDrawCell struct {
	ch   byte
	attr byte
}

func decodeGrid(payload []byte, startOff int) [][]theDrawCell {
	const (
		rowStride = Cols * 2
		opFGMax   = 0x10
		opBGMax   = 0x18
		opNewline = 0x18
		opSpace   = 0x19
		opRepeat  = 0x1A
		opBlink   = 0x1B
	)

	baseRow := min(startOff/rowStride, maxRows-1)
	curRow := baseRow
	curCol := (startOff % rowStride) / 2
	var attr byte
	var rows [][]theDrawCell

	ensureRow := func(r int) {
		for len(rows) <= r {
			rows = append(rows, make([]theDrawCell, Cols))
		}
	}
	putCell := func(ch, a byte) {
		if curRow >= maxRows {
			return
		}
		if curCol >= Cols {
			curRow += curCol / Cols
			curCol %= Cols
			if curRow >= maxRows {
				return
			}
		}
		ensureRow(curRow)
		rows[curRow][curCol] = theDrawCell{ch: ch, attr: a}
		curCol++
	}

	for i := 0; i < len(payload); {
		b := payload[i]
		i++
		switch {
		case b == opBlink:
			attr ^= 0x80
		case b < opFGMax:
			attr = (attr & 0xF0) | b
		case b == opNewline:
			baseRow++
			curRow = baseRow
			curCol = 0
		case b < opBGMax:
			attr = (attr & 0x8F) | ((b - opFGMax) << 4)
		case b == opSpace:
			if i >= len(payload) {
				return rows
			}
			count := int(payload[i]) + 1
			i++
			for range count {
				putCell(' ', attr)
			}
		case b == opRepeat:
			if i+1 >= len(payload) {
				return rows
			}
			count := int(payload[i]) + 1
			ch := payload[i+1]
			i += 2
			for range count {
				putCell(ch, attr)
			}
		default:
			putCell(b, attr)
		}
	}
	return rows
}

func isBlankCell(c theDrawCell) bool {
	return (c.ch == 0 || c.ch == ' ') && (c.attr&0xF0) == 0
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
	case 0, '\r', '\n', dosEOF, 0x1B:
		return ' '
	default:
		return b
	}
}

func appendSGR(buf []byte, prevAttr int, next byte) ([]byte, int) {
	// CGA attribute bits: 0..2 FG (BGR), 3 bold/bright, 4..6 BG (BGR), 7 blink.
	cgaToANSI := [8]int{0, 4, 2, 6, 1, 5, 3, 7}
	const defaultAttr = 0x07

	if prevAttr == int(next) {
		return buf, prevAttr
	}

	nextBold := (next & 0x08) != 0
	nextBlink := (next & 0x80) != 0
	nextFG := int(next & 0x07)
	nextBG := int((next >> 4) & 0x07)

	needReset := prevAttr < 0
	var curBold, curBlink bool
	curFG, curBG := 7, 0
	if !needReset {
		curBold = (prevAttr & 0x08) != 0
		curBlink = (prevAttr & 0x80) != 0
		curFG = prevAttr & 0x07
		curBG = (prevAttr >> 4) & 0x07
		if (curBold && !nextBold) || (curBlink && !nextBlink) || next == defaultAttr {
			needReset = true
		}
	}

	var params []string
	if needReset {
		params = append(params, "0")
		curBold, curBlink = false, false
		curFG, curBG = 7, 0
	}
	if nextBold && !curBold {
		params = append(params, "1")
	}
	if nextBlink && !curBlink {
		params = append(params, "5")
	}
	if nextFG != curFG {
		params = append(params, strconv.Itoa(30+cgaToANSI[nextFG]))
	}
	if nextBG != curBG {
		params = append(params, strconv.Itoa(40+cgaToANSI[nextBG]))
	}

	if len(params) > 0 {
		buf = append(buf, "\x1b["...)
		buf = append(buf, strings.Join(params, ";")...)
		buf = append(buf, 'm')
	}
	return buf, int(next)
}

func makeANSI(rows [][]theDrawCell, cols int) []byte {
	const defaultAttr = 0x07
	rows = trimBlank(rows)
	if len(rows) == 0 {
		return nil
	}

	var out []byte
	curAttr := -1

	for i, row := range rows {
		used := usedLen(row)
		for _, cell := range row[:used] {
			out, curAttr = appendSGR(out, curAttr, cell.attr)
			out = append(out, sanitize(cell.ch))
		}
		if i < len(rows)-1 && used < cols {
			if curAttr >= 0 && (curAttr&0xF0) != 0 {
				out = append(out, "\x1b[0m"...)
				curAttr = defaultAttr
			}
			out = append(out, "\r\n"...)
		}
	}
	if curAttr >= 0 && curAttr != defaultAttr {
		out = append(out, "\x1b[0m"...)
	}
	return out
}
