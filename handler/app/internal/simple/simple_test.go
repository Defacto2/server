package simple_test

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/Defacto2/server/handler/app/internal/simple"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/volatiletech/null/v8"
)

func testImage(t *testing.T) string {
	_, file, _, ok := runtime.Caller(0)
	require.True(t, ok)
	return filepath.Join(filepath.Dir(file), "testdata", "TEST.png")
}

func TestAssetSrc(t *testing.T) {
	t.Parallel()
	s := simple.AssetSrc("", "", "", "")
	assert.Equal(t, "integrity os.readfile open : no such file or directory", s)
	_, file, _, ok := runtime.Caller(0)
	require.True(t, ok)
	s = simple.AssetSrc("", file, "", "")
	assert.Contains(t, s, "sha384-")
}

func TestDownloadB(t *testing.T) {
	t.Parallel()
	x := simple.DownloadB("")
	assert.Contains(t, x, "received an invalid type")
	x = simple.DownloadB("a string")
	assert.Contains(t, x, "received an invalid type")
	x = simple.DownloadB("1")
	assert.Contains(t, x, "received an invalid type")
	x = simple.DownloadB(null.Int64From(1))
	assert.Contains(t, x, "1 B")
	x = simple.DownloadB(1024)
	assert.Contains(t, x, "(1k)")
}

func TestLinkRelations(t *testing.T) {
	t.Parallel()
	x := simple.LinkRelations("")
	assert.Empty(t, x)
	x = simple.LinkRelations("nfo file;aa2165c")
	assert.Contains(t, x, "/f/aa2165c")
	x = simple.LinkRelations("nfo file;aa2165c|readme;a822ea8")
	assert.Contains(t, x, "/f/aa2165c")
	assert.Contains(t, x, "/f/a822ea8")
	x = simple.LinkRelations("nfo file;xxxxx")
	assert.Contains(t, x, "invalid download path")
}

func TestLinkSites(t *testing.T) {
	t.Parallel()
	x := simple.LinkSites("")
	assert.Empty(t, x)
	x = simple.LinkSites("a string")
	assert.Empty(t, x)
	x = simple.LinkSites("example.com")
	assert.Empty(t, x)
	x = simple.LinkSites("example.com|example.org")
	assert.Empty(t, x)
	x = simple.LinkSites("example;example.org")
	assert.Contains(t, x, "https://example.org")
	x = simple.LinkSites("example;example.org|another example;example.net")
	assert.Contains(t, x, "https://example.org")
	assert.Contains(t, x, "https://example.net")
	x = simple.LinkSites("example.com|||example.org")
	assert.Empty(t, x)
	x = simple.LinkSites("example.com;;;example.org")
	assert.Empty(t, x)
}

func TestLinkPreviewTip(t *testing.T) {
	t.Parallel()
	s := simple.LinkPreviewTip("", "")
	assert.Empty(t, s)
	s = simple.LinkPreviewTip(".zip", "windows")
	assert.Empty(t, s)
	s = simple.LinkPreviewTip(".txt", "windows")
	assert.Equal(t, "Read this as text", s)
}

func TestReleaserPair(t *testing.T) {
	t.Parallel()
	s := simple.ReleaserPair(nil, nil)
	assert.Empty(t, s)
	s = simple.ReleaserPair("1", "2")
	assert.Equal(t, "1", s[0])
	assert.Equal(t, "2", s[1])
	s = simple.ReleaserPair(nil, "2")
	assert.Equal(t, "2", s[0])
	assert.Empty(t, s[1])
}

func TestUpdated(t *testing.T) {
	t.Parallel()
	s := simple.Updated(nil, "")
	assert.Empty(t, s)
	s = simple.Updated(time.Now(), "")
	assert.Contains(t, s, "Time less than 5 seconds ago")
}

func TestDemozooGetLink(t *testing.T) {
	t.Parallel()
	html := simple.DemozooGetLink("", "", "", "")
	assert.Empty(t, html)
	fn := null.String{}
	fs := null.Int64{}
	dz := null.Int64{}
	un := null.String{}
	html = simple.DemozooGetLink(fn, fs, dz, un)
	assert.Empty(t, html)

	fn = null.StringFrom("file")
	html = simple.DemozooGetLink(fn, fs, dz, un)
	assert.Empty(t, html)

	fn = null.String{}
	fs = null.Int64From(1000)
	html = simple.DemozooGetLink(fn, fs, dz, un)
	assert.Empty(t, html)

	fn = null.String{}
	fs = null.Int64{}
	dz = null.Int64From(1)
	un = null.StringFrom("user")
	html = simple.DemozooGetLink(fn, fs, dz, un)
	assert.NotEmpty(t, html)
}

func TestImageSample(t *testing.T) {
	t.Parallel()
	const missing = "No preview image file"
	x := simple.ImageSample("", "")
	assert.Contains(t, x, missing)
	// note: the filename extension is case-sensitive.
	x = simple.ImageSample("", filepath.Join("testdata", "TEST.png"))
	assert.Contains(t, x, missing)
	abs, err := filepath.Abs("testdata")
	require.NoError(t, err)
	const filenameNoExt = "TEST"
	x = simple.ImageSample(filenameNoExt, abs)
	assert.Contains(t, x, "sha384-SK3qCpS11QMhNxUUnyeUeWWXBMPORDgLTI")
}

func TestImageSampleStat(t *testing.T) {
	t.Parallel()
	x := simple.ImageSampleStat("", "")
	assert.False(t, x)
	name := filepath.Base(testImage(t))
	name = strings.TrimSuffix(name, filepath.Ext(name))
	dir := filepath.Dir(testImage(t))
	x = simple.ImageSampleStat(name, dir)
	assert.True(t, x)
}

func TestImageXY(t *testing.T) {
	t.Parallel()
	s := simple.ImageXY("")
	assert.Contains(t, s, "stat : no such file or directory")
	img := testImage(t)
	s = simple.ImageXY(img)
	fmt.Println(img)
	assert.Equal(t, s[0], "4,163")
	assert.Equal(t, s[1], "500x500")
}

func TestLinkID(t *testing.T) {
	t.Parallel()
	s, err := simple.LinkID("", "")
	require.Error(t, err)
	assert.Empty(t, s)
	s, err = simple.LinkID("a string", "a string")
	require.Error(t, err)
	assert.Empty(t, s)
	s, err = simple.LinkID(1, "")
	assert.NoError(t, err)
	assert.Equal(t, "/9b1c6", s)
}
