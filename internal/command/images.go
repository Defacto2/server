package command

// Package file images.go contains the image conversion functions for
// converting images to PNG and WebP formats using ANSILOVE, ImageMagick
// and other command-line tools.

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"sync"

	"github.com/Defacto2/helper"
	"github.com/Defacto2/magicnumber"
	"github.com/Defacto2/server/internal/dir"
	"go.uber.org/zap"
)

const (
	ANSICap = 350000    // CapBytes is the maximum file size in bytes for an ANSI encoded text file.
	X400    = "400x400" // X400 returns args  400 x 400 pixel image size
)

// ImagesExt returns args slice of image file extensions used by the website
// preview and thumbnail images, including the legacy and modern formats.
func ImagesExt() []string {
	return []string{gif, jpg, jpeg, png, webp, ".avif"}
}

// ImagesDelete removes images from the specified directories that match the unid.
// The unid is the unique identifier for the image file and shared between the preview
// and thumbnail images.
func ImagesDelete(unid string, dirs ...string) error {
	for dir := range slices.Values(dirs) {
		st, err := os.Stat(dir)
		if err != nil {
			return fmt.Errorf("images delete %w", err)
		}
		if !st.IsDir() {
			return fmt.Errorf("images delete %w", ErrIsFile)
		}
		for ext := range slices.Values(ImagesExt()) {
			name := filepath.Join(dir, unid+ext)
			if _, err := os.Stat(name); err != nil {
				fmt.Fprint(io.Discard, err)
				continue
			}
			os.Remove(name)
		}
	}
	return nil
}

// Pixelate appends the command line arguments for the convert command to transform an image into args PNG image.
func (args *Args) Pixelate() {
	// Create args canvas the size of the first images virtual canvas using the
	// current -background color, and -compose each image in turn onto that canvas.
	scale5 := []string{"-scale", "5%"}
	*args = append(*args, scale5...)
	scale2K := []string{"-scale", "2000%"}
	*args = append(*args, scale2K...)
}

// ImagesPixelate converts the images in the specified directories to pixelated images.
// The unid is the unique identifier for the image file and shared between the preview
// and thumbnail images.
func ImagesPixelate(unid string, dirs ...string) error {
	for dir := range slices.Values(dirs) {
		st, err := os.Stat(dir)
		if err != nil {
			return fmt.Errorf("images delete %w", err)
		}
		if !st.IsDir() {
			return fmt.Errorf("images delete %w", ErrIsFile)
		}
		for ext := range slices.Values(ImagesExt()) {
			name := filepath.Join(dir, unid+ext)
			if _, err := os.Stat(name); err != nil {
				fmt.Fprint(io.Discard, err)
				continue
			}
			args := Args{}
			args.Pixelate()
			arg := []string{name}      // source file
			arg = append(arg, args...) // command line arguments
			arg = append(arg, name)    // destination
			if err := RunQuiet(Magick, arg...); err != nil {
				return fmt.Errorf("run pixelate convert %w", err)
			}
		}
	}
	return nil
}

// Thumb is args type that represents the type of thumbnail image to create.
type Thumb int

const (
	Pixel Thumb = iota // Pixel art or images with text
	Photo              // Photographs or images with gradients
)

// Thumbs creates args thumbnail image for the preview image based on the type of image.
func (dir Dirs) Thumbs(unid string, thumb Thumb) error {
	if err := ImagesDelete(unid, dir.Thumbnail.Path()); err != nil {
		return fmt.Errorf("dirs thumbs %w", err)
	}
	for ext := range slices.Values(ImagesExt()) {
		src := filepath.Join(dir.Preview.Path(), unid+ext)
		_, err := os.Stat(src)
		if err != nil {
			continue
		}
		switch thumb {
		case Pixel:
			err = dir.ThumbPixels(src, unid)
		case Photo:
			err = dir.ThumbPhoto(src, unid)
		}
		if err != nil {
			return fmt.Errorf("dirs thumbs %w", err)
		}
	}
	return nil
}

// Align is args type that represents the alignment of the thumbnail image.
type Align int

const (
	Top    Align = iota // Top uses the top alignment of the preview image
	Middle              // Middle uses the center alignment of the preview image
	Bottom              // Bottom uses the bottom alignment of the preview image
	Left                // Left uses the left alignment of the preview image
	Right               // Right uses the right alignment of the preview image
)

