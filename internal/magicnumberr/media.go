package magicnumberr

import (
	"bytes"
	"encoding/hex"
	"io"
	"strconv"
)

// Package file media.go contains the functions that parse bytes as commom image, digital audio and video formats.
// A number of these media containers could support multiple modes,
// such as audio only, audio+video, video only, static images, animated images, etc.

// AAC matches the Advanced Audio Coding audio format.
func AAC(r io.ReaderAt) bool {
	const size = 3
	p := make([]byte, size)
	sr := io.NewSectionReader(r, 0, size)
	if n, err := sr.Read(p); err != nil || n < size {
		return false
	}
	if !bytes.Equal(p[:2], []byte{0xff, 0xfb}) {
		return false
	}
	const x90, xb0, xe0 = 0x90, 0xb0, 0xe0
	switch p[2] {
	case x90, xb0, xe0:
		return true
	default:
		return false
	}
}

// Avi matches the Microsoft Audio Video Interleave video format.
func Avi(r io.ReaderAt) bool {
	if !RIFF(r) {
		return false
	}
	const offset, size = 8, 8
	p := make([]byte, size)
	sr := io.NewSectionReader(r, offset, size)
	if n, err := sr.Read(p); err != nil || n < size {
		return false
	}
	return bytes.Equal(p, []byte{'A', 'V', 'I', 0x20, 'L', 'I', 'S', 'T'})
}

// Avif matches the AV1 Image File image format in the byte slice, also known as AVIF.
// This is a new image format based on the AV1 video codec from the Alliance for Open Media.
// But the detection method is not accurate and should be used as a hint.
func Avif(r io.ReaderAt) bool {
	const size, offset = 8, 4
	p := make([]byte, size)
	sr := io.NewSectionReader(r, offset, size)
	if n, err := sr.Read(p); err != nil || n < size {
		return false
	}
	// Gary Kessler's File Signatures suggests the AVIF image format is 0x0A 0x00 0x00
	// but this maybe out dated and definitely causes false positives.
	// According to the AV1 Image File Format specification there is no magic number.
	// https://aomediacodec.github.io/av1-avif/v1.0.0.html
	//
	// As a workaround, we detect the AVIF image format by checking for the 'ftypavif' string.
	// 'ftyp' matches the HEIF container and 'avif' is the brand.
	return bytes.Equal(p, []byte{'f', 't', 'y', 'p', 'a', 'v', 'i', 'f'})
}

// Bmp matches the BMP image format.
func Bmp(r io.ReaderAt) bool {
	const size = 2
	p := make([]byte, size)
	sr := io.NewSectionReader(r, 0, size)
	if n, err := sr.Read(p); err != nil || n < size {
		return false
	}
	return bytes.Equal(p, []byte{'B', 'M'})
}

// Flac matches the Free Lossless Audio Codec audio format.
func Flac(r io.ReaderAt) bool {
	const size = 8
	p := make([]byte, size)
	sr := io.NewSectionReader(r, 0, size)
	if n, err := sr.Read(p); err != nil || n < size {
		return false
	}
	return bytes.Equal(p, []byte{'f', 'L', 'a', 'C', 0x0, 0x0, 0x0, 0x2})
}

// Flv matches the Shockwave Flash Video format.
func Flv(r io.ReaderAt) bool {
	const size = 4
	p := make([]byte, size)
	sr := io.NewSectionReader(r, 0, size)
	if n, err := sr.Read(p); err != nil || n < size {
		return false
	}
	return bytes.Equal(p, []byte{'F', 'L', 'V', 0x1})
}

// Gif matches the image Graphics Interchange Format.
// There are two versions of the GIF format, GIF87a and GIF89a.
func Gif(r io.ReaderAt) bool {
	const size = 6
	p := make([]byte, size)
	sr := io.NewSectionReader(r, 0, size)
	if n, err := sr.Read(p); err != nil || n < size {
		return false
	}
	gif87a := []byte{0x47, 0x49, 0x46, 0x38, 0x37, 0x61}
	gif89a := []byte{0x47, 0x49, 0x46, 0x38, 0x39, 0x61}
	return bytes.Equal(p, gif87a) ||
		bytes.Equal(p, gif89a)
}

