package option_test

import (
	"slices"
	"testing"

	"github.com/Defacto2/server/internal/command/option"
	"github.com/nalgeon/be"
)

func TestPNGPixel(t *testing.T) {
	t.Parallel()

	var preview option.Opts
	preview.PNGPixel(false, "in.png", "out.png")
	be.True(t, !slices.Contains(preview, "-thumbnail"))
	be.True(t, !slices.Contains(preview, "-colors"))

	var thumb option.Opts
	thumb.PNGPixel(true, "in.png", "out.png")
	be.True(t, slices.Contains(thumb, "Box"))
	be.True(t, slices.Contains(thumb, "400x400"))
	be.True(t, slices.Contains(thumb, "+dither"))
	be.True(t, slices.Contains(thumb, "-colors"))
	be.True(t, slices.Contains(thumb, "256"))
}

func TestWebpPhoto(t *testing.T) {
	t.Parallel()

	var opts option.Opts
	opts.WebpPhoto("in.jpg", "out.webp")
	be.True(t, !slices.Contains(opts, "-lossless"))
	be.True(t, slices.Contains(opts, "photo"))
	be.True(t, slices.Contains(opts, "-q"))
	be.True(t, slices.Contains(opts, "in.jpg"))
	be.True(t, slices.Contains(opts, "out.webp"))
}

func TestWebpPixel(t *testing.T) {
	t.Parallel()

	var opts option.Opts
	opts.WebpPixel("in.png", "out.webp")
	be.True(t, slices.Contains(opts, "text"))
	be.True(t, slices.Contains(opts, "-z"))
	be.True(t, slices.Contains(opts, "in.png"))
	be.True(t, slices.Contains(opts, "out.webp"))
}

func TestPhotoOptions(t *testing.T) {
	t.Parallel()

	var jpg option.Opts
	jpg.JPGPhoto(false, "in.jpg", "out.jpg")
	be.True(t, slices.Contains(jpg, "-flatten"))
	be.True(t, slices.Contains(jpg, "-quality"))

	var thumb option.Opts
	thumb.PhotoThumb("in.webp", "out.png")
	be.True(t, slices.Contains(thumb, "-flatten"))
	be.True(t, slices.Contains(thumb, "-strip"))
	be.True(t, slices.Contains(thumb, "Triangle"))
	be.True(t, slices.Contains(thumb, "400x400"))
	be.True(t, !slices.Contains(thumb, "-quality"))
}

func TestThumbAlignment(t *testing.T) {
	t.Parallel()

	var opts option.Opts
	opts.ThumbAlignment("in.png", "out.png", 1)
	be.True(t, slices.Contains(opts, "-flatten"))
	be.True(t, slices.Contains(opts, "-strip"))
	be.True(t, slices.Contains(opts, "-trim"))
	be.True(t, slices.Contains(opts, "+repage"))
	be.True(t, slices.Contains(opts, "400x400^>"))
	be.True(t, slices.Contains(opts, "#999"))
	be.True(t, slices.Contains(opts, "center"))
	be.True(t, slices.Contains(opts, "+dither"))
	be.True(t, slices.Contains(opts, "-colors"))
	be.True(t, slices.Contains(opts, "256"))
}
