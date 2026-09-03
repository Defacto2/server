// Package htmx handles the routes and views for the AJAX responses using the htmx library.
package htmx

import (
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
	"github.com/Defacto2/server/handler/areacode"
	"github.com/Defacto2/server/handler/cache"
	"github.com/Defacto2/server/handler/demozoo"
	"github.com/Defacto2/server/handler/fulltext"
	"github.com/Defacto2/server/handler/pouet"
	"github.com/Defacto2/server/handler/releaser"
	"github.com/Defacto2/server/handler/releaser/lism"
	"github.com/Defacto2/server/internal/dir"
	"github.com/Defacto2/server/internal/nils"
	"github.com/Defacto2/server/internal/postgres"
	"github.com/Defacto2/server/model"
	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
)

var (
	ErrIsDir      = errors.New("htmx: the file is a directory")
	ErrPath       = errors.New("htmx: the file path is invalid")
	ErrYMD        = errors.New("htmx: invalid ymd format")
	ErrYouTube    = errors.New("htmx: youtube watch video id needs to be empty or 11 characters")
	ErrFormRead   = errors.New("htmx form: parameters could not be read")
	ErrFormInsert = errors.New("htmx form: submission could not be inserted into the database")
	ErrFormUpdate = errors.New("htmx form: submission could not update a database record")
)

const (
	maximum = "maximum"
	nme     = "name"
	result  = "result"
)

