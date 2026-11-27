package config_test

import (
	"bytes"
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Defacto2/server/internal/config"
	"github.com/Defacto2/server/internal/dir"
	"github.com/Defacto2/server/internal/logs"
	"github.com/Defacto2/server/internal/nils"
	"github.com/nalgeon/be"
)

var testLog = logs.Discard()

func TestRemove(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		data string
		want string
		keep string
	}{
		{"diz", "FILE_ID.DIZ content", "readme.txt", "file_id.diz"},
		{"too wide", strings.Repeat("1234567890", 100), "file_id.diz", "readme.txt"},
		{"too long", strings.Repeat("1234567890\n", 15), "file_id.diz", "readme.txt"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tmp := t.TempDir()
			diz := filepath.Join(tmp, "file_id.diz")
			txt := filepath.Join(tmp, "readme.txt")
			be.Err(t, os.WriteFile(diz, []byte(tt.data), 0o600), nil)
			be.Err(t, os.WriteFile(txt, []byte(tt.data), 0o600), nil)

			got, err := config.Remove(diz, txt)
			be.Err(t, err, nil)
			be.Equal(t, got, tt.want)

			_, err = os.Stat(filepath.Join(tmp, tt.want))
			be.True(t, errors.Is(err, os.ErrNotExist))
			_, err = os.Stat(filepath.Join(tmp, tt.keep))
			be.Err(t, err, nil)
		})
	}

	t.Run("missing diz", func(t *testing.T) {
		t.Parallel()

		tmp := t.TempDir()
		_, err := config.Remove(filepath.Join(tmp, "missing.diz"), filepath.Join(tmp, "readme.txt"))
		be.True(t, errors.Is(err, os.ErrNotExist))
	})

	t.Run("missing txt", func(t *testing.T) {
		t.Parallel()

		tmp := t.TempDir()
		diz := filepath.Join(tmp, "file_id.diz")
		be.Err(t, os.WriteFile(diz, []byte("short diz\n"), 0o600), nil)

		_, err := config.Remove(diz, filepath.Join(tmp, "missing.txt"))
		be.True(t, errors.Is(err, os.ErrNotExist))
	})
}

func TestFileID(t *testing.T) {
	t.Parallel()

	t.Run("nil", func(t *testing.T) {
		t.Parallel()
		be.Equal(t, config.FileID(nil), false)
	})

	tests := []struct {
		name string
		body string
		want bool
	}{
		{"empty", "", false},
		{"max valid", strings.Repeat(strings.Repeat("a", 45)+"\n", 10), true},
		{"too long", strings.Repeat("a\n", 11), false},
		{"too wide", strings.Repeat("a", 46), false},
		{"utf8 valid", strings.Repeat("é", 45), true},
		{"utf8 too wide", strings.Repeat("é", 46), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			be.Equal(t, config.FileID(strings.NewReader(tt.body)), tt.want)
		})
	}
}

func TestRepairString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		r    config.Repair
		want string
	}{
		{config.Zip, "zip"},
		{config.LHA, "lha"},
		{config.Arc, "arc"},
		{config.Arj, "arj"},
		{config.Repair(-1), "unknown(-1)"},
		{config.Repair(99), "unknown(99)"},
	}

	for _, tt := range tests {
		be.Equal(t, tt.r.String(), tt.want)
	}
}

func TestRemoveDir(t *testing.T) {
	t.Parallel()

	t.Run("invalid args", func(t *testing.T) {
		t.Parallel()

		be.True(t, errors.Is(config.RemoveDir(nil, "a", "/a", "/"), nils.ErrSlogLogger))
		be.True(t, errors.Is(config.RemoveDir(testLog, "", "/a", "/"), config.ErrNoPath))
	})

	t.Run("root", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer
		sl := slog.New(slog.NewTextHandler(&buf, nil))
		root := t.TempDir()

		be.Err(t, config.RemoveDir(sl, filepath.Base(root), root, root), nil)
		be.Equal(t, buf.String(), "")
	})

	t.Run("same name nested", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer
		sl := slog.New(slog.NewTextHandler(&buf, nil))
		root := filepath.Join(t.TempDir(), "downloads")
		nested := filepath.Join(root, "downloads")
		be.Err(t, os.MkdirAll(nested, 0o700), nil)

		be.Err(t, config.RemoveDir(sl, "downloads", nested, root), nil)
		be.True(t, strings.Contains(buf.String(), "unknown_path"))
	})

	t.Run("syncthing", func(t *testing.T) {
		t.Parallel()

		root := t.TempDir()
		st := filepath.Join(root, ".stfolder")
		be.Err(t, os.Mkdir(st, 0o700), nil)

		err := config.RemoveDir(testLog, ".stfolder", st, root)
		be.Err(t, err, filepath.SkipDir)
		_, statErr := os.Stat(st)
		be.True(t, errors.Is(statErr, os.ErrNotExist))
	})

	t.Run("unknown", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer
		sl := slog.New(slog.NewTextHandler(&buf, nil))
		root := t.TempDir()
		sub := filepath.Join(root, "unexpected")
		be.Err(t, os.Mkdir(sub, 0o700), nil)

		be.Err(t, config.RemoveDir(sl, "unexpected", sub, root), nil)
		be.True(t, strings.Contains(buf.String(), "unknown_path"))
		_, statErr := os.Stat(sub)
		be.Err(t, statErr, nil)
	})
}

