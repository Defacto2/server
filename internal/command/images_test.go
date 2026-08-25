package command_test

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/Defacto2/helper"
	"github.com/Defacto2/server/internal/command"
	"github.com/Defacto2/server/internal/dir"
	"github.com/Defacto2/server/internal/logs"
	"github.com/google/uuid"
	"github.com/nalgeon/be"
)

const (
	testdata      = "testdata"
	testdataCount = 9
	testCount     = 7
	invalid       = "this-is#invalid!"
)

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
		fmt.Println(e.Name(), strings.HasPrefix(e.Name(), "TEST."))
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
	if n != testCount {
		t.Fatalf("found %d test files in the temp directory, wanted %d: %s", n, testCount, tempDir)
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
		if entry.Type().IsRegular() {
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

	const want = 3 // want 3 as ".ascii", ".bmp", ".pcx" are ignored by the deleter

	got := countFiles(t, dir)
	be.Equal(t, got, want)

	err = command.ImagesDelete("", dir)
	be.Err(t, err)

	err = command.ImagesDelete(unid, "")
	be.Err(t, err)
}

func TestImagesPixelate(t *testing.T) {
	t.Parallel()

	sl := slog.Default()

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

	sl := slog.Default()
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
}

func TestAlignThumbs(t *testing.T) {
	t.Parallel()

	unid, prevdir := setupTestDir(t, t.TempDir())
	thumbdir := t.TempDir()
	preview := dir.Directory(prevdir)
	thumbnail := dir.Directory(thumbdir)

	count := countFiles(t, thumbdir)
	be.Equal(t, count, 0)

	sl := slog.Default()

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
}

func TestCropImages(t *testing.T) {
	t.Parallel()

	thumbdir := t.TempDir()
	unid, path := setupTestDir(t, thumbdir)

	sl := slog.Default()
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

	prevdir := t.TempDir()
	thumbdir := t.TempDir()

	dirs := command.Dirs{
		Preview:   dir.Directory(prevdir),  // previews will be generated here
		Thumbnail: dir.Directory(thumbdir), // thumbnails will be generated here
	}

	sl := slog.Default()
	id := uuid.New()
	unid := id.String()

	bmp, err := filepath.Abs(filepath.Join(testdata, "TEST.BMP"))
	be.Err(t, err, nil)
	err = dirs.PictureImager(t.Context(), sl, bmp, unid)
	be.Err(t, err, nil)

	bmpSt, err := os.Stat(bmp)
	be.Err(t, err, nil)
	bmpSize := bmpSt.Size()
	const bmpBytes = 750054 // confirm the test files have not been modified
	be.Equal(t, bmpSize, bmpBytes)

	preImg := filepath.Join(prevdir, unid+".png")
	preSt, err := os.Stat(preImg)
	be.Err(t, err, nil)
	preSize := preSt.Size()
	const preBytes = 1629 // this could change depending on the tool set?
	be.Equal(t, preSize, preBytes)

	gif, err := filepath.Abs(filepath.Join(testdata, "TEST.GIF"))
	be.Err(t, err, nil)
	err = dirs.PictureImager(t.Context(), sl, gif, unid)
	be.Err(t, err, nil)
	gifSt, err := os.Stat(gif)
	be.Err(t, err, nil)
	gifSize := gifSt.Size()
	const gifBytes = 2646
	be.Equal(t, gifSize, gifBytes)

	jpg, err := filepath.Abs(filepath.Join(testdata, "TEST.JPG"))
	be.Err(t, err, nil)
	err = dirs.PictureImager(t.Context(), sl, jpg, unid)
	be.Err(t, err, nil)
	jpgSt, err := os.Stat(jpg)
	be.Err(t, err, nil)
	jpgSize := jpgSt.Size()
	const jpgBytes = 16461
	be.Equal(t, jpgSize, jpgBytes)

	pcx, err := filepath.Abs(filepath.Join(testdata, "TEST.PCX"))
	be.Err(t, err, nil)
	err = dirs.PictureImager(t.Context(), sl, pcx, unid)
	be.Err(t, err, nil)
	pcxSt, err := os.Stat(pcx)
	be.Err(t, err, nil)
	pcxSize := pcxSt.Size()
	const pcxBytes = 29530
	be.Equal(t, pcxSize, pcxBytes)

	png, err := filepath.Abs(filepath.Join(testdata, "TEST.PNG"))
	be.Err(t, err, nil)
	err = dirs.PictureImager(t.Context(), sl, png, unid)
	be.Err(t, err, nil)
	pngSt, err := os.Stat(png)
	be.Err(t, err, nil)
	pngSize := pngSt.Size()
	const pngBytes = 4163
	be.Equal(t, pngSize, pngBytes)

	web, err := filepath.Abs(filepath.Join(testdata, "TEST.WEBP"))
	be.Err(t, err, nil)
	err = dirs.PictureImager(t.Context(), sl, web, unid)
	be.Err(t, err, nil)
	webSt, err := os.Stat(web)
	be.Err(t, err, nil)
	webSize := webSt.Size()
	const webBytes = 2768
	be.Equal(t, webSize, webBytes)
}

func TestCropText(t *testing.T) {
	t.Parallel()

	_, textdir := setupTestDir(t, t.TempDir())
	src := filepath.Join(textdir, "TEST.ASCII")
	srcSt, err := os.Stat(src)
	be.Err(t, err, nil)
	const srcSize = 931
	be.Equal(t, srcSt.Size(), srcSize)

	sl := slog.Default()
	id := uuid.New()

	txt := command.Text{UUID: id.String()}
	dst, err := txt.Crop(sl, src)
	be.Err(t, err, nil)
	dstSt, err := os.Stat(dst)
	be.Err(t, err, nil)
	wants := int64(481)
	be.Equal(t, dstSt.Size(), wants)

	txt.MaxRows = 1
	txt.MaxCols = 1
	dst, err = txt.Crop(sl, src)
	be.Err(t, err, nil)
	dstSt, err = os.Stat(dst)
	be.Err(t, err, nil)
	wants = int64(2)
	be.Equal(t, dstSt.Size(), wants)

	txt.MaxRows = 0
	txt.MaxCols = 40
	dst, err = txt.Crop(sl, src)
	be.Err(t, err, nil)
	dstSt, err = os.Stat(dst)
	be.Err(t, err, nil)
	wants = int64(246)
	be.Equal(t, dstSt.Size(), wants)

	txt.MaxRows = 2
	txt.MaxCols = 80
	dst, err = txt.Crop(sl, src)
	be.Err(t, err, nil)
	dstSt, err = os.Stat(dst)
	be.Err(t, err, nil)
	wants = int64(162)
	be.Equal(t, dstSt.Size(), wants)
}

func TestTextImagerVgaFont(t *testing.T) {
	t.Parallel()

	prevdir := t.TempDir()
	thumbdir := t.TempDir()
	dirs := command.Dirs{
		Preview:   dir.Directory(prevdir),
		Thumbnail: dir.Directory(thumbdir),
	}

	sl := slog.Default()
	id := uuid.New()
	unid := id.String()
	src, got := filepath.Abs(filepath.Join(testdata, "TEST.ASCII"))
	be.Err(t, got, nil)

	const amigaFont = false
	got = dirs.TextImager(t.Context(), sl, "", unid, amigaFont)
	be.Err(t, got)

	got = dirs.TextImager(t.Context(), sl, src, "", amigaFont)
	be.Err(t, got)

	got = dirs.TextImager(t.Context(), sl, src, unid, amigaFont)
	be.Err(t, got, nil)

	// check for the preview image
	name := filepath.Join(prevdir, unid+".png")
	pst, got := os.Stat(name)
	be.Err(t, got, nil)
	const pstSize = 2421
	be.Equal(t, pst.Size(), pstSize)

	// check for the thumbnail
	name = filepath.Join(thumbdir, unid+".webp")
	tst, got := os.Stat(name)
	be.Err(t, got, nil)
	const tstSize = 2746
	be.Equal(t, tst.Size(), tstSize)
}

func TestTextImagerAmigaFont(t *testing.T) {
	t.Parallel()

	prevdir := t.TempDir()
	thumbdir := t.TempDir()
	dirs := command.Dirs{
		Preview:   dir.Directory(prevdir),
		Thumbnail: dir.Directory(thumbdir),
	}

	sl := slog.Default()
	id := uuid.New()
	unid := id.String()
	src, got := filepath.Abs(filepath.Join(testdata, "TEST.ASCII"))
	be.Err(t, got, nil)

	const amigaFont = true
	got = dirs.TextImager(t.Context(), sl, "", unid, amigaFont)
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
	be.Equal(t, pst.Size(), pstSize)

	// check for the thumbnail
	name = filepath.Join(thumbdir, unid+".webp")
	tst, got := os.Stat(name)
	be.Err(t, got, nil)
	const tstSize = 1854
	be.Equal(t, tst.Size(), tstSize)
}

func TestOptimizePNG(t *testing.T) {
	t.Parallel()

	sl := slog.Default()

	got := command.OptimizePNG(t.Context(), sl, "")
	be.Err(t, got)

	got = command.OptimizePNG(t.Context(), sl, invalid)
	be.Err(t, got)

	name, dirs := setupTestDir(t, t.TempDir())
	bmp, got := filepath.Abs(filepath.Join(dirs, name+".png"))
	be.Err(t, got, nil)
	got = command.OptimizePNG(t.Context(), sl, bmp)
	be.Err(t, got)

	png, got := filepath.Abs(filepath.Join(dirs, name+".PNG"))
	be.Err(t, got, nil)

	got = command.OptimizePNG(t.Context(), sl, png)
	be.Err(t, got, nil)
}

func TestTextDeferred(t *testing.T) {
	t.Parallel()

	extradir := t.TempDir()
	prevdir := t.TempDir()
	thumbdir := t.TempDir()
	dirs := command.Dirs{
		Extra:     dir.Directory(extradir),
		Preview:   dir.Directory(prevdir),
		Thumbnail: dir.Directory(thumbdir),
	}

	sl := slog.Default()
	id := uuid.New()
	unid := id.String()
	src, got := filepath.Abs(filepath.Join(testdata, "TEST.ASCII"))
	be.Err(t, got, nil)

	got = dirs.TextDeferred(t.Context(), sl, "", "")
	be.Err(t, got)
	got = dirs.TextDeferred(t.Context(), sl, src, "")
	be.Err(t, got)
	got = dirs.TextDeferred(t.Context(), sl, "", unid)
	be.Err(t, got)
	got = dirs.TextDeferred(t.Context(), sl, src, unid)
	be.Err(t, got, nil)

	// check for the preview
	name := filepath.Join(prevdir, unid+".png")
	pst, got := os.Stat(name)
	be.Err(t, got, nil)
	const pstSize = 2421
	be.Equal(t, pst.Size(), pstSize)

	// check for the thumbnail
	name = filepath.Join(thumbdir, unid+".webp")
	tst, got := os.Stat(name)
	be.Err(t, got, nil)
	const tstSize = 2746
	be.Equal(t, tst.Size(), tstSize)

	// confirm the text was copied to the extra directory
	name = filepath.Join(extradir, unid+".txt")
	est, got := os.Stat(name)
	be.Err(t, got, nil)
	const estSize = 931
	be.Equal(t, est.Size(), estSize)
}

func TestThumbs(t *testing.T) {
	t.Parallel()

	dirs := command.Dirs{}
	sl := logs.Discard()
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
	be.True(t, st.Size() > 10_000)
}

func TestAlign(t *testing.T) {
	t.Parallel()

	prevdir := t.TempDir()
	thumbdir := t.TempDir()
	unid, _ := setupTestDir(t, prevdir)

	sl := slog.Default()
	got := command.Top.Thumbs(t.Context(), sl, "", "", "")
	be.Err(t, got)

	preview := dir.Directory(prevdir)
	thumbnail := dir.Directory(thumbdir)
	got = command.Top.Thumbs(t.Context(), sl, unid, preview, thumbnail)
	be.Err(t, got, nil)
}

func TestCrop(t *testing.T) {
	t.Parallel()

	sl := slog.Default()
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
	be.Equal(t, st.Size(), 1940)

	got = command.OneTwo.Images(t.Context(), sl, unid, preview)
	be.Err(t, got, nil)
	st, got = os.Stat(name)
	be.Err(t, got, nil)
	be.Equal(t, st.Size(), 1940)
}
