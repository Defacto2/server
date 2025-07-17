package simple_test

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/Defacto2/server/handler/app/internal/simple"
	"github.com/Defacto2/server/internal/dir"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/volatiletech/null/v8"
)

func imagefiler(t *testing.T) string {
	t.Helper()
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
	s = simple.Updated("9:30pm", "")
	assert.Contains(t, s, "error")
	s = simple.Updated(time.Now(), "")
	assert.Contains(t, s, "Time just now")
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
	x = simple.ImageSample("", dir.Directory(filepath.Join("testdata", "TEST.png")))
	assert.Contains(t, x, missing)
	abs, err := filepath.Abs("testdata")
	require.NoError(t, err)
	const filenameNoExt = "TEST"
	x = simple.ImageSample(filenameNoExt, dir.Directory(abs))
	assert.Contains(t, x, "sha384-SK3qCpS11QMhNxUUnyeUeWWXBMPORDgLTI")
}

func TestImageSampleStat(t *testing.T) {
	t.Parallel()
	x := simple.ImageSampleStat("", "")
	assert.False(t, x)
	name := filepath.Base(imagefiler(t))
	name = strings.TrimSuffix(name, filepath.Ext(name))
	prev := filepath.Dir(imagefiler(t))
	x = simple.ImageSampleStat(name, dir.Directory(prev))
	assert.True(t, x)
}

func TestImageXY(t *testing.T) {
	t.Parallel()
	missing := [2]string{"0", ""}
	s := simple.ImageXY("")
	assert.Equal(t, missing, s)
	img := imagefiler(t)
	s = simple.ImageXY(img)
	assert.Equal(t, "4,163", s[0])
	assert.Equal(t, "500x500", s[1])
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
	require.NoError(t, err)
	assert.Equal(t, "/9b1c6", s)
}

func TestLinkRelr(t *testing.T) {
	t.Parallel()
	s, err := simple.LinkRelr("")
	require.Error(t, err)
	assert.Empty(t, s)
	s, err = simple.LinkRelr("a string")
	require.NoError(t, err)
	assert.Equal(t, "/g/a-string", s)
}

func TestMakeLink(t *testing.T) {
	t.Parallel()
	s, err := simple.MakeLink("", "", true)
	require.Error(t, err)
	assert.Empty(t, s)
	s, err = simple.MakeLink("tport", "", true)
	require.NoError(t, err)
	assert.Contains(t, s, "Tport")
	s, err = simple.MakeLink("tport", "", false)
	require.NoError(t, err)
	assert.Contains(t, s, "tPORt")
}

func TestMagicAsTitle(t *testing.T) {
	t.Parallel()
	s := simple.MagicAsTitle("")
	assert.Equal(t, "file not found", s)
	s = simple.MagicAsTitle(imagefiler(t))
	assert.Contains(t, s, "Portable Network Graphics")
}

func TestMIME(t *testing.T) {
	t.Parallel()
	s := simple.MIME("")
	assert.Equal(t, "file not found", s)
	s = simple.MIME(imagefiler(t))
	assert.Equal(t, "image/png", s)
}

func TestMkContent(t *testing.T) {
	t.Parallel()
	s := simple.MkContent("")
	assert.Empty(t, s)
	s = simple.MkContent("a string")
	assert.Contains(t, s, "a string")
	defer os.Remove(s)
}

func TestReleasers(t *testing.T) {
	t.Parallel()
	s := simple.Releasers("", "", true)
	assert.Empty(t, s)
	s = simple.Releasers("group 1", "group 2", false)
	assert.Contains(t, s, "group 1")
	assert.Contains(t, s, "group 2")
	s = simple.Releasers("group 1", "group 2", true)
	assert.Contains(t, s, "group 1")
	assert.Contains(t, s, "group 2")
	assert.Contains(t, s, "published by")
}

func TestScreenshot(t *testing.T) {
	t.Parallel()
	s := simple.Screenshot("", "", "")
	assert.Empty(t, s)
	prev := filepath.Dir(imagefiler(t))
	s = simple.Screenshot("TEST", "test", dir.Directory(prev))
	assert.Contains(t, s, `alt="test screenshot"`)
	assert.Contains(t, s, `<img src="/public/image`)
}

func TestStatHumanize(t *testing.T) {
	t.Parallel()
	x, y, z := simple.StatHumanize("")
	const none = "file not found"
	assert.Equal(t, none, x)
	assert.Equal(t, none, y)
	assert.Equal(t, none, z)
	x, y, z = simple.StatHumanize(imagefiler(t))
	assert.Contains(t, x, "202") // a year prefix
	assert.Equal(t, "4,163", y)
	assert.Contains(t, z, "4.2 kB")
}

func TestThumb(t *testing.T) {
	t.Parallel()
	s := simple.Thumb("", "", "", false)
	assert.Contains(t, "<!-- no thumbnail found -->", s)
	name := filepath.Base(imagefiler(t))
	name = strings.TrimSuffix(name, filepath.Ext(name))
	thumb := dir.Directory(filepath.Dir(imagefiler(t)))
	s = simple.Thumb(name, "a description", thumb, false)
	assert.Contains(t, s, `alt="a description thumbnail"`)
}

func TestThumbSample(t *testing.T) {
	t.Parallel()
	const missing = "No thumbnail"
	x := simple.ThumbSample("", "")
	assert.Contains(t, x, missing)
	name := filepath.Base(imagefiler(t))
	name = strings.TrimSuffix(name, filepath.Ext(name))
	thumb := filepath.Dir(imagefiler(t))
	x = simple.ThumbSample(name, dir.Directory(thumb))
	assert.Contains(t, x, "sha384-SK3qCpS11QMhNxUUnyeUeWWXBMPORDgLTI")
}