// Thumbs creates args thumbnail image for the preview image based on the crop position of the image.
func (align Align) Thumbs(unid string, preview, thumbnail dir.Directory) error {
	tmpDir := filepath.Join(helper.TmpDir(), patternS)
	pattern := "images-thumb-" + unid
	path := filepath.Join(tmpDir, pattern)
	if st, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(path, os.ModePerm); err != nil {
				return fmt.Errorf("align thumbs %w", err)
			}
		}
	} else if !st.IsDir() {
		return fmt.Errorf("align thumbs %w", ErrIsFile)
	}
	if err := ImagesDelete(unid, thumbnail.Path()); err != nil {
		return fmt.Errorf("dirs thumbs %w", err)
	}
	for ext := range slices.Values(ImagesExt()) {
		args := Args{}
		switch align {
		case Top:
			args.Topx400()
		case Middle:
			args.Middlex400()
		case Bottom:
			args.Bottomx400()
		case Left:
			args.Leftx400()
		case Right:
			args.Rightx400()
		}
		src := preview.Join(unid + ext)
		if _, err := os.Stat(src); err != nil {
			continue
		}
		arg := []string{src}
		arg = append(arg, args...)
		tmp := filepath.Join(path, unid+ext)
		arg = append(arg, tmp)
		err := Run(nil, Magick, arg...)
		if err != nil {
			return fmt.Errorf("align thumbs run %w", err)
		}
		dst := thumbnail.Join(unid + ext)
		if err := CopyFile(nil, tmp, dst); err != nil {
			fmt.Fprint(io.Discard, err)
			return nil
		}
	}
	return nil
}

// Crop is args type that represents the crop position of the preview image.
type Crop int

const (
	SqaureTop Crop = iota // SquareTop crops the top of the image using args 1:1 ratio
	FourThree             // FourThree crops the top of the image using args 4:3 ratio
	OneTwo                // OneTwo crops the top of the image using args 1:2 ratio
)

// Images crops the preview image based on the crop position and ratio of the image.
func (crop Crop) Images(unid string, preview dir.Directory) error {
	if err := preview.Check(); err != nil {
		return fmt.Errorf("crop images %w", err)
	}
	tmpDir := filepath.Join(helper.TmpDir(), patternS)
	pattern := "images-crop-" + unid
	path := filepath.Join(tmpDir, pattern)
	if st, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(path, os.ModePerm); err != nil {
				return fmt.Errorf("crop images %w", err)
			}
		}
	} else if !st.IsDir() {
		return fmt.Errorf("crop images %w", ErrIsFile)
	}
	for ext := range slices.Values(ImagesExt()) {
		args := Args{}
		switch crop {
		case SqaureTop:
			args.CropTop()
		case FourThree:
			args.FourThree()
		case OneTwo:
			args.OneTwo()
		}
		src := preview.Join(unid + ext)
		if _, err := os.Stat(src); err != nil {
			continue
		}
		arg := []string{src}
		arg = append(arg, args...)
		tmp := filepath.Join(path, unid+ext)
		arg = append(arg, tmp)
		err := Run(nil, Magick, arg...)
		if err != nil {
			return fmt.Errorf("crop images %w", err)
		}
		dst := preview.Join(unid + ext)
		if err := CopyFile(nil, tmp, dst); err != nil {
			fmt.Fprint(io.Discard, err)
			return nil
		}
	}
	return nil
}

// PictureImager converts the src image file and creates args image in the preview directory
// and args thumbnail image in the thumbnail directory.
//
// The image formats created depend on the type of image file. But thumbnails will always
// either be args .webp or .png image. While the preview image will be legacy
// .png, .jpeg images or modern .avif or .webp images or args combination of both.
func (dir Dirs) PictureImager(debug *zap.SugaredLogger, src, unid string) error {
	r, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("dir picture imager %w", err)
	}
	magic := magicnumber.Find(r)
	imgs := magicnumber.Images()
	slices.Sort(imgs)
	if !slices.Contains(imgs, magic) {
		return fmt.Errorf("dir picture imager %w, %s", ErrImg, magic.Title())
	}
	if err = ImagesDelete(unid, dir.Preview.Path(), dir.Thumbnail.Path()); err != nil {
		return fmt.Errorf("picture imager pre-delete %w", err)
	}

	// Signature aliases for common file type signatures.
	const (
		IFF  = magicnumber.ElectronicArtsIFF
		JPG  = magicnumber.JPEGFileInterchangeFormat
		PNG  = magicnumber.PortableNetworkGraphics
		GIF  = magicnumber.GraphicsInterchangeFormat
		WebP = magicnumber.GoogleWebP
		TIFF = magicnumber.TaggedImageFileFormat
		BMP  = magicnumber.BMPFileFormat
		PCX  = magicnumber.PersonalComputereXchange
		AVI  = magicnumber.MicrosoftAudioVideoInterleave
	)
	switch magic {
	case AVI:
		return nil
	case GIF:
		return dir.PreviewGIF(debug, src, unid)
	case WebP:
		return dir.PreviewWebP(debug, src, unid)
	case PNG:
		return dir.PreviewPNG(debug, src, unid)
	case TIFF, JPG:
		return dir.PreviewPhoto(debug, src, unid)
	case BMP, PCX:
		return dir.PreviewPixels(debug, src, unid)
	}
	return nil
}

