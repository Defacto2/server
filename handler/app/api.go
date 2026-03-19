package app

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Defacto2/helper"
	"github.com/Defacto2/releaser"
	"github.com/Defacto2/server/handler/app/internal/filerecord"
	"github.com/Defacto2/server/handler/app/internal/fileslice"
	"github.com/Defacto2/server/handler/app/internal/simple"
	"github.com/Defacto2/server/handler/areacode"
	"github.com/Defacto2/server/handler/csdb"
	"github.com/Defacto2/server/handler/demozoo"
	"github.com/Defacto2/server/handler/janeway"
	"github.com/Defacto2/server/handler/site"
	"github.com/Defacto2/server/handler/sixteen"
	"github.com/Defacto2/server/internal/panics"
	"github.com/Defacto2/server/internal/postgres"
	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/Defacto2/server/internal/tags"
	"github.com/Defacto2/server/model"
	"github.com/Defacto2/server/model/html3"
	"github.com/aarondl/null/v8"
	"github.com/labstack/echo/v4"
)

const (
	APIBase = "/api/v1" // API Base URI (any changes need to be reflected in apiinfo.tmpl)
	APIVer  = "1.0.0"   // API Version gets shown in the HTTP header replies

	apiLimit = 1000
)

const (
	cacheDuration = 1 * time.Hour // Cache duration for query results
	cacheMaxItems = 1000          // Maximum number of items to keep in cache
)

// cacheQueryItem represents a cached query result.
type cacheQueryItem struct {
	data    any
	expires time.Time
}

var (
	apiCache   = make(map[string]cacheQueryItem) //nolint:gochecknoglobals
	apiCacheMu sync.RWMutex                      //nolint:gochecknoglobals
)

// ArtifactAPI represents an artifact file for API responses.
type ArtifactAPI struct {
	Summary       artifactAPI   `json:"artifact"`
	FileMeta      filemetaAPI   `json:"download"`
	ArtMeta       artmetaAPI    `json:"meta"`
	Relationships []relationAPI `json:"relationships"`
}

// artifactAPI represents an artifact file summary for API responses.
type artifactAPI struct {
	UUID          string       `json:"uuid"`
	Filename      string       `json:"filename"`
	DatePublished publishedAPI `json:"datePublished"`
	PostedDate    *time.Time   `json:"postedDate,omitempty"`
	Size          struct {
		Formatted string `json:"formatted"`
		Bytes     int64  `json:"bytes"`
	} `json:"size"`
	Description string `json:"description,omitempty"`
	Tags        struct {
		Category    string `json:"category"`
		Platform    string `json:"platform"`
		Description string `json:"description"`
	} `json:"tags"`
	URLs urlAPI `json:"urls"`
}

type urlAPI struct {
	API       string `json:"api"`
	Download  string `json:"download"`
	HTML      string `json:"html"`
	Thumbnail string `json:"thumbnail,omitempty"`
}

type filemetaAPI struct {
	Checksum     string     `json:"checksum"`
	LastModified *time.Time `json:"lastModified"`
	LastModAgo   string     `json:"lastModAgo,omitempty"`
	MimeType     string     `json:"mimeType"`
}

type artmetaAPI struct {
	Comment   string        `json:"comment"`
	Title     string        `json:"title"`
	Releasers []releaserAPI `json:"releasers"`
}

type relationAPI struct {
	Link string `json:"link"`
	Desc string `json:"description"`
}

type releaserAPI struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	API  string `json:"api"`
	HTML string `json:"html"`
}

// EntityAPI represents a group, scener, or releaser for API responses.
type EntityAPI struct {
	ID       string    `json:"id"`
	Name     string    `json:"name"`
	Title    string    `json:"title"`
	URLs     enityURLs `json:"urls"`
	Stats    totalsAPI `json:"statistics"`
	Websites any       `json:"websites,omitempty"`
	Sixteen  string    `json:"sixteen,omitempty"`
	Janeway  string    `json:"janeway,omitempty"`
	Demozoo  string    `json:"demozoo,omitempty"`
	Csdb     string    `json:"csdb,omitempty"`
}

type artifactsAPI struct {
	Files []artifactAPI `json:"artifacts"`
	Stats totalsAPI     `json:"statistics"`
}

// areacodeAPI represents an area code for API responses.
type areacodeAPI struct {
	Code        int      `json:"code"`
	Territories []string `json:"territories"`
	Notes       string   `json:"notes,omitempty"`
}

// enityURLs represents a collection releaser or group URLs.
type enityURLs struct {
	API   string `json:"api"`
	HTML3 string `json:"html3"`
	HTML  string `json:"html"`
}

type publishedAPI struct {
	Year  int16 `json:"year,omitempty"`
	Month int16 `json:"month,omitempty"`
	Day   int16 `json:"day,omitempty"`
}

// scenerAPI represents a scener for API responses.
type scenerAPI struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Title string `json:"title"`
	URLs  struct {
		API  string `json:"api"`
		HTML string `json:"html"`
	} `json:"urls"`
}

// tagAPI represents a tag category or tag platform for API responses.
type tagAPI struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Title       string    `json:"title,omitempty"`
	Description string    `json:"description,omitempty"`
	URLs        enityURLs `json:"urls"`
	Stats       totalsAPI `json:"statistics"`
}

// territoryAPI represents a territory for API responses.
type territoryAPI struct {
	Name         string `json:"name"`
	Abbreviation string `json:"abbreviation"`
	AreaCodes    []int  `json:"areaCodes"`
}

// totalsAPI for total file count, humanized size, and byte sizes.
type totalsAPI struct {
	TotalFiles     int64  `json:"totalFiles,omitempty"`
	TotalSize      string `json:"totalSize,omitempty"`
	TotalSizeBytes int64  `json:"totalSizeBytes,omitempty"`
}

// webpageAPI represents a group website for API responses.
type webpageAPI struct {
	URL     string `json:"url"`
	Name    string `json:"name"`
	Working bool   `json:"working"`
}

// websiteAPI represents a website for API responses.
type websiteAPI struct {
	Title    string `json:"title"`
	URL      string `json:"url"`
	Info     string `json:"info"`
	Category string `json:"category"`
}

