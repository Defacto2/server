// Package filerecord provides functions for the file model which is an artifact record.
package filerecord

import (
	"errors"
	"fmt"
	"html/template"
	"image"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/Defacto2/archive"
	"github.com/Defacto2/helper"
	"github.com/Defacto2/magicnumber"
	"github.com/Defacto2/releaser"
	"github.com/Defacto2/server/handler/app/internal/extensions"
	"github.com/Defacto2/server/handler/app/internal/simple"
	"github.com/Defacto2/server/handler/jsdos/msdos"
	"github.com/Defacto2/server/handler/readme"
	"github.com/Defacto2/server/internal/command"
	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/Defacto2/server/internal/tags"
	"github.com/Defacto2/server/model"
	"github.com/dustin/go-humanize"
	_ "golang.org/x/image/bmp"  // Register BMP image format
	_ "golang.org/x/image/tiff" // Register TIFF image format
	"golang.org/x/text/encoding/charmap"
)

const (
	YYYYMMDD = "2006-Jan-02"

	epoch                   = model.EpochYear // epoch is the default year for MS-DOS files without a timestamp
	textamiga               = "textamiga"
	arrowLink template.HTML = `<svg class="bi" aria-hidden="true">` +
		`<use xlink:href="/svg/bootstrap-icons.svg#arrow-right"></use></svg>`
	br  = "<br>"
	bat = ".bat"
	cmd = ".cmd"
	com = ".com"
	exe = ".exe"
	ini = ".ini"
)

// ListEntry is a struct for the directory item that is used to generate the HTML.
type ListEntry struct {
	RelativeName            string
	Signature               string
	Filesize                string
	ImageConfig             string
	MusicConfig             string
	UniqueID                string
	Executable              magicnumber.Windows
	Images, Programs, Texts bool
	name                    string
	platform                string
	section                 string
	bytes                   int64
}

// HTML returns the HTML for an file item in the "Download content" section of the File editor.
func (m *ListEntry) HTML(bytes int64, platform, section string) string {
	m.name = url.QueryEscape(m.RelativeName)
	m.bytes = bytes
	m.platform = platform
	m.section = section
	displayname := m.RelativeName
	if strings.EqualFold(platform, tags.DOS.String()) {
		if msdos.Rename(displayname) != displayname {
			displayname = `<span class="text-danger">` + displayname + `</span>`
		}
	}
	htm := fmt.Sprintf(`<div class="col d-inline-block text-truncate">%s</div>`,
		displayname)
	return m.Column1(htm)
}

func (m ListEntry) Column1(htm string) string {
	const blank = `<div class="col col-1"></div>`
	ext := strings.ToLower(filepath.Ext(m.name))
	switch {
	case osTool(ext):
		htm += blank
	case m.Images:
		htm += previewcopy(m.UniqueID, m.name)
	case m.Texts:
		htm += readmepreview(m.UniqueID, m.name, m.platform)
	default:
		htm += blank
	}
	return m.Column2(htm)
}

func (m ListEntry) Column2(htm string) string {
	soloText := func() bool {
		if !strings.EqualFold(m.platform, tags.Text.String()) &&
			!strings.EqualFold(m.platform, textamiga) {
			return false
		}
		return strings.EqualFold(m.section, tags.Nfo.String())
	}
	const blank = `<div class="col col-1"></div>`
	name := url.QueryEscape(m.RelativeName)
	ext := strings.ToLower(filepath.Ext(name))
	switch {
	case strings.EqualFold(m.RelativeName, "file_id.diz"):
		htm += dizcopy(m.UniqueID, name)
	case m.Programs, ext == exe, ext == com:
		htm += `<div class="col col-1 text-end" ` +
			`data-bs-toggle="tooltip" data-bs-title="Known program or executable">` +
			`<svg width="16" height="16" fill="currentColor" aria-hidden="true">` +
			`<use xlink:href="/svg/bootstrap-icons.svg#terminal-plus"></use></svg></div>`
	case osTool(ext):
		htm += blank
	case m.Texts || soloText():
		htm += readmecopy(m.UniqueID, name)
	default:
		htm += blank
	}
	htm += fmt.Sprintf(`<div><small data-bs-toggle="tooltip" data-bs-title="%d bytes">%s</small>`,
		m.bytes, m.Filesize)
	return m.Column1and2(ext, htm)
}

func (m ListEntry) Column1and2(ext, htm string) string {
	switch {
	case m.Texts && (ext == bat || ext == cmd):
		htm += fmt.Sprintf(` <small class="">%s</small></div>`, "command script")
	case m.Texts && (ext == ini):
		htm += fmt.Sprintf(` <small class="">%s</small></div>`, "configuration textfile")
	case m.Programs || ext == com:
		htm = progr(m.Executable, ext, htm, m.bytes)
	case m.MusicConfig != "":
		htm += fmt.Sprintf(` <small class="">%s</small></div>`, m.MusicConfig)
	case m.Images:
		htm += fmt.Sprintf(` <small class="">%s</small></div>`, m.ImageConfig)
	default:
		htm += fmt.Sprintf(` <small class="">%s</small></div>`, m.Signature)
	}
	htm = fmt.Sprintf(`<div class="border-bottom row mb-1">%s</div>`, htm)
	return htm
}

