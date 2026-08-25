package command

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/Defacto2/helper"
	"github.com/Defacto2/magicnumber"
	"github.com/Defacto2/server/internal/command/option"
	"github.com/Defacto2/server/internal/dir"
	"github.com/Defacto2/server/internal/nils"
	"golang.org/x/sync/errgroup"
)

// Package file dirs.go houses all the command funcs that rely on the Dirs struct.

// Dirs points to the download, preview, thumbnail, and extra directories.
type Dirs struct {
	Download  dir.Directory // Download is the directory path for the file downloads.
	Preview   dir.Directory // Preview is the directory path for the image previews.
	Thumbnail dir.Directory // Thumbnail is the directory path for the image thumbnails.
	Extra     dir.Directory // Extra is the directory path for the extra files.
}

// PictureImager creates both a preview image and a thumbnail image using the srcImage.
//
// Thumbnails are generated as .webp or .png files, while the preview image format depends
// on the input format (.png, .jpeg, .avif, .webp, etc.).
func (ds Dirs) PictureImager(ctx context.Context, sl *slog.Logger, srcImage, unid string) error {
	const format = "picture imager %s: %w"

	if err := nils.Check(ctx, sl); err != nil {
		return fmt.Errorf(format, check, err)
	}
	if err := ds.Preview.Check(sl); err != nil {
		return fmt.Errorf(format, prevChk, err)
	}
	if err := ds.Thumbnail.Check(sl); err != nil {
		return fmt.Errorf(format, thumbChk, err)
	}

	// inspect magic bytes
	r, err := os.Open(srcImage)
	if err != nil {
		return fmt.Errorf(format, "open source file", err)
	}
	defer r.Close()
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

	if magic == AVI {
		return nil // unsupported
	}

	if ctxErr := ctx.Err(); ctxErr != nil {
		return fmt.Errorf(format, "context", ctxErr)
	}

	switch magic { //nolint:exhaustive
	case IFF, JPG, PNG, GIF, WebP, TIFF, BMP, PCX: // do nothing
	default:
		return fmt.Errorf(format, magic.Title(), ErrUnknownImg)
	}

	// remove existing thumbnails
	err = ImagesDelete(unid, ds.Preview.Path(), ds.Thumbnail.Path())
	if err != nil && !errors.Is(err, ErrNoImages) {
		return fmt.Errorf(format, "delete existing images", err)
	}

	switch magic { //nolint:exhaustive
	case IFF:
		return ds.previewPixels(ctx, sl, srcImage, unid)
	case JPG:
		return ds.previewPhoto(ctx, sl, srcImage, unid)
	case PNG:
		return ds.previewPNG(ctx, sl, srcImage, unid)
	case GIF:
		return ds.previewGif(ctx, sl, srcImage, unid)
	case WebP:
		const makeThumb = true
		return ds.previewWebP(ctx, sl, srcImage, unid, makeThumb)
	case TIFF:
		return ds.previewPhoto(ctx, sl, srcImage, unid)
	case BMP:
		return ds.previewPixels(ctx, sl, srcImage, unid)
	case PCX:
		return ds.previewPixels(ctx, sl, srcImage, unid)
	default:
		return fmt.Errorf(format, magic.Title(), ErrUnknownImg)
	}
}

// BinTextImager converts a binary text source file into a PNG preview using Ansilove,
// then processes the generated PNG through the standard text imager pipeline.
func (ds Dirs) BinTextImager(ctx context.Context, sl *slog.Logger, srcBinary, unid string) error {
	const format = "binary text imager %s: %w"
	if err := nils.Check(ctx, sl); err != nil {
		return fmt.Errorf(format, "", err)
	}

	src := filepath.Clean(srcBinary)
	st, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf(format, "stat source", err)
	}
	if st.Size() == 0 {
		return fmt.Errorf(format, "zero size: "+src, ErrIsEmpty)
	}

	// create an isolated, safe temporary file
	f, err := dir.CreateTemp("bintext")
	if err != nil {
		return fmt.Errorf(format, "create temp", err)
	}
	dst := f.Name()
	_ = f.Close()

	defer func() {
		if err := os.Remove(dst); !errors.Is(err, os.ErrNotExist) {
			sl.Error("binary text imager could not remove temporary file",
				slog.String("file", dst), slog.Any("error", err))
		}
	}()

	// command line arguments: [src, "-o", dst]
	arg := []string{src, "-o", dst}
	if _, err := Run(ctx, sl, Ansilove, arg...); err != nil {
		return fmt.Errorf(format, "run ansilove", err)
	}

	if err := ds.optiAnsilove(ctx, sl, unid, dst); err != nil {
		return fmt.Errorf(format, "optimize ansilove", err)
	}

	return nil
}

