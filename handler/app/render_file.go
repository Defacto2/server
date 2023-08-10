package app

// Package file render_file.go contains the handler functions for the file and files routes.

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"

	"github.com/Defacto2/server/model"
	"github.com/Defacto2/server/pkg/helpers"
	"github.com/Defacto2/server/pkg/postgres"
	"github.com/Defacto2/server/pkg/postgres/models"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

const records = "records"

// File is the handler for the file categories page.
func File(z *zap.SugaredLogger, c echo.Context, stats bool) error {
	if z == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("%w: handler app file", ErrLogger))
	}
	data := empty()
	ctx := context.Background()
	db, err := postgres.ConnectDB()
	if err != nil {
		z.Warnf("%s: %s", errConn, err)
		return echo.NewHTTPError(http.StatusServiceUnavailable, errConn)
	}
	defer db.Close()
	counter := Stats{}
	if err := counter.Get(ctx, db); err != nil {
		z.Warnf("%w: %w", errConn, err)
		return echo.NewHTTPError(http.StatusServiceUnavailable, errConn)
	}

	const title = "File categories"
	data["title"] = title
	data["description"] = "Table of contents for the files."
	data["logo"] = title
	data["h1"] = title
	data["stats"] = stats
	data["counter"] = counter
	if stats {
		data["h1sub"] = "with statistics"
		data["logo"] = title + " + stats"
		data["lead"] = "This page shows the file categories with selected statistics, " +
			"such as the number of files in the category or platform." +
			fmt.Sprintf(" The total number of files in the database is %d.", counter.All.Count) +
			fmt.Sprintf(" The total size of all files in the database is %s.", helpers.ByteCount(int64(counter.All.Bytes)))
	}
	err = c.Render(http.StatusOK, "file", data)
	if err != nil {
		z.Errorf("%s: %s", ErrTmpl, err)
		return echo.NewHTTPError(http.StatusInternalServerError, ErrTmpl)
	}
	return nil
}

// Files is the handler for the files page.
func Files(z *zap.SugaredLogger, c echo.Context, id string) error {
	if z == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("%w: handler app files", ErrLogger))
	}
	if !IsURI(id) {
		// TODO: redirect to File categories with custom alert 404 message?
		// replace this message: The page you are looking for might have been removed, had its name changed, or is temporarily unavailable.
		// with something about the file categories page.
		return StatusErr(z, c, http.StatusNotFound, c.Param("uri"))
	}

	const (
		limit = 99
		page  = 1
	)
	data := empty()
	ctx := context.Background()
	db, err := postgres.ConnectDB()
	if err != nil {
		z.Warnf("%s: %s", errConn, err)
		return echo.NewHTTPError(http.StatusServiceUnavailable, errConn)
	}
	defer db.Close()
	counter := Stats{}
	if err := counter.All.Stat(ctx, db); err != nil {
		z.Warnf("%s: %s", errConn, err)
	}

	data["title"] = "Files placeholder"
	data["logo"] = "Files placeholder"
	data["description"] = "Table of contents for the files."
	data[records], err = Records(ctx, db, id, page, limit)
	if err != nil {
		z.Warnf("%s: %s", ErrTmpl, err)
		return echo.NewHTTPError(http.StatusInternalServerError, ErrTmpl)
	}
	err = c.Render(http.StatusOK, "files", data)
	if err != nil {
		z.Errorf("%s: %s", ErrTmpl, err)
		return echo.NewHTTPError(http.StatusInternalServerError, ErrTmpl)
	}
	return nil
}

// G is the handler for the files page.
// TODO: move this to _releaser.go
func G(z *zap.SugaredLogger, c echo.Context, id string) error {
	if z == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("%w: handler app files", ErrLogger))
	}
	// if !IsURI(id) {
	// 	// TODO: redirect to File categories with custom alert 404 message?
	// 	// replace this message: The page you are looking for might have been removed, had its name changed, or is temporarily unavailable.
	// 	// with something about the file categories page.
	// 	return StatusErr(s, c, http.StatusNotFound, c.Param("uri"))
	// }

	fmt.Fprintln(os.Stdout, "G", id)

	const (
		limit = 99
		page  = 1
	)
	data := empty()
	ctx := context.Background()
	db, err := postgres.ConnectDB()
	if err != nil {
		z.Warnf("%s: %s", errConn, err)
		return echo.NewHTTPError(http.StatusServiceUnavailable, errConn)
	}
	defer db.Close()
	counter := Stats{}
	if err := counter.All.Stat(ctx, db); err != nil {
		z.Warnf("%s: %s", errConn, err)
	}

	data["title"] = "Releaser files placeholder"
	data["logo"] = "Releaser files placeholder"
	data["description"] = "Table of contents for the files."
	rel := model.Releasers{}
	data[records], err = rel.List(ctx, db, id)

	x := data[records].(models.FileSlice)
	fmt.Fprintln(os.Stdout, "G len", len(x))

	if err != nil {
		z.Warnf("%s: %s", ErrTmpl, err)
		return echo.NewHTTPError(http.StatusInternalServerError, ErrTmpl)
	}
	err = c.Render(http.StatusOK, "files", data)
	if err != nil {
		z.Errorf("%s: %s", ErrTmpl, err)
		return echo.NewHTTPError(http.StatusInternalServerError, ErrTmpl)
	}
	return nil
}

