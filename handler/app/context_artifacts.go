//nolint:exhaustive,exhaustruct_v5,wrapcheck
package app

// Package file context_artifacts.go contains handlers for paginated artifact listings
// and category statistics. Individual artifact pages live in dirs.go.

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"net/http"
	"strconv"
	"strings"

	"github.com/Defacto2/helper"
	"github.com/Defacto2/server/handler/internal/fileslice"
	"github.com/Defacto2/server/internal/nils"
	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/Defacto2/server/model"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/dustin/go-humanize"
	"github.com/labstack/echo/v5"
)

const (
	limit      = 198 // per-page record limit
	categories = "categories"
)

// Artifacts is the handler for the list and preview of the files page.
// The uri is the category or collection of files to display.
// The page is the page number of the results to display.
func Artifacts(sl *slog.Logger, c *echo.Context, db *sql.DB, uri, page string) error {
	const format = "artifacts context: %w"
	if err := nils.Check(sl, c, db); err != nil {
		return fmt.Errorf(format, err)
	}

	ctx := c.Request().Context()
	switch {
	case !fileslice.Valid(uri):
		return ArtifactsErr(sl, c, uri)

	case page == "":
		return artifactsTable(ctx, sl, c, db, uri, 1)

	default:
		p, err := strconv.Atoi(page)
		if err != nil || p < 1 {
			return PageErr(sl, c, uri, page)
		}

		return artifactsTable(ctx, sl, c, db, uri, p)
	}
}

// Deletions is the handler to list the files that have been marked for deletion.
func Deletions(sl *slog.Logger, c *echo.Context, db *sql.DB, page string) error {
	return Artifacts(sl, c, db, fileslice.Deletions.String(), page)
}

// Unwanted is the handler to list the files that have been marked as unwanted.
func Unwanted(sl *slog.Logger, c *echo.Context, db *sql.DB, page string) error {
	return Artifacts(sl, c, db, fileslice.Unwanted.String(), page)
}

// ForApproval is the handler to list files awaiting editor approval.
func ForApproval(sl *slog.Logger, c *echo.Context, db *sql.DB, page string) error {
	return Artifacts(sl, c, db, fileslice.ForApproval.String(), page)
}

func zeroOk(uri string) bool {
	switch fileslice.Match(uri) {
	case fileslice.ForApproval, fileslice.Deletions, fileslice.Unwanted:
		return true
	default:
		return false
	}
}

func pageLimit(uri string) int {
	const oneOffsLimit = 250
	switch fileslice.Match(uri) {
	case fileslice.OneOffs, fileslice.OneOffsAZ:
		return oneOffsLimit
	default:
		return limit
	}
}

// artifactsTable is a helper function for Artifacts that renders the paginated artifacts list.
func artifactsTable(ctx context.Context, sl *slog.Logger, c *echo.Context, db *sql.DB, uri string, page int) error {
	const format = "sub-artifacts context: %w"
	if err := nils.Check(ctx, sl, c, db); err != nil {
		return fmt.Errorf(format, err)
	}

	const title = "Artifacts"
	const descr = "Table of contents for the collection of artifacts."
	logo, subhead, lead := fileslice.FileInfo(uri)

	data := emptyFiles(c)
	data.SetHeader(title, title, lead, logo, descr)
	data.SetCanonical(strings.Join([]string{files, uri}, "/"))
	data.SetSubheading(subhead)
	data["unknownYears"] = true
	data["forApproval"] = false

	spage := strconv.Itoa(page)
	errURI := "artifacts page " + spage + " for '" + uri + "'"

	d, sum, err := stats(ctx, db, uri)
	if err != nil {
		return DatabaseErr(sl, c, errURI, err)
	}
	lim := pageLimit(uri)
	lastPage := int(math.Ceil(float64(sum) / float64(lim)))
	if lastPage == 0 && zeroOk(uri) {
		lastPage = 1
	}
	if page < 1 || page > lastPage {
		return PageErr(sl, c, uri, spage)
	}

	data = artifactsDesc(uri, d[years], sum, data)
	data.SetStats(d)

	data.SetRecords([]models.FileSlice{})
	r, err := fileslice.Records(ctx, db, uri, page, lim)
	if err != nil {
		return DatabaseErr(sl, c, errURI, err)
	}
	if len(r) == 0 {
		if err = c.Render(http.StatusOK, artifacts, data); err != nil {
			return InternalErr(sl, c, errURI, err)
		}
		return nil
	}
	data.SetRecords(r)

	baseURL := "/files/" + uri
	if fileslice.Match(uri) == fileslice.OneOffsAZ {
		baseURL = "/files/one-offs/a-z"
		data.SetCanonical("files/one-offs/a-z")
	}

	const pages = 2
	data.SetPagination(Pagination{
		TwoAfter: page + pages,
		NextPage: page + 1,
		CurrPage: page,
		PrevPage: page - 1,
		TwoBelow: page - pages,
		SumPages: lastPage,
		BaseURL:  baseURL,
	})

	if err = c.Render(http.StatusOK, artifacts, data); err != nil {
		return InternalErr(sl, c, errURI, err)
	}

	return nil
}

