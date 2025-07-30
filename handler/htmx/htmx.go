// Package htmx handles the routes and views for the AJAX responses using the htmx library.
package htmx

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"html"
	"html/template"
	"log/slog"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/Defacto2/helper"
	"github.com/Defacto2/releaser"
	"github.com/Defacto2/releaser/initialism"
	"github.com/Defacto2/server/handler/areacode"
	"github.com/Defacto2/server/handler/cache"
	"github.com/Defacto2/server/handler/demozoo"
	"github.com/Defacto2/server/handler/pouet"
	"github.com/Defacto2/server/internal/dir"
	"github.com/Defacto2/server/internal/postgres"
	"github.com/Defacto2/server/model"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

var (
	ErrDB     = errors.New("database connection is nil")
	ErrFormat = errors.New("invalid format")
	ErrKey    = errors.New("numeric record key is invalid")
)

func htmxpanic(c echo.Context, db *sql.DB, sl *slog.Logger) error {
	if c == nil {
		return ErrNoEcho
	}
	if db == nil {
		return ErrNoDB
	}
	if sl == nil {
		return ErrNoSlog
	}
	return nil
}

func dbpanic(c echo.Context, db *sql.DB) error {
	if c == nil {
		return ErrNoEcho
	}
	if db == nil {
		return ErrNoDB
	}
	return nil
}

// Areacodes is the handler for the /areacodes route.
func Areacodes(c echo.Context) error {
	htm := template.HTML("")
	search := c.FormValue("htmx-search")
	search = strings.TrimSpace(search)
	if search == "" {
		return c.HTML(http.StatusOK, "")
	}
	searches := strings.Split(search, ",")
	query := areacode.Queries(searches...)
	if len(query) == 0 {
		return c.HTML(http.StatusOK,
			"<small>No results for '"+html.EscapeString(search)+"'.</small><br>")
	}
	for val := range slices.Values(query) {
		if val.AreaCode.Valid() {
			htm += val.AreaCode.HTML() + "<br>"
		}
		if len(val.Terr) > 0 {
			for terr := range slices.Values(val.Terr) {
				htm += terr.HTML() + "<br>"
			}
		}
	}
	htm += "<hr>"
	return c.HTML(http.StatusOK, string(htm))
}

// DemozooLookup is the handler for the /demozoo/production route.
// This looks up the Demozoo production ID and returns a form button to submit
// the ID to the server for processing. If the Demozoo production ID is
// already in use, an error message is returned.
//
// This also acts as the string constructor for the summary of a successful lookup
// for the "Demozoo production or graphic" form.
func DemozooLookup(c echo.Context, db *sql.DB, prodMode bool) error {
	zoo := c.FormValue("demozoo-submission")
	id, err := strconv.Atoi(zoo)
	if err != nil {
		return c.String(http.StatusNotAcceptable,
			"The Demozoo production ID must be a numeric value, "+zoo)
	}
	ctx := context.Background()
	deleted, key, err := model.OneDemozoo(ctx, db, int64(id))
	if err != nil {
		return c.String(http.StatusServiceUnavailable,
			"error, the database query failed")
	}
	if prodInUse := key != 0 && !deleted; prodInUse {
		html := fmt.Sprintf("This Demozoo production is already <a href=\"/f/%s\">in use</a>.", helper.ObfuscateID(key))
		return c.HTML(http.StatusOK, html)
	}
	if prodInUse := key != 0 && deleted; prodInUse {
		return c.HTML(http.StatusOK, "This Demozoo production is already in use.")
	}
	prod, err := DemozooValid(c, prodMode, id)
	if err != nil {
		return err
	}
	if invalid := prod.ID < 1; invalid {
		return nil
	}
	info := []string{prod.Title, "<br>"}
	if len(prod.Authors) > 0 {
		info = append(info, "by")
		for _, val := range prod.Authors {
			name := strings.TrimSpace(val.Name)
			if name == "" {
				continue
			}
			info = append(info, name)
		}
	}
	if relDate := strings.TrimSpace(prod.ReleaseDate); relDate != "" {
		info = append(info, "on", relDate)
	}
	if prod.Platforms != nil {
		for _, val := range prod.Platforms {
			name := strings.TrimSpace(val.Name)
			if name == "" {
				continue
			}
			info = append(info, "for", name)
		}
	}
	return c.HTML(http.StatusOK, btn(id, info...))
}

