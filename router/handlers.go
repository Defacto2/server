package router

import (
	"context"
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/bengarrett/df2023/helpers"
	"github.com/bengarrett/df2023/models"
	"github.com/bengarrett/df2023/postgres"
	"github.com/labstack/echo/v4"
)

func Download(c echo.Context) error {
	// https://go.dev/src/net/http/status.go
	uri := c.Param("id")
	// get id
	id := helpers.Deobfuscate(uri)
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
	// build filepath
	file := filepath.Join("public", "images", "html3", "burst.gifx")
	if !helpers.IsExist(file) {
		return echo.NewHTTPError(http.StatusNotFound,
			"The file for download cannot be located on the server")
	}
	// check local file exists
	fmt.Printf("\nFILE DOWNLOAD: %s\n",
		res.Filename.String)
	// print log to console
	name := res.Filename.String
	if name == "" {
		// log
		name = file
	}
	return c.Attachment(file, name)
}

func DownloadX(c echo.Context) error {
	return Download(c)
	// return c.Render(http.StatusOK, "categories", map[string]interface{}{
	// TODO: if err then render a HTML3 template error
}
