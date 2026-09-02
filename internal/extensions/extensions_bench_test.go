package extensions_test

import (
	"slices"
	"testing"

	"github.com/Defacto2/server/internal/extensions"
)

func BenchmarkArchive(b *testing.B) {
	b.Run("", func(b *testing.B) {
		for range b.N {
			_ = extensions.Archive()
		}
	})
}

func BenchmarkDocument(b *testing.B) {
	b.Run("", func(b *testing.B) {
		for range b.N {
			_ = extensions.Document()
		}
	})
}

func BenchmarkImage(b *testing.B) {
	b.Run("", func(b *testing.B) {
		for range b.N {
			_ = extensions.Image()
		}
	})
}

func BenchmarkMedia(b *testing.B) {
	b.Run("", func(b *testing.B) {
		for range b.N {
			_ = extensions.Media()
		}
	})
}

func BenchmarkContainsArchive(b *testing.B) {
	b.Run("", func(b *testing.B) {
		for range b.N {
			_ = slices.Contains(extensions.Archive(), ".rar")
		}
	})
}

func BenchmarkContainsDocument(b *testing.B) {
	b.Run("", func(b *testing.B) {
		for range b.N {
			_ = slices.Contains(extensions.Document(), ".pdf")
		}
	})
}

func BenchmarkContainsImage(b *testing.B) {
	b.Run("", func(b *testing.B) {
		for range b.N {
			_ = slices.Contains(extensions.Image(), ".png")
		}
	})
}

func BenchmarkContainsMedia(b *testing.B) {
	b.Run("", func(b *testing.B) {
		for range b.N {
			_ = slices.Contains(extensions.Media(), ".mp3")
		}
	})
}

func BenchmarkFileListingWith100Files(b *testing.B) {
	exts := [...]string{".zip", ".pdf", ".png", ".mp3", ".rar", ".txt", ".jpg", ".gif", ".7z"}
	b.Run("", func(b *testing.B) {
		for range b.N {
			for _, ext := range exts {
				_ = slices.Contains(extensions.Archive(), ext)
				_ = slices.Contains(extensions.Document(), ext)
				_ = slices.Contains(extensions.Image(), ext)
				_ = slices.Contains(extensions.Media(), ext)
			}
		}
	})
}

func BenchmarkAllFunctionsCalled(b *testing.B) {
	b.Run("", func(b *testing.B) {
		for range b.N {
			_ = extensions.Archive()
			_ = extensions.Document()
			_ = extensions.Image()
			_ = extensions.Media()
		}
	})
}
