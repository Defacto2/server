package app

import (
	"bytes"
	"database/sql"
	"encoding/binary"
	"errors"
	"image"
	"image/png"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Defacto2/magicnumber"
	"github.com/Defacto2/server/internal/dir"
	"github.com/Defacto2/server/internal/logs"
	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/Defacto2/server/internal/tags"
	"github.com/aarondl/null/v8"
	"github.com/labstack/echo/v5"
	"github.com/nalgeon/be"
	"golang.org/x/text/encoding/unicode"
)

var errTestDirs = errors.New("dirs test error")

func newSAUCE(title, author, group, yyyymmdd string) []byte {
	rec := make([]byte, 128)
	copy(rec[0:7], "SAUCE00")
	copy(rec[7:42], title)
	copy(rec[42:62], author)
	copy(rec[62:82], group)
	copy(rec[82:90], yyyymmdd)
	return rec
}

func tinyPNG(t *testing.T, path string) {
	t.Helper()

	var buf bytes.Buffer
	img := image.NewRGBA(image.Rect(0, 0, 2, 3))
	be.Err(t, png.Encode(&buf, img), nil)
	be.Err(t, os.WriteFile(path, buf.Bytes(), 0o600), nil)
}

func TestTheDrawCOM(t *testing.T) {
	t.Parallel()

	const stubLen = 0xF0
	const payloadLen = 4
	payload := []byte{0x0F, 'H', 'I', 0x18}
	be.Equal(t, len(payload), payloadLen)
	td := make([]byte, stubLen+payloadLen)
	copy(td[0:3], []byte{0xEB, 0x3D, 0x90})
	binary.LittleEndian.PutUint16(td[4:6], payloadLen)
	copy(td[9:], []byte("TheDraw COM file Screen Save\x1A"))
	copy(td[stubLen:], payload)

	dlDir := t.TempDir()
	const unid = "b13e2a00-0000-0000-0000-000000000001"
	be.Err(t, os.WriteFile(filepath.Join(dlDir, unid), td, 0o600), nil)

	ds := Dirs{Download: dir.Directory(dlDir)}
	art := &models.File{
		UUID:           null.StringFrom(unid),
		DateIssuedYear: null.Int16From(1992),
	}
	data := map[string]any{
		"magic":    "Binary data or binary text",
		"platform": "dos",
	}
	elems := []string{"reader-invert", "border"}
	buf := bytes.NewBufferString("<span>HI</span>")

	got := ds.addTextBinary(art, buf, elems, data)
	be.Equal(t, got["contentBinarySwappers"], true)

	classes, ok := got["preElementClass"].(string)
	be.True(t, ok)
	be.True(t, strings.Contains(classes, "font-dos"))
	be.True(t, strings.Contains(classes, "render"))
	be.True(t, !strings.Contains(classes, "text-bg-dark"))
}

func TestBinarySAUCE(t *testing.T) {
	t.Parallel()

	ds := Dirs{}
	art := &models.File{
		UUID:           null.StringFrom("b13e2a00-0000-0000-0000-000000000002"),
		DateIssuedYear: null.Int16From(1995),
	}
	elems := []string{"reader-invert", "border"}
	buf := bytes.NewBufferString("<span>BIN</span>")

	for _, magic := range []string{"Binary data or binary text", "SAUCE binary text", "IBM CGA video RAM dump"} {
		data := map[string]any{
			"magic":    magic,
			"platform": "ansi",
		}
		got := ds.addTextBinary(art, buf, elems, data)
		be.Equal(t, got["contentBinarySwappers"], false)
		classes, ok := got["preElementClass"].(string)
		be.True(t, ok)
		be.True(t, strings.Contains(classes, "font-squared"))
		be.True(t, strings.Contains(classes, "reader-hires"))
		be.True(t, strings.Contains(classes, "text-bg-dark"))
	}
}

