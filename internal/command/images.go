package command

// Package file images.go contains the image conversion functions for
// converting images to PNG and WebP formats using ANSILOVE, ImageMagick
// and other command-line tools.

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/Defacto2/helper"
	"github.com/Defacto2/magicnumber"
	"github.com/Defacto2/server/internal/dir"
	"github.com/Defacto2/server/internal/panics"
	"golang.org/x/sync/errgroup"
)

const (
	AnsiCap = 350000    // ANSICap is the maximum file size in bytes for an ANSI encoded text file.
	X400    = "400x400" // X400 returns args 400 x 400 pixel image size
	argCap  = 2         // argCap is the fixed buffer size for command arguments (source + destination)
)

const (
	argExtent  = "-extent"
	argGravity = "-gravity"
	argTrim    = "-trim"
	north      = "North"
)

var ErrNoImages = errors.New("no images found")

// ImagesExt returns args slice of image file extensions used by the website
// preview and thumbnail images, including the legacy and modern formats.
func ImagesExt() []string {
	return []string{gif, jpg, jpeg, png, webp, ".avif"}
}

// ImagesDelete removes images matching the unid from the specified directories.
// The unid is the unique identifier shared across preview and thumbnail variants.
// Returns ErrNoImages if no matching files were found and removed.
func ImagesDelete(unid string, dirs ...string) error {
	const msg = "images delete"
	deletedAny := false
	// range each directory
	for _, dir := range dirs {
		st, err := os.Stat(dir)
		if err != nil {
			return fmt.Errorf("%s stat dir %s: %w", msg, dir, err)
		}
		if !st.IsDir() {
			return fmt.Errorf("%s %s: %w", msg, dir, ErrIsFile)
		}
		// range each image file extension looking for matches
		for _, ext := range ImagesExt() {
			name := filepath.Join(dir, unid+ext)
			if err := os.Remove(name); err == nil {
				deletedAny = true
			} else if !errors.Is(err, os.ErrNotExist) {
				return fmt.Errorf("%s remove %s: %w", msg, name, err)
			}
		}
	}
	if !deletedAny {
		return fmt.Errorf("%s: %w", msg, ErrNoImages)
	}
	return nil
}

// Pixelate appends ImageMagick flags to downscale and upscale an image,
// producing a blocky pixelated effect.
func (args *Args) Pixelate() {
	*args = append(*args, "-scale", "5%", "-scale", "2000%")
}

