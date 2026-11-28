package filerecord

import (
	"archive/zip"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Defacto2/archive"
	"github.com/Defacto2/server/internal/command"
	"github.com/Defacto2/server/internal/dir"
	"github.com/Defacto2/server/internal/logs"
	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/aarondl/null/v8"
	"github.com/nalgeon/be"
)

var (
	errExtractBoom = errors.New("extract failed")
	errExtractXSS  = errors.New("bad <script>alert(1)</script>")
	errWalkBoom    = errors.New("walk failed")
)

func TestRenameExtractedUUID(t *testing.T) {
	t.Parallel()

	const unid = "0192883a-b587-7659-befd-eb90e395953d"
	src1 := filepath.Join("/downloads", unid)

	// Single unpacked UUID file is renamed to the artifact filename.
	dir1 := t.TempDir()
	be.Err(t, os.WriteFile(filepath.Join(dir1, unid), []byte("unpacked"), 0o600), nil)

	renUUID(logs.Discard(), dir1, src1, "KURUPT.COM")

	_, err := os.Stat(filepath.Join(dir1, unid))
	be.True(t, os.IsNotExist(err))
	got, err := os.ReadFile(filepath.Join(dir1, "KURUPT.COM"))
	be.Err(t, err, nil)
	be.Equal(t, string(got), "unpacked")

	// Stale existing target file alongside unpacked UUID file is replaced cleanly.
	dir2 := t.TempDir()
	be.Err(t, os.WriteFile(filepath.Join(dir2, "KURUPT.COM"), []byte("stale-packed"), 0o600), nil)
	be.Err(t, os.WriteFile(filepath.Join(dir2, unid), []byte("fresh-unpacked"), 0o600), nil)

	renUUID(logs.Discard(), dir2, src1, "KURUPT.COM")

	_, err = os.Stat(filepath.Join(dir2, unid))
	be.True(t, os.IsNotExist(err))
	got, err = os.ReadFile(filepath.Join(dir2, "KURUPT.COM"))
	be.Err(t, err, nil)
	be.Equal(t, string(got), "fresh-unpacked")

	// Normal archives with other files are left untouched.
	dir3 := t.TempDir()
	be.Err(t, os.WriteFile(filepath.Join(dir3, unid), []byte("x"), 0o600), nil)
	be.Err(t, os.WriteFile(filepath.Join(dir3, "OTHER.NFO"), []byte("y"), 0o600), nil)

	renUUID(logs.Discard(), dir3, src1, "KURUPT.COM")

	_, err = os.Stat(filepath.Join(dir3, unid))
	be.Err(t, err, nil)

	// Early-return edge cases: invalid args, empty/missing dir, >2 entries, dir named unid, missing unid.
	renUUID(logs.Discard(), "", src1, "KURUPT.COM")
	renUUID(logs.Discard(), dir3, src1, "")
	renUUID(logs.Discard(), dir3, src1, ".")
	renUUID(logs.Discard(), dir3, src1, "..")
	renUUID(logs.Discard(), dir3, src1, unid)
	renUUID(logs.Discard(), filepath.Join(t.TempDir(), "missing"), src1, "KURUPT.COM")
	renUUID(logs.Discard(), t.TempDir(), src1, "KURUPT.COM")

	dirMany := t.TempDir()
	for _, f := range []string{unid, "a.txt", "b.txt"} {
		be.Err(t, os.WriteFile(filepath.Join(dirMany, f), []byte("x"), 0o600), nil)
	}
	renUUID(logs.Discard(), dirMany, src1, "KURUPT.COM")
	_, err = os.Stat(filepath.Join(dirMany, unid))
	be.Err(t, err, nil)

	dirSub := t.TempDir()
	be.Err(t, os.Mkdir(filepath.Join(dirSub, unid), 0o755), nil)
	renUUID(logs.Discard(), dirSub, src1, "KURUPT.COM")
	info, err := os.Stat(filepath.Join(dirSub, unid))
	be.Err(t, err, nil)
	be.True(t, info.IsDir())

	dirNoSrc := t.TempDir()
	be.Err(t, os.WriteFile(filepath.Join(dirNoSrc, "KURUPT.COM"), []byte("keep"), 0o600), nil)
	renUUID(logs.Discard(), dirNoSrc, src1, "KURUPT.COM")
	got, err = os.ReadFile(filepath.Join(dirNoSrc, "KURUPT.COM"))
	be.Err(t, err, nil)
	be.Equal(t, string(got), "keep")
}

