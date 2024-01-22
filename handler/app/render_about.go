package app

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"image"
	_ "image/gif"  // gif format decoder
	_ "image/jpeg" // jpeg format decoder
	_ "image/png"  // png format decoder
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/Defacto2/releaser"
	"github.com/Defacto2/server/internal/helper"
	"github.com/Defacto2/server/internal/magic"
	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/Defacto2/server/internal/render"
	"github.com/Defacto2/server/model"
	"github.com/dustin/go-humanize"
	"github.com/h2non/filetype"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
	_ "golang.org/x/image/webp" // webp format decoder
	"golang.org/x/text/encoding/charmap"
)

// Dirs contains the directories used by the about pages.
type Dirs struct {
	Download  string // path to the artifact download directory
	Preview   string // path to the preview and screenshot directory
	Thumbnail string // path to the file thumbnail directory
	URI       string // the URI of the file record
}

// About is the handler for the about page of the file record.
func (dir Dirs) About(z *zap.SugaredLogger, c echo.Context, readonly bool) error { //nolint:funlen
	const name = "about"
	if z == nil {
		return InternalErr(z, c, name, ErrZap)
	}
	res, err := model.OneRecord(z, c, dir.URI)
	if err != nil {
		if errors.Is(err, model.ErrID) {
			return AboutErr(z, c, dir.URI)
		}
		return DatabaseErr(z, c, "f/"+dir.URI, err)
	}
	fname := res.Filename.String
	uuid := res.UUID.String
	abs := filepath.Join(dir.Download, uuid)
	data := empty(c)
	// about editor
	if !readonly {
		data["readonly"] = false
		data["recID"] = res.ID
		data["recTitle"] = res.RecordTitle.String
		data["recOnline"] = res.Deletedat.Time.IsZero()
		data["recReleasers"] = string(RecordRels(res.GroupBrandBy, res.GroupBrandFor))
		data["recYear"] = res.DateIssuedYear.Int16
		data["recMonth"] = res.DateIssuedMonth.Int16
		data["recDay"] = res.DateIssuedDay.Int16
		data["recLastMod"] = res.FileLastModified.IsZero()
		data["recLastModValue"] = res.FileLastModified.Time.Format("2006-1-2") // value should not have no leading zeros
		data["recAbsDownload"] = abs
		data["recKind"] = aboutMagic(abs)
		data["recStatMod"] = aboutStat(abs)[0]
		data["recStatSize"] = aboutStat(abs)[1]
		data["recAssets"] = dir.aboutAssets(uuid)
		data["recNoReadme"] = res.RetrotxtNoReadme.Int16 != 0
		data["recReadmeList"] = OptionsReadme(res.FileZipContent.String)
		data["recPreviewList"] = OptionsPreview(res.FileZipContent.String)
		data["recAnsiLoveList"] = OptionsAnsiLove(res.FileZipContent.String)
		data["recReadmeSug"] = readmeSuggest(res)
		data["recZipContent"] = strings.TrimSpace(res.FileZipContent.String)
		data["recOS"] = strings.ToLower(strings.TrimSpace(res.Platform.String))
		data["recTag"] = strings.ToLower(strings.TrimSpace(res.Section.String))
	}
	// page metadata
	data["uuid"] = uuid
	data["download"] = helper.ObfuscateID(res.ID)
	data["title"] = fname
	data["description"] = aboutDesc(res)
	data["h1"] = aboutIssue(res)
	data["lead"] = aboutLead(res)
	data["comment"] = res.Comment.String
	// file metadata
	data["filename"] = fname
	data["filesize"] = helper.ByteCount(res.Filesize.Int64)
	data["filebyte"] = res.Filesize.Int64
	data["lastmodified"] = aboutLM(res)
	data["lastmodifiedAgo"] = aboutModAgo(res)
	data["checksum"] = strings.TrimSpace(res.FileIntegrityStrong.String)
	data["magic"] = res.FileMagicType.String
	data["releasers"] = string(LinkRels(res.GroupBrandBy, res.GroupBrandFor))
	data["published"] = model.PublishedFmt(res)
	data["section"] = strings.TrimSpace(res.Section.String)
	data["platform"] = strings.TrimSpace(res.Platform.String)
	data["alertURL"] = res.FileSecurityAlertURL.String
	// attributions and credits
	data["writers"] = res.CreditText.String
	data["artists"] = res.CreditIllustration.String
	data["programmers"] = res.CreditProgram.String
	data["musicians"] = res.CreditAudio.String
	// links to other records and sites
	data["listLinks"] = aboutLinks(res)
	data["listReleases"] = res.ListRelations.String
	data["listWebsites"] = res.ListLinks.String
	data["demozoo"] = aboutID(res.WebIDDemozoo.Int64)
	data["pouet"] = aboutID(res.WebIDPouet.Int64)
	data["sixteenColors"] = res.WebID16colors.String
	data["youtube"] = res.WebIDYoutube.String
	data["github"] = res.WebIDGithub.String
	// file archive content
	data["jsdos"] = aboutJSDos(res)
	ctt := aboutCtt(res)
	data["content"] = ctt
	data["contentDesc"] = ""
	if len(ctt) == 1 {
		data["contentDesc"] = "contains one file"
	}
	if len(ctt) > 1 {
		data["contentDesc"] = fmt.Sprintf("contains %d files", len(ctt))
	}
	// record metadata
	data["linkpreview"] = LinkPreviewHref(res.ID, res.Filename.String, res.Platform.String)
	data["linkpreviewTip"] = LinkPreviewTip(res.Filename.String, res.Platform.String)
	switch {
	case res.Createdat.Valid && res.Updatedat.Valid:
		c := Updated(res.Createdat.Time, "")
		u := Updated(res.Updatedat.Time, "")
		if c != u {
			c = Updated(res.Createdat.Time, "Created")
			u = Updated(res.Updatedat.Time, "Updated")
			data["filentry"] = c + "<br>" + u
		} else {
			c = Updated(res.Createdat.Time, "Created")
			data["filentry"] = c
		}
	case res.Createdat.Valid:
		c := Updated(res.Createdat.Time, "Created")
		data["filentry"] = c
	case res.Updatedat.Valid:
		u := Updated(res.Updatedat.Time, "Updated")
		data["filentry"] = u
	}
	d, err := dir.aboutReadme(res)
	if err != nil {
		return InternalErr(z, c, name, err)
	}
	maps.Copy(data, d)
	err = c.Render(http.StatusOK, name, data)
	if err != nil {
		return InternalErr(z, c, name, err)
	}
	return nil
}

