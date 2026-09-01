// Package readme provides functions for reading and suggesting readme files.
package readme

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"slices"
	"strings"
	"unicode/utf8"

	"github.com/Defacto2/helper"
	"github.com/Defacto2/magicnumber"
	"github.com/Defacto2/server/internal/dir"
	"github.com/Defacto2/server/internal/logs"
	"github.com/Defacto2/server/internal/nils"
	"github.com/bengarrett/ansibump"
	"github.com/bengarrett/binbump"
	"github.com/bengarrett/sauce"
	"golang.org/x/text/encoding/charmap"
)

var (
	ErrDownload = errors.New("readme: cannot stat the downloaded file")
	ErrFilename = errors.New("readme: file model filename is empty")
	ErrUUID     = errors.New("readme: file model uuid is empty")
	ErrTooLong  = errors.New("readme: reader is larger than 1MB")
)

type Text struct {
	Download dir.Directory
	Extra    dir.Directory
	UUID     string
	Filename string
	Platform string
	Section  string
	Year     int16
	MaxSize  int64
	Sign     magicnumber.Signature
	Record   sauce.Record
}

// Buffers parses the text files use by a file artifact.
//
// Two buffers are created and returned:
//
//   - The first buffer is used for CP-1252 and ISO-8859-1 text
//   - The second buffer is used for UTF-8 text
//
// Errors are generally logged except some critical issues.
func (t *Text) Buffers(sl *slog.Logger) (*bytes.Buffer, *bytes.Buffer, error) {
	if sl == nil {
		sl = logs.Discard()
	}
	const format = "readme %s: %w"

	textBuf := new(bytes.Buffer)
	runeBuf := new(bytes.Buffer)
	descBuf := new(bytes.Buffer)
	helpBuf := new(bytes.Buffer)

	defer func() {
		descBuf.Reset()
		helpBuf.Reset()
	}()

	var err error
	if dErr := t.descriptor(descBuf); dErr != nil {
		err = errors.Join(err, fmt.Errorf(format, "descriptor", dErr))
	}
	if pErr := t.primary(textBuf, runeBuf); pErr != nil {
		err = errors.Join(err, fmt.Errorf(format, "primary", pErr))
	}
	if hErr := t.helper(helpBuf); hErr != nil {
		err = errors.Join(err, fmt.Errorf(format, "helper", hErr))
	}

	if emptyBufs(textBuf, runeBuf, descBuf, helpBuf) {
		if uErr := t.handleUnknown(sl); uErr != nil {
			err = errors.Join(err, uErr)
		}
		t.Error(sl, err)
		return nil, nil, nil
	}

	if notText := t.signature(sl, textBuf); notText {
		t.Error(sl, err)
		textBuf.Reset()
		runeBuf.Reset()
		return nil, nil, nil
	}

	t.handleSAUCE(textBuf)

	// text with ANSI escape codes use a custom readme template
	if ansi, aErr := hasANSI(bytes.NewReader(textBuf.Bytes())); aErr != nil {
		// handle the error and continue to try to display the buffer elsewhere
		err = errors.Join(err, fmt.Errorf(format, "incompatible ansi", aErr))
	} else if ansi {
		sl.Info("readme will render the buffer as ansi encoded")
		buf, err := t.handleANSI(textBuf)
		if err != nil {
			return nil, nil, err
		}
		return buf, nil, nil
	}

	// log any previous errors
	t.Error(sl, err)

	// binary texts can cause false positives
	if t.Sign == magicnumber.Unknown {
		sl.Info("readme will render the buffer as binary text")
		buf, err := t.handleBIN(textBuf)
		if err != nil {
			return nil, nil, err
		}
		return buf, nil, nil
	}

	sl.Info("readme will render the buffer as raw or plain text")
	return t.handleRAW(textBuf, runeBuf, descBuf, helpBuf)
}

