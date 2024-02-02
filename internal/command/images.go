package command

import (
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
	*a = append(*a,
		"-f", "topaz+", // Output font.
		"-m", "workbench", // Rendering mode set to Amiga Workbench palette.
		"-S", // Use SAUCE record for render options.
	)
}

// AnsiDOS appends the command line arguments for the [ansilove command] to
// transform an ANSI text file into a PNG image.
//
// [ansilove command]: https://github.com/ansilove/ansilove
func (a *Args) AnsiDOS() {
	*a = append(*a,
		"-d",          // DOS aspect ratio.
		"-f", "80x25", // Output font.
		"-i", // Use iCE colors.
		"-S", // Use SAUCE record for render options.
	)
}

// Jpeg appends the command line arguments for the convert command to
// transform an image into a JPEG image.
func (a *Args) Jpeg() {
	*a = append(*a,
		"-sampling-factor", "4:2:0", // Horizontal and vertical sampling factors to be used by the JPEG encoder for chroma downsampling.
		"-strip",         // Strip the image of any profiles and comments.
		"-quality", "90", // See: https://imagemagick.org/script/command-line-options.php#quality
		"-interlace", "plane", // Type of interlacing scheme, see: https://imagemagick.org/script/command-line-options.php#interlace
		"-gaussian-blur", "0.05", // Blur the image with a Gaussian operator.
		"-colorspace", "RGB", // Set the image colorspace.
	)
}

// Png appends the command line arguments for the convert command to transform an image into a PNG image.
func (a *Args) Png() {
	*a = append(*a,
		// Defined PNG compression options, these replace the -quality option.
		"-define", "png:compression-filter=5",
		"-define", "png:compression-level=9",
		"-define", "png:compression-strategy=1",
		"-define", "png:exclude-chunk=all",
		"-flatten",          // Create a canvas the size of the first images virtual canvas using the current -background color, and -compose each image in turn onto that canvas.
		"-strip",            // Strip the image of any profiles, comments or PNG chunks.
		"-posterize", "136", // Reduce the image to a limited number of color levels per channel.
	)
}

// Thumb appends the command line arguments for the convert command to transform an image into a thumbnail image.
func (a *Args) Thumb() {
	*a = append(*a,
		"-filter", "Triangle", // Use this type of filter when resizing or distorting an image.
		"-thumbnail", "400x400", // Create a thumbnail of the image, more performant than -resize.
		"-background", "#999", // Set the background color.
		"-gravity", "center", // Sets the current gravity suggestion for various other settings and options.
		"-extent", "400x400", // Set the image size and offset.
	)
}

// CWebp appends the command line arguments for the [cwebp command] to transform an image into a webp image.
//
// [cwebp command]: https://developers.google.com/speed/webp/docs/cwebp
func (a *Args) CWebp() {
	*a = append(*a,
		"-af",    // Auto-filter will spend additional time optimizing the filtering strength to reach a well-balanced quality.
		"-exact", // Preserve RGB values in transparent area. The default is off, to help compressibility.
		// "-v", // Print extra information.
	)
}

// GWebp appends the command line arguments for the [gif2webp command] to transform a GIF image into a webp image.
//
// [gif2webp command]: https://developers.google.com/speed/webp/docs/gif2webp
func (a *Args) GWebp() {
	*a = append(*a,
		"-q", "100", // Compression factor for RGB channels between 0 and 100.
		"-mt", // Use multi-threading if available.
		// "-v", // Print extra information.
	)
}

// AnsiLove converts the src text file and creates a PNG image in the preview directory.
// A webp thumbnail image is also created and copied to the thumbnail directory.
func (dir Dirs) AnsiLove(z *zap.SugaredLogger, src, uuid string) error {
	if z == nil {
		return ErrZap
	}

	args := Args{}
	args.AnsiDOS()
	arg := []string{src}           // source file
	arg = append(arg, args...)     // command line arguments
	tmp := BaseNamePath(src) + png // destination
	arg = append(arg, "-o", tmp)
	if err := Run(z, Ansilove, arg...); err != nil {
		return err
	}

	dst := filepath.Join(dir.Preview, uuid+png)
	if err := CopyFile(z, tmp, dst); err != nil {
		return err
	}
	defer func() {
		err := OptimizePNG(z, dst)
		if err != nil {
			z.Warnln("ansilove: ", err)
		}
	}()
	defer func() {
		err := dir.AnsiThumbnail(z, tmp, uuid)
		if err != nil {
			z.Warnln("ansilove: ", err)
		}
	}()
	return nil
}

// PreviewPNG copies and optimizes the src PNG image to the screenshot directory.
// A webp thumbnail image is also created and copied to the thumbnail directory.
func (dir Dirs) PreviewGIF(z *zap.SugaredLogger, src, uuid string) error {
	if z == nil {
		return ErrZap
	}

	args := Args{}
	args.GWebp()
	arg := []string{src}            // source file
	arg = append(arg, args...)      // command line arguments
	tmp := BaseNamePath(src) + webp // destination
	arg = append(arg, "-o", tmp)
	if err := Run(z, Gwebp, arg...); err != nil {
		return err
	}

	dst := filepath.Join(dir.Preview, uuid+webp)
	if err := CopyFile(z, tmp, dst); err != nil {
		return err
	}

	defer func() {
		err := dir.WebpThumbnail(z, tmp, uuid)
		if err != nil {
			z.Warnln("gif: ", err)
		}
	}()
	return nil
}