func (dir Dirs) aboutReadme(res *models.File) (map[string]interface{}, error) { //nolint:funlen
	data := map[string]interface{}{}
	if res.RetrotxtNoReadme.Int16 != 0 {
		return data, nil
	}
	platform := strings.TrimSpace(res.Platform.String)
	switch platform {
	case "markup", "pdf":
		return data, nil
	}
	if render.NoScreenshot(dir.Preview, res) {
		data["noScreenshot"] = true
	}
	// the bbs era, remote images protcol is not supported
	// example: /f/b02392f
	const ripScrip = ".rip"
	if filepath.Ext(strings.ToLower(res.Filename.String)) == ripScrip {
		return data, nil
	}

	b, err := render.Read(dir.Download, res)
	if errors.Is(err, render.ErrDownload) {
		data["noDownload"] = true
		return data, nil
	}
	if err != nil {
		return data, err
	}
	if b == nil || render.IsUTF16(b) {
		return data, nil
	}

	contentType := http.DetectContentType(b)
	switch contentType {
	case "archive/zip", "application/zip", "application/octet-stream":
		return data, nil
	}

	// Remove control codes and metadata from byte array
	const (
		reAnsi  = `\x1b\[[0-9;]*[a-zA-Z]` // ANSI escape codes
		reAmiga = `\x1b\[[0-9;]*[ ]p`     // unknown control code found in Amiga texts
		reSauce = `SAUCE00.*`             // SAUCE metadata that is appended to some files
	)
	re := regexp.MustCompile(reAnsi + `|` + reAmiga + `|` + reSauce)
	b = re.ReplaceAll(b, []byte{})

	e := render.Encoder(res, b...)
	const (
		sp      = 0x20 // space
		hyphen  = 0x2d // hyphen-minus
		shy     = 0xad // soft hyphen for ISO8859-1
		nbsp    = 0xa0 // non-breaking space for ISO8859-1
		nbsp437 = 0xff // non-breaking space for CP437
	)
	switch e {
	case charmap.ISO8859_1:
		data["readmeLatin1Cls"] = ""
		data["readmeCP437Cls"] = "d-none"
		data["topazCheck"] = "checked"
		b = bytes.ReplaceAll(b, []byte{nbsp}, []byte{sp})
		b = bytes.ReplaceAll(b, []byte{shy}, []byte{hyphen})
	case charmap.CodePage437:
		data["readmeLatin1Cls"] = "d-none"
		data["readmeCP437Cls"] = ""
		data["vgaCheck"] = "checked"
		b = bytes.ReplaceAll(b, []byte{nbsp437}, []byte{sp})
	}

	// render both ISO8859 and CP437 encodings of the readme
	// and let the user choose which one to display
	r := charmap.ISO8859_1.NewDecoder().Reader(bytes.NewReader(b))
	out := strings.Builder{}
	if _, err := io.Copy(&out, r); err != nil {
		return data, err
	}
	if !strings.HasSuffix(out.String(), "\n\n") {
		out.WriteString("\n")
	}
	data["readmeLatin1"] = out.String()
	r = charmap.CodePage437.NewDecoder().Reader(bytes.NewReader(b))
	out = strings.Builder{}
	if _, err := io.Copy(&out, r); err != nil {
		return data, err
	}
	if !strings.HasSuffix(out.String(), "\n\n") {
		out.WriteString("\n")
	}
	data["readmeCP437"] = out.String()

	data["readmeLines"] = strings.Count(out.String(), "\n")
	data["readmeRows"] = helper.MaxLineLength(out.String())

	return data, nil
}