// Error logs the error as a warning with useful Text metadata.
func (t *Text) Error(sl *slog.Logger, err error) {
	if sl == nil || err == nil {
		return
	}
	sl.Warn("handler buffer problem",
		slog.String("uuid", t.UUID), slog.String("filename", t.Filename),
		slog.String("platform", t.Platform), slog.String("section", t.Section),
		slog.Int("year", int(t.Year)), slog.Any("errors", err),
	)
}

func emptyBufs(textBuf, runeBuf, descBuf, helpBuf *bytes.Buffer) bool {
	if err := nils.Check(textBuf, descBuf, helpBuf, runeBuf); err != nil {
		return true
	}
	return descBuf.Len() == 0 && textBuf.Len() == 0 && helpBuf.Len() == 0 && runeBuf.Len() == 0
}

func (t *Text) handleRAW(textBuf, runeBuf, descBuf, helpBuf *bytes.Buffer) (
	*bytes.Buffer, *bytes.Buffer, error,
) {
	if err := nils.Check(textBuf, descBuf, helpBuf, runeBuf); err != nil {
		return nil, nil, fmt.Errorf("plain texts: %w", err)
	}

	b := trimBytes(textBuf.Bytes()) // usually the readme or nfo text

	if descBuf.Len() > 0 { // usually the file_id or other header text
		// avoid edge cases where two buffers might have the same content
		if !bytes.Equal(descBuf.Bytes(), textBuf.Bytes()) {
			b = AddPrefix(b, descBuf.Bytes())
		}
		descBuf.Reset()
	}

	if helpBuf.Len() > 0 { // usually a manual or secondary text
		if !bytes.Equal(helpBuf.Bytes(), textBuf.Bytes()) {
			b = AddSuffix(b, helpBuf.Bytes())
		}
		helpBuf.Reset()
	}

	b = removeControls(b)

	if len(bytes.TrimSpace(b)) == 0 {
		textBuf.Reset()
		runeBuf.Reset()
		return nil, nil, nil
	}
	if len(b) > 0 {
		textBuf.Reset()
		textBuf.Write(b)
	}
	return textBuf, runeBuf, nil
}

// handleUnknown will open the UUID named file in Download
// and save any found SAUCE metadata to [Text.Record].
func (t *Text) handleUnknown(sl *slog.Logger) error {
	const format = "handle unknown %s: %w"
	if t.UUID == "" {
		return ErrUUID
	}

	name := t.Download.Join(t.UUID)
	sl.Info("readme attempted to show binary data", slog.String("name", name))

	// reopen the uuid download file
	r, err := os.Open(name)
	if err != nil {
		return fmt.Errorf(format, "open "+name, err)
	}
	defer r.Close()

	rec, err := sauce.Read(r)
	if err != nil {
		return fmt.Errorf(format, "sauce read", err)
	}
	if rec != nil && rec.ID == "SAUCE" {
		t.Record = *rec
	}
	return nil
}

func (t *Text) handleBIN(textBuf *bytes.Buffer) (*bytes.Buffer, error) {
	const format = "handle binary text %s: %w"
	if err := nils.Check(textBuf); err != nil {
		return nil, fmt.Errorf(format, "buffer", err)
	}

	width := 0 // use default
	maxRows := 0
	palette := binbump.StandardCGA
	year := t.Year
	if y := int(year); helper.Year(y) && y < 1992 {
		width = 80
		maxRows = 25
		palette = binbump.RevisedCGA
	}

	bintext, err := binbump.Buffer(
		bytes.NewReader(textBuf.Bytes()), width, maxRows, palette, nil,
	)
	if err != nil {
		return nil, fmt.Errorf(format, "binbump", err)
	}
	textBuf.Reset()
	return bintext, nil
}

