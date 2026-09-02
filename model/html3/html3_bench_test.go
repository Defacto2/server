package html3_test

import (
	"strings"
	"testing"

	"github.com/Defacto2/server/model/html3"
)

// Use the following command to run all:
// go test -bench=Benchmark -benchmem

// BenchmarkOrderStringNew uses package-level map.
func BenchmarkOrderStringNew(b *testing.B) {
	b.Run("", func(b *testing.B) {
		for range b.N {
			// Optimized: just call String() which uses pre-allocated map
			_ = html3.NameAsc.String()
		}
	})
}

// BenchmarkOrderStringLookup tests rapid order lookups.
func BenchmarkOrderStringLookup(b *testing.B) {
	orders := []html3.Order{
		html3.NameAsc, html3.NameDes,
		html3.PublAsc, html3.PublDes,
		html3.PostAsc, html3.PostDes,
		html3.SizeAsc, html3.SizeDes,
		html3.DescAsc, html3.DescDes,
	}
	b.Run("", func(b *testing.B) {
		for range b.N {
			for _, o := range orders {
				_ = o.String()
			}
		}
	})
}

// BenchmarkCreated tests the Created() function performance.
func BenchmarkCreated(b *testing.B) {
	// This is a fast function, but benchmark for regression detection
	b.Run("", func(b *testing.B) {
		for range b.N {
			_ = html3.Created(nil)
		}
	})
}

// BenchmarkPublished tests the Published() function performance.
func BenchmarkPublished(b *testing.B) {
	b.Run("", func(b *testing.B) {
		for range b.N {
			_ = html3.Published(nil)
		}
	})
}

// BenchmarkLeadStr tests string padding performance.
func BenchmarkLeadStr(b *testing.B) {
	b.Run("", func(b *testing.B) {
		for range b.N {
			_ = html3.LeadStr(20, "test")
		}
	})
}

// BenchmarkIcon tests icon lookup performance.
func BenchmarkIcon(b *testing.B) {
	b.Run("", func(b *testing.B) {
		for range b.N {
			_ = html3.Icon(nil)
		}
	})
}

// BenchmarkLeadStrOld simulates old string allocation every call.
func BenchmarkLeadStrOld(b *testing.B) {
	b.Run("", func(b *testing.B) {
		for range b.N {
			// Simulating old approach: strings.Repeat every call
			_ = strings.Repeat(" ", 3)
			_ = strings.Repeat(" ", 7)
		}
	})
}

// BenchmarkLeadStrOptimized uses pre-computed padding.
func BenchmarkLeadStrOptimized(b *testing.B) {
	b.Run("", func(b *testing.B) {
		for range b.N {
			// Optimized: pre-computed constants
			_ = html3.LeadStr(3, "x")
			_ = html3.LeadStr(7, "y")
		}
	})
}

// BenchmarkPublishedWithFlags demonstrates state machine approach.
func BenchmarkPublishedOptimized(b *testing.B) {
	b.Run("", func(b *testing.B) {
		for range b.N {
			_ = html3.Published(nil)
		}
	})
}
