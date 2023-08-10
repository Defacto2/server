package app

// Package file render_file.go contains the handler functions for the file and files routes.

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/Defacto2/server/model"
	"github.com/Defacto2/server/pkg/helpers"
	"github.com/Defacto2/server/pkg/postgres"
	"github.com/Defacto2/server/pkg/postgres/models"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

const records = "records"

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

// Files is the handler for the files page.
func Files(z *zap.SugaredLogger, c echo.Context, uri string) error {
	if z == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("%w: handler app files", ErrLogger))
	}
	if !IsURI(uri) {
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
		h1sub = RecordsSub(uri)
	}
	data := empty()
	data["title"] = title
	data["description"] = "Table of contents for the files."
	data["logo"] = logo
	data["h1"] = title
	data["h1sub"] = h1sub
	data["lead"] = lead
	data[records] = []models.FileSlice{}

	const (
		limit = 99
		page  = 1
	)
	ctx := context.Background()
	db, err := postgres.ConnectDB()
	if err != nil {
		z.Warnf("%s: %s", errConn, err)
		return echo.NewHTTPError(http.StatusServiceUnavailable, errConn)
	}
	defer db.Close()
	counter := Stats{}
	if err := counter.All.Stat(ctx, db); err != nil {
		z.Warnf("%s: %s", errConn, err)
	}
	// fetch the records by category
	data[records], err = Records(ctx, db, uri, page, limit)
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

// G is the handler for the files page.
// TODO: move this to _releaser.go
func G(z *zap.SugaredLogger, c echo.Context, id string) error {
	if z == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("%w: handler app files", ErrLogger))
	}
	// if !IsURI(id) {
	// 	// TODO: redirect to File categories with custom alert 404 message?
	// 	// replace this message: The page you are looking for might have been removed, had its name changed, or is temporarily unavailable.
	// 	// with something about the file categories page.
	// 	return StatusErr(s, c, http.StatusNotFound, c.Param("uri"))
	// }

	fmt.Fprintln(os.Stdout, "G", id)

	const (
		limit = 99
		page  = 1
	)
	data := empty()
	ctx := context.Background()
	db, err := postgres.ConnectDB()
	if err != nil {
		z.Warnf("%s: %s", errConn, err)
		return echo.NewHTTPError(http.StatusServiceUnavailable, errConn)
	}
	defer db.Close()
	counter := Stats{}
	if err := counter.All.Stat(ctx, db); err != nil {
		z.Warnf("%s: %s", errConn, err)
	}

	data["title"] = "Releaser files placeholder"
	data["logo"] = "Releaser files placeholder"
	data["description"] = "Table of contents for the files."
	rel := model.Releasers{}
	data[records], err = rel.List(ctx, db, id)

	x := data[records].(models.FileSlice)
	fmt.Fprintln(os.Stdout, "G len", len(x))

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
