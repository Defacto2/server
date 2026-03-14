package app

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/Defacto2/helper"
	"github.com/Defacto2/releaser"
	"github.com/Defacto2/server/handler/app/internal/filerecord"
	"github.com/Defacto2/server/handler/app/internal/simple"
	"github.com/Defacto2/server/handler/areacode"
	"github.com/Defacto2/server/internal/panics"
	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/Defacto2/server/internal/tags"
	"github.com/Defacto2/server/model"
	"github.com/Defacto2/server/model/html3"
	"github.com/labstack/echo/v4"
)

// ArtifactSumAPI represents an artifact file summary for API responses.
type ArtifactSumAPI struct {
	ID            int64  `json:"id"`
	Filename      string `json:"filename"`
	DatePublished struct {
		Year  int16 `json:"year,omitempty"`
		Month int16 `json:"month,omitempty"`
		Day   int16 `json:"day,omitempty"`
	} `json:"datePublished"`
	PostedDate *time.Time `json:"postedDate,omitempty"`
	Size       struct {
		Formatted string `json:"formatted"`
		Bytes     int64  `json:"bytes"`
	} `json:"size"`
	Description string `json:"description,omitempty"`
	Tags        struct {
		Category    string `json:"category"`
		Platform    string `json:"platform"`
		Description string `json:"description"`
	} `json:"tags"`
	URLs struct {
		Download  string `json:"download"`
		HTML      string `json:"html"`
		Thumbnail string `json:"thumbnail,omitempty"`
	} `json:"urls"`
}

// SceneEntityAPI represents a releaser or group for API responses.
type SceneEntityAPI struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Title string `json:"title"`
	URLs  struct {
		API   string `json:"api"`
		HTML3 string `json:"html3"`
		HTML  string `json:"html"`
	} `json:"urls"`
	Statistics struct {
		TotalFiles     int64  `json:"totalFiles"`
		TotalSize      string `json:"totalSize"`
		TotalSizeBytes int64  `json:"totalSizeBytes"`
	} `json:"statistics"`
}

type artifactsSumStat struct {
	Files []ArtifactSumAPI `json:"files"`
	Stats struct {
		TotalFiles     int64  `json:"totalFiles"`
		TotalSize      string `json:"totalSize"`
		TotalSizeBytes int64  `json:"totalSizeBytes"`
	} `json:"statistics"`
}

// areacodeAPI represents an area code for API responses.
type areacodeAPI struct {
	Code        int      `json:"code"`
	Territories []string `json:"territories"`
	Notes       string   `json:"notes,omitempty"`
}

// tagAPI represents a tag category or tag platform for API responses.
type tagAPI struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
	URLs        struct {
		API   string `json:"api,omitempty"`
		HTML3 string `json:"html3,omitempty"`
		HTML  string `json:"html,omitempty"`
	} `json:"urls"`
	Stats struct {
		TotalFiles     int64  `json:"totalFiles"`
		TotalSize      string `json:"totalSize"`
		TotalSizeBytes int64  `json:"totalSizeBytes"`
	} `json:"statistics"`
}

// territoryAPI represents a territory for API responses.
type territoryAPI struct {
	Name         string `json:"name"`
	Abbreviation string `json:"abbreviation"`
	AreaCodes    []int  `json:"areaCodes"`
}

// APIMarkup removes CSS classes and attributes from HTML for API responses.
// Keeps semantic HTML tags but removes presentation-specific markup.
func APIMarkup(src string) string {
	if src == "" {
		return src
	}

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

	// Try to parse as area code first
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

	// Try a territory lookup
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
	milestones := Collection()
	// Clean milestones for API response
	cleaned := make(Milestones, len(milestones))
	for i, m := range milestones {
		cleaned[i] = milestoneFix(m)
	}
	return c.JSON(http.StatusOK, cleaned)
}

// MilestoneHighlightsAPI returns only highlighted milestones.
func MilestoneHighlightsAPI(c echo.Context) error {
	var result Milestones
	for _, m := range Collection() {
		if m.Highlight {
			result = append(result, milestoneFix(m))
		}
	}

	if len(result) == 0 {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "No highlighted milestones found",
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
			result = append(result, milestoneFix(m))
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
			result = append(result, milestoneFix(m))
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
			result = append(result, milestoneFix(m))
		}
	}

	if len(result) == 0 {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "No milestones found for this decade",
		})
	}

	return c.JSON(http.StatusOK, result)
}