func TestAddTextBinary(t *testing.T) {
	t.Parallel()

	ds := Dirs{}
	buf := bytes.NewBufferString("<span>ART</span>")
	elems := []string{"reader-invert", "border"}

	// nil guard
	got := ds.addTextBinary(nil, buf, elems, map[string]any{})
	be.Equal(t, len(got), 0)

	// Amiga ANSI text
	amiga := &models.File{DateIssuedYear: null.Int16From(1993)}
	got = ds.addTextBinary(amiga, buf, elems, map[string]any{"platform": textamiga, "magic": "ANSI"})
	be.True(t, strings.Contains(got["preElementClass"].(string), "font-amiga"))

	// CGA era (<= 1988) ANSI text
	cga := &models.File{DateIssuedYear: null.Int16From(1986)}
	got = ds.addTextBinary(cga, buf, elems, map[string]any{"platform": "dos", "magic": "ANSI"})
	be.Equal(t, got["contentBinarySwappers"], false)
	be.True(t, strings.Contains(got["preElementClass"].(string), "font-dos-cga"))

	// CGA era binary text
	got = ds.addTextBinary(cga, buf, elems, map[string]any{
		"platform": "dos",
		"magic":    magicnumber.Unknown.Title(),
	})
	be.Equal(t, got["contentBinarySwappers"], false)
	be.True(t, strings.Contains(got["preElementClass"].(string), "reader"))
	be.True(t, strings.Contains(got["preElementClass"].(string), "text-bg-dark"))

	// High-res binary text with empty elems must not panic and keeps text-bg-dark.
	hiRes := &models.File{DateIssuedYear: null.Int16From(1995)}
	got = ds.addTextBinary(hiRes, buf, nil, map[string]any{
		"platform": "ansi",
		"magic":    magicnumber.Unknown.Title(),
	})
	be.Equal(t, got["preElementClass"], "font-squared reader-hires text-bg-dark")

	// Caller slice with spare capacity must not be mutated.
	backing := []string{"reader-invert", "border", "sentinel"}
	shortElems := backing[:2]
	_ = ds.addTextBinary(hiRes, buf, shortElems, map[string]any{
		"platform": "ansi",
		"magic":    magicnumber.Unknown.Title(),
	})
	be.Equal(t, backing[2], "sentinel")
}

func TestBuildRoutes(t *testing.T) {
	t.Parallel()

	entries, counts := buildRoutes(echo.Routes{
		{Method: http.MethodHead, Path: "/ping"},
		{Method: http.MethodGet, Path: "/"},
		{Method: http.MethodOptions, Path: "/ping"},
	})
	be.Equal(t, len(entries), 3)
	be.Equal(t, len(counts), 3)
	be.Equal(t, counts[0].Method, http.MethodGet)
	be.Equal(t, counts[1].Method, http.MethodHead)
	be.Equal(t, counts[1].Count, 1)
	be.Equal(t, counts[2].Method, http.MethodOptions)
	be.Equal(t, counts[2].Count, 1)
}

func TestLogErr(t *testing.T) {
	t.Parallel()

	ds := Dirs{URI: "abc12", ID: 42, UUID: "u-1"}
	ds.logErr(nil, "ignored", errTestDirs)

	var buf bytes.Buffer
	sl := slog.New(slog.NewTextHandler(&buf, nil))
	ds.logErr(sl, "no err", nil)
	be.True(t, strings.Contains(buf.String(), "uri=abc12"))
	be.True(t, !strings.Contains(buf.String(), "error="))

	buf.Reset()
	ds.logErr(sl, "with err", errTestDirs)
	be.True(t, strings.Contains(buf.String(), "id=42"))
	be.True(t, strings.Contains(buf.String(), "dirs test error"))
}