// TextImager generates two images based on the text file provided by the src path.
// The provided unid must be a valid universal unique identifier.
//
// If the amigaFont is set to true, a Commodore Amiga Tapaz font is used to represent
// the text, otherwise an IBM VGA font is used. The amigaFont also displays less rows
// of text due to the Topaz font being taller.
func (ds Dirs) TextImager(ctx context.Context, sl *slog.Logger, src, unid string, amigaFont bool) error {
	const format = "dirs text imager %s: %w"
	if err := nils.Check(ctx, sl); err != nil {
		return fmt.Errorf(format, check, err)
	}

	if err := ds.Thumbnail.Check(sl); err != nil {
		return fmt.Errorf(format, thumbChk, err)
	}
	if err := ds.Preview.Check(sl); err != nil {
		return fmt.Errorf(format, prevChk, err)
	}

	const maxColumns = 80
	const vgaRows = 50
	t := Text{
		UUID:    unid,
		MaxRows: vgaRows,
		MaxCols: maxColumns,
	}

	if amigaFont {
		const topazRows = 35
		t.MaxRows = topazRows
		return ds.textImager(ctx, sl, src, true, t)
	}

	return ds.textImager(ctx, sl, src, false, t)
}

func (ds Dirs) textImager(ctx context.Context, sl *slog.Logger, src string, amigaFont bool, t Text) error {
	const format = "text imager %s: %w"
	if err := nils.Check(ctx, sl); err != nil {
		return fmt.Errorf(format, check, err)
	}

	// create a cropped text file
	src = filepath.Clean(src)
	srcPath, err := t.Crop(sl, src)
	if err != nil {
		return fmt.Errorf(format, "crop text", err)
	}
	if srcPath != src {
		defer func() {
			if err := os.Remove(srcPath); err != nil && !os.IsNotExist(err) {
				sl.Error("text imager could not remove source", slog.String("file", srcPath), slog.Any("error", err))
			}
		}()
	}

	st, err := os.Stat(srcPath)
	if err != nil {
		return fmt.Errorf(format, "stat cropped", err)
	}
	if st.Size() == 0 {
		return fmt.Errorf(format, "zero size: "+srcPath, ErrIsEmpty)
	}

	// create an isolated temporary PNG destination
	tmpFile, err := dir.CreateTemp("aldos-*.png")
	if err != nil {
		return fmt.Errorf(format, "create temp", err)
	}
	tmp := tmpFile.Name()
	_ = tmpFile.Close()
	defer func() {
		if err := os.Remove(tmp); err != nil && !os.IsNotExist(err) {
			sl.Error("text imager could not remove temporary file", slog.String("file", tmp), slog.Any("error", err))
		}
	}()

	arg := option.Opts{}
	if amigaFont {
		arg.AnsiAmiga(srcPath, tmp)
	} else {
		arg.AnsiDOS(srcPath, tmp)
	}
	fmt.Println("ANSI:", srcPath)
	r := Runner{Log: sl}
	if _, err := r.Run(ctx, Ansilove, arg...); err != nil {
		return fmt.Errorf(format, "run ansilove", err)
	}

	if err := ds.optiAnsilove(ctx, sl, t.UUID, tmp); err != nil {
		return fmt.Errorf(format, "optimize ansilove", err)
	}
	return nil
}