func previewcopy(uniqueID, name string) string {
	return `<div class="col col-1 text-end" ` +
		`data-bs-toggle="tooltip" data-bs-title="Use image for preview">` +
		fmt.Sprintf(`<a class="icon-link align-text-bottom" name="artifact-editor-comp-previewcopy" `+
			`hx-indicator="#artifact-editor-comp-htmx-indicator" `+
			`hx-target="#artifact-editor-comp-feedback" `+
			`hx-patch="/editor/preview/copy/%s/%s">`, uniqueID, name) +
		`<svg width="16" height="16" fill="currentColor" aria-hidden="true">` +
		`<use xlink:href="/svg/bootstrap-icons.svg#images"></use></svg></a></div>`
}

func readmepreview(uniqueID, name, platform string) string {
	uri := "preview"
	if strings.EqualFold(platform, tags.TextAmiga.String()) {
		uri = "preview-amiga"
	}
	return `<div class="col col-1 text-end" ` +
		`data-bs-toggle="tooltip" data-bs-title="Use file for preview">` +
		fmt.Sprintf(`<a class="icon-link align-text-bottom" name="artifact-editor-comp-previewtext" `+
			`hx-indicator="#artifact-editor-comp-htmx-indicator" `+
			`hx-target="#artifact-editor-comp-feedback" `+
			`hx-patch="/editor/readme/%s/%s/%s">`, uri, uniqueID, name) +
		`<svg width="16" height="16" fill="currentColor" aria-hidden="true">` +
		`<use xlink:href="/svg/bootstrap-icons.svg#images"></use></svg></a></div>`
}

func readmecopy(uniqueID, name string) string {
	return `<div class="col col-1 text-end" ` +
		`data-bs-toggle="tooltip" data-bs-title="Use file as readme">` +
		fmt.Sprintf(`<a class="icon-link align-text-bottom" name="artifact-editor-comp-textcopy" `+
			`hx-indicator="#artifact-editor-comp-htmx-indicator" `+
			`hx-target="#artifact-editor-comp-feedback" `+
			`hx-patch="/editor/readme/copy/%s/%s">`, uniqueID, name) +
		`<svg class="bi" width="16" height="16" fill="currentColor" aria-hidden="true">` +
		`<use xlink:href="/svg/bootstrap-icons.svg#file-text"></use></svg></a></div>`
}

func dizcopy(uniqueID, name string) string {
	return `<div class="col col-1 text-end" ` +
		`data-bs-toggle="tooltip" data-bs-title="Use file as the FILE_ID.DIZ">` +
		fmt.Sprintf(`<a class="icon-link align-text-bottom" name="artifact-editor-comp-dizcopy" `+
			`hx-indicator="#artifact-editor-comp-htmx-indicator" `+
			`hx-target="#artifact-editor-comp-feedback" `+
			`hx-patch="/editor/diz/copy/%s/%s">`, uniqueID, name) +
		`<svg class="bi" width="16" height="16" fill="currentColor" aria-hidden="true">` +
		`<use xlink:href="/svg/bootstrap-icons.svg#file-text"></use></svg></a></div>`
}

func progr(exec magicnumber.Windows, ext, htm string, bytes int64) string {
	const epochYear = 1980
	const x8086 = 64 * 1024
	dosProg := (ext == exe || ext == com)
	var s string
	switch {
	case dosProg && exec.PE != magicnumber.UnknownPE:
		s = exec.String() + " executable"
	case dosProg && exec.NE == magicnumber.UnknownNE:
		s = progrDos(x8086, bytes)
	case dosProg && exec.NE != magicnumber.NoneNE:
		s = exec.String() + " executable"
	case dosProg:
		s = "MS Dos program"
	case ext == ".dll" && exec.PE != magicnumber.UnknownPE:
		s = "Windows dynamic-link library"
	case exec.NE != magicnumber.NoneNE:
		s = "NE program data"
	default:
		s = "PE program data"
	}
	if y := exec.TimeDateStamp.Year(); y >= epochYear && y <= time.Now().Year() {
		s += ", built " + exec.TimeDateStamp.Format("2006-01-2")
	}
	htm += fmt.Sprintf(` <small class="">%s</small></div>`, s)
	return htm
}

func progrDos(x8086 int, bytes int64) string {
	if x8086 >= int(bytes) {
		return "Dos command"
	}
	return "Dos executable"
}

// osTool returns true if the file extension matches the known operating system tools.
// This includes batch scripts, executables, commands and ini configurations files.
func osTool(ext string) bool {
	switch ext {
	case bat, exe, com, ini:
		return true
	default:
		return false
	}
}

// AlertURL returns the VirusTotal URL for the security alert for the file record.
// This will normally return an empty string unless the file has a security alert.
func AlertURL(art *models.File) string {
	if art == nil {
		return ""
	}
	if !art.FileSecurityAlertURL.Valid {
		return ""
	}
	raw := strings.TrimSpace(art.FileSecurityAlertURL.String)
	u, err := url.ParseRequestURI(raw)
	if err != nil {
		return ""
	}
	if host := u.Hostname(); host == "" {
		u.Host = "www.virustotal.com"
	}
	if u.Scheme != "https" {
		u.Scheme = "https"
	}
	return u.String()
}

// AttrArtist returns the attributed artist names for the file record.
func AttrArtist(art *models.File) string {
	if art == nil {
		return ""
	}
	if art.CreditIllustration.Valid {
		return art.CreditIllustration.String
	}
	return ""
}

// AttrMusic returns the attributed musician names for the file record.
func AttrMusic(art *models.File) string {
	if art == nil {
		return ""
	}
	if art.CreditAudio.Valid {
		return art.CreditAudio.String
	}
	return ""
}