func artifactsDesc(uri, years string, sum int, data PageData) PageData {
	match := fileslice.Match(uri)
	if match == -1 {
		return data
	}

	data.SetNoIndex(true)

	switch match {
	case fileslice.NewUploads:
		data.SetDescription("These are the most recent additions of scene history to the site.")
		data.SetTitle("New additions")
		data["unknownYears"] = false
	case fileslice.NewUpdates:
		data.SetDescription("Artifacts that have been recently updated or revised.")
		data.SetTitle("Artifact revisions and changes")
		data["unknownYears"] = false
	case fileslice.Deletions:
		data.SetTitle("Deleted artifacts")
		data["unknownYears"] = false
	case fileslice.Unwanted:
		data.SetTitle("Dangerous files artifacts")
		data["unknownYears"] = false
	case fileslice.ForApproval:
		data.SetTitle("For Approval artifacts")
		data["forApproval"] = true
	case fileslice.Oldest:
		data.SetDescription("These are the oldest known artifacts held by the site.")
		data.SetTitle("Oldest artifacts")
		data.SetNoIndex(false)
	case fileslice.Newest:
		data.SetDescription("These are more recent artifacts held by the site.")
		data.SetTitle("Recent artifacts")
	case fileslice.OneOffs:
		data.SetDescription("Artifacts from groups, releasers, BBSes, and FTP sites with only a single documented file.")
		data.SetTitle("One-off releases")
		data["noGroupLinks"] = true
	case fileslice.OneOffsAZ:
		data.SetDescription("Artifacts from groups, releasers, BBSes, and FTP sites with only a single documented file, ordered alphabetically.")
		data.SetTitle("One-off releases, A-Z")
		data["unknownYears"] = false
		data["noGroupLinks"] = true
	case fileslice.Sensenstahl:
		data.SetTitle("Sensenstahl artifacts")
	default:
		// catch all other matches
		s := strings.TrimSpace(fileslice.RecordsSub(uri))
		data.SetTitle(helper.Capitalize(s) + " artifacts")

		desc := "The collection of " +
			strconv.Itoa(sum) + " " + s + " artifacts"
		if years != "" {
			desc += " from " + years
		}
		data.SetDescription(desc + ".")
	}

	return data
}

// Categories is the handler for the artifact categories page.
func Categories(sl *slog.Logger, c *echo.Context, db *sql.DB, withStats bool) error {
	const format = "categories context: %w"
	if err := nils.Check(sl, c, db); err != nil {
		return fmt.Errorf(format, err)
	}

	const title = "Artifact categories"
	const descr = "A table of contents for the collection."
	const leadr = "This page shows the categories and platforms in the collection of file artifacts."

	data := empty(c)
	data.SetNoIndex(true)
	data.SetHeader(title, title, leadr, title, descr)
	data["stats"] = withStats
	data["counter"] = fileslice.Statistics()

	ctx := c.Request().Context()
	data, err := fileWStats(ctx, sl, db, data, withStats)
	if err != nil {
		sl.Warn("context_categories", slog.Any("error", err))
		data.SetDatabaseErr(true)
		data["stats"] = false
	}

	err = c.Render(http.StatusOK, categories, data)
	if err != nil {
		return InternalErr(sl, c, categories, err)
	}
	return nil
}

type categoryStatSummary struct {
	CountFmt  string
	BytesFmt  string
	MinYear   int
	MaxYear   int
	SpanYears int
}

type categoryStatRow struct {
	Name     string
	URI      string
	Count    int
	CountFmt string
	CountPct int
	Bytes    int
	BytesFmt string
	BytesPct int
	MinYear  int
	MaxYear  int
}

