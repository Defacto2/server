package app

// Package file render.go contains the handler functions for the app pages.
// The BBS, FTP, Magazine and Releaser handlers can be found in render_releaser.go.

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

const demo = "demo"

const errSQL = "Database connection problem or a SQL error" // fix

// empty is a map of default values for the app templates.
func empty() map[string]interface{} {
	// the keys are listed in order of appearance in the templates.
	// * marked keys are required.
	// ! marked keys are suggested.
	return map[string]interface{}{
		"title":       "", // * The title of the page that get inserted into the title meta element.
		"canonical":   "", //   A canonical URL is the URL of the best representative page from a group of duplicate pages.
		"description": "", // * A short description of the page that get inserted into the description meta element.

		"logo":     "", // ! Text to insert into the monospaced, ASCII art logo.
		"h1":       "", // ! The H1 heading of the page.
		"h1sub":    "", //   The H1 sub-heading of the page.
		"lead":     "", // ! The enlarged, lead paragraph of the page.
		"carousel": "", //   The ID of the carousel to display.

		"counter": Statistics(), // The database counts for files and categories.
	}
}

func emptyFiles() map[string]interface{} {
	data := empty()
	data["demozoo"] = "0"
	data["sixteen"] = ""
	data["scener"] = ""
	return data
}

// TODO: reorder by menu order

// Interview is the handler for the People Interviews page.
func Interview(z *zap.SugaredLogger, c echo.Context) error {
	if z == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, ErrLogger)
	}
	data := empty()
	data["description"] = demo
	data["title"] = demo
	err := c.Render(http.StatusOK, "interview", data)
	if err != nil {
		z.Errorf("%s: %s", ErrTmpl, err)
		return echo.NewHTTPError(http.StatusInternalServerError, ErrTmpl)
	}
	return nil
}

// Index is the handler for the Home page.
func Index(z *zap.SugaredLogger, c echo.Context) error {
	if z == nil {
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
		z.Errorf("%s: %s", ErrTmpl, err)
		return echo.NewHTTPError(http.StatusInternalServerError, ErrTmpl)
	}
	return nil
}

// History is the handler for the History page.
func History(z *zap.SugaredLogger, c echo.Context) error {
	if z == nil {
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
		z.Errorf("%s: %s", ErrTmpl, err)
		return echo.NewHTTPError(http.StatusInternalServerError, ErrTmpl)
	}
	return nil
}

// Thanks is the handler for the Thanks page.
func Thanks(z *zap.SugaredLogger, c echo.Context) error {
	if z == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, ErrLogger)
	}
	data := empty()
	data["description"] = "Defacto2 thankyous."
	data["h1"] = "Thank you!"
	data["lead"] = "Thanks to the hundreds of people who have contributed to Defacto2 over the decades with file submissions, hard drive donations, interviews, corrections, artwork and monetiary donations!"
	data["title"] = "Thanks!"
	err := c.Render(http.StatusOK, "thanks", data)
	if err != nil {
		z.Errorf("%s: %s", ErrTmpl, err)
		return echo.NewHTTPError(http.StatusInternalServerError, ErrTmpl)
	}
	return nil
}

// TheScene is the handler for the The Scene page.
func TheScene(z *zap.SugaredLogger, c echo.Context) error {
	if z == nil {
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
		z.Errorf("%s: %s", ErrTmpl, err)
		return echo.NewHTTPError(http.StatusInternalServerError, ErrTmpl)
	}
	return nil
}
