// Package mf provides functions for the file model which is an artifact record.
package mf

import (
	"fmt"
	"html/template"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/Defacto2/releaser"
	"github.com/Defacto2/server/handler/app/internal/exts"
	"github.com/Defacto2/server/handler/app/internal/readme"
	"github.com/Defacto2/server/handler/app/internal/str"
	"github.com/Defacto2/server/internal/archive"
	"github.com/Defacto2/server/internal/helper"
	"github.com/Defacto2/server/internal/magicnumber"
	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/Defacto2/server/internal/tags"
	"github.com/Defacto2/server/model"
	"github.com/dustin/go-humanize"
)

const (
	YYYYMMDD = "2006-Jan-02"

	epoch                   = model.EpochYear // epoch is the default year for MS-DOS files without a timestamp
	textamiga               = "textamiga"
	arrowLink template.HTML = `<svg class="bi" aria-hidden="true">` +
		`<use xlink:href="/svg/bootstrap-icons.svg#arrow-right"></use></svg>`
	br = "<br>"
)

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

// ListContent returns a list of the files contained in the archive file.
func ListContent(art *models.File, src string) template.HTML {
	if art == nil {
		return template.HTML(model.ErrModel.Error())
	}
	if !art.Platform.Valid {
		return "error, no platform"
	}

	// TODO: validate against string
	platform := strings.ToLower(art.Platform.String)

	unid := art.UUID.String
	if !art.UUID.Valid {
		return "error, no UUID"
	}

	const mb150 = 150 * 1024 * 1024
	if st, err := os.Stat(src); err != nil {
		return template.HTML(err.Error())
	} else if st.IsDir() {
		return "error, directory"
	} else if st.Size() > mb150 {
		return "will not decompress this archive as it is very large"
	}
	dst, err := helper.MkContent(src)
	if err != nil {
		return template.HTML(err.Error())
	}

	if entries, _ := os.ReadDir(dst); len(entries) == 0 {
		if err := archive.ExtractAll(src, dst); err != nil {
			defer os.RemoveAll(dst)
			return template.HTML(err.Error())
		}
	}

	files := 0
	walkerCount := func(path string, d fs.DirEntry, err error) error {
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
	items, zeroByteFiles := 0, 0
	const maxItems = 200
	walkerFunc := func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return fs.SkipDir
		}
		rel, err := filepath.Rel(dst, path)
		if err != nil {
			debug := fmt.Sprintf(`<div class="border-bottom row mb-1">... %v more files</div>`, err)
			b.WriteString(debug)
			return fs.SkipDir
		}
		if d.IsDir() {
			return fs.SkipDir
		}
		info, err := d.Info()
		if err != nil {
			return fs.SkipDir
		}
		bytes := info.Size()
		if bytes == 0 {
			zeroByteFiles++
			return fs.SkipDir
		}
		size := humanize.Bytes(uint64(info.Size()))
		image := false
		texts := false
		program := false
		r, err := os.Open(path)
		if err != nil {
			return fs.SkipDir
		}
		defer r.Close()
		sign, err := magicnumber.Find512B(r)
		if err != nil {
			return fs.SkipDir
		}
		for _, v := range magicnumber.Images() {
			if v == sign {
				image = true
				break
			}
		}
		for _, v := range magicnumber.Texts() {
			if v == sign {
				texts = true
				break
			}
		}
		for _, v := range magicnumber.Programs() {
			if strings.EqualFold(platform, tags.DOS.String()) {
				break
			}
			if v == sign {
				program = true
				break
			}
		}
		items++
		htm := fmt.Sprintf(`<div class="col d-inline-block text-truncate" data-bs-toggle="tooltip" data-bs-title="%s">%s</div>`,
			rel, rel)
		if image || texts {
			htm += `<div class="col col-1 text-end"><svg width="16" height="16" fill="currentColor" aria-hidden="true">` +
				`<use xlink:href="/svg/bootstrap-icons.svg#images"></use></svg></div>`
		} else {
			htm += `<div class="col col-1"></div>`
		}
		switch {
		case texts:
			name := url.QueryEscape(rel)
			htm += `<div class="col col-1 text-end">` +
				fmt.Sprintf(`<a class="icon-link align-text-bottom" hx-patch="/editor/readme/copy/%s/%s">`, unid, name) +
				`<svg class="bi" width="16" height="16" fill="currentColor" aria-hidden="true">` +
				`<use xlink:href="/svg/bootstrap-icons.svg#file-text"></use></svg></a></div>`
		case program:
			htm += `<div class="col col-1 text-end"><svg width="16" height="16" fill="currentColor" aria-hidden="true">` +
				`<use xlink:href="/svg/bootstrap-icons.svg#terminal-plus"></use></svg></div>`
		default:
			htm += `<div class="col col-1"></div>`
		}
		htm += fmt.Sprintf(`<div><small data-bs-toggle="tooltip" data-bs-title="%d bytes">%s</small>`, bytes, size)
		htm += fmt.Sprintf(` <small class="">%s</small></div>`, sign)
		htm = fmt.Sprintf(`<div class="border-bottom row mb-1">%s</div>`, htm)
		b.WriteString(htm)
		if items > maxItems {
			more := fmt.Sprintf(`<div class="border-bottom row mb-1">... %d more files</div>`, files-items)
			b.WriteString(more)
			return filepath.SkipAll
		}
		return nil
	}
	err = filepath.WalkDir(dst, walkerFunc)
	if err != nil {
		return template.HTML(err.Error())
	}
	if zeroByteFiles > 0 {
		zero := fmt.Sprintf(`<div class="border-bottom row mb-1">... skipped %d empty (0 B) files</div>`, zeroByteFiles)
		b.WriteString(zero)
	}
	return template.HTML(b.String())
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
	switch {
	case art.Createdat.Valid && art.Updatedat.Valid:
		c := str.Updated(art.Createdat.Time, "")
		u := str.Updated(art.Updatedat.Time, "")
		if c != u {
			c = str.Updated(art.Createdat.Time, "Created")
			u = str.Updated(art.Updatedat.Time, "Updated")
			return c + br + u
		}
		c = str.Updated(art.Createdat.Time, "Created")
		return c
	case art.Createdat.Valid:
		c := str.Updated(art.Createdat.Time, "Created")
		return c
	case art.Updatedat.Valid:
		u := str.Updated(art.Updatedat.Time, "Updated")
		return u
	}
	return ""
}