func TestExtractErr(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	dlc := &DownloadContent{
		Source:   filepath.Join(tmp, "sample.lha"),
		platform: "dos",
		section:  "release",
	}

	// Generic extraction error with nil logger returns HTML-escaped error text.
	be.Equal(t, string(dlc.extractErr(nil, errExtractBoom)), errExtractBoom.Error())
	be.Equal(t, string(dlc.extractErr(nil, errExtractXSS)), "bad &lt;script&gt;alert(1)&lt;/script&gt;")

	// ErrNotImplemented with empty file returns empty-byte message.
	emptyFile := filepath.Join(tmp, "empty.bin")
	be.Err(t, os.WriteFile(emptyFile, nil, 0o600), nil)
	dlc.Source = emptyFile
	be.Equal(t, string(dlc.extractErr(logs.Discard(), archive.ErrNotImplemented)), "error, empty byte file")

	// ErrNotImplemented with non-empty file renders fallback HTML entry.
	nonEmpty := filepath.Join(tmp, "readme.txt")
	be.Err(t, os.WriteFile(nonEmpty, []byte("hello scene world\n"), 0o600), nil)
	dlc.Source = nonEmpty
	got := string(dlc.extractErr(logs.Discard(), archive.ErrNotImplemented))
	be.True(t, strings.Contains(got, "<div"))

	// extractNone on a missing file returns empty HTML instead of a blank row.
	be.Equal(t, string(dlc.extractNone(filepath.Join(tmp, "missing.txt"))), "")
}

func TestOldestModTime(t *testing.T) {
	t.Parallel()

	t1970 := time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
	t1992 := time.Date(1992, 3, 1, 0, 0, 0, 0, time.UTC)
	t1995 := time.Date(1995, 6, 15, 0, 0, 0, 0, time.UTC)

	be.Equal(t, oldestModTime(t1995, null.Time{}), t1995)
	be.Equal(t, oldestModTime(t1995, null.TimeFrom(t1970)), t1995)
	be.Equal(t, oldestModTime(time.Time{}, null.TimeFrom(t1992)), t1992)
	be.Equal(t, oldestModTime(t1970, null.TimeFrom(t1992)), t1992)
	be.Equal(t, oldestModTime(t1995, null.TimeFrom(t1992)), t1992)
	be.Equal(t, oldestModTime(t1992, null.TimeFrom(t1995)), t1992)
}

func TestWalkersAndIndexDiz(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	sub := filepath.Join(tmp, "sub")
	be.Err(t, os.Mkdir(sub, 0o755), nil)
	f1 := filepath.Join(tmp, "a.txt")
	f2 := filepath.Join(tmp, "b.txt")
	fZero := filepath.Join(tmp, "zero.txt")
	be.Err(t, os.WriteFile(f1, []byte("hello"), 0o600), nil)
	be.Err(t, os.WriteFile(f2, []byte("world"), 0o600), nil)
	be.Err(t, os.WriteFile(fZero, nil, 0o600), nil)

	dirInfo, err := os.Stat(sub)
	be.Err(t, err, nil)
	dirEntry := fs.FileInfoToDirEntry(dirInfo)

	f1Info, err := os.Stat(f1)
	be.Err(t, err, nil)
	f1Entry := fs.FileInfoToDirEntry(f1Info)

	f2Info, err := os.Stat(f2)
	be.Err(t, err, nil)
	f2Entry := fs.FileInfoToDirEntry(f2Info)

	fZeroInfo, err := os.Stat(fZero)
	be.Err(t, err, nil)
	fZeroEntry := fs.FileInfoToDirEntry(fZeroInfo)

	dlc := &DownloadContent{
		tempRoot: tmp,
		MaxItems: 1,
		platform: "dos",
		section:  "release",
		unid:     "test-unid",
	}

	// countFn
	be.Err(t, dlc.countFn(sub, dirEntry, errWalkBoom), fs.SkipDir)
	be.Err(t, dlc.countFn(sub, dirEntry, nil), nil)
	be.Equal(t, dlc.files, 0)
	be.Err(t, dlc.countFn(f1, f1Entry, nil), nil)
	be.Equal(t, dlc.files, 1)

	// listFn
	dlc.elems = make([]string, 2)
	dlc.index = -1
	be.Err(t, dlc.listFn(sub, dirEntry, errWalkBoom), filepath.SkipDir)
	be.Err(t, dlc.listFn(sub, dirEntry, nil), nil)
	be.Err(t, dlc.listFn(f1, f1Entry, nil), nil)
	be.Equal(t, dlc.index, 0)
	be.Equal(t, dlc.elems[0], "a.txt")

	// htmlFn: error, unselected file, zero-byte skip, and MaxItems SkipAll.
	dlc.results = []string{"a.txt", "b.txt", "zero.txt"}
	be.Err(t, dlc.htmlFn(f1, f1Entry, errWalkBoom), filepath.SkipDir)
	be.Err(t, dlc.htmlFn(sub, dirEntry, nil), nil)
	be.Err(t, dlc.htmlFn(fZero, fZeroEntry, nil), nil)
	be.Equal(t, dlc.zeroByteFiles, 1)
	dlc.MaxItems = 2
	be.Err(t, dlc.htmlFn(f1, f1Entry, nil), nil)
	be.Equal(t, dlc.entries, 1)
	be.Err(t, dlc.htmlFn(f2, f2Entry, nil), filepath.SkipAll)
	be.Equal(t, dlc.entries, 2)

	// indexDiz
	be.Equal(t, (&DownloadContent{}).indexDiz(), -1)
	be.Equal(t, (&DownloadContent{names: []string{"readme.nfo"}}).indexDiz(), -1)
	be.Equal(t, (&DownloadContent{names: []string{"readme.nfo", "FILE_ID.DIZ"}}).indexDiz(), 1)
	be.Equal(t, (&DownloadContent{names: []string{" sub/file_id.diz "}}).indexDiz(), 0)
}

