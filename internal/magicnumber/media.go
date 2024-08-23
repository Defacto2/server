package magicnumber

import (
	"bytes"
	"encoding/hex"
	"io"
	"strconv"
)

// Package file media.go contains the functions that parse bytes as commom image, digital audio and video formats.
// A number of these media containers could support multiple modes,
// such as audio only, audio+video, video only, static images, animated images, etc.

// AAC matches the Advanced Audio Coding audio format in the byte slice.
func AAC(p []byte) bool {
	const twoBytes = 2
	const min = twoBytes + 1
	if len(p) < min {
		return false
	}
	if !bytes.Equal(p[:twoBytes], []byte{0xff, 0xfb}) {
		return false
	}
	const x90, xb0, xe0 = 0x90, 0xb0, 0xe0
	switch p[twoBytes] {
	case x90, xb0, xe0:
		return true
	default:
		return false
	}
}

// Avi matches the Microsoft Audio Video Interleave video format in the byte slice.
func Avi(p []byte) bool {
	const min = 16
	if len(p) < min {
		return false
	}
	return bytes.Equal(p[:4], []byte{'R', 'I', 'F', 'F'}) &&
		bytes.Equal(p[8:16], []byte{'A', 'V', 'I', 0x20, 'L', 'I', 'S', 'T'})
}

// Avif matches the AV1 Image File image format in the byte slice, also known as AVIF.
// This is a new image format based on the AV1 video codec from the Alliance for Open Media.
// But the detection method is not accurate and should be used as a hint.
func Avif(p []byte) bool {
	const min, offset = 8, 4
	if len(p) < min+offset {
		return false
	}
	// Gary Kessler's File Signatures suggests the AVIF image format is 0x0A 0x00 0x00
	// but this maybe out dated and definitely causes false positives.
	// According to the AV1 Image File Format specification there is no magic number.
	// https://aomediacodec.github.io/av1-avif/v1.0.0.html
	//
	// As a workaround, we detect the AVIF image format by checking for the 'ftypavif' string.
	// 'ftyp' matches the HEIF container and 'avif' is the brand.
	return bytes.Equal(p[offset:min+offset], []byte{'f', 't', 'y', 'p', 'a', 'v', 'i', 'f'})
}

// Bmp matches the BMP image format in the byte slice.
func Bmp(p []byte) bool {
	const min = 2
	if len(p) < min {
		return false
	}
	return bytes.Equal(p[:min], []byte{'B', 'M'})
}

// Flac matches the Free Lossless Audio Codec audio format in the byte slice.
func Flac(p []byte) bool {
	const min = 8
	if len(p) < min {
		return false
	}
	return bytes.Equal(p[:min], []byte{'f', 'L', 'a', 'C', 0x0, 0x0, 0x0, 0x2})
}

// Flv matches the Shockwave Flash Video format in the byte slice.
func Flv(p []byte) bool {
	const min = 4
	if len(p) < min {
		return false
	}
	return bytes.Equal(p[:min], []byte{'F', 'L', 'V', 0x1})
}

// Gif matches the image Graphics Interchange Format in the byte slice.
// There are two versions of the GIF format, GIF87a and GIF89a.
func Gif(p []byte) bool {
	const min = 6
	if len(p) < min {
		return false
	}
	gif87a := []byte{0x47, 0x49, 0x46, 0x38, 0x37, 0x61}
	gif89a := []byte{0x47, 0x49, 0x46, 0x38, 0x39, 0x61}
	return bytes.Equal(p[:min], gif87a) ||
		bytes.Equal(p[:min], gif89a)
}

// Ico matches the Microsoft Icon image format in the byte slice.
func Ico(p []byte) bool {
	const min = 4
	if len(p) < min {
		return false
	}
	return bytes.Equal(p[:min], []byte{0x0, 0x0, 0x1, 0x0})
}

// Iff matches the Interchange File Format image in the byte slice.
// This is a generic wrapper format originally created by Electronic Arts
// for storing data in chunks.
func Iff(p []byte) bool {
	const min = 4
	if len(p) < min {
		return false
	}
	return bytes.Equal(p[:4], []byte{'C', 'A', 'T', 0x20})
}