func aboutDesc(res *models.File) string {
	s := res.Filename.String
	if res.RecordTitle.String != "" {
		s = aboutIssue(res)
	}
	r1 := releaser.Clean(strings.ToLower(res.GroupBrandBy.String))
	r2 := releaser.Clean(strings.ToLower(res.GroupBrandFor.String))
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
	y := res.DateIssuedYear.Int16
	if y > 0 {
		s = fmt.Sprintf("%s in %d", s, y)
	}
	return s
}

func aboutIssue(res *models.File) string {
	sect := strings.TrimSpace(strings.ToLower(res.Section.String))
	if sect != "magazine" {
		return res.RecordTitle.String
	}
	s := res.RecordTitle.String
	if i, err := strconv.Atoi(s); err == nil {
		return fmt.Sprintf("Issue %d", i)
	}
	return s
}

func aboutLead(res *models.File) string {
	fname := res.Filename.String
	span := fmt.Sprintf("<span class=\"font-monospace fs-6 fw-light\">%s</span> ", fname)
	rels := string(LinkRels(res.GroupBrandBy, res.GroupBrandFor))
	return fmt.Sprintf("%s<br>%s", rels, span)
}

func aboutLM(res *models.File) string {
	const none = "no timestamp"
	if !res.FileLastModified.Valid {
		return none
	}
	year, _ := strconv.Atoi(res.FileLastModified.Time.Format("2006"))
	const epoch = 1980
	if year <= epoch {
		// 1980 is the default date for MS-DOS files without a timestamp
		return none
	}
	lm := res.FileLastModified.Time.Format("2006 Jan 2, 15:04")
	if lm == "0001 Jan 1, 00:00" {
		return none
	}
	return lm
}

func aboutMagic(name string) string {
	file, err := os.Open(name)
	if err != nil {
		return err.Error()
	}
	defer file.Close()

	const sample = 512
	head := make([]byte, sample)
	_, err = file.Read(head)
	if err != nil {
		return err.Error()
	}

	// add custom magic matchers
	// filetype.AddMatcher(magic.ANSIType, magic.ANSIMatcher) // todo: this is creating false positives with ZIP archives
	filetype.AddMatcher(magic.ArcSeaType(), magic.ArcSeaMatcher)
	filetype.AddMatcher(magic.ARJType(), magic.ARJMatcher)
	filetype.AddMatcher(magic.DOSComType(), magic.DOSComMatcher)
	filetype.AddMatcher(magic.InterchangeType(), magic.InterchangeMatcher)
	filetype.AddMatcher(magic.PCXType(), magic.PCXMatcher)
	kind, err := filetype.Match(head)
	if err != nil {
		return err.Error()
	}
	if kind != filetype.Unknown {
		return kind.MIME.Value
	}

	return http.DetectContentType(head)
}

