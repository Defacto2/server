package magicnumber

import (
	"bytes"
	"fmt"
	"strings"
)

// Package file synthesismusic.go contains the functions that parse bytes as common synthesis and tracker music formats.

func Midi(p []byte) bool {
	const min = 4
	if len(p) < min {
		return false
	}
	return bytes.Equal(p[:min], []byte{'M', 'T', 'h', 'd'})
}

// Mod matches common tracked music formats in the byte slice.
func Mod(p []byte) bool {
	return MusicMod(p) != ""
}

// MusicMod returns the tracked music format in the byte slice and
// the name or title of the song if available.
// The tracked music formats include MultiTracker, Impulse Tracker,
// Extended Module, and 4 channel MODule music.
//
// [Modland] has a large collection of tracked music format documentation.
//
// [Modland]: https://ftp.modland.com/pub/documents/format_documentation/
func MusicMod(p []byte) string {
	const offset, length = 1080, 4
	if len(p) < offset+length {
		return ""
	}
	mtnHeader := p[0:3]
	switch {
	case bytes.Equal(mtnHeader, []byte{'M', 'T', 'M'}):
		s := "MultiTrack module music"
		name := string(bytes.Trim(p[4:20+4], "\x00"))
		if name != "" {
			s += fmt.Sprintf(", %q", strings.TrimSpace(name))
		}
		return s
	}
	impulse := p[0:4]
	switch {
	case bytes.Equal(impulse, []byte{'I', 'M', 'P', 'M'}):
		s := "Impulse Tracker module music"
		name := string(bytes.Trim(p[4:20+4], "\x00"))
		if name != "" {
			s += fmt.Sprintf(", %q", strings.TrimSpace(name))
		}
		return s
	}
	xmHeader := p[0:17]
	fmt.Println(xmHeader, fmt.Sprintf("%q", xmHeader))
	switch { //Extended module:
	// Extended Module:
	case bytes.Equal(xmHeader, []byte{'E', 'x', 't', 'e', 'n', 'd', 'e', 'd', 0x20,
		'M', 'o', 'd', 'u', 'l', 'e', ':', 0x20}):
		s := "extended module tracked music"
		name := strings.TrimSpace(string(bytes.Trim(p[17:17+20], "\x00")))
		if name != "" {
			s += fmt.Sprintf(", %q", name)
		}
		return s
	}
	modHeader := p[offset : offset+length]
	switch {
	case bytes.Equal(modHeader, []byte{'M', '.', 'K', '.'}):
		// The original Amiga ProTracker MOD format had no signature.
		// The M.K. signature was added by Mahoney & Kaktus in their MOD samples,
		// and became a common signature in the MOD format.
		s := "ProTracker module music"
		name := string(bytes.Trim(p[0:20], "\x00"))
		name = strings.TrimSpace(name)
		if name != "" {
			s += fmt.Sprintf(", %q", name)
		}
		return s
	}
	return ""
}
