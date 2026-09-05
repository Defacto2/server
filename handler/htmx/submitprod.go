package htmx

// Package file submitprod.go provides functions for handling the HTMX requests
// for submitting Demozoo and Pouet productions.

import (
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"slices"
	"strconv"
	"strings"

	"github.com/Defacto2/helper"
	"github.com/Defacto2/server/handler/app"
	"github.com/Defacto2/server/handler/cache"
	"github.com/Defacto2/server/handler/demozoo"
	"github.com/Defacto2/server/handler/pouet"
	"github.com/Defacto2/server/handler/sess"
	"github.com/Defacto2/server/internal/dir"
	"github.com/Defacto2/server/internal/nils"
	"github.com/Defacto2/server/model"
	"github.com/labstack/echo/v5"
)

type Prod int

const (
	Demozoo Prod = iota
	Pouet
)

const (
	dz = "demozoo"
	pt = "pouet"
)

func (prod Prod) String() string {
	return [...]string{dz, pt}[prod]
}

// Submit handles the PUT production routes for Demozoo and Pouet.
// This will attempt to insert a new file record into the database using
// the production ID. If the ID is already in use, an error message is returned.
func (prod Prod) Submit(sl *slog.Logger, c *echo.Context, tx *sql.Tx, download dir.Directory) error {
	const msg = "htmx transfer submit"
	if err := nils.Check(sl, c, tx); err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}

	logErr := func(s string, err error) {
		sl.Error(msg, slog.String("problem", s), slog.Any("error", err))
	}

	prodID, err := prod.ID(c)
	if err != nil {
		return err
	}

	ctx := c.Request().Context()
	var exist bool
	var eErr error
	switch prod {
	case Demozoo:
		exist, eErr = model.ExistDemozoo(ctx, tx, prodID)
	case Pouet:
		exist, eErr = model.ExistPouet(ctx, tx, prodID)
	}
	if eErr != nil {
		return c.String(http.StatusServiceUnavailable, "error, the database query failed")
	}
	if exist {
		return c.String(http.StatusForbidden, "error, the "+prod.String()+" key is already in use")
	}

	var key int64
	var unid string
	switch prod {
	case Demozoo:
		key, unid, err = model.InsertDemozoo(ctx, tx, prodID)
	case Pouet:
		key, unid, err = model.InsertPouet(ctx, tx, prodID)
	}
	if err != nil || key == 0 {
		logErr(fmt.Sprintf("cannot insert record id %d", prodID), err)
		return c.String(http.StatusServiceUnavailable, "error, the database insert failed")
	}

	name := strings.ToTitle(prod.String())
	const format = `<div class="text-success">Thanks for the submission of %s production, %d</div>`
	html := fmt.Sprintf(format, name, prodID)
	if sess.Editor(c) {
		uri := helper.ObfuscateID(key)
		const format = `<p data-bs-toggle="tooltip" data-bs-placement="top" data-bs-title="ctrl + alt + enter">` +
			`<a id="go-to-the-new-artifact-record" href="/f/%s" autofocus>Go to the new artifact record</a></p>`
		html += fmt.Sprintf(format, uri)
	}

	// see Download in handler/app/internal/remote/remote.go
	switch prod {
	case Demozoo:
		if err := app.GetDemozoo(ctx, sl, c, tx, prodID, unid, download); err != nil {
			logErr("cannot fetch remote demozoo api", err)
			const format = `<p class="text-danger">error, cannot fetch the remote download linked by %s</p>`
			html += fmt.Sprintf(format, prod.String())
			return c.String(http.StatusServiceUnavailable, html)
		}
	case Pouet:
		if err := app.GetPouet(ctx, sl, c, tx, prodID, unid, download); err != nil {
			logErr("cannot fetch remote pouet api", err)
			const format = `<p class="text-danger">error, cannot fetch the remote download linked by %s</p>`
			html += fmt.Sprintf(format, prod.String())
			return c.String(http.StatusServiceUnavailable, html)
		}
	}

	sl.Info(msg,
		slog.String("okay", "the production has been submitted"),
		slog.String("remote", name), slog.Int("new_id", prodID))

	return c.String(http.StatusOK, html)
}