// Submit ID button saves the Demozoo production ID to the database and fetches the file.
// htmx.DemozooSubmit is the handler for the /demozoo/production put route,
// which uses htmx.submit, found in transfer.go, to insert the new file record into the database.
func btn(id int, info ...string) string {
	htm := `<div class="d-grid gap-2">`
	htm += fmt.Sprintf(`<button type="button" class="btn btn-outline-success" `+
		`hx-put="/demozoo/production/%d" `+
		`hx-indicator="#demozoo-indicator" `+
		`hx-target="#demozoo-submission-results" hx-trigger="click once delay:500ms" `+
		`hx-target-error="#demozoo-submission-error" `+
		`autofocus>Submit ID %d</button>`, id, id)
	htm += `</div>`
	htm += fmt.Sprintf(`<p class="mt-3">%s</p>`, strings.Join(info, " "))
	return htm
}

// DemozooValid looks up the Demozoo production ID and confirms that the
// production is suitable for Defacto2. If a production is not suitable,
// an message is returned.
//
// A valid production requires at least one download link and must be a suitable type
// such as an intro, demo or cracktro for MS-DOS, Windows etc.
func DemozooValid(c echo.Context, prodMode bool, id int) (demozoo.Production, error) {
	const msg = "htmx demozoo valid"
	none := demozoo.Production{}
	if c == nil {
		return none, fmt.Errorf("%s: %w", msg, ErrNoEcho)
	}
	if invalid := id < 1; invalid {
		return none,
			c.String(http.StatusNotAcceptable, fmt.Sprintf("invalid id: %d", id))
	}
	sid := strconv.Itoa(id)
	if s, err := cache.DemozooProduction.Read(sid); err == nil {
		if prodMode && s != "" {
			return none,
				c.String(http.StatusOK,
					fmt.Sprintf("Production %d is probably not suitable for Defacto2!<br>Types: %s", id, s))
		}
	}
	var prod demozoo.Production
	// Get the production data from Demozoo.
	// This func can be found in /internal/demozoo/demozoo.go
	if code, err := prod.Get(id); err != nil {
		return none, c.String(code, err.Error())
	}
	plat, sect := prod.SuperType()
	if plat == -1 || sect == -1 {
		s := []string{}
		for _, val := range prod.Platforms {
			s = append(s, val.Name)
		}
		for _, val := range prod.Types {
			s = append(s, val.Name)
		}
		sid := strconv.Itoa(id)
		_ = cache.DemozooProduction.WriteNoExpire(sid, strings.Join(s, " - "))
		return none, c.HTML(http.StatusOK,
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
		return none,
			c.String(http.StatusOK,
				"This Demozoo production has no suitable download links.")
	}
	return prod, nil
}

// DemozooSubmit is the handler for the /demozoo/production put route.
// This will attempt to insert a new file record into the database using
// the Demozoo production ID. If the Demozoo production ID is already in
// use, an error message is returned.
func DemozooSubmit(c echo.Context, db *sql.DB, sl *slog.Logger, download dir.Directory) error {
	const msg = "htmx demozoo submit"
	if err := htmxpanic(c, db, sl); err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}
	return Demozoo.Submit(c, db, sl, download)
}

// DBConnections is the handler for the database connections page.
func DBConnections(c echo.Context, db *sql.DB) error {
	const msg = "htmx db connections"
	if err := dbpanic(c, db); err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}
	conns, maxConn, err := postgres.Connections(db)
	if err != nil {
		return c.String(http.StatusOK, err.Error())
	}
	currentTime := time.Now()
	return c.String(http.StatusOK, fmt.Sprintf("%d of %d, <small>%s</small>",
		conns, maxConn, currentTime.Format("15:04:05")))
}

