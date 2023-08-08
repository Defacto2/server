package app

// Package file render.go contains the handler functions for the app pages.

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Defacto2/server/model"
	"github.com/Defacto2/server/pkg/postgres"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

const demo = "demo"

const errSQL = "Database connection problem or a SQL error" // fix

// empty is a map of default values for the app templates.
func empty() map[string]interface{} {
	return map[string]interface{}{
		"canonical":   "",           // A canonical URL is the URL of the best representative page from a group of duplicate pages.
		"carousel":    "",           // The ID of the carousel to display.
		"description": "",           // A short description of the page that get inserted into the description meta element.
		"h1":          "",           // The H1 heading of the page.
		"h1sub":       "",           // The H1 sub-heading of the page.
		"lead":        "",           // The enlarged, lead paragraph of the page.
		"logo":        "",           // Text to insert into the monospaced, ASCII art logo.
		"title":       "",           // The title of the page that get inserted into the title meta element.
		"counter":     Statistics(), // The database counts for files and categories.
	}
}

// TODO: reorder by menu order

// Artist is the handler for the Artist page.
func Artist(s *zap.SugaredLogger, c echo.Context) error {
	if s == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, ErrLogger)
	}
	data := empty()
	data["description"] = demo
	data["title"] = demo
	err := c.Render(http.StatusOK, "artist", data)
	if err != nil {
		s.Errorf("%s: %s", ErrTmpl, err)
		return echo.NewHTTPError(http.StatusInternalServerError, ErrTmpl)
	}
	return nil
}

// BBS is the handler for the BBS page.
func BBS(s *zap.SugaredLogger, c echo.Context) error {
	if s == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, ErrLogger)
	}
	data := empty()
	data["description"] = demo
	data["title"] = demo

	data["itemName"] = "issue"

	// TODO: groups data
	ctx := context.Background()
	db, err := postgres.ConnectDB()
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, errConn)
	}
	defer db.Close()
	// Groups are the distinct groups from the file table.
	var sceners model.Releasers //nolint:gochecknoglobals
	if err := sceners.BBS(ctx, db, 0, 0, model.NameAsc); err != nil {
		s.Errorf("%s: %s %d", errConn, err)
		const errSQL = "Database connection problem or a SQL error" // fix
		return echo.NewHTTPError(http.StatusNotFound, errSQL)
	}
	data["sceners"] = sceners // model.Grps.List

	err = c.Render(http.StatusOK, "bbs", data)
	if err != nil {
		s.Errorf("%s: %s", ErrTmpl, err)
		return echo.NewHTTPError(http.StatusInternalServerError, ErrTmpl)
	}
	return nil
}

func Coder(s *zap.SugaredLogger, c echo.Context) error {
	if s == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, ErrLogger)
	}
	data := empty()
	data["description"] = demo
	data["title"] = demo
	err := c.Render(http.StatusOK, "coder", data)
	if err != nil {
		s.Errorf("%s: %s", ErrTmpl, err)
		return echo.NewHTTPError(http.StatusInternalServerError, "coder")
	}
	return nil
}

// FTP is the handler for the FTP page.
func FTP(s *zap.SugaredLogger, c echo.Context) error {
	if s == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, ErrLogger)
	}
	data := empty()
	data["description"] = demo
	data["title"] = demo
	data["itemName"] = "file"

	// TODO: groups data
	ctx := context.Background()
	db, err := postgres.ConnectDB()
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, errConn)
	}
	defer db.Close()
	// Groups are the distinct groups from the file table.
	var sceners model.Releasers //nolint:gochecknoglobals
	if err := sceners.FTP(ctx, db, 0, 0, model.NameAsc); err != nil {
		s.Errorf("%s: %s %d", errConn, err)
		const errSQL = "Database connection problem or a SQL error" // fix
		return echo.NewHTTPError(http.StatusNotFound, errSQL)
	}
	data["sceners"] = sceners // model.Grps.List

	err = c.Render(http.StatusOK, "ftp", data)
	if err != nil {
		s.Errorf("%s: %s", ErrTmpl, err)
		return echo.NewHTTPError(http.StatusInternalServerError, ErrTmpl)
	}
	return nil
}

