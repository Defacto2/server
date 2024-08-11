package magicnumber

import (
	"bytes"
	"fmt"
	"strings"
)

// Package file id3.go contains the functions that parse bytes as common ID3 tag formats usually found in MP3 files.

// ID3v1Size is the minimum buffer size of an ID3 v1 tag.
const ID3v1Size = 128

// MusicID3v1 reads the [ID3 v1] tag in the byte slice and returns the title and artist.
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
	song := string(bytes.Trim(p[3:33], "\x00"))
	song = strings.TrimSpace(song)
	artist := string(bytes.Trim(p[33:63], "\x00"))
	artist = strings.TrimSpace(artist)
	year := string(bytes.Trim(p[93:97], "\x00"))
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

func headerSize(p []byte) int {
	const length = 10
	if len(p) < length {
		return 0
	}
	b0 := int(p[6]) * 2097152
	b1 := int(p[7]) * 16384
	b2 := int(p[8]) * 128
	b3 := int(p[9]) * 1
	return b0 + b1 + b2 + b3
}

func MusicID3v2(p []byte) string {
	const length = 10
	if len(p) < length {
		return ""
	}
	if !bytes.Equal(p[0:3], []byte{'I', 'D', '3'}) {
		return ""
	}
	hsize := headerSize(p)
	if hsize+length > len(p) {
		return ""
	}
	data := p[length : length+hsize]
	switch version := p[3]; version {
	case 0x2:
		return "ID3v2.2"
	case 0x3:
		return ID3v230(data...)
	case 0x4:
		return "ID3v2.4"
	}
	return ""
}

func ID3v230(data ...byte) string {
	ablumTitle := [4]byte{'T', 'A', 'L', 'B'}
	leadPerformer := [4]byte{'T', 'P', 'E', '1'}
	contentGroup := [4]byte{'T', 'I', 'T', '1'}
	songName := [4]byte{'T', 'I', 'T', '2'}
	year := [4]byte{'T', 'Y', 'E', 'R'}
	s := ID3v2Frame(songName, data...)
	if s != "" {
		if lp := ID3v2Frame(leadPerformer, data...); lp != "" {
			s += fmt.Sprintf(" by %s", lp)
		} else if cg := ID3v2Frame(contentGroup, data...); cg != "" {
			s += fmt.Sprintf(" by %s", cg)
		}
	} else if ab := ID3v2Frame(ablumTitle, data...); ab == "" {
		return ""
	}
	if y := ID3v2Frame(year, data...); y != "" {
		s += fmt.Sprintf(" (%s)", y)
	}
	return strings.TrimSpace(s)
}

func ID3v2Frame(id [4]byte, data ...byte) string {
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
	return strings.TrimSpace(string(data[offset+header : offset+header+frameLen]))
}