// ImagesPixelate converts images matching unid across the specified directories
// into pixelated versions in-place.
func ImagesPixelate(ctx context.Context, unid string, dirs ...string) error {
	const msg = "images pixelate"
	// range each directory
	for _, dir := range dirs {
		st, err := os.Stat(dir)
		if err != nil {
			return fmt.Errorf("%s stat dir %s: %w", msg, dir, err)
		}
		if !st.IsDir() {
			return fmt.Errorf("%s %s: %w", msg, dir, ErrIsFile)
		}
		// range each image file extension looking for matches
		for _, ext := range ImagesExt() {
			name := filepath.Join(dir, unid+ext)
			if _, err := os.Stat(name); err != nil {
				if errors.Is(err, os.ErrNotExist) {
					continue
				}
				return fmt.Errorf("%s stat image %s: %w", msg, name, err)
			}

			var flags Args
			flags.Pixelate()
			// construct ordered command arguments: [input, flags..., output]
			cmdArgs := make([]string, 0, 2+len(flags))
			cmdArgs = append(cmdArgs, name)
			cmdArgs = append(cmdArgs, flags...)
			cmdArgs = append(cmdArgs, name)
			if err := RunQuiet(ctx, Magick, cmdArgs...); err != nil {
				return fmt.Errorf("%s convert %s: %w", msg, name, err)
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

// Thumbs creates a thumbnail image from the corresponding preview image based on the thumb type.
func (dir Dirs) Thumbs(ctx context.Context, sl *slog.Logger, unid string, thumb Thumb) error {
	const msg = "thumbs"
	switch thumb {
	case Pixel, Photo:
	default:
		return fmt.Errorf("%s: invalid thumb type: %d", msg, thumb)
	}
	if err := dir.Thumbnail.Check(sl); err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}
	if err := dir.Preview.Check(sl); err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}

	// remove any existing thumbnails; ignore expected "not found" errors
	if err := ImagesDelete(unid, dir.Thumbnail.Path()); err != nil && !errors.Is(err, ErrNoImages) {
		return fmt.Errorf("%s delete existing: %w", msg, err)
	}
	// range each image file extension looking for matches
	for _, ext := range ImagesExt() {
		src := filepath.Join(dir.Preview.Path(), unid+ext)
		if _, err := os.Stat(src); err != nil {
			continue
		}
		var err error
		switch thumb {
		case Pixel:
			err = dir.thumbPixels(ctx, sl, src, unid)
		case Photo:
			err = dir.thumbPhoto(ctx, sl, src, unid)
		}
		if err != nil {
			return fmt.Errorf("%s conversion failed: %w", msg, err)
		}
		return nil
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

// Thumbs generates a cropped thumbnail from the source preview image using the specified alignment.
func (align Align) Thumbs(ctx context.Context, sl *slog.Logger, unid string, preview, thumbnail dir.Directory) error {
	const msg = "thumbnail realignment"
	if err := preview.Check(sl); err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}
	if err := thumbnail.Check(sl); err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}
	var args Args
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
	default:
		return fmt.Errorf("%s: invalid alignment: %d", msg, align)
	}
	// prep an isolated temporary directory
	tmpDir := filepath.Join(helper.TmpDir(), patternS, "images-thumb-"+unid)
	if err := os.MkdirAll(tmpDir, 0o755); err != nil {
		return fmt.Errorf("%s create temp dir: %w", msg, err)
	}
	defer func(sl *slog.Logger) {
		if err := os.RemoveAll(tmpDir); err != nil {
			slog.Info(msg, slog.String("tmp dir", tmpDir), slog.Any("os remove all error", err))
		}
	}(sl)
	// remove existing thumbnails
	if err := ImagesDelete(unid, thumbnail.Path()); err != nil && !errors.Is(err, ErrNoImages) {
		return fmt.Errorf("%s delete existing: %w", msg, err)
	}
	// range each image file extension looking for matches
	for _, ext := range ImagesExt() {
		src := preview.Join(unid + ext)
		if _, err := os.Stat(src); err != nil {
			continue
		}
		tmpDst := filepath.Join(tmpDir, unid+ext)
		// construct command arguments: [src, args..., tmpDst]
		cmdArgs := make([]string, 0, 2+len(args))
		cmdArgs = append(cmdArgs, src)
		cmdArgs = append(cmdArgs, args...)
		cmdArgs = append(cmdArgs, tmpDst)
		if err := Run(ctx, sl, Magick, cmdArgs...); err != nil {
			return fmt.Errorf("%s run magick: %w", msg, err)
		}
		// copy thumbs to their final destination
		finalDst := thumbnail.Join(unid + ext)
		if err := CopyFile(sl, tmpDst, finalDst); err != nil {
			return fmt.Errorf("%s copy file: %w", msg, err)
		}
		return nil
	}
	return fmt.Errorf("%s: %w", msg, ErrNoImages)
}

// Crop is an args type that represents the crop position of the preview image.
type Crop int

const (
	SquareTop Crop = iota // SquareTop crops the top of the image using args 1:1 ratio
	FourThree             // FourThree crops the top of the image using args 4:3 ratio
	OneTwo                // OneTwo crops the top of the image using args 1:2 ratio
)

// Images crops the preview image based on the crop position and ratio of the image.
func (crop Crop) Images(ctx context.Context, sl *slog.Logger, unid string, preview dir.Directory) error {
	const msg = "crop images"
	if sl == nil {
		return fmt.Errorf("%s: %w", msg, panics.ErrNoSlog)
	}
	if err := preview.Check(sl); err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}
	switch crop {
	case SquareTop, FourThree, OneTwo:
	default:
		return fmt.Errorf("%s: invalid crop: %d", msg, crop)
	}
	tmpDir := filepath.Join(helper.TmpDir(), patternS)
	pattern := "images-crop-" + unid
	path := filepath.Join(tmpDir, pattern)
	if st, err := os.Stat(path); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			if err := os.MkdirAll(path, os.ModePerm); err != nil {
				return fmt.Errorf("%s: %w", msg, err)
			}
		}
	} else if !st.IsDir() {
		return fmt.Errorf("%s: %w", msg, ErrIsFile)
	}
	imagesNotFound := true
	for ext := range slices.Values(ImagesExt()) {
		args := Args{}
		switch crop {
		case SquareTop:
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
		imagesNotFound = false
		arg := make([]string, 1, argCap+len(args))
		arg[0] = src
		arg = append(arg, args...)
		tmp := filepath.Join(path, unid+ext)
		arg = append(arg, tmp)
		err := Run(ctx, sl, Magick, arg...)
		if err != nil {
			return fmt.Errorf("%s: %w", msg, err)
		}
		dst := preview.Join(unid + ext)
		if err := CopyFile(sl, tmp, dst); err != nil {
			slog.Info(msg+" copyfile tmp to dst error", slog.String("src temp", tmp),
				slog.String("dst", dst), slog.Any("error", err))
			return nil
		}
	}
	if imagesNotFound {
		return fmt.Errorf("%s cannot use: %w", msg, ErrNoImages)
	}
	return nil
}

// PictureImager converts src into a preview image and a thumbnail image in their
// respective directories.
//
// Thumbnails are generated as .webp or .png files, while preview images depend on
// the input format (.png, .jpeg, .avif, .webp, etc.).
func (dir Dirs) PictureImager(ctx context.Context, sl *slog.Logger, src, unid string) error {
	const msg = "picture imager"
	if sl == nil {
		return fmt.Errorf("%s: %w", msg, panics.ErrNoSlog)
	}
	if err := dir.Preview.Check(sl); err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}
	if err := dir.Thumbnail.Check(sl); err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}
	// inspect magic bytes
	r, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("%s open src: %w", msg, err)
	}
	defer func() {
		if err := r.Close(); err != nil {
			slog.Info(msg, slog.String("opened file", src), slog.Any("close error", err))
		}
	}()
	magic := magicnumber.Find(r)
	// format signature aliases
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
	switch magic { //nolint:exhaustive
	case AVI:
		return nil // unsupported, so do not delete existing assets
	case GIF, WebP, PNG, TIFF, JPG, BMP, PCX:
	// do nothing
	default:
		return fmt.Errorf("%s: %w: %s", msg, ErrUnknownImg, magic.Title())
	}

	// remove existing thumbnails
	if err := ImagesDelete(unid, dir.Preview.Path(), dir.Thumbnail.Path()); err != nil && !errors.Is(err, ErrNoImages) {
		return fmt.Errorf("%s delete existing: %w", msg, err)
	}
	switch magic { //nolint:exhaustive
	case GIF:
		return dir.previewGif(ctx, sl, src, unid)
	case WebP:
		const makeThumb = true
		return dir.previewWebP(ctx, sl, src, unid, makeThumb)
	case PNG:
		return dir.previewPNG(ctx, sl, src, unid)
	case TIFF, JPG:
		return dir.previewPhoto(ctx, sl, src, unid)
	case BMP, PCX:
		return dir.previewPixels(ctx, sl, src, unid)
	default:
		return fmt.Errorf("%s: %w: %s", msg, ErrUnknownImg, magic.Title())
	}
}