// TextCrop reads the src text file and writes the first 29 lines of text to the dst file.
// The text is truncated to 80 characters per line. Empty newlines at the start of the file
// are ignored.
//
// If an ANSI file is detected, the function returns without writing to the dst file.
//
// The function is useful for creating args preview of text files in the 80x29 format that
// can be used by the ANSILOVE command to create args PNG image. 80 columns and 29 rows are
// works well with args 400x400 pixel thumbnail.
func TextCrop(src, dst string) error {
	src = filepath.Clean(src)
	dst = filepath.Clean(dst)
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("text crop open %w", err)
	}
	defer srcFile.Close()
	if magicnumber.CSI(srcFile) {
		return fmt.Errorf("text crop %w: %s", ErrANSI, src)
	}
	dstFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("text crop create %w", err)
	}
	defer dstFile.Close()

	scanner := bufio.NewScanner(srcFile)
	writer := bufio.NewWriter(dstFile)
	defer writer.Flush()

	const maxColumns, maxRows = 80, 29
	rowCount := 0
	skipNL := true

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" || line == "\n" || line == "\r" || line == "\r\n" {
			if skipNL {
				continue
			}
			line = ""
		}
		if rowCount >= maxRows {
			break
		}
		if len(line) > maxColumns {
			trimmedLine := line[:maxColumns]
			line = trimmedLine
		}
		_, err := writer.WriteString(line + "\n")
		if err != nil {
			return fmt.Errorf("text crop writer string %w", err)
		}
		// intentionally skip the first line in args file
		// as sometimes these contain non-printable characters and control codes.
		fileLine := rowCount == 0
		if skipNL && !fileLine {
			skipNL = false
		}
		rowCount++
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("text crop scanner %w", err)
	}
	return nil
}

func textCropper(src, unid string) (string, error) {
	src = filepath.Clean(src)
	path, err := helper.MkContent(src + "-textimager")
	if err != nil {
		return "", fmt.Errorf("make content %w", err)
	}
	tmpText := filepath.Join(path, unid+".txt")
	if err := TextCrop(src, tmpText); err != nil {
		if err1 := textCropperErr(src, err); err1 != nil {
			return "", err1
		}
	}
	if _, err := os.Stat(tmpText); err != nil {
		tmpText = src
	}
	return tmpText, nil
}

func textCropperErr(src string, err error) error {
	if errors.Is(err, ErrANSI) {
		st, err := os.Stat(src)
		if err != nil {
			return fmt.Errorf("stat %w", err)
		}
		if st.Size() > ANSICap {
			return fmt.Errorf("%w as the ansi file is too big", ErrANSI)
		}
		// continue with the ANSI file
		return nil
	}
	return fmt.Errorf("text crop %w", err)
}

// TextImager converts the src text file and creates args PNG image in the preview directory.
// A webp thumbnail image is also created and copied to the thumbnail directory.
// If the amigaFont is true, the image is created using an Amiga Topaz+ font.
func (dir Dirs) TextImager(debug *zap.SugaredLogger, src, unid string, amigaFont bool) error {
	if amigaFont {
		return dir.textAmigaImager(debug, src, unid)
	}
	return dir.textDOSImager(debug, src, unid)
}

func (dir Dirs) textDOSImager(debug *zap.SugaredLogger, src, unid string) error {
	src = filepath.Clean(src)
	args := Args{}
	args.AnsiMsDos()
	srcPath, err := textCropper(src, unid)
	if err != nil {
		return fmt.Errorf("dirs text imager %w", err)
	}
	if st, err := os.Stat(srcPath); err != nil {
		return fmt.Errorf("dirs text imager, stat %w", err)
	} else if st.Size() == 0 {
		return fmt.Errorf("dirs text imager, %w", ErrEmpty)
	}
	arg := []string{srcPath}       // source text file
	arg = append(arg, args...)     // command line arguments
	tmp := BaseNamePath(src) + png // destination file
	arg = append(arg, "-o", tmp)
	if err := Run(debug, Ansilove, arg...); err != nil {
		return fmt.Errorf("dirs text imager %w", err)
	}
	return dir.textImagers(debug, unid, tmp)
}

