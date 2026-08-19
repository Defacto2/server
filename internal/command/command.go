// Package command provides interfaces for the shell commands and programs on the host.
package command

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/Defacto2/server/internal/dir"
	"github.com/Defacto2/server/internal/nils"
)

const (
	CmdTimeout = 10 * time.Second

	patternS = "defacto2-server"
	gif      = ".gif"  // gif file extension
	jpg      = ".jpg"  // jpg file extension
	jpeg     = ".jpeg" // jpeg file extension
	png      = ".png"  // png file extension
	webp     = ".webp" // webp file extension
)

var (
	ErrAlign      = errors.New("invalid align choice")
	ErrCrop       = errors.New("invalid crop choice")
	ErrThumb      = errors.New("invalid thumb choice")
	ErrIsAnsi     = errors.New("text is ansi encoded, cannot crop")
	ErrIsDir      = errors.New("file is a directory")
	ErrIsEmpty    = errors.New("file is empty")
	ErrIsFile     = errors.New("directory path points to a file")
	ErrNoMatch    = errors.New("no match value is present")
	ErrPath       = errors.New("path is not permitted")
	ErrUnknownImg = errors.New("file is not a known image format")
	ErrValue      = errors.New("argument is empty")
	ErrVersion    = errors.New("application version mismatch")
)

// Dirs points to the download, preview, thumbnail, and extra directories.
type Dirs struct {
	Download  dir.Directory // Download is the directory path for the file downloads.
	Preview   dir.Directory // Preview is the directory path for the image previews.
	Thumbnail dir.Directory // Thumbnail is the directory path for the image thumbnails.
	Extra     dir.Directory // Extra is the directory path for the extra files.
}

// NOTE: For unrar on linux, the installation cannot use the unrar-free package,
// which is a poor substitute for the files this application needs to handle.
// The unrar binary should return:
// "UNRAR 6.24 freeware, Copyright (c) 1993-2023 Alexander Roshal".

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
	Unrar    = "unrar"    // Unrar is the rar decompression command.
	Unzip    = "unzip"    // Unzip is the zip decompression command.
	Zip7     = "7zz"      // Zip7 is the 7-Zip decompression command.
	ZipInfo  = "zipinfo"  // ZipInfo is the zip information command.
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
func CopyFile(sl *slog.Logger, src, dst string) error {
	const format = "copy file %s: %w"
	if err := nils.Check(sl); err != nil {
		return fmt.Errorf(format, "check", err)
	}

	s, err := os.Open(src)
	if err != nil {
		return fmt.Errorf(format, "open source", err)
	}
	defer func() { _ = s.Close() }()

	d, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf(format, "create dest", err)
	}

	defer func() {
		if cErr := d.Close(); cErr != nil && err == nil {
			err = fmt.Errorf(format, "close dest", cErr)
		}
	}()

	n, err := io.Copy(d, s)
	if err != nil {
		return fmt.Errorf(format, "copy bytes", err)
	}
	if err := d.Sync(); err != nil {
		return fmt.Errorf(format, "sync dest", err)
	}

	sl.Debug("file.copied", slog.String("path", dst), slog.Int64("bytes", n))
	return nil
}

// BaseName returns the base name of the file without the extension.
// Both the directory and extension are removed.
func BaseName(path string) string {
	if path == "" {
		return ""
	}
	base := filepath.Base(path)
	return strings.TrimSuffix(base, filepath.Ext(base))
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
	const format = "command lookup%s: %w"
	_, err := exec.LookPath(name)
	if errors.Is(err, exec.ErrDot) {
		err = nil
	}
	if err != nil {
		return fmt.Errorf(format, " path", err)
	}
	return nil
}

// LookVersion returns an error when the match string is not found in the named command output.
func LookVersion(name, flag, match string) error {
	const format = "command version lookup%s: %w"
	if err := LookCmd(name); err != nil {
		return fmt.Errorf(format, "", err)
	}
	if match == "" {
		return ErrNoMatch
	}
	cmd := exec.Command(name, flag) //nolint:noctx // legacy code, context not available in this function signature
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf(format, " stdout pipe", err)
	}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf(format, " start", err)
	}
	b, err := io.ReadAll(stdout)
	if err != nil {
		return fmt.Errorf(format, " read all", err)
	}
	if !bytes.Contains(b, []byte(match)) {
		return fmt.Errorf(format, " "+name, ErrVersion)
	}
	if err := cmd.Wait(); err != nil {
		return fmt.Errorf(format, " wait", err)
	}
	return nil
}

// Run looks for the command in the system path and executes it with the arguments.
// Any output to stderr is logged as a debug message.
func Run(ctx context.Context, sl *slog.Logger, name string, arg ...string) error {
	return run(ctx, sl, name, "", arg...)
}

// RunStdOut looks for the command in the system path and executes it with the arguments.
// Any output is sent to the stdout buffer.
func RunStdOut(name string, arg ...string) ([]byte, error) {
	const format = "command to stdout execute: %w"
	if err := LookCmd(name); err != nil {
		return nil, fmt.Errorf(format, err)
	}
	var out bytes.Buffer
	ctx, cancel := context.WithTimeout(context.Background(), CmdTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, name, arg...)
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf(format, err)
	}
	return out.Bytes(), nil
}

// RunQuiet looks for the command in the system path and executes it with the arguments.
func RunQuiet(ctx context.Context, name string, arg ...string) error {
	const format = "command to discard execute: %w"
	if err := nils.Check(ctx); err != nil {
		return fmt.Errorf(format, err)
	}
	if err := LookCmd(name); err != nil {
		return fmt.Errorf(format, err)
	}
	cmd := exec.CommandContext(ctx, name, arg...)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf(format, err)
	}
	return nil
}

// RunWorkdir looks for the command in the system path and executes it with the arguments.
// An optional working directory is set for the command.
// Any output to stderr is logged as a debug message.
func RunWorkdir(ctx context.Context, sl *slog.Logger, name, wdir string, arg ...string) error {
	return run(ctx, sl, name, wdir, arg...)
}

func run(ctx context.Context, sl *slog.Logger, name, wdir string, arg ...string) error {
	const msg = "command run"
	const format = msg + "%s: %w"
	if err := nils.Check(ctx, sl); err != nil {
		return fmt.Errorf(format, "", err)
	}
	if err := LookCmd(name); err != nil {
		return fmt.Errorf(format, "", err)
	}
	cmd := exec.CommandContext(ctx, name, arg...)
	cmd.Dir = wdir
	p, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf(format, " cannot start "+name, err)
	}
	sl.Debug(msg,
		slog.String("command.name", cmd.String()),
		slog.String("output", string(p)))
	return nil
}

// LockPath prevents directory traversal attacks by returning an error if
// any of the following are found in the path:
//
//   - forward slash
//   - back slash
//   - a pair of dots (..)
func LockPath(path string) error {
	if strings.ContainsAny(path, "/\\") || strings.Contains(path, "..") {
		return ErrPath
	}
	return nil
}
