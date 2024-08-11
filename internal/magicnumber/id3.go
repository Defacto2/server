package magicnumber

import (
	"bytes"
	"fmt"
	"strings"
)

// Package file id3.go contains the functions that parse bytes as common ID3 tag formats usually found in MP3 files.

// MusicID3v1 reads the [ID3 v1] tag in the byte slice and returns the title and artist.
// The ID3 v1 tag is a 128 byte tag at the end of an MP3 audio file.
//
// [ID3 v1]: http://id3.org/ID3v1
func MusicID3v1(p []byte) string {
	const length = 128
	if len(p) < length {
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

func MusicID3v2(p []byte) string {
	const length = 10
	if len(p) < length {
		return ""
	}
	if !bytes.Equal(p[0:3], []byte{'I', 'D', '3'}) {
		return ""
	}
	ha := int(p[6]) * 2097152
	hb := int(p[7]) * 16384
	hc := int(p[8]) * 128
	hd := int(p[9])
	headerSize := ha + hb + hc + hd
	fmt.Println(headerSize, "header size", headerSize, "length", len(p))
	if headerSize+length > len(p) {
		return ""
	}
	data := p[length : headerSize+length]
	fmt.Println(headerSize, "header size", headerSize)

	// frame id 4 bytes
	// size 4 bytes
	// flags 2 bytes
	talb := bytes.Index(data, []byte{'T', 'A', 'L', 'B'})
	fmt.Println("talb", talb)
	// if talb != -1 && talb+10 < len(data) {
	// 	//tablSize := binary.LittleEndian.Uint32(data[talb+4 : talb+8])
	// 	b0 := int(data[talb+4]) * 2097152
	// 	b1 := int(data[talb+5]) * 16384
	// 	b2 := int(data[talb+6]) * 128
	// 	b3 := int(data[talb+7])
	// 	tablSize := b0 + b1 + b2 + b3
	// 	fmt.Println("tabl size u32", tablSize)
	// 	title := string(data[talb+10 : talb+10+tablSize])
	// 	fmt.Println("title", title)
	// }
	// h0 := p[6]*2 ^ 21
	// h1 := p[7]*2 ^ 14
	// h2 := p[8]*2 ^ 7
	// headerSize := int(h0) + int(h1) + int(h2) + int(p[9])
	//b := p[0x06:0xa]
	//
	//headerSize := binary.LittleEndian.Uint16(b)
	// An easy way of calculating the tag size is
	// A*2^21+B*2^14+C*2^7+D = A*2097152+B*16384+C*128+D,
	//where A is the first byte, B the second, C the third and D the fourth byte.
	tabl := [4]byte{'T', 'A', 'L', 'B'}
	fmt.Println("TABL", ID3v3Frame(tabl, data...))
	tpe1 := [4]byte{'T', 'P', 'E', '1'}
	fmt.Println("TPE1", ID3v3Frame(tpe1, data...))
	tit1 := [4]byte{'T', 'I', 'T', '1'}
	fmt.Println("TIT1", ID3v3Frame(tit1, data...))
	tit2 := [4]byte{'T', 'I', 'T', '2'}
	fmt.Println("TIT2", ID3v3Frame(tit2, data...))
	tyer := [4]byte{'T', 'Y', 'E', 'R'}
	fmt.Println("TYER", ID3v3Frame(tyer, data...))

	fmt.Println(headerSize, "header size", p[6], p[7], p[8], p[9])
	// lookup version
	version := p[3]
	switch version {
	case 0x2:
		return "ID3v2.2"
	case 0x3:
		return "ID3v2.3"
	case 0x4:
		return "ID3v2.4"
	}
	// 0 * 2^21 = 0
	// 0 * 2^14 = 0
	// 2 * 2^7 = 256
	// 1 = 1
	return ""
}

func ID3v3Frame(id [4]byte, data ...byte) string {
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
