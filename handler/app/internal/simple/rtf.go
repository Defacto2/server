package simple

import (
	"bytes"
	"regexp"
)

// Just a heads up that this rtf.go and the rtf_test.go were mostly coded using
// Mistral devstral-2, and then was modified by me to make it more efficient.
// You may wish to remove this functionality for licensing requirements.

// RTF detects if the content is likely RTF format by checking for a RTF header pattern.
// RTF files typically start with {\rtf followed by version number.
func RTF(b []byte) bool {
	const minLen = 10
	if len(b) < minLen {
		return false
	}
	return len(b) >= 7 &&
		b[0] == '{' &&
		b[1] == '\\' &&
		b[2] == 'r' &&
		b[3] == 't' &&
		b[4] == 'f' &&
		(b[5] >= '0' && b[5] <= '9')
}

// StripRTF removes RTF control words and formatting from text content.
// This handles RTF version 1.0 standard patterns comprehensively.
//
// This first runs [RTF] and when false returns b unmodified.
func StripRTF(b []byte) []byte {
	if len(b) == 0 || !RTF(b) {
		return b
	}

	// Comprehensive RTF 1.0 pattern removal
	// Remove RTF group delimiters
	b = bytes.ReplaceAll(b, []byte("{"), []byte(""))
	b = bytes.ReplaceAll(b, []byte("}"), []byte(""))

	// Replace paragraph marks with newlines to preserve paragraph structure (before removing other control words)
	b = bytes.ReplaceAll(b, []byte("\\par"), []byte("\n"))

	// Use regex to remove RTF control words and their numeric parameters
	// Pattern: backslash followed by letters, optional digits (including negative numbers)
	re := regexp.MustCompile(`\\[a-zA-Z]+(?:-\d+|\d*)`)
	b = re.ReplaceAll(b, []byte(""))

	// Additional pass to catch any remaining control words with parameters
	re2 := regexp.MustCompile(`\\[a-zA-Z]+`)
	b = re2.ReplaceAll(b, []byte(""))

	// Remove any remaining single letters that were part of control words (like 'd' from \deflang)
	reSingle := regexp.MustCompile(`(?:^|\s)([a-zA-Z])(?:\s|$)`)
	b = reSingle.ReplaceAll(b, []byte(" "))

	// Remove common font names that appear in RTF files
	b = bytes.ReplaceAll(b, []byte("MS Sans Serif"), []byte(""))
	b = bytes.ReplaceAll(b, []byte("Times New Roman"), []byte(""))
	b = bytes.ReplaceAll(b, []byte("Symbol"), []byte(""))
	b = bytes.ReplaceAll(b, []byte("System"), []byte(""))
	b = bytes.ReplaceAll(b, []byte("Arial"), []byte(""))
	b = bytes.ReplaceAll(b, []byte("Courier New"), []byte(""))
	b = bytes.ReplaceAll(b, []byte("Courier"), []byte(""))
	b = bytes.ReplaceAll(b, []byte("Helvetica"), []byte(""))

	// Remove remaining backslashes and semicolons
	b = bytes.ReplaceAll(b, []byte("\\"), []byte(""))
	b = bytes.ReplaceAll(b, []byte(";"), []byte(""))

	// Clean up multiple spaces but preserve line breaks
	// Replace multiple spaces with single spaces, but preserve newlines
	reSpaces := regexp.MustCompile(`[ \t\f\v\r]+`)
	b = reSpaces.ReplaceAll(b, []byte(" "))

	// Split into lines to handle paragraph structure
	lines := bytes.Split(b, []byte("\n"))
	var cleanLines [][]byte
	for _, line := range lines {
		// Clean up each line but preserve non-empty lines
		trimLine := bytes.TrimSpace(line)
		if len(trimLine) > 0 {
			cleanLines = append(cleanLines, trimLine)
		}
	}
	b = bytes.Join(cleanLines, []byte("\n"))

	return b
}
