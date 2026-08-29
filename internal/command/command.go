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
	"time"

	"github.com/Defacto2/server/internal/nils"
)

// go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out

const (
	CmdTimeout = 10 * time.Second

	gif  = ".gif"  // gif file extension
	jpg  = ".jpg"  // jpg file extension
	jpeg = ".jpeg" // jpeg file extension
	png  = ".png"  // png file extension
	webp = ".webp" // webp file extension
)

var (
	ErrAlign      = errors.New("command: invalid align choice")
	ErrCrop       = errors.New("command: invalid crop choice")
	ErrThumb      = errors.New("command: invalid thumb choice")
	ErrIsAnsi     = errors.New("command: text is ansi encoded, cannot crop")
	ErrIsEmpty    = errors.New("command: file is empty")
	ErrIsFile     = errors.New("command: directory path points to a file")
	ErrNoImages   = errors.New("command: no images found")
	ErrNoMatch    = errors.New("command: no match value is present")
	ErrUnknownImg = errors.New("command: file is not a known image format")
	ErrValue      = errors.New("command: argument is empty")
	ErrVersion    = errors.New("command: application version mismatch")
)

// NOTE: For unrar on linux, the installation cannot use the unrar-free package,
// which is a poor substitute for the files this application needs to handle.
// The unrar binary should return:
// "UNRAR 6.24 freeware, Copyright (c) 1993-2023 Alexander Roshal".

const (
	Arc      = "arc"      // Arc is the arc decompression command.
	Arj      = "arj"      // Arj is the arj decompression command.
	Ansilove = "ansilove" // Ansilove is the ansilove text to image command.
	Cwebp    = "cwebp"    // Cwebp is the Google create webp command.
	Gif2webp = "gif2webp" // Gif2webp is the Google gif to webp command.
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
var Lookups = [...]string{ //nolint:gochecknoglobals
	Arc,
	Arj,
	Ansilove,
	Cwebp,
	Gif2webp,
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

// Infos returns details for the list of the execute command names used by the application.
var Infos = [...]string{ //nolint:gochecknoglobals
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

// Lookup returns an error if the command is not found in the system path.
func Lookup(file string) (string, error) {
	const format = "command lookup: %w"
	if file == "" {
		return "", nil
	}

	s, err := exec.LookPath(file)
	if errors.Is(err, exec.ErrDot) {
		err = nil
	}
	if err != nil {
		return "", fmt.Errorf(format, err)
	}

	return s, nil
}

// LookupS returns an error when the match string is not found in the named command output.
func LookupS(ctx context.Context, name, flag, match string) error {
	const format = "command version lookup %s: %w"

	if match == "" {
		return ErrNoMatch
	}

	r := Runner{} //nolint:exhaustruct
	out, err := r.Run(ctx, name, flag)
	if err != nil {
		return fmt.Errorf(format, name, err)
	}

	if !bytes.Contains(out, []byte(match)) {
		return fmt.Errorf(format, name, ErrVersion)
	}

	return nil
}

// LookupUnrar returns an error if the name Alexander Roshal is not found in the unrar version output.
func LookupUnrar(ctx context.Context) error {
	return LookupS(ctx, Unrar, "", "Alexander Roshal")
}

type Runner struct {
	Timeout    time.Duration // Command timeout value, or leave empty to use 10 seconds
	Log        *slog.Logger  // Logger output or leave empty to use default
	WorkingDir string        // Working directory to execute the command or leave empty
}

// Run looks for the command in the system path and executes it with the arguments.
// The name of the command to run must be provided, and any options provided using arg.
func Run(ctx context.Context, sl *slog.Logger, name string, arg ...string) ([]byte, error) {
	r := Runner{
		Timeout:    0,
		Log:        sl,
		WorkingDir: "",
	}
	return r.Run(ctx, name, arg...)
}

// Run looks for the command in the system path and executes it with the arguments.
// The name of the command to run must be provided, and any options provided using arg.
func (r *Runner) Run(ctx context.Context, name string, arg ...string) ([]byte, error) {
	if r.Log == nil {
		r.Log = slog.Default()
	}

	if _, ok := ctx.Deadline(); !ok {
		if r.Timeout <= 0 {
			r.Timeout = CmdTimeout
		}
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, r.Timeout)
		defer cancel()
	}

	cmd := exec.CommandContext(ctx, name, arg...)
	if r.WorkingDir != "" {
		cmd.Dir = r.WorkingDir
	}
	out, err := cmd.CombinedOutput() // runs the command

	if r.Log.Enabled(ctx, slog.LevelDebug) {
		r.Log.Debug("command executed", slog.String("cmd", cmd.String()),
			slog.Int("exit_code", cmd.ProcessState.ExitCode()), slog.String("output", string(out)),
		)
	}

	if err != nil {
		const format = "command run %s: %w"
		if len(out) > 0 {
			return out, fmt.Errorf(format, name, err)
		}
		return out, fmt.Errorf(format, name, err)
	}

	return out, nil
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
