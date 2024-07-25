// Package mf provides functions for the file model which is an artifact record
package mf

import (
	"fmt"
	"html/template"
	"io/fs"
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
	epoch                   = model.EpochYear // epoch is the default year for MS-DOS files without a timestamp
	textamiga               = "textamiga"
	arrowLink template.HTML = `<svg class="bi" aria-hidden="true">` +
		`<use xlink:href="/svg/bootstrap-icons.svg#arrow-right"></use></svg>`
	br = "<br>"
)

func AlertURL(art *models.File) string {
	if art == nil {
		return ""
	}
	// todo: confirm link is a valid url?
	if art.FileSecurityAlertURL.Valid {
		return strings.TrimSpace(art.FileSecurityAlertURL.String)
	}
	return ""
}

func AttrArtist(art *models.File) string {
	if art == nil {
		return ""
	}
	if art.CreditIllustration.Valid {
		return art.CreditIllustration.String
	}
	return ""
}

func AttrMusic(art *models.File) string {
	if art == nil {
		return ""
	}
	if art.CreditAudio.Valid {
		return art.CreditAudio.String
	}
	return ""
}

func AttrProg(art *models.File) string {
	if art == nil {
		return ""
	}
	if art.CreditProgram.Valid {
		return art.CreditProgram.String
	}
	return ""
}

func AttrWriter(art *models.File) string {
	if art == nil {
		return ""
	}
	if art.CreditText.Valid {
		return art.CreditText.String
	}
	return ""
}

func Basename(art *models.File) string {
	if art == nil {
		return ""
	}
	if art.Filename.Valid {
		return art.Filename.String
	}
	return ""
}

func Checksum(art *models.File) string {
	if art == nil {
		return ""
	}
	if art.FileIntegrityStrong.Valid {
		return strings.TrimSpace(art.FileIntegrityStrong.String)
	}
	return ""
}

func Comment(art *models.File) string {
	if art == nil {
		return ""
	}
	if art.Comment.Valid {
		return art.Comment.String
	}
	return ""
}

