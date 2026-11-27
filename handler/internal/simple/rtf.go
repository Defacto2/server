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
	rePar          = regexp.MustCompile(`\\par\b`)
	reControlWords = regexp.MustCompile(`\\[a-zA-Z]+(?:-\d+|\d*)`)
	reSpaces       = regexp.MustCompile(`[ \t\f\v\r]+`)
	empty          = []byte("")
	space          = []byte(" ")
	byteLF         = []byte("\n")
)

// stripRTFGroups removes nested RTF metadata groups (such as {\fonttbl...}
// and {\colortbl...}) before braces are stripped so font names and style
// definitions do not leak into the plain text output.
func stripRTFGroups(b []byte) []byte {
	headers := [...][]byte{
		[]byte(`{\fonttbl`),
		[]byte(`{\colortbl`),
		[]byte(`{\stylesheet`),
		[]byte(`{\info`),
	}
	out := make([]byte, 0, len(b))
	for i := 0; i < len(b); {
		if b[i] == '{' {
			matched := false
			for _, hdr := range headers {
				if bytes.HasPrefix(b[i:], hdr) {
					matched = true
					break
				}
			}
			if matched {
				depth := 1
				i++
				for i < len(b) && depth > 0 {
					switch b[i] {
					case '\\':
						i++ // skip escaped character
					case '{':
						depth++
					case '}':
						depth--
					}
					i++
				}
				continue
			}
		}
		out = append(out, b[i])
		i++
	}
	return out
}

// StripRTF removes RTF control words and formatting from text content.
// This handles RTF version 1.0 standard patterns comprehensively.
//
// This first runs [RTF] and when false returns b unmodified.
func StripRTF(b []byte) []byte {
	if !RTF(b) {
		return b
	}

	// remove font/color/stylesheet/info header groups
	b = stripRTFGroups(b)

	// replace paragraph marks with newlines (without matching \pard, etc.)
	b = rePar.ReplaceAll(b, byteLF)

	// remove RTF control words and their numeric parameters
	b = reControlWords.ReplaceAll(b, nil)

	// group delimiters and remaining backslashes
	b = bytes.ReplaceAll(b, []byte("{"), empty)
	b = bytes.ReplaceAll(b, []byte("}"), empty)
	b = bytes.ReplaceAll(b, []byte("\\"), empty)

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
