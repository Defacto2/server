package command

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"go.uber.org/zap"
)

// UnzipOne extracts athe named file from a zip archive.
//
// The extracted file is copied to the src with the ext extension appended.
// It requires the [unzip] command to be available on the host system.
// This allows for better compatibility with retro zip archives,
// such as those that use the [compression methods] prior to zip deflate.
//
// The src argument is the path to the zip archive.
// The dst argument is the destination filepath and should end with
// a file extension, eg. ".txt".
// The name argument is the name of the one file to unzip and copy.
//
// [unzip]: https://sourceforge.net/projects/infozip
// [compression methods]: https://www.hanshq.net/zip.html
func UnZipOne(z *zap.SugaredLogger, src, dst, name string) error {
	if z == nil {
		return ErrZap
	}

	st, err := os.Stat(src)
	if err != nil {
		return err
	}
	if st.IsDir() {
		return fmt.Errorf("%w: %q", ErrIsDir, src)
	}
	if st.Size() == 0 {
		return fmt.Errorf("%w: %q", ErrEmpty, src)
	}

	tmp, err := os.MkdirTemp(os.TempDir(), pattern)
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)

	arg := []string{src}         // source zip archive
	arg = append(arg, name)      // target file to extract
	arg = append(arg, "-d", tmp) // extract destination
	if err := Run(z, Unzip, arg...); err != nil {
		return err
	}

	extracted := filepath.Join(tmp, name)
	st, err = os.Stat(extracted)
	if err != nil {
		return err
	}
	if st.IsDir() {
		return fmt.Errorf("%w: %q", ErrIsDir, extracted)
	}
	if st.Size() == 0 {
		return fmt.Errorf("%w: %q", ErrEmpty, extracted)
	}

	return CopyFile(z, extracted, dst)
}

// extract extracts the named file from a zip archive and returns the path to the file.
func extract(z *zap.SugaredLogger, src, uuid, name string) (string, error) {
	if z == nil {
		return "", ErrZap
	}
	tmp, err := os.MkdirTemp(os.TempDir(), pattern)
	if err != nil {
		return "", err
	}

	dst := filepath.Join(tmp, filepath.Base(name))
	if err = UnZipOne(z, src, dst, name); err != nil {
		return "", err
	}

	st, err := os.Stat(dst)
	if err != nil {
		return "", err
	}
	if st.IsDir() {
		return "", fmt.Errorf("%w: %q", ErrIsDir, dst)
	}
	return dst, nil
}

// ExtractImage extracts the named image file from a zip archive.
// Based on the file extension, the image is converted to a webp preview and thumbnails.
// Named files with a PNG extension are optimized but kept as the preview image.
func (dir Dirs) ExtractImage(z *zap.SugaredLogger, src, uuid, name string) error {
	if z == nil {
		return ErrZap
	}

	dst, err := extract(z, src, uuid, name)
	if err != nil {
		return err
	}
	defer os.RemoveAll(dst)

	switch filepath.Ext(strings.ToLower(dst)) {
	case gif:
		err = dir.PreviewGIF(z, dst, uuid)
	case bmp:
		err = dir.PreviewLossy(z, dst, uuid)
	case png:
		// optimize but keep the original png file as preview
		err = dir.PreviewPNG(z, dst, uuid)
	case jpeg, jpg, tiff, webp:
		// convert to the optimal webp format
		// as of 2023, webp is supported by all current browsers
		// these format cases are supported by cwebp conversion tool
		err = dir.PreviewWebP(z, dst, uuid)
	default:
		return fmt.Errorf("%w: %q", ErrImg, filepath.Ext(dst))
		// use lossless compression (but larger file size)
		//err = dir.LosslessScreenshot(z, dst, uuid)
	}
	if err != nil {
		return err
	}
	return nil
}

// ExtractAnsiLove extracts the named text file from a zip archive.
// The text file is converted to a PNG preview and a webp thumbnails.
// Any text file usable by the ansilove command is supported,
// including ANSI, codepage plain text, PCBoard, etc.
func (dir Dirs) ExtractAnsiLove(z *zap.SugaredLogger, src, uuid, name string) error {
	if z == nil {
		return ErrZap
	}

	dst, err := extract(z, src, uuid, name)
	if err != nil {
		return err
	}
	defer os.RemoveAll(dst)

	if err := dir.AnsiLove(z, dst, uuid); err != nil {
		return err
	}
	return nil
}
