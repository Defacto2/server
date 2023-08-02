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

const errConn = "Sorry, at the moment the server cannot connect to the database"

// Stats are the database statistics.
type Stats struct { //nolint:gochecknoglobals
	All       model.All
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

func Statistics() Stats {
	return Stats{}
}

// File is the handler for the file categories page.
func File(s *zap.SugaredLogger, c echo.Context, stats bool) error {
	data := initData()

	ctx := context.Background()
	db, err := postgres.ConnectDB()
	if err != nil {
		s.Warnf("%s: %s", errConn, err)
		return echo.NewHTTPError(http.StatusServiceUnavailable, errConn)
	}
	defer db.Close()
	counter := Stats{}
	if err := counter.Get(ctx, db); err != nil {
		s.Warnf("%s: %s", errConn, err)
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
	data := initData()
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

	// err := ctx.Render(http.StatusOK, "file", data)
	// if err != nil {
	// 	s.Errorf("%s: %s", ErrTmpl, err)
	// 	return echo.NewHTTPError(http.StatusInternalServerError, ErrTmpl)
	// }
	switch id {
	case "", "newest", "oldest", "new-uploads",
		"intro", "intro-windows", "intro-msdos", "installer", "demoscene",
		"nfo", "proof",
		"ansi", "ansi-brand", "ansi-bbs", "ansi-ftp", "ansi-nfo",
		"bbs", "bbstro", "bbs-image", "bbs-text",
		"ftp",
		"magazine",
		"ansi-pack", "text-pack", "nfo-pack", "image-pack", "windows-pack", "msdos-pack",
		"database",
		"text", "text-amiga", "text-apple2", "text-atari-st", "pdf", "html",
		"windows", "msdos", "macos", "linux", "script", "java",
		"news-article", "standards", "announcement", "job-advert", "trial-crackme",
		"hack", "tool", "nfo-tool", "takedown", "drama", "advert", "restrict", "how-to",
		"ansi-tool", "image", "music", "video":

		const (
			limit = 99
			page  = 1
		)
		var all model.All
		data["records"], err = all.List(ctx, db, page, limit)
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
	default:
		// TODO: redirect to File categories with custom alert 404 message?
		// replace this message: The page you are looking for might have been removed, had its name changed, or is temporarily unavailable.
		// with something about the file categories page.
		return Status(nil, c, http.StatusNotFound, c.Param("uri"))
	}
}
