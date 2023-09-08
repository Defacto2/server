package app

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/Defacto2/server/model"
	"github.com/Defacto2/server/pkg/exts"
	"github.com/Defacto2/server/pkg/helper"
	"github.com/Defacto2/server/pkg/postgres/models"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"
)

// AboutConf contains required data for the about file page.
type AboutConf struct {
	DownloadDir string
	URI         string
}

// About is the handler for the about page of the file record.
func (a AboutConf) About(z *zap.SugaredLogger, c echo.Context) error {
	const name = "about"
	if z == nil {
		return InternalErr(z, c, name, ErrZap)
	}
	res, err := model.OneRecord(z, c, a.URI)
	if err != nil {
		return err
	}
	title := res.RecordTitle.String
	fname := res.Filename.String
	uuid := res.UUID.String
	platform := strings.TrimSpace(res.Platform.String)
	data := empty()
	data["uuid"] = uuid
	data["download"] = helper.ObfuscateID(int64(res.ID))
	data["title"] = fname
	data["description"] = aboutDesc(res)
	data["h1"] = title
	data["lead"] = aboutLead(res)
	data["comment"] = res.Comment.String
	// file metadata
	data["filename"] = fname
	data["filesize"] = helper.ByteCount(res.Filesize.Int64)
	data["lastmodified"] = aboutLM(res)
	data["checksum"] = res.FileIntegrityStrong.String
	data["magic"] = res.FileMagicType.String
	data["releasers"] = string(LinkRelrs(res.GroupBrandBy, res.GroupBrandFor))
	data["published"] = model.PublishedFmt(res)
	data["section"] = res.Section.String
	data["platform"] = platform
	// attributions and credits
	data["writers"] = res.CreditText.String
	data["artists"] = res.CreditIllustration.String
	data["programmers"] = res.CreditProgram.String
	data["musicians"] = res.CreditAudio.String
	// links to other records and sites
	data["listLinks"] = aboutLinks(res)
	data["demozoo"] = res.WebIDDemozoo.Int64
	data["pouet"] = res.WebIDPouet.Int64
	data["youtube"] = res.WebIDYoutube.String
	data["github"] = res.WebIDGithub.String
	// file archive content
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
	switch {
	case res.Createdat.Valid && res.Updatedat.Valid:
		c := Updated(res.Createdat.Time, "")
		u := Updated(res.Updatedat.Time, "")
		if c != u {
			c = Updated(res.Createdat.Time, "Created")
			u = Updated(res.Updatedat.Time, "Updated")
			data["filentry"] = c + "<br>" + u
		} else {
			data["filentry"] = c
		}
	case res.Createdat.Valid:
		c := Updated(res.Createdat.Time, "Created")
		data["filentry"] = c
	case res.Updatedat.Valid:
		u := Updated(res.Updatedat.Time, "Updated")
		data["filentry"] = u
	}

	//txt := filepath.Join(a.DownloadDir, uuid+".txt")

	// switch platform {
	// case "textamiga", "text":
	// 	if !exts.IsArchive(fname) {
	// 		data["noScreenshot"] = true
	// 	}
	// }

	// fmt.Println(res.Platform.String)

	// if strings.TrimSpace(res.Platform.String) == "textamiga" {
	// 	fmt.Println("hello bonjour")
	// 	helper.ReadFile(filepath.Join(a.DownloadDir, uuid))
	// }

	//if helper.IsStat(txt) && res.RetrotxtNoReadme.Int16 == 0 {

	//data["readmeFont"] = "font-dos"

	// check if utf8 and then check if ISO8859?

	// switch {
	// // case e == nil:
	// // 	data["readme"] = string(b)
	// // 	data["readmeFont"] = "font-dos"
	// case e == charmap.ISO8859_1:
	// 	r := e.NewDecoder().Reader(bytes.NewReader(b))
	// 	out := strings.Builder{}
	// 	if _, err := io.Copy(&out, r); err != nil {
	// 		z.Info(err)
	// 	}
	// 	data["readmeLatin1"] = out.String()
	// 	data["readmeFont"] = "font-amiga"
	// case e == charmap.CodePage437:
	// 	r := e.NewDecoder().Reader(bytes.NewReader(b))
	// 	out := strings.Builder{}
	// 	if _, err := io.Copy(&out, r); err != nil {
	// 		z.Info(err)
	// 	}
	// 	data["readmeCP437"] = out.String()
	// 	data["readmeFont"] = "font-dos"
	// }
	//}
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
	if res.RetrotxtNoReadme.Int16 != 0 {
		return nil, nil
	}

	// todo: make into func
	// if platform amigatext only show topaz pre
	// use filedownload for all text platform files
	// - except known archives extensions
	// - also do a scan to confirm is not a binary file

	fname := res.Filename.String
	uuid := res.UUID.String
	platform := strings.TrimSpace(res.Platform.String)
	section := strings.TrimSpace(res.Section.String)

	switch platform {
	case "markup", "pdf":
		return nil, nil
	}

	//file := filepath.Join(a.DownloadDir, uuid)
	txt := filepath.Join(a.DownloadDir, uuid+".txt")
	readPath := filepath.Join(a.DownloadDir, uuid)
	data := map[string]interface{}{}

	isTextfile := !exts.IsArchive(fname)
	switch platform {
	case "textamiga", "text", "atarist":
		if isTextfile {
			data["noScreenshot"] = true
			break
		}
		readPath = txt
	}

	if !helper.IsStat(readPath) {
		return data, nil
	}

	b, err := os.ReadFile(readPath)
	if err != nil {
		return nil, err
	}

	var e encoding.Encoding
	switch platform {
	case "textamiga":
		e = charmap.ISO8859_1
	default:
		switch section {
		case "appleii", "atarist":
			e = charmap.ISO8859_1
		default:
			e = helper.DetermineEncoding(b)
		}
	}

	data["readmeName"] = res.RetrotxtReadme.String
	fmt.Println("DetermineEncoding", e)

	r := charmap.ISO8859_1.NewDecoder().Reader(bytes.NewReader(b))
	out := strings.Builder{}
	if _, err := io.Copy(&out, r); err != nil {
		return nil, err
	}
	data["readmeLatin1"] = out.String()
	//data["readmeFont"] = "font-amiga"

	r = charmap.CodePage437.NewDecoder().Reader(bytes.NewReader(b))
	out = strings.Builder{}
	if _, err := io.Copy(&out, r); err != nil {
		return nil, err
	}
	data["readmeCP437"] = out.String()

	switch e {
	case charmap.ISO8859_1:
		data["readmeLatin1Cls"] = ""
		data["readmeCP437Cls"] = "d-none"
		data["topazCheck"] = "checked"
	case charmap.CodePage437:
		data["readmeLatin1Cls"] = "d-none"
		data["readmeCP437Cls"] = ""
		data["vgaCheck"] = "checked"
	}

	if err = helper.ReadFile(readPath); err != nil {
		return nil, err
	}

	return data, nil

}

