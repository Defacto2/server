package render

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Defacto2/server/internal/exts"
	"github.com/Defacto2/server/internal/helper"
	"github.com/Defacto2/server/internal/postgres/models"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"
)

var (
	ErrDownload  = errors.New("download file cannot be stat")
	ErrFileModel = errors.New("file model is nil")
)

const textamiga = "textamiga"

// Encoder returns the encoding for the model file entry.
// Based on the platform and section.
// Otherwise it will attempt to determine the encoding from the file byte content.
func Encoder(res *models.File, b ...byte) encoding.Encoding {
	if res == nil {
		return nil
	}

	platform := strings.ToLower(strings.TrimSpace(res.Platform.String))
	section := strings.ToLower(strings.TrimSpace(res.Section.String))

	switch platform {
	case textamiga:
		return charmap.ISO8859_1
	default:
		switch section {
		case "appleii", "atarist":
			return charmap.ISO8859_1
		}
	}
	return helper.DetermineEncoding(b)
}

// Read returns the content of either the file download or an extracted text file.
// The text is intended to be used as a readme, preview or an in-browser viewer.
func Read(path string, res *models.File) ([]byte, error) {
	if res == nil {
		return nil, ErrFileModel
	}

	fname := res.Filename.String
	uuid := res.UUID.String

	var files struct {
		uuidTxt string
		uutxtOk bool
		filepth string
		filepOk bool
		txt     bool
	}
	files.uuidTxt = filepath.Join(path, uuid+".txt")
	files.uutxtOk = helper.IsStat(files.uuidTxt)
	files.filepth = filepath.Join(path, uuid)
	files.filepOk = helper.IsStat(files.filepth)
	files.txt = !exts.IsArchive(fname)

	if !files.uutxtOk && !files.filepOk {
		return nil, fmt.Errorf("%w: %s", ErrDownload, filepath.Join(path, uuid))
	}

	if !files.uutxtOk && !Viewer(res) {
		return nil, nil
	}

	name := files.filepth
	if files.uutxtOk {
		name = files.uuidTxt
	}

	b, err := os.ReadFile(name)
	if err != nil {
		return nil, err
	}

	const nul = 0x00
	b = bytes.ReplaceAll(b, []byte{nul}, []byte(" "))
	return b, nil
}

// IsUTF16 returns true if the byte slice is encoded in UTF-16.
func IsUTF16(p []byte) bool {
	const minimum = 2
	if len(p) < minimum {
		return false
	}
	if p[0] == 0xff && p[1] == 0xfe {
		return true
	}
	if p[0] == 0xfe && p[1] == 0xff {
		return true
	}
	return false
}

// Viewer returns true if the file entry should display the file download in the browser plain text viewer.
func Viewer(res *models.File) bool {
	if res == nil {
		return false
	}
	platform := strings.ToLower(strings.TrimSpace(res.Platform.String))
	switch platform {
	case "text", textamiga:
		return true
	}
	return false
}

// NoScreenshot returns true when the file entry should not attempt to display a screenshot.
// This is based on the platform, section or if the screenshot is missing on the server.
func NoScreenshot(path string, res *models.File) bool {
	if res == nil {
		return true
	}
	platform := strings.ToLower(strings.TrimSpace(res.Platform.String))
	switch platform {
	case textamiga, "text":
		return true
	}
	uuid := res.UUID.String
	webp := strings.Join([]string{path, fmt.Sprintf("%s.webp", uuid)}, "/")
	png := strings.Join([]string{path, fmt.Sprintf("%s.png", uuid)}, "/")
	if helper.IsStat(webp) || helper.IsStat(png) {
		return false
	}
	return true
}
