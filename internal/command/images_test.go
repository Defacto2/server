package command_test

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Defacto2/server/internal/command"
	"github.com/Defacto2/server/internal/dir"
	"github.com/Defacto2/server/internal/logs"
	"github.com/google/uuid"
	"github.com/nalgeon/be"
)

const (
	imageCount = 7
	invalid    = "this-is#invalid!"
)

// copyTestData copies all files from testdata/ into a newly created t.TempDir()
// and returns the base file name and path to the temporary directory.
func setupTestDir(t *testing.T) (string, string) {
	t.Helper()
	dstDir := t.TempDir()
	srcDir := "testdata"
	baseN := ""

	entries, err := os.ReadDir(srcDir)
	if err != nil {
		t.Fatalf("failed to read testdata directory: %v", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		srcPath := filepath.Join(srcDir, entry.Name())
		dstPath := filepath.Join(dstDir, strings.ToLower(entry.Name()))
		if baseN == "" {
			baseN = filepath.Base(dstPath)
			baseN = strings.TrimSuffix(baseN, path.Ext(baseN))
		}

		data, err := os.ReadFile(srcPath)
		if err != nil {
			t.Fatalf("failed to read fixture %s: %v", srcPath, err)
		}

		if err := os.WriteFile(dstPath, data, 0o644); err != nil {
			t.Fatalf("failed to write fixture to temp dir %s: %v", dstPath, err)
		}
	}

	return baseN, dstDir
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
		if !entry.IsDir() {
			count++
		}
	}
	return count
}

func TestImagesDelete(t *testing.T) {
	t.Parallel()
	unid, dir := setupTestDir(t)
	count := countFiles(t, dir)
	be.Equal(t, count, imageCount)

	err := command.ImagesDelete(unid, dir)
	be.Err(t, err, nil)

	// wantCnt is 3 as the TEST.BMP, TEST.PCX, TEST. are ignored by ImagesDelete
	const wantCnt = 3
	got := countFiles(t, dir)
	be.Equal(t, got, wantCnt)

	err = command.ImagesDelete("", dir)
	be.Err(t, err)
	err = command.ImagesDelete(unid, "")
	be.Err(t, err)
}

func TestImagesPixelate(t *testing.T) {
	t.Parallel()
	unid, dir := setupTestDir(t)
	count := countFiles(t, dir)
	be.Equal(t, count, imageCount)

	err := command.ImagesPixelate(t.Context(), unid, dir)
	be.Err(t, err, nil)

	img := filepath.Join(dir, unid+".png")
	original, err := filepath.Abs(filepath.Join("testdata", "TEST.PNG"))
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

	// not found images including empty unid values are skipped
	err = command.ImagesPixelate(t.Context(), "", dir)
	be.Err(t, err, nil)
	err = command.ImagesPixelate(t.Context(), unid, "")
	be.Err(t, err)
}

func TestDirsThumbs(t *testing.T) {
	t.Parallel()
	unid, path := setupTestDir(t)
	count := countFiles(t, path)
	be.Equal(t, count, imageCount)

	thumbdir := t.TempDir()
	dirs := command.Dirs{
		Preview:   dir.Directory(path),
		Thumbnail: dir.Directory(thumbdir),
	}

	count = countFiles(t, thumbdir)
	be.Equal(t, count, 0)

	sl := slog.Default()
	err := dirs.Thumbs(t.Context(), sl, unid, command.Pixel)
	be.Err(t, err, nil)
	count = countFiles(t, thumbdir)
	const wantThumb = 1 // because unid is named "TEST" only 1 thumbnail should be created
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
	unid, path := setupTestDir(t)
	count := countFiles(t, path)
	be.Equal(t, count, imageCount)

	thumbdir := t.TempDir()
	preview := dir.Directory(path)
	thumbnail := dir.Directory(thumbdir)

	count = countFiles(t, thumbdir)
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
	unid, path := setupTestDir(t)
	count := countFiles(t, path)
	be.Equal(t, count, imageCount)

	thumbdir := t.TempDir()
	preview := dir.Directory(path)

	count = countFiles(t, thumbdir)
	be.Equal(t, count, 0)

	sl := slog.Default()

	crop := command.SquareTop
	err := crop.Images(t.Context(), sl, unid, preview)
	be.Err(t, err, nil)

	crop = command.FourThree
	err = crop.Images(t.Context(), sl, unid, preview)
	be.Err(t, err, nil)

	crop = command.OneTwo
	err = crop.Images(t.Context(), sl, unid, preview)
	be.Err(t, err, nil)

	crop = -1
	err = crop.Images(t.Context(), sl, unid, preview)
	be.Err(t, err)
}

