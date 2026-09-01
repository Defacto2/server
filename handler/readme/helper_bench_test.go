package readme_test

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"testing"
)

type ansitest struct {
	name string
	data []byte
}

func ansitests(b *testing.B) []ansitest {
	b.Helper()

	makeData := func(size int, ansi string, pos int) []byte {
		buf := bytes.Repeat([]byte("Lorem ipsum dolor sit amet, consectetur. "), (size/40)+1)[:size]
		if pos >= 0 && pos < size {
			copy(buf[pos:], ansi)
		}
		return buf
	}
	const kb, mb = 1024, 1024 * 1024
	tests := []ansitest{
		{"ImmediateMatch", []byte("\x1b[31mRed Text")},
		{"Match1KB      ", makeData(1*kb, "\x1b[1A", 500)},
		{"Match500KB    ", makeData(500*kb, "\x1b[10;20H", 450*kb)},
		{"NoMatch32KB   ", makeData(32*kb, "", -1)},
		{"NoMatch1MB    ", makeData(1*mb, "", -1)},
		{"BoundarySplit ", func() []byte {
			d := makeData(64*kb, "", -1)
			d[32*kb-1], d[32*kb] = '\x1b', '['
			return d
		}()},
	}
	return tests
}

func BenchmarkRM(b *testing.B) {
	for n, tt := range ansitests(b) {
		b.Run(tt.name+" 00 #"+strconv.Itoa(n), func(b *testing.B) {
			b.ReportAllocs()
			b.SetBytes(int64(len(tt.data)))
			for b.Loop() {
				_ = remove00(tt.data)
			}
		})
		b.Run(tt.name+" 01 #"+strconv.Itoa(n), func(b *testing.B) {
			b.ReportAllocs()
			b.SetBytes(int64(len(tt.data)))
			for b.Loop() {
				_ = remove01(tt.data)
			}
		})
	}
}

func BenchmarkHas(b *testing.B) {
	// These patterns share the same iteration to use same input data.
	for n, tt := range ansitests(b) {
		b.Run(tt.name+" 00 #"+strconv.Itoa(n), func(b *testing.B) {
			b.ReportAllocs()
			b.SetBytes(int64(len(tt.data)))
			r := bytes.NewReader(nil)
			for b.Loop() {
				r.Reset(tt.data)
				_, _ = hasANSI00(r)
			}
		})

		b.Run(tt.name+" 01 #"+strconv.Itoa(n), func(b *testing.B) {
			b.ReportAllocs()
			b.SetBytes(int64(len(tt.data)))
			r := bytes.NewReader(nil)
			for b.Loop() {
				r.Reset(tt.data)
				_, _ = hasANSI01(r)
			}
		})

		b.Run(tt.name+" 02 #"+strconv.Itoa(n), func(b *testing.B) {
			b.ReportAllocs()
			b.SetBytes(int64(len(tt.data)))
			r := bytes.NewReader(nil)
			for b.Loop() {
				r.Reset(tt.data)
				_, _ = hasANSI02(r)
			}
		})
	}
}

func BenchmarkEOF(b *testing.B) {
	// These patterns share the same iteration to use same input data.
	for n, tt := range ansitests(b) {
		b.Run(tt.name+" EOF 00 #"+strconv.Itoa(n), func(b *testing.B) {
			b.ReportAllocs()
			b.SetBytes(int64(len(tt.data)))
			for b.Loop() {
				_ = trimEOF00(tt.data)
			}
		})
		b.Run(tt.name+" EOF 01 #"+strconv.Itoa(n), func(b *testing.B) {
			b.ReportAllocs()
			b.SetBytes(int64(len(tt.data)))
			for b.Loop() {
				_ = trimEOF01(tt.data)
			}
		})
	}
}

func trimEOF00(s []byte) []byte {
	const (
		e   = 0x8a // è
		eof = 0x1a // msdos end-of-file marker
	)
	match := bytes.LastIndexByte(s, e)
	if none := match == -1; none {
		return s
	}
	for i := match + 1; i < len(s); i++ {
		if s[i] != eof {
			return s
		}
	}
	return s[:match]
}

func trimEOF01(s []byte) []byte {
	const (
		e   = 0x8a // CP437 'è'
		eof = 0x1a // DOS EOF marker
	)

	n := len(s)
	if n == 0 {
		return s
	}

	// 1. Strip all trailing 0x1a (EOF) bytes from the right
	i := n - 1
	for i >= 0 && s[i] == eof {
		i--
	}

	// 2. If the byte immediately preceding the 0x1a sequence is 0x8a ('è'), strip it too
	if i >= 0 && s[i] == e {
		i--
	}

	return s[:i+1]
}