// ID validates the production ID and ensures that it is a valid numeric value.
func (prod Prod) ID(c *echo.Context) (int, error) {
	const format = "transfer sanitize id: %w"
	if err := nils.Check(c); err != nil {
		return 0, fmt.Errorf(format, err)
	}

	name := strings.ToTitle(prod.String())
	id, err := echo.PathParam[int](c, "id")
	if err != nil {
		return 0, c.String(http.StatusNotAcceptable,
			"The "+name+" production ID must be a numeric value")
	}

	var sanity int
	switch prod {
	case Demozoo:
		sanity = demozoo.Sanity
	case Pouet:
		sanity = pouet.Sanity
	}
	if id < 1 || id > sanity {
		const format = `The %q production ID is invalid, %d`
		return 0, c.String(http.StatusNotAcceptable, fmt.Sprintf(format, name, id))
	}

	return id, nil
}

func (prod Prod) Lookup(c *echo.Context, db *sql.DB, useCache bool) error {
	switch prod {
	case Demozoo:
		return prod.lookupDemozoo(c, db, useCache)
	case Pouet:
		return prod.lookupPouet(c, db, useCache)
	}
	return nil
}

func (prod Prod) ButtonOK(c *echo.Context, prodID int, info ...string) error {
	demozoo := func() string {
		const format = `<button type="button" class="btn btn-outline-success" ` +
			`hx-put="/demozoo/production/%d" ` +
			`hx-indicator="#demozoo-remote-indicator" ` +
			`hx-target="#demozoo-submission-results" ` +
			`hx-swap="innerHTML" ` +
			`hx-trigger="click once delay:500ms" ` +
			`hx-target-error="#demozoo-submission-error" ` +
			`autofocus>Submit ID %d</button>`

		const did = `demozoo-remote-indicator`
		const dclass = `htmx-indicator text-secondary pt-2`
		const sclass = `spinner-border spinner-border-sm`
		const text = `Fetching Download linked by Demozoo...`

		button := fmt.Sprintf(format, prodID, prodID)
		button += `<div id="` + did + `" class="` + dclass + `" role="status">` +
			`  <span class="` + sclass + `"></span> <span>` + text + `</span></div>`
		button += fmt.Sprintf(`<div>%s</div>`, strings.Join(info, " "))

		return `<form class="d-grid">` + button + `</form>`
	}

	pouet := func() string {
		const format = `<button type="button" class="btn btn-outline-success" ` +
			`hx-put="/pouet/production/%d" ` +
			`hx-indicator="#pouet-remote-indicator" ` +
			`hx-target="#pouet-submission-results" ` +
			`hx-swap="innerHTML" ` +
			`hx-trigger="click once delay:500ms" ` +
			`hx-target-error="#pouet-submission-error" ` +
			`autofocus>Submit ID %d</button>`
		const did = `pouet-remote-indicator`
		const dclass = `htmx-indicator text-secondary pt-2`
		const sclass = `spinner-border spinner-border-sm`
		const text = `Fetching Download linked by Pouet...`

		button := fmt.Sprintf(format, prodID, prodID)
		button += `<div id="` + did + `" class="` + dclass + `" role="status">` +
			`  <span class="` + sclass + `"></span> <span>` + text + `</span></div>`
		button += fmt.Sprintf(`<div>%s</div>`, strings.Join(info, " "))
		return `<form class="d-grid">` + button + `</form>`
	}

	const code = http.StatusOK
	switch prod {
	case Demozoo:
		return c.HTML(code, demozoo())
	case Pouet:
		return c.HTML(code, pouet())
	default:
		return nil
	}
}

func (prod Prod) ValidFn(c *echo.Context, prodID int, useCache bool,
	fn func(string) (string, error),
) (bool, error) {
	const format = "valid fn: %w"
	if err := nils.Check(c); err != nil {
		return false, fmt.Errorf(format, err)
	}

	sid := strconv.Itoa(prodID)
	if ok := prodID > 0; !ok {
		s := "invalid id: " + sid
		return true, c.String(http.StatusNotAcceptable, s)
	}

	if useCache {
		if s, err := fn(sid); err == nil && s != "" {
			s := "Production " + sid + " is probably not suitable for Defacto2!<br>" +
				"Types: " + s
			return true, c.String(http.StatusOK, s)
		}
	}
	return false, nil
}

