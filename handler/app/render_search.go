package app

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strings"

	"github.com/Defacto2/releaser/initialism"
	namer "github.com/Defacto2/releaser/name"
	"github.com/Defacto2/server/internal/helper"
	"github.com/Defacto2/server/internal/postgres"
	"github.com/Defacto2/server/model"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// FileSearch is the type of search to perform.
type FileSearch int

const (
	Filenames    FileSearch = iota // Filenames is the search for filenames.
	Descriptions                   // Descriptions is the search for file descriptions and titles.
)

// SearchDesc is the handler for the Search for file descriptions page.
func SearchDesc(z *zap.SugaredLogger, c echo.Context) error {
	const title, name = "Search titles and descriptions", "searchPost"
	if z == nil {
		return InternalErr(z, c, name, ErrZap)
	}
	data := empty()
	data["description"] = "Search form to scan through file descriptions."
	data["logo"] = title
	data["title"] = title
	data["info"] = "A search for file descriptions"
	err := c.Render(http.StatusOK, name, data)
	if err != nil {
		return InternalErr(z, c, name, err)
	}
	return nil
}

// PostDesc is the handler for the Search for file descriptions form post page.
func PostDesc(z *zap.SugaredLogger, c echo.Context, input string) error {
	const name = "files"
	ctx := context.Background()
	db, err := postgres.ConnectDB()
	if err != nil {
		return DatabaseErr(z, c, name, err)
	}
	defer db.Close()

	//input := c.FormValue("search-term-query")
	terms := helper.SearchTerm(input)
	rel := model.Files{}

	fs, _ := rel.SearchDescription(ctx, db, terms)
	if err != nil {
		return DatabaseErr(z, c, name, err)
	}
	d := Descriptions.postStats(ctx, db, terms)
	s := strings.Join(terms, ", ")
	data := emptyFiles()
	data["title"] = "Title and description results"
	data["h1"] = "Title and description search"
	data["lead"] = fmt.Sprintf("Results for %q", s)
	data["logo"] = s + " results"
	data["description"] = "Title and description search results for " + s + "."
	data["unknownYears"] = false
	data[records] = fs
	data["stats"] = d
	err = c.Render(http.StatusOK, "files", data)
	if err != nil {
		return InternalErr(z, c, name, err)
	}
	return nil
}

// SearchFile is the handler for the Search for files page.
func SearchFile(z *zap.SugaredLogger, c echo.Context) error {
	const title, name = "Search for filenames", "searchPost"
	if z == nil {
		return InternalErr(z, c, name, ErrZap)
	}
	data := empty()
	data["description"] = "Search form to discover files."
	data["logo"] = title
	data["title"] = title
	data["info"] = "A search for filenames or extensions"
	err := c.Render(http.StatusOK, name, data)
	if err != nil {
		return InternalErr(z, c, name, err)
	}
	return nil
}

// PostFilename is the handler for the Search for filenames form post page.
func PostFilename(z *zap.SugaredLogger, c echo.Context) error {
	return PostName(z, c, Filenames)
}

// PostName is the handler for the Search for filenames form post page.
func PostName(z *zap.SugaredLogger, c echo.Context, mode FileSearch) error {
	const name = "files"
	ctx := context.Background()
	db, err := postgres.ConnectDB()
	if err != nil {
		return DatabaseErr(z, c, name, err)
	}
	defer db.Close()

	input := c.FormValue("search-term-query")
	terms := helper.SearchTerm(input)
	rel := model.Files{}

	fs, _ := rel.SearchFilename(ctx, db, terms)
	if err != nil {
		return DatabaseErr(z, c, name, err)
	}
	d := mode.postStats(ctx, db, terms)
	s := strings.Join(terms, ", ")
	data := emptyFiles()
	data["title"] = "Filename results"
	data["h1"] = "Filename search"
	data["lead"] = fmt.Sprintf("Results for %q", s)
	data["logo"] = s + " results"
	data["description"] = "Filename search results for " + s + "."
	data["unknownYears"] = false
	data[records] = fs
	data["stats"] = d
	err = c.Render(http.StatusOK, "files", data)
	if err != nil {
		return InternalErr(z, c, name, err)
	}
	return nil
}

func (mode FileSearch) postStats(ctx context.Context, db *sql.DB, terms []string) map[string]string {
	if db == nil {
		return nil
	}
	none := func() map[string]string {
		return map[string]string{
			"files": "no files found",
			"years": "",
		}
	}
	// fetch the statistics of the category
	m := model.Summary{}
	switch mode {
	case Filenames:
		if err := m.SearchFilename(ctx, db, terms); err != nil {
			return none()
		}
	case Descriptions:
		if err := m.SearchDesc(ctx, db, terms); err != nil {
			return none()
		}
	}
	if m.SumCount.Int64 == 0 {
		return none()
	}
	// add the statistics to the data
	d := map[string]string{
		"files": string(ByteFileS("file", m.SumCount.Int64, m.SumBytes.Int64)),
		"years": helper.Years(m.MinYear.Int16, m.MaxYear.Int16),
	}
	return d
}

// PostReleaser is the handler for the releaser search form post page.
func PostReleaser(z *zap.SugaredLogger, c echo.Context) error {
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

// SearchReleaser is the handler for the Releaser Search page.
func SearchReleaser(z *zap.SugaredLogger, c echo.Context) error {
	const title, name = "Search for releasers", "searchList"
	if z == nil {
		return InternalErr(z, c, name, ErrZap)
	}
	data := empty()
	data["description"] = "Search form to discover releasers."
	data["logo"] = title
	data["title"] = title
	data["info"] = "A search for a group, initalism, magazine, board or site"
	ctx := context.Background()
	db, err := postgres.ConnectDB()
	if err != nil {
		return DatabaseErr(z, c, name, err)
	}
	defer db.Close()
	x := model.ReleaserStr{}
	if err := x.List(ctx, db); err != nil {
		return DatabaseErr(z, c, name, err)
	}
	s := make([]string, len(x))
	for i, v := range x {
		id := strings.TrimSpace(v.Name)
		slug := helper.Slug(id)
		// use namer.Humanized instead of the releaser.link func as it is far more performant
		name, _ := namer.Humanize(namer.Path(slug))
		ism := initialism.Initialism(initialism.Path(slug))
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
