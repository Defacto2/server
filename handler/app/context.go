//nolint:exhaustruct_v5,wrapcheck
package app

// Package file context.go contains the typed PageData template wrapper,
// shared defaults, and pagination types used by context_*.go handlers.

import (
	"fmt"
	"html/template"
	"log/slog"
	"net/http"

	"github.com/Defacto2/server/handler/sess"
	"github.com/Defacto2/server/handler/site"
	"github.com/Defacto2/server/internal/nils"
	"github.com/labstack/echo/v5"
)

// Pagination holds the page numbers and base URL for paginated artifact listings.
type Pagination struct {
	BaseURL  string // BaseURL is the base URL for the pagination links.
	CurrPage int    // CurrPage is the current page number.
	SumPages int    // SumPages is the total number of pages.
	PrevPage int    // PrevPage is the previous page number.
	NextPage int    // NextPage is the next page number.
	TwoBelow int    // TwoBelow is the page number two below the current page.
	TwoAfter int    // TwoAfter is the page number two after the current page.
}

// PageData is a typed wrapper around map[string]any for layout template data.
// Because its underlying type is map[string]any, existing templates continue to
// read keys directly while Go handlers benefit from compile-time key and type safety.
type PageData map[string]any

func (p PageData) SetTitle(s string)           { p["title"] = s }
func (p PageData) SetH1(s string)              { p["h1"] = s }
func (p PageData) SetSubheading(s string)      { p["subheading"] = s }
func (p PageData) SetLead(s string)            { p["lead"] = s }
func (p PageData) SetLeadHTML(h template.HTML) { p["lead"] = h }
func (p PageData) SetLogo(s string)            { p["logo"] = s }
func (p PageData) SetDescription(s string)     { p["description"] = s }
func (p PageData) SetCanonical(s string)       { p[canonical] = s }
func (p PageData) SetCarousel(s string)        { p["carousel"] = s }
func (p PageData) SetOGTitle(s string)         { p["ogtitle"] = s }
func (p PageData) SetNoIndex(v bool)           { p["noindex"] = v }
func (p PageData) SetDatabaseErr(v bool)       { p["databaseErr"] = v }
func (p PageData) SetPagination(pg Pagination) { p["Pagination"] = pg }
func (p PageData) SetStats(s map[string]string) { p["stats"] = s }
func (p PageData) SetRecords(r any)            { p[records] = r }

// SetHeader sets the core page metadata fields in a single call.
func (p PageData) SetHeader(title, h1, lead, logo, description string) {
	p.SetTitle(title)
	p.SetH1(h1)
	p.SetLead(lead)
	p.SetLogo(logo)
	p.SetDescription(description)
}

const (
	az        = ", a-z"
	byyear    = " by year"
	artifact  = "artifact"
	artifacts = "artifacts"
	ffile     = "file"
	status    = "status"
)

// empty returns a PageData populated with default values for layout.tmpl.
//
// All keys are optional except for "description" and "title":
//   - "description" is used by the meta description element.
//   - "title" is used by the title element.
//
// The following optional keys are recommended:
//   - "h1" is the H1 heading of the page.
//   - "lead" is the lead or introduction paragraph of the page.
//   - "logo" is the brief text inserted into the ASCII art logo.
//
// Other optional keys are also available:
//   - "canonical" is the canonical URL of the best representative page from a group of duplicate pages.
//   - "carousel" is the ID of the carousel to display.
//   - "databaseErr" is true if the database is not available.
//   - "subheading" is the H1 sub-heading of the page.
//   - "jsdos6" is true if the js-dos v6.22 emulator files are to be loaded.
//   - "readonlymode" is true if the application is in read-only mode.
//
// These keys are autofilled:
//   - "cachefiles" is the total number of records and used by the defacto2:file-count meta element.
//   - "editor" is true if the editor mode is enabled for the browser session.
func empty(c *echo.Context) PageData {
	const minSize = 20
	m := make(PageData, minSize)
	m["cachefiles"] = Caching.RecordCount
	m[canonical] = ""
	m["carousel"] = ""
	m["databaseErr"] = false
	m["description"] = ""
	m["editor"] = sess.Editor(c)
	m["h1"] = ""
	m["indexJS"] = false
	m["subheading"] = ""
	m["jsdos6"] = false
	m["lead"] = ""
	m["logo"] = ""
	m["readonlymode"] = true
	m["title"] = ""
	m["ogtitle"] = ""
	m["noindex"] = false
	m["uploader"] = true
	return m
}

// emptyFiles returns a PageData populated with defaults specific to artifact listing templates.
func emptyFiles(c *echo.Context) PageData {
	data := empty(c)
	data["bbs"] = false
	data["demozoo"] = "0"
	data["csdb"] = 0
	data["janeway"] = 0
	data["scener"] = ""
	data["sixteen"] = ""
	data["tidbits"] = ""
	data["website"] = []site.Website(nil)
	data["unknownYears"] = true
	data["noGroupLinks"] = false
	return data
}

// EmptyTester is a map of defaults for the app template tests.
func EmptyTester(c *echo.Context) map[string]any {
	return empty(c)
}

// staticPage holds the layout metadata for a static/editorial page.
type staticPage struct {
	template    string
	errContext  string
	title       string
	h1          string
	lead        string
	logo        string
	description string
	carousel    string
	noindex     bool
}

func renderStatic(sl *slog.Logger, c *echo.Context, p staticPage) error {
	if err := nils.Check(sl, c); err != nil {
		return fmt.Errorf("%s: %w", p.errContext, err)
	}

	data := empty(c)
	data.SetHeader(p.title, p.h1, p.lead, p.logo, p.description)
	if p.carousel != "" {
		data.SetCarousel(p.carousel)
	}
	if p.noindex {
		data.SetNoIndex(true)
	}

	if err := c.Render(http.StatusOK, p.template, data); err != nil {
		return InternalErr(sl, c, p.template, err)
	}
	return nil
}
