// Package filerecord provides functions for the file model which is an artifact record.
//
//nolint:exhaustive,gochecknoglobals,nonamedreturns
package filerecord

import (
	"fmt"
	"html/template"
	"image"
	"io/fs"
	"math"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/Defacto2/helper"
	"github.com/Defacto2/magicnumber"
	"github.com/Defacto2/server/handler/internal/simple"
	"github.com/Defacto2/server/handler/jsdos/msdos"
	"github.com/Defacto2/server/handler/readme"
	"github.com/Defacto2/server/handler/releaser"
	"github.com/Defacto2/server/internal/dir"
	"github.com/Defacto2/server/internal/extensions"
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
	blank = `<div class="col col-1"></div>`
	br    = "<br>"
	bat   = ".bat"
	cmd   = ".cmd"
	com   = ".com"
	exe   = ".exe"
	ini   = ".ini"
)

var enforceSimpleText = sync.OnceValue(func() map[string]struct{} {
	return map[string]struct{}{
		"aa97833330f4a27f0c7888ae633de652be5a37840fc87cc364b5c90908d027d855d66fe60d8b2b23b02fb0fe482ddcf1": {},
		"d312eba8773eff6f6d535f55f4c8cbbe10977ce2075297efca4773186cd3630699a855a1172050aaa0e22375ed644ac6": {},
	}
})

// ForceSimpleText returns true if an artifact should only display the readme
// as a plain text file.
//
// This is a fallback mechanism to deal with false positives with character
// combinations within texts that get mistaken as ANSI or BBS color codes
// and so need an manual override.
//
// Generally, a file hash comparison will be used for the false positive listed items.
func ForceSimpleText(art *models.File) bool {
	if art == nil {
		return false
	}

	hash := fileIntegrity(art)
	if hash == "" {
		return false
	}

	_, found := enforceSimpleText()[hash]
	return found
}

func fileIntegrity(art *models.File) string {
	if art == nil {
		return ""
	}
	if art.FileIntegrityStrong.Valid {
		return art.FileIntegrityStrong.String
	}
	return ""
}

// ListEntry is a struct for the directory item that is used to generate the HTML.
type ListEntry struct {
	RelativeName                      string
	Signature                         string
	Filesize                          string
	ImageConfig                       string
	MusicConfig                       string
	UniqueID                          string
	name                              string
	platform                          string
	section                           string
	Executable                        magicnumber.Windows
	bytes                             int64
	Images, Programs, Texts, BINtexts bool
}

// HTML returns the HTML for an file item in the "Download content" section of the File editor.
func (m *ListEntry) HTML(bytes int64, platform, section string) string {
	s := m.RelativeName
	m.name = url.QueryEscape(s)
	m.bytes = bytes
	m.platform = platform
	m.section = section

	title, err := simple.CleanFname(s)
	if err != nil {
		title = ""
	}
	if strings.EqualFold(platform, tags.DOS.String()) &&
		msdos.Rename(title) != title {
		title = `<span class="text-danger">` + title + `</span>`
	}

	var htm strings.Builder
	const size = 512
	htm.Grow(size)

	htm.WriteString(`<div class="border-bottom row mb-1">`)
	htm.WriteString(`<div class="col d-inline-block text-truncate">`)
	htm.WriteString(title)
	htm.WriteString(`</div>`)
	htm.WriteString(m.column1())
	htm.WriteString(m.column2())
	htm.WriteString(m.column3())
	htm.WriteString(m.columnFooter())
	htm.WriteString(`</div>`)
	return htm.String()
}

// systemfile returns true if the file extension matches the known operating system tools.
// This includes batch scripts, executables, commands and ini configurations files.
//
// This is to stop the preview text content button being displayed for known false-positives.
func systemfile(ext string) bool {
	switch ext {
	case bat, exe, com, ini:
		return true
	default:
		return false
	}
}