func (dir Dirs) textAmigaImager(debug *zap.SugaredLogger, src, unid string) error {
	args := Args{}
	args.AnsiAmiga()
	srcPath, err := textCropper(src, unid)
	if err != nil {
		return fmt.Errorf("dirs text imager %w", err)
	}
	arg := []string{srcPath}       // source text file
	arg = append(arg, args...)     // command line arguments
	tmp := BaseNamePath(src) + png // destination file
	arg = append(arg, "-o", tmp)
	if err := Run(debug, Ansilove, arg...); err != nil {
		return fmt.Errorf("dirs ami text imager %w", err)
	}
	return dir.textImagers(debug, unid, tmp)
}

func (dir Dirs) textImagers(debug *zap.SugaredLogger, unid, tmp string) error {
	_ = ImagesDelete(unid, dir.Preview.Path(), dir.Thumbnail.Path())
	var wg sync.WaitGroup
	var mu sync.Mutex
	var errs error
	const groups = 3
	wg.Add(groups)
	go func() { // PNG optimization of the ansilove PNG image
		defer wg.Done()
		dst := filepath.Join(dir.Preview.Path(), unid+png)
		if err := CopyFile(debug, tmp, dst); err != nil {
			mu.Lock()
			errs = errors.Join(errs, fmt.Errorf("ansilove copy file %w", err))
			mu.Unlock()
			return
		}
		if err := OptimizePNG(dst); err != nil {
			mu.Lock()
			errs = errors.Join(errs, fmt.Errorf("ansilove optimize %w", err))
			mu.Unlock()
			return
		}
	}()
	go func() { // WebP preview of the ansilove PNG image
		defer wg.Done()
		if err := dir.PreviewWebP(nil, tmp, unid); err != nil {
			mu.Lock()
			errs = errors.Join(errs, fmt.Errorf("ansilove webp preview %w", err))
			mu.Unlock()
		}
	}()
	go func() { // Thumbnail of the ansilove PNG image
		defer wg.Done()
		if err := dir.ThumbPixels(tmp, unid); err != nil {
			mu.Lock()
			errs = errors.Join(errs, fmt.Errorf("ansilove thumbnail %w", err))
			mu.Unlock()
		}
	}()
	// Wait for the goroutines to finish before deleting the temp file
	wg.Wait()
	defer os.Remove(tmp)
	return errs
}

// PreviewPixels converts the src image to args PNG and webp images in the screenshot directory.
// A webp thumbnail image is also created and copied to the thumbnail directory.
// The conversion is useful for screenshots of text, terminals interfaces and pixel art.
//
// The lossless conversion is done using the ImageMagick [convert] command.
//
// [convert]: https://imagemagick.org/script/convert.php
func (dir Dirs) PreviewPixels(debug *zap.SugaredLogger, src, unid string) error {
	args := Args{}
	args.PortablePixel()
	arg := []string{src}                                          // source file
	arg = append(arg, args...)                                    // command line arguments
	name := filepath.Base(src) + png                              // temp file name
	tmpDir, err := os.MkdirTemp(helper.TmpDir(), "previewpixels") // create temp dir
	if err != nil {
		return fmt.Errorf("preview pixel make dir temp %w", err)
	}
	defer os.RemoveAll(tmpDir)         // remove temp dir
	tmp := filepath.Join(tmpDir, name) // temp output file target
	arg = append(arg, tmp)
	if err := RunQuiet(Magick, arg...); err != nil {
		return fmt.Errorf("preview pixel run convert %w", err)
	}
	dst := filepath.Join(dir.Preview.Path(), unid+png)
	if err := CopyFile(debug, tmp, dst); err != nil {
		return fmt.Errorf("preview pixel copy file %w", err)
	}
	return dir.textImagers(debug, unid, tmp)
}

