package command

// Package file images.go contains the image conversion functions for
// converting images to PNG and WebP formats using ANSILOVE, ImageMagick
// and other command-line tools.

import (
	"fmt"
	"os"
	"path/filepath"

	"go.uber.org/zap"
)

// ansilove may find -extent and -extract useful
// https://imagemagick.org/script/command-line-options.php#extent

// Args is a slice of strings that represents the command line arguments.
// Each argument and its value is a separate string in the slice.
type Args []string

// AnsiDOS appends the command line arguments for the [ansilove command]
// to transform an Commodore Amiga ANSI text file into a PNG image.
//
// [ansilove command]: https://github.com/ansilove/ansilove
func (a *Args) AnsiAmiga() {
	// Output font.
	const f = "-f"
	// Rendering mode set to Amiga Workbench palette.
	const m = "-m"
	// Use SAUCE record for render options.
	const s = "-S"
	*a = append(*a,
		f, "topaz+", m, "workbench", s,
	)
}

// AnsiDOS appends the command line arguments for the [ansilove command] to
// transform an ANSI text file into a PNG image.
//
// [ansilove command]: https://github.com/ansilove/ansilove
func (a *Args) AnsiDOS() {
	// DOS aspect ratio.
	const d = "-d"
	// Output font.
	const f = "-f"
	// Use iCE colors.
	const i = "-i"
	// Use SAUCE record for render options.
	const s = "-S"
	*a = append(*a,
		d, f, "80x25", i, s,
	)
}

// Jpeg appends the command line arguments for the convert command to
// transform an image into a JPEG image.
func (a *Args) Jpeg() {
	// Horizontal and vertical sampling factors to be used by the JPEG encoder for chroma downsampling.
	const sampleFactor = "-sampling-factor"
	// Strip the image of any profiles and comments.
	const strip = "-strip"
	// See: https://imagemagick.org/script/command-line-options.php#quality
	const quality = "-quality"
	// Type of interlacing scheme, see: https://imagemagick.org/script/command-line-options.php#interlace
	const interlace = "-interlace"
	// Blur the image with a Gaussian operator.
	const gaussianBlur = "-gaussian-blur"
	// Set the image colorspace.
	const colorspace = "-colorspace"
	*a = append(*a,
		sampleFactor, "4:2:0", strip,
		quality, "90",
		interlace, "plane",
		gaussianBlur, "0.05",
		colorspace, "RGB",
	)
}

// Png appends the command line arguments for the convert command to transform an image into a PNG image.
func (a *Args) Png() {
	// Defined PNG compression options, these replace the -quality option.
	const define = "-define"
	// Create a canvas the size of the first images virtual canvas using the
	// current -background color, and -compose each image in turn onto that canvas.
	const flatten = "-flatten"
	// Strip the image of any profiles, comments or PNG chunks.
	const strip = "-strip"
	// Reduce the image to a limited number of color levels per channel.
	const posterize = "-posterize"
	*a = append(*a,
		define, "png:compression-filter=5",
		define, "png:compression-level=9",
		define, "png:compression-strategy=1",
		define, "png:exclude-chunk=all",
		flatten,
		strip,
		posterize, "136",
	)
}

// Thumb appends the command line arguments for the convert command to transform an image into a thumbnail image.
func (a *Args) Thumb() {
	// Use this type of filter when resizing or distorting an image.
	const filter = "-filter"
	// Create a thumbnail of the image, more performant than -resize.
	const thumbnail = "-thumbnail"
	// Set the background color.
	const background = "-background"
	// Sets the current gravity suggestion for various other settings and options.
	const gravity = "-gravity"
	// Set the image size and offset.
	const extent = "-extent"
	*a = append(*a,
		filter, "Triangle",
		thumbnail, "400x400",
		background, "#999",
		gravity, "center",
		extent, "400x400",
	)
}

// CWebp appends the command line arguments for the [cwebp command] to transform an image into a webp image.
//
// [cwebp command]: https://developers.google.com/speed/webp/docs/cwebp
func (a *Args) CWebp() {
	// Auto-filter will spend additional time optimizing the
	// filtering strength to reach a well-balanced quality.
	const af = "-af"
	// Preserve RGB values in transparent area. The default is off, to help compressibility.
	const exact = "-exact"
	*a = append(*a,
		af, exact,
		// "-v", // Print extra information.
	)
}

// GWebp appends the command line arguments for the [gif2webp command] to transform a GIF image into a webp image.
//
// [gif2webp command]: https://developers.google.com/speed/webp/docs/gif2webp
func (a *Args) GWebp() {
	// Compression factor for RGB channels between 0 and 100.
	const q = "-q"
	// Use multi-threading if available.
	const mt = "-mt"
	*a = append(*a,
		q, "100",
		mt,
		// "-v", // Print extra information.
	)
}