func TestAddMetadata(t *testing.T) {
	t.Parallel()

	ds := Dirs{Extra: dir.Directory(t.TempDir())}

	be.Equal(t, len(ds.addAttribution(nil, map[string]any{})), 0)
	be.Equal(t, len(ds.addLinks(nil, map[string]any{})), 0)
	be.Equal(t, len(ds.addDownload(nil, map[string]any{})), 0)

	art := &models.File{
		Filename:           null.StringFrom("demo.zip"),
		Filesize:           null.Int64From(2048),
		Platform:           null.StringFrom("dos"),
		Section:            null.StringFrom(tags.Interview.String()),
		CreditText:         null.StringFrom("Alice"),
		CreditIllustration: null.StringFrom("Bob"),
		CreditProgram:      null.StringFrom("Carol"),
		CreditAudio:        null.StringFrom("Dave"),
		WebIDDemozoo:       null.Int64From(123),
		WebIDPouet:         null.Int64From(456),
		WebID16colors:      null.StringFrom("pack/file.ans"),
		WebIDYoutube:       null.StringFrom("dQw4w9WgXcQ"),
		WebIDGithub:        null.StringFrom("Defacto2/server"),
		Comment:            null.StringFrom("Broken link https://example.com/alert"),
	}

	attr := ds.addAttribution(art, map[string]any{})
	be.Equal(t, attr["writers"], "Alice")
	be.Equal(t, attr["artists"], "Bob")
	be.Equal(t, attr["programmers"], "Carol")
	be.Equal(t, attr["musicians"], "Dave")

	links := ds.addLinks(art, map[string]any{})
	be.True(t, links["demozoo"] != "")
	be.True(t, links["pouet"] != "")
	be.True(t, links["github"] != "")
	be.True(t, links["youtube"] != "")
	be.True(t, links["sixteenColors"] != "")

	dl := ds.addDownload(art, map[string]any{})
	be.Equal(t, dl["filename"], "demo.zip")
	be.Equal(t, dl["filebyte"], any(int64(2048)))
	be.Equal(t, dl["platform"], "dos")
	be.Equal(t, dl["section"], any(tags.Interview.String()))
	be.Equal(t, dl["extraZip"], false)
	be.Equal(t, dl["noDownload"], nil)

	// Configured Download dir with missing vs existing file sets noDownload accurately.
	dlDir := t.TempDir()
	const unid = "b13e2a00-0000-0000-0000-000000000005"
	dlDS := Dirs{Download: dir.Directory(dlDir), Extra: ds.Extra}
	art.UUID = null.StringFrom(unid)
	missing := dlDS.addDownload(art, map[string]any{})
	be.Equal(t, missing["noDownload"], true)

	be.Err(t, os.WriteFile(filepath.Join(dlDir, unid), []byte("ok"), 0o600), nil)
	present := dlDS.addDownload(art, map[string]any{})
	be.Equal(t, present["noDownload"], nil)
}

func TestAddEditor(t *testing.T) {
	t.Parallel()

	dlDir := t.TempDir()
	prevDir := t.TempDir()
	thumbDir := t.TempDir()
	extraDir := t.TempDir()
	const unid = "b13e2a00-0000-0000-0000-000000000010"

	be.Err(t, os.WriteFile(filepath.Join(dlDir, unid), []byte("hello world\n"), 0o600), nil)
	tinyPNG(t, filepath.Join(prevDir, unid+".png"))
	tinyPNG(t, filepath.Join(thumbDir, unid+".png"))
	be.Err(t, os.WriteFile(filepath.Join(extraDir, unid+".txt"), []byte("readme\n"), 0o600), nil)

	ds := Dirs{
		Download:  dir.Directory(dlDir),
		Preview:   dir.Directory(prevDir),
		Thumbnail: dir.Directory(thumbDir),
		Extra:     dir.Directory(extraDir),
		UUID:      unid,
		Platform:  "dos",
	}
	ds.Maximum.Items = maxArchiveItems

	art := &models.File{
		ID:       99,
		UUID:     null.StringFrom(unid),
		Filename: null.StringFrom("readme.txt"),
		Platform: null.StringFrom("dos"),
		Section:  null.StringFrom("releaseinfo"),
	}
	got := ds.addEditor(t.Context(), logs.Discard(), echoCtx(t), art, map[string]any{})
	be.Equal(t, got["modID"], any(int64(99)))
	be.Equal(t, got["epochYear"], PcEpoch)
	be.Equal(t, got["readonlymode"], false)
	be.Equal(t, got["missingAssets"], "")
}

func TestAddEmu(t *testing.T) {
	t.Parallel()

	sl := logs.Discard()
	ds := Dirs{}

	// nil guard
	be.Equal(t, len(ds.addEmu(nil, nil, map[string]any{})), 0)

	// non-DOS file
	nonDOS := &models.File{Platform: null.StringFrom("windows"), Filename: null.StringFrom("app.exe")}
	got := ds.addEmu(sl, nonDOS, map[string]any{})
	be.Equal(t, got["jsdos6"], false)

	// DOS executable
	dosExe := &models.File{Platform: null.StringFrom("dos"), Filename: null.StringFrom("intro.com")}
	got = ds.addEmu(sl, dosExe, map[string]any{})
	be.Equal(t, got["jsdos6"], true)
	be.True(t, got["jsdos6Config"] != "")
	be.Equal(t, got["jsdos6Zip"], false)
}

