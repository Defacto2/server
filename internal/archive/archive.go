// Package archive provides compressed and stored archive file extraction and content listing.
//
// The file archive formats supported are ARJ, LHA, LZH, RAR, TAR, and ZIP.
//
// The package uses the [mholt/archiver/v3] package and the following Linux programs
// as a fallback for legacy file support.
//
//  1. [arj] - open-source ARJ v3.10
//  2. [lha] - Lhasa v0.4 LHA tool found in the jlha-utils or lhasa packages
//  3. [unrar] - 6.24 freeware by Alexander Roshal, not the common [unrar-free] which is feature incomplete
//  4. [zipinfo] - ZipInfo v3 by the Info-ZIP workgroup
//
// [mholt/archiver/v3]: https://github.com/mholt/archiver/tree/v3.5.1
// [arj]: https://arj.sourceforge.net/
// [lha]: https://fragglet.github.io/lhasa/
// [unrar]: https://www.rarlab.com/rar_add.htm
// [unrar-free]: https://gitlab.com/bgermann/unrar-free
// [zipinfo]: https://infozip.sourceforge.net/
package archive

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/Defacto2/server/internal/archive/internal"
	"github.com/Defacto2/server/internal/archive/pkzip"
	"github.com/Defacto2/server/internal/command"
	"github.com/Defacto2/server/internal/helper"
	"github.com/Defacto2/server/internal/magicnumber"
	"github.com/mholt/archiver/v3"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

const (
	// permitted archives on the site:
	// 7z,arc,ark,arj,cab,gz,lha,lzh,rar,tar,tar.gz,zip.
	arjx = ".arj" // Archived by Robert Jung
	lhax = ".lha" // LHarc by Haruyasu Yoshizaki (Yoshi)
	lhzx = ".lzh" // LHArc by Haruyasu Yoshizaki (Yoshi)
	rarx = ".rar" // Roshal ARchive by Alexander Roshal
	zipx = ".zip" // Phil Katz's ZIP for MS-DOS systems
)

var (
	ErrDest           = errors.New("destination is empty")
	ErrExt            = errors.New("extension is not a supported archive format")
	ErrNotArchive     = errors.New("file is not an archive")
	ErrNotImplemented = errors.New("archive format is not implemented")
	ErrRead           = errors.New("could not read the file archive")
	ErrProg           = errors.New("program error")
	ErrFile           = errors.New("path is a directory")
	ErrPath           = errors.New("path is a file")
	ErrPanic          = errors.New("extract panic")
	ErrMissing        = errors.New("path does not exist")
)

// CheckyPath checks the byte slice for valid UTF-8 encoding.
// If the byte slice is not valid, it will attempt to decode
// the byte slice using the MS-DOS, [charmap.CodePage437] character set.
//
// Needed for historical oddities found in BBS file archives, the
// file and folders were sometimes named in [leetspeak] using untypable
// characters and symbols. For example the valid filename ¿ædmé.ñôw could not be
// easily typed out on a standard North American keyboard in MS-DOS.
//
// [leetspeak]: https://www.oed.com/dictionary/leetspeak_n
func CheckyPath(p []byte) string {
	if utf8.Valid(p) {
		return string(p)
	}
	r := transform.NewReader(bytes.NewReader(p), charmap.CodePage437.NewDecoder())
	result, err := io.ReadAll(r)
	if err != nil {
		return ""
	}
	return string(result)
}

// List returns the files within an rar, tar, lha, or zip archive.
// This filename extension is used to determine the archive format.
func List(src, filename string) ([]string, error) {
	st, err := os.Stat(src)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, fmt.Errorf("archive list %w: %s", ErrMissing, filepath.Base(src))
	}
	if st.IsDir() {
		return nil, fmt.Errorf("archive list %w: %s", ErrFile, filepath.Base(src))
	}
	files, err := walker(src, filename)
	if err != nil {
		return commander(src, filename)
	}
	return files, nil
}