// CropText reads the src text file and writes the number of maxRows (lines) of text to the dst file.
// The dst file is stored in the same directory as the src but is given the unid file name with a ".txt" extension.
// The text is truncated to maxColumns (characters per line), but any leading empty lines are ignored.
//
// If maxRows is set to 0, a default of 29 is used. And if maxColumns is set to 0, a default of 80 is used.
// As these defaults allow for a 400x400 pixel thumbnail to be generated with the ANSILOVE text to image generator.
//
//   - If successful, the full path to the cropped text is returned.
//   - If unsuccessful the src path is returned.
//   - If a large number of ANSI sequences are detected then an ErrIsAnsi error is returned.
func CropText(sl *slog.Logger, maxColumns, maxRows int, src, unid string) (string, error) {
	const msg = "crop text imager"
	if sl == nil {
		return "", fmt.Errorf("%s: %w", msg, panics.ErrNoSlog)
	}
	if unid == "" {
		return "", fmt.Errorf("%s: unid %w", msg, ErrValue)
	}
	src = filepath.Clean(src)
	path, err := helper.MkContent(src + "-textimager")
	if err != nil {
		return "", fmt.Errorf("%s make content: %w", msg, err)
	}
	tmpText := filepath.Join(path, unid+".txt")
	if err := cropText(sl, maxColumns, maxRows, src, tmpText); err != nil {
		if err1 := ansiCheck(src, err); err1 != nil {
			return "", fmt.Errorf("%s: %w", msg, err1)
		}
	}
	if !exist(tmpText) {
		return src, nil
	}
	return tmpText, nil
}

func ansiCheck(src string, err error) error {
	if !errors.Is(err, ErrIsAnsi) {
		return err
	}
	file, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("stat src: %w", err)
	}
	if file.Size() > AnsiCap {
		return fmt.Errorf("%w: file size exceeds maximum ansi size", ErrIsAnsi)
	}
	// allow processing to continue for small ANSI files
	return nil
}

func cropText(sl *slog.Logger, maxColumns, maxRows int, src, dst string) error {
	const msg = "text crop scanner"
	if sl == nil {
		return fmt.Errorf("%s: %w", msg, panics.ErrNoSlog)
	}
	src = filepath.Clean(src)
	dst = filepath.Clean(dst)

	r, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("%s open src: %w", msg, err)
	}
	defer func(sl *slog.Logger) {
		if err := r.Close(); err != nil {
			sl.Info(msg, slog.String("error closing src file", src), slog.Any("error", err))
		}
	}(sl)

	// return an error if any ansi sequences are found
	if magicnumber.CSI(r) {
		return fmt.Errorf("%s: %w: %s", msg, ErrIsAnsi, src)
	}
	const reset = 0
	if _, err := r.Seek(reset, io.SeekStart); err != nil {
		return fmt.Errorf("%s seek reset: %w", msg, err)
	}

	w, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("%s create dst: %w", msg, err)
	}
	defer func(sl *slog.Logger) {
		if err := w.Close(); err != nil {
			sl.Info(msg, slog.String("error closing dst file", dst), slog.Any("error", err))
		}
	}(sl)

	scanner := bufio.NewScanner(r)
	writer := bufio.NewWriter(w)
	defer func(sl *slog.Logger) {
		if err := writer.Flush(); err != nil {
			sl.Info(msg, slog.Any("error flushing writer", err))
		}
	}(sl)

	if maxColumns < 1 {
		maxColumns = 80
	}
	if maxRows < 1 {
		maxRows = 29
	}
	cntRows := 0
	skipLeadingEmpty := true

	for scanner.Scan() {
		line := scanner.Text()
		// skip leading empty lines
		if skipLeadingEmpty {
			if strings.TrimSpace(line) == "" {
				continue
			}
			skipLeadingEmpty = false
		}
		// end of crop
		if cntRows >= maxRows {
			break
		}
		// truncate lines of text, howver this is not valid with unicode multi-bytes.
		if len(line) > maxColumns {
			line = line[:maxColumns]
		}
		// write the truncated text and append newline
		if _, err := writer.WriteString(line + "\n"); err != nil {
			return fmt.Errorf("%s write string: %w", msg, err)
		}
		cntRows++
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("%s scan line: %w", msg, err)
	}
	if err := writer.Flush(); err != nil {
		return fmt.Errorf("%s flush writer: %w", msg, err)
	}
	return nil
}

// BinTextImager converts a binary text source file into a PNG preview using Ansilove,
// then processes the generated PNG through the standard text imager pipeline.
func (dir Dirs) BinTextImager(ctx context.Context, sl *slog.Logger, src, unid string) error {
	const msg = "binary text imager"
	if sl == nil {
		return fmt.Errorf("%s: %w", msg, panics.ErrNoSlog)
	}

	srcPath := filepath.Clean(src)
	st, err := os.Stat(srcPath)
	if err != nil {
		return fmt.Errorf("%s stat src: %w", msg, err)
	}
	if st.Size() == 0 {
		return fmt.Errorf("%s %s: %w", msg, srcPath, ErrIsEmpty)
	}

	// create an isolated, safe temporary file
	tmpFile, err := os.CreateTemp("", "ansilove-*.png")
	if err != nil {
		return fmt.Errorf("%s create temp: %w", msg, err)
	}
	tmp := tmpFile.Name()
	_ = tmpFile.Close()
	defer func() {
		if err := os.Remove(tmp); err != nil && !os.IsNotExist(err) {
			sl.Error(msg, slog.String("tmp", tmp), slog.Any("error", err))
		}
	}()

	// command line arguments: [src, "-o", dst]
	ansiloveArgs := []string{srcPath, "-o", tmp}
	if err := Run(ctx, sl, Ansilove, ansiloveArgs...); err != nil {
		return fmt.Errorf("%s run ansilove: %w", msg, err)
	}
	if err := dir.optimizeAnsilove(ctx, sl, unid, tmp); err != nil {
		return fmt.Errorf("%s text imagers: %w", msg, err)
	}
	return nil
}

