package app

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/Defacto2/server/handler/download"
	"github.com/Defacto2/server/model"
	"github.com/Defacto2/server/pkg/config"
	"github.com/Defacto2/server/pkg/helper"
	"github.com/Defacto2/server/pkg/postgres"
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
	data := empty()

	// move to a func of download.go
	id := helper.DeobfuscateID(c.Param("id"))
	if id < 1 {
		return fmt.Errorf("%w: %d ~ %s", download.ErrID, id, c.Param("id"))
	}
	// get record id, filename, uuid
	ctx := context.Background()
	db, err := postgres.ConnectDB()
	if err != nil {
		return ErrDB
	}
	defer db.Close()
	res, err := model.One(ctx, db, id)
	if err != nil {
		return ErrDB
	}
	if res.ID != int64(id) {
		return fmt.Errorf("%w: %d ~ %s", download.ErrID, id, c.Param("id"))
	}
	// build the source filepath
	title := res.RecordTitle.String
	fname := res.Filename.String
	// uid := strings.TrimSpace(res.UUID.String)

	data["title"] = fname
	data["description"] = "About the filename x released by x + x."
	//data["logo"] = title
	data["h1"] = title
	fn := fmt.Sprintf("<span class=\"font-monospace fw-light\">%s</span> ", fname)
	data["lead"] = string(LinkRelrs(res.GroupBrandBy, res.GroupBrandFor)) + "<br>" + fn
	// data["lead"] = fmt.Sprintf("<code>%s</code> ", fname) +
	// 	fmt.Sprintf("%s ", model.Published(res)) +
	// 	"<br>" +
	// 	string(LinkRelrs(res.GroupBrandBy, res.GroupBrandFor)) +
	// 	"<br>" +
	// 	fmt.Sprintf("%s", res.Comment.String)
	data["filename"] = fname
	data["filesize"] = helper.ByteCount(res.Filesize.Int64)
	lm := res.FileLastModified.Time.Format("2006 Jan 2, 15:04")
	if lm == "0001 Jan 1, 00:00" {
		lm = "not set"
	}
	data["lastmodified"] = lm
	data["checksum"] = res.FileIntegrityWeak.String
	data["magic"] = res.FileMagicType.String
	data["releasers"] = string(LinkRelrs(res.GroupBrandBy, res.GroupBrandFor))
	data["published"] = model.PublishedFmt(res)
	data["section"] = res.Section.String
	data["platform"] = res.Platform.String
	data["writers"] = res.CreditText.String
	data["artists"] = res.CreditIllustration.String
	data["programmers"] = res.CreditProgram.String
	data["musicians"] = res.CreditAudio.String
	conts := strings.Split(res.FileZipContent.String, "\n")
	conts = slices.DeleteFunc(conts, func(s string) bool {
		return strings.TrimSpace(s) == "" // delete empty lines
	})
	conts = slices.Compact(conts)
	data["content"] = conts
	data["contentDesc"] = ""
	if len(conts) == 1 {
		data["contentDesc"] = "contains one file"
	}
	if len(conts) > 1 {
		data["contentDesc"] = fmt.Sprintf("contains %d files", len(conts))
	}
	data["download"] = helper.ObfuscateID(int64(id))
	data["uuid"] = res.UUID.String // /images/uuid/original/filename
	// tODO make a func that handles webp and png etc
	data["image"] = strings.Join([]string{config.StaticOriginal(), fmt.Sprintf("%s.png", res.UUID.String)}, "/")

	data["demozoo"] = res.WebIDDemozoo.Int64
	data["pouet"] = res.WebIDPouet.Int64
	data["youtube"] = res.WebIDYoutube.String
	data["createdat"] = res.Createdat.Time
	data["updatedat"] = res.Updatedat.Time

	// todo: use different templates based on the platform and category.
	err = c.Render(http.StatusOK, name, data)
	if err != nil {
		return InternalErr(z, c, name, err)
	}
	return nil
}
