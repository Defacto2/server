//nolint:gochecknoglobals
package app

import (
	"bytes"
	"cmp"
	"database/sql"
	"fmt"
	"io"
	"math"
	"path/filepath"
	"slices"
	"strings"

	"github.com/Defacto2/archive/pkzip"
	"github.com/Defacto2/helper"
	"github.com/Defacto2/magicnumber"
	"github.com/Defacto2/server/handler/internal/simple"
	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/Defacto2/server/internal/tags"
	"github.com/aarondl/null/v8"
	"github.com/bengarrett/bbs"
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
	if content == "" {
		return nil
	}

	x := strings.Split(content, "\n")
	if len(x) == 1 {
		return x
	}

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

func discard(err error) {
	_, _ = fmt.Fprint(io.Discard, err)
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
	return releasersHrefs(art) +
		`<br>` +
		`<span class="font-monospace fs-5 fw-light">` + string(a) + `</span>`
}

func legacyArchiving(modMagic any) bool {
	val, ok := modMagic.(string)
	if !ok {
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

	const unmod = 3
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

	const size = 2
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

// plainText returns true when matching utf8 text, ansi escape text, and plain text.
func plainText(modMagic any) bool {
	val, ok := modMagic.(string)
	if !ok {
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
	if name == "" {
		return false
	}

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

func parseValS(val any) (string, bool) {
	switch v := val.(type) {
	case string:
		return v, true
	case null.String:
		if v.Valid {
			return v.String, true
		}
		return "", true
	default:
		return "", false
	}
}

func parseValI(val any) (int, bool) {
	switch v := val.(type) {
	case int:
		return castSigned(v)
	case int8:
		return castSigned(v)
	case int16:
		return castSigned(v)
	case int32:
		return castSigned(v)
	case int64:
		return castSigned(v)
	case uint:
		return castUnsigned(v)
	case uint8:
		return castUnsigned(v)
	case uint16:
		return castUnsigned(v)
	case uint32:
		return castUnsigned(v)
	case uint64:
		return castUnsigned(v)
	case null.Int:
		if !v.Valid {
			return 0, false
		}
		return castSigned(v.Int)
	case null.Int16:
		if !v.Valid {
			return 0, false
		}
		return castSigned(v.Int16)
	case null.Int32:
		if !v.Valid {
			return 0, false
		}
		return castSigned(v.Int32)
	case null.Int64:
		if !v.Valid {
			return 0, false
		}
		return castSigned(v.Int64)
	case null.Byte:
		if !v.Valid {
			return 0, false
		}
		return castUnsigned(v.Byte)
	case sql.NullInt16:
		if !v.Valid {
			return 0, false
		}
		return castSigned(v.Int16)
	case sql.NullInt32:
		if !v.Valid {
			return 0, false
		}
		return castSigned(v.Int32)
	case sql.NullInt64:
		if !v.Valid {
			return 0, false
		}
		return castSigned(v.Int64)
	case sql.NullByte:
		if !v.Valid {
			return 0, false
		}
		return castUnsigned(v.Byte)

	default:
		return 0, false
	}
}

func parseValI64(val any) (int64, bool) {
	switch v := val.(type) {
	case int:
		return castSigned64(v)
	case int8:
		return castSigned64(v)
	case int16:
		return castSigned64(v)
	case int32:
		return castSigned64(v)
	case int64:
		return castSigned64(v)
	case uint:
		return castUnsigned64(v)
	case uint8:
		return castUnsigned64(v)
	case uint16:
		return castUnsigned64(v)
	case uint32:
		return castUnsigned64(v)
	case uint64:
		return castUnsigned64(v)
	default:
		return 0, false
	}
}

func castSigned64[T Integer](val T) (int64, bool) {
	if val < 0 {
		return 0, false
	}
	return int64(val), true
}

func castUnsigned64[T Integer](val T) (int64, bool) {
	if uint64(val) > math.MaxInt64 {
		return 0, false // Protects against overflow
	}
	return int64(val), true
}

type Integer interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64
}

func castSigned[T Integer](val T) (int, bool) {
	if val < 0 {
		return 0, false
	}
	return int(val), true
}

func castUnsigned[T Integer](val T) (int, bool) {
	if uint64(val) > math.MaxInt {
		return 0, false
	}
	return int(val), true
}
