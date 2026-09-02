package dir_test

import (
	"errors"
	"os"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/Defacto2/server/internal/dir"
	"github.com/Defacto2/server/internal/logs"
	"github.com/nalgeon/be"
)

// checked in Sep 26, test coverage was good at 65%

func TestJoin(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		dirPath  string
		fileName string
		want     string
	}{
		{
			name:     "simple file",
			dirPath:  "/tmp",
			fileName: "test.zip",
			want:     "/tmp/test.zip",
		},
		{
			name:     "nested path",
			dirPath:  "/var/lib",
			fileName: "subdir/file.txt",
			want:     "/var/lib/subdir/file.txt",
		},
		{
			name:     "empty filename",
			dirPath:  "/tmp",
			fileName: "",
			want:     "/tmp",
		},
		{
			name:     "relative directory",
			dirPath:  ".",
			fileName: "file.txt",
			want:     "file.txt",
		},
		{
			name:     "path with trailing slash",
			dirPath:  "/tmp/",
			fileName: "file.txt",
			want:     "/tmp/file.txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			d := dir.Directory(tt.dirPath)
			got := d.Join(tt.fileName)
			be.Equal(t, got, tt.want)
		})
	}
}

func TestPath(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		path string
	}{
		{"simple path", "/tmp"},
		{"nested path", "/var/lib/defacto2"},
		{"empty path", ""},
		{"relative path", "."},
		{"home path", "~"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			d := dir.Directory(tt.path)
			got := d.Path()
			be.Equal(t, got, tt.path)
		})
	}
}

func TestIsDir(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	t.Run("valid directory", func(t *testing.T) {
		t.Parallel()
		d := dir.Directory(tmpDir)
		err := d.IsDir()
		be.True(t, err == nil)
	})

	t.Run("empty path returns ErrNoPath", func(t *testing.T) {
		t.Parallel()
		d := dir.Directory("")
		err := d.IsDir()
		be.True(t, errors.Is(err, dir.ErrNoPath))
	})

	t.Run("non-existent directory returns ErrNoDir", func(t *testing.T) {
		t.Parallel()
		d := dir.Directory("/nonexistent/path/that/does/not/exist")
		err := d.IsDir()
		be.True(t, errors.Is(err, dir.ErrNoDir))
	})

	t.Run("file instead of directory returns ErrFile", func(t *testing.T) {
		t.Parallel()
		// Create a temp file
		tmpFile, err := dir.CreateTemp("test-*.txt")
		be.True(t, err == nil)

		defer func() {
			err := tmpFile.Close()
			t.Log(err)
		}()
		defer func() {
			err := os.Remove(tmpFile.Name())
			t.Log(err)
		}()
		d := dir.Directory(tmpFile.Name())
		err = d.IsDir()
		be.True(t, errors.Is(err, dir.ErrFile))
	})
}

func TestPaths(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		dirs []dir.Directory
		want []string
	}{
		{
			name: "single directory",
			dirs: []dir.Directory{"/tmp"},
			want: []string{"/tmp"},
		},
		{
			name: "multiple directories",
			dirs: []dir.Directory{"/tmp", "/var", "/home"},
			want: []string{"/tmp", "/var", "/home"},
		},
		{
			name: "empty slice",
			dirs: []dir.Directory{},
			want: []string{},
		},
		{
			name: "empty path",
			dirs: []dir.Directory{""},
			want: []string{""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := dir.Paths(tt.dirs...)
			be.Equal(t, len(got), len(tt.want))
			for i, path := range got {
				be.Equal(t, path, tt.want[i])
			}
		})
	}
}

func TestCheck(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	sl := logs.Discard()

	t.Run("valid writable directory", func(t *testing.T) {
		t.Parallel()
		d := dir.Directory(tmpDir)
		err := d.Check(sl)
		be.True(t, err == nil)
	})

	t.Run("non-existent directory returns error", func(t *testing.T) {
		t.Parallel()
		d := dir.Directory("/nonexistent/path/for/check")
		err := d.Check(sl)
		be.Err(t, err)
	})

	t.Run("file instead of directory returns error", func(t *testing.T) {
		t.Parallel()
		// Create a temp file
		tmpFile, err := dir.CreateTemp("test-*.txt")
		be.True(t, err == nil)
		defer func() {
			if err := tmpFile.Close(); err != nil {
				t.Log(err)
			}
		}()
		defer func() {
			if err := os.Remove(tmpFile.Name()); err != nil {
				t.Log(err)
			}
		}()
		d := dir.Directory(tmpFile.Name())
		err = d.Check(sl)
		be.Err(t, err)
	})

	t.Run("nil logger returns error", func(t *testing.T) {
		t.Parallel()
		d := dir.Directory(tmpDir)
		err := d.Check(nil)
		be.Err(t, err)
	})
}

