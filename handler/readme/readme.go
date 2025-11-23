// Package readme provides functions for reading and suggesting readme files.
package readme

import (
	"bufio"
	"bytes"
	"cmp"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	uni "unicode"

	"github.com/Defacto2/helper"
	"github.com/Defacto2/magicnumber"
	"github.com/Defacto2/server/handler/render"
	"github.com/Defacto2/server/internal/dir"
	"github.com/Defacto2/server/internal/panics"
	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/bengarrett/ansibump"
	"github.com/bengarrett/binbump"
	"github.com/bengarrett/sauce"
	"golang.org/x/text/encoding/charmap"
)

// Suggest returns a suggested readme file name for the record.
// It prioritizes the filename and group name with a priority extension,
// such as ".nfo", ".txt", etc. If no priority extension is found,
// it will return the first text file in the content list.
//
// The archive should be the filename of the archive artifact such as the
// zip filename.
// The group should be a name or common abbreviation of the group that
// released the artifact. The content should be a list of files contained
// in the artifact.
//
// To save memory, content is not split into a slice until we need to handle it.
//
// This is a port of the CFML function, variables.findTextfile found in File.cfc.
func Suggest(archive, group string, content string) string {
	finds := SortList(true, content)
	if len(finds) == 0 {
		return ""
	}
	if len(finds) == 1 {
		return finds[0]
	}
	base := filepath.Base(archive)

	// match priority file extensions, ".nfo", ".txt", etc
	for ext := range slices.Values(priority()) {
		for name := range slices.Values(finds) {
			// use the group name as the base name
			// ie: group = "Defacto2" will match "defacto2.nfo", "DeFacto2.txt", etc
			if strings.EqualFold(group+ext, name) {
				return name
			}
			// use the archive filename as the base name
			// ie: archive = "mycollection.zip" will match "mycollection.nfo", etc
			if strings.EqualFold(base+ext, name) {
				return name
			}
		}
	}
	// match candidate filename extensions, ".diz", ".asc", etc
	for ext := range slices.Values(candidate()) {
		for name := range slices.Values(finds) {
			if strings.EqualFold(group+ext, name) {
				return name
			}
			if strings.EqualFold(base+ext, name) {
				return name
			}
		}
	}
	// match any other filenames that use a priority extension
	for name := range slices.Values(finds) {
		s := strings.ToLower(name)
		ext := filepath.Ext(s)
		if slices.Contains(priority(), ext) {
			return name
		}
	}
	// else, return the first text filename on the list
	return finds[0]
}

// SortList returns a sorted list of possible readme text files found in the file archive.
// The first result is the closes filename to root that has a priority
// filename extension such as ".nfo", then ordered alphabetically.
//
// When compact is true all filenames using extensions that are not known textfiles,
// are removed from the slice.
//
// To save memory, content is not split into a slice until we need to handle it.
func SortList(compact bool, content string) []string {
	list := strings.Split(content, "\n")
	slices.SortFunc(list, func(a, b string) int {
		a = strings.ToLower(a)
		b = strings.ToLower(b)
		aExt := strings.ToLower(filepath.Ext(a))
		ap := slices.Index(priority(), aExt)
		ac := slices.Index(candidate(), aExt)
		bExt := strings.ToLower(filepath.Ext(b))
		bp := slices.Index(priority(), bExt)
		bc := slices.Index(candidate(), bExt)
		aPaths := strings.Count(a, "/")
		bPaths := strings.Count(b, "/")
		return cmp.Or(
			cmp.Compare(aPaths, bPaths), // compare the depth of sub-directories, less is better
			cmp.Compare(ap, bp),         // compare which filename uses a priority file extension
			cmp.Compare(ac, bc),         // compare which filename uses a candidate file extension
			strings.Compare(aExt, bExt), // order all other file extensions alphabetically
			strings.Compare(a, b),       // order all other file paths alphabetically
		)
	})
	// filter out known bad filenames, such as file_id.diz or website advertising injections
	paths := make([]string, len(list))
	index := -1
	for s := range slices.Values(list) {
		path := strings.TrimSpace(s)
		if path == "" {
			continue
		}
		switch filepath.Base(path) {
		case
			"file_id.diz",
			"scene.org",
			"scene.org.txt":
			continue
		}
		v := strings.ToLower(filepath.Ext(path))
		if compact && !slices.Contains(priority(), v) && !slices.Contains(candidate(), v) {
			continue
		}
		index++
		paths[index] = path
	}
	return paths
}