// milestoneFix prepares a milestone for API response.
func milestoneFix(m Milestone) Milestone {
	m.ContentHTML = APIMarkup(m.Content)
	m.Content = simple.CleanHTML(m.Content)
	m.Lead = APIMarkup(m.Lead)

	// Clean any links
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
	const limit = 1000                   // Fixed limit per page
	pages := (count + limit - 1) / limit // Ceiling division
	rels := model.Releasers{}
	if err := rels.Limit(ctx, db, model.Alphabetical, limit, page); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to query releasers",
		})
	}
	if len(rels) == 0 {
		return c.JSON(http.StatusOK, map[string]any{
			"releasers":  []SceneEntityAPI{},
			"page":       page,
			"totalPages": pages,
		})
	}

	releasersWithStats := ReleasersAPI(rels)
	return c.JSON(http.StatusOK, map[string]any{
		"releasers":  releasersWithStats,
		"page":       page,
		"totalPages": pages,
	})
}

// MagazinesAPI is the handler for the magazines API endpoint.
func MagazinesAPI(c echo.Context, db *sql.DB, sl *slog.Logger) error {
	const msg = "magazines api"
	if err := panics.EchoContextDS(c, db, sl); err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}

	ctx := context.Background()
	rels := model.Releasers{}
	if err := rels.Magazine(ctx, db); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to query magazines",
		})
	}
	if len(rels) == 0 {
		return c.JSON(http.StatusOK, map[string]any{
			"releasers":  []SceneEntityAPI{},
			"page":       1,
			"totalPages": 1,
		})
	}

	releasersWithStats := ReleasersAPI(rels)
	return c.JSON(http.StatusOK, map[string]any{
		"releasers":  releasersWithStats,
		"page":       1,
		"totalPages": 1,
	})
}

