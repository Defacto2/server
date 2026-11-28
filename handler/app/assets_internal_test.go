//nolint:paralleltest
package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Defacto2/helper"
	"github.com/Defacto2/server/internal/config"
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
	// Symlinks are non-regular and must be skipped.
	be.Err(t, os.Symlink(filepath.Join(root, "a.PNG"), filepath.Join(root, "link.png")), nil)

	count, total, exts, err := scanAssetDir(root)
	be.Err(t, err, nil)
	be.Equal(t, count, 5)
	be.Equal(t, total, int64(5+3+7+2+4))
	be.Equal(t, exts, []helper.Extension{
		{Name: ".png", Count: 3},
		{Name: ".webp", Count: 1},
		{Name: "uuid", Count: 1},
	})

	// Empty directory
	emptyCount, emptyTotal, emptyExts, err := scanAssetDir(t.TempDir())
	be.Err(t, err, nil)
	be.Equal(t, emptyCount, 0)
	be.Equal(t, emptyTotal, int64(0))
	be.Equal(t, len(emptyExts), 0)

	// Missing root directory
	_, _, _, err = scanAssetDir(filepath.Join(t.TempDir(), "missing"))
	be.Err(t, err)
}

func TestProcessAsset(t *testing.T) {
	t.Parallel()

	// Missing directory returns check error and zero defaults.
	badSpec := assetSpec{
		path:   filepath.Join(t.TempDir(), "missing"),
		desc:   "downloads",
		prefix: "Downloads",
		cache:  &AssetCache{c: -1, b: -1},
	}
	bad := processAsset(badSpec, false)
	be.True(t, bad["checkDownloads"] != nil)
	be.Equal(t, bad["countDownloads"], 0)

	// Valid directory: cold scan -> warm cache hit -> forced refresh.
	dirPath := t.TempDir()
	be.Err(t, os.WriteFile(filepath.Join(dirPath, "one.png"), []byte("1234"), 0o600), nil)
	spec := assetSpec{
		path:   dirPath,
		desc:   "previews",
		prefix: "Previews",
		cache:  &AssetCache{c: -1, b: -1},
	}

	cold := processAsset(spec, false)
	be.Equal(t, cold["checkPreviews"], nil)
	be.Equal(t, cold["countPreviews"], "1")

	// Add a second file: non-refresh read returns cached count.
	be.Err(t, os.WriteFile(filepath.Join(dirPath, "two.png"), []byte("5678"), 0o600), nil)
	warm := processAsset(spec, false)
	be.Equal(t, warm["countPreviews"], "1")

	// Refresh rescans and picks up the second file.
	refreshed := processAsset(spec, true)
	be.Equal(t, refreshed["countPreviews"], "2")
}

func TestHostStats(t *testing.T) {
	t.Parallel()

	conf := config.Config{AbsDownload: config.Absdown(t.TempDir())}
	got := HostStats(conf)

	for _, key := range []string{
		"wdTotal", "wdFree", "wdPercent",
		"tmpTotal", "tmpFree", "tmpPercent",
		"osTotal", "osFree", "osPercent",
		"metric0", "metric1",
		"memTotal", "memFree", "memPercent",
		"swapTotal", "swapFree", "swapPercent",
	} {
		_, ok := got[key]
		be.True(t, ok)
	}
	be.True(t, strings.HasSuffix(got["memPercent"].(string), "%"))
	be.True(t, strings.HasSuffix(got["swapPercent"].(string), "%"))
}

func TestAssetsData(t *testing.T) {
	t.Cleanup(func() {
		for _, c := range []*AssetCache{downloadCache, previewCache, thumbnailCache, extraCache, orphanCache} {
			c.Lock()
			c.c = -1
			c.b = -1
			c.exts = nil
			c.Unlock()
		}
	})

	conf := config.Config{
		AbsDownload:  config.Absdown(t.TempDir()),
		AbsPreview:   config.Absprev(t.TempDir()),
		AbsThumbnail: config.Absthumb(t.TempDir()),
		AbsExtra:     config.Absextra(t.TempDir()),
		AbsOrphaned:  config.Absorphan(t.TempDir()),
	}
	be.Err(t, os.WriteFile(filepath.Join(string(conf.AbsDownload), "a.zip"), []byte("zip"), 0o600), nil)

	specs := assetSpecs(conf)
	be.Equal(t, len(specs), 5)

	be.True(t, !AssetsCached())

	got := AssetsData(conf, true)
	be.True(t, AssetsCached())
	be.Equal(t, got["countDownloads"], "1")
	be.Equal(t, got["countPreviews"], "0")
	be.True(t, got["memTotal"] != "")
}
