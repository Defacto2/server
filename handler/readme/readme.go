// Package readme provides functions for reading and suggesting readme files.
package readme

import (
	"bufio"
	"bytes"
	"cmp"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	uni "unicode"

	"github.com/Defacto2/magicnumber"
	"github.com/Defacto2/server/handler/render"
	"github.com/Defacto2/server/internal/dir"
	"github.com/Defacto2/server/internal/postgres/models"
)

var ErrNoModel = errors.New("no model")

// Suggest returns a suggested readme file name for the record.
// It prioritizes the filename and group name with a priority extension,
// such as ".nfo", ".txt", etc. If no priority extension is found,
// it will return the first textfile in the content list.
//
// The filename should be the name of the file archive artifact.
// The group should be a name or common abbreviation of the group that
// released the artifact. The content should be a list of files contained
// in the artifact.
//
// This is a port of the CFML function, variables.findTextfile found in File.cfc.
func Suggest(filename, group string, content ...string) string {
	finds := List(content...)
	if len(finds) == 1 {
		return finds[0]
	}
	finds = SortContent(finds...)

	// match either the filename or the group name with a priority extension
	// e.g. .nfo, .txt, .unp, .doc
	base := filepath.Base(filename)
	for ext := range slices.Values(priority()) {
		for name := range slices.Values(finds) {
			if strings.EqualFold(base+ext, name) {
				return name
			}
			if strings.EqualFold(group+ext, name) {
				return name
			}
		}
	}
	// match either the filename or the group name with a candidate extension
	for ext := range slices.Values(candidate()) {
		for name := range slices.Values(finds) {
			if strings.EqualFold(base+ext, name) {
				return name
			}
			if strings.EqualFold(group+ext, name) {
				return name
			}
		}
	}
	// match any finds that use a priority extension
	for name := range slices.Values(finds) {
		s := strings.ToLower(name)
		ext := filepath.Ext(s)
		if slices.Contains(priority(), ext) {
			return name
		}
	}
	// match the first file in the list
	for name := range slices.Values(finds) {
		return name
	}
	return ""
}

// List returns a list of readme text files found in the file archive.
func List(content ...string) []string {
	finds := []string{}
	skip := []string{"scene.org", "scene.org.txt"}
	for name := range slices.Values(content) {
		if name == "" {
			continue
		}
		s := strings.ToLower(name)
		if slices.Contains(skip, s) {
			continue
		}
		ext := filepath.Ext(s)
		if slices.Contains(priority(), ext) {
			finds = append(finds, name)
			continue
		}
		if slices.Contains(candidate(), ext) {
			finds = append(finds, name)
		}
	}
	return finds
}

// priority returns a list of readme text file extensions in priority order.
func priority() []string {
	return []string{".nfo", ".txt", ".unp", ".doc"}
}

// candidate returns a list of other, common text file extensions in priority order.
func candidate() []string {
	return []string{".diz", ".asc", ".1st", ".dox", ".me", ".cap", ".ans", ".pcb"}
}

// SortContent sorts the content list by the number of slashes in each string.
// It prioritizes strings with fewer slashes (i.e., closer to the root).
// If the number of slashes is the same, it sorts alphabetically.
func SortContent(content ...string) []string {
	const windowsPath = "\\"
	const pathSeparator = "/"
	slices.SortFunc(content, func(a, b string) int {
		a = strings.ReplaceAll(a, windowsPath, pathSeparator)
		b = strings.ReplaceAll(b, windowsPath, pathSeparator)
		aCount := strings.Count(a, pathSeparator)
		bCount := strings.Count(b, pathSeparator)
		if aCount != bCount {
			return aCount - bCount
		}
		return cmp.Compare(strings.ToLower(a), strings.ToLower(b))
	})
	return content
}

// Read returns the content of the readme file or the text of the file download.
func Read(art *models.File, download, extra dir.Directory) ([]byte, []rune, error) {
	if art == nil {
		return nil, nil, fmt.Errorf("art in read, %w", ErrNoModel)
	}
	b, r, err := render.Read(art, download, extra)
	if err != nil {
		if errors.Is(err, render.ErrFilename) {
			return nil, nil, nil
		}
		if errors.Is(err, render.ErrDownload) {
			return nil, nil, err
		}
		return nil, nil, err
	}
	if b == nil {
		return nil, nil, nil
	}
	nr := bytes.NewReader(b)
	// check the bytes are plain text but not utf16 or utf32
	if sign, err := magicnumber.Text(nr); err != nil {
		return nil, nil, fmt.Errorf("magicnumber.Text: %w", err)
	} else if sign == magicnumber.Unknown ||
		sign == magicnumber.UTF16Text ||
		sign == magicnumber.UTF32Text {
		return nil, nil, nil
	}
	// trim trailing whitespace and MS-DOS era EOF marker
	b = bytes.TrimRightFunc(b, uni.IsSpace)
	const endOfFile = 0x1a // Ctrl+Z
	if bytes.HasSuffix(b, []byte{endOfFile}) {
		b = bytes.TrimSuffix(b, []byte{endOfFile})
	}
	incompatible, err := IncompatibleANSI(nr)
	if err != nil {
		return nil, nil, fmt.Errorf("incompatibleANSI: %w", err)
	} else if incompatible {
		b = nil
	}
	// insert the file_id.diz content into the readme text
	diz, err := render.Diz(art, extra)
	if err != nil {
		return nil, nil, fmt.Errorf("render.Diz: %w", err)
	}
	if diz != nil {
		b = render.InsertDiz(b, diz)
	}
	return RemoveCtrls(b), r, nil
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
	)
	const sep = `|`
	controlCodes := regexp.MustCompile(reAnsi + sep + reDEC + sep + reAmiga + sep + reSauce)
	b = controlCodes.ReplaceAll(b, []byte{})
	b = bytes.ReplaceAll(b, []byte(nlWindows), []byte(nlUnix))
	return b
}

// IncompatibleANSI scans for HTML incompatible, ANSI cursor escape codes in the reader.
func IncompatibleANSI(r io.Reader) (bool, error) {
	if r == nil {
		return false, nil
	}
	mcur, mpos := moveCursor(), moveCursorToPos()
	reMoveCursor := regexp.MustCompile(mcur)
	reMoveCursorToPos := regexp.MustCompile(mpos)

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		if reMoveCursor.Match(scanner.Bytes()) {
			return true, nil
		}
		if reMoveCursorToPos.Match(scanner.Bytes()) {
			return true, nil
		}
	}
	err := scanner.Err()
	if err != nil && !errors.Is(err, bufio.ErrTooLong) {
		return false, fmt.Errorf("incompatible ansi cursor scanner: %w", err)
	} else if err == nil {
		return false, nil
	}
	// handle files that are too long for the scanner buffer
	// examples would be texts or ansi files with no newlines
	scanner = bufio.NewScanner(r)
	const sixtyfourK = 64 * 1024
	buf := make([]byte, 0, sixtyfourK)
	const oneMegabyte = 1024 * 1024
	scanner.Buffer(buf, oneMegabyte)
	scanner = bufio.NewScanner(r)
	for scanner.Scan() {
		if reMoveCursor.Match(scanner.Bytes()) {
			return true, nil
		}
		if reMoveCursorToPos.Match(scanner.Bytes()) {
			return true, nil
		}
	}
	if err := scanner.Err(); err != nil {
		return false, fmt.Errorf("incompatible ansi, file is too large for the 1MB scanner: %w", err)
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
