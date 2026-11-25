//nolint:exhaustive,wrapcheck
package app

// Package file context_releasers.go contains handlers for scene groups, releasers,
// BBS sites, FTP sites, and magazines.

import (
	"context"
	"database/sql"
	"fmt"
	"html/template"
	"io/fs"
	"log/slog"
	"net/http"
	"slices"
	"strconv"
	"strings"

	"github.com/Defacto2/helper"
	"github.com/Defacto2/server/handler/csdb"
	"github.com/Defacto2/server/handler/demozoo"
	"github.com/Defacto2/server/handler/janeway"
	"github.com/Defacto2/server/handler/releaser"
	"github.com/Defacto2/server/handler/releaser/lism"
	"github.com/Defacto2/server/handler/site"
	"github.com/Defacto2/server/handler/sixteen"
	"github.com/Defacto2/server/handler/tidbit"
	"github.com/Defacto2/server/internal/logs"
	"github.com/Defacto2/server/internal/nils"
	"github.com/Defacto2/server/model"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/labstack/echo/v5"
)

const (
	bbsx      = "bbs"
	ftp       = "ftp"
	mag       = "magazine"
	releaserx = "releaser"
)

// BBS is the handler for the BBS page ordered by the most files.
func BBS(sl *slog.Logger, c *echo.Context, db *sql.DB) error {
	ctx := c.Request().Context()
	return bbsHandler(ctx, sl, c, db, model.Prolific)
}

// BBSAZ is the handler for the BBS page ordered alphabetically.
func BBSAZ(sl *slog.Logger, c *echo.Context, db *sql.DB) error {
	ctx := c.Request().Context()
	return bbsHandler(ctx, sl, c, db, model.AlphabeticalTwo)
}

// BBSYear is the handler for the BBS page ordered by the year.
func BBSYear(sl *slog.Logger, c *echo.Context, db *sql.DB) error {
	ctx := c.Request().Context()
	return bbsHandler(ctx, sl, c, db, model.Oldest)
}

// bbsHandler is the handler for the BBS page.
func bbsHandler(ctx context.Context, sl *slog.Logger, c *echo.Context, db *sql.DB, orderBy model.OrderBy) error {
	const format = "bbs handler context: %w"
	if err := nils.Check(ctx, sl, c, db); err != nil {
		return fmt.Errorf(format, err)
	}

	const title = "BBS"
	const leadr = "Bulletin Board Systems were personal computers networked using the copper telephone network " +
		"and provided communication services, file hosting and exchanges."
	const logo = "Bulletin Board Systems"
	const key = "releasers"

	data := empty(c)
	data["noindex"] = true
	data["title"] = "Former " + title
	data["description"] = leadr
	data["logo"] = logo
	data["h1"] = title
	data["lead"] = leadr
	data["itemName"] = bbsx
	data[key] = model.Releasers{}
	data["stats"] = map[string]string{}

	r := model.Releasers{}
	if err := orderBy.BBS(ctx, db, &r); err != nil {
		return DatabaseErr(sl, c, bbsx, err)
	}

	data[key] = r
	tmpl := bbsx
	var order string
	switch orderBy {
	case model.Alphabetical, model.AlphabeticalTwo:
		s := logo + az
		data["logo"] = s
		data["title"] = title + az
		order = alpha
	case model.Prolific:
		data["noindex"] = false
		s := logo + ", by count"
		data["logo"] = s
		order = "by file artifact count"
	case model.Oldest:
		tmpl = "bbs-year"
		s := logo + byyear
		data["title"] = title + byyear
		data["logo"] = s
		order = byYear
	default:
		// Handle unknown order types
		order = alpha
	}

	data["stats"] = map[string]string{
		pubs:   fmt.Sprintf("%d boards", len(r)),
		ordrby: order,
	}

	if err := c.Render(http.StatusOK, tmpl, data); err != nil {
		return InternalErr(sl, c, bbsx, err)
	}
	return nil
}

