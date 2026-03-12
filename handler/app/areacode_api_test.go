// Package app provides the application handlers for the web server.
// This file contains tests for the areacode API handlers.

package app_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Defacto2/server/handler/app"
	"github.com/labstack/echo/v4"
	"github.com/nalgeon/be"
)

func TestGetAllAreacodes(t *testing.T) {
	t.Parallel()
	// Setup
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/areacodes", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Test
	err := app.GetAllAreacodes(c)
	be.Equal(t, err, nil)
	be.Equal(t, http.StatusOK, rec.Code)
	be.True(t, len(rec.Body.String()) > 0)
	be.True(t, rec.Header().Get("Content-Type") == "application/json")
}

func TestGetAreacodeByCode(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name          string
		code          string
		expectStatus  int
		expectContain string
	}{
		{
			name:          "valid area code",
			code:          "212",
			expectStatus:  http.StatusOK,
			expectContain: "New York",
		},
		{
			name:          "invalid area code",
			code:          "999",
			expectStatus:  http.StatusNotFound,
			expectContain: "area code not found",
		},
		{
			name:          "empty code",
			code:          "",
			expectStatus:  http.StatusBadRequest,
			expectContain: "area code parameter is required",
		},
		{
			name:          "non-numeric code",
			code:          "abc",
			expectStatus:  http.StatusBadRequest,
			expectContain: "invalid area code format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/api/areacodes/"+tt.code, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("code")
			c.SetParamValues(tt.code)

			// Test
			err := app.GetAreacodeByCode(c)
			be.Equal(t, err, nil)
			be.Equal(t, tt.expectStatus, rec.Code)
			be.True(t, len(rec.Body.String()) > 0)
			be.True(t, rec.Header().Get("Content-Type") == "application/json")
			be.True(t, len(rec.Body.String()) > 0)
			be.True(t, strings.Contains(rec.Body.String(), tt.expectContain))
		})
	}
}

func TestGetTerritories(t *testing.T) {
	t.Parallel()
	// Setup
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/areacodes/territories", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Test
	err := app.GetTerritories(c)
	be.Equal(t, err, nil)
	be.Equal(t, http.StatusOK, rec.Code)
	be.True(t, len(rec.Body.String()) > 0)
	be.True(t, rec.Header().Get("Content-Type") == "application/json")
}

func TestGetTerritoryByAbbr(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name          string
		abbr          string
		expectStatus  int
		expectContain string
	}{
		{
			name:          "valid abbreviation",
			abbr:          "CA",
			expectStatus:  http.StatusOK,
			expectContain: "California",
		},
		{
			name:          "invalid abbreviation",
			abbr:          "XX",
			expectStatus:  http.StatusNotFound,
			expectContain: "territory not found",
		},
		{
			name:          "short abbreviation",
			abbr:          "C",
			expectStatus:  http.StatusBadRequest,
			expectContain: "abbreviation must be 2 characters",
		},
		{
			name:          "long abbreviation",
			abbr:          "CAL",
			expectStatus:  http.StatusBadRequest,
			expectContain: "abbreviation must be 2 characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/api/areacodes/territories/"+tt.abbr, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("abbr")
			c.SetParamValues(tt.abbr)

			// Test
			err := app.GetTerritoryByAbbr(c)
			be.Equal(t, err, nil)
			be.Equal(t, tt.expectStatus, rec.Code)
			be.True(t, len(rec.Body.String()) > 0)
			be.True(t, rec.Header().Get("Content-Type") == "application/json")
			be.True(t, strings.Contains(rec.Body.String(), tt.expectContain))
		})
	}
}

func TestSearchAreacodes(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name          string
		query         string
		expectStatus  int
		expectContain string
	}{
		{
			name:          "search by area code",
			query:         "212",
			expectStatus:  http.StatusOK,
			expectContain: "areacodes",
		},
		{
			name:          "search by state name",
			query:         "california",
			expectStatus:  http.StatusOK,
			expectContain: "territories",
		},
		{
			name:          "search by abbreviation",
			query:         "ny",
			expectStatus:  http.StatusOK,
			expectContain: "territories",
		},
		{
			name:          "empty query",
			query:         "",
			expectStatus:  http.StatusBadRequest,
			expectContain: "search query is required",
		},
		{
			name:          "not found",
			query:         "xyz123",
			expectStatus:  http.StatusOK,
			expectContain: "[]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/api/areacodes/search/"+tt.query, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("query")
			c.SetParamValues(tt.query)

			// Test
			err := app.SearchAreacodes(c)
			be.Equal(t, err, nil)
			be.Equal(t, tt.expectStatus, rec.Code)
			be.True(t, len(rec.Body.String()) > 0)
			be.True(t, rec.Header().Get("Content-Type") == "application/json")
			be.True(t, strings.Contains(rec.Body.String(), tt.expectContain))
		})
	}
}