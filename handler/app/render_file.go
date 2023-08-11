package app

// Package file render_file.go contains the handler functions for the file and files routes.

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"github.com/Defacto2/sceners/pkg/rename"
	"github.com/Defacto2/server/model"
	"github.com/Defacto2/server/pkg/helpers"
	"github.com/Defacto2/server/pkg/initialism"
	"github.com/Defacto2/server/pkg/postgres"
	"github.com/Defacto2/server/pkg/postgres/models"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

const (
	limit   = 198
	page    = 1
	records = "records"
)

// File is the handler for the file categories page.
func File(z *zap.SugaredLogger, c echo.Context, stats bool) error {
	if z == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("%w: handler app file", ErrLogger))
	}
	const title = "File categories"
	data := empty()
	data["title"] = title
	data["description"] = "A table of contents for the collection."
	data["logo"] = title
	data["h1"] = title
	data["lead"] = "This page shows the file categories and platforms in the collection."
	data["stats"] = stats
	data["counter"] = Stats{}

	if stats {
		c, err := counter()
		if err != nil {
			return echo.NewHTTPError(http.StatusServiceUnavailable, err)
		}
		data["counter"] = c
		data["logo"] = "File category statistics"
		data["lead"] = "This page shows the file categories with selected statistics, " +
			"such as the number of files in the category or platform." +
			fmt.Sprintf(" The total number of files in the database is %d.", c.All.Count) +
			fmt.Sprintf(" The total size of all files in the database is %s.", helpers.ByteCount(int64(c.All.Bytes)))
	}
	err := c.Render(http.StatusOK, "file", data)
	if err != nil {
		z.Errorf("%s: %s", ErrTmpl, err)
		return echo.NewHTTPError(http.StatusInternalServerError, ErrTmpl)
	}
	return nil
}

func counter() (Stats, error) {
	ctx := context.Background()
	db, err := postgres.ConnectDB()
	if err != nil {
		return Stats{}, fmt.Errorf("%w: %s", ErrConn, err)
	}
	defer db.Close()
	counter := Stats{}
	if err := counter.Get(ctx, db); err != nil {
		return Stats{}, fmt.Errorf("%w: %s", ErrConn, err)
	}
	return counter, nil
}

// Files is the handler for the list and preview of the files page.
func Files(z *zap.SugaredLogger, c echo.Context, uri string) error {
	if z == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("%w: handler app files", ErrLogger))
	}
	if !IsFiles(uri) {
		return FilesErr(z, c, uri)
	}

	title, logo, h1sub, lead := "Files", "", "", ""
	switch uri {
	case "new-uploads":
		logo = "new uploads"
		h1sub = "the new uploads"
		lead = "These are the files that have been recently uploaded to Defacto2."
	case "new-updates":
		logo = "new updates"
		h1sub = "the new updates"
		lead = "These are the file records that have been recently uploaded or modified on Defacto2."
	case "oldest":
		logo = "oldest releases"
		h1sub = "the oldest releases"
		lead = "These are the earliest, historical products from The Scene in the collection."
	case "newest":
		logo = "newest releases"
		h1sub = "the newest releases"
		lead = "These are the most recent products from The Scene in the collection."
	default:
		s := RecordsSub(uri)
		h1sub = s
		logo = s
	}
	data := empty()
	data["title"] = title
	data["description"] = "Table of contents for the files."
	data["logo"] = logo
	data["h1"] = title
	data["h1sub"] = h1sub
	data["lead"] = lead
	data[records] = []models.FileSlice{}

	ctx := context.Background()
	db, err := postgres.ConnectDB()
	if err != nil {
		z.Warnf("%s: %s", errConn, err)
		return echo.NewHTTPError(http.StatusServiceUnavailable, errConn)
	}
	defer db.Close()
	// fetch the records by category
	r, err := Records(ctx, db, uri, page, limit)
	if err != nil {
		z.Warnf("%s: %s", ErrTmpl, err)
		return echo.NewHTTPError(http.StatusInternalServerError, ErrTmpl)
	}
	data[records] = r
	d, err := stats(ctx, db, uri)
	if err != nil {
		z.Warnf("%s: %s", ErrTmpl, err)
		return echo.NewHTTPError(http.StatusInternalServerError, ErrTmpl)
	}
	data["stats"] = d

	err = c.Render(http.StatusOK, "files", data)
	if err != nil {
		z.Errorf("%s: %s", ErrTmpl, err)
		return echo.NewHTTPError(http.StatusInternalServerError, ErrTmpl)
	}
	return nil
}

func stats(ctx context.Context, db *sql.DB, uri string) (map[string]string, error) {
	if db == nil {
		return nil, fmt.Errorf("%w: %s", ErrConn, "nil database connection")
	}
	// fetch the statistics of the category
	m := model.Summary{}
	if err := m.All(ctx, db); err != nil {
		return nil, err
	}
	// add the statistics to the data
	d := map[string]string{
		"files": string(FmtByteName("file", m.SumCount, m.SumBytes)),
		"years": fmt.Sprintf("%d - %d", m.MinYear, m.MaxYear),
	}
	switch uri {
	case "new-updates", "newest":
		d["years"] = fmt.Sprintf("%d - %d", m.MaxYear, m.MinYear)
	}
	return d, nil
}

// G is the handler for the files page.
// TODO: move this to _releaser.go
func G(z *zap.SugaredLogger, c echo.Context, uri string) error {
	if z == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("%w: handler app files", ErrLogger))
	}
	ctx := context.Background()
	db, err := postgres.ConnectDB()
	if err != nil {
		z.Warnf("%s: %s", errConn, err)
		return echo.NewHTTPError(http.StatusServiceUnavailable, errConn)
	}
	defer db.Close()

	name := rename.DeObfuscateURL(uri)
	rel := model.Releasers{}
	fs, err := rel.List(ctx, db, name)
	if err != nil {
		z.Warnf("%s: %s", ErrTmpl, err)
		return echo.NewHTTPError(http.StatusInternalServerError, ErrTmpl)
	}
	if len(fs) == 0 {
		return GErr(z, c, uri) // releaser not found
	}

	data := empty()
	data["title"] = "Files for " + name
	data["h1"] = name
	data["lead"] = initialism.Join(uri)
	data["logo"] = name
	data["description"] = "The collection of files for " + name + "."
	data[records] = fs

	d, err := releaserSum(ctx, db, name)
	if err != nil {
		z.Warnf("%s: %s", ErrTmpl, err)
		return echo.NewHTTPError(http.StatusInternalServerError, ErrTmpl)
	}
	data["stats"] = d

	if err != nil {
		z.Warnf("%s: %s", ErrTmpl, err)
		return echo.NewHTTPError(http.StatusInternalServerError, ErrTmpl)
	}
	err = c.Render(http.StatusOK, "files", data)
	if err != nil {
		z.Errorf("%s: %s", ErrTmpl, err)
		return echo.NewHTTPError(http.StatusInternalServerError, ErrTmpl)
	}
	return nil
}

func releaserSum(ctx context.Context, db *sql.DB, uri string) (map[string]string, error) {
	if db == nil {
		return nil, fmt.Errorf("%w: %s", ErrConn, "nil database connection")
	}
	// fetch the statistics of the category
	m := model.Summary{}
	if err := m.Releaser(ctx, db, uri); err != nil {
		return nil, err
	}
	// add the statistics to the data
	d := map[string]string{
		"files": string(FmtByteName("file", m.SumCount, m.SumBytes)),
		"years": FmtYears(m.MinYear, m.MaxYear),
	}
	return d, nil
}