// Ico matches the Microsoft Icon image format.
func Ico(r io.ReaderAt) bool {
	const size = 4
	p := make([]byte, size)
	sr := io.NewSectionReader(r, 0, size)
	if n, err := sr.Read(p); err != nil || n < size {
		return false
	}
	return bytes.Equal(p, []byte{0x0, 0x0, 0x1, 0x0})
}

// Iff matches the Interchange File Format image.
// This is a generic wrapper format originally created by Electronic Arts
// for storing data in chunks.
func Iff(r io.ReaderAt) bool {
	const size = 4
	p := make([]byte, size)
	sr := io.NewSectionReader(r, 0, size)
	if n, err := sr.Read(p); err != nil || n < size {
		return false
	}
	return bytes.Equal(p, []byte{'C', 'A', 'T', 0x20})
}

// Ivr matches the RealPlayer video format.
func Ivr(r io.ReaderAt) bool {
	const size = 4
	p := make([]byte, size)
	sr := io.NewSectionReader(r, 0, size)
	if n, err := sr.Read(p); err != nil || n < size {
		return false
	}
	const fullstop = 0x2e
	return bytes.Equal(p, []byte{fullstop, 'R', 'E', 'C'}) ||
		bytes.Equal(p, []byte{fullstop, 'R', 'M', 'F'})
}

// Jpeg matches the JPEG File Interchange Format v1 image.
func Jpeg(r io.ReaderAt) bool {
	return jpeg(r, true)
}

// JpegNoSuffix matches the JPEG File Interchange Format v1 image.
// This is a less accurate method than Jpeg as it does not check the final bytes.
func JpegNoSuffix(r io.ReaderAt) bool {
	return jpeg(r, false)
}

func jpeg(r io.ReaderAt, suffix bool) bool {
	const size = 3
	p := make([]byte, size)
	sr := io.NewSectionReader(r, 0, size)
	if n, err := sr.Read(p); err != nil || n < size {
		return false
	}
	if !bytes.Equal(p, []byte{0xff, 0xd8, 0xff}) {
		return false
	}
	p = make([]byte, 1)
	sr = io.NewSectionReader(r, 3, 1)
	if n, err := sr.Read(p); err != nil || n < 1 {
		return false
	}
	if p[0] != 0xe0 && p[0] != 0xe1 {
		return false
	}
	const jsize = 5
	p = make([]byte, jsize)
	sr = io.NewSectionReader(r, 6, jsize)
	if n, err := sr.Read(p); err != nil || n < jsize {
		return false
	}
	if !bytes.Equal(p, []byte{'J', 'F', 'I', 'F', 0x0}) &&
		!bytes.Equal(p, []byte{'E', 'x', 'i', 'f', 0x0}) {
		return false
	}
	if !suffix {
		return true
	}
	length := Length(r)
	const sufSize = int64(2)
	offset := length - sufSize
	p = make([]byte, sufSize)
	sr = io.NewSectionReader(r, offset, sufSize)
	if n, err := sr.Read(p); err != nil || int64(n) < sufSize {
		return false
	}
	return bytes.HasSuffix(p, []byte{0xff, 0xd9})
}

// Jpeg2000 matches the JPEG 2000 image format.
func Jpeg2000(r io.ReaderAt) bool {
	const size = 10
	p := make([]byte, size)
	sr := io.NewSectionReader(r, 0, size)
	if n, err := sr.Read(p); err != nil || n < size {
		return false
	}
	return bytes.Equal(p, []byte{0x0, 0x0, 0x0, 0xc, 0x6a, 0x50, 0x20, 0x20, 0xd, 0xa})
}