// Ivr matches the RealPlayer video format in the byte slice.
func Ivr(p []byte) bool {
	const min = 4
	if len(p) < min {
		return false
	}
	return bytes.Equal(p[:min], []byte{0x2e, 'R', 'E', 'C'}) ||
		bytes.Equal(p[:min], []byte{0x2e, 'R', 'M', 'F'})
}

// Jpeg matches the JPEG File Interchange Format v1 image in the byte slice.
func Jpeg(p []byte) bool {
	return jpeg(p, true)
}

// JpegNoSuffix matches the JPEG File Interchange Format v1 image in the byte slice.
// This is a less accurate method than Jpeg as it does not check the final bytes.
func JpegNoSuffix(p []byte) bool {
	return jpeg(p, false)
}

func jpeg(p []byte, suffix bool) bool {
	const min = 11
	if len(p) < min {
		return false
	}
	if !bytes.Equal(p[:3], []byte{0xff, 0xd8, 0xff}) {
		return false
	}
	if p[3] != 0xe0 && p[3] != 0xe1 {
		return false
	}
	if !bytes.Equal(p[6:11], []byte{'J', 'F', 'I', 'F', 0x0}) &&
		!bytes.Equal(p[6:11], []byte{'E', 'x', 'i', 'f', 0x0}) {
		return false
	}
	if !suffix {
		return true
	}
	return bytes.HasSuffix(p, []byte{0xff, 0xd9})
}

// Jpeg2000 matches the JPEG 2000 image format in the byte slice.
func Jpeg2000(p []byte) bool {
	const min = 10
	if len(p) < min {
		return false
	}
	return bytes.Equal(p[:min], []byte{0x0, 0x0, 0x0, 0xc, 0x6a, 0x50, 0x20, 0x20, 0xd, 0xa})
}

// Ilbm matches the InterLeaved Bitmap image format in the byte slice.
// Created by Electronic Arts it conforms to the IFF standard.
func Ilbm(p []byte) bool {
	const min = 12
	if len(p) < min {
		return false
	}
	return bytes.Equal(p[:4], []byte{'F', 'O', 'R', 'M'}) &&
		bytes.Equal(p[8:12], []byte{'I', 'L', 'B', 'M'})
}

// IlbmDecode reads the InterLeaved Bitmap image format in the reader and returns the width and height.
func IlbmDecode(r io.Reader) (int, int) {
	const min = 24
	p := make([]byte, min)
	if _, err := io.ReadFull(r, p); err != nil {
		return 0, 0
	}
	return IlbmConfig(p)
}

// IlbmConfig reads the InterLeaved Bitmap image format in the byte slice and returns the width and height.
func IlbmConfig(p []byte) (int, int) {
	const min = 24
	if len(p) < min {
		return 0, 0
	}
	hw := hex.EncodeToString([]byte{p[20], p[21]})
	hh := hex.EncodeToString([]byte{p[22], p[23]})
	w, err := strconv.ParseInt(hw, 16, 64)
	if err != nil {
		return 0, 0
	}
	h, err := strconv.ParseInt(hh, 16, 64)
	if err != nil {
		return 0, 0
	}
	return int(w), int(h)
}

// M4v matches the QuickTime M4V video format in the byte slice.
func M4v(p []byte) bool {
	const min, offset = 8, 4
	if len(p) < min+offset {
		return false
	}
	return bytes.Equal(p[offset:min+offset], []byte{'f', 't', 'y', 'p', 'm', 'p', '4', '2'})
}

// Mp4 matches the MPEG-4 video format in the byte slice.
func Mp4(p []byte) bool {
	const min, offset = 8, 4
	if len(p) < min+offset {
		return false
	}
	ftypMSNV := []byte{'f', 't', 'y', 'p', 'M', 'S', 'N', 'V'}
	ftypisom := []byte{'f', 't', 'y', 'p', 'i', 's', 'o', 'm'}
	switch {
	case bytes.Equal(p[offset:min+offset], ftypMSNV):
		return true
	case bytes.Equal(p[offset:min+offset], ftypisom):
		return true
	default:
		return false
	}
}

