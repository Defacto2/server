package app_test

import (
	"embed"
	"maps"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/Defacto2/server/handler/app"
	"github.com/Defacto2/server/internal/tags"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/volatiletech/null/v8"
)

//go:embed*
var emptyFS embed.FS //nolint:gochecknoglobals

const (
	exampleURL  = "https://example.com"
	exampleWiki = "/some/wiki/page"
)

func TestTrimSpace(t *testing.T) {
	t.Parallel()
	s := app.TrimSpace(nil)
	assert.Empty(t, s)
	s = app.TrimSpace("")
	assert.Empty(t, s)
	s = app.TrimSpace("  ")
	assert.Empty(t, s)
	s = app.TrimSpace("  a  ")
	assert.Equal(t, "a", s)
	x := null.StringFrom("  a  ")
	s = app.TrimSpace(x)
	assert.Equal(t, "a", s)
}

func TestTagOption(t *testing.T) {
	t.Parallel()
	s := app.TagOption(nil, nil)
	assert.Contains(t, `<option value="">`, s)
	s = app.TagOption("", tags.Interview.String())
	assert.Contains(t, s, `<option value="interview">`)
	s = app.TagOption(tags.Interview.String(), tags.Interview.String())
	assert.Contains(t, s, `<option value="interview" selected>`)
}

func TestTagBrief(t *testing.T) {
	t.Parallel()
	s := app.TagBrief("")
	assert.Empty(t, s)
	s = app.TagBrief(tags.Interview.String())
	assert.Contains(t, s, "conversations with")
}

func TestSubTitle(t *testing.T) {
	t.Parallel()
	x := null.StringFrom("")
	s := app.SubTitle(x, nil, false)
	assert.Empty(t, s)

	const sub = "A second title."
	s = app.SubTitle(x, sub, false)
	assert.Contains(t, s, sub)

	mag := null.StringFrom("magazine")
	s = app.SubTitle(mag, "1", false)
	assert.Contains(t, s, "Issue 1")

	s = app.SubTitle(mag, 1, false)
	assert.Empty(t, s)
}

func TestRecordRels(t *testing.T) {
	t.Parallel()
	s := app.RecordRels(nil, nil)
	assert.Empty(t, s)
	s = app.RecordRels("1", "2")
	assert.Equal(t, "1 + 2", s)
}

func TestMonth(t *testing.T) {
	t.Parallel()
	s := app.Month(nil)
	assert.Empty(t, s)
	s = app.Month(0)
	assert.Empty(t, s)
	s = app.Month(1)
	assert.Contains(t, s, "Jan")
	s = app.Month(13)
	assert.Contains(t, s, "error")
}

func TestLinkRelsPerf(t *testing.T) {
	t.Parallel()
	s := app.LinkRelsPerf("", "")
	assert.Empty(t, s)
	s = app.LinkRelsPerf("Group 1", "Group 2")
	assert.Contains(t, s, "Group 1")
	assert.Contains(t, s, "Group 2")
	assert.Contains(t, s, `href="/g/group-1"`)
	assert.Contains(t, s, `href="/g/group-2"`)
}

func TestLastUpdated(t *testing.T) {
	t.Parallel()
	s := app.LastUpdated(nil)
	assert.Empty(t, s)
	oneHourAgo := time.Now().Add(-time.Hour)
	s = app.LastUpdated(oneHourAgo)
	assert.Equal(t, "Last updated about 1 hour ago", s)
}

func TestLinkHref(t *testing.T) {
	t.Parallel()
	s, err := app.LinkHref(nil)
	assert.Empty(t, s)
	require.Error(t, err)

	s, err = app.LinkHref(0)
	assert.Empty(t, s)
	require.Error(t, err)

	s, err = app.LinkHref(1)
	assert.Contains(t, s, "/f/9b1c6")
	require.NoError(t, err)
}

func TestDescribe(t *testing.T) {
	t.Parallel()
	s := app.Describe("", "", "", "")
	assert.Contains(t, s, "error")
	s = app.Describe("", "", 1900, 50)
	assert.Contains(t, s, "unknown release")
	s = app.Describe("x", "y", 1980, 1)
	assert.Contains(t, s, "Unknown platform")
	assert.Contains(t, s, "Jan, 1980")
	plat := null.StringFrom(tags.ANSI.String())
	s = app.Describe(plat, "y", 1980, 1)
	assert.Contains(t, s, "Unknown section")
	sect := null.StringFrom(tags.BBS.String())
	year := null.Int16From(1990)
	month := null.Int16From(12)
	s = app.Describe(plat, sect, year, month)
	assert.Contains(t, s, "BBS ansi advert published in")
	assert.Contains(t, s, "Dec, 1990")
}

