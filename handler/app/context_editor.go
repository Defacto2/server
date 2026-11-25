//nolint:gochecknoglobals,wrapcheck
package app

// Package file context_editor.go contains handlers for editor/operator tools,
// diagnostics (configs, dependencies, routes), batch fixers, and inline metadata edits.

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"html/template"
	"log/slog"
	"maps"
	"net/http"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"sync"

	"github.com/Defacto2/helper"
	"github.com/Defacto2/server/internal/command"
	"github.com/Defacto2/server/internal/config"
	"github.com/Defacto2/server/internal/nils"
	"github.com/Defacto2/server/internal/postgres"
	"github.com/Defacto2/server/internal/tags"
	"github.com/Defacto2/server/model"
	"github.com/Defacto2/server/model/fix"
	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/labstack/echo/v5"
)

const (
	configs    = "configs"
	dep        = "dependencies"
	duplicates = "duplicates"
	fixers     = "fixers"
	routes     = "routes"
)

// Configurations is the handler for the Configuration page.
func Configurations(sl *slog.Logger, c *echo.Context, db *sql.DB, conf config.Config) error { //nolint:funlen
	const format = "configurations context: %w"
	if err := nils.Check(sl, c, db); err != nil {
		return fmt.Errorf(format, err)
	}

	const title = "Configs"
	const descr = "Defacto2 configurations."
	const leadr = "The web application configurations, tools and links to special records."

	data := empty(c)
	data["description"] = descr
	data["h1"] = "Configurations"
	data["lead"] = leadr
	data["title"] = title
	data["configurations"] = conf
	data["countArtifacts"] = 0
	data["countPublic"] = 0
	data["countNewUpload"] = 0
	data["countHidden"] = 0
	data["uuidVersions"] = ""

	ctx := c.Request().Context()
	// As we are collecting stats of both the file system and database, we may as well do it concurrently.
	var wg sync.WaitGroup
	var mu sync.Mutex

	wg.Go(func() {
		ca, cp, cnu, err := model.Counts(ctx, db)
		if err != nil {
			return
		}
		mu.Lock()
		data["countArtifacts"] = ca
		data["countPublic"] = cp
		data["countNewUpload"] = cnu
		data["countHidden"] = ca - cp - cnu
		mu.Unlock()
	})

	wg.Go(func() {
		host := HostStats(conf)
		var dirs map[string]any
		loading := !AssetsCached()
		if !loading {
			dirs = AssetDirsData(conf, false)
		}
		mu.Lock()
		maps.Copy(data, host)
		if loading {
			data["assetsLoading"] = true
		} else {
			maps.Copy(data, dirs)
		}
		mu.Unlock()
	})

	wg.Go(func() {
		vers, err := uuidCache.Get(ctx, db)
		if err != nil {
			return
		}
		mu.Lock()
		data["uuidVersions"] = vers
		mu.Unlock()
	})

	wg.Go(func() {
		idxStats, err := postgres.Indexes(ctx, db)
		mu.Lock()
		if err != nil {
			data["dbIndexesErr"] = err.Error()
		} else {
			data["dbIndexes"] = idxStats
		}
		mu.Unlock()
	})

	wg.Wait()

	count, maximum, err := postgres.Connections(ctx, db)
	if err != nil {
		data["dbConnections"] = err.Error()
	} else {
		data["dbConnections"] = fmt.Sprintf("%d of %d", count, maximum)
	}

	if err = c.Render(http.StatusOK, configs, data); err != nil {
		return InternalErr(sl, c, configs, err)
	}
	return nil
}

// UUIDCache caches the stats for the different UUID versions in use.
type UUIDCache struct {
	sync.RWMutex

	vers model.UUIDVers
	used bool
}

var uuidCache UUIDCache

