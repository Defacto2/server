package command

import (
	"os"
	"path/filepath"

	"go.uber.org/zap"
)

// Args is a slice of strings that represents the command line arguments.
// Each argument and its value is a separate string in the slice.
type Args []string

// Jpeg appends the command line arguments for the convert command to transform an image into a JPEG image.
func (a *Args) Jpeg() {
	*a = append(*a,
		"-sampling-factor", "4:2:0",
		"-strip",
		"-quality", "85",
		"-interlace", "Plane",
		"-gaussian-blur", "0.05",
		"-colorspace", "RGB",
	)
}

// Png appends the command line arguments for the convert command to transform an image into a PNG image.
func (a *Args) Png() {
	*a = append(*a,
		"-define", "png:compression-filter=5",
		"-define", "png:compression-level=9",
		"-define", "png:compression-strategy=1",
		"-define", "png:exclude-chunk=all",
		"-filter", "Triangle",
		"-flatten",
		"-posterize", "136", // max colours
	)
}

// Thumb appends the command line arguments for the convert command to transform an image into a thumbnail image.
func (a *Args) Thumb() {
	*a = append(*a,
		"-thumbnail", "400x400",
		"-background", "#999",
		"-gravity", "center",
		"-extent", "400x400",
	)
}

// Webp appends the command line arguments for the cwebp command to transform an image into a webp image.
func (a *Args) Webp() {
	*a = append(*a,
		"-af",
		"-v",
		"-exact",
	)
}

// PngScreenshot copies and optimizes the src PNG image to the screenshot directory.
// A webp thumbnail image is also created and copied to the thumbnail directory.
func (dir Dirs) PngScreenshot(z *zap.SugaredLogger, src, uuid string) error {
	if z == nil {
		return ErrZap
	}

	dst := filepath.Join(dir.Screenshot, uuid+png)
	if err := CopyFile(z, src, dst); err != nil {
		return err
	}

	defer func() {
		err := OptimizePNG(z, dst)
		if err != nil {
			z.Warnln("png screenshot: ", err)
		}
	}()

	defer func() {
		err := dir.WebpThumbnail(z, src, uuid)
		if err != nil {
			z.Warnln("png screenshot: ", err)
		}
	}()
	return nil
}

// WebpScreenshot converts the src image to a webp image in the screenshot directory.
// A webp thumbnail image is also created and copied to the thumbnail directory.
//
// The conversion is done using the cwebp command, which supports either
// a PNG, JPEG, TIFF or WebP source image file.
func (dir Dirs) WebpScreenshot(z *zap.SugaredLogger, src, uuid string) error {
	if z == nil {
		return ErrZap
	}

	args := Args{}
	args.Webp()
	arg := []string{src}            // source file
	arg = append(arg, args...)      // command line arguments
	tmp := BaseNamePath(src) + webp // destination
	arg = append(arg, "-o", tmp)
	if err := RunQuiet(z, Cwebp, arg...); err != nil {
		return err
	}

	dst := filepath.Join(dir.Screenshot, uuid+webp)
	if err := CopyFile(z, tmp, dst); err != nil {
		return err
	}

	defer func() {
		err := dir.WebpThumbnail(z, tmp, uuid)
		if err != nil {
			z.Warnln("webp screenshot: ", err)
		}
	}()
	return nil
}

// WebpThumbnail converts the src image to a webp image in the thumbnail directory.
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
	args.Webp()
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
	arg := []string{src}           // source file
	arg = append(arg, args...)     // command line arguments
	tmp := BaseNamePath(src) + png // destination
	arg = append(arg, tmp)
	if err := RunQuiet(z, Convert, arg...); err != nil {
		return err
	}

	dst := filepath.Join(dir.Screenshot, uuid+png)
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

// LossyScreenshot converts the src image to a lossy Webp image in the screenshot directory.
// A webp thumbnail image is also created and copied to the thumbnail directory.
// The lossy conversion is useful for photographs.
//
// The lossy conversion is done using the ImageMagick [convert] command.
//
// [convert]: https://imagemagick.org/script/convert.php
func (dir Dirs) LossyScreenshot(z *zap.SugaredLogger, src, uuid string) error {
	if z == nil {
		return ErrZap
	}

	tmp := BaseNamePath(src) + jpg
	args := Args{}
	args.Jpeg()
	arg := []string{src}       // source file
	arg = append(arg, args...) // command line arguments
	arg = append(arg, tmp)     // destination
	if err := RunQuiet(z, Convert, arg...); err != nil {
		return err
	}

	dst := filepath.Join(dir.Screenshot, uuid+webp)
	args = Args{}
	args.Webp()
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
	if err := RunQuiet(z, Optipng, arg...); err != nil {
		return err
	}

	return nil
}