// column1 is the list entry, first column button of the "Download content" list.
func (m *ListEntry) column1() string {
	switch {
	case systemfile(strings.ToLower(filepath.Ext(m.name))):
		return blank
	case m.Images:
		// use Image
		return buttonImage(m.UniqueID, m.name)
	case m.Texts:
		// use Text
		return buttonText(m.UniqueID, m.name, m.platform, m.Signature)
	case m.BINtexts || m.xbinary():
		// use Text
		return buttonTextBinary(m.UniqueID, m.name)
	default:
		return blank
	}
}

// column2 is the list entry, second column button of the "Download content" list.
func (m *ListEntry) column2() string {
	filename := url.QueryEscape(m.RelativeName)
	ext := strings.ToLower(filepath.Ext(filename))
	switch {
	case systemfile(ext):
		return blank
	case m.briefDescription():
		// use DIZ
		return buttonDIZ(m.UniqueID, filename)
	case m.Programs || ext == exe || ext == com: // FIX: conflicts with systemfile
		// use EXE
		return buttonEXE()
	case m.Texts || m.BINtexts || m.textNFO():
		// use TEXT
		return buttonReadme(m.UniqueID, filename)
	default:
		return blank
	}
}

// column3 is the list entry, third column button of the "Download content" list.
func (m *ListEntry) column3() string {
	filename := url.QueryEscape(m.RelativeName)
	ext := strings.ToLower(filepath.Ext(filename))
	if useDIZ := m.briefDescription(); useDIZ {
		return blank
	}
	if systemfile(ext) {
		return blank
	}

	return buttonExtra(m.UniqueID, url.QueryEscape(m.RelativeName))
}

// buttonText creates a link to "/editor/readme/URI/ID/FILENAME".
func buttonText(id, filename, platform, sign string) string {
	if strings.EqualFold(platform, tags.TextAmiga.String()) ||
		strings.EqualFold(platform, tags.Console.String()) {
		if !strings.Contains(strings.ToLower(sign), "ansi") {
			// ansilove does not color ANSI using "ced" or "workbench"
			// instead, it renders the files as ASCII text files
			return buttonPreview("preview-amiga", id, filename)
		}
	}

	return buttonPreview("preview", id, filename)
}

// buttonTextBinary creates a link to "/editor/readme/URI/ID/FILENAME".
func buttonTextBinary(id, filename string) string {
	return buttonPreview("preview-binary", id, filename)
}

// buttonEXE returns a non-interactive terminal icon.
func buttonEXE() string {
	const title = `Known program or executable`
	return `<div class="col col-1 text-end" ` +
		`data-bs-toggle="tooltip" data-bs-title="` + title + `">` +
		`<svg width="16" height="16" fill="currentColor" aria-hidden="true">` +
		`<use xlink:href="/svg/bootstrap-icons.svg#terminal-plus"></use></svg></div>`
}

// buttonExtra creates a link to "/editor/helper/copy/ID/FILENAME".
func buttonExtra(id, filename string) string {
	const title = `Use file as an extra readme`
	const href = `#file-editor`
	const class = `icon-link align-text-bottom`
	const name = `artifact-editor-comp-hlpcopy`
	const indicator = `#artifact-editor-comp-htmx-indicator`
	const target = `#artifact-editor-comp-feedback`
	// while hard to read, this is the performant method of returning a complex string
	return `<div class="col col-1 text-end" data-bs-toggle="tooltip" data-bs-title="` + title + `">` +
		`<a href="` + href + `" class="` + class + `" name="` + name + `" ` +
		`hx-indicator="` + indicator + `" hx-target="` + target + `" ` +
		`hx-patch="/editor/helper/copy/` + id + `/` + filename + `">` +
		`<span class="badge bg-success text-dark">` +
		`<svg width="16" height="16" fill="currentColor" aria-hidden="true">` +
		`<use xlink:href="/svg/bootstrap-icons.svg#file-text"></use></svg>` +
		`</span></a></div>`
}

