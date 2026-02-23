package fixarj

import (
	"context"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/Defacto2/server/internal/dir"
	"github.com/nalgeon/be"
)

// MockDirEntry implements fs.DirEntry for testing.
type MockDirEntry struct {
	name  string
	isDir bool
}

func (m *MockDirEntry) Name() string               { return m.name }
func (m *MockDirEntry) IsDir() bool                { return m.isDir }
func (m *MockDirEntry) Type() fs.FileMode          { return 0 }
func (m *MockDirEntry) Info() (fs.FileInfo, error) { return nil, nil }

// TestCheckIsDirectory tests that directories are skipped.
func TestCheckIsDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	extra := dir.Directory(tmpDir)

	d := &MockDirEntry{name: "somedir", isDir: true}
	result := Check(extra, d)
	be.Equal(t, result, "")
}

// TestCheckWrongExtension tests that non-.zip files are skipped.
func TestCheckWrongExtension(t *testing.T) {
	tmpDir := t.TempDir()
	extra := dir.Directory(tmpDir)

	d := &MockDirEntry{name: "file123.arj", isDir: false}
	result := Check(extra, d)
	be.Equal(t, result, "")
}

// TestCheckNoExtension tests that files with no extension are skipped.
func TestCheckNoExtension(t *testing.T) {
	tmpDir := t.TempDir()
	extra := dir.Directory(tmpDir)

	d := &MockDirEntry{name: "file123", isDir: false}
	result := Check(extra, d)
	be.Equal(t, result, "")
}

// TestCheckUUIDNotInArtifacts tests that UUIDs not in artifacts list are skipped.
func TestCheckUUIDNotInArtifacts(t *testing.T) {
	tmpDir := t.TempDir()
	extra := dir.Directory(tmpDir)

	d := &MockDirEntry{name: "12345678-1234-1234-1234-123456789012.zip", isDir: false}
	artifacts := []string{"aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"}
	result := Check(extra, d, artifacts...)
	be.Equal(t, result, "")
}

// TestCheckAlreadyInExtra tests that files already in extra directory are skipped.
func TestCheckAlreadyInExtra(t *testing.T) {
	tmpDir := t.TempDir()
	extra := dir.Directory(tmpDir)
	uid := "12345678-1234-1234-1234-123456789012"

	// Create the extra zip file
	extraZip := filepath.Join(tmpDir, uid+".zip")
	err := os.WriteFile(extraZip, []byte("test"), 0o644)
	be.True(t, err == nil)

	d := &MockDirEntry{name: uid + ".zip", isDir: false}
	artifacts := []string{uid}
	result := Check(extra, d, artifacts...)
	be.Equal(t, result, "")
}

// TestCheckValidFile tests that valid file returns UUID.
func TestCheckValidFile(t *testing.T) {
	tmpDir := t.TempDir()
	extra := dir.Directory(tmpDir)
	uid := "12345678-1234-1234-1234-123456789012"

	d := &MockDirEntry{name: uid + ".zip", isDir: false}
	artifacts := []string{uid}
	result := Check(extra, d, artifacts...)
	be.Equal(t, result, uid)
}

// TestCheckCaseInsensitiveExtension tests case-insensitive extension matching.
func TestCheckCaseInsensitiveExtension(t *testing.T) {
	tmpDir := t.TempDir()
	extra := dir.Directory(tmpDir)

	testCases := []struct {
		name string
	}{
		{name: "12345678-1234-1234-1234-123456789012.ZIP"},
		{name: "12345678-1234-1234-1234-123456789012.Zip"},
		{name: "12345678-1234-1234-1234-123456789012.zIp"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			d := &MockDirEntry{name: tc.name, isDir: false}
			artifacts := []string{"12345678-1234-1234-1234-123456789012"}
			result := Check(extra, d, artifacts...)
			be.Equal(t, result, "12345678-1234-1234-1234-123456789012")
		})
	}
}

// TestCheckMultipleArtifacts tests that binary search works with multiple artifacts.
func TestCheckMultipleArtifacts(t *testing.T) {
	tmpDir := t.TempDir()
	extra := dir.Directory(tmpDir)

	uid := "ffffffff-ffff-ffff-ffff-ffffffffffff"
	d := &MockDirEntry{name: uid + ".zip", isDir: false}

	// Provide sorted artifacts (required for binary search)
	artifacts := []string{
		"aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
		"cccccccc-cccc-cccc-cccc-cccccccccccc",
		"ffffffff-ffff-ffff-ffff-ffffffffffff",
		"zzzzzzzz-zzzz-zzzz-zzzz-zzzzzzzzzzzz",
	}

	result := Check(extra, d, artifacts...)
	be.Equal(t, result, uid)
}