// ArtifactsAPI returns a list of all files ordered by "oldest".
func ArtifactsAPI(c echo.Context, db *sql.DB, sl *slog.Logger) error {
	return ArtifactAPIs(c, db, sl, "oldest")
}

// ArtifactsNewAPI returns a list of all files ordered by "new-uploads".
func ArtifactsNewAPI(c echo.Context, db *sql.DB, sl *slog.Logger) error {
	return ArtifactAPIs(c, db, sl, "new-uploads")
}

// ArtifactAPIs returns a list of all files filtered by the provided uri string.
func ArtifactAPIs(c echo.Context, db *sql.DB, sl *slog.Logger, uri string) error {
	const msg = "artifacts api"
	if err := panics.EchoContextDS(c, db, sl); err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}

	const limit = apiLimit
	page := 1
	if s := c.QueryParam("page"); s != "" {
		var err error
		page, err = strconv.Atoi(s)
		if err != nil || page < 1 {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "Invalid page parameter",
			})
		}
	}

	ctx := context.Background()
	fs, err := fileslice.Records(ctx, db, uri, page, limit)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to query files",
		})
	}
	count, err := model.Count(ctx, db)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to query count",
		})
	}
	files := artifactSummaries(fs)
	pages := (count + limit - 1) / limit // Ceiling division
	response := artifactsAPI{
		Files: files,
		Stats: totalsAPI{
			TotalFiles:     count,
			TotalSize:      "",
			TotalSizeBytes: 0,
		},
	}

	return c.JSON(http.StatusOK, map[string]any{
		"artifacts":  response.Files,
		"statistics": response.Stats,
		"page":       page,
		"totalPages": pages,
		"limit":      apiLimit,
	})
}

// FileAPI returns a single file by its obfuscated ID.
func FileAPI(c echo.Context, db *sql.DB, sl *slog.Logger) error {
	const msg = "file api"
	if err := panics.EchoContextDS(c, db, sl); err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}

	hash := c.Param("id")
	if hash == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Missing id parameter",
		})
	}

	ctx := context.Background()
	fileID := helper.DeobfuscateID(hash)
	if fileID == 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid file hash",
		})
	}

	// Get the file record by ID
	record, err := models.FindFile(ctx, db, int64(fileID))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "File not found",
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to query file",
		})
	}

	file := artifact(record)
	return c.JSON(http.StatusOK, file)
}

// APIMarkup removes CSS classes and attributes from HTML for API responses.
// Keeps semantic HTML tags but removes presentation-specific markup.
func APIMarkup(src string) string {
	if src == "" {
		return src
	}
	// This func was generated by Minstal using the devstral-2 model.

	// First, remove anchor tags without href attribute (keep content)
	// This must be done BEFORE cleaning other anchor tags
	re := regexp.MustCompile(`<a\b[^>]*>([^<]*)<\/a>`)
	// Only remove anchors that don't contain href=
	src = re.ReplaceAllStringFunc(src, func(match string) string {
		if strings.Contains(match, `href="`) {
			return match // Keep anchors with href
		}
		// Extract content from anchors without href
		re2 := regexp.MustCompile(`<a\b[^>]*>([^<]*)<\/a>`)
		if matches := re2.FindStringSubmatch(match); len(matches) > 1 {
			return matches[1]
		}
		return match
	})

	// Now clean anchor tags with href - remove all attributes except href
	re = regexp.MustCompile(`<a\s+([^>]*href="([^"]*)"[^>]*)>([^<]*)<\/a>`)
	src = re.ReplaceAllString(src, `<a href="$2">$3</a>`)

	// Remove specific attributes from all tags (class, style, id, title, data-*)
	// But preserve href attributes on anchor tags
	re = regexp.MustCompile(`(class|style|id|title|data-[a-z-]*)="[^"]*"`)
	src = re.ReplaceAllString(src, "")

	// Remove any remaining empty attributes or whitespace in tags
	re = regexp.MustCompile(`<\s*([a-z0-9]+)\s+>`)
	src = re.ReplaceAllString(src, "<$1>")

	// Remove empty element pairs (e.g., "<span></span>" or "<span> </span>" -> "")
	// We'll handle common empty tags specifically, including those with only whitespace
	emptyTags := []string{"span", "div", "p", "i", "b", "strong", "em", "a", "q", "code"}
	for _, tag := range emptyTags {
		re := regexp.MustCompile(`<` + tag + `[^>]*>\s*<\/` + tag + `>`)
		src = re.ReplaceAllString(src, "")
	}

	// Remove spaces between closing and opening tags (e.g., "</div> <p>" -> "</div><p>")
	re = regexp.MustCompile(`>\s+<`)
	src = re.ReplaceAllString(src, "><")

	// Clean up multiple spaces (but preserve line breaks for readability)
	re = regexp.MustCompile(`[\t\r\n]+`)
	src = re.ReplaceAllString(src, " ")
	re = regexp.MustCompile(`\s{2,}`)
	src = re.ReplaceAllString(src, " ")

	return strings.TrimSpace(src)
}

// AreacodesAPI returns all North American Numbering Plan (NANP) area codes.
func AreacodesAPI(c echo.Context) error {
	codes := areacode.AreaCodes()
	if len(codes) == 0 {
		return c.JSON(http.StatusOK, []areacodeAPI{})
	}
	result := make([]areacodeAPI, 0, len(codes))
	for _, code := range codes {
		territories := areacode.TerritoryByCode(code)
		names := make([]string, 0, len(territories))
		for _, t := range territories {
			names = append(names, t.Name)
		}
		apiCode := areacodeAPI{
			Code:        int(code),
			Territories: names,
			Notes:       "",
		}
		if note, ok := areacode.Notes()[code]; ok {
			apiCode.Notes = note
		}
		result = append(result, apiCode)
	}

	return c.JSON(http.StatusOK, result)
}

