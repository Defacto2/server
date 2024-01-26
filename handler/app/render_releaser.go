package app

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Defacto2/server/internal/helper"
	"github.com/Defacto2/server/internal/postgres"
	"github.com/Defacto2/server/model"
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
	const title, name = "BBS", "bbs"
	if z == nil {
		return InternalErr(z, c, name, ErrZap)
	}
	const lead = "Bulletin Board Systems are historical, " +
		"networked personal computer servers connected using the landline telephone network and provide forums, " +
		"real-time chat, mail, and file sharing for The Scene \"elites.\""
	const key = "releasers"
	data := empty(c)
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
	if err := r.BBS(ctx, db, prolific); err != nil {
		return DatabaseErr(z, c, name, err)
	}
	m := model.Summary{}
	if err := m.BBS(ctx, db); err != nil {
		return DatabaseErr(z, c, name, err)
	}
	data[key] = r
	data["stats"] = map[string]string{
		"pubs":   fmt.Sprintf("%d boards", len(r)),
		"issues": string(ByteFileS("file artifact", m.SumCount.Int64, m.SumBytes.Int64)),
		"years":  fmt.Sprintf("%d - %d", m.MinYear.Int16, m.MaxYear.Int16),
	}
	err = c.Render(http.StatusOK, name, data)
	if err != nil {
		return InternalErr(z, c, name, err)
	}
	return nil
}

// FTP is the handler for the FTP page.
func FTP(z *zap.SugaredLogger, c echo.Context) error {
	const title, name = "FTP", "ftp"
	if z == nil {
		return InternalErr(z, c, name, ErrZap)
	}
	data := empty(c)
	const lead = "FTP sites are historical, internet-based file servers for uploading and downloading \"elite\" scene releases."
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
	if err := r.FTP(ctx, db); err != nil {
		return DatabaseErr(z, c, name, err)
	}
	m := model.Summary{}
	if err := m.FTP(ctx, db); err != nil {
		return DatabaseErr(z, c, name, err)
	}
	data[key] = r
	data["stats"] = map[string]string{
		"pubs":   fmt.Sprintf("%d sites", len(r)),
		"issues": string(ByteFileS("file artifact", m.SumCount.Int64, m.SumBytes.Int64)),
		"years":  fmt.Sprintf("%d - %d", m.MinYear.Int16, m.MaxYear.Int16),
	}
	err = c.Render(http.StatusOK, name, data)
	if err != nil {
		return InternalErr(z, c, name, err)
	}
	return nil
}

// Magazine is the handler for the Magazine page.
func Magazine(z *zap.SugaredLogger, c echo.Context) error {
	return mag(z, c, true)
}

// MagazineAZ is the handler for the Magazine page ordered chronologically.
func MagazineAZ(z *zap.SugaredLogger, c echo.Context) error {
	return mag(z, c, false)
}

// mag is the handler for the magazine page.
func mag(z *zap.SugaredLogger, c echo.Context, chronological bool) error {
	const title, name = "Magazine", "magazine"
	if z == nil {
		return InternalErr(z, c, name, ErrZap)
	}
	data := empty(c)
	const lead = "The magazines are newsletters, reports, and publications about activities within The Scene subculture."
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
	if chronological {
		if err := r.Magazine(ctx, db); err != nil {
			return DatabaseErr(z, c, name, err)
		}
	} else {
		if err := r.MagazineAZ(ctx, db); err != nil {
			return DatabaseErr(z, c, name, err)
		}
	}
	m := model.Summary{}
	if err := m.Magazine(ctx, db); err != nil {
		return DatabaseErr(z, c, name, err)
	}
	data[key] = r
	data["stats"] = map[string]string{
		"pubs":   fmt.Sprintf("%d publications", len(r)),
		"issues": string(ByteFileS(issue, m.SumCount.Int64, m.SumBytes.Int64)),
		"years":  fmt.Sprintf("%d - %d", m.MinYear.Int16, m.MaxYear.Int16),
	}
	err = c.Render(http.StatusOK, name, data)
	if err != nil {
		return InternalErr(z, c, name, err)
	}
	return nil
}

// Releaser is the handler for the releaser page ordered by the most files.
func Releaser(z *zap.SugaredLogger, c echo.Context) error {
	return rel(z, c, true)
}

// ReleaserAZ is the handler for the releaser page ordered alphabetically.
func ReleaserAZ(z *zap.SugaredLogger, c echo.Context) error {
	return rel(z, c, false)
}

// rel is the handler for the Releaser page.
func rel(z *zap.SugaredLogger, c echo.Context, prolific bool) error {
	const title, name = "Releaser", "releaser"
	if z == nil {
		return InternalErr(z, c, name, ErrZap)
	}
	data := empty(c)
	const lead = "A releaser is a brand or a collective group of sceners responsible for releasing or distributing products."
	const key = "releasers"
	data["title"] = title
	data["description"] = fmt.Sprint(title, " ", lead)
	data["logo"] = "Groups, organizations and publications"
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
	if err := r.All(ctx, db, prolific); err != nil {
		return DatabaseErr(z, c, name, err)
	}
	var m model.Summary
	if err := m.All(ctx, db); err != nil {
		return DatabaseErr(z, c, name, err)
	}
	data[key] = r
	data["stats"] = map[string]string{
		"pubs":   fmt.Sprintf("%d releasers and groups", len(r)),
		"issues": string(ByteFileS("file artifact", m.SumCount.Int64, m.SumBytes.Int64)),
		"years":  helper.Years(m.MinYear.Int16, m.MaxYear.Int16),
	}
	err = c.Render(http.StatusOK, name, data)
	if err != nil {
		return InternalErr(z, c, name, err)
	}
	return nil
}