// AttrProg returns the attributed programmer names for the file record.
func AttrProg(art *models.File) string {
	if art == nil {
		return ""
	}
	if art.CreditProgram.Valid {
		return art.CreditProgram.String
	}
	return ""
}

// AttrWriter returns the attributed text writer names for the file record.
func AttrWriter(art *models.File) string {
	if art == nil {
		return ""
	}
	if art.CreditText.Valid {
		return art.CreditText.String
	}
	return ""
}

// Basename returns the name of the file given to the artifact file record.
func Basename(art *models.File) string {
	if art == nil {
		return ""
	}
	if art.Filename.Valid {
		return art.Filename.String
	}
	return ""
}

// Checksum returns the strong SHA386 hash checksum for the file record.
func Checksum(art *models.File) string {
	if art == nil {
		return ""
	}
	if art.FileIntegrityStrong.Valid {
		return strings.TrimSpace(art.FileIntegrityStrong.String)
	}
	return ""
}

// Comment returns the optional comment for the file record.
func Comment(art *models.File) string {
	if art == nil {
		return ""
	}
	if art.Comment.Valid {
		return art.Comment.String
	}
	return ""
}

type entry struct {
	module  string
	size    string
	format  string
	exec    magicnumber.Windows
	sign    magicnumber.Signature
	zeros   int
	bytes   int64
	image   bool
	text    bool
	program bool
}

// ParseFile parses the file at the given path and returns true if it should be skipped.
func (e *entry) ParseFile(path string, platform string) bool {
	const skipEntry = true
	info, err := os.Stat(path)
	if err != nil {
		return skipEntry
	}
	if info.IsDir() {
		return skipEntry
	}
	return e.parse(path, platform, info)
}

// ParseDirEntry parses the directory entry and returns true if it should be skipped.
func (e *entry) ParseDirEntry(path string, d fs.DirEntry, platform string) bool {
	const skipEntry = true
	if d.IsDir() {
		return skipEntry
	}
	info, _ := d.Info()
	if info == nil {
		return skipEntry
	}
	return e.parse(path, platform, info)
}

func (e *entry) parse(path, platform string, info fs.FileInfo) bool {
	const skipEntry = true
	e.bytes = info.Size()
	if e.bytes == 0 {
		e.zeros++
		return skipEntry
	}
	e.size = humanize.Bytes(uint64(info.Size()))
	r, _ := os.Open(path)
	if r == nil {
		return skipEntry
	}
	defer r.Close()
	e.sign = magicnumber.Find(r)
	platform = strings.TrimSpace(platform)
	e.image = isImage(e.sign)
	e.text = isText(e.sign)
	e.program = isProgram(e.sign, platform)
	switch {
	case e.image:
		return e.parseImage(e.sign, path)
	case e.program:
		return e.parseProgram(path)
	case
		e.sign == magicnumber.MusicExtendedModule,
		e.sign == magicnumber.MusicMultiTrackModule,
		e.sign == magicnumber.MusicImpulseTracker,
		e.sign == magicnumber.MusicProTracker:
		return e.parseMusicMod(path)
	case
		e.sign == magicnumber.MPEG1AudioLayer3,
		platform == tags.Audio.String():
		return e.parseMusicID3(path)
	}
	return !skipEntry
}

func (e *entry) parseImage(sign magicnumber.Signature, path string) bool {
	const skipEntry = true
	r, _ := os.Open(path)
	if r == nil {
		return skipEntry
	}
	defer r.Close()
	config, format, err := image.DecodeConfig(r)
	if err == nil {
		e.format = fmt.Sprintf("%s image, %dx%d", format, config.Width, config.Height)
		return !skipEntry
	}
	switch sign {
	case magicnumber.InterleavedBitmap:
		r, _ := os.Open(path)
		if r == nil {
			return skipEntry
		}
		defer r.Close()
		x, y := magicnumber.IlbmDecode(r)
		e.format = fmt.Sprintf("ILBM image, %dx%d", x, y)
	default:
		e.format = sign.Title() + " image"
	}
	return !skipEntry
}

func (e *entry) parseProgram(path string) bool {
	const skipEntry = true
	r, _ := os.Open(path)
	if r == nil {
		return skipEntry
	}
	defer r.Close()
	exec, err := magicnumber.FindExecutable(r)
	if err == nil {
		e.exec = exec
	}
	return !skipEntry
}

func (e *entry) parseMusicMod(path string) bool {
	const skipEntry = true
	r, _ := os.Open(path)
	if r == nil {
		return skipEntry
	}
	defer r.Close()
	e.module = magicnumber.MusicTracker(r)
	return !skipEntry
}

// ParseMusicID3 parses the ID3 tag in the byte slice and returns the title, artist and year if available.
// It looks up in order the ID3v2.3, ID3v2.2 and ID3v1 tags in the byte slice with the priority being
// the newer versions of the tag.
//
// ID3v1 is a completely different tag format to ID3v2 and has serious limitations,
// so it is only used as a last resort.
func (e *entry) parseMusicID3(path string) bool {
	const skipEntry = true
	// ID3 v2.x tags are located at the start of the file.
	id3, _ := os.Open(path)
	if id3 == nil {
		return skipEntry
	}
	defer id3.Close()
	if s := magicnumber.MusicID3v2(id3); s != "" {
		e.module = s
		return !skipEntry
	}
	// ID3 v1 tags are located at the end of the file.
	if s := magicnumber.MusicID3v1(id3); s != "" {
		e.module = s
		return !skipEntry
	}
	return !skipEntry
}

