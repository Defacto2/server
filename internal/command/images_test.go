package command_test

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"uuid"

	"github.com/Defacto2/helper"
	"github.com/Defacto2/server/internal/command"
	"github.com/Defacto2/server/internal/dir"
	"github.com/Defacto2/server/internal/testutil"
	"github.com/nalgeon/be"
)

const invalid = "this-is#invalid!"

var (
	testdata      = testutil.Testdata()
	testdataCount = testutil.CountTest()
)

func requireTool(t *testing.T, name string) {
	t.Helper()
	if _, err := command.Lookup(name); err != nil {
		t.Skipf("%s not available: %v", name, err)
	}
}

// setupTestDir copies all files from testdata/ to tempDir. A string that should normally
// be the result of [testing.TempDir]. After the transfer there is a file count to confirm
// the copies. Note that the copying keeps the exact casing of the filenames.
//
// Returned are the base name of the first copied file,
// and the temporary directory used.
func setupTestDir(t *testing.T, tempDir string) (string, string) {
	t.Helper()

	var baseN string

	entries, err := os.ReadDir(testdata)
	if err != nil {
		t.Fatalf("failed to read testdata directory: %v", err)
	}
	entries = slices.DeleteFunc(entries, func(e os.DirEntry) bool {
		skip := !strings.HasPrefix(e.Name(), "TEST.")
		return skip
	})

	if n := countFiles(t, testdata); n != testdataCount {
		t.Fatalf("expected %d test files in testdata, but got %d", testdataCount, n)
	}

	src, err := os.OpenRoot(testdata)
	if err != nil {
		t.Fatalf("failed to open testdata root: %v", err)
	}
	defer src.Close()

	dst, err := os.OpenRoot(tempDir)
	if err != nil {
		t.Fatalf("failed to open tempDir root: %v", err)
	}
	defer dst.Close()

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()

		if baseN == "" {
			filename := filepath.Base(name)
			baseN = strings.TrimSuffix(filename, filepath.Ext(filename))
		}

		data, err := src.ReadFile(name)
		if err != nil {
			t.Fatalf("failed to read fixture %s: %v", name, err)
		}

		if err := dst.WriteFile(name, data, 0o600); err != nil {
			t.Fatalf("failed to write fixture to temp dir %s: %v", name, err)
		}
	}

	if baseN == "" {
		t.Fatalf("setupTestDir: no files found in %s", testdata)
	}
	n := countFiles(t, tempDir)
	if n != testdataCount {
		t.Fatalf("found %d test files in the temp directory, wanted %d: %s",
			n, testdataCount, tempDir)
	}

	return baseN, tempDir
}

// countFiles returns the number of regular files in dir (excluding subdirectories).
func countFiles(t *testing.T, dir string) int {
	t.Helper()

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("failed to read directory %s: %v", dir, err)
	}

	count := 0
	for _, entry := range entries {
		if entry.Type().IsRegular() && strings.HasPrefix(entry.Name(), "TEST.") {
			count++
		}
	}
	return count
}

func TestImagesDelete(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	unid, _ := setupTestDir(t, dir)

	err := command.ImagesDelete(unid, dir)
	be.Err(t, err, nil)

	const want = 5 // want 3 as ".ascii", ".bmp", ".pcx" are ignored by the deleter

	got := countFiles(t, dir)
	be.Equal(t, got, want)

	err = command.ImagesDelete("", dir)
	be.Err(t, err)

	err = command.ImagesDelete(unid, "")
	be.Err(t, err)
}

func TestImagesPixelate(t *testing.T) {
	t.Parallel()

	sl := testLog

	dir := t.TempDir()
	unid, _ := setupTestDir(t, dir)

	err := command.ImagesPixelate(t.Context(), sl, unid, dir)
	be.Err(t, err, nil)

	img := filepath.Join(dir, unid+".PNG")
	original, err := filepath.Abs(filepath.Join(testdata, "TEST.PNG"))
	be.Err(t, err, nil)

	st, err := os.Stat(img)
	be.Err(t, err, nil)
	imgSize := st.Size()

	st, err = os.Stat(original)
	be.Err(t, err, nil)
	ogSize := st.Size()

	// pixelated images should be smaller in file size because there is less image details
	pixelated := ogSize > 0 && ogSize > imgSize
	be.True(t, pixelated)

	err = command.ImagesPixelate(t.Context(), sl, "", dir)
	be.Err(t, err)

	err = command.ImagesPixelate(t.Context(), sl, unid, "")
	be.Err(t, err)
}