// AreaCodeAPI returns details for a specific area code.
func AreaCodeAPI(c echo.Context) error {
	s := c.Param("code")
	if s == "" {
		return c.JSON(http.StatusBadRequest, "area code parameter is required")
	}
	code, err := strconv.Atoi(s)
	if err != nil {
		return c.JSON(http.StatusBadRequest, "invalid area code format")
	}

	nancode := areacode.NAN(code)
	if !nancode.Valid() {
		return c.JSON(http.StatusNotFound, "area code not found")
	}
	territories := areacode.TerritoryByCode(nancode)
	names := make([]string, 0, len(territories))
	for _, t := range territories {
		names = append(names, t.Name)
	}
	result := areacodeAPI{
		Code:        code,
		Territories: names,
		Notes:       "",
	}
	if note, ok := areacode.Notes()[nancode]; ok {
		result.Notes = note
	}

	return c.JSON(http.StatusOK, result)
}

// AreacodeSearchAPI searches for area codes or territories by query.
func AreacodeSearchAPI(c echo.Context) error {
	query := c.Param("query")
	if query == "" {
		return c.JSON(http.StatusBadRequest, "search query is required")
	}

	// Parse as an area code
	if code, err := strconv.Atoi(query); err == nil {
		nancode := areacode.NAN(code)
		if nancode.Valid() {
			territories := areacode.TerritoryByCode(nancode)
			names := make([]string, 0, len(territories))
			for _, t := range territories {
				names = append(names, t.Name)
			}
			result := areacodeAPI{
				Code:        code,
				Territories: names,
				Notes:       "",
			}
			if note, ok := areacode.Notes()[nancode]; ok {
				result.Notes = note
			}
			return c.JSON(http.StatusOK, map[string]any{
				"areacodes":   []areacodeAPI{result},
				"territories": []territoryAPI{},
			})
		}
	}

	// Territory lookup
	territories := areacode.Lookup(query)
	if len(territories) > 0 {
		results := make([]territoryAPI, 0, len(territories))
		for _, t := range territories {
			areaCodes := make([]int, 0, len(t.AreaCodes))
			for _, ac := range t.AreaCodes {
				areaCodes = append(areaCodes, int(ac))
			}
			results = append(results, territoryAPI{
				Name:         t.Name,
				Abbreviation: string(t.Abbreviation),
				AreaCodes:    areaCodes,
			})
		}

		return c.JSON(http.StatusOK, map[string]any{
			"areacodes":   []areacodeAPI{},
			"territories": results,
		})
	}

	return c.JSON(http.StatusOK, map[string]any{
		"areacodes":   []areacodeAPI{},
		"territories": []territoryAPI{},
	})
}

// MilestonesAPI returns all milestones.
func MilestonesAPI(c echo.Context) error {
	return milestonesAPI(c, false)
}

// MilestoneHighlightsAPI returns only highlighted milestones.
func MilestoneHighlightsAPI(c echo.Context) error {
	return milestonesAPI(c, true)
}

// MilestonesAPI returns all milestones.
// When highlights is true, only the highlighted milestones will be returned.
func milestonesAPI(c echo.Context, highlights bool) error {
	all := Collection()
	result := make(Milestones, len(all))
	for i, m := range all {
		if highlights && !m.Highlight {
			continue
		}
		result[i] = milestoneFmt(m)
	}
	if highlights {
		// delete empty slots created by the highlights bool, these default to Highlight=false
		result = slices.DeleteFunc(result, func(m Milestone) bool {
			return !m.Highlight
		})
	}
	return c.JSON(http.StatusOK, result)
}

// MilestoneYearAPI returns milestones for a specific year.
func MilestoneYearAPI(c echo.Context) error {
	year, err := strconv.Atoi(c.Param("year"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid year format",
		})
	}

	var result Milestones
	for _, m := range Collection() {
		if m.Year == year {
			result = append(result, milestoneFmt(m))
		}
	}
	if len(result) == 0 {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "No milestones found for this year",
		})
	}

	return c.JSON(http.StatusOK, result)
}

// MilestoneYearsAPI returns milestones within a year range.
func MilestoneYearsAPI(c echo.Context) error {
	rangeParam := c.Param("range")
	parts := strings.Split(rangeParam, "-")

	const expectedParts = 2
	if len(parts) != expectedParts {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid range format. Use format: start-end",
		})
	}

	start, err := strconv.Atoi(parts[0])
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid start year",
		})
	}

	end, err := strconv.Atoi(parts[1])
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid end year",
		})
	}

	if start > end {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Start year must be less than or equal to end year",
		})
	}

	var result Milestones
	for _, m := range Collection() {
		if m.Year >= start && m.Year <= end {
			result = append(result, milestoneFmt(m))
		}
	}

	if len(result) == 0 {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "No milestones found for this year range",
		})
	}

	return c.JSON(http.StatusOK, result)
}

// MilestoneDecadeAPI returns milestones for a specific decade.
func MilestoneDecadeAPI(c echo.Context) error {
	decadeParam := c.Param("decade")
	decade, err := strconv.Atoi(strings.TrimSuffix(decadeParam, "s"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid decade format. Use format like '1980s'",
		})
	}

	const years = 9
	startYear := decade
	endYear := decade + years

	var result Milestones
	for _, m := range Collection() {
		if m.Year >= startYear && m.Year <= endYear {
			result = append(result, milestoneFmt(m))
		}
	}

	if len(result) == 0 {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "No milestones found for this decade",
		})
	}

	return c.JSON(http.StatusOK, result)
}

// milestoneFmt prepares a milestone for API response.
func milestoneFmt(m Milestone) Milestone {
	m.ContentHTML = APIMarkup(m.Content)
	m.Content = simple.CleanHTML(m.Content)
	m.Lead = APIMarkup(m.Lead)
	if len(m.List) > 0 {
		for i := range m.List {
			m.List[i].LinkTitle = APIMarkup(m.List[i].LinkTitle)
			m.List[i].SubTitle = APIMarkup(m.List[i].SubTitle)
		}
	}

	return m
}

// CategoriesAPI returns all categories.
func CategoriesAPI(c echo.Context, db *sql.DB) error {
	return TagsAPI(c, db, true, false)
}

// PlatformsAPI returns all platforms.
func PlatformsAPI(c echo.Context, db *sql.DB) error {
	return TagsAPI(c, db, false, true)
}

