package app

import (
	"bytes"
	"cmp"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"path/filepath"
	"slices"
	"strings"

	"github.com/Defacto2/archive/pkzip"
	"github.com/Defacto2/helper"
	"github.com/Defacto2/magicnumber"
	"github.com/Defacto2/server/handler/app/internal/simple"
	"github.com/Defacto2/server/internal/nils"
	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/Defacto2/server/internal/tags"
	"github.com/bengarrett/bbs"
	"github.com/labstack/echo/v5"
)

var (
	byteNBSP    = []byte{0xa0} // non-breaking space for ISO8859-1
	byteNBSP437 = []byte{0xff} // non-breaking space for CP437
	byteSpace   = []byte{0x20}
	byteSHY     = []byte{0xad} // soft hyphen for ISO8859-1
	byteHyphen  = []byte{0x2d} // hyphen-minus
)

// LockIn80Columns returns true if the readme viewer should probably lock the
// width of text to a maximum of 80 columns, the traditional screen width on
// Microsoft DOS. However, with the popularization of Microsoft Windows and the
// use of notepad.exe, many newer texts break with this lock. But there are
// reasons to use it, as BBS era texts sometimes lack newlines.
//
// In the future this could be expanded to count the number of newlines vs the
// size of the byte sec.
//
// This func does the following checks:
//
//   - confirms the text isn't PCBoard
//   - confirms the text isn't newer than 1992
func LockIn80Columns(year int16, src ...byte) bool {
	const epoch = 1992 // Windows 3.1 release year
	switch {
	case len(src) == 0:
		return false
	case bbs.IsPCBoard(src):
		return false
	case year >= epoch:
		return false
	}
	return true
}

func SortContent(content string) []string {
	x := strings.Split(content, "\n")

	slices.SortFunc(x, func(a, b string) int {
		// sort by filename, but behave like Windows and ignore case ordering
		x := strings.TrimSpace(strings.ToLower(a))
		y := strings.TrimSpace(strings.ToLower(b))

		// sort by directory length
		lenX := strings.Count(x, "/")
		lenY := strings.Count(y, "/")

		// sort by filename extensions
		extX := filepath.Ext(x)
		extY := filepath.Ext(y)

		return cmp.Or(
			cmp.Compare(lenX, lenY),
			strings.Compare(extX, extY),
			strings.Compare(x, y),
		)
	})

	items := make([]string, 0, len(x))

	for _, s := range x {
		if strings.HasSuffix(s, "/") {
			continue
		}
		items = append(items, s)
	}

	return items
}

// artifact404 renders the error page for the artifact links.
func artifact404(sl *slog.Logger, c *echo.Context, id string) error {
	const title = "Artifact not found"
	const probl = "The artifact page does not exist, there is probably a typo with the URL."
	const msg = "artifact 404 context"
	const format = msg + ": %w"
	if err := nils.Check(c, sl); err != nil {
		return fmt.Errorf(format, err)
	}
	const name = "status"
	data := empty(c)
	data["title"] = fmt.Sprintf("%d error, artifact page not found", http.StatusNotFound)
	data["description"] = fmt.Sprintf("HTTP status %d error", http.StatusNotFound)
	data["code"] = http.StatusNotFound
	data["logo"] = title
	data["alert"] = fmt.Sprintf("Artifact %q cannot be found", strings.ToLower(id))
	data["probl"] = probl
	data["uriOkay"] = "f/"
	data["uriErr"] = id
	err := c.Render(http.StatusNotFound, name, data)
	if err != nil {
		return InternalErr(sl, c, name, errorWithID(err, id, nil))
	}
	return nil
}

// decode decodes the text content from the reader.
func decode(src io.Reader) (string, error) {
	if src == nil {
		return "", nil
	}

	out := strings.Builder{}
	if _, err := io.Copy(&out, src); err != nil {
		return "", fmt.Errorf("decode copy: %w", err)
	}

	if !strings.HasSuffix(out.String(), "\n\n") {
		out.WriteString("\n")
	}

	return out.String(), nil
}

