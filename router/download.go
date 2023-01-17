package router

import (
	"context"
	"net/http"
	"path/filepath"

	"github.com/Defacto2/server/helpers"
	"github.com/Defacto2/server/logger"
	"github.com/Defacto2/server/models"
	"github.com/Defacto2/server/postgres"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type Log struct {
	IsProduction bool
	ConfigDir    string
}

const (
	invalidID = 0
	dbdown    = "The database is temporarily down"
	missing   = "The file for download cannot be located on the server"
	notfound  = "The download record cannot be located on the server"
)

// Download serves files to the user and prompts for a save location.
// The download relies on the URL ID parameter to determine the requested file.
func (l Log) Download(c echo.Context) error {
	// Logger
	var log *zap.SugaredLogger
	switch l.IsProduction {
	case true:
		log = logger.Production(l.ConfigDir).Sugar()
		defer log.Sync()
	default:
		log = logger.Development().Sugar()
		defer log.Sync()
	}
	return download(log, c)
}

func download(log *zap.SugaredLogger, c echo.Context) error {
	// https://go.dev/src/net/http/status.go
	// get id
	id := helpers.Deobfuscate(c.Param("id"))
	if id <= invalidID {
		return echo.NewHTTPError(http.StatusNotFound, notfound)
	}
	// get record id, filename, uuid
	ctx := context.Background()
	db, err := postgres.ConnectDB()
	if err != nil {
		return echo.NewHTTPError(http.StatusServiceUnavailable, dbdown)
	}
	defer db.Close()
	res, err := models.File(id, ctx, db)
	if err != nil {
		return echo.NewHTTPError(http.StatusServiceUnavailable, dbdown)
	}
	if res.ID != int64(id) {
		return echo.NewHTTPError(http.StatusNotFound, notfound)
	}
	// build the source filepath
	file := filepath.Join("public", "images", "html3", "burst.xgif") // TODO: replace this placeholder
	if !helpers.IsStat(file) {
		log.Warnf("The hosted file download %q, for record %d does not exist.", res.Filename.String, res.ID)
		return echo.NewHTTPError(http.StatusFailedDependency, missing)
	}
	// pass the original filename to the client browser
	name := res.Filename.String
	if name == "" {
		log.Warnf("No filename exists for the record %d.", res.ID)
		name = file
	}
	return c.Attachment(file, name)
}
