package option_test

import (
	"slices"
	"testing"

	"github.com/Defacto2/server/internal/command/option"
	"github.com/nalgeon/be"
)

func hasPair(opts option.Opts, flag, val string) bool {
	for i := range len(opts) - 1 {
		if opts[i] == flag && opts[i+1] == val {
			return true
		}
	}
	return false
}

func TestPNGPixel(t *testing.T) {
	t.Parallel()

	var preview option.Opts
	preview.PNGPixel(false, "in.png", "out.png")
	be.Equal(t, preview[0], "in.png")
	be.Equal(t, preview[len(preview)-1], "out.png")
	be.True(t, !slices.Contains(preview, "-thumbnail"))
	be.True(t, !slices.Contains(preview, "-colors"))

	var thumb option.Opts
	thumb.PNGPixel(true, "in.png", "out.png")
	be.Equal(t, thumb[0], "in.png")
	be.Equal(t, thumb[len(thumb)-1], "out.png")
	be.True(t, hasPair(thumb, "-filter", "Box"))
	be.True(t, hasPair(thumb, "-thumbnail", "400x400"))
	be.True(t, slices.Contains(thumb, "+dither"))
	be.True(t, hasPair(thumb, "-colors", "256"))
}

func TestWebpPhoto(t *testing.T) {
	t.Parallel()

	var opts option.Opts
	opts.WebpPhoto("in.jpg", "out.webp")
	be.True(t, !slices.Contains(opts, "-lossless"))
	be.True(t, hasPair(opts, "-preset", "photo"))
	be.True(t, slices.Contains(opts, "-q"))
	be.True(t, slices.Contains(opts, "in.jpg"))
	be.True(t, hasPair(opts, "-o", "out.webp"))
}

func TestWebpPixel(t *testing.T) {
	t.Parallel()

	var opts option.Opts
	opts.WebpPixel("in.png", "out.webp")
	be.True(t, hasPair(opts, "-preset", "text"))
	be.True(t, slices.Contains(opts, "-z"))
	be.True(t, slices.Contains(opts, "in.png"))
	be.True(t, hasPair(opts, "-o", "out.webp"))
}

func TestPhotoOptions(t *testing.T) {
	t.Parallel()

	var jpg option.Opts
	jpg.JPGPhoto(false, "in.jpg", "out.jpg")
	be.Equal(t, jpg[0], "in.jpg")
	be.Equal(t, jpg[len(jpg)-1], "out.jpg")
	be.True(t, slices.Contains(jpg, "-flatten"))
	be.True(t, slices.Contains(jpg, "-quality"))

	var thumb option.Opts
	thumb.PhotoThumb("in.webp", "out.png")
	be.Equal(t, thumb[0], "in.webp")
	be.Equal(t, thumb[len(thumb)-1], "out.png")
	be.True(t, slices.Contains(thumb, "-flatten"))
	be.True(t, slices.Contains(thumb, "-strip"))
	be.True(t, hasPair(thumb, "-filter", "Triangle"))
	be.True(t, hasPair(thumb, "-thumbnail", "400x400"))
	be.True(t, !slices.Contains(thumb, "-quality"))

	var nilOpts *option.Opts
	nilOpts.Thumb()
}

func TestThumbAlignment(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		align   int
		gravity string
	}{
		{"top", 0, "North"},
		{"middle", 1, "center"},
		{"bottom", 2, "South"},
		{"left", 3, "West"},
		{"right", 4, "East"},
		{"invalid", 99, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var opts option.Opts
			opts.ThumbAlignment("in.png", "out.png", tt.align)
			be.Equal(t, opts[0], "in.png")
			be.Equal(t, opts[len(opts)-1], "out.png")
			be.True(t, hasPair(opts, "-trim", "+repage"))
			be.True(t, hasPair(opts, "-thumbnail", "400x400^>"))
			be.True(t, hasPair(opts, "-colors", "256"))
			if tt.gravity == "" {
				be.True(t, !slices.Contains(opts, "-gravity"))
			} else {
				be.True(t, hasPair(opts, "-gravity", tt.gravity))
			}
		})
	}
}

func TestCropAlignment(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		crop   int
		extent string
	}{
		{"1to1", 0, "1:1"},
		{"4to3", 1, "4:3"},
		{"1to2", 2, "1:2"},
		{"invalid", 99, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var opts option.Opts
			opts.CropAlignment("in.png", "out.png", tt.crop)
			be.Equal(t, opts[0], "in.png")
			be.Equal(t, opts[len(opts)-1], "out.png")
			be.True(t, hasPair(opts, "-gravity", "North"))
			if tt.extent == "" {
				be.True(t, !slices.Contains(opts, "-extent"))
			} else {
				be.True(t, hasPair(opts, "-extent", tt.extent))
			}
		})
	}
}