func TestDirsThumbs(t *testing.T) {
	t.Parallel()

	unid, path := setupTestDir(t, t.TempDir())

	thumbdir := t.TempDir()
	dirs := command.Dirs{
		Preview:   dir.Directory(path),
		Thumbnail: dir.Directory(thumbdir),
	}

	count := countFiles(t, thumbdir)
	be.Equal(t, count, 0)

	sl := testLog
	err := dirs.Thumbs(t.Context(), sl, unid, command.Pixel)
	be.Err(t, err, nil)

	count = countFiles(t, thumbdir)
	const wantThumb = 1 // as unid is named "TEST", and all the testdata are named "TEST.*", only 1 thumbnail is created
	be.Equal(t, count, wantThumb)

	// check invalid thumb value
	err = dirs.Thumbs(t.Context(), sl, unid, -1)
	be.Err(t, err)
	// check invalid dirs values
	dirs.Preview = ""
	err = dirs.Thumbs(t.Context(), sl, unid, command.Pixel)
	be.Err(t, err)

	dirs.Preview = dir.Directory(path)
	dirs.Thumbnail = dir.Directory(invalid)
	err = dirs.Thumbs(t.Context(), sl, unid, command.Pixel)
	be.Err(t, err)

	dirs.Preview = dir.Directory(t.TempDir())
	dirs.Thumbnail = dir.Directory(thumbdir)
	err = dirs.Thumbs(t.Context(), sl, "missing-preview", command.Pixel)
	be.True(t, errors.Is(err, command.ErrNoImages))
}

func TestAlignThumbs(t *testing.T) {
	t.Parallel()

	unid, prevdir := setupTestDir(t, t.TempDir())
	thumbdir := t.TempDir()
	preview := dir.Directory(prevdir)
	thumbnail := dir.Directory(thumbdir)

	count := countFiles(t, thumbdir)
	be.Equal(t, count, 0)

	sl := testLog

	align := command.Top
	err := align.Thumbs(t.Context(), sl, unid, preview, thumbnail)
	be.Err(t, err, nil)

	align = command.Middle
	err = align.Thumbs(t.Context(), sl, unid, preview, thumbnail)
	be.Err(t, err, nil)

	align = command.Bottom
	err = align.Thumbs(t.Context(), sl, unid, preview, thumbnail)
	be.Err(t, err, nil)

	align = command.Left
	err = align.Thumbs(t.Context(), sl, unid, preview, thumbnail)
	be.Err(t, err, nil)

	align = command.Right
	err = align.Thumbs(t.Context(), sl, unid, preview, thumbnail)
	be.Err(t, err, nil)

	count = countFiles(t, thumbdir)
	const wantThumb = 1 // because unid is named "TEST" only 1 thumbnail should be created
	be.Equal(t, count, wantThumb)

	align = -1
	err = align.Thumbs(t.Context(), sl, unid, preview, thumbnail)
	be.Err(t, err)

	// Alignment with missing preview should return ErrNoImages without deleting existing thumbnail.
	err = command.Middle.Thumbs(t.Context(), sl, "missing-preview", preview, thumbnail)
	be.True(t, errors.Is(err, command.ErrNoImages))
	_, statErr := os.Stat(filepath.Join(thumbdir, unid+".webp"))
	be.Err(t, statErr, nil)
}

func TestCropImages(t *testing.T) {
	t.Parallel()

	thumbdir := t.TempDir()
	unid, path := setupTestDir(t, thumbdir)

	sl := testLog
	preview := dir.Directory(path)

	crop := command.SquareTop
	got := crop.Images(t.Context(), sl, unid, preview)
	be.Err(t, got, nil)

	crop = command.FourThree
	got = crop.Images(t.Context(), sl, unid, preview)
	be.Err(t, got, nil)

	crop = command.OneTwo
	got = crop.Images(t.Context(), sl, unid, preview)
	be.Err(t, got, nil)

	crop = -1
	got = crop.Images(t.Context(), sl, unid, preview)
	be.Err(t, got)
}

