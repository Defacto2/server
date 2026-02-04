package model

import (
	"testing"
)

// Benchmark getColumns() caching optimization.
func BenchmarkGetColumns(b *testing.B) {
	for range b.N {
		_ = getColumns()
	}
}

// Benchmark that subsequent calls hit the cache.
func BenchmarkGetColumnsCached(b *testing.B) {
	getColumns() // Prime the cache
	b.ResetTimer()
	for range b.N {
		_ = getColumns()
	}
}

// Benchmark parallel getColumns() access for thread-safety.
func BenchmarkGetColumnsParallel(b *testing.B) {
	getColumns() // Prime the cache
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = getColumns()
		}
	})
}