// walker uses the mholt/archiver package to walk the src archive file.
func walker(src, filename string) ([]string, error) {
	name := strings.ToLower(filename) // ByExtension is case sensitive
	format, err := archiver.ByExtension(name)
	if err != nil {
		return nil, fmt.Errorf("walker by extension %w", err)
	}
	w, walkerExists := format.(archiver.Walker)
	if !walkerExists {
		return nil, fmt.Errorf("walker %w, %q", ErrExt, filename)
	}
	files := []string{}
	err = w.Walk(src, func(f archiver.File) error {
		if f.IsDir() {
			return nil
		}
		if strings.TrimSpace(f.Name()) == "" {
			return nil
		}
		name := CheckyPath([]byte(f.Name()))
		files = append(files, name)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walker %w", err)
	}
	return files, nil
}

// commander uses system archiver and decompression programs to read the src archive file.
func commander(src, filename string) ([]string, error) {
	c := Content{}
	if err := c.Read(src); err != nil {
		return nil, fmt.Errorf("commander failed with %s (%q): %w", filename, c.Ext, err)
	}
	// remove empty entries
	files := c.Files
	files = slices.DeleteFunc(files, func(s string) bool {
		return strings.TrimSpace(s) == ""
	})
	return files, nil
}

// Extract the filename targets from the source archive file to the destination folder.
// If no targets are provided, all files are extracted.
// The filename extension is used to determine the archive format.
func Extract(src, dst, filename string, targets ...string) error {
	name := strings.ToLower(filename)
	f, err := archiver.ByExtension(name)
	if err != nil {
		return extractor(src, dst, targets...)
	}
	// recover from panic caused by mholt/archiver.
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("archive extract %w %s: %v", ErrPanic, name, r)
		}
	}()
	extractAll := len(targets) == 0
	if extractAll {
		all, unarchiverExists := f.(archiver.Unarchiver)
		if !unarchiverExists {
			return fmt.Errorf("archive extract %w, %q", ErrExt, filename)
		}
		if err = all.Unarchive(src, dst); err == nil {
			return nil
		}
		return extractor(src, dst, targets...)
	}
	target, extractorExists := f.(archiver.Extractor)
	if !extractorExists {
		return fmt.Errorf("archive extract %w, %q", ErrExt, filename)
	}
	t := strings.Join(targets, " ")
	if err = target.Extract(src, t, dst); err == nil {
		return nil
	}
	return extractor(src, dst, targets...)
}

// extractor second attempt at extraction using a system archiver program.
func extractor(src, dst string, targets ...string) error {
	x := Extractor{Source: src, Destination: dst}
	err := x.Extract(targets...)
	if err != nil {
		return fmt.Errorf("command extract: %w", err)
	}
	return nil
}

