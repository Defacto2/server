package simple_test

import (
	"fmt"
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

func BenchmarkCleanHTML(b *testing.B) {
	html := `<div class="content">
		<p class="lead">This is a <strong>test</strong> with <a href="https://example.com" class="link" id="test">links</a> and <span style="color: red;">formatting</span>.</p>
		<p>Another paragraph with &nbsp; non-breaking &amp; spaces and <data-info="test">data attributes</data-info>.</p>
	</div>`

	b.Run("", func(b *testing.B) {
		for range b.N {
			simple.CleanHTML(html)
		}
	})
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

func TestStripHTMLTags(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Removes all HTML tags",
			input:    `<p>Test <strong>content</strong> here</p>`,
			expected: `Test content here`,
		},
		{
			name:     "Handles HTML entities",
			input:    `Test &nbsp; with &amp; entities`,
			expected: `Test with & entities`,
		},
		{
			name:     "Adds spacing after punctuation",
			input:    `Test.content.with.punctuation!marks?here`,
			expected: `Test. content. with. punctuation! marks? here`,
		},
		{
			name:     "Collapses multiple spaces",
			input:    `Test    multiple     spaces`,
			expected: `Test multiple spaces`,
		},
		{
			name:     "Handles anchor tags",
			input:    `Test <a href="https://example.com">link</a> content`,
			expected: `Test link content`,
		},
		{
			name:     "Handles complex HTML",
			input:    `<div><p>Test <span>content</span> <a href="#">link</a></p></div>`,
			expected: `Test content link`,
		},
		{
			name:     "Removes spaces before punctuation",
			input:    `Test content with spaces , before commas . and periods !`,
			expected: `Test content with spaces, before commas. and periods!`,
		},
		{
			name:     "Removes spaces around parentheses",
			input:    `Test ( content ) with ( spaces ) around ( parentheses )`,
			expected: `Test (content) with (spaces) around (parentheses)`,
		},
		{
			name:     "Handles real milestone content",
			input:    `<p>Ron Rosenbaum writes the first mainstream article on phone freaks, primarily kids who'd hack and experiment with the global telephone network.</p><p>The piece coins them as phone-freaks (<strong>phreaks</strong>) and introduces the reader to the kids' use of <strong>pseudonyms</strong> or codenames within their cliques and <strong>groups</strong> of friends. It gives an early example of <strong>social engineering</strong>, defines the community of phreakers as the phone-phreak <strong>underground</strong>, and mentions the newer trend of <strong>computer phreaking</strong>, which we call <u>computer&nbsp;hacking</u> today.</p>`,
			expected: `Ron Rosenbaum writes the first mainstream article on phone freaks, primarily kids who'd hack and experiment with the global telephone network. The piece coins them as phone-freaks (phreaks) and introduces the reader to the kids' use of pseudonyms or codenames within their cliques and groups of friends. It gives an early example of social engineering, defines the community of phreakers as the phone-phreak underground, and mentions the newer trend of computer phreaking, which we call computer hacking today.`,
		},
		{
			name:     "Converts <q> tags to quotes",
			input:    `<p>He said <q>Hello world</q> to everyone.</p>`,
			expected: `He said "Hello world" to everyone.`,
		},
		{
			name:     "Handles multiple <q> tags",
			input:    `<p>Multiple <q>quotes</q> in <q>one</q> sentence.</p>`,
			expected: `Multiple "quotes" in "one" sentence.`,
		},
		{
			name:     "Handles nested <q> tags",
			input:    `<p>Nested <q>quotes <q>inside</q> quotes</q> should work.</p>`,
			expected: `Nested "quotes inside" quotes should work.`,
		},
		{
			name:     "Handles <q> tags with attributes",
			input:    `<p>The famous quote <q cite="https://example.com">To be or not to be</q> is from Shakespeare.</p>`,
			expected: `The famous quote "To be or not to be" is from Shakespeare.`,
		},
		{
			name:     "Complex real-world example with quotes",
			input:    `<p>As Steve Jobs famously said <q>Stay hungry, stay foolish</q>, which was inspired by the <q>Whole Earth Catalog</q> manifesto that stated <q>Stay hungry. Stay foolish.</q> This philosophy became a cornerstone of Apple's <q cite="https://apple.com">Think Different</q> campaign.</p>`,
			expected: `As Steve Jobs famously said "Stay hungry, stay foolish", which was inspired by the "Whole Earth Catalog" manifesto that stated "Stay hungry. Stay foolish." This philosophy became a cornerstone of Apple's "Think Different" campaign.`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := simple.CleanHTML(tt.input)
			if result != tt.expected {
				t.Errorf("CleanHTMLTags(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
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

func TestHash(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Empty string",
			input:    "",
			expected: "y_Kc5IQiIyU=",
		},
		{
			name:     "Simple string",
			input:    "test",
			expected: "-ebm7xl8KyU=",
		},
		{
			name:     "Case sensitivity",
			input:    "Test",
			expected: "JHTn-xrsnwU=",
		},
		{
			name:     "Special characters",
			input:    "hello@world.com",
			expected: "VLO_NWka7Yg=",
		},
		{
			name:     "Unicode characters",
			input:    "こんにちは",
			expected: "NQtCHdj8ka0=",
		},
		{
			name:     "Long string",
			input:    "This is a longer test string to verify the hash function works with more substantial input",
			expected: "B1QZc8sq1O4=",
		},
		{
			name:     "Consistency check",
			input:    "consistency",
			expected: "n-09Zb9aUwc=",
		},
		{
			name:     "URL-safe characters",
			input:    "user@example.com",
			expected: "uBab6YHzyts=",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := simple.Hash(tt.input)
			if result != tt.expected {
				t.Errorf("Hash(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestHashDeterministic(t *testing.T) {
	t.Parallel()
	// Test that the same input always produces the same output
	input := "deterministic-test"
	hash1 := simple.Hash(input)
	hash2 := simple.Hash(input)
	be.Equal(t, hash1, hash2)

	// Test different inputs produce different outputs
	input2 := "deterministic-test-2"
	hash3 := simple.Hash(input2)
	be.True(t, hash1 != hash3)
}

func TestHashProperties(t *testing.T) {
	t.Parallel()
	// Test that hash output has consistent length
	hash1 := simple.Hash("short")
	hash2 := simple.Hash("this is a much longer input string for testing")

	// Both should be base64 encoded FNV-64a hashes (12 characters)
	be.Equal(t, 12, len(hash1))
	be.Equal(t, 12, len(hash2))

	// Test that hash only contains URL-safe base64 characters
	validChars := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_="
	for _, char := range hash1 {
		be.True(t, strings.ContainsRune(validChars, char))
	}
	for _, char := range hash2 {
		be.True(t, strings.ContainsRune(validChars, char))
	}
}

func BenchmarkHash(b *testing.B) {
	testStrings := []string{
		"short",
		"medium length string for benchmarking",
		"This is a longer string that would be more typical of real-world usage in the application for generating stable identifiers",
		strings.Repeat("a", 100), // 100 character string
	}

	for _, str := range testStrings {
		b.Run(fmt.Sprintf("length-%d", len(str)), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = simple.Hash(str)
			}
		})
	}
}