// optiAnsilove concurrently generates an optimized PNG preview, a WebP preview,
// and the thumbnail from an temporary, unoptimized ANSILOVE generated image.
func (ds Dirs) optiAnsilove(ctx context.Context, sl *slog.Logger, unid, tmp string) error {
	const msg = "optimize ansilove conversion"
	const format = msg + "%s: %w"
	if err := nils.Check(ctx, sl); err != nil {
		return fmt.Errorf(format, check, err)
	}

	defer func() {
		if err := os.Remove(tmp); err != nil && !os.IsNotExist(err) {
			sl.Error(msg+" could not remove temporary file", slog.String("file", tmp), slog.Any("error", err))
		}
	}()

	// remove existing preview & thumbnail images
	if err := ImagesDelete(unid, ds.Preview.Path(), ds.Thumbnail.Path()); err != nil && !errors.Is(err, ErrNoImages) {
		return fmt.Errorf(format, "delete existing", err)
	}

	g, ctx := errgroup.WithContext(ctx)

	// task 1: optimized PNG preview
	g.Go(func() error {
		dst := filepath.Join(ds.Preview.Path(), unid+png)
		if err := CopyFile(sl, tmp, dst); err != nil {
			return fmt.Errorf(format, "copy file", err)
		}
		if err := OptimizePNG(ctx, sl, dst); err != nil {
			return fmt.Errorf(format, "optimize png", err)
		}
		return nil
	})

	// task 2: webp preview
	g.Go(func() error {
		const makeThumb = false
		if err := ds.previewWebP(ctx, sl, tmp, unid, makeThumb); err != nil {
			return fmt.Errorf(format, "webp preview", err)
		}
		return nil
	})

	// task 3: thumbnail
	g.Go(func() error {
		if err := ds.thumbPixels(ctx, sl, tmp, unid); err != nil {
			return fmt.Errorf(format, "thumbnail pixels", err)
		}
		return nil
	})

	if err := g.Wait(); err != nil {
		return fmt.Errorf(format, "group wait", err)
	}
	return nil
}

// previewPixels converts src into PNG and WebP preview images in the preview directory.
// A WebP thumbnail image is also generated in the thumbnail directory.
// This lossless conversion is optimal for screenshots of text, terminal interfaces, and pixel art.
func (ds Dirs) previewPixels(ctx context.Context, sl *slog.Logger, src, unid string) error {
	const msg = "pixel image preview"
	const format = msg + " %s: %w"
	if err := nils.Check(ctx, sl); err != nil {
		return fmt.Errorf(format, check, err)
	}

	tmpDir, err := dir.MkdirTemp("pixel")
	if err != nil {
		return fmt.Errorf(format, "create temp directory", err)
	}
	defer func() {
		if err := dir.RemoveAll(tmpDir); err != nil {
			sl.Error(msg+" could not apply remove all to temporary directory",
				slog.String("directory", tmpDir), slog.Any("error", err))
		}
	}()

	name := filepath.Base(src) + png
	tmp := filepath.Join(tmpDir, name)

	// command flags: [src, flags..., tmpPath]
	arg := option.Opts{}
	arg.PNGPixel(false, src, tmp)
	r := Runner{Log: sl}
	if _, err := r.Run(ctx, Magick, arg...); err != nil {
		return fmt.Errorf(format, "run magick", err)
	}

	dst := filepath.Join(ds.Preview.Path(), unid+png)
	if err := CopyFile(sl, tmp, dst); err != nil {
		return fmt.Errorf(format, "copyfile png preview", err)
	}
	if err := ds.optiAnsilove(ctx, sl, unid, tmp); err != nil {
		return fmt.Errorf(format, "optimize ansilove", err)
	}
	return nil
}

// previewPhoto converts src into a lossy JPEG and WebP image in a temporary directory.
// It compares their output sizes, copies the smaller format to the preview directory,
// and generates a WebP thumbnail in the thumbnail directory.
//
// This lossy conversion is optimal for continuous-tone photographs.
func (ds Dirs) previewPhoto(ctx context.Context, sl *slog.Logger, src, unid string) error {
	const msg = "photo image preview"
	const format = msg + " %s: %w"
	if err := nils.Check(ctx, sl); err != nil {
		return fmt.Errorf(format, check, err)
	}

	tmpDir, err := dir.MkdirTemp("photo")
	if err != nil {
		return fmt.Errorf(format, "create temp directory", err)
	}
	defer func() {
		if err := dir.RemoveAll(tmpDir); err != nil {
			sl.Error(msg+" could not apply remove all to temporary directory",
				slog.String("directory", tmpDir), slog.Any("error", err))
		}
	}()

	// convert source image to JPEG using ImageMagick
	jtmp := filepath.Join(tmpDir, filepath.Base(src)+jpg)
	arg := option.Opts{}
	arg.JPGPhoto(false, src, jtmp)
	r := Runner{Log: sl}
	if _, err := r.Run(ctx, Magick, arg...); err != nil {
		return fmt.Errorf(format, "convert jpeg", err)
	}

	// convert generated JPEG to WebP using cwebp
	wtmp := filepath.Join(tmpDir, unid+webp)
	arg = option.Opts{}
	arg.WebpPhoto(jtmp, wtmp)
	if _, err := r.Run(ctx, Cwebp, arg...); err != nil {
		return fmt.Errorf(format, "run cwebp", err)
	}

	// compare the image file sizes and pick the smaller format for use as the preview
	jst, err := os.Stat(jtmp)
	if err != nil {
		return fmt.Errorf(format, "stat jpeg", err)
	}
	wst, err := os.Stat(wtmp)
	if err != nil {
		return fmt.Errorf(format, "stat webp", err)
	}
	srcPath := wtmp
	dst := filepath.Join(ds.Preview.Path(), unid+webp)
	if jst.Size() < wst.Size() {
		srcPath = jtmp
		dst = filepath.Join(ds.Preview.Path(), unid+jpg)
	}
	if err := CopyFile(sl, srcPath, dst); err != nil {
		return fmt.Errorf(format, "copy preview", err)
	}
	if err := ds.thumbPhoto(ctx, sl, srcPath, unid); err != nil {
		return fmt.Errorf(format, "thumbnail", err)
	}
	return nil
}

