package filerecord

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Defacto2/server/internal/logs"
	"github.com/nalgeon/be"
)

func TestRenameExtractedUUID(t *testing.T) {
	t.Parallel()

	const unid = "0192883a-b587-7659-befd-eb90e395953d"

	// Single unpacked UUID file is renamed to the artifact filename.
	dir1 := t.TempDir()
	src1 := filepath.Join("/downloads", unid)
	be.Err(t, os.WriteFile(filepath.Join(dir1, unid), []byte("unpacked"), 0o600), nil)

	renUUID(logs.Discard(), dir1, src1, "KURUPT.COM")

	_, err := os.Stat(filepath.Join(dir1, unid))
	be.True(t, os.IsNotExist(err))
	got, err := os.ReadFile(filepath.Join(dir1, "KURUPT.COM"))
	be.Err(t, err, nil)
	be.Equal(t, string(got), "unpacked")

	// Stale existing target file alongside unpacked UUID file is replaced cleanly.
	dir2 := t.TempDir()
	be.Err(t, os.WriteFile(filepath.Join(dir2, "KURUPT.COM"), []byte("stale-packed"), 0o600), nil)
	be.Err(t, os.WriteFile(filepath.Join(dir2, unid), []byte("fresh-unpacked"), 0o600), nil)

	renUUID(logs.Discard(), dir2, src1, "KURUPT.COM")

	_, err = os.Stat(filepath.Join(dir2, unid))
	be.True(t, os.IsNotExist(err))
	got, err = os.ReadFile(filepath.Join(dir2, "KURUPT.COM"))
	be.Err(t, err, nil)
	be.Equal(t, string(got), "fresh-unpacked")

	// Normal archives with other files are left untouched.
	dir3 := t.TempDir()
	be.Err(t, os.WriteFile(filepath.Join(dir3, unid), []byte("x"), 0o600), nil)
	be.Err(t, os.WriteFile(filepath.Join(dir3, "OTHER.NFO"), []byte("y"), 0o600), nil)

	renUUID(logs.Discard(), dir3, src1, "KURUPT.COM")

	_, err = os.Stat(filepath.Join(dir3, unid))
	be.Err(t, err, nil)
}