func aboutModAgo(res *models.File) string {
	if !res.FileLastModified.Valid {
		return ""
	}
	year, _ := strconv.Atoi(res.FileLastModified.Time.Format("2006"))
	const epoch = 1980
	if year <= epoch {
		// 1980 is the default date for MS-DOS files without a timestamp
		return ""
	}
	return Updated(res.FileLastModified.Time, "Modified")
}

func aboutCtt(res *models.File) []string {
	conts := strings.Split(res.FileZipContent.String, "\n")
	conts = slices.DeleteFunc(conts, func(s string) bool {
		return strings.TrimSpace(s) == "" // delete empty lines
	})
	conts = slices.Compact(conts)
	return conts
}

func aboutLinks(res *models.File) template.HTML {
	s := res.ListLinks.String
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
		rows += fmt.Sprintf("<tr><th scope=\"row\"><small>Link</small></th>"+
			"<td><small><a class=\"text-truncate\" href=\"%s\">%s</a></small></td></tr>", href, name)
	}
	return template.HTML(rows) //nolint:gosec
}

func aboutJSDos(res *models.File) bool {
	if strings.TrimSpace(strings.ToLower(res.Platform.String)) != "dos" {
		return false
	}
	// check supported filename extensions
	ext := filepath.Ext(strings.ToLower(res.Filename.String))
	switch ext {
	case ".zip":
		// ".exe", ".com", /f/b03550
		// legacy zip, not supported, /f/a319104
		return true
	default:
		return false
	}
}

func aboutID(id int64) string {
	if id == 0 {
		return ""
	}
	return strconv.FormatInt(id, 10)
}

// aboutStat returns the file last modified date and formatted file size.
func aboutStat(name string) [2]string {
	stat, err := os.Stat(name)
	if err != nil {
		return [2]string{err.Error(), err.Error()}
	}
	return [2]string{
		stat.ModTime().Format("2006-Jan-02"),
		fmt.Sprintf("%s bytes - %s - %s",
			humanize.Comma(stat.Size()),
			humanize.Bytes(uint64(stat.Size())),
			humanize.IBytes(uint64(stat.Size()))),
	}
}

// aboutAssets returns a list of downloads and image assets belonging to the file record.
// any errors are appended to the list.
func (dir Dirs) aboutAssets(uuid string) map[string]string {
	matches := map[string]string{}

	downloads, err := os.ReadDir(dir.Download)
	if err != nil {
		matches[err.Error()] = ""
	}
	images, err := os.ReadDir(dir.Preview)
	if err != nil {
		matches[err.Error()] = ""
	}
	thumbs, err := os.ReadDir(dir.Thumbnail)
	if err != nil {
		matches[err.Error()] = ""
	}

	for _, file := range downloads {
		if strings.HasPrefix(file.Name(), uuid) {
			if filepath.Ext(file.Name()) == "" {
				continue
			}
			s := strings.ToUpper(filepath.Ext(file.Name()))
			st, err := file.Info()
			if err != nil {
				matches[err.Error()] = err.Error()
			}
			switch s {
			case ".TXT":
				s = ".TXT readme"
				i, _ := helper.Lines(filepath.Join(dir.Download, file.Name()))
				matches[s] = fmt.Sprintf("%s bytes - %d lines", humanize.Comma(st.Size()), i)
			case ".ZIP":
				s = ".ZIP for emulator"
				matches[s] = fmt.Sprintf("%s bytes", humanize.Comma(st.Size()))
			}
		}
	}
	for _, file := range images {
		if strings.HasPrefix(file.Name(), uuid) {
			s := strings.ToUpper(filepath.Ext(file.Name()))
			if s == ".WEBP" {
				s = ".WebP"
			}
			matches[s+" preview "] = aboutImgInfo(filepath.Join(dir.Preview, file.Name()))
		}
	}
	for _, file := range thumbs {
		if strings.HasPrefix(file.Name(), uuid) {
			s := strings.ToUpper(filepath.Ext(file.Name()))
			if s == ".WEBP" {
				s = ".WebP"
			}
			matches[s+" thumb"] = aboutImgInfo(filepath.Join(dir.Thumbnail, file.Name()))
		}
	}

	return matches
}

