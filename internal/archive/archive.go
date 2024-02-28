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
	zipx = ".zip" // Phil Katz's ZIP for MSDOS systems
)

var (
	ErrDest    = errors.New("destination is empty")
	ErrExt     = errors.New("extension is not a supported archive format")
	ErrRead    = errors.New("could not read the file archive")
	ErrProg    = errors.New("program error")
	ErrFile    = errors.New("path is a directory")
	ErrPath    = errors.New("path is a file")
	ErrPanic   = errors.New("extract panic")
	ErrMissing = errors.New("path does not exist")
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
func CheckyPath(b []byte) string {
	if utf8.Valid(b) {
		return string(b)
	}
	r := transform.NewReader(bytes.NewReader(b), charmap.CodePage437.NewDecoder())
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
		return nil, fmt.Errorf("read %s: %w", filepath.Base(src), ErrMissing)
	}
	if st.IsDir() {
		return nil, fmt.Errorf("read %s: %w", filepath.Base(src), ErrFile)
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
		return nil, err
	}
	w, ok := format.(archiver.Walker)
	if !ok {
		return nil, fmt.Errorf("%w, %q", ErrExt, filename)
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
		return nil, err
	}
	return files, err
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
		return extractor(src, dst, filename, targets...)
	}
	// recover from panic caused by mholt/archiver.
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%w %s: %v", ErrPanic, name, r)
		}
	}()
	extractAll := len(targets) == 0
	if extractAll {
		all, ok := f.(archiver.Unarchiver)
		if !ok {
			return fmt.Errorf("%w, %q", ErrExt, filename)
		}
		if err = all.Unarchive(src, dst); err == nil {
			return nil
		}
		return extractor(src, dst, filename, targets...)
	}
	target, ok := f.(archiver.Extractor)
	if !ok {
		return fmt.Errorf("%w, %q", ErrExt, filename)
	}
	t := strings.Join(targets, " ")
	if err = target.Extract(src, t, dst); err == nil {
		return nil
	}
	return extractor(src, dst, filename, targets...)
}

// extractor second attempt at extraction using a system archiver program.
func extractor(src, dst, filename string, targets ...string) error {
	x := Extractor{Source: src, Destination: dst, Filename: filename}
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
		return "", fmt.Errorf("magic file type: %w", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cmd := exec.CommandContext(ctx, prog, "--brief", src)
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("magic file type: %w", err)
	}
	if len(out) == 0 {
		return "", fmt.Errorf("magic file type: %w", ErrRead)
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
	return "", fmt.Errorf("%w: %q", ErrExt, magic)
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
	Files []string // Files returns list of files within the archive.
	Ext   string   // Ext returns file extension of the archive.
}

// ARJ returns the content of the src ARJ archive,
// credited to Robert Jung, using the [arj program].
//
// [arj program]: https://arj.sourceforge.net/
func (c *Content) ARJ(src string) error {
	prog, err := exec.LookPath("arj")
	if err != nil {
		return fmt.Errorf("arj reader: %w", err)
	}

	const verboselist = "v"
	var b bytes.Buffer
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cmd := exec.CommandContext(ctx, prog, verboselist, src)
	cmd.Stderr = &b
	out, err := cmd.Output()
	if err != nil {
		return err
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
	prog, err := exec.LookPath("lha")
	if err != nil {
		return fmt.Errorf("lha reader: %w", err)
	}

	const list = "-l"
	var b bytes.Buffer
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cmd := exec.CommandContext(ctx, prog, list, src)
	cmd.Stderr = &b
	out, err := cmd.Output()
	if err != nil {
		return err
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
	prog, err := exec.LookPath("unrar")
	if err != nil {
		return fmt.Errorf("unrar reader: %w", err)
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
		return fmt.Errorf("%q: %w", src, err)
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
		return fmt.Errorf("system reader: %w", err)
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
	return fmt.Errorf("system reader: %w", ErrRead)
}

// Zip returns the content of the src ZIP archive, credited to Phil Katz,
// using the [zipinfo program].
//
// [zipinfo program]: https://infozip.sourceforge.net/
func (c *Content) Zip(src string) error {
	prog, err := exec.LookPath("zipinfo")
	if err != nil {
		return fmt.Errorf("zipinfo reader: %w", err)
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
		return fmt.Errorf("%q: %w", src, err)
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

// Extractor uses system archiver programs to extract the targets from the src file archive.
type Extractor struct {
	Source      string // The source archive file.
	Destination string // The extraction destination directory.

	// The original filename of the archive, used by Extract to determine the archive format.
	Filename string
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
	prog, err := exec.LookPath("arj")
	if err != nil {
		return fmt.Errorf("arj extract: %w", err)
	}
	var b bytes.Buffer
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// example command: arj x archive destdir/ *
	const extract = "x"
	args := []string{extract, src, dst}
	args = append(args, targets...)
	cmd := exec.CommandContext(ctx, prog, args...)
	cmd.Stderr = &b
	if err = cmd.Run(); err != nil {
		if b.String() != "" {
			return fmt.Errorf("%w: %s: %q", ErrProg, prog, strings.TrimSpace(b.String()))
		}
		return fmt.Errorf("%w: %s", err, prog)
	}
	return nil
}

// Extract the targets from the source file archive
// to the destination directory a system archive program.
// If the targets are empty then all files are extracted.
//
// The required Filename string is used to determine the archive format.
func (x Extractor) Extract(targets ...string) error {
	ext := strings.ToLower(filepath.Ext(x.Filename))
	switch ext {
	case arjx:
		return x.ARJ(targets...)
	case lhax, lhzx:
		return x.LHA(targets...)
	case zipx:
		return x.Zip(targets...)
	default:
		return ErrExt
	}
}

// LHA extracts the targets from the source LHA/LZH archive
// to the destination directory using an lha program.
// If the targets are empty then all files are extracted.
//
// On Linux either the jlha-utils or lhasa work.
func (x Extractor) LHA(targets ...string) error {
	src, dst := x.Source, x.Destination
	prog, err := exec.LookPath("lha")
	if err != nil {
		return fmt.Errorf("lha extract: %w", err)
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
			return fmt.Errorf("%w: %s: %s", ErrProg, prog, strings.TrimSpace(b.String()))
		}
		return fmt.Errorf("%s: %w", prog, err)
	}
	if len(out) == 0 {
		return ErrRead
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
	prog, err := exec.LookPath("unzip")
	if err != nil {
		return fmt.Errorf("unzip extract: %w", err)
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
		notimestamps    = "-D"  // skip restoration of timestamps
		junkpaths       = "-j"  // junk paths, ignore directory structures
		overwrite       = "-o"  // overwrite existing files without prompting
		quiet           = "-q"  // quiet
		quieter         = "-qq" // quieter
		targetDir       = "-d"  // target directory to extract files to
	)
	// unzip [-options] file[.zip] [file(s)...] [-x files(s)] [-d exdir]
	// file[.zip]		path to the zip archive
	// [file(s)...]		optional list of archived files to process, sep by spaces.
	// [-x files(s)]	optional files to be excluded.
	// [-d exdir]		optional target directory to extract files in.
	args := []string{quieter, junkpaths, overwrite, src}
	args = append(args, targets...)
	args = append(args, targetDir, dst)
	cmd := exec.CommandContext(ctx, prog, args...)
	cmd.Stderr = &b
	if err = cmd.Run(); err != nil {
		if b.String() != "" {
			return fmt.Errorf("%w: %s: %s", ErrProg, prog, strings.TrimSpace(b.String()))
		}
		return fmt.Errorf("%s: %w", prog, err)
	}
	return nil
}