// GroupsAPI returns a list of all releasers/groups with pagination.
func GroupsAPI(c echo.Context, db *sql.DB, sl *slog.Logger) error {
	const msg = "groups api"
	if err := panics.EchoContextDS(c, db, sl); err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}

	// parse page parameter or default to page 1
	page := 1
	if s := c.QueryParam("page"); s != "" {
		var err error
		page, err = strconv.Atoi(s)
		if err != nil || page < 1 {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "Invalid page parameter",
			})
		}
	}

	ctx := context.Background()
	count, err := groupsCount(ctx, db)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to count releasers",
		})
	}
	pages := (count + apiLimit - 1) / apiLimit // Ceiling division
	rels := model.Releasers{}
	if err := rels.Limit(ctx, db, model.Alphabetical, apiLimit, page); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to query releasers",
		})
	}
	if len(rels) == 0 {
		return c.JSON(http.StatusOK, map[string]any{
			"releasers":  []EntityAPI{},
			"page":       page,
			"totalPages": pages,
		})
	}

	result := ReleasersAPI(rels)
	return c.JSON(http.StatusOK, map[string]any{
		"releasers":  result,
		"page":       page,
		"totalPages": pages,
	})
}

// MagazinesAPI is the handler for the magazines API endpoint.
func MagazinesAPI(c echo.Context, db *sql.DB, sl *slog.Logger) error {
	return cachedReleasersAPI(c, db, sl, "magazines_all", func(ctx context.Context, db *sql.DB) error {
		rels := model.Releasers{}
		return rels.Magazine(ctx, db)
	})
}

// BoardsAPI is the handler for the BBS API endpoint.
func BoardsAPI(c echo.Context, db *sql.DB, sl *slog.Logger) error {
	return cachedReleasersAPI(c, db, sl, "boards_all", func(ctx context.Context, db *sql.DB) error {
		rels := model.Releasers{}
		return rels.BBS(ctx, db, model.Oldest)
	})
}

// SitesAPI is the handler for the FTP sites API endpoint.
func SitesAPI(c echo.Context, db *sql.DB, sl *slog.Logger) error {
	return cachedReleasersAPI(c, db, sl, "sites_all", func(ctx context.Context, db *sql.DB) error {
		rels := model.Releasers{}
		return rels.FTP(ctx, db)
	})
}

// groupsCount returns the total number of releasers.
func groupsCount(ctx context.Context, db *sql.DB) (int, error) {
	var names model.ReleaserNames
	if err := names.DistinctGroups(ctx, db); err != nil {
		return 0, fmt.Errorf("groups counter: %w", err)
	}
	return len(names), nil
}

// ReleasersAPI builds the ReleaserAPI list from model data.
func ReleasersAPI(rels model.Releasers) []EntityAPI {
	results := make([]EntityAPI, 0, len(rels))
	for _, rel := range rels {
		title := releaser.Link(rel.Unique.Name)
		name := releaser.Obfuscate(rel.Unique.Name)
		count := rel.Unique.Count
		bytes := rel.Unique.Bytes
		// As there is no unique ids for releasers, only a unique uri,
		// create stable ID from the url using an obfuscated name
		id := simple.Hash(name)

		result := EntityAPI{
			ID:    id,
			Name:  name,
			Title: title,
			URLs: enityURLs{
				API:   APIBase + "/releaser/" + name,
				HTML3: "/html3/group/" + name,
				HTML:  "/g/" + name,
			},
			Stats: totalsAPI{
				TotalFiles:     int64(count),
				TotalSize:      helper.ByteCount(int64(bytes)),
				TotalSizeBytes: int64(bytes),
			},
			Websites: site.Find(name),
			Sixteen:  linkSixteen(name),
			Janeway:  linkJaneway(name),
			Csdb:     linkCsdb(name),
			Demozoo:  linkDemozoo(name),
		}
		results = append(results, result)
	}
	return results
}

// artifactSum creates an ArtifactSumAPI from a file model.
func artifactSum(f *models.File) artifactAPI {
	category := filerecord.TagCategory(f)
	platform := filerecord.TagProgram(f)
	categoryTag := tags.TagByURI(category)
	platformTag := tags.TagByURI(platform)
	humanized := tags.Humanize(platformTag, categoryTag)

	art := &models.File{ //nolint:exhaustruct
		Filename:       f.Filename,
		RecordTitle:    f.RecordTitle,
		GroupBrandBy:   f.GroupBrandBy,
		GroupBrandFor:  f.GroupBrandFor,
		DateIssuedYear: f.DateIssuedYear,
	}

	result := artifactAPI{
		UUID:     filerecord.UnID(f),
		Filename: filerecord.Basename(f),
		DatePublished: publishedAPI{
			Year:  f.DateIssuedYear.Int16,
			Month: f.DateIssuedMonth.Int16,
			Day:   f.DateIssuedDay.Int16,
		},
		PostedDate: f.Createdat.Ptr(),
		Size: struct {
			Formatted string `json:"formatted"`
			Bytes     int64  `json:"bytes"`
		}{Formatted: helper.ByteCount(f.Filesize.Int64), Bytes: f.Filesize.Int64},
		Description: filerecord.Description(art),
		Tags: struct {
			Category    string `json:"category"`
			Platform    string `json:"platform"`
			Description string `json:"description"`
		}{
			Category:    category,
			Platform:    platform,
			Description: humanized,
		},
		URLs: urlAPI{
			API:       APIBase + "/artifact/" + helper.ObfuscateID(f.ID),
			Download:  "/d/" + helper.ObfuscateID(f.ID),
			HTML:      "/f/" + helper.ObfuscateID(f.ID),
			Thumbnail: "/public/image/thumb/" + f.UUID.String,
		},
	}
	if f.Createdat.Valid {
		result.PostedDate = &f.Createdat.Time
	}

	return result
}

