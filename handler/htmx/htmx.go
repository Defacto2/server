// Package htmx handles the routes and views for the AJAX responses using the htmx library.
package htmx

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/Defacto2/releaser/initialism"
	"github.com/Defacto2/server/internal/demozoo"
	"github.com/Defacto2/server/internal/helper"
	"github.com/Defacto2/server/internal/postgres"
	"github.com/Defacto2/server/internal/pouet"
	"github.com/Defacto2/server/model"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

var (
	ErrCreators     = errors.New("invalid reset creators format")
	ErrDate         = errors.New("invalid reset date format")
	ErrDB           = errors.New("database connection is nil")
	ErrExist        = errors.New("file already exists")
	ErrFile         = errors.New("cannot be a file")
	ErrFileHead     = errors.New("file header is nil")
	ErrKey          = errors.New("numeric record key is invalid")
	ErrRoutes       = errors.New("echo instance is nil")
	ErrUploaderDest = errors.New("invalid uploader destination")
	ErrUploaderSave = errors.New("cannot save a file to the uploader destination")
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

	deleted, key, err := model.OneDemozoo(ctx, db, int64(id))
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

	deleted, key, err := model.OnePouet(ctx, db, int64(id))
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
		return fmt.Errorf("PouetValid: %w", err)
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
	return c.HTML(http.StatusOK, htmler(id, info...))
}

func htmler(id int, info ...string) string {
	s := `<div class="d-grid gap-2">`
	s += fmt.Sprintf(`<button type="button" class="btn btn-outline-success" `+
		`hx-post="/pouet/production/submit/%d" hx-target="#pouet-submission-results" hx-trigger="click once delay:500ms" `+
		`autofocus>Submit ID %d</button>`, id, id)
	s += `</div>`
	s += fmt.Sprintf(`<p class="mt-3">%s</p>`, strings.Join(info, " "))
	return s
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
		if strings.Contains(link.Link, "youtube") {
			continue
		}
		if strings.Contains(link.Link, "sourceforge") {
			continue
		}
		if strings.Contains(link.Link, "github") {
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
func PouetSubmit(c echo.Context, logger *zap.SugaredLogger) error {
	return submit(c, logger, "pouet")
}

// SearchReleaser is a handler for the /search/releaser route.
func SearchReleaser(c echo.Context, logger *zap.SugaredLogger) error {
	const maxResults = 14
	ctx := context.Background()
	db, err := postgres.ConnectDB()
	if err != nil {
		logger.Error(err)
		return c.String(http.StatusServiceUnavailable,
			"cannot connect to the database")
	}
	defer db.Close()

	input := c.FormValue("releaser-search")
	slug := helper.Slug(helper.TrimRoundBraket(input))
	if slug == "" {
		return c.HTML(http.StatusOK, "<!-- empty search query -->")
	}

	lookup := []string{}
	for key, values := range initialism.Initialisms() {
		for _, value := range values {
			if strings.Contains(strings.ToLower(value), strings.ToLower(slug)) {
				lookup = append(lookup, string(key))
			}
		}
	}
	lookup = append(lookup, slug)
	var r model.Releasers
	if err := r.Similar(ctx, db, maxResults, lookup...); err != nil {
		logger.Error(err)
		return c.String(http.StatusServiceUnavailable,
			"the search query failed")
	}
	if len(r) == 0 {
		return c.HTML(http.StatusOK, "No releasers found.")
	}
	err = c.Render(http.StatusOK, "releasers", map[string]interface{}{
		"maximum": maxResults,
		"name":    slug,
		"result":  r,
	})
	if err != nil {
		return c.String(http.StatusInternalServerError,
			"cannot render the htmx template")
	}
	return nil
}

// DataListReleasers is a handler for the /datalist/releasers route.
func DataListReleasers(c echo.Context, logger *zap.SugaredLogger, input string) error {
	return datalist(c, logger, input, false)
}

// DataListMagazines is a handler for the /datalist/magazines route.
func DataListMagazines(c echo.Context, logger *zap.SugaredLogger, input string) error {
	return datalist(c, logger, input, true)
}

// datalist is a shared handler for the /datalist/releasers and /datalist/magazines routes.
func datalist(c echo.Context, logger *zap.SugaredLogger, input string, magazine bool) error {
	const maxResults = 14
	ctx := context.Background()
	db, err := postgres.ConnectDB()
	if err != nil {
		logger.Error(err)
		return c.String(http.StatusServiceUnavailable,
			"cannot connect to the database")
	}
	defer db.Close()

	slug := helper.Slug(helper.TrimRoundBraket(input))
	if slug == "" {
		return c.HTML(http.StatusOK, "")
	}

	lookup := []string{}
	for key, values := range initialism.Initialisms() {
		for _, value := range values {
			if strings.Contains(strings.ToLower(value), strings.ToLower(slug)) {
				lookup = append(lookup, string(key))
			}
		}
	}
	lookup = append(lookup, slug)
	var r model.Releasers
	if magazine {
		err = r.SimilarMagazine(ctx, db, maxResults, lookup...)
	} else {
		err = r.Similar(ctx, db, maxResults, lookup...)
	}
	if err != nil {
		logger.Error(err)
		return c.String(http.StatusServiceUnavailable,
			"cannot connect to the database")
	}
	if len(r) == 0 {
		return c.HTML(http.StatusOK, "")
	}
	err = c.Render(http.StatusOK, "releasersdl", map[string]interface{}{
		"maximum": maxResults,
		"name":    slug,
		"result":  r,
	})
	if err != nil {
		return c.String(http.StatusInternalServerError,
			"cannot render the htmx template")
	}
	return nil
}