// TextImager generates two images based on the text file provided by the src path.
// The provided unid must be a valid universal unique identifier.
//
// If the amigaFont is set to true, a Commodore Amiga Tapaz font is used to represent
// the text, otherwise an IBM VGA font is used. The amigaFont also displays less rows
// of text due to the Topaz font being taller.
func (dir Dirs) TextImager(ctx context.Context, sl *slog.Logger, src, unid string, amigaFont bool) error {
	const msg = "dirs text imager"
	if sl == nil {
		return fmt.Errorf("%s: %w", msg, panics.ErrNoSlog)
	}
	if err := dir.Thumbnail.Check(sl); err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}
	if err := dir.Preview.Check(sl); err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}
	const maxColumns = 80
	cfg := imagerCfg{
		maxColumns: maxColumns,
		unid:       unid,
	}
	if amigaFont {
		const topaz = 35
		cfg.amigaFont = true
		cfg.maxRows = topaz
		return dir.textImager(ctx, sl, src, cfg)
	}
	const vga = 50
	cfg.amigaFont = false
	cfg.maxRows = vga
	return dir.textImager(ctx, sl, src, cfg)
}

type imagerCfg struct {
	maxColumns int
	maxRows    int
	unid       string
	amigaFont  bool
}

func (dir Dirs) textImager(ctx context.Context, sl *slog.Logger, src string, cfg imagerCfg) error {
	const msg = "text imager"
	if sl == nil {
		return fmt.Errorf("%s: %w", msg, panics.ErrNoSlog)
	}
	maxColumns := cfg.maxColumns
	maxRows := cfg.maxRows
	unid := cfg.unid
	amigaFont := cfg.amigaFont

	// create a cropped text file
	src = filepath.Clean(src)
	srcPath, err := CropText(sl, maxColumns, maxRows, src, unid)
	if err != nil {
		return fmt.Errorf("%s crop text: %w", msg, err)
	}
	if srcPath != src {
		defer func() {
			if err := os.Remove(srcPath); err != nil && !os.IsNotExist(err) {
				sl.Error(msg, slog.String("remove temp cropped text file", srcPath), slog.Any("error", err))
			}
		}()
	}
	st, err := os.Stat(srcPath)
	if err != nil {
		return fmt.Errorf("%s stat cropped src: %w", msg, err)
	}
	if st.Size() == 0 {
		return fmt.Errorf("%s %s: %w", msg, srcPath, ErrIsEmpty)
	}

	// create an isolated temporary PNG destination
	tmpFile, err := os.CreateTemp("", "ansilove-dos-*.png")
	if err != nil {
		return fmt.Errorf("%s create temp: %w", msg, err)
	}
	tmp := tmpFile.Name()
	_ = tmpFile.Close()
	defer func() {
		if err := os.Remove(tmp); err != nil && !os.IsNotExist(err) {
			sl.Error(msg, slog.String("remove temp image file", tmp), slog.Any("error", err))
		}
	}()

	// command arguments: [srcPath, flags..., "-o", tmp]
	args := Args{}
	if amigaFont {
		args.AnsiAmiga()
	} else {
		args.AnsiMsDos()
	}
	ansiloveArgs := make([]string, 0, 3+len(args))
	ansiloveArgs = append(ansiloveArgs, srcPath)
	ansiloveArgs = append(ansiloveArgs, args...)
	ansiloveArgs = append(ansiloveArgs, "-o", tmp)
	if err := Run(ctx, sl, Ansilove, ansiloveArgs...); err != nil {
		return fmt.Errorf("%s run ansilove: %w", msg, err)
	}
	if err := dir.optimizeAnsilove(ctx, sl, unid, tmp); err != nil {
		return fmt.Errorf("%s text imagers: %w", msg, err)
	}
	return nil
}

// optimizeAnsilove concurrently generates an optimized PNG preview, a WebP preview,
// and the thumbnail from an temporary, unoptimized ANSILOVE generated image.
func (dir Dirs) optimizeAnsilove(ctx context.Context, sl *slog.Logger, unid, tmp string) error {
	const msg = "optimize ansilove conversion"
	if sl == nil {
		return fmt.Errorf("%s: %w", msg, panics.ErrNoSlog)
	}

	defer func() {
		if err := os.Remove(tmp); err != nil && !os.IsNotExist(err) {
			sl.Error(msg, slog.String("ansilove conversion tmp removal", tmp), slog.Any("error", err))
		}
	}()

	// remove existing preview & thumbnail images
	if err := ImagesDelete(unid, dir.Preview.Path(), dir.Thumbnail.Path()); err != nil && !errors.Is(err, ErrNoImages) {
		return fmt.Errorf("%s delete existing: %w", msg, err)
	}

	g, ctx := errgroup.WithContext(ctx)

	// task 1: optimized PNG preview
	g.Go(func() error {
		dst := filepath.Join(dir.Preview.Path(), unid+png)
		if err := CopyFile(sl, tmp, dst); err != nil {
			return fmt.Errorf("%s copy file: %w", msg, err)
		}
		if err := OptimizePNG(ctx, dst); err != nil {
			return fmt.Errorf("%s optimize png: %w", msg, err)
		}
		return nil
	})

	// task 2: webp preview
	g.Go(func() error {
		const makeThumb = false
		if err := dir.previewWebP(ctx, sl, tmp, unid, makeThumb); err != nil {
			return fmt.Errorf("%s webp preview: %w", msg, err)
		}
		return nil
	})

	// task 3: thumbnail
	g.Go(func() error {
		if err := dir.thumbPixels(ctx, sl, tmp, unid); err != nil {
			return fmt.Errorf("%s thumbnail: %w", msg, err)
		}
		return nil
	})

	if err := g.Wait(); err != nil {
		return err
	}
	return nil
}

