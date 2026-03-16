// Package app provides the application handlers for the web server.
// This file contains tests for the areacode API handlers.

package app_test

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Defacto2/server/handler/app"
	_ "github.com/jackc/pgx/v5"
	"github.com/labstack/echo/v4"
	"github.com/nalgeon/be"
)

const (
	contentTypeJSON = "application/json"
	dataSourceName  = "postgres://root:example@localhost:5432/defacto2_ps?sslmode=disable" //nolint:gosec
	driverName      = "pgx"
)

func BenchmarkApiMarkup(b *testing.B) {
	html := `<div class="content">
		<p class="lead">This is a <strong>test</strong> with <a href="https://example.com" class="link" id="test">links</a> and <span style="color: red;">formatting</span>.</p>
		<p>Another paragraph with <a name="anchor">anchor</a> and <data-info="test">data attributes</data-info>.</p>
	</div>`

	b.Run("", func(b *testing.B) {
		for range b.N {
			app.APIMarkup(html)
		}
	})
}

// BenchmarkCategoriesAPIWithRealStats benchmarks the CategoriesAPI with realistic stats calculation.
func BenchmarkCategoriesAPIWithRealStats(b *testing.B) {
	e := echo.New()
	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, api+"/categories", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Create database connection using default credentials
	db, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		b.Logf("Could not create database connection: %v", err)
		b.Skipf("Could not create database connection: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			panic(err)
		}
	}()
	if err := db.PingContext(context.Background()); err != nil {
		b.Logf("Could not ping database: %v", err)
		b.Skipf("Could not ping database: %v", err)
	}
	b.ResetTimer()
	for b.Loop() {
		_ = app.CategoriesAPI(c, db)
		rec.Body.Reset()
	}
}

// BenchmarkPlatformsAPIWithRealStats benchmarks the PlatformsAPI with realistic stats calculation.
func BenchmarkPlatformsAPIWithRealStats(b *testing.B) {
	e := echo.New()
	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Create database connection using default credentials
	db, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		b.Logf("Could not create database connection: %v", err)
		b.Skipf("Could not create database connection: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			panic(err)
		}
	}()
	if err := db.PingContext(context.Background()); err != nil {
		b.Logf("Could not ping database: %v", err)
		b.Skipf("Could not ping database: %v", err)
	}

	b.ResetTimer()
	for b.Loop() {
		_ = app.PlatformsAPI(c, db)
		rec.Body.Reset()
	}
}

func TestApiMarkup(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Preserves anchor tags with href",
			input:    `<p>Test <a href="https://example.com" class="link">link</a> here</p>`,
			expected: `<p>Test <a href="https://example.com">link</a> here</p>`,
		},
		{
			name:     "Removes anchor tags without href",
			input:    `<p>Test <a name="anchor">link</a> here</p>`,
			expected: `<p>Test link here</p>`,
		},
		{
			name:     "Removes class attributes",
			input:    `<p class="test">Content</p>`,
			expected: `<p>Content</p>`,
		},
		{
			name:     "Removes style attributes",
			input:    `<p style="color: red;">Content</p>`,
			expected: `<p>Content</p>`,
		},
		{
			name:     "Removes id attributes",
			input:    `<p id="test">Content</p>`,
			expected: `<p>Content</p>`,
		},
		{
			name:     "Removes title attributes",
			input:    `<p title="tooltip">Content</p>`,
			expected: `<p>Content</p>`,
		},
		{
			name:     "Removes data attributes",
			input:    `<p data-test="value">Content</p>`,
			expected: `<p>Content</p>`,
		},
		{
			name:     "Preserves semantic HTML",
			input:    `<p>Test <strong>bold</strong> and <em>italic</em> text</p>`,
			expected: `<p>Test <strong>bold</strong> and <em>italic</em> text</p>`,
		},
		{
			name:     "Handles complex anchor tags",
			input:    `<a href="https://example.com" class="link" id="test" title="tooltip" data-info="test">Complex Link</a>`,
			expected: `<a href="https://example.com">Complex Link</a>`,
		},
		{
			name:     "Handles multiple anchor tags",
			input:    `<p><a href="https://example1.com">Link 1</a> and <a href="https://example2.com">Link 2</a></p>`,
			expected: `<p><a href="https://example1.com">Link 1</a> and <a href="https://example2.com">Link 2</a></p>`,
		},
		{
			name:     "Removes empty tags",
			input:    `<p><span> </span>Content</p>`,
			expected: `<p>Content</p>`,
		},
		{
			name:     "Removes style attribute from h1",
			input:    `<h1 style="color: blue;">Welcome</h1>`,
			expected: `<h1>Welcome</h1>`,
		},
		{
			name: "Remove newlines",
			input: `<div class="content" id="main">` + "\n\t\t" +
				`<h1 style="color: blue;">Welcome</h1>` + "\n" + `</div>`,
			expected: `<div><h1>Welcome</h1></div>`,
		},
		{
			name:     "Removes various empty elements",
			input:    `<p><span></span>Text <div></div> more <i></i> text</p>`,
			expected: `<p>Text more text</p>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := app.APIMarkup(tt.input)
			be.Equal(t, result, tt.expected)
		})
	}
}

func TestGetAllAreacodes(t *testing.T) {
	t.Parallel()
	e := echo.New()
	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := app.AreacodesAPI(c)
	be.Equal(t, err, nil)
	be.Equal(t, http.StatusOK, rec.Code)
	be.True(t, len(rec.Body.String()) > 0)
	be.True(t, rec.Header().Get("Content-Type") == contentTypeJSON)
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
			t.Parallel()
			// Setup
			e := echo.New()
			req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("code")
			c.SetParamValues(tt.code)

			// Test
			err := app.AreaCodeAPI(c)
			be.Equal(t, err, nil)
			be.Equal(t, tt.expectStatus, rec.Code)
			be.True(t, len(rec.Body.String()) > 0)
			be.True(t, rec.Header().Get("Content-Type") == contentTypeJSON)
			be.True(t, len(rec.Body.String()) > 0)
			be.True(t, strings.Contains(rec.Body.String(), tt.expectContain))
		})
	}
}

func TestGetTerritories(t *testing.T) {
	t.Parallel()
	// Setup
	e := echo.New()
	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Test
	err := app.TerritoriesAPI(c)
	be.Equal(t, err, nil)
	be.Equal(t, http.StatusOK, rec.Code)
	be.True(t, len(rec.Body.String()) > 0)
	be.True(t, rec.Header().Get("Content-Type") == contentTypeJSON)
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
			t.Parallel()
			// Setup
			e := echo.New()
			req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("abbr")
			c.SetParamValues(tt.abbr)

			// Test
			err := app.TerritoryAPI(c)
			be.Equal(t, err, nil)
			be.Equal(t, tt.expectStatus, rec.Code)
			be.True(t, len(rec.Body.String()) > 0)
			be.True(t, rec.Header().Get("Content-Type") == contentTypeJSON)
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
			t.Parallel()
			// Setup
			e := echo.New()
			req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("query")
			c.SetParamValues(tt.query)

			// Test
			err := app.AreacodeSearchAPI(c)
			be.Equal(t, err, nil)
			be.Equal(t, tt.expectStatus, rec.Code)
			be.True(t, len(rec.Body.String()) > 0)
			be.True(t, rec.Header().Get("Content-Type") == contentTypeJSON)
			be.True(t, strings.Contains(rec.Body.String(), tt.expectContain))
		})
	}
}
