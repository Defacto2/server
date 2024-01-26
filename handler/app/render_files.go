package app

// Package file render_files.go contains the renderers that use the files.html template.

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"net/http"
	"strconv"

	"github.com/Defacto2/releaser"
	"github.com/Defacto2/releaser/initialism"
	"github.com/Defacto2/server/internal/helper"
	"github.com/Defacto2/server/internal/postgres"
	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/Defacto2/server/internal/sixteen"
	"github.com/Defacto2/server/internal/web"
	"github.com/Defacto2/server/internal/zoo"
	"github.com/Defacto2/server/model"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

const (
	limit   = 198 // per-page record limit
	page    = 1   // default page number
	records = "records"
)

// Files is the handler for the list and preview of the files page.
// The uri is the category or collection of files to display.
// The page is the page number of the results to display.
func Files(z *zap.SugaredLogger, c echo.Context, uri, page string) error {
	if z == nil {
		return InternalErr(z, c, "files", ErrZap)
	}
	// check the uri is valid
	if !IsFiles(uri) {
		return FilesErr(z, c, uri)
	}
	// check the page is valid
	if page == "" {
		return files(z, c, uri, 1)
	}
	p, err := strconv.Atoi(page)
	if err != nil {
		return PageErr(z, c, uri, page)
	}
	return files(z, c, uri, p)
}

// fileInfo is a helper function for Files that returns the page title, h1 title and lead text.
func fileInfo(uri string) (string, string, string) {
	var logo, h1sub, lead string
	switch uri {
	case newUploads.String():
		logo = "new uploads"
		h1sub = "the new uploads"
		lead = "These are the recent file artifacts that have been submitted to Defacto2."
	case newUpdates.String():
		logo = "new changes"
		h1sub = "the new changes"
		lead = "These are the recent file artifacts that have been modified or submitted on Defacto2."
	case oldest.String():
		logo = "oldest releases"
		h1sub = "the oldest releases"
		lead = "These are the earliest, historical file artifacts in the collection."
	case newest.String():
		logo = "newest releases"
		h1sub = "the newest releases"
		lead = "These are the most recent file artifacts in the collection."
	default:
		s := RecordsSub(uri)
		h1sub = s
		logo = s
	}
	return logo, h1sub, lead
}

// files is a helper function for Files that returns the data map for the files page.
func files(z *zap.SugaredLogger, c echo.Context, uri string, page int) error {
	const title, name = "Files", "files"
	logo, h1sub, lead := fileInfo(uri)
	data := emptyFiles(c)
	data["title"] = title
	data["description"] = "Table of contents for the files."
	data["logo"] = logo
	data["h1"] = title
	data["h1Sub"] = h1sub
	data["lead"] = lead
	data[records] = []models.FileSlice{}
	data["unknownYears"] = true
	switch uri {
	case newUploads.String(), newUpdates.String():
		data["unknownYears"] = false
	}

	ctx := context.Background()
	db, err := postgres.ConnectDB()
	if err != nil {
		return InternalErr(z, c, name, err)
	}
	defer db.Close()
	// fetch the records by category
	r, err := Records(ctx, db, uri, page, limit)
	if err != nil {
		return DatabaseErr(z, c, name, err)
	}
	data[records] = r
	d, sum, err := stats(ctx, db, uri)
	if err != nil {
		return DatabaseErr(z, c, name, err)
	}
	data["stats"] = d
	lastPage := math.Ceil(float64(sum) / float64(limit))
	if page > int(lastPage) {
		i := strconv.Itoa(page)
		return PageErr(z, c, uri, i)
	}
	const pages = 2
	data["Pagination"] = model.Pagination{
		TwoAfter: page + pages,
		NextPage: page + 1,
		CurrPage: page,
		PrevPage: page - 1,
		TwoBelow: page - pages,
		SumPages: int(lastPage),
		BaseURL:  "/files/" + uri,
	}
	err = c.Render(http.StatusOK, name, data)
	if err != nil {
		return InternalErr(z, c, name, err)
	}
	return nil
}