// DeleteForever is a handler for the /delete/forever route.
func DeleteForever(c echo.Context, db *sql.DB, sl *slog.Logger, id string) error {
	const msg = "htmx delete forever"
	if err := htmxpanic(c, db, sl); err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}
	key, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return c.String(http.StatusNotFound, err.Error())
	}
	ctx := context.Background()
	tx, err := db.Begin()
	if err != nil {
		sl.Error(msg, slog.String("database", "could not start transaction"), slog.Any("error", err))
		return c.String(http.StatusServiceUnavailable,
			"cannot begin a transaction")
	}
	if err = model.DeleteOne(ctx, tx, key); err != nil {
		defer func() {
			if err := tx.Rollback(); err != nil && sl != nil {
				sl.Error(msg,
					slog.String("database", "delete one transaction rollback problem"),
					slog.Any("error", err))
			}
		}()
		sl.Error(msg, slog.String("database", "delete one transaction problem"),
			slog.Any("error", err))
		return c.String(http.StatusServiceUnavailable,
			"cannot delete the record")
	}
	//
	// There is no need to delete any file assets from the file system.
	// As the file assets will be deleted by the next cleanup job.
	//
	if err = tx.Commit(); err != nil {
		if sl != nil {
			sl.Error(msg,
				slog.String("database", "transaction commit failed"),
				slog.Any("error", err))
		}
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
	const msg = "htmx pings"
	if c == nil {
		return fmt.Errorf("%s, %w", msg, ErrNoEcho)
	}
	pings := pings()
	results := make([]string, 0, len(pings))
	for ping := range slices.Values(pings) {
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

// PouetLookup fetches the multiple download_links values from the
// Pouet production API and attempts to download and save one of the
// linked files. If multiple links are found, the first link is used as
// they should all point to the same asset.
//
// Both the Pouet production ID param and the Defacto2 UUID query
// param values are required as params to fetch the production data and
// to save the file to the correct filename.
func PouetLookup(c echo.Context, db *sql.DB) error {
	const msg = "htmx pouet lookup"
	if err := dbpanic(c, db); err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}
	pouet := c.FormValue("pouet-submission")
	id, err := strconv.Atoi(pouet)
	if err != nil {
		return c.String(http.StatusNotAcceptable,
			"The Pouet production ID must be a numeric value, "+pouet)
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
	resp, err := PouetValid(c, id, false)
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
		for _, val := range prod.Groups {
			info = append(info, val.Name)
		}
	}
	if prod.ReleaseDate != "" {
		info = append(info, "on", prod.ReleaseDate)
	}
	platforms := strings.Split(prod.Platforms.String(), ",")
	if len(platforms) > 0 {
		info = append(info, "for")
		for val := range slices.Values(platforms) {
			info = append(info, " ", strings.TrimSpace(val))
		}
	}
	return c.HTML(http.StatusOK, htmler(id, info...))
}

func htmler(id int, info ...string) string {
	s := `<div class="d-grid gap-2">`
	s += fmt.Sprintf(`<button type="button" class="btn btn-outline-success" `+
		`hx-put="/pouet/production/%d" `+
		`hx-indicator="#pouet-indicator" `+
		`hx-target="#pouet-submission-results" hx-trigger="click once delay:500ms" `+
		`hx-target-error="#pouet-submission-error" `+
		`autofocus>Submit ID %d</button>`, id, id)
	s += `</div>`
	s += fmt.Sprintf(`<p class="mt-3">%s</p>`, strings.Join(info, " "))
	return s
}

// PouetValid fetches the first usable download link from the Pouet API.
// The production ID is validated and the production is checked to see if it
// is suitable for Defacto2. If the production is not suitable, an empty
// production is returned with a htmx message.
func PouetValid(c echo.Context, id int, useCache bool) (pouet.Response, error) {
	const msg = "htmx pouet valid"
	none := pouet.Response{}
	if c == nil {
		return none, fmt.Errorf("%s: %w", msg, ErrNoEcho)
	}
	if invalid := id < 1; invalid {
		return none,
			c.String(http.StatusNotAcceptable, fmt.Sprintf("invalid id: %d", id))
	}
	if useCache {
		sid := strconv.Itoa(id)
		if s, err := cache.PouetProduction.Read(sid); err == nil {
			if s != "" {
				return none,
					c.String(http.StatusOK,
						fmt.Sprintf("Production %d is probably not suitable for Defacto2.", id)+
							"<br>A production must an intro, demo or cracktro either for MsDos or Windows.")
			}
		}
	}
	var prod pouet.Response
	if _, err := prod.Get(id); err != nil {
		return none, c.String(http.StatusInternalServerError, err.Error())
	}
	platOkay := pouet.PlatformsValid(prod.Prod.Platforms.String())
	typeOkay := false
	for _, val := range prod.Prod.Types {
		if val.Valid() {
			typeOkay = true
			break
		}
	}
	if valid := platOkay && typeOkay; !valid {
		sid := strconv.Itoa(id)
		_ = cache.PouetProduction.WriteNoExpire(sid, "invalid")
		return none, c.HTML(http.StatusOK,
			fmt.Sprintf("Production %d is probably not suitable for Defacto2.", id)+
				"<br>A production must an intro, demo or cracktro either for MsDos or Windows.")
	}
	if valid := validation(prod); valid == "" {
		return none,
			c.String(http.StatusOK, "This Pouet production has no suitable download links.")
	}
	return prod, nil
}

func validation(prod pouet.Response) string {
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
	return valid
}

// PouetSubmit is the handler for the /pouet/production PUT route.
// This will attempt to insert a new file record into the database using
// the Pouet production ID. If the Pouet production ID is already in
// use, an error message is returned.
func PouetSubmit(c echo.Context, db *sql.DB, sl *slog.Logger, download dir.Directory) error {
	const msg = "htmx pouet submit"
	if err := htmxpanic(c, db, sl); err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}
	return Pouet.Submit(c, db, sl, download)
}

