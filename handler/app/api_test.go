// Package app provides the application handlers for the web server.
// This file contains tests for the areacode API handlers.

package app_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Defacto2/server/handler/app"
	_ "github.com/jackc/pgx/v5"
	"github.com/labstack/echo/v5"
	"github.com/nalgeon/be"
)

const (
	contentTypeJSON = "application/json"
	dataSourceName  = "postgres://root:example@localhost:5432/defacto2_ps?sslmode=disable" //nolint:gosec
	driverName      = "pgx"
)

func newRequest(tb testing.TB, target string) (*httptest.ResponseRecorder, *echo.Context) {
	tb.Helper()

	e := echo.New()
	r := httptest.NewRequestWithContext(tb.Context(), http.MethodGet, target, nil)
	w := httptest.NewRecorder()
	c := e.NewContext(r, w)

	return w, c
}

func clientDo(tb testing.TB, url string) (*http.Response, error) {
	tb.Helper()

	client := http.Client{}
	req, _ := http.NewRequestWithContext(tb.Context(), http.MethodGet, url, nil)
	return client.Do(req)
}

func TestApiMarkup(t *testing.T) {
	t.Parallel()

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
			t.Parallel()
			got := app.APIMarkup(tt.input)
			be.Equal(t, got, tt.expected)
		})
	}
}

func TestGetAllAreacodes(t *testing.T) {
	t.Parallel()

	w, c := newRequest(t, "/")
	err := app.AreacodesAPI(c)
	be.Equal(t, err, nil)
	be.Equal(t, http.StatusOK, w.Code)
	be.True(t, len(w.Body.String()) > 0)
	be.True(t, w.Header().Get("Content-Type") == contentTypeJSON)
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

			w, c := newRequest(t, "/")
			c.PathValues()
			c.SetPathValues(echo.PathValues{
				{Name: "code", Value: tt.code},
			})
			err := app.AreaCodeAPI(c)
			be.Equal(t, err, nil)

			be.Equal(t, w.Code, tt.expectStatus)
			be.True(t, len(w.Body.String()) > 0)
			be.True(t, w.Header().Get("Content-Type") == contentTypeJSON)
			be.True(t, len(w.Body.String()) > 0)
			be.True(t, strings.Contains(w.Body.String(), tt.expectContain))
		})
	}
}

func TestGetTerritories(t *testing.T) {
	t.Parallel()

	w, c := newRequest(t, "/")
	err := app.RegionsAPI(c)
	be.Equal(t, err, nil)

	be.Equal(t, http.StatusOK, w.Code)
	be.True(t, len(w.Body.String()) > 0)
	be.True(t, w.Header().Get("Content-Type") == contentTypeJSON)
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
			expectContain: "region not found",
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

			w, c := newRequest(t, "/")
			c.SetPathValues(echo.PathValues{
				{Name: "abbr", Value: tt.abbr},
			})
			err := app.RegionAPI(c)
			be.Equal(t, err, nil)

			be.Equal(t, tt.expectStatus, w.Code)
			be.True(t, len(w.Body.String()) > 0)
			be.True(t, w.Header().Get("Content-Type") == contentTypeJSON)
			be.True(t, strings.Contains(w.Body.String(), tt.expectContain))
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
			expectContain: "regions",
		},
		{
			name:          "search by abbreviation",
			query:         "ny",
			expectStatus:  http.StatusOK,
			expectContain: "regions",
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

			w, c := newRequest(t, "/")
			c.SetPathValues(echo.PathValues{
				{Name: "query", Value: tt.query},
			})
			err := app.AreacodeSearchAPI(c)
			be.Equal(t, err, nil)

			be.Equal(t, tt.expectStatus, w.Code)
			be.True(t, len(w.Body.String()) > 0)
			be.True(t, w.Header().Get("Content-Type") == contentTypeJSON)
			be.True(t, strings.Contains(w.Body.String(), tt.expectContain))
		})
	}
}
