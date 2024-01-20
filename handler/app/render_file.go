package app

// Package file render_file.go contains the renderers that use the file.html template.

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Defacto2/server/internal/helper"
	"github.com/Defacto2/server/internal/postgres"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// File is the handler for the file categories page.
func File(z *zap.SugaredLogger, c echo.Context, stats bool) error {
	const title, name = "File categories", "file"
	if z == nil {
		return InternalErr(z, c, name, ErrZap)
	}
	data := empty(c)
	data["title"] = title
	data["description"] = "A table of contents for the collection."
	data["logo"] = title
	data["h1"] = title
	data["lead"] = "This page shows the categories and platforms in the collection of file artifacts."
	data["stats"] = stats
	data["counter"] = Stats{}

	data, err := fileWStats(data, stats)
	if err != nil {
		z.Warn(err)
		data["dberror"] = true
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
		return data, err
	}
	data["counter"] = c
	data["logo"] = "File category statistics"
	data["lead"] = "This page shows the file categories with selected statistics, " +
		"such as the number of files in the category or platform." +
		fmt.Sprintf(" The total number of files in the database is %d.", c.Record.Count) +
		fmt.Sprintf(" The total size of all file artifacts are %s.", helper.ByteCount(int64(c.Record.Bytes)))
	return data, nil
}

func counter() (Stats, error) {
	ctx := context.Background()
	db, err := postgres.ConnectDB()
	if err != nil {
		return Stats{}, err
	}
	defer db.Close()
	counter := Stats{}
	if err := counter.Get(ctx, db); err != nil {
		return Stats{}, err
	}
	return counter, nil
}
