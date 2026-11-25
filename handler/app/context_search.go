//nolint:wrapcheck
package app

// Package file context_search.go contains handlers for search forms and artifact search results.

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/Defacto2/helper"
	"github.com/Defacto2/server/internal/nils"
	"github.com/Defacto2/server/model"
	"github.com/labstack/echo/v5"
)

// FileSearch is the type of search to perform.
type FileSearch int

const (
	Filenames    FileSearch = iota // Filenames is the search for filenames.
	Descriptions                   // Descriptions is the search for file descriptions and titles.
)

const (
	searchhtmx = "searchhtmx"
	searchpost = "searchpost"
)

// cleanSearchTerms splits and trims comma-separated search terms, dropping blanks.
func cleanSearchTerms(input string) []string {
	inputs := helper.SearchTerm(input)
	terms := make([]string, 0, len(inputs))
	for _, term := range inputs {
		if trimmed := strings.TrimSpace(term); trimmed != "" {
			terms = append(terms, trimmed)
		}
	}
	return terms
}

// PostDesc is the handler for the Search for file descriptions form post page.
func PostDesc(sl *slog.Logger, c *echo.Context, db *sql.DB, input string) error {
	const format = "post desc context: %w"
	if err := nils.Check(sl, c, db); err != nil {
		return fmt.Errorf(format, err)
	}

	terms := cleanSearchTerms(input)
	if len(terms) == 0 {
		// Redirect back to the search form
		return c.Redirect(http.StatusFound, "/search/desc")
	}

	errs := "post desc search for, " + input
	ctx := c.Request().Context()
	fs, err := model.OnlyDescriptions(ctx, sl, db, terms)
	if err != nil {
		return DatabaseErr(sl, c, errs, err)
	}
	d := Descriptions.postStats(ctx, sl, db, terms)
	s := strings.Join(terms, ", ")

	data := emptyFiles(c)
	const brief = "Search titles & comments"
	data["noindex"] = true
	data["title"] = brief + " results"
	data["h1"] = brief
	data["lead"] = "Find games, apps, descriptions, and artifact comments."
	data["logo"] = s + " results"
	data["description"] = "Game or app titles search results for " + s + "."
	data["unknownYears"] = false
	data["searchMode"] = "desc"
	data["searchQuery"] = s
	data["searchLimit"] = model.Maximum
	data["searchTruncated"] = len(fs) >= model.Maximum
	data[records] = fs
	data["stats"] = d

	if err = c.Render(http.StatusOK, artifacts, data); err != nil {
		return InternalErr(sl, c, errs, err)
	}
	return nil
}

// PostFilename is the handler for the Search for filenames form post page.
func PostFilename(sl *slog.Logger, c *echo.Context, db *sql.DB) error {
	return PostName(sl, c, db, Filenames)
}

// PostName is the handler for the Search for filenames form post page.
func PostName(sl *slog.Logger, c *echo.Context, db *sql.DB, mode FileSearch) error {
	const format = "post name context: %w"
	if err := nils.Check(sl, c, db); err != nil {
		return fmt.Errorf(format, err)
	}

	const title = "Search filenames results"
	const h1 = "Search filenames"
	input := c.FormValue("search-term-query")
	if input == "" {
		input = c.QueryParam("query")
	}
	terms := cleanSearchTerms(input)
	if len(terms) == 0 {
		// Redirect back to the search form
		return c.Redirect(http.StatusFound, "/search/file")
	}
	errs := "post name search for, " + input
	ctx := c.Request().Context()
	fs, err := model.OnlyFilenames(ctx, db, terms)
	if err != nil {
		return DatabaseErr(sl, c, errs, err)
	}
	d := mode.postStats(ctx, sl, db, terms)
	s := strings.Join(terms, ", ")

	data := emptyFiles(c)
	data["noindex"] = true
	data["title"] = title
	data["h1"] = h1
	data["lead"] = "Find artifacts by archive filename or file extension."
	data["logo"] = s + " results"
	data["description"] = "Filename search results for " + s + "."
	data["unknownYears"] = false
	data["searchMode"] = ffile
	data["searchQuery"] = s
	data["searchLimit"] = model.Maximum
	data["searchTruncated"] = len(fs) >= model.Maximum
	data[records] = fs
	data["stats"] = d

	if err = c.Render(http.StatusOK, artifacts, data); err != nil {
		return InternalErr(sl, c, errs, err)
	}
	return nil
}

