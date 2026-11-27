package config_test

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	"github.com/Defacto2/server/internal/config"
	"github.com/Defacto2/server/internal/dir"
	"github.com/Defacto2/server/internal/testutil"
	"github.com/nalgeon/be"
)

func TestTempInfo(t *testing.T) {
	t.Parallel()

	tmp, err := dir.MkdirTemp("tempinfo")
	be.Err(t, err, nil)
	t.Cleanup(func() {
		_ = dir.RemoveAll(tmp)
	})
	be.Err(t, os.WriteFile(filepath.Join(tmp, "sample.bin"), make([]byte, 1024), 0o600), nil)

	buff := testutil.Buffer(t)

	config.TempInfo(buff.Log)

	be.True(t, buff.Contains("Temporary directory"))
	be.True(t, buff.Contains("Usage="))
	be.True(t, !buff.Contains(`Usage="0 B"`))
}

func TestRecordCount(t *testing.T) {
	t.Parallel()

	t.Run("nil ctx", func(t *testing.T) {
		t.Parallel()
		defer func() {
			if r := recover(); r == nil {
				t.Fatal("expected panic")
			}
		}()
		_ = config.RecordCount(nil, new(sql.DB)) //nolint:staticcheck // testing nil context
	})

	t.Run("nil exec", func(t *testing.T) {
		t.Parallel()
		defer func() {
			if r := recover(); r == nil {
				t.Fatal("expected panic")
			}
		}()
		_ = config.RecordCount(t.Context(), nil)
	})

	t.Run("db", func(t *testing.T) {
		t.Parallel()
		db := testutil.DB(t)
		got := config.RecordCount(t.Context(), db)
		be.True(t, got > 2)
	})
}