func TestPictureImager(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		file      string
		wantBytes int64
	}{
		{"bmp", "TEST.BMP", 750054},
		{"gif", "TEST.GIF", 2646},
		{"jpg", "TEST.JPG", 16461},
		{"pcx", "TEST.PCX", 29530},
		{"png", "TEST.PNG", 4163},
		{"webp", "TEST.WEBP", 2768},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			prevdir := t.TempDir()
			thumbdir := t.TempDir()
			dirs := command.Dirs{
				Preview:   dir.Directory(prevdir),
				Thumbnail: dir.Directory(thumbdir),
			}

			src, err := filepath.Abs(filepath.Join(testdata, tt.file))
			be.Err(t, err, nil)

			unid := uuid.New().String()
			be.Err(t, dirs.PictureImager(t.Context(), testLog, src, unid), nil)

			// Confirm the source fixture was not modified.
			srcSt, err := os.Stat(src)
			be.Err(t, err, nil)
			be.Equal(t, srcSt.Size(), tt.wantBytes)

			// Confirm a non-empty preview and WebP thumbnail were generated.
			prevEntries, err := os.ReadDir(prevdir)
			be.Err(t, err, nil)
			be.True(t, len(prevEntries) > 0)
			for _, e := range prevEntries {
				info, statErr := e.Info()
				be.Err(t, statErr, nil)
				be.True(t, info.Size() > 0)
			}

			thumbSt, err := os.Stat(filepath.Join(thumbdir, unid+".webp"))
			be.Err(t, err, nil)
			be.True(t, thumbSt.Size() > 0)
		})
	}
}

func TestCropText(t *testing.T) {
	t.Parallel()

	src := testutil.FileData("testascii").Abs()
	srcSt, err := os.Stat(src)
	be.Err(t, err, nil)

	const srcSize = 440
	be.True(t, srcSt.Size() >= srcSize)

	sl := testLog
	id := uuid.New()
	cleanupCrop := func(dst string) {
		if dst != "" && dst != src {
			t.Cleanup(func() { _ = dir.RemoveAll(filepath.Dir(dst)) })
		}
	}

	txt := command.Text{UUID: id.String()}
	dst, err := txt.Crop(sl, src)
	be.Err(t, err, nil)
	cleanupCrop(dst)

	dstSt, err := os.Stat(dst)
	be.Err(t, err, nil)

	wants := int64(440)
	be.True(t, dstSt.Size() >= wants)

	txt.MaxRows = 1
	txt.MaxCols = 1
	dst, err = txt.Crop(sl, src)
	be.Err(t, err, nil)
	cleanupCrop(dst)
	dstSt, err = os.Stat(dst)
	be.Err(t, err, nil)
	wants = int64(2)
	be.Equal(t, dstSt.Size(), wants)

	txt.MaxRows = 0
	txt.MaxCols = 40
	dst, err = txt.Crop(sl, src)
	be.Err(t, err, nil)
	cleanupCrop(dst)
	dstSt, err = os.Stat(dst)
	be.Err(t, err, nil)
	wants = int64(246)
	be.Equal(t, dstSt.Size(), wants)

	txt.MaxRows = 2
	txt.MaxCols = 80
	dst, err = txt.Crop(sl, src)
	be.Err(t, err, nil)
	cleanupCrop(dst)
	dstSt, err = os.Stat(dst)
	be.Err(t, err, nil)
	wants = int64(150)
	be.True(t, dstSt.Size() >= wants)
}

func TestTextVGA(t *testing.T) {
	t.Parallel()
	requireTool(t, command.Ansilove)

	prevdir := t.TempDir()
	thumbdir := t.TempDir()
	dirs := command.Dirs{
		Preview:   dir.Directory(prevdir),
		Thumbnail: dir.Directory(thumbdir),
	}

	sl := testLog
	id := uuid.New()
	unid := id.String()
	src := testutil.FileData("testascii").Abs()

	const amigaFont = false
	got := dirs.TextImager(t.Context(), sl, "", unid, amigaFont)
	be.Err(t, got)

	got = dirs.TextImager(t.Context(), sl, src, "", amigaFont)
	be.Err(t, got)

	got = dirs.TextImager(t.Context(), sl, src, unid, amigaFont)
	be.Err(t, got, nil)

	// check for the preview image
	name := filepath.Join(prevdir, unid+".png")
	pst, got := os.Stat(name)
	be.Err(t, got, nil)
	const pstSize = 2400
	be.True(t, pst.Size() >= pstSize)

	// check for the thumbnail
	name = filepath.Join(thumbdir, unid+".webp")
	tst, got := os.Stat(name)
	be.Err(t, got, nil)
	const tstSize = 2746
	be.True(t, tst.Size() >= tstSize)
}

