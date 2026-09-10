package app_test

import (
	"embed"
	"strings"
	"testing"
	"time"

	"github.com/Defacto2/server/handler/app"
	"github.com/Defacto2/server/internal/tags"
	"github.com/Defacto2/server/internal/testutil"
	"github.com/aarondl/null/v8"
	"github.com/nalgeon/be"
)

//go:embed*
var emptyFS embed.FS

const (
	exampleURL  = "https://example.com"
	exampleWiki = "/some/wiki/page"
)

func TestAttribute(t *testing.T) {
	t.Parallel()

	got := app.Attribute("", "", "", "", "")
	be.Equal(t, got, "")

	got = app.Attribute("writer1",
		"", "", "", "")
	be.Equal(t, got, "")

	got = app.Attribute("writer",
		"", "", "", "some scener")
	be.True(t, strings.Contains(got, "error:"))

	got = app.Attribute("another person,writer,some scener",
		"", "", "", "some scener")
	be.Equal(t, got, "Writer attribution")

	got = app.Attribute("another person,writer,some scener",
		"", "some scener", "", "some scener")
	be.Equal(t, got, "Writer and artist attributions")

	got = app.Attribute("another person,writer,ben",
		"ben", "", "", "ben")
	be.Equal(t, got, "Writer and programmer attributions")
}

func TestBrief(t *testing.T) {
	t.Parallel()

	got := app.Brief("", "")
	be.Equal(t, got, "an unknown release")

	got = app.Brief("a string", "")
	be.True(t, strings.Contains(got, "unknown platform"))
	plat := null.StringFrom("windows")

	got = app.Brief(plat, "")
	be.True(t, strings.Contains(got, "unknown section"))

	sect := null.StringFrom(tags.Intro.String())
	got = app.Brief(plat, sect)
	be.True(t, strings.Contains(got, "a Windows intro"))
}

func TestByteBytes(t *testing.T) {
	t.Parallel()

	s := app.ByteBytes("")
	be.True(t, strings.Contains(string(s), "error"))

	s = app.ByteBytes(1)
	be.True(t, !strings.Contains(string(s), "error"))

	s = app.ByteBytes(1023)
	h := string(s)
	be.True(t, strings.Contains(h, "1 kB "))
	be.True(t, strings.Contains(h, "(1023B)"))
}

func TestByteFile(t *testing.T) {
	t.Parallel()

	s := app.ByteFile("", "")
	be.True(t, strings.Contains(string(s), "error"))

	s = app.ByteFile(1, "")
	be.True(t, strings.Contains(string(s), "error"))

	s = app.ByteFile("", 1)
	be.True(t, strings.Contains(string(s), "error"))

	s = app.ByteFile(12, 1023)
	h := string(s)
	be.True(t, strings.Contains(h, "12 "))
	be.True(t, strings.Contains(h, "(1 kB)"))
}

func TestByteFileS(t *testing.T) {
	t.Parallel()

	const intro = "intro"
	got := app.ByteFileS("", "", "")
	be.True(t, strings.Contains(string(got), "error"))

	got = app.ByteFileS(intro, 1, "")
	be.True(t, strings.Contains(string(got), "error"))

	got = app.ByteFileS(intro, "", 1)
	be.True(t, strings.Contains(string(got), "error"))

	got = app.ByteFileS(intro, 1, 50000)
	be.True(t, strings.Contains(string(got), "1 intro"))
	be.True(t, strings.Contains(string(got), "50 kB"))

	got = app.ByteFileS(intro, 12, 1023)
	be.True(t, strings.Contains(string(got), "12 intros"))
	be.True(t, strings.Contains(string(got), "1 kB"))
}

func TestDay(t *testing.T) {
	t.Parallel()

	got := app.Day("")
	be.True(t, strings.Contains(got, "error"))

	got = app.Day("1")
	be.True(t, strings.Contains(got, "error"))

	got = app.Day(1)
	be.True(t, strings.Contains(got, " 1"))

	got = app.Day(100)
	be.True(t, strings.Contains(got, "error"))
}

func TestDescribe(t *testing.T) {
	t.Parallel()

	s := app.Describe("", "", "", "")
	be.True(t, strings.Contains(string(s), "error"))

	s = app.Describe("", "", 1900, 50)
	be.True(t, strings.Contains(string(s), "unknown release"))

	s = app.Describe("x", "y", 1980, 1)
	be.True(t, strings.Contains(string(s), "Unknown platform"))
	be.True(t, strings.Contains(string(s), "Jan, 1980"))

	plat := null.StringFrom(tags.ANSI.String())
	s = app.Describe(plat, "y", 1980, 1)
	be.True(t, strings.Contains(string(s), "Unknown section"))

	sect := null.StringFrom(tags.BBS.String())
	year := null.Int16From(1990)
	month := null.Int16From(12)

	s = app.Describe(plat, sect, year, month)
	h := string(s)

	be.True(t, strings.Contains(h, "ansi BBS advert"))
	be.True(t, strings.Contains(h, "Dec, 1990"))
}