// previewGif converts a GIF image into an animated or static WebP preview image
// in the preview directory, and generates a WebP thumbnail in the thumbnail directory.
func (ds Dirs) previewGif(ctx context.Context, sl *slog.Logger, src, unid string) error {
	const msg = "gif preview"
	const format = msg + " %s: %w"
	if err := nils.Check(ctx, sl); err != nil {
		return fmt.Errorf(format, "", err)
	}

	src = filepath.Clean(src)
	tmpFile, err := dir.CreateTemp("previewgif-*.webp")
	if err != nil {
		return fmt.Errorf(format, "create temp", err)
	}
	tmp := tmpFile.Name()
	_ = tmpFile.Close()

	defer func() {
		if err := os.Remove(tmp); err != nil && !os.IsNotExist(err) {
			sl.Error(msg+" could not remove temporary file", slog.String("file", tmp), slog.Any("error", err))
		}
	}()

	// command arguments: [src, flags..., "-o", tmp]
	arg := option.Opts{}
	arg.Gif2webp(src, tmp)
	r := Runner{Log: sl}
	if _, err := r.Run(ctx, Gif2webp, arg...); err != nil {
		return fmt.Errorf(format, "run gif2webp", err)
	}

	dst := filepath.Join(ds.Preview.Path(), unid+webp)
	if err := CopyFile(sl, tmp, dst); err != nil {
		return fmt.Errorf(format, "copy preview", err)
	}
	if err := ds.thumbPixels(ctx, sl, tmp, unid); err != nil {
		return fmt.Errorf(format, "thumbnail", err)
	}
	return nil
}

// previewPNG copies the src PNG image to the preview directory.
// It concurrently optimizes the preview and generates a WebP thumbnail .
func (ds Dirs) previewPNG(ctx context.Context, sl *slog.Logger, src, unid string) error {
	const msg = "preview png"
	const format = msg + " %s: %w"
	if err := nils.Check(ctx, sl); err != nil {
		return fmt.Errorf(format, check, err)
	}

	src = filepath.Clean(src)
	dst := filepath.Join(ds.Preview.Path(), unid+png)
	if err := CopyFile(sl, src, dst); err != nil {
		return fmt.Errorf(format, "copy file", err)
	}

	g, ctx := errgroup.WithContext(ctx)
	// task 1: png optimization
	g.Go(func() error {
		if err := OptimizePNG(ctx, sl, dst); err != nil {
			return fmt.Errorf(format, "optimize", err)
		}
		return nil
	})
	// task 2: thumbnail generation
	g.Go(func() error {
		if err := ds.thumbPixels(ctx, sl, src, unid); err != nil {
			return fmt.Errorf(format, "thumbnail", err)
		}
		return nil
	})

	if err := g.Wait(); err != nil {
		_ = os.Remove(dst)
		return fmt.Errorf(format, "group wait", err)
	}
	return nil
}

