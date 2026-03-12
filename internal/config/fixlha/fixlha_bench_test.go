package fixlha_test

import (
	"log/slog"
	"testing"

	"github.com/Defacto2/server/internal/config/fixlha"
	"github.com/Defacto2/server/internal/config/testconst"
	"github.com/Defacto2/server/internal/dir"
)

// BenchmarkCheckValid benchmarks the Check function with a valid file.
func BenchmarkCheckValid(b *testing.B) {
	sl := slog.New(slog.DiscardHandler)
	extra := dir.Directory(b.TempDir())
	d := &MockDirEntry{name: testconst.TestUUID + ".zip", isDir: false}
	artifacts := []string{testconst.TestUUID}

	b.Run("", func(b *testing.B) {
		for range b.N {
			_ = fixlha.Check(sl, extra, d, artifacts...)
		}
	})
}

// BenchmarkCheckInvalidExtension benchmarks the Check function with an invalid extension.
func BenchmarkCheckInvalidExtension(b *testing.B) {
	sl := slog.New(slog.DiscardHandler)
	extra := dir.Directory(b.TempDir())
	d := &MockDirEntry{name: testconst.TestUUID + ".lha", isDir: false}
	artifacts := []string{testconst.TestUUID}

	b.Run("", func(b *testing.B) {
		for range b.N {
			_ = fixlha.Check(sl, extra, d, artifacts...)
		}
	})
}

// BenchmarkCheckDirectory benchmarks the Check function with a directory.
func BenchmarkCheckDirectory(b *testing.B) {
	sl := slog.New(slog.DiscardHandler)
	extra := dir.Directory(b.TempDir())
	d := &MockDirEntry{name: "somedir", isDir: true}

	b.Run("", func(b *testing.B) {
		for range b.N {
			_ = fixlha.Check(sl, extra, d)
		}
	})
}

// BenchmarkCheckUppercaseExtension benchmarks the Check function with uppercase extension.
func BenchmarkCheckUppercaseExtension(b *testing.B) {
	sl := slog.New(slog.DiscardHandler)
	extra := dir.Directory(b.TempDir())
	d := &MockDirEntry{name: testconst.TestUUID + ".ZIP", isDir: false}
	artifacts := []string{testconst.TestUUID}

	b.Run("", func(b *testing.B) {
		for range b.N {
			_ = fixlha.Check(sl, extra, d, artifacts...)
		}
	})
}

// BenchmarkCheckManyArtifacts benchmarks the Check function with many artifacts.
func BenchmarkCheckManyArtifacts(b *testing.B) {
	sl := slog.New(slog.DiscardHandler)
	extra := dir.Directory(b.TempDir())
	d := &MockDirEntry{name: testconst.TestUUID + ".zip", isDir: false}

	// Generate many sorted artifacts (binary search requires sorted)
	artifacts := make([]string, 0, 100)
	for i := range 100 {
		artifacts = append(artifacts, "00000000-0000-0000-0000-000000000000"+string(rune('a'+i%26)))
	}
	artifacts = append(artifacts, testconst.TestUUID)

	b.Run("", func(b *testing.B) {
		for range b.N {
			_ = fixlha.Check(sl, extra, d, artifacts...)
		}
	})
}