// postStats returns the summary statistics for filename or description search results.
func (mode FileSearch) postStats(ctx context.Context, sl *slog.Logger, db *sql.DB, terms []string) map[string]string {
	none := func(err error) map[string]string {
		if sl == nil {
			sl = slog.Default()
		}
		if err != nil {
			sl.Warn("file search post stats, no files found", slog.Any("error", err))
		}
		return map[string]string{files: "no files found", years: ""}
	}
	if len(terms) == 0 {
		return none(nil)
	}
	if err := nils.Check(ctx, db); err != nil {
		return none(err)
	}

	m := model.Summary{
		SumBytes: sql.NullInt64{Int64: 0, Valid: false},
		SumCount: sql.NullInt64{Int64: 0, Valid: false},
		MinYear:  sql.NullInt16{Int16: 0, Valid: false},
		MaxYear:  sql.NullInt16{Int16: 0, Valid: false},
	}
	switch mode {
	case Filenames:
		if err := m.ByFilename(ctx, db, terms); err != nil {
			return none(err)
		}
	case Descriptions:
		if err := m.ByDescription(ctx, db, terms); err != nil {
			return none(err)
		}
	default:
		// Handle unknown search modes
		return none(nil)
	}
	if m.SumCount.Int64 == 0 {
		return none(nil)
	}

	d := map[string]string{
		files: string(ByteFileS(ffile, m.SumCount.Int64, m.SumBytes.Int64)),
		years: helper.Years(m.MinYear.Int16, m.MaxYear.Int16),
	}
	return d
}

// SearchDesc is the handler for the Search for file descriptions page.
func SearchDesc(sl *slog.Logger, c *echo.Context) error {
	if err := nils.Check(c, sl); err != nil {
		return fmt.Errorf("search desc context: %w", err)
	}

	const title = "Search titles & comments"
	const lead = "Find games, apps, descriptions, and artifact comments."

	data := empty(c)
	data["noindex"] = true
	data["description"] = "Use this search to uncover named applications, games, and descriptions of artifacts."
	data["logo"] = search
	data["title"] = title
	data["h1"] = title
	data["lead"] = lead
	data["searchMode"] = "desc"

	if err := c.Render(http.StatusOK, searchpost, data); err != nil {
		return InternalErr(sl, c, searchpost, err)
	}
	return nil
}

// SearchID is the handler for the Record by ID Search page.
func SearchID(sl *slog.Logger, c *echo.Context) error {
	if err := nils.Check(c, sl); err != nil {
		return fmt.Errorf("search id context: %w", err)
	}

	const title = "Search by ID or UUID"
	const lead = "Look up artifacts by record ID, URL key, or UUID."

	data := empty(c)
	data["noindex"] = true
	data["description"] = "Use this search to lookup artifacts by their database identities."
	data["logo"] = search
	data["title"] = title
	data["h1"] = title
	data["lead"] = lead
	data["hxPost"] = "/editor/search/id"
	data["searchMode"] = "id"
	data["inputPlaceholder"] = "Type ID, URL key, or UUID (space or comma separated)..."

	if err := c.Render(http.StatusOK, searchhtmx, data); err != nil {
		return InternalErr(sl, c, searchhtmx, err)
	}
	return nil
}

// SearchFile is the handler for the Search for files page.
func SearchFile(sl *slog.Logger, c *echo.Context) error {
	if err := nils.Check(c, sl); err != nil {
		return fmt.Errorf("search file context: %w", err)
	}

	const title = "Search filenames"
	const lead = "Find artifacts by archive filename or file extension."

	data := empty(c)
	data["noindex"] = true
	data["description"] = "Use this search to lookup artifacts by their filenames."
	data["logo"] = search
	data["title"] = title
	data["h1"] = title
	data["lead"] = lead
	data["searchMode"] = ffile

	if err := c.Render(http.StatusOK, searchpost, data); err != nil {
		return InternalErr(sl, c, searchpost, err)
	}
	return nil
}

// SearchReleaser is the handler for the Releaser Search page.
func SearchReleaser(sl *slog.Logger, c *echo.Context) error {
	if err := nils.Check(c, sl); err != nil {
		return fmt.Errorf("search releaser context: %w", err)
	}

	const title = "Search groups & names"
	const lead = "Find groups, magazines, BBS boards, and FTP sites."

	data := empty(c)
	data["noindex"] = true
	data["description"] = "Lookup groups, magazines, BBS boards, and FTP sites by their names."
	data["logo"] = search
	data["title"] = title
	data["h1"] = title
	data["lead"] = lead
	data["helpText"] = "Acronyms match exact initialisms: 'rc' and 'rcn' return different results."
	data["hxPost"] = "/search/releaser"
	data["searchMode"] = "releaser"
	data["inputPlaceholder"] = "Type a group, magazine, BBS, or FTP name..."

	if err := c.Render(http.StatusOK, searchhtmx, data); err != nil {
		return InternalErr(sl, c, searchhtmx, err)
	}
	return nil
}