func TestWalkerChmod(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	sub := filepath.Join(tmp, "sub")
	be.Err(t, os.Mkdir(sub, 0o700), nil)
	file := filepath.Join(sub, "item.txt")
	be.Err(t, os.WriteFile(file, []byte("data"), 0o600), nil)

	root, err := os.OpenRoot(tmp)
	be.Err(t, err, nil)
	defer func() { _ = root.Close() }()

	// Error input returns fs.SkipDir.
	be.Err(t, WalkerChmod(root, sub, nil, errWalkBoom), fs.SkipDir)

	// Directory is chmodded to 0o755 and file to 0o644.
	subInfo, err := os.Stat(sub)
	be.Err(t, err, nil)
	be.Err(t, WalkerChmod(root, sub, fs.FileInfoToDirEntry(subInfo), nil), nil)
	subInfo, err = os.Stat(sub)
	be.Err(t, err, nil)
	be.Equal(t, subInfo.Mode().Perm(), fs.FileMode(0o755))

	fileInfo, err := os.Stat(file)
	be.Err(t, err, nil)
	be.Err(t, WalkerChmod(root, file, fs.FileInfoToDirEntry(fileInfo), nil), nil)
	fileInfo, err = os.Stat(file)
	be.Err(t, err, nil)
	be.Equal(t, fileInfo.Mode().Perm(), fs.FileMode(0o644))

	// Missing path inside root returns an error.
	be.Err(t, WalkerChmod(root, filepath.Join(tmp, "missing.txt"), fs.FileInfoToDirEntry(fileInfo), nil))
	be.Err(t, WalkerChmod(root, filepath.Join(tmp, "missingdir"), fs.FileInfoToDirEntry(subInfo), nil))
}

func TestListMaxItems(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	zipPath := filepath.Join(tmp, t.Name()+".zip")
	f, err := os.Create(zipPath)
	be.Err(t, err, nil)
	zw := zip.NewWriter(f)
	for _, entry := range []struct {
		name string
		body []byte
	}{
		{"a.dat", []byte{0x00, 0x01, 0x02, 0x03}},
		{"b.dat", []byte{0x00, 0x04, 0x05, 0x06}},
		{"c.dat", []byte{0x00, 0x07, 0x08, 0x09}},
		{"FILE_ID.DIZ", []byte("scene description")},
	} {
		w, err := zw.Create(entry.name)
		be.Err(t, err, nil)
		_, err = w.Write(entry.body)
		be.Err(t, err, nil)
	}
	be.Err(t, zw.Close(), nil)
	be.Err(t, f.Close(), nil)

	art := &models.File{
		UUID:     null.StringFrom("0192883a-b587-7659-befd-eb90e395953d"),
		Platform: null.StringFrom("dos"),
		Section:  null.StringFrom("release"),
		Filename: null.StringFrom("pack.zip"),
	}
	thumbDir := t.TempDir()
	be.Err(t, os.WriteFile(filepath.Join(thumbDir, art.UUID.String+".png"), []byte("png"), 0o600), nil)
	extraDir := t.TempDir()
	be.Err(t, os.WriteFile(filepath.Join(extraDir, art.UUID.String+".diz"), []byte("diz"), 0o600), nil)
	dirs := command.Dirs{
		Preview:   dir.Directory(t.TempDir()),
		Thumbnail: dir.Directory(thumbDir),
		Extra:     dir.Directory(extraDir),
	}

	// Truncates to MaxItems after sorting (preserving FILE_ID.DIZ) and reports skipped count.
	dlc := DownloadContent{Source: zipPath, MaxItems: 2, Dirs: dirs}
	got := string(dlc.List(t.Context(), logs.Discard(), art))
	t.Cleanup(func() { _ = os.RemoveAll(dlc.tempRoot) })
	be.True(t, strings.Contains(got, "skipped 2 other files"))
	be.Equal(t, len(dlc.results), 2)
	be.Equal(t, dlc.results[0], "FILE_ID.DIZ")
	be.Equal(t, dlc.entries, 2)
	be.True(t, strings.Contains(strings.ToLower(got), "file_id.diz"))

	// Negative MaxItems must not panic and renders all entries without truncation.
	dlcNeg := DownloadContent{Source: zipPath, MaxItems: -1, Dirs: dirs}
	gotNeg := string(dlcNeg.List(t.Context(), logs.Discard(), art))
	be.True(t, !strings.Contains(gotNeg, "skipped"))
	be.Equal(t, dlcNeg.entries, 4)
}
