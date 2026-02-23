package fixzip_test

import (
	"context"
	"errors"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/Defacto2/server/internal/config/fixzip"
	"github.com/Defacto2/server/internal/dir"
	"github.com/nalgeon/be"
)

var errMockInfoNotSupported = errors.New("mock: Info() not supported")

// MockDirEntry implements fs.DirEntry for testing.
type MockDirEntry struct {
	name  string
	isDir bool
}

func (m *MockDirEntry) Name() string               { return m.name }
func (m *MockDirEntry) IsDir() bool                { return m.isDir }
func (m *MockDirEntry) Type() fs.FileMode          { return 0 }
func (m *MockDirEntry) Info() (fs.FileInfo, error) { return nil, errMockInfoNotSupported }

// TestCheckIsDirectory tests that directories are skipped.
func TestCheckIsDirectory(t *testing.T) {
	sl := slog.New(slog.DiscardHandler)
	tmpDir := t.TempDir()
	extra := dir.Directory(tmpDir)

	d := &MockDirEntry{name: "somedir", isDir: true}
	result := fixzip.Check(sl, "", extra, d)
	be.Equal(t, result, "")
}

// TestCheckWrongExtension tests that non-.zip files are skipped.
func TestCheckWrongExtension(t *testing.T) {
	sl := slog.New(slog.DiscardHandler)
	tmpDir := t.TempDir()
	extra := dir.Directory(tmpDir)

	d := &MockDirEntry{name: "file123.rar", isDir: false}
	result := fixzip.Check(sl, "", extra, d)
	be.Equal(t, result, "")
}

// TestCheckNoExtension tests that files with no extension are skipped.
func TestCheckNoExtension(t *testing.T) {
	sl := slog.New(slog.DiscardHandler)
	tmpDir := t.TempDir()
	extra := dir.Directory(tmpDir)

	d := &MockDirEntry{name: "file123", isDir: false}
	result := fixzip.Check(sl, "", extra, d)
	be.Equal(t, result, "")
}

// TestCheckUUIDNotInArtifacts tests that UUIDs not in artifacts list are skipped.
func TestCheckUUIDNotInArtifacts(t *testing.T) {
	sl := slog.New(slog.DiscardHandler)
	tmpDir := t.TempDir()
	extra := dir.Directory(tmpDir)

	d := &MockDirEntry{name: "12345678-1234-1234-1234-123456789012.zip", isDir: false}
	artifacts := []string{"aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"}
	result := fixzip.Check(sl, "", extra, d, artifacts...)
	be.Equal(t, result, "")
}

// TestCheckAlreadyInExtra tests that files already in extra directory are skipped.
func TestCheckAlreadyInExtra(t *testing.T) {
	sl := slog.New(slog.DiscardHandler)
	tmpDir := t.TempDir()
	extra := dir.Directory(tmpDir)
	uid := "12345678-1234-1234-1234-123456789012"

	// Create the extra zip file
	extraZip := filepath.Join(tmpDir, uid+".zip")
	err := os.WriteFile(extraZip, []byte("test"), 0o644)
	be.True(t, err == nil)

	d := &MockDirEntry{name: uid + ".zip", isDir: false}
	artifacts := []string{uid}
	result := fixzip.Check(sl, "", extra, d, artifacts...)
	be.Equal(t, result, "")
}

// TestCheckNilLogger tests that Check panics with nil logger.
func TestCheckNilLogger(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic with nil logger")
		}
	}()

	tmpDir := t.TempDir()
	extra := dir.Directory(tmpDir)
	d := &MockDirEntry{name: "test.zip", isDir: false}
	fixzip.Check(nil, "", extra, d)
}

// TestCheckUppercaseExtension tests that uppercase extensions are handled correctly.
func TestCheckUppercaseExtension(t *testing.T) {
	sl := slog.New(slog.DiscardHandler)
	tmpDir := t.TempDir()
	extra := dir.Directory(tmpDir)

	d := &MockDirEntry{name: "test.ZIP", isDir: false}
	result := fixzip.Check(sl, "", extra, d)
	be.Equal(t, result, "")
}

// TestCheckMixedCaseExtension tests that mixed case extensions are handled correctly.
func TestCheckMixedCaseExtension(t *testing.T) {
	sl := slog.New(slog.DiscardHandler)
	tmpDir := t.TempDir()
	extra := dir.Directory(tmpDir)
	uid := "12345678-1234-1234-1234-123456789012"

	d := &MockDirEntry{name: uid + ".Zip", isDir: false}
	artifacts := []string{uid}
	result := fixzip.Check(sl, "", extra, d, artifacts...)
	be.Equal(t, result, "")
}

// TestCheckEmptyArtifacts tests that empty artifacts list works correctly.
func TestCheckEmptyArtifacts(t *testing.T) {
	sl := slog.New(slog.DiscardHandler)
	tmpDir := t.TempDir()
	extra := dir.Directory(tmpDir)
	uid := "12345678-1234-1234-1234-123456789012"

	d := &MockDirEntry{name: uid + ".zip", isDir: false}
	result := fixzip.Check(sl, "", extra, d)
	be.Equal(t, result, "")
}

// TestInvalidNilLogger tests that Invalid panics with nil logger.
func TestInvalidNilLogger(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic with nil logger")
		}
	}()

	ctx := context.Background()
	fixzip.Invalid(ctx, nil, "/tmp/test.zip")
}

// TestInvalidContextCancellation tests that Invalid respects context cancellation.
func TestInvalidContextCancellation(t *testing.T) {
	sl := slog.New(slog.DiscardHandler)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Should return true (invalid) due to context being cancelled
	result := fixzip.Invalid(ctx, sl, "/tmp/nonexistent.zip")
	be.True(t, result)
}

// TestInvalidNonexistentFile tests that Invalid returns true for nonexistent files.
func TestInvalidNonexistentFile(t *testing.T) {
	sl := slog.New(slog.DiscardHandler)
	ctx := context.Background()

	result := fixzip.Invalid(ctx, sl, "/tmp/nonexistent_zip_file_"+t.Name()+".zip")
	be.True(t, result)
}
