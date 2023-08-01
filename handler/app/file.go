package app

import (
	"context"
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
var Stats struct { //nolint:gochecknoglobals
	All       model.All
	Demo      model.Demo
	Intro     model.Intro
	IntroD    model.IntroDOS
	IntroW    model.IntroWindows
	Installer model.Installer
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
	if err := Stats.All.Stat(ctx, db); err != nil {
		s.Warnf("%s: %s", errConn, err)
	}
	if err := Stats.Intro.Stat(ctx, db); err != nil {
		s.Warnf("%s: %s", errConn, err)
	}
	if err := Stats.IntroD.Stat(ctx, db); err != nil {
		s.Warnf("%s: %s", errConn, err)
	}
	if err := Stats.IntroW.Stat(ctx, db); err != nil {
		s.Warnf("%s: %s", errConn, err)
	}
	if err := Stats.Installer.Stat(ctx, db); err != nil {
		s.Warnf("%s: %s", errConn, err)
	}
	if err := Stats.Demo.Stat(ctx, db); err != nil {
		s.Warnf("%s: %s", errConn, err)
	}

	const title = "File categories"
	data["title"] = title
	data["description"] = "Table of contents for the files."
	data["logo"] = title
	data["h1"] = title
	data["stats"] = stats
	data["counter"] = Stats

	if stats {
		data["h1sub"] = "with statistics"
		data["logo"] = title + " + stats"
		data["lead"] = "This page shows the file categories with selected statistics, such as the number of files in the category or platform." +
			fmt.Sprintf(" The total number of files in the database is %d.", Stats.All.Count) +
			fmt.Sprintf(" The total size of all files in the database is %s.", helpers.ByteCount(int64(Stats.All.Bytes)))
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
	if err := Stats.All.Stat(ctx, db); err != nil {
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
		return c.String(http.StatusOK, fmt.Sprintf("ToDo!, %q", id))
	default:
		return Status(nil, c, http.StatusNotFound, c.Param("uri"))
	}
}