func TestPictureImager(t *testing.T) {
	prevdir := t.TempDir()
	thumbdir := t.TempDir()
	dirs := command.Dirs{
		Preview:   dir.Directory(prevdir),
		Thumbnail: dir.Directory(thumbdir),
	}
	sl := slog.Default()
	id := uuid.New()
	unid := id.String()

	bmp, err := filepath.Abs(filepath.Join("testdata", "TEST.BMP"))
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

	thbImg := filepath.Join(thumbdir, unid+".webp")
	thbSt, err := os.Stat(thbImg)
	be.Err(t, err, nil)
	thbSize := thbSt.Size()
	const thbBytes = 2664 // this could change depending on the tool set?
	be.Equal(t, thbSize, thbBytes)

	gif, err := filepath.Abs(filepath.Join("testdata", "TEST.GIF"))
	be.Err(t, err, nil)
	err = dirs.PictureImager(t.Context(), sl, gif, unid)
	be.Err(t, err, nil)
	gifSt, err := os.Stat(gif)
	be.Err(t, err, nil)
	gifSize := gifSt.Size()
	const gifBytes = 2646
	be.Equal(t, gifSize, gifBytes)

	jpg, err := filepath.Abs(filepath.Join("testdata", "TEST.JPG"))
	be.Err(t, err, nil)
	err = dirs.PictureImager(t.Context(), sl, jpg, unid)
	be.Err(t, err, nil)
	jpgSt, err := os.Stat(jpg)
	be.Err(t, err, nil)
	jpgSize := jpgSt.Size()
	const jpgBytes = 16461
	be.Equal(t, jpgSize, jpgBytes)

	pcx, err := filepath.Abs(filepath.Join("testdata", "TEST.PCX"))
	be.Err(t, err, nil)
	err = dirs.PictureImager(t.Context(), sl, pcx, unid)
	be.Err(t, err, nil)
	pcxSt, err := os.Stat(pcx)
	be.Err(t, err, nil)
	pcxSize := pcxSt.Size()
	const pcxBytes = 29530
	be.Equal(t, pcxSize, pcxBytes)

	png, err := filepath.Abs(filepath.Join("testdata", "TEST.PNG"))
	be.Err(t, err, nil)
	err = dirs.PictureImager(t.Context(), sl, png, unid)
	be.Err(t, err, nil)
	pngSt, err := os.Stat(png)
	be.Err(t, err, nil)
	pngSize := pngSt.Size()
	const pngBytes = 4163
	be.Equal(t, pngSize, pngBytes)

	web, err := filepath.Abs(filepath.Join("testdata", "TEST.WEBP"))
	be.Err(t, err, nil)
	err = dirs.PictureImager(t.Context(), sl, web, unid)
	be.Err(t, err, nil)
	webSt, err := os.Stat(web)
	be.Err(t, err, nil)
	webSize := webSt.Size()
	const webBytes = 2768
	be.Equal(t, webSize, webBytes)
	// because previewweb has a special makeThumb flag, test for the thumbnail generation
	// also, sometimes the thumbnails are larger in file size than the source image
	thbImg = filepath.Join(thumbdir, unid+".webp")
	thbSt, err = os.Stat(thbImg)
	be.Err(t, err, nil)
	thbSize = thbSt.Size()
	const wtBytes = 8772 // this could change depending on the tool set?
	be.Equal(t, thbSize, wtBytes)
}