// SearchByID is a handler for the /editor/search/id route.
func SearchByID(c echo.Context, db *sql.DB, sl *slog.Logger) error {
	const msg = "search by id"
	if err := htmxpanic(c, db, sl); err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}
	const maxResults = 50
	ctx := context.Background()
	ids := []int{}
	uuids := []uuid.UUID{}
	search := c.FormValue("htmx-search")
	inputs := strings.Split(search, " ")
	for input := range slices.Values(inputs) {
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
		if sl != nil {
			sl.Error(msg, slog.String("lookup", "something went wrong with the search"), slog.Any("error", err))
		}
		return c.String(http.StatusServiceUnavailable,
			"the search query failed")
	}

	if len(fs) == 0 {
		return c.HTML(http.StatusOK, "No artifacts found.")
	}
	err = c.Render(http.StatusOK, "searchids", map[string]any{
		"maximum": maxResults,
		"name":    search,
		"result":  fs,
	})
	if err != nil {
		if sl != nil {
			sl.Error(msg, slog.String("lookup", "could not render the htmx search template"), slog.Any("error", err))
		}
		return c.String(http.StatusInternalServerError,
			"cannot render the htmx search by id template")
	}
	return nil
}

// Alternatives returns a slice of possible matching alternative names,
// spellings, acronyms and initialisms for the s string.
func Alternatives(s string) []string {
	const minChars = 4
	lookups := []string{s}
	// examples of key and values:
	// "tristar-ampersand-red-sector-inc": {"TRSi", "TRS", "Tristar"},
	key := ""
	for path, initialisms := range *initialism.Initialisms() {
		key = releaser.Index(string(path))
		if key == "" {
			continue
		}
		// value is usually an initialism, however it can be alternative spellings etc.
		for value := range slices.Values(initialisms) {
			//	if s and value are an exact match, use the key as a lookup
			//	ie: s = "trs" and value is "TRS" and key is "tristar-ampersand-red-sector-inc"
			if strings.EqualFold(value, s) {
				lookups = append(lookups, key)
				continue
			}
			if len(s) < minChars {
				continue
			}
			if strings.Contains(strings.ToLower(value), strings.ToLower(s)) {
				lookups = append(lookups, key, value)
			}
		}
	}
	t := releaser.Humanize(s)
	if t != "" && !strings.EqualFold(s, t) {
		lookups = append(lookups, t)
	}
	return lookups
}