func TestTextAmiga(t *testing.T) {
	t.Parallel()
	requireTool(t, command.Ansilove)

	prevdir := t.TempDir()
	thumbdir := t.TempDir()
	dirs := command.Dirs{
		Preview:   dir.Directory(prevdir),
		Thumbnail: dir.Directory(thumbdir),
	}

	sl := testLog
	id := uuid.New()
	unid := id.String()
	src := testutil.FileData("testascii").Abs()

	const amigaFont = true
	got := dirs.TextImager(t.Context(), sl, "", unid, amigaFont)
	be.Err(t, got)
	got = dirs.TextImager(t.Context(), sl, src, "", amigaFont)
	be.Err(t, got)
	got = dirs.TextImager(t.Context(), sl, src, unid, amigaFont)
	be.Err(t, got, nil)

	// check for the preview image
	name := filepath.Join(prevdir, unid+".png")
	pst, got := os.Stat(name)
	be.Err(t, got, nil)
	const pstSize = 1232
	be.True(t, pst.Size() >= pstSize)

	// check for the thumbnail
	name = filepath.Join(thumbdir, unid+".webp")
	tst, got := os.Stat(name)
	be.Err(t, got, nil)
	const tstSize = 1854
	be.True(t, tst.Size() >= tstSize)
}

func TestTextUTF8(t *testing.T) {
	t.Parallel()

	prevdir := t.TempDir()
	thumbdir := t.TempDir()
	dirs := command.Dirs{
		Preview:   dir.Directory(prevdir),
		Thumbnail: dir.Directory(thumbdir),
	}

	sl := testLog
	unid := uuid.New().String()
	src := testutil.FileData("testuni").Abs()

	got := dirs.TextImager(t.Context(), sl, src, unid, false)
	be.Err(t, got, nil)

	// check for the preview image
	name := filepath.Join(prevdir, unid+".png")
	pst, got := os.Stat(name)
	be.Err(t, got, nil)
	be.True(t, pst.Size() > 0)

	// check for the thumbnail
	name = filepath.Join(thumbdir, unid+".webp")
	tst, got := os.Stat(name)
	be.Err(t, got, nil)
	be.True(t, tst.Size() > 0)
}

func TestTextAnim(t *testing.T) {
	t.Parallel()

	prevdir := t.TempDir()
	thumbdir := t.TempDir()
	dirs := command.Dirs{
		Preview:   dir.Directory(prevdir),
		Thumbnail: dir.Directory(thumbdir),
	}

	sl := testLog
	unid := uuid.New().String()
	src := testutil.FileData("maximum3ans").Abs()

	got := dirs.TextImager(t.Context(), sl, src, unid, false)
	be.Err(t, got, nil)

	// check for the animated preview image (.webp)
	name := filepath.Join(prevdir, unid+".webp")
	pst, got := os.Stat(name)
	be.Err(t, got, nil)
	be.True(t, pst.Size() > 0)

	// check for the thumbnail (.webp)
	name = filepath.Join(thumbdir, unid+".webp")
	tst, got := os.Stat(name)
	be.Err(t, got, nil)
	be.True(t, tst.Size() > 0)
}

func TestTextBlink(t *testing.T) {
	t.Parallel()

	prevdir := t.TempDir()
	thumbdir := t.TempDir()
	dirs := command.Dirs{
		Preview:   dir.Directory(prevdir),
		Thumbnail: dir.Directory(thumbdir),
	}

	sl := testLog
	unid := uuid.New().String()
	blinkSrc := filepath.Join(t.TempDir(), "blink.ans")
	be.Err(t, os.WriteFile(blinkSrc, []byte("\x1b[5;31mHello World\x1b[0m\r\n"), 0o600), nil)

	got := dirs.TextImager(t.Context(), sl, blinkSrc, unid, false)
	be.Err(t, got, nil)

	// check for the animated preview image (.webp) and absence of ansilove .png
	name := filepath.Join(prevdir, unid+".webp")
	pst, err := os.Stat(name)
	be.Err(t, err, nil)
	be.True(t, pst.Size() > 0)

	_, err = os.Stat(filepath.Join(prevdir, unid+".png"))
	be.True(t, os.IsNotExist(err))

	// check for the thumbnail (.webp)
	name = filepath.Join(thumbdir, unid+".webp")
	tst, err := os.Stat(name)
	be.Err(t, err, nil)
	be.True(t, tst.Size() > 0)
}

