package app

import (
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"github.com/Defacto2/server/handler/download"
	"github.com/Defacto2/server/model"
	"github.com/Defacto2/server/pkg/helper"
	"github.com/Defacto2/server/pkg/postgres/models"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"golang.org/x/exp/slices"
)

// About is the handler for the about page of the file record.
func About(z *zap.SugaredLogger, c echo.Context, uri string) error {
	const name = "about"
	if z == nil {
		return InternalErr(z, c, name, ErrZap)
	}
	res, err := download.OneRecord(z, c, uri)
	if err != nil {
		return err
	}
	title := res.RecordTitle.String
	fname := res.Filename.String
	data := empty()
	data["uuid"] = res.UUID.String
	data["download"] = helper.ObfuscateID(int64(res.ID))
	data["title"] = fname
	data["description"] = desc(res)
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
	data["createdat"] = res.Createdat.Time
	data["updatedat"] = res.Updatedat.Time

	err = c.Render(http.StatusOK, name, data)
	if err != nil {
		return InternalErr(z, c, name, err)
	}
	return nil
}

func desc(res *models.File) string {
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
	span := fmt.Sprintf("<span class=\"font-monospace fw-light\">%s</span> ", fname)
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
