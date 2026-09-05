//nolint:exhaustruct_v5,gochecknoglobals
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
	"strings"
	"unicode/utf8"

	"github.com/Defacto2/magicnumber"
	"github.com/Defacto2/server/internal/command/option"
	"github.com/Defacto2/server/internal/dir"
	"github.com/Defacto2/server/internal/nils"
)

const (
	ANSICap = 350_000 // ANSICap is the maximum file size in bytes for an ANSI encoded text file.
)

const (
	check    = "check"
	prevChk  = "preview check"
	thumbChk = "thumbnail check"
)

// ImagesExt returns args slice of image file extensions used by the website
// preview and thumbnail images, including the legacy and modern formats.
var imagesExt = [...]string{
	gif, ".GIF", jpg, ".JPG", jpeg, ".JPEG", png, ".PNG", webp, ".WEBP", ".avif", ".AVIF",
}

// ImagesDelete removes images matching the unid from the specified directories.
// The unid is the unique identifier shared across preview and thumbnail variants.
// Returns ErrNoImages if no matching files were found and removed.
func ImagesDelete(unid string, dirs ...string) error {
	const format = "images delete %s: %w"
	if unid == "" {
		return fmt.Errorf(format, "no unid", ErrIsEmpty)
	}

	exts := imagesExt[:]
	deletedAny := false

	for _, dir := range dirs {
		if dir == "" {
			continue
		}

		st, err := os.Stat(dir)
		if err != nil {
			return fmt.Errorf(format, dir, err)
		}
		if !st.IsDir() {
			return fmt.Errorf(format, dir, ErrIsFile)
		}

		// delete all matching extension variants
		for _, ext := range exts {
			name := filepath.Join(dir, unid+ext)

			err := os.Remove(name)
			if err == nil {
				deletedAny = true
				continue
			}
			if !errors.Is(err, os.ErrNotExist) {
				return fmt.Errorf(format, name, err)
			}
		}
	}

	if !deletedAny {
		return fmt.Errorf(format, dirs, ErrNoImages)
	}

	return nil
}

