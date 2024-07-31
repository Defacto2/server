package str_test

import (
	"html/template"
	"testing"
	"time"

	"github.com/Defacto2/server/handler/app/internal/str"
	"github.com/stretchr/testify/assert"
	"github.com/volatiletech/null/v8"
)

func TestDownloadB(t *testing.T) {
	t.Parallel()
	x := str.DownloadB("")
	assert.Contains(t, x, "received an invalid type")
	x = str.DownloadB("a string")
	assert.Contains(t, x, "received an invalid type")
	x = str.DownloadB("1")
	assert.Contains(t, x, "received an invalid type")
	x = str.DownloadB(null.Int64From(1))
	assert.Contains(t, x, "1 B")
	x = str.DownloadB(1024)
	assert.Contains(t, x, "(1k)")
}

func TestLinkRelations(t *testing.T) {
	t.Parallel()
	x := str.LinkRelations("")
	assert.Empty(t, x)
	x = str.LinkRelations("nfo file;aa2165c")
	assert.Contains(t, x, "/f/aa2165c")
	x = str.LinkRelations("nfo file;aa2165c|readme;a822ea8")
	assert.Contains(t, x, "/f/aa2165c")
	assert.Contains(t, x, "/f/a822ea8")
	x = str.LinkRelations("nfo file;xxxxx")
	assert.Contains(t, x, "invalid download path")
}

func TestLinkSites(t *testing.T) {
	t.Parallel()
	x := str.LinkSites("")
	assert.Empty(t, x)
	x = str.LinkSites("a string")
	assert.Empty(t, x)
	x = str.LinkSites("example.com")
	assert.Empty(t, x)
	x = str.LinkSites("example.com|example.org")
	assert.Empty(t, x)
	x = str.LinkSites("example;example.org")
	assert.Contains(t, x, "https://example.org")
	x = str.LinkSites("example;example.org|another example;example.net")
	assert.Contains(t, x, "https://example.org")
	assert.Contains(t, x, "https://example.net")
	x = str.LinkSites("example.com|||example.org")
	assert.Empty(t, x)
	x = str.LinkSites("example.com;;;example.org")
	assert.Empty(t, x)
}

func TestLinkPreviewTip(t *testing.T) {
	t.Parallel()
	s := str.LinkPreviewTip("", "")
	assert.Empty(t, s)
	s = str.LinkPreviewTip(".zip", "windows")
	assert.Empty(t, s)
	s = str.LinkPreviewTip(".txt", "windows")
	assert.Equal(t, "Read this as text", s)
}

func TestReleaserPair(t *testing.T) {
	t.Parallel()
	s := str.ReleaserPair(nil, nil)
	assert.Empty(t, s)
	s = str.ReleaserPair("1", "2")
	assert.Equal(t, "1", s[0])
	assert.Equal(t, "2", s[1])
	s = str.ReleaserPair(nil, "2")
	assert.Equal(t, "2", s[0])
	assert.Empty(t, s[1])
}

func TestUpdated(t *testing.T) {
	t.Parallel()
	s := str.Updated(nil, "")
	assert.Empty(t, s)
	s = str.Updated(time.Now(), "")
	assert.Contains(t, s, "Time less than 5 seconds ago")
}

func TestDemozooGetLink(t *testing.T) {
	t.Parallel()
	html := str.DemozooGetLink("", "", "", "")
	assert.Equal(t, template.HTML("no id provided"), html)
	fn := null.String{}
	fs := null.Int64{}
	dz := null.Int64{}
	un := null.String{}
	html = str.DemozooGetLink(fn, fs, dz, un)
	assert.Empty(t, html)

	fn = null.StringFrom("file")
	html = str.DemozooGetLink(fn, fs, dz, un)
	assert.Empty(t, html)

	fn = null.String{}
	fs = null.Int64From(1000)
	html = str.DemozooGetLink(fn, fs, dz, un)
	assert.Empty(t, html)

	fn = null.String{}
	fs = null.Int64{}
	dz = null.Int64From(1)
	un = null.StringFrom("user")
	html = str.DemozooGetLink(fn, fs, dz, un)
	assert.NotEmpty(t, html)
}

func TestImageSample(t *testing.T) {
	t.Parallel()
	x := str.ImageSample("", "")
	assert.Contains(t, x, "no such file")
	x = str.ImageSample("", "testdata/TEST.PNG")
	assert.Contains(t, x, "no such file")
	x = str.ImageSample("", "testdata/test")
	assert.Contains(t, x, "sha384-SK3qCpS11QMhNxUUnyeUeWWXBMPORDgLTI")
}