// ListContent returns a list of the files contained in the archive file.
// This is used to generate the HTML for the "Download content" section of the File editor.
func ListContent(art *models.File, dirs command.Dirs, src string) template.HTML { //nolint:funlen
	if art == nil {
		return ""
	}
	entries, files, zeroByteFiles := 0, 0, 0
	unid := art.UUID.String
	if !art.UUID.Valid {
		return "error, no UUID"
	}
	platform := strings.TrimSpace(strings.ToLower(art.Platform.String))
	if !tags.IsPlatform(platform) {
		return "error, invalid platform"
	}
	section := strings.TrimSpace(strings.ToLower(art.Section.String))
	dst, err := archive.ExtractSource(src, art.Filename.String)
	if err != nil {
		return extractErr(src, platform, section, zeroByteFiles, err)
	}
	walkerCount := func(_ string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return fs.SkipDir
		}
		files++
		return nil
	}
	if err := filepath.WalkDir(dst, walkerCount); err != nil {
		return template.HTML(err.Error())
	}
	var b strings.Builder
	name := ""
	names := []string{}
	walkerFunc := func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return filepath.SkipDir
		}
		var skipEntry error
		rel, err := filepath.Rel(dst, path)
		if err != nil {
			debug := fmt.Sprintf(`<div class="border-bottom row mb-1">... %v more files</div>`, err)
			b.WriteString(debug)
			return skipEntry
		}
		e := entry{zeros: zeroByteFiles}
		if e.ParseDirEntry(path, d, platform) {
			zeroByteFiles = e.zeros
			return skipEntry
		}
		entries++
		if e.text {
			name = d.Name()
			names = append(names, name)
		}
		le := listEntry(e, rel, unid)
		b.WriteString(le.HTML(e.bytes, platform, section))
		if maxItems := 200; entries > maxItems {
			more := fmt.Sprintf(`<div class="border-bottom row mb-1">... %d more files</div>`, files-entries)
			b.WriteString(more)
			return filepath.SkipAll
		}
		return nil
	}
	if err = filepath.WalkDir(dst, walkerFunc); err != nil {
		b.Reset()
		return template.HTML(err.Error())
	}
	const maxItems = 2
	if l := len(names); l > 0 && l <= maxItems {
		i := indexDiz(names...)
		diz, src := "", ""
		useDizAndTxt := l == 2 && i != -1
		useTxt := l == 1
		if useDizAndTxt {
			diz = names[i]
			x := 1 - i
			src = filepath.Join(dst, names[x])
		}
		if useTxt {
			src = filepath.Join(dst, name)
		}
		if src != "" {
			if err := dirs.TextDeferred(src, unid); err != nil {
				b.Reset()
				return template.HTML(err.Error())
			}
		}
		if diz != "" {
			if err := dirs.DizDeferred(src, unid); err != nil {
				b.Reset()
				return template.HTML(err.Error())
			}
		}
	}
	b.WriteString(skippedEmpty(zeroByteFiles))
	return template.HTML(b.String())
}

func indexDiz(names ...string) int {
	for i, name := range names {
		if strings.EqualFold(name, "file_id.diz") {
			return i
		}
	}
	return -1
}

func extractErr(src, platform, section string, zeroByteFiles int, err error) template.HTML {
	if !errors.Is(err, archive.ErrNotArchive) && !errors.Is(err, archive.ErrNotImplemented) {
		return template.HTML(err.Error())
	}
	e := entry{zeros: zeroByteFiles}
	if e.ParseFile(src, platform) {
		return "error, empty byte file"
	}
	le := listErr(e)
	var b strings.Builder
	b.WriteString(le.HTML(e.bytes, platform, section))
	return template.HTML(b.String())
}

func listErr(e entry) ListEntry {
	return ListEntry{
		Executable:   e.exec,
		Images:       e.image,
		Programs:     e.program,
		Texts:        e.text,
		MusicConfig:  e.module,
		ImageConfig:  e.format,
		RelativeName: "",
		Signature:    e.sign.String(),
		Filesize:     e.size,
		UniqueID:     "",
	}
}

func listEntry(e entry, rel, unid string) ListEntry {
	return ListEntry{
		Executable:   e.exec,
		Images:       e.image,
		Programs:     e.program,
		Texts:        e.text,
		MusicConfig:  e.module,
		ImageConfig:  e.format,
		RelativeName: LegacyString(rel),
		Signature:    e.sign.String(),
		Filesize:     e.size,
		UniqueID:     unid,
	}
}

// LegacyString returns a string that is converted to UTF-8 if it is not already.
// Intended for filenames in archives that may have been encoded using a legacy charset,
// such as ISO-8859-1 (Commodore Amiga) or Windows-1252 (Windows 9x) and using non-ASCII characters.
func LegacyString(s string) string {
	if valid := utf8.ValidString(s); valid {
		return s
	}
	undefinedChr := func(b byte) bool {
		const euroSymbol, yDiaeresis = 0x80, 0x9f
		return b >= euroSymbol && b <= yDiaeresis
	}
	if windows1252 := slices.ContainsFunc([]byte(s), undefinedChr); windows1252 {
		decoder := charmap.Windows1252.NewDecoder()
		x, _ := decoder.String(s)
		if valid := utf8.ValidString(x); valid {
			return x
		}
	}
	decoder := charmap.ISO8859_1.NewDecoder()
	x, _ := decoder.String(s)
	if valid := utf8.ValidString(x); valid {
		return x
	}
	return s
}

