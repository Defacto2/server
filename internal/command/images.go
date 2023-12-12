package command

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"go.uber.org/zap"
)

// Args is a slice of strings that represents the command line arguments.
// Each argument and its value is a separate string in the slice.
type Args []string

// Jpeg sets the command line arguments for the convert command to transform an image into a JPEG image.
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

// Png sets the command line arguments for the convert command to transform an image into a PNG image.
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

// Thumb sets the command line arguments for the convert command to transform an image into a thumbnail image.
func (a *Args) Thumb() {
	*a = append(*a,
		"-thumbnail", "400x400",
		"-background", "#999",
		"-gravity", "center",
		"-extent", "400x400",
	)
}

// Webp sets the command line arguments for the cwebp command to transform an image into a webp image.
func (a *Args) Webp() {
	*a = append(*a,
		"-af",
		"-v",
		"-exact",
	)
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
	arg := []string{src}                                        // source file
	arg = append(arg, args...)                                  // command line arguments
	tmp := filepath.Join(filepath.Dir(src), BaseName(src)+webp) // destination
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
			z.Warnln("images webp: ", err)
		}
	}()
	return nil
}

// WebpThumbnail converts the src image to a webp image in the thumbnail directory.
func (dir Dirs) WebpThumbnail(z *zap.SugaredLogger, src, uuid string) error {
	if z == nil {
		return ErrZap
	}

	tmp := filepath.Join(dir.Thumbnail, BaseName(src)+jpg)
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

// ConvertLossless converts the src image to a webp image in the screenshot directory.
// A webp thumbnail image is also created and copied to the thumbnail directory.
//
// The lossless conversion is done using the ImageMagick [convert] command
// and transforms the src into a .PNG image before converting to webp.
func (dir Dirs) ConvertLossless(z *zap.SugaredLogger, src, uuid string) error {
	if z == nil {
		return ErrZap
	}

	const png = ".png"

	name := Convert
	_, err := exec.LookPath(name)
	if errors.Is(err, exec.ErrDot) {
		err = nil
	}
	if err != nil {
		return err
	}

	tmp := src + png
	args := Args{}
	args.Png()
	arg := []string{src}       // source file
	arg = append(arg, args...) // command line arguments
	arg = append(arg, tmp)     // destination
	cmd := exec.Command(name, arg...)
	if err := cmd.Run(); err != nil {
		return err
	}

	dst := filepath.Join(dir.Screenshot, uuid+png)
	if err := CopyFile(z, tmp, dst); err != nil {
		return err
	}
	// run these conversions in the background for faster frontend response
	defer func() {
		err := ConvertThumbnail(z, dst, filepath.Join(dir.Thumbnail, uuid+png))
		if err != nil {
			z.Error("convertLossless thumbnail: ", err)
		}
	}()
	defer func() {
		err = ConvertWebP(z, dst)
		if err != nil {
			z.Error("convertLossless webp: ", err)
		}
	}()
	return nil
}

func (dir Dirs) ConvertLossy(z *zap.SugaredLogger, src, uuid string) error {
	if z == nil {
		return ErrZap
	}

	//a PNG, JPEG, TIFF or WebP file.

	const jpg = ".jpg"

	name := Convert
	_, err := exec.LookPath(name)
	if errors.Is(err, exec.ErrDot) {
		err = nil
	}
	if err != nil {
		return err
	}

	tmp := src + jpg
	args := Args{}
	args.Jpeg()
	arg := []string{src}       // source file
	arg = append(arg, args...) // command line arguments
	arg = append(arg, tmp)     // destination
	cmd := exec.Command(name, arg...)
	if err := cmd.Run(); err != nil {
		return err
	}

	dst := filepath.Join(dir.Screenshot, uuid+jpg)
	if err := CopyFile(z, tmp, dst); err != nil {
		return err
	}

	// all these should be deferred and errors printed to the log

	err = ThumbnailLossy(z, dst, filepath.Join(dir.Thumbnail, uuid+jpg))
	if err != nil {
		return err
	}
	err = ConvertWebP(z, dst)
	if err != nil {
		return err
	}
	return nil
}

func ThumbnailLossy(z *zap.SugaredLogger, src, dst string) error {
	if z == nil {
		return ErrZap
	}

	const name = "convert"
	_, err := exec.LookPath(name)
	if errors.Is(err, exec.ErrDot) {
		err = nil
	}
	if err != nil {
		return err
	}

	args := Args{}
	args.Thumb()
	arg := []string{src}       // source file
	arg = append(arg, args...) // command line arguments
	arg = append(arg, dst)     // destination
	cmd := exec.Command(name, arg...)
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return err
	}

	slurp, _ := io.ReadAll(stderr)
	z.Debugln("thumbnaniler: %s %s", cmd, string(slurp))

	if err := cmd.Wait(); err != nil {
		return err
	}

	err = ConvertWebP(z, dst)
	if err != nil {
		return err
	}
	defer os.Remove(dst)

	return nil
}

func ConvertThumbnail(z *zap.SugaredLogger, src, dst string) error {
	if z == nil {
		return ErrZap
	}

	const name = "convert"
	_, err := exec.LookPath(name)
	if errors.Is(err, exec.ErrDot) {
		err = nil
	}
	if err != nil {
		return err
	}

	cmd := exec.Command(name, src, "-thumbnail", "400x400", "-background", "#999", "-gravity", "center", "-extent", "400x400",
		"-define", "png:compression-filter=5", "-define", "png:compression-level=9", "-define", "png:compression-strategy=1", "-define", "png:exclude-chunk=all",
		"-filter", "Triangle",
		"-posterize", "136", // max colours
		dst) // 239151 bytes
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return err
	}

	slurp, _ := io.ReadAll(stderr)
	z.Debugln("thumbnaniler: %s %s", cmd, string(slurp))

	if err := cmd.Wait(); err != nil {
		return err
	}

	err = ConvertWebP(z, dst)
	if err != nil {
		return err
	}
	defer os.Remove(dst)

	return nil
}

func ConvertWebP(z *zap.SugaredLogger, src string) error {
	if z == nil {
		return ErrZap
	}

	const name = "cwebp"
	_, err := exec.LookPath(name)
	if errors.Is(err, exec.ErrDot) {
		err = nil
	}
	if err != nil {
		return err
	}

	// arguments='cwebp -near_lossless 70 "#arguments.source#" -o "#dest#"'
	filename := strings.TrimSuffix(filepath.Base(src), filepath.Ext(filepath.Base(src)))
	dst := filepath.Join(filepath.Dir(src), filename+".webp")
	fmt.Println("conv webp ->", dst)
	cmd := exec.Command(name, src, "-lossless", "-o", dst)
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return err
	}

	slurp, _ := io.ReadAll(stderr)
	z.Debugln("cwebp: %s %s", cmd, string(slurp))

	if err := cmd.Wait(); err != nil {
		return err
	}

	return nil
}

func OptimizePNG(z *zap.SugaredLogger, src string) error {
	if z == nil {
		return ErrZap
	}

	const name = "optipng"
	_, err := exec.LookPath(name)
	if errors.Is(err, exec.ErrDot) {
		err = nil
	}
	if err != nil {
		return err
	}

	cmd := exec.Command(name, src)
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return err
	}

	slurp, _ := io.ReadAll(stderr)
	z.Debugln("optipng: %s %s", cmd, string(slurp))

	if err := cmd.Wait(); err != nil {
		return err
	}

	return nil
}