// Ilbm matches the InterLeaved Bitmap image format.
// Created by Electronic Arts it conforms to the IFF standard.
func Ilbm(r io.ReaderAt) bool {
	const size = 4
	p := make([]byte, size)
	sr := io.NewSectionReader(r, 0, size)
	if n, err := sr.Read(p); err != nil || n < size {
		return false
	}
	if !bytes.Equal(p, []byte{'F', 'O', 'R', 'M'}) {
		return false
	}
	const offset = 8
	p = make([]byte, size)
	sr = io.NewSectionReader(r, offset, size)
	if n, err := sr.Read(p); err != nil || n < size {
		return false
	}
	return bytes.Equal(p, []byte{'I', 'L', 'B', 'M'})
}

// IlbmDecode reads the InterLeaved Bitmap image format in the reader and returns the width and height.
func IlbmDecode(r io.ReaderAt) (int, int) {
	const offset, size = 20, 4
	p := make([]byte, size)
	sr := io.NewSectionReader(r, offset, size)
	if n, err := sr.Read(p); err != nil || n < size {
		return 0, 0
	}
	hw := hex.EncodeToString([]byte{p[0], p[1]})
	hh := hex.EncodeToString([]byte{p[2], p[3]})
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

// M4v matches the QuickTime M4V video format.
func M4v(r io.ReaderAt) bool {
	const offset, size = 4, 8
	p := make([]byte, size)
	sr := io.NewSectionReader(r, offset, size)
	if n, err := sr.Read(p); err != nil || n < size {
		return false
	}
	return bytes.Equal(p, []byte{'f', 't', 'y', 'p', 'm', 'p', '4', '2'})
}

// Mp4 matches the MPEG-4 video format.
func Mp4(r io.ReaderAt) bool {
	const offset, size = 4, 8
	p := make([]byte, size)
	sr := io.NewSectionReader(r, offset, size)
	if n, err := sr.Read(p); err != nil || n < size {
		return false
	}
	ftypMSNV := []byte{'f', 't', 'y', 'p', 'M', 'S', 'N', 'V'}
	ftypisom := []byte{'f', 't', 'y', 'p', 'i', 's', 'o', 'm'}
	return bytes.Equal(p, ftypMSNV) ||
		bytes.Equal(p, ftypisom)
}

// Mp3 matches the MPEG-1 Audio Layer 3 audio format.
// This only checks for the ID3v2 tag and not the audio data.
// Songs with no ID3v2 tag will not be detected including files with ID3v1 tags.
func Mp3(r io.ReaderAt) bool {
	const size = 3
	p := make([]byte, size)
	sr := io.NewSectionReader(r, 0, size)
	if n, err := sr.Read(p); err != nil || n < size {
		return false
	}
	return bytes.Equal(p, []byte{'I', 'D', '3'})
}

// Mpeg matches the MPEG video format.
func Mpeg(r io.ReaderAt) bool {
	const size = 4
	p := make([]byte, size)
	sr := io.NewSectionReader(r, 0, size)
	if n, err := sr.Read(p); err != nil || n < size {
		return false
	}
	return bytes.Equal(p[:3], []byte{0x0, 0x0, 0x1}) && p[3] >= 0xba && p[3] <= 0xbf
}

// Ogg matches the Ogg Vorbis audio format.
func Ogg(r io.ReaderAt) bool {
	const size = 14
	oggs := []byte{
		'O', 'g', 'g', 'S',
		0x0, 0x2, 0x0, 0x0,
		0x0, 0x0, 0x0, 0x0,
		0x0, 0x0,
	}
	p := make([]byte, size)
	sr := io.NewSectionReader(r, 0, size)
	if n, err := sr.Read(p); err != nil || n < size {
		return false
	}
	return bytes.Equal(p, oggs)
}

// Pcx matches the Personal Computer eXchange image format.
func Pcx(r io.ReaderAt) bool {
	const size = 3
	p := make([]byte, size)
	sr := io.NewSectionReader(r, 0, size)
	if n, err := sr.Read(p); err != nil || n < size {
		return false
	}
	id := p[0]
	ver := p[1] // version of PCX v0 through to v5
	enc := p[2] // encoding (0 = uncompressed, 1 = run-length encoding compressed)
	return id == 0x0a && ver <= 0x5 && (enc == 0x0 || enc == 0x1)
}

// Png matches the Portable Network Graphics image format.
func Png(r io.ReaderAt) bool {
	const size = 8
	p := make([]byte, size)
	sr := io.NewSectionReader(r, 0, size)
	if n, err := sr.Read(p); err != nil || n < size {
		return false
	}
	return bytes.Equal(p, []byte{0x89, 0x50, 0x4E, 0x47, 0x0d, 0x0a, 0x1a, 0x0a})
}

// QTMov matches the QuickTime Movie video format.
func QTMov(r io.ReaderAt) bool {
	const offset, size = 4, 10
	p := make([]byte, size)
	sr := io.NewSectionReader(r, offset, size)
	if n, err := sr.Read(p); err != nil || n < size {
		return false
	}
	return bytes.Equal(p[:4], []byte{'m', 'o', 'o', 'v'}) ||
		bytes.Equal(p, []byte{'f', 't', 'y', 'p', 'q', 't'})
}

func RIFF(r io.ReaderAt) bool {
	const size = 4
	p := make([]byte, size)
	sr := io.NewSectionReader(r, 0, size)
	if n, err := sr.Read(p); err != nil || n < size {
		return false
	}
	return bytes.Equal(p, []byte{'R', 'I', 'F', 'F'})
}

// Ripscrip returns true if the reader contains the RIPscrip signature.
// This is a vector graphics format used in BBS systems in the early 1990s.
func Ripscrip(r io.ReaderAt) bool {
	const size = 3
	p := make([]byte, size)
	sr := io.NewSectionReader(r, 0, size)
	if n, err := sr.Read(p); err != nil || n < size {
		return false
	}
	if head := bytes.Equal(p[:2], []byte{'!', '|'}); !head {
		return false
	}
	i := p[2]
	return i >= '0' && i <= '9'
}

// Tiff matches the Tagged Image File Format.
func Tiff(r io.ReaderAt) bool {
	const size = 4
	p := make([]byte, size)
	sr := io.NewSectionReader(r, 0, size)
	if n, err := sr.Read(p); err != nil || n < size {
		return false
	}
	le := []byte{0x49, 0x49, 0x2a, 0x00}
	be := []byte{0x4d, 0x4d, 0x00, 0x2a}
	return bytes.Equal(p, le) || bytes.Equal(p, be)
}

// Wave matches the IBM / Microsoft Waveform audio format.
func Wave(r io.ReaderAt) bool {
	if !RIFF(r) {
		return false
	}
	const offset, size = 8, 8
	p := make([]byte, size)
	sr := io.NewSectionReader(r, offset, size)
	if n, err := sr.Read(p); err != nil || n < size {
		return false
	}
	return bytes.Equal(p, []byte{'W', 'A', 'V', 'E', 'f', 'm', 't', 0x20})
}

// Webp matches the Google WebP image format.
func Webp(r io.ReaderAt) bool {
	if !RIFF(r) {
		return false
	}
	const offset, size = 8, 4
	p := make([]byte, size)
	sr := io.NewSectionReader(r, offset, size)
	if n, err := sr.Read(p); err != nil || n < size {
		return false
	}
	return bytes.Equal(p, []byte{'W', 'E', 'B', 'P'})
}

// Wmv matches the Microsoft Windows Media video format.
func Wmv(r io.ReaderAt) bool {
	const size = 16
	p := make([]byte, size)
	sr := io.NewSectionReader(r, 0, size)
	if n, err := sr.Read(p); err != nil || n < size {
		return false
	}
	return bytes.Equal(p, []byte{
		0x30, 0x26, 0xb2, 0x75, 0x8e, 0x66, 0xcf, 0x11,
		0xa6, 0xd9, 0x00, 0xaa, 0x00, 0x62, 0xce, 0x6c,
	})
}
