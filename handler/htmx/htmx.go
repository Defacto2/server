// Package htmx handles the routes and views for the AJAX responses using the htmx library.
package htmx

import (
	"database/sql"
	"errors"
	"fmt"
	"html"
	"html/template"
	"log/slog"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"slices"
	"strconv"
	"strings"
	"time"
	"unicode"
	"uuid"

	"github.com/Defacto2/helper"
	"github.com/Defacto2/server/handler/areacode"
	"github.com/Defacto2/server/handler/fulltext"
	"github.com/Defacto2/server/handler/releaser"
	"github.com/Defacto2/server/handler/releaser/lism"
	"github.com/Defacto2/server/internal/nils"
	"github.com/Defacto2/server/internal/postgres"
	"github.com/Defacto2/server/model"
	"github.com/Defacto2/server/model/fix"
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
			`<div class="alert alert-light border small mb-4">No results for '`+html.EscapeString(search)+`'.</div>`)
	}

	htm += `<div class="p-3 rounded border bg-body-tertiary mb-4">`
	for val := range slices.Values(query) {
		if val.AreaCode.Valid() {
			htm += val.AreaCode.HTML()
		}
		if len(val.Region) > 0 {
			for terr := range slices.Values(val.Region) {
				htm += terr.HTML()
			}
		}
	}
	htm += `</div>`
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
		"/files/one-offs",
		"/files/one-offs/a-z",
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
// It streams each route ping result as a multipart/mixed HTML part for htmx hx-multipart.
//
// The proto string should either be "http" or "https".
func Pings(c *echo.Context, proto string, port int) error {
	const format = "htmx pings context: %w"
	if err := nils.Check(c); err != nil {
		return fmt.Errorf(format, err)
	}

	ctx := c.Request().Context()
	pings := pings()

	w := c.Response()
	mw := multipart.NewWriter(w)
	w.Header().Set("Content-Type", "multipart/mixed; boundary="+mw.Boundary())
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)

	rc := http.NewResponseController(w)
	partHeader := textproto.MIMEHeader{"Content-Type": {"text/html; charset=utf-8"}}
	writePart := func(body string) error {
		part, err := mw.CreatePart(partHeader)
		if err != nil {
			return err
		}
		if _, err := part.Write([]byte(body)); err != nil {
			return err
		}
		_ = rc.Flush()
		return nil
	}

	count := 0
	for ping := range slices.Values(pings) {
		if ctx.Err() != nil {
			break
		}
		count++
		code, size, err := helper.LocalHostPing(ctx, ping, proto, port)
		if err != nil {
			if werr := writePart(fmt.Sprintf("<div>%s: %v</div>", html.EscapeString(ping), err)); werr != nil {
				return fmt.Errorf(format, werr)
			}
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

		const spanFmt = `<div><span class="%s">%d</span> %s <span class="text-secondary">%s</span></div>`
		if werr := writePart(fmt.Sprintf(spanFmt, class, code, ping, helper.ByteCount(size))); werr != nil {
			return fmt.Errorf(format, werr)
		}
	}

	summary := `<div><small>` + strconv.Itoa(count) + ` URLs were pinged</small></div>`
	if err := writePart(summary); err != nil {
		return fmt.Errorf(format, err)
	}
	if err := mw.Close(); err != nil {
		return fmt.Errorf(format, err)
	}
	_ = rc.Flush()

	return nil
}

// SearchByID is a handler for the /editor/search/id route.
func SearchByID(sl *slog.Logger, c *echo.Context, db *sql.DB) error {
	const format = "search by id context: %w"
	if err := nils.Check(sl, c, db); err != nil {
		return fmt.Errorf(format, err)
	}

	const maxResults = 50
	search := strings.TrimSpace(c.FormValue("htmx-search"))
	if search == "" {
		return c.HTML(http.StatusOK, `<!-- empty search query -->`)
	}

	ids, uuids, prefixes := ParseSearchIDs(search)
	if len(ids) == 0 && len(uuids) == 0 && len(prefixes) == 0 {
		return c.HTML(http.StatusOK, "No artifacts found.")
	}

	ctx := c.Request().Context()
	fs, err := model.OnlyUniqueIDsWithPrefixes(ctx, db, ids, uuids, prefixes)
	if err != nil {
		sl.Error("search by id query failed", slog.Any("error", err))
		return c.String(http.StatusServiceUnavailable,
			"the search by id query failed")
	}

	if len(fs) == 0 {
		return c.HTML(http.StatusOK, "No artifacts found.")
	}
	if len(fs) > maxResults {
		fs = fs[:maxResults]
	}

	err = c.Render(http.StatusOK, "searchids", map[string]any{
		maximum: maxResults,
		nme:     search,
		result:  fs,
	})
	if err != nil {
		sl.Error("cannot render htmx search by id template", slog.Any("error", err))
		return c.String(http.StatusInternalServerError, "cannot render the htmx search by id template")
	}

	return nil
}