// BBSAPI is the handler for the BBS API endpoint.
func BBSAPI(c echo.Context, db *sql.DB, sl *slog.Logger) error {
	const msg = "bbs api"
	if err := panics.EchoContextDS(c, db, sl); err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}

	ctx := context.Background()
	rels := model.Releasers{}
	if err := rels.BBS(ctx, db, model.Oldest); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to query BBS sites",
		})
	}
	if len(rels) == 0 {
		return c.JSON(http.StatusOK, map[string]any{
			"releasers":  []SceneEntityAPI{},
			"page":       1,
			"totalPages": 1,
		})
	}

	releasersWithStats := ReleasersAPI(rels)
	return c.JSON(http.StatusOK, map[string]any{
		"releasers":  releasersWithStats,
		"page":       1,
		"totalPages": 1,
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
func ReleasersAPI(rels model.Releasers) []SceneEntityAPI {
	results := make([]SceneEntityAPI, 0, len(rels))
	for _, rel := range rels {
		title := releaser.Link(rel.Unique.Name)
		name := releaser.Obfuscate(rel.Unique.Name)
		count := rel.Unique.Count
		bytes := rel.Unique.Bytes
		// As there is no unique ids for releasers, only a unique uri,
		// create stable ID from the url using an obfuscated name
		id := simple.Hash(name)

		result := SceneEntityAPI{
			ID:    id,
			Name:  name,
			Title: title,
			URLs: struct {
				API   string `json:"api"`
				HTML3 string `json:"html3"`
				HTML  string `json:"html"`
			}{
				API:   "/api/releaser/" + name,
				HTML3: "/html3/group/" + name,
				HTML:  "/g/" + name,
			},
			Statistics: struct {
				TotalFiles     int64  `json:"totalFiles"`
				TotalSize      string `json:"totalSize"`
				TotalSizeBytes int64  `json:"totalSizeBytes"`
			}{
				TotalFiles:     int64(count),
				TotalSize:      helper.ByteCount(int64(bytes)),
				TotalSizeBytes: int64(bytes),
			},
		}
		results = append(results, result)
	}
	return results
}

// artifactSum creates an ArtifactSumAPI from a file model.
func artifactSum(f *models.File) ArtifactSumAPI {
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

	artifact := ArtifactSumAPI{
		ID:       f.ID,
		Filename: f.Filename.String,
		DatePublished: struct {
			Year  int16 `json:"year,omitempty"`
			Month int16 `json:"month,omitempty"`
			Day   int16 `json:"day,omitempty"`
		}{
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
		URLs: struct {
			Download  string `json:"download"`
			HTML      string `json:"html"`
			Thumbnail string `json:"thumbnail,omitempty"`
		}{
			Download:  "/d/" + helper.ObfuscateID(f.ID),
			HTML:      "/f/" + helper.ObfuscateID(f.ID),
			Thumbnail: "/public/image/thumb/" + f.UUID.String,
		},
	}
	if f.Createdat.Valid {
		artifact.PostedDate = &f.Createdat.Time
	}

	return artifact
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
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "Releaser not found",
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

	artifacts := make([]ArtifactSumAPI, 0, len(fs))
	for _, f := range fs {
		artifacts = append(artifacts, artifactSum(f))
	}
	return c.JSON(http.StatusOK, map[string]any{
		"group": SceneEntityAPI{
			ID:    simple.Hash(name),
			Name:  name,
			Title: releaser.Link(name),
			URLs: struct {
				API   string `json:"api"`
				HTML3 string `json:"html3"`
				HTML  string `json:"html"`
			}{
				API:   "/api/releaser/" + name,
				HTML3: "/html3/group/" + name,
				HTML:  "/g/" + name,
			},
			Statistics: struct {
				TotalFiles     int64  `json:"totalFiles"`
				TotalSize      string `json:"totalSize"`
				TotalSizeBytes int64  `json:"totalSizeBytes"`
			}{
				TotalFiles:     sum.SumCount.Int64,
				TotalSize:      helper.ByteCount(sum.SumBytes.Int64),
				TotalSizeBytes: sum.SumBytes.Int64,
			},
		},
		"files": artifacts,
	})
}

// TagsAPI returns artifact tags.
//
//   - Set categories true to return all categories.
//   - Set platform true to return all platforms.
//   - Set both to true to return all category and platform tags.
//
// Setting both to false will return an empty JSON response.
func TagsAPI(c echo.Context, db *sql.DB, category, platform bool) error {
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
			Stats: struct {
				TotalFiles     int64  `json:"totalFiles"`
				TotalSize      string `json:"totalSize"`
				TotalSizeBytes int64  `json:"totalSizeBytes"`
			}{
				TotalFiles:     count,
				TotalSize:      helper.ByteCount(byteSum),
				TotalSizeBytes: byteSum,
			},
			URLs: struct {
				API   string `json:"api,omitempty"`
				HTML3 string `json:"html3,omitempty"`
				HTML  string `json:"html,omitempty"`
			}{
				API:   "/api/files/" + slug,
				HTML3: "/html3/" + slug,
				HTML:  "/files/" + slug,
			},
		}
		results = append(results, result)
	}

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

	var records models.FileSlice
	var err error
	var byteSum int64
	var count int64
	ctx := context.Background()
	order := html3.PublAsc

	if category {
		records, err = order.ByCategory(ctx, db, 0, 0, name)
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
		records, err = order.ByPlatform(ctx, db, 0, 0, name)
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

	files := artifactsSum(records)
	response := artifactsSumStat{
		Files: files,
		Stats: struct {
			TotalFiles     int64  `json:"totalFiles"`
			TotalSize      string `json:"totalSize"`
			TotalSizeBytes int64  `json:"totalSizeBytes"`
		}{
			TotalFiles:     count,
			TotalSize:      helper.ByteCount(byteSum),
			TotalSizeBytes: byteSum,
		},
	}

	return c.JSON(http.StatusOK, response)
}

// artifactsSum transforms database records to the API format.
func artifactsSum(records []*models.File) []ArtifactSumAPI {
	files := make([]ArtifactSumAPI, len(records))
	for i, record := range records {
		var datePublished struct {
			Year  int16 `json:"year,omitempty"`
			Month int16 `json:"month,omitempty"`
			Day   int16 `json:"day,omitempty"`
		}
		if record.DateIssuedYear.Valid && record.DateIssuedMonth.Valid && record.DateIssuedDay.Valid {
			datePublished.Year = record.DateIssuedYear.Int16
			datePublished.Month = record.DateIssuedMonth.Int16
			datePublished.Day = record.DateIssuedDay.Int16
		}

		// Handle postedDate using Createdat field
		var postedDate *time.Time
		if record.Createdat.Valid {
			t := record.Createdat.Time
			postedDate = &t
		}

		fileRecord := &models.File{ //nolint:exhaustruct
			Filename:       record.Filename,
			Section:        record.Section,
			Platform:       record.Platform,
			GroupBrandBy:   record.GroupBrandBy,
			GroupBrandFor:  record.GroupBrandFor,
			RecordTitle:    record.RecordTitle,
			DateIssuedYear: record.DateIssuedYear,
		}

		// Get tags for the file
		category := filerecord.TagCategory(record)
		platform := filerecord.TagProgram(record)

		// Get humanized description for the tags
		categoryTag := tags.TagByURI(category)
		platformTag := tags.TagByURI(platform)
		humanized := tags.Humanize(platformTag, categoryTag)

		files[i] = ArtifactSumAPI{
			ID:            record.ID,
			Filename:      record.Filename.String,
			DatePublished: datePublished,
			PostedDate:    postedDate,
			Size: struct {
				Formatted string `json:"formatted"`
				Bytes     int64  `json:"bytes"`
			}{
				Formatted: helper.ByteCount(record.Filesize.Int64),
				Bytes:     record.Filesize.Int64,
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
			URLs: struct {
				Download  string `json:"download"`
				HTML      string `json:"html"`
				Thumbnail string `json:"thumbnail,omitempty"`
			}{
				Download:  "/d/" + helper.ObfuscateID(record.ID),
				HTML:      "/f/" + helper.ObfuscateID(record.ID),
				Thumbnail: "/public/image/thumb/" + record.UUID.String,
			},
		}
	}
	return files
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
