package htmx

// Package file demozoo.go provides functions for handling the HTMX requests for the Demozoo production uploader.

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/Defacto2/server/internal/demozoo"
	"github.com/Defacto2/server/internal/helper"
	"github.com/Defacto2/server/internal/postgres"
	"github.com/Defacto2/server/model"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// DemozooProd fetches the multiple download_links values from the
// Demozoo production API and attempts to download and save one of the
// linked files. If multiple links are found, the first link is used as
// they should all point to the same asset.
//
// Both the Demozoo production ID param and the Defacto2 UUID query
// param values are required as params to fetch the production data and
// to save the file to the correct filename.
func DemozooProd(c echo.Context) error {
	sid := c.FormValue("demozoo-submission")
	id, err := strconv.Atoi(sid)
	if err != nil {
		return c.String(http.StatusNotAcceptable,
			"The Demozoo production ID must be a numeric value, "+sid)
	}

	db, err := postgres.ConnectDB()
	if err != nil {
		return ErrDB
	}
	defer db.Close()
	ctx := context.Background()

	deleted, key, err := model.FindDemozooFile(ctx, db, int64(id))
	if err != nil {
		return c.String(http.StatusServiceUnavailable,
			"error, the database query failed")
	}
	if key != 0 && !deleted {
		html := fmt.Sprintf("This Demozoo production is already <a href=\"/f/%s\">in use</a>.", helper.ObfuscateID(key))
		return c.HTML(http.StatusOK, html)
	}
	if key != 0 && deleted {
		return c.HTML(http.StatusOK, "This Demozoo production is already in use.")
	}

	prod, err := DemozooValid(c, id)
	if err != nil {
		return fmt.Errorf("demozoo.DemozooValid: %w", err)
	}
	if prod.ID < 1 {
		return nil
	}

	info := []string{prod.Title}
	if len(prod.Authors) > 0 {
		info = append(info, "by")
		for _, a := range prod.Authors {
			info = append(info, a.Name)
		}
	}
	if prod.ReleaseDate != "" {
		info = append(info, "on", prod.ReleaseDate)
	}
	if prod.Platforms != nil {
		info = append(info, "for")
		for _, p := range prod.Platforms {
			info = append(info, p.Name)
		}
	}
	html := `<div class="d-grid gap-2">`
	html += fmt.Sprintf(`<button type="button" class="btn btn-outline-success" `+
		`hx-post="/demozoo/production/submit/%d" `+
		`hx-target="#demozoo-submission-results" hx-trigger="click once delay:500ms" `+
		`autofocus>Submit ID %d</button>`, id, id)
	html += `</div>`
	html += fmt.Sprintf(`<p class="mt-3">%s</p>`, strings.Join(info, " "))
	return c.HTML(http.StatusOK, html)
}

// DemozooValid fetches the first usable download link from the Demozoo API.
// The production ID is validated and the production is checked to see if it
// is suitable for Defacto2. If the production is not suitable, an empty
// production is returned with a htmx message.
func DemozooValid(c echo.Context, id int) (demozoo.Production, error) {
	if id < 1 {
		return demozoo.Production{},
			c.String(http.StatusNotAcceptable, fmt.Sprintf("invalid id: %d", id))
	}

	var prod demozoo.Production
	if code, err := prod.Get(id); err != nil {
		return demozoo.Production{}, c.String(code, err.Error())
	}
	plat, sect := prod.SuperType()
	if plat == -1 || sect == -1 {
		s := []string{}
		for _, p := range prod.Platforms {
			s = append(s, p.Name)
		}
		for _, t := range prod.Types {
			s = append(s, t.Name)
		}
		return demozoo.Production{}, c.HTML(http.StatusOK,
			fmt.Sprintf("Production %d is probably not suitable for Defacto2.<br>Types: %s",
				id, strings.Join(s, " - ")))
	}

	var valid string
	for _, link := range prod.DownloadLinks {
		if link.URL == "" {
			continue
		}
		valid = link.URL
		break
	}
	if valid == "" {
		return demozoo.Production{},
			c.String(http.StatusOK, "This Demozoo production has no suitable download links.")
	}
	return prod, nil
}

// DemozooSubmit is the handler for the /demozoo/production/submit route.
// This will attempt to insert a new file record into the database using
// the Demozoo production ID. If the Demozoo production ID is already in
// use, an error message is returned.
func DemozooSubmit(c echo.Context, logger *zap.SugaredLogger) error {
	return submit(c, logger, "demozoo")
}
