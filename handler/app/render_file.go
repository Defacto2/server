package app

// Package file render_file.go contains the renderers that use the file.html template.

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Defacto2/server/pkg/helper"
	"github.com/Defacto2/server/pkg/postgres"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// File is the handler for the file categories page.
func File(z *zap.SugaredLogger, c echo.Context, stats bool) error {
	if z == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("%w: handler app file", ErrLogger))
	}
	const title = "File categories"
	data := empty()
	data["title"] = title
	data["description"] = "A table of contents for the collection."
	data["logo"] = title
	data["h1"] = title
	data["lead"] = "This page shows the file categories and platforms in the collection."
	data["stats"] = stats
	data["counter"] = Stats{}

	if stats {
		c, err := counter()
		if err != nil {
			return echo.NewHTTPError(http.StatusServiceUnavailable, err)
		}
		data["counter"] = c
		data["logo"] = "File category statistics"
		data["lead"] = "This page shows the file categories with selected statistics, " +
			"such as the number of files in the category or platform." +
			fmt.Sprintf(" The total number of files in the database is %d.", c.All.Count) +
			fmt.Sprintf(" The total size of all files in the database is %s.", helper.ByteCount(int64(c.All.Bytes)))
	}
	err := c.Render(http.StatusOK, "file", data)
	if err != nil {
		z.Errorf("%s: %s", ErrTmpl, err)
		return echo.NewHTTPError(http.StatusInternalServerError, ErrTmpl)
	}
	return nil
}

func counter() (Stats, error) {
	ctx := context.Background()
	db, err := postgres.ConnectDB()
	if err != nil {
		return Stats{}, fmt.Errorf("%w: %s", ErrConn, err)
	}
	defer db.Close()
	counter := Stats{}
	if err := counter.Get(ctx, db); err != nil {
		return Stats{}, fmt.Errorf("%w: %s", ErrConn, err)
	}
	return counter, nil
}