func TestCropText(t *testing.T) {
	t.Parallel()
	_, textdir := setupTestDir(t)
	src := filepath.Join(textdir, "test.ascii")
	srcSt, err := os.Stat(src)
	be.Err(t, err, nil)
	const srcSize = 931
	be.Equal(t, srcSt.Size(), srcSize)

	sl := slog.Default()
	id := uuid.New()
	unid := id.String()
	dst, err := command.CropText(sl, 0, 0, src, unid)
	be.Err(t, err, nil)
	dstSt, err := os.Stat(dst)
	be.Err(t, err, nil)
	wants := int64(481)
	be.Equal(t, dstSt.Size(), wants)

	dst, err = command.CropText(sl, 1, 1, src, unid)
	be.Err(t, err, nil)
	dstSt, err = os.Stat(dst)
	be.Err(t, err, nil)
	wants = int64(2)
	be.Equal(t, dstSt.Size(), wants)

	dst, err = command.CropText(sl, 40, 0, src, unid)
	be.Err(t, err, nil)
	dstSt, err = os.Stat(dst)
	be.Err(t, err, nil)
	wants = int64(246)
	be.Equal(t, dstSt.Size(), wants)

	dst, err = command.CropText(sl, 80, 2, src, unid)
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
	src, err := filepath.Abs(filepath.Join("testdata", "TEST.ASCII"))
	be.Err(t, err, nil)

	const amigaFont = false
	err = dirs.TextImager(t.Context(), sl, "", unid, amigaFont)
	be.Err(t, err)
	err = dirs.TextImager(t.Context(), sl, src, "", amigaFont)
	be.Err(t, err)
	err = dirs.TextImager(t.Context(), sl, src, unid, amigaFont)
	be.Err(t, err, nil)
	// check for the preview image
	name := filepath.Join(prevdir, unid+".png")
	pst, err := os.Stat(name)
	be.Err(t, err, nil)
	const pstSize = 2421
	be.Equal(t, pst.Size(), pstSize)
	// check for the thumbnail
	name = filepath.Join(thumbdir, unid+".webp")
	tst, err := os.Stat(name)
	be.Err(t, err, nil)
	const tstSize = 2784
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
	src, err := filepath.Abs(filepath.Join("testdata", "TEST.ASCII"))
	be.Err(t, err, nil)

	const amigaFont = true
	err = dirs.TextImager(t.Context(), sl, "", unid, amigaFont)
	be.Err(t, err)
	err = dirs.TextImager(t.Context(), sl, src, "", amigaFont)
	be.Err(t, err)
	err = dirs.TextImager(t.Context(), sl, src, unid, amigaFont)
	be.Err(t, err, nil)
	// check for the preview image
	name := filepath.Join(prevdir, unid+".png")
	pst, err := os.Stat(name)
	be.Err(t, err, nil)
	const pstSize = 1232
	be.Equal(t, pst.Size(), pstSize)
	// check for the thumbnail
	name = filepath.Join(thumbdir, unid+".webp")
	tst, err := os.Stat(name)
	be.Err(t, err, nil)
	const tstSize = 2738
	be.Equal(t, tst.Size(), tstSize)
}

func TestPixelate(t *testing.T) {
	t.Parallel()
	a := command.Args{}
	a.Pixelate()
	s := fmt.Sprintf("%+v", a)
	find := strings.Contains(s, "-scale")
	be.True(t, find)
	s = fmt.Sprintf("%+v", a)
	find = strings.Contains(s, "5%")
	be.True(t, find)
}

func TestThumbs(t *testing.T) {
	t.Parallel()
	dir := command.Dirs{}
	d := logs.Discard()
	err := dir.Thumbs(context.TODO(), d, "", -1)
	be.Err(t, err)
	err = dir.Thumbs(context.TODO(), d, "", command.Photo)
	be.Err(t, err)
}

func TestAlign(t *testing.T) {
	t.Parallel()
	err := command.Top.Thumbs(context.TODO(), nil, "", "", "")
	be.Err(t, err)
}

func TestCrop(t *testing.T) {
	t.Parallel()
	d := logs.Discard()
	err := command.OneTwo.Images(context.TODO(), d, "", "")
	be.Err(t, err)
	wd, err := os.Getwd()
	be.Err(t, err, nil)
	err = command.OneTwo.Images(context.TODO(), d, "", dir.Directory(wd))
	be.Err(t, err)
}

