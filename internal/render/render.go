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
	case "textamiga":
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

	if !Viewer(res) {
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
	b = bytes.ToValidUTF8(b, []byte("x"))
	// TODO: func to replace all non-standard controls with a space
	b = bytes.ReplaceAll(b, []byte{0x0}, []byte(" "))

	return b, nil
}

// Viewer returns true if the file entry should display the file download in the browser plain text viewer.
func Viewer(res *models.File) bool {
	if res == nil {
		return false
	}
	platform := strings.ToLower(strings.TrimSpace(res.Platform.String))
	switch platform {
	case "text":
		return true
	}
	return false
}

// NoReadme returns true if the file entry has a "no readme" flagged.
func NoReadme(res *models.File) bool {
	return res.RetrotxtNoReadme.Int16 != 0
}

// NoScreenshot returns true when the file entry should not attempt to display a screenshot.
// This is based on the platform.
func NoScreenshot(res *models.File) bool {
	if res == nil {
		return false
	}
	platform := strings.ToLower(strings.TrimSpace(res.Platform.String))
	switch platform {
	case "textamiga", "text", "atarist":
		return true
	}

	return false
}
