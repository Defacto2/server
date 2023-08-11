package app

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Defacto2/server/model"
	"github.com/Defacto2/server/pkg/postgres"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// The file render_releaser.go contains the renderers that use the releaser.html template.

// ReleaserAZ is the handler for the releaser page ordered alphabetically.
func ReleaserAZ(z *zap.SugaredLogger, c echo.Context) error {
	return releaser(z, c, false)
}

// Releaser is the handler for the releaser page ordered by the most files.
func Releaser(z *zap.SugaredLogger, c echo.Context) error {
	return releaser(z, c, true)
}

// releaser is the handler for the Releaser page.
func releaser(z *zap.SugaredLogger, c echo.Context, prolific bool) error {
	if z == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, ErrLogger)
	}
	data := empty()
	const h1 = "Releaser"
	const lead = "A releaser is a collective or group of sceners who are responsible for releasing or distributing product."
	const name = "file"
	const key = "releasers"
	data["title"] = h1
	data["description"] = fmt.Sprint(h1, " ", lead)
	data["logo"] = "Groups, organisations and publications"
	data["h1"] = h1
	data["lead"] = lead
	// releaser.html specific data items
	data["itemName"] = "file"
	data[key] = model.Releasers{}
	data["stats"] = map[string]string{}

	ctx := context.Background()
	db, err := postgres.ConnectDB()
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, errConn)
	}
	defer db.Close()
	var r model.Releasers
	if err := r.All(ctx, db, 0, 0, prolific); err != nil {
		z.Errorf("%s: %s %d", errConn, err)
		return echo.NewHTTPError(http.StatusNotFound, errSQL)
	}
	var m model.Summary
	if err := m.All(ctx, db); err != nil {
		return err
	}
	data[key] = r
	data["stats"] = map[string]string{
		"pubs":   fmt.Sprintf("%d releasers and groups", len(r)),
		"issues": string(FmtByteName(name, m.SumCount, m.SumBytes)),
		"years":  FmtYears(m.MinYear, m.MaxYear),
	}

	err = c.Render(http.StatusOK, "releaser", data)
	if err != nil {
		z.Errorf("%s: %s", ErrTmpl, err)
		return echo.NewHTTPError(http.StatusInternalServerError, ErrTmpl)
	}
	return nil
}

// Magazine is the handler for the Magazine page.
func Magazine(z *zap.SugaredLogger, c echo.Context) error {
	if z == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, ErrLogger)
	}
	data := empty()
	const h1 = "Magazines"
	const lead = "Newsletters, reports and publications written about activities within The Scene subculture."
	const name = "issue"
	const key = "releasers"
	data["title"] = h1
	data["description"] = lead
	data["logo"] = h1
	data["h1"] = h1
	data["lead"] = lead
	// releaser.html specific data items
	data["itemName"] = name
	data[key] = model.Releasers{}
	data["stats"] = map[string]string{}

	ctx := context.Background()
	db, err := postgres.ConnectDB()
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, errConn)
	}
	defer db.Close()
	var r model.Releasers
	if err := r.Magazine(ctx, db, 0, 0, model.NameAsc); err != nil {
		z.Errorf("%s: %s %d", errConn, err)
		return echo.NewHTTPError(http.StatusNotFound, errSQL)
	}
	var m model.Summary
	if err := m.Magazine(ctx, db); err != nil {
		return err
	}
	data[key] = r
	data["stats"] = map[string]string{
		"pubs":   fmt.Sprintf("%d publications", len(r)),
		"issues": string(FmtByteName(name, m.SumCount, m.SumBytes)),
		"years":  fmt.Sprintf("%d - %d", m.MinYear, m.MaxYear),
	}

	err = c.Render(http.StatusOK, "magazine", data)
	if err != nil {
		z.Errorf("%s: %s", ErrTmpl, err)
		return echo.NewHTTPError(http.StatusInternalServerError, ErrTmpl)
	}
	return nil
}

func BBS(z *zap.SugaredLogger, c echo.Context) error {
	return bbsH(z, c, true)
}

func BBSAZ(z *zap.SugaredLogger, c echo.Context) error {
	return bbsH(z, c, false)
}

// bbsH is the handler for the BBS page.
func bbsH(z *zap.SugaredLogger, c echo.Context, prolific bool) error {
	if z == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, ErrLogger)
	}
	data := empty()
	const h1 = "BBS"
	const lead = "Historical, telephone based \"elite\" Bulletin Board Systems for communication and file sharing."
	const name = "file"
	const key = "releasers"
	data["title"] = h1
	data["description"] = lead
	data["logo"] = "Bulletin Board Systems"
	data["h1"] = h1
	data["lead"] = lead
	// releaser.html specific data items
	data["itemName"] = name
	data[key] = model.Releasers{}
	data["stats"] = map[string]string{}

	ctx := context.Background()
	db, err := postgres.ConnectDB()
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, errConn)
	}
	defer db.Close()
	var r model.Releasers
	if err := r.BBS(ctx, db, 0, 0, prolific); err != nil {
		z.Errorf("%s: %s %d", errConn, err)
		return echo.NewHTTPError(http.StatusNotFound, errSQL)
	}
	var m model.Summary
	if err := m.BBS(ctx, db); err != nil {
		return err
	}
	data[key] = r
	data["stats"] = map[string]string{
		"pubs":   fmt.Sprintf("%d boards", len(r)),
		"issues": string(FmtByteName(name, m.SumCount, m.SumBytes)),
		"years":  fmt.Sprintf("%d - %d", m.MinYear, m.MaxYear),
	}

	err = c.Render(http.StatusOK, "bbs", data)
	if err != nil {
		z.Errorf("%s: %s", ErrTmpl, err)
		return echo.NewHTTPError(http.StatusInternalServerError, ErrTmpl)
	}
	return nil
}

// FTP is the handler for the FTP page.
func FTP(z *zap.SugaredLogger, c echo.Context) error {
	if z == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, ErrLogger)
	}
	data := empty()
	const h1 = "FTP"
	const lead = "Historical, internet based \"elite\" FTP sites for the upload and download of scene releases."
	const name = "file"
	const key = "releasers"
	data["title"] = h1
	data["description"] = lead
	data["logo"] = "FTP sites"
	data["h1"] = h1
	data["lead"] = lead
	// releaser.html specific data items
	data["itemName"] = name
	data[key] = model.Releasers{}
	data["stats"] = map[string]string{}

	ctx := context.Background()
	db, err := postgres.ConnectDB()
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, errConn)
	}
	defer db.Close()
	var r model.Releasers
	if err := r.FTP(ctx, db, 0, 0, model.NameAsc); err != nil {
		z.Errorf("%s: %s %d", errConn, err)
		return echo.NewHTTPError(http.StatusNotFound, errSQL)
	}
	var m model.Summary
	if err := m.FTP(ctx, db); err != nil {
		return err
	}
	data[key] = r
	data["stats"] = map[string]string{
		"pubs":   fmt.Sprintf("%d sites", len(r)),
		"issues": string(FmtByteName(name, m.SumCount, m.SumBytes)),
		"years":  fmt.Sprintf("%d - %d", m.MinYear, m.MaxYear),
	}

	err = c.Render(http.StatusOK, "ftp", data)
	if err != nil {
		z.Errorf("%s: %s", ErrTmpl, err)
		return echo.NewHTTPError(http.StatusInternalServerError, ErrTmpl)
	}
	return nil
}
