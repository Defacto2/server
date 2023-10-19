package app

// Package file render.go contains the handler functions for the app pages.
// The BBS, FTP, Magazine and Releaser handlers can be found in render_releaser.go.

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/Defacto2/server/handler/download"
	"github.com/Defacto2/server/internal/pouet"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

const demo = "demo"

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

		"counter":      Statistics(),        // Empty database counts for files and categories.
		"df2FileCount": Caching.RecordCount, // The number of records of files in the database.
		"dberror":      false,               // If true, the database is not available.
	}
}

// emptyFiles is a map of default values specific to the files templates.
func emptyFiles() map[string]interface{} {
	data := empty()
	data["demozoo"] = "0"
	data["sixteen"] = ""
	data["scener"] = ""
	data["unknownYears"] = true
	return data
}

// Pouet is the handler for the Pouet production votes JSON page.
func Pouet(z *zap.SugaredLogger, c echo.Context, id string) error {
	const title, name = "Pouet", "pouet"
	if z == nil {
		return InternalErr(z, c, name, ErrZap)
	}
	data := pouet.Pouet{}
	val, err := strconv.Atoi(id)
	if err != nil {
		return c.String(http.StatusNotFound, err.Error())
	}
	err = data.Votes(val)
	if err != nil {
		return c.String(http.StatusNotFound, err.Error())
	}
	err = c.JSON(http.StatusOK, data)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return nil
}

// Checksum is the handler for the Checksum file record page.
func Checksum(z *zap.SugaredLogger, c echo.Context, id string) error {
	return download.Checksum(z, c, id)
}

// Download is the handler for the Download file record page.
func Download(z *zap.SugaredLogger, c echo.Context, path string) error {
	d := download.Download{
		Inline: false,
		Path:   path,
	}
	err := d.HTTPSend(z, c)
	if err != nil {
		return DownloadErr(z, c, "d", err)
	}
	return nil
}

// Inline is the handler for the Download file record page.
func Inline(z *zap.SugaredLogger, c echo.Context, path string) error {
	d := download.Download{
		Inline: true,
		Path:   path,
	}
	err := d.HTTPSend(z, c)
	if err != nil {
		return DownloadErr(z, c, "v", err)
	}
	return nil
}

// Interview is the handler for the People Interviews page.
func Interview(z *zap.SugaredLogger, c echo.Context) error {
	const title, name = "Interviews with sceners", "interview"
	if z == nil {
		return InternalErr(z, c, name, ErrZap)
	}
	data := empty()
	data["title"] = title
	data["description"] = "Discussions with scene members."
	data["logo"] = title
	data["h1"] = title
	data["lead"] = "An incomplete list of discussions and unedited interviews with sceners, crackers and demo makers."
	data["interviews"] = Interviewees()
	err := c.Render(http.StatusOK, name, data)
	if err != nil {
		return InternalErr(z, c, name, err)
	}
	return nil
}

// Index is the handler for the Home page.
func Index(z *zap.SugaredLogger, c echo.Context) error {
	const name = "index"
	if z == nil {
		return InternalErr(z, c, name, ErrZap)
	}
	const lead = "a website maintaining the historic PC cracking and warez scene subcultures." +
		" It covers digital objects including text files, demos, music, art, magazines, and other projects."
	const desc = "Defacto2 is " + lead
	data := empty()
	data["title"] = demo
	data["description"] = desc
	data["h1"] = "Welcome,"
	data["lead"] = "You're at " + lead
	data["milestones"] = Collection()
	err := c.Render(http.StatusOK, name, data)
	if err != nil {
		return InternalErr(z, c, name, err)
	}
	return nil
}

// History is the handler for the History page.
func History(z *zap.SugaredLogger, c echo.Context) error {
	const name = "history"
	if z == nil {
		return InternalErr(z, c, name, ErrZap)
	}
	const lead = "Defacto2 started in late February or early March of 1996," +
		" as an electronic magazine that wrote about The Scene subculture." +
		" In the past, alternative iterations of the name have included" +
		" De Facto, DF, DeFacto, Defacto II, Defacto 2, and the defacto2.com domain."
	const h1 = "The history of the brand"
	data := empty()
	data["carousel"] = "#carouselDf2Artpacks"
	data["description"] = lead
	data["logo"] = "The history of Defacto"
	data["h1"] = h1
	data["lead"] = lead
	data["title"] = h1
	err := c.Render(http.StatusOK, name, data)
	if err != nil {
		return InternalErr(z, c, name, err)
	}
	return nil
}

// Interview is the handler for the People Interviews page.
func Reader(z *zap.SugaredLogger, c echo.Context, id string) error {
	const title, name = "Textfile reader", "reader"
	if z == nil {
		return InternalErr(z, c, name, ErrZap)
	}
	data := empty()
	data["title"] = title
	data["description"] = "Discussions with scene members."
	data["logo"] = title
	data["h1"] = title
	data["lead"] = "An incomplete list of discussions and unedited interviews with sceners, crackers and demo makers."
	data["interviews"] = Interviewees()
	err := c.Render(http.StatusOK, name, data)
	if err != nil {
		return InternalErr(z, c, name, err)
	}
	return nil
}

// Thanks is the handler for the Thanks page.
func Thanks(z *zap.SugaredLogger, c echo.Context) error {
	const name = "thanks"
	if z == nil {
		return InternalErr(z, c, name, ErrZap)
	}
	data := empty()
	data["description"] = "Defacto2 thankyous."
	data["h1"] = "Thank you!"
	data["lead"] = "Thanks to the hundreds of people who have contributed" +
		" to Defacto2 over the decades with file submissions, hard drive donations," +
		" interviews, corrections, artwork and monetiary donations!"
	data["title"] = "Thanks!"
	err := c.Render(http.StatusOK, name, data)
	if err != nil {
		return InternalErr(z, c, name, err)
	}
	return nil
}

// TheScene is the handler for the The Scene page.
func TheScene(z *zap.SugaredLogger, c echo.Context) error {
	const name = "thescene"
	if z == nil {
		return InternalErr(z, c, name, ErrZap)
	}
	const h1 = "The Scene?"
	const lead = "Collectively referred to as The Scene," +
		" it is a subculture of different computer activities where participants actively share ideas and creations."
	data := empty()
	data["description"] = fmt.Sprint(h1, " ", lead)
	data["logo"] = "The underground"
	data["h1"] = h1
	data["lead"] = lead
	data["title"] = h1
	err := c.Render(http.StatusOK, name, data)
	if err != nil {
		return InternalErr(z, c, name, err)
	}
	return nil
}
