package tags_test

import (
	"strings"
	"testing"

	"github.com/Defacto2/server/internal/tags"
)

// BenchmarkTagByURI tests the optimized O(1) lookup
func BenchmarkTagByURI(b *testing.B) {
	slugs := []string{"dos", "windows", "demo", "ansi", "text", "image", "java", "linux"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, slug := range slugs {
			_ = tags.TagByURI(slug)
		}
	}
}

// BenchmarkTagByURILinearSearch simulates the old O(n) approach for comparison
func BenchmarkTagByURILinearSearch(b *testing.B) {
	slugs := []string{"dos", "windows", "demo", "ansi", "text", "image", "java", "linux"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, slug := range slugs {
			oldTagByURI(slug)
		}
	}
}

// oldTagByURI is the original O(n) implementation for benchmarking comparison
func oldTagByURI(slug string) tags.Tag {
	for key, value := range tags.URIs() {
		if strings.ToLower(slug) == value {
			return key
		}
	}
	return -1
}

// BenchmarkURIsCalls tests the optimized cached map
func BenchmarkURIsCalls(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = tags.URIs()
	}
}

// BenchmarkNamesCalls tests the optimized cached map
func BenchmarkNamesCalls(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = tags.Names()
	}
}

// BenchmarkInfosCalls tests the optimized cached map
func BenchmarkInfosCalls(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = tags.Infos()
	}
}

// BenchmarkDeterminerCalls tests the optimized cached map
func BenchmarkDeterminerCalls(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = tags.Determiner()
	}
}

// BenchmarkBuildSimulation simulates old Build() accessing maps 40+ times each
func BenchmarkBuildSimulation(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		for j := 0; j < 40; j++ {
			_ = tags.URIs()
			_ = tags.Names()
			_ = tags.Infos()
		}
	}
}

// BenchmarkIsCategory tests the optimized O(1) lookup
func BenchmarkIsCategory(b *testing.B) {
	names := []string{"announcements", "demo", "text", "ansi", "linux"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, name := range names {
			_ = tags.IsCategory(name)
		}
	}
}

// BenchmarkIsCategoryOld simulates old O(n) implementation
func BenchmarkIsCategoryOld(b *testing.B) {
	names := []string{"announcements", "demo", "text", "ansi", "linux"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, name := range names {
			oldIsCategory(name)
		}
	}
}

// oldIsCategory simulates the old O(n) iteration
func oldIsCategory(name string) bool {
	name = strings.TrimSpace(name)
	if name == "" {
		return false
	}
	for _, tag := range tags.List() {
		if strings.EqualFold(tag.String(), name) {
			return tag >= tags.FirstCategory && tag <= tags.LastCategory
		}
	}
	return false
}

// BenchmarkIsPlatform tests the optimized O(1) lookup
func BenchmarkIsPlatform(b *testing.B) {
	names := []string{"ansi", "dos", "windows", "linux", "java"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, name := range names {
			_ = tags.IsPlatform(name)
		}
	}
}

// BenchmarkIsTag tests the optimized O(1) lookup
func BenchmarkIsTag(b *testing.B) {
	names := []string{"ansi", "demo", "windows", "text", "java"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, name := range names {
			_ = tags.IsTag(name)
		}
	}
}