// previewPixels converts src into PNG and WebP preview images in the preview directory.
// A WebP thumbnail image is also generated in the thumbnail directory.
// This lossless conversion is optimal for screenshots of text, terminal interfaces, and pixel art.
func (dir Dirs) previewPixels(ctx context.Context, sl *slog.Logger, src, unid string) error {
	const msg = "pixel image preview"
	if sl == nil {
		return fmt.Errorf("%s: %w", msg, panics.ErrNoSlog)
	}

	tmpDir, err := os.MkdirTemp(helper.TmpDir(), "previewpixels-*")
	if err != nil {
		return fmt.Errorf("%s create temp dir: %w", msg, err)
	}
	defer func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			sl.Error(msg, slog.String("remove temp directory", tmpDir), slog.Any("error", err))
		}
	}()

	tmpName := filepath.Base(src) + png
	tmpPath := filepath.Join(tmpDir, tmpName)
	// command flags: [src, flags..., tmpPath]
	args := Args{}
	args.PortablePixel()
	magickArgs := make([]string, 0, 2+len(args))
	magickArgs = append(magickArgs, src)
	magickArgs = append(magickArgs, args...)
	magickArgs = append(magickArgs, tmpPath)
	if err := RunQuiet(ctx, Magick, magickArgs...); err != nil {
		return fmt.Errorf("%s run magick: %w", msg, err)
	}

	dst := filepath.Join(dir.Preview.Path(), unid+png)
	if err := CopyFile(sl, tmpPath, dst); err != nil {
		return fmt.Errorf("%s copy preview png: %w", msg, err)
	}
	if err := dir.optimizeAnsilove(ctx, sl, unid, tmpPath); err != nil {
		return fmt.Errorf("%s optimization: %w", msg, err)
	}
	return nil
}

// previewPhoto converts src into a lossy JPEG and WebP image in a temporary directory.
// It compares their output sizes, copies the smaller format to the preview directory,
// and generates a WebP thumbnail in the thumbnail directory.
//
// This lossy conversion is optimal for continuous-tone photographs.
func (dir Dirs) previewPhoto(ctx context.Context, sl *slog.Logger, src, unid string) error {
	const msg = "photo image preview"
	if sl == nil {
		return fmt.Errorf("%s: %w", msg, panics.ErrNoSlog)
	}

	tmpDir, err := os.MkdirTemp(helper.TmpDir(), "previewphoto-*")
	if err != nil {
		return fmt.Errorf("%s create temp dir: %w", msg, err)
	}
	defer func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			sl.Error(msg, slog.String("remove temp directory", tmpDir), slog.Any("error", err))
		}
	}()

	// convert source image to JPEG using ImageMagick
	jargs := Args{}
	jargs.JpegPhoto()
	jtmp := filepath.Join(tmpDir, filepath.Base(src)+jpg)
	magickArgs := make([]string, 0, 2+len(jargs))
	magickArgs = append(magickArgs, src)
	magickArgs = append(magickArgs, jargs...)
	magickArgs = append(magickArgs, jtmp)
	if err := RunQuiet(ctx, Magick, magickArgs...); err != nil {
		return fmt.Errorf("%s convert jpeg: %w", msg, err)
	}

	// convert generated JPEG to WebP using cwebp
	wargs := Args{}
	wargs.CWebp()
	wtmp := filepath.Join(tmpDir, unid+webp)
	cwebpArgs := make([]string, 0, 3+len(wargs))
	cwebpArgs = append(cwebpArgs, jtmp)
	cwebpArgs = append(cwebpArgs, wargs...)
	cwebpArgs = append(cwebpArgs, "-o", wtmp)
	if err := RunQuiet(ctx, Cwebp, cwebpArgs...); err != nil {
		return fmt.Errorf("%s run cwebp: %w", msg, err)
	}

	// compare the image file sizes and pick the smaller format for use as the preview
	jst, err := os.Stat(jtmp)
	if err != nil {
		return fmt.Errorf("%s stat jpeg: %w", msg, err)
	}
	wst, err := os.Stat(wtmp)
	if err != nil {
		return fmt.Errorf("%s stat webp: %w", msg, err)
	}
	srcPath := wtmp
	dst := filepath.Join(dir.Preview.Path(), unid+webp)
	if jst.Size() < wst.Size() {
		srcPath = jtmp
		dst = filepath.Join(dir.Preview.Path(), unid+jpg)
	}
	if err := CopyFile(sl, srcPath, dst); err != nil {
		return fmt.Errorf("%s copy preview: %w", msg, err)
	}
	if err := dir.thumbPhoto(ctx, sl, srcPath, unid); err != nil {
		return fmt.Errorf("%s thumbnail: %w", msg, err)
	}
	return nil
}

// previewGif converts a GIF image into an animated or static WebP preview image
// in the preview directory, and generates a WebP thumbnail in the thumbnail directory.
func (dir Dirs) previewGif(ctx context.Context, sl *slog.Logger, src, unid string) error {
	const msg = "gif preview"
	if sl == nil {
		return fmt.Errorf("%s: %w", msg, panics.ErrNoSlog)
	}

	src = filepath.Clean(src)
	tmpFile, err := os.CreateTemp("", "preview-gif-*.webp")
	if err != nil {
		return fmt.Errorf("%s create temp: %w", msg, err)
	}
	tmp := tmpFile.Name()
	_ = tmpFile.Close()

	defer func() {
		if err := os.Remove(tmp); err != nil && !os.IsNotExist(err) {
			sl.Error(msg, slog.String("remove temp image", tmp), slog.Any("error", err))
		}
	}()

	// command arguments: [src, flags..., "-o", tmp]
	args := Args{}
	args.GWebp()
	gwebpArgs := make([]string, 0, 3+len(args))
	gwebpArgs = append(gwebpArgs, src)
	gwebpArgs = append(gwebpArgs, args...)
	gwebpArgs = append(gwebpArgs, "-o", tmp)
	if err := Run(ctx, sl, Gwebp, gwebpArgs...); err != nil {
		return fmt.Errorf("%s run gif2webp: %w", msg, err)
	}

	dst := filepath.Join(dir.Preview.Path(), unid+webp)
	if err := CopyFile(sl, tmp, dst); err != nil {
		return fmt.Errorf("%s copy preview: %w", msg, err)
	}
	if err := dir.thumbPixels(ctx, sl, tmp, unid); err != nil {
		return fmt.Errorf("%s thumbnail: %w", msg, err)
	}
	return nil
}