func TestArgs(t *testing.T) {
	t.Parallel()
	a := command.Args{}
	a.Topx400()
	s := fmt.Sprintf("%+v", a)
	find := strings.Contains(s, "-gravity")
	be.True(t, find)
	find = strings.Contains(s, "North")
	a.Middlex400()
	be.True(t, find)
	s = fmt.Sprintf("%+v", a)
	find = strings.Contains(s, "-gravity")
	be.True(t, find)
	find = strings.Contains(s, "center")
	be.True(t, find)
	a.Bottomx400()
	s = fmt.Sprintf("%+v", a)
	find = strings.Contains(s, "-gravity")
	be.True(t, find)
	find = strings.Contains(s, "South")
	be.True(t, find)
	a.Leftx400()
	s = fmt.Sprintf("%+v", a)
	find = strings.Contains(s, "-gravity")
	be.True(t, find)
	find = strings.Contains(s, "West")
	be.True(t, find)
	a.Rightx400()
	s = fmt.Sprintf("%+v", a)
	find = strings.Contains(s, "-gravity")
	be.True(t, find)
	find = strings.Contains(s, "East")
	be.True(t, find)
	a.CropTop()
	s = fmt.Sprintf("%+v", a)
	find = strings.Contains(s, "-gravity")
	be.True(t, find)
	find = strings.Contains(s, "North")
	be.True(t, find)
	a = command.Args{}
	a.FourThree()
	s = fmt.Sprintf("%+v", a)
	find = strings.Contains(s, "-gravity")
	be.True(t, find)
	find = strings.Contains(s, "North")
	be.True(t, find)
	a = command.Args{}
	a.OneTwo()
	s = fmt.Sprintf("%+v", a)
	find = strings.Contains(s, "-gravity")
	be.True(t, find)
	find = strings.Contains(s, "North")
	be.True(t, find)
	a = command.Args{}
	a.AnsiAmiga()
	s = fmt.Sprintf("%+v", a)
	find = strings.Contains(s, "topaz+")
	be.True(t, find)
	a = command.Args{}
	a.AnsiMsDos()
	s = fmt.Sprintf("%+v", a)
	find = strings.Contains(s, "80x50")
	be.True(t, find)
	a = command.Args{}
	a.JpegPhoto()
	s = fmt.Sprintf("%+v", a)
	find = strings.Contains(s, "75")
	be.True(t, find)
	a = command.Args{}
	a.PortablePixel()
	s = fmt.Sprintf("%+v", a)
	find = strings.Contains(s, "png:compression-filter=5")
	be.True(t, find)
	a = command.Args{}
	a.Thumbnail()
	s = fmt.Sprintf("%+v", a)
	find = strings.Contains(s, "#999")
	be.True(t, find)
	a = command.Args{}
	a.CWebp()
	s = fmt.Sprintf("%+v", a)
	find = strings.Contains(s, "-exact")
	be.True(t, find)
	a = command.Args{}
	a.CWebpText()
	s = fmt.Sprintf("%+v", a)
	find = strings.Contains(s, "text")
	be.True(t, find)
	a = command.Args{}
	a.GWebp()
	s = fmt.Sprintf("%+v", a)
	find = strings.Contains(s, "-mt")
	be.True(t, find)
}

func TestOptimizePNG(t *testing.T) {
	t.Parallel()
	err := command.OptimizePNG(t.Context(), "")
	be.Err(t, err)
	err = command.OptimizePNG(t.Context(), invalid)
	be.Err(t, err)

	name, dirs := setupTestDir(t)
	bmp, err := filepath.Abs(filepath.Join(dirs, name+".bmp"))
	be.Err(t, err, nil)
	err = command.OptimizePNG(t.Context(), bmp)
	be.Err(t, err)
	png, err := filepath.Abs(filepath.Join(dirs, name+".png"))
	be.Err(t, err, nil)
	err = command.OptimizePNG(t.Context(), png)
	be.Err(t, err, nil)
}
