package app_test

import (
	"strings"
	"testing"

	"github.com/Defacto2/server/handler/app"
	"github.com/stretchr/testify/assert"
)

const (
	exampleURL  = "https://example.com"
	exampleWiki = "/some/wiki/page"
)

func TestValid(t *testing.T) {
	t.Parallel()
	assert.False(t, app.Valid("not-a-valid-uri"))
	assert.False(t, app.Valid("/files/newest"))
	assert.True(t, app.Valid("newest"))
	assert.True(t, app.Valid("windows-pack"))
	assert.True(t, app.Valid("advert"))
}

func TestMatch(t *testing.T) {
	t.Parallel()
	assert.Equal(t, app.URI(-1), app.Match("not-a-valid-uri"))
	assert.Equal(t, app.URI(35), app.Match("newest"))
	assert.Equal(t, app.URI(57), app.Match("windows-pack"))
	assert.Equal(t, app.URI(1), app.Match("advert"))
}

func TestRecordsSub(t *testing.T) {
	t.Parallel()
	s := app.RecordsSub("")
	assert.Equal(t, "unknown uri", s)

	// range to 57
	for i := 1; i <= int(57); i++ {
		assert.NotEqual(t, "unknown uri", app.URI(i).String())
	}
}

func TestMilestone(t *testing.T) {
	t.Parallel()
	ms := app.Collection()
	assert.Equal(t, 108, ms.Len())
	assert.Equal(t, 108, len(ms))

	one := ms[0]
	assert.Equal(t, 1971, one.Year)
	assert.Equal(t, "Secrets of the Little Blue Box", one.Title)

	for _, record := range ms {
		assert.NotEqual(t, 0, record.Year)
	}
}

func TestInterviewees(t *testing.T) {
	t.Parallel()
	i := app.Interviewees()
	assert.Equal(t, 9, len(i))

	for _, x := range i {
		assert.NotEmpty(t, x.Name)
	}
}

func TestExternalLink(t *testing.T) {
	t.Parallel()
	x := app.LinkRemote("", "")
	assert.Contains(t, x, "error:")
	x = app.LinkRemote(exampleURL, "")
	assert.Contains(t, x, "error:")
	x = app.LinkRemote(exampleURL, "Example")
	assert.Contains(t, x, exampleURL)
}

func TestWikiLink(t *testing.T) {
	t.Parallel()
	x := app.LinkWiki("", "")
	assert.Contains(t, x, "error:")
	x = app.LinkWiki(exampleWiki, "")
	assert.Contains(t, x, "error:")
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
}

func TestSortContent(t *testing.T) {
	tests := []struct {
		content  []string
		expected []string
	}{
		{
			content:  nil,
			expected: nil,
		},
		{
			content: []string{
				"dir1/file1",
				"dir2/file2",
				"dir1/subdir/file3",
				"file4",
			},
			expected: []string{
				"file4",
				"dir1/file1",
				"dir2/file2",
				"dir1/subdir/file3",
			},
		},
		{
			content: []string{
				"dir1/file1",
				"dir1/subdir/file2",
				"dir2/file3",
				"file4",
			},
			expected: []string{
				"file4",
				"dir1/file1",
				"dir2/file3",
				"dir1/subdir/file2",
			},
		},
	}

	for _, tt := range tests {
		t.Run(strings.Join(tt.content, ","), func(t *testing.T) {
			// Make a copy of the original content
			originalContent := make([]string, len(tt.content))
			copy(originalContent, tt.content)

			// Sort the content using the SortContent function
			sortedContent := app.SortContent(tt.content...)

			for i, v := range sortedContent {
				assert.Equal(t, tt.expected[i], v)
			}
		})
	}
}

func TestReadmeSug(t *testing.T) {
	tests := []struct {
		filename string
		group    string
		content  []string
		expected string
	}{
		{
			filename: "file1",
			group:    "group1",
			content: []string{
				"file1.nfo",
				"file1.txt",
				"file1.unp",
				"file1.doc",
			},
			expected: "file1.nfo",
		},
		{
			filename: "file2",
			group:    "group2",
			content: []string{
				"file.diz",
				"file.asc",
				"file.1st",
				"group2.dox",
			},
			expected: "group2.dox",
		},
		{
			filename: "file3",
			group:    "group3",
			content: []string{
				"file3.nfo",
				"file.txt",
				"file30.unp",
				"file3x.doc",
				"filex3.diz",
				"file3.asc",
				"file3.1st",
				"file3.dox",
			},
			expected: "file3.nfo",
		},
		{
			filename: "file4",
			group:    "group4",
			content:  []string{},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.filename+"_"+tt.group, func(t *testing.T) {
			result := app.ReadmeSug(tt.filename, tt.group, tt.content...)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestList(t *testing.T) {
	t.Parallel()
	list := app.List()
	assert.Len(t, list, 9)
}

func TestAsset(t *testing.T) {
	t.Parallel()

	x, y := app.Bootstrap, app.Uploader
	assert.Equal(t, app.Asset(0), x)
	assert.Equal(t, app.Asset(14), y)

	hrefs := app.Hrefs()
	for i, href := range hrefs {
		assert.NotEmpty(t, href)
		switch i {
		case 0, 9:
			ext := href[len(href)-8:]
			assert.Equal(t, ".min.css", ext)
		case 6, 7, 8:
		default:
			ext := href[len(href)-7:]
			assert.Equal(t, ".min.js", ext)
		}
	}
}

func TestNames(t *testing.T) {
	t.Parallel()

	x := app.Names()
	assert.Equal(t, "public/css/bootstrap.min.css", x[0])
}

func TestFontRefs(t *testing.T) {
	t.Parallel()

	x := app.FontRefs()
	assert.Equal(t, "/pxplus_ibm_vga8.woff2", x[app.VGA8])

	n := app.FontNames()
	assert.Equal(t, "public/font/pxplus_ibm_vga8.woff2", n[app.VGA8])
}

func TestGlobTo(t *testing.T) {
	t.Parallel()

	x := app.GlobTo("file.css")
	assert.Equal(t, "view/app/file.css", x)
}

func TestTemplates(t *testing.T) {
	t.Parallel()

	w := app.Web{}
	_, err := w.Templates()
	assert.Error(t, err)
}
