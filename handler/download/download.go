// Package download handles the client file downloads.
package download

import (
	"context"
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/Defacto2/server/model"
	"github.com/Defacto2/server/pkg/helpers"
	"github.com/Defacto2/server/pkg/logger"
	"github.com/Defacto2/server/pkg/postgres"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// Log configuration.
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

// Send serves files to the user and prompts for a save location.
// The download relies on the URL ID parameter to determine the requested file.
func (l Log) Send(c echo.Context) error {
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
	return Send(log, c)
}

// Send serves files to the user and prompts for a save location.
// The download relies on the URL ID parameter to determine the requested file.
func Send(log *zap.SugaredLogger, c echo.Context) error {
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
	res, err := model.One(id, ctx, db)
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
		return echo.NewHTTPError(http.StatusFailedDependency,
			fmt.Sprintf("%s: %s", missing, filepath.Base(file)))
	}
	// pass the original filename to the client browser
	name := res.Filename.String
	if name == "" {
		log.Warnf("No filename exists for the record %d.", res.ID)
		name = file
	}
	return c.Attachment(file, name)
}
