package app_test

import (
	"database/sql"
	"embed"
	"maps"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/Defacto2/server/handler/app"
	"github.com/Defacto2/server/internal/tags"
	"github.com/aarondl/null/v8"
	"github.com/nalgeon/be"
)

//go:embed*
var emptyFS embed.FS //nolint:gochecknoglobals

const (
	exampleURL  = "https://example.com"
	exampleWiki = "/some/wiki/page"
)

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

func TestTagOption(t *testing.T) {
	t.Parallel()

	s := app.TagOption(nil, nil)
	be.True(t, strings.Contains(string(s), `<option value="">`))

	s = app.TagOption("", tags.Interview.String())
	be.True(t, strings.Contains(string(s), `<option value="interview">`))

	s = app.TagOption(tags.Interview.String(), tags.Interview.String())
	be.True(t, strings.Contains(string(s), `<option value="interview" selected>`))
}

func TestTagBrief(t *testing.T) {
	t.Parallel()

	s := app.TagBrief("")
	be.Equal(t, s, "")

	s = app.TagBrief(tags.Interview.String())
	be.True(t, strings.Contains(s, "conversations with"))
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

	s = app.SubTitle(mag, 1, false)
	be.Equal(t, s, "")
}

func TestRecordRels(t *testing.T) {
	t.Parallel()

	s := app.RecordRels(nil, nil)
	be.Equal(t, s, "")

	s = app.RecordRels("1", "2")
	be.Equal(t, "1 + 2", s)
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
	be.True(t, strings.Contains(s, "error"))
}

func TestLinkRelsPerf(t *testing.T) {
	t.Parallel()

	s := app.LinkRelsPerf("", "")
	be.Equal(t, s, "")

	s = app.LinkRelsPerf("Group 1", "Group 2")
	be.True(t, strings.Contains(string(s), "Group 1"))
	be.True(t, strings.Contains(string(s), "Group 2"))
	be.True(t, strings.Contains(string(s), `href="/g/group-1"`))
	be.True(t, strings.Contains(string(s), `href="/g/group-2"`))
}

func TestLastUpdated(t *testing.T) {
	t.Parallel()

	got := app.LastUpdated(nil)
	be.Equal(t, got, "")

	oneHourAgo := time.Now().Add(-time.Hour)
	got = app.LastUpdated(oneHourAgo)
	be.Equal(t, got, "Last updated about 1 hour ago")
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

func TestDay(t *testing.T) {
	t.Parallel()

	x := app.Day("")
	be.True(t, strings.Contains(x, "error"))

	x = app.Day("1")
	be.True(t, strings.Contains(x, "error"))

	x = app.Day(1)
	be.True(t, strings.Contains(x, " 1"))

	x = app.Day(100)
	be.True(t, strings.Contains(x, "error"))
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
	s := app.ByteFileS("", "", "")
	be.True(t, strings.Contains(string(s), "error"))

	s = app.ByteFileS(intro, 1, "")
	be.True(t, strings.Contains(string(s), "error"))

	s = app.ByteFileS(intro, "", 1)
	be.True(t, strings.Contains(string(s), "error"))

	s = app.ByteFileS(intro, 1, 50000)
	be.True(t, strings.Contains(string(s), "1 intro "))
	be.True(t, strings.Contains(string(s), "(49k)"))

	s = app.ByteFileS(intro, 12, 1023)
	be.True(t, strings.Contains(string(s), "12 intros "))
	be.True(t, strings.Contains(string(s), "(1023B)"))
}

func TestFuncMap(t *testing.T) {
	t.Parallel()

	w := app.Templ{}
	db := &sql.DB{}
	x := w.FuncMap(t.Context(), db)
	keys := slices.Sorted(maps.Keys(*x))
	be.True(t, slices.Contains(keys, "add"))
	be.True(t, slices.Contains(keys, "version"))
	be.True(t, slices.Contains(keys, "az"))
	be.True(t, slices.Contains(keys, "msdos"))
}

func TestLinkSamples(t *testing.T) {
	t.Parallel()

	x := app.LinkPreviews("1", "2", "3", "4", "5", "6", "7")
	be.True(t, len(x) == 7)
	be.True(t, strings.Contains(x[0], "youtube.com/watch?v=1"))
	be.True(t, strings.Contains(x[1], "demozoo.org/productions/2"))
}

func TestMilestone(t *testing.T) {
	t.Parallel()

	ms := app.Collection()

	const expectedMileStones = 100
	be.True(t, ms.Len() > expectedMileStones)

	one := ms[0]
	const expectedYear = 1971
	be.Equal(t, expectedYear, one.Year)
	be.Equal(t, "Secrets of the Little Blue Box", one.Title)

	for _, record := range ms {
		be.True(t, record.Year != 0)
	}
}

func TestInterviewees(t *testing.T) {
	t.Parallel()
	i := app.Interviewees()
	l := len(i)
	const expected = 12
	be.Equal(t, l, expected)

	for _, x := range i {
		be.True(t, x.Name != "")
	}
}

func TestExternalLink(t *testing.T) {
	t.Parallel()

	s := app.LinkRemote("", "")
	be.True(t, strings.Contains(string(s), "error"))

	s = app.LinkRemote(exampleURL, "")
	be.True(t, strings.Contains(string(s), "error"))

	s = app.LinkRemote(exampleURL, "Example")
	be.True(t, strings.Contains(string(s), exampleURL))
}

func TestWikiLink(t *testing.T) {
	t.Parallel()

	s := app.LinkWiki("", "")
	be.True(t, strings.Contains(string(s), "error"))

	s = app.LinkWiki(exampleWiki, "")
	be.True(t, strings.Contains(string(s), "error"))

	s = app.LinkWiki(exampleWiki, "Example")
	be.True(t, strings.Contains(string(s), exampleWiki))
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

func TestList(t *testing.T) {
	t.Parallel()

	list := app.List()
	const expectedCount = 9
	be.True(t, len(list) == expectedCount)
}

func TestNames(t *testing.T) {
	t.Parallel()

	x := *app.Names()
	be.Equal(t, "public/css/bootstrap.min.css", x[0])
}

func TestGlobTo(t *testing.T) {
	t.Parallel()

	x := app.GlobTo("file.css")
	be.Equal(t, "view/app/file.css", x)
}

func TestTemplates(t *testing.T) {
	t.Parallel()

	w := app.Templ{}
	_, err := w.Templates(t.Context(), nil)
	be.Err(t, err)
}

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

func TestLinkDownload(t *testing.T) {
	t.Parallel()

	s := string(app.LinkDownload("", ""))
	be.True(t, strings.Contains(s, "invalid"))

	s = string(app.LinkDownload(1, ""))
	be.True(t, strings.Contains(s, "/d/9b1c6"))
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

func TestLinkPreview(t *testing.T) {
	t.Parallel()

	s := app.LinkPreview("", "", "")
	be.Equal(t, s, "")

	s = app.LinkPreview(1, "readme.txt", "text")
	be.True(t, strings.Contains(string(s), "Preview"))
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

func TestURLEncode(t *testing.T) {
	t.Parallel()

	got := app.URLEncode("")
	be.Equal(t, got, "")

	got = app.URLEncode("Some text.txt")
	be.Equal(t, got, "Some+text.txt")
}

func TestYMDEdit(t *testing.T) {
	t.Parallel()

	s := app.YMDEdit(t.Context(), nil, nil)
	be.Err(t, s)
}

func TestVerify(t *testing.T) {
	t.Parallel()

	sri := app.SRI{}
	err := sri.Verify(emptyFS)
	be.Err(t, err)
}