// AnsiLove converts the src text file and creates a PNG image in the preview directory.
// A webp thumbnail image is also created and copied to the thumbnail directory.
func (dir Dirs) AnsiLove(logger *zap.SugaredLogger, src, unid string) error {
	if logger == nil {
		return ErrZap
	}
	args := Args{}
	args.AnsiDOS()
	arg := []string{src}           // source file
	arg = append(arg, args...)     // command line arguments
	tmp := BaseNamePath(src) + png // destination
	arg = append(arg, "-o", tmp)
	if err := Run(logger, Ansilove, arg...); err != nil {
		return fmt.Errorf("ansilove: %w", err)
	}

	dst := filepath.Join(dir.Preview, unid+png)
	if err := CopyFile(logger, tmp, dst); err != nil {
		return fmt.Errorf("copy file: %w", err)
	}
	defer func() {
		err := OptimizePNG(dst)
		if err != nil {
			logger.Warnln("ansilove: ", err)
		}
	}()
	defer func() {
		err := dir.AnsiThumbnail(tmp, unid)
		if err != nil {
			logger.Warnln("ansilove: ", err)
		}
	}()
	return nil
}

// PreviewPNG copies and optimizes the src PNG image to the screenshot directory.
// A webp thumbnail image is also created and copied to the thumbnail directory.
func (dir Dirs) PreviewGIF(logger *zap.SugaredLogger, src, unid string) error {
	if logger == nil {
		return ErrZap
	}

	args := Args{}
	args.GWebp()
	arg := []string{src}            // source file
	arg = append(arg, args...)      // command line arguments
	tmp := BaseNamePath(src) + webp // destination
	arg = append(arg, "-o", tmp)
	if err := Run(logger, Gwebp, arg...); err != nil {
		return fmt.Errorf("gif2webp: %w", err)
	}

	dst := filepath.Join(dir.Preview, unid+webp)
	if err := CopyFile(logger, tmp, dst); err != nil {
		return fmt.Errorf("copy file: %w", err)
	}

	defer func() {
		err := dir.WebpThumbnail(tmp, unid)
		if err != nil {
			logger.Warnln("gif: ", err)
		}
	}()
	return nil
}

// PreviewPNG copies and optimizes the src PNG image to the screenshot directory.
// A webp thumbnail image is also created and copied to the thumbnail directory.
func (dir Dirs) PreviewPNG(logger *zap.SugaredLogger, src, unid string) error {
	if logger == nil {
		return ErrZap
	}

	dst := filepath.Join(dir.Preview, unid+png)
	if err := CopyFile(logger, src, dst); err != nil {
		return fmt.Errorf("copy file: %w", err)
	}
	defer func() {
		err := OptimizePNG(dst)
		if err != nil {
			logger.Warnln("png: ", err)
		}
	}()
	defer func() {
		err := dir.WebpThumbnail(src, unid)
		if err != nil {
			logger.Warnln("png: ", err)
		}
	}()
	return nil
}

// PreviewWebP converts the src image to a webp image in the screenshot directory.
// A webp thumbnail image is also created and copied to the thumbnail directory.
//
// The conversion is done using the cwebp command, which supports either
// a PNG, JPEG, TIFF or WebP source image file.
func (dir Dirs) PreviewWebP(logger *zap.SugaredLogger, src, unid string) error {
	if logger == nil {
		return ErrZap
	}

	args := Args{}
	args.CWebp()
	arg := []string{src}            // source file
	arg = append(arg, args...)      // command line arguments
	tmp := BaseNamePath(src) + webp // destination
	arg = append(arg, "-o", tmp)
	if err := RunQuiet(Cwebp, arg...); err != nil {
		return fmt.Errorf("cwebp: %w", err)
	}

	dst := filepath.Join(dir.Preview, unid+webp)
	if err := CopyFile(logger, tmp, dst); err != nil {
		return fmt.Errorf("copy file: %w", err)
	}
	defer func() {
		err := dir.WebpThumbnail(tmp, unid)
		if err != nil {
			logger.Warnln("webp: ", err)
		}
	}()
	return nil
}

// AnsiThumbnail converts the src image to a 400x400 pixel, webp image in the thumbnail directory.
// The conversion is done using a temporary, lossless PNG image.
func (dir Dirs) AnsiThumbnail(src, unid string) error {
	tmp := filepath.Join(dir.Thumbnail, unid+png)
	args := Args{}
	args.Thumb()
	args.Png()
	arg := []string{src}       // source file
	arg = append(arg, args...) // command line arguments
	arg = append(arg, tmp)     // destination
	if err := RunQuiet(Convert, arg...); err != nil {
		return fmt.Errorf("convert: %w", err)
	}

	dst := filepath.Join(dir.Thumbnail, unid+webp)
	args = Args{}
	args.CWebp()
	arg = []string{tmp}          // source file
	arg = append(arg, args...)   // command line arguments
	arg = append(arg, "-o", dst) // destination
	if err := RunQuiet(Cwebp, arg...); err != nil {
		return fmt.Errorf("cwebp: %w", err)
	}
	defer os.Remove(tmp)
	return nil
}

