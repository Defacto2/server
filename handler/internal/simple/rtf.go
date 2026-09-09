//nolint:gochecknoglobals
package simple

import (
	"bytes"
	"regexp"
)

// Just a heads up that this rtf.go and the rtf_test.go were originally coded using
// Mistral devstral-2, and then was modified by me to make it more efficient.
// You may wish to remove this functionality for licensing requirements.

// RTF detects if the content is likely RTF format by checking for a RTF header pattern.
// RTF files typically start with {\rtf followed by version number.
func RTF(b []byte) bool {
	const minLen = 10
	if len(b) < minLen {
		return false
	}

	return bytes.HasPrefix(b, []byte("{\\rtf")) && b[5] >= '0' && b[5] <= '9'
}

var (
	reControlWords  = regexp.MustCompile(`\\[a-zA-Z]+(?:-\d+|\d*)`)
	reLeftovers     = regexp.MustCompile(`\\[a-zA-Z]+`)
	reSingleLetters = regexp.MustCompile(`(?:^|\s)([a-zA-Z])(?:\s|$)`)
	reSpaces        = regexp.MustCompile(`[ \t\f\v\r]+`)
	empty           = []byte("")
	space           = []byte(" ")
	byteLF          = []byte("\n")
)

// StripRTF removes RTF control words and formatting from text content.
// This handles RTF version 1.0 standard patterns comprehensively.
//
// This first runs [RTF] and when false returns b unmodified.
func StripRTF(b []byte) []byte {
	if len(b) == 0 || !RTF(b) {
		return b
	}

	// group delimiters
	b = bytes.ReplaceAll(b, []byte("{"), empty)
	b = bytes.ReplaceAll(b, []byte("}"), empty)

	// replace paragraph marks with newlines
	b = bytes.ReplaceAll(b, []byte("\\par"), []byte("\n"))

	// remove RTF control words and their numeric parameters
	b = reControlWords.ReplaceAll(b, nil)

	// catch remaining control words with parameters
	b = reLeftovers.ReplaceAll(b, nil)

	// remove any remaining single letters  control words
	b = reSingleLetters.ReplaceAll(b, space)

	// common font names that appear in RTF files
	b = bytes.ReplaceAll(b, []byte("MS Sans Serif"), empty)
	b = bytes.ReplaceAll(b, []byte("Times New Roman"), empty)
	b = bytes.ReplaceAll(b, []byte("Symbol"), empty)
	b = bytes.ReplaceAll(b, []byte("System"), empty)
	b = bytes.ReplaceAll(b, []byte("Arial"), empty)
	b = bytes.ReplaceAll(b, []byte("Courier New"), empty)
	b = bytes.ReplaceAll(b, []byte("Courier"), empty)
	b = bytes.ReplaceAll(b, []byte("Helvetica"), empty)

	// remaining backslashes and semicolons
	b = bytes.ReplaceAll(b, []byte("\\"), empty)
	b = bytes.ReplaceAll(b, []byte(";"), empty)

	// clean up multiple spaces but preserve line breaks
	b = reSpaces.ReplaceAll(b, space)

	lines := bytes.Split(b, byteLF)
	cleanLines := make([][]byte, 0, len(lines))

	for _, line := range lines {
		trimmed := bytes.TrimSpace(line)
		if len(trimmed) > 0 {
			cleanLines = append(cleanLines, trimmed)
		}
	}

	return bytes.Join(cleanLines, byteLF)
}