func TestAddReadme(t *testing.T) {
	t.Parallel()

	dlDir := t.TempDir()
	extraDir := t.TempDir()
	const unid = "b13e2a00-0000-0000-0000-000000000020"

	ds := Dirs{
		Download: dir.Directory(dlDir),
		Extra:    dir.Directory(extraDir),
		UUID:     unid,
	}
	ds.Maximum.Readme = sizeLimitBytes

	// addSAUCE guards
	be.Equal(t, len(ds.addSAUCE(nil, map[string]any{})), 0)
	art := &models.File{
		UUID:           null.StringFrom(unid),
		Filename:       null.StringFrom("info.nfo"),
		Platform:       null.StringFrom("text"),
		Section:        null.StringFrom("releaseinfo"),
		DateIssuedYear: null.Int16From(1994),
	}
	// Missing file -> unchanged
	be.Equal(t, len(ds.addSAUCE(art, map[string]any{})), 0)

	// File without SAUCE
	dlPath := filepath.Join(dlDir, unid)
	be.Err(t, os.WriteFile(dlPath, []byte("plain text without sauce"), 0o600), nil)
	be.Equal(t, len(ds.addSAUCE(art, map[string]any{})), 0)

	// File with valid SAUCE record
	withSauce := append([]byte("hello text\x1a"), newSAUCE("Test Title", "Test Author", "Test Group", "19940506")...)
	be.Err(t, os.WriteFile(dlPath, withSauce, 0o600), nil)
	got := ds.addSAUCE(art, map[string]any{})
	be.Equal(t, got["readmeSAUCE"], true)
	be.Equal(t, got["sauceTitle"], "Test Title")
	be.Equal(t, got["sauceAuthor"], "Test Author")
	be.Equal(t, got["sauceGroup"], "Test Group")
	be.Equal(t, got["sauceDate"], "1994 May 06")

	// addReadme nil args
	_, err := ds.addReadme(nil, art, map[string]any{})
	be.Err(t, err)

	// addReadme with plain text in Extra (<uuid>.txt)
	be.Err(t, os.WriteFile(filepath.Join(extraDir, unid+".txt"), []byte("Hello from extra readme!\n"), 0o600), nil)
	rdData, err := ds.addReadme(logs.Discard(), art, map[string]any{"platform": "text"})
	be.Err(t, err, nil)
	be.True(t, strings.Contains(rdData["contentUTF8"].(string), "Hello from extra readme!"))
}

func TestAddText(t *testing.T) {
	t.Parallel()

	sl := logs.Discard()
	ds := Dirs{}

	// addText8bit nil & disabled & empty guards
	be.Equal(t, len(ds.addText8bit(nil, nil, nil, map[string]any{})), 0)
	disabled := &models.File{RetrotxtNoReadme: null.Int16From(1)}
	be.Equal(t, len(ds.addText8bit(sl, disabled, bytes.NewBufferString("abc"), map[string]any{})), 0)
	plain := &models.File{DateIssuedYear: null.Int16From(1985)}
	be.Equal(t, len(ds.addText8bit(sl, plain, bytes.NewBuffer(nil), map[string]any{})), 0)

	// CGA year + PC platform + CP437
	got := ds.addText8bit(sl, plain, bytes.NewBufferString("Line 1\nLine 2\n"), map[string]any{"platform": "dos"})
	be.Equal(t, got["vgaCheck"], "checked")
	be.Equal(t, got["preClassCP437"], "font-dos-mda")
	be.True(t, strings.Contains(got["contentCP437"].(string), "Line 1"))

	// Amiga platform + ISO-8859-1 NBSP/SHY replacement
	amigaDS := Dirs{Platform: textamiga}
	amigaArt := &models.File{DateIssuedYear: null.Int16From(1993)}
	got = amigaDS.addText8bit(sl, amigaArt, bytes.NewBuffer([]byte("foo\xa0bar\xadbaz\n")), map[string]any{"platform": textamiga})
	be.Equal(t, got["topazCheck"], "checked")
	be.Equal(t, got["preClassLatin1"], "")
	be.True(t, strings.Contains(got["contentLatin1"].(string), "foo bar-baz"))

	// addTextUTF8 branches
	be.Equal(t, len(ds.addTextUTF8(nil, nil, map[string]any{})), 0)
	got = ds.addTextUTF8(bytes.NewBufferString("fallback"), bytes.NewBufferString("unicode"), map[string]any{})
	be.Equal(t, got["contentUTF8"], "unicode")
	got = ds.addTextUTF8(bytes.NewBufferString("fallback"), bytes.NewBuffer(nil), map[string]any{})
	be.Equal(t, got["contentUTF8"], "fallback")
}