// errorWithID returns an error with the artifact ID appended to the error message.
// The key string is expected any will always be displayed in the error message.
// The id can be an integer or string value and should be the database numeric ID.
func errorWithID(err error, key string, id any) error {
	if err == nil {
		return nil
	}
	key = strings.TrimSpace(key)
	const format = "%w: caused by artifact %s (%v)"
	switch id.(type) {
	case int, int64:
		return fmt.Errorf(format, err, key, id)
	case string:
		return fmt.Errorf(format, err, key, id)
	default:
		return fmt.Errorf(format, err, key, id)
	}
}

// firstLead returns the lead for the file record which is the filename and releasers.
func firstLead(art *models.File) string {
	if art == nil {
		return ""
	}
	fname, err := simple.CleanFname(art.Filename.String)
	if err != nil {
		fname = ""
	}
	a := helper.MaskTerm([]byte(fname)...)
	const format = `<span class="font-monospace fs-5 fw-light">%s</span> `
	span := fmt.Sprintf(format, a)
	return fmt.Sprintf("%s<br>%s", releasersHrefs(art), span)
}

func legacyArchiving(modMagic any) bool {
	switch modMagic.(type) {
	case string:
	default:
		return false
	}
	val, valid := modMagic.(string)
	if !valid {
		return false
	}
	switch val {
	case
		magicnumber.ARChiveSEA.Title(),
		magicnumber.YoshiLHA.Title(),
		magicnumber.ArchiveRobertJung.Title(),
		magicnumber.PKWAREZipImplode.Title(),
		magicnumber.PKWAREZipReduce.Title(),
		magicnumber.PKWAREZipShrink.Title():
		return true
	default:
		return false
	}
}

// lockWidth returns the byte array with an enforced maximum width of printed characters per line.
// If there are 3 or more tabs found in the byte array, it will be returned unmodified.
//
// The maxWidth should usually be a value of 80 representing the standard terminal column value.
func lockWidth(maxWidth int, b []byte) []byte {
	tabs := 0
	index := 0
	const size, unmod = 2, 3
	for index < len(b) {
		if tabs >= unmod {
			return b
		}
		index = bytes.IndexByte(b[index:], byte('\t'))
		if index < 0 {
			break
		}
		tabs++
	}
	var builder bytes.Buffer
	for line := range bytes.Lines(b) {
		// builder.Write(line) // uncomment to debug
		total := len(line) - 1
		if total <= maxWidth {
			builder.Write(line)
			continue
		}
		cut := 0
		for n := range line {
			if n%maxWidth == 0 {
				p := make([]byte, 1, len(line[cut:n])+1)
				p[0] = '\n'
				p = append(p, line[cut:n]...)
				builder.Write(p)
				cut = n
				continue
			}
			if n >= total {
				p := make([]byte, 1, len(line[cut:n])+size)
				p[0] = '\n'
				p = append(p, line[cut:n]...)
				p = append(p, byte('\n'))
				builder.Write(p)
				break
			}
		}
	}
	return builder.Bytes()
}

// plainText returns true when matching utf8 text,
// ansi escape text, and plain text.
func plainText(modMagic any) bool {
	switch modMagic.(type) {
	case string:
	default:
		return false
	}
	val, valid := modMagic.(string)
	if !valid {
		return false
	}
	switch val {
	case
		magicnumber.UTF8Text.Title(),
		magicnumber.ANSIEscapeText.Title(),
		magicnumber.PlainText.Title():
		return true
	default:
		return false
	}
}

// releasersHrefs returns the releasers for the file record as a string of HTML links.
func releasersHrefs(art *models.File) string {
	if art == nil {
		return ""
	}
	magazine := strings.TrimSpace(art.Section.String) == tags.Mag.String()
	return string(LinkRelrs(magazine, art.GroupBrandBy, art.GroupBrandFor))
}

func requireReplacementZip(name string) bool {
	methods, err := pkzip.Methods(name)
	if err != nil {
		return false
	}
	for method := range slices.Values(methods) {
		if !method.Zip() {
			return true
		}
	}
	return false
}