// Interview is the handler for the People Interviews page.
func Interview(s *zap.SugaredLogger, c echo.Context) error {
	if s == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, ErrLogger)
	}
	data := empty()
	data["description"] = demo
	data["title"] = demo
	err := c.Render(http.StatusOK, "interview", data)
	if err != nil {
		s.Errorf("%s: %s", ErrTmpl, err)
		return echo.NewHTTPError(http.StatusInternalServerError, ErrTmpl)
	}
	return nil
}

// Index is the handler for the Home page.
func Index(s *zap.SugaredLogger, c echo.Context) error {
	if s == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, ErrLogger)
	}
	const lead = "a website committed to preserving the historic PC cracking and warez scene subcultures." +
		" It covers digital objects including text files, demos, music, art, magazines, and other projects."
	const desc = "Defacto2 is " + lead
	data := empty()
	data["title"] = demo
	data["description"] = desc
	data["h1"] = "Welcome,"
	data["lead"] = "You're at " + lead
	data["milestones"] = Collection()
	err := c.Render(http.StatusOK, "index", data)
	if err != nil {
		s.Errorf("%s: %s", ErrTmpl, err)
		return echo.NewHTTPError(http.StatusInternalServerError, ErrTmpl)
	}
	return nil
}

// History is the handler for the History page.
func History(s *zap.SugaredLogger, c echo.Context) error {
	if s == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, ErrLogger)
	}
	const lead = "Defacto founded in late February or early March of 1996, as an electronic magazine that wrote about The Scene subculture."
	const h1 = "Our history"
	data := empty()
	data["carousel"] = "#carouselDf2Artpacks"
	data["description"] = lead
	data["logo"] = "The history of Defacto"
	data["h1"] = h1
	data["lead"] = lead
	data["title"] = h1
	err := c.Render(http.StatusOK, "history", data)
	if err != nil {
		s.Errorf("%s: %s", ErrTmpl, err)
		return echo.NewHTTPError(http.StatusInternalServerError, ErrTmpl)
	}
	return nil
}

// Magazine is the handler for the Magazine page.
func Magazine(s *zap.SugaredLogger, c echo.Context) error {
	if s == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, ErrLogger)
	}
	const h1 = "Magazines"
	const lead = "Reports and publications written about The Scene subculture."
	data := empty()
	data["description"] = lead
	data["lead"] = lead
	data["title"] = h1
	data["h1"] = h1
	data["itemName"] = "issue"

	ctx := context.Background()
	db, err := postgres.ConnectDB()
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, errConn)
	}
	defer db.Close()

	var r model.Releasers
	if err := r.Magazine(ctx, db, 0, 0, model.NameAsc); err != nil {
		s.Errorf("%s: %s %d", errConn, err)
		return echo.NewHTTPError(http.StatusNotFound, errSQL)
	}
	data["sceners"] = r

	mags := model.Mag{}
	if err := mags.Stat(ctx, db); err != nil {
		return err
	}
	fmt.Println(mags, "rel", len(r))

	err = c.Render(http.StatusOK, "magazine", data)
	if err != nil {
		s.Errorf("%s: %s", ErrTmpl, err)
		return echo.NewHTTPError(http.StatusInternalServerError, ErrTmpl)
	}
	return nil
}

// Musician is the handler for the Musician page.
func Musician(s *zap.SugaredLogger, c echo.Context) error {
	if s == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, ErrLogger)
	}
	data := empty()
	data["description"] = demo
	data["title"] = demo
	err := c.Render(http.StatusOK, "musician", data)
	if err != nil {
		s.Errorf("%s: %s", ErrTmpl, err)
		return echo.NewHTTPError(http.StatusInternalServerError, "musician")
	}
	return nil
}