func TestDay(t *testing.T) {
	t.Parallel()
	x := app.Day("")
	assert.Contains(t, x, "error")
	x = app.Day("1")
	assert.Contains(t, x, "error")
	x = app.Day(1)
	assert.Contains(t, x, " 1")
	x = app.Day(100)
	assert.Contains(t, x, "error")
}

func TestByteFile(t *testing.T) {
	t.Parallel()
	s := app.ByteFile("", "")
	assert.Contains(t, s, "error")
	s = app.ByteFile(1, "")
	assert.Contains(t, s, "error")
	s = app.ByteFile("", 1)
	assert.Contains(t, s, "error")
	s = app.ByteFile(12, 1023)
	assert.Contains(t, s, "12 ")
	assert.Contains(t, s, "(1023B)")
}

func TestByteFileS(t *testing.T) {
	t.Parallel()
	const intro = "intro"
	s := app.ByteFileS("", "", "")
	assert.Contains(t, s, "error")
	s = app.ByteFileS(intro, 1, "")
	assert.Contains(t, s, "error")
	s = app.ByteFileS(intro, "", 1)
	assert.Contains(t, s, "error")
	s = app.ByteFileS(intro, 1, 50000)
	assert.Contains(t, s, "1 intro ")
	assert.Contains(t, s, "(49k)")
	s = app.ByteFileS(intro, 12, 1023)
	assert.Contains(t, s, "12 intros ")
	assert.Contains(t, s, "(1023B)")
}

func TestFuncMap(t *testing.T) {
	t.Parallel()
	w := app.Templ{}
	x := w.FuncMap(nil)
	keys := slices.Sorted(maps.Keys(*x))
	assert.Contains(t, keys, "add")
	assert.Contains(t, keys, "version")
	assert.Contains(t, keys, "az")
	assert.Contains(t, keys, "msdos")
}

func TestLinkSamples(t *testing.T) {
	t.Parallel()
	x := app.LinkPreviews("1", "2", "3", "4", "5", "6", "7")
	assert.Len(t, x, 7)
	assert.Contains(t, x[0], "youtube.com/watch?v=1")
	assert.Contains(t, x[1], "demozoo.org/productions/2")
}

func TestMilestone(t *testing.T) {
	t.Parallel()
	ms := app.Collection()

	const expectedMileStones = 100
	assert.Greater(t, ms.Len(), expectedMileStones)

	one := ms[0]
	const expectedYear = 1971
	assert.Equal(t, expectedYear, one.Year)
	assert.Equal(t, "Secrets of the Little Blue Box", one.Title)

	for _, record := range ms {
		assert.NotEqual(t, 0, record.Year)
	}
}

func TestInterviewees(t *testing.T) {
	t.Parallel()
	i := app.Interviewees()
	l := len(i)
	const expectedInterviewees = 11
	assert.Equal(t, expectedInterviewees, l)

	for _, x := range i {
		assert.NotEmpty(t, x.Name)
	}
}

func TestExternalLink(t *testing.T) {
	t.Parallel()
	x := app.LinkRemote("", "")
	assert.Contains(t, x, "error")
	x = app.LinkRemote(exampleURL, "")
	assert.Contains(t, x, "error")
	x = app.LinkRemote(exampleURL, "Example")
	assert.Contains(t, x, exampleURL)
}

func TestWikiLink(t *testing.T) {
	t.Parallel()
	x := app.LinkWiki("", "")
	assert.Contains(t, x, "error")
	x = app.LinkWiki(exampleWiki, "")
	assert.Contains(t, x, "error")
	x = app.LinkWiki(exampleWiki, "Example")
	assert.Contains(t, x, exampleWiki)
}

func TestLogoText(t *testing.T) {
	t.Parallel()
	const leftPad = 6
	const want1 = "      :                             ·· X ··                             ·"
	const want2 = "      :                             ·· XY ··                            ·"
	const want3 = "      :                            ·· XYZ ··                            ·"
	const wantR = "      : ·· I'M MEANT TO BE WRITING AT THIS MOMENT. WHAT I MEAN IS, I ·· ·"
	x := app.LogoText("")
	want := strings.Repeat(" ", leftPad) + app.Welcome
	assert.Equal(t, want, x)
	x = app.LogoText("X")
	assert.Equal(t, want1, x)
	x = app.LogoText("XY")
	assert.Equal(t, want2, x)
	x = app.LogoText("xyz")
	assert.Equal(t, want3, x)
	const rand = "I'm meant to be writing at this moment. " +
		"What I mean is, I'm meant to be writing something else at this moment."
	x = app.LogoText(rand)
	assert.Equal(t, wantR, x)
	x = app.LogoText("abc")
	assert.Contains(t, x, "      :                            ·· ABC ··                            ·")
}