// Records returns the records for the file category URI.
func Records(ctx context.Context, db *sql.DB, uri string, page, limit int) (models.FileSlice, error) {
	if db == nil {
		return nil, ErrDB
	}
	switch Match(uri) {
	// pulldown menu matches
	case newUploads:
		r := model.Files{}
		return r.List(ctx, db, page, limit)
	case newUpdates:
		r := model.Files{}
		return r.ListUpdates(ctx, db, page, limit)
	case oldest:
		r := model.Files{}
		return r.ListOldest(ctx, db, page, limit)
	case newest:
		r := model.Files{}
		return r.ListNewest(ctx, db, page, limit)
	// file categories matches
	case advert:
		r := model.Advert{}
		return r.List(ctx, db, page, limit)
	case announcement:
		r := model.Announcement{}
		return r.List(ctx, db, page, limit)
	case ansi:
		r := model.Ansi{}
		return r.List(ctx, db, page, limit)
	case ansiBrand:
		r := model.AnsiBrand{}
		return r.List(ctx, db, page, limit)
	case ansiBBS:
		r := model.AnsiBBS{}
		return r.List(ctx, db, page, limit)
	case ansiFTP:
		r := model.AnsiFTP{}
		return r.List(ctx, db, page, limit)
	case ansiNfo:
		r := model.AnsiNfo{}
		return r.List(ctx, db, page, limit)
	case ansiPack:
		r := model.AnsiPack{}
		return r.List(ctx, db, page, limit)
	case bbs:
		r := model.BBS{}
		return r.List(ctx, db, page, limit)
	case bbsImage:
		r := model.BBSImage{}
		return r.List(ctx, db, page, limit)
	case bbstro:
		r := model.BBStro{}
		return r.List(ctx, db, page, limit)
	case bbsText:
		r := model.BBSText{}
		return r.List(ctx, db, page, limit)
	case database:
		r := model.Database{}
		return r.List(ctx, db, page, limit)
	case demoscene:
		r := model.Demo{}
		return r.List(ctx, db, page, limit)
	case drama:
		r := model.Drama{}
		return r.List(ctx, db, page, limit)
	case ftp:
		r := model.FTP{}
		return r.List(ctx, db, page, limit)
	case hack:
		r := model.Hack{}
		return r.List(ctx, db, page, limit)
	case html:
		r := model.HTML{}
		return r.List(ctx, db, page, limit)
	case howTo:
		r := model.HowTo{}
		return r.List(ctx, db, page, limit)
	case image:
		r := model.Image{}
		return r.List(ctx, db, page, limit)
	case imagePack:
		r := model.ImagePack{}
		return r.List(ctx, db, page, limit)
	case installer:
		r := model.Installer{}
		return r.List(ctx, db, page, limit)
	case intro:
		r := model.Intro{}
		return r.List(ctx, db, page, limit)
	case linux:
		r := model.Linux{}
		return r.List(ctx, db, page, limit)
	case java:
		r := model.Java{}
		return r.List(ctx, db, page, limit)
	case jobAdvert:
		r := model.JobAdvert{}
		return r.List(ctx, db, page, limit)
	case macos:
		r := model.Mac{}
		return r.List(ctx, db, page, limit)
	case msdosPack:
		r := model.DosPack{}
		return r.List(ctx, db, page, limit)
	case music:
		r := model.Music{}
		return r.List(ctx, db, page, limit)
	case newsArticle:
		r := model.NewsArticle{}
		return r.List(ctx, db, page, limit)
	case nfo:
		r := model.Nfo{}
		return r.List(ctx, db, page, limit)
	case nfoTool:
		r := model.NfoTool{}
		return r.List(ctx, db, page, limit)
	case standards:
		r := model.Standard{}
		return r.List(ctx, db, page, limit)
	case script:
		r := model.Script{}
		return r.List(ctx, db, page, limit)
	case introMsdos:
		r := model.IntroDOS{}
		return r.List(ctx, db, page, limit)
	case introWindows:
		r := model.IntroWindows{}
		return r.List(ctx, db, page, limit)
	case magazine:
		r := model.Mag{}
		return r.List(ctx, db, page, limit)
	case msdos:
		r := model.DOS{}
		return r.List(ctx, db, page, limit)
	case pdf:
		r := model.PDF{}
		return r.List(ctx, db, page, limit)
	case proof:
		r := model.Proof{}
		return r.List(ctx, db, page, limit)
	case restrict:
		r := model.Restrict{}
		return r.List(ctx, db, page, limit)
	case takedown:
		r := model.Takedown{}
		return r.List(ctx, db, page, limit)
	case text:
		r := model.Text{}
		return r.List(ctx, db, page, limit)
	case textAmiga:
		r := model.TextAmiga{}
		return r.List(ctx, db, page, limit)
	case textApple2:
		r := model.TextAppleII{}
		return r.List(ctx, db, page, limit)
	case textAtariST:
		r := model.TextAtariST{}
		return r.List(ctx, db, page, limit)
	case textPack:
		r := model.TextPack{}
		return r.List(ctx, db, page, limit)
	case tool:
		r := model.Tool{}
		return r.List(ctx, db, page, limit)
	case trialCrackme:
		r := model.TrialCrackme{}
		return r.List(ctx, db, page, limit)
	case video:
		r := model.Video{}
		return r.List(ctx, db, page, limit)
	case windows:
		r := model.Windows{}
		return r.List(ctx, db, page, limit)
	case windowsPack:
		r := model.WindowsPack{}
		return r.List(ctx, db, page, limit)
	default:
		return nil, fmt.Errorf("unknown file category: %s", uri)
	}
}

