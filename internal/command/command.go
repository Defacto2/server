// Package command provides interfaces for the shell commands and programs on the host.
package command

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/Defacto2/server/internal/dir"
	"go.uber.org/zap"
)

const (
	cmdTimeout = 10 * time.Second
	pattern    = "defacto2-" // prefix for temporary directories
	patternS   = "defacto2-server"
	arc        = ".arc"  // arc file extension
	arj        = ".arj"  // arj file extension
	bmp        = ".bmp"  // bmp file extension
	gif        = ".gif"  // gif file extension
	gzip       = ".gz"   // gzip file extension
	jpg        = ".jpg"  // jpg file extension
	jpeg       = ".jpeg" // jpeg file extension
	png        = ".png"  // png file extension
	rar        = ".rar"  // rar file extension
	tar        = ".tar"  // tar file extension
	tiff       = ".tiff" // tiff file extension
	txt        = ".txt"  // txt file extension
	webp       = ".webp" // webp file extension
	zip        = ".zip"  // zip file extension
	zip7       = ".7z"   // 7zip file extension
)

var (
	ErrANSI   = errors.New("text is ansi encoded, cannot crop")
	ErrEmpty  = errors.New("file is empty")
	ErrImg    = errors.New("file is not an known image format")
	ErrIsDir  = errors.New("file is a directory")
	ErrIsFile = errors.New("directory path points to a file")
	ErrMatch  = errors.New("no match value is present")
	ErrPath   = errors.New("path is not permitted")
	ErrVers   = errors.New("version mismatch")
	ErrZap    = errors.New("zap logger instance is nil")
)

// Dirs is a struct of the download, preview and thumbnail directories.
type Dirs struct {
	Download  dir.Directory // Download is the directory path for the file downloads.
	Preview   dir.Directory // Preview is the directory path for the image previews.
	Thumbnail dir.Directory // Thumbnail is the directory path for the image thumbnails.
	Extra     dir.Directory // Extra is the directory path for the extra files.
}

const (
	Arc      = "arc"      // Arc is the arc decompression command.
	Arj      = "arj"      // Arj is the arj decompression command.
	Ansilove = "ansilove" // Ansilove is the ansilove text to image command.
	Cwebp    = "cwebp"    // Cwebp is the Google create webp command.
	Gwebp    = "gif2webp" // Gwebp is the Google gif to webp command.
	HWZip    = "hwzip"    // Hwzip the zip decompression command for files using obsolete methods.
	Lha      = "lha"      // Lha is the lha/lzh decompression command.
	Magick   = "magick"   // Magick is the ImageMagick v7+ command.
	Optipng  = "optipng"  // Optipng is the PNG optimizer command.
	Tar      = "tar"      // Tar is the tar decompression command.
	// A note about unrar on linux, the installation cannot use the unrar-free package,
	// which is a poor substitute for the files this application needs to handle.
	// The unrar binary should return:
	// "UNRAR 6.24 freeware, Copyright (c) 1993-2023 Alexander Roshal".
	Unrar   = "unrar"   // Unrar is the rar decompression command.
	Unzip   = "unzip"   // Unzip is the zip decompression command.
	Zip7    = "7zz"     // Zip7 is the 7-Zip decompression command.
	ZipInfo = "zipinfo" // ZipInfo is the zip information command.
)

// Lookups returns a list of the execute command names used by the application.
func Lookups() []string {
	return []string{
		Arc,
		Arj,
		Ansilove,
		Cwebp,
		Gwebp,
		HWZip,
		Lha,
		Magick,
		Optipng,
		Tar,
		Unrar,
		Unzip,
		Zip7,
		ZipInfo,
	}
}

// Infos returns details for the list of the execute command names used by the application.
func Infos() []string {
	return []string{
		"archive utility ver 5+",
		"arj32 ver 3+",
		"ansilove/c ver 4+",
		"Google WebP ver 1+",
		"Google GIF to WebP ver 1+",
		"HWZip ver 2+",
		"Lhasa command line LHA tool",
		"ImageMagick ver 7+",
		"OptiPNG optimizer ver 0.7+",
		"GNU tar ver 1+",
		"UNRAR freeware (c) Alexander Roshal",
		"UnZip Info-ZIP ver 6+",
		"7-Zip ver 24+",
		"ZipInfo Info-ZIP ver 3+",
	}
}

