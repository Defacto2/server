package app

import (
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/Defacto2/server/handler/app/internal/simple"
	"github.com/Defacto2/server/handler/areacode"
	"github.com/Defacto2/server/internal/tags"
	"github.com/labstack/echo/v4"
)

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
	Description string `json:"description,omitempty"`
	URI         string `json:"uri,omitempty"`
	URLs        struct {
		API   string `json:"api,omitempty"`
		HTML3 string `json:"html3,omitempty"`
		HTML  string `json:"html,omitempty"`
	} `json:"urls"`
	Count int `json:"count,omitempty"`
}

// territoryAPI represents a territory for API responses.
type territoryAPI struct {
	Name         string `json:"name"`
	Abbreviation string `json:"abbreviation"`
	AreaCodes    []int  `json:"areaCodes"`
}

const (
	yearsInDecade = 9
)

// ApiMarkup removes CSS classes and attributes from HTML for API responses.
// Keeps semantic HTML tags but removes presentation-specific markup.
func ApiMarkup(src string) string {
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
	re = regexp.MustCompile(`<([a-z]+)\s+>`)
	src = re.ReplaceAllString(src, "<$1>")

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

	startYear := decade
	endYear := decade + yearsInDecade

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
	m.ContentHTML = ApiMarkup(m.Content)
	m.Content = simple.CleanHTML(m.Content)
	m.Lead = ApiMarkup(m.Lead)

	// Clean any links
	if len(m.List) > 0 {
		for i := range m.List {
			m.List[i].LinkTitle = ApiMarkup(m.List[i].LinkTitle)
			m.List[i].SubTitle = ApiMarkup(m.List[i].SubTitle)
		}
	}

	return m
}

// CategoriesAPI returns all categories.
func CategoriesAPI(c echo.Context) error {
	return TagsAPI(c, true, false)
}

// PlatformsAPI returns all platforms.
func PlatformsAPI(c echo.Context) error {
	return TagsAPI(c, false, true)
}

// TagsAPI returns artifact tags.
//
//   - Set categories true to return all categories.
//   - Set platform true to return all platforms.
//   - Set both to true to return all category and platform tags.
//
// Setting both to false will return an empty JSON response.
func TagsAPI(c echo.Context, category, platform bool) error {
	names := tags.Names()
	infos := tags.Infos()
	uris := tags.URIs()
	if len(names) == 0 || !category && !platform {
		return c.JSON(http.StatusOK, []tagAPI{})
	}

	results := make([]tagAPI, 0, len(names))
	for tag, name := range names {
		switch {
		case category && !tags.IsCategory(name):
			continue
		case platform && !tags.IsPlatform(name):
			continue
		default:
			// return all tags
		}

		desc := infos[tag]
		uri := uris[tag]

		linkHtm3 := "/html3/" + uri
		linkHtml := "/files/" + uri
		linkApi := "/api/files/" + uri
		result := tagAPI{
			ID:          int(tag),
			Name:        name,
			Description: desc,
			URI:         uri,
			Count:       0, // TODO: Will be populated later if needed
			URLs: struct {
				API   string `json:"api,omitempty"`
				HTML3 string `json:"html3,omitempty"`
				HTML  string `json:"html,omitempty"`
			}{
				API:   linkApi,
				HTML3: linkHtm3,
				HTML:  linkHtml,
			},
		}
		results = append(results, result)
	}

	return c.JSON(http.StatusOK, results)
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
