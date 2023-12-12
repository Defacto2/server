package command

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

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
func UnzipOne(z *zap.SugaredLogger, src, dst, name string) error {
	if z == nil {
		return ErrZap
	}

	const cmd = "unzip"
	_, err := exec.LookPath(cmd)
	if errors.Is(err, exec.ErrDot) {
		err = nil
	}
	if err != nil {
		return err
	}

	st, err := os.Stat(src)
	if err != nil {
		return err
	}
	if st.IsDir() {
		return ErrIsDir
	}
	if st.Size() == 0 {
		return ErrEmpty
	}

	tmp, err := os.MkdirTemp(os.TempDir(), pattern)
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)

	const exdir = "-d" // Directory to which to extract files.
	out, err := exec.Command(cmd, src, name, exdir, tmp).Output()
	if errors.Is(err, exec.ErrDot) {
		err = nil
	}
	if err != nil {
		return err
	}
	z.Info("unzipone: ", cmd, string(out))

	extracted := filepath.Join(tmp, name)
	st, err = os.Stat(extracted)
	if err != nil {
		return err
	}
	if st.IsDir() {
		return ErrIsDir
	}
	if st.Size() == 0 {
		return ErrEmpty
	}

	return CopyFile(z, extracted, dst)
}

func (dir Dirs) UnzipImg(z *zap.SugaredLogger, src, uuid, name string) error {
	if z == nil {
		return ErrZap
	}

	tmp, err := os.MkdirTemp(os.TempDir(), pattern)
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)

	dst := filepath.Join(tmp, filepath.Base(name))
	if err = UnzipOne(z, src, dst, name); err != nil {
		return err
	}

	st, err := os.Stat(dst)
	if err != nil {
		return err
	}
	if st.IsDir() {
		return ErrIsDir
	}

	switch filepath.Ext(dst) {
	case ".gif":
		// use lossless compression (but larger file size)
		err = dir.ConvertLossless(z, dst, uuid)
	case ".bmp":
		// use lossy compression that removes some details (but smaller file size)
		err = dir.ConvertLossy(z, dst, uuid)
	case ".jpeg", ".jpg", ".tiff", ".webp":
		// keep the same format, but optimize the file size
		err = dir.WebpScreenshot(z, dst, uuid)
	default:
		return fmt.Errorf("%w: %q", ErrImg, filepath.Ext(dst))
	}
	if err != nil {
		return err
	}
	return nil
}