// Mp3 matches the MPEG-1 Audio Layer 3 audio format in the byte slice.
// This only checks for the ID3v2 tag and not the audio data.
// Songs with no ID3v2 tag will not be detected including files with ID3v1 tags.
func Mp3(p []byte) bool {
	const min = 3
	if len(p) < min {
		return false
	}
	return bytes.Equal(p[:min], []byte{'I', 'D', '3'})
}

// Mpeg matches the MPEG video format in the byte slice.
func Mpeg(p []byte) bool {
	const min = 4
	if len(p) < min {
		return false
	}
	return bytes.Equal(p[:3], []byte{0x0, 0x0, 0x1}) && p[4] >= 0xba && p[4] <= 0xbf
}

// Ogg matches the Ogg Vorbis audio format in the byte slice.
func Ogg(p []byte) bool {
	b := []byte{
		'O', 'g', 'g', 'S', 0x0, 0x2, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	}
	min := len(b)
	if len(p) < min {
		return false
	}
	return bytes.Equal(p[:min], b)
}

// Pcx matches the Personal Computer eXchange image format in the byte slice.
func Pcx(p []byte) bool {
	const min = 3
	if len(p) < min {
		return false
	}
	id := p[0]
	ver := p[1] // version of PCX v0 through to v5
	enc := p[2] // encoding (0 = uncompressed, 1 = run-length encoding compressed)
	return id == 0x0a && ver <= 0x5 && (enc == 0x0 || enc == 0x1)
}

// Png matches the Portable Network Graphics image format in the byte slice.
func Png(p []byte) bool {
	const min = 8
	if len(p) < min {
		return false
	}
	return bytes.Equal(p[:min], []byte{0x89, 0x50, 0x4E, 0x47, 0x0d, 0x0a, 0x1a, 0x0a})
}

// QTMov matches the QuickTime Movie video format in the byte slice.
func QTMov(p []byte) bool {
	const min = 8
	if len(p) < min {
		return false
	}
	const offset = 4
	return bytes.Equal(p[offset:8], []byte{'m', 'o', 'o', 'v'}) ||
		bytes.Equal(p[offset:10], []byte{'f', 't', 'y', 'p', 'q', 't'})
}

// Ripscrip returns true if the reader contains the RIPscrip signature.
// This is a vector graphics format used in BBS systems in the early 1990s.
func Ripscrip(p []byte) bool {
	const min = 3
	if len(p) < min {
		return false
	}
	head := bytes.Equal(p[:2], []byte{'!', '|'})
	if !head {
		return false
	}
	i := p[2]
	return i >= '0' && i <= '9'
}

// Tiff matches the Tagged Image File Format in the byte slice.
func Tiff(p []byte) bool {
	const min = 4
	if len(p) < min {
		return false
	}
	le := []byte{0x49, 0x49, 0x2a, 0x0}
	be := []byte{0x4d, 0x4d, 0x0, 0x2a}
	return bytes.Equal(p[:min], le) ||
		bytes.Equal(p[:min], be)
}

// Wave matches the IBM / Microsoft Waveform audio format in the byte slice.
func Wave(p []byte) bool {
	const min = 12
	if len(p) < min {
		return false
	}
	return bytes.Equal(p[:4], []byte{'R', 'I', 'F', 'F'}) &&
		bytes.Equal(p[8:16], []byte{'W', 'A', 'V', 'E', 'f', 'm', 't', 0x20})
}

// Webp matches the Google WebP image format in the byte slice.
func Webp(p []byte) bool {
	const min = 12
	if len(p) < min {
		return false
	}
	return bytes.Equal(p[:4], []byte{'R', 'I', 'F', 'F'}) &&
		bytes.Equal(p[8:12], []byte{'W', 'E', 'B', 'P'})
}

// Wmv matches the Microsoft Windows Media video format in the byte slice.
func Wmv(p []byte) bool {
	const min = 16
	if len(p) < min {
		return false
	}
	return bytes.Equal(p[:min],
		[]byte{
			0x30, 0x26, 0xb2, 0x75, 0x8e, 0x66, 0xcf, 0x11,
			0xa6, 0xd9, 0x0, 0xaa, 0x0, 0x62, 0xce, 0x6c,
		})
}
