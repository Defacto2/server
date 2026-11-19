//nolint:cyclop,gochecknoglobals
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
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/Defacto2/archive/pkzip"
	"github.com/Defacto2/helper"
	"github.com/Defacto2/magicnumber"
	"github.com/Defacto2/server/handler/areacode"
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
	chip := ""
	if len(a) > 0 {
		chip = `<span class="bg-body-tertiary border rounded px-2 py-0 font-monospace fs-6 text-body-secondary">` +
			string(a) + `</span>`
	}

	rel := releasersHrefs(art)
	switch {
	case rel != "" && chip != "":
		return `<span class="d-inline-flex flex-wrap align-items-center gap-2">` + rel + chip + `</span>`
	case rel != "":
		return rel
	default:
		return chip
	}
}

// lockWidth returns the byte array with an enforced maximum width of printed characters per line.
// If there are 3 or more tabs found in the byte array, it will be returned unmodified.
//
// The maxWidth should usually be a value of 80 representing the standard terminal column value.
func lockWidth(maxWidth int, b []byte) []byte {
	if maxWidth <= 0 {
		return b
	}

	// Exit early if 3 or more tabs are present
	const tabs = 3
	if bytes.Count(b, []byte{'\t'}) >= tabs {
		return b
	}

	var builder bytes.Buffer

	for line := range bytes.Lines(b) {
		// INFO: line includes all terminating newlines.

		if len(line) <= maxWidth {
			builder.Write(line)
			continue
		}

		// Chunk line into maxWidth pieces
		for i := 0; i < len(line); i += maxWidth {
			end := min(i+maxWidth, len(line))
			builder.WriteByte('\n')
			builder.Write(line[i:end])
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
		magicnumber.ANSIEscapeText.Title() + simple.ANSIBlink,
		simple.ANSImation,
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

func parseValI(val any) (int, bool) { //nolint:funlen
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

// ScenerGroup holds a single alphabetical bucket of sceners for directory rendering.
type ScenerGroup struct {
	ID    string
	Label string
	Names []string
}

// ScenerIndexItem represents an entry in the alphabetical quick-jump index.
type ScenerIndexItem struct {
	ID     string
	Label  string
	Active bool
}

// BuildScenerIndex groups sorted scener names into "#" and "A"-"Z" sections and returns
// the non-empty groups alongside the full 27-entry quick-jump index.
func BuildScenerIndex(names []string) ([]ScenerGroup, []ScenerIndexItem) {
	buckets := make(map[string][]string, 27)
	for _, name := range names {
		key := scenerBucketKey(name)
		buckets[key] = append(buckets[key], name)
	}

	const letters = "#ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	groups := make([]ScenerGroup, 0, len(letters))
	index := make([]ScenerIndexItem, 0, len(letters))
	for _, r := range letters {
		label := string(r)
		id := label
		if label == "#" {
			id = "num"
		}
		items := buckets[label]
		index = append(index, ScenerIndexItem{
			ID:     id,
			Label:  label,
			Active: len(items) > 0,
		})
		if len(items) > 0 {
			groups = append(groups, ScenerGroup{
				ID:    id,
				Label: label,
				Names: items,
			})
		}
	}
	return groups, index
}

func scenerBucketKey(name string) string {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return "#"
	}
	r, _ := utf8.DecodeRuneInString(trimmed)
	u := unicode.ToUpper(r)
	if u >= 'A' && u <= 'Z' {
		return string(u)
	}
	return "#"
}

// AreaCodeEntry holds structured display fields for a single NANP area code.
type AreaCodeEntry struct {
	Code    areacode.NAN
	Regions string
	Note    string
}

// AreaRegionEntry holds structured display fields for a NANP state, province, or territory.
type AreaRegionEntry struct {
	Name  string
	Abbr  string
	Codes string
}

// BuildAreaCodeEntries returns structured entries for all NANP area codes sorted ascending.
func BuildAreaCodeEntries() []AreaCodeEntry {
	codes := areacode.AreaCodes()
	notes := areacode.Copy()
	entries := make([]AreaCodeEntry, 0, len(codes))
	for _, code := range codes {
		regs := areacode.RegionByCode(code)
		names := make([]string, 0, len(regs))
		for _, r := range regs {
			if r.Abbreviation != "" {
				names = append(names, fmt.Sprintf("%s (%s)", r.Name, r.Abbreviation))
			} else {
				names = append(names, r.Name)
			}
		}
		entries = append(entries, AreaCodeEntry{
			Code:    code,
			Regions: strings.Join(names, " + "),
			Note:    notes[code],
		})
	}
	return entries
}

// BuildAreaRegionEntries returns structured entries for all NANP regions sorted by name.
func BuildAreaRegionEntries() []AreaRegionEntry {
	regs := areacode.Regions()
	entries := make([]AreaRegionEntry, 0, len(regs))
	for _, r := range regs {
		entries = append(entries, AreaRegionEntry{
			Name:  r.Name,
			Abbr:  string(r.Abbreviation),
			Codes: formatAreaCodes(r.AreaCodes),
		})
	}
	return entries
}

// BuildAreaAbbrEntries returns structured entries for all 2-letter NANP abbreviations.
func BuildAreaAbbrEntries() []AreaRegionEntry {
	abbrs := areacode.Abbreviations()
	entries := make([]AreaRegionEntry, 0, len(abbrs))
	for _, abbr := range abbrs {
		r, ok := areacode.RegionByAbbr(abbr)
		if !ok {
			entries = append(entries, AreaRegionEntry{Abbr: string(abbr), Codes: "n/a"})
			continue
		}
		entries = append(entries, AreaRegionEntry{
			Name:  r.Name,
			Abbr:  string(r.Abbreviation),
			Codes: formatAreaCodes(r.AreaCodes),
		})
	}
	return entries
}

func formatAreaCodes(codes []areacode.NAN) string {
	if len(codes) == 0 {
		return "n/a"
	}
	parts := make([]string, len(codes))
	for i, c := range codes {
		parts[i] = strconv.Itoa(int(c))
	}
	return strings.Join(parts, ", ")
}