// priority returns a list of readme text file extensions in priority order.
func priority() []string {
	return []string{".nfo", ".txt", ".unp", ".doc"}
}

// candidate returns a list of other, common text file extensions in priority order.
func candidate() []string {
	return []string{".diz", ".asc", ".1st", ".dox", ".me", ".cap", ".ans", ".pcb"}
}

// PlainTextBuffers returns the content of the readme file or the text of the file download.
// The first buffer is used for CP1252 and ISO-8859-1 texts while the second buffer
// is used for UTF-8 texts.
//
// The CP1252 and ISO-8859-1 Buffer may also include a FILE_ID.DIZ prefixed metadata.
// However, the UTF-8 Buffer does get the FILE_ID.DIZ prefix.
func PlainTextBuffers(
	art *models.File, sizeLimit int64, download, extra dir.Directory,
) (*bytes.Buffer, *bytes.Buffer, sauce.Record, error) {
	const msg = "readme pool"
	nosauce := sauce.Record{}
	if art == nil {
		return nil, nil, nosauce, fmt.Errorf("%s: %w", msg, panics.ErrNoArtM)
	}
	buf := new(bytes.Buffer)
	diz := new(bytes.Buffer)
	ruf := new(bytes.Buffer)
	// This might be useful if we want to force Go to not use the garbage collector.
	// buf.Reset() diz.Reset() ruf.Reset()
	err1 := render.DescriptionInZIP(diz, art, extra)
	err2 := render.InformationText(buf, ruf, sizeLimit, art, download, extra)
	var errs error
	if err1 != nil {
		errs = errors.Join(errs, fmt.Errorf("%s render diz: %w", msg, err1))
	}
	if err2 != nil {
		if errors.Is(err2, render.ErrFilename) {
			err2 = nil
		}
		if err2 != nil {
			errs = errors.Join(errs, fmt.Errorf("%s render read: %w", msg, err2))
		}
	}
	knownData := diz.Len() == 0 && buf.Len() == 0 && ruf.Len() == 0
	if knownData {
		name := download.Join(art.UUID.String)
		return knownBinaries(msg, name, errs)
	}
	// check the bytes to confirm they can be displayed as text
	r := bytes.NewReader(buf.Bytes())
	sign := magicnumber.Find(r)
	// reset text buffer for utf-16 or utf-32 text which won't be displayed
	if incompatible := sign == magicnumber.UTF16Text ||
		sign == magicnumber.UTF32Text; incompatible {
		buf.Reset()
	} else {
		// reset buffers for known images
		skip := magicnumber.Images()
		skip = append(skip, magicnumber.XBinaryText)
		if slices.Contains(skip, sign) {
			buf.Reset()
			ruf.Reset()
		}
	}
	sr := sauceData(buf)
	// text with ANSI escape codes use a custom readme template
	if match, err := UseANSICodes(bytes.NewReader(buf.Bytes())); err != nil {
		errs = errors.Join(errs, fmt.Errorf("%s incompatible ansi: %w", msg, err))
		buf.Reset()
	} else if match {
		platform := strings.TrimSpace(strings.ToLower(art.Platform.String))
		return ansiTexts(buf, diz, ruf, platform, sr, errs)
	}
	// binary texts can also cause false positives
	if binaryText := sign == magicnumber.Unknown; binaryText {
		y := art.DateIssuedYear.Int16
		return binaryTexts(buf, diz, ruf, y, sr, errs)
	}
	// modify the buffer bytes for cleanup
	return plainTexts(buf, diz, ruf, sr, errs)
}

