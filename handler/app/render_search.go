package app

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strings"

	"github.com/Defacto2/releaser"
	"github.com/Defacto2/releaser/initialism"
	"github.com/Defacto2/server/internal/helper"
	"github.com/Defacto2/server/internal/postgres"
	"github.com/Defacto2/server/model"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// SearchFile is the handler for the Search for files page.
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

// PostFile is the handler for the Search for files form post page.
func PostFile(z *zap.SugaredLogger, c echo.Context) error {
	const name = "files"
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

	s := strings.Join(terms, " ")
	data := emptyFiles()
	data["title"] = "Filename results"
	data["h1"] = "Filename search"
	data["lead"] = fmt.Sprintf("Results for %q", s)
	data["logo"] = s + " results"
	data["description"] = "Filename search results for " + s + "."
	data[records] = fs

	d := noFiles()
	if len(fs) > 0 {
		d, err = postFileStats(ctx, db, terms)
		if err != nil {
			return InternalErr(z, c, name, err)
		}
	}
	data["stats"] = d
	err = c.Render(http.StatusOK, "files", data)
	if err != nil {
		return InternalErr(z, c, name, err)
	}
	return nil
}

func noFiles() map[string]string {
	return map[string]string{
		"files": "no files found",
		"years": "",
	}
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
		"files": string(ByteFileS("file", m.SumCount, m.SumBytes)),
		"years": helper.Years(m.MinYear, m.MaxYear),
	}
	return d, nil
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
		name := releaser.Link(slug)
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
