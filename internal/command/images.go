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
	"go.uber.org/zap"
)

const X400 = "400x400" // X400 returns a  400 x 400 pixel image size

// ImagesExt returns a slice of image file extensions used by the website
// preview and thumbnail images, including the legacy and modern formats.
func ImagesExt() []string {
	return []string{gif, jpg, jpeg, png, webp, ".avif"}
}

// ImagesDelete removes images from the specified directories that match the unid.
// The unid is the unique identifier for the image file and shared between the preview
// and thumbnail images.
func ImagesDelete(unid string, dirs ...string) error {
	for _, dir := range dirs {
		st, err := os.Stat(dir)
		if err != nil {
			return fmt.Errorf("images delete %w", err)
		}
		if !st.IsDir() {
			return fmt.Errorf("images delete %w", ErrIsFile)
		}
		for _, ext := range ImagesExt() {
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

// Pixelate appends the command line arguments for the convert command to transform an image into a PNG image.
func (a *Args) Pixelate() {
	// Create a canvas the size of the first images virtual canvas using the
	// current -background color, and -compose each image in turn onto that canvas.
	scale5 := []string{"-scale", "5%"}
	*a = append(*a, scale5...)
	scale2K := []string{"-scale", "2000%"}
	*a = append(*a, scale2K...)
}

// ImagesPixelate converts the images in the specified directories to pixelated images.
// The unid is the unique identifier for the image file and shared between the preview
// and thumbnail images.
func ImagesPixelate(unid string, dirs ...string) error {
	for _, dir := range dirs {
		st, err := os.Stat(dir)
		if err != nil {
			return fmt.Errorf("images delete %w", err)
		}
		if !st.IsDir() {
			return fmt.Errorf("images delete %w", ErrIsFile)
		}
		for _, ext := range ImagesExt() {
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

// Thumb is a type that represents the type of thumbnail image to create.
type Thumb int

const (
	Pixel Thumb = iota // Pixel art or images with text
	Photo              // Photographs or images with gradients
)

// Thumbs creates a thumbnail image for the preview image based on the type of image.
func (dir Dirs) Thumbs(unid string, thumb Thumb) error {
	if err := ImagesDelete(unid, dir.Thumbnail); err != nil {
		return fmt.Errorf("dirs thumbs %w", err)
	}
	for _, ext := range ImagesExt() {
		src := filepath.Join(dir.Preview, unid+ext)
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

// Align is a type that represents the alignment of the thumbnail image.
type Align int

const (
	Top    Align = iota // Top uses the top alignment of the preview image
	Middle              // Middle uses the center alignment of the preview image
	Bottom              // Bottom uses the bottom alignment of the preview image
	Left                // Left uses the left alignment of the preview image
	Right               // Right uses the right alignment of the preview image
)

// Thumbs creates a thumbnail image for the preview image based on the crop position of the image.
func (align Align) Thumbs(unid, previewDir, thumbnailDir string) error {
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
	if err := ImagesDelete(unid, thumbnailDir); err != nil {
		return fmt.Errorf("dirs thumbs %w", err)
	}
	for _, ext := range ImagesExt() {
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
		src := filepath.Join(previewDir, unid+ext)
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
		dst := filepath.Join(thumbnailDir, unid+ext)
		if err := CopyFile(nil, tmp, dst); err != nil {
			fmt.Fprint(io.Discard, err)
			return nil
		}
	}
	return nil
}

// Crop is a type that represents the crop position of the preview image.
type Crop int

const (
	SqaureTop Crop = iota // SquareTop crops the top of the image using a 1:1 ratio
	FourThree             // FourThree crops the top of the image using a 4:3 ratio
	OneTwo                // OneTwo crops the top of the image using a 1:2 ratio
)

// Images crops the preview image based on the crop position and ratio of the image.
func (crop Crop) Images(unid, previewDir string) error {
	st, err := os.Stat(previewDir)
	if err != nil {
		return fmt.Errorf("crop images %w", err)
	}
	if !st.IsDir() {
		return fmt.Errorf("crop images %w", ErrIsFile)
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
	for _, ext := range ImagesExt() {
		args := Args{}
		switch crop {
		case SqaureTop:
			args.CropTop()
		case FourThree:
			args.FourThree()
		case OneTwo:
			args.OneTwo()
		}
		src := filepath.Join(previewDir, unid+ext)
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
		dst := filepath.Join(previewDir, unid+ext)
		if err := CopyFile(nil, tmp, dst); err != nil {
			fmt.Fprint(io.Discard, err)
			return nil
		}
	}
	return nil
}

// PictureImager converts the src image file and creates a image in the preview directory
// and a thumbnail image in the thumbnail directory.
//
// The image formats created depend on the type of image file. But thumbnails will always
// either be a .webp or .png image. While the preview image will be legacy
// .png, .jpeg images or modern .avif or .webp images or a combination of both.
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
	if err = ImagesDelete(unid, dir.Preview, dir.Thumbnail); err != nil {
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

// Write80x29 reads the src text file and writes the first 29 lines of text to the dst file.
// The text is truncated to 80 characters per line. Empty newlines at the start of the file
// are ignored.
//
// The function is useful for creating a preview of text files in the 80x29 format that
// can be used by the ANSILOVE command to create a PNG image. 80 columns and 29 rows are
// works well with a 400x400 pixel thumbnail.
func Write80x29(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("write 80x29 open %w", err)
	}
	defer srcFile.Close()
	dstFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("write 80x29 create %w", err)
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
			return fmt.Errorf("write 80x29 writer string %w", err)
		}
		// intentionally skip the first line in a file
		// as sometimes these contain non-printable characters and control codes.
		fileLine := rowCount == 0
		if skipNL && !fileLine {
			skipNL = false
		}
		rowCount++
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("write 80x29 scanner %w", err)
	}
	return nil
}

// TextImager converts the src text file and creates a PNG image in the preview directory.
// A webp thumbnail image is also created and copied to the thumbnail directory.
func (dir Dirs) TextImager(debug *zap.SugaredLogger, src, unid string) error {
	args := Args{}
	args.AnsiMsDos()

	path, err := helper.MkContent(src)
	if err != nil {
		return err
	}
	defer os.RemoveAll(path)
	tmpText := filepath.Join(path, unid+".txt")
	if err = Write80x29(src, tmpText); err != nil {
		return fmt.Errorf("dirs text imager %w", err)
	}
	arg := []string{tmpText}       // source text file
	arg = append(arg, args...)     // command line arguments
	tmp := BaseNamePath(src) + png // destination file
	arg = append(arg, "-o", tmp)
	if err := Run(nil, Ansilove, arg...); err != nil {
		return fmt.Errorf("dirs text imager %w", err)
	}
	return dir.textImagers(debug, unid, tmp)
}

// TextImager converts the src text file and creates a PNG image using an Amiga Topaz+ font
// and stores it in the preview directory.
// A webp thumbnail image is also created and copied to the thumbnail directory.
func (dir Dirs) TextAmigaImager(debug *zap.SugaredLogger, src, unid string) error {
	args := Args{}
	args.AnsiMsDos()

	path, err := helper.MkContent(src)
	if err != nil {
		return err
	}
	defer os.RemoveAll(path)
	tmpText := filepath.Join(path, unid+".txt")
	if err = Write80x29(src, tmpText); err != nil {
		return fmt.Errorf("dirs text imager %w", err)
	}

	arg := []string{tmpText}       // source text file
	arg = append(arg, args...)     // command line arguments
	tmp := BaseNamePath(src) + png // destination file
	arg = append(arg, "-o", tmp)
	if err := Run(nil, Ansilove, arg...); err != nil {
		return fmt.Errorf("dirs ami text imager %w", err)
	}
	return dir.textImagers(debug, unid, tmp)
}

func (dir Dirs) textImagers(debug *zap.SugaredLogger, unid, tmp string) error {
	_ = ImagesDelete(unid, dir.Preview, dir.Thumbnail)
	var wg sync.WaitGroup
	var mu sync.Mutex
	var errs error
	const groups = 3
	wg.Add(groups)
	go func() { // PNG optimization of the ansilove PNG image
		defer wg.Done()
		dst := filepath.Join(dir.Preview, unid+png)
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

// PreviewPixels converts the src image to a PNG and webp images in the screenshot directory.
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
	dst := filepath.Join(dir.Preview, unid+png)
	if err := CopyFile(debug, tmp, dst); err != nil {
		return fmt.Errorf("preview pixel copy file %w", err)
	}
	return dir.textImagers(debug, unid, tmp)
}

// PreviewPhoto converts the src image to lossy jpeg or a webp image in the screenshot directory.
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
	dst := filepath.Join(dir.Preview, unid+webp)
	if jpegSmaller := jst.Size() < wst.Size(); jpegSmaller {
		srcPath = jtmp
		dst = filepath.Join(dir.Preview, unid+jpg)
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

// PreviewGIF converts the src GIF image to a webp image the screenshot directory.
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
	dst := filepath.Join(dir.Preview, unid+webp)
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
		err = dir.ThumbPhoto(tmp, unid)
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
	dst := filepath.Join(dir.Preview, unid+png)
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
		err := dir.ThumbPhoto(src, unid)
		if err != nil {
			mu.Lock()
			errs = errors.Join(errs, fmt.Errorf("thumbnail png %w", err))
			mu.Unlock()
		}
	}()
	wg.Wait()
	return errs
}

// PreviewWebP runs cwebp text preset on a supported image and copies the result to the screenshot directory.
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
	dst := filepath.Join(dir.Preview, unid+webp)
	if err := CopyFile(debug, tmp, dst); err != nil {
		return fmt.Errorf("preview webp copy file %w", err)
	}
	defer os.Remove(tmp)
	return nil
}

// ansilove may find -extent and -extract useful
// https://imagemagick.org/script/command-line-options.php#extent

// Args is a slice of strings that represents the command line arguments.
// Each argument and its value is a separate string in the slice.
type Args []string

// Topx400 appends the command line arguments for the magick command to transform
// an image into a 400x400 pixel image using the "North" top alignment.
func (a *Args) Topx400() {
	// Set the gravity suggestion for various other settings and options.
	gravity := []string{"-gravity", "North"}
	*a = append(*a, gravity...)
	// Set the image size and offset.
	extent := []string{"-trim", "-extent", X400}
	*a = append(*a, extent...)
}

// Middlex400 appends the command line arguments for the magick command to transform
// an image into a 400x400 pixel image using the "Center" alignment.
func (a *Args) Middlex400() {
	// Set the gravity suggestion for various other settings and options.
	gravity := []string{"-gravity", "center"}
	*a = append(*a, gravity...)
	// Set the image size and offset.
	extent := []string{"-trim", "-extent", X400}
	*a = append(*a, extent...)
}

// Bottomx400 appends the command line arguments for the magick command to transform
// an image into a 400x400 pixel image using the "South" bottom alignment.
func (a *Args) Bottomx400() {
	// Set the gravity suggestion for various other settings and options.
	gravity := []string{"-gravity", "South"}
	*a = append(*a, gravity...)
	// Set the image size and offset.
	extent := []string{"-trim", "-extent", X400}
	*a = append(*a, extent...)
}

// Leftx400 appends the command line arguments for the magick command to transform
// an image into a 400x400 pixel image using the "South" bottom alignment.
func (a *Args) Leftx400() {
	// Set the gravity suggestion for various other settings and options.
	gravity := []string{"-gravity", "West"}
	*a = append(*a, gravity...)
	// Set the image size and offset.
	extent := []string{"-trim", "-extent", X400}
	*a = append(*a, extent...)
}

// Rightx400 appends the command line arguments for the magick command to transform
// an image into a 400x400 pixel image using the "South" bottom alignment.
func (a *Args) Rightx400() {
	// Set the gravity suggestion for various other settings and options.
	gravity := []string{"-gravity", "East"}
	*a = append(*a, gravity...)
	// Set the image size and offset.
	extent := []string{"-trim", "-extent", X400}
	*a = append(*a, extent...)
}

// CropTop appends the command line arguments for the magick command to transform
// an image into a 1:1 square image using the "North" top alignment.
func (a *Args) CropTop() {
	// Set the gravity suggestion for various other settings and options.
	gravity := []string{"-gravity", "North"}
	*a = append(*a, gravity...)
	// Set the image size and offset.
	extent := []string{"-extent", "1:1"}
	*a = append(*a, extent...)
}

// FourThree appends the command line arguments for the magick command to transform
// an image into a 4:3 image using the "North" top alignment.
func (a *Args) FourThree() {
	// Set the gravity suggestion for various other settings and options.
	gravity := []string{"-gravity", "North"}
	*a = append(*a, gravity...)
	// Set the image size and offset.
	extent := []string{"-extent", "4:3"}
	*a = append(*a, extent...)
}

// OneTwo appends the command line arguments for the magick command to transform
// an image into a 1:2 image using the "North" top alignment.
func (a *Args) OneTwo() {
	// Set the gravity suggestion for various other settings and options.
	gravity := []string{"-gravity", "North"}
	*a = append(*a, gravity...)
	// Set the image size and offset.
	extent := []string{"-extent", "1:2"}
	*a = append(*a, extent...)
}

// AnsiAmiga appends the command line arguments for the [ansilove command]
// to transform an Commodore Amiga ANSI text file into a PNG image.
//
// [ansilove command]: https://github.com/ansilove/ansilove
func (a *Args) AnsiAmiga() {
	// Output font.
	f := []string{"-f", "topaz+"}
	*a = append(*a, f...)
	// Rendering mode set to Amiga palette.
	m := []string{"-m", "ced"}
	*a = append(*a, m...)
	// Use SAUCE record for render options.
	const s = "-S"
	*a = append(*a, s)
}

// AnsiMsDos appends the command line arguments for the [ansilove command] to
// transform an ANSI text file into a PNG image.
//
// [ansilove command]: https://github.com/ansilove/ansilove
func (a *Args) AnsiMsDos() {
	// DOS aspect ratio.
	const d = "-d"
	*a = append(*a, d)
	// Output font.
	f := []string{"-f", "80x25"}
	*a = append(*a, f...)
	// Use iCE colors.
	const i = "-i"
	*a = append(*a, i)
	// Use SAUCE record for render options.
	const s = "-S"
	*a = append(*a, s)
}

// JpegPhoto appends the command line arguments for the convert command to
// transform an image into a JPEG image.
func (a *Args) JpegPhoto() {
	// Horizontal and vertical sampling factors to be used by the JPEG encoder for chroma downsampling.
	sampleFactor := []string{"-sampling-factor", "4:2:0"}
	*a = append(*a, sampleFactor...)
	// Strip the image of any profiles and comments.
	const strip = "-strip"
	*a = append(*a, strip)
	// See: https://imagemagick.org/script/command-line-options.php#quality
	quality := []string{"-quality", "90"}
	*a = append(*a, quality...)
	// Type of interlacing scheme, see: https://imagemagick.org/script/command-line-options.php#interlace
	interlace := []string{"-interlace", "plane"}
	*a = append(*a, interlace...)
	// Blur the image with a Gaussian operator.
	gaussianBlur := []string{"-gaussian-blur", "0.05"}
	*a = append(*a, gaussianBlur...)
	// Set the image colorspace.
	colorspace := []string{"-colorspace", "RGB"}
	*a = append(*a, colorspace...)
}

// PortablePixel appends the command line arguments for the convert command to transform an image into a PNG image.
func (a *Args) PortablePixel() {
	// Defined PNG compression options, these replace the -quality option.
	const define = "-define"
	*a = append(*a,
		define, "png:compression-filter=5",
		define, "png:compression-level=9",
		define, "png:compression-strategy=1",
		define, "png:exclude-chunk=all",
	)
	// Create a canvas the size of the first images virtual canvas using the
	// current -background color, and -compose each image in turn onto that canvas.
	const flatten = "-flatten"
	*a = append(*a, flatten)
	// Strip the image of any profiles, comments or PNG chunks.
	const strip = "-strip"
	*a = append(*a, strip)
	// Reduce the image to a limited number of color levels per channel.
	posterize := []string{"-posterize", "136"}
	*a = append(*a, posterize...)
}

// Thumbnail appends the command line arguments for the convert command to transform an image into a thumbnail image.
func (a *Args) Thumbnail() {
	// Use this type of filter when resizing or distorting an image.
	filter := []string{"-filter", "Triangle"}
	*a = append(*a, filter...)
	// Create a thumbnail of the image, more performant than -resize.
	thumbnail := []string{"-thumbnail", X400}
	*a = append(*a, thumbnail...)
	// Set the background color.
	background := []string{"-background", "#999"}
	*a = append(*a, background...)
	// Sets the current gravity suggestion for various other settings and options.
	gravity := []string{"-gravity", "center"}
	*a = append(*a, gravity...)
	// Set the image size and offset.
	extent := []string{"-extent", X400}
	*a = append(*a, extent...)
}

// CWebp appends the command line arguments for the [cwebp command] to transform an image into a webp image.
//
// [cwebp command]: https://developers.google.com/speed/webp/docs/cwebp
func (a *Args) CWebp() {
	// Auto-filter will spend additional time optimizing the
	// filtering strength to reach a well-balanced quality.
	const af = "-af"
	*a = append(*a, af)
	// Preserve RGB values in transparent area. The default is off, to help compressibility.
	const exact = "-exact"
	*a = append(*a, exact)
	// Use multi-threading if available.
	const mt = "-mt"
	*a = append(*a, mt)
}

// CWebpText appends the command line arguments for the [cwebp command] to transform a text image into a webp image.
//
// [cwebp command]: https://developers.google.com/speed/webp/docs/cwebp
func (a *Args) CWebpText() {
	// Preset parameters for various types of images.
	preset := []string{"-preset", "text"}
	*a = append(*a, preset...)
	// Lossless compression mode, between 0 and 9, "a good default is 6".
	compression := []string{"-z", "6"}
	*a = append(*a, compression...)
	// Use multi-threading if available.
	const mt = "-mt"
	*a = append(*a, mt)
}

// GWebp appends the command line arguments for the [gif2webp command] to transform a GIF image into a webp image.
//
// [gif2webp command]: https://developers.google.com/speed/webp/docs/gif2webp
func (a *Args) GWebp() {
	// Compression factor for RGB channels between 0 and 100.
	q := []string{"-q", "100"}
	*a = append(*a, q...)
	// Use multi-threading if available.
	const mt = "-mt"
	*a = append(*a, mt)
}

// ThumbPixels converts the src image to a 400x400 pixel, webp image in the thumbnail directory.
// The conversion is done using a temporary, lossless PNG image.
func (dir Dirs) ThumbPixels(src, unid string) error {
	tmp := filepath.Join(dir.Thumbnail, unid+png)
	args := Args{}
	args.Thumbnail()
	args.PortablePixel()
	arg := []string{src}       // source file
	arg = append(arg, args...) // command line arguments
	arg = append(arg, tmp)     // destination
	if err := RunQuiet(Magick, arg...); err != nil {
		return fmt.Errorf("run ansi convert %w", err)
	}

	dst := filepath.Join(dir.Thumbnail, unid+webp)
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

// ThumbPhoto converts the src image to a 400x400 pixel, webp image in the thumbnail directory.
// The conversion is done using a temporary, lossy PNG image.
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

	dst := filepath.Join(dir.Thumbnail, unid+webp)
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
// It should be used in a deferred function.
func OptimizePNG(src string) error {
	args := Args{}
	arg := []string{src}       // source file
	arg = append(arg, args...) // command line arguments
	return RunQuiet(Optipng, arg...)
}