// SearchReleaser is a handler for the /search/releaser route.
func SearchReleaser(c echo.Context, db *sql.DB, sl *slog.Logger) error {
	const msg = "htmx search releaser"
	if err := htmxpanic(c, db, sl); err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}
	const limit = 14
	ctx := context.Background()
	input := c.FormValue("htmx-search")
	name := helper.TrimRoundBraket(input)
	name = releaser.Clean(name) // required to stop 503 errors with invalid characters
	if name == "" {
		return c.HTML(http.StatusOK, "<!-- empty search query -->")
	}
	// Obtain a list of alternative lookups and remove any possible duplicates.
	lookup := Alternatives(name)
	slices.Sort(lookup)
	lookup = slices.Compact(lookup)
	// matchZeroOrMore is an SQL "LIKE" expression, to return zero (exact match) or more matches.
	// see: https://www.postgresql.org/docs/current/functions-matching.html#FUNCTIONS-LIKE
	const matchZeroOrMore = "%"
	lookup = slices.Insert(lookup, 0, name+matchZeroOrMore)
	// lookup exact match initialisms
	var r model.Releasers
	if err := r.Initialism(ctx, db, limit, lookup...); err != nil {
		sl.Error(msg, slog.String("task", "releaser match initialisms"),
			slog.Any("error", err))
		return c.String(http.StatusServiceUnavailable,
			"the search query failed")
	}
	// lookup similar named releasers
	if len(r) == 0 {
		if err := r.Similar(ctx, db, limit, lookup...); err != nil {
			sl.Error(msg, slog.String("task", "similar named releaser matches"),
				slog.Any("error", err))
			return c.String(http.StatusServiceUnavailable,
				"the search query failed")
		}
	}
	// no results
	if len(r) == 0 {
		return c.HTML(http.StatusOK, "No initialisms or releasers found.")
	}
	err := c.Render(http.StatusOK, "searchreleasers", map[string]any{
		"maximum": limit,
		"name":    name,
		"result":  r,
	})
	if err != nil {
		return c.String(http.StatusInternalServerError,
			"cannot render the htmx search releases template")
	}
	return nil
}

func dlpanic(c echo.Context, db *sql.DB, sl *slog.Logger) error {
	if c == nil {
		return ErrNoEcho
	}
	if db == nil {
		return ErrNoDB
	}
	if sl == nil {
		return ErrNoSlog
	}
	return nil
}

// DataListReleasers is a handler for the /datalist/releasers route.
func DataListReleasers(c echo.Context, db *sql.DB, sl *slog.Logger, input string) error {
	const msg = "htmx datalist releasers"
	if err := dlpanic(c, db, sl); err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}
	return datalist(c, db, sl, input, false)
}

// DataListMagazines is a handler for the /datalist/magazines route.
func DataListMagazines(c echo.Context, db *sql.DB, sl *slog.Logger, input string) error {
	const msg = "htmx datalist magazines"
	if err := dlpanic(c, db, sl); err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}
	return datalist(c, db, sl, input, true)
}

// datalist is a shared handler for the /datalist/releasers and /datalist/magazines routes.
func datalist(c echo.Context, db *sql.DB, sl *slog.Logger, input string, magazine bool) error {
	const msg = "htmx datalist"
	if err := dlpanic(c, db, sl); err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}
	const maxResults = 14
	ctx := context.Background()
	slug := helper.Slug(helper.TrimRoundBraket(input))
	if slug == "" {
		return c.HTML(http.StatusOK, "")
	}
	lookups := []string{releaser.Cell(input)}
	if inits := initialism.Match(slug); len(inits) > 0 {
		for uri := range slices.Values(inits) {
			val := releaser.Humanize(string(uri))
			lookups = append(lookups, val)
		}
	}
	lookups = append(lookups, slug) // slug is the last lookup and must be present.
	var r model.Releasers
	var err error
	if magazine {
		err = r.SimilarMagazine(ctx, db, maxResults, lookups...)
	} else {
		err = r.Similar(ctx, db, maxResults, lookups...)
	}
	if err != nil {
		sl.Error(msg, slog.String("model", "similar releasers lookup failure"),
			slog.String("lookups", strings.Join(lookups, ",")),
			slog.Bool("magazine lookup", magazine),
			slog.Any("error", err))
		return c.String(http.StatusServiceUnavailable,
			"cannot connect to the database")
	}
	if len(r) == 0 {
		return c.HTML(http.StatusOK, "")
	}
	err = c.Render(http.StatusOK, "datalistreleasers", map[string]any{
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
