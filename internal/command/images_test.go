package command_test

import (
	"os"
	"testing"

	"github.com/Defacto2/server/internal/command"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestImages(t *testing.T) {
	t.Parallel()
	err := command.ImagesDelete("", "")
	require.Error(t, err)
	err = command.ImagesPixelate("", "")
	require.Error(t, err)
}

func TestPixelate(t *testing.T) {
	t.Parallel()
	a := command.Args{}
	a.Pixelate()
	assert.Contains(t, a, "-scale")
	assert.Contains(t, a, "5%")
}

func TestThumbs(t *testing.T) {
	t.Parallel()
	dir := command.Dirs{}
	err := dir.Thumbs("", -1)
	require.Error(t, err)
}

func TestAlign(t *testing.T) {
	t.Parallel()
	err := command.Top.Thumbs("", "", "")
	require.Error(t, err)
}

func TestCrop(t *testing.T) {
	t.Parallel()
	err := command.OneTwo.Images("", "")
	require.Error(t, err)
	dir, err := os.Getwd()
	require.NoError(t, err)
	err = command.OneTwo.Images("", dir)
	require.NoError(t, err)
}

func TestDirs(t *testing.T) {
	t.Parallel()
	dir := command.Dirs{}
	err := dir.PictureImager(nil, "", "")
	require.Error(t, err)
	err = dir.TextImager(nil, "", "")
	require.Error(t, err)
	err = dir.TextAmigaImager(nil, "", "")
	require.Error(t, err)
	err = dir.PreviewPhoto(nil, "", "")
	require.Error(t, err)
	err = dir.PreviewGIF(nil, "", "")
	require.Error(t, err)
	err = dir.PreviewPNG(nil, "", "")
	require.Error(t, err)
	err = dir.PreviewWebP(nil, "", "")
	require.Error(t, err)
	err = dir.ThumbPixels("", "")
	require.Error(t, err)
	err = dir.ThumbPhoto("", "")
	require.Error(t, err)
}

func TestArgs(t *testing.T) {
	t.Parallel()
	a := command.Args{}
	a.Topx400()
	assert.Contains(t, a, "-gravity")
	assert.Contains(t, a, "North")
	a.Middlex400()
	assert.Contains(t, a, "-gravity")
	assert.Contains(t, a, "center")
	a.Bottomx400()
	assert.Contains(t, a, "-gravity")
	assert.Contains(t, a, "South")
	a.Leftx400()
	assert.Contains(t, a, "-gravity")
	assert.Contains(t, a, "West")
	a.Rightx400()
	assert.Contains(t, a, "-gravity")
	assert.Contains(t, a, "East")
	a.CropTop()
	assert.Contains(t, a, "-gravity")
	assert.Contains(t, a, "North")
	a = command.Args{}
	a.FourThree()
	assert.Contains(t, a, "-gravity")
	assert.Contains(t, a, "North")
	a = command.Args{}
	a.OneTwo()
	assert.Contains(t, a, "-gravity")
	assert.Contains(t, a, "North")
	a = command.Args{}
	a.AnsiAmiga()
	assert.Contains(t, a, "topaz+")
	a = command.Args{}
	a.AnsiMsDos()
	assert.Contains(t, a, "80x25")
	a = command.Args{}
	a.JpegPhoto()
	assert.Contains(t, a, "-gaussian-blur")
	a = command.Args{}
	a.PortablePixel()
	assert.Contains(t, a, "png:compression-filter=5")
	a = command.Args{}
	a.Thumbnail()
	assert.Contains(t, a, "#999")
	a = command.Args{}
	a.CWebp()
	assert.Contains(t, a, "-exact")
	a = command.Args{}
	a.CWebpText()
	assert.Contains(t, a, "text")
	a = command.Args{}
	a.GWebp()
	assert.Contains(t, a, "-mt")
}

func TestOptimizePNG(t *testing.T) {
	t.Parallel()
	err := command.OptimizePNG("")
	require.NoError(t, err)
}
