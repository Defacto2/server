package simple_test

import (
	"encoding/binary"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/Defacto2/magicnumber"
	"github.com/Defacto2/server/handler/internal/simple"
	"github.com/Defacto2/server/internal/config"
	"github.com/Defacto2/server/internal/dir"
	"github.com/Defacto2/server/internal/logs"
	"github.com/Defacto2/server/internal/testutil"
	"github.com/aarondl/null/v8"
	"github.com/nalgeon/be"
)

var testLog = logs.Discard()

func testpng(t *testing.T) string {
	t.Helper()

	return filepath.Join(testutil.Testdata(), "thumb", "TEST.png")
}

func TestAssetSrc(t *testing.T) {
	t.Parallel()

	s := simple.AssetSrc("", "", "", "")
	be.True(t, strings.Contains(s, "no such file"))
	_, file, _, ok := runtime.Caller(0)
	be.True(t, ok)

	s = simple.AssetSrc("", file, "", "")
	be.True(t, strings.Contains(s, "sha384-"))

	d := filepath.Dir(testpng(t))
	prev := simple.AssetSrc(config.AbsPreview, d, "TEST", ".PNG")
	be.True(t, strings.HasPrefix(prev, config.StaticOriginal()+"/TEST.png?sha384-"))

	thumb := simple.AssetSrc(config.AbsThumbnail, d, "TEST", ".png")
	be.True(t, strings.HasPrefix(thumb, config.StaticThumb()+"/TEST.png?sha384-"))
}

func TestBytesHuman(t *testing.T) {
	t.Parallel()

	be.Equal(t, simple.BytesHuman(0), "(n/a)")
	be.Equal(t, simple.BytesHuman(1000), "1.0 kB")
	be.Equal(t, simple.BytesHuman(-1000), "1.0 kB")
}

func TestCleanFname(t *testing.T) {
	t.Parallel()

	got, err := simple.CleanFname("")
	be.Err(t, err, nil)
	be.Equal(t, got, "")

	got, err = simple.CleanFname("readme.nfo")
	be.Err(t, err, nil)
	be.Equal(t, got, "readme.nfo")

	got, err = simple.CleanFname(`<script>"x"</script>`)
	be.Err(t, err, nil)
	be.Equal(t, got, `&lt;script&gt;&#34;x&#34;&lt;/script&gt;`)
}