// Stats are the database statistics for the file categories.
type Stats struct { //nolint:gochecknoglobals
	All       model.Files
	Ansi      model.Ansi
	AnsiBBS   model.AnsiBBS
	BBS       model.BBS
	BBSText   model.BBSText
	BBStro    model.BBStro
	Demo      model.Demo
	DOS       model.DOS
	Intro     model.Intro
	IntroD    model.IntroDOS
	IntroW    model.IntroWindows
	Installer model.Installer
	Java      model.Java
	Linux     model.Linux
	Mag       model.Mag
	Mac       model.Mac
	Nfo       model.Nfo
	NfoTool   model.NfoTool
	Proof     model.Proof
	Script    model.Script
	Text      model.Text
	Windows   model.Windows
}

// Get and store the database statistics for the file categories.
func (s *Stats) Get(ctx context.Context, db *sql.DB) error {
	if db == nil {
		return ErrDB
	}
	if err := s.All.Stat(ctx, db); err != nil {
		return err
	}
	if err := s.Ansi.Stat(ctx, db); err != nil {
		return err
	}
	if err := s.AnsiBBS.Stat(ctx, db); err != nil {
		return err
	}
	if err := s.BBS.Stat(ctx, db); err != nil {
		return err
	}
	if err := s.BBSText.Stat(ctx, db); err != nil {
		return err
	}
	if err := s.BBStro.Stat(ctx, db); err != nil {
		return err
	}
	if err := s.DOS.Stat(ctx, db); err != nil {
		return err
	}
	if err := s.Intro.Stat(ctx, db); err != nil {
		return err
	}
	if err := s.IntroD.Stat(ctx, db); err != nil {
		return err
	}
	if err := s.IntroW.Stat(ctx, db); err != nil {
		return err
	}
	if err := s.Installer.Stat(ctx, db); err != nil {
		return err
	}
	if err := s.Java.Stat(ctx, db); err != nil {
		return err
	}
	if err := s.Linux.Stat(ctx, db); err != nil {
		return err
	}
	if err := s.Demo.Stat(ctx, db); err != nil {
		return err
	}
	if err := s.Mac.Stat(ctx, db); err != nil {
		return err
	}
	if err := s.Mag.Stat(ctx, db); err != nil {
		return err
	}
	if err := s.Nfo.Stat(ctx, db); err != nil {
		return err
	}
	if err := s.NfoTool.Stat(ctx, db); err != nil {
		return err
	}
	if err := s.Proof.Stat(ctx, db); err != nil {
		return err
	}
	if err := s.Script.Stat(ctx, db); err != nil {
		return err
	}
	if err := s.Text.Stat(ctx, db); err != nil {
		return err
	}
	if err := s.Windows.Stat(ctx, db); err != nil {
		return err
	}
	return nil
}

// Statistics returns the empty database statistics for the file categories.
func Statistics() Stats {
	return Stats{}
}
