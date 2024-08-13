// Package htmx handles the routes and views for the AJAX responses using the htmx library.
package htmx

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Defacto2/releaser/initialism"
	"github.com/Defacto2/server/internal/demozoo"
	"github.com/Defacto2/server/internal/helper"
	"github.com/Defacto2/server/internal/postgres"
	"github.com/Defacto2/server/internal/pouet"
	"github.com/Defacto2/server/model"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

var (
	ErrDB     = errors.New("database connection is nil")
	ErrFormat = errors.New("invalid format")
	ErrKey    = errors.New("numeric record key is invalid")
)

// DemozooProd fetches the multiple download_links values from the
// Demozoo production API and attempts to download and save one of the
// linked files. If multiple links are found, the first link is used as
// they should all point to the same asset.
//
// Both the Demozoo production ID param and the Defacto2 UUID query
// param values are required as params to fetch the production data and
// to save the file to the correct filename.
func DemozooProd(c echo.Context, db *sql.DB) error {
	sid := c.FormValue("demozoo-submission")
	id, err := strconv.Atoi(sid)
	if err != nil {
		return c.String(http.StatusNotAcceptable,
			"The Demozoo production ID must be a numeric value, "+sid)
	}
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
		`hx-put="/demozoo/production/%d" `+
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

// DemozooSubmit is the handler for the /demozoo/production put route.
// This will attempt to insert a new file record into the database using
// the Demozoo production ID. If the Demozoo production ID is already in
// use, an error message is returned.
func DemozooSubmit(c echo.Context, db *sql.DB, logger *zap.SugaredLogger) error {
	return submit(c, db, logger, "demozoo")
}

// DBConnections is the handler for the database connections page.
func DBConnections(c echo.Context, db *sql.DB) error {
	conns, max, err := postgres.Connections(db)
	if err != nil {
		return c.String(http.StatusOK, err.Error())
	}
	currentTime := time.Now()
	return c.String(http.StatusOK, fmt.Sprintf("%d of %d, <small>%s</small>",
		conns, max, currentTime.Format("15:04:05")))
}

// DeleteForever is a handler for the /delete/forever route.
func DeleteForever(c echo.Context, db *sql.DB, logger *zap.SugaredLogger, id string) error {
	key, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return c.String(http.StatusNotFound, err.Error())
	}
	ctx := context.Background()
	tx, err := db.Begin()
	if err != nil {
		logger.Error(err)
		return c.String(http.StatusServiceUnavailable,
			"cannot begin a transaction")
	}
	if err = model.DeleteOne(ctx, tx, key); err != nil {
		defer func() {
			if err := tx.Rollback(); err != nil {
				logger.Error(err)
			}
		}()
		logger.Error(err)
		return c.String(http.StatusServiceUnavailable,
			"cannot delete the record")
	}
	//
	// There is no need to delete any file assets from the file system.
	// As the file assets will be deleted by the next cleanup job.
	//
	if err = tx.Commit(); err != nil {
		logger.Error(err)
		return c.String(http.StatusServiceUnavailable,
			"cannot commit the transaction")
	}
	return c.String(http.StatusOK,
		"The artifact is gone, and reloading this page will result in a 404 error.")
}

func pings() []string {
	return []string{
		"/this-is-an-invalid-url",
		"/html3",
		"/html3/groups",
		"/html3/group/2000ad",
		"/html3/group/2000ad?C=N&O=D",
		"/html3/platform/audio?C=N&O=A",
		"/html3/platform/audio?C=N&O=D",
		"/html3/platform/audio?C=D&O=A",
		"/html3/platform/audio?C=D&O=D",
		"/html3/platform/audio?C=P&O=A",
		"/html3/platform/audio?C=P&O=D",
		"/html3/platform/audio?C=S&O=A",
		"/html3/platform/audio?C=S&O=D",
		"/html3/platform/audio?C=I&O=A",
		"/html3/platform/audio?C=I&O=D",
		"/html3/categories",
		"/html3/category/ansieditor",
		"/html3/category/ansieditor?C=N&O=D",
		"/html3/art/1",
		"/html3/art/1?C=N&O=D",
		"/html3/documents",
		"/html3/software",
		"/html3/all",
		"/editor/for-approval",
		"/files/new-uploads",
		"/files/new-updates",
		"/files/oldest",
		"/files/newest",
		"/file",
		"/file/stats",
		"/files/installer",
		"/files/installer/2",
		"/releaser",
		"/releaser/a-z",
		"/releaser/year",
		"/g/the-grand-council",
		"/magazine",
		"/magazine/a-z",
		"/ftp",
		"/bbs",
		"/bbs/a-z",
		"/bbs/year",
		"/scener",
		"/interview",
		"/artist",
		"/coder",
		"/musician",
		"/writer",
		"/p/200mhz",
		"/website",
		"/website/hide",
		"/search/releaser",
		"/search/file",
		"/search/desc",
		"/editor/search/id",
		"/history",
		"/thescene",
		"/thanks",
	}
}

