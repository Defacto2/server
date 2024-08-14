package command

// Package file compress.go contains the functions for extracting files from archives
// using external command-line tools like 7-Zip, Unrar, Unzip, etc.

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/Defacto2/server/internal/helper"
	"go.uber.org/zap"
)

// ExtractOne extracts the named file from the src archive.
//
// The extracted file is copied to the dst. It uses [exec.Command] and
// relies on extractor being available on the system host path.
// Using non-Go apps allows for better compatibility with retro zip archives,
// such as those that use the [compression methods] prior to zip deflate.
//
// The src argument is the path to the zip archive.
// The dst argument is the destination filepath and should end with
// a file extension, eg. ".txt".
// The optional extHint arg is a file extension hint for the extractor.
// Valid hints are: ".arc", ".arj", ".rar", ".tar", ".zip", otherwise the
// extractor will use the 7-Zip command.
// The name argument is the name of the one file to unzip and copy.
//
// [compression methods]: https://www.hanshq.net/zip.html
//
// [unzip]: https://sourceforge.net/projects/infozip
func ExtractOne(debug *zap.SugaredLogger, src, dst, extHint, name string) error {
	st, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("extract one stat %w", err)
	}
	if st.IsDir() {
		return fmt.Errorf("extract one %w: %q", ErrIsDir, src)
	}
	if st.Size() == 0 {
		return fmt.Errorf("extract one %w: %q", ErrEmpty, src)
	}

	tmp, err := os.MkdirTemp(helper.TmpDir(), pattern)
	if err != nil {
		return fmt.Errorf("extract one make temp dir %w", err)
	}
	defer os.RemoveAll(tmp)
	r := runner{src: src, tmp: tmp, name: name, debug: debug}
	if err = r.extract(extHint); err != nil {
		return fmt.Errorf("extract one extraction %w", err)
	}
	extracted := filepath.Join(tmp, r.name)
	st, err = os.Stat(extracted)
	if err != nil {
		return fmt.Errorf("extract one extracted stat %w", err)
	}
	if st.IsDir() {
		return fmt.Errorf("extract one extracted %w: %q", ErrIsDir, extracted)
	}
	if st.Size() == 0 {
		return fmt.Errorf("extract one extracted %w: %q", ErrEmpty, extracted)
	}
	if err := CopyFile(debug, extracted, dst); err != nil {
		return fmt.Errorf("extract one copy extracted %w", err)
	}
	return nil
}

type runner struct {
	debug *zap.SugaredLogger
	src   string // src is the absolute path to the source archive.
	tmp   string // tmp is the absolute path to a temporary, destination directory.
	name  string // name is the name of the file to extract from the archive.
}

// extract extracts the named file from the src archive.
func (r runner) extract(ext string) error {
	switch strings.ToLower(ext) {
	case arc:
		return r.arc()
	case arj:
		return r.arj()
	case rar:
		return r.rar()
	case tar, gzip:
		return r.tar()
	case zip:
		return r.zip()
	default:
		return r.zip7()
	}
}

// arc extracts the named file from the src arc archive.
func (r runner) arc() error {
	// the arc command doesn't offer a target directory option
	tmpArc := filepath.Join(r.tmp, "archive.arc")
	if err := CopyFile(r.debug, r.src, tmpArc); err != nil {
		return fmt.Errorf("decompress arc copy file %w", err)
	}
	arg := []string{
		"xwo",  // Extract files from archive.
		tmpArc, // Source archive.
		r.name, // File to extract from the archive.
	}
	return RunWD(r.debug, Arc, r.tmp, arg...)
}

// arj extracts the named file from the src arj archive.
func (r runner) arj() error {
	// the arj command requires the source archive to have an .arj extension
	tmpArj := filepath.Join(r.tmp, "archive.arj")
	if err := CopyFile(r.debug, r.src, tmpArj); err != nil {
		return fmt.Errorf("decompress arj copy file %w", err)
	}
	arg := []string{
		"e",           // Extract files from archive.
		tmpArj,        // Source archive with the required .arj extension.
		r.name,        // File to extract from the archive.
		"-ht" + r.tmp, // Set Target directory, ie: "ht/destdir".
	}
	if err := Run(r.debug, Arj, arg...); err != nil {
		s := ArjExitStatus(err)
		r.debug.Warnf("arj exit status: %s", s)
		return fmt.Errorf("decompress arj %w: %s", err, s)
	}
	return nil
}