// Looks up the Demozoo production ID and returns a form button to submit
// the ID to the server for processing. If the Demozoo production ID is
// already in use, an error message is returned.
//
// This also acts as the string constructor for the summary of a successful lookup
// for the "Demozoo production or graphic" form.
func (prod Prod) lookupDemozoo(c *echo.Context, db *sql.DB, useCache bool) error {
	const format = "demozoo lookup htmx context: %w"
	if err := nils.Check(c, db); err != nil {
		return fmt.Errorf(format, err)
	}

	id := c.FormValue("demozoo-submission")
	prodID, err := strconv.Atoi(id)
	if err != nil {
		return c.String(http.StatusNotAcceptable,
			"The Demozoo production ID must be a numeric value, "+id)
	}

	ctx := c.Request().Context()
	deleted, key, err := model.OneDemozoo(ctx, db, int64(prodID))
	if err != nil {
		return c.String(http.StatusServiceUnavailable,
			"error, the database query failed")
	}

	if prodUsed := key != 0 && !deleted; prodUsed {
		const format = `This Demozoo production is already <a href="/f/%s">in use</a>.`
		html := fmt.Sprintf(format, helper.ObfuscateID(key))
		return c.HTML(http.StatusOK, html)
	}
	if prodUsed := key != 0 && deleted; prodUsed {
		return c.HTML(http.StatusOK, "This Demozoo production is already in use.")
	}

	product, err := ValidateDemozoo(c, prodID, useCache)
	if err != nil {
		return err
	}
	if invalid := product.ID < 1; invalid {
		return nil
	}

	info := []string{product.Title, "<br>"}
	if len(product.Authors) > 0 {
		info = append(info, "by")
		for _, author := range product.Authors {
			name := strings.TrimSpace(author.Name)
			if name == "" {
				continue
			}
			info = append(info, name)
		}
	}

	if prodRelDate := strings.TrimSpace(product.ReleaseDate); prodRelDate != "" {
		info = append(info, "on", prodRelDate)
	}

	if product.Platforms != nil {
		for _, val := range product.Platforms {
			name := strings.TrimSpace(val.Name)
			if name == "" {
				continue
			}
			info = append(info, "for", name)
		}
	}

	return prod.ButtonOK(c, prodID, info...)
}

// Fetches the multiple download_links values from the
// Pouet production API and attempts to download and save one of the
// linked files. If multiple links are found, the first link is used as
// they should all point to the same asset.
//
// Both the Pouet production ID param and the Defacto2 UUID query
// param values are required as params to fetch the production data and
// to save the file to the correct filename.
func (prod Prod) lookupPouet(c *echo.Context, db *sql.DB, useCache bool) error {
	const format = "htmx pouet lookup context: %w"
	if err := nils.Check(c, db); err != nil {
		return fmt.Errorf(format, err)
	}

	pouet := c.FormValue("pouet-submission")
	prodID, err := strconv.Atoi(pouet)
	if err != nil {
		return c.String(http.StatusNotAcceptable,
			"The Pouet production ID must be a numeric value, "+pouet)
	}

	ctx := c.Request().Context()
	deleted, key, err := model.OnePouet(ctx, db, int64(prodID))
	if err != nil {
		return c.String(http.StatusServiceUnavailable,
			"error, the database query failed")
	}

	if key != 0 && !deleted {
		const format = `This Pouet production is already <a href="/f/%s">in use</a>.`
		html := fmt.Sprintf(format, helper.ObfuscateID(key))
		return c.HTML(http.StatusOK, html)
	}
	if key != 0 && deleted {
		return c.HTML(http.StatusOK, "This Pouet production is already in use.")
	}

	resp, err := ValidatePouet(c, prodID, useCache)
	switch {
	case err != nil:
		return fmt.Errorf("PouetValid: %w", err)
	case resp.Prod.ID == "":
		return nil
	case !resp.Success:
		return c.String(http.StatusNotFound, "error, the Pouet production ID is not found")
	}

	product := resp.Prod
	if pid, err := strconv.Atoi(product.ID); err != nil {
		return c.String(http.StatusNotFound, "error, the Pouet production ID is invalid")
	} else if pid < 1 {
		return nil
	}

	info := []string{product.Title}
	if len(product.Groups) > 0 {
		info = append(info, "by")
		for _, val := range product.Groups {
			info = append(info, val.Name)
		}
	}
	if product.ReleaseDate != "" {
		info = append(info, "on", product.ReleaseDate)
	}

	platforms := strings.Split(product.Platforms.String(), ",")
	if len(platforms) > 0 {
		info = append(info, "for")
		for val := range slices.Values(platforms) {
			info = append(info, " ", strings.TrimSpace(val))
		}
	}

	return prod.ButtonOK(c, prodID, info...)
}

