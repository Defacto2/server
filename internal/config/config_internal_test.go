package config

import (
	"bytes"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Defacto2/server/internal/dir"
	"github.com/nalgeon/be"
)

func TestCheckLogDir(t *testing.T) {
	t.Parallel()

	c := Config{}
	err := c.checkLogDir()
	be.Err(t, err, nil)
	be.Equal(t, "", c.AbsLog.String())

	tmp := t.TempDir()
	c.AbsLog = Abslog(tmp)
	err = c.checkLogDir()
	be.Err(t, err, nil)

	st, err := os.Stat(c.AbsLog.String())
	be.Err(t, err, nil)
	be.True(t, st.IsDir())
}

func TestDryRunLeavesFilesUntouched(t *testing.T) {
	t.Parallel()

	dl := t.TempDir()
	preview := t.TempDir()
	thumb := t.TempDir()
	extra := t.TempDir()
	orphan := t.TempDir()

	badImg := filepath.Join(preview, "not-a-uuid.jpg")
	badDL := filepath.Join(dl, "junk.exe")
	zipDL := filepath.Join(dl, "keep.zip")
	stDir := filepath.Join(preview, syncthing)

	be.Err(t, os.WriteFile(badImg, []byte("x"), 0o600), nil)
	be.Err(t, os.WriteFile(badDL, []byte("x"), 0o600), nil)
	be.Err(t, os.WriteFile(zipDL, []byte("x"), 0o600), nil)
	be.Err(t, os.Mkdir(stDir, 0o700), nil)

	var buf bytes.Buffer
	sl := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo}))
	c := Config{
		AbsDownload:  Absdown(dl),
		AbsPreview:   Absprev(preview),
		AbsThumbnail: Absthumb(thumb),
		AbsExtra:     Absextra(extra),
		AbsOrphaned:  Absorphan(orphan),
	}

	be.Err(t, c.imageDirs(sl, true), nil)
	be.Err(t, downloadDir(sl, dir.Directory(dl), dir.Directory(orphan), dir.Directory(extra), true), nil)

	// Nothing on disk was moved, renamed, or deleted.
	for _, p := range []string{badImg, badDL, zipDL, stDir} {
		_, err := os.Stat(p)
		be.Err(t, err, nil)
	}
	entries, err := os.ReadDir(orphan)
	be.Err(t, err, nil)
	be.Equal(t, 0, len(entries))
	be.True(t, strings.Contains(buf.String(), "[dry-run]"))
}

func TestRanges(t *testing.T) {
	t.Parallel()

	for name := range informations {
		s := Configured(name, "")
		be.True(t, s != "")
	}
}
