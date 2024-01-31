package app

// Package file render_file.go contains the renderers that use the file.html template.

import (
	"context"
	"fmt"

	"github.com/Defacto2/server/internal/helper"
	"github.com/Defacto2/server/internal/postgres"
)

// fileWStats is a helper function for File that adds the statistics to the data map.
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

// counter returns the statistics for the file categories.
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