// stats is a helper function for Files that returns the statistics for the files page.
func stats(ctx context.Context, db *sql.DB, uri string) (map[string]string, int, error) {
	if db == nil {
		return nil, 0, ErrDB
	}
	if !IsFiles(uri) {
		return nil, 0, nil
	}
	// fetch the statistics of the uri
	m := model.Summary{}
	err := m.URI(ctx, db, uri)
	if err != nil && !errors.Is(err, model.ErrURI) {
		return nil, 0, err
	}
	if errors.Is(err, model.ErrURI) {
		if err := m.All(ctx, db); err != nil {
			return nil, 0, err
		}
	}
	// add the statistics to the data
	d := map[string]string{
		"files": string(ByteFileS("file", m.SumCount.Int64, m.SumBytes.Int64)),
		"years": fmt.Sprintf("%d - %d", m.MinYear.Int16, m.MaxYear.Int16),
	}
	switch uri {
	case "new-updates", "newest":
		d["years"] = fmt.Sprintf("%d - %d", m.MaxYear.Int16, m.MinYear.Int16)
	}
	return d, int(m.SumCount.Int64), nil
}

// Sceners is the handler for the list and preview of files credited to a scener.
func Sceners(z *zap.SugaredLogger, c echo.Context, uri string) error {
	const name = "files"
	if z == nil {
		return InternalErr(z, c, name, ErrZap)
	}
	ctx := context.Background()
	db, err := postgres.ConnectDB()
	if err != nil {
		return InternalErr(z, c, name, err)
	}
	defer db.Close()

	s := releaser.Link(uri)
	var rel model.Scener
	fs, err := rel.List(ctx, db, uri)
	if err != nil {
		return InternalErr(z, c, name, err)
	}
	if len(fs) == 0 {
		return ScenerErr(z, c, uri)
	}
	data := emptyFiles(c)
	data["title"] = s + " attributions"
	data["h1"] = s
	data["lead"] = "Files attributed to " + s + "."
	data["logo"] = s
	data["description"] = "The collection of files attributed to " + s + "."
	data["scener"] = s
	data[records] = fs
	d, err := scenerSum(ctx, db, uri)
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

// scenerSum is a helper function for Sceners that returns the statistics for the files page.
func scenerSum(ctx context.Context, db *sql.DB, uri string) (map[string]string, error) {
	if db == nil {
		return nil, ErrDB
	}
	// fetch the statistics of the category
	m := model.Summary{}
	if err := m.Scener(ctx, db, uri); err != nil {
		return nil, err
	}
	// add the statistics to the data
	d := map[string]string{
		"files": string(ByteFileS("file", m.SumCount.Int64, m.SumBytes.Int64)),
		"years": helper.Years(m.MinYear.Int16, m.MaxYear.Int16),
	}
	return d, nil
}

// Releasers is the handler for the list and preview of files credited to a releaser.
func Releasers(z *zap.SugaredLogger, c echo.Context, uri string) error {
	const name = "files"
	if z == nil {
		return InternalErr(z, c, name, ErrZap)
	}
	ctx := context.Background()
	db, err := postgres.ConnectDB()
	if err != nil {
		return InternalErr(z, c, name, err)
	}
	defer db.Close()

	s := releaser.Link(uri)
	rel := model.Releasers{}
	fs, err := rel.List(ctx, db, uri)
	if err != nil {
		return InternalErr(z, c, name, err)
	}
	if len(fs) == 0 {
		return ReleaserErr(z, c, uri)
	}
	data := emptyFiles(c)
	data["title"] = "Files for " + s
	data["h1"] = s
	data["lead"] = initialism.Join(initialism.Path(uri))
	data["logo"] = s
	data["description"] = "The collection of files for " + s + "."
	data["demozoo"] = strconv.Itoa(int(zoo.Find(uri)))
	data["sixteen"] = sixteen.Find(uri)
	data["website"] = web.Find(uri)
	data[records] = fs
	switch uri {
	case "independent":
		data["lead"] = initialism.Join(initialism.Path(uri)) +
			", independent releases are files with no group or releaser affiliation." +
			`<br><small class="fw-lighter">In the scene's early years,` +
			` releasing documents or software cracks under a personal alias or a` +
			` real-name attribution was commonplace.</small>`
	default:
		// placeholder to handle other releaser types
	}
	d, err := releaserSum(ctx, db, uri)
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

// releaserSum is a helper function for Releasers that returns the statistics for the files page.
func releaserSum(ctx context.Context, db *sql.DB, uri string) (map[string]string, error) {
	if db == nil {
		return nil, ErrDB
	}
	// fetch the statistics of the category
	m := model.Summary{}
	if err := m.Releaser(ctx, db, uri); err != nil {
		return nil, err
	}
	// add the statistics to the data
	d := map[string]string{
		"files": string(ByteFileS("file", m.SumCount.Int64, m.SumBytes.Int64)),
		"years": helper.Years(m.MinYear.Int16, m.MaxYear.Int16),
	}
	return d, nil
}