func skippedEmpty(zeroByteFiles int) string {
	if zeroByteFiles == 0 {
		return ""
	}
	return fmt.Sprintf(`<div class="border-bottom row mb-1">... skipped %d empty (0 B) files</div>`, zeroByteFiles)
}

func isImage(sign magicnumber.Signature) bool {
	for _, v := range magicnumber.Images() {
		if v == sign {
			return true
		}
	}
	return false
}

func isProgram(sign magicnumber.Signature, platform string) bool {
	for _, v := range magicnumber.Programs() {
		if strings.EqualFold(platform, tags.DOS.String()) {
			break
		}
		if v == sign {
			return true
		}
	}
	return false
}

func isText(sign magicnumber.Signature) bool {
	for _, v := range magicnumber.Texts() {
		if v == sign {
			return true
		}
	}
	return false
}

// Date returns a formatted date string for the published date for the artifact.
func Date(art *models.File) template.HTML {
	if art == nil {
		return template.HTML(model.ErrModel.Error())
	}
	ys, ms, ds := "", "", ""
	if art.DateIssuedYear.Valid {
		if i := int(art.DateIssuedYear.Int16); helper.Year(i) {
			ys = strconv.Itoa(i)
		}
	}
	if art.DateIssuedMonth.Valid {
		if s := time.Month(art.DateIssuedMonth.Int16); s.String() != "" {
			ms = s.String()
		}
	}
	if art.DateIssuedDay.Valid {
		if i := int(art.DateIssuedDay.Int16); helper.Day(i) {
			ds = strconv.Itoa(i)
		}
	}
	strong := func(s string) template.HTML {
		return template.HTML("<strong>" + s + "</strong>")
	}
	if isYearOnly := ys != "" && ms == "" && ds == ""; isYearOnly {
		return strong(ys)
	}
	if isInvalidDay := ys != "" && ms != "" && ds == ""; isInvalidDay {
		return strong(ys) + template.HTML(" "+ms)
	}
	if isInvalid := ys == "" && ms == "" && ds == ""; isInvalid {
		return "unknown date"
	}
	return strong(ys) + template.HTML(fmt.Sprintf(" %s %s", ms, ds))
}

// Dates returns the year, month and day for the published date for the artifact.
func Dates(art *models.File) (int16, int16, int16) {
	if art == nil {
		return 0, 0, 0
	}
	y, m, d := int16(0), int16(0), int16(0)
	if art.DateIssuedYear.Valid {
		y = art.DateIssuedYear.Int16
	}
	if art.DateIssuedMonth.Valid {
		m = art.DateIssuedMonth.Int16
	}
	if art.DateIssuedDay.Valid {
		d = art.DateIssuedDay.Int16
	}
	return y, m, d
}

// Description returns a human readable description for the artifact.
// This includes the title, the releaser and the year of release.
func Description(art *models.File) string {
	if art == nil {
		return ""
	}
	s := art.Filename.String
	if art.RecordTitle.String != "" {
		s = FirstHeader(art)
	}
	r1 := releaser.Clean(strings.ToLower(art.GroupBrandBy.String))
	r2 := releaser.Clean(strings.ToLower(art.GroupBrandFor.String))
	r := ""
	switch {
	case r1 != "" && r2 != "":
		r = fmt.Sprintf("%s + %s", r1, r2)
	case r1 != "":
		r = r1
	case r2 != "":
		r = r2
	}
	s = fmt.Sprintf("%s released by %s", s, r)
	y := art.DateIssuedYear.Int16
	if y > 0 {
		s = fmt.Sprintf("%s in %d", s, y)
	}
	return s
}

// DownloadID returns the obfuscated ID for the file record.
// This is used to create a unique download link for the file based on its ID database key.
func DownloadID(art *models.File) string {
	if art == nil {
		return ""
	}
	return helper.ObfuscateID(art.ID)
}

// ExtraZip returns true if the file record has repacked zip file offering in the extra directory.
// This repackage gets used by the DOS emulator and also offered as an secondary download when
// the original artifact file uses a defunct compression method or format.
//
// The original artifact must always be preserved and offered as the primary download.
// But the extra zip file is a convenience for users who may not have the tools to decompress the original.
func ExtraZip(art *models.File, extraDir string) bool {
	if art == nil {
		return false
	}
	extraZip := 0
	unid := UnID(art)
	st, err := os.Stat(filepath.Join(extraDir, unid+".zip"))
	if err == nil && !st.IsDir() {
		extraZip = int(st.Size())
	}
	return extraZip > 0
}

// FileEntry returns the created and updated date and time for the file record using
// the "time ago" format.
//
// For example, "Created 2 days ago" or "Updated 1 month ago".
func FileEntry(art *models.File) string {
	if art == nil {
		return ""
	}
	switch {
	case art.Createdat.Valid && art.Updatedat.Valid:
		c := simple.Updated(art.Createdat.Time, "")
		u := simple.Updated(art.Updatedat.Time, "")
		if c != u {
			c = simple.Updated(art.Createdat.Time, "Created")
			u = simple.Updated(art.Updatedat.Time, "Updated")
			return c + br + u
		}
		c = simple.Updated(art.Createdat.Time, "Created")
		return c
	case art.Createdat.Valid:
		c := simple.Updated(art.Createdat.Time, "Created")
		return c
	case art.Updatedat.Valid:
		u := simple.Updated(art.Updatedat.Time, "Updated")
		return u
	}
	return ""
}