// MagicExt uses the Linux [file] program to determine the src archive file type.
// The returned string will be a file separator and extension.
// For example a file with the magic string "gzip compressed data" will return ".tar.gz".
//
// Note both bzip2 and gzip archives return the .tar extension prefix.
//
// [file]: https://www.darwinsys.com/file/
func MagicExt(src string) (string, error) {
	prog, err := exec.LookPath("file")
	if err != nil {
		return "", fmt.Errorf("archive magic file lookup %w", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cmd := exec.CommandContext(ctx, prog, "--brief", src)
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("archive magic file command %w", err)
	}
	if len(out) == 0 {
		return "", fmt.Errorf("archive magic file type: %w", ErrRead)
	}
	magics := map[string]string{
		"7-zip archive data":    ".7z",
		"arj archive data":      arjx,
		"bzip2 compressed data": ".tar.bz2",
		"gzip compressed data":  ".tar.gz",
		"rar archive data":      ".rar",
		"posix tar archive":     ".tar",
		"zip archive data":      zipx,
	}
	s := strings.Split(strings.ToLower(string(out)), ",")
	magic := strings.TrimSpace(s[0])
	if internal.MagicLHA(magic) {
		return lhax, nil
	}
	for magic, ext := range magics {
		if strings.TrimSpace(s[0]) == magic {
			return ext, nil
		}
	}
	return "", fmt.Errorf("archive magic file %w: %q", ErrExt, magic)
}

// Replace the filename file extension with the ext string.
// Leaving ext empty returns the filename without a file extension.
func Replace(ext, filename string) string {
	const sep = "."
	s := strings.Split(filename, sep)
	if ext == "" && len(s) == 1 {
		return filename
	}
	if ext == "" {
		return strings.Join(s[:len(s)-1], sep)
	}
	if len(s) == 1 {
		s = append(s, ".tmp")
	}
	s[len(s)-1] = strings.Join(strings.Split(ext, sep), "")
	return strings.Join(s, sep)
}

// Content are the result of using system programs to read the file archives.
type Content struct {
	Ext   string   // Ext returns file extension of the archive.
	Files []string // Files returns list of files within the archive.
}

// ARJ returns the content of the src ARJ archive,
// credited to Robert Jung, using the [arj program].
//
// [arj program]: https://arj.sourceforge.net/
func (c *Content) ARJ(src string) error {
	prog, err := exec.LookPath(command.Arj)
	if err != nil {
		return fmt.Errorf("archive arj reader %w", err)
	}

	const verboselist = "v"
	var b bytes.Buffer
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cmd := exec.CommandContext(ctx, prog, verboselist, src)
	cmd.Stderr = &b
	out, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("archive arj output %w", err)
	}
	if len(out) == 0 {
		return ErrRead
	}
	outs := strings.Split(string(out), "\n")
	files := []string{}
	const start = len("001) ")
	for _, s := range outs {
		if !internal.ARJItem(s) {
			continue
		}
		files = append(files, s[start:])
	}
	c.Files = slices.DeleteFunc(files, func(s string) bool {
		return strings.TrimSpace(s) == ""
	})
	c.Ext = arjx
	return nil
}

// LHA returns the content of the src LHA or LZH archive,
// credited to Haruyasu Yoshizaki (Yoshi), using the [lha program].
//
// [lha program]: https://fragglet.github.io/lhasa/
func (c *Content) LHA(src string) error {
	prog, err := exec.LookPath(command.Lha)
	if err != nil {
		return fmt.Errorf("archive lha reader %w", err)
	}

	const list = "-l"
	var b bytes.Buffer
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cmd := exec.CommandContext(ctx, prog, list, src)
	cmd.Stderr = &b
	out, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("archive lha output %w", err)
	}
	if len(out) == 0 {
		return ErrRead
	}
	outs := strings.Split(string(out), "\n")

	// LHA list command outputs with a MSDOS era, fixed-width layout table
	const (
		sizeS = len("[generic]              ")
		sizeL = len("-------")
		start = len("[generic]                   12 100.0% Apr 10 17:03 ")
		dir   = 0
	)

	files := []string{}
	for _, s := range outs {
		if len(s) < start {
			continue
		}
		size := strings.TrimSpace(s[sizeS : sizeS+sizeL])
		if i, err := strconv.Atoi(size); err != nil {
			continue
		} else if i == dir {
			continue
		}
		files = append(files, s[start:])
	}
	c.Files = slices.DeleteFunc(files, func(s string) bool {
		return strings.TrimSpace(s) == ""
	})
	c.Ext = lhax
	return nil
}