func TestStripHTMLTags(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Empty string",
			input:    "",
			expected: "",
		},
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
			t.Parallel()

			result := simple.CleanHTML(tt.input)
			if result != tt.expected {
				t.Errorf("CleanHTMLTags(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestDownloadB(t *testing.T) {
	t.Parallel()

	x := simple.DownloadInBytes("")
	be.True(t, strings.Contains(string(x), "received an invalid type"))

	x = simple.DownloadInBytes("a string")
	be.True(t, strings.Contains(string(x), "received an invalid type"))

	x = simple.DownloadInBytes("1")
	be.True(t, strings.Contains(string(x), "received an invalid type"))

	x = simple.DownloadInBytes(null.Int64{})
	be.True(t, strings.Contains(string(x), "(n/a)"))

	x = simple.DownloadInBytes(null.Int64From(1))
	be.True(t, strings.Contains(string(x), "1B"))

	for _, v := range []any{
		int(1024), int64(1024), int32(1024), int16(1024), int8(10),
		uint(1024), uint64(1024), uint32(1024), uint16(1024), uint8(10),
		uint(math.MaxUint), uint64(math.MaxUint64),
	} {
		got := string(simple.DownloadInBytes(v))
		be.True(t, strings.Contains(got, "text-body-secondary"))
	}
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

	be.Equal(t, simple.LinkPreviewTip("", ""), "")
	be.Equal(t, simple.LinkPreviewTip("file.zip", "windows"), "")
	be.Equal(t, simple.LinkPreviewTip("index.htm", "markup"), "Read this as HTML")
	be.Equal(t, simple.LinkPreviewTip("file.nfo", "text"), "Read this as text")
	be.Equal(t, simple.LinkPreviewTip("file.nfo", "textamiga"), "Read this as text")
	be.Equal(t, simple.LinkPreviewTip("readme.txt", "windows"), "Read this as text")
	be.Equal(t, simple.LinkPreviewTip("pic.png", "image"), "View this as an image or photo")
	be.Equal(t, simple.LinkPreviewTip("song.mp3", "audio"), "Play this as media")
	be.Equal(t, simple.LinkPreviewTip("prog.bin", "dos"), "")
}

func TestReleaserPair(t *testing.T) {
	t.Parallel()

	be.Equal(t, simple.ReleaserPair(nil, nil), [2]string{})

	s := simple.ReleaserPair("1", "2")
	be.Equal(t, s[0], "1")
	be.Equal(t, s[1], "2")

	s = simple.ReleaserPair(nil, "2")
	be.Equal(t, s[0], "2")
	be.Equal(t, s[1], "")

	s = simple.ReleaserPair("1", nil)
	be.Equal(t, s[0], "1")
	be.Equal(t, s[1], "")

	s = simple.ReleaserPair(null.StringFrom("alpha"), null.StringFrom("beta"))
	be.Equal(t, s, [2]string{"alpha", "beta"})
	be.Equal(t, simple.ReleaserPair(null.String{}, null.String{}), [2]string{})
}

func TestUpdated(t *testing.T) {
	t.Parallel()

	be.Equal(t, simple.Updated(nil, ""), "")
	be.Equal(t, simple.Updated(null.Time{}, ""), "")
	be.True(t, strings.Contains(simple.Updated("9:30pm", ""), "error"))
	be.True(t, strings.Contains(simple.Updated(time.Now(), ""), "Time just now"))
	be.True(t, strings.Contains(simple.Updated(time.Now().Add(-2*time.Hour), "Updated"), "Updated "))
	be.True(t, strings.Contains(simple.Updated(null.TimeFrom(time.Now()), "Saved"), "Saved just now"))
	be.True(t, strings.Contains(simple.Updated(null.TimeFrom(time.Now().Add(-2*time.Hour)), ""), " ago"))
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
	be.Equal(t, simple.DemozooGetLink(fn, fs, dz, null.String{String: "ignored", Valid: false}), "")
	be.Equal(t, simple.DemozooGetLink(fn, fs, dz, null.StringFrom("")), "")

	un = null.StringFrom("user")
	html = simple.DemozooGetLink(fn, fs, dz, un)
	be.True(t, html != "")
}

func TestImageSample(t *testing.T) {
	t.Parallel()

	const missing = "No preview"
	x := simple.ImageSample("", "")
	be.True(t, strings.Contains(string(x), missing))

	// note: the filename extension is case-sensitive.
	preview := dir.Directory(testpng(t))
	x = simple.ImageSample("", preview)
	be.True(t, strings.Contains(string(x), missing))

	const filenameNoExt = "TEST"
	abs := dir.Directory(filepath.Dir(testpng(t)))
	x = simple.ImageSample(filenameNoExt, abs)
	be.True(t, strings.Contains(string(x), "sha384-SK3qCpS11QMhNxUUnyeUeWWXBMPORDgLTI"))
}

func TestImageSampleStat(t *testing.T) {
	t.Parallel()

	x := simple.ImageSampleStat("", "")
	be.True(t, !x)

	name := filepath.Base(testpng(t))
	name = strings.TrimSuffix(name, filepath.Ext(name))
	prev := filepath.Dir(testpng(t))
	x = simple.ImageSampleStat(name, dir.Directory(prev))
	be.True(t, x)
}

func TestImageXY(t *testing.T) {
	t.Parallel()

	missing := [2]string{"0", ""}
	be.Equal(t, simple.ImageXY(nil, ""), missing)

	img := testpng(t)
	s := simple.ImageXY(testLog, img)
	be.Equal(t, s[0], "4,163")
	be.Equal(t, s[1], "500x500")

	tmp := t.TempDir()
	txt := filepath.Join(tmp, "note.txt")
	be.Err(t, os.WriteFile(txt, []byte("hello"), 0o600), nil)
	be.Equal(t, simple.ImageXY(testLog, txt), [2]string{"5", ""})

	badPNG := filepath.Join(tmp, "bad.png")
	be.Err(t, os.WriteFile(badPNG, []byte("not a png"), 0o600), nil)
	got := simple.ImageXY(testLog, badPNG)
	be.True(t, got[0] != "" && got[0] != "0")
	be.Equal(t, got[1], "")

	be.Equal(t, simple.ImageXY(testLog, filepath.Join(tmp, "missing.png")), missing)
}

func TestLinkID(t *testing.T) {
	t.Parallel()

	s, err := simple.LinkID(nil, "f")
	be.Err(t, err, simple.ErrLinkType)
	be.Equal(t, s, "")

	s, err = simple.LinkID("", "")
	be.Err(t, err, simple.ErrLinkType)
	be.Equal(t, s, "")

	s, err = simple.LinkID("a string", "a string")
	be.Err(t, err, simple.ErrLinkType)
	be.Equal(t, s, "")

	s, err = simple.LinkID(0, "f")
	be.Err(t, err, simple.ErrNegative)
	be.Equal(t, s, "")

	s, err = simple.LinkID(-1, "f")
	be.Err(t, err, simple.ErrNegative)
	be.Equal(t, s, "")

	s, err = simple.LinkID(uint(0), "f")
	be.Err(t, err, simple.ErrNegative)
	be.Equal(t, s, "")

	s, err = simple.LinkID(uint64(1)<<63, "f")
	be.Err(t, err, simple.ErrLinkType)
	be.Equal(t, s, "")

	s, err = simple.LinkID(uint(1), "")
	be.Err(t, err, nil)
	be.Equal(t, s, "/9b1c6")

	s, err = simple.LinkID(1, "")
	be.Err(t, err, nil)
	be.Equal(t, s, "/9b1c6")
}

func TestLinkRelr(t *testing.T) {
	t.Parallel()

	s, err := simple.LinkRelr("")
	be.Err(t, err)
	be.Equal(t, s, "")

	s, err = simple.LinkRelr("a string")
	be.Err(t, err, nil)
	be.Equal(t, s, "/g/a-string")
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

	s := simple.MagicAsTitle(testLog, "")
	be.Equal(t, s, "file not found")

	s = simple.MagicAsTitle(testLog, testpng(t))
	be.True(t, strings.Contains(s, "Portable Network Graphics"))

	dir := t.TempDir()

	staticANSI := filepath.Join(dir, "static.ans")
	be.Err(t, os.WriteFile(staticANSI, []byte("\x1b[31mHello World\x1b[0m\r\n"), 0o600), nil)
	be.Equal(t, simple.MagicAsTitle(testLog, staticANSI), "ANSI escaped text")

	blinkANSI := filepath.Join(dir, "blink.ans")
	be.Err(t, os.WriteFile(blinkANSI, []byte("\x1b[5;31mHello World\x1b[0m\r\n"), 0o600), nil)
	be.Equal(t, simple.MagicAsTitle(testLog, blinkANSI), "ANSI escaped text with blinking")

	be.Equal(t, simple.MagicAsTitle(testLog, testutil.FileData("maximum3ans").Abs()), simple.ANSImation)

	f, err := os.Open(staticANSI)
	be.Err(t, err, nil)
	be.Equal(t, simple.MagicString(f, magicnumber.ANSIEscapeText), "ANSI text")
	be.Err(t, f.Close(), nil)

	f, err = os.Open(blinkANSI)
	be.Err(t, err, nil)
	be.Equal(t, simple.MagicString(f, magicnumber.ANSIEscapeText), "ANSI text with blinking")
	be.Err(t, f.Close(), nil)

	f, err = os.Open(testutil.FileData("maximum3ans").Abs())
	be.Err(t, err, nil)
	be.Equal(t, simple.MagicString(f, magicnumber.ANSIEscapeText), simple.ANSImation)
	be.Err(t, f.Close(), nil)

	// Single-file PKLITE COM wrapping a TheDraw screen appends the inner type.
	td := make([]byte, 40)
	td[0], td[1], td[2] = 0xEB, 0x3D, 0x90
	copy(td[9:], "TheDraw COM file Screen Save\x1A")
	pkTheDraw := filepath.Join(dir, "thedraw_pk.com")
	be.Err(t, os.WriteFile(pkTheDraw, buildTestPKLITECOM(td), 0o600), nil)
	be.Equal(t, simple.MagicAsTitle(testLog, pkTheDraw),
		"PKLITE compressed executable of a TheDraw COM screen")

	// Single-file PKLITE COM wrapping unknown binary code keeps base PKLITE title.
	pkUnknown := filepath.Join(dir, "unknown_pk.com")
	be.Err(t, os.WriteFile(pkUnknown, buildTestPKLITECOM([]byte{0xB4, 0x09, 0xCD, 0x21, 0xC3}), 0o600), nil)
	be.Equal(t, simple.MagicAsTitle(testLog, pkUnknown), "PKLITE compressed executable")
}

// buildTestPKLITECOM encodes raw as literal-only small-mode PKLITE v1.12 COM data.
func buildTestPKLITECOM(raw []byte) []byte {
	const (
		comBase  = 0x100
		decomp   = 0x50
		dataOff  = 0x60
		pktTotal = dataOff
	)
	var hdr [pktTotal]byte
	// Copier pattern matched by pklite.findDataPos: and ax,fff0; mov di,ax; mov cx,N; mov si,decomp+100h
	copy(hdr[0x0f:], []byte{0x25, 0xf0, 0xff, 0x8b, 0xf8, 0xb9, 0x10, 0x00, 0xbe})
	binary.LittleEndian.PutUint16(hdr[0x18:], uint16(comBase+decomp))
	copy(hdr[46:], "PKLITE Copr. 1991")
	// Decompressor sig matched by pklite.findDataPos: std; mov di,ax; dec di; dec di; mov si,dataOff-2+100h
	copy(hdr[decomp:], []byte{0xfd, 0x8b, 0xf8, 0x4f, 0x4f, 0xbe})
	binary.LittleEndian.PutUint16(hdr[decomp+6:], uint16(comBase+dataOff-2))

	// Encode bitstream: control bits live in 16-bit LE words; bit() refills eagerly as soon as nbits reaches 16.
	stream := []byte{0, 0}
	var word uint16
	wordOff := 0
	var nbits uint
	pushBit := func(b uint16, refill bool) {
		word |= (b & 1) << nbits
		nbits++
		if nbits == 16 {
			binary.LittleEndian.PutUint16(stream[wordOff:], word)
			word, nbits = 0, 0
			if refill {
				wordOff = len(stream)
				stream = append(stream, 0, 0)
			} else {
				wordOff = -1
			}
		}
	}
	for _, b := range raw {
		pushBit(0, true)
		stream = append(stream, b)
	}
	// Match flag 1 + smallLengthCodes[7] (3 bits: 0, 1, 1 -> MSB-first code 0b011) + byte 0xFF stop marker.
	pushBit(1, true)
	pushBit(0, true)
	pushBit(1, true)
	pushBit(1, false)
	stream = append(stream, 0xFF)
	if wordOff >= 0 {
		binary.LittleEndian.PutUint16(stream[wordOff:], word)
	}
	return slices.Concat(hdr[:], stream)
}

func TestMIME(t *testing.T) {
	t.Parallel()

	s := simple.MIME(testLog, "")
	be.Equal(t, s, "file not found")

	s = simple.MIME(testLog, testpng(t))
	be.Equal(t, s, "image/png")

	tmp := t.TempDir()
	shortTxt := filepath.Join(tmp, "short.txt")
	be.Err(t, os.WriteFile(shortTxt, []byte("hello scene\n"), 0o600), nil)
	be.Equal(t, simple.MIME(testLog, shortTxt), "text/plain; charset=utf-8")

	emptyFile := filepath.Join(tmp, "empty.txt")
	be.Err(t, os.WriteFile(emptyFile, nil, 0o600), nil)
	be.Equal(t, simple.MIME(testLog, emptyFile), "text/plain; charset=utf-8")
}

func TestMkdirStale(t *testing.T) {
	t.Parallel()

	s := simple.MkdirStale(testLog, "")
	be.Equal(t, s, "")

	s1 := simple.MkdirStale(testLog, "a_string")
	be.True(t, strings.Contains(s1, "a_string"))
	s2 := simple.MkdirStale(testLog, "a_string")
	be.True(t, strings.Contains(s2, "a_string"))
	be.Equal(t, s1, s2)

	t.Cleanup(func() {
		_ = os.RemoveAll(s1)
		_ = os.RemoveAll(s2)
	})
}

func TestReleasers(t *testing.T) {
	t.Parallel()

	be.Equal(t, simple.Releasers("", "", true), "")
	be.Equal(t, simple.Releasers("group 1", "", false), "group 1")
	be.Equal(t, simple.Releasers("", "group 2", false), "group 2")

	s := simple.Releasers("group 1", "group 2", false)
	be.True(t, strings.Contains(string(s), "group 1"))
	be.True(t, strings.Contains(string(s), "group 2"))

	s = simple.Releasers("group 1", "group 2", true)
	be.True(t, strings.Contains(string(s), "group 1"))
	be.True(t, strings.Contains(string(s), "group 2"))
	be.True(t, strings.Contains(string(s), "published by"))
}

func TestOpenGraphImg(t *testing.T) {
	t.Parallel()

	empty := dir.Directory(t.TempDir())
	be.Equal(t, simple.OpenGraphImg("missing", empty, empty), "/image/layout/defacto2-ascii.png")

	prevDir := dir.Directory(filepath.Dir(testpng(t)))
	gotPrev := simple.OpenGraphImg("TEST", prevDir, empty)
	be.True(t, strings.HasPrefix(gotPrev, config.StaticOriginal()+"/TEST.png?sha384-"))

	gotThumb := simple.OpenGraphImg("TEST", empty, prevDir)
	be.True(t, strings.HasPrefix(gotThumb, config.StaticThumb()+"/TEST.png?sha384-"))
}

func TestAnimPreview(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	d := dir.Directory(tmp)
	be.Equal(t, simple.AnimPreview("", d), "")
	be.Equal(t, simple.AnimPreview("missing", d), "")

	be.Err(t, os.WriteFile(filepath.Join(tmp, "anim.webp"), []byte("webp"), 0o600), nil)
	be.Equal(t, simple.AnimPreview("anim", d), config.StaticOriginal()+"/anim.webp")
}

func TestScreenshot(t *testing.T) {
	t.Parallel()

	be.Equal(t, simple.Screenshot("", "", ""), "")

	prev := filepath.Dir(testpng(t))
	s := simple.Screenshot("TEST", "test", dir.Directory(prev))
	be.True(t, strings.Contains(string(s), `alt="test screenshot"`))
	be.True(t, strings.Contains(string(s), `<img src="/public/image`))

	// Escapes HTML characters in alt attribute.
	escaped := simple.Screenshot("TEST", `quoted "title" <tag>`, dir.Directory(prev))
	be.True(t, strings.Contains(string(escaped), `alt="quoted &#34;title&#34; &lt;tag&gt; screenshot"`))

	// Picture element with avif/webp + png/jpg fallback and standalone formats.
	tmp := t.TempDir()
	d := dir.Directory(tmp)
	be.Err(t, os.WriteFile(filepath.Join(tmp, "a.avif"), []byte("avifdata"), 0o600), nil)
	be.True(t, strings.Contains(string(simple.Screenshot("a", "av", d)), "a.avif?"))

	be.Err(t, os.WriteFile(filepath.Join(tmp, "a.png"), []byte("pngdata"), 0o600), nil)
	pic := string(simple.Screenshot("a", "av", d))
	be.True(t, strings.HasPrefix(pic, "<picture>"))
	be.True(t, strings.Contains(pic, `type="image/avif"`))
	be.True(t, strings.Contains(pic, "a.png?"))

	be.Err(t, os.WriteFile(filepath.Join(tmp, "w.webp"), []byte("webpdata"), 0o600), nil)
	be.True(t, strings.Contains(string(simple.Screenshot("w", "wp", d)), "w.webp?"))

	be.Err(t, os.WriteFile(filepath.Join(tmp, "w.jpg"), []byte("jpgdata"), 0o600), nil)
	picW := string(simple.Screenshot("w", "wp", d))
	be.True(t, strings.HasPrefix(picW, "<picture>"))
	be.True(t, strings.Contains(picW, `type="image/webp"`))
	be.True(t, strings.Contains(picW, "w.jpg?"))

	be.Err(t, os.WriteFile(filepath.Join(tmp, "j.jpg"), []byte("jpgdata"), 0o600), nil)
	be.True(t, strings.Contains(string(simple.Screenshot("j", "jp", d)), "j.jpg?"))
}

func TestStatHumanize(t *testing.T) {
	t.Parallel()

	x, y, z := simple.StatHumanize("")
	const none = "file not found"
	be.Equal(t, x, none)
	be.Equal(t, y, none)
	be.Equal(t, z, none)

	x, y, z = simple.StatHumanize(testpng(t))
	_, err := time.Parse(simple.YYYYMMDD, x)
	be.Err(t, err, nil)
	be.Equal(t, y, "4,163")
	be.True(t, strings.Contains(z, "4.2 kB"))
}

func TestThumb(t *testing.T) {
	t.Parallel()

	s := simple.Thumb("", "", "", false)
	be.True(t, strings.Contains(string(s), "<!-- no thumbnail found -->"))

	name := filepath.Base(testpng(t))
	name = strings.TrimSuffix(name, filepath.Ext(name))
	thumb := dir.Directory(filepath.Dir(testpng(t)))
	s = simple.Thumb(name, "a description", thumb, false)
	be.True(t, strings.Contains(string(s), `alt="a description thumbnail"`))
	be.True(t, strings.Contains(string(s), `class="card-img-top"`))

	s = simple.Thumb(name, `quoted "title" <tag>`, thumb, true)
	be.True(t, strings.Contains(string(s), `alt="quoted &#34;title&#34; &lt;tag&gt; thumbnail"`))
	be.True(t, strings.Contains(string(s), `class="card-img-bottom"`))

	tmp := t.TempDir()
	d := dir.Directory(tmp)
	be.Err(t, os.WriteFile(filepath.Join(tmp, "item.webp"), []byte("webp"), 0o600), nil)
	onlyWebp := string(simple.Thumb("item", "desc", d, false))
	be.True(t, strings.Contains(onlyWebp, "item.webp"))
	be.True(t, !strings.Contains(onlyWebp, "<picture"))

	be.Err(t, os.WriteFile(filepath.Join(tmp, "item.png"), []byte("png"), 0o600), nil)
	both := string(simple.Thumb("item", "desc", d, false))
	be.True(t, strings.HasPrefix(both, "<picture"))
	be.True(t, strings.Contains(both, `type="image/webp"`))
	be.True(t, strings.Contains(both, "item.png"))
}

func TestThumbSample(t *testing.T) {
	t.Parallel()

	const missing = "No thumbnail"
	x := simple.ThumbSample("", "")
	be.True(t, strings.Contains(string(x), missing))

	name := filepath.Base(testpng(t))
	name = strings.TrimSuffix(name, filepath.Ext(name))
	thumb := filepath.Dir(testpng(t))
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
			t.Parallel()
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
	be.Equal(t, len(hash1), 12)
	be.Equal(t, len(hash2), 12)

	// Test that hash only contains URL-safe base64 characters
	validChars := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_="
	for _, char := range hash1 {
		be.True(t, strings.ContainsRune(validChars, char))
	}
	for _, char := range hash2 {
		be.True(t, strings.ContainsRune(validChars, char))
	}
}
