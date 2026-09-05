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
	"github.com/Defacto2/server/handler/fulltext"
	"github.com/Defacto2/server/handler/releaser"
	"github.com/Defacto2/server/handler/releaser/lism"
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

// DBConnections is the handler for the database connections page.
func DBConnections(c *echo.Context, db *sql.DB) error {
	const format = "htmx db connections context: %w"
	if err := nils.Check(c, db); err != nil {
		return fmt.Errorf(format, err)
	}

	ctx := c.Request().Context()
	conns, maxConn, err := postgres.Connections(ctx, db)
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