// previewWebP runs the cwebp text preset on a supported source image (.png, .jpg, .tiff, .webp),
// copies the resulting WebP to the preview directory, and optionally generates a thumbnail.
func (ds Dirs) previewWebP(ctx context.Context, sl *slog.Logger, src, unid string, makeThumb bool) error {
	const msg = "preview webp"
	const format = msg + " %s: %w"
	if err := nils.Check(ctx, sl); err != nil {
		return fmt.Errorf(format, check, err)
	}

	src = filepath.Clean(src)
	tmpFile, err := dir.CreateTemp("previewwebp-*.webp")
	if err != nil {
		return fmt.Errorf(format, "create temp", err)
	}
	tmp := tmpFile.Name()
	_ = tmpFile.Close()

	defer func() {
		if err := os.Remove(tmp); err != nil && !os.IsNotExist(err) {
			sl.Error(msg+" could not remove temporary image", slog.String("file", tmp), slog.Any("error", err))
		}
	}()

	// cwebp arguments: [src, flags..., "-o", tmp]
	arg := option.Opts{}
	arg.WebpPixel(src, tmp)
	r := Runner{Log: sl}
	if out, err := r.Run(ctx, Cwebp, arg...); err != nil {
		return fmt.Errorf(format, "run cwebp "+string(out), err)
	}
	dst := filepath.Join(ds.Preview.Path(), unid+webp)
	if err := CopyFile(sl, tmp, dst); err != nil {
		return fmt.Errorf(format, "copy preview", err)
	}
	if !makeThumb {
		return nil
	}
	if err := ds.thumbPixels(ctx, sl, tmp, unid); err != nil {
		return fmt.Errorf(format, "thumbnail", err)
	}
	return nil
}

// thumbPixels converts src to a 400x400 pixel WebP image in the thumbnail directory.
// It uses a temporary lossless image format during conversion to preserve crisp edges.
//
// This is intended for text and pixel art images.
func (ds Dirs) thumbPixels(ctx context.Context, sl *slog.Logger, src, unid string) error {
	const format = "thumb as pixel capture %s: %w"
	if err := nils.Check(ctx, sl); err != nil {
		return fmt.Errorf(format, check, err)
	}

	t := Thumb{
		Source:  src,
		UUID:    unid,
		Pattern: "thumb-pixel-*.png",
		JPEG:    false,
	}
	return t.make(ctx, sl, ds.Thumbnail)
}

// thumbPhoto converts the src image to args 400x400 pixel WebP image in the thumbnail directory.
// It uses a temporary JPEG intermediate image during conversion.
//
// This is used for photographs and images that are not text or pixel art.
func (ds Dirs) thumbPhoto(ctx context.Context, sl *slog.Logger, src, unid string) error {
	const format = "thumb as photograph %s: %w"
	if err := nils.Check(ctx, sl); err != nil {
		return fmt.Errorf(format, check, err)
	}

	t := Thumb{
		Source:  src,
		UUID:    unid,
		Pattern: "thumb-photo-*.jpg",
		JPEG:    true,
	}
	return t.make(ctx, sl, ds.Thumbnail)
}

// TextDeferred creates a thumbnail (if one does not exist) and
// copies the source text file into the extra directory.
func (ds Dirs) TextDeferred(ctx context.Context, sl *slog.Logger, srcText, unid string) error {
	const format = "text deferred %s: %w"

	if err := nils.Check(ctx, sl); err != nil {
		return fmt.Errorf(format, check, err)
	}
	if unid == "" {
		return fmt.Errorf(format, "unid", ErrValue)
	}

	if err := ds.Thumbnail.Check(sl); err != nil {
		return fmt.Errorf(format, thumbChk, err)
	}
	if err := ds.Extra.Check(sl); err != nil {
		return fmt.Errorf(format, "extra check", err)
	}

	st, err := os.Stat(srcText)
	if err != nil {
		return fmt.Errorf(format, srcText, err)
	}
	if st.Size() == 0 {
		return fmt.Errorf(format, srcText, ErrIsEmpty)
	}

	base := filepath.Base(unid)
	thumb := false
	for _, ext := range imagesExt {
		name := filepath.Join(ds.Thumbnail.Path(), base+ext)
		if info, err := os.Stat(name); err == nil && info.Size() > 0 {
			thumb = true
			break
		}
	}

	// generate thumbnail if missing
	if !thumb {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return fmt.Errorf(format, "context canceled", ctxErr)
		}
		if err := ds.TextImager(ctx, sl, srcText, base, false); err != nil {
			return fmt.Errorf(format, srcText, err)
		}
	}

	// copy text file to the extra directory if it does not already exist
	newpath := filepath.Join(ds.Extra.Path(), base+".txt")
	if info, err := os.Stat(newpath); err == nil && info.Size() > 0 {
		return nil
	}

	if _, err := helper.DuplicateOW(srcText, newpath); err != nil {
		return duplicateSecondTry(sl, err, srcText, newpath, "text")
	}

	return nil
}

