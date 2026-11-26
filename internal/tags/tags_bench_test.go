package tags_test

import (
	"strings"
	"testing"

	"github.com/Defacto2/server/internal/tags"
)

// Use the following command to run all:
// go test -bench=Benchmark -benchmem

func BenchmarkTagByURI(b *testing.B) {
	slugs := [...]string{"dos", "Windows", "demo", "ANSI", "text", "image", "Java", "linux"}
	b.Run("linear", func(b *testing.B) {
		for b.Loop() {
			for _, slug := range slugs {
				_ = tagbyuri00(slug)
			}
		}
	})
	b.Run("map", func(b *testing.B) {
		for b.Loop() {
			for _, slug := range slugs {
				_ = tags.TagByURI(slug)
			}
		}
	})
}

func tagbyuri00(slug string) tags.Tag {
	slug = strings.ToLower(slug)
	for key, value := range tags.URIs() {
		if slug == value {
			return key
		}
	}
	return -1
}

func BenchmarkURIsCalls(b *testing.B) {
	for b.Loop() {
		_ = tags.URIs()
	}
}

func BenchmarkNamesCalls(b *testing.B) {
	for b.Loop() {
		_ = tags.Names()
	}
}

// BenchmarkInfosCalls tests the optimized cached map.
func BenchmarkInfosCalls(b *testing.B) {
	for b.Loop() {
		_ = tags.Infos()
	}
}

func BenchmarkDeterminerCalls(b *testing.B) {
	for b.Loop() {
		_ = tags.Determiner()
	}
}

func BenchmarkBuildSimulation(b *testing.B) {
	for b.Loop() {
		for range 40 {
			_ = tags.URIs()
			_ = tags.Names()
			_ = tags.Infos()
		}
	}
}

func BenchmarkIsCategory(b *testing.B) {
	names := [...]string{"announcements", " DEMO ", "text", "ANSI", "linux"}
	b.Run("linear", func(b *testing.B) {
		for b.Loop() {
			for _, name := range names {
				_ = iscat00(name)
			}
		}
	})
	b.Run("map", func(b *testing.B) {
		for b.Loop() {
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
	names := [...]string{"ansi", "DOS", " WINDOWS ", "linux", "java"}
	for b.Loop() {
		for _, name := range names {
			_ = tags.IsPlatform(name)
		}
	}
}

func BenchmarkIsTag(b *testing.B) {
	names := [...]string{"ansi", " DEMO ", "WINDOWS", "text", "java"}
	for b.Loop() {
		for _, name := range names {
			_ = tags.IsTag(name)
		}
	}
}