// zip7 extracts the named file from the src 7-Zip archive.
// The tool also supports the following archive formats:
// LZMA2, XZ, ZIP, Zip64, CAB, ARJ, GZIP, BZIP2, TAR, CPIO, RPM, ISO,
// most filesystem images and DEB formats.
func (r *runner) zip7() error {
	// zip7 may use incompatible forward slashes for Windows paths
	name := strings.ReplaceAll(r.name, "\\", "/")
	arg := []string{
		"e",          // Extract files from archive.
		"-y",         // Assume Yes on all queries.
		"-o" + r.tmp, // Set output directory.
		r.src,        // Source archive.
		name,         // File to extract from the archive.
	}
	if err := Run(r.debug, Zip7, arg...); err != nil {
		return fmt.Errorf("decompress 7z run %w", err)
	}
	// handle file extraction from a directory in the archive
	r.name = filepath.Base(name)
	if err := Run(r.debug, Zip7, arg...); err != nil {
		return fmt.Errorf("decompress 7z subdirectory run %w, %s", err, r.name)
	}
	return nil
}

// rar extracts the named file from the src rar archive.
func (r *runner) rar() error {
	// unrar <command> -<switch 1> -<switch N> <archive> <files...>
	arg := []string{
		"e",    // Extract files.
		r.src,  // Source archive.
		r.name, // File to extract from the archive.
	}
	if err := RunWD(r.debug, Unrar, r.tmp, arg...); err != nil {
		s := UnRarExitStatus(err)
		r.debug.Warnf("unrar exit status: %s", s)
		return fmt.Errorf("decompress rar run %w: %s", err, s)
	}
	// handle file extraction from a directory in the archive
	r.name = filepath.Base(r.name)
	return nil
}

// tar extracts the named file from the src tar archive.
func (r runner) tar() error {
	arg := []string{
		"-x",        // Extract files from archive.
		"-a",        // Auto detect archive type (for gzip support).
		"-f", r.src, // Source archive.
		"-C", r.tmp, // Target directory.
		r.name, // File to extract from the archive.
	}
	return Run(r.debug, Tar, arg...)
}

// zip extracts the named file from the src zip archive.
func (r runner) zip() error {
	arg := []string{r.src}         // source zip archive
	arg = append(arg, r.name)      // target file to extract
	arg = append(arg, "-d", r.tmp) // extract destination
	if err := Run(r.debug, Unzip, arg...); err != nil {
		s := unzipExitStatus(err)
		r.debug.Warnf("unzip exit status: %s", s)
		return fmt.Errorf("decompress zip run %w: %s", err, s)
	}
	return nil
}

// ArjExitStatus returns the exit status of the arj command error.
func ArjExitStatus(err error) string {
	if err == nil {
		return ""
	}
	statuses := map[int]string{
		0:  "success",
		1:  "warning",
		2:  "fatal error",
		3:  "crc error (header, file or bad password)",
		4:  "arj-security error",
		5:  "disk full or write error",
		6:  "cannot open archive or file",
		7:  "user error, bad command line parameters",
		8:  "not enough memory",
		9:  "not an arj archive",
		10: "MS-DOS XMS memory error",
		11: "user control break",
		12: "too many chapters (over 250)",
	}
	var exitError *exec.ExitError
	if errors.As(err, &exitError) {
		if waitStatus, statusExists := exitError.Sys().(syscall.WaitStatus); statusExists {
			if status, exitStatus := statuses[waitStatus.ExitStatus()]; exitStatus {
				return status
			}
		}
	}
	return err.Error()
}

