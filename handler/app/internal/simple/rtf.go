package simple

import (
	"regexp"
	"strings"
)

// Just a heads up that this rtf.go and the rtf_test.go was mostly coded using
// Mistral devstral-2. You may wish to remove this functionality for licencing requirements.

const minRTFLength = 10

// IsRTF detects if the content is likely RTF format by checking for the RTF header pattern.
// RTF files typically start with {\rtf followed by version number.
func IsRTF(input []byte) bool {
	if len(input) < minRTFLength {
		return false
	}

	// Check for RTF header pattern: {\rtf followed by version number
	// We look for the literal sequence: '{', '\', 'r', 't', 'f', digit
	return len(input) >= 7 &&
		input[0] == '{' &&
		len(input) > 1 && input[1] == '\\' &&
		len(input) > 2 && input[2] == 'r' &&
		len(input) > 3 && input[3] == 't' &&
		len(input) > 4 && input[4] == 'f' &&
		len(input) > 5 && (input[5] >= '0' && input[5] <= '9')
}

// StripRTFBytes removes RTF control words and formatting from text content.
// This handles RTF version 1.0 standard patterns comprehensively.
// It ONLY processes content that is detected as RTF format to avoid false positives.
// It preserves the actual text content while removing all RTF control sequences.
func StripRTFBytes(input []byte) []byte {
	if len(input) == 0 {
		return input
	}

	// Only process if this is actually RTF content
	if !IsRTF(input) {
		return input
	}

	// Comprehensive RTF 1.0 pattern removal
	s := string(input)

	// Remove RTF group delimiters
	s = strings.ReplaceAll(s, "{", "")
	s = strings.ReplaceAll(s, "}", "")

	// Replace paragraph marks with newlines to preserve paragraph structure (before removing other control words)
	s = strings.ReplaceAll(s, "\\par", "\n")

	// Use regex to remove RTF control words and their numeric parameters
	// Pattern: backslash followed by letters, optional digits (including negative numbers)
	re := regexp.MustCompile(`\\[a-zA-Z]+(?:-\d+|\d*)`)
	s = re.ReplaceAllString(s, "")

	// Additional pass to catch any remaining control words with parameters
	re2 := regexp.MustCompile(`\\[a-zA-Z]+`)
	s = re2.ReplaceAllString(s, "")

	// Remove any remaining single letters that were part of control words (like 'd' from \deflang)
	reSingle := regexp.MustCompile(`(?:^|\s)([a-zA-Z])(?:\s|$)`)
	s = reSingle.ReplaceAllString(s, " ")

	// Remove common font names that appear in RTF files
	s = strings.ReplaceAll(s, "MS Sans Serif", "")
	s = strings.ReplaceAll(s, "Times New Roman", "")
	s = strings.ReplaceAll(s, "Symbol", "")
	s = strings.ReplaceAll(s, "System", "")
	s = strings.ReplaceAll(s, "Arial", "")
	s = strings.ReplaceAll(s, "Courier New", "")
	s = strings.ReplaceAll(s, "Courier", "")
	s = strings.ReplaceAll(s, "Helvetica", "")

	// Remove remaining backslashes and semicolons
	s = strings.ReplaceAll(s, "\\", "")
	s = strings.ReplaceAll(s, ";", "")

	// Clean up multiple spaces but preserve line breaks
	// Replace multiple spaces with single spaces, but preserve newlines
	reSpaces := regexp.MustCompile(`[ \t\f\v\r]+`)
	s = reSpaces.ReplaceAllString(s, " ")

	// Split into lines to handle paragraph structure
	var cleanLines []string
	lines := strings.Split(s, "\n")
	for _, line := range lines {
		// Clean up each line but preserve non-empty lines
		trimLine := strings.TrimSpace(line)
		if trimLine != "" {
			cleanLines = append(cleanLines, trimLine)
		}
	}
	s = strings.Join(cleanLines, "\n")

	return []byte(s)
}

// StripRTF is the string version of StripRTFBytes for convenience.
func StripRTF(input string) string {
	return string(StripRTFBytes([]byte(input)))
}