func TestANSImation(t *testing.T) {
	t.Parallel()

	prevdir := t.TempDir()
	thumbdir := t.TempDir()
	dirs := command.Dirs{
		Preview:   dir.Directory(prevdir),
		Thumbnail: dir.Directory(thumbdir),
	}

	sl := testLog
	unid := uuid.New().String()
	src := testutil.FileData("maximum3ans").Abs()

	got := dirs.ANSImation(t.Context(), sl, "", unid)
	be.Err(t, got)

	got = dirs.ANSImation(t.Context(), sl, src, "")
	be.Err(t, got)

	got = dirs.ANSImation(t.Context(), sl, src, unid)
	be.Err(t, got, nil)

	// check preview
	pName := filepath.Join(prevdir, unid+".webp")
	st, err := os.Stat(pName)
	be.Err(t, err, nil)
	be.True(t, st.Size() > 0)

	// check thumbnail
	tName := filepath.Join(thumbdir, unid+".webp")
	st, err = os.Stat(tName)
	be.Err(t, err, nil)
	be.True(t, st.Size() > 0)
}

func TestOptimizePNG(t *testing.T) {
	t.Parallel()

	sl := testLog

	got := command.OptimizePNG(t.Context(), sl, "")
	be.Err(t, got)

	got = command.OptimizePNG(t.Context(), sl, invalid)
	be.Err(t, got)

	name, dirs := setupTestDir(t, t.TempDir())
	notImg, got := filepath.Abs(filepath.Join(dirs, name+".ASCII"))
	be.Err(t, got, nil)
	got = command.OptimizePNG(t.Context(), sl, notImg)
	be.Err(t, got)

	png, got := filepath.Abs(filepath.Join(dirs, name+".PNG"))
	be.Err(t, got, nil)

	got = command.OptimizePNG(t.Context(), sl, png)
	be.Err(t, got, nil)
}

func TestTextDeferred(t *testing.T) {
	t.Parallel()
	requireTool(t, command.Ansilove)

	extradir := t.TempDir()
	prevdir := t.TempDir()
	thumbdir := t.TempDir()
	dirs := command.Dirs{
		Extra:     dir.Directory(extradir),
		Preview:   dir.Directory(prevdir),
		Thumbnail: dir.Directory(thumbdir),
	}

	sl := testLog
	id := uuid.New()
	unid := id.String()
	src := testutil.FileData("testascii").Abs()

	got := dirs.TextDeferred(t.Context(), sl, "", "")
	be.Err(t, got)
	got = dirs.TextDeferred(t.Context(), sl, src, "")
	be.Err(t, got)
	got = dirs.TextDeferred(t.Context(), sl, "", unid)
	be.Err(t, got)
	got = dirs.TextDeferred(t.Context(), sl, src, unid)
	be.Err(t, got, nil)

	// NOTE: we test be.True >= instead of be.Equal because different
	// platforms and OSes using the same image tools give varied outputs.

	// check for the preview
	name := filepath.Join(prevdir, unid+".png")
	pst, got := os.Stat(name)
	be.Err(t, got, nil)
	const pstSize = 2400
	be.True(t, pst.Size() >= pstSize)

	// check for the thumbnail
	name = filepath.Join(thumbdir, unid+".webp")
	tst, got := os.Stat(name)
	be.Err(t, got, nil)
	const tstSize = 2746
	be.True(t, tst.Size() >= tstSize)

	// confirm the text was copied to the extra directory
	name = filepath.Join(extradir, unid+".txt")
	est, got := os.Stat(name)
	be.Err(t, got, nil)
	const estSize = 440
	be.True(t, est.Size() >= estSize)
}