// FirstHeader returns the title of the file,
// unless the artifact is marked as a magazine issue, in which case it returns the issue number.
func FirstHeader(art *models.File) string {
	if art == nil {
		return ""
	}
	sect := strings.TrimSpace(strings.ToLower(art.Section.String))
	if sect != "magazine" {
		return art.RecordTitle.String
	}
	s := art.RecordTitle.String
	if i, err := strconv.Atoi(s); err == nil {
		return fmt.Sprintf("Issue %d", i)
	}
	return s
}

// Idenfication16C returns the 16 color identification for the file record.
// This is usually a partial URL to the 16 color website.
func Idenfication16C(art *models.File) string {
	if art == nil {
		return ""
	}
	if art.WebID16colors.Valid {
		return art.WebID16colors.String
	}
	return ""
}

// IdenficationDZ returns the Demozoo production ID for the file record.
func IdenficationDZ(art *models.File) string {
	if art == nil {
		return ""
	}
	if art.WebIDDemozoo.Valid {
		id := art.WebIDDemozoo.Int64
		return strconv.FormatInt(id, 10)
	}
	return ""
}

// IdenficationGitHub returns the GitHub repository for the file record.
func IdenficationGitHub(art *models.File) string {
	if art == nil {
		return ""
	}
	if art.WebIDGithub.Valid {
		return art.WebIDGithub.String
	}
	return ""
}

// IdenficationPouet returns the Pouet production ID for the file record.
func IdenficationPouet(art *models.File) string {
	if art == nil {
		return ""
	}
	if art.WebIDPouet.Valid {
		id := art.WebIDPouet.Int64
		return strconv.FormatInt(id, 10)
	}
	return ""
}

// IdenficationYT returns the YouTube video watch ID for the file record.
func IdenficationYT(art *models.File) string {
	if art == nil {
		return ""
	}
	if art.WebIDYoutube.Valid {
		return strings.TrimSpace(art.WebIDYoutube.String)
	}
	return ""
}

// JsdosArchive returns true if the file record is a known MS-DOS archive file.
func JsdosArchive(art *models.File) bool {
	if art == nil {
		return false
	}
	switch filepath.Ext(strings.ToLower(art.Filename.String)) {
	case ".zip", ".lhz", ".lzh", ".arc", ".arj":
		return true
	}
	return false
}

// JsdosBroken returns true if the MsDos artifact is known to be incompatible with the js-dos emulator.
func JsdosBroken(art *models.File) bool {
	if art == nil {
		return false
	}
	if art.DoseeIncompatible.Valid {
		return art.DoseeIncompatible.Int16 != 0
	}
	return false
}

// JsdosCPU returns the js-dos CPU type for the file record.
func JsdosCPU(art *models.File) string {
	if art == nil {
		return ""
	}
	if art.DoseeHardwareCPU.Valid {
		return art.DoseeHardwareCPU.String
	}
	return ""
}

// JsdosMachine returns the js-dos machine type for the file record.
// This is usually the graphic card type but can also be a unique machine
// type such as "tandy" that is range of hardware.
func JsdosMachine(art *models.File) string {
	if art == nil {
		return ""
	}
	if art.DoseeHardwareGraphic.Valid {
		return art.DoseeHardwareGraphic.String
	}
	return ""
}

// JsdosMemory returns true if js-dos should disable the XMS, EMS and UMB memory options.
func JsdosMemory(art *models.File) (bool, bool, bool) {
	if art == nil {
		return false, false, false
	}
	x, e, u := false, false, false
	if art.DoseeNoXMS.Valid {
		x = art.DoseeNoXMS.Int16 == 0
	}
	if art.DoseeNoEms.Valid {
		e = art.DoseeNoEms.Int16 == 0
	}
	if art.DoseeNoUmb.Valid {
		u = art.DoseeNoUmb.Int16 == 0
	}
	return x, e, u
}

// JsdosRun returns the program name or sequence of commands to launch in the js-dos emulator.
func JsdosRun(art *models.File) string {
	if art == nil {
		return ""
	}
	if art.DoseeRunProgram.Valid {
		return art.DoseeRunProgram.String
	}
	return ""
}

// JsdosSound returns the js-dos sound card or built-in audio for the file record.
func JsdosSound(art *models.File) string {
	if art == nil {
		return ""
	}
	if art.DoseeHardwareAudio.Valid {
		return art.DoseeHardwareAudio.String
	}
	return ""
}

// jsdosUse returns true if the file record is a known, MS-DOS executable.
// The supported file types are .zip archives and .exe, .com. binaries.
// Script files such as .bat and .cmd are not supported.
func JsdosUse(art *models.File) bool {
	if art == nil {
		return false
	}
	if strings.TrimSpace(strings.ToLower(art.Platform.String)) != "dos" {
		return false
	}
	if JsdosArchive(art) {
		return true
	}
	ext := filepath.Ext(strings.ToLower(art.Filename.String))
	switch ext {
	case exe, com:
		return true
	case bat, cmd:
		return false
	default:
		return false
	}
}

// JsdosUsage returns true if the js-dos emulator should be used with the filename.
func JsdosUsage(filename, platform string) bool {
	filename = strings.ToLower(filename)
	ext := filepath.Ext(filename)
	platform = strings.TrimSpace(strings.ToLower(platform))
	if platform != "dos" {
		return false
	}
	switch ext {
	case ".zip", ".lhz", ".lzh", ".arc", ".arj":
		return true
	}
	switch ext {
	case exe, com:
		return true
	case bat, cmd:
		return false
	default:
		return false
	}
}