// ReleaserAPI returns details for a specific releaser or group.
func ReleaserAPI(c echo.Context, db *sql.DB, sl *slog.Logger) error {
	const msg = "releaser api"
	if err := panics.EchoContextDS(c, db, sl); err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}

	name := c.Param("name")
	if name == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Releaser name parameter is required",
		})
	}
	ctx := context.Background()
	rels := model.Releasers{}
	fs, err := rels.Where(ctx, db, name)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to query releaser",
		})
	}
	if len(fs) == 0 {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "Releaser not found",
		})
	}

	sum := model.Summary{ //nolint:exhaustruct // Fields are set by ByReleaser method
	}
	if err := sum.ByReleaser(ctx, db, name); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to get releaser statistics",
		})
	}

	result := make([]artifactAPI, 0, len(fs))
	for _, f := range fs {
		result = append(result, artifactSum(f))
	}
	sites := site.Find(name)
	websites := make([]webpageAPI, len(sites))
	for i, site := range sites {
		websites[i] = webpageAPI{
			URL:     site.URL,
			Name:    site.Name,
			Working: !site.NotWorking,
		}
	}
	return c.JSON(http.StatusOK, map[string]any{
		"group": EntityAPI{
			ID:    simple.Hash(name),
			Name:  name,
			Title: releaser.Link(name),
			URLs: enityURLs{
				API:   APIBase + "releaser/" + name,
				HTML3: "/html3/group/" + name,
				HTML:  "/g/" + name,
			},
			Stats: totalsAPI{
				TotalFiles:     sum.SumCount.Int64,
				TotalSize:      helper.ByteCount(sum.SumBytes.Int64),
				TotalSizeBytes: sum.SumBytes.Int64,
			},
			Websites: websites,
			Sixteen:  linkSixteen(name),
			Janeway:  linkJaneway(name),
			Demozoo:  linkDemozoo(name),
			Csdb:     linkCsdb(name),
		},
		"artifacts": result,
	})
}

func linkSixteen(uri string) string {
	if tag := sixteen.Find(uri); tag != "" {
		return "https://16colo.rs/" + string(tag)
	}
	return ""
}

func linkJaneway(uri string) string {
	if id := janeway.Find(uri); id != 0 {
		return "https://janeway.exotica.org.uk/author.php?id=" + strconv.Itoa(int(id))
	}
	return ""
}

func linkCsdb(uri string) string {
	if id := csdb.Find(uri); id != 0 {
		return "https://csdb.dk/group/?id=" + strconv.Itoa(int(id))
	}
	return ""
}

func linkDemozoo(uri string) string {
	if id := demozoo.Find(uri); id != 0 {
		return "https://demozoo.org/groups/" + strconv.Itoa(int(id))
	}
	return ""
}

// ScenerAPI returns details for a specific scener.
func ScenerAPI(c echo.Context, db *sql.DB, sl *slog.Logger) error {
	const msg = "scener api"
	if err := panics.EchoContextDS(c, db, sl); err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}

	name := c.Param("name")
	if name == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Scener name parameter is required",
		})
	}

	ctx := context.Background()
	srs := model.Scener(name)
	fs, err := srs.Where(ctx, db, name)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "Failed to fetch scener",
		})
	}
	if len(fs) == 0 {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "Scener not found",
		})
	}

	sum := model.Summary{ //nolint:exhaustruct // Fields are set by ByReleaser method
	}
	if err := sum.ByScener(ctx, db, name); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to get scener statistics",
		})
	}

	result := make([]artifactAPI, 0, len(fs))
	for _, f := range fs {
		result = append(result, artifactSum(f))
	}
	return c.JSON(http.StatusOK, map[string]any{
		"scener": EntityAPI{
			ID:    simple.Hash(name),
			Name:  name,
			Title: releaser.Link(name),
			URLs: struct {
				API   string `json:"api"`
				HTML3 string `json:"html3"`
				HTML  string `json:"html"`
			}{
				API:   APIBase + "/scener/" + name,
				HTML3: "",
				HTML:  "/p/" + name,
			},
			Stats: totalsAPI{
				TotalFiles:     sum.SumCount.Int64,
				TotalSize:      helper.ByteCount(sum.SumBytes.Int64),
				TotalSizeBytes: sum.SumBytes.Int64,
			},
			Websites: nil,
			Sixteen:  "",
			Janeway:  "",
			Demozoo:  "",
			Csdb:     "",
		},
		"artifacts": result,
	})
}

// ScenersAPI returns a list of Sceners
// ScenersAPI builds the ReleaserAPI list from model data.
func ScenersAPI(c echo.Context, db *sql.DB, sl *slog.Logger) error {
	return roleAPI(c, db, sl, postgres.Roles())
}

func ArtistsAPI(c echo.Context, db *sql.DB, sl *slog.Logger) error {
	return roleAPI(c, db, sl, postgres.Artist)
}

func CodersAPI(c echo.Context, db *sql.DB, sl *slog.Logger) error {
	return roleAPI(c, db, sl, postgres.Coder)
}

func MusiciansAPI(c echo.Context, db *sql.DB, sl *slog.Logger) error {
	return roleAPI(c, db, sl, postgres.Musician)
}

func WritersAPI(c echo.Context, db *sql.DB, sl *slog.Logger) error {
	return roleAPI(c, db, sl, postgres.Writer)
}

// roleAPI returns a list of all releasers/groups with pagination.
func roleAPI(c echo.Context, db *sql.DB, sl *slog.Logger, r postgres.Role) error {
	const msg = "sceners api"
	if err := panics.EchoContextDS(c, db, sl); err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}

	ctx := context.Background()
	srs := model.Sceners{}
	var err error
	switch r {
	case postgres.Writer:
		err = srs.Writer(ctx, db)
	case postgres.Artist:
		err = srs.Artist(ctx, db)
	case postgres.Musician:
		err = srs.Musician(ctx, db)
	case postgres.Coder:
		err = srs.Coder(ctx, db)
	case postgres.Roles():
		err = srs.Distinct(ctx, db)
	default:
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Sceners role is unknown",
		})
	}
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to fetch sceners",
		})
	}
	if len(srs) == 0 {
		return c.JSON(http.StatusOK, map[string]any{
			"sceners":    []scenerAPI{},
			"page":       1,
			"totals":     0,
			"totalPages": 1,
		})
	}
	result := scenersAPI(srs)
	return c.JSON(http.StatusOK, map[string]any{
		"sceners":    result,
		"page":       1,
		"totals":     len(result),
		"totalPages": 1,
	})
}

