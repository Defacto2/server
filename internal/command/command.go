// Package command provides interfaces for the shell commands and programs on the host.
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
	ErrMatch = errors.New("no match value is present")
	ErrVers  = errors.New("version mismatch")
	ErrZap   = errors.New("zap logger instance is nil")
)

// Dirs is a struct of the download, preview and thumbnail directories.
type Dirs struct {
	Download  string // Download is the directory path for the file downloads.
	Preview   string // Preview is the directory path for the image previews.
	Thumbnail string // Thumbnail is the directory path for the image thumbnails.
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
	// A note about unrar on linux, the installation cannot use the unrar-free package,
	// which is a poor substitute for the files this application needs to handle.
	// The unrar binary should return:
	// "UNRAR 6.24 freeware, Copyright (c) 1993-2023 Alexander Roshal".
	Unrar = "unrar" // Unrar is the rar decompression command.
	Unzip = "unzip" // Unzip is the zip decompression command.
)

// Lookups returns a list of the execute command names used by the application.
func Lookups() []string {
	return []string{
		Arc,
		Arj,
		Ansilove,
		Convert,
		Cwebp,
		Gwebp,
		Optipng,
		P7zip,
		Tar,
		Unrar,
		Unzip,
	}
}

// Infos returns details for the list of the execute command names used by the application.
func Infos() []string {
	return []string{
		"archive utility ver 5+",
		"arj32 ver 3+",
		"ansilove/c ver 4+",
		"ImageMagick ver 7+",
		"Google WebP ver 1+",
		"Google GIF to WebP ver 1+",
		"OptiPNG optimizer ver 0.7+",
		"\t7-Zip ver 16+",
		"GNU tar ver 1+",
		"UNRAR freeware (c) Alexander Roshal",
		"Info-ZIP ver 6+",
	}
}

// LookupUnrar returns an error if the name Alexander Roshal is not found in the unrar version output.
func LookupUnrar() error {
	return LookVersion(Unrar, "", "Alexander Roshal")
}

// RemoveImgs removes unid named images with .jpg, .png and .webp extensions from the directory paths.
// A nil is returned if the directory or the named unid files do not exist.
func RemoveImgs(unid string, dirs ...string) error {
	exts := []string{jpg, png, webp}
	for _, path := range dirs {
		// remove images
		for _, ext := range exts {
			name := filepath.Join(path, unid+ext)
			st, err := os.Stat(name)
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			if st.IsDir() {
				return ErrIsDir
			}
			if err = os.Remove(name); err != nil {
				return fmt.Errorf("os.Remove: %w", err)
			}
		}
	}
	return nil
}

// RemoveMe removes the file with the unid name combined with a ".txt" extension
// from the download directory path. It returns nil if the file does not exist.
func RemoveMe(unid, dir string) error {
	name := filepath.Join(dir, unid+txt)
	st, err := os.Stat(name)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("os.Stat: %w", err)
	}
	if st.IsDir() {
		return ErrIsDir
	}
	if err := os.Remove(name); err != nil {
		return fmt.Errorf("os.Remove: %w", err)
	}
	return nil
}

// CopyFile copies the src file to the dst file and path.
func CopyFile(logger *zap.SugaredLogger, src, dst string) error {
	if logger == nil {
		return ErrZap
	}

	s, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("os.Open: %w", err)
	}
	defer s.Close()

	d, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("os.Create: %w", err)
	}
	defer d.Close()

	i, err := io.Copy(d, s)
	if err != nil {
		return fmt.Errorf("io.Copy: %w", err)
	}
	logger.Infof("copyfile: copied %d bytes to %s\n", i, dst)

	if err := d.Sync(); err != nil {
		return fmt.Errorf("d.Sync: %w", err)
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
		return fmt.Errorf("exec.LookPath: %w", err)
	}
	return nil
}

// LookVersion returns an error when the match string is not found in the named command output.
func LookVersion(name, flag, match string) error {
	if err := LookCmd(name); err != nil {
		return fmt.Errorf("lookcmd: %w", err)
	}
	if match == "" {
		return ErrMatch
	}
	cmd := exec.Command(name, flag)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("cmd.StdoutPipe: %w", err)
	}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("cmd.Start: %w", err)
	}
	b, err := io.ReadAll(stdout)
	if err != nil {
		return fmt.Errorf("io.ReadAll: %w", err)
	}
	if !bytes.Contains(b, []byte(match)) {
		return fmt.Errorf("%w: %s", ErrVers, name)
	}
	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("cmd.Wait: %w", err)
	}
	return nil
}

// Run looks for the command in the system path and executes it with the arguments.
// Any output to stderr is logged as a debug message.
func Run(logger *zap.SugaredLogger, name string, arg ...string) error {
	if logger == nil {
		return ErrZap
	}
	return run(logger, name, "", arg...)
}

// RunOut looks for the command in the system path and executes it with the arguments.
// Any output is sent to the stdout buffer.
func RunOut(name string, arg ...string) ([]byte, error) {
	if err := LookCmd(name); err != nil {
		return nil, fmt.Errorf("lookcmd: %w", err)
	}
	var out bytes.Buffer
	cmd := exec.Command(name, arg...)
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("cmd.Run: %w", err)
	}
	return out.Bytes(), nil
}

// RunQuiet looks for the command in the system path and executes it with the arguments.
func RunQuiet(name string, arg ...string) error {
	if err := LookCmd(name); err != nil {
		return fmt.Errorf("lookcmd: %w", err)
	}
	cmd := exec.Command(name, arg...)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("cmd.Start: %w", err)
	}
	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("cmd.Wait: %w", err)
	}
	return nil
}

// RunWD looks for the command in the system path and executes it with the arguments.
// An optional working directory is set for the command.
// Any output to stderr is logged as a debug message.
func RunWD(logger *zap.SugaredLogger, name, wdir string, arg ...string) error {
	if logger == nil {
		return ErrZap
	}
	return run(logger, name, wdir, arg...)
}

func run(logger *zap.SugaredLogger, name, wdir string, arg ...string) error {
	if err := LookCmd(name); err != nil {
		return fmt.Errorf("lookcmd: %w", err)
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
		logger.Debugf("run %q: %s\n", cmd, string(b))
	}

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("cmd.Wait: %w", err)
	}
	return nil
}