// Areacodes is the handler for the /areacodes route.
func Areacodes(c *echo.Context) error {
	const format = "htmx area codes: %w"
	if err := nils.Check(c); err != nil {
		return fmt.Errorf(format, err)
	}

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
			`<small>No results for '`+html.EscapeString(search)+`'.</small><br>`)
	}

	for val := range slices.Values(query) {
		if val.AreaCode.Valid() {
			htm += val.AreaCode.HTML() + "<br>"
		}
		if len(val.Region) > 0 {
			for terr := range slices.Values(val.Region) {
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
func DemozooLookup(c *echo.Context, db *sql.DB, prodMode bool) error {
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

	prod, err := DemozooValid(c, prodMode, prodID)
	if err != nil {
		return err
	}
	if invalid := prod.ID < 1; invalid {
		return nil
	}

	info := []string{prod.Title, "<br>"}
	if len(prod.Authors) > 0 {
		info = append(info, "by")
		for _, author := range prod.Authors {
			name := strings.TrimSpace(author.Name)
			if name == "" {
				continue
			}
			info = append(info, name)
		}
	}

	if prodRelDate := strings.TrimSpace(prod.ReleaseDate); prodRelDate != "" {
		info = append(info, "on", prodRelDate)
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

	return c.HTML(http.StatusOK, demozooBtn(prodID, info...))
}

// Submit ID button saves the Demozoo production ID to the database and fetches the file.
// htmx.DemozooSubmit is the handler for the /demozoo/production put route,
// which uses htmx.submit, found in transfer.go, to insert the new file record into the database.
func demozooBtn(prodID int, info ...string) string {
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

// DemozooValid looks up the Demozoo production ID and confirms that the
// production is suitable for Defacto2. If a production is not suitable,
// an message is returned.
//
// A valid production requires at least one download link and must be a suitable type
// such as an intro, demo or cracktro for MS-DOS, Windows etc.
func DemozooValid(c *echo.Context, prodMode bool, prodID int) (demozoo.Production, error) {
	const format = "htmx demozoo valid: %w"
	none := demozoo.Production{} //nolint:exhaustruct
	if err := nils.Check(c); err != nil {
		return none, fmt.Errorf(format, err)
	}

	sid := strconv.Itoa(prodID)
	if invalid := prodID < 1; invalid {
		s := "invalid id: " + sid
		return none, c.String(http.StatusNotAcceptable, s)
	}

	if val, err := cache.DemozooProduction.Read(sid); err == nil {
		if prodMode && val != "" {
			s := "Production " + sid + " is probably not suitable for Defacto2!<br>Types: " + val
			return none, c.String(http.StatusOK, s)
		}
	}

	// Get the production data from Demozoo.
	// This func can be found in /internal/demozoo/demozoo.go
	ctx := c.Request().Context()
	var prod demozoo.Production
	if code, err := prod.Get(ctx, prodID); err != nil {
		return none, c.String(code, err.Error())
	}

	plat, sect := prod.SuperType()
	if plat == -1 || sect == -1 {
		elems := []string{}
		for _, val := range prod.Platforms {
			elems = append(elems, val.Name)
		}
		for _, val := range prod.Types {
			elems = append(elems, val.Name)
		}

		sid := strconv.Itoa(prodID)
		_ = cache.DemozooProduction.WriteNoExpire(sid, strings.Join(elems, " - "))
		s := "Production " + sid + " is probably not suitable for Defacto2!<br>Types: " +
			strings.Join(elems, " - ")

		return none, c.HTML(http.StatusOK, s)
	}

	for _, link := range prod.DownloadLinks {
		if link.URL == "" {
			continue
		}
		return prod, nil
	}

	return none,
		c.String(http.StatusOK,
			"This Demozoo production has no suitable download links.")
}

// DemozooSubmit is the handler for the /demozoo/production put route.
// This will attempt to insert a new file record into the database using
// the Demozoo production ID. If the Demozoo production ID is already in
// use, an error message is returned.
func DemozooSubmit(sl *slog.Logger, c *echo.Context, tx *sql.Tx, download dir.Directory) error {
	const format = "htmx demozoo submit context: %w"
	if err := nils.Check(sl, c, tx); err != nil {
		return fmt.Errorf(format, err)
	}

	return Demozoo.Submit(sl, c, tx, download)
}

// DBConnections is the handler for the database connections page.
func DBConnections(c *echo.Context, db *sql.DB) error {
	const format = "htmx db connections context: %w"
	if err := nils.Check(c, db); err != nil {
		return fmt.Errorf(format, err)
	}

	conns, maxConn, err := postgres.Connections(db)
	if err != nil {
		return c.String(http.StatusOK, err.Error())
	}

	currentTime := time.Now()

	const feedback = `%d of %d, <small>%s</small>`
	return c.String(http.StatusOK, fmt.Sprintf(feedback,
		conns, maxConn, currentTime.Format("15:04:05")))
}

// DeleteForever is a handler for the /delete/forever route.
// The recordKey is the numeric ID used by the record database table.
func DeleteForever(sl *slog.Logger, c *echo.Context, tx *sql.Tx, recordKey string) error {
	const msg = "htmx delete forever"
	if err := nils.Check(sl, c, tx); err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}

	key, err := strconv.ParseInt(recordKey, 10, 64)
	if err != nil {
		return c.String(http.StatusNotFound, err.Error())
	}

	ctx := c.Request().Context()
	if err = model.DeleteOne(ctx, tx, key); err != nil {
		sl.Error(msg+" database delete one transaction problem", slog.Any("error", err))
		return c.String(http.StatusServiceUnavailable,
			"cannot delete the record")
	}
	//
	// INFO: There is no need to delete any leftover file assets from the host system.
	// As any orphaned file assets will be deleted during the next cleanup job.
	//

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
//
// The proto string should either be "http" or "https".
func Pings(c *echo.Context, proto string, port int) error {
	const format = "htmx pings context: %w"
	if err := nils.Check(c); err != nil {
		return fmt.Errorf(format, err)
	}

	ctx := c.Request().Context()
	pings := pings()
	results := make([]string, 0, len(pings))
	for ping := range slices.Values(pings) {
		code, size, err := helper.LocalHostPing(ctx, ping, proto, port)
		if err != nil {
			results = append(results, fmt.Sprintf("%s: %v", ping, err))
			continue
		}

		var class string
		switch {
		case code == http.StatusOK:
			class = `text-success`
		case code == http.StatusNotFound:
			class = `text-danger-emphasis`
		case code >= http.StatusInternalServerError:
			class = `text-danger`
		default:
			class = `text-warning-emphasis`
		}

		const format = `<span class="%s">%d</span> %s <span class="text-secondary">%s</span>`
		spans := fmt.Sprintf(format, class, code, ping, helper.ByteCount(size))
		results = append(results, "<div>", spans, "</div>")
	}

	output := strings.Join(results, "")
	output += `<div><small>` + strconv.Itoa(len(pings)) + `  URLs were pinged</small></div>`

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
func PouetLookup(c *echo.Context, db *sql.DB) error {
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

	resp, err := PouetValid(c, prodID, false)
	switch {
	case err != nil:
		return fmt.Errorf("PouetValid: %w", err)
	case resp.Prod.ID == "":
		return nil
	case !resp.Success:
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

	return c.HTML(http.StatusOK, pouetBtn(prodID, info...))
}

func pouetBtn(prodID int, info ...string) string {
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

// PouetValid fetches the first usable download link from the Pouet API.
// The production ID is validated and the production is checked to see if it
// is suitable for Defacto2. If the production is not suitable, an empty
// production is returned with a htmx message.
func PouetValid(c *echo.Context, prodID int, useCache bool) (pouet.Response, error) {
	const msg = "htmx pouet valid context"
	const format = `Production %d is probably not suitable for Defacto2.`
	const helper = `<br>A production must an intro, demo or cracktro either for MsDos or Windows.`
	none := pouet.Response{} //nolint:exhaustruct
	if err := nils.Check(c); err != nil {
		return none, fmt.Errorf("%s: %w", msg, err)
	}

	sid := strconv.Itoa(prodID)
	if invalid := prodID < 1; invalid {
		return none,
			c.String(http.StatusNotAcceptable, "invalid id: "+sid)
	}

	if useCache {
		if s, err := cache.PouetProduction.Read(sid); err == nil {
			if s != "" {
				return none, c.String(http.StatusOK, fmt.Sprintf(format, prodID)+helper)
			}
		}
	}

	var prod pouet.Response
	ctx := c.Request().Context()
	if _, err := prod.Get(ctx, prodID); err != nil {
		return none, c.String(http.StatusInternalServerError, err.Error())
	}

	okPlat := pouet.PlatformsValid(prod.Prod.Platforms.String())
	okType := false
	for _, val := range prod.Prod.Types {
		if val.Valid() {
			okType = true
			break
		}
	}
	if valid := okPlat && okType; !valid {
		sid := strconv.Itoa(prodID)
		_ = cache.PouetProduction.WriteNoExpire(sid, "invalid")
		return none, c.String(http.StatusOK, fmt.Sprintf(format, prodID)+helper)
	}

	if valid := validation(prod) != ""; !valid {
		const s = `This Pouet production has no suitable download links.`
		return none, c.String(http.StatusOK, s)
	}

	return prod, nil
}

func validation(prod pouet.Response) string {
	if s := prod.Prod.Download; s != "" {
		return s
	}

	skips := [...]string{"", "youtube", "sourceforge", "github"}
	for _, link := range prod.Prod.DownloadLinks {
		for _, skip := range skips {
			if strings.Contains(strings.ToLower(link.Link), skip) {
				continue
			}
		}

		return link.Link
	}

	return ""
}

// PouetSubmit is the handler for the /pouet/production PUT route.
// This will attempt to insert a new file record into the database using
// the Pouet production ID. If the Pouet production ID is already in
// use, an error message is returned.
func PouetSubmit(sl *slog.Logger, c *echo.Context, tx *sql.Tx, download dir.Directory) error {
	const format = "htmx pouet submit context: %w"
	if err := nils.Check(sl, c, tx); err != nil {
		return fmt.Errorf(format, err)
	}

	return Pouet.Submit(sl, c, tx, download)
}

// SearchByID is a handler for the /editor/search/id route.
func SearchByID(sl *slog.Logger, c *echo.Context, db *sql.DB) error {
	const format = "search by id context: %w"
	if err := nils.Check(sl, c, db); err != nil {
		return fmt.Errorf(format, err)
	}

	const maxResults = 50
	ids := []int{}
	uuids := []uuid.UUID{}
	search := c.FormValue("htmx-search")
	inputs := strings.Split(search, " ")
	for input := range slices.Values(inputs) {
		s := strings.ToLower(strings.TrimSpace(input))
		if id, _ := strconv.Atoi(s); id > 0 {
			ids = append(ids, id)
			continue
		}
		if id := helper.DeobfuscateID(s); id > 0 {
			ids = append(ids, id)
			continue
		}
		if uid, err := uuid.Parse(s); err == nil {
			uuids = append(uuids, uid)
			continue
		}
	}

	ctx := c.Request().Context()
	fs, err := model.OnlyUniqueIDs(ctx, db, ids, uuids...)
	if err != nil {
		if sl != nil {
			sl.Error("something went wrong with the pouet lookup search", slog.Any("error", err))
		}
		return c.String(http.StatusServiceUnavailable,
			"the search by id query failed")
	}

	if len(fs) == 0 {
		return c.HTML(http.StatusOK, "No artifacts found.")
	}

	err = c.Render(http.StatusOK, "searchids", map[string]any{
		maximum: maxResults,
		nme:     search,
		result:  fs,
	})
	if err != nil {
		if sl != nil {
			sl.Error("could not render the pouet htmx search template", slog.Any("error", err))
		}
		return c.String(http.StatusInternalServerError, "cannot render the htmx search by id template")
	}

	return nil
}

// Alternatives returns a slice of possible matching alternative names,
// spellings, acronyms and initialisms for the s string.
func Alternatives(s string) []string {
	if s == "" {
		return []string{}
	}

	const minChars = 4
	lookups := []string{s}
	// examples of key and values:
	// "tristar-ampersand-red-sector-inc": {"TRSi", "TRS", "Tristar"},
	key := ""
	for path, initialisms := range lism.Copy() {
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
func SearchReleaser(sl *slog.Logger, c *echo.Context, db *sql.DB, ft *fulltext.Tidbits) error {
	const format = "htmx search releaser context: %w"
	if err := nils.Check(sl, c, db, ft); err != nil {
		return fmt.Errorf(format, err)
	}

	const limit = 14
	input := c.FormValue("htmx-search")
	name := helper.TrimRoundBracket(input)
	name = releaser.Clean(name) // required to stop 503 errors with invalid characters
	if name == "" {
		const comment = `<!-- empty search query -->`
		return c.HTML(http.StatusOK, comment)
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
	ctx := c.Request().Context()
	if err := r.Initialism(ctx, db, limit, lookup...); err != nil {
		sl.Error("task releaser match initialisms", slog.Any("error", err))
		return c.String(http.StatusServiceUnavailable,
			"the search exact query failed")
	}

	// lookup similar named releasers
	remaining := limit - len(r)
	if remaining > 0 {
		if err := r.Similar(ctx, db, remaining, lookup...); err != nil {
			sl.Error("task similar named releaser matches", slog.Any("error", err))
			return c.String(http.StatusServiceUnavailable,
				"the search similar query failed")
		}
	}
	// lookup markdown
	const maxResults = 50
	tidbits := ft.Search(input, maxResults)

	// no results
	if len(r) == 0 && len(tidbits) == 0 {
		return c.HTML(http.StatusOK, "No initialisms or releasers found.")
	}

	err := c.Render(http.StatusOK, "searchreleasers", map[string]any{
		"fulltext": tidbits,
		"input":    input,
		maximum:    limit,
		nme:        name,
		result:     r,
	})
	if err != nil {
		return c.String(http.StatusInternalServerError,
			"cannot render the htmx search releases template")
	}

	return nil
}

// DataListReleasers is a handler for the /datalist/releasers route.
func DataListReleasers(sl *slog.Logger, c *echo.Context, db *sql.DB, input string) error {
	return datalist(sl, c, db, input, false)
}

// DataListMagazines is a handler for the /datalist/magazines route.
func DataListMagazines(sl *slog.Logger, c *echo.Context, db *sql.DB, input string) error {
	return datalist(sl, c, db, input, true)
}

// datalist is a shared handler for the /datalist/releasers and /datalist/magazines routes.
func datalist(sl *slog.Logger, c *echo.Context, db *sql.DB, input string, magazine bool) error {
	const format = "htmx datalist context: %w"
	if err := nils.Check(sl, c, db); err != nil {
		return fmt.Errorf(format, err)
	}

	slug := helper.Slug(helper.TrimRoundBracket(input))
	if slug == "" {
		return c.HTML(http.StatusOK, "")
	}

	lookups := []string{releaser.Cell(input)}
	if inits := lism.Find(slug); len(inits) > 0 {
		for uri := range slices.Values(inits) {
			val := releaser.Humanize(string(uri))
			lookups = append(lookups, val)
		}
	}
	lookups = append(lookups, slug) // slug is the last lookup and must be present.

	const maxResults = 14
	var r model.Releasers
	var err error
	ctx := c.Request().Context()
	if magazine {
		err = r.SimilarMagazine(ctx, db, maxResults, lookups...)
	} else {
		err = r.Similar(ctx, db, maxResults, lookups...)
	}
	if err != nil {
		sl.Error("model similar releasers lookup failure",
			slog.String("lookups", strings.Join(lookups, ",")),
			slog.Bool("magazine_lookup", magazine), slog.Any("error", err))
		return c.String(http.StatusServiceUnavailable,
			"cannot connect to the database")
	}
	if len(r) == 0 {
		return c.HTML(http.StatusOK, "")
	}

	err = c.Render(http.StatusOK, "datalistreleasers", map[string]any{
		maximum: maxResults,
		nme:     slug,
		result:  r,
	})
	if err != nil {
		return c.String(http.StatusInternalServerError,
			"cannot render the htmx datalist releases template")
	}

	return nil
}
