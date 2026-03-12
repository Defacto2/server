package app

// This file contains the API handlers for areacode data.

import (
	"net/http"
	"strconv"

	"github.com/Defacto2/server/handler/areacode"
	"github.com/labstack/echo/v4"
)

// areacodeAPI represents an area code for API responses.
type areacodeAPI struct {
	Code        int      `json:"code"`
	Territories []string `json:"territories"`
	Notes       string   `json:"notes,omitempty"`
}

// territoryAPI represents a territory for API responses.
type territoryAPI struct {
	Name         string `json:"name"`
	Abbreviation string `json:"abbreviation"`
	AreaCodes    []int  `json:"areaCodes"`
}

// GetAllAreacodes returns all North American Numbering Plan area codes.
// @Summary Get all area codes
// @Description Get a list of all NANP area codes with their associated territories
// @Tags areacodes
// @Produce json
// @Success 200 {array} areacodeAPI
// @Router /api/areacodes [get].
func GetAllAreacodes(c echo.Context) error {
	codes := areacode.AreaCodes()
	if len(codes) == 0 {
		return c.JSON(http.StatusOK, []areacodeAPI{})
	}

	result := make([]areacodeAPI, 0, len(codes))
	for _, code := range codes {
		territories := areacode.TerritoryByCode(code)
		tNames := make([]string, 0, len(territories))
		for _, t := range territories {
			tNames = append(tNames, t.Name)
		}

		apiCode := areacodeAPI{
			Code:        int(code),
			Territories: tNames,
			Notes:       "",
		}

		if note, ok := areacode.Notes()[code]; ok {
			apiCode.Notes = note
		}

		result = append(result, apiCode)
	}

	return c.JSON(http.StatusOK, result)
}

// GetAreacodeByCode returns details for a specific area code.
// @Summary Get area code by code
// @Description Get details for a specific NANP area code
// @Tags areacodes
// @Produce json
// @Param code path int true "Area code" format(int32)
// @Success 200 {object} areacodeAPI
// @Failure 400 {object} string
// @Failure 404 {object} string
// @Router /api/areacodes/{code} [get].
func GetAreacodeByCode(c echo.Context) error {
	codeStr := c.Param("code")
	if codeStr == "" {
		return c.JSON(http.StatusBadRequest, "area code parameter is required")
	}

	code, err := strconv.Atoi(codeStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, "invalid area code format")
	}

	nanCode := areacode.NAN(code)
	if !nanCode.Valid() {
		return c.JSON(http.StatusNotFound, "area code not found")
	}

	territories := areacode.TerritoryByCode(nanCode)
	tNames := make([]string, 0, len(territories))
	for _, t := range territories {
		tNames = append(tNames, t.Name)
	}

	result := areacodeAPI{
		Code:        code,
		Territories: tNames,
		Notes:       "",
	}

	if note, ok := areacode.Notes()[nanCode]; ok {
		result.Notes = note
	}

	return c.JSON(http.StatusOK, result)
}

// GetTerritories returns all territories in the North American Numbering Plan.
// @Summary Get all territories
// @Description Get a list of all territories (states, provinces) with their area codes
// @Tags areacodes
// @Produce json
// @Success 200 {array} territoryAPI
// @Router /api/areacodes/territories [get].
func GetTerritories(c echo.Context) error {
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

// GetTerritoryByAbbr returns a specific territory by its abbreviation.
// @Summary Get territory by abbreviation
// @Description Get territory details by two-letter abbreviation
// @Tags areacodes
// @Produce json
// @Param abbr path string true "Territory abbreviation" minlength(2) maxlength(2)
// @Success 200 {object} territoryAPI
// @Failure 400 {object} string
// @Failure 404 {object} string
// @Router /api/areacodes/territories/{abbr} [get].
func GetTerritoryByAbbr(c echo.Context) error {
	abbr := c.Param("abbr")
	const territoryAbbreviationLength = 2
	if len(abbr) != territoryAbbreviationLength {
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

// SearchAreacodes searches for area codes or territories by query.
// @Summary Search areacodes and territories
// @Description Search for area codes, territories, or abbreviations
// @Tags areacodes
// @Produce json
// @Param query path string true "Search query"
// @Success 200 {object} object{areacodes=[]areacodeAPI,territories=[]territoryAPI}
// @Failure 400 {object} string
// @Router /api/areacodes/search/{query} [get].
func SearchAreacodes(c echo.Context) error {
	query := c.Param("query")
	if query == "" {
		return c.JSON(http.StatusBadRequest, "search query is required")
	}

	// Try to parse as area code first
	if code, err := strconv.Atoi(query); err == nil {
		nanCode := areacode.NAN(code)
		if nanCode.Valid() {
			territories := areacode.TerritoryByCode(nanCode)
			tNames := make([]string, 0, len(territories))
			for _, t := range territories {
				tNames = append(tNames, t.Name)
			}

			result := areacodeAPI{
				Code:        code,
				Territories: tNames,
				Notes:       "",
			}

			if note, ok := areacode.Notes()[nanCode]; ok {
				result.Notes = note
			}

			return c.JSON(http.StatusOK, map[string]any{
				"areacodes":   []areacodeAPI{result},
				"territories": []territoryAPI{},
			})
		}
	}

	// Try territory lookup
	territories := areacode.Lookup(query)
	if len(territories) > 0 {
		resultTerrs := make([]territoryAPI, 0, len(territories))
		for _, t := range territories {
			areaCodes := make([]int, 0, len(t.AreaCodes))
			for _, ac := range t.AreaCodes {
				areaCodes = append(areaCodes, int(ac))
			}

			resultTerrs = append(resultTerrs, territoryAPI{
				Name:         t.Name,
				Abbreviation: string(t.Abbreviation),
				AreaCodes:    areaCodes,
			})
		}

		return c.JSON(http.StatusOK, map[string]any{
			"areacodes":   []areacodeAPI{},
			"territories": resultTerrs,
		})
	}

	return c.JSON(http.StatusOK, map[string]any{
		"areacodes":   []areacodeAPI{},
		"territories": []territoryAPI{},
	})
}