// ValidateDemozoo looks up the Demozoo production ID and confirms that the
// production is suitable for Defacto2. If a production is not suitable,
// an message is returned.
//
// A valid production requires at least one download link and must be a suitable type
// such as an intro, demo or cracktro for MS-DOS, Windows etc.
func ValidateDemozoo(c *echo.Context, prodID int, useCache bool) (demozoo.Production, error) {
	none := demozoo.Production{} //nolint:exhaustruct_v5
	const format = "htmx demozoo valid: %w"
	if err := nils.Check(c); err != nil {
		return none, fmt.Errorf(format, err)
	}

	exit, err := Demozoo.ValidFn(c, prodID, useCache, cache.DemozooProduction.Read)
	if exit || err != nil {
		return none, err
	}

	var prod demozoo.Production
	ctx := c.Request().Context()
	if code, err := prod.Get(ctx, prodID); err != nil {
		return none, c.String(code, err.Error())
	}

	plat, sect := prod.SuperType()
	const missing = -1
	if plat == missing || sect == missing {
		elems := []string{}
		for _, val := range prod.Platforms {
			elems = append(elems, val.Name)
		}
		for _, val := range prod.Types {
			elems = append(elems, val.Name)
		}

		key := strconv.Itoa(prodID)
		_ = cache.DemozooProduction.WriteNoExpire(key, strings.Join(elems, " - "))
		s := "Production " + key + " is probably not suitable for Defacto2!<br>Types: " +
			strings.Join(elems, " - ")

		return none, c.HTML(http.StatusOK, s)
	}

	for _, link := range prod.DownloadLinks {
		if link.URL == "" {
			continue
		}
		return prod, nil
	}

	return none, c.String(http.StatusOK,
		"This Demozoo production has no suitable download links.")
}

// ValidatePouet fetches the first usable download link from the Pouet API.
// The production ID is validated and the production is checked to see if it
// is suitable for Defacto2. If the production is not suitable, an empty
// production is returned with a htmx message.
func ValidatePouet(c *echo.Context, prodID int, useCache bool) (pouet.Response, error) {
	const msg = "htmx pouet valid context"
	const format = `Production %d is probably not suitable for Defacto2.`
	const helper = `<br>A production must an intro, demo or cracktro either for MsDos or Windows.`
	none := pouet.Response{} //nolint:exhaustruct_v5
	if err := nils.Check(c); err != nil {
		return none, fmt.Errorf("%s: %w", msg, err)
	}

	exit, err := Demozoo.ValidFn(c, prodID, useCache, cache.PouetProduction.Read)
	if exit || err != nil {
		return none, err
	}

	var prod pouet.Response
	ctx := c.Request().Context()
	if _, err := prod.Get(ctx, prodID); err != nil {
		return none, c.String(http.StatusInternalServerError, err.Error())
	}

	plat := pouet.PlatformsValid(prod.Prod.Platforms.String())
	types := false
	for _, prodType := range prod.Prod.Types {
		if prodType.Valid() {
			types = true
			break
		}
	}
	if ok := plat && types; !ok {
		key := strconv.Itoa(prodID)
		_ = cache.PouetProduction.WriteNoExpire(key, "invalid")
		return none, c.String(http.StatusOK, fmt.Sprintf(format, prodID)+helper)
	}

	if ok := prodLink(prod) != ""; !ok {
		const s = `This Pouet production has no suitable download links.`
		return none, c.String(http.StatusOK, s)
	}

	return prod, nil
}

func prodLink(prod pouet.Response) string {
	if s := prod.Prod.Download; s != "" {
		return s
	}

	unwanted := [...]string{"", "youtube", "sourceforge", "github"}
	for _, link := range prod.Prod.DownloadLinks {
		for _, substr := range unwanted {
			if strings.Contains(strings.ToLower(link.Link), substr) {
				continue
			}
		}

		return link.Link
	}

	return ""
}
