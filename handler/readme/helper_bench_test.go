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

	"github.com/Defacto2/server/internal/testutil"
)

// Use the following command to run all:
// go test -bench=Benchmark -benchmem

func BenchmarkRM(b *testing.B) {
	for n, tt := range testutil.ANSITests(b) {
		b.Run(tt.Name+" 00 #"+strconv.Itoa(n), func(b *testing.B) {
			b.ReportAllocs()
			b.SetBytes(int64(len(tt.Data)))
			for b.Loop() {
				_ = remove00(tt.Data)
			}
		})
		b.Run(tt.Name+" 01 #"+strconv.Itoa(n), func(b *testing.B) {
			b.ReportAllocs()
			b.SetBytes(int64(len(tt.Data)))
			for b.Loop() {
				_ = remove01(tt.Data)
			}
		})
	}
}

func BenchmarkHas(b *testing.B) {
	// These patterns share the same iteration to use same input data.
	for n, tt := range testutil.ANSITests(b) {
		b.Run(tt.Name+" 00 #"+strconv.Itoa(n), func(b *testing.B) {
			b.ReportAllocs()
			b.SetBytes(int64(len(tt.Data)))
			r := bytes.NewReader(nil)
			for b.Loop() {
				r.Reset(tt.Data)
				_, _ = hasANSI00(r)
			}
		})

		b.Run(tt.Name+" 01 #"+strconv.Itoa(n), func(b *testing.B) {
			b.ReportAllocs()
			b.SetBytes(int64(len(tt.Data)))
			r := bytes.NewReader(nil)
			for b.Loop() {
				r.Reset(tt.Data)
				_, _ = hasANSI01(r)
			}
		})

		b.Run(tt.Name+" 02 #"+strconv.Itoa(n), func(b *testing.B) {
			b.ReportAllocs()
			b.SetBytes(int64(len(tt.Data)))
			r := bytes.NewReader(nil)
			for b.Loop() {
				r.Reset(tt.Data)
				_, _ = hasANSI02(r)
			}
		})
	}
}

func BenchmarkEOF(b *testing.B) {
	// These patterns share the same iteration to use same input data.
	for n, tt := range testutil.ANSITests(b) {
		b.Run(tt.Name+" EOF 00 #"+strconv.Itoa(n), func(b *testing.B) {
			b.ReportAllocs()
			b.SetBytes(int64(len(tt.Data)))
			for b.Loop() {
				_ = trimEOF00(tt.Data)
			}
		})
		b.Run(tt.Name+" EOF 01 #"+strconv.Itoa(n), func(b *testing.B) {
			b.ReportAllocs()
			b.SetBytes(int64(len(tt.Data)))
			for b.Loop() {
				_ = trimEOF01(tt.Data)
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

	i := n - 1
	for i >= 0 && s[i] == eof {
		i--
	}

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

	b = reRemove01.ReplaceAll(b, nil)

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

	return bytes.TrimRight(out, trimCutset)
}

const (
	reAnsi  = `\x1b\[[0-9;]*[a-zA-Z]`
	reAmiga = `\x1b\[[0-9;]* p`
	reDEC   = `\x1b\[\?[0-9]+\w`
	reSauce = `SAUCE00`
)

var (
	byteCRLF   = []byte("\r\n")
	byteLF     = []byte("\n")
	byteNull   = []byte{0x00}
	byteSpace  = []byte(" ")
	emptyBytes = []byte{}
	trimCutset = " \x1a" // Space + SUB character
)

var reControlCodes = regexp.MustCompile(reAnsi + `|` + reDEC + `|` + reAmiga + `|` + reSauce)

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

var reANSI = regexp.MustCompile("(?:" + reMove + "|" + reMovePos + "|" + reSGR + ")")

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
	ErrTooLong   = errors.New("reader is larger than 1MB")
	ansiSequence = []byte("\x1b[")
)

const (
	maxBytes  = 1024 * 1024 // 1MB limit
	chunkSize = 8 * 1024    // 8KB for stack
)

func hasANSI01(r io.Reader) (bool, error) {
	if r == nil {
		return false, nil
	}

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

			if bytes.Contains(buf[:total], ansiSequence) {
				return true, nil
			}

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

			if matchCSI(buf[:total]) {
				return true, nil
			}

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
	for i := range b {
		if b[i] != '\x1b' {
			continue
		}

		if i+1 >= len(b) || b[i+1] != '[' {
			continue
		}

		idx := i + 2
		for idx < len(b) && ((b[idx] >= '0' && b[idx] <= '9') || b[idx] == ';') {
			idx++
		}

		if idx < len(b) {
			term := b[idx]
			if (term >= 'A' && term <= 'G') || term == 'H' || term == 'f' || term == 'm' {
				return true
			}
		}
	}
	return false
}