func aboutDesc(res *models.File) string {
	s := res.Filename.String
	if res.RecordTitle.String != "" {
		s = res.RecordTitle.String
	}
	r1 := helper.Capitalize(res.GroupBrandBy.String)
	r2 := helper.Capitalize(res.GroupBrandFor.String)
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

func aboutLead(res *models.File) string {
	fname := res.Filename.String
	span := fmt.Sprintf("<span class=\"font-monospace fs-6 fw-light\">%s</span> ", fname)
	rels := string(LinkRelrs(res.GroupBrandBy, res.GroupBrandFor))
	return fmt.Sprintf("%s<br>%s", rels, span)
}

func aboutLM(res *models.File) string {
	const none = "not set"
	if !res.FileLastModified.Valid {
		return none
	}
	if res.FileLastModified.Time.Format("2006") == "1980" {
		// 1980 is the default date for MS-DOS files without a timestamp
		return none
	}
	lm := res.FileLastModified.Time.Format("2006 Jan 2, 15:04")
	if lm == "0001 Jan 1, 00:00" {
		return none
	}
	return lm
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
	for _, link := range links {
		x := strings.Split(link, ";")
		if len(x) != 2 {
			continue
		}
		name, href := x[0], x[1]
		rows += fmt.Sprintf("<tr><th scope=\"row\"><small>Link</small></th>"+
			"<td><small><a class=\"text-truncate\" href=\"%s\">%s</a></small></td></tr>", href, name)
	}
	return template.HTML(rows)
}