// LookupUnrar returns an error if the name Alexander Roshal is not found in the unrar version output.
func LookupUnrar() error {
	return LookVersion(Unrar, "", "Alexander Roshal")
}

// CopyFile copies the src file to the dst file and path.
func CopyFile(debug *zap.SugaredLogger, src, dst string) error {
	s, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("copy file open %w", err)
	}
	defer func() { _ = s.Close() }()

	d, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("copy file create %w", err)
	}
	defer func() { _ = d.Close() }()

	// io.CopyBuffer is faster than io.Copy
	const size = 4 * 1024
	buf := make([]byte, size)
	i, err := io.CopyBuffer(d, s, buf)
	if err != nil {
		return fmt.Errorf("copy file io.copybuffer %w", err)
	}

	if debug != nil {
		debug.Infof("copyfile: copied %d bytes to %s\n", i, dst)
	}
	if err := d.Sync(); err != nil {
		return fmt.Errorf("copy file sync %w", err)
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
		return fmt.Errorf("exec look path %w", err)
	}
	return nil
}

// LookVersion returns an error when the match string is not found in the named command output.
func LookVersion(name, flag, match string) error {
	if err := LookCmd(name); err != nil {
		return fmt.Errorf("look version %w", err)
	}
	if match == "" {
		return ErrMatch
	}
	cmd := exec.Command(name, flag)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("look version stdout pipe %w", err)
	}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("look version start %w", err)
	}
	b, err := io.ReadAll(stdout)
	if err != nil {
		return fmt.Errorf("look version read all %w", err)
	}
	if !bytes.Contains(b, []byte(match)) {
		return fmt.Errorf("look version %w: %s", ErrVers, name)
	}
	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("look version wait %w", err)
	}
	return nil
}

// Run looks for the command in the system path and executes it with the arguments.
// Any output to stderr is logged as a debug message.
func Run(debug *zap.SugaredLogger, name string, arg ...string) error {
	return run(debug, name, "", arg...)
}

// RunStdOut looks for the command in the system path and executes it with the arguments.
// Any output is sent to the stdout buffer.
func RunStdOut(name string, arg ...string) ([]byte, error) {
	if err := LookCmd(name); err != nil {
		return nil, fmt.Errorf("run output %w", err)
	}
	var out bytes.Buffer
	ctx, cancel := context.WithTimeout(context.Background(), cmdTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, name, arg...)
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("run output cmd.run %w", err)
	}
	return out.Bytes(), nil
}

// RunQuiet looks for the command in the system path and executes it with the arguments.
func RunQuiet(name string, arg ...string) error {
	if err := LookCmd(name); err != nil {
		return fmt.Errorf("run quiet %w", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), cmdTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, name, arg...)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("run quiet start %w", err)
	}
	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("run quiet wait %w", err)
	}
	return nil
}

// RunWorkdir looks for the command in the system path and executes it with the arguments.
// An optional working directory is set for the command.
// Any output to stderr is logged as a debug message.
func RunWorkdir(debug *zap.SugaredLogger, name, wdir string, arg ...string) error {
	return run(debug, name, wdir, arg...)
}

func run(debug *zap.SugaredLogger, name, wdir string, arg ...string) error {
	if err := LookCmd(name); err != nil {
		return fmt.Errorf("run %w", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), cmdTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, name, arg...)
	cmd.Dir = wdir
	if debug != nil {
		p, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("run could not start command %w", err)
		}
		debug.Debugf("run %q: %s\n", cmd, string(p))
		return nil
	}
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("run could not start command %w", err)
	}
	return nil
}

// UncontrolledPath returns an error if the path contains any of the following characters:
//   - /
//   - \
//   - ..
//
// This is to prevent directory traversal attacks.
func UncontrolledPath(path string) error {
	if strings.Contains(path, "/") || strings.Contains(path, "\\") || strings.Contains(path, "..") {
		return ErrPath
	}
	return nil
}