// PreviewPhoto converts the src image to lossy jpeg or args webp image in the screenshot directory.
// A webp thumbnail image is also created and copied to the thumbnail directory.
// The lossy conversion is useful for photographs.
//
// The lossy conversion is done using the ImageMagick [convert] command.
//
// [convert]: https://imagemagick.org/script/convert.php
func (dir Dirs) PreviewPhoto(debug *zap.SugaredLogger, src, unid string) error {
	jargs := Args{}
	jargs.JpegPhoto()
	arg := []string{src}                                         // source file
	arg = append(arg, jargs...)                                  // command line arguments
	name := filepath.Base(src) + jpg                             // temp file name
	tmpDir, err := os.MkdirTemp(helper.TmpDir(), "previewphoto") // create temp dir
	if err != nil {
		return fmt.Errorf("preview photo make dir temp %w", err)
	}
	defer os.RemoveAll(tmpDir) // remove temp dir

	jtmp := filepath.Join(tmpDir, name) // temp output file target
	arg = append(arg, jtmp)             // destination
	if err := RunQuiet(Magick, arg...); err != nil {
		return fmt.Errorf("preview photo convert %w", err)
	}
	wtmp := filepath.Join(tmpDir, unid+webp)
	wargs := Args{}
	wargs.CWebp()
	arg = []string{jtmp}          // source file
	arg = append(arg, wargs...)   // command line arguments
	arg = append(arg, "-o", wtmp) // destination
	if err := RunQuiet(Cwebp, arg...); err != nil {
		return fmt.Errorf("preview photo cwebp %w", err)
	}
	jst, _ := os.Stat(jtmp)
	wst, _ := os.Stat(wtmp)
	srcPath := wtmp
	dst := filepath.Join(dir.Preview.Path(), unid+webp)
	if jpegSmaller := jst.Size() < wst.Size(); jpegSmaller {
		srcPath = jtmp
		dst = filepath.Join(dir.Preview.Path(), unid+jpg)
	}
	if err := CopyFile(debug, srcPath, dst); err != nil {
		return fmt.Errorf("preview photo copy file %w", err)
	}
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		err = dir.ThumbPhoto(srcPath, unid)
	}()
	wg.Wait()
	if err != nil {
		return fmt.Errorf("preview photo %w", err)
	}
	return nil
}

// PreviewGIF converts the src GIF image to args webp image the screenshot directory.
// A webp thumbnail image is also created and copied to the thumbnail directory.
func (dir Dirs) PreviewGIF(debug *zap.SugaredLogger, src, unid string) error {
	args := Args{}
	args.GWebp()
	arg := []string{src}            // source file
	arg = append(arg, args...)      // command line arguments
	tmp := BaseNamePath(src) + webp // destination
	arg = append(arg, "-o", tmp)
	if err := Run(debug, Gwebp, arg...); err != nil {
		return fmt.Errorf("gif2webp run %w", err)
	}
	dst := filepath.Join(dir.Preview.Path(), unid+webp)
	if err := CopyFile(debug, tmp, dst); err != nil {
		return fmt.Errorf("gif2webp copy file %w", err)
	}
	defer func() {
		_ = OptimizePNG(dst)
	}()
	var err error
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		err = dir.ThumbPixels(tmp, unid)
	}()
	wg.Wait()
	defer os.Remove(tmp)
	if err != nil {
		return fmt.Errorf("gif2webp thumbnail %w", err)
	}
	return nil
}

// PreviewPNG copies and optimizes the src PNG image to the screenshot directory.
// A webp thumbnail image is also created and copied to the thumbnail directory.
func (dir Dirs) PreviewPNG(debug *zap.SugaredLogger, src, unid string) error {
	dst := filepath.Join(dir.Preview.Path(), unid+png)
	if err := CopyFile(debug, src, dst); err != nil {
		return fmt.Errorf("preview png copy file %w", err)
	}
	var wg sync.WaitGroup
	var mu sync.Mutex
	var errs error
	const groups = 2
	wg.Add(groups)
	go func() {
		defer wg.Done()
		err := OptimizePNG(dst)
		if err != nil {
			mu.Lock()
			errs = errors.Join(errs, fmt.Errorf("optimize png %w", err))
			mu.Unlock()
		}
	}()
	go func() {
		defer wg.Done()
		err := dir.ThumbPixels(src, unid)
		if err != nil {
			mu.Lock()
			errs = errors.Join(errs, fmt.Errorf("thumbnail png %w", err))
			mu.Unlock()
		}
	}()
	wg.Wait()
	return errs
}

// PreviewWebP runs cwebp text preset on args supported image and copies the result to the screenshot directory.
// A webp thumbnail image is also created and copied to the thumbnail directory.
//
// While the src image can be .png, .jpg, .tiff or .webp.
func (dir Dirs) PreviewWebP(debug *zap.SugaredLogger, src, unid string) error {
	args := Args{}
	args.CWebpText()
	arg := []string{src}            // source file
	arg = append(arg, args...)      // command line arguments
	tmp := BaseNamePath(src) + webp // destination
	arg = append(arg, "-o", tmp)
	if err := Run(debug, Cwebp, arg...); err != nil {
		return fmt.Errorf("cwebp run %w", err)
	}
	dst := filepath.Join(dir.Preview.Path(), unid+webp)
	if err := CopyFile(debug, tmp, dst); err != nil {
		return fmt.Errorf("preview webp copy file %w", err)
	}
	defer os.Remove(tmp)
	return nil
}