// Get returns cached versions or gets them from the database.
func (c *UUIDCache) Get(ctx context.Context, db *sql.DB) (model.UUIDVers, error) {
	c.RLock()
	if c.used {
		v := c.vers
		c.RUnlock()
		return v, nil
	}
	c.RUnlock()

	c.Lock()
	defer c.Unlock()

	if !c.used {
		vers, err := model.UUIDs(ctx, db)
		if err != nil {
			return model.UUIDVers{}, err
		}
		c.vers = vers
	}
	c.used = true

	return c.vers, nil
}

type depSection struct {
	Title   string
	Caption string
	Tools   []command.ToolInfo
	OK      int
	Total   int
}

func newDepSection(title, caption string, tools []command.ToolInfo) depSection {
	ok := 0
	for _, tool := range tools {
		if tool.Found && !tool.Warning {
			ok++
		}
	}
	return depSection{
		Title:   title,
		Caption: caption,
		Tools:   tools,
		OK:      ok,
		Total:   len(tools),
	}
}

// Dependencies is the handler for the dependencies page.
func Dependencies(sl *slog.Logger, c *echo.Context) error {
	const format = "dependencies context: %w"
	if err := nils.Check(sl, c); err != nil {
		return fmt.Errorf(format, err)
	}

	const title = "Dependencies"
	const descr = "Defacto2 dependencies."
	const leadr = "The web application host system dependencies and tools."

	data := empty(c)
	data["description"] = descr
	data["h1"] = "Dependencies"
	data["lead"] = leadr
	data["title"] = title

	ctx := c.Request().Context()
	arcTools, imgTools := command.DependencyInfos(ctx)
	arcSec := newDepSection(
		"Archive & Compression Tools",
		"File and directory archive, compression, and package tools.",
		arcTools,
	)
	imgSec := newDepSection(
		"Image & Text Tools",
		"Image and text conversion tools.",
		imgTools,
	)
	data["depSections"] = []depSection{arcSec, imgSec}
	data["depOK"] = arcSec.OK + imgSec.OK
	data["depTotal"] = arcSec.Total + imgSec.Total

	if err := c.Render(http.StatusOK, dep, data); err != nil {
		return InternalErr(sl, c, dep, err)
	}
	return nil
}

type routeEntry struct {
	Method   string
	Path     string
	Badge    string
	Linkable bool
}

type routeMethodCount struct {
	Method string
	Badge  string
	Count  int
}

func routeBadge(method string) string {
	switch method {
	case http.MethodGet:
		return "bg-success-subtle text-success-emphasis border border-success-subtle"
	case http.MethodPost:
		return "bg-primary-subtle text-primary-emphasis border border-primary-subtle"
	case http.MethodPatch:
		return "bg-warning-subtle text-warning-emphasis border border-warning-subtle"
	case http.MethodPut:
		return "bg-info-subtle text-info-emphasis border border-info-subtle"
	case http.MethodDelete:
		return "bg-danger-subtle text-danger-emphasis border border-danger-subtle"
	default:
		return "bg-secondary-subtle text-secondary-emphasis border border-secondary-subtle"
	}
}

func buildRoutes(r echo.Routes) ([]routeEntry, []routeMethodCount) {
	entries := make([]routeEntry, 0, len(r))
	counts := make(map[string]int)
	for _, rt := range r {
		entries = append(entries, routeEntry{
			Method:   rt.Method,
			Path:     rt.Path,
			Badge:    routeBadge(rt.Method),
			Linkable: rt.Method == http.MethodGet && !strings.ContainsAny(rt.Path, ":*"),
		})
		counts[rt.Method]++
	}
	slices.SortFunc(entries, func(a, b routeEntry) int {
		if cmp := strings.Compare(a.Path, b.Path); cmp != 0 {
			return cmp
		}
		return strings.Compare(a.Method, b.Method)
	})

	preferred := []string{
		http.MethodGet,
		http.MethodPost,
		http.MethodPatch,
		http.MethodPut,
		http.MethodDelete,
	}
	methodCounts := make([]routeMethodCount, 0, len(counts))
	for _, m := range preferred {
		if cnt, ok := counts[m]; ok {
			methodCounts = append(methodCounts, routeMethodCount{
				Method: m,
				Badge:  routeBadge(m),
				Count:  cnt,
			})
			delete(counts, m)
		}
	}

	extra := make([]string, 0, len(counts))
	for m := range counts {
		extra = append(extra, m)
	}
	slices.Sort(extra)
	for _, m := range extra {
		methodCounts = append(methodCounts, routeMethodCount{
			Method: m,
			Badge:  routeBadge(m),
			Count:  counts[m],
		})
	}

	return entries, methodCounts
}

