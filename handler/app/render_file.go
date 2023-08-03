package app

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"github.com/Defacto2/server/model"
	"github.com/Defacto2/server/pkg/helpers"
	"github.com/Defacto2/server/pkg/postgres"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

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

// File is the handler for the file categories page.
func File(s *zap.SugaredLogger, c echo.Context, stats bool) error {
	if s == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("%w: handler app file", ErrLogger))
	}
	data := empty()

	ctx := context.Background()
	db, err := postgres.ConnectDB()
	if err != nil {
		s.Warnf("%s: %s", errConn, err)
		return echo.NewHTTPError(http.StatusServiceUnavailable, errConn)
	}
	defer db.Close()

	counter := Stats{}
	if err := counter.Get(ctx, db); err != nil {
		s.Warnf("%w: %w", errConn, err)
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
		data["lead"] = "This page shows the file categories with selected statistics, such as the number of files in the category or platform." +
			fmt.Sprintf(" The total number of files in the database is %d.", counter.All.Count) +
			fmt.Sprintf(" The total size of all files in the database is %s.", helpers.ByteCount(int64(counter.All.Bytes)))
	}
	err = c.Render(http.StatusOK, "file", data)
	if err != nil {
		s.Errorf("%s: %s", ErrTmpl, err)
		return echo.NewHTTPError(http.StatusInternalServerError, ErrTmpl)
	}
	return nil
}

// Files is the handler for the files page.
func Files(s *zap.SugaredLogger, c echo.Context, id string) error {
	const (
		limit = 99
		page  = 1
	)
	data := empty()
	data["title"] = "Files placeholder"
	data["logo"] = "Files placeholder"
	data["description"] = "Table of contents for the files."

	ctx := context.Background()
	db, err := postgres.ConnectDB()
	if err != nil {
		s.Warnf("%s: %s", errConn, err)
		return echo.NewHTTPError(http.StatusServiceUnavailable, errConn)
	}
	defer db.Close()

	counter := Stats{}
	if err := counter.All.Stat(ctx, db); err != nil {
		s.Warnf("%s: %s", errConn, err)
	}
	if !IsURI(id) {
		// TODO: redirect to File categories with custom alert 404 message?
		// replace this message: The page you are looking for might have been removed, had its name changed, or is temporarily unavailable.
		// with something about the file categories page.
		return Status(s, c, http.StatusNotFound, c.Param("uri"))
	}

	switch Match(id) {
	case demoscene:
		r := model.Demo{}
		data["records"], err = r.List(ctx, db, page, limit)
	case installer:
		r := model.Installer{}
		data["records"], err = r.List(ctx, db, page, limit)
	case intro:
		r := model.Intro{}
		data["records"], err = r.List(ctx, db, page, limit)
	case linux:
		r := model.Linux{}
		data["records"], err = r.List(ctx, db, page, limit)
	case java:
		r := model.Java{}
		data["records"], err = r.List(ctx, db, page, limit)
	case macos:
		r := model.Mac{}
		data["records"], err = r.List(ctx, db, page, limit)
	case script:
		r := model.Script{}
		data["records"], err = r.List(ctx, db, page, limit)
	case introMsdos:
		r := model.IntroDOS{}
		data["records"], err = r.List(ctx, db, page, limit)
	case introWindows:
		r := model.IntroWindows{}
		data["records"], err = r.List(ctx, db, page, limit)
	case msdos:
		r := model.DOS{}
		data["records"], err = r.List(ctx, db, page, limit)
	case windows:
		r := model.Windows{}
		data["records"], err = r.List(ctx, db, page, limit)
	default:
		return echo.NewHTTPError(http.StatusInternalServerError, "file category not found")
	}
	if err != nil {
		s.Warnf("%s: %s", ErrTmpl, err)
		return echo.NewHTTPError(http.StatusInternalServerError, ErrTmpl)
	}

	err = c.Render(http.StatusOK, "files", data)
	if err != nil {
		s.Errorf("%s: %s", ErrTmpl, err)
		return echo.NewHTTPError(http.StatusInternalServerError, ErrTmpl)
	}
	return nil
}