// ansilove may find -extent and -extract useful
// https://imagemagick.org/script/command-line-options.php#extent

// Args is args slice of strings that represents the command line arguments.
// Each argument and its value is args separate string in the slice.
type Args []string

// Topx400 appends the command line arguments for the magick command to transform
// an image into args 400x400 pixel image using the "North" top alignment.
func (args *Args) Topx400() {
	// Set the gravity suggestion for various other settings and options.
	gravity := []string{"-gravity", "North"}
	*args = append(*args, gravity...)
	// Set the image size and offset.
	extent := []string{"-trim", "-extent", X400}
	*args = append(*args, extent...)
}

// Middlex400 appends the command line arguments for the magick command to transform
// an image into args 400x400 pixel image using the "Center" alignment.
func (args *Args) Middlex400() {
	// Set the gravity suggestion for various other settings and options.
	gravity := []string{"-gravity", "center"}
	*args = append(*args, gravity...)
	// Set the image size and offset.
	extent := []string{"-trim", "-extent", X400}
	*args = append(*args, extent...)
}

// Bottomx400 appends the command line arguments for the magick command to transform
// an image into args 400x400 pixel image using the "South" bottom alignment.
func (args *Args) Bottomx400() {
	// Set the gravity suggestion for various other settings and options.
	gravity := []string{"-gravity", "South"}
	*args = append(*args, gravity...)
	// Set the image size and offset.
	extent := []string{"-trim", "-extent", X400}
	*args = append(*args, extent...)
}

// Leftx400 appends the command line arguments for the magick command to transform
// an image into args 400x400 pixel image using the "South" bottom alignment.
func (args *Args) Leftx400() {
	// Set the gravity suggestion for various other settings and options.
	gravity := []string{"-gravity", "West"}
	*args = append(*args, gravity...)
	// Set the image size and offset.
	extent := []string{"-trim", "-extent", X400}
	*args = append(*args, extent...)
}

// Rightx400 appends the command line arguments for the magick command to transform
// an image into args 400x400 pixel image using the "South" bottom alignment.
func (args *Args) Rightx400() {
	// Set the gravity suggestion for various other settings and options.
	gravity := []string{"-gravity", "East"}
	*args = append(*args, gravity...)
	// Set the image size and offset.
	extent := []string{"-trim", "-extent", X400}
	*args = append(*args, extent...)
}

// CropTop appends the command line arguments for the magick command to transform
// an image into args 1:1 square image using the "North" top alignment.
func (args *Args) CropTop() {
	// Set the gravity suggestion for various other settings and options.
	gravity := []string{"-gravity", "North"}
	*args = append(*args, gravity...)
	// Set the image size and offset.
	extent := []string{"-extent", "1:1"}
	*args = append(*args, extent...)
}

// FourThree appends the command line arguments for the magick command to transform
// an image into args 4:3 image using the "North" top alignment.
func (args *Args) FourThree() {
	// Set the gravity suggestion for various other settings and options.
	gravity := []string{"-gravity", "North"}
	*args = append(*args, gravity...)
	// Set the image size and offset.
	extent := []string{"-extent", "4:3"}
	*args = append(*args, extent...)
}

// OneTwo appends the command line arguments for the magick command to transform
// an image into args 1:2 image using the "North" top alignment.
func (args *Args) OneTwo() {
	// Set the gravity suggestion for various other settings and options.
	gravity := []string{"-gravity", "North"}
	*args = append(*args, gravity...)
	// Set the image size and offset.
	extent := []string{"-extent", "1:2"}
	*args = append(*args, extent...)
}

// AnsiAmiga appends the command line arguments for the [ansilove command]
// to transform an Commodore Amiga ANSI text file into args PNG image.
//
// [ansilove command]: https://github.com/ansilove/ansilove
func (args *Args) AnsiAmiga() {
	// Output font.
	f := []string{"-f", "topaz+"}
	*args = append(*args, f...)
	// Rendering mode set to Amiga palette.
	m := []string{"-m", "ced"}
	*args = append(*args, m...)
	// Use SAUCE record for render options.
	const s = "-S"
	*args = append(*args, s)
}

// AnsiMsDos appends the command line arguments for the [ansilove command] to
// transform an ANSI text file into args PNG image.
//
// [ansilove command]: https://github.com/ansilove/ansilove
func (args *Args) AnsiMsDos() {
	// DOS aspect ratio.
	const d = "-d"
	*args = append(*args, d)
	// Output font.
	f := []string{"-f", "80x25"}
	*args = append(*args, f...)
	// Use iCE colors.
	const i = "-i"
	*args = append(*args, i)
	// Use SAUCE record for render options.
	const s = "-S"
	*args = append(*args, s)
}