// FTP is the handler for the FTP page.
func FTP(sl *slog.Logger, c *echo.Context, db *sql.DB) error {
	const format = "ftp context: %w"
	if err := nils.Check(sl, c, db); err != nil {
		return fmt.Errorf(format, err)
	}

	const title = "FTP"
	const descr = "FTP sites were Internet-based file exchange servers that would host and share Scene releases."
	data := empty(c)
	data["title"] = "Former " + title + " sites"
	data["description"] = descr
	data["logo"] = "FTP sites, A-Z"
	data["h1"] = title
	data["lead"] = descr

	// releaser.tmpl specific data items
	data["itemName"] = ftp
	const key = "releasers"
	data[key] = model.Releasers{}
	data["stats"] = map[string]string{}

	r := model.Releasers{}
	ctx := c.Request().Context()
	if err := r.FTP(ctx, db); err != nil {
		return DatabaseErr(sl, c, ftp, err)
	}
	data[key] = r

	data["stats"] = map[string]string{
		pubs:   fmt.Sprintf("%d sites", len(r)),
		ordrby: alpha,
	}

	if err := c.Render(http.StatusOK, ftp, data); err != nil {
		return InternalErr(sl, c, ftp, err)
	}
	return nil
}

// Magazine is the handler for the Magazine page ordered chronologically.
func Magazine(sl *slog.Logger, c *echo.Context, db *sql.DB) error {
	return magazines(c.Request().Context(), sl, c, db, true)
}

// MagazineAZ is the handler for the Magazine page ordered alphabetically.
func MagazineAZ(sl *slog.Logger, c *echo.Context, db *sql.DB) error {
	return magazines(c.Request().Context(), sl, c, db, false)
}

// magazines is the handler for the magazine page.
func magazines(ctx context.Context, sl *slog.Logger, c *echo.Context, db *sql.DB, chronological bool) error {
	const format = "magazines context: %w"
	if err := nils.Check(ctx, sl, c, db); err != nil {
		return fmt.Errorf(format, err)
	}

	const title = "Magazines"
	const descr = "Scene magazines are the newsletters, reports, " +
		"and publications on the activities of the subculture community."
	const issue = "issue"
	const key = "releasers"

	data := empty(c)
	data["title"] = title
	data["description"] = descr
	data["logo"] = title
	data["h1"] = title
	data["lead"] = descr
	data["itemName"] = issue
	data[key] = model.Releasers{}
	data["stats"] = map[string]string{}

	var order string
	r := model.Releasers{}
	render := mag

	if chronological {
		if err := r.Magazine(ctx, db); err != nil {
			return DatabaseErr(sl, c, mag, err)
		}
		s := title + byyear
		data["logo"] = s
		data["title"] = title + byyear
		order = byYear
	} else {
		render = mag + "-az"
		if err := r.MagazineAZ(ctx, db); err != nil {
			return DatabaseErr(sl, c, mag, err)
		}
		data["noindex"] = true
		s := title + az
		data["logo"] = s
		data["title"] = title + az
		order = alpha
	}

	data[key] = r
	data["stats"] = map[string]string{
		pubs:   fmt.Sprintf("%d publications", len(r)),
		ordrby: order,
	}

	if err := c.Render(http.StatusOK, render, data); err != nil {
		return InternalErr(sl, c, mag, err)
	}
	return nil
}

// Releasers is the handler for the releaser page ordered by the most files.
func Releasers(sl *slog.Logger, c *echo.Context, db *sql.DB) error {
	return releasers(c.Request().Context(), sl, c, db, model.Prolific)
}

// ReleasersAZ is the handler for the releaser page ordered alphabetically.
func ReleasersAZ(sl *slog.Logger, c *echo.Context, db *sql.DB) error {
	return releasers(c.Request().Context(), sl, c, db, model.AlphabeticalTwo)
}

// ReleasersYear is the handler for the releaser page ordered by year of the first release.
func ReleasersYear(sl *slog.Logger, c *echo.Context, db *sql.DB) error {
	return releasers(c.Request().Context(), sl, c, db, model.Oldest)
}