func (t *Text) handleANSI(textBuf *bytes.Buffer) (*bytes.Buffer, error) {
	const format = "handle ansi %s: %w"
	if err := nils.Check(textBuf); err != nil {
		return nil, fmt.Errorf(format, "buffers", err)
	}

	const defaultColumns = 80
	width := defaultColumns
	ti1 := t.Record.Info.Info1
	if ti1.Info == "character width" && ti1.Value > 0 {
		width = int(ti1.Value)
	}

	charset := charmap.CodePage437
	palette := ansibump.CGA16
	amiga := false
	if t.Platform == "textamiga" {
		charset = charmap.ISO8859_1
		palette = ansibump.DP2
		amiga = true
	}
	config := ansibump.Customizer{
		Width:       width,
		AmigaParser: amiga,
		Strict:      false,
		Color:       palette,
		CharSet:     charset,
	}
	ansitext, err := config.Buffer(bytes.NewReader(textBuf.Bytes()))
	if err != nil {
		return nil, fmt.Errorf(format, "customizer", err)
	}
	// reset all other buffers and return the ansi buffer
	textBuf.Reset()
	return ansitext, nil
}

func (t *Text) handleSAUCE(textBuf *bytes.Buffer) {
	const minimum = 128 // SAUCE records are exactly 128 bytes at EOF
	if textBuf == nil || textBuf.Len() < minimum {
		return
	}

	b := textBuf.Bytes()
	if sr := sauce.Decode(b); sr.ID == "SAUCE" {
		t.Record = sr
	}
}

// signature checks the bytes to confirm they can be displayed as text.
func (t *Text) signature(sl *slog.Logger, textBuf *bytes.Buffer) bool {
	if sl == nil || textBuf == nil {
		return true
	}

	r := bytes.NewReader(textBuf.Bytes())
	t.Sign = magicnumber.Find(r)
	sl.Info("readme matched sign", slog.String("sign", t.Sign.String()))

	// skip UTF-16 or UTF-32 encodings which cannot be rendered natively in browser
	if t.Sign == magicnumber.UTF16Text || t.Sign == magicnumber.UTF32Text {
		sl.Info("readme found incompatible text", slog.String("sign", t.Sign.String()))
		return true
	}

	// skip known images and XBinary text without dynamic slice allocations
	if t.Sign == magicnumber.XBinaryText || slices.Contains(magicnumber.Images(), t.Sign) {
		sl.Info("readme found known image or binary format", slog.String("sign", t.Sign.String()))
		return true
	}

	return false
}

// descriptor (File ID - Description In ZIP) returns the content of archive file descriptor.
// Usually this brief summary text is named 'FILE_ID.descriptor' and is a legacy of the BBS
// era of file hosting.
//
// The summary text can be used as a readme, preview, or viewed in the browser.
func (t *Text) descriptor(descBuf *bytes.Buffer) error {
	const extension = ".diz"
	return t.secondary(descBuf, extension)
}

// helper returns the content of a helper text file.
// These optional texts can be secondary NFOs, READMEs, or additional instructions.
func (t *Text) helper(helpBuf *bytes.Buffer) error {
	const extension = ".hlp"
	return t.secondary(helpBuf, extension)
}

// secondary inserts any extra files such as file_id.diz or help files to the buf buffer.
func (t *Text) secondary(buf *bytes.Buffer, extension string) error {
	const format = "secondary handler %s: %w"

	if err := nils.Check(buf); err != nil {
		return fmt.Errorf(format, "arguments", err)
	}
	if t.UUID == "" {
		return ErrUUID
	}

	name := t.Extra.Join(t.UUID + extension)
	src, err := os.Open(name)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // non-fatal
		}
		buf.Reset()
		buf.WriteString("error could not read the extra text file")
		return nil
	}
	defer src.Close()

	buf.Reset()
	if _, err = io.Copy(buf, src); err != nil {
		return fmt.Errorf(format, "io copy", err)
	}

	b := buf.Bytes()
	b = bytes.ReplaceAll(b, byteNull, byteSpace)
	b = bytes.ReplaceAll(b, byteEOF, emptyBytes)
	b = bytes.ReplaceAll(b, byteCR, byteLF) // must go first before crlf
	b = bytes.ReplaceAll(b, byteCRLF, byteLF)
	b = helper.MaskTerm(b...)

	buf.Reset()
	buf.Write(b)

	return nil
}