// ImagesPixelate converts images matching unid across the specified directories
// into pixelated versions in-place.
func ImagesPixelate(ctx context.Context, sl *slog.Logger, unid string, dirs ...string) error {
	const format = "images pixelate %s: %w"
	if err := nils.Check(ctx, sl); err != nil {
		return fmt.Errorf(format, "check", err)
	}

	base := filepath.Base(unid)
	if base == "" || base == "." {
		return fmt.Errorf(format, "base", ErrValue)
	}

	for _, dir := range dirs {
		st, err := os.Stat(dir)
		if err != nil {
			return fmt.Errorf(format, "stat dir", err)
		}
		if !st.IsDir() {
			return fmt.Errorf(format, "is dir check", ErrIsFile)
		}

		// range each image file extension looking for matches
		for _, ext := range imagesExt[:] {
			if ctxErr := ctx.Err(); ctxErr != nil {
				return fmt.Errorf(format, "context canceled", ctxErr)
			}

			name := filepath.Join(dir, base+ext)
			if _, err := os.Stat(name); err != nil {
				if errors.Is(err, os.ErrNotExist) {
					continue
				}
				return fmt.Errorf(format, "stat image", err)
			}

			arg := option.Opts{}
			arg.Pixelate(name)
			r := Runner{Log: sl}
			if _, err := r.Run(ctx, Magick, arg...); err != nil {
				return fmt.Errorf(format, "convert", err)
			}
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

func (align Align) String() string {
	return [...]string{"top", "middle", "bottom", "left", "right"}[align]
}

// Thumbs generates a cropped thumbnail from the source preview image using the specified alignment.
func (align Align) Thumbs(ctx context.Context, sl *slog.Logger, unid string, preview, thumbnail dir.Directory) error {
	const msg = "thumbs re-alignment"
	const format = msg + " %s: %w"

	if err := nils.Check(ctx, sl); err != nil {
		return fmt.Errorf(format, check, err)
	}
	if err := preview.Check(sl); err != nil {
		return fmt.Errorf(format, prevChk, err)
	}
	if err := thumbnail.Check(sl); err != nil {
		return fmt.Errorf(format, thumbChk, err)
	}

	switch align {
	case Top, Middle, Bottom, Left, Right: // sanity check
	default:
		return fmt.Errorf(format, fmt.Sprintf("alignment value %d", align), ErrAlign)
	}

	// prep an isolated temporary directory
	tmpDir, err := dir.MkdirTemp(unid)
	if err != nil {
		return fmt.Errorf(format, "create temporary directory", err)
	}
	defer func() {
		if err := dir.RemoveAll(tmpDir); err != nil {
			slog.Info(msg+" remove all temporary directory",
				slog.String("directory", tmpDir), slog.Any("error", err))
		}
	}()

	// remove existing thumbnails
	if err := ImagesDelete(unid, thumbnail.Path()); err != nil && !errors.Is(err, ErrNoImages) {
		return fmt.Errorf(format, "delete existing thumbs", err)
	}

	// range each image file extension looking for matches
	for _, ext := range imagesExt[:] {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return fmt.Errorf(format, "context canceled", ctxErr)
		}

		src := preview.Join(unid + ext)
		if _, err := os.Stat(src); err != nil {
			continue
		}

		tmp := filepath.Join(tmpDir, unid+ext)
		arg := option.Opts{}
		arg.ThumbAlignment(src, tmp, int(align))
		r := Runner{Log: sl}
		if _, err := r.Run(ctx, Magick, arg...); err != nil {
			return fmt.Errorf(format, "run magick", err)
		}

		// copy thumbs to their final destination
		finalDst := thumbnail.Join(unid + ext)
		if err := CopyFile(sl, tmp, finalDst); err != nil {
			return fmt.Errorf(format, "copy file", err)
		}

		return nil
	}
	return fmt.Errorf(format, "none", ErrNoImages)
}

// Crop is an args type that represents the crop position of the preview image.
type Crop int

const (
	SquareTop Crop = iota // SquareTop crops the top of the image using args 1:1 ratio
	FourThree             // FourThree crops the top of the image using args 4:3 ratio
	OneTwo                // OneTwo crops the top of the image using args 1:2 ratio
)

func (crop Crop) String() string {
	return [...]string{"1:1", "4:3", "1:2"}[crop]
}

// Images crops the preview image based on the crop position and ratio of the image.
func (crop Crop) Images(ctx context.Context, sl *slog.Logger, unid string, preview dir.Directory) error {
	const format = "crop images %s: %w"
	if err := nils.Check(ctx, sl); err != nil {
		return fmt.Errorf(format, check, err)
	}
	if err := preview.Check(sl); err != nil {
		return fmt.Errorf(format, prevChk, err)
	}

	switch crop {
	case SquareTop, FourThree, OneTwo:
	default:
		return fmt.Errorf(format, fmt.Sprintf("crop value %d", crop), ErrCrop)
	}

	path, err := dir.MkdirTemp(unid)
	if err != nil {
		return fmt.Errorf(format, "make "+unid+" temp directory", err)
	}

	find := false
	for _, ext := range imagesExt {
		src := preview.Join(unid + ext)
		if _, err := os.Stat(src); err != nil {
			continue
		}

		find = true
		arg := option.Opts{}
		tmp := filepath.Join(path, unid+ext)
		arg.CropAlignment(src, tmp, int(crop))
		r := Runner{Log: sl}
		_, err := r.Run(ctx, Magick, arg...)
		if err != nil {
			return fmt.Errorf(format, "", err)
		}

		dst := preview.Join(unid + ext)
		if err := CopyFile(sl, tmp, dst); err != nil {
			slog.Info("crop image copyfile error", slog.String("source", tmp),
				slog.String("dest", dst), slog.Any("error", err))
			return nil
		}
	}

	if !find {
		return fmt.Errorf(format, "cannot use", ErrNoImages)
	}

	return nil
}

type Text struct {
	UUID    string // unique universal ID used as the base filename for the output
	MaxRows int    // maximum number of rows to crop
	MaxCols int    // maximum number of to crop, or leave as 0 to use the default of 80
	UTF8    bool   // when true, MaxCols will crop  to Unicode rune count instead of ASCII characters
}

// Crop reads the src text file and writes the number of MaxRows (lines) of text to a new file.
// The new file is stored in the same directory as the src but is given file name
// of UUID appended with a ".txt" extension.
// The text is truncated to MaxCols (runes or characters per line),
// but any leading empty lines are ignored.
//
// If MaxRows is 0, a default value of 29 is used.
// If MaxCols is set to 0, a default of 80 is used.
// Both defaults allow for a 400x400 pixel thumbnail to be generated.
//
//   - When successful the full path to the new file with the cropped text is returned.
//   - If unsuccessful the src path is returned.
//   - If a large number of ANSI sequences are detected then an ErrIsAnsi error is returned.
func (t Text) Crop(sl *slog.Logger, src string) (string, error) {
	const format = "crop text imager %s: %w"
	if err := nils.Check(sl); err != nil {
		return "", fmt.Errorf(format, check, err)
	}

	if t.UUID == "" {
		return "", fmt.Errorf(format, "uuid", ErrValue)
	}

	newpath, err := dir.MkdirTemp(t.UUID + "_txt")
	if err != nil {
		return "", fmt.Errorf(format, "make content", err)
	}

	dst := filepath.Join(newpath, t.UUID+".txt")
	err = t.crop(sl, src, dst)
	if err != nil {
		if !errors.Is(err, ErrIsAnsi) {
			return "", fmt.Errorf(format, "cropper", err)
		}
		if err := ansiCheck(src); err != nil {
			return "", fmt.Errorf(format, "ansi check", err)
		}
	}

	if !exist(dst) {
		return src, nil
	}

	return dst, nil
}

func ansiCheck(src string) error {
	file, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("src: %s: %w", src, err)
	}
	if file.Size() > ANSICap {
		return ErrIsAnsi
	}
	// allow processing to continue for small ANSI files
	return nil
}

// exists is a special case where the error value should not be returned.
func exist(name string) bool {
	if name == "" {
		return false
	}
	st, err := os.Stat(name)
	if err != nil {
		return false
	}

	return st.Size() > 0
}

func (t Text) crop(sl *slog.Logger, src, dst string) error { //nolint:funlen
	const format = "text scanner %s: %w"
	if err := nils.Check(sl); err != nil {
		return fmt.Errorf(format, check, err)
	}

	src = filepath.Clean(src)
	dst = filepath.Clean(dst)

	r, err := os.Open(src)
	if err != nil {
		return fmt.Errorf(format, "open source", err)
	}
	defer func() {
		if err := r.Close(); err != nil {
			sl.Info("text crop error closing source file", slog.String("file", src), slog.Any("error", err))
		}
	}()

	// return an error if any ansi sequences are found
	if magicnumber.CSI(r) {
		return fmt.Errorf(format, " "+src, ErrIsAnsi)
	}
	const reset = 0
	if _, err := r.Seek(reset, io.SeekStart); err != nil {
		return fmt.Errorf(format, "seek reset", err)
	}

	w, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf(format, "create destination", err)
	}
	defer func() {
		if err := w.Close(); err != nil {
			sl.Info("text crop closing destination file", slog.String("file", dst), slog.Any("error", err))
		}
	}()

	if t.MaxCols < 1 {
		t.MaxCols = 80
	}
	if t.MaxRows < 1 {
		t.MaxRows = 29
	}

	scanner := bufio.NewScanner(r)
	writer := bufio.NewWriter(w)
	defer func() {
		if err := writer.Flush(); err != nil {
			sl.Info("text crop flush writer", slog.Any("error", err))
		}
	}()

	rows := 0
	skip := true

	for scanner.Scan() {
		line := scanner.Text()
		if skip {
			// skip leading empty lines
			if strings.TrimSpace(line) == "" {
				continue
			}
			skip = false
		}
		if rows >= t.MaxRows {
			// end of crop
			break
		}
		line = t.truncate(line)

		if _, err := writer.WriteString(line); err != nil {
			return fmt.Errorf(format, "writer write string", err)
		}
		if err := writer.WriteByte('\n'); err != nil {
			return fmt.Errorf("text crop write newline: %w", err)
		}

		rows++
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf(format, "scanner line", err)
	}

	if err := writer.Flush(); err != nil {
		return fmt.Errorf(format, "flush writer", err)
	}

	return nil
}