// previewPNG copies the src PNG image to the preview directory.
// It concurrently optimizes the preview and generates a WebP thumbnail .
func (dir Dirs) previewPNG(ctx context.Context, sl *slog.Logger, src, unid string) error {
	const msg = "preview png"
	if sl == nil {
		return fmt.Errorf("%s: %w", msg, panics.ErrNoSlog)
	}

	src = filepath.Clean(src)
	dst := filepath.Join(dir.Preview.Path(), unid+png)
	if err := CopyFile(sl, src, dst); err != nil {
		return fmt.Errorf("%s copy file: %w", msg, err)
	}

	g, ctx := errgroup.WithContext(ctx)
	// task 1: png optimization
	g.Go(func() error {
		if err := OptimizePNG(ctx, dst); err != nil {
			return fmt.Errorf("%s optimize: %w", msg, err)
		}
		return nil
	})
	// task 2: thumbnail generation
	g.Go(func() error {
		if err := dir.thumbPixels(ctx, sl, src, unid); err != nil {
			return fmt.Errorf("%s thumbnail: %w", msg, err)
		}
		return nil
	})

	if err := g.Wait(); err != nil {
		_ = os.Remove(dst)
		return err
	}
	return nil
}

// previewWebP runs the cwebp text preset on a supported source image (.png, .jpg, .tiff, .webp),
// copies the resulting WebP to the preview directory, and optionally generates a thumbnail.
func (dir Dirs) previewWebP(ctx context.Context, sl *slog.Logger, src, unid string, makeThumb bool) error {
	const msg = "preview webp"
	if sl == nil {
		return fmt.Errorf("%s: %w", msg, panics.ErrNoSlog)
	}

	src = filepath.Clean(src)
	tmpFile, err := os.CreateTemp("", "preview-webp-*.webp")
	if err != nil {
		return fmt.Errorf("%s create temp: %w", msg, err)
	}
	tmp := tmpFile.Name()
	_ = tmpFile.Close()

	defer func() {
		if err := os.Remove(tmp); err != nil && !os.IsNotExist(err) {
			sl.Error(msg, slog.String("remove temp image", tmp), slog.Any("error", err))
		}
	}()

	// cwebp arguments: [src, flags..., "-o", tmp]
	args := Args{}
	args.CWebpText()
	cwebpArgs := make([]string, 0, 3+len(args))
	cwebpArgs = append(cwebpArgs, src)
	cwebpArgs = append(cwebpArgs, args...)
	cwebpArgs = append(cwebpArgs, "-o", tmp)
	if err := Run(ctx, sl, Cwebp, cwebpArgs...); err != nil {
		return fmt.Errorf("%s run cwebp: %w", msg, err)
	}
	dst := filepath.Join(dir.Preview.Path(), unid+webp)
	if err := CopyFile(sl, tmp, dst); err != nil {
		return fmt.Errorf("%s copy preview: %w", msg, err)
	}

	if !makeThumb {
		return nil
	}
	if err := dir.thumbPixels(ctx, sl, tmp, unid); err != nil {
		return fmt.Errorf("%s thumbnail: %w", msg, err)
	}
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
	gravity := []string{argGravity, north}
	*args = append(*args, gravity...)
	// Set the image size and offset.
	extent := []string{argTrim, argExtent, X400}
	*args = append(*args, extent...)
}

// Middlex400 appends the command line arguments for the magick command to transform
// an image into args 400x400 pixel image using the "Center" alignment.
func (args *Args) Middlex400() {
	// Set the gravity suggestion for various other settings and options.
	gravity := []string{argGravity, "center"}
	*args = append(*args, gravity...)
	// Set the image size and offset.
	extent := []string{argTrim, argExtent, X400}
	*args = append(*args, extent...)
}

// Bottomx400 appends the command line arguments for the magick command to transform
// an image into args 400x400 pixel image using the "South" bottom alignment.
func (args *Args) Bottomx400() {
	// Set the gravity suggestion for various other settings and options.
	gravity := []string{argGravity, "South"}
	*args = append(*args, gravity...)
	// Set the image size and offset.
	extent := []string{argTrim, argExtent, X400}
	*args = append(*args, extent...)
}

// Leftx400 appends the command line arguments for the magick command to transform
// an image into args 400x400 pixel image using the "South" bottom alignment.
func (args *Args) Leftx400() {
	// Set the gravity suggestion for various other settings and options.
	gravity := []string{argGravity, "West"}
	*args = append(*args, gravity...)
	// Set the image size and offset.
	extent := []string{argTrim, argExtent, X400}
	*args = append(*args, extent...)
}

// Rightx400 appends the command line arguments for the magick command to transform
// an image into args 400x400 pixel image using the "South" bottom alignment.
func (args *Args) Rightx400() {
	// Set the gravity suggestion for various other settings and options.
	gravity := []string{argGravity, "East"}
	*args = append(*args, gravity...)
	// Set the image size and offset.
	extent := []string{argTrim, argExtent, X400}
	*args = append(*args, extent...)
}

// CropTop appends the command line arguments for the magick command to transform
// an image into args 1:1 square image using the "North" top alignment.
func (args *Args) CropTop() {
	// Set the gravity suggestion for various other settings and options.
	gravity := []string{argGravity, north}
	*args = append(*args, gravity...)
	// Set the image size and offset.
	extent := []string{argExtent, "1:1"}
	*args = append(*args, extent...)
}