func TestGlobTo(t *testing.T) {
	t.Parallel()

	got := app.GlobTo("file.css")
	be.Equal(t, got, "view/app/file.css")
}

func TestLastUpdated(t *testing.T) {
	t.Parallel()

	got := app.LastUpdated(nil)
	be.Equal(t, got, "")

	oneHourAgo := time.Now().Add(-time.Hour)
	got = app.LastUpdated(oneHourAgo)
	be.Equal(t, got, "Last updated about 1 hour ago")
}

func TestLinkDownload(t *testing.T) {
	t.Parallel()

	s := string(app.LinkDownload("", ""))
	be.True(t, strings.Contains(s, "invalid"))

	s = string(app.LinkDownload(1, ""))
	be.True(t, strings.Contains(s, "/d/9b1c6"))
}

func TestLinkHref(t *testing.T) {
	t.Parallel()

	s, err := app.LinkHref(nil)
	be.Equal(t, s, "")
	be.Err(t, err)

	s, err = app.LinkHref(0)
	be.Equal(t, s, "")
	be.Err(t, err)

	s, err = app.LinkHref(1)
	be.True(t, strings.Contains(s, "/f/9b1c6"))
	be.Err(t, err, nil)
}

func TestLinkInterview(t *testing.T) {
	t.Parallel()

	s := string(app.LinkInterview(""))
	be.True(t, strings.Contains(s, "error"))

	s = string(app.LinkInterview("x"))
	be.Equal(t, s, "")

	s = string(app.LinkInterview("https://example.com"))
	be.True(t, strings.Contains(s, "#arrow-right"))
}

func TestLinkPage(t *testing.T) {
	t.Parallel()

	got := app.LinkPage(nil, nil)
	be.Equal(t, got, "")

	got = app.LinkPage(1, nil)
	be.Equal(t, got, `<a class="card-link" href="/f/9b1c6" rel="nofollow">Artifact</a>`)

	got = app.LinkPage(1, "c")
	be.True(t, !strings.Contains(string(got), "tooltip"))

	got = app.LinkPage(1, 10)
	be.True(t, strings.Contains(string(got), "tooltip"))
}

func TestLinkRunApp(t *testing.T) {
	t.Parallel()

	got := app.LinkRunApp(nil)
	be.Equal(t, got, "")

	got = app.LinkRunApp(1)
	be.True(t, strings.HasPrefix(string(got), `&nbsp; &nbsp; <a class="card-link" href="`))
	be.True(t, strings.Contains(string(got), "#runapp"))
	be.True(t, strings.Contains(string(got), "Run app"))
}

func TestLinkPreview(t *testing.T) {
	t.Parallel()

	s := app.LinkPreview("", "", "")
	be.Equal(t, s, "")

	s = app.LinkPreview(1, "readme.txt", "text")
	be.True(t, strings.Contains(string(s), "Preview"))
}

func TestLinkRemote(t *testing.T) {
	t.Parallel()

	s := app.LinkRemote("", "")
	be.True(t, strings.Contains(string(s), "error"))

	s = app.LinkRemote(exampleURL, "")
	be.True(t, strings.Contains(string(s), "error"))

	s = app.LinkRemote(exampleURL, "Example")
	be.True(t, strings.Contains(string(s), exampleURL))
}

func TestLinkRemoteTip(t *testing.T) {
	t.Parallel()

	s := app.LinkRemoteTip("", "", "")
	be.True(t, strings.Contains(string(s), "error"))

	s = app.LinkRemoteTip(exampleURL, "unique", "test tip")
	be.True(t, strings.Contains(string(s), "tooltip"))
	be.True(t, strings.Contains(string(s), "unique"))
	be.True(t, strings.Contains(string(s), exampleURL))
}

func TestLinkScnr(t *testing.T) {
	t.Parallel()

	got, err := app.LinkScnr("")
	be.Err(t, err, nil)
	be.Equal(t, got, "")

	got, err = app.LinkScnr("some scener")
	be.Err(t, err, nil)
	be.Equal(t, got, "/p/some-scener")
}

func TestLinkScnrs(t *testing.T) {
	t.Parallel()

	got := app.LinkScnrs("")
	be.Equal(t, got, "")

	got = app.LinkScnrs("somescener")
	be.True(t, strings.Contains(string(got), "somescener"))
	be.True(t, strings.Contains(string(got), "link-underline"))
}

