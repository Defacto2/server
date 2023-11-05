package app

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/Defacto2/releaser"
	"github.com/Defacto2/server/internal/helper"
	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/Defacto2/server/internal/render"
	"github.com/Defacto2/server/model"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
	"golang.org/x/text/encoding/charmap"
)

// AboutConf contains required data for the about file page.
type AboutConf struct {
	DownloadDir   string // path to the file download directory
	ScreenshotDir string // path to the file screenshot directory
	URI           string // the URI of the file record
}

// About is the handler for the about page of the file record.
func (a AboutConf) About(z *zap.SugaredLogger, c echo.Context) error {
	const name = "about"
	if z == nil {
		return InternalErr(z, c, name, ErrZap)
	}
	res, err := model.OneRecord(z, c, a.URI)
	if err != nil {
		if errors.Is(err, model.ErrID) {
			return AboutErr(z, c, a.URI)
		}
		return DatabaseErr(z, c, "f/"+a.URI, err)
	}
	fname := res.Filename.String
	uuid := res.UUID.String
	data := empty()
	// about editor
	data["recID"] = res.ID
	data["recTitle"] = res.RecordTitle.String
	data["recOnline"] = res.Deletedat.Time.IsZero()
	data["recReleasers"] = string(RecordRels(res.GroupBrandBy, res.GroupBrandFor))
	data["recYear"] = res.DateIssuedYear.Int16
	data["recMonth"] = res.DateIssuedMonth.Int16
	data["recDay"] = res.DateIssuedDay.Int16
	data["recLastMod"] = res.FileLastModified.IsZero()
	data["recLastModValue"] = res.FileLastModified.Time.Format("2006-1-2") // value should not have no leading zeros
	// page metadata
	data["uuid"] = uuid
	data["download"] = helper.ObfuscateID(int64(res.ID))
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
	data["linkpreviewTip"] = LinkPreviewTip(res.ID, res.Filename.String, res.Platform.String)
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
	d, err := a.aboutReadme(res)
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

func (a AboutConf) aboutReadme(res *models.File) (map[string]interface{}, error) {
	data := map[string]interface{}{}
	if res.RetrotxtNoReadme.Int16 != 0 {
		return data, nil
	}
	platform := strings.TrimSpace(res.Platform.String)
	switch platform {
	case "markup", "pdf":
		return data, nil
	}
	if render.NoScreenshot(a.ScreenshotDir, res) {
		data["noScreenshot"] = true
	}
	// the bbs era, remote images protcol is not supported
	// example: /f/b02392f
	const ripScrip = ".rip"
	if filepath.Ext(strings.ToLower(res.Filename.String)) == ripScrip {
		return data, nil
	}

	b, err := render.Read(a.DownloadDir, res)
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
	if r1 != "" && r2 != "" {
		r = fmt.Sprintf("%s + %s", r1, r2)
	} else if r1 != "" {
		r = r1
	} else if r2 != "" {
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
	return template.HTML(rows)
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
	}
	return false
}

func aboutID(id int64) string {
	if id == 0 {
		return ""
	}
	return strconv.FormatInt(id, 10)
}