// releasers is the handler for the Releaser listing page.
func releasers(ctx context.Context, sl *slog.Logger, c *echo.Context, db *sql.DB, orderBy model.OrderBy) error {
	const format = "releaser context: %w"
	if err := nils.Check(ctx, sl, c, db); err != nil {
		return fmt.Errorf(format, err)
	}

	const title = "Releaser"
	const descr = "A linked list of the former Scene releasers and groups, that were collectives of people " +
		"who would work together and operate under a common brand."
	const leadr = "A releaser is a brand or a collective group of " +
		"sceners responsible for releasing or distributing products."
	const logo = "Former groups and releasers"
	const key = "releasers"

	data := empty(c)
	data["noindex"] = true
	data["title"] = title + "s and groups"
	data["description"] = descr
	data["logo"] = logo
	data["h1"] = title + "s"
	data["lead"] = leadr
	data["itemName"] = ffile
	data[key] = model.Releasers{}
	data["stats"] = map[string]string{}

	var r model.Releasers
	if err := orderBy.Limit(ctx, db, &r, 0, 0); err != nil {
		return DatabaseErr(sl, c, releaserx, err)
	}

	data[key] = r
	tmpl := releaserx
	var order string
	switch orderBy {
	case model.Alphabetical, model.AlphabeticalTwo:
		s := logo + az
		data["logo"] = s
		data["title"] = title + az
		order = alpha
	case model.Prolific:
		s := logo + ", by count"
		data["logo"] = s
		order = "by file artifact count"
	case model.Oldest:
		data["noindex"] = false
		tmpl = "releaser-year"
		s := logo + byyear
		data["logo"] = s
		data["title"] = title + byyear
		order = byYear
	default:
		// Handle unknown order types
		order = alpha
	}

	data["stats"] = map[string]string{
		pubs:   fmt.Sprintf("%d releasers and groups", len(r)),
		ordrby: order,
	}

	if err := c.Render(http.StatusOK, tmpl, data); err != nil {
		return InternalErr(sl, c, tmpl, err)
	}
	return nil
}

// Releaser is the handler for the list and preview of files credited to a releaser.
func Releaser(sl *slog.Logger, c *echo.Context, db *sql.DB, uri string, public fs.FS) error {
	const msg = "releasers context handler"
	const format = msg + ": %w"
	if err := nils.Check(sl, c, db); err != nil {
		return fmt.Errorf(format, err)
	}

	const code = http.StatusNotFound
	ctx := c.Request().Context()
	fs, err := model.ReleasersWhere(ctx, db, uri)
	if err != nil {
		logErr(sl, msg+" cannot lookup releaser", uri, code, err)
		return ReleaserErr(sl, c, uri)
	}
	if len(fs) == 0 {
		return ReleaserErr(sl, c, uri)
	}

	relname := releaser.Link(uri)
	data := emptyFiles(c)
	data["title"] = relname + " artifacts"
	data[canonical] = strings.Join([]string{"g", uri}, "/")
	data["h1"] = relname
	altnames := lism.String(lism.Path(uri))
	data["lead"] = altnames
	data["logo"] = relname
	data["demozoo"] = strconv.Itoa(int(demozoo.Find(uri)))
	data["sixteen"] = sixteen.Find(uri)
	data["csdb"] = csdb.Find(uri)
	data["janeway"] = janeway.Find(uri)
	data["website"] = site.Find(uri)
	data["tidbits"] = template.HTML(renderTidbits(sl, uri, public))

	if strings.HasSuffix(uri, "-bbs") {
		data["bbs"] = true
		relname = "the " + relname
	}
	if strings.HasSuffix(uri, "-ftp") {
		relname = "the " + relname
	}

	data["uploader-releaser-index"] = releaser.Index(uri)
	data[records] = fs
	data = releaserLead(uri, data)

	d, err := releaserSum(ctx, db, uri)
	if err != nil {
		logErr(sl, msg+" cannot lookup releaser stats", uri, code, err)
		return ReleaserErr(sl, c, uri)
	}

	data = releasersDesc(relname, altnames, d, data)

	if err = c.Render(http.StatusOK, artifacts, data); err != nil {
		return InternalErr(sl, c, "releasers page for, "+uri, err)
	}
	return nil
}