// buttonImage creates a link to "/editor/preview/copy/ID/FILENAME".
func buttonImage(id, filename string) string {
	const title = `Use image for preview`
	const class = `icon-link align-text-bottom`
	const name = `artifact-editor-comp-previewcopy`
	const indicator = `#artifact-editor-comp-htmx-indicator`
	const target = `#artifact-editor-comp-feedback`
	return `<div class="col col-1 text-end" data-bs-toggle="tooltip" data-bs-title="` + title + `">` +
		`<a class="` + class + `" name="` + name + `" ` +
		`hx-indicator="` + indicator + `" hx-target="` + target + `" ` +
		`hx-patch="/editor/preview/copy/` + id + `/` + filename + `">` +
		`<span class="badge text-bg-primary">` +
		`<svg width="16" height="16" fill="currentColor" aria-hidden="true">` +
		`<use xlink:href="/svg/bootstrap-icons.svg#images"></use></svg></span></a></div>`
}

// buttonPreview creates a link to "/editor/readme/URI/ID/FILENAME".
func buttonPreview(uri, id, filename string) string {
	const title = `Use file for preview`
	const href = `#file-editor`
	const class = `icon-link align-text-bottom`
	const name = `artifact-editor-comp-previewtext`
	const indicator = `#artifact-editor-comp-htmx-indicator`
	const target = `#artifact-editor-comp-feedback`
	return `<div class="col col-1 text-end" data-bs-toggle="tooltip" data-bs-title="` + title + `">` +
		`<a href="` + href + `" class="` + class + `" name="` + name + `" ` +
		`hx-indicator="` + indicator + `" hx-target="` + target + `" ` +
		`hx-patch="/editor/readme/` + uri + `/` + id + `/` + filename + `">` +
		`<span class="badge text-bg-secondary">` +
		`<svg width="16" height="16" fill="currentColor" aria-hidden="true">` +
		`<use xlink:href="/svg/bootstrap-icons.svg#images"></use></svg></span></a></div>`
}

// buttonReadme creates a link to "/editor/readme/copy/ID/FILENAME".
func buttonReadme(id, filename string) string {
	const title = `Use file as the readme`
	const href = `#file-editor`
	const class = `icon-link align-text-bottom`
	const name = `artifact-editor-comp-textcopy`
	const indicator = `#artifact-editor-comp-htmx-indicator`
	const target = `#artifact-editor-comp-feedback`
	return `<div class="col col-1 text-end" data-bs-toggle="tooltip" data-bs-title="` + title + `">` +
		`<a href="` + href + `" class="` + class + `" name="` + name + `" ` +
		`hx-indicator="` + indicator + `" hx-target="` + target + `" ` +
		`hx-patch="/editor/readme/copy/` + id + `/` + filename + `">` +
		`<span class="badge text-bg-success">` +
		`<svg class="bi" width="16" height="16" fill="currentColor" aria-hidden="true">` +
		`<use xlink:href="/svg/bootstrap-icons.svg#file-text"></use></svg></span></a></div>`
}

// buttonDIZ creates a link to "/editor/diz/copy/ID/FILENAME".
func buttonDIZ(id, filename string) string {
	const title = `Use file as the FILE_ID.DIZ`
	const href = `#file-editor`
	const class = `icon-link align-text-bottom`
	const name = `artifact-editor-comp-dizcopy`
	const indicator = `#artifact-editor-comp-htmx-indicator`
	const target = `#artifact-editor-comp-feedback`
	return `<div class="col col-1 text-end" data-bs-toggle="tooltip" data-bs-title="` + title + `">` +
		`<a href="` + href + `" class="` + class + `" name="` + name + `" ` +
		`hx-indicator="` + indicator + `" hx-target="` + target + `" ` +
		`hx-patch="/editor/diz/copy/` + id + `/` + filename + `">` +
		`<span class="badge text-bg-primary">` +
		`<svg class="bi" width="16" height="16" fill="currentColor" aria-hidden="true">` +
		`<use xlink:href="/svg/bootstrap-icons.svg#file-text"></use></svg></span></a></div>`
}

