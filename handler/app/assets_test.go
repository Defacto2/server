package app

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Defacto2/helper"
	"github.com/nalgeon/be"
)

func TestScanAssetDir(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	sub := filepath.Join(root, "nested")
	be.Err(t, os.Mkdir(sub, 0o750), nil)

	files := map[string][]byte{
		filepath.Join(root, "a.PNG"):          []byte("12345"),
		filepath.Join(root, "b.png"):          []byte("123"),
		filepath.Join(root, "c.webp"):         []byte("1234567"),
		filepath.Join(root, "noext-uuid-key"): []byte("12"),
		filepath.Join(root, helper.DSStore):   []byte("ignored"),
		filepath.Join(sub, "d.png"):           []byte("1234"),
	}
	for p, body := range files {
		be.Err(t, os.WriteFile(p, body, 0o600), nil)
	}

	count, total, exts, err := scanAssetDir(root)
	be.Err(t, err, nil)
	be.Equal(t, count, 5)
	be.Equal(t, total, int64(5+3+7+2+4))
	be.Equal(t, exts, []helper.Extension{
		{Name: ".png", Count: 3},
		{Name: ".webp", Count: 1},
		{Name: "uuid", Count: 1},
	})
}
