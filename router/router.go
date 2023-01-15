// Package router provides all the functions for the Echo web framework.
// https://echo.labstack.com
package router

import (
	"context"
	"net/http"
	"path/filepath"

	"github.com/Defacto2/server/helpers"
	"github.com/Defacto2/server/models"
	"github.com/Defacto2/server/postgres"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
)

// Download serves files to the user and prompts for a save location.
// The download relies on the URL ID parameter to determine the requested file.
func Download(c echo.Context) error {
	// https://go.dev/src/net/http/status.go
	// get id
	id := helpers.Deobfuscate(c.Param("id"))
	if id <= 0 {
		return echo.NewHTTPError(http.StatusNotFound,
			"The download record cannot be located on the server")
	}
	// get record id, filename, uuid
	ctx := context.Background()
	db, err := postgres.ConnectDB()
	if err != nil {
		return echo.NewHTTPError(http.StatusServiceUnavailable,
			"The database is temporarily down")
	}
	defer db.Close()
	res, err := models.Download(id, ctx, db)
	if err != nil {
		return echo.NewHTTPError(http.StatusServiceUnavailable,
			"The database is temporarily down")
	}
	if res.ID != int64(id) {
		return echo.NewHTTPError(http.StatusNotFound,
			"The download record cannot be located on the server")
	}
	// build the source filepath
	file := filepath.Join("public", "images", "html3", "burst.gif")
	if !helpers.IsStat(file) {
		return echo.NewHTTPError(http.StatusNotFound,
			"The file for download cannot be located on the server")
	}
	// pass the original filename to the client browser
	name := res.Filename.String
	if name == "" {
		log.Info("no filename exists for record: %d", id)
		name = file
	}
	return c.Attachment(file, name)
}