// columnFooter returns the brief metadata descriptions of the list entry.
func (m *ListEntry) columnFooter() string {
	ext := strings.ToLower(filepath.Ext(m.name))

	var sm string
	switch {
	case m.Texts && (ext == bat || ext == cmd):
		sm = "command script"
	case m.Texts && (ext == ini):
		sm = "configuration textfile"
	case m.Programs || ext == com:
		sm = progr(m.Executable, ext, m.bytes)
	case m.MusicConfig != "":
		sm = m.MusicConfig
	case m.Images:
		sm = m.ImageConfig
	default:
		sm = m.Signature
	}

	return `<div><small data-bs-toggle="tooltip" data-bs-title="` +
		strconv.Itoa(int(m.bytes)) + ` bytes">` + m.Filesize + `</small><span>` + sm + `</span></div>`
}

func progr(exec magicnumber.Windows, ext string, bytes int64) string {
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
	return ` <small>` + s + `</small>`
}

func progrDos(x8086 int, bytes int64) string {
	if x8086 >= int(bytes) {
		return "Dos command"
	}
	return "Dos executable"
}

// briefDescription returns true for known BBS/FTP site descriptor files.
func (m *ListEntry) briefDescription() bool {
	name := strings.TrimSpace(m.RelativeName)
	names := []string{"file_id.diz"} // FIX: add console and amiga
	for valid := range slices.Values(names) {
		if strings.EqualFold(name, valid) {
			return true
		}
	}

	return false
}

// xbinary, xbin or extended binary text is only partially supported.
func (m *ListEntry) xbinary() bool {
	return strings.EqualFold(m.Signature, magicnumber.XBinaryText.String())
}

