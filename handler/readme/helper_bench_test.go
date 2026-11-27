//nolint:testpackage
package readme

import (
	"bytes"
	"strings"
	"testing"

	"github.com/Defacto2/server/internal/testutil"
)

func BenchmarkAddPrefix(b *testing.B) {
	prefix := []byte("FILE_ID.DIZ content")
	for _, tt := range testutil.ANSITests(b) {
		b.Run(strings.TrimSpace(tt.Name), func(b *testing.B) {
			b.SetBytes(int64(len(tt.Data)))
			for b.Loop() {
				_ = AddPrefix(tt.Data, prefix)
			}
		})
	}
}

func BenchmarkAddSuffix(b *testing.B) {
	suffix := []byte("helper text content")
	for _, tt := range testutil.ANSITests(b) {
		b.Run(strings.TrimSpace(tt.Name), func(b *testing.B) {
			b.SetBytes(int64(len(tt.Data)))
			for b.Loop() {
				_ = AddSuffix(tt.Data, suffix)
			}
		})
	}
}

func BenchmarkTrimE(b *testing.B) {
	for _, tt := range testutil.ANSITests(b) {
		b.Run(strings.TrimSpace(tt.Name), func(b *testing.B) {
			b.SetBytes(int64(len(tt.Data)))
			for b.Loop() {
				_ = trimE(tt.Data)
			}
		})
	}
}

func BenchmarkTrimBytes(b *testing.B) {
	for _, tt := range testutil.ANSITests(b) {
		b.Run(strings.TrimSpace(tt.Name), func(b *testing.B) {
			b.SetBytes(int64(len(tt.Data)))
			for b.Loop() {
				_ = trimBytes(tt.Data)
			}
		})
	}
}

func BenchmarkRemoveControls(b *testing.B) {
	for _, tt := range testutil.ANSITests(b) {
		b.Run(strings.TrimSpace(tt.Name), func(b *testing.B) {
			b.SetBytes(int64(len(tt.Data)))
			for b.Loop() {
				_ = removeControls(tt.Data)
			}
		})
	}
}

func BenchmarkHasANSI(b *testing.B) {
	for _, tt := range testutil.ANSITests(b) {
		b.Run(strings.TrimSpace(tt.Name), func(b *testing.B) {
			b.SetBytes(int64(len(tt.Data)))
			r := bytes.NewReader(nil)
			for b.Loop() {
				r.Reset(tt.Data)
				_, _ = hasANSI(r)
			}
		})
	}
}
