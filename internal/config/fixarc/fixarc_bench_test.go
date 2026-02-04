package fixarc

import (
	"io"
	"log/slog"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Defacto2/server/internal/dir"
)

// BenchmarkExtensionExtractionOld simulates the old approach (Ext called twice + full ToLower)
func BenchmarkExtensionExtractionOld(b *testing.B) {
	names := []string{
		"file123.ZIP",
		"archive.Zip",
		"data.zIp",
		"document.pdf",
		"archive.tar.gz",
		"unknown",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
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
}

// BenchmarkExtensionExtractionNew simulates the new optimized approach
func BenchmarkExtensionExtractionNew(b *testing.B) {
	names := []string{
		"file123.ZIP",
		"archive.Zip",
		"data.zIp",
		"document.pdf",
		"archive.tar.gz",
		"unknown",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, name := range names {
			// New approach: filepath.Ext called once, extension cached
			ext := filepath.Ext(name)
			if strings.ToLower(ext) != ".zip" && ext != "" {
				continue
			}
			_ = strings.TrimSuffix(name, ext)
		}
	}
}

// BenchmarkCheckFunctionOptimized measures the optimized Check function
func BenchmarkCheckFunctionOptimized(b *testing.B) {
	sl := slog.New(slog.NewTextHandler(io.Discard, nil))
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

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, entry := range entries {
			d := &MockDirEntry{name: entry.name, isDir: entry.isDir}
			_ = Check(sl, "", extra, d)
		}
	}
}