// PreviewPNG copies and optimizes the src PNG image to the screenshot directory.
// A webp thumbnail image is also created and copied to the thumbnail directory.
func (dir Dirs) PreviewPNG(z *zap.SugaredLogger, src, uuid string) error {
	if z == nil {
		return ErrZap
	}

	dst := filepath.Join(dir.Preview, uuid+png)
	if err := CopyFile(z, src, dst); err != nil {
		return err
	}
	defer func() {
		err := OptimizePNG(z, dst)
		if err != nil {
			z.Warnln("png: ", err)
		}
	}()
	defer func() {
		err := dir.WebpThumbnail(z, src, uuid)
		if err != nil {
			z.Warnln("png: ", err)
		}
	}()
	return nil
}

// PreviewWebP converts the src image to a webp image in the screenshot directory.
// A webp thumbnail image is also created and copied to the thumbnail directory.
//
// The conversion is done using the cwebp command, which supports either
// a PNG, JPEG, TIFF or WebP source image file.
func (dir Dirs) PreviewWebP(z *zap.SugaredLogger, src, uuid string) error {
	if z == nil {
		return ErrZap
	}

	args := Args{}
	args.CWebp()
	arg := []string{src}            // source file
	arg = append(arg, args...)      // command line arguments
	tmp := BaseNamePath(src) + webp // destination
	arg = append(arg, "-o", tmp)
	if err := RunQuiet(z, Cwebp, arg...); err != nil {
		return err
	}

	dst := filepath.Join(dir.Preview, uuid+webp)
	if err := CopyFile(z, tmp, dst); err != nil {
		return err
	}
	defer func() {
		err := dir.WebpThumbnail(z, tmp, uuid)
		if err != nil {
			z.Warnln("webp: ", err)
		}
	}()
	return nil
}

// AnsiThumbnail converts the src image to a 400x400 pixel, webp image in the thumbnail directory.
// The conversion is done using a temporary, lossless PNG image.
func (dir Dirs) AnsiThumbnail(z *zap.SugaredLogger, src, uuid string) error {
	if z == nil {
		return ErrZap
	}

	tmp := filepath.Join(dir.Thumbnail, uuid+png)
	args := Args{}
	args.Thumb()
	args.Png()
	arg := []string{src}       // source file
	arg = append(arg, args...) // command line arguments
	arg = append(arg, tmp)     // destination
	if err := RunQuiet(z, Convert, arg...); err != nil {
		return err
	}

	dst := filepath.Join(dir.Thumbnail, uuid+webp)
	args = Args{}
	args.CWebp()
	arg = []string{tmp}          // source file
	arg = append(arg, args...)   // command line arguments
	arg = append(arg, "-o", dst) // destination
	if err := RunQuiet(z, Cwebp, arg...); err != nil {
		return err
	}
	defer os.Remove(tmp)
	return nil
}

// WebpThumbnail converts the src image to a 400x400 pixel, webp image in the thumbnail directory.
// The conversion is done using a temporary, lossy PNG image.
func (dir Dirs) WebpThumbnail(z *zap.SugaredLogger, src, uuid string) error {
	if z == nil {
		return ErrZap
	}

	tmp := BaseNamePath(src) + jpg
	args := Args{}
	args.Thumb()
	args.Jpeg()
	arg := []string{src}       // source file
	arg = append(arg, args...) // command line arguments
	arg = append(arg, tmp)     // destination
	if err := RunQuiet(z, Convert, arg...); err != nil {
		return err
	}

	dst := filepath.Join(dir.Thumbnail, uuid+webp)
	args = Args{}
	args.CWebp()
	arg = []string{tmp}          // source file
	arg = append(arg, args...)   // command line arguments
	arg = append(arg, "-o", dst) // destination
	if err := RunQuiet(z, Cwebp, arg...); err != nil {
		return err
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
func (dir Dirs) LosslessScreenshot(z *zap.SugaredLogger, src, uuid string) error {
	if z == nil {
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
		return err
	}
	defer os.RemoveAll(tmp)        // remove temp dir
	tmp = filepath.Join(tmp, name) // temp output file target
	arg = append(arg, tmp)
	if err := RunQuiet(z, Convert, arg...); err != nil {
		return err
	}

	dst := filepath.Join(dir.Preview, uuid+png)
	if err := CopyFile(z, tmp, dst); err != nil {
		return err
	}

	defer func() {
		err := dir.WebpThumbnail(z, tmp, uuid)
		if err != nil {
			z.Warnln("lossless screenshot: ", err)
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
func (dir Dirs) PreviewLossy(z *zap.SugaredLogger, src, uuid string) error {
	if z == nil {
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
		return err
	}
	defer os.RemoveAll(tmp)        // remove temp dir
	tmp = filepath.Join(tmp, name) // temp output file target
	arg = append(arg, tmp)         // destination
	if err := RunQuiet(z, Convert, arg...); err != nil {
		return err
	}

	dst := filepath.Join(dir.Preview, uuid+webp)
	args = Args{}
	args.CWebp()
	arg = []string{tmp}          // source file
	arg = append(arg, args...)   // command line arguments
	arg = append(arg, "-o", dst) // destination
	if err := RunQuiet(z, Cwebp, arg...); err != nil {
		return err
	}
	defer os.Remove(tmp)

	defer func() {
		err := dir.WebpThumbnail(z, tmp, uuid)
		if err != nil {
			z.Warnln("lossy screenshot: ", err)
		}
	}()
	return nil
}

// OptimizePNG optimizes the src PNG image using the optipng command.
// The optimization is done in-place, overwriting the src file.
// It should be used in a deferred function.
func OptimizePNG(z *zap.SugaredLogger, src string) error {
	if z == nil {
		return ErrZap
	}

	args := Args{}
	arg := []string{src}       // source file
	arg = append(arg, args...) // command line arguments
	return RunQuiet(z, Optipng, arg...)
}