func scenersAPI(srs model.Sceners) []scenerAPI {
	sceners := srs.Sort()
	results := make([]scenerAPI, 0, len(sceners))
	for _, s := range sceners {
		title := helper.Capitalize(strings.ToLower(s)) // scener // releaser.Link(scener.Unique.Name)
		name := helper.Slug(s)
		html, _ := LinkScnr(s) // any errors can be ignored as it will leave "HTML" blank
		// create stable ID from the url using an obfuscated name
		id := simple.Hash(name)

		result := scenerAPI{
			ID:    id,
			Name:  name,
			Title: title,
			URLs: struct {
				API  string `json:"api"`
				HTML string `json:"html"`
			}{
				API:  APIBase + "/scener/" + name,
				HTML: html,
			},
		}
		results = append(results, result)
	}
	return results
}

// cachedResults returns cached API results if available.
func cachedResults(key string) (any, bool) {
	apiCacheMu.RLock()
	defer apiCacheMu.RUnlock()

	if item, exists := apiCache[key]; exists && time.Now().Before(item.expires) {
		return item.data, true
	}
	return nil, false
}

// cacheResults stores API results in the cache.
func cacheResults(key string, data any) {
	apiCacheMu.Lock()
	defer apiCacheMu.Unlock()

	// Clean up old cache entries if we're approaching the limit
	if len(apiCache) >= cacheMaxItems {
		for key, item := range apiCache {
			if time.Now().After(item.expires) {
				delete(apiCache, key)
			}
		}
	}

	apiCache[key] = cacheQueryItem{
		data:    data,
		expires: time.Now().Add(cacheDuration),
	}
}

// cachedReleasersAPI handles the common pattern for caching
// releaser-based API endpoints.
func cachedReleasersAPI(
	c echo.Context,
	db *sql.DB,
	sl *slog.Logger,
	key string,
	queryFunc func(ctx context.Context, db *sql.DB) error,
) error {
	const msg = "cached releasers api"
	if err := panics.EchoContextDS(c, db, sl); err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}

	if data, found := cachedResults(key); found {
		if i, ok := data.(map[string]any); ok {
			return c.JSON(http.StatusOK, i)
		}
	}

	ctx := context.Background()
	rels := model.Releasers{}
	if err := queryFunc(ctx, db); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to query releasers",
		})
	}
	if len(rels) == 0 {
		result := map[string]any{
			"releasers":  []EntityAPI{},
			"page":       1,
			"totalPages": 1,
		}
		cacheResults(key, result)
		return c.JSON(http.StatusOK, result)
	}

	result := map[string]any{
		"releasers":  ReleasersAPI(rels),
		"page":       1,
		"totalPages": 1,
	}
	cacheResults(key, result)

	return c.JSON(http.StatusOK, result)
}

// cachedTags returns cached tag results if available.
func cachedTags(category, platform bool) ([]tagAPI, bool) {
	cacheKey := fmt.Sprintf("tags_category=%t_platform=%t", category, platform)

	if data, found := cachedResults(cacheKey); found {
		if results, ok := data.([]tagAPI); ok {
			return results, true
		}
	}
	return nil, false
}

// tagsCache stores tag results in the cache.
func tagsCache(category, platform bool, results []tagAPI) {
	cacheKey := fmt.Sprintf("tags_category=%t_platform=%t", category, platform)
	cacheResults(cacheKey, results)
}

// TagsAPI returns artifact tags.
//
//   - Set categories true to return all categories.
//   - Set platform true to return all platforms.
//   - Set both to true to return all category and platform tags.
//
// Setting both to false will return an empty JSON response.
func TagsAPI(c echo.Context, db *sql.DB, category, platform bool) error {
	// Try to get cached results first
	if i, found := cachedTags(category, platform); found {
		return c.JSON(http.StatusOK, i)
	}

	items := tags.List()
	infos := tags.Infos()
	if len(items) == 0 || !category && !platform {
		return c.JSON(http.StatusOK, []tagAPI{})
	}

	results := make([]tagAPI, 0, len(items))
	for _, tag := range items {
		slug := tag.String()
		title := tags.NameByURI(slug)
		switch {
		case category && !tags.IsCategory(slug):
			continue
		case platform && !tags.IsPlatform(slug):
			continue
		default:
			// return all tags
		}
		ctx := context.Background()
		var byteSum int64
		var count int64
		if category {
			count, _ = model.CategoryCount(ctx, db, slug)
			byteSum, _ = model.CategoryByteSum(ctx, db, slug)
		}
		if platform {
			c, _ := model.PlatformCount(ctx, db, slug)
			count = c + count
			b, _ := model.PlatformByteSum(ctx, db, slug)
			byteSum = b + byteSum
		}
		result := tagAPI{
			ID:          int(tag),
			Name:        slug,
			Description: infos[tag],
			Title:       title,
			Stats: totalsAPI{
				TotalFiles:     count,
				TotalSize:      helper.ByteCount(byteSum),
				TotalSizeBytes: byteSum,
			},
			URLs: enityURLs{
				API:   APIBase + "/files/" + slug,
				HTML3: "/html3/" + slug,
				HTML:  "/files/" + slug,
			},
		}
		results = append(results, result)
	}

	// Cache the results before returning
	tagsCache(category, platform, results)

	return c.JSON(http.StatusOK, results)
}

// CategoryAPI returns a list of files from any category tag.
func CategoryAPI(c echo.Context, db *sql.DB, sl *slog.Logger) error {
	name := c.Param("category")
	if name == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Category parameter is required",
		})
	}
	if !tags.IsCategory(name) {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Category is not known",
		})
	}
	return TagAPI(c, db, sl, name)
}

// PlatformAPI returns a list of files from any category tag.
func PlatformAPI(c echo.Context, db *sql.DB, sl *slog.Logger) error {
	name := c.Param("platform")
	if name == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Platform parameter is required",
		})
	}
	if !tags.IsPlatform(name) {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Platform is not known",
		})
	}
	return TagAPI(c, db, sl, name)
}