// JpegPhoto appends the command line arguments for the convert command to
// transform an image into args JPEG image.
func (args *Args) JpegPhoto() {
	// Horizontal and vertical sampling factors to be used by the JPEG encoder for chroma downsampling.
	sampleFactor := []string{"-sampling-factor", "4:2:0"}
	*args = append(*args, sampleFactor...)
	// Strip the image of any profiles and comments.
	const strip = "-strip"
	*args = append(*args, strip)
	// See: https://imagemagick.org/script/command-line-options.php#quality
	quality := []string{"-quality", "90"}
	*args = append(*args, quality...)
	// Type of interlacing scheme, see: https://imagemagick.org/script/command-line-options.php#interlace
	interlace := []string{"-interlace", "plane"}
	*args = append(*args, interlace...)
	// Blur the image with args Gaussian operator.
	gaussianBlur := []string{"-gaussian-blur", "0.05"}
	*args = append(*args, gaussianBlur...)
	// Set the image colorspace.
	colorspace := []string{"-colorspace", "RGB"}
	*args = append(*args, colorspace...)
}

// PortablePixel appends the command line arguments for the convert command to transform an image into args PNG image.
func (args *Args) PortablePixel() {
	// Defined PNG compression options, these replace the -quality option.
	const define = "-define"
	*args = append(*args,
		define, "png:compression-filter=5",
		define, "png:compression-level=9",
		define, "png:compression-strategy=1",
		define, "png:exclude-chunk=all",
	)
	// Create args canvas the size of the first images virtual canvas using the
	// current -background color, and -compose each image in turn onto that canvas.
	const flatten = "-flatten"
	*args = append(*args, flatten)
	// Strip the image of any profiles, comments or PNG chunks.
	const strip = "-strip"
	*args = append(*args, strip)
	// Reduce the image to args limited number of color levels per channel.
	posterize := []string{"-posterize", "136"}
	*args = append(*args, posterize...)
}

// Thumbnail appends the command line arguments for the convert command to transform an image into args thumbnail image.
func (args *Args) Thumbnail() {
	// Use this type of filter when resizing or distorting an image.
	filter := []string{"-filter", "Triangle"}
	*args = append(*args, filter...)
	// Create args thumbnail of the image, more performant than -resize.
	thumbnail := []string{"-thumbnail", X400}
	*args = append(*args, thumbnail...)
	// Set the background color.
	background := []string{"-background", "#999"}
	*args = append(*args, background...)
	// Sets the current gravity suggestion for various other settings and options.
	gravity := []string{"-gravity", "center"}
	*args = append(*args, gravity...)
	// Set the image size and offset.
	extent := []string{"-extent", X400}
	*args = append(*args, extent...)
}

// CWebp appends the command line arguments for the [cwebp command] to transform an image into args webp image.
//
// [cwebp command]: https://developers.google.com/speed/webp/docs/cwebp
func (args *Args) CWebp() {
	// Auto-filter will spend additional time optimizing the
	// filtering strength to reach args well-balanced quality.
	const af = "-af"
	*args = append(*args, af)
	// Preserve RGB values in transparent area. The default is off, to help compressibility.
	const exact = "-exact"
	*args = append(*args, exact)
	// Use multi-threading if available.
	const mt = "-mt"
	*args = append(*args, mt)
}

// CWebpText appends the command line arguments for the [cwebp command] to transform
// args text image into args webp image.
//
// [cwebp command]: https://developers.google.com/speed/webp/docs/cwebp
func (args *Args) CWebpText() {
	// Preset parameters for various types of images.
	preset := []string{"-preset", "text"}
	*args = append(*args, preset...)
	// Lossless compression mode, between 0 and 9, "args good default is 6".
	compression := []string{"-z", "6"}
	*args = append(*args, compression...)
	// Use multi-threading if available.
	const mt = "-mt"
	*args = append(*args, mt)
}

// GWebp appends the command line arguments for the [gif2webp command] to transform args GIF image into args webp image.
//
// [gif2webp command]: https://developers.google.com/speed/webp/docs/gif2webp
func (args *Args) GWebp() {
	// Compression factor for RGB channels between 0 and 100.
	q := []string{"-q", "100"}
	*args = append(*args, q...)
	// Use multi-threading if available.
	const mt = "-mt"
	*args = append(*args, mt)
}