// Routes is the handler for the listing of all the routes page.
func Routes(sl *slog.Logger, c *echo.Context, r echo.Routes) error {
	const format = "routes context: %w"
	if err := nils.Check(sl, c); err != nil {
		return fmt.Errorf(format, err)
	}

	const title = "List of routes"
	const descr = "Web browser routes and parameters registered by the web application."

	entries, methodCounts := buildRoutes(r)

	data := empty(c)
	data["title"] = title
	data["description"] = descr
	data["logo"] = "Routes"
	data["h1"] = title
	data["lead"] = descr
	data["routesList"] = entries
	data["routesCount"] = len(entries)
	data["routeMethods"] = methodCounts

	if err := c.Render(http.StatusOK, routes, data); err != nil {
		return InternalErr(sl, c, routes, err)
	}
	return nil
}

// EditFn applies a single-column update inside a transaction and returns the updated artifact as JSON.
func EditFn(sl *slog.Logger, c *echo.Context, db *sql.DB,
	fn func(context.Context, boil.ContextExecutor, int64, string) error,
) error {
	const format = "editfn context %s: %w"
	if err := nils.Check(sl, c, db); err != nil {
		return fmt.Errorf(format, "check", err)
	}

	var f Form
	if err := c.Bind(&f); err != nil {
		return badRequest(c, err)
	}

	ctx := c.Request().Context()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return c.String(http.StatusServiceUnavailable, err.Error())
	}
	// Rollback is a no-op once tx.Commit succeeds; ensures cleanup on early return or panic.
	defer func() {
		_ = tx.Rollback()
	}()
	if err = fn(ctx, tx, int64(f.ID), f.Value); err != nil {
		return badRequest(c, err)
	}
	r, err := model.One(ctx, tx, true, f.ID)
	if err != nil {
		return fmt.Errorf(format, strconv.Itoa(f.ID), err)
	}
	if err = tx.Commit(); err != nil {
		return c.String(http.StatusServiceUnavailable, err.Error())
	}

	return c.JSON(http.StatusOK, r)
}

// PlatformEdit handles the post submission for the Platform selection field.
func PlatformEdit(sl *slog.Logger, c *echo.Context, db *sql.DB) error {
	return EditFn(sl, c, db, model.Platform.Update)
}

// TagEdit handles the post submission for the Tag selection field.
func TagEdit(sl *slog.Logger, c *echo.Context, db *sql.DB) error {
	return EditFn(sl, c, db, model.Section.Update)
}

// PlatformTagInfo handles the POST submission for the platform and tag info.
func PlatformTagInfo(c *echo.Context) error {
	const format = "platform tag info: %w"
	if err := nils.Check(c); err != nil {
		return fmt.Errorf(format, err)
	}

	var f Form
	if err := c.Bind(&f); err != nil {
		return badRequest(c, err)
	}

	info, err := tags.Platform(f.Platform, f.Tag)
	if err != nil {
		return badRequest(c, err)
	}

	return c.String(http.StatusOK, info)
}

// TagInfo handles the POST submission for the tag description info.
func TagInfo(c *echo.Context) error {
	if err := nils.Check(c); err != nil {
		return fmt.Errorf("tag info context: %w", err)
	}

	var f Form
	if err := c.Bind(&f); err != nil {
		return badRequest(c, err)
	}

	info, err := tags.Description(f.Tag)
	if err != nil {
		return badRequest(c, err)
	}
	return c.String(http.StatusOK, info)
}

