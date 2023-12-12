package command

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"go.uber.org/zap"
)

// UnzipOne extracts a single file from a zip archive.
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
	if err := RunQuiet(z, Unzip, arg...); err != nil {
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

func (dir Dirs) UnZipImage(z *zap.SugaredLogger, src, uuid, name string) error {
	if z == nil {
		return ErrZap
	}

	tmp, err := os.MkdirTemp(os.TempDir(), pattern)
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)

	dst := filepath.Join(tmp, filepath.Base(name))
	if err = UnZipOne(z, src, dst, name); err != nil {
		return err
	}

	st, err := os.Stat(dst)
	if err != nil {
		return err
	}
	if st.IsDir() {
		return fmt.Errorf("%w: %q", ErrIsDir, dst)
	}

	switch filepath.Ext(strings.ToLower(dst)) {
	case gif:
		// use lossless compression (but larger file size)
		err = dir.LosslessScreenshot(z, dst, uuid)
	case bmp:
		// use lossy compression that removes some details (but smaller file size)
		err = dir.LossyScreenshot(z, dst, uuid)
	case png:
		// optimize but keep the original png file as preview
		err = dir.PngScreenshot(z, dst, uuid)
	case jpeg, jpg, tiff, webp:
		// convert to the optimal webp format
		// as of 2023, webp is supported by all current browsers
		// these format cases are supported by cwebp conversion tool
		err = dir.WebpScreenshot(z, dst, uuid)
	default:
		return fmt.Errorf("%w: %q", ErrImg, filepath.Ext(dst))
	}
	if err != nil {
		return err
	}
	return nil
}

func (dir Dirs) UnZipAnsiLove(z *zap.SugaredLogger, src, uuid, name string) error {
	if z == nil {
		return ErrZap
	}
	return nil
}