var reRemove01 = regexp.MustCompile(`\x1b\[(?:\?[0-9]+\w|[0-9;]*[a-zA-Z]|[0-9;]* p)|SAUCE00`)

func remove01(b []byte) []byte {
	if len(b) == 0 {
		return b
	}

	// Step 1: Strip ANSI/DEC/Amiga/SAUCE sequences in a single regex sweep
	b = reRemove01.ReplaceAll(b, nil)

	// Step 2: Single-pass byte replacement for CRLF -> LF and NULL -> Space
	// Pre-allocate destination slice to avoid dynamic re-allocations
	out := make([]byte, 0, len(b))
	for i := 0; i < len(b); i++ {
		switch {
		case b[i] == '\x00':
			out = append(out, ' ')
		case b[i] == '\r' && i+1 < len(b) && b[i+1] == '\n':
			out = append(out, '\n')
			i++ // Skip '\n' in CRLF pair
		default:
			out = append(out, b[i])
		}
	}

	// Step 3: Trim trailing bytes in-place (0 allocations)
	return bytes.TrimRight(out, trimCutset)
}

const (
	reAnsi  = `\x1b\[[0-9;]*[a-zA-Z]`
	reAmiga = `\x1b\[[0-9;]* p`
	reDEC   = `\x1b\[\?[0-9]+\w`
	reSauce = `SAUCE00`
)

var (
	byteCR   = []byte("\r")
	byteCRLF = []byte("\r\n")
	// byteCRLFx2 = []byte("\n\n").
	byteEOF    = []byte("\x1a")
	byteLF     = []byte("\n")
	byteNull   = []byte{0x00}
	byteSpace  = []byte(" ")
	emptyBytes = []byte{}
	trimCutset = " \x1a" // Space + SUB character
)

var reControlCodes = regexp.MustCompile(reAnsi + `|` + reDEC + `|` + reAmiga + `|` + reSauce)

// removeControls removes known problematic characters and controls, including:
//   - ASCII control codes, NULL, SUB
//   - Amiga control operators
//   - DEC control codes
//   - ANSI escape codes
//   - Windows and PC DOS newlines
func remove00(b []byte) []byte {
	if len(b) == 0 {
		return b
	}

	b = reControlCodes.ReplaceAll(b, emptyBytes)
	b = bytes.ReplaceAll(b, byteCRLF, byteLF)
	b = bytes.ReplaceAll(b, byteNull, byteSpace)
	b = bytes.TrimRight(b, trimCutset)

	return b
}

const (
	reMove    = `\x1b\[\d*?[ABCDEFG]`
	reMovePos = `\x1b\[\d+;\d+[Hf]`
	reSGR     = `\x1b\[`
)

var reANSI = regexp.MustCompile("(?:" + reMove + "|" + reMovePos + "|" + reSGR + ")") //, reMove, reMovePos, reSGR))

func hasANSI00(r io.Reader) (bool, error) {
	if r == nil {
		return false, nil
	}

	const (
		format    = "match ansi reader %s: %w"
		oneKB     = 1024
		maxBytes  = oneKB * 1024
		chunkSize = oneKB * 32
	)

	buf := make([]byte, chunkSize)
	var read int

	for {
		n, err := r.Read(buf)
		if n > 0 {
			read += n
			if read > maxBytes {
				return false, fmt.Errorf(format, "reader is larger than 1MB", bufio.ErrTooLong)
			}

			if reANSI.Match(buf[:n]) {
				return true, nil
			}
		}

		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return false, fmt.Errorf(format, "chunk read", err)
		}
	}

	return false, nil
}

var (
	ErrTooLong   = fmt.Errorf("reader is larger than 1MB")
	ansiSequence = []byte("\x1b[")
)

func hasANSI01(r io.Reader) (bool, error) {
	if r == nil {
		return false, nil
	}

	const (
		maxBytes  = 1024 * 1024 // 1MB limit
		chunkSize = 8 * 1024    // 8KB fits comfortably on the stack
	)

	// Stack-allocated buffer (0 heap allocations)
	var buf [chunkSize]byte
	var read, carry int

	for {
		// Read into buffer starting after carried-over byte
		n, err := r.Read(buf[carry:])
		if n > 0 {
			total := carry + n
			read += n

			if read > maxBytes {
				return false, fmt.Errorf("match ansi reader: reader is larger than 1MB: %w", ErrTooLong)
			}

			// Fast SIMD search for "\x1b["
			if bytes.Contains(buf[:total], ansiSequence) {
				return true, nil
			}

			// Retain the last byte in case '\x1b' was at the chunk boundary
			if buf[total-1] == '\x1b' {
				buf[0] = '\x1b'
				carry = 1
			} else {
				carry = 0
			}
		}

		if err != nil {
			if err == io.EOF {
				break
			}
			return false, fmt.Errorf("match ansi reader: chunk read: %w", err)
		}
	}

	return false, nil
}