// JsdosUtilities returns true the js-dos emulator should also load the utilities archive
// as an internal hard disk drive.
func JsdosUtilities(art *models.File) bool {
	if art == nil {
		return false
	}
	if art.DoseeLoadUtilities.Valid {
		return art.DoseeLoadUtilities.Int16 != 0
	}
	return false
}

// LastModification returns the last modified date and time for the file record.
func LastModification(art *models.File) string {
	if art == nil {
		return ""
	}
	const none = "no timestamp"
	if !art.FileLastModified.Valid {
		return none
	}
	year, _ := strconv.Atoi(art.FileLastModified.Time.Format("2006"))
	if year <= epoch {
		return none
	}
	lm := art.FileLastModified.Time.Format("2006 Jan 2, 15:04")
	if lm == "0001 Jan 1, 00:00" {
		return none
	}
	return lm
}

// LastModificationDate returns the last modified date for the file record.
func LastModificationDate(art *models.File) string {
	if art == nil {
		return ""
	}
	const none = "no timestamp"
	if !art.FileLastModified.Valid {
		return none
	}
	year, _ := strconv.Atoi(art.FileLastModified.Time.Format("2006"))
	if year <= epoch {
		return none
	}
	lm := art.FileLastModified.Time.Format(YYYYMMDD)
	if lm == "0001-01-01" {
		return none
	}
	return lm
}

// LastModifications returns the year, month and day for the last modified date for the file record.
func LastModifications(art *models.File) (int, int, int) {
	if art == nil {
		return 0, 0, 0
	}
	if !art.FileLastModified.Valid || art.FileLastModified.IsZero() {
		return 0, 0, 0
	}
	y := art.FileLastModified.Time.Year()
	m := int(art.FileLastModified.Time.Month())
	d := art.FileLastModified.Time.Day()
	return y, m, d
}

// LastModificationAgo returns the last modified date in a human readable format.
func LastModificationAgo(art *models.File) string {
	if art == nil {
		return ""
	}
	const none = "No recorded timestamp"
	if !art.FileLastModified.Valid {
		return none
	}
	year, _ := strconv.Atoi(art.FileLastModified.Time.Format("2006"))
	if year <= epoch {
		return none
	}
	return simple.Updated(art.FileLastModified.Time, "Modified")
}

// LinkPreview returns a URL path to link to the file record in tab, to use as a preview.
// A preview link is only available for certain file types such as images, text, documents,
// and renders the whole item in its own browser tab without any HTML or CSS from the website.
func LinkPreview(art *models.File) string {
	if art == nil {
		return ""
	}
	if art.ID == 0 {
		return ""
	}
	id := art.ID
	name := ""
	platform := ""
	if art.Filename.Valid {
		name = art.Filename.String
	}
	if art.Platform.Valid {
		platform = art.Platform.String
	}
	return LinkPreviewHref(id, name, platform)
}

// LinkPreviewHref creates a URL path to link to the file record in tab, to use as a preview.
//
// A list of supported file types: https://developer.mozilla.org/en-US/docs/Web/Media/Formats/Image_types
func LinkPreviewHref(id any, name, platform string) string {
	if id == nil || name == "" {
		return ""
	}
	platform = strings.TrimSpace(platform)
	// supported formats
	// https://developer.mozilla.org/en-US/docs/Web/Media/Formats/Image_types
	ext := strings.ToLower(filepath.Ext(name))
	switch {
	case slices.Contains(extensions.Archive(), ext):
		// this must always be first
		return ""
	case platform == textamiga, platform == "text":
		break
	case slices.Contains(extensions.Document(), ext):
		break
	case slices.Contains(extensions.Image(), ext):
		break
	case slices.Contains(extensions.Media(), ext):
		break
	default:
		return ""
	}
	s, err := simple.LinkID(id, "v")
	if err != nil {
		return fmt.Sprint("error: ", err)
	}
	return s
}

// LinkPreviewTip returns a tooltip for the link preview.
func LinkPreviewTip(art *models.File) string {
	if art == nil {
		return ""
	}
	name := ""
	platform := ""
	if art.Filename.Valid {
		name = art.Filename.String
	}
	if art.Platform.Valid {
		platform = art.Platform.String
	}
	return simple.LinkPreviewTip(name, platform)
}

// LinkSVG returns an right-arrow SVG icon.
func LinkSVG() template.HTML {
	return arrowLink
}

// Magic returns the magic number or guessed file type for the file record.
func Magic(art *models.File) string {
	if art == nil {
		return ""
	}
	if art.FileMagicType.Valid {
		return strings.TrimSpace(art.FileMagicType.String)
	}
	return ""
}

// Readme returns a guessed or suggested readme file name to use for the record.
func Readme(r *models.File) string {
	if r == nil {
		return ""
	}
	filename := r.Filename.String
	group := r.GroupBrandFor.String
	if group == "" {
		group = r.GroupBrandBy.String
	}
	if x := strings.Split(group, " "); len(x) > 1 {
		group = x[0]
	}
	cont := strings.ReplaceAll(r.FileZipContent.String, "\r\n", "\n")
	content := strings.Split(cont, "\n")
	return readme.Suggest(filename, group, content...)
}

// RecordIsNew returns true if the file record is a new upload.
func RecordIsNew(art *models.File) bool {
	if art == nil {
		return false
	}
	return !art.Deletedat.IsZero() && art.Deletedby.IsZero()
}

