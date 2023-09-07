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
	"github.com/Defacto2/server/pkg/helper"
	"github.com/Defacto2/server/pkg/postgres/models"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"golang.org/x/exp/slices"
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
	data := empty()
	data["uuid"] = uuid
	data["download"] = helper.ObfuscateID(int64(res.ID))
	data["title"] = fname
	data["description"] = aboutDesc(res)
	data["h1"] = title
	data["lead"] = aboutLead(res)
	data["filename"] = fname
	data["comment"] = res.Comment.String
	data["filesize"] = helper.ByteCount(res.Filesize.Int64)
	data["lastmodified"] = aboutLM(res)
	data["checksum"] = res.FileIntegrityStrong.String
	data["magic"] = res.FileMagicType.String
	data["releasers"] = string(LinkRelrs(res.GroupBrandBy, res.GroupBrandFor))
	data["published"] = model.PublishedFmt(res)
	data["section"] = res.Section.String
	data["platform"] = res.Platform.String
	data["writers"] = res.CreditText.String
	data["artists"] = res.CreditIllustration.String
	data["programmers"] = res.CreditProgram.String
	data["musicians"] = res.CreditAudio.String
	ctt := aboutCtt(res)
	data["content"] = ctt
	data["contentDesc"] = ""
	if len(ctt) == 1 {
		data["contentDesc"] = "contains one file"
	}
	if len(ctt) > 1 {
		data["contentDesc"] = fmt.Sprintf("contains %d files", len(ctt))
	}
	data["listLinks"] = aboutLinks(res)
	data["demozoo"] = res.WebIDDemozoo.Int64
	data["pouet"] = res.WebIDPouet.Int64
	data["youtube"] = res.WebIDYoutube.String
	data["github"] = res.WebIDGithub.String
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
	txt := filepath.Join(a.DownloadDir, uuid+".txt")

	// todo: make into func
	// if platform amigatext only show topaz pre
	// use filedownload for all text platform files
	// - except known archives extensions
	// - also do a scan to confirm is not a binary file

	if helper.IsStat(txt) && res.RetrotxtNoReadme.Int16 == 0 {
		b, err := os.ReadFile(txt)
		if err != nil {
			z.Error(err)
		}
		e := helper.DetermineEncoding(b)
		data["readmeName"] = res.RetrotxtReadme.String
		fmt.Println("DetermineEncoding", e)

		r := charmap.ISO8859_1.NewDecoder().Reader(bytes.NewReader(b))
		out := strings.Builder{}
		if _, err := io.Copy(&out, r); err != nil {
			z.Info(err)
		}
		data["readmeLatin1"] = out.String()
		//data["readmeFont"] = "font-amiga"

		r = charmap.CodePage437.NewDecoder().Reader(bytes.NewReader(b))
		out = strings.Builder{}
		if _, err := io.Copy(&out, r); err != nil {
			z.Info(err)
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
	}
	err = c.Render(http.StatusOK, name, data)
	if err != nil {
		return InternalErr(z, c, name, err)
	}
	return nil
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
	lm := res.FileLastModified.Time.Format("2006 Jan 2, 15:04")
	if lm == "0001 Jan 1, 00:00" {
		lm = "not set"
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
		rows += fmt.Sprintf("<tr><th scope=\"row\"><small>%s</small></th>"+
			"<td><small><a href=\"%s\">%s</a></small></td></tr>", name, href, href)
	}
	return template.HTML(rows)
}