// Rar returns the content of the src RAR archive, credited to Alexander Roshal,
// using the [unrar program].
//
// [unrar program]: https://www.rarlab.com/rar_add.htm
func (c *Content) Rar(src string) error {
	prog, err := exec.LookPath(command.Unrar)
	if err != nil {
		return fmt.Errorf("archive unrar reader %w", err)
	}
	const (
		listBrief  = "lb"
		noComments = "-c-"
	)
	var b bytes.Buffer
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cmd := exec.CommandContext(ctx, prog, listBrief, "-ep", noComments, src)
	cmd.Stderr = &b
	out, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("archive unrar output %w: %s", err, src)
	}
	if len(out) == 0 {
		return ErrRead
	}
	c.Files = strings.Split(string(out), "\n")
	c.Files = slices.DeleteFunc(c.Files, func(s string) bool {
		return strings.TrimSpace(s) == ""
	})
	c.Ext = rarx
	return nil
}

// Read returns the content of the src file archive using the system archiver programs.
// The filename is used to determine the archive format.
// Supported formats are ARJ, LHA, LZH, RAR, and ZIP.
func (c *Content) Read(src string) error {
	ext, err := MagicExt(src)
	if err != nil {
		return fmt.Errorf("read %w", err)
	}
	// if !strings.EqualFold(ext, filepath.Ext(filename)) {
	// 	// retry using correct filename extension
	// 	return fmt.Errorf("system reader: %w", ErrWrongExt)
	// }
	switch strings.ToLower(ext) {
	case arjx:
		return c.ARJ(src)
	case lhax, lhzx:
		return c.LHA(src)
	case rarx:
		return c.Rar(src)
	case zipx:
		return c.Zip(src)
	}
	return fmt.Errorf("read %w", ErrRead)
}

// Zip returns the content of the src ZIP archive, credited to Phil Katz,
// using the [zipinfo program].
//
// [zipinfo program]: https://infozip.sourceforge.net/
func (c *Content) Zip(src string) error {
	prog, err := exec.LookPath(command.ZipInfo)
	if err != nil {
		return fmt.Errorf("archive zipinfo reader %w", err)
	}
	const list = "-1"
	var b bytes.Buffer
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cmd := exec.CommandContext(ctx, prog, list, src)
	cmd.Stderr = &b
	out, err := cmd.Output()
	if err != nil {
		// handle broken zips that still contain some valid files
		if b.String() != "" && len(out) > 0 {
			// return files, zipx, nil
			return nil
		}
		// otherwise the zipinfo threw an error
		return fmt.Errorf("archive zipinfo %w: %s", err, src)
	}
	if len(out) == 0 {
		return ErrRead
	}
	c.Files = strings.Split(string(out), "\n")
	c.Files = slices.DeleteFunc(c.Files, func(s string) bool {
		return strings.TrimSpace(s) == ""
	})
	c.Ext = zipx
	return nil
}

// ExtractAll extracts all files from the src archive file to the destination directory.
func ExtractAll(src, dst string) error {
	e := Extractor{Source: src, Destination: dst}
	if err := e.Extract(); err != nil {
		return fmt.Errorf("extract all %w", err)
	}
	return nil
}

// Extractor uses system archiver programs to extract the targets from the src file archive.
type Extractor struct {
	Source      string // The source archive file.
	Destination string // The extraction destination directory.
}

