package extensions_test

import (
	"slices"
	"testing"

	"github.com/Defacto2/server/internal/extensions"
)

// Benchmark single calls to each function.
func BenchmarkArchive(b *testing.B) {
	for range b.N {
		_ = extensions.Archive()
	}
}

func BenchmarkDocument(b *testing.B) {
	for range b.N {
		_ = extensions.Document()
	}
}

func BenchmarkImage(b *testing.B) {
	for range b.N {
		_ = extensions.Image()
	}
}

func BenchmarkMedia(b *testing.B) {
	for range b.N {
		_ = extensions.Media()
	}
}

// Benchmark realistic use case: checking if extension is in archive list.
// This simulates the common pattern in filerecord.go and simple.go.
func BenchmarkContainsArchive(b *testing.B) {
	for range b.N {
		_ = slices.Contains(extensions.Archive(), ".rar")
	}
}

// Benchmark realistic use case: checking if extension is in document list.
func BenchmarkContainsDocument(b *testing.B) {
	for range b.N {
		_ = slices.Contains(extensions.Document(), ".pdf")
	}
}

// Benchmark realistic use case: checking if extension is in image list.
func BenchmarkContainsImage(b *testing.B) {
	for range b.N {
		_ = slices.Contains(extensions.Image(), ".png")
	}
}

// Benchmark realistic use case: checking if extension is in media list.
func BenchmarkContainsMedia(b *testing.B) {
	for range b.N {
		_ = slices.Contains(extensions.Media(), ".mp3")
	}
}

// Benchmark simulating a realistic page with 100 files being checked.
// This shows the cumulative impact of the optimization.
func BenchmarkFileListingWith100Files(b *testing.B) {
	exts := []string{".zip", ".pdf", ".png", ".mp3", ".rar", ".txt", ".jpg", ".gif", ".7z"}
	b.ResetTimer()
	for range b.N {
		for _, ext := range exts {
			_ = slices.Contains(extensions.Archive(), ext)
			_ = slices.Contains(extensions.Document(), ext)
			_ = slices.Contains(extensions.Image(), ext)
			_ = slices.Contains(extensions.Media(), ext)
		}
	}
}

// Benchmark all four functions called together (simulating type detection).
func BenchmarkAllFunctionsCalled(b *testing.B) {
	for range b.N {
		_ = extensions.Archive()
		_ = extensions.Document()
		_ = extensions.Image()
		_ = extensions.Media()
	}
}
