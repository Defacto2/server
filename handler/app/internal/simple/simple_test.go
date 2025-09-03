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
	"github.com/aarondl/null/v8"
	"github.com/nalgeon/be"
)

func imagefiler(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	be.True(t, ok)
	return filepath.Join(filepath.Dir(file), "testdata", "TEST.png")
}

func TestAssetSrc(t *testing.T) {
	t.Parallel()
	s := simple.AssetSrc("", "", "", "")
	be.Equal(t, "integrity os.readfile open : no such file or directory", s)
	_, file, _, ok := runtime.Caller(0)
	be.True(t, ok)
	s = simple.AssetSrc("", file, "", "")
	be.True(t, strings.Contains(s, "sha384-"))
}

func TestDownloadB(t *testing.T) {
	t.Parallel()
	x := simple.DownloadB("")
	be.True(t, strings.Contains(string(x), "received an invalid type"))
	x = simple.DownloadB("a string")
	be.True(t, strings.Contains(string(x), "received an invalid type"))
	x = simple.DownloadB("1")
	be.True(t, strings.Contains(string(x), "received an invalid type"))
	x = simple.DownloadB(null.Int64From(1))
	be.True(t, strings.Contains(string(x), "1 B"))
	x = simple.DownloadB(1024)
	be.True(t, strings.Contains(string(x), "(1k)"))
}

func TestLinkRelations(t *testing.T) {
	t.Parallel()
	x := simple.LinkRelations("")
	be.True(t, string(x) == "")
	x = simple.LinkRelations("nfo file;aa2165c")
	be.True(t, strings.Contains(string(x), "/f/aa2165c"))
	x = simple.LinkRelations("nfo file;aa2165c|readme;a822ea8")
	be.True(t, strings.Contains(string(x), "/f/aa2165c"))
	be.True(t, strings.Contains(string(x), "/f/a822ea8"))
	x = simple.LinkRelations("nfo file;xxxxx")
	be.True(t, strings.Contains(string(x), "invalid download path"))
}

func TestLinkSites(t *testing.T) {
	t.Parallel()
	x := simple.LinkSites("")
	be.True(t, string(x) == "")
	x = simple.LinkSites("a string")
	be.True(t, string(x) == "")
	x = simple.LinkSites("example.com")
	be.True(t, string(x) == "")
	x = simple.LinkSites("example.com|example.org")
	be.True(t, string(x) == "")
	x = simple.LinkSites("example;example.org")
	be.True(t, strings.Contains(string(x), "https://example.org"))
	x = simple.LinkSites("example;example.org|another example;example.net")
	be.True(t, strings.Contains(string(x), "https://example.org"))
	be.True(t, strings.Contains(string(x), "https://example.net"))
	x = simple.LinkSites("example.com|||example.org")
	be.True(t, string(x) == "")
	x = simple.LinkSites("example.com;;;example.org")
	be.True(t, string(x) == "")
}

func TestLinkPreviewTip(t *testing.T) {
	t.Parallel()
	s := simple.LinkPreviewTip("", "")
	be.Equal(t, s, "")
	s = simple.LinkPreviewTip(".zip", "windows")
	be.Equal(t, s, "")
	s = simple.LinkPreviewTip(".txt", "windows")
	be.Equal(t, "Read this as text", s)
}

func TestReleaserPair(t *testing.T) {
	t.Parallel()
	s := simple.ReleaserPair(nil, nil)
	be.Equal(t, s, [2]string{})
	s = simple.ReleaserPair("1", "2")
	be.Equal(t, "1", s[0])
	be.Equal(t, "2", s[1])
	s = simple.ReleaserPair(nil, "2")
	be.Equal(t, "2", s[0])
	be.Equal(t, s[1], "")
}

func TestUpdated(t *testing.T) {
	t.Parallel()
	s := simple.Updated(nil, "")
	be.Equal(t, s, "")
	s = simple.Updated("9:30pm", "")
	be.True(t, strings.Contains(s, "error"))
	s = simple.Updated(time.Now(), "")
	be.True(t, strings.Contains(s, "Time just now"))
}

func TestDemozooGetLink(t *testing.T) {
	t.Parallel()
	html := simple.DemozooGetLink("", "", "", "")
	be.Equal(t, html, "")
	fn := null.String{}
	fs := null.Int64{}
	dz := null.Int64{}
	un := null.String{}
	html = simple.DemozooGetLink(fn, fs, dz, un)
	be.Equal(t, html, "")

	fn = null.StringFrom("file")
	html = simple.DemozooGetLink(fn, fs, dz, un)
	be.Equal(t, html, "")

	fn = null.String{}
	fs = null.Int64From(1000)
	html = simple.DemozooGetLink(fn, fs, dz, un)
	be.Equal(t, html, "")

	fn = null.String{}
	fs = null.Int64{}
	dz = null.Int64From(1)
	un = null.StringFrom("user")
	html = simple.DemozooGetLink(fn, fs, dz, un)
	be.True(t, html != "")
}

func TestImageSample(t *testing.T) {
	t.Parallel()
	const missing = "No preview image file"
	x := simple.ImageSample("", "")
	be.True(t, strings.Contains(string(x), missing))
	// note: the filename extension is case-sensitive.
	x = simple.ImageSample("", dir.Directory(filepath.Join("testdata", "TEST.png")))
	be.True(t, strings.Contains(string(x), missing))
	abs, err := filepath.Abs("testdata")
	be.Err(t, err, nil)
	const filenameNoExt = "TEST"
	x = simple.ImageSample(filenameNoExt, dir.Directory(abs))
	be.True(t, strings.Contains(string(x), "sha384-SK3qCpS11QMhNxUUnyeUeWWXBMPORDgLTI"))
}