// Extract the targets from the source file archive
// to the destination directory a system archive program.
// If the targets are empty then all files are extracted.
//
// The required Filename string is used to determine the archive format.
//
// Some archive formats that could be impelmented if needed in the future:
// freearc, zoo
func (x Extractor) Extract(targets ...string) error {
	r, err := os.Open(x.Source)
	if err != nil {
		return err
	}
	defer r.Close()
	sign, err := magicnumber.Archive(r)
	if err != nil {
		return err
	}
	switch sign {
	case
		magicnumber.Bzip2CompressArchive,
		magicnumber.GzipCompressArchive,
		magicnumber.MicrosoftCABinet,
		magicnumber.TapeARchive,
		magicnumber.XZCompressArchive,
		magicnumber.ZStandardArchive:
		return x.Bsdtar(targets...)
	case
		magicnumber.PKWAREZip,
		magicnumber.PKWAREZip64,
		magicnumber.PKWAREZipShrink,
		magicnumber.PKWAREZipReduce,
		magicnumber.PKWAREZipImplode:
		return x.extractZip(targets...)
	case
		magicnumber.PKLITE,
		magicnumber.PKSFX,
		magicnumber.PKWAREMultiVolume:
		return fmt.Errorf("%w, %s", ErrNotImplemented, sign)
	case magicnumber.ARChiveSEA:
		return x.ARC(targets...)
	case magicnumber.ArchiveRobertJung:
		return x.ARJ(targets...)
	case magicnumber.YoshiLHA:
		return x.LHA(targets...)
	case magicnumber.RoshalARchive,
		magicnumber.RoshalARchivev5:
		return x.Rar(targets...)
	case magicnumber.X7zCompressArchive:
		return x.Zip7(targets...)
	case magicnumber.Unknown:
		return fmt.Errorf("%w, %s", ErrNotArchive, sign)
	default:
		return fmt.Errorf("%w, %s", ErrNotImplemented, sign)
	}
}

// ExtractZip delegates the extraction of the source archive to the correct program
// based on its compression method and the original operating system used to create it.
// As some valid filenames set by MS-DOS codepages are not valid UTF-8 filenames.
func (x Extractor) extractZip(targets ...string) error {
	if deflate, _ := pkzip.Zip(x.Source); !deflate {
		return x.ZipHW(targets...)
	}
	if err := x.Zip(targets...); err != nil {
		return x.Bsdtar(targets...)
	}
	return nil
}

// Bsdtar extracts the targets from the source archive
// to the destination directory using the [bsdtar program].
// If the targets are empty then all files are extracted.
// bsdtar uses the performant [libarchive library] for archive extraction
// and is the recommended program for extracting the following formats:
//
// gzip, bzip2, compress, xz, lzip, lzma, tar, iso9660, zip, ar, xar,
// lha/lzh, rar, rar v5, Microsoft Cabinet, 7-zip.
//
// [bsdtar program]: https://man.freebsd.org/cgi/man.cgi?query=bsdtar&sektion=1&format=html
// [libarchive library]: http://www.libarchive.org/
func (x Extractor) Bsdtar(targets ...string) error {
	src, dst := x.Source, x.Destination
	prog, err := exec.LookPath("bsdtar")
	if err != nil {
		return fmt.Errorf("archive tar extract %w", err)
	}
	if dst == "" {
		return ErrDest
	}
	var b bytes.Buffer
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	const (
		extract   = "-x"                    // -x extract files
		source    = "--file"                // -f file path to extract
		targetDir = "--cd"                  // -C target directory
		noAcls    = "--no-acls"             // --no-acls
		noFlags   = "--no-fflags"           // --no-fflags
		noModTime = "--modification-time"   // --modification-time
		noSafeW   = "--no-safe-writes"      // --no-safe-writes
		noOwner   = "--no-same-owner"       // --no-same-owner
		noPerms   = "--no-same-permissions" // --no-same-permissions
		noXattrs  = "--no-xattrs"           // --no-xattrs
	)
	args := []string{extract, source, src}
	args = append(args, noAcls, noFlags, noSafeW, noModTime, noOwner, noPerms, noXattrs)
	args = append(args, targetDir, dst)
	args = append(args, targets...)
	cmd := exec.CommandContext(ctx, prog, args...)
	cmd.Stderr = &b
	if err = cmd.Run(); err != nil {
		if b.String() != "" {
			return fmt.Errorf("archive tar %w: %s: %s", ErrProg, prog, strings.TrimSpace(b.String()))
		}
		return fmt.Errorf("archive tar %w: %s", err, prog)
	}
	return nil
}