// Scener is the handler for the Scener page.
func Scener(s *zap.SugaredLogger, c echo.Context) error {
	if s == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, ErrLogger)
	}
	data := empty()
	data["description"] = demo
	data["title"] = demo
	err := c.Render(http.StatusOK, "scener", data)
	if err != nil {
		s.Errorf("%s: %s", ErrTmpl, err)
		return echo.NewHTTPError(http.StatusInternalServerError, ErrTmpl)
	}
	return nil
}

// Releaser is the handler for the Releaser page.
func Releaser(s *zap.SugaredLogger, c echo.Context) error {
	if s == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, ErrLogger)
	}
	const h1 = "Releaser"
	const lead = "A releaser is a member of The Scene who is responsible for releasing new content."
	data := empty()
	data["description"] = fmt.Sprint(h1, " ", lead)
	data["logo"] = "The underground"
	data["h1"] = h1
	data["lead"] = lead
	data["title"] = h1
	data["itemName"] = "file"

	// TODO: groups data
	ctx := context.Background()
	db, err := postgres.ConnectDB()
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, errConn)
	}
	defer db.Close()
	// Groups are the distinct groups from the file table.
	var Groups model.Releasers //nolint:gochecknoglobals
	if err := Groups.All(ctx, db, 0, 0, model.NameAsc); err != nil {
		s.Errorf("%s: %s %d", errConn, err)
		const errSQL = "Database connection problem or a SQL error" // fix
		return echo.NewHTTPError(http.StatusNotFound, errSQL)
	}
	data["sceners"] = Groups // model.Grps.List

	err = c.Render(http.StatusOK, "releaser", data)
	if err != nil {
		s.Errorf("%s: %s", ErrTmpl, err)
		return echo.NewHTTPError(http.StatusInternalServerError, ErrTmpl)
	}
	return nil
}

// Thanks is the handler for the Thanks page.
func Thanks(s *zap.SugaredLogger, c echo.Context) error {
	if s == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, ErrLogger)
	}
	data := empty()
	data["description"] = "Defacto2 thankyous."
	data["h1"] = "Thank you!"
	data["lead"] = "Thanks to the hundreds of people who have contributed to Defacto2 over the decades with file submissions, hard drive donations, interviews, corrections, artwork and monetiary donations!"
	data["title"] = "Thanks!"
	err := c.Render(http.StatusOK, "thanks", data)
	if err != nil {
		s.Errorf("%s: %s", ErrTmpl, err)
		return echo.NewHTTPError(http.StatusInternalServerError, ErrTmpl)
	}
	return nil
}

// TheScene is the handler for the The Scene page.
func TheScene(s *zap.SugaredLogger, c echo.Context) error {
	if s == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, ErrLogger)
	}
	const h1 = "The Scene?"
	const lead = "Collectively referred to as The Scene, it is a subculture of different computer activities where participants actively share ideas and creations."
	data := empty()
	data["description"] = fmt.Sprint(h1, " ", lead)
	data["logo"] = "The underground"
	data["h1"] = h1
	data["lead"] = lead
	data["title"] = h1
	err := c.Render(http.StatusOK, "thescene", data)
	if err != nil {
		s.Errorf("%s: %s", ErrTmpl, err)
		return echo.NewHTTPError(http.StatusInternalServerError, ErrTmpl)
	}
	return nil
}

// Writer is the handler for the Writer page.
func Writer(s *zap.SugaredLogger, c echo.Context) error {
	if s == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, ErrLogger)
	}
	data := empty()
	data["description"] = demo
	data["title"] = "demo"
	err := c.Render(http.StatusOK, "writer", data)
	if err != nil {
		s.Errorf("%s: %s", ErrTmpl, err)
		return echo.NewHTTPError(http.StatusInternalServerError, "writer")
	}
	return nil
}
