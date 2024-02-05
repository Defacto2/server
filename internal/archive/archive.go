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

	"github.com/mholt/archiver/v3"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

const (
	// permitted archives on the site:
	// 7z,arc,ark,arj,cab,gz,lha,lzh,rar,tar,tar.gz,zip.
	arjx = ".arj" // Archived by Robert Jung
	lhax = ".lha" // LHarc by Haruyasu Yoshizaki (Yoshi)
	rarx = ".rar" // Roshal ARchive by Alexander Roshal
	zipx = ".zip" // Phil Katz's ZIP for MSDOS systems
)

var (
	ErrArchive    = errors.New("format specified by source filename is not an archive format")
	ErrDest       = errors.New("dest directory points to a file")
	ErrDir        = errors.New("is a directory")
	ErrFile       = errors.New("no such file")
	ErrMagic      = errors.New("no unsupport for magic file type")
	ErrReadr      = errors.New("system could not read the file archive")
	ErrSilent     = errors.New("archiver program silently failed, it return no output or errors")
	ErrProg       = errors.New("archive program error")
	ErrTypeOut    = errors.New("magic file program result is empty")
	ErrWriter     = errors.New("writer must be a file object")
	ErrWrongExt   = errors.New("filename has the wrong file extension")
	ErrUnknownExt = errors.New("the archive uses an unsupported file extension")
)

// ArjItem returns true if the string is a row from an ARJ list.
func ARJItem(s string) bool {
	const minLen = 6
	if len(s) < minLen {
		return false
	}
	if s[3:4] != ")" {
		return false
	}
	x := s[:3]
	if _, err := strconv.Atoi(x); err != nil {
		return false
	}
	return true
}

// CheckyCharset checks the byte slice for valid UTF-8 encoding.
// If the byte slice is not valid UTF-8, it will attempt to decode
// the byte slice using the MS-DOS era, IBM CP-437 character set.
func CheckyCharset(b []byte) string {
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

// Content returns both a list of files within an rar, tar, or zip archive;
// as-well as a suitable filename string for the archive. This filename is
// useful when the original archive filename has been given an invalid file
// extension.
//
// An absolute path is required by src that points to the archive file named as a unique id.
//
// The original archive filename with extension is required to determine text compression format.
func Content(src, filename string) ([]string, string, error) {
	st, err := os.Stat(src)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, "", fmt.Errorf("read %s: %w", filepath.Base(src), ErrFile)
	}
	if st.IsDir() {
		return nil, "", fmt.Errorf("read %s: %w", filepath.Base(src), ErrDir)
	}
	files, fname, err := Readr(src, filename)
	if err != nil {
		return nil, "", fmt.Errorf("read uuid/filename: %w", err)
	}
	return files, fname, nil
}

// Readr returns both a list of files within an rar, tar or zip archive,
// and a suitable archive filename string.
// If there are problems reading the archive due to an incorrect filename
// extension, the returned filename string will be corrected.
func Readr(src, filename string) ([]string, string, error) {
	files, err := readr(src, filename)
	if err != nil {
		fmt.Println("readr error:", err)
		return readCommand(src, filename)
	}
	return files, filename, nil
}

func readCommand(src, filename string) ([]string, string, error) {
	files, ext, err := Readr(src, filename)
	if errors.Is(err, ErrWrongExt) {
		newname := Rename(ext, filename)
		files, err = readr(src, newname)
		if err != nil {
			return nil, "", fmt.Errorf("readr fix: %w", err)
		}
		return files, newname, nil
	}
	if err != nil {
		return nil, "", fmt.Errorf("readr: %w", err)
	}
	// remove empty entries
	files = slices.DeleteFunc(files, func(s string) bool {
		return strings.TrimSpace(s) == ""
	})
	return files, filename, nil
}

func readr(src, filename string) ([]string, error) {
	name := strings.ToLower(filename) // ByExtension is case sensitive
	format, err := archiver.ByExtension(name)
	if err != nil {
		return nil, fmt.Errorf("by extension: %w", err)
	}
	w, ok := format.(archiver.Walker)
	if !ok {
		return nil, fmt.Errorf("readr %s (%T): %w", filename, format, ErrArchive)
	}
	files := []string{}
	err = w.Walk(src, func(f archiver.File) error {
		if f.IsDir() {
			return nil
		}
		if strings.TrimSpace(f.Name()) == "" {
			return nil
		}
		name := CheckyCharset([]byte(f.Name()))
		files = append(files, name)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return files, err
}

// Extract the targets file from src archive to the destination folder.
// The archive format is selected implicitly.
//
// Archiver relies on the filename extension to determine which
// decompression format to use, which must be supplied using filename.
func Extract(src, dst, filename, targets string) error {
	name := strings.ToLower(filename)
	f, err := archiver.ByExtension(name)
	if err != nil {
		return fmt.Errorf("extract %q: %w", name, err)
	}
	format, ok := f.(archiver.Extractor)
	if !ok {
		return fmt.Errorf("extract %s (%T): %w", name, f, ErrArchive)
	}
	// recover from panic caused by mholt/archiver.
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("extract panic %s: %v", name, r)
		}
	}()
	// err := format.Extract(ctx, input, fileList, handler)
	if err := format.Extract(src, targets, dst); err != nil {
		// second attempt at extraction using a system archiver program
		x := Extractor{Source: src, Destination: dst, OriginalName: filename}
		if err := x.Extract(targets); err != nil {
			return fmt.Errorf("command extract: %w", err)
		}
		return fmt.Errorf("extract: %w", err)
	}
	return nil
}