func TestLinkWiki(t *testing.T) {
	t.Parallel()

	s := app.LinkWiki("", "")
	be.True(t, strings.Contains(string(s), "error"))

	s = app.LinkWiki(exampleWiki, "")
	be.True(t, strings.Contains(string(s), "error"))

	s = app.LinkWiki(exampleWiki, "Example")
	be.True(t, strings.Contains(string(s), exampleWiki))
}

func TestLinkWikiTip(t *testing.T) {
	t.Parallel()

	s := app.LinkWikiTip("", "", "")
	be.True(t, strings.Contains(string(s), "error"))

	s = app.LinkWikiTip("abc/", "unique", "test tip")
	be.True(t, strings.Contains(string(s), "tooltip"))
	be.True(t, strings.Contains(string(s), "unique"))
	be.True(t, strings.Contains(string(s), "abc/"))
}

func TestLogoText(t *testing.T) {
	t.Parallel()

	const leftPad = 6
	const want1 = "      :                             ·· X ··                             ·"
	const want2 = "      :                             ·· XY ··                            ·"
	const want3 = "      :                            ·· XYZ ··                            ·"
	const wantR = "      : ·· I'M MEANT TO BE WRITING AT THIS MOMENT. WHAT I MEAN IS, I ·· ·"

	got := app.LogoText("")
	want := strings.Repeat(" ", leftPad) + app.Welcome
	be.Equal(t, got, want)

	got = app.LogoText("X")
	be.Equal(t, got, want1)

	got = app.LogoText("XY")
	be.Equal(t, got, want2)

	got = app.LogoText("xyz")
	be.Equal(t, got, want3)

	const rand = "I'm meant to be writing at this moment. " +
		"What I mean is, I'm meant to be writing something else at this moment."
	got = app.LogoText(rand)
	be.Equal(t, got, wantR)

	got = app.LogoText("abc")
	be.True(t, strings.Contains(got, "      :                            ·· ABC ··                            ·"))
}

func TestMarkAll(t *testing.T) {
	t.Parallel()

	got := app.MarkAll("", "")
	be.Equal(t, got, "")

	got = app.MarkAll("a", "xyz")
	be.Equal(t, got, "xyz")

	got = app.MarkAll("y", "xyz")
	be.Equal(t, got, `x<mark>y</mark>z`)

	got = app.MarkAll("😊", "x😊z")
	be.Equal(t, got, `x<mark>😊</mark>z`)
}

func TestMonth(t *testing.T) {
	t.Parallel()

	got := app.Month(nil)
	be.Equal(t, got, "")

	got = app.Month(0)
	be.Equal(t, got, "")

	s := app.Month(1)
	be.True(t, strings.Contains(s, "Jan"))

	s = app.Month(13)
	be.Equal(t, s, "")
}

func TestMusicModule(t *testing.T) {
	t.Parallel()

	got := app.MusicModule("")
	be.True(t, !got)

	got = app.MusicModule("zyx")
	be.True(t, !got)

	got = app.MusicModule("mod music")
	be.True(t, got)
}

func TestPrefix(t *testing.T) {
	t.Parallel()

	got := app.Prefix("")
	be.Equal(t, got, "")

	got = app.Prefix("abc")
	be.Equal(t, got, " abc")
}

func TestRecordRels(t *testing.T) {
	t.Parallel()

	got := app.RecordRels(nil, nil)
	be.Equal(t, got, "")

	got = app.RecordRels("A", "B")
	be.Equal(t, got, "A + B")

	got = app.RecordRels("A", nil)
	be.Equal(t, got, "A")

	got = app.RecordRels(nil, "B")
	be.Equal(t, got, "B")
}

func TestSafeBBS(t *testing.T) {
	t.Parallel()

	got := app.SafeBBS(nil)
	be.Equal(t, got, "")

	got = app.SafeBBS(testutil.RTF)
	be.Equal(t, got, "This is some bold text.")

	got = app.SafeBBS("<strong>xyz</strong>")
	be.Equal(t, got, "&lt;strong>xyz&lt;/strong>")
}

func TestSafeDocument(t *testing.T) {
	t.Parallel()

	got := app.SafeDocument(nil)
	be.Equal(t, got, "")

	got = app.SafeDocument(testutil.RTF)
	be.Equal(t, got, "This is some bold text.")

	got = app.SafeDocument("<strong>xyz</strong>")
	be.Equal(t, got, "&lt;strong>xyz&lt;/strong>")
}

func TestRemovePCBoard(t *testing.T) {
	t.Parallel()

	got := app.RemovePCBoard([]byte("This @X0FT@X0AE@X77X@X8BT!"))
	be.Equal(t, got, []byte("This TEXT!"))
}