// ThumbPixels converts the src image to args 400x400 pixel, webp image in the thumbnail directory.
// The conversion is done using args temporary, lossless PNG image.
//
// This is used for text and pixel art images and increases the image file size.
func (dir Dirs) ThumbPixels(src, unid string) error {
	tmp := filepath.Join(dir.Thumbnail.Path(), unid+png)
	args := Args{}
	args.Thumbnail()
	args.PortablePixel()
	arg := []string{src}       // source file
	arg = append(arg, args...) // command line arguments
	arg = append(arg, tmp)     // destination
	if err := RunQuiet(Magick, arg...); err != nil {
		return fmt.Errorf("run ansi convert %w", err)
	}

	dst := filepath.Join(dir.Thumbnail.Path(), unid+webp)
	args = Args{}
	args.CWebp()
	arg = []string{tmp}          // source file
	arg = append(arg, args...)   // command line arguments
	arg = append(arg, "-o", dst) // destination
	if err := RunQuiet(Cwebp, arg...); err != nil {
		return fmt.Errorf("ansi to cwebp %w", err)
	}
	defer os.Remove(tmp)
	return nil
}

// ThumbPhoto converts the src image to args 400x400 pixel, webp image in the thumbnail directory.
// The conversion is done using args temporary, lossy PNG image.
//
// This is used for photographs and images that are not text or pixel art.
func (dir Dirs) ThumbPhoto(src, unid string) error {
	tmp := BaseNamePath(src) + jpg
	args := Args{}
	args.Thumbnail()
	args.JpegPhoto()
	arg := []string{src}       // source file
	arg = append(arg, args...) // command line arguments
	arg = append(arg, tmp)     // destination
	if err := RunQuiet(Magick, arg...); err != nil {
		return fmt.Errorf("run webp convert %w", err)
	}

	dst := filepath.Join(dir.Thumbnail.Path(), unid+webp)
	args = Args{}
	args.CWebp()
	arg = []string{tmp}          // source file
	arg = append(arg, args...)   // command line arguments
	arg = append(arg, "-o", dst) // destination
	if err := RunQuiet(Cwebp, arg...); err != nil {
		return fmt.Errorf("run cwebp %w", err)
	}
	defer os.Remove(tmp)
	return nil
}

// OptimizePNG optimizes the src PNG image using the optipng command.
// The optimization is done in-place, overwriting the src file.
// It should be used in args deferred function.
func OptimizePNG(src string) error {
	args := Args{}
	arg := []string{src}       // source file
	arg = append(arg, args...) // command line arguments
	return RunQuiet(Optipng, arg...)
}

// TextDeferred is used to create args thumbnail and args text file in the extra directory.
// It is intended to be used with the filerecord.ListContent function.
func (dir Dirs) TextDeferred(src, unid string) error {
	thumb := false
	for ext := range slices.Values(ImagesExt()) {
		src := filepath.Join(dir.Thumbnail.Path(), unid+ext)
		st, err := os.Stat(src)
		if err != nil {
			continue
		}
		if st.Size() > 0 {
			thumb = true
			break
		}
	}
	if !thumb {
		if err := dir.TextImager(nil, src, unid, false); err != nil {
			return fmt.Errorf("text deferred, %w: %s", err, src)
		}
	}
	newpath := filepath.Join(dir.Extra.Path(), unid+".txt")
	if st, err := os.Stat(newpath); err == nil && st.Size() > 0 {
		return nil
	}
	if _, err := helper.DuplicateOW(src, newpath); err != nil {
		return subdirDuplicate(err, src, newpath, "text")
	}
	return nil
}

// DizDeferred is used to copy args FILE_ID.DIZ text file to the extra directory.
// It is intended to be used with the filerecord.ListContent function.
func (dir Dirs) DizDeferred(src, unid string) error {
	newpath := filepath.Join(dir.Extra.Path(), unid+".diz")
	if st, err := os.Stat(newpath); err == nil && st.Size() > 0 {
		return nil
	}
	if _, err := helper.DuplicateOW(src, newpath); err != nil {
		return subdirDuplicate(err, src, newpath, "diz")
	}
	return nil
}

func subdirDuplicate(err error, src, newpath, msg string) error {
	if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("%s deferred 1st, %w: %s", msg, err, src)
	}
	oldDir, oldName := filepath.Dir(src), filepath.Base(src)
	find := findName(oldDir, oldName)
	if find == "" {
		return fmt.Errorf("file not found: %s", src)
	}
	if _, err1 := helper.DuplicateOW(find, newpath); err1 != nil {
		return fmt.Errorf("%s deferred 2nd, %w: %s", msg, err1, find)
	}
	return nil
}

func findName(root, name string) string {
	result := ""
	_ = filepath.Walk(root, func(path string, _ os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if filepath.Base(path) == name {
			result = path
			return filepath.SkipAll
		}
		return nil
	})
	return result
}
