package app

// Package file render.go contains the handler functions for the app pages.
// The BBS, FTP, Magazine and Releaser handlers can be found in render_releaser.go.

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strings"

	"github.com/Defacto2/server/model"
	"github.com/Defacto2/server/pkg/fmts"
	"github.com/Defacto2/server/pkg/helper"
	"github.com/Defacto2/server/pkg/initialism"
	"github.com/Defacto2/server/pkg/postgres"
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

		"counter": Statistics(), // The database counts for files and categories.
		"dberror": false,        // If true, the database is not available.
	}
}

// emptyFiles is a map of default values specific to the files templates.
func emptyFiles() map[string]interface{} {
	data := empty()
	data["demozoo"] = "0"
	data["sixteen"] = ""
	data["scener"] = ""
	return data
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
	const lead = "a website committed to preserving the historic PC cracking and warez scene subcultures." +
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
	const lead = "Defacto2 started in late February or early March of 1996, as an electronic magazine that wrote about The Scene subculture. " +
		"In the past, alternative iterations of the name have included DF, De Facto, DeFacto, Defacto II, Defacto 2, and defacto2.com."
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

// Thanks is the handler for the Thanks page.
func Thanks(z *zap.SugaredLogger, c echo.Context) error {
	const name = "thanks"
	if z == nil {
		return InternalErr(z, c, name, ErrZap)
	}
	data := empty()
	data["description"] = "Defacto2 thankyous."
	data["h1"] = "Thank you!"
	data["lead"] = "Thanks to the hundreds of people who have contributed to Defacto2 over the decades with file submissions, hard drive donations, interviews, corrections, artwork and monetiary donations!"
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
	const lead = "Collectively referred to as The Scene, it is a subculture of different computer activities where participants actively share ideas and creations."
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

// SearchPoster is the handler for the releaser search form post page.
func SearchPoster(z *zap.SugaredLogger, c echo.Context) error {
	const name = "searchList"
	if z == nil {
		return InternalErr(z, c, name, ErrZap)
	}
	input := c.FormValue("releaser-data-list")
	val := helper.TrimRoundBraket(input)
	slug := helper.Slug(val)
	if slug == "" {
		return SearchReleaser(z, c)
	}
	// note, the redirect to a GET only works with 301 and 404 status codes.
	return c.Redirect(http.StatusMovedPermanently, "/g/"+slug)
}

func SearchFile(z *zap.SugaredLogger, c echo.Context) error {
	const title, name = "Search for files", "searchPost"
	if z == nil {
		return InternalErr(z, c, name, ErrZap)
	}
	data := empty()
	data["description"] = "Search form to discover files."
	data["logo"] = title
	data["title"] = title
	data["info"] = "A search can be for a filename, description, year or ?"
	err := c.Render(http.StatusOK, name, data)
	if err != nil {
		return InternalErr(z, c, name, err)
	}
	return nil
}

func PostFile(z *zap.SugaredLogger, c echo.Context) error {
	const name = "files"
	const uri = "defacto"
	ctx := context.Background()
	db, err := postgres.ConnectDB()
	if err != nil {
		return InternalErr(z, c, name, err)
	}
	defer db.Close()

	input := c.FormValue("search-term-query")
	terms := helper.SearchTerm(input)
	rel := model.Files{}
	fs, err := rel.Search(ctx, db, terms)
	if err != nil {
		return InternalErr(z, c, name, err)
	}
	if len(fs) == 0 {
		return ReleaserErr(z, c, uri)
	}

	s := strings.Join(terms, " ")
	data := emptyFiles()
	data["title"] = "Filename results"
	data["h1"] = "Filename results"
	data["lead"] = s
	data["logo"] = s
	data["description"] = "Filename search results for " + s + "."
	data[records] = fs
	d, err := postFileStats(ctx, db, terms)
	if err != nil {
		return InternalErr(z, c, name, err)
	}
	data["stats"] = d
	err = c.Render(http.StatusOK, "files", data)
	if err != nil {
		return InternalErr(z, c, name, err)
	}
	return nil
}

func postFileStats(ctx context.Context, db *sql.DB, terms []string) (map[string]string, error) {
	if db == nil {
		return nil, ErrDB
	}
	// fetch the statistics of the category
	m := model.Summary{}
	if err := m.Search(ctx, db, terms); err != nil {
		return nil, err
	}
	// add the statistics to the data
	d := map[string]string{
		"files": string(FmtByteName("file", m.SumCount, m.SumBytes)),
		"years": FmtYears(m.MinYear, m.MaxYear),
	}
	return d, nil
}

// SearchReleaser is the handler for the Releaser Search page.
func SearchReleaser(z *zap.SugaredLogger, c echo.Context) error {
	const title, name = "Search for releasers", "search"
	if z == nil {
		return InternalErr(z, c, name, ErrZap)
	}
	data := empty()
	data["description"] = "Search form to discover releasers."
	data["logo"] = title
	data["title"] = title
	data["info"] = "A search can be for a group, magazine, board or site"
	ctx := context.Background()
	db, err := postgres.ConnectDB()
	if err != nil {
		return InternalErr(z, c, name, err)
	}
	defer db.Close()
	x := model.ReleaserStr{}
	if err := x.List(ctx, db); err != nil {
		return InternalErr(z, c, name, err)
	}
	s := make([]string, len(x))
	for i, v := range x {
		id := strings.TrimSpace(v.Name)
		slug := helper.Slug(id)
		name := fmts.Name(slug)
		ism := initialism.Initialism(slug)
		opt := name
		if len(ism) > 0 {
			opt = fmt.Sprintf("%s (%s)", name, strings.Join(ism, ", "))
		}
		s[i] = opt
	}
	data["releasers"] = s

	err = c.Render(http.StatusOK, name, data)
	if err != nil {
		return InternalErr(z, c, name, err)
	}
	return nil
}
