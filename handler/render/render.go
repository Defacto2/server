// Package render provides the file content rendering for the web server.
package render

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/Defacto2/helper"
	"github.com/Defacto2/server/internal/dir"
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
func Read(art *models.File, download, extra dir.Directory) ([]byte, []rune, error) {
	if art == nil {
		return nil, nil, ErrFileModel
	}

	fname := art.Filename.String
	unid := art.UUID.String

	if fname == "" {
		return nil, nil, ErrFilename
	}
	if unid == "" {
		return nil, nil, ErrUUID
	}

	var files struct {
		uuidTxt string
		filepth string
		uutxtOk bool
		filepOk bool
	}
	files.uuidTxt = extra.Join(unid + ".txt")
	files.uutxtOk = helper.Stat(files.uuidTxt)
	files.filepth = extra.Join(unid)
	files.filepOk = helper.Stat(files.filepth)

	if !files.uutxtOk && !files.filepOk {
		return nil, nil, fmt.Errorf("render read %w: %s", ErrDownload, download.Join(unid))
	}

	if !files.uutxtOk && !Viewer(art) {
		doNothing := []byte{}
		return doNothing, nil, nil
	}

	name := files.filepth
	if files.uutxtOk {
		name = files.uuidTxt
	}

	b, err := os.ReadFile(name)
	if err != nil {
		b = []byte("error could not read the readme text file")
	}

	r := []rune{}
	if utf8.Valid(b) {
		r = bytes.Runes(b)
	}
	const nul = 0x00
	b = bytes.ReplaceAll(b, []byte{nul}, []byte(" "))
	return b, r, nil
}

// Diz returns the content of the FILE_ID.DIZ file.
// The text is intended to be used as a readme, preview or an in-browser viewer.
func Diz(art *models.File, extra dir.Directory) ([]byte, error) {
	if art == nil {
		return nil, ErrFileModel
	}

	unid := art.UUID.String
	if unid == "" {
		return nil, ErrUUID
	}

	diz := extra.Join(unid + ".diz")
	if !helper.Stat(diz) {
		return nil, nil
	}

	b, err := os.ReadFile(diz)
	if err != nil {
		b = []byte("error could not read the diz file")
	}

	const nul = 0x00
	b = bytes.ReplaceAll(b, []byte{nul}, []byte(" "))
	return b, nil
}

// InsertDiz inserts the FILE_ID.DIZ content into the extisting byte content.
func InsertDiz(b []byte, diz []byte) []byte {
	if bytes.TrimSpace(diz) == nil {
		return b
	}
	x := diz
	if bytes.TrimSpace(b) == nil {
		return x
	}
	x = append(x, []byte("\n\n")...)
	x = append(x, b...)
	return x
}

// Viewer returns true if the file entry should display the file download in the browser plain text viewer.
// The result is based on the platform and section such as "text" or "textamiga" will return true.
// If the filename is "file_id.diz" then it will return false.
func Viewer(art *models.File) bool {
	if art == nil {
		return false
	}
	if strings.EqualFold(art.Filename.String, "file_id.diz") {
		// avoid displaying the file_id.diz twice in the browser viewer
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
func NoScreenshot(art *models.File, previewPath string) bool {
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
