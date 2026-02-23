package fixarj_test

import (
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/Defacto2/server/internal/config/fixarj"
	"github.com/Defacto2/server/internal/dir"
)

// BenchmarkCheckOptimizations measures optimization improvements
func BenchmarkCheckOptimizations(b *testing.B) {
	tmpDir := b.TempDir()
	extra := dir.Directory(tmpDir)

	uid := "12345678-1234-1234-1234-123456789012"
	d := &MockDirEntry{name: uid + ".zip", isDir: false}
	artifacts := []string{uid}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = fixarj.Check(extra, d, artifacts...)
	}
}

// BenchmarkInvalidBytesOptimization measures string vs bytes.Contains optimization
func BenchmarkInvalidBytesOptimization(b *testing.B) {
	sl := slog.New(slog.DiscardHandler)
	tmpDir := b.TempDir()

	arjPath := filepath.Join(tmpDir, "test.arj")
	err := os.WriteFile(arjPath, []byte("dummy arj content"), 0o644)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = fixarj.Invalid(sl, arjPath)
	}
}

// BenchmarkCheckWithExtraZip tests the optimization where extra zip exists
func BenchmarkCheckWithExtraZip(b *testing.B) {
	tmpDir := b.TempDir()
	extra := dir.Directory(tmpDir)
	uid := "12345678-1234-1234-1234-123456789012"

	// Create the extra zip file (optimization path)
	extraZip := filepath.Join(tmpDir, uid+".zip")
	err := os.WriteFile(extraZip, []byte("zip"), 0o644)
	if err != nil {
		b.Fatal(err)
	}

	d := &MockDirEntry{name: uid + ".zip", isDir: false}
	artifacts := []string{uid}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = fixarj.Check(extra, d, artifacts...)
	}
}

// BenchmarkCheckWithoutExtraZip tests the non-optimization path
func BenchmarkCheckWithoutExtraZip(b *testing.B) {
	tmpDir := b.TempDir()
	extra := dir.Directory(tmpDir)
	uid := "12345678-1234-1234-1234-123456789012"

	d := &MockDirEntry{name: uid + ".zip", isDir: false}
	artifacts := []string{uid}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = fixarj.Check(extra, d, artifacts...)
	}
}
