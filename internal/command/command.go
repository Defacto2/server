// Package command provides interfaces for shell commands and applications on the host system.
package command

import (
	"bytes"
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
	arc     = ".arc"      // arc file extension
	arj     = ".arj"      // arj file extension
	bmp     = ".bmp"      // bmp file extension
	gif     = ".gif"      // gif file extension
	gzip    = ".gz"       // gzip file extension
	jpg     = ".jpg"      // jpg file extension
	jpeg    = ".jpeg"     // jpeg file extension
	p7zip   = ".7z"       // 7zip file extension
	png     = ".png"      // png file extension
	rar     = ".rar"      // rar file extension
	tar     = ".tar"      // tar file extension
	tiff    = ".tiff"     // tiff file extension
	txt     = ".txt"      // txt file extension
	webp    = ".webp"     // webp file extension
	zip     = ".zip"      // zip file extension
)

var (
	ErrEmpty = errors.New("file is empty")
	ErrImg   = errors.New("file is not an known image format")
	ErrIsDir = errors.New("file is a directory")
	ErrZap   = errors.New("zap logger instance is nil")
)

// Dirs is a struct of the download, preview and thumbnail directories.
type Dirs struct {
	Download  string
	Preview   string
	Thumbnail string
}

const (
	Arc      = "arc"      // Arc is the arc decompression command.
	Arj      = "arj"      // Arj is the arj decompression command.
	Ansilove = "ansilove" // Ansilove is the ansilove text to image command.
	Convert  = "convert"  // Convert is the ImageMagick convert command.
	Cwebp    = "cwebp"    // Cwebp is the Google create webp command.
	Gwebp    = "gif2webp" // Gwebp is the Google gif to webp command.
	Optipng  = "optipng"  // Optipng is the PNG optimizer command.
	P7zip    = "7z"       // P7zip is the 7-Zip decompression command.
	Tar      = "tar"      // Tar is the tar decompression command.
	// A note about unrar on linux, this cannot be the unrar-free package
	// which is a poor substitute for the files this application needs to handle.
	// The unrar binary should return "UNRAR 6.24 freeware      Copyright (c) 1993-2023 Alexander Roshal".
	Unrar = "unrar" // Unrar is the rar decompression command.
	Unzip = "unzip" // Unzip is the zip decompression command.
)

// Lookups returns a list of the execute command names used by the application.
func Lookups() []string {
	return []string{Convert, Cwebp, Optipng, Unzip}
}

// RemoveImgs removes the preview and thumbnail images from the preview and
// thumbnail directories associated with the uuid.
// It returns nil if the files do not exist.
func RemoveImgs(preview, thumb, uuid string) error {
	exts := []string{jpg, png, webp}
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
	name := filepath.Join(download, uuid+txt)
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

// CopyFile copies the src file to the dst file and path.
func CopyFile(z *zap.SugaredLogger, src, dst string) error {
	if z == nil {
		return ErrZap
	}

	s, err := os.Open(src)
	if err != nil {
		return err
	}
	defer s.Close()

	d, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer d.Close()

	i, err := io.Copy(d, s)
	if err != nil {
		return err
	}
	z.Infof("copyfile: copied %d bytes to %s\n", i, dst)

	return d.Sync()
}

// BaseName returns the base name of the file without the extension.
// Both the directory and extension are removed.
func BaseName(path string) string {
	if path == "" {
		return ""
	}
	return strings.TrimSuffix(filepath.Base(path), filepath.Ext(filepath.Base(path)))
}

// BaseNamePath returns the directory and base name of the file without the extension.
func BaseNamePath(path string) string {
	if path == "" {
		return ""
	}
	return filepath.Join(filepath.Dir(path), BaseName(path))
}

// LookCmd returns an error if the named command is not found in the system path.
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

// Run looks for the command in the system path and executes it with the arguments.
// Any output to stderr is logged as a debug message.
func Run(z *zap.SugaredLogger, name string, arg ...string) error {
	if z == nil {
		return ErrZap
	}
	return run(z, name, "", arg...)
}

func RunOut(z *zap.SugaredLogger, name string, arg ...string) ([]byte, error) {
	if err := LookCmd(name); err != nil {
		return nil, err
	}

	var out bytes.Buffer
	cmd := exec.Command(name, arg...)
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}

// Run looks for the command in the system path and executes it with the arguments.
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

// RunWD looks for the command in the system path and executes it with the arguments.
// The working directory is set for the command.
// Any output to stderr is logged as a debug message.
func RunWD(z *zap.SugaredLogger, name, wdir string, arg ...string) error {
	if z == nil {
		return ErrZap
	}
	return run(z, name, wdir, arg...)
}

func run(z *zap.SugaredLogger, name, wdir string, arg ...string) error {
	if err := LookCmd(name); err != nil {
		return err
	}

	cmd := exec.Command(name, arg...)
	cmd.Dir = wdir
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("could not get stderr pipe: %w", err)
	}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("could not start command: %w", err)
	}
	b, err := io.ReadAll(stderr)
	if err != nil {
		return fmt.Errorf("could not read stderr: %w", err)
	}
	if len(b) > 0 {
		z.Debugf("run %q: %s\n", cmd, string(b))
	}
	if err := cmd.Wait(); err != nil {
		return err
	}
	return nil
}
