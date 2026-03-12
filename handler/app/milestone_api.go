package app

import (
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
)

// Constants for decade calculations.
const (
	yearsInDecade = 9
)

// CleanHTMLForAPI removes CSS classes and attributes from HTML for API responses.
// Keeps semantic HTML tags but removes presentation-specific markup.
func CleanHTMLForAPI(html string) string {
	if html == "" {
		return html
	}

	// First, remove anchor tags without href attribute (keep content)
	// This must be done BEFORE cleaning other anchor tags
	re := regexp.MustCompile(`<a\b[^>]*>([^<]*)<\/a>`)
	// Only remove anchors that don't contain href=
	html = re.ReplaceAllStringFunc(html, func(match string) string {
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
	html = re.ReplaceAllString(html, `<a href="$2">$3</a>`)

	// Remove specific attributes from all tags (class, style, id, title, data-*)
	// But preserve href attributes on anchor tags
	re = regexp.MustCompile(`(class|style|id|title|data-[a-z-]*)="[^"]*"`)
	html = re.ReplaceAllString(html, "")

	// Remove any remaining empty attributes or whitespace in tags
	re = regexp.MustCompile(`<([a-z]+)\s+>`)
	html = re.ReplaceAllString(html, "<$1>")

	// Clean up multiple spaces (but preserve line breaks for readability)
	re = regexp.MustCompile(`[\t\r\n]+`)
	html = re.ReplaceAllString(html, " ")
	re = regexp.MustCompile(`\s{2,}`)
	html = re.ReplaceAllString(html, " ")

	return strings.TrimSpace(html)
}

// StripHTMLTags removes all HTML tags from content, returning plain text.
func StripHTMLTags(html string) string {
	if html == "" {
		return html
	}

	// First, handle <q> tags specially - convert to quoted text (non-greedy)
	re := regexp.MustCompile(`<q\b[^>]*>(.*?)<\/q>`)
	html = re.ReplaceAllString(html, `"$1"`)

	// Convert common HTML entities to regular characters
	html = strings.ReplaceAll(html, "&amp;", "&")
	html = strings.ReplaceAll(html, "&lt;", "<")
	html = strings.ReplaceAll(html, "&gt;", ">")

	// Remove all HTML tags and replace with single space
	re = regexp.MustCompile(`<[^>]*>`)
	result := re.ReplaceAllString(html, " ")

	// Fix common spacing issues
	// Remove spaces before punctuation
	re = regexp.MustCompile(`\s+([.,;:!?])`)
	result = re.ReplaceAllString(result, "${1}")

	// Remove spaces after opening parentheses and before closing parentheses
	re = regexp.MustCompile(`\(\s+`)
	result = re.ReplaceAllString(result, "(")
	re = regexp.MustCompile(`\s+\)`)
	result = re.ReplaceAllString(result, ")")

	// Add space after punctuation if missing (but not if already there)
	re = regexp.MustCompile(`([.!?])(\w)`)
	result = re.ReplaceAllString(result, "${1} ${2}")

	// Handle &nbsp; by converting to single space (preserves intent without double spacing)
	result = strings.ReplaceAll(result, "&nbsp;", " ")

	// Clean up all multiple spaces
	re = regexp.MustCompile(`[\s\n\r\t]+`)
	result = re.ReplaceAllString(result, " ")

	return strings.TrimSpace(result)
}

// cleanMilestoneForAPI prepares a milestone for API response.
func cleanMilestoneForAPI(m Milestone) Milestone {
	m.Content = CleanHTMLForAPI(m.Content)
	m.ContentPlain = StripHTMLTags(m.Content)
	m.Lead = CleanHTMLForAPI(m.Lead)

	// Clean links if they exist
	if len(m.List) > 0 {
		for i := range m.List {
			m.List[i].LinkTitle = CleanHTMLForAPI(m.List[i].LinkTitle)
			m.List[i].SubTitle = CleanHTMLForAPI(m.List[i].SubTitle)
		}
	}

	return m
}

// GetAllMilestones returns all milestones.
func GetAllMilestones(c echo.Context) error {
	milestones := Collection()
	// Clean milestones for API response
	cleaned := make(Milestones, len(milestones))
	for i, m := range milestones {
		cleaned[i] = cleanMilestoneForAPI(m)
	}
	return c.JSON(http.StatusOK, cleaned)
}

// GetMilestonesByYear returns milestones for a specific year.
func GetMilestonesByYear(c echo.Context) error {
	year, err := strconv.Atoi(c.Param("year"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid year format",
		})
	}

	var result Milestones
	for _, m := range Collection() {
		if m.Year == year {
			result = append(result, cleanMilestoneForAPI(m))
		}
	}

	if len(result) == 0 {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "No milestones found for this year",
		})
	}

	return c.JSON(http.StatusOK, result)
}

// GetMilestonesByYearRange returns milestones within a year range.
func GetMilestonesByYearRange(c echo.Context) error {
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
			result = append(result, cleanMilestoneForAPI(m))
		}
	}

	if len(result) == 0 {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "No milestones found for this year range",
		})
	}

	return c.JSON(http.StatusOK, result)
}

// GetHighlightedMilestones returns only highlighted milestones.
func GetHighlightedMilestones(c echo.Context) error {
	var result Milestones
	for _, m := range Collection() {
		if m.Highlight {
			result = append(result, cleanMilestoneForAPI(m))
		}
	}

	if len(result) == 0 {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "No highlighted milestones found",
		})
	}

	return c.JSON(http.StatusOK, result)
}

// GetMilestonesByDecade returns milestones for a specific decade.
func GetMilestonesByDecade(c echo.Context) error {
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
			result = append(result, cleanMilestoneForAPI(m))
		}
	}

	if len(result) == 0 {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "No milestones found for this decade",
		})
	}

	return c.JSON(http.StatusOK, result)
}
