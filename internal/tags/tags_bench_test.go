package tags_test

import (
	"strings"
	"testing"

	"github.com/Defacto2/server/internal/tags"
)

// Use the following command to run all:
// go test -bench=Benchmark -benchmem

func BenchmarkTagByURI00(b *testing.B) {
	slugs := []string{"dos", "windows", "demo", "ansi", "text", "image", "java", "linux"}
	b.Run("", func(b *testing.B) {
		for range b.N {
			for _, slug := range slugs {
				tagbyuri00(b, slug)
			}
		}
	})
}

func BenchmarkTagByURI01(b *testing.B) {
	slugs := []string{"dos", "windows", "demo", "ansi", "text", "image", "java", "linux"}
	b.Run("", func(b *testing.B) {
		for range b.N {
			for _, slug := range slugs {
				_ = tags.TagByURI(slug)
			}
		}
	})
}

func tagbyuri00(b *testing.B, slug string) tags.Tag {
	b.Helper()
	for key, value := range tags.URIs() {
		if strings.ToLower(slug) == value {
			return key
		}
	}
	return -1
}

func BenchmarkURIsCalls(b *testing.B) {
	b.ReportAllocs()
	b.Run("", func(b *testing.B) {
		for range b.N {
			_ = tags.URIs()
		}
	})
}

func BenchmarkNamesCalls(b *testing.B) {
	b.ReportAllocs()
	b.Run("", func(b *testing.B) {
		for range b.N {
			_ = tags.Names()
		}
	})
}

// BenchmarkInfosCalls tests the optimized cached map.
func BenchmarkInfosCalls(b *testing.B) {
	b.ReportAllocs()
	b.Run("", func(b *testing.B) {
		for range b.N {
			_ = tags.Infos()
		}
	})
}

func BenchmarkDeterminerCalls(b *testing.B) {
	b.ReportAllocs()
	b.Run("", func(b *testing.B) {
		for range b.N {
			_ = tags.Determiner()
		}
	})
}

func BenchmarkBuildSimulation(b *testing.B) {
	b.ReportAllocs()
	b.Run("", func(b *testing.B) {
		for range b.N {
			for range 40 {
				_ = tags.URIs()
				_ = tags.Names()
				_ = tags.Infos()
			}
		}
	})
}

func BenchmarkIsCat00(b *testing.B) {
	names := []string{"announcements", "demo", "text", "ansi", "linux"}
	b.Run("", func(b *testing.B) {
		for range b.N {
			for _, name := range names {
				iscat00(name)
			}
		}
	})
}

func BenchmarkIsCat01(b *testing.B) {
	names := []string{"announcements", "demo", "text", "ansi", "linux"}
	b.Run("", func(b *testing.B) {
		for range b.N {
			for _, name := range names {
				_ = tags.IsCategory(name)
			}
		}
	})
}

func iscat00(name string) bool {
	name = strings.TrimSpace(name)
	if name == "" {
		return false
	}
	for _, tag := range tags.List {
		if strings.EqualFold(tag.String(), name) {
			return tag >= tags.FirstCategory && tag <= tags.LastCategory
		}
	}
	return false
}

func BenchmarkIsPlatform(b *testing.B) {
	names := []string{"ansi", "dos", "windows", "linux", "java"}
	b.Run("", func(b *testing.B) {
		for range b.N {
			for _, name := range names {
				_ = tags.IsPlatform(name)
			}
		}
	})
}

func BenchmarkIsTag(b *testing.B) {
	names := []string{"ansi", "demo", "windows", "text", "java"}
	b.Run("", func(b *testing.B) {
		for range b.N {
			for _, name := range names {
				_ = tags.IsTag(name)
			}
		}
	})
}