// Fixers is the handler for the editor, batch-fixers page.
func Fixers(sl *slog.Logger, c *echo.Context, db *sql.DB) error {
	if err := nils.Check(sl, c, db); err != nil {
		return fmt.Errorf("fixers context: %w", err)
	}

	const title = "Fixers"
	data := empty(c)
	data["description"] = "Defacto2 fixers tool."
	data["h1"] = title
	data["lead"] = "Artifact fixes using batch-friendly tools."
	data["title"] = title

	// Get files with numeric suffixes
	ctx := c.Request().Context()
	filesFix, err := fix.NumSuffix(ctx, db)
	if err != nil {
		sl.Error("failed to get files with numeric suffixes", slog.String("error", err.Error()))
		// Don't return error, just continue without the data
	} else {
		data["numericSuffixCount"] = filesFix.Count
		// Add obfuscated IDs for the /f/ route
		for i := range filesFix.Files {
			filesFix.Files[i].ObfuscatedID = helper.ObfuscateID(filesFix.Files[i].ID)
		}
		// Pass the full file data (including ID, UUID, and obfuscated ID) to the template
		data["numericSuffixFiles"] = filesFix.Files
	}

	if err = c.Render(http.StatusOK, fixers, data); err != nil {
		return InternalErr(sl, c, fixers, err)
	}
	return nil
}

// Duplicates is the handler for the potential duplicate releasers, BBSes, and FTP sites report.
func Duplicates(sl *slog.Logger, c *echo.Context, db *sql.DB) error {
	if err := nils.Check(sl, c, db); err != nil {
		return fmt.Errorf("duplicates context: %w", err)
	}

	const title = "Potential duplicate releasers"
	data := empty(c)
	data["noindex"] = true
	data["description"] = "In-memory similarity report for detecting duplicate groups, BBSes, and FTP sites."
	data["logo"] = "Duplicate releasers"
	data["h1"] = title
	data["lead"] = "Groups, BBSes, and FTP sites with matching normalized fingerprints, swapped words, or 1-character typos."
	data["title"] = title

	ctx := c.Request().Context()
	rep, err := fix.Duplicates(ctx, db)
	if err != nil {
		sl.Error("failed to analyze duplicate releasers", slog.String("error", err.Error()))
	} else {
		data["dupReport"] = rep
	}

	if err := c.Render(http.StatusOK, duplicates, data); err != nil {
		return InternalErr(sl, c, duplicates, err)
	}
	return nil
}

// MergeDuplicate merges a duplicate releaser name into a canonical target name in a transaction.
func MergeDuplicate(sl *slog.Logger, c *echo.Context, db *sql.DB) error {
	if err := nils.Check(sl, c, db); err != nil {
		return fmt.Errorf("merge duplicate: %w", err)
	}
	from := c.FormValue("from")
	to := c.FormValue("to")
	if from == "" || to == "" || from == to {
		return c.String(http.StatusBadRequest, "invalid merge parameters")
	}

	ctx := c.Request().Context()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		sl.Error("merge duplicate begin tx", slog.Any("error", err))
		return c.String(http.StatusServiceUnavailable, "database transaction failed")
	}
	defer func() {
		if rErr := tx.Rollback(); rErr != nil && !errors.Is(rErr, sql.ErrTxDone) {
			sl.Error("merge duplicate rollback", slog.Any("error", rErr))
		}
	}()

	updated, err := fix.MergeReleaser(ctx, tx, from, to)
	if err != nil {
		sl.Error("merge duplicate update", slog.String("from", from), slog.String("to", to), slog.Any("error", err))
		return c.String(http.StatusInternalServerError, "merge failed")
	}
	if err := tx.Commit(); err != nil {
		sl.Error("merge duplicate commit", slog.Any("error", err))
		return c.String(http.StatusServiceUnavailable, "database commit failed")
	}

	sl.Info("merged duplicate releaser", slog.String("from", from), slog.String("to", to), slog.Int64("updated", updated))
	return c.HTML(http.StatusOK, fmt.Sprintf(
		`<li class="list-group-item list-group-item-success py-2 px-3 small">`+
			`Merged <code>%s</code> &rarr; <code>%s</code> (%d updated)</li>`,
		template.HTMLEscapeString(from), template.HTMLEscapeString(to), updated,
	))
}