// Pings is a handler for the /pings route.
func Pings(c echo.Context, proto string, port int) error {
	pings := pings()
	results := make([]string, 0, len(pings))
	for _, ping := range pings {
		code, size, err := helper.LocalHostPing(ping, proto, port)
		if err != nil {
			results = append(results, fmt.Sprintf("%s: %v", ping, err))
			continue
		}
		var elm string
		switch {
		case code == http.StatusOK:
			elm = "<span class=\"text-success\">"
		case code >= http.StatusInternalServerError:
			elm = "<span class=\"text-danger\">"
		default:
			elm = "<span class=\"text-warning\">"
		}
		results = append(results,
			"<div>", elm,
			fmt.Sprintf("%d</span> %s %s", code, ping, helper.ByteCount(size)),
			"</div>")
	}
	output := strings.Join(results, "")
	output += fmt.Sprintf("<div><small>%d URLs were pinged</small></div>", len(pings))
	return c.HTML(http.StatusOK, output)
}

// PouetProd fetches the multiple download_links values from the
// Pouet production API and attempts to download and save one of the
// linked files. If multiple links are found, the first link is used as
// they should all point to the same asset.
//
// Both the Pouet production ID param and the Defacto2 UUID query
// param values are required as params to fetch the production data and
// to save the file to the correct filename.
func PouetProd(c echo.Context, db *sql.DB) error {
	sid := c.FormValue("pouet-submission")
	id, err := strconv.Atoi(sid)
	if err != nil {
		return c.String(http.StatusNotAcceptable,
			"The Pouet production ID must be a numeric value, "+sid)
	}
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
		`hx-put="/pouet/production/%d" hx-target="#pouet-submission-results" hx-trigger="click once delay:500ms" `+
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

// PouetSubmit is the handler for the /pouet/production PUT route.
// This will attempt to insert a new file record into the database using
// the Pouet production ID. If the Pouet production ID is already in
// use, an error message is returned.
func PouetSubmit(c echo.Context, db *sql.DB, logger *zap.SugaredLogger) error {
	return submit(c, db, logger, "pouet")
}

// SearchByID is a handler for the /editor/search/id route.
func SearchByID(c echo.Context, db *sql.DB, logger *zap.SugaredLogger) error {
	const maxResults = 50
	ctx := context.Background()
	ids := []int{}
	uuids := []uuid.UUID{}
	search := c.FormValue("htmx-search")
	inputs := strings.Split(search, " ")
	for _, input := range inputs {
		x := strings.ToLower(strings.TrimSpace(input))
		if id, _ := strconv.Atoi(x); id > 0 {
			ids = append(ids, id)
			continue
		}
		if id := helper.DeobfuscateID(x); id > 0 {
			ids = append(ids, id)
			continue
		}
		if uid, err := uuid.Parse(x); err == nil {
			uuids = append(uuids, uid)
			continue
		}
	}

	var r model.Artifacts
	fs, err := r.ID(ctx, db, ids, uuids...)
	if err != nil {
		logger.Error(err)
		return c.String(http.StatusServiceUnavailable,
			"the search query failed")
	}

	if len(fs) == 0 {
		return c.HTML(http.StatusOK, "No artifacts found.")
	}
	err = c.Render(http.StatusOK, "searchids", map[string]interface{}{
		"maximum": maxResults,
		"name":    search,
		"result":  fs,
	})
	if err != nil {
		logger.Errorf("search by id htmx template: %v", err)
		return c.String(http.StatusInternalServerError,
			"cannot render the htmx search by id template")
	}
	return nil
}

// SearchReleaser is a handler for the /search/releaser route.
func SearchReleaser(c echo.Context, db *sql.DB, logger *zap.SugaredLogger) error {
	const maxResults = 14
	ctx := context.Background()
	input := c.FormValue("htmx-search")
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
	err := c.Render(http.StatusOK, "searchreleasers", map[string]interface{}{
		"maximum": maxResults,
		"name":    slug,
		"result":  r,
	})
	if err != nil {
		return c.String(http.StatusInternalServerError,
			"cannot render the htmx search releases template")
	}
	return nil
}

// DataListReleasers is a handler for the /datalist/releasers route.
func DataListReleasers(c echo.Context, db *sql.DB, logger *zap.SugaredLogger, input string) error {
	return datalist(c, db, logger, input, false)
}

// DataListMagazines is a handler for the /datalist/magazines route.
func DataListMagazines(c echo.Context, db *sql.DB, logger *zap.SugaredLogger, input string) error {
	return datalist(c, db, logger, input, true)
}

// datalist is a shared handler for the /datalist/releasers and /datalist/magazines routes.
func datalist(c echo.Context, db *sql.DB, logger *zap.SugaredLogger, input string, magazine bool) error {
	const maxResults = 14
	ctx := context.Background()
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
	var err error
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
	err = c.Render(http.StatusOK, "datalistreleasers", map[string]interface{}{
		"maximum": maxResults,
		"name":    slug,
		"result":  r,
	})
	if err != nil {
		return c.String(http.StatusInternalServerError,
			"cannot render the htmx datalist releases template")
	}
	return nil
}
