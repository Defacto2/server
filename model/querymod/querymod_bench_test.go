package querymod

import (
	"testing"
)

// Benchmarks for tag functions with caching optimization.

func BenchmarkSAdvert(b *testing.B) {
	for range b.N {
		_ = SAdvert()
	}
}

func BenchmarkPWindows(b *testing.B) {
	for range b.N {
		_ = PWindows()
	}
}

func BenchmarkPDos(b *testing.B) {
	for range b.N {
		_ = PDos()
	}
}

func BenchmarkAdvertExpr(b *testing.B) {
	for range b.N {
		_ = AdvertExpr()
	}
}

func BenchmarkDOSExpr(b *testing.B) {
	for range b.N {
		_ = DOSExpr()
	}
}

func BenchmarkAnsiBBSExpr(b *testing.B) {
	for range b.N {
		_ = AnsiBBSExpr()
	}
}

func BenchmarkWindowsPackExpr(b *testing.B) {
	for range b.N {
		_ = WindowsPackExpr()
	}
}

func BenchmarkURICaching(b *testing.B) {
	// Benchmark that getURIs() returns cached result on subsequent calls.
	getURIs() // Prime the cache.
	b.ResetTimer()
	for range b.N {
		_ = getURIs()
	}
}

// Parallel benchmark to ensure thread-safe caching.
func BenchmarkURICachingParallel(b *testing.B) {
	getURIs() // Prime the cache.
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = getURIs()
		}
	})
}
