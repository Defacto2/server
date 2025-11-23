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
	"github.com/Defacto2/magicnumber"
	"github.com/Defacto2/server/internal/dir"
	"github.com/Defacto2/server/internal/panics"
	"github.com/Defacto2/server/internal/postgres/models"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/encoding/unicode"
)

var (
	ErrDownload = errors.New("cannot stat the downloaded file")
	ErrFilename = errors.New("file model filename is empty")
	ErrUUID     = errors.New("file model uuid is empty")
)

const textamiga = "textamiga"

// Encoder returns the encoding for the model file entry.
// Based on the platform and section.
// Otherwise it will attempt to determine the encoding from the file byte content.
func Encoder(art *models.File, r io.Reader) encoding.Encoding {
	const msg = "render encoder"
	if art == nil {
		panic(fmt.Errorf("%s: %w", msg, panics.ErrNoArtM))
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
	// Guess and determine CP437 or Latin1, etc
	guess := helper.Determine(r)
	return guess
}

// InformationText writes the content of either the file download or an extracted text file to the buffers.
// The text is intended to be used as a readme, preview or an in-browser viewer.
//
// Both the buf buffer and the ruf rune buffer are reset before writing.
func InformationText(buf, ruf *bytes.Buffer, sizeLimit int64, art *models.File, download, extra dir.Directory) error {
	const msg = "render information text"
	if err := infopanic(buf, ruf, art, msg); err != nil {
		return err
	}
	name, err := infoFilename(buf, art, download, extra)
	if err != nil {
		return err
	} else if name == "" {
		return nil
	}
	st, err := os.Stat(name)
	if err != nil {
		b := []byte("error could not describe the information text file")
		buf.Write(b)
		return nil
	}
	if st.Size() > sizeLimit {
		b := []byte("skipped, text is too long")
		buf.Write(b)
		return nil
	}
	f, err := os.Open(name)
	if err != nil {
		b := []byte("error could not read the information text file")
		buf.Write(b)
		return nil
	}
	defer func() { _ = f.Close() }()
	buf.Reset()
	_, err = io.Copy(buf, f)
	if err != nil {
		return fmt.Errorf("information text copy %w: %q", err, name)
	}
	var p []byte
	if sign, _ := magicnumber.Text(f); sign != magicnumber.Unknown {
		p = normalize(buf)
	} else {
		p = buf.Bytes()
	}
	buf.Reset()
	buf.Write(p)
	if utf8.Valid(p) {
		ruf.Reset()
		ruf.Write(p)
	}
	return nil
}

func infopanic(buf, ruf *bytes.Buffer, art *models.File, msg string) error {
	if buf == nil {
		return fmt.Errorf("%s: buf %w", msg, panics.ErrNoBuffer)
	}
	if ruf == nil {
		return fmt.Errorf("%s: ruf %w", msg, panics.ErrNoBuffer)
	}
	if art == nil {
		return fmt.Errorf("%s: %w", msg, panics.ErrNoArtM)
	}
	return nil
}

func infoFilename(buf *bytes.Buffer, art *models.File, download, extra dir.Directory) (string, error) {
	const msg = "render readme pool"
	fname := art.Filename.String
	if fname == "" {
		return "", ErrFilename
	}
	unid := art.UUID.String
	if unid == "" {
		return "", ErrUUID
	}
	var files struct {
		artifact struct {
			okay bool
			path string
		}
		readmeText struct {
			okay bool
			path string
		}
	}
	files.readmeText.path = extra.Join(unid + ".txt")
	files.readmeText.okay = helper.Stat(files.readmeText.path)
	files.artifact.path = download.Join(unid)
	files.artifact.okay = helper.Stat(files.artifact.path)
	if !files.artifact.okay && !files.readmeText.okay {
		return "", fmt.Errorf("%s: %w %q", msg, ErrDownload, download.Join(unid))
	}
	if !files.readmeText.okay && !Viewer(art) {
		buf.Reset()
		return "", nil
	}
	name := files.artifact.path
	if files.readmeText.okay {
		name = files.readmeText.path
	}
	return name, nil
}

func normalize(buf *bytes.Buffer) []byte {
	b := buf.Bytes()
	const nul = 0x00
	b = bytes.ReplaceAll(b, []byte{nul}, []byte(" "))
	// normalize the line feeds to attempt to fix any breakages with the layout
	b = bytes.ReplaceAll(b, []byte("\r\n"), []byte("\n"))
	b = helper.Mask(b...)
	return b
}

// DescriptionInZIP returns the content of the description in archive file.
// Usually this brief summary text is named 'file_id.diz' and is a legacy of the BBS
// era of file hosting.
//
// The summary text can be used as a readme, preview, or viewed in the browser.
func DescriptionInZIP(buf *bytes.Buffer, art *models.File, extra dir.Directory) error {
	const msg = "description in zip"
	if buf == nil {
		return fmt.Errorf("%s: %w", msg, panics.ErrNoBuffer)
	}
	if art == nil {
		return fmt.Errorf("%s: %w", msg, panics.ErrNoArtM)
	}
	unid := art.UUID.String
	if unid == "" {
		return ErrUUID
	}
	const extension = ".diz"
	diz := extra.Join(unid + extension)
	if !helper.Stat(diz) {
		return nil
	}
	f, err := os.Open(diz)
	if err != nil {
		b := []byte("error could not read the description file")
		buf.Write(b)
	}
	defer func() { _ = f.Close() }()

	buf.Reset()
	_, err = io.Copy(buf, f)
	if err != nil {
		return fmt.Errorf("%s copy %w: %q", msg, err, diz)
	}
	b := buf.Bytes()
	const nul, eof = 0x00, "\x1a"
	b = bytes.ReplaceAll(b, []byte{nul}, []byte(" "))
	// normalize the line feeds as often courier groups injecting their tags would break the layout
	b = bytes.ReplaceAll(b, []byte("\r\r"), []byte("\n")) // this should be before \r\n
	b = bytes.ReplaceAll(b, []byte("\r\n"), []byte("\n"))
	b = bytes.ReplaceAll(b, []byte("\n\r"), []byte("\n"))
	b = bytes.ReplaceAll(b, []byte("\n\n"), []byte("\n")) // this should be after all \r replacements
	b = bytes.ReplaceAll(b, []byte(eof), []byte(""))      // there maybe more than one injected end-of-file char
	b = helper.MaskTerm(b...)
	buf.Reset()
	buf.Write(b)
	return nil
}

// InsertDiz inserts the FILE_ID.DIZ content into the existing byte content.
func InsertDiz(b []byte, diz []byte) []byte {
	if bytes.TrimSpace(diz) == nil {
		return b
	}
	if bytes.TrimSpace(b) == nil {
		return diz
	}
	sep := []byte("\n\n")
	size := len(diz) + len(sep) + len(b)
	buf := bytes.NewBuffer(make([]byte, 0, size))
	buf.Write(diz)
	buf.Write(sep)
	buf.Write(b)
	return buf.Bytes()
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
	section := strings.ToLower(strings.TrimSpace(art.Section.String))
	if section == "package" {
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
	jpg := strings.Join([]string{previewPath, unid + ".jpg"}, "/")
	if helper.Stat(webp) || helper.Stat(png) || helper.Stat(jpg) {
		return false
	}
	return true
}