func TestAddZip(t *testing.T) {
	t.Parallel()

	// Zero-value Maximum.Items defaults to maxArchiveItems instead of wiping items.
	zeroDS := Dirs{}
	zeroGot := zeroDS.addZip("a.txt\nb.txt\n", map[string]any{})
	be.Equal(t, zeroGot["contentDesc"], "contains 2 files")
	be.Equal(t, len(zeroGot["content"].([]string)), 2)

	ds := Dirs{}
	ds.Maximum.Items = 2

	be.Equal(t, len(ds.addZip("", map[string]any{})), 0)

	onlyBlank := ds.addZip("   \n\t\n", map[string]any{})
	be.Equal(t, onlyBlank["contentDesc"], "")

	one := ds.addZip("readme.txt", map[string]any{})
	be.Equal(t, one["contentDesc"], "contains one file")
	be.Equal(t, len(one["content"].([]string)), 1)

	many := ds.addZip("a.txt\nb.txt\nc.txt\n", map[string]any{})
	be.Equal(t, many["contentDesc"], "contains 3 files")
	be.Equal(t, len(many["content"].([]string)), 2)
}

func TestEncodingUTF(t *testing.T) {
	t.Parallel()

	for _, magic := range []string{"UTF-8 Unicode text", "UTF-16 Unicode text"} {
		ds := Dirs{Magic: magic}
		be.Equal(t, ds.encoding(nil), unicode.UTF8)
	}
}

func TestFixGuards(t *testing.T) {
	t.Parallel()

	sl := logs.Discard()
	db := &sql.DB{}
	ds := Dirs{ID: 1, UUID: "u-1"}

	// fixANSI guards & early exits
	be.Equal(t, len(ds.fixANSI(t.Context(), nil, nil, map[string]any{})), 0)
	be.Equal(t, len(ds.fixANSI(t.Context(), sl, db, map[string]any{})), 0)
	be.Equal(t, len(ds.fixANSI(t.Context(), sl, db, map[string]any{"modOS": "text"})), 1)
	got := ds.fixANSI(t.Context(), sl, db, map[string]any{
		"modOS":          "dos",
		"modMagicNumber": magicnumber.PlainText.Title(),
	})
	be.Equal(t, got["platform"], nil)

	// fixMagic guards & early exits (no DB call when recMagic == modMagic)
	be.Equal(t, len(ds.fixMagic(t.Context(), nil, nil, map[string]any{})), 0)
	// non-string modMagic
	got = ds.fixMagic(t.Context(), sl, db, map[string]any{"magic": "a", "modMagicNumber": 123})
	be.Equal(t, got["magic"], 123)
	// missing extraZip
	got = ds.fixMagic(t.Context(), sl, db, map[string]any{"magic": "a", "modMagicNumber": "a"})
	be.Equal(t, got["magic"], "a")
	// extraZip already true
	got = ds.fixMagic(t.Context(), sl, db, map[string]any{"magic": "a", "modMagicNumber": "a", "extraZip": true})
	be.Equal(t, got["extraZip"], true)
	// missing modDecompressLoc
	got = ds.fixMagic(t.Context(), sl, db, map[string]any{"magic": "a", "modMagicNumber": "a", "extraZip": false})
	be.Equal(t, got["extraZip"], false)
	// nonexistent modDecompressLoc
	got = ds.fixMagic(t.Context(), sl, db, map[string]any{
		"magic":            "a",
		"modMagicNumber":   "a",
		"extraZip":         false,
		"modDecompressLoc": filepath.Join(t.TempDir(), "does-not-exist"),
	})
	be.Equal(t, got["extraZip"], false)
}