// infotext returns true if the artifact is set as a "Text" or "Amiga text" platform,
// and the section is set to "NFO".
//
// Note, the following unsupported assets will always return false, "xbinary".
func (m *ListEntry) textNFO() bool {
	if m.xbinary() {
		return false // unsupported
	}

	simpleText := strings.EqualFold(m.platform, tags.Text.String()) || strings.EqualFold(m.platform, textamiga)
	if !simpleText {
		return false
	}

	return strings.EqualFold(m.section, tags.Nfo.String())
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

type Entry struct {
	module  string
	size    string
	format  string
	exec    magicnumber.Windows
	sign    magicnumber.Signature
	zeros   int
	bytes   int64
	image   bool
	text    bool
	bintext bool
	program bool
}

// SkipFile true if the named file should be skipped.
func (e *Entry) SkipFile(name, platform string) bool {
	info, err := os.Stat(name)
	if err != nil {
		return true
	}
	if info.IsDir() {
		return true
	}

	return e.skipByInfo(name, platform, info)
}

// SkipEntry parses the directory entry and returns true if it should be skipped.
// This is used to skip directories and files that are not relevant to the artifact,
// such as common DOS file extensions like .bat, .com, .exe, .cmd and .ini files.
func (e *Entry) SkipEntry(path string, d fs.DirEntry, platform string) bool {
	if d.IsDir() {
		return true
	}
	info, _ := d.Info()
	if info == nil {
		return true
	}

	const (
		batchScript = ".bat"
		batchCmd    = ".cmd"
		command     = ".com"
		executable  = ".exe"
		config      = ".ini"
	)
	switch strings.ToLower(filepath.Ext(path)) {
	case batchScript, batchCmd, command, executable, config:
		return true
	}

	return e.skipByInfo(path, platform, info)
}

func (e *Entry) skipByInfo(path, platform string, info fs.FileInfo) bool {
	if path == "" {
		return true
	}

	platform = strings.TrimSpace(platform)

	e.bytes = info.Size()
	if e.bytes == 0 {
		e.zeros++
		return true
	}

	e.size = humanize.Bytes(uint64(math.Abs(float64(info.Size()))))
	r, _ := os.Open(path)
	if r == nil {
		return true
	}
	defer r.Close()

	e.sign = magicnumber.Find(r)
	e.image = e.isImage()
	e.text = e.isText()
	e.bintext = e.isBinaryText(platform)
	e.program = e.isProgram(platform)
	switch {
	case e.image:
		return e.skipImage(path)
	case e.program:
		return e.skipProgram(path)
	case
		e.sign == magicnumber.MusicExtendedModule,
		e.sign == magicnumber.MusicMultiTrackModule,
		e.sign == magicnumber.MusicImpulseTracker,
		e.sign == magicnumber.MusicProTracker:
		return e.skipMusicMOD(path)
	case
		e.sign == magicnumber.MPEG1AudioLayer3,
		platform == tags.Audio.String():
		return e.skipMusic(path)
	}

	return false
}

// isImage returns truf if the magic signature is a known image format,
// both web friendly formats as well as obsolute formats.
func (e *Entry) isImage() bool {
	// unsupported formats that cannot be converted to browser friendly, .png,.webp,etc.
	if e.sign == magicnumber.RIPscrip {
		return false
	}

	for val := range slices.Values(magicnumber.Images()) {
		if val == e.sign {
			return true
		}
	}

	return false
}

func (e *Entry) isProgram(platform string) bool {
	for val := range slices.Values(magicnumber.Programs()) {
		if strings.EqualFold(platform, tags.DOS.String()) {
			break
		}

		if val == e.sign {
			return true
		}
	}

	return false
}

// isBinaryText returns true if the magic signature is binary data that does not match
// a known signature, and the platform is set to ANSI or Text.
//
// If unsupported XBIN binary text is detected, false is returned.
func (e *Entry) isBinaryText(platform string) bool {
	if e.sign == magicnumber.XBinaryText {
		return false
	}

	s := []tags.Tag{tags.Text, tags.ANSI}
	for tag := range slices.Values(s) {
		if strings.EqualFold(platform, tag.String()) {
			// unknown binary data
			if e.sign == magicnumber.Unknown {
				return true
			}
		}
	}

	return false
}

// isText returns true if the magic signature is a
// simple character encoded text file, or a
// unicoded encoded text file, or a text file with ansi escape codes.
func (e *Entry) isText() bool {
	for val := range slices.Values(magicnumber.Texts()) {
		if val == e.sign {
			return true
		}
	}

	return false
}

func (e *Entry) skipImage(path string) bool {
	r, _ := os.Open(path)
	if r == nil {
		return true
	}
	defer r.Close()

	config, imgtype, err := image.DecodeConfig(r)
	if err == nil {
		e.format = imgtype + " image, " + strconv.Itoa(config.Width) + "x" + strconv.Itoa(config.Height)
		return false
	}

	switch e.sign {
	case magicnumber.InterleavedBitmap:
		r, _ := os.Open(path)
		if r == nil {
			return true
		}
		defer func() { _ = r.Close() }()
		x, y := magicnumber.IlbmDecode(r)
		e.format = imgtype + "ILBM image, " + strconv.Itoa(x) + "x" + strconv.Itoa(y)
	default:
		e.format = e.sign.Title() + " image"
	}

	return false
}

func (e *Entry) skipProgram(path string) bool {
	r, _ := os.Open(path)
	if r == nil {
		return true
	}
	defer r.Close()

	exec, err := magicnumber.FindExecutable(r)
	if err == nil {
		e.exec = exec
	}

	return false
}

func (e *Entry) skipMusicMOD(path string) bool {
	r, _ := os.Open(path)
	if r == nil {
		return true
	}
	defer r.Close()

	e.module = magicnumber.MusicTracker(r)

	return false
}

// skipMusic parses the ID3 tag in the byte slice and returns the title, artist and year if available.
// It looks up in order the ID3v2.3, ID3v2.2 and ID3v1 tags in the byte slice with the priority being
// the newer versions of the tag.
//
// ID3v1 is a completely different tag format to ID3v2 and has serious limitations,
// so it is only used as a last resort.
func (e *Entry) skipMusic(path string) bool {
	// ID3 v2.x tags are located at the start of the file.
	id3, _ := os.Open(path)
	if id3 == nil {
		return true
	}
	defer id3.Close()

	if s := magicnumber.MusicID3v2(id3); s != "" {
		e.module = s
		return false
	}

	// ID3 v1 tags are located at the end of the file.
	if s := magicnumber.MusicID3v1(id3); s != "" {
		e.module = s
	}
	return false
}

func listErr(e Entry) ListEntry {
	return ListEntry{
		RelativeName: "",
		Signature:    e.sign.String(),
		Filesize:     e.size,
		ImageConfig:  e.format,
		MusicConfig:  e.module,
		UniqueID:     "",
		name:         "",
		platform:     "",
		section:      "",
		Executable:   e.exec,
		bytes:        e.bytes,
		Images:       e.image,
		Programs:     e.program,
		Texts:        e.text,
		BINtexts:     e.bintext,
	}
}

func listEntry(e Entry, rel, unid string) ListEntry {
	return ListEntry{
		RelativeName: LegacyString(rel),
		Signature:    e.sign.String(),
		Filesize:     e.size,
		ImageConfig:  e.format,
		MusicConfig:  e.module,
		UniqueID:     unid,
		name:         "",
		platform:     "",
		section:      "",
		Executable:   e.exec,
		bytes:        e.bytes,
		Images:       e.image,
		Programs:     e.program,
		Texts:        e.text,
		BINtexts:     e.bintext,
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
	const format = `<div class="border-bottom row mb-1">... skipped %d empty (0 B) files</div>`
	return fmt.Sprintf(format, zeroByteFiles)
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

	return strong(ys) + template.HTML(" "+ms+" "+ds)
}

// Dates returns the year, month and day for the published date for the artifact.
func Dates(art *models.File) (y int16, m int16, d int16) {
	if art == nil {
		return 0, 0, 0
	}

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
		r = r1 + ` + ` + r2
	case r1 != "":
		r = r1
	case r2 != "":
		r = r2
	}

	if strings.EqualFold(r, "independent") {
		s += " independently released"
	} else if !strings.EqualFold(r, "none") {
		s = s + " released by " + r
	}

	y := art.DateIssuedYear.Int16
	if y > 0 {
		s = s + " in " + strconv.Itoa(int(y))
	}

	return s + "."
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
func ExtraZip(art *models.File, extra dir.Directory) bool {
	if art == nil {
		return false
	}

	unid := UnID(art)
	name := filepath.Join(extra.Path(), unid+".zip")
	st, err := os.Stat(name)

	extraZip := 0
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
		const format = "Issue %d"
		return fmt.Sprintf(format, i)
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
		if valid := id > 0; valid {
			return strconv.FormatInt(id, 10)
		}
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
func JsdosMemory(art *models.File) (xms bool, ems bool, umb bool) {
	if art == nil {
		return false, false, false
	}

	if art.DoseeNoXMS.Valid {
		xms = art.DoseeNoXMS.Int16 == 0
	}
	if art.DoseeNoEms.Valid {
		ems = art.DoseeNoEms.Int16 == 0
	}
	if art.DoseeNoUmb.Valid {
		umb = art.DoseeNoUmb.Int16 == 0
	}

	return xms, ems, umb
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

// JsdosUse returns true if the file record is a known, MS-DOS executable.
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
	case exe, com: // programs
		return true
	case bat, cmd: // scripts
		return false
	default:
		return false
	}
}

// JsdosUsage returns true if the js-dos emulator should be used with the filename.
func JsdosUsage(filename, platform string) bool {
	platform = strings.TrimSpace(strings.ToLower(platform))
	if platform != "dos" {
		return false
	}

	ext := filepath.Ext(strings.ToLower(filename))
	switch ext {
	case ".zip", ".lhz", ".lzh", ".arc", ".arj":
		return true
	}
	switch ext {
	case exe, com: // programs
		return true
	case bat, cmd: // scripts
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
func LastModifications(art *models.File) (y int, m int, d int) {
	if art == nil {
		return 0, 0, 0
	}
	if !art.FileLastModified.Valid || art.FileLastModified.IsZero() {
		return 0, 0, 0
	}

	y = art.FileLastModified.Time.Year()
	m = int(art.FileLastModified.Time.Month())
	d = art.FileLastModified.Time.Day()

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
	if art == nil || art.ID == 0 {
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
	if x := strings.IndexByte(group, ' '); x > 0 {
		group = group[:x]
	}

	entries := strings.ReplaceAll(r.FileZipContent.String, "\r\n", "\n")

	// Avoid splitting the entire content into a slice to save memory
	return readme.Suggest(filename, group, entries)
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

	s := strings.Split(validate.Error(), ",")
	vals := make([]string, 0, len(s))
	for val := range slices.Values(s) {
		if strings.TrimSpace(val) == "" {
			continue
		}
		vals = append(vals, val)
	}

	vals = slices.Clip(vals)
	return strings.Join(vals, " + ")
}

// Relations returns the list of relationships for the file record.
func Relations(art *models.File) template.HTML {
	if art == nil {
		return ""
	}

	rels := art.ListRelations.String
	if rels == "" {
		return ""
	}

	links := strings.Split(rels, "|")
	if len(links) == 0 {
		return ""
	}

	const expected = 2
	const title = `Link to`
	const route = `/f/`
	const class = `fw-light text-secondary`

	var rows strings.Builder
	for link := range slices.Values(links) {
		s := strings.Split(link, ";")
		if len(s) != expected {
			continue
		}

		name, href := s[0], s[1]
		id := helper.DeObfuscate(href)
		if invalidID := id == href; invalidID {
			continue
		}

		if !strings.HasPrefix(href, route) {
			href = route + href
		}

		rows.WriteString(`<tr><th scope="row"><small class="` + class + `">` + title + `</small></th>` +
			`<td><small><a class="text-truncate" href="` + href + `">` + name + `</a></small></td></tr>`)
	}

	return template.HTML(rows.String())
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
func ReleaserPair(art *models.File) (relfor string, relby string) {
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

// UnsupportedFile returns true if an artifact should not be processed by the
// binary text or plain text file templates shown in the artifact pages.
//
// This includes artifacts that are classified as "documents" but include rich syntax
// such a HTML markup, PDF documents, or BBS RIP images.
func UnsupportedFile(art *models.File) bool {
	if art == nil {
		return true
	}

	const bbsRipImage = ".rip"
	if filepath.Ext(strings.ToLower(art.Filename.String)) == bbsRipImage {
		// the bbs era, remote images protcol is not supported
		// example: /f/b02392f
		return true
	}

	switch strings.TrimSpace(art.Platform.String) {
	case "markup", "pdf":
		return true
	}

	magic := strings.ToLower(strings.TrimSpace(art.FileMagicType.String))
	skips := slices.Concat(
		magicnumber.Images(),
		magicnumber.Programs(),
		magicnumber.Videos(),
	)

	// skips = append(skips, magicnumber.Unknown) // "Binary data"
	for skip := range slices.Values(skips) {
		if strings.EqualFold(skip.Title(), magic) {
			return true
		}
	}

	return false
}

// DisableReadme returns true if the readme or diz text files should not be displayed in the artifact page.
// This should be used sparingly and only for artifacts that have a readme file that is not useful or relevant.
func DisableReadme(art *models.File) bool {
	if art == nil {
		return false
	}
	return art.RetrotxtNoReadme.Int16 != 0
}

// Websites returns the list of links for the file record.
func Websites(art *models.File) template.HTML {
	if art == nil {
		return ""
	}

	lls := art.ListLinks.String
	if lls == "" {
		return ""
	}
	links := strings.Split(lls, "|")
	if len(links) == 0 {
		return ""
	}

	const expected = 2
	const title = `Link to`
	const class = `fw-light text-secondary`
	const aclass = `link-offset-3 icon-link icon-link-hover`

	var rows strings.Builder
	for link := range slices.Values(links) {
		s := strings.Split(link, ";")
		if len(s) != expected {
			continue
		}

		name, href := s[0], s[1]
		// Generally a stored URL will not include the protocol,
		// and will need to be prefixed with "https://".
		// There are some exceptions for websites that refuse to
		// implement HTTPS, such as http://textfiles.com.
		if !strings.HasPrefix(href, "http") {
			href = "https://" + href
		}
		if val, err := url.Parse(href); err != nil || val.Host == "" {
			continue
		}

		rows.WriteString(`<tr><th scope="row"><small class="` + class + `">` + title + `</small></th>` +
			`<td><small><a class="` + aclass + `" href="` + href + `">` + name + ` ` +
			string(LinkSVG()) + `</a></small></td></tr>`)
	}

	return template.HTML(rows.String())
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