// FirstHeader returns the title of the file,
// unless the artifact is marked as a magazine issue, in which case it returns the issue number.
func FirstHeader(art *models.File) string {
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
	case ".exe", ".com":
		return true
	case ".bat", ".cmd":
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
	const none = "No recorded timestamp"
	if !art.FileLastModified.Valid {
		return none
	}
	year, _ := strconv.Atoi(art.FileLastModified.Time.Format("2006"))
	if year <= epoch {
		return none
	}
	return str.Updated(art.FileLastModified.Time, "Modified")
}

// LinkPreview returns a URL path to link to the file record in tab, to use as a preview.
// A preview link is only available for certain file types such as images, text, documents, and
// renders the whole item in its own browser tab without any HTML or CSS from the website.
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
	case slices.Contains(exts.Archives(), ext):
		// this must always be first
		return ""
	case platform == textamiga, platform == "text":
		break
	case slices.Contains(exts.Documents(), ext):
		break
	case slices.Contains(exts.Images(), ext):
		break
	case slices.Contains(exts.Media(), ext):
		break
	default:
		return ""
	}
	s, err := str.LinkID(id, "v")
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
	return str.LinkPreviewTip(name, platform)
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

// ReadmeNone returns true if the file record should not display the text file content in the artifact page.
func ReadmeNone(art *models.File) bool {
	if art == nil {
		return false
	}
	if art.RetrotxtNoReadme.Valid {
		return art.RetrotxtNoReadme.Int16 != 0
	}
	return false
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
	return art.Deletedat.Time.IsZero()
}

// RecordProblems returns a list of validation problems for the file record.
func RecordProblems(art *models.File) string {
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
	pair := str.ReleaserPair(art.GroupBrandFor, art.GroupBrandBy)
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
	if !art.ListLinks.Valid {
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