func TestMakeAssets(t *testing.T) {
	t.Parallel()

	sl := logs.Discard()
	dlDir := t.TempDir()
	extraDir := t.TempDir()
	thumbDir := t.TempDir()
	srcDir := t.TempDir()
	const unid = "b13e2a00-0000-0000-0000-000000000030"

	be.Err(t, os.WriteFile(filepath.Join(srcDir, "FILE_ID.DIZ"), []byte("hello archive\n"), 0o600), nil)

	ds := Dirs{
		Download:  dir.Directory(dlDir),
		Extra:     dir.Directory(extraDir),
		Thumbnail: dir.Directory(thumbDir),
		UUID:      unid,
	}

	// makeAssets nil & default no-op
	be.Equal(t, len(ds.makeAssets(t.Context(), nil, srcDir, "unknown", map[string]any{})), 0)
	got := ds.makeAssets(t.Context(), sl, srcDir, "unknown", map[string]any{})
	be.Equal(t, got["extraZip"], nil)

	// makeAssets repacking a legacy archive via makeZipfile
	got = ds.makeAssets(t.Context(), sl, srcDir, magicnumber.ArchiveRobertJung.Title(), map[string]any{})
	be.Equal(t, got["extraZip"], true)
	_, err := os.Stat(filepath.Join(extraDir, unid+".zip"))
	be.Err(t, err, nil)

	// makeZipfile error on missing source dir
	_, err = ds.makeZipfile(filepath.Join(t.TempDir(), "missing"))
	be.Err(t, err)

	// makeTextImages nil guard, error preserves missingAssets, and existing-thumbnail short-circuit
	be.Equal(t, len(ds.makeTextImages(t.Context(), nil, map[string]any{})), 0)
	badThumb := Dirs{UUID: unid}
	got = badThumb.makeTextImages(t.Context(), sl, map[string]any{"missingAssets": "create a thumbnail image"})
	be.Equal(t, got["missingAssets"], "create a thumbnail image")
	be.Err(t, os.WriteFile(filepath.Join(thumbDir, unid+".png"), []byte("stub"), 0o600), nil)
	got = ds.makeTextImages(t.Context(), sl, map[string]any{"missingAssets": "keep"})
	be.Equal(t, got["missingAssets"], "keep")
}

func TestPaths(t *testing.T) {
	t.Parallel()

	dlDir := t.TempDir()
	prevDir := t.TempDir()
	thumbDir := t.TempDir()
	extraDir := t.TempDir()
	const unid = "b13e2a00-0000-0000-0000-000000000040"

	ds := Dirs{
		Download:  dir.Directory(dlDir),
		Preview:   dir.Directory(prevDir),
		Thumbnail: dir.Directory(thumbDir),
		Extra:     dir.Directory(extraDir),
		UUID:      unid,
	}

	// pathsSuggest: all missing
	be.Equal(t, ds.pathsSuggest(), "offer a file for download + create a preview image + create a thumbnail image")

	// pathsSuggest: audio only needs download
	ds.Platform = tags.Audio.String()
	be.Equal(t, ds.pathsSuggest(), "offer a file for download")

	// pathsSuggest: text skips preview suggestion
	ds.Platform = tags.Text.String()
	be.Equal(t, ds.pathsSuggest(), "offer a file for download + create a thumbnail image")
	be.True(t, !ds.screenshot())

	// Create all assets
	ds.Platform = "dos"
	be.Err(t, os.WriteFile(filepath.Join(dlDir, unid), []byte("bin"), 0o600), nil)
	tinyPNG(t, filepath.Join(prevDir, unid+".png"))
	be.Err(t, os.WriteFile(filepath.Join(prevDir, unid+".avif"), []byte("avifdata"), 0o600), nil)
	tinyPNG(t, filepath.Join(thumbDir, unid+".png"))
	be.Err(t, os.WriteFile(filepath.Join(extraDir, unid+".diz"), []byte("line1\nline2\n"), 0o600), nil)
	be.Err(t, os.WriteFile(filepath.Join(extraDir, unid+".hlp"), []byte("help\n"), 0o600), nil)
	be.Err(t, os.WriteFile(filepath.Join(extraDir, unid+".txt"), []byte("readme\n"), 0o600), nil)
	be.Err(t, os.WriteFile(filepath.Join(extraDir, unid+".zip"), []byte("zipbytes"), 0o600), nil)

	be.Equal(t, ds.pathsSuggest(), "")

	previews := ds.pathsPreview(nil)
	be.Equal(t, previews["AVIF"][0], "8")
	be.Equal(t, previews["PNG"][1], "2x3")

	thumbs := ds.pathsThumb(nil)
	be.Equal(t, thumbs["PNG"][1], "2x3")

	extras := ds.pathsExtras()
	be.True(t, extras["FILE_ID"][0] != "")
	be.True(t, extras["HELPER"][0] != "")
	be.True(t, extras["README"][0] != "")
	be.Equal(t, extras["Repacked ZIP"][1], "Deflate compression")
}

func TestOneByKey(t *testing.T) {
	t.Parallel()

	ds := Dirs{URI: "abc"}
	art, err := ds.oneByKey(t.Context(), logs.Discard(), echoCtx(t), nil)
	be.Err(t, err)
	be.True(t, art == nil)
}
