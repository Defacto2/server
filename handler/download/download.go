// Package download handles the client file downloads.
package download

import (
	"context"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/Defacto2/server/model"
	"github.com/Defacto2/server/pkg/helpers"
	"github.com/Defacto2/server/pkg/postgres"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// Download configuration.
type Download struct {
	Path string // Path is the absolute path to the download directory.
}

const (
	invalidID = 0 // invalidID is the default out of range ID value.
	dbdown    = "The database is temporarily down"
	missing   = "The file for download cannot be located on the server"
	notfound  = "The download record cannot be located on the server"
)

// HTTPSend serves files to the user and prompts for a save location.
// The download relies on the URL ID parameter to determine the requested file.
func (d Download) HTTPSend(log *zap.SugaredLogger, c echo.Context) error {
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
	res, err := model.One(ctx, db, id)
	if err != nil {
		return echo.NewHTTPError(http.StatusServiceUnavailable, dbdown)
	}
	if res.ID != int64(id) {
		return echo.NewHTTPError(http.StatusNotFound, notfound)
	}
	// build the source filepath
	name := res.Filename.String
	uid := strings.TrimSpace(res.UUID.String)
	file := filepath.Join(d.Path, uid)
	if !helpers.IsStat(file) {
		log.Warnf("The hosted file download %q, for record %d does not exist.\nAbsolute path: %q",
			res.Filename.String, res.ID, file)
		return echo.NewHTTPError(http.StatusFailedDependency,
			fmt.Sprintf("%s: %s", missing, name))
	}
	// pass the original filename to the client browser
	if name == "" {
		log.Warnf("No filename exists for the record %d.", res.ID)
		name = file
	}
	return c.Attachment(file, name)
}