func aboutImgInfo(name string) string {
	switch filepath.Ext(strings.ToLower(name)) {
	case ".jpg", ".jpeg", ".gif", ".png", ".webp":
	default:
		st, err := os.Stat(name)
		if err != nil {
			return err.Error()
		}
		return fmt.Sprintf("%s bytes", humanize.Comma(st.Size()))
	}
	reader, err := os.Open(name)
	if err != nil {
		return err.Error()
	}
	defer reader.Close()
	st, err := reader.Stat()
	if err != nil {
		return err.Error()
	}
	config, _, err := image.DecodeConfig(reader)
	if err != nil {
		return err.Error()
	}
	return fmt.Sprintf("%s bytes - %d x %d pixels", humanize.Comma(st.Size()), config.Width, config.Height)
}

// readmeSuggest returns a suggested readme file name for the record.
func readmeSuggest(r *models.File) string {
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
	return ReadmeSug(filename, group, content...)
}

func readmeFinds(content ...string) []string {
	finds := []string{}
	skip := []string{"scene.org", "scene.org.txt"}
	for _, name := range content {
		if name == "" {
			continue
		}
		s := strings.ToLower(name)
		if slices.Contains(skip, s) {
			continue
		}
		ext := filepath.Ext(s)
		if slices.Contains(priority(), ext) {
			finds = append(finds, name)
			continue
		}
		if slices.Contains(candidate(), ext) {
			finds = append(finds, name)
		}
	}
	return finds
}

// ReadmeSug returns a suggested readme file name for the record.
// It prioritizes the filename and group name with a priority extension,
// such as ".nfo", ".txt", etc. If no priority extension is found,
// it will return the first textfile in the content list.
//
// The filename should be the name of the file archive artifact.
// The group should be a name or common abbreviation of the group that
// released the artifact. The content should be a list of files contained
// in the artifact.
func ReadmeSug(filename, group string, content ...string) string {
	// this is a port of the CFML function, variables.findTextfile found in File.cfc

	finds := readmeFinds(content...)
	if len(finds) == 1 {
		return finds[0]
	}
	finds = SortContent(finds)

	// match either the filename or the group name with a priority extension
	// e.g. .nfo, .txt, .unp, .doc
	base := filepath.Base(filename)
	for _, ext := range priority() {
		for _, name := range finds {
			// match the filename + extension
			if strings.EqualFold(base+ext, name) {
				return name
			}
			// match the group name + extension
			if strings.EqualFold(group+ext, name) {
				return name
			}
		}
	}
	// match file_id.diz
	for _, name := range finds {
		if strings.EqualFold("file_id.diz", name) {
			return name
		}
	}
	// match either the filename or the group name with a candidate extension
	for _, ext := range candidate() {
		for _, name := range finds {
			// match the filename + extension
			if strings.EqualFold(base+ext, name) {
				return name
			}
			// match the group name + extension
			if strings.EqualFold(group+ext, name) {
				return name
			}
		}
	}
	// match any finds that use a priority extension
	for _, name := range finds {
		s := strings.ToLower(name)
		ext := filepath.Ext(s)
		if slices.Contains(priority(), ext) {
			return name
		}
	}
	// match the first file in the list
	for _, name := range finds {
		return name
	}
	return ""
}

// priority returns a list of readme text file extensions in priority order.
func priority() []string {
	return []string{".nfo", ".txt", ".unp", ".doc"}
}

// candidate returns a list of other, common text file extensions in priority order.
func candidate() []string {
	return []string{".diz", ".asc", ".1st", ".dox", ".me", ".cap", ".ans", ".pcb"}
}

// SortContent sorts the content list by the number of slashes in each string.
// It prioritizes strings with fewer slashes (i.e., closer to the root).
// If the number of slashes is the same, it sorts alphabetically.
func SortContent(content []string) []string {
	sort.Slice(content, func(i, j int) bool {
		// Fix any Windows path separators
		content[i] = strings.ReplaceAll(content[i], "\\", "/")
		content[j] = strings.ReplaceAll(content[j], "\\", "/")
		// Count the number of slashes in each string
		iCount := strings.Count(content[i], "/")
		jCount := strings.Count(content[j], "/")

		// Prioritize strings with fewer slashes (i.e., closer to the root)
		if iCount != jCount {
			return iCount < jCount
		}

		// If the number of slashes is the same, sort alphabetically
		return content[i] < content[j]
	})

	return content
}