func buildStatRows(c fileslice.Stats) (categoryStatSummary, []categoryStatRow) {
	span := 0
	if c.Record.MinYear > 0 && c.Record.MaxYear >= c.Record.MinYear {
		span = c.Record.MaxYear - c.Record.MinYear + 1
	}
	summary := categoryStatSummary{
		CountFmt:  humanize.Comma(int64(c.Record.Count)),
		BytesFmt:  helper.ByteCountFloat(int64(c.Record.Bytes)),
		MinYear:   c.Record.MinYear,
		MaxYear:   c.Record.MaxYear,
		SpanYears: span,
	}

	sorted := c.SortCount()
	maxCount, maxBytes := 1, 1
	for _, item := range sorted {
		if item.Count > maxCount {
			maxCount = item.Count
		}
		if item.Bytes > maxBytes {
			maxBytes = item.Bytes
		}
	}

	rows := make([]categoryStatRow, 0, len(sorted))
	for _, item := range sorted {
		countPct := item.Count * 100 / maxCount
		if item.Count > 0 && countPct == 0 {
			countPct = 1
		}
		bytesPct := item.Bytes * 100 / maxBytes
		if item.Bytes > 0 && bytesPct == 0 {
			bytesPct = 1
		}
		rows = append(rows, categoryStatRow{
			Name:     item.Name,
			URI:      item.URI,
			Count:    item.Count,
			CountFmt: humanize.Comma(int64(item.Count)),
			CountPct: countPct,
			Bytes:    item.Bytes,
			BytesFmt: helper.ByteCountFloat(int64(item.Bytes)),
			BytesPct: bytesPct,
			MinYear:  item.MinYear,
			MaxYear:  item.MaxYear,
		})
	}
	return summary, rows
}

// fileWStats is a helper function for Categories that adds the statistics to the data map.
func fileWStats(ctx context.Context, sl *slog.Logger, db *sql.DB, data PageData, withStats bool) (
	PageData, error,
) {
	emptyMap := make(PageData)
	if err := nils.Check(ctx, sl, db); err != nil {
		return emptyMap, fmt.Errorf("file with stats: %w", err)
	}

	const title = "Artifact category statistics"
	const descr = "This page shows the artifact categories with selected statistics, " +
		"such as the number of files, storage size, and publishing span in each category or platform."
	if data == nil {
		data = emptyMap
	}
	if !withStats {
		return data, nil
	}

	c, err := fileslice.Counter(ctx, sl, db)
	if err != nil {
		return data, fmt.Errorf("counter: %w", err)
	}

	summary, rows := buildStatRows(c)
	data["counter"] = c
	data["statSummary"] = summary
	data["statRows"] = rows
	if len(c.Yearly) > 0 {
		if b, mErr := json.Marshal(c.Yearly); mErr == nil {
			data["statYearly"] = string(b)
		}
	}
	data.SetHeader(title, title, descr, title, descr)

	return data, nil
}

// stats is a helper function for Artifacts that returns the statistics for the files page.
func stats(ctx context.Context, exec boil.ContextExecutor, uri string) (map[string]string, int, error) {
	const format = `context artifacts stats %q: %w`
	if err := nils.Check(ctx, exec); err != nil {
		return nil, 0, fmt.Errorf(format, uri, err)
	}

	if !fileslice.Valid(uri) {
		return nil, 0, nil
	}

	m := model.Summary{
		SumBytes: sql.NullInt64{Int64: 0, Valid: false},
		SumCount: sql.NullInt64{Int64: 0, Valid: false},
		MinYear:  sql.NullInt16{Int16: 0, Valid: false},
		MaxYear:  sql.NullInt16{Int16: 0, Valid: false},
	}
	err := m.ByMatch(ctx, exec, uri)
	if err != nil {
		if !errors.Is(err, model.ErrURI) {
			return nil, 0, fmt.Errorf(format, uri, err)
		}
		if err := statsByURI(ctx, exec, uri, &m); err != nil {
			return nil, 0, err
		}
	}

	d := map[string]string{
		files: string(ByteFileS(ffile, m.SumCount.Int64, m.SumBytes.Int64)),
		years: helper.Years(m.MinYear.Int16, m.MaxYear.Int16),
	}
	return d, int(m.SumCount.Int64), nil
}

// statsByURI handles the different URI types for statistics calculation.
func statsByURI(ctx context.Context, exec boil.ContextExecutor, uri string, m *model.Summary) error {
	const format = "artifacts stats uri %s %q: %w"
	if err := nils.Check(ctx, exec, m); err != nil {
		return fmt.Errorf(format, "check", uri, err)
	}

	switch fileslice.Match(uri) {
	case fileslice.ForApproval:
		if err := m.ByForApproval(ctx, exec); err != nil {
			return fmt.Errorf(format, "by for approval", uri, err)
		}
	case fileslice.Deletions:
		if err := m.ByHidden(ctx, exec); err != nil {
			return fmt.Errorf(format, "by hidden", uri, err)
		}
	case fileslice.Unwanted:
		if err := m.ByUnwanted(ctx, exec); err != nil {
			return fmt.Errorf(format, "by unwanted", uri, err)
		}
	case
		fileslice.NewUploads,
		fileslice.NewUpdates,
		fileslice.Oldest,
		fileslice.Newest,
		fileslice.Sensenstahl:
		// For these cases, use the public artifacts method as fallback
		if err := m.ByPublic(ctx, exec); err != nil {
			return fmt.Errorf(format, "by public fallback", uri, err)
		}
	default:
		if err := m.ByPublic(ctx, exec); err != nil {
			return fmt.Errorf(format, "by public", uri, err)
		}
	}

	return nil
}