func TestRemoveDownload(t *testing.T) {
	t.Parallel()

	t.Run("invalid args", func(t *testing.T) {
		t.Parallel()

		tmp := dir.Directory(t.TempDir())
		be.True(t, errors.Is(config.RemoveDownload(nil, "a", "/a", tmp, tmp), nils.ErrSlogLogger))
		be.True(t, errors.Is(config.RemoveDownload(testLog, "", "/a", tmp, tmp), config.ErrNoPath))
	})

	t.Run("no ext", func(t *testing.T) {
		t.Parallel()

		srcDir := t.TempDir()
		backup := dir.Directory(t.TempDir())
		extra := dir.Directory(t.TempDir())
		file := filepath.Join(srcDir, "11111111-2222-3333-4444-555566667777")
		be.Err(t, os.WriteFile(file, []byte("bin"), 0o600), nil)

		be.Err(t, config.RemoveDownload(testLog, filepath.Base(file), file, backup, extra), nil)
		_, err := os.Stat(file)
		be.Err(t, err, nil)
	})

	t.Run("valid ext", func(t *testing.T) {
		t.Parallel()

		for _, name := range []string{"a.txt", "b.zip", "c.chiptune"} {
			srcDir := t.TempDir()
			backup := dir.Directory(t.TempDir())
			extra := dir.Directory(t.TempDir())
			file := filepath.Join(srcDir, name)
			be.Err(t, os.WriteFile(file, []byte("data"), 0o600), nil)

			be.Err(t, config.RemoveDownload(testLog, name, file, backup, extra), nil)
			_, err := os.Stat(file)
			be.True(t, errors.Is(err, os.ErrNotExist))
			_, err = os.Stat(extra.Join(name))
			be.Err(t, err, nil)
		}
	})

	t.Run("invalid ext", func(t *testing.T) {
		t.Parallel()

		srcDir := t.TempDir()
		backup := dir.Directory(t.TempDir())
		extra := dir.Directory(t.TempDir())
		file := filepath.Join(srcDir, "junk.exe")
		be.Err(t, os.WriteFile(file, []byte("data"), 0o600), nil)

		be.Err(t, config.RemoveDownload(testLog, "junk.exe", file, backup, extra), nil)
		_, err := os.Stat(file)
		be.True(t, errors.Is(err, os.ErrNotExist))

		entries, err := os.ReadDir(backup.Path())
		be.Err(t, err, nil)
		be.Equal(t, len(entries), 1)
	})
}

