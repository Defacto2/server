// Package command provides interfaces for shell commands and applications on the host system.
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

const (
	pattern = "defacto2-" // prefix for temporary directories
	gif     = ".gif"      // gif file extension
	jpg     = ".jpg"      // jpg file extension
	jpeg    = ".jpeg"     // jpeg file extension
	png     = ".png"      // png file extension
	webp    = ".webp"     // webp file extension
)

var (
	ErrEmpty = errors.New("file is empty")
	ErrImg   = errors.New("file is not an known image format")
	ErrIsDir = errors.New("file is a directory")
	ErrZap   = errors.New("zap logger instance is nil")
)

// Dirs is a struct of the download, screenshot and thumbnail directories.
type Dirs struct {
	Download   string
	Screenshot string
	Thumbnail  string
}

const (
	Convert = "convert" // Convert is the ImageMagick convert command.
	Cwebp   = "cwebp"   // Cwebp is the Google create webp command.
)

// RemoveImgs removes the preview and thumbnail images from the preview and
// thumbnail directories associated with the uuid.
// It returns nil if the files do not exist.
func RemoveImgs(preview, thumb, uuid string) error {
	exts := []string{".jpg", ".png", ".gif", ".webp"}
	// remove previews
	for _, ext := range exts {
		name := filepath.Join(preview, uuid+ext)
		st, err := os.Stat(name)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if st.IsDir() {
			return ErrIsDir
		}
		if err = os.Remove(name); err != nil {
			return err
		}
	}
	// remove thumbnails
	for _, ext := range exts {
		name := filepath.Join(thumb, uuid+ext)
		st, err := os.Stat(name)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if st.IsDir() {
			return ErrIsDir
		}
		if err = os.Remove(name); err != nil {
			return err
		}
	}
	return nil
}

// RemoveMe removes the file with the uuid name combined with a ".txt" extension
// from the download directory. It returns nil if the file does not exist.
func RemoveMe(download, uuid string) error {
	name := filepath.Join(download, uuid+".txt")
	st, err := os.Stat(name)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	if st.IsDir() {
		return ErrIsDir
	}
	return os.Remove(name)
}

func CopyFile(z *zap.SugaredLogger, src, dst string) error {
	if z == nil {
		return ErrZap
	}

	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	b, err := io.Copy(dstFile, srcFile)
	if err != nil {
		return err
	}
	z.Info("copyfile: " + fmt.Sprintf("copied %d bytes to %s", b, dst))

	return dstFile.Sync()
}

func LookCmd(name string) error {
	_, err := exec.LookPath(name)
	if errors.Is(err, exec.ErrDot) {
		err = nil
	}
	if err != nil {
		return err
	}
	return nil
}

// BaseName returns the base name of the file without the extension.
// Both the directory and extension are removed.
func BaseName(path string) string {
	if path == "" {
		return ""
	}
	return strings.TrimSuffix(filepath.Base(path), filepath.Ext(filepath.Base(path)))
}

func BaseNamePath(path string) string {
	if path == "" {
		return ""
	}
	return filepath.Join(filepath.Dir(path), BaseName(path))
}

func Run(z *zap.SugaredLogger, name string, arg ...string) error {
	if z == nil {
		return ErrZap
	}

	if err := LookCmd(name); err != nil {
		return err
	}

	cmd := exec.Command(name, arg...)
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	if b, _ := io.ReadAll(stderr); len(b) > 0 {
		z.Debugln("run %q: %s", cmd, string(b))
	}
	if err := cmd.Wait(); err != nil {
		return err
	}
	return nil
}

func RunQuiet(z *zap.SugaredLogger, name string, arg ...string) error {
	if z == nil {
		return ErrZap
	}

	if err := LookCmd(name); err != nil {
		return err
	}

	cmd := exec.Command(name, arg...)
	if err := cmd.Start(); err != nil {
		return err
	}
	if err := cmd.Wait(); err != nil {
		return err
	}
	return nil
}
