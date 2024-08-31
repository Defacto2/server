package magicnumberr

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"

	"golang.org/x/text/encoding/charmap"
)

// Package file id3.go contains the functions that parse bytes as common ID3 tag formats usually found in MP3 files.

// ID3v1Size is the minimum buffer size of an ID3 v1 tag.
const ID3v1Size = 128

const nul = "\x00"

// Length returns the length of the reader.
func Length(r io.ReaderAt) int64 {
	seeker, ok := r.(io.Seeker)
	if !ok {
		return 0
	}
	length, err := seeker.Seek(0, io.SeekEnd)
	if err != nil {
		return 0
	}
	_, err = seeker.Seek(0, io.SeekStart)
	if err != nil {
		return 0
	}
	return length
}

// MusicID3v1 reads the [ID3 v1] tag in the byte slice and returns the song, artist and year.
// The ID3 v1 tag is a 128 byte tag at the end of an MP3 audio file.
//
// [ID3 v1]: http://id3.org/ID3v1
func MusicID3v1(r io.ReaderAt) string {
	offset := Length(r) - ID3v1Size
	if offset < 0 {
		return ""
	}
	const size = 3
	p := make([]byte, size)
	sr := io.NewSectionReader(r, offset, size)
	if n, err := sr.Read(p); err != nil || n < size {
		return ""
	}
	if !bytes.Equal(p, []byte{'T', 'A', 'G'}) {
		return ""
	}
	const songSize = 30
	p = make([]byte, songSize)
	sr = io.NewSectionReader(r, offset+3, songSize)
	if n, err := sr.Read(p); err != nil || n < size {
		return ""
	}
	song := string(bytes.Trim(p, nul))
	song = strings.TrimSpace(song)
	const artistSize = 30
	p = make([]byte, artistSize)
	sr = io.NewSectionReader(r, offset+33, artistSize)
	if n, err := sr.Read(p); err != nil || n < size {
		return ""
	}
	artist := string(bytes.Trim(p, nul))
	artist = strings.TrimSpace(artist)
	const yearSize = 4
	p = make([]byte, yearSize)
	sr = io.NewSectionReader(r, offset+93, yearSize)
	if n, err := sr.Read(p); err != nil || n < size {
		return ""
	}
	year := string(bytes.Trim(p, nul))
	year = strings.TrimSpace(year)
	s := song
	if artist != "" {
		s += " by " + artist
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
func MusicID3v2(r io.ReaderAt) string {
	const size = 4
	p := make([]byte, size)
	sr := io.NewSectionReader(r, 0, size)
	if n, err := sr.Read(p); err != nil || n < size {
		return ""
	}
	if !bytes.Equal(p[:3], []byte{'I', 'D', '3'}) {
		return ""
	}
	const (
		ver220 = 0x02
		ver230 = 0x03
		ver240 = 0x04
	)
	switch v := p[3]; v {
	case ver220:
		return ID3v220(r)
	case ver230, ver240:
		return ID3v230(r)
	}
	return "d"
}

// ID3v220 reads the [ID3 v2.2] tags in the byte slice and returns the song, artist and year.
// The v2.2 tag is obsolete but still found in the wild.
//
// [ID3 v2.2]: https://id3.org/id3v2-00
func ID3v220(r io.ReaderAt) string {
	const size = 4
	p := make([]byte, size)
	sr := io.NewSectionReader(r, 0, size)
	if n, err := sr.Read(p); err != nil || n < size {
		return ""
	}
	if !bytes.Equal(p[:3], []byte{'I', 'D', '3'}) {
		return ""
	}

	const offset = 6
	p = make([]byte, size)
	sr = io.NewSectionReader(r, offset, size)
	if n, err := sr.Read(p); err != nil || n < size {
		return ""
	}
	tagSize := ConvSize(p)
	p = make([]byte, tagSize)
	sr = io.NewSectionReader(r, 0, tagSize)
	if n, err := sr.Read(p); err != nil || int64(n) < tagSize {
		return ""
	}

	albumTitle := [3]byte{'T', 'A', 'L'}
	leadPerformer := [3]byte{'T', 'P', '1'}
	band := [3]byte{'T', 'P', '2'}
	songName := [3]byte{'T', 'T', '2'}
	year := [3]byte{'T', 'Y', 'E'}
	s := ID3v22Frame(songName, p...)
	if s != "" {
		if lp := ID3v22Frame(leadPerformer, p...); lp != "" {
			s += " by " + lp
		} else if band := ID3v22Frame(band, p...); band != "" {
			s += " by " + band
		}
	} else if ab := ID3v22Frame(albumTitle, p...); ab == "" {
		return ""
	}
	s = strings.TrimSpace(s)
	if y := ID3v22Frame(year, p...); y != "" {
		if _, err := strconv.Atoi(y); err != nil {
			return s
		}
		s += fmt.Sprintf(" (%s)", y)
	}
	return strings.TrimSpace(s)
}

// ID3v230 reads the [ID3 v2.3] and ID3 v2.4 tags in the byte slice and returns the song, artist and year.
// The v2.3 and v2.4 tags are the most common ID3 tags found in MP3 files.
// For our purposes, we treat v2.3 and v2.4 tags the same as there's no difference for the metadata used.
//
// [ID3 v2.3]: https://id3.org/id3v2.3.0
func ID3v230(r io.ReaderAt) string {
	const size = 4
	p := make([]byte, size)
	sr := io.NewSectionReader(r, 0, size)
	if n, err := sr.Read(p); err != nil || n < size {
		return ""
	}
	if !bytes.Equal(p[:3], []byte{'I', 'D', '3'}) {
		return ""
	}

	const offset = 6
	p = make([]byte, size)
	sr = io.NewSectionReader(r, offset, size)
	if n, err := sr.Read(p); err != nil || n < size {
		return ""
	}
	tagSize := ConvSize(p)
	p = make([]byte, tagSize)
	sr = io.NewSectionReader(r, 0, tagSize)
	if n, err := sr.Read(p); err != nil || int64(n) < tagSize {
		return ""
	}

	albumTitle := [4]byte{'T', 'A', 'L', 'B'}
	leadPerformer := [4]byte{'T', 'P', 'E', '1'}
	contentGroup := [4]byte{'T', 'I', 'T', '1'}
	songName := [4]byte{'T', 'I', 'T', '2'}
	year := [4]byte{'T', 'Y', 'E', 'R'}
	s := ID3v23Frame(songName, p...)
	if s != "" {
		if lp := ID3v23Frame(leadPerformer, p...); lp != "" {
			s += " by " + lp
		} else if cg := ID3v23Frame(contentGroup, p...); cg != "" {
			s += " by " + cg
		}
	} else if ab := ID3v23Frame(albumTitle, p...); ab == "" {
		return ""
	}
	s = strings.TrimSpace(s)
	if y := ID3v23Frame(year, p...); y != "" {
		if _, err := strconv.Atoi(y); err != nil {
			return s
		}
		s += fmt.Sprintf(" (%s)", y)
	}
	return strings.TrimSpace(s)
}

// ID3v22Frame reads the ID3 v2.2 frame in the byte slice and returns the frame data as a string.
// The frame header contains a 3 byte identifier followed by a 3 byte size.
func ID3v22Frame(id [3]byte, data ...byte) string {
	const header, size = 6, 3
	frameID := []byte{id[0], id[1], id[2]}
	return id3Frame(frameID, header, size, data...)
}

// ID3v23Frame reads the ID3 v2.3 and v2.4 frame in the byte slice and returns the frame data as a string.
// The frame header contains a 4 byte identifier followed by a 4 byte size.
func ID3v23Frame(id [4]byte, data ...byte) string {
	const header, size = 10, 4
	frameID := []byte{id[0], id[1], id[2], id[3]}
	return id3Frame(frameID, header, size, data...)
}

func id3Frame(frameID []byte, header, size int, data ...byte) string {
	offset := bytes.Index(data, frameID)
	if offset == -1 {
		return ""
	}
	sizeIndex := offset + size
	sizeData := data[sizeIndex : sizeIndex+size]
	length := ConvSize(sizeData)
	b := bytes.Trim(data[offset+header:int64(offset+header)+length], nul)
	s, _ := ConvLatin1(b)
	return strings.TrimSpace(s)
}

// ConvLatin1 converts a byte slice to a Latin-1 (ISO-8859-1) string.
func ConvLatin1(p []byte) (string, error) {
	decoder := charmap.ISO8859_1.NewDecoder()
	s, err := decoder.Bytes(p)
	if err != nil {
		return "", fmt.Errorf("magicnumber iso 8859-1 decoder: %w", err)
	}
	return string(s), nil
}

func ConvSize(p []byte) int64 {
	const synchSafeSizeBase = 7
	sizeBase := uint(synchSafeSizeBase)
	var size int64
	for _, b := range p {
		if b&128 > 0 {
			return 0
		}
		size = (size << sizeBase) | int64(b)
	}
	return size
}