func TestCheckTempFileCreation(t *testing.T) {
	// Test that temp file creation works and cleanup happens
	t.Parallel()
	tmpDir := t.TempDir()
	sl := logs.Discard()

	// Count files before check
	filesBefore, err := os.ReadDir(tmpDir)
	be.True(t, err == nil)

	// Run check
	d := dir.Directory(tmpDir)
	err = d.Check(sl)
	be.True(t, err == nil)

	// Count files after check - should be same (temp file cleaned up)
	filesAfter, err := os.ReadDir(tmpDir)
	be.True(t, err == nil)

	be.Equal(t, len(filesBefore), len(filesAfter))
}

func TestJoinNormalization(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		dirPath  string
		fileName string
		check    func(string) bool
	}{
		{
			name:     "removes double slashes",
			dirPath:  "/tmp/",
			fileName: "file.txt",
			check: func(result string) bool {
				// Should not have double slashes
				for i := range len(result) - 1 {
					if result[i] == '/' && result[i+1] == '/' {
						return false
					}
				}
				return true
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			d := dir.Directory(tt.dirPath)
			result := d.Join(tt.fileName)
			be.True(t, tt.check(result))
		})
	}
}

func TestErrorTypes(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		err  error
	}{
		{"ErrFile", dir.ErrFile},
		{"ErrSave", dir.ErrSave},
		{"ErrNoPath", dir.ErrNoPath},
		{"ErrNoDir", dir.ErrNoDir},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			be.True(t, tt.err != nil)
			be.True(t, len(tt.err.Error()) > 0)
		})
	}
}

func TestIsTemp(t *testing.T) {
	t.Parallel()

	const wantMatch = dir.Pattern + "-random"
	// use a virtual file system to test os.DirEntry
	fsys := fstest.MapFS{
		"temp.tmp":    &fstest.MapFile{Mode: 0o644},
		"regular.txt": &fstest.MapFile{Mode: 0o644},
		"temp_folder": &fstest.MapFile{Mode: os.ModeDir | 0o755},
		dir.Pattern:   &fstest.MapFile{Mode: os.ModeDir | 0o755},
		wantMatch:     &fstest.MapFile{Mode: os.ModeDir | 0o755},
	}

	entries, err := fsys.ReadDir(".")
	if err != nil {
		t.Fatalf("failed to read mapfs: %v", err)
	}

	for n, entry := range entries {
		t.Run(entry.Name(), func(t *testing.T) {
			t.Parallel()
			got := dir.IsTemp(entry)
			t.Log(n, entry.Name(), got)

			if entry.Name() == wantMatch {
				be.True(t, got)
			} else {
				be.True(t, !got)
			}
		})
	}
}

func TestMkdirStale(t *testing.T) {
	t.Parallel()

	random, err := dir.MkdirStale("")
	be.Err(t, err, nil)

	st, err := os.Stat(random)
	be.Err(t, err, nil)
	be.True(t, st.IsDir())

	const include = "hello_world"
	prefix, err := dir.MkdirStale(include)
	be.Err(t, err, nil)
	be.True(t, strings.Contains(prefix, include))

	st, err = os.Stat(prefix)
	be.Err(t, err, nil)
	be.True(t, st.IsDir())

	const smiley = "hello🙂world"
	emoji, err := dir.MkdirStale(smiley)
	be.Err(t, err, nil)
	be.True(t, strings.Contains(emoji, smiley))

	st, err = os.Stat(emoji)
	be.Err(t, err, nil)
	be.True(t, st.IsDir())

	t.Cleanup(func() {
		err := dir.RemoveAll(random)
		be.Err(t, err, nil)
		err = dir.RemoveAll(prefix)
		be.Err(t, err, nil)
		err = dir.RemoveAll(emoji)
		be.Err(t, err, nil)
	})
}