// primary writes the content of either the file download or an extracted text file to the buffers.
// The text is intended to be used as a readme, preview or an in-browser viewer.
//
// Both the buf buffer and the ruf rune buffer are reset before writing.
func (t *Text) primary(textBuf, runeBuf *bytes.Buffer) error {
	const format = "render primary text %s: %w"

	if err := nils.Check(textBuf, runeBuf); err != nil {
		return fmt.Errorf(format, "buffers", err)
	}

	// always fulfill the contract to reset both buffers first
	textBuf.Reset()
	runeBuf.Reset()

	name, err := t.filename()
	if err != nil {
		return fmt.Errorf(format, "named", err)
	}
	if name == "" {
		return nil
	}

	f, err := os.Open(name)
	if err != nil {
		textBuf.WriteString("error could not read the information text file")
		return nil
	}
	defer f.Close()

	st, err := f.Stat()
	if err != nil {
		textBuf.WriteString("error could not describe the information text file")
		return nil
	}

	if st.Size() > t.MaxSize {
		textBuf.WriteString("skipped, text is too long")
		return nil
	}

	// read directly into memory once
	textBuf.Grow(int(st.Size()))
	if _, err := io.Copy(textBuf, f); err != nil {
		textBuf.Reset()
		return fmt.Errorf(format, "information text copy", err)
	}

	b := textBuf.Bytes()
	r := bytes.NewReader(b)
	var p []byte

	if sign, _ := magicnumber.Text(r); sign != magicnumber.Unknown {
		p = t.normalize(textBuf)
	} else {
		_, _ = r.Seek(0, io.SeekStart)
		if sign, _ := magicnumber.Archive(r); sign != magicnumber.Unknown {
			textBuf.Reset()
			return nil
		}
		p = b
	}

	// update textBuf only if normalized bytes modified the slice
	textBuf.Reset()
	p = trimEOF(p)
	textBuf.Write(p)

	if utf8.Valid(p) {
		runeBuf.Write(p)
	}

	return nil
}

func (t *Text) filename() (string, error) {
	const format = "readme filename %s: %w"

	if t.Filename == "" {
		return "", ErrFilename
	}
	if t.UUID == "" {
		return "", ErrUUID
	}

	readmePath := t.Extra.Join(t.UUID + ".txt")
	readmeOkay := helper.Stat(readmePath)

	if readmeOkay {
		return readmePath, nil
	}

	// when the readme text file does not exist
	// check if we should fall back to using the artifact
	if !t.useViewer() {
		return "", nil
	}

	artifactPath := t.Download.Join(t.UUID)
	if !helper.Stat(artifactPath) {
		return "", fmt.Errorf(format, artifactPath, ErrDownload)
	}

	return artifactPath, nil
}

// useViewer returns true if the file entry should display the file download in the browser plain text viewer.
// The result is based on the platform and section such as "text" or "textamiga" will return true.
func (t *Text) useViewer() bool {
	if strings.EqualFold(strings.TrimSpace(t.Filename), "file_id.diz") {
		return true
	}

	section := strings.TrimSpace(t.Section)
	if strings.EqualFold(section, "package") {
		return false
	}

	platform := strings.TrimSpace(t.Platform)
	return strings.EqualFold(platform, "text") || strings.EqualFold(platform, "textamiga")
}

func (t *Text) normalize(textBuf *bytes.Buffer) []byte {
	if textBuf == nil || textBuf.Len() == 0 {
		return []byte{}
	}

	b := textBuf.Bytes()
	b = bytes.ReplaceAll(b, byteNull, byteSpace)
	b = bytes.ReplaceAll(b, byteCRLF, byteLF)
	b = bytes.TrimRight(b, "\x1a")
	return helper.Mask(b...)
}
