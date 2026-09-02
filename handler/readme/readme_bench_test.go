package readme_test

import (
	"bytes"
	"strconv"
	"testing"

	"github.com/Defacto2/helper"
	"github.com/Defacto2/server/internal/testutil"
)

// Benchmarks ending with 00 were the original unoptimized implementations
// in use until September 2026. The patterns numbered 01+ were different
// suggestions sourced from Gemini and online.
//
// Use the following command to run all:
// go test -bench=Benchmark -benchmem

func BenchmarkNorm(b *testing.B) {
	for n, tt := range testutil.ANSITests(b) {
		b.Run(tt.Name+" 00 #"+strconv.Itoa(n), func(b *testing.B) {
			b.ReportAllocs()
			b.SetBytes(int64(len(tt.Data)))
			for b.Loop() {
				textBuf := bytes.NewBuffer(tt.Data)
				_ = norm00(textBuf)
				textBuf.Reset()
			}
		})
		b.Run(tt.Name+" 01 #"+strconv.Itoa(n), func(b *testing.B) {
			b.ReportAllocs()
			b.SetBytes(int64(len(tt.Data)))
			for b.Loop() {
				textBuf := bytes.NewBuffer(tt.Data)
				_ = norm01(textBuf)
				textBuf.Reset()
			}
		})
		b.Run(tt.Name+" 02 #"+strconv.Itoa(n), func(b *testing.B) {
			b.ReportAllocs()
			b.SetBytes(int64(len(tt.Data)))
			for b.Loop() {
				textBuf := bytes.NewBuffer(tt.Data)
				_ = norm02(textBuf)
				textBuf.Reset()
			}
		})
	}
}

func norm00(textBuf *bytes.Buffer) []byte {
	if textBuf == nil || textBuf.Len() == 0 {
		return []byte{}
	}

	b := textBuf.Bytes()
	b = bytes.ReplaceAll(b, byteNull, byteSpace)
	b = bytes.ReplaceAll(b, byteCRLF, byteLF)
	b = bytes.TrimRight(b, "\x1a")
	return helper.Mask(b...)
}

func norm01(textBuf *bytes.Buffer) []byte {
	if textBuf == nil || textBuf.Len() == 0 {
		return nil
	}

	b := textBuf.Bytes()

	end := len(b)
	for end > 0 && b[end-1] == '\x1a' {
		end--
	}

	w := 0
	for r := 0; r < end; r++ {
		switch {
		case b[r] == '\x00':
			b[w] = ' '
			w++
		case b[r] == '\r' && r+1 < end && b[r+1] == '\n':
			b[w] = '\n'
			w++
			r++
		default:
			b[w] = b[r]
			w++
		}
	}

	return helper.Mask(b[:w]...)
}

func norm02(textBuf *bytes.Buffer) []byte {
	if textBuf == nil || textBuf.Len() == 0 {
		return nil
	}

	src := textBuf.Bytes()
	end := len(src)
	for end > 0 && src[end-1] == '\x1a' {
		end--
	}
	src = src[:end]

	if len(src) == 0 {
		return nil
	}
	dst := make([]byte, 0, len(src))
	for i := 0; i < len(src); i++ {
		switch {
		case src[i] == '\x00':
			dst = append(dst, ' ')
		case src[i] == '\r' && i+1 < len(src) && src[i+1] == '\n':
			dst = append(dst, '\n')
			i++
		default:
			dst = append(dst, src[i])
		}
	}

	return helper.Mask(dst...)
}