// WebpThumbnail converts the src image to a 400x400 pixel, webp image in the thumbnail directory.
// The conversion is done using a temporary, lossy PNG image.
func (dir Dirs) WebpThumbnail(src, unid string) error {
	tmp := BaseNamePath(src) + jpg
	args := Args{}
	args.Thumb()
	args.Jpeg()
	arg := []string{src}       // source file
	arg = append(arg, args...) // command line arguments
	arg = append(arg, tmp)     // destination
	if err := RunQuiet(Convert, arg...); err != nil {
		return fmt.Errorf("convert: %w", err)
	}

	dst := filepath.Join(dir.Thumbnail, unid+webp)
	args = Args{}
	args.CWebp()
	arg = []string{tmp}          // source file
	arg = append(arg, args...)   // command line arguments
	arg = append(arg, "-o", dst) // destination
	if err := RunQuiet(Cwebp, arg...); err != nil {
		return fmt.Errorf("cwebp: %w", err)
	}
	defer os.Remove(tmp)
	return nil
}

// LosslessScreenshot converts the src image to a lossless PNG image in the screenshot directory.
// A webp thumbnail image is also created and copied to the thumbnail directory.
// The lossless conversion is useful for screenshots of text, terminals interfaces and pixel art.
//
// The lossless conversion is done using the ImageMagick [convert] command.
//
// [convert]: https://imagemagick.org/script/convert.php
func (dir Dirs) LosslessScreenshot(logger *zap.SugaredLogger, src, unid string) error {
	if logger == nil {
		return ErrZap
	}

	args := Args{}
	args.Png()
	arg := []string{src}       // source file
	arg = append(arg, args...) // command line arguments
	// create a temporary target file in the temp dir
	name := filepath.Base(src) + png                             // temp file name
	tmp, err := os.MkdirTemp(os.TempDir(), "losslessscreenshot") // create temp dir
	if err != nil {
		return fmt.Errorf("os.MkdirTemp: %w", err)
	}
	defer os.RemoveAll(tmp)        // remove temp dir
	tmp = filepath.Join(tmp, name) // temp output file target
	arg = append(arg, tmp)
	if err := RunQuiet(Convert, arg...); err != nil {
		return fmt.Errorf("convert: %w", err)
	}

	dst := filepath.Join(dir.Preview, unid+png)
	if err := CopyFile(logger, tmp, dst); err != nil {
		return fmt.Errorf("copy file: %w", err)
	}

	defer func() {
		err := dir.WebpThumbnail(tmp, unid)
		if err != nil {
			logger.Warnln("lossless screenshot: ", err)
		}
	}()
	return nil
}

// PreviewLossy converts the src image to a lossy Webp image in the screenshot directory.
// A webp thumbnail image is also created and copied to the thumbnail directory.
// The lossy conversion is useful for photographs.
//
// The lossy conversion is done using the ImageMagick [convert] command.
//
// [convert]: https://imagemagick.org/script/convert.php
func (dir Dirs) PreviewLossy(logger *zap.SugaredLogger, src, unid string) error {
	if logger == nil {
		return ErrZap
	}

	args := Args{}
	args.Jpeg()
	arg := []string{src}       // source file
	arg = append(arg, args...) // command line arguments
	// create a temporary target file in the temp dir
	name := filepath.Base(src) + jpg                       // temp file name
	tmp, err := os.MkdirTemp(os.TempDir(), "lossypreview") // create temp dir
	if err != nil {
		return fmt.Errorf("os.MkdirTemp: %w", err)
	}
	defer os.RemoveAll(tmp)        // remove temp dir
	tmp = filepath.Join(tmp, name) // temp output file target
	arg = append(arg, tmp)         // destination
	if err := RunQuiet(Convert, arg...); err != nil {
		return fmt.Errorf("convert: %w", err)
	}

	dst := filepath.Join(dir.Preview, unid+webp)
	args = Args{}
	args.CWebp()
	arg = []string{tmp}          // source file
	arg = append(arg, args...)   // command line arguments
	arg = append(arg, "-o", dst) // destination
	if err := RunQuiet(Cwebp, arg...); err != nil {
		return fmt.Errorf("cwebp: %w", err)
	}
	defer os.Remove(tmp)

	defer func() {
		err := dir.WebpThumbnail(tmp, unid)
		if err != nil {
			logger.Warnln("lossy screenshot: ", err)
		}
	}()
	return nil
}

// OptimizePNG optimizes the src PNG image using the optipng command.
// The optimization is done in-place, overwriting the src file.
// It should be used in a deferred function.
func OptimizePNG(src string) error {
	args := Args{}
	arg := []string{src}       // source file
	arg = append(arg, args...) // command line arguments
	return RunQuiet(Optipng, arg...)
}
