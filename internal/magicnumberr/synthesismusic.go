package magicnumberr

// Package file synthesismusic.go contains the functions that parse bytes as common synthesis and tracker music formats.

import (
	"bytes"
	"fmt"
	"io"
	"strings"
)

func Midi(r io.ReaderAt) bool {
	const size = 4
	p := make([]byte, size)
	sr := io.NewSectionReader(r, 0, size)
	if n, err := sr.Read(p); err != nil || n < size {
		return false
	}
	return bytes.Equal(p, []byte{'M', 'T', 'h', 'd'})
}

// Mod matches common tracked music formats.
func Mod(r io.ReaderAt) bool {
	return MusicTracker(r) != ""
}

// MTM matches the MultiTracker music format.
func MTM(r io.ReaderAt) bool {
	return MusicMTM(r) != ""
}

// XM matches the eXtended Module tracked music format.
func XM(r io.ReaderAt) bool {
	return MusicXM(r) != ""
}

// IT matches the Impulse Tracker music format.
func IT(r io.ReaderAt) bool {
	return MusicIT(r) != ""
}

func MK(r io.ReaderAt) bool {
	return MusicMK(r) != ""
}

// MusicMod returns the tracked music format in the byte slice and
// the name or title of the song if available.
// The tracked music formats include MultiTracker, Impulse Tracker,
// Extended Module, and 4 channel MODule music.
//
// [Modland] has a large collection of tracked music format documentation.
//
// [Modland]: https://ftp.modland.com/pub/documents/format_documentation/
func MusicTracker(r io.ReaderAt) string {
	if s := MusicMTM(r); s != "" {
		return s
	}
	if s := MusicIT(r); s != "" {
		return s
	}
	if s := MusicXM(r); s != "" {
		return s
	}
	if s := MusicMK(r); s != "" {
		return s
	}
	return ""
}

// MusicMTM returns the [MultiTracker] song or title in the byte slice if available.
// The MultiTracker format is a tracked music format created by the scene group Renaissance.
//
// [MultiTracker]: https://ftp.modland.com/pub/documents/format_documentation/MultiTracker%20(.mtm).txt
func MusicMTM(r io.ReaderAt) string {
	const sizeID = 3
	p := make([]byte, sizeID)
	sr := io.NewSectionReader(r, 0, sizeID)
	if n, err := sr.Read(p); err != nil || n < sizeID {
		return ""
	}
	if !bytes.Equal(p, []byte{'M', 'T', 'M'}) {
		return ""
	}
	const size = 20
	p = make([]byte, size)
	sr = io.NewSectionReader(r, 4, size)
	if n, err := sr.Read(p); err != nil || n < size {
		return ""
	}
	s := "MultiTrack song"
	song := string(bytes.Trim(p, "\x00"))
	song = strings.TrimSpace(song)
	if song != "" {
		s += fmt.Sprintf(", %q", song)
	}
	return s
}

// MusicIT returns the [Impulse Tracker] song or title in the byte slice if available.
// The Impulse Tracker format is a tracked music format created by Jeffrey Lim.
//
// [Impulse Tracker]: https://ftp.modland.com/pub/documents/format_documentation/Impulse%20Tracker%20v2.04%20(.it).html
func MusicIT(r io.ReaderAt) string {
	const sizeID = 4
	p := make([]byte, sizeID)
	sr := io.NewSectionReader(r, 0, sizeID)
	if n, err := sr.Read(p); err != nil || n < sizeID {
		return ""
	}
	if !bytes.Equal(p, []byte{'I', 'M', 'P', 'M'}) {
		return ""
	}
	const size = 20
	p = make([]byte, size)
	sr = io.NewSectionReader(r, 4, size)
	if n, err := sr.Read(p); err != nil || n < size {
		return ""
	}
	s := "Impulse Tracker song"
	song := string(bytes.Trim(p, "\x00"))
	song = strings.TrimSpace(song)
	if song != "" {
		s += fmt.Sprintf(", %q", song)
	}
	return s
}

// MusicXM returns the [eXtended Module] song or title in the byte slice if available.
// The XM format was originally used by FastTracker II (FT2) and later modified by other trackers.
//
// [eXtended Module]: https://ftp.modland.com/pub/documents/format_documentation/FastTracker%202%20v2.04%20(.xm).html
func MusicXM(r io.ReaderAt) string {
	const sizeID = 17
	p := make([]byte, sizeID)
	sr := io.NewSectionReader(r, 0, sizeID)
	if n, err := sr.Read(p); err != nil || n < sizeID {
		return ""
	}
	if !bytes.Equal(p, []byte{
		'E', 'x', 't', 'e', 'n', 'd', 'e', 'd', 0x20,
		'M', 'o', 'd', 'u', 'l', 'e', ':', 0x20,
	}) {
		return ""
	}
	const size = 20
	p = make([]byte, size)
	sr = io.NewSectionReader(r, sizeID, size)
	if n, err := sr.Read(p); err != nil || n < size {
		return ""
	}
	s := "extended module tracked music"
	song := string(bytes.Trim(p, "\x00"))
	song = strings.TrimSpace(song)
	if song != "" {
		s += fmt.Sprintf(", %q", song)
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
func MusicMK(r io.ReaderAt) string {
	const size = 4
	const offset = 1080
	p := make([]byte, size)
	sr := io.NewSectionReader(r, offset, size)
	if n, err := sr.Read(p); err != nil || n < size {
		return ""
	}
	switch {
	case
		bytes.Equal(p, []byte{'2', 'C', 'H', 'N'}):
		return music2Chan(r)
	case
		bytes.Equal(p, []byte{'M', '.', 'K', '.'}),
		bytes.Equal(p, []byte{'M', '!', 'K', '!'}),
		bytes.Equal(p, []byte{'4', 'C', 'H', 'N'}),
		bytes.Equal(p, []byte{'F', 'L', 'T', '4'}):
		return music4Chan(r)
	case
		bytes.Equal(p, []byte{'6', 'C', 'H', 'N'}):
		return music6Chan(r)
	case
		bytes.Equal(p, []byte{'F', 'L', 'T', '8'}),
		bytes.Equal(p, []byte{'O', 'C', 'T', 'A'}),
		bytes.Equal(p, []byte{'8', 'C', 'H', 'N'}):
		return music8Chan(r)
	default:
		return ""
	}
}

func music2Chan(r io.ReaderAt) string {
	s := "ProTracker 2-channel song"
	return modSong(s, r)
}

func music4Chan(r io.ReaderAt) string {
	s := "ProTracker 4-channel song"
	return modSong(s, r)
}

func music6Chan(r io.ReaderAt) string {
	s := "ProTracker 6-channel song"
	return modSong(s, r)
}

func music8Chan(r io.ReaderAt) string {
	s := "ProTracker 8-channel song"
	return modSong(s, r)
}

func modSong(info string, r io.ReaderAt) string {
	const size = 20
	const offset = 1084
	p := make([]byte, size)
	sr := io.NewSectionReader(r, offset, size)
	if n, err := sr.Read(p); err != nil || n < size {
		return info
	}
	song := string(bytes.Trim(p, "\x00"))
	song = strings.TrimSpace(song)
	s := info
	if song != "" {
		s += fmt.Sprintf(", %q", song)
	}
	return s
}
