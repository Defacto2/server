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
