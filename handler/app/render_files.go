package app

// Package file render_files.go contains the renderers that use the files.html template.

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	"github.com/Defacto2/server/model"
	"github.com/Defacto2/server/pkg/fmts"
	"github.com/Defacto2/server/pkg/initialism"
	"github.com/Defacto2/server/pkg/postgres"
	"github.com/Defacto2/server/pkg/postgres/models"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// Files is the handler for the list and preview of the files page.
func Files(z *zap.SugaredLogger, c echo.Context, uri string) error {
	if z == nil {
		return echo.NewHTTPError(http.StatusInternalServerError,
			fmt.Errorf("%w: handler app files", ErrLogger))
	}
	if !IsFiles(uri) {
		return FilesErr(z, c, uri)
	}

	title, logo, h1sub, lead := "Files", "", "", ""
	switch uri {
	case newUploads.String():
		logo = "new uploads"
		h1sub = "the new uploads"
		lead = "These are the files that have been recently uploaded to Defacto2."
	case newUpdates.String():
		logo = "new updates"
		h1sub = "the new updates"
		lead = "These are the file records that have been recently uploaded or modified on Defacto2."
	case oldest.String():
		logo = "oldest releases"
		h1sub = "the oldest releases"
		lead = "These are the earliest, historical products from The Scene in the collection."
	case newest.String():
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
	if !IsFiles(uri) {
		return nil, nil
	}
	// fetch the statistics of the uri
	m := model.Summary{}
	err := m.URI(ctx, db, uri)
	if err != nil && !errors.Is(err, model.ErrURI) {
		return nil, err
	}
	if errors.Is(err, model.ErrURI) {
		if err := m.All(ctx, db); err != nil {
			return nil, err
		}
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

// Releasers is the handler for the list and preview of files credited to a releaser.
func Releasers(z *zap.SugaredLogger, c echo.Context, uri string) error {
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

	name := fmts.Name(uri)
	rel := model.Releasers{}
	fs, err := rel.List(ctx, db, uri)
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

	switch uri {
	case "independent":
		data["lead"] = initialism.Join(uri) +
			", independent releases are files with no group or releaser affiliation"
	}

	d, err := releaserSum(ctx, db, uri)
	if err != nil {
		z.Warnf("releaserSum %s: %s", ErrTmpl, err)
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