func hasANSI02(r io.Reader) (bool, error) {
	if r == nil {
		return false, nil
	}

	const (
		maxBytes  = 1024 * 1024 // 1MB
		chunkSize = 8 * 1024    // 8KB stack-allocated buffer
	)

	var buf [chunkSize]byte
	var read, carry int

	for {
		n, err := r.Read(buf[carry:])
		if n > 0 {
			total := carry + n
			read += n

			if read > maxBytes {
				return false, fmt.Errorf("match ansi reader: larger than 1MB: %w", ErrTooLong)
			}

			// Scan the buffer for valid sequence
			if matchCSI(buf[:total]) {
				return true, nil
			}

			// Carry over up to 32 bytes in case an ANSI sequence was split across buffer chunks
			carry = copy(buf[:], buf[max(0, total-32):total])
		}

		if err != nil {
			if err == io.EOF {
				break
			}
			return false, fmt.Errorf("match ansi reader: chunk read: %w", err)
		}
	}

	return false, nil
}

func matchCSI(b []byte) bool {
	for i := range len(b) {
		// Fast skip to '\x1b'
		if b[i] != '\x1b' {
			continue
		}

		// Check for '[' after '\x1b'
		if i+1 >= len(b) || b[i+1] != '[' {
			continue
		}

		// Replicates `reSGR` (\x1b\[) or validates sequence params
		// If you ONLY want specific sequences (e.g. Move or MovePos, but NOT bare \x1b\[):
		// Parse parameter digits/semicolons and validate terminal char:

		idx := i + 2
		// 1. Scan optional numbers and semicolons
		for idx < len(b) && ((b[idx] >= '0' && b[idx] <= '9') || b[idx] == ';') {
			idx++
		}

		// 2. Validate final terminal character
		if idx < len(b) {
			term := b[idx]
			// Check for move (A-G), pos (H, f), or SGR/other CSI terminals (m, etc.)
			if (term >= 'A' && term <= 'G') || term == 'H' || term == 'f' || term == 'm' {
				return true
			}
		}
	}
	return false
}

func hasANSI03(r io.Reader) (bool, error) {
	if r == nil {
		return false, nil
	}

	const (
		maxBytes  = 1024 * 1024 // 1MB
		chunkSize = 8 * 1024    // 8KB stack buffer (0 allocs)
	)

	var buf [chunkSize]byte
	var read, carry int

	for {
		n, err := r.Read(buf[carry:])
		if n > 0 {
			total := carry + n
			read += n

			if read > maxBytes {
				return false, fmt.Errorf("match ansi reader: reader is larger than 1MB: %w", ErrTooLong)
			}

			matched, matchedIdx := matchCSIPlus(buf[:total])
			if matched {
				return true, nil
			}

			// Smart carryover: only carry over trailing bytes if an ESC (\x1b) was seen near the tail
			carry = 0
			if matchedIdx >= 0 && total-matchedIdx < 32 {
				carry = copy(buf[:], buf[matchedIdx:total])
			}
		}

		if err != nil {
			if err == io.EOF {
				break
			}
			return false, fmt.Errorf("match ansi reader: chunk read: %w", err)
		}
	}

	return false, nil
}

// matchCSI uses SIMD (bytes.IndexByte) to jump between ESC characters,
// eliminating manual byte-by-byte iteration over non-ANSI payload bytes.
func matchCSIPlus(b []byte) (bool, int) {
	offset := 0
	lastEsc := -1

	for offset < len(b) {
		// SIMD-accelerated jump to the next ESC (\x1b) byte
		idx := bytes.IndexByte(b[offset:], '\x1b')
		if idx == -1 {
			break
		}

		absIdx := offset + idx
		lastEsc = absIdx
		offset = absIdx + 1

		// Need at least '[' after '\x1b'
		if offset >= len(b) || b[offset] != '[' {
			continue
		}
		offset++ // Skip '['

		// Parse parameter bytes (digits and semicolons)
		paramStart := offset
		for offset < len(b) && ((b[offset] >= '0' && b[offset] <= '9') || b[offset] == ';') {
			offset++
		}

		// Check terminal character if present in current buffer
		if offset < len(b) {
			term := b[offset]
			// Strict match matching original intent:
			// - Move: \x1b\[\d*?[ABCDEFG]
			// - Pos:  \x1b\[\d+;\d+[Hf]
			// - SGR / bare CSI: \x1b\[
			if (term >= 'A' && term <= 'G') || term == 'H' || term == 'f' || term == 'm' || (offset == paramStart) {
				return true, absIdx
			}
		}
	}

	return false, lastEsc
}
