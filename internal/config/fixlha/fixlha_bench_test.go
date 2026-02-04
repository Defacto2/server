package fixlha

import (
	"io"
	"log/slog"
	"testing"

	"github.com/Defacto2/server/internal/dir"
)

// BenchmarkCheckValid benchmarks the Check function with a valid file.
func BenchmarkCheckValid(b *testing.B) {
	sl := slog.New(slog.NewTextHandler(io.Discard, nil))
	extra := dir.Directory(b.TempDir())
	d := &MockDirEntry{name: "12345678-1234-1234-1234-123456789012.zip", isDir: false}
	artifacts := []string{"12345678-1234-1234-1234-123456789012"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Check(sl, extra, d, artifacts...)
	}
}

// BenchmarkCheckInvalidExtension benchmarks the Check function with an invalid extension.
func BenchmarkCheckInvalidExtension(b *testing.B) {
	sl := slog.New(slog.NewTextHandler(io.Discard, nil))
	extra := dir.Directory(b.TempDir())
	d := &MockDirEntry{name: "12345678-1234-1234-1234-123456789012.lha", isDir: false}
	artifacts := []string{"12345678-1234-1234-1234-123456789012"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Check(sl, extra, d, artifacts...)
	}
}

// BenchmarkCheckDirectory benchmarks the Check function with a directory.
func BenchmarkCheckDirectory(b *testing.B) {
	sl := slog.New(slog.NewTextHandler(io.Discard, nil))
	extra := dir.Directory(b.TempDir())
	d := &MockDirEntry{name: "somedir", isDir: true}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Check(sl, extra, d)
	}
}

// BenchmarkCheckUppercaseExtension benchmarks the Check function with uppercase extension.
func BenchmarkCheckUppercaseExtension(b *testing.B) {
	sl := slog.New(slog.NewTextHandler(io.Discard, nil))
	extra := dir.Directory(b.TempDir())
	d := &MockDirEntry{name: "12345678-1234-1234-1234-123456789012.ZIP", isDir: false}
	artifacts := []string{"12345678-1234-1234-1234-123456789012"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Check(sl, extra, d, artifacts...)
	}
}

// BenchmarkCheckManyArtifacts benchmarks the Check function with many artifacts.
func BenchmarkCheckManyArtifacts(b *testing.B) {
	sl := slog.New(slog.NewTextHandler(io.Discard, nil))
	extra := dir.Directory(b.TempDir())
	d := &MockDirEntry{name: "12345678-1234-1234-1234-123456789012.zip", isDir: false}

	// Generate many sorted artifacts (binary search requires sorted)
	artifacts := make([]string, 0, 100)
	for i := 0; i < 100; i++ {
		artifacts = append(artifacts, "00000000-0000-0000-0000-000000000000"+string(rune('a'+i%26)))
	}
	artifacts = append(artifacts, "12345678-1234-1234-1234-123456789012")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Check(sl, extra, d, artifacts...)
	}
}
