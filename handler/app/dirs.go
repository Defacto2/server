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
	"strconv"
	"strings"

	"github.com/Defacto2/releaser"
	"github.com/Defacto2/server/handler/sess"
	"github.com/Defacto2/server/internal/command"
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

// Dirs contains the directories used by the artifact pages.
type Dirs struct {
	Download  string // path to the artifact download directory
	Preview   string // path to the preview and screenshot directory
	Thumbnail string // path to the file thumbnail directory
	URI       string // the URI of the file record
}

type extract int // extract target format for the file archive extractor

const (
	imgs  extract = iota // extract image
	ansis                // extract ansilove compatible text
)

// Artifact is the handler for the of the file record.
func (dir Dirs) Artifact(logr *zap.SugaredLogger, c echo.Context, readonly bool) error { //nolint:funlen
	const name = "artifact"
	if logr == nil {
		return InternalErr(logr, c, name, ErrZap)
	}
	var res *models.File
	var err error
	if sess.Editor(c) {
		res, err = model.EditObf(dir.URI)
	} else {
		res, err = model.FindObf(dir.URI)
	}
	if err != nil {
		if errors.Is(err, model.ErrID) {
			return ArtifactErr(logr, c, dir.URI)
		}
		return DatabaseErr(logr, c, "f/"+dir.URI, err)
	}
	fname := res.Filename.String
	uuid := res.UUID.String
	abs := filepath.Join(dir.Download, uuid)
	data := empty(c)
	// artifact editor
	if !readonly {
		data["readOnly"] = false
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
		data["recKind"] = artifactMagic(abs)
		data["recStatMod"] = artifactStat(abs)[0]
		data["recStatSize"] = artifactStat(abs)[1]
		data["recAssets"] = dir.artifactAssets(uuid)
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
	data["description"] = artifactDesc(res)
	data["h1"] = artifactIssue(res)
	data["lead"] = artifactLead(res)
	data["comment"] = res.Comment.String
	// file metadata
	data["filename"] = fname
	data["filesize"] = artifactByteCount(res.Filesize)
	data["filebyte"] = res.Filesize
	data["lastmodified"] = artifactLM(res)
	data["lastmodifiedAgo"] = artifactModAgo(res)
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
	data["listLinks"] = artifactLinks(res)
	data["listReleases"] = res.ListRelations.String
	data["listWebsites"] = res.ListLinks.String
	data["demozoo"] = artifactID(res.WebIDDemozoo.Int64)
	data["pouet"] = artifactID(res.WebIDPouet.Int64)
	data["sixteenColors"] = res.WebID16colors.String
	data["youtube"] = res.WebIDYoutube.String
	data["github"] = res.WebIDGithub.String
	// file archive content
	data["jsdos"] = artifactJSDos(res)
	nameBin := model.JsDosBinary(res)
	data["jsdosBinary"] = nameBin
	data["jsdosZip"] = filepath.Ext(strings.ToLower(fname)) == ".zip"
	ctt := artifactCtt(res)
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
	d, err := dir.artifactReadme(res)
	if err != nil {
		return InternalErr(logr, c, name, err)
	}
	maps.Copy(data, d)
	err = c.Render(http.StatusOK, name, data)
	if err != nil {
		return InternalErr(logr, c, name, err)
	}
	return nil
}

// AnsiLovePost handles the post submission for the Preview from text in archive.
func (dir Dirs) AnsiLovePost(logr *zap.SugaredLogger, c echo.Context) error {
	const name = "editor ansilove"
	if logr == nil {
		return InternalErr(logr, c, name, ErrZap)
	}
	return dir.extractor(logr, c, ansis)
}

// PreviewDel handles the post submission for the Delete complementary images button.
func (dir Dirs) PreviewDel(logr *zap.SugaredLogger, c echo.Context) error {
	const name = "editor preview remove"
	if logr == nil {
		return InternalErr(logr, c, name, ErrZap)
	}

	var f Form
	if err := c.Bind(&f); err != nil {
		return badRequest(c, err)
	}
	r, err := model.EditFind(f.ID)
	if err != nil {
		return badRequest(c, err)
	}
	if err = command.RemoveImgs(r.UUID.String, dir.Preview, dir.Thumbnail); err != nil {
		return badRequest(c, err)
	}
	return c.JSON(http.StatusOK, r)
}

// PreviewPost handles the post submission for the Preview from image in archive.
func (dir Dirs) PreviewPost(logr *zap.SugaredLogger, c echo.Context) error {
	const name = "editor preview"
	if logr == nil {
		return InternalErr(logr, c, name, ErrZap)
	}
	return dir.extractor(logr, c, imgs)
}

// artifactReadme returns the readme data for the file record.
func (dir Dirs) artifactReadme(res *models.File) (map[string]interface{}, error) { //nolint:funlen
	data := map[string]interface{}{}
	if res.RetrotxtNoReadme.Int16 != 0 {
		return data, nil
	}
	platform := strings.TrimSpace(res.Platform.String)
	switch platform {
	case "markup", "pdf":
		return data, nil
	}
	if render.NoScreenshot(res, dir.Preview) {
		data["noScreenshot"] = true
	}
	// the bbs era, remote images protcol is not supported
	// example: /f/b02392f
	const ripScrip = ".rip"
	if filepath.Ext(strings.ToLower(res.Filename.String)) == ripScrip {
		return data, nil
	}

	b, err := render.Read(res, dir.Download)
	if errors.Is(err, render.ErrDownload) {
		data["noDownload"] = true
		return data, nil
	}
	if errors.Is(err, render.ErrFilename) {
		return data, nil
	}
	if err != nil {
		return data, err
	}
	if b == nil || render.IsUTF16(b) {
		return data, nil
	}

	// check if the file is a zip archive.
	// if unknown "application/octet-stream" is returned,
	// but this can be a false positives with other legacy text files.
	contentType := http.DetectContentType(b)
	switch contentType {
	case "archive/zip", "application/zip":
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
	// and let the client choose which one to display
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

// extractor is a helper function for the PreviewPost and AnsiLovePost handlers.
func (dir Dirs) extractor(logr *zap.SugaredLogger, c echo.Context, p extract) error {
	if logr == nil {
		return InternalErr(logr, c, "extractor", ErrZap)
	}

	var f Form
	if err := c.Bind(&f); err != nil {
		return badRequest(c, err)
	}
	r, err := model.EditFind(f.ID)
	if err != nil {
		return badRequest(c, err)
	}

	list := strings.Split(r.FileZipContent.String, "\n")
	target := ""
	for _, x := range list {
		s := strings.TrimSpace(x)
		if s == "" {
			continue
		}
		if strings.EqualFold(s, f.Target) {
			target = s
		}
	}
	if target == "" {
		return badRequest(c, ErrTarget)
	}
	src := filepath.Join(dir.Download, r.UUID.String)
	cmd := command.Dirs{Download: dir.Download, Preview: dir.Preview, Thumbnail: dir.Thumbnail}
	ext := filepath.Ext(strings.ToLower(r.Filename.String))
	switch p {
	case imgs:
		err = cmd.ExtractImage(src, ext, r.UUID.String, target)
	case ansis:
		err = cmd.ExtractAnsiLove(src, ext, r.UUID.String, target)
	default:
		return InternalErr(logr, c, "extractor", fmt.Errorf("%w: %d", ErrExtract, p))
	}
	if err != nil {
		return badRequest(c, err)
	}
	return c.JSON(http.StatusOK, r)
}

// artifactByteCount returns the file size for the file record.
func artifactByteCount(i int64) string {
	if i == 0 {
		return "(n/a)"
	}
	return humanize.Bytes(uint64(i))
}

// artifactDesc returns the description for the file record.
func artifactDesc(res *models.File) string {
	s := res.Filename.String
	if res.RecordTitle.String != "" {
		s = artifactIssue(res)
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

// artifactIssue returns the title of the file,
// unless the file is a magazine issue, in which case it returns the issue number.
func artifactIssue(res *models.File) string {
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

// artifactLead returns the lead for the file record which is the filename and releasers.
func artifactLead(res *models.File) string {
	fname := res.Filename.String
	span := fmt.Sprintf("<span class=\"font-monospace fs-6 fw-light\">%s</span> ", fname)
	rels := string(LinkRels(res.GroupBrandBy, res.GroupBrandFor))
	return fmt.Sprintf("%s<br>%s", rels, span)
}

// artifactLM returns the last modified date for the file record.
func artifactLM(res *models.File) string {
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

// artifactMagic returns the MIME type for the file record.
func artifactMagic(name string) string {
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
	filetype.AddMatcher(magic.ANSIType(), magic.ANSIMatcher)
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

// artifactModAgo returns the last modified date in a human readable format.
func artifactModAgo(res *models.File) string {
	if !res.FileLastModified.Valid {
		return "No recorded timestamp"
	}
	year, _ := strconv.Atoi(res.FileLastModified.Time.Format("2006"))
	const epoch = 1980
	if year <= epoch {
		// 1980 is the default date for MS-DOS files without a timestamp
		return ""
	}
	return Updated(res.FileLastModified.Time, "Modified")
}

// artifactCtt returns the file archive content for the file record.
func artifactCtt(res *models.File) []string {
	conts := strings.Split(res.FileZipContent.String, "\n")
	conts = slices.DeleteFunc(conts, func(s string) bool {
		return strings.TrimSpace(s) == "" // delete empty lines
	})
	conts = slices.Compact(conts)
	return conts
}

// artifactLinks returns the list of links for the file record.
func artifactLinks(res *models.File) template.HTML {
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

// artifactJSDos returns true if the file record is a known, MS-DOS executable.
func artifactJSDos(res *models.File) bool {
	if strings.TrimSpace(strings.ToLower(res.Platform.String)) != "dos" {
		return false
	}
	// check supported filename extensions
	ext := filepath.Ext(strings.ToLower(res.Filename.String))
	switch ext {
	case ".zip": // js-dos only supports zip archives
		return true
	case ".exe", ".com":
		return true
	case ".bat", ".cmd":
		// do not support the emulation of batch scripts
		return false
	default:
		return false
	}
}

// artifactID returns the record ID as a string.
func artifactID(id int64) string {
	if id == 0 {
		return ""
	}
	return strconv.FormatInt(id, 10)
}

// artifactStat returns the file last modified date and formatted file size.
func artifactStat(name string) [2]string {
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

// artifactAssets returns a list of downloads and image assets belonging to the file record.
// any errors are appended to the list.
func (dir Dirs) artifactAssets(uuid string) map[string]string {
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
				matches[s] = humanize.Comma(st.Size()) + " bytes"
			}
		}
	}
	for _, file := range images {
		if strings.HasPrefix(file.Name(), uuid) {
			s := strings.ToUpper(filepath.Ext(file.Name()))
			if s == ".WEBP" {
				s = ".WebP"
			}
			matches[s+" preview "] = artifactImgInfo(filepath.Join(dir.Preview, file.Name()))
		}
	}
	for _, file := range thumbs {
		if strings.HasPrefix(file.Name(), uuid) {
			s := strings.ToUpper(filepath.Ext(file.Name()))
			if s == ".WEBP" {
				s = ".WebP"
			}
			matches[s+" thumb"] = artifactImgInfo(filepath.Join(dir.Thumbnail, file.Name()))
		}
	}

	return matches
}

// artifactImgInfo returns the image file size and dimensions.
func artifactImgInfo(name string) string {
	switch filepath.Ext(strings.ToLower(name)) {
	case ".jpg", ".jpeg", ".gif", ".png", ".webp":
	default:
		st, err := os.Stat(name)
		if err != nil {
			return err.Error()
		}
		return humanize.Comma(st.Size()) + " bytes"
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

// readmeFinds returns a list of readme text files found in the file archive.
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

// priority returns a list of readme text file extensions in priority order.
func priority() []string {
	return []string{".nfo", ".txt", ".unp", ".doc"}
}

// candidate returns a list of other, common text file extensions in priority order.
func candidate() []string {
	return []string{".diz", ".asc", ".1st", ".dox", ".me", ".cap", ".ans", ".pcb"}
}