func TestImageSampleStat(t *testing.T) {
	t.Parallel()
	x := simple.ImageSampleStat("", "")
	be.True(t, !x)
	name := filepath.Base(imagefiler(t))
	name = strings.TrimSuffix(name, filepath.Ext(name))
	prev := filepath.Dir(imagefiler(t))
	x = simple.ImageSampleStat(name, dir.Directory(prev))
	be.True(t, x)
}

func TestImageXY(t *testing.T) {
	t.Parallel()
	missing := [2]string{"0", ""}
	s := simple.ImageXY("")
	be.Equal(t, missing, s)
	img := imagefiler(t)
	s = simple.ImageXY(img)
	be.Equal(t, "4,163", s[0])
	be.Equal(t, "500x500", s[1])
}

func TestLinkID(t *testing.T) {
	t.Parallel()
	s, err := simple.LinkID("", "")
	be.Err(t, err)
	be.Equal(t, s, "")
	s, err = simple.LinkID("a string", "a string")
	be.Err(t, err)
	be.Equal(t, s, "")
	s, err = simple.LinkID(1, "")
	be.Err(t, err, nil)
	be.Equal(t, "/9b1c6", s)
}

func TestLinkRelr(t *testing.T) {
	t.Parallel()
	s, err := simple.LinkRelr("")
	be.Err(t, err)
	be.Equal(t, s, "")
	s, err = simple.LinkRelr("a string")
	be.Err(t, err, nil)
	be.Equal(t, "/g/a-string", s)
}

func TestMakeLink(t *testing.T) {
	t.Parallel()
	s, err := simple.MakeLink("", "", "", true)
	be.Err(t, err)
	be.Equal(t, s, "")
	s, err = simple.MakeLink("", "tport", "", true)
	be.Err(t, err, nil)
	be.True(t, strings.Contains(s, "Tport"))
	s, err = simple.MakeLink("", "tport", "", false)
	be.Err(t, err, nil)
	be.True(t, strings.Contains(s, "tPORt"))
}

func TestMagicAsTitle(t *testing.T) {
	t.Parallel()
	s := simple.MagicAsTitle("")
	be.Equal(t, "file not found", s)
	s = simple.MagicAsTitle(imagefiler(t))
	be.True(t, strings.Contains(s, "Portable Network Graphics"))
}

func TestMIME(t *testing.T) {
	t.Parallel()
	s := simple.MIME("")
	be.Equal(t, "file not found", s)
	s = simple.MIME(imagefiler(t))
	be.Equal(t, "image/png", s)
}

func TestMkContent(t *testing.T) {
	t.Parallel()
	s := simple.MkContent("")
	be.Equal(t, s, "")
	s = simple.MkContent("a string")
	be.True(t, strings.Contains(s, "a string"))
	defer func() { _ = os.Remove(s) }()
}

func TestReleasers(t *testing.T) {
	t.Parallel()
	s := simple.Releasers("", "", true)
	be.Equal(t, s, "")
	s = simple.Releasers("group 1", "group 2", false)
	be.True(t, strings.Contains(string(s), "group 1"))
	be.True(t, strings.Contains(string(s), "group 2"))
	s = simple.Releasers("group 1", "group 2", true)
	be.True(t, strings.Contains(string(s), "group 1"))
	be.True(t, strings.Contains(string(s), "group 2"))
	be.True(t, strings.Contains(string(s), "published by"))
}

func TestScreenshot(t *testing.T) {
	t.Parallel()
	s := simple.Screenshot("", "", "")
	be.Equal(t, s, "")
	prev := filepath.Dir(imagefiler(t))
	s = simple.Screenshot("TEST", "test", dir.Directory(prev))
	be.True(t, strings.Contains(string(s), `alt="test screenshot"`))
	be.True(t, strings.Contains(string(s), `<img src="/public/image`))
}

func TestStatHumanize(t *testing.T) {
	t.Parallel()
	x, y, z := simple.StatHumanize("")
	const none = "file not found"
	be.Equal(t, none, x)
	be.Equal(t, none, y)
	be.Equal(t, none, z)
	x, y, z = simple.StatHumanize(imagefiler(t))
	be.True(t, strings.Contains(x, "202")) // a year prefix
	be.Equal(t, "4,163", y)
	be.True(t, strings.Contains(z, "4.2 kB"))
}

func TestThumb(t *testing.T) {
	t.Parallel()
	s := simple.Thumb("", "", "", false)
	be.True(t, strings.Contains(string(s), "<!-- no thumbnail found -->"))
	name := filepath.Base(imagefiler(t))
	name = strings.TrimSuffix(name, filepath.Ext(name))
	thumb := dir.Directory(filepath.Dir(imagefiler(t)))
	s = simple.Thumb(name, "a description", thumb, false)
	be.True(t, strings.Contains(string(s), `alt="a description thumbnail"`))
}

func TestThumbSample(t *testing.T) {
	t.Parallel()
	const missing = "No thumbnail"
	x := simple.ThumbSample("", "")
	be.True(t, strings.Contains(string(x), missing))
	name := filepath.Base(imagefiler(t))
	name = strings.TrimSuffix(name, filepath.Ext(name))
	thumb := filepath.Dir(imagefiler(t))
	x = simple.ThumbSample(name, dir.Directory(thumb))
	be.True(t, strings.Contains(string(x), "sha384-SK3qCpS11QMhNxUUnyeUeWWXBMPORDgLTI"))
}
