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

// BBS is the handler for the BBS page ordered by the most files.
func BBS(z *zap.SugaredLogger, c echo.Context) error {
	return bbsHandler(z, c, true)
}

// BBSAZ is the handler for the BBS page ordered alphabetically.
func BBSAZ(z *zap.SugaredLogger, c echo.Context) error {
	return bbsHandler(z, c, false)
}

// bbsHandler is the handler for the BBS page.
func bbsHandler(z *zap.SugaredLogger, c echo.Context, prolific bool) error {
	const title, name = "BBS", "file"
	if z == nil {
		return InternalErr(z, c, name, ErrZap)
	}
	const lead = "Historical, telephone based \"elite\" Bulletin Board Systems for communication and file sharing."
	const key = "releasers"
	data := empty()
	data["title"] = title
	data["description"] = lead
	data["logo"] = "Bulletin Board Systems"
	data["h1"] = title
	data["lead"] = lead
	// releaser.html specific data items
	data["itemName"] = name
	data[key] = model.Releasers{}
	data["stats"] = map[string]string{}

	ctx := context.Background()
	db, err := postgres.ConnectDB()
	if err != nil {
		return InternalErr(z, c, name, err)
	}
	defer db.Close()
	r := model.Releasers{}
	if err := r.BBS(ctx, db, 0, 0, prolific); err != nil {
		return InternalErr(z, c, name, err)
	}
	m := model.Summary{}
	if err := m.BBS(ctx, db); err != nil {
		return InternalErr(z, c, name, err)
	}
	data[key] = r
	data["stats"] = map[string]string{
		"pubs":   fmt.Sprintf("%d boards", len(r)),
		"issues": string(FmtByteName(name, m.SumCount, m.SumBytes)),
		"years":  fmt.Sprintf("%d - %d", m.MinYear, m.MaxYear),
	}
	err = c.Render(http.StatusOK, name, data)
	if err != nil {
		return InternalErr(z, c, name, err)
	}
	return nil
}

// FTP is the handler for the FTP page.
func FTP(z *zap.SugaredLogger, c echo.Context) error {
	const title, name = "FTP", "file"
	if z == nil {
		return InternalErr(z, c, name, ErrZap)
	}
	data := empty()
	const lead = "Historical, internet based \"elite\" FTP sites for the upload and download of scene releases."
	const key = "releasers"
	data["title"] = title
	data["description"] = lead
	data["logo"] = "FTP sites"
	data["h1"] = title
	data["lead"] = lead
	// releaser.html specific data items
	data["itemName"] = name
	data[key] = model.Releasers{}
	data["stats"] = map[string]string{}
	ctx := context.Background()
	db, err := postgres.ConnectDB()
	if err != nil {
		return InternalErr(z, c, name, err)
	}
	defer db.Close()
	r := model.Releasers{}
	if err := r.FTP(ctx, db, 0, 0, model.NameAsc); err != nil {
		return InternalErr(z, c, name, err)
	}
	m := model.Summary{}
	if err := m.FTP(ctx, db); err != nil {
		return InternalErr(z, c, name, err)
	}
	data[key] = r
	data["stats"] = map[string]string{
		"pubs":   fmt.Sprintf("%d sites", len(r)),
		"issues": string(FmtByteName(name, m.SumCount, m.SumBytes)),
		"years":  fmt.Sprintf("%d - %d", m.MinYear, m.MaxYear),
	}
	err = c.Render(http.StatusOK, name, data)
	if err != nil {
		return InternalErr(z, c, name, err)
	}
	return nil
}

// Magazine is the handler for the Magazine page.
func Magazine(z *zap.SugaredLogger, c echo.Context) error {
	const title, name = "Magazine", "magazine"
	if z == nil {
		return InternalErr(z, c, name, ErrZap)
	}
	data := empty()
	const lead = "Newsletters, reports and publications written about activities within The Scene subculture."
	const issue = "issue"
	const key = "releasers"
	data["title"] = title
	data["description"] = lead
	data["logo"] = title
	data["h1"] = title
	data["lead"] = lead
	// releaser.html specific data items
	data["itemName"] = issue
	data[key] = model.Releasers{}
	data["stats"] = map[string]string{}

	ctx := context.Background()
	db, err := postgres.ConnectDB()
	if err != nil {
		return InternalErr(z, c, name, err)
	}
	defer db.Close()
	r := model.Releasers{}
	if err := r.Magazine(ctx, db, 0, 0, model.NameAsc); err != nil {
		return InternalErr(z, c, name, err)
	}
	m := model.Summary{}
	if err := m.Magazine(ctx, db); err != nil {
		return InternalErr(z, c, name, err)
	}
	data[key] = r
	data["stats"] = map[string]string{
		"pubs":   fmt.Sprintf("%d publications", len(r)),
		"issues": string(FmtByteName(issue, m.SumCount, m.SumBytes)),
		"years":  fmt.Sprintf("%d - %d", m.MinYear, m.MaxYear),
	}
	err = c.Render(http.StatusOK, name, data)
	if err != nil {
		return InternalErr(z, c, name, err)
	}
	return nil
}

// Releaser is the handler for the releaser page ordered by the most files.
func Releaser(z *zap.SugaredLogger, c echo.Context) error {
	return releaser(z, c, true)
}

// ReleaserAZ is the handler for the releaser page ordered alphabetically.
func ReleaserAZ(z *zap.SugaredLogger, c echo.Context) error {
	return releaser(z, c, false)
}

// releaser is the handler for the Releaser page.
func releaser(z *zap.SugaredLogger, c echo.Context, prolific bool) error {
	const title, name = "Releaser", "releaser"
	if z == nil {
		return InternalErr(z, c, name, ErrZap)
	}
	data := empty()
	const lead = "A releaser is a collective or group of sceners who are responsible for releasing or distributing product."
	const key = "releasers"
	data["title"] = title
	data["description"] = fmt.Sprint(title, " ", lead)
	data["logo"] = "Groups, organisations and publications"
	data["h1"] = title
	data["lead"] = lead
	// releaser.html specific data items
	data["itemName"] = "file"
	data[key] = model.Releasers{}
	data["stats"] = map[string]string{}

	ctx := context.Background()
	db, err := postgres.ConnectDB()
	if err != nil {
		return InternalErr(z, c, name, err)
	}
	defer db.Close()
	var r model.Releasers
	if err := r.All(ctx, db, 0, 0, prolific); err != nil {
		return InternalErr(z, c, name, err)
	}
	var m model.Summary
	if err := m.All(ctx, db); err != nil {
		return InternalErr(z, c, name, err)
	}
	data[key] = r
	data["stats"] = map[string]string{
		"pubs":   fmt.Sprintf("%d releasers and groups", len(r)),
		"issues": string(FmtByteName("file", m.SumCount, m.SumBytes)),
		"years":  FmtYears(m.MinYear, m.MaxYear),
	}
	err = c.Render(http.StatusOK, name, data)
	if err != nil {
		return InternalErr(z, c, name, err)
	}
	return nil
}
