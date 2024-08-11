package magicnumber

import (
	"bytes"
	"fmt"
	"strings"
)

// Package file id3.go contains the functions that parse bytes as common ID3 tag formats usually found in MP3 files.

// ID3v1Size is the minimum buffer size of an ID3 v1 tag.
const ID3v1Size = 128

const nul = "\x00"

// MusicID3v1 reads the [ID3 v1] tag in the byte slice and returns the song, artist and year.
// The ID3 v1 tag is a 128 byte tag at the end of an MP3 audio file.
//
// [ID3 v1]: http://id3.org/ID3v1
func MusicID3v1(p []byte) string {
	if len(p) < ID3v1Size {
		return ""
	}
	if !bytes.Equal(p[0:3], []byte{'T', 'A', 'G'}) {
		return ""
	}
	song := string(bytes.Trim(p[3:33], nul))
	song = strings.TrimSpace(song)
	artist := string(bytes.Trim(p[33:63], nul))
	artist = strings.TrimSpace(artist)
	year := string(bytes.Trim(p[93:97], nul))
	year = strings.TrimSpace(year)
	s := song
	if artist != "" {
		s += fmt.Sprintf(" by %s", artist)
	}
	if year != "" {
		s += fmt.Sprintf(" (%s)", year)
	}
	return strings.TrimSpace(s)
}

// MusicID3v2 reads the [ID3 v2] tag in the byte slice and returns the song, artist and year.
// The ID3 v2 tag is a variable length tag at the start of an MP3 audio file.
//
// [ID3 v2]: https://id3.org/id3v2-00
func MusicID3v2(p []byte) string {
	const length = 10
	if len(p) < length {
		return ""
	}
	if !bytes.Equal(p[0:3], []byte{'I', 'D', '3'}) {
		return ""
	}
	if length > len(p) {
		return ""
	}
	data := p[length:]
	switch version := p[3]; version {
	case 0x2:
		return ID3v220(data...)
	case 0x3:
		return ID3v230(data...)
	case 0x4:
		return ID3v230(data...)
	}
	return ""
}

// ID3v220 reads the [ID3 v2.2] tags in the byte slice and returns the song, artist and year.
// The v2.2 tag is obsolete but still found in the wild.
//
// [ID3 v2.2]: https://id3.org/id3v2-00
func ID3v220(data ...byte) string {
	albumTitle := [3]byte{'T', 'A', 'L'}
	leadPerformer := [3]byte{'T', 'P', '1'}
	band := [3]byte{'T', 'P', '2'}
	songName := [3]byte{'T', 'T', '2'}
	year := [3]byte{'T', 'Y', 'E'}
	s := ID3v22Frame(songName, data...)
	if s != "" {
		if lp := ID3v22Frame(leadPerformer, data...); lp != "" {
			s += fmt.Sprintf(" by %s", lp)
		} else if band := ID3v22Frame(band, data...); band != "" {
			s += fmt.Sprintf(" by %s", band)
		}
	} else if ab := ID3v22Frame(albumTitle, data...); ab == "" {
		return ""
	}
	if y := ID3v22Frame(year, data...); y != "" {
		s += fmt.Sprintf(" (%s)", y)
	}
	return strings.TrimSpace(s)
}

// ID3v22Frame reads the ID3 v2.2 frame in the byte slice and returns the frame data as a string.
// The frame header contains a 3 byte identifier followed by a 3 byte size.
func ID3v22Frame(id [3]byte, data ...byte) string {
	const header = 6
	frameID := []byte{id[0], id[1], id[2]}
	offset := bytes.Index(data, frameID)
	if offset == -1 || offset+10 > len(data) {
		return ""
	}
	b0 := int(data[offset+3]) * 16384
	b1 := int(data[offset+4]) * 128
	b2 := int(data[offset+5])
	frameLen := b0 + b1 + b2
	if offset+header+frameLen > len(data) {
		return ""
	}
	b := bytes.Trim(data[offset+header:offset+header+frameLen], nul)
	return strings.TrimSpace(string(b))
}

// TODO: handle non-ascii characters, look for extended iso-8859-1 1 byte characters and replace them with utf-8.
// https://en.wikipedia.org/wiki/ISO/IEC_8859-1

// ID3v230 reads the [ID3 v2.3] and ID3 v2.4 tags in the byte slice and returns the song, artist and year.
// The v2.3 and v2.4 tags are the most common ID3 tags found in MP3 files.
// For our purposes, we treat v2.3 and v2.4 tags the same as there's no difference for the metadata used.
//
// [ID3 v2.3]: https://id3.org/id3v2.3.0
func ID3v230(data ...byte) string {
	albumTitle := [4]byte{'T', 'A', 'L', 'B'}
	leadPerformer := [4]byte{'T', 'P', 'E', '1'}
	contentGroup := [4]byte{'T', 'I', 'T', '1'}
	songName := [4]byte{'T', 'I', 'T', '2'}
	year := [4]byte{'T', 'Y', 'E', 'R'}
	s := ID3v23Frame(songName, data...)
	if s != "" {
		if lp := ID3v23Frame(leadPerformer, data...); lp != "" {
			s += fmt.Sprintf(" by %s", lp)
		} else if cg := ID3v23Frame(contentGroup, data...); cg != "" {
			s += fmt.Sprintf(" by %s", cg)
		}
	} else if ab := ID3v23Frame(albumTitle, data...); ab == "" {
		return ""
	}
	if y := ID3v23Frame(year, data...); y != "" {
		s += fmt.Sprintf(" (%s)", y)
	}
	return strings.TrimSpace(s)
}

// ID3v23Frame reads the ID3 v2.3 and v2.4 frame in the byte slice and returns the frame data as a string.
// The frame header contains a 4 byte identifier followed by a 4 byte size.
func ID3v23Frame(id [4]byte, data ...byte) string {
	const header = 10
	frameID := []byte{id[0], id[1], id[2], id[3]}
	offset := bytes.Index(data, frameID)
	if offset == -1 || offset+10 > len(data) {
		return ""
	}
	b0 := int(data[offset+4]) * 2097152
	b1 := int(data[offset+5]) * 16384
	b2 := int(data[offset+6]) * 128
	b3 := int(data[offset+7])
	frameLen := b0 + b1 + b2 + b3
	if offset+header+frameLen > len(data) {
		return ""
	}
	b := bytes.Trim(data[offset+header:offset+header+frameLen], nul)
	return strings.TrimSpace(string(b))
}