var numericSuffixRegexes = []*regexp.Regexp{
	regexp.MustCompile(` \([0-9]{1,3}\)`), // Pattern with space: " (123)"
	regexp.MustCompile(`\([0-9]{1,3}\)`),  // Pattern without space: "(123)"
}

// FixNumericSuffix handles the fixing of numeric suffixes in filenames.
func FixNumericSuffix(sl *slog.Logger, c *echo.Context, db *sql.DB) error {
	if err := nils.Check(sl, c, db); err != nil {
		return fmt.Errorf("fix numeric suffix: %w", err)
	}

	obfuscatedID := c.Param("id")
	if obfuscatedID == "" {
		return fmt.Errorf("missing obfuscated ID: %w", ErrMissingObfuscatedID)
	}
	fileID := int64(helper.DeobfuscateID(obfuscatedID))
	if fileID == 0 {
		return fmt.Errorf("invalid obfuscated ID %s: %w", obfuscatedID, ErrInvalidObfuscatedID)
	}

	// Get the file from the database using the standard One function
	// Use deleted=true to allow fixing soft-deleted files
	ctx := c.Request().Context()
	file, err := model.One(ctx, db, true, int(fileID))
	if err != nil {
		logID(sl, "failure to find file in fix handler", obfuscatedID, fileID, err)
		return fmt.Errorf("failed to find file %s: %w", obfuscatedID, err)
	}
	if file == nil {
		logID(sl, "failure to find file in fix handler", obfuscatedID, fileID, ErrFileNotFound)
		return fmt.Errorf("file not found: %w", ErrFileNotFound)
	}

	// Remove the numeric suffix from the filename
	originalFilename := file.Filename.String
	// Try both patterns: with space and without space before the parenthesis
	var matched bool
	baseFilename := originalFilename
	for _, regex := range numericSuffixRegexes {
		if regex.MatchString(originalFilename) {
			baseFilename = regex.ReplaceAllString(originalFilename, "")
			matched = true
			break
		}
	}

	if !matched {
		return fmt.Errorf("filename does not match numeric suffix pattern: %s: %w",
			originalFilename, ErrInvalidFilenamePattern)
	}

	// Update the filename in the database
	file.Filename = null.StringFrom(baseFilename)
	_, err = file.Update(ctx, db, boil.Infer())
	if err != nil {
		return fmt.Errorf("failed to update filename %s: %w", baseFilename, err)
	}

	// Return the updated file info as HTML to replace the list item
	obfuscatedID = helper.ObfuscateID(file.ID)
	escBase := template.HTMLEscapeString(baseFilename)
	escOrig := template.HTMLEscapeString(originalFilename)
	escUUID := template.HTMLEscapeString(file.UUID.String)
	return c.HTML(http.StatusOK, fmt.Sprintf(`
		<div class="list-group-item list-group-item-success">
			<div class="d-flex justify-content-between align-items-center">
				<div>
					<code>%s</code>
					<small class="text-body-secondary d-block">
						ID: %d | UUID: %s
					</small>
					<small class="text-success d-block">
						Fixed: %s to %s
					</small>
				</div>
				<div>
					<a href="/f/%s" class="btn btn-sm btn-outline-secondary" target="_blank" rel="noopener">View</a>
				</div>
			</div>
		</div>
		`, escBase, file.ID, escUUID, escOrig, escBase, template.HTMLEscapeString(obfuscatedID)))
}