// ExtractAll decompresses the given archive file into the destination folder.
// The archive format is selected implicitly.
//
// Archiver relies on the filename extension to determine which
// decompression format to use, which must be supplied using filename.
func ExtractAll(src, dst, filename string) error {
	name := strings.ToLower(filename)
	f, err := archiver.ByExtension(name)
	if err != nil {
		return fmt.Errorf("%s: %w", filename, err)
	}
	format, ok := f.(archiver.Unarchiver)
	if !ok {
		return fmt.Errorf("extract all %s (%T): %w", filename, f, ErrArchive)
	}
	// recover from panic caused by mholt/archiver.
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("extract all panic %s: %v", name, r)
		}
	}()
	if err := format.Unarchive(src, dst); err != nil {
		// second attempt at extraction using a system archiver program
		x := Extractor{Source: src, Destination: dst, OriginalName: filename}
		if err := x.Extract(); err != nil {
			return fmt.Errorf("command extract all: %w", err)
		}
		return fmt.Errorf("extract all: %w", err)
	}
	return nil
}

// MagicExt uses the Linux file program to determine the src archive file type.
// The returned string will be a file separator and extension.
// Note both bzip2 and gzip archives return a .tar extension prefix.
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
		return "", fmt.Errorf("magic file type: %w", ErrTypeOut)
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
	for magic, ext := range magics {
		if strings.TrimSpace(s[0]) == magic {
			return ext, nil
		}
	}
	if MagicLHA(magic) {
		return lhax, nil
	}
	return "", fmt.Errorf("%w: %q", ErrMagic, magic)
}

// MagicLHA returns true if the LHA file type is matched in the magic string.
func MagicLHA(magic string) bool {
	s := strings.Split(magic, " ")
	const lha = "lha"
	if s[0] != lha {
		return false
	}
	if len(s) < len(lha) {
		return false
	}
	if strings.Join(s[0:3], " ") == "lha archive data" {
		return true
	}
	if strings.Join(s[2:4], " ") == "archive data" {
		return true
	}
	return false
}

// Rename the filename by replacing the file extension with the ext string.
// Leaving ext empty returns the filename without a file extension.
func Rename(ext, filename string) string {
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

type Contents struct {
	Files []string
	Ext   string
}

// Readr attempts to use programs on the host operating system to determine
// the src archive content and a usable filename based on its format.
func (c *Contents) Read(src, filename string) error {
	ext, err := MagicExt(src)
	if err != nil {
		return fmt.Errorf("system reader: %w", err)
	}
	if !strings.EqualFold(ext, filepath.Ext(filename)) {
		// retry using correct filename extension
		return fmt.Errorf("system reader: %w", ErrWrongExt)
	}
	switch strings.ToLower(ext) {
	case arjx:
		return c.ARJ(src)
	case lhax:
		return c.LHA(src)
	case rarx:
		return c.Rar(src)
	case zipx:
		return c.Zip(src)
	}
	return fmt.Errorf("system reader: %w", ErrReadr)
}

// ARJReader returns the content of the src ARJ archive.
// There is an internal limit of 999 items.
func (c *Contents) ARJ(src string) error {
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
		return ErrReadr
	}
	outs := strings.Split(string(out), "\n")
	files := []string{}
	const start = len("001) ")
	for _, s := range outs {
		if !ARJItem(s) {
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

// LHAReader returns the content of the src LHA/LZH archive.
func (c *Contents) LHA(src string) error {
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
		return ErrReadr
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

// Rar returns the content of the src RAR archive.
func (c *Contents) Rar(src string) error {
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
		return ErrReadr
	}
	c.Files = strings.Split(string(out), "\n")
	c.Files = slices.DeleteFunc(c.Files, func(s string) bool {
		return strings.TrimSpace(s) == ""
	})
	c.Ext = rarx
	return nil
}

// Zip returns a list of files within the src zip archive.
func (c *Contents) Zip(src string) error {
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
			//return files, zipx, nil
			return nil
		}
		// otherwise the zipinfo threw an error
		return fmt.Errorf("%q: %w", src, err)
	}
	if len(out) == 0 {
		return ErrReadr
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
	Source       string // The source archive file.
	Destination  string // The extraction destination directory.
	OriginalName string // The original filename of the archive, used to determine the archive format.
}

// Extract the targets from the src file archive
// to the dest directory using an Linux archive program.
// The program used is determined by the extension of the
// provided archive filename, which maybe different to src.
// If the targets are empty, all files are extracted.
func (x Extractor) Extract(targets ...string) error {
	ext := strings.ToLower(filepath.Ext(x.OriginalName))
	switch ext {
	case arjx:
		return x.ARJ(targets...)
	case lhax:
		return x.LHA(targets...)
	case zipx:
		return x.Zip(targets...)
	default:
		return ErrUnknownExt
	}
}

// ARJ extracts the targets from the Source ARJ archive
// to the Destination directory using the Linux arj program.
// If the targets are empty, all files are extracted.
func (x Extractor) ARJ(targets ...string) error {
	src, dst := x.Source, x.Destination
	if st, err := os.Stat(dst); err != nil {
		return fmt.Errorf("%w: %s", err, dst)
	} else if !st.IsDir() {
		return fmt.Errorf("%w: %s", ErrDest, dst)
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

// LHA extracts the targets from the src LHA/LZH archive
// to the dest directory using a Linux lha program.
// Either jlha-utils or lhasa work.
// Targets with spaces in their names are ignored by the program.
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
		return ErrSilent
	}
	return nil
}

// Zip extracts the target filenames from the src ZIP archive
// to the dest directory using the Linux unzip program.
// Multiple filenames can be separated by spaces.
func (x Extractor) Zip(targets ...string) error {
	src, dst := x.Source, x.Destination
	prog, err := exec.LookPath("unzip")
	if err != nil {
		return fmt.Errorf("unzip extract: %w", err)
	}
	if dst == "" {
		return fmt.Errorf("unzip destination is empty")
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