// RecordOffline returns true if the file record is marked as offline.
// This means the artifact has been soft deleted and is no longer available for download.
func RecordOffline(art *models.File) bool {
	if art == nil {
		return false
	}
	return !art.Deletedat.IsZero() && !art.Deletedby.IsZero()
}

// RecordOnline returns true if the artifact file record is available for download.
func RecordOnline(art *models.File) bool {
	if art == nil {
		return false
	}
	return art.Deletedat.Time.IsZero()
}

// RecordProblems returns a list of validation problems for the file record.
func RecordProblems(art *models.File) string {
	if art == nil {
		return ""
	}
	validate := model.Validate(art)
	if validate == nil {
		return ""
	}
	x := strings.Split(validate.Error(), ",")
	s := make([]string, 0, len(x))
	for _, v := range x {
		if strings.TrimSpace(v) == "" {
			continue
		}
		s = append(s, v)
	}
	s = slices.Clip(s)
	return strings.Join(s, " + ")
}

// Relations returns the list of relationships for the file record.
func Relations(art *models.File) template.HTML {
	if art == nil {
		return ""
	}
	s := art.ListRelations.String
	if s == "" {
		return ""
	}
	links := strings.Split(s, "|")
	if len(links) == 0 {
		return ""
	}
	rows := ""
	const expected = 2
	const route = "/f/"
	for _, link := range links {
		x := strings.Split(link, ";")
		if len(x) != expected {
			continue
		}
		name, href := x[0], x[1]
		if !strings.HasPrefix(href, route) {
			href = route + href
		}
		rows += fmt.Sprintf("<tr><th scope=\"row\"><small>Link to</small></th>"+
			"<td><small><a class=\"text-truncate\" href=\"%s\">%s</a></small></td></tr>", href, name)
	}
	return template.HTML(rows)
}

// RelationsStr returns the list of relationships for the file record as a string.
func RelationsStr(art *models.File) string {
	if art == nil {
		return ""
	}
	if art.ListRelations.Valid {
		return strings.TrimSpace(art.ListRelations.String)
	}
	return ""
}

// ReleaserPair returns the pair of releaser names for the file record.
// The first name is the releaser "for" and the second name is the releaser "by".
func ReleaserPair(art *models.File) (string, string) {
	if art == nil {
		return "", ""
	}
	pair := simple.ReleaserPair(art.GroupBrandFor, art.GroupBrandBy)
	return pair[0], pair[1]
}

// TagCategory returns the "Tag as category" for the file record,
// which is used to group similar artifacts together.
func TagCategory(art *models.File) string {
	if art == nil {
		return ""
	}
	if !art.Section.Valid {
		return ""
	}
	name := strings.ToLower(strings.TrimSpace(art.Section.String))
	if tags.IsCategory(name) {
		return name
	}
	return ""
}

// TagProgram returns the "Programs or apps" for the file record,
// which is the platform or operating system the artifact is intended for.
func TagProgram(art *models.File) string {
	if art == nil {
		return ""
	}
	if !art.Platform.Valid {
		return ""
	}
	name := strings.ToLower(strings.TrimSpace(art.Platform.String))
	if tags.IsPlatform(name) {
		return name
	}
	return ""
}

// Title returns the brief title of the file record or a issue number for a magazine.
func Title(art *models.File) string {
	if art == nil {
		return ""
	}
	return art.RecordTitle.String
}

// UnID returns the universal unique ID for the file record commonly known as a UUID.
func UnID(art *models.File) string {
	if art == nil {
		return ""
	}
	if art.UUID.Valid {
		return art.UUID.String
	}
	return ""
}

// EmbedReadme returns false if a text file artifact should not be displayed in the page as a readme or textfile.
// This includes artifacts that are set as documents such a HTML, PDF or BBS RIP images.
func EmbedReadme(art *models.File) bool {
	if art == nil {
		return false
	}
	const bbsRipImage = ".rip"
	if filepath.Ext(strings.ToLower(art.Filename.String)) == bbsRipImage {
		// the bbs era, remote images protcol is not supported
		// example: /f/b02392f
		return false
	}
	switch strings.TrimSpace(art.Platform.String) {
	case "markup", "pdf":
		return false
	}
	return true
}

// Websites returns the list of links for the file record.
func Websites(art *models.File) template.HTML {
	if art == nil {
		return ""
	}
	s := art.ListLinks.String
	if s == "" {
		return ""
	}
	links := strings.Split(s, "|")
	if len(links) == 0 {
		return ""
	}
	rows := ""
	const expected = 2
	for _, link := range links {
		x := strings.Split(link, ";")
		if len(x) != expected {
			continue
		}
		name, href := x[0], x[1]
		if !strings.HasPrefix(href, "http") {
			href = "https://" + href
		}
		rows += fmt.Sprintf("<tr><th scope=\"row\"><small>Link to</small></th>"+
			"<td><small><a class=\"link-offset-3 icon-link icon-link-hover\" "+
			"href=\"%s\">%s %s</a></small></td></tr>", href, name, LinkSVG())
	}
	return template.HTML(rows)
}

// WebsitesStr returns the list of links for the file record as a string.
func WebsitesStr(art *models.File) string {
	if art == nil {
		return ""
	}
	if art.ListLinks.Valid {
		return strings.TrimSpace(art.ListLinks.String)
	}
	return ""
}

// ZipContent returns the archive content of the file download, or an empty string if not an archive file.
func ZipContent(art *models.File) string {
	if art == nil {
		return ""
	}
	if art.FileZipContent.Valid {
		return strings.TrimSpace(art.FileZipContent.String)
	}
	return ""
}
