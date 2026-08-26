package fixarc_test

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Defacto2/server/internal/config/fixarc"
	"github.com/Defacto2/server/internal/config/testconst"
	"github.com/Defacto2/server/internal/dir"
)

// BenchmarkCheck measures Check function performance.
func BenchmarkCheck(b *testing.B) {
	sl := slog.New(slog.DiscardHandler)
	tmpDir := b.TempDir()
	extra := dir.Directory(tmpDir)

	uid := testconst.TestUUID
	d := &MockDirEntry{name: uid + ".zip", isDir: false}
	artifacts := []string{uid}

	zipPath := filepath.Join(tmpDir, uid+".zip")
	err := os.WriteFile(zipPath, []byte("not zip"), 0o600)
	if err != nil {
		b.Fatal(err)
	}

	b.Run("", func(b *testing.B) {
		for range b.N {
			_ = fixarc.Check(sl, zipPath, extra, d, artifacts...)
		}
	})
}

// BenchmarkInvalid measures Invalid function performance.
func BenchmarkInvalid(b *testing.B) {
	sl := slog.New(slog.DiscardHandler)
	tmpDir := b.TempDir()

	arcPath := filepath.Join(tmpDir, "test.arc")
	err := os.WriteFile(arcPath, []byte("dummy"), 0o600)
	if err != nil {
		b.Fatal(err)
	}

	b.Run("", func(b *testing.B) {
		for range b.N {
			_ = fixarc.Invalid(context.Background(), sl, arcPath)
		}
	})
}

// BenchmarkExtensionExtractionOld simulates the old approach (Ext called twice + full ToLower).
func BenchmarkExtensionExtractionOld(b *testing.B) {
	names := [...]string{
		"file123.ZIP",
		"archive.Zip",
		"data.zIp",
		"document.pdf",
		"archive.tar.gz",
		"unknown",
	}

	b.Run("", func(b *testing.B) {
		for range b.N {
			for _, name := range names {
				// Old approach: filepath.Ext(strings.ToLower(name)) called on line 38
				ext1 := filepath.Ext(strings.ToLower(name))
				if ext1 != ".zip" && ext1 != "" {
					continue
				}
				// Old approach: filepath.Ext(name) called again on line 41
				ext2 := filepath.Ext(name)
				_ = strings.TrimSuffix(name, ext2)
			}
		}
	})
}

// BenchmarkExtensionExtractionNew simulates the new optimized approach.
func BenchmarkExtensionExtractionNew(b *testing.B) {
	names := [...]string{
		"file123.ZIP",
		"archive.Zip",
		"data.zIp",
		"document.pdf",
		"archive.tar.gz",
		"unknown",
	}

	b.Run("", func(b *testing.B) {
		for range b.N {
			for _, name := range names {
				// New approach: filepath.Ext called once, extension cached
				ext := filepath.Ext(name)
				if strings.ToLower(ext) != ".zip" && ext != "" {
					continue
				}
				_ = strings.TrimSuffix(name, ext)
			}
		}
	})
}

// BenchmarkCheckFunctionOptimized measures the optimized Check function.
func BenchmarkCheckFunctionOptimized(b *testing.B) {
	sl := slog.New(slog.DiscardHandler)
	tmpDir := b.TempDir()
	extra := dir.Directory(tmpDir)

	// Create various test files
	entries := []struct {
		name  string
		isDir bool
	}{
		{"file.ZIP", false},
		{"archive.Zip", false},
		{"data.zIp", false},
		{"document.pdf", false},
		{"unknown", false},
	}

	b.Run("", func(b *testing.B) {
		for range b.N {
			for _, entry := range entries {
				d := &MockDirEntry{name: entry.name, isDir: entry.isDir}
				_ = fixarc.Check(sl, "", extra, d)
			}
		}
	})
}