// ARC extracts the targets from the source ARC archive
// to the destination directory using the [arc program].
// If the targets are empty then all files are extracted.
//
// ARC is a DOS era archive format that is not widely supported.
// It also does not support extracting to a target directory.
// To work around this, this copies the source archive
// to the destination directory, uses that as the working directory
// and extracts the files. The copied source archive is then removed.
//
// [arc program]: https://arj.sourceforge.net/
func (x Extractor) ARC(targets ...string) error {
	src, dst := x.Source, x.Destination
	if st, err := os.Stat(dst); err != nil {
		return fmt.Errorf("%w: %s", err, dst)
	} else if !st.IsDir() {
		return fmt.Errorf("%w: %s", ErrPath, dst)
	}
	prog, err := exec.LookPath(command.Arc)
	if err != nil {
		return fmt.Errorf("archive arc extract %w", err)
	}

	srcInDst := filepath.Join(dst, filepath.Base(src))
	if _, err := helper.Duplicate(src, srcInDst); err != nil {
		return fmt.Errorf("archive arc duplicate %w", err)
	}
	defer os.Remove(srcInDst)

	var b bytes.Buffer
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	const (
		extract = "x" // x extract files
	)
	args := []string{extract, filepath.Base(src)}
	args = append(args, targets...)
	cmd := exec.CommandContext(ctx, prog, args...)
	cmd.Dir = dst
	cmd.Stderr = &b
	if err = cmd.Run(); err != nil {
		if b.String() != "" {
			return fmt.Errorf("archive arc %w: %s: %q",
				ErrProg, prog, strings.TrimSpace(b.String()))
		}
		return fmt.Errorf("archive arc %w: %s", err, prog)
	}
	return nil
}

// ARJ extracts the targets from the source ARJ archive
// to the destination directory using the [arj program].
// If the targets are empty then all files are extracted.
//
// [arj program]: https://arj.sourceforge.net/
func (x Extractor) ARJ(targets ...string) error {
	src, dst := x.Source, x.Destination
	if st, err := os.Stat(dst); err != nil {
		return fmt.Errorf("%w: %s", err, dst)
	} else if !st.IsDir() {
		return fmt.Errorf("%w: %s", ErrPath, dst)
	}
	// note: only use arj, as unarj offers limited functionality
	prog, err := exec.LookPath(command.Arj)
	if err != nil {
		return fmt.Errorf("archive arj extract %w", err)
	}

	// arj REQUIRES a file extension for the source archive
	srcWithExt := src + arjx
	if err := os.Symlink(src, srcWithExt); err != nil {
		defer os.Remove(srcWithExt)
		return fmt.Errorf("archive arj symlink %w", err)
	}
	defer os.Remove(srcWithExt)

	var b bytes.Buffer
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// example command: arj x archive destdir/ *
	const (
		extract   = "x"   // x extract files
		rmPaths   = "r"   // r remove paths
		yes       = "-y"  // -y assume yes to all queries
		noProg    = "-i"  // -i do not display progress
		excBase   = "-e1" // -e exclude base directory
		noExt     = "-hx" // -hx default extension
		targetDir = "-ht" // -ht target directory
		dosCompat = "-2d" // -2d DOS compatibility mode
	)
	args := []string{extract, srcWithExt}
	args = append(args, targets...)
	args = append(args, targetDir+dst)
	fmt.Println("arj", args)
	cmd := exec.CommandContext(ctx, prog, args...)
	cmd.Stderr = &b
	if err = cmd.Run(); err != nil {
		if b.String() != "" {
			return fmt.Errorf("archive arj %w: %s: %q",
				ErrProg, prog, strings.TrimSpace(b.String()))
		}
		return fmt.Errorf("archive arj %w: %s", err, prog)
	}
	return nil
}