// FourThree appends the command line arguments for the magick command to transform
// an image into args 4:3 image using the "North" top alignment.
func (args *Args) FourThree() {
	// Set the gravity suggestion for various other settings and options.
	gravity := []string{argGravity, north}
	*args = append(*args, gravity...)
	// Set the image size and offset.
	extent := []string{argExtent, "4:3"}
	*args = append(*args, extent...)
}

// OneTwo appends the command line arguments for the magick command to transform
// an image into args 1:2 image using the "North" top alignment.
func (args *Args) OneTwo() {
	// Set the gravity suggestion for various other settings and options.
	gravity := []string{argGravity, north}
	*args = append(*args, gravity...)
	// Set the image size and offset.
	extent := []string{argExtent, "1:2"}
	*args = append(*args, extent...)
}

// AnsiAmiga appends the command line arguments for the [ansilove command]
// to transform an Commodore Amiga ANSI text file into args PNG image.
//
// [ansilove command]: https://github.com/ansilove/ansilove
func (args *Args) AnsiAmiga() {
	type font string // amiga font
	const (
		microknight font = "microknight+"
		mosoul      font = "mosoul"
		potnoodle   font = "pot-noodle"
		topaz       font = "topaz+"
		topaz500    font = "topaz500+"
	)
	type mode string // rendering mode
	const (
		ced         mode = "ced"         // black on gray
		transparent mode = "transparent" // transparent background
		workbench   mode = "workbench"   // Amiga Workbench colors
	)
	// Output font.
	f := []string{"-f", string(topaz)}
	*args = append(*args, f...)
	// Rendering mode set to Amiga palette.
	m := []string{"-m", string(workbench)}
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
	type font string // pc font
	const (
		f80x25   font = "80x25" // vga 80x25
		f80x50   font = "80x50" // vga 80x50 (hires)
		spleen   font = "spleen"
		terminus font = "terminus"
	)
	// DOS aspect ratio.
	const d = "-d"
	*args = append(*args, d)
	// Output font.
	f := []string{"-f", string(f80x50)}
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
	// This options will vary based on the type of screenshot being used as a thumbnail.
	// A test with a DOS VGA image screenshot that was 11,074B in size at 400x400 pixels,
	// has the following results when reprocessed into a Jpeg with the following options:
	// 6,012B with gaussianBlur and quality at 75.
	// 6,334B with gaussianBlur and quality at 90.
	// 8,824B with no gaussianBlur config and quality at 75.
	// 9,234B with no gaussianBlur config and quality at 90.
	const jpegQuality = "75"
	const gaussian = false
	// Strip the image of any profiles and comments.
	const strip = "-strip"
	*args = append(*args, strip)
	// See: https://imagemagick.org/script/command-line-options.php#quality
	quality := []string{"-quality", jpegQuality}
	*args = append(*args, quality...)
	if gaussian {
		// Blur the image with args Gaussian operator.
		gaussianBlur := []string{"-gaussian-blur", "0.05"}
		*args = append(*args, gaussianBlur...)
	}
	// NOTE: Oct-25, this has been disabled as it breaks certain images causing them to be far too.
	// Set the image colorspace.
	// colorspace := []string{"-colorspace", "RGB"}
	// *args = append(*args, colorspace...)
}

