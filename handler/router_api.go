package handler

// Package file router_api.go contains the API router URIs for the website.

import (
	"database/sql"
	"embed"
	"log/slog"

	"github.com/Defacto2/server/handler/app"
	"github.com/Defacto2/server/internal/panics"
	"github.com/labstack/echo/v4"
)

// APIRoutes defines the API routes for the web server.
func (c *Configuration) APIRoutes(e *echo.Echo, db *sql.DB, sl *slog.Logger, public embed.FS) (*echo.Echo, error) {
	const msg = "api routes"
	if err := panics.EchoDSP(e, db, sl, public); err != nil {
		panic(msg + ": " + err.Error())
	}

	// Milestone API routes
	e.GET("/api/milestones", func(c echo.Context) error { return app.GetAllMilestones(c) })
	e.GET("/api/milestones/highlights", func(c echo.Context) error { return app.GetHighlightedMilestones(c) })
	e.GET("/api/milestones/year/:year", func(c echo.Context) error { return app.GetMilestonesByYear(c) })
	e.GET("/api/milestones/years/:range", func(c echo.Context) error { return app.GetMilestonesByYearRange(c) })
	e.GET("/api/milestones/decade/:decade", func(c echo.Context) error { return app.GetMilestonesByDecade(c) })

	return e, nil
}