// DizDeferred copies a FILE_ID.DIZ text file into the extra directory.
func (ds Dirs) DizDeferred(sl *slog.Logger, srcDIZ, unid string) error {
	const format = "text diz defer %s: %w"

	if err := nils.Check(sl); err != nil {
		return fmt.Errorf(format, check, err)
	}
	if unid == "" {
		return fmt.Errorf(format, "unid", ErrValue)
	}
	if err := ds.Extra.Check(sl); err != nil {
		return fmt.Errorf(format, "extra check", err)
	}

	base := filepath.Base(unid)
	newpath := filepath.Join(ds.Extra.Path(), base+".diz")
	if st, err := os.Stat(newpath); err == nil && !st.IsDir() {
		return nil
	}

	if _, err := helper.DuplicateOW(srcDIZ, newpath); err != nil {
		return duplicateSecondTry(sl, err, srcDIZ, newpath, "diz")
	}

	return nil
}

// duplicateSecondTry is used when the source missing,
// so attempt a case-insensitive or a loose name resolution in the parent directory.
func duplicateSecondTry(sl *slog.Logger, err error, path, newpath, category string) error {
	if err := nils.Check(sl); err != nil {
		return fmt.Errorf("images duplicate second try: %w", err)
	}

	const format = "%s deferred %s: %w"
	if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf(format, category, path, err)
	}

	root, target := filepath.Dir(path), filepath.Base(path)
	oldpath := findName(sl, root, target)
	if oldpath == "" {
		return fmt.Errorf(format, category+" source missing", path, os.ErrNotExist)
	}

	if _, dErr := helper.DuplicateOW(oldpath, newpath); dErr != nil {
		return fmt.Errorf(format, category+" fallback copy", oldpath, dErr)
	}

	return nil
}

// findName walks root directory searching for a file matching target name.
func findName(sl *slog.Logger, root, target string) string {
	if sl == nil {
		return ""
	}

	entries, err := os.ReadDir(root)
	if err != nil {
		sl.Info("find name read dir", slog.String("root", root), slog.Any("error", err))
		return ""
	}

	// First pass: exact case match
	for _, d := range entries {
		if d.IsDir() {
			continue
		}
		if d.Name() == target {
			return filepath.Join(root, d.Name())
		}
	}

	// Second pass: case-insensitive match
	for _, d := range entries {
		if d.IsDir() {
			continue
		}
		if strings.EqualFold(d.Name(), target) {
			return filepath.Join(root, d.Name())
		}
	}

	return ""
}

// Generate is the type type of thumbnail image to create.
type Generate int

const (
	Pixel Generate = iota // Pixel art or images with text
	Photo                 // Photographs or images with gradients
)

// Thumbs creates a thumbnail image from the corresponding preview image based on the thumb type.
func (ds Dirs) Thumbs(ctx context.Context, sl *slog.Logger, unid string, generate Generate) error {
	const format = "thumb creator %s: %w"
	if err := nils.Check(ctx, sl); err != nil {
		return fmt.Errorf(format, check, err)
	}

	switch generate {
	case Pixel, Photo:
	default:
		return fmt.Errorf(format, fmt.Sprintf(" thumb value %d", generate), ErrThumb)
	}
	if err := ds.Thumbnail.Check(sl); err != nil {
		return fmt.Errorf(format, thumbChk, err)
	}
	if err := ds.Preview.Check(sl); err != nil {
		return fmt.Errorf(format, prevChk, err)
	}

	// remove any existing thumbnails; ignore expected "not found" errors
	if ds.Thumbnail.Path() != ds.Preview.Path() {
		if err := ImagesDelete(unid, ds.Thumbnail.Path()); err != nil && !errors.Is(err, ErrNoImages) {
			return fmt.Errorf(format, "delete existing", err)
		}
	}

	// range each image file extension looking for matches
	for _, ext := range imagesExt[:] {
		src := filepath.Join(ds.Preview.Path(), unid+ext)
		if _, err := os.Stat(src); err != nil {
			continue
		}
		var err error
		switch generate {
		case Pixel:
			err = ds.thumbPixels(ctx, sl, src, unid)
		case Photo:
			err = ds.thumbPhoto(ctx, sl, src, unid)
		}
		if err != nil {
			return fmt.Errorf(format, "conversion failed", err)
		}
		return nil
	}
	return nil
}