// PortablePixel appends the command line arguments for the convert command to transform an image into args PNG image.
func (args *Args) PortablePixel() {
	// Defined PNG compression options, these replace the -quality option.
	const define = "-define"
	*args = append(
		*args,
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
	gravity := []string{argGravity, "center"}
	*args = append(*args, gravity...)
	// Set the image size and offset.
	extent := []string{argExtent, X400}
	*args = append(*args, extent...)
}

// CWebp appends the command line arguments for the [cwebp command] to transform an image into args webp image.
//
// [cwebp command]: https://developers.google.com/speed/webp/docs/cwebp
func (args *Args) CWebp() {
	// Auto-filter will spend additional time optimizing the
	// filtering strength to reach args well-balanced quality.
	const ll = "-lossless"
	*args = append(*args, ll)
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

// thumbPixels converts src to a 400x400 pixel WebP image in the thumbnail directory.
// It uses a temporary lossless image format during conversion to preserve crisp edges.
//
// This is intended for text and pixel art images.
func (dir Dirs) thumbPixels(ctx context.Context, sl *slog.Logger, src, unid string) error {
	const msg = "thumb as pixel capture"
	if sl == nil {
		return fmt.Errorf("%s: %w", msg, panics.ErrNoSlog)
	}
	if unid == "" {
		return fmt.Errorf("%s: unid %w", msg, ErrValue)
	}

	// Create a safe, unique temporary file
	tmpFile, err := os.CreateTemp("", "thumb-pixel-*.png")
	if err != nil {
		return fmt.Errorf("%s create temp file: %w", msg, err)
	}
	tmp := tmpFile.Name()
	_ = tmpFile.Close()

	// Always cleanup
	defer func() {
		if err := os.Remove(tmp); err != nil && !os.IsNotExist(err) {
			sl.Error(msg, slog.String("remove temp file", tmp), slog.Any("error", err))
		}
	}()

	// 1. Render source to intermediate lossless image
	args := Args{}
	args.Thumbnail()
	args.PortablePixel()

	magickArgs := make([]string, 0, 2+len(args))
	magickArgs = append(magickArgs, src)
	magickArgs = append(magickArgs, args...)
	magickArgs = append(magickArgs, tmp)

	if err := RunQuiet(ctx, Magick, magickArgs...); err != nil {
		return fmt.Errorf("%s run magick convert: %w", msg, err)
	}

	// 2. Convert intermediate image to WebP
	dst := filepath.Join(dir.Thumbnail.Path(), unid+webp)
	args = Args{}
	args.CWebp()

	cwebpArgs := make([]string, 0, 3+len(args))
	cwebpArgs = append(cwebpArgs, tmp)
	cwebpArgs = append(cwebpArgs, args...)
	cwebpArgs = append(cwebpArgs, "-o", dst)

	if err := RunQuiet(ctx, Cwebp, cwebpArgs...); err != nil {
		return fmt.Errorf("%s run cwebp: %w", msg, err)
	}

	return nil
}

// thumbPhoto converts the src image to args 400x400 pixel WebP image in the thumbnail directory.
// It uses a temporary JPEG intermediate image during conversion.
//
// This is used for photographs and images that are not text or pixel art.
func (dir Dirs) thumbPhoto(ctx context.Context, sl *slog.Logger, src, unid string) error {
	const msg = "thumb as photograph"
	if sl == nil {
		return fmt.Errorf("%s: %w", msg, panics.ErrNoSlog)
	}
	if unid == "" {
		return fmt.Errorf("%s: unid %w", msg, ErrValue)
	}

	// Create a safe, unique temporary file
	tmpFile, err := os.CreateTemp("", "thumb-*.jpg")
	if err != nil {
		return fmt.Errorf("%s create temp file: %w", msg, err)
	}
	tmp := tmpFile.Name()
	_ = tmpFile.Close()

	// Always cleanup
	defer func() {
		if err := os.Remove(tmp); err != nil && !os.IsNotExist(err) {
			sl.Error(msg, slog.String("tmp", tmp), slog.Any("error", err))
		}
	}()

	// 1. Convert source to intermediate JPEG
	args := Args{}
	args.Thumbnail()
	args.JpegPhoto()

	magickArgs := make([]string, 0, 2+len(args))
	magickArgs = append(magickArgs, src)
	magickArgs = append(magickArgs, args...)
	magickArgs = append(magickArgs, tmp)
	if err := RunQuiet(ctx, Magick, magickArgs...); err != nil {
		return fmt.Errorf("%s run magick convert: %w", msg, err)
	}

	// 2. Convert intermediate JPEG to WebP
	dst := filepath.Join(dir.Thumbnail.Path(), unid+webp)
	args = Args{}
	args.CWebp()

	cwebpArgs := make([]string, 0, 3+len(args))
	cwebpArgs = append(cwebpArgs, tmp)
	cwebpArgs = append(cwebpArgs, args...)
	cwebpArgs = append(cwebpArgs, "-o", dst)

	if err := RunQuiet(ctx, Cwebp, cwebpArgs...); err != nil {
		return fmt.Errorf("%s run cwebp: %w", msg, err)
	}
	return nil
}

// OptimizePNG optimizes the src PNG image in-place using optipng.
// It is safe to call within a deferred function.
func OptimizePNG(ctx context.Context, src string) error {
	const format = "optimize png using optipng %s: %w"
	const minPNG = 67
	if st, err := os.Stat(src); err != nil {
		return fmt.Errorf(format, src, err)
	} else if st.Size() < minPNG {
		return fmt.Errorf(format, src, ErrIsEmpty)
	}
	args := Args{}
	// command args: [flags..., src]
	cmdArgs := make([]string, 0, len(args)+1)
	cmdArgs = append(cmdArgs, args...)
	cmdArgs = append(cmdArgs, src)

	if err := RunQuiet(ctx, Optipng, cmdArgs...); err != nil {
		return fmt.Errorf(format, src, err)
	}
	return nil
}

// TextDeferred creates a thumbnail (if one does not exist) and copies the source text file
// into the extra directory.
func (dir Dirs) TextDeferred(ctx context.Context, sl *slog.Logger, src, unid string) error {
	const msg = "text deferred"
	if sl == nil {
		return fmt.Errorf("%s: %w", msg, panics.ErrNoSlog)
	}

	// Check if a non-empty thumbnail image already exists
	hasThumb := false
	for _, ext := range ImagesExt() {
		thumbPath := filepath.Join(dir.Thumbnail.Path(), unid+ext)
		if exist(thumbPath) {
			hasThumb = true
			break
		}
	}

	if !hasThumb {
		if err := dir.TextImager(ctx, sl, src, unid, false); err != nil {
			return fmt.Errorf("%s text imager: %w: %s", msg, err, src)
		}
	}

	// Copy to extra directory if destination does not already exist
	dst := filepath.Join(dir.Extra.Path(), unid+".txt")
	if exist(dst) {
		return nil
	}

	if _, err := helper.DuplicateOW(src, dst); err != nil {
		return subdirDuplicate(err, src, dst, "text")
	}

	return nil
}

// DizDeferred copies a FILE_ID.DIZ text file into the extra directory.
func (dir Dirs) DizDeferred(src, unid string) error {
	dst := filepath.Join(dir.Extra.Path(), unid+".diz")
	if exist(dst) {
		return nil
	}

	if _, err := helper.DuplicateOW(src, dst); err != nil {
		return subdirDuplicate(err, src, dst, "diz")
	}

	return nil
}

func subdirDuplicate(err error, src, dst, category string) error {
	if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("%s deferred copy: %w: %s", category, err, src)
	}

	// Source missing; attempt case-insensitive or loose name resolution in parent directory
	oldDir, oldName := filepath.Dir(src), filepath.Base(src)
	foundPath := findName(oldDir, oldName)
	if foundPath == "" {
		return fmt.Errorf("%s deferred source missing: %w: %s", category, os.ErrNotExist, src)
	}

	if _, err1 := helper.DuplicateOW(foundPath, dst); err1 != nil {
		return fmt.Errorf("%s deferred fallback copy: %w: %s", category, err1, foundPath)
	}

	return nil
}

// findName walks root directory searching for a file matching target name.
func findName(root, target string) string {
	var result string

	_ = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil // Skip unreadable entries and keep searching
		}
		if d.Name() == target {
			result = path
			return filepath.SkipAll // Immediately stop walking
		}
		return nil
	})

	return result
}

func exist(path string) bool {
	st, err := os.Stat(path)
	return err == nil && st.Size() > 0
}
