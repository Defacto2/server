package magicnumber

import (
	"bytes"
	"fmt"
	"strings"
)

// Package file synthesismusic.go contains the functions that parse bytes as common synthesis and tracker music formats.

// MusicTrackerSize is the minimum buffer size of a tracked music metadata.
const MusicTrackerSize = 1024 * 2

func Midi(p []byte) bool {
	const min = 4
	if len(p) < min {
		return false
	}
	return bytes.Equal(p[:min], []byte{'M', 'T', 'h', 'd'})
}

// Mod matches common tracked music formats in the byte slice.
func Mod(p []byte) bool {
	return MusicTracker(p) != ""
}

// MusicMod returns the tracked music format in the byte slice and
// the name or title of the song if available.
// The tracked music formats include MultiTracker, Impulse Tracker,
// Extended Module, and 4 channel MODule music.
//
// [Modland] has a large collection of tracked music format documentation.
//
// [Modland]: https://ftp.modland.com/pub/documents/format_documentation/
func MusicTracker(p []byte) string {
	if s := MusicMTM(p); s != "" {
		return s
	}
	if s := MusicIT(p); s != "" {
		return s
	}
	if s := MusicXM(p); s != "" {
		return s
	}
	if s := MusicMK(p); s != "" {
		return s
	}
	return ""
}

// MusicMTM returns the [MultiTracker] song or title in the byte slice if available.
// The MultiTracker format is a tracked music format created by the scene group Renaissance.
//
// [MultiTracker]: https://ftp.modland.com/pub/documents/format_documentation/MultiTracker%20(.mtm).txt
func MusicMTM(p []byte) string {
	const id, offset, headerLen = 3, 4, 20
	if len(p) < id {
		return ""
	}
	mtnHeader := p[0:id]
	if !bytes.Equal(mtnHeader, []byte{'M', 'T', 'M'}) {
		return ""
	}
	s := "MultiTrack song"
	song := string(bytes.Trim(p[offset:headerLen+offset], "\x00"))
	if song != "" {
		s += fmt.Sprintf(", %q", strings.TrimSpace(song))
	}
	return s
}

// MusicIT returns the [Impulse Tracker] song or title in the byte slice if available.
// The Impulse Tracker format is a tracked music format created by Jeffrey Lim.
//
// [Impulse Tracker]: https://ftp.modland.com/pub/documents/format_documentation/Impulse%20Tracker%20v2.04%20(.it).html
func MusicIT(p []byte) string {
	const id, offset, headerLen = 4, 4, 20
	impulse := p[0:id]
	if !bytes.Equal(impulse, []byte{'I', 'M', 'P', 'M'}) {
		return ""
	}
	s := "Impulse Tracker song"
	song := string(bytes.Trim(p[offset:headerLen+offset], "\x00"))
	if song != "" {
		s += fmt.Sprintf(", %q", strings.TrimSpace(song))
	}
	return s
}

// MusicXM returns the [eXtended Module] song or title in the byte slice if available.
// The XM format was originally used by FastTracker II (FT2) and later modified by other trackers.
//
// [eXtended Module]: https://ftp.modland.com/pub/documents/format_documentation/FastTracker%202%20v2.04%20(.xm).html
func MusicXM(p []byte) string {
	const id, offset, headerLen = 17, 17, 20
	if len(p) < id {
		return ""
	}
	xmHeader := p[0:id]
	if !bytes.Equal(xmHeader, []byte{'E', 'x', 't', 'e', 'n', 'd', 'e', 'd', 0x20,
		'M', 'o', 'd', 'u', 'l', 'e', ':', 0x20}) {
		return ""
	}
	s := "extended module tracked music"
	song := string(bytes.Trim(p[offset:headerLen+offset], "\x00"))
	if song != "" {
		s += fmt.Sprintf(", %q", strings.TrimSpace(song))
	}
	return s
}

// MusicMK returns the MOD song or title in the byte slice if available.
// The Soundtracker MOD format is a tracked music format created by Karsten Obarski on the Commodore Amiga.
// The original MOD format had no signature, but the M.K. signature was added by Mahoney & Kaktus
// in their MOD samples and became a common signature in the MOD format.
//
// Common MOD formats include the original The Ultimate Soundtracker, Protracker, FastTracker II...
//
// [ProTracker]: https://ftp.modland.com/pub/documents/format_documentation/ProTracker%20v1.0%20(.mod).html
func MusicMK(p []byte) string {
	const offset, length, headerLen = 1080, 4, 20
	if len(p) < offset+length+headerLen {
		return ""
	}
	modHeader := p[offset : offset+length]
	if !bytes.Equal(modHeader, []byte{'M', '.', 'K', '.'}) {
		return ""
	}
	s := "ProTracker song"
	song := string(bytes.Trim(p[0:headerLen], "\x00"))
	song = strings.TrimSpace(song)
	fmt.Println("MusicMK", string(modHeader), "<>", song)
	if song != "" {
		s += fmt.Sprintf(", %q", song)
	}
	return s
}