func Content(art *models.File, src string) template.HTML {
	if art == nil {
		return template.HTML(model.ErrModel.Error())
	}
	if !art.Platform.Valid {
		return "error, no platform"
	}

	// TODO: validate against string
	platform := strings.ToLower(art.Platform.String)

	const mb150 = 150 * 1024 * 1024
	if st, err := os.Stat(src); err != nil {
		return template.HTML(err.Error())
	} else if st.IsDir() {
		return "error, directory"
	} else if st.Size() > mb150 {
		return "will not decompress this archive as it is very large"
	}
	dst, err := str.ContentSRC(src)
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
	var walkerCount = func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			return nil
		}
		files++
		return nil
	}
	if err := filepath.WalkDir(dst, walkerCount); err != nil {
		return template.HTML(err.Error())
	}

	var b strings.Builder
	items, zeroByteFiles := 0, 0
	var walkerFunc = func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		rel, err := filepath.Rel(dst, path)
		if err != nil {
			debug := fmt.Sprintf(`<div class="border-bottom row mb-1">... %v more files</div>`, err)
			b.WriteString(debug)
			return nil
		}
		if d.IsDir() {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return nil
		}
		bytes := info.Size()
		if bytes == 0 {
			zeroByteFiles++
			return nil
		}
		size := humanize.Bytes(uint64(info.Size()))
		image := false
		texts := false
		program := false
		r, err := os.Open(path)
		if err != nil {
			return nil
		}
		defer r.Close()
		sign, err := magicnumber.Find512B(r)
		if err != nil {
			return nil
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
		if texts {
			htm += `<div class="col col-1 text-end"><svg width="16" height="16" fill="currentColor" aria-hidden="true">` +
				`<use xlink:href="/svg/bootstrap-icons.svg#file-text"></use></svg></div>`
		} else if program {
			htm += `<div class="col col-1 text-end"><svg width="16" height="16" fill="currentColor" aria-hidden="true">` +
				`<use xlink:href="/svg/bootstrap-icons.svg#terminal-plus"></use></svg></div>`
		} else {
			htm += `<div class="col col-1"></div>`
		}
		htm += fmt.Sprintf(`<div><small data-bs-toggle="tooltip" data-bs-title="%d bytes">%s</small>`, bytes, size)
		htm += fmt.Sprintf(` <small class="">%s</small></div>`, sign)
		htm = fmt.Sprintf(`<div class="border-bottom row mb-1">%s</div>`, htm)
		b.WriteString(htm)
		if items > 200 {
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

// Date returns a formatted date string for the artifact's published date.
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

func DownloadID(art *models.File) string {
	if art == nil {
		return ""
	}
	return helper.ObfuscateID(art.ID)
}

func ExtraZip(art *models.File, extraDir string) bool {
	extraZip := 0
	unid := UnID(art)
	st, err := os.Stat(filepath.Join(extraDir, unid+".zip"))
	if err == nil && !st.IsDir() {
		extraZip = int(st.Size())
	}
	return extraZip > 0
}

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
// unless the file is a magazine issue, in which case it returns the issue number.
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

func Idenfication16C(art *models.File) string {
	if art == nil {
		return ""
	}
	if art.WebID16colors.Valid {
		return art.WebID16colors.String
	}
	return ""
}

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

func IdenficationGitHub(art *models.File) string {
	if art == nil {
		return ""
	}
	if art.WebIDGithub.Valid {
		return art.WebIDGithub.String
	}
	return ""
}

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

func IdenficationYT(art *models.File) string {
	if art == nil {
		return ""
	}
	if art.WebIDYoutube.Valid {
		return strings.TrimSpace(art.WebIDYoutube.String)
	}
	return ""
}

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

func JsdosBroken(art *models.File) bool {
	if art == nil {
		return false
	}
	if art.DoseeIncompatible.Valid {
		return art.DoseeIncompatible.Int16 != 0
	}
	return false
}

func JsdosCPU(art *models.File) string {
	if art == nil {
		return ""
	}
	if art.DoseeHardwareCPU.Valid {
		return art.DoseeHardwareCPU.String
	}
	return ""
}

func JsdosMachine(art *models.File) string {
	if art == nil {
		return ""
	}
	if art.DoseeHardwareGraphic.Valid {
		return art.DoseeHardwareGraphic.String
	}
	return ""
}

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

func JsdosRun(art *models.File) string {
	if art == nil {
		return ""
	}
	if art.DoseeRunProgram.Valid {
		return art.DoseeRunProgram.String
	}
	return ""
}

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

func JsdosUtilities(art *models.File) bool {
	if art == nil {
		return false
	}
	if art.DoseeLoadUtilities.Valid {
		return art.DoseeLoadUtilities.Int16 != 0
	}
	return false
}

// LastModification returns the last modified date for the file record.
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

// lastModificationAgo returns the last modified date in a human readable format.
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

func Magic(art *models.File) string {
	if art == nil {
		return ""
	}
	if art.FileMagicType.Valid {
		return strings.TrimSpace(art.FileMagicType.String)
	}
	return ""
}

func Platform(art *models.File) string {
	if art == nil {
		return ""
	}
	// todo test against tag library
	if art.Platform.Valid {
		return strings.TrimSpace(art.Platform.String)
	}
	return ""
}

func ReadmeNone(art *models.File) bool {
	if art == nil {
		return false
	}
	if art.RetrotxtNoReadme.Valid {
		return art.RetrotxtNoReadme.Int16 != 0
	}
	return false
}

// Readme returns a suggested readme file name for the record.
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

func RecordIsNew(art *models.File) bool {
	if art == nil {
		return false
	}
	return !art.Deletedat.IsZero() && art.Deletedby.IsZero()
}

func RecordOffline(art *models.File) bool {
	if art == nil {
		return false
	}
	return !art.Deletedat.IsZero() && !art.Deletedby.IsZero()
}

func RecordOnline(art *models.File) bool {
	return art.Deletedat.Time.IsZero()
}

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

func RelationsStr(art *models.File) string {
	if art == nil {
		return ""
	}
	if art.ListRelations.Valid {
		return strings.TrimSpace(art.ListRelations.String)
	}
	return ""
}

func ReleaserPair(art *models.File) (string, string) {
	if art == nil {
		return "", ""
	}
	pair := str.ReleaserPair(art.GroupBrandFor, art.GroupBrandBy)
	return pair[0], pair[1]

}

func Section(art *models.File) string {
	if art == nil {
		return ""
	}
	// TODO: validate using the tag pkg?
	if art.Section.Valid {
		return strings.TrimSpace(art.Section.String)
	}
	return ""
}

func TagCategory(art *models.File) string {
	if art == nil {
		return ""
	}
	// todo: validate against tags library
	if art.Section.Valid {
		return strings.ToLower(strings.TrimSpace(art.Section.String))
	}
	return ""
}

func TagOS(art *models.File) string {
	if art == nil {
		return ""
	}
	// todo: validate against tags library
	if art.Platform.Valid {
		return strings.ToLower(strings.TrimSpace(art.Platform.String))
	}
	return ""
}

func Title(art *models.File) string {
	if art == nil {
		return ""
	}
	return art.RecordTitle.String
}

func UnID(art *models.File) string {
	if art == nil {
		return ""
	}
	if art.UUID.Valid {
		return art.UUID.String
	}
	return ""
}

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

func WebsitesStr(art *models.File) string {
	if art == nil {
		return ""
	}
	if !art.ListLinks.Valid {
		return strings.TrimSpace(art.ListLinks.String)
	}
	return ""
}

func ZipContent(art *models.File) string {
	if art == nil {
		return ""
	}
	if art.FileZipContent.Valid {
		return strings.TrimSpace(art.FileZipContent.String)
	}
	return ""
}
