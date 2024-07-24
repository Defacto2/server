// Package render provides the file content rendering for the web server.
package render

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/Defacto2/server/internal/ext"
	"github.com/Defacto2/server/internal/helper"
	"github.com/Defacto2/server/internal/postgres/models"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/encoding/unicode"
)

var (
	ErrDownload  = errors.New("download file cannot be stat")
	ErrFileModel = errors.New("file model is nil")
	ErrFilename  = errors.New("file model filename is empty")
	ErrUUID      = errors.New("file model uuid is empty")
)

const textamiga = "textamiga"

// Encoder returns the encoding for the model file entry.
// Based on the platform and section.
// Otherwise it will attempt to determine the encoding from the file byte content.
func Encoder(art *models.File, r io.Reader) encoding.Encoding {
	if art == nil {
		return nil
	}
	platform := strings.ToLower(strings.TrimSpace(art.Platform.String))
	section := strings.ToLower(strings.TrimSpace(art.Section.String))
	switch platform {
	case textamiga:
		return charmap.ISO8859_1
	default:
		switch section {
		case "appleii", "atarist":
			return charmap.ISO8859_1
		}
	}
	magic := strings.ToLower(strings.TrimSpace(art.FileMagicType.String))
	if strings.Contains(magic, "utf-8") {
		return unicode.UTF8
	}
	guess := helper.Determine(r)
	return guess
}

// Read returns the content of either the file download or an extracted text file.
// The text is intended to be used as a readme, preview or an in-browser viewer.
func Read(art *models.File, downloadDir, extraDir string) ([]byte, error) {
	if art == nil {
		return nil, ErrFileModel
	}

	fname := art.Filename.String
	unid := art.UUID.String

	if fname == "" {
		return nil, ErrFilename
	}
	if unid == "" {
		return nil, ErrUUID
	}

	var files struct {
		uuidTxt string
		filepth string
		uutxtOk bool
		filepOk bool
		txt     bool
	}
	files.uuidTxt = filepath.Join(extraDir, unid+".txt")
	files.uutxtOk = helper.Stat(files.uuidTxt)
	files.filepth = filepath.Join(downloadDir, unid)
	files.filepOk = helper.Stat(files.filepth)
	files.txt = !ext.IsArchive(fname)

	if !files.uutxtOk && !files.filepOk {
		return nil, fmt.Errorf("render read %w: %s", ErrDownload, filepath.Join(downloadDir, unid))
	}

	if !files.uutxtOk && !Viewer(art) {
		doNothing := []byte{}
		return doNothing, nil
	}

	name := files.filepth
	if files.uutxtOk {
		name = files.uuidTxt
	}

	b, err := os.ReadFile(name)
	if err != nil {
		return nil, fmt.Errorf("render read file %w", err)
	}

	const nul = 0x00
	b = bytes.ReplaceAll(b, []byte{nul}, []byte(" "))
	return b, nil
}

// Viewer returns true if the file entry should display the file download in the browser plain text viewer.
func Viewer(art *models.File) bool {
	if art == nil {
		return false
	}
	platform := strings.ToLower(strings.TrimSpace(art.Platform.String))
	switch platform {
	case "text", textamiga:
		return true
	}
	return false
}

// NoScreenshot returns true when the file entry should not attempt to display a screenshot.
// This is based on the platform, section or if the screenshot is missing on the server.
func NoScreenshot(art *models.File, downloadPath, previewPath string) bool {
	if art == nil {
		return true
	}
	platform := strings.ToLower(strings.TrimSpace(art.Platform.String))
	switch platform {
	case textamiga, "text":
		return true
	}
	unid := art.UUID.String
	webp := strings.Join([]string{previewPath, unid + ".webp"}, "/")
	png := strings.Join([]string{previewPath, unid + ".png"}, "/")
	if helper.Stat(webp) || helper.Stat(png) {
		return false
	}
	return true
}

// UTF16 returns true if the byte slice is embedded with a UTF-16 BOM (byte order mark).
func UTF16(r io.Reader) bool {
	if r == nil {
		return false
	}
	const minimum = 2
	p := make([]byte, minimum)
	if _, err := io.ReadFull(r, p); err != nil {
		return false
	}
	if len(p) < minimum {
		return false
	}
	const y, thorn = 0xff, 0xfe
	littleEndian := p[0] == y && p[1] == thorn
	if littleEndian {
		return true
	}
	bigEndian := p[0] == thorn && p[1] == y
	return bigEndian
}
