// Package msdos provides functions for working with MS-DOS FAT 12/16 file system filenames.
package msdos

import (
	"path/filepath"
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

const (
	// ExtensionLen is the maximum length of a filename extension on a FAT 12/16 file system including the dot.
	ExtensionLen = 4
	// BaseLen is the maximum length of a filename without the extension on a FAT 12/16 file system.
	BaseLen = 8
)

// Special returns true if the rune is a non alphanumeric character that is allowed in a
// MS-DOS FAT 12/16 format file system.
func special(r rune) bool {
	const (
		underscore  = '_'
		caret       = '^'
		dollar      = '$'
		tilde       = '~'
		exclamation = '!'
		number      = '#'
		percent     = '%'
		ampersand   = '&'
		hyphen      = '-'
		open        = '{'
		closer      = '}'
		at          = '@'
		quote       = '`'
		apostrophe  = '\''
		openParen   = '('
		closeParen  = ')'
	)
	switch r {
	case underscore, caret, dollar, tilde, exclamation, number,
		percent, ampersand, hyphen, open, closer, at, quote,
		apostrophe, openParen, closeParen:
		return true
	}
	return false
}

// DirName returns a FAT 16 compatible string based on the provided named directory. It uses
// the Rename function to convert the directory name into a FAT 16 compatible format and trims
// the result to the maximum length of 8 characters. Directories can have an optional extension
// of up to 3 characters long.
func DirName(name string) string {
	dir := Rename(name)
	ext := filepath.Ext(dir)
	base := strings.TrimSuffix(dir, ext)
	if len(base) > BaseLen {
		base = base[:BaseLen]
	}
	return base + ext
}

// Rename returns a FAT 16 compatible string based on the provided filename. It replaces
// accented characters with their closest Latin equivalent and all other unsupported characters
// with an 'X'. The resulting filename has all spaces replaced with underscores and letters
// returned as uppercase. Any provided filename extension can up to 3 characters long.
//
// Many legacy archive formats such as ZIP and LHA were usable on multiple operating systems and
// file systems. These archives can contain filenames that can be listed by PKZIP or LHA on an
// MS-DOS system but are not viewable on the platform or in a emulator. Fat16Rename can be used to
// rename these filenames into a format that is compatible with the MS-DOS FAT 12/16 file system.
//
// This function does not impose the maximum length limit of 8 characters for the base filename.
// The list of supported characters were taken from the [MS-DOS 6 Concise User's Guide].
//
// [MS-DOS 6 Concise User's Guide]: https://archive.org/details/microsoft-ms-dos-6/page/n25/mode/2up
func Rename(filename string) string {
	s := strings.TrimSpace(strings.ToUpper(filename))
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	x, _, _ := transform.String(t, s)

	name := []rune(x)
	l := len(name)
	for i, r := range name {
		if unicode.Is(unicode.Latin, r) || unicode.Is(unicode.Number, r) {
			continue
		}
		if special(r) {
			continue
		}
		if unicode.Is(unicode.Space, r) {
			name[i] = '_'
			continue
		}
		if r == '.' && !linuxHideMarker(i) {
			if validExtension(i, l) {
				continue
			}
		}
		name[i] = 'X'
	}
	return string(name)
}

// On many systems, files starting with a dot are marked as hidden.
// But in MS-DOS, the dot can only be used once as the filename extension separator.
func linuxHideMarker(i int) bool {
	return i == 0
}

// validExtension returns true if the index is within the last 4 characters of the filename and
// the index is not the last character.
// MS-DOS only allow a single dot as the extension separator and the extension must be 1-3 characters long.
func validExtension(i, l int) bool {
	return i >= l-ExtensionLen && i < l-1
}

// Truncate returns the filename in a MS-DOS 8.3 friendly format, truncating the name if necessary.
// For example, "my backup collection.7zip" would return "my bac~1.7zi".
// The base filename is permitted to be up to 8 characters long and the optional file extension
// is 1 to 3 characters long plus a "." separator.
func Truncate(filename string) string {
	ext := filepath.Ext(filename)
	name := strings.TrimSuffix(filename, ext)
	if len(ext) > ExtensionLen {
		ext = ext[:ExtensionLen]
	}
	if len(name) > BaseLen {
		return name[:BaseLen-2] + "~1" + ext
	}
	return name + ext
}
