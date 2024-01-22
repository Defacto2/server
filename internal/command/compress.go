package command

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"go.uber.org/zap"
)

// ExtractOne extracts the named file from a zip archive.
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
func ExtractOne(z *zap.SugaredLogger, src, dst, ext, name string) error {
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

	r := runner{src: src, tmp: tmp, name: name}
	switch strings.ToLower(ext) {
	case arc:
		err = r.arc(z)
	case arj:
		err = r.arj(z)
	case rar:
		err = r.rar(z)
	case tar, gzip:
		err = r.tar(z)
	case zip:
		err = r.zip(z)
	default:
		err = r.p7zip(z)
	}
	if err != nil {
		return err
	}

	extracted := filepath.Join(tmp, r.name) // p7zip manipulates the r.name
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

type runner struct {
	src  string // src is the absolute path to the source archive.
	tmp  string // tmp is the absolute path to a temporary, destination directory.
	name string // name is the name of the file to extract from the archive.
}

// arc extracts the named file from the src arc archive.
func (r runner) arc(z *zap.SugaredLogger) error {
	// the arc command doesn't offer a target directory option
	tmpArc := filepath.Join(r.tmp, "archive.arc")
	if err := CopyFile(z, r.src, tmpArc); err != nil {
		return err
	}
	arg := []string{
		"xwo",  // Extract files from archive.
		tmpArc, // Source archive.
		r.name, // File to extract from the archive.
	}
	return RunWD(z, Arc, r.tmp, arg...)
}

// arj extracts the named file from the src arj archive.
func (r runner) arj(z *zap.SugaredLogger) error {
	// the arj command requires the source archive to have an .arj extension
	tmpArj := filepath.Join(r.tmp, "archive.arj")
	if err := CopyFile(z, r.src, tmpArj); err != nil {
		return err
	}
	arg := []string{
		"e",           // Extract files from archive.
		tmpArj,        // Source archive with the required .arj extension.
		r.name,        // File to extract from the archive.
		"-ht" + r.tmp, // Set Target directory, ie: "ht/destdir".
	}
	if err := Run(z, Arj, arg...); err != nil {
		s := arjExitStatus(err)
		z.Warnf("arj exit status: %s", s)
		return fmt.Errorf("%w: %s", err, s)
	}
	return nil
}

// p7zip extracts the named file from the src 7-Zip archive.
// The tool also supports the following archive formats:
// LZMA2, XZ, ZIP, Zip64, CAB, ARJ, GZIP, BZIP2, TAR, CPIO, RPM, ISO,
// most filesystem images and DEB formats.
func (r *runner) p7zip(z *zap.SugaredLogger) error {
	// p7zip may use incompatible forward slashes for Windows paths
	name := strings.ReplaceAll(r.name, "\\", "/")
	arg := []string{
		"e",          // Extract files from archive.
		"-y",         // Assume Yes on all queries.
		"-o" + r.tmp, // Set output directory.
		r.src,        // Source archive.
		name,         // File to extract from the archive.
	}
	if err := Run(z, P7zip, arg...); err != nil {
		return err
	}
	// handle file extraction from a directory in the archive
	r.name = filepath.Base(name)
	return Run(z, P7zip, arg...)
}

// rar extracts the named file from the src rar archive.
func (r *runner) rar(z *zap.SugaredLogger) error {
	// unrar <command> -<switch 1> -<switch N> <archive> <files...>
	arg := []string{
		"e",    // Extract files.
		r.src,  // Source archive.
		r.name, // File to extract from the archive.
	}
	if err := RunWD(z, Unrar, r.tmp, arg...); err != nil {
		s := unrarExitStatus(err)
		z.Warnf("unrar exit status: %s", s)
		return fmt.Errorf("%w: %s", err, s)
	}
	// handle file extraction from a directory in the archive
	r.name = filepath.Base(r.name)
	return nil
}

// tar extracts the named file from the src tar archive.
func (r runner) tar(z *zap.SugaredLogger) error {
	arg := []string{
		"-x",        // Extract files from archive.
		"-a",        // Auto detect archive type (for gzip support).
		"-f", r.src, // Source archive.
		"-C", r.tmp, // Target directory.
		r.name, // File to extract from the archive.
	}
	return Run(z, Tar, arg...)
}

// zip extracts the named file from the src zip archive.
func (r runner) zip(z *zap.SugaredLogger) error {
	arg := []string{r.src}         // source zip archive
	arg = append(arg, r.name)      // target file to extract
	arg = append(arg, "-d", r.tmp) // extract destination
	if err := Run(z, Unzip, arg...); err != nil {
		s := unzipExitStatus(err)
		z.Warnf("unzip exit status: %s", s)
		return fmt.Errorf("%w: %s", err, s)
	}
	return nil
}

// arjExitStatus returns the exit status of the arj command error.
func arjExitStatus(err error) string {
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
		if status, ok := exitError.Sys().(syscall.WaitStatus); ok {
			if s, ok := statuses[status.ExitStatus()]; ok {
				return s
			}
		}
	}
	return err.Error()
}

// unrarExitStatus returns the exit status of the unrar command error.
func unrarExitStatus(err error) string {
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
		if status, ok := exitError.Sys().(syscall.WaitStatus); ok {
			if s, ok := statuses[status.ExitStatus()]; ok {
				return s
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
		if status, ok := exitError.Sys().(syscall.WaitStatus); ok {
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
			if s, ok := statuses[status.ExitStatus()]; ok {
				return s
			}
		}
	}
	return err.Error()
}

// extract extracts the named file from a zip archive and returns the path to the file.
func extract(z *zap.SugaredLogger, src, ext, name string) (string, error) {
	if z == nil {
		return "", ErrZap
	}
	tmp, err := os.MkdirTemp(os.TempDir(), pattern)
	if err != nil {
		return "", err
	}

	dst := filepath.Join(tmp, filepath.Base(name))
	if err = ExtractOne(z, src, dst, ext, name); err != nil {
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
func (dir Dirs) ExtractImage(z *zap.SugaredLogger, src, uuid, ext, name string) error {
	if z == nil {
		return ErrZap
	}

	dst, err := extract(z, src, ext, name)
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
		// err = dir.LosslessScreenshot(z, dst, uuid)
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
func (dir Dirs) ExtractAnsiLove(z *zap.SugaredLogger, src, uuid, ext, name string) error {
	if z == nil {
		return ErrZap
	}

	dst, err := extract(z, src, ext, name)
	if err != nil {
		return err
	}
	defer os.RemoveAll(dst)
	if err := dir.AnsiLove(z, dst, uuid); err != nil {
		return err
	}
	return nil
}