// LHA extracts the targets from the source LHA/LZH archive
// to the destination directory using an lha program.
// If the targets are empty then all files are extracted.
//
// On Linux either the jlha-utils or lhasa work.
func (x Extractor) LHA(targets ...string) error {
	src, dst := x.Source, x.Destination
	prog, err := exec.LookPath(command.Lha)
	if err != nil {
		return fmt.Errorf("archive lha extract %w", err)
	}
	var b bytes.Buffer
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// example command: lha -eq2w=destdir/ archive *
	const (
		extract     = "e"
		ignorepaths = "i"
		overwrite   = "f"
		quiet       = "q1"
		quieter     = "q2"
	)
	param := fmt.Sprintf("-%s%s%sw=%s", extract, overwrite, ignorepaths, dst)
	args := []string{param, src}
	args = append(args, targets...)
	cmd := exec.CommandContext(ctx, prog, args...)
	cmd.Stderr = &b
	out, err := cmd.Output()
	if err != nil {
		if b.String() != "" {
			return fmt.Errorf("archive lha %w: %s: %s", ErrProg, prog, strings.TrimSpace(b.String()))
		}
		return fmt.Errorf("archive lha %w: %s", err, prog)
	}
	if len(out) == 0 {
		return ErrRead
	}
	return nil
}

// Rar extracts the targets from the source RAR archive
// to the destination directory using the [unrar program].
// If the targets are empty then all files are extracted.
//
// On Linux there are two versions of the unrar program, the freeware
// version by Alexander Roshal and the feature incomplete [unrar-free].
// The freeware version is the recommended program for extracting RAR archives.
//
// [unrar program]: https://www.rarlab.com/rar_add.htm
func (x Extractor) Rar(targets ...string) error {
	src, dst := x.Source, x.Destination
	prog, err := exec.LookPath(command.Unrar)
	if err != nil {
		return fmt.Errorf("archive unrar extract %w", err)
	}
	if dst == "" {
		return ErrDest
	}
	var b bytes.Buffer
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	const (
		eXtract    = "x"   // x extract files with full path
		noPaths    = "-ep" // -ep do not preserve paths
		noComments = "-c-" // -c- do not display comments
		rename     = "-or" // -or rename files automatically
		yes        = "-y"  // -y assume yes to all queries
		outputPath = "-op" // -op output path
	)
	args := []string{eXtract, noPaths, noComments, rename, yes, src}
	args = append(args, targets...)
	args = append(args, outputPath+dst)
	cmd := exec.CommandContext(ctx, prog, args...)
	cmd.Stderr = &b
	if err = cmd.Run(); err != nil {
		if b.String() != "" {
			return fmt.Errorf("archive unrar %w: %s: %s", ErrProg, prog, strings.TrimSpace(b.String()))
		}
		return fmt.Errorf("archive unrar %w: %s", err, prog)
	}
	return nil

}

// Zip extracts the targets from the source Zip archive
// to the destination directory using the [unzip program].
// If the targets are empty then all files are extracted.
//
// [unzip program]: https://www.linux.org/docs/man1/unzip.html
func (x Extractor) Zip(targets ...string) error {
	src, dst := x.Source, x.Destination
	prog, err := exec.LookPath(command.Unzip)
	if err != nil {
		return fmt.Errorf("archive zip extract %w", err)
	}
	if dst == "" {
		return ErrDest
	}
	var b bytes.Buffer
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// [-options]
	const (
		test            = "-t"  // test archive files
		caseinsensitive = "-C"  // use case-insensitive matching
		notimestamps    = "-DD" // skip restoration of timestamps
		junkpaths       = "-j"  // junk paths, ignore directory structures
		overwrite       = "-o"  // overwrite existing files without prompting
		quiet           = "-q"  // quiet
		quieter         = "-qq" // quieter
		targetDir       = "-d"  // target directory to extract files to
		allowCtrlChars  = "-^"  // allow control characters in filenames
	)
	// unzip [-options] file[.zip] [file(s)...] [-x files(s)] [-d exdir]
	// file[.zip]		path to the zip archive
	// [file(s)...]		optional list of archived files to process, sep by spaces.
	// [-x files(s)]	optional files to be excluded.
	// [-d exdir]		optional target directory to extract files in.
	args := []string{quieter, notimestamps, allowCtrlChars, overwrite, src}
	args = append(args, targets...)
	args = append(args, targetDir, dst)
	cmd := exec.CommandContext(ctx, prog, args...)
	cmd.Stderr = &b
	if err = cmd.Run(); err != nil {
		if b.String() != "" {
			return fmt.Errorf("archive zip %w: %s: %s", ErrProg, prog, strings.TrimSpace(b.String()))
		}
		return fmt.Errorf("archive zip %w: %s", err, prog)
	}
	return nil
}