func TestList(t *testing.T) {
	t.Parallel()
	list := app.List()
	const expectedCount = 9
	assert.Len(t, list, expectedCount)
}

func TestNames(t *testing.T) {
	t.Parallel()

	x := *app.Names()
	assert.Equal(t, "public/css/bootstrap.min.css", x[0])
}

func TestFontRefs(t *testing.T) {
	t.Parallel()

	x := *app.FontRefs()
	assert.Equal(t, "/pxplus_ibm_vga8.woff2", x[app.VGA8])

	n := *app.FontNames()
	assert.Equal(t, "public/font/pxplus_ibm_vga8.woff2", n[app.VGA8])
}

func TestGlobTo(t *testing.T) {
	t.Parallel()

	x := app.GlobTo("file.css")
	assert.Equal(t, "view/app/file.css", x)
}

func TestTemplates(t *testing.T) {
	t.Parallel()

	w := app.Templ{}
	_, err := w.Templates(nil)
	require.Error(t, err)
}

func TestAttribute(t *testing.T) {
	t.Parallel()
	s := app.Attribute("", "", "", "", "")
	assert.Empty(t, s)
	s = app.Attribute("writer1",
		"", "", "", "")
	assert.Empty(t, s)
	s = app.Attribute("writer",
		"", "", "", "some scener")
	assert.Contains(t, s, "error:")
	s = app.Attribute("another person,writer,some scener",
		"", "", "", "some scener")
	require.Equal(t, "Writer attribution", s)
	s = app.Attribute("another person,writer,some scener",
		"", "some scener", "", "some scener")
	assert.Equal(t, "Writer and artist attributions", s)
	s = app.Attribute("another person,writer,ben",
		"ben", "", "", "ben")
	assert.Equal(t, "Writer and programmer attributions", s)
}

func TestBrief(t *testing.T) {
	t.Parallel()
	s := app.Brief("", "")
	assert.Equal(t, "an unknown release", s)

	s = app.Brief("a string", "")
	assert.Contains(t, s, "unknown platform")
	plat := null.StringFrom("windows")

	s = app.Brief(plat, "")
	assert.Contains(t, s, "unknown section")

	sect := null.StringFrom(tags.Intro.String())
	s = app.Brief(plat, sect)
	assert.Contains(t, s, "a Windows intro")
}

func TestLinkDownload(t *testing.T) {
	t.Parallel()
	s := app.LinkDownload("", "")
	assert.Contains(t, s, "invalid")
	s = app.LinkDownload(1, "")
	assert.Contains(t, s, "/d/9b1c6")
}

func TestLinkInterview(t *testing.T) {
	t.Parallel()
	s := app.LinkInterview("")
	assert.Contains(t, s, "error")
	s = app.LinkInterview("x")
	assert.Empty(t, s)
	s = app.LinkInterview("https://example.com")
	assert.Contains(t, s, "#arrow-right")
}

func TestLinkPreview(t *testing.T) {
	t.Parallel()
	s := app.LinkPreview("", "", "")
	assert.Empty(t, s)
	s = app.LinkPreview(1, "readme.txt", "text")
	assert.Contains(t, s, "Preview")
}

func TestLinkScnr(t *testing.T) {
	t.Parallel()
	s, err := app.LinkScnr("")
	require.NoError(t, err)
	assert.Empty(t, s)
	s, err = app.LinkScnr("some scener")
	require.NoError(t, err)
	assert.Equal(t, "/p/some-scener", s)
}

func TestTagWithOS(t *testing.T) {
	t.Parallel()
	s := app.TagWithOS("", "")
	assert.Contains(t, s, "unknown")
	s = app.TagWithOS("windows", "")
	assert.Contains(t, s, "unknown")
	s = app.TagWithOS("dos", "magazine")
	assert.Equal(t, "a Dos magazine", s)
}

func TestTrimSiteSuffix(t *testing.T) {
	t.Parallel()
	s := app.TrimSiteSuffix("Some text")
	assert.Equal(t, "Some text", s)
	s = app.TrimSiteSuffix("abc")
	assert.Equal(t, "abc", s)
	s = app.TrimSiteSuffix("My super BBS")
	assert.Equal(t, "My super", s)
}

func TestURLEncode(t *testing.T) {
	t.Parallel()
	s := app.URLEncode("")
	assert.Empty(t, s)
	s = app.URLEncode("Some text.txt")
	assert.Equal(t, "Some+text.txt", s)
}

func TestYMDEdit(t *testing.T) {
	t.Parallel()
	s := app.YMDEdit(nil, nil)
	require.Error(t, s)
}

func TestVerify(t *testing.T) {
	t.Parallel()
	sri := app.SRI{}
	err := sri.Verify(emptyFS)
	require.Error(t, err)
}