// UnRarExitStatus returns the exit status of the unrar command error.
func UnRarExitStatus(err error) string {
	if err == nil {
		return ""
	}
	statuses := map[int]string{
		0:   "success",
		1:   "success with warning",
		2:   "fatal error",
		3:   "invalid checksum, data damage",
		4:   "attempt to modify a locked archive",
		5:   "write error",
		6:   "file open error",
		7:   "wrong command line option",
		8:   "not enough memory",
		9:   "file create error",
		10:  "no files matching the specified mask and options were found",
		11:  "incorrect password",
		255: "user stopped the process with control-C",
	}
	var exitError *exec.ExitError
	if errors.As(err, &exitError) {
		if waitStatus, statusExists := exitError.Sys().(syscall.WaitStatus); statusExists {
			if status, exitStatus := statuses[waitStatus.ExitStatus()]; exitStatus {
				return status
			}
		}
	}
	return err.Error()
}

// unzipExitStatus returns the exit status of the unzip command error.
func unzipExitStatus(err error) string {
	if err == nil {
		return ""
	}
	var exitError *exec.ExitError
	if errors.As(err, &exitError) {
		if waitStatus, statusExists := exitError.Sys().(syscall.WaitStatus); statusExists {
			statuses := map[int]string{
				0:  "success",
				1:  "success with warning",
				2:  "generic error in the zipfile format",
				3:  "severe error in zipfile format",
				4:  "unable to allocate memory for buffers",
				5:  "unable to allocate memory or tty to read decryption password",
				6:  "unable to allocate memory during decompression to disk",
				7:  "unable to allocate memory during in-memory decompression",
				8:  "unused",
				9:  "the specified zip file was not found",
				10: "invalid command arguments",
				11: "no matching files were found",
				12: "possible zip-bomb detected, aborting",
				50: "the disk is full during extraction",
				51: "the end of the zip archive was encountered prematurely",
				80: "user stopped the process with control-C",
				81: "testing or extraction of one or more files failed due to " +
					"unsupported compression methods or unsupported decryption",
				82: "no files were found due to bad decryption password",
			}
			if status, exitStatus := statuses[waitStatus.ExitStatus()]; exitStatus {
				return status
			}
		}
	}
	return err.Error()
}

// ExtractAnsiLove extracts the named text file from a zip archive.
// The text file is converted to a PNG preview and a webp thumbnails.
// Any text file usable by the ansilove command is supported,
// including ANSI, codepage plain text, PCBoard, etc.
func (dir Dirs) ExtractAnsiLove(debug *zap.SugaredLogger, src, extHint, unid, name string) error {
	dst, err := extract(debug, src, extHint, name)
	if err != nil {
		return fmt.Errorf("decompress extract ansilove: %w", err)
	}
	defer os.RemoveAll(dst)
	return dir.TextImager(debug, dst, unid)
}

// ExtractImage extracts the named image file from a zip archive.
// Based on the file extension, the image is converted to a webp preview and thumbnails.
// Named files with a PNG extension are optimized but kept as the preview image.
func (dir Dirs) ExtractImage(debug *zap.SugaredLogger, src, extHint, unid, name string) error {
	dst, err := extract(debug, src, extHint, name)
	if err != nil {
		return fmt.Errorf("decompress extract image: %w", err)
	}
	defer os.RemoveAll(dst)
	return dir.PictureImager(debug, dst, unid)
}

// extract extracts the named file from a zip archive and returns the path to the file.
func extract(debug *zap.SugaredLogger, src, extHint, name string) (string, error) {
	tmp, err := os.MkdirTemp(helper.TmpDir(), pattern)
	if err != nil {
		return "", fmt.Errorf("decompress make dir temp %w", err)
	}

	dst := filepath.Join(tmp, filepath.Base(name))
	if err = ExtractOne(debug, src, dst, extHint, name); err != nil {
		return "", fmt.Errorf("decompress extract one %w", err)
	}
	st, err := os.Stat(dst)
	if err != nil {
		return "", fmt.Errorf("decompress extract stat %w", err)
	}
	if st.IsDir() {
		return "", fmt.Errorf("decompress extract %w: %q", ErrIsDir, dst)
	}
	return dst, nil
}