// ParseSearchIDs splits space- or comma-separated tokens into numeric record IDs
// (including deobfuscated URL keys), exact non-nil UUIDs, and partial UUID prefixes (6-35 chars).
func ParseSearchIDs(search string) ([]int, []uuid.UUID, []string) {
	inputs := strings.FieldsFunc(search, func(r rune) bool {
		return unicode.IsSpace(r) || r == ',' || r == ';'
	})
	ids := make([]int, 0, len(inputs))
	uuids := make([]uuid.UUID, 0, len(inputs))
	prefixes := make([]string, 0, len(inputs))
	for _, input := range inputs {
		s := strings.ToLower(strings.TrimSpace(input))
		if s == "" {
			continue
		}
		if id, _ := strconv.Atoi(s); id > 0 {
			ids = append(ids, id)
			continue
		}
		if uid, err := uuid.Parse(s); err == nil && uid != uuid.Nil() {
			uuids = append(uuids, uid)
			continue
		}
		if id := helper.DeobfuscateID(s); id > 0 {
			ids = append(ids, id)
		}
		if isUUIDPrefix(s) {
			prefixes = append(prefixes, s)
		}
	}
	return ids, uuids, prefixes
}

// isUUIDPrefix reports whether s is a valid canonical UUID prefix of 6 to 35 lowercase hex/hyphen characters.
func isUUIDPrefix(s string) bool {
	const minPrefix, maxPrefix = 6, 35
	if len(s) < minPrefix || len(s) > maxPrefix {
		return false
	}
	for i := range len(s) {
		c := s[i]
		if i == 8 || i == 13 || i == 18 || i == 23 {
			if c != '-' {
				return false
			}
			continue
		}
		if (c < '0' || c > '9') && (c < 'a' || c > 'f') {
			return false
		}
	}
	return true
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

	// lookup exact match initialisms
	var r model.Releasers
	ctx := c.Request().Context()
	if err := r.Initialism(ctx, db, limit, lookup...); err != nil {
		sl.Error("task releaser match initialisms", slog.Any("error", err))
		return c.String(http.StatusServiceUnavailable,
			"the search exact query failed")
	}

	// lookup similar named releasers using a fresh slice (Lookup.releasers skips if len > 0)
	if len(r) < limit {
		var sim model.Releasers
		if err := sim.Similar(ctx, db, limit, lookup...); err != nil {
			sl.Error("task similar named releaser matches", slog.Any("error", err))
			return c.String(http.StatusServiceUnavailable,
				"the search similar query failed")
		}
		r = AppendUniqueReleasers(r, sim, limit)
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

// AppendUniqueReleasers merges additional releasers into dst without duplicate URIs,
// stopping once limit entries are reached.
func AppendUniqueReleasers(dst, add model.Releasers, limit int) model.Releasers {
	if len(dst) >= limit || len(add) == 0 {
		return dst
	}
	seen := make(map[string]struct{}, len(dst))
	for _, rel := range dst {
		seen[rel.Unique.URI] = struct{}{}
	}
	for _, rel := range add {
		if _, ok := seen[rel.Unique.URI]; ok {
			continue
		}
		seen[rel.Unique.URI] = struct{}{}
		dst = append(dst, rel)
		if len(dst) >= limit {
			break
		}
	}
	return dst
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
	ctx := c.Request().Context()
	matches, err := fix.SuggestReleasers(ctx, db, magazine, maxResults, lookups...)
	if err != nil {
		sl.Error("model similar releasers lookup failure",
			slog.String("lookups", strings.Join(lookups, ",")),
			slog.Bool("magazine_lookup", magazine), slog.Any("error", err))
		return c.String(http.StatusServiceUnavailable,
			"cannot connect to the database")
	}
	if len(matches) == 0 {
		return c.HTML(http.StatusOK, "")
	}

	r := make(model.Releasers, len(matches))
	for i, m := range matches {
		r[i] = &struct {
			Unique model.Releaser `boil:",bind"`
		}{
			Unique: model.Releaser{
				Name:  m.Releaser,
				URI:   helper.Slug(m.Releaser),
				Count: m.Count,
			},
		}
	}

	err = c.Render(http.StatusOK, "datalistreleasers", map[string]any{
		maximum: maxResults,
		nme:     slug,
		"query": strings.TrimSpace(input),
		result:  r,
	})
	if err != nil {
		return c.String(http.StatusInternalServerError,
			"cannot render the htmx datalist releases template")
	}

	return nil
}