// truncate safely clips a UTF-8 string to a maximum number of runes (characters).
//
// when UTF8 is false, it instead clips the maximum number of bytes.
func (t Text) truncate(line string) string {
	cols := t.MaxCols
	if !t.UTF8 {
		if len(line) > cols {
			line = line[:cols]
		}
		return line
	}
	if utf8.RuneCountInString(line) <= cols {
		return line
	}
	r := []rune(line)
	return string(r[:cols])
}

type Thumb struct {
	Source  string // Source image file
	UUID    string // unique universal ID used as the base filename for the output
	Pattern string // filename pattern for the temporary directory
	JPEG    bool   // make JPEG thumbnails instead of the default PNG format
}

func (t Thumb) make(ctx context.Context, sl *slog.Logger, thumb dir.Directory) error {
	const format = "thumb make %s: %w"

	if err := nils.Check(ctx, sl); err != nil {
		return fmt.Errorf(format, check, err)
	}

	base := filepath.Base(t.UUID)
	if base == "" || base == "." {
		return fmt.Errorf(format, "unid", ErrValue)
	}

	pattern := t.Pattern
	if pattern == "" {
		pattern = "thumb-*"
	}

	tmpFile, err := dir.CreateTemp(pattern)
	if err != nil {
		return fmt.Errorf(format, "create temp file", err)
	}
	tmp := tmpFile.Name()

	// close immediately so external processes can overwrite/read the file
	if closeErr := tmpFile.Close(); closeErr != nil && sl != nil {
		sl.Debug("thumb maker could not close initial temp handle", slog.Any("error", closeErr))
	}

	defer func() {
		if remErr := os.Remove(tmp); remErr != nil && !os.IsNotExist(remErr) && sl != nil {
			sl.Error("thumb maker could not remove temporary file", slog.String("file", tmp), slog.Any("error", remErr))
		}
	}()

	// render source to an intermediate image
	arg := option.Opts{}
	if t.JPEG {
		arg.JPGPhoto(true, t.Source, tmp)
	} else {
		arg.PNGPixel(true, t.Source, tmp)
	}
	r := Runner{Log: sl}
	if out, err := r.Run(ctx, Magick, arg...); err != nil {
		return fmt.Errorf(format, "run magick convert "+string(out), err)
	}

	if ctxErr := ctx.Err(); ctxErr != nil {
		return fmt.Errorf(format, "context canceled", ctxErr)
	}

	// convert intermediate image to WebP
	dst := filepath.Join(thumb.Path(), base+webp)
	arg = option.Opts{}
	arg.WebpPhoto(tmp, dst)
	if _, err := r.Run(ctx, Cwebp, arg...); err != nil {
		return fmt.Errorf(format, "run cwebp", err)
	}

	return nil
}

// OptimizePNG optimizes the src PNG image in-place using optipng.
// It is safe to call within a deferred function.
func OptimizePNG(ctx context.Context, sl *slog.Logger, src string) error {
	const format = "optimize png using optipng %s: %w"
	if err := nils.Check(ctx, sl); err != nil {
		return fmt.Errorf(format, check, err)
	}

	st, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf(format, src, err)
	}
	if st.Size() == 0 {
		return fmt.Errorf(format, src, ErrIsEmpty)
	}

	const tiny = 67
	// skip optimization for tiny files
	if st.Size() < tiny {
		return nil
	}

	r := Runner{Log: sl}
	if _, err := r.Run(ctx, Optipng, src); err != nil {
		return fmt.Errorf(format, src, err)
	}

	return nil
}