// Zip7 extracts the targets from the source 7z archive
// to the destination directory using the [7z program].
// If the targets are empty then all files are extracted.
//
// On some Linux distributions the 7z program is named 7zz.
// The legacy version of the 7z program, the p7zip package
// should not be used!
//
// [7z program]: https://www.7-zip.org/
func (x Extractor) Zip7(targets ...string) error {
	src, dst := x.Source, x.Destination
	prog, err := exec.LookPath(command.Zip7)
	if err != nil {
		return fmt.Errorf("archive 7z extract %w", err)
	}
	if dst == "" {
		return ErrDest
	}
	var b bytes.Buffer
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	const (
		extract   = "x"    // x extract files without paths
		overwrite = "-aoa" // -aoa overwrite all
		quiet     = "-bb0" // -bb0 quiet
		targetDir = "-o"   // -o output directory
		yes       = "-y"   // -y assume yes to all queries
	)
	args := []string{extract, overwrite, quiet, yes, targetDir + dst, src}
	args = append(args, targets...)
	cmd := exec.CommandContext(ctx, prog, args...)
	cmd.Stderr = &b
	if err = cmd.Run(); err != nil {
		if b.String() != "" {
			return fmt.Errorf("archive 7z %w: %s: %s", ErrProg, prog, strings.TrimSpace(b.String()))
		}
		return fmt.Errorf("archive 7z %w: %s", err, prog)
	}
	return nil
}

// ZipHW extracts the targets from the source zip archive
// to the destination directory using the [hwzip program].
// If the targets are empty then all files are extracted.
//
// hwzip is used to handle DOS era, zip archive compression methods
// that are not widely supported.
// It also does not support extracting to a target directory.
// To work around this, this copies the source archive
// to the destination directory, uses that as the working directory
// and extracts the files. The copied source archive is then removed.
//
// [arc program]: https://arj.sourceforge.net/
func (x Extractor) ZipHW(targets ...string) error {
	src, dst := x.Source, x.Destination
	if st, err := os.Stat(dst); err != nil {
		return fmt.Errorf("%w: %s", err, dst)
	} else if !st.IsDir() {
		return fmt.Errorf("%w: %s", ErrPath, dst)
	}
	prog, err := exec.LookPath(command.HWZip)
	if err != nil {
		return fmt.Errorf("archive hwzip extract %w", err)
	}

	srcInDst := filepath.Join(dst, filepath.Base(src))
	if _, err := helper.Duplicate(src, srcInDst); err != nil {
		return fmt.Errorf("archive hwzip duplicate %w", err)
	}
	defer os.Remove(srcInDst)

	var b bytes.Buffer
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	const (
		extract = "extract" // x extract files
	)
	args := []string{extract, filepath.Base(src)}
	args = append(args, targets...)
	cmd := exec.CommandContext(ctx, prog, args...)
	cmd.Dir = dst
	cmd.Stderr = &b
	if err = cmd.Run(); err != nil {
		if b.String() != "" {
			return fmt.Errorf("archive arc %w: %s: %q",
				ErrProg, prog, strings.TrimSpace(b.String()))
		}
		return fmt.Errorf("archive arc %w: %s", err, prog)
	}
	return nil
}