func knownBinaries(msg, name string, errs error) (
	*bytes.Buffer, *bytes.Buffer, sauce.Record, error,
) {
	nosauce := sauce.Record{}
	file, err := os.Open(name)
	if err != nil {
		errs = errors.Join(errs, fmt.Errorf("%s not text open: %w", msg, err))
		return nil, nil, nosauce, errs
	}
	defer func() {
		_ = file.Close()
	}()
	rec, err := sauce.Read(file)
	if err != nil {
		errs = errors.Join(errs, fmt.Errorf("%s not text sauce read: %w", msg, err))
	}
	if rec == nil {
		rec = &nosauce
	}
	return nil, nil, *rec, errs
}

func sauceData(buf *bytes.Buffer) sauce.Record {
	b := buf.Bytes()
	if sauce.Contains(b) {
		return sauce.Decode(b)
	}
	return sauce.Record{}
}

func plainTexts(
	buf *bytes.Buffer, diz *bytes.Buffer, ruf *bytes.Buffer,
	sr sauce.Record, errs error) (
	*bytes.Buffer, *bytes.Buffer, sauce.Record, error,
) {
	b := trimBytes(buf.Bytes())
	if diz.Len() > 0 {
		if diz.Len() == buf.Len() {
			// for performance we want to use the bytes equal as a last resort.
			// do not use the diz buffer if it is identical the existing buf buffer.
			if !bytes.Equal(diz.Bytes(), buf.Bytes()) {
				b = render.InsertDiz(b, diz.Bytes())
			}
		} else {
			b = render.InsertDiz(b, diz.Bytes())
		}
		diz.Reset()
	}
	b = RemoveCtrls(b)
	if bytes.TrimSpace(b) == nil {
		buf.Reset()
		diz.Reset()
		ruf.Reset()
		return nil, nil, sr, errs
	}
	if len(b) > 0 {
		buf.Reset()
		buf.Write(b)
	}
	return buf, ruf, sr, nil
}

func ansiTexts(
	buf *bytes.Buffer, diz *bytes.Buffer, ruf *bytes.Buffer,
	platform string, sr sauce.Record, errs error) (
	*bytes.Buffer, *bytes.Buffer, sauce.Record, error,
) {
	width := 80 // TODO: add SAUCE width, this needs changes to the sauce package
	charset := charmap.CodePage437
	if platform == "textamiga" {
		charset = charmap.ISO8859_1
	}
	ansi, err := ansibump.Buffer(
		bytes.NewReader(buf.Bytes()), width, false, ansibump.CGA16, charset)
	if err != nil {
		errs = errors.Join(errs, err)
		return nil, nil, sr, errs
	}
	// for now we reset all other buffers
	buf.Reset()
	diz.Reset()
	ruf.Reset()
	return ansi, nil, sr, nil
}

func binaryTexts(
	buf *bytes.Buffer, diz *bytes.Buffer, ruf *bytes.Buffer,
	year int16, sr sauce.Record, errs error) (
	*bytes.Buffer, *bytes.Buffer, sauce.Record, error,
) {
	width := 0 // use default
	maxRows := 0
	pal := binbump.StandardCGA
	if y := int(year); helper.Year(y) && y < 1992 {
		width = 80
		maxRows = 25
		pal = binbump.RevisedCGA
	}
	binbuf, err := binbump.Buffer(
		bytes.NewReader(buf.Bytes()), width, maxRows, pal, nil)
	if err != nil {
		errs = errors.Join(errs, err)
		return nil, nil, sr, errs
	}
	// for now we reset all other buffers
	buf.Reset()
	diz.Reset()
	ruf.Reset()
	return binbuf, nil, sr, nil
}

func trimBytes(b []byte) []byte {
	if b == nil {
		return nil
	}
	// trim trailing whitespace and MS-DOS era EOF marker
	b = bytes.TrimRightFunc(b, uni.IsSpace)
	const endOfFile = 0x1a // Ctrl+Z
	b, _ = bytes.CutSuffix(b, []byte{endOfFile})
	return b
}