func TestSafety(t *testing.T) {
	t.Parallel()

	got := app.Safety(nil, nil)
	be.True(t, !got)

	got = app.Safety("placeholder", "")
	be.True(t, got)

	got = app.Safety("", "magazine")
	be.True(t, got)
}

func TestSubTitle(t *testing.T) {
	t.Parallel()

	x := null.StringFrom("")
	s := app.SubTitle(x, nil, false)
	be.Equal(t, s, "")

	const sub = "A second title."
	s = app.SubTitle(x, sub, false)
	be.True(t, strings.Contains(string(s), sub))

	mag := null.StringFrom("magazine")
	s = app.SubTitle(mag, "1", false)
	be.True(t, strings.Contains(string(s), "Issue 1"))

	s = app.SubTitle(mag, "1", true)
	be.True(t, strings.Contains(string(s), "fs-5"))

	s = app.SubTitle(mag, 1, false)
	be.Equal(t, s, "")
}

func TestTagBrief(t *testing.T) {
	t.Parallel()

	got := app.TagBrief("")
	be.Equal(t, got, "")

	got = app.TagBrief(tags.Interview.String())
	be.True(t, strings.Contains(got, "conversations with"))
}

func TestTagOption(t *testing.T) {
	t.Parallel()

	s := app.TagOption(nil, nil)
	be.Equal(t, s, "")

	s = app.TagOption("", tags.Interview.String())
	be.True(t, strings.Contains(string(s), `<option value="interview">`))

	s = app.TagOption(tags.Interview.String(), tags.Interview.String())
	be.True(t, strings.Contains(string(s), `<option value="interview" selected>`))
}

func TestTagWithOS(t *testing.T) {
	t.Parallel()

	got := app.TagWithOS("", "")
	be.True(t, strings.Contains(got, "unknown"))

	got = app.TagWithOS("windows", "")
	be.True(t, strings.Contains(got, "unknown"))

	got = app.TagWithOS("dos", "magazine")
	be.Equal(t, got, "a Dos magazine")
}

func TestTrimSiteSuffix(t *testing.T) {
	t.Parallel()

	got := app.TrimSiteSuffix("Some text")
	be.Equal(t, got, "Some text")

	got = app.TrimSiteSuffix("abc")
	be.Equal(t, got, "abc")

	got = app.TrimSiteSuffix("My super BBS")
	be.Equal(t, got, "My super")
}

func TestTrimSpace(t *testing.T) {
	t.Parallel()

	got := app.TrimSpace(nil)
	be.Equal(t, got, "")

	got = app.TrimSpace("")
	be.Equal(t, got, "")

	got = app.TrimSpace("  ")
	be.Equal(t, got, "")

	got = app.TrimSpace("  a  ")
	be.Equal(t, got, "a")

	x := null.StringFrom("  a  ")
	got = app.TrimSpace(x)
	be.Equal(t, got, "a")
}

func TestURLEncode(t *testing.T) {
	t.Parallel()

	got := app.URLEncode("")
	be.Equal(t, got, "")

	got = app.URLEncode("Some text.txt")
	be.Equal(t, got, "Some+text.txt")
}

func TestWebsiteIcon(t *testing.T) {
	t.Parallel()

	got := app.WebsiteIcon("")
	be.Equal(t, got, "")

	icons := map[string]string{
		"archive.org":  "bank2",
		"reddit.com":   "reddit",
		"youtube.com":  "youtube",
		"defacto2.net": "arrow-right",
	}
	for url, icon := range icons {
		got = app.WebsiteIcon(url)
		be.True(t, strings.Contains(string(got), icon))
	}
}

func TestStripSub(t *testing.T) {
	t.Parallel()

	got, _ := app.StripSup("xyz")
	be.Equal(t, got["text"], "xyz")
	be.Equal(t, got["sup"], "")

	got, _ = app.StripSup("x<sup>y</sup>z")
	be.Equal(t, got["text"], "xz")
	be.Equal(t, got["sup"], "<sup>y</sup>")
}

func TestYMDEdit(t *testing.T) {
	t.Parallel()

	got := app.YMDEdit(nil, nil)
	be.Err(t, got)

	tx := testutil.Tx(t)
	c := testutil.NewContext(t, "")
	got = app.YMDEdit(c, tx)
	be.Err(t, got)

	c = testutil.NewContext(t, "/abc?id=1&year=2000")
	got = app.YMDEdit(c, tx)
	be.Err(t, got, nil)
}

func TestVerify(t *testing.T) {
	t.Parallel()

	sri := app.SRI{}
	err := sri.Verify(emptyFS)
	be.Err(t, err)
}
