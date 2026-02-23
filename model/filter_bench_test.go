package model_test

import (
	"testing"

	"github.com/Defacto2/server/model"
)

// Benchmark model.GetColumns() caching optimization.
func BenchmarkGetColumns(b *testing.B) {
	for range b.N {
		_ = model.GetColumns()
	}
}

// Benchmark that subsequent calls hit the cache.
func BenchmarkGetColumnsCached(b *testing.B) {
	model.GetColumns() // Prime the cache
	b.ResetTimer()
	for range b.N {
		_ = model.GetColumns()
	}
}

// Benchmark parallel model.GetColumns() access for thread-safety.
func BenchmarkGetColumnsParallel(b *testing.B) {
	model.GetColumns() // Prime the cache
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = model.GetColumns()
		}
	})
}