func TestThumbs(t *testing.T) {
	t.Parallel()

	dirs := command.Dirs{}
	sl := testLog
	got := dirs.Thumbs(t.Context(), sl, "", -1)
	be.Err(t, got)

	unid, prevdir := setupTestDir(t, t.TempDir())

	extradir := t.TempDir()
	thumbdir := t.TempDir()
	dirs = command.Dirs{
		Extra:     dir.Directory(extradir),
		Preview:   dir.Directory(prevdir),
		Thumbnail: dir.Directory(thumbdir),
	}
	got = dirs.Thumbs(t.Context(), sl, unid, command.Photo)
	be.Err(t, got, nil)

	count := countFiles(t, thumbdir)
	be.Equal(t, count, 1)
	name := filepath.Join(thumbdir, unid+".webp")
	st, got := os.Stat(name)
	be.Err(t, got, nil)
	be.True(t, st.Size() > 0)
	be.True(t, st.Size() < 10_000)
}

func TestThumbsModes(t *testing.T) {
	t.Parallel()

	prevdir := t.TempDir()
	thumbdir := t.TempDir()
	sl := testLog
	unid := uuid.New().String()
	src := filepath.Join(prevdir, unid+".png")
	testutil.CopyPNG(t, src)

	dirs := command.Dirs{
		Preview:   dir.Directory(prevdir),
		Thumbnail: dir.Directory(thumbdir),
	}

	got := dirs.Thumbs(t.Context(), sl, unid, command.Pixel)
	be.Err(t, got, nil)
	pixelSt, err := os.Stat(filepath.Join(thumbdir, unid+".webp"))
	be.Err(t, err, nil)
	be.True(t, pixelSt.Size() > 0)
	be.True(t, pixelSt.Size() < 45_000)

	got = dirs.Thumbs(t.Context(), sl, unid, command.Photo)
	be.Err(t, got, nil)
	photoSt, err := os.Stat(filepath.Join(thumbdir, unid+".webp"))
	be.Err(t, err, nil)
	be.True(t, photoSt.Size() > 0)
	be.True(t, photoSt.Size() < 25_000)

	// Verify multi-frame preview (e.g. animated WebP) succeeds in Photo mode.
	animID := uuid.New().String()
	animSrc := testutil.FileData("maximum3ans").Abs()
	be.Err(t, dirs.ANSImation(t.Context(), sl, animSrc, animID), nil)
	be.Err(t, dirs.Thumbs(t.Context(), sl, animID, command.Photo), nil)
	animThumbSt, err := os.Stat(filepath.Join(thumbdir, animID+".webp"))
	be.Err(t, err, nil)
	be.True(t, animThumbSt.Size() > 0)
}

func TestAlign(t *testing.T) {
	t.Parallel()

	prevdir := t.TempDir()
	thumbdir := t.TempDir()
	unid, _ := setupTestDir(t, prevdir)

	sl := testLog
	got := command.Top.Thumbs(t.Context(), sl, "", "", "")
	be.Err(t, got)

	preview := dir.Directory(prevdir)
	thumbnail := dir.Directory(thumbdir)
	got = command.Top.Thumbs(t.Context(), sl, unid, preview, thumbnail)
	be.Err(t, got, nil)
}

func TestCrop(t *testing.T) {
	t.Parallel()

	sl := testLog
	got := command.OneTwo.Images(t.Context(), sl, "", "")
	be.Err(t, got)

	wd, got := os.Getwd()
	be.Err(t, got, nil)
	got = command.OneTwo.Images(t.Context(), sl, "", dir.Directory(wd))
	be.Err(t, got)

	prevdir := t.TempDir()
	oldpath := filepath.Join(testdata, "TEST.PNG")
	newpath := filepath.Join(prevdir, "TEST.PNG")
	n, got := helper.Duplicate(oldpath, newpath)
	be.Err(t, got, nil)
	be.Equal(t, n, 4163)

	preview := dir.Directory(prevdir)
	unid := "TEST"
	got = command.OneTwo.Images(t.Context(), sl, unid, preview)
	be.Err(t, got, nil)

	name := filepath.Join(prevdir, unid+".PNG")
	st, got := os.Stat(name)
	be.Err(t, got, nil)
	be.True(t, st.Size() >= 1940)

	got = command.OneTwo.Images(t.Context(), sl, unid, preview)
	be.Err(t, got, nil)
	st, got = os.Stat(name)
	be.Err(t, got, nil)
	be.True(t, st.Size() >= 1940)
}