// RemoveCtrls removes ANSI escape codes and converts Windows line endings to Unix.
func RemoveCtrls(b []byte) []byte {
	const (
		reAnsi    = `\x1b\[[0-9;]*[a-zA-Z]` // ANSI escape codes
		reAmiga   = `\x1b\[[0-9;]*[ ]p`     // unknown control code found in Amiga texts
		reDEC     = `\x1b\[\?[0-9+]h`       // DEC control codes
		reSauce   = `SAUCE00.*`             // SAUCE metadata that is appended to some files
		nlWindows = "\x01\x0a"              // Windows line endings
		nlUnix    = "\x0a"                  // Unix line endings
		null      = "\x00"                  // null byte
	)
	const sep = `|`
	controlCodes := regexp.MustCompile(reAnsi + sep + reDEC + sep + reAmiga + sep + reSauce)
	b = controlCodes.ReplaceAll(b, []byte{})
	b = bytes.ReplaceAll(b, []byte(nlWindows), []byte(nlUnix))
	b = bytes.ReplaceAll(b, []byte(null), []byte(" "))
	return b
}

// UseANSICodes scans the reader and returns true if ANSI escape codes are found.
// If the reader is too large to render in a HTML template, an error is returned.
func UseANSICodes(r io.Reader) (bool, error) {
	const maximumBytes = 1024 * 1024
	const msg = "match ansi reader"
	if r == nil {
		return false, nil
	}
	mcur, mpos, sgr := moveCursor(), moveCursorToPos(), sgrWithoutReset()
	reMoveCursor := regexp.MustCompile(mcur)
	reMoveCursorToPos := regexp.MustCompile(mpos)
	reSGR := regexp.MustCompile(sgr)

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		if reMoveCursor.Match(scanner.Bytes()) {
			return true, nil
		}
		if reMoveCursorToPos.Match(scanner.Bytes()) {
			return true, nil
		}
		if reSGR.Match(scanner.Bytes()) {
			return true, nil
		}
	}
	err := scanner.Err()
	if err != nil && !errors.Is(err, bufio.ErrTooLong) {
		return false, fmt.Errorf("%s cursor scanner: %w", msg, err)
	} else if err == nil {
		return false, nil
	}
	// handle files that are too long for the scanner buffer
	// examples would be texts or ansi files with no newlines
	scanner = bufio.NewScanner(r)
	const sixtyfourK = 64 * 1024
	buf := make([]byte, 0, sixtyfourK)
	scanner.Buffer(buf, maximumBytes)
	scanner = bufio.NewScanner(r)
	for scanner.Scan() {
		if reMoveCursor.Match(scanner.Bytes()) {
			return true, nil
		}
		if reMoveCursorToPos.Match(scanner.Bytes()) {
			return true, nil
		}
		if reSGR.Match(scanner.Bytes()) {
			return true, nil
		}
	}
	if err := scanner.Err(); err != nil {
		return false, fmt.Errorf("%s, file is too large for the 1MB scanner: %w", msg, err)
	}
	return false, nil
}

// moveCursor returns a regular expression for ANSI cursor movement escape codes.
//   - match "1B" (Escape)
//   - match "[" (Left Bracket)
//   - match optional digits or if no digits, then the cursor moves 1 position
//   - match "A", "B", "C", "D", "E", "F", "G" for cursor movement up, down, left, right, etc.
func moveCursor() string {
	return `\x1b\[\d*?[ABCDEFG]`
}

// moveCursorToPos returns a regular expression for ANSI cursor position escape codes.
//   - match "1B" (Escape)
//   - match "[" (Left Bracket)
//   - match the digits for line number
//   - match ";" (semicolon)
//   - match the digits for column number
//   - match "H" cursor position or "f" cursor position
func moveCursorToPos() string {
	return `\x1b\[\d+;\d+[Hf]`
}

func sgrWithoutReset() string {
	return `\x1b\[`
}
