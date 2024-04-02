package htmx

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/Defacto2/server/internal/helper"
	"github.com/Defacto2/server/internal/postgres"
	"github.com/Defacto2/server/internal/pouet"
	"github.com/Defacto2/server/model"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// PouetProd fetches the multiple download_links values from the
// Pouet production API and attempts to download and save one of the
// linked files. If multiple links are found, the first link is used as
// they should all point to the same asset.
//
// Both the Pouet production ID param and the Defacto2 UUID query
// param values are required as params to fetch the production data and
// to save the file to the correct filename.
func PouetProd(c echo.Context) error {
	sid := c.FormValue("pouet-submission")
	id, err := strconv.Atoi(sid)
	if err != nil {
		return c.String(http.StatusNotAcceptable,
			"The Pouet production ID must be a numeric value, "+sid)
	}

	db, err := postgres.ConnectDB()
	if err != nil {
		return ErrDB
	}
	defer db.Close()
	ctx := context.Background()

	deleted, key, err := model.FindPouetFile(ctx, db, int64(id))
	if err != nil {
		return c.String(http.StatusServiceUnavailable,
			"error, the database query failed")
	}
	if key != 0 && !deleted {
		html := fmt.Sprintf("This Pouet production is already <a href=\"/f/%s\">in use</a>.", helper.ObfuscateID(key))
		return c.HTML(http.StatusOK, html)
	}
	if key != 0 && deleted {
		return c.HTML(http.StatusOK, "This Pouet production is already in use.")
	}

	resp, err := PouetValid(c, id)
	if err != nil {
		return err
	} else if resp.Prod.ID == "" {
		return nil
	}
	if !resp.Success {
		return c.String(http.StatusNotFound, "error, the Pouet production ID is not found")
	}

	prod := resp.Prod
	if pid, err := strconv.Atoi(prod.ID); err != nil {
		return c.String(http.StatusNotFound, "error, the Pouet production ID is invalid")
	} else if pid < 1 {
		return nil
	}

	info := []string{prod.Title}
	if len(prod.Groups) > 0 {
		info = append(info, "by")
		for _, a := range prod.Groups {
			info = append(info, a.Name)
		}
	}
	if prod.ReleaseDate != "" {
		info = append(info, "on", prod.ReleaseDate)
	}
	if prod.Platfs.String() != "" {
		info = append(info, "for", prod.Platfs.String())
	}

	html := `<div class="d-grid gap-2">`
	html += fmt.Sprintf(`<button type="button" class="btn btn-outline-success" `+
		`hx-post="/pouet/production/submit/%d" hx-target="#pouet-submission-results" hx-trigger="click once delay:500ms" `+
		`autofocus>Submit ID %d</button>`, id, id)
	html += `</div>`
	html += fmt.Sprintf(`<p class="mt-3">%s</p>`, strings.Join(info, " "))
	return c.HTML(http.StatusOK, html)
}

// PouetValid fetches the first usable download link from the Pouet API.
// The production ID is validated and the production is checked to see if it
// is suitable for Defacto2. If the production is not suitable, an empty
// production is returned with a htmx message.
func PouetValid(c echo.Context, id int) (pouet.Response, error) {
	if id < 1 {
		return pouet.Response{},
			c.String(http.StatusNotAcceptable, fmt.Sprintf("invalid id: %d", id))
	}

	var prod pouet.Response
	if err := prod.Get(id); err != nil {
		return pouet.Response{}, c.String(http.StatusNotFound, err.Error())
	}

	plat := prod.Prod.Platfs
	sect := prod.Prod.Types
	if !plat.Valid() || !sect.Valid() {
		return pouet.Response{}, c.HTML(http.StatusOK,
			fmt.Sprintf("Production %d is probably not suitable for Defacto2."+
				"<br>A production must an intro, demo or cracktro either for MsDos or Windows.", id))
	}

	var valid string
	if prod.Prod.Download != "" {
		valid = prod.Prod.Download
	}

	for _, link := range prod.Prod.DownloadLinks {
		if valid != "" {
			break
		}
		if link.Link == "" {
			continue
		}
		switch strings.ToLower(link.Type) {
		case "youtube":
			continue
		}
		valid = link.Link
		break
	}
	if valid == "" {
		return pouet.Response{},
			c.String(http.StatusOK, "This Pouet production has no suitable download links.")
	}
	return prod, nil
}

// PouetSubmit is the handler for the /pouet/production/submit route.
// This will attempt to insert a new file record into the database using
// the Pouet production ID. If the Pouet production ID is already in
// use, an error message is returned.
func PouetSubmit(logr *zap.SugaredLogger, c echo.Context) error {
	if logr == nil {
		return c.String(http.StatusInternalServerError,
			"error, pouet submit logger is nil")
	}

	sid := c.Param("id")
	id, err := strconv.ParseUint(sid, 10, 64)
	if err != nil {
		return c.String(http.StatusNotAcceptable,
			"The Pouet production ID must be a numeric value, "+sid)
	}
	if id < 1 || id > pouet.Sanity {
		return c.String(http.StatusNotAcceptable,
			"The Pouet production ID is invalid, "+sid)
	}

	db, err := postgres.ConnectDB()
	if err != nil {
		return ErrDB
	}
	defer db.Close()
	ctx := context.Background()

	if exist, err := model.ExistPouetFile(ctx, db, int64(id)); err != nil {
		return c.String(http.StatusServiceUnavailable,
			"error, the database query failed")
	} else if exist {
		return c.String(http.StatusForbidden,
			"error, the pouet key is already in use")
	}

	key, err := model.InsertPouetFile(ctx, db, int64(id))
	if err != nil || key == 0 {
		logr.Error(err, id)
		return c.String(http.StatusServiceUnavailable,
			"error, the database insert failed")
	}

	html := fmt.Sprintf("Thanks for the submission of Pouet production: %d", id)
	return c.HTML(http.StatusOK, html)
}