func TestRemoveImage(t *testing.T) {
	t.Parallel()

	const (
		validUUID = "11111111-2222-3333-4444-555566667777"
		validCFID = "11111111-2222-3333-4444555566667777"
		badCFID   = "zzzzzzzz-zzzz-zzzz-zzzzzzzzzzzzzzzz"
	)

	t.Run("invalid args", func(t *testing.T) {
		t.Parallel()

		backup := dir.Directory(t.TempDir())
		be.True(t, errors.Is(config.RemoveImage(nil, "a.png", "/a.png", backup), nils.ErrSlogLogger))
		be.True(t, errors.Is(config.RemoveImage(testLog, "", "/a.png", backup), config.ErrNoPath))
	})

	t.Run("valid uuid", func(t *testing.T) {
		t.Parallel()

		for _, ext := range []string{".png", ".webp"} {
			srcDir := t.TempDir()
			backup := dir.Directory(t.TempDir())
			name := validUUID + ext
			file := filepath.Join(srcDir, name)
			be.Err(t, os.WriteFile(file, []byte("img"), 0o600), nil)

			be.Err(t, config.RemoveImage(testLog, name, file, backup), nil)
			_, err := os.Stat(file)
			be.Err(t, err, nil)
		}
	})

	t.Run("bad ext", func(t *testing.T) {
		t.Parallel()

		srcDir := t.TempDir()
		backup := dir.Directory(t.TempDir())
		name := validUUID + ".jpg"
		file := filepath.Join(srcDir, name)
		be.Err(t, os.WriteFile(file, []byte("img"), 0o600), nil)

		be.Err(t, config.RemoveImage(testLog, name, file, backup), nil)
		_, err := os.Stat(file)
		be.True(t, errors.Is(err, os.ErrNotExist))

		entries, err := os.ReadDir(backup.Path())
		be.Err(t, err, nil)
		be.Equal(t, len(entries), 1)
	})

	t.Run("invalid uuid", func(t *testing.T) {
		t.Parallel()

		srcDir := t.TempDir()
		backup := dir.Directory(t.TempDir())
		file := filepath.Join(srcDir, "not-a-uuid.png")
		be.Err(t, os.WriteFile(file, []byte("img"), 0o600), nil)

		be.Err(t, config.RemoveImage(testLog, "not-a-uuid.png", file, backup), nil)
		_, err := os.Stat(file)
		be.True(t, errors.Is(err, os.ErrNotExist))

		entries, err := os.ReadDir(backup.Path())
		be.Err(t, err, nil)
		be.Equal(t, len(entries), 1)
	})

	t.Run("cfid", func(t *testing.T) {
		t.Parallel()

		srcDir := t.TempDir()
		backup := dir.Directory(t.TempDir())
		name := validCFID + ".png"
		file := filepath.Join(srcDir, name)
		be.Err(t, os.WriteFile(file, []byte("img"), 0o600), nil)

		be.Err(t, config.RemoveImage(testLog, name, file, backup), nil)
		_, err := os.Stat(file)
		be.True(t, errors.Is(err, os.ErrNotExist))
		_, err = os.Stat(filepath.Join(srcDir, validUUID+".png"))
		be.Err(t, err, nil)
	})

	t.Run("bad 35-char", func(t *testing.T) {
		t.Parallel()

		srcDir := t.TempDir()
		backup := dir.Directory(t.TempDir())
		name := badCFID + ".png"
		file := filepath.Join(srcDir, name)
		be.Err(t, os.WriteFile(file, []byte("img"), 0o600), nil)

		be.Err(t, config.RemoveImage(testLog, name, file, backup), nil)
		_, err := os.Stat(filepath.Join(srcDir, ".png"))
		be.True(t, errors.Is(err, os.ErrNotExist))

		entries, err := os.ReadDir(backup.Path())
		be.Err(t, err, nil)
		be.Equal(t, len(entries), 1)
	})
}

func TestRenameDownload(t *testing.T) {
	t.Parallel()

	const (
		validUUID = "11111111-2222-3333-4444-555566667777"
		validCFID = "11111111-2222-3333-4444555566667777"
		badCFID   = "zzzzzzzz-zzzz-zzzz-zzzzzzzzzzzzzzzz"
	)

	t.Run("invalid args", func(t *testing.T) {
		t.Parallel()

		be.True(t, errors.Is(config.RenameDownload(nil, "a", "/a"), nils.ErrSlogLogger))
		be.True(t, errors.Is(config.RenameDownload(testLog, "", "/a"), config.ErrNoPath))
	})

	t.Run("cfid", func(t *testing.T) {
		t.Parallel()

		tmp := t.TempDir()
		file := filepath.Join(tmp, validCFID)
		be.Err(t, os.WriteFile(file, []byte("bin"), 0o600), nil)

		be.Err(t, config.RenameDownload(testLog, validCFID, file), nil)
		_, err := os.Stat(file)
		be.True(t, errors.Is(err, os.ErrNotExist))
		_, err = os.Stat(filepath.Join(tmp, validUUID))
		be.Err(t, err, nil)
	})

	t.Run("regular uuid", func(t *testing.T) {
		t.Parallel()

		tmp := t.TempDir()
		file := filepath.Join(tmp, validUUID)
		be.Err(t, os.WriteFile(file, []byte("bin"), 0o600), nil)

		be.Err(t, config.RenameDownload(testLog, validUUID, file), nil)
		_, err := os.Stat(file)
		be.Err(t, err, nil)
	})

	t.Run("bad 35-char", func(t *testing.T) {
		t.Parallel()

		tmp := t.TempDir()
		file := filepath.Join(tmp, badCFID)
		be.Err(t, os.WriteFile(file, []byte("bin"), 0o600), nil)

		be.Err(t, config.RenameDownload(testLog, badCFID, file))
		_, err := os.Stat(file)
		be.Err(t, err, nil)
	})
}
