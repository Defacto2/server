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

	"github.com/Defacto2/server/model"
	"github.com/Defacto2/server/pkg/fmts"
	"github.com/Defacto2/server/pkg/initialism"
	"github.com/Defacto2/server/pkg/postgres"
	"github.com/Defacto2/server/pkg/postgres/models"
	"github.com/Defacto2/server/pkg/sixteen"
	"github.com/Defacto2/server/pkg/zoo"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// Files is the handler for the list and preview of the files page.
// The uri is the category or collection of files to display.
// The page is the page number of the results to display.
func Files(z *zap.SugaredLogger, c echo.Context, uri, page string) error {
	if z == nil {
		return echo.NewHTTPError(http.StatusInternalServerError,
			fmt.Errorf("%w: handler app files", ErrLogger))
	}
	// check the uri is valid
	if !IsFiles(uri) {
		return FilesErr(z, c, uri)
	}
	// check the page is valid
	p := 1
	if page == "" {
		return files(z, c, uri, 1)
	}
	p, _ = strconv.Atoi(page)

	// uri := c.Request().URL.Path
	// page := 1
	// if p := c.QueryParam("page"); p != "" {
	// 	page = fmts.ParseInt(p)
	// }
	return files(z, c, uri, p)
}

func files(z *zap.SugaredLogger, c echo.Context, uri string, page int) error {
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
	r, err := Records(ctx, db, uri, int(page), limit)
	if err != nil {
		z.Warnf("%s: %s", ErrTmpl, err)
		return echo.NewHTTPError(http.StatusInternalServerError, ErrTmpl)
	}
	data[records] = r
	d, sum, err := stats(ctx, db, uri)
	if err != nil {
		z.Warnf("%s: %s", ErrTmpl, err)
		return echo.NewHTTPError(http.StatusInternalServerError, ErrTmpl)
	}
	data["stats"] = d

	//fmt.Println(float64(sum/limit), " // ", page, ">>", sum, limit)
	lastPage := math.Ceil(float64(sum) / float64(limit))
	fmt.Printf("%d / %d = %f\n", sum, limit, float64(sum)/float64(limit))
	if page > int(lastPage) {
		return echo.NewHTTPError(http.StatusNotFound)
	}

	data["Pagination"] = model.Pagination{
		TwoAfter: page + 2,
		NextPage: page + 1,
		CurrPage: page,
		PrevPage: page - 1,
		TwoBelow: page - 2,
		SumPages: int(lastPage),
		BaseURL:  "/files/" + uri,
	}

	err = c.Render(http.StatusOK, "files", data)
	if err != nil {
		z.Errorf("%s: %s", ErrTmpl, err)
		return echo.NewHTTPError(http.StatusInternalServerError, ErrTmpl)
	}
	return nil
}

func stats(ctx context.Context, db *sql.DB, uri string) (map[string]string, int, error) {
	if db == nil {
		return nil, 0, fmt.Errorf("%w: %s", ErrConn, "nil database connection")
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
		"files": string(FmtByteName("file", m.SumCount, m.SumBytes)),
		"years": fmt.Sprintf("%d - %d", m.MinYear, m.MaxYear),
	}
	switch uri {
	case "new-updates", "newest":
		d["years"] = fmt.Sprintf("%d - %d", m.MaxYear, m.MinYear)
	}
	return d, m.SumCount, nil
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
	data["demozoo"] = strconv.Itoa(int(zoo.Find(uri)))
	data["sixteen"] = sixteen.Find(uri)
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