// releasersDesc appends stats to the description.
func releasersDesc(relname, altnames string, d map[string]string, data map[string]any,
) map[string]any {
	data["stats"] = d
	// append stats to the description
	dfiles := d["sum"]
	dyears := d[years]

	var desc string
	switch {
	case strings.EqualFold(relname, "independent"):
		desc = "The collection of " + dfiles + " unaffiliated artifacts"
	case strings.EqualFold(relname, "none"):
		desc = "The collection of the " + dfiles + " artifacts that were not authored by the Scene,"
		data["title"] = "Non-Scene artifacts"
		data["h1"] = "Non-Scene artifacts"
	case dfiles == "1":
		desc = "The single artifact for " + relname
	case dfiles != "":
		desc = "The collection of " + dfiles + " artifacts for " + relname
	default:
		desc = "The collection of artifacts for " + relname
	}
	if altnames != "" {
		desc += " (" + altnames + ")"
	}
	if dyears != "" {
		desc += " from " + dyears
	}
	data["description"] = desc + "."

	return data
}

func renderTidbits(sl *slog.Logger, uri string, public fs.FS) string {
	if sl == nil {
		sl = logs.Discard()
	}

	var htm strings.Builder

	tidbits := tidbit.Find(uri)
	slices.Sort(tidbits)
	for value := range slices.Values(tidbits) {
		s := strings.TrimSpace(value.String(sl, public))
		if s == "" {
			continue
		}
		htm.WriteString(`<div class="tidbit-entry"><div class="tidbit-prose">` + s + `</div>`)
		if rel := strings.TrimSpace(string(value.URL(uri))); rel != "" {
			htm.WriteString(`<div class="tidbit-related"><span class="tidbit-related-label">Related:</span> ` + rel + `</div>`)
		}
		htm.WriteString(`</div>`)
	}

	return htm.String()
}

func releaserLead(uri string, data map[string]any) map[string]any {
	switch uri {
	case "independent":
		data["lead"] = lism.String(lism.Path(uri)) +
			", independent releases are files with no group or releaser affiliation." +
			`<br><small class="fw-lighter">In the scene's early years,` +
			` releasing documents or software cracks under a personal alias or a` +
			` real-name attribution was commonplace.</small>`
	case "none":
		data["lead"] = "None, are files which were never intended for the scene." +
			`<br><small class="fw-lighter">These can include commercial or free software` +
			` applications, articles for the general public, and are often credited to a real name author.</small>`
	default:
		// placeholder to handle other releaser types
	}
	return data
}

// releaserSum is a helper function for Releaser that returns the statistics for the files page.
func releaserSum(ctx context.Context, exec boil.ContextExecutor, uri string) (map[string]string, error) {
	const format = "releaser sum %q: %w"
	if err := nils.Check(ctx, exec); err != nil {
		return nil, fmt.Errorf(format, uri, err)
	}

	m := model.Summary{
		SumBytes: sql.NullInt64{Int64: 0, Valid: false},
		SumCount: sql.NullInt64{Int64: 0, Valid: false},
		MinYear:  sql.NullInt16{Int16: 0, Valid: false},
		MaxYear:  sql.NullInt16{Int16: 0, Valid: false},
	}
	if err := m.ByReleaser(ctx, exec, uri); err != nil {
		return nil, fmt.Errorf(format, uri, err)
	}

	d := map[string]string{
		files: string(ByteFileS(ffile, m.SumCount.Int64, m.SumBytes.Int64)),
		years: helper.Years(m.MinYear.Int16, m.MaxYear.Int16),
		"sum": strconv.Itoa(int(m.SumCount.Int64)),
	}
	return d, nil
}