// TagAPI returns a list of files from any category or platform tag.
func TagAPI(c echo.Context, db *sql.DB, sl *slog.Logger, name string) error { //nolint:funlen
	const msg = "get files by tag"
	if err := panics.EchoContextDS(c, db, sl); err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}
	if name == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "TagAPI param string is missing",
		})
	}
	category, platform := tags.IsCategory(name), tags.IsPlatform(name)
	if !category && !platform {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "TagAPI param string is not a known category or platform",
		})
	}

	var fs models.FileSlice
	var err error
	var byteSum int64
	var count int64
	ctx := context.Background()
	order := html3.PublAsc

	if category {
		fs, err = order.ByCategory(ctx, db, 0, 0, name)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "Failed to query category files",
			})
		}
		count, err = model.CategoryCount(ctx, db, name)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "Failed to count category files",
			})
		}
		byteSum, err = model.CategoryByteSum(ctx, db, name)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "Failed to calculate platform file sizes",
			})
		}
	}
	if platform {
		fs, err = order.ByPlatform(ctx, db, 0, 0, name)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "Failed to query platform files",
			})
		}
		count, err = model.PlatformCount(ctx, db, name)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "Failed to count platform files",
			})
		}
		byteSum, err = model.PlatformByteSum(ctx, db, name)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "Failed to calculate platform file sizes",
			})
		}
	}

	result := artifactSummaries(fs)
	response := artifactsAPI{
		Files: result,
		Stats: totalsAPI{
			TotalFiles:     count,
			TotalSize:      helper.ByteCount(byteSum),
			TotalSizeBytes: byteSum,
		},
	}

	return c.JSON(http.StatusOK, response)
}

func datePublished(record *models.File) publishedAPI {
	dp := publishedAPI{ //nolint:exhaustruct
	}
	if record.DateIssuedYear.Valid && record.DateIssuedMonth.Valid && record.DateIssuedDay.Valid {
		dp.Year = record.DateIssuedYear.Int16
		dp.Month = record.DateIssuedMonth.Int16
		dp.Day = record.DateIssuedDay.Int16
		return dp
	}
	if record.DateIssuedYear.Valid && record.DateIssuedMonth.Valid {
		dp.Year = record.DateIssuedYear.Int16
		dp.Month = record.DateIssuedMonth.Int16
		return dp
	}
	if record.DateIssuedYear.Valid {
		dp.Year = record.DateIssuedYear.Int16
		return dp
	}
	return dp
}

func artifact(art *models.File) ArtifactAPI {
	return ArtifactAPI{
		Summary: artifactSummary(art),
		FileMeta: filemetaAPI{
			Checksum:     filerecord.Checksum(art),
			LastModified: timedTimer(art.FileLastModified),
			LastModAgo:   filerecord.LastModificationAgo(art),
			MimeType:     filerecord.Magic(art),
		},
		ArtMeta: artmetaAPI{
			Comment:   simple.CleanHTML(filerecord.Comment(art)),
			Title:     filerecord.FirstHeader(art),
			Releasers: releasersAPI(art),
		},
		Relationships: relationshipsAPI(art),
	}
}

func releasersAPI(art *models.File) []releaserAPI {
	n1, n2 := filerecord.ReleaserPair(art)
	u1, u2 := helper.Slug(n1), helper.Slug(n2)
	const size = 2
	pair := make([]releaserAPI, size)
	pair[0] = releaserAPI{
		ID:   u1,
		Name: releaser.Link(u1),
		API:  APIBase + "/releaser/" + u1,
		HTML: "/g/" + u1,
	}
	pair[1] = releaserAPI{
		ID:   u2,
		Name: releaser.Link(u2),
		API:  APIBase + "/releaser/" + u2,
		HTML: u2,
	}
	return pair
}

func relationshipsAPI(art *models.File) []relationAPI {
	results := []relationAPI{}
	if r := relationsAPI(art); len(r) > 0 {
		results = append(results, r...)
	}
	if r := linksAPI(art); len(r) > 0 {
		results = append(results, r...)
	}
	if s := filerecord.IdenficationDZ(art); s != "" {
		r := relationAPI{
			Link: "https://demozoo.org/productions/" + s + "/",
			Desc: "Demozoo production",
		}
		results = append(results, r)
	}
	if s := filerecord.IdenficationPouet(art); s != "" {
		r := relationAPI{
			Link: "https://www.pouet.net/prod.php?which=" + s + "/",
			Desc: "Pouet production",
		}
		results = append(results, r)
	}
	if s := filerecord.Idenfication16C(art); s != "" {
		r := relationAPI{
			Link: "https://16colo.rs/" + s + "/",
			Desc: "16colors link",
		}
		results = append(results, r)
	}
	if s := filerecord.IdenficationYT(art); s != "" {
		r := relationAPI{
			Link: "https://www.youtube.com/watch?v=" + s + "/",
			Desc: "YouTube video",
		}
		results = append(results, r)
	}
	if s := filerecord.IdenficationGitHub(art); s != "" {
		r := relationAPI{
			Link: "https://github.com/" + s + "/",
			Desc: "GitHub repository",
		}
		results = append(results, r)
	}
	return results
}

// relationsAPI returns the list of relationships for the file record.
func relationsAPI(art *models.File) []relationAPI {
	if art == nil {
		return nil
	}
	rels := art.ListRelations.String
	if rels == "" {
		return nil
	}
	links := strings.Split(rels, "|")
	if len(links) == 0 {
		return nil
	}
	const expected = 2
	const route = "https://defacto2.net/f/"
	results := []relationAPI{}
	for link := range slices.Values(links) {
		s := strings.Split(link, ";")
		if len(s) != expected {
			continue
		}
		name, href := s[0], s[1]
		id := helper.DeObfuscate(href)
		if invalidID := id == href; invalidID {
			continue
		}
		if !strings.HasPrefix(href, route) {
			href = route + href + "/"
		}
		result := relationAPI{
			Link: href,
			Desc: name,
		}
		results = append(results, result)
	}
	return results
}

// Websites returns the list of links for the file record.
func linksAPI(art *models.File) []relationAPI {
	if art == nil {
		return nil
	}
	lls := art.ListLinks.String
	if lls == "" {
		return nil
	}
	links := strings.Split(lls, "|")
	if len(links) == 0 {
		return nil
	}
	results := []relationAPI{}
	const expected = 2
	for link := range slices.Values(links) {
		s := strings.Split(link, ";")
		if len(s) != expected {
			continue
		}
		name, href := s[0], s[1]
		// Generally a stored URL will not include the protocol,
		// and will need to be prefixed with "https://".
		// There are some exceptions for websites that refuse to
		// implement HTTPS, such as http://textfiles.com.
		if !strings.HasPrefix(href, "http") {
			href = "https://" + href
		}
		if val, err := url.Parse(href); err != nil || val.Host == "" {
			continue
		}
		result := relationAPI{
			Link: href,
			Desc: name,
		}
		results = append(results, result)
	}
	return results
}