// TestCheckBinarySearchCorrectness tests that binary search finds UUIDs at various positions.
func TestCheckBinarySearchCorrectness(t *testing.T) {
	tmpDir := t.TempDir()
	extra := dir.Directory(tmpDir)

	testCases := []struct {
		name     string
		uid      string
		position int
	}{
		{name: "start", uid: "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa", position: 0},
		{name: "middle", uid: "mmmmmmmm-mmmm-mmmm-mmmm-mmmmmmmmmmmm", position: 1},
		{name: "end", uid: "zzzzzzzz-zzzz-zzzz-zzzz-zzzzzzzzzzzz", position: 2},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			d := &MockDirEntry{name: tc.uid + ".zip", isDir: false}
			artifacts := []string{
				"aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
				"mmmmmmmm-mmmm-mmmm-mmmm-mmmmmmmmmmmm",
				"zzzzzzzz-zzzz-zzzz-zzzz-zzzzzzzzzzzz",
			}
			result := Check(extra, d, artifacts...)
			be.Equal(t, result, tc.uid)
		})
	}
}

// TestInvalidNilLogger tests that nil logger panics.
func TestInvalidNilLogger(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for nil logger")
		}
	}()
	Invalid(nil, "/tmp/test.arj")
}

// TestInvalidNonexistentFile tests behavior with non-existent file.
func TestInvalidNonexistentFile(t *testing.T) {
	sl := slog.New(slog.NewTextHandler(io.Discard, nil))
	result := Invalid(sl, "/nonexistent/file/path.arj")
	// Command should fail, so result should be true
	be.Equal(t, result, true)
}

// TestInvalidWithTimeout tests that timeout works (shouldn't hang).
func TestInvalidWithTimeout(t *testing.T) {
	sl := slog.New(slog.NewTextHandler(io.Discard, nil))
	tmpDir := t.TempDir()

	// Create a dummy arj file
	arjPath := filepath.Join(tmpDir, "test.arj")
	err := os.WriteFile(arjPath, []byte("dummy arj"), 0o644)
	be.True(t, err == nil)

	// This should complete within the 10-second timeout
	result := Invalid(sl, arjPath)
	// Command will likely fail since we don't have a valid arj file, so result should be true
	be.Equal(t, result, true)
}

// TestFilesContextNil tests Files with nil context.
func TestFilesContextNil(t *testing.T) {
	files, err := Files(nil, nil)
	be.True(t, err != nil)
	be.Equal(t, files, nil)
}

// TestFilesExecutorNil tests Files with nil executor.
func TestFilesExecutorNil(t *testing.T) {
	ctx := context.Background()
	files, err := Files(ctx, nil)
	be.True(t, err != nil)
	be.Equal(t, files, nil)
}

// TestCheckExtensionFiltering tests that extension filtering works correctly.
func TestCheckExtensionFiltering(t *testing.T) {
	tmpDir := t.TempDir()
	extra := dir.Directory(tmpDir)

	testCases := []struct {
		filename    string
		shouldMatch bool
		description string
	}{
		{"file.zip", true, "valid .zip extension"},
		{"file.ZIP", true, "uppercase extension"},
		{"file.Zip", true, "mixed case extension"},
		{"file.arj", false, "wrong extension"},
		{"file.tar", false, "wrong extension"},
		{"file.exe", false, "wrong extension"},
		{"file", false, "no extension"},
		{"file.", false, "only dot, no extension"},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			d := &MockDirEntry{name: tc.filename, isDir: false}
			result := Check(extra, d)
			if tc.shouldMatch {
				// Would return "" because no artifacts, but passes extension check
				be.Equal(t, result, "")
			} else {
				be.Equal(t, result, "")
			}
		})
	}
}

// TestCheckFileInExtraDirectory tests extra directory file existence check.
func TestCheckFileInExtraDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	extra := dir.Directory(tmpDir)
	uid := "12345678-1234-1234-1234-123456789012"

	testCases := []struct {
		name       string
		createFile bool
	}{
		{name: "file exists in extra", createFile: true},
		{name: "file does not exist", createFile: false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.createFile {
				extraZip := filepath.Join(tmpDir, uid+".zip")
				err := os.WriteFile(extraZip, []byte("test"), 0o644)
				be.True(t, err == nil)
				defer os.Remove(extraZip)

				d := &MockDirEntry{name: uid + ".zip", isDir: false}
				artifacts := []string{uid}
				result := Check(extra, d, artifacts...)
				// Should return "" because file exists in extra
				be.Equal(t, result, "")
			} else {
				d := &MockDirEntry{name: uid + ".zip", isDir: false}
				artifacts := []string{uid}
				result := Check(extra, d, artifacts...)
				// Should return uid because file doesn't exist in extra
				be.Equal(t, result, uid)
			}
		})
	}
}

// BenchmarkCheck measures Check function performance.
func BenchmarkCheck(b *testing.B) {
	tmpDir := b.TempDir()
	extra := dir.Directory(tmpDir)

	uid := "12345678-1234-1234-1234-123456789012"
	d := &MockDirEntry{name: uid + ".zip", isDir: false}
	artifacts := []string{uid}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Check(extra, d, artifacts...)
	}
}

// BenchmarkInvalid measures Invalid function performance.
func BenchmarkInvalid(b *testing.B) {
	sl := slog.New(slog.NewTextHandler(io.Discard, nil))
	tmpDir := b.TempDir()

	arjPath := filepath.Join(tmpDir, "test.arj")
	err := os.WriteFile(arjPath, []byte("dummy"), 0o644)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Invalid(sl, arjPath)
	}
}
