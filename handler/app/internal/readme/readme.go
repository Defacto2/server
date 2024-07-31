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

	"github.com/Defacto2/server/internal/magicnumber"
	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/Defacto2/server/internal/render"
)

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
	for _, ext := range priority() {
		for _, name := range finds {
			if strings.EqualFold(base+ext, name) {
				return name
			}
			if strings.EqualFold(group+ext, name) {
				return name
			}
		}
	}
	const matchFileID = "file_id.diz"
	for _, name := range finds {
		if strings.EqualFold(matchFileID, name) {
			return name
		}
	}
	// match either the filename or the group name with a candidate extension
	for _, ext := range candidate() {
		for _, name := range finds {
			if strings.EqualFold(base+ext, name) {
				return name
			}
			if strings.EqualFold(group+ext, name) {
				return name
			}
		}
	}
	// match any finds that use a priority extension
	for _, name := range finds {
		s := strings.ToLower(name)
		ext := filepath.Ext(s)
		if slices.Contains(priority(), ext) {
			return name
		}
	}
	// match the first file in the list
	for _, name := range finds {
		return name
	}
	return ""
}

// List returns a list of readme text files found in the file archive.
func List(content ...string) []string {
	finds := []string{}
	skip := []string{"scene.org", "scene.org.txt"}
	for _, name := range content {
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
func Read(art *models.File, downloadPath, extraPath string) ([]byte, error) {
	if art == nil || art.RetrotxtNoReadme.Int16 != 0 {
		return nil, nil
	}
	b, err := render.Read(art, downloadPath, extraPath)
	if err != nil {
		if errors.Is(err, render.ErrFilename) {
			return nil, nil
		}
		if errors.Is(err, render.ErrDownload) {
			return nil, render.ErrDownload
		}
		return nil, fmt.Errorf("render.Read: %w", err)
	}
	if b == nil {
		return nil, nil
	}
	r := bufio.NewReader(bytes.NewReader(b))
	// check the bytes are plain text but not utf16 or utf32
	if sign, err := magicnumber.Text(r); err != nil {
		return nil, fmt.Errorf("magicnumber.Text: %w", err)
	} else if sign == magicnumber.Unknown ||
		sign == magicnumber.UTF16Text ||
		sign == magicnumber.UTF32Text {
		return nil, nil
	}
	// trim trailing whitespace and MS-DOS era EOF marker
	b = bytes.TrimRightFunc(b, uni.IsSpace)
	const endOfFile = 0x1a // Ctrl+Z
	if bytes.HasSuffix(b, []byte{endOfFile}) {
		b = bytes.TrimSuffix(b, []byte{endOfFile})
	}
	if incompatible, err := IncompatibleANSI(r); err != nil {
		return nil, fmt.Errorf("incompatibleANSI: %w", err)
	} else if incompatible {
		return nil, nil
	}
	return RemoveCtrls(b), nil
}

// RemoveCtrls removes ANSI escape codes and converts Windows line endings to Unix.
func RemoveCtrls(b []byte) []byte {
	const (
		reAnsi    = `\x1b\[[0-9;]*[a-zA-Z]` // ANSI escape codes
		reAmiga   = `\x1b\[[0-9;]*[ ]p`     // unknown control code found in Amiga texts
		reSauce   = `SAUCE00.*`             // SAUCE metadata that is appended to some files
		nlWindows = "\r\n"                  // Windows line endings
		nlUnix    = "\n"                    // Unix line endings
	)
	controlCodes := regexp.MustCompile(reAnsi + `|` + reAmiga + `|` + reSauce)
	b = controlCodes.ReplaceAll(b, []byte{})
	b = bytes.ReplaceAll(b, []byte(nlWindows), []byte(nlUnix))
	return b
}

// IncompatibleANSI scans for HTML incompatible, ANSI cursor escape codes in the reader.
func IncompatibleANSI(r io.Reader) (bool, error) {
	scanner := bufio.NewScanner(r)
	mcur, mpos := moveCursor(), moveCursorToPos()
	reMoveCursor := regexp.MustCompile(mcur)
	reMoveCursorToPos := regexp.MustCompile(mpos)
	for scanner.Scan() {
		if reMoveCursor.Match(scanner.Bytes()) {
			return true, nil
		}
		if reMoveCursorToPos.Match(scanner.Bytes()) {
			return true, nil
		}
	}
	if err := scanner.Err(); err != nil {
		return false, fmt.Errorf("moves cursor scanner: %w", err)
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