func timedTimer(t null.Time) *time.Time {
	var tt *time.Time
	if t.Valid {
		t := t.Time
		tt = &t
		return tt
	}
	return nil
}

// artifactSummaries transforms database records to the API format.
func artifactSummaries(fs []*models.File) []artifactAPI {
	result := make([]artifactAPI, len(fs))
	for i, art := range fs {
		result[i] = artifactSummary(art)
	}
	return result
}

func artifactSummary(art *models.File) artifactAPI {
	// Handle postedDate using Createdat field
	var postedDate *time.Time
	if art.Createdat.Valid {
		t := art.Createdat.Time
		postedDate = &t
	}

	fileRecord := &models.File{ //nolint:exhaustruct
		Filename:       art.Filename,
		Section:        art.Section,
		Platform:       art.Platform,
		GroupBrandBy:   art.GroupBrandBy,
		GroupBrandFor:  art.GroupBrandFor,
		RecordTitle:    art.RecordTitle,
		DateIssuedYear: art.DateIssuedYear,
	}

	// Get tags for the file
	category := filerecord.TagCategory(art)
	platform := filerecord.TagProgram(art)

	// Get humanized description for the tags
	categoryTag := tags.TagByURI(category)
	platformTag := tags.TagByURI(platform)
	humanized := tags.Humanize(platformTag, categoryTag)

	return artifactAPI{
		UUID:          filerecord.UnID(art),
		Filename:      art.Filename.String,
		DatePublished: datePublished(art),
		PostedDate:    postedDate,
		Size: struct {
			Formatted string `json:"formatted"`
			Bytes     int64  `json:"bytes"`
		}{
			Formatted: helper.ByteCount(art.Filesize.Int64),
			Bytes:     art.Filesize.Int64,
		},
		Description: filerecord.Description(fileRecord),
		Tags: struct {
			Category    string `json:"category"`
			Platform    string `json:"platform"`
			Description string `json:"description"`
		}{
			Category:    category,
			Platform:    platform,
			Description: humanized,
		},
		URLs: urlAPI{
			API:       APIBase + "/artifact/" + helper.ObfuscateID(art.ID),
			Download:  "/d/" + helper.ObfuscateID(art.ID),
			HTML:      "/f/" + helper.ObfuscateID(art.ID),
			Thumbnail: "/public/image/thumb/" + art.UUID.String,
		},
	}
}

// TerritoriesAPI returns all territories in the North American Numbering Plan (NANP).
func TerritoriesAPI(c echo.Context) error {
	territories := areacode.Territories()
	if len(territories) == 0 {
		return c.JSON(http.StatusOK, []territoryAPI{})
	}

	result := make([]territoryAPI, 0, len(territories))
	for _, t := range territories {
		areaCodes := make([]int, 0, len(t.AreaCodes))
		for _, ac := range t.AreaCodes {
			areaCodes = append(areaCodes, int(ac))
		}

		result = append(result, territoryAPI{
			Name:         t.Name,
			Abbreviation: string(t.Abbreviation),
			AreaCodes:    areaCodes,
		})
	}

	return c.JSON(http.StatusOK, result)
}

// TerritoryAPI returns a specific territory by its abbreviation.
func TerritoryAPI(c echo.Context) error {
	abbr := c.Param("abbr")
	const twoChrs = 2
	if len(abbr) != twoChrs {
		return c.JSON(http.StatusBadRequest, "abbreviation must be 2 characters")
	}

	territory := areacode.TerritoryByAbbr(areacode.Abbreviation(abbr))
	if territory.Name == "" {
		return c.JSON(http.StatusNotFound, "territory not found")
	}
	areaCodes := make([]int, 0, len(territory.AreaCodes))
	for _, ac := range territory.AreaCodes {
		areaCodes = append(areaCodes, int(ac))
	}

	result := territoryAPI{
		Name:         territory.Name,
		Abbreviation: string(territory.Abbreviation),
		AreaCodes:    areaCodes,
	}

	return c.JSON(http.StatusOK, result)
}

// WebsitesAPI returns all websites from the website page.
func WebsitesAPI(c echo.Context) error {
	list := List()
	if len(list) == 0 {
		return c.JSON(http.StatusOK, map[string]any{
			"websites": []websiteAPI{},
		})
	}
	var result []websiteAPI
	for _, category := range list {
		for _, site := range category.Sites {
			result = append(result, websiteAPI{
				Title:    site.Title,
				URL:      site.URL,
				Info:     site.Info,
				Category: category.Name,
			})
		}
	}
	sort.Slice(result, func(i, j int) bool {
		si := result[i].URL
		sj := result[j].URL
		return strings.ToLower(si) < strings.ToLower(sj)
	})
	return c.JSON(http.StatusOK, map[string]any{
		"websites": result,
		"count":    len(result),
	})
}

// DemozooAPI returns a list of all groups with their Demozoo IDs.
func DemozooAPI(c echo.Context) error {
	groups := demozoo.FindAll()
	if len(groups) == 0 {
		return c.JSON(http.StatusOK, map[string]any{
			"groups": []map[string]any{},
		})
	}
	result := make([]map[string]any, 0, len(groups))
	for uri, id := range groups {
		result = append(result, map[string]any{
			"uri": string(uri),
			"id":  int(id),
			"url": fmt.Sprintf("https://demozoo.org/groups/%d/", id),
		})
	}
	sort.Slice(result, func(i, j int) bool {
		si, ok := result[i]["uri"].(string)
		if !ok {
			return false
		}
		sj, ok := result[j]["uri"].(string)
		if !ok {
			return false
		}
		return strings.ToLower(si) < strings.ToLower(sj)
	})

	return c.JSON(http.StatusOK, map[string]any{
		"groups": result,
		"count":  len(result),
	})
}
