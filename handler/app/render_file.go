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
	const title, name = "File categories", "file"
	if z == nil {
		return InternalErr(z, c, name, ErrZap)
	}
	data := empty()
	data["title"] = title
	data["description"] = "A table of contents for the collection."
	data["logo"] = title
	data["h1"] = title
	data["lead"] = "This page shows the file categories and platforms in the collection."
	data["stats"] = stats
	data["counter"] = Stats{}

	data, err := fileWStats(data, stats)
	if err != nil {
		return InternalErr(z, c, name, err)
	}
	err = c.Render(http.StatusOK, name, data)
	if err != nil {
		return InternalErr(z, c, name, err)
	}
	return nil
}

func fileWStats(data map[string]interface{}, stats bool) (map[string]interface{}, error) {
	if !stats {
		return data, nil
	}
	c, err := counter()
	if err != nil {
		return nil, err
	}
	data["counter"] = c
	data["logo"] = "File category statistics"
	data["lead"] = "This page shows the file categories with selected statistics, " +
		"such as the number of files in the category or platform." +
		fmt.Sprintf(" The total number of files in the database is %d.", c.All.Count) +
		fmt.Sprintf(" The total size of all files in the database is %s.", helper.ByteCount(int64(c.All.Bytes)))
	return data, nil
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
