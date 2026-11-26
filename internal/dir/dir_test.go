package dir_test

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
	"time"

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

	t.Run("valid", func(t *testing.T) {
		t.Parallel()
		d := dir.Directory(tmpDir)
		err := d.IsDir()
		be.True(t, err == nil)
	})

	t.Run("empty", func(t *testing.T) {
		t.Parallel()
		d := dir.Directory("")
		err := d.IsDir()
		be.True(t, errors.Is(err, dir.ErrNoPath))
	})

	t.Run("missing", func(t *testing.T) {
		t.Parallel()
		d := dir.Directory("/nonexistent/path/that/does/not/exist")
		err := d.IsDir()
		be.True(t, errors.Is(err, dir.ErrNoDir))
	})

	t.Run("file", func(t *testing.T) {
		t.Parallel()
		filePath := filepath.Join(t.TempDir(), "test.txt")
		be.Err(t, os.WriteFile(filePath, nil, 0o600), nil)

		d := dir.Directory(filePath)
		err := d.IsDir()
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

	t.Run("writable", func(t *testing.T) {
		t.Parallel()
		d := dir.Directory(tmpDir)
		err := d.Check(sl)
		be.True(t, err == nil)
	})

	t.Run("missing", func(t *testing.T) {
		t.Parallel()
		d := dir.Directory("/nonexistent/path/for/check")
		err := d.Check(sl)
		be.Err(t, err)
	})

	t.Run("file", func(t *testing.T) {
		t.Parallel()
		filePath := filepath.Join(t.TempDir(), "test.txt")
		be.Err(t, os.WriteFile(filePath, nil, 0o600), nil)

		d := dir.Directory(filePath)
		err := d.Check(sl)
		be.Err(t, err)
	})

	t.Run("nil logger", func(t *testing.T) {
		t.Parallel()
		d := dir.Directory(tmpDir)
		err := d.Check(nil)
		be.Err(t, err)
	})

	t.Run("read-only", func(t *testing.T) {
		t.Parallel()
		if os.Geteuid() == 0 {
			t.Skip("root bypasses directory permissions")
		}
		roDir := filepath.Join(t.TempDir(), "ro")
		be.Err(t, os.Mkdir(roDir, 0o500), nil)
		t.Cleanup(func() {
			_ = os.Chmod(roDir, 0o700)
		})

		err := dir.Directory(roDir).Check(sl)
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

func TestIsTemp(t *testing.T) {
	t.Parallel()

	be.True(t, !dir.IsTemp(nil))

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

	// Calling again with the same name reuses the existing directory.
	reused, err := dir.MkdirStale(include)
	be.Err(t, err, nil)
	be.Equal(t, reused, prefix)

	st, err = os.Stat(prefix)
	be.Err(t, err, nil)
	be.True(t, st.IsDir())

	// If a regular file occupies the target name, fall back to a randomized subdir.
	const collide = "stale_file_collision"
	blocker := filepath.Join(os.TempDir(), dir.Pattern+"-"+collide)
	be.Err(t, os.WriteFile(blocker, nil, 0o600), nil)
	fallback, err := dir.MkdirStale(collide)
	be.Err(t, err, nil)
	be.True(t, fallback != blocker)
	st, err = os.Stat(fallback)
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
		err = dir.RemoveAll(fallback)
		be.Err(t, err, nil)
		_ = os.Remove(blocker)
		err = dir.RemoveAll(emoji)
		be.Err(t, err, nil)
	})
}

func TestRemoveAllOutsideTemp(t *testing.T) {
	t.Parallel()

	be.Err(t, dir.RemoveAll("/etc/hosts"))
	be.Err(t, dir.RemoveAll(filepath.Join(os.TempDir(), "..")))
}

func TestPurgeTemp(t *testing.T) {
	t.Setenv("TMPDIR", t.TempDir())

	tmp, err := dir.MkdirTemp("purgetest")
	be.Err(t, err, nil)
	be.Err(t, os.WriteFile(filepath.Join(tmp, "file.bin"), make([]byte, 512), 0o600), nil)

	// Fresh dir is skipped when stale threshold is large.
	rep, err := dir.PurgeTemp(time.Hour, false)
	be.Err(t, err, nil)
	be.Equal(t, rep.Found, 1)
	be.Equal(t, rep.FoundBytes, int64(512))
	be.Equal(t, rep.Skipped, 1)
	be.Equal(t, rep.Removed, 0)

	// Dry run reports removal without deleting.
	rep, err = dir.PurgeTemp(0, true)
	be.Err(t, err, nil)
	be.Equal(t, rep.Found, 1)
	be.Equal(t, rep.Removed, 1)
	_, statErr := os.Stat(tmp)
	be.Err(t, statErr, nil)

	// Real purge deletes the directory.
	rep, err = dir.PurgeTemp(0, false)
	be.Err(t, err, nil)
	be.Equal(t, rep.Found, 1)
	be.Equal(t, rep.Removed, 1)
	be.Equal(t, rep.RemovedBytes, int64(512))
	_, statErr = os.Stat(tmp)
	be.True(t, os.IsNotExist(statErr))

	// CleanTemp removes matching directories via PurgeTemp.
	tmp2, err := dir.MkdirTemp("cleantest")
	be.Err(t, err, nil)
	be.Err(t, dir.CleanTemp(0), nil)
	_, statErr = os.Stat(tmp2)
	be.True(t, os.IsNotExist(statErr))
}
