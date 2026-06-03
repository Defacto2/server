package command_test

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/Defacto2/server/internal/command"
	"github.com/Defacto2/server/internal/dir"
	"github.com/Defacto2/server/internal/logs"
	"github.com/nalgeon/be"
)

func TestImages(t *testing.T) {
	t.Parallel()
	err := command.ImagesDelete("", "")
	be.Err(t, err)
	err = command.ImagesPixelate("", "")
	be.Err(t, err)
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
	err := dir.Thumbs(d, "", -1)
	be.Err(t, err, nil)
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

func TestDirs(t *testing.T) {
	t.Parallel()
	dir := command.Dirs{}
	err := dir.PictureImager(context.TODO(), nil, "", "")
	be.Err(t, err)
	err = dir.TextImager(context.TODO(), nil, "", "", false)
	be.Err(t, err)
	err = dir.TextImager(context.TODO(), nil, "", "", true)
	be.Err(t, err)
	err = dir.PreviewPhoto(nil, "", "")
	be.Err(t, err)
	err = dir.PreviewGIF(context.TODO(), nil, "", "")
	be.Err(t, err)
	err = dir.PreviewPNG(nil, "", "")
	be.Err(t, err)
	err = dir.PreviewWebP(context.TODO(), nil, "", "")
	be.Err(t, err)
	d := logs.Discard()
	err = dir.ThumbPixels(d, "", "")
	be.Err(t, err)
	err = dir.ThumbPhoto(d, "", "")
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
	err := command.OptimizePNG("")
	be.Err(t, err, nil)
}
