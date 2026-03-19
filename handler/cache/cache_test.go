package cache_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Defacto2/server/handler"
	"github.com/Defacto2/server/handler/app"
	"github.com/labstack/echo/v4"
	"github.com/nalgeon/be"
)

// TestCacheMiddleware tests the cache middleware functionality.
func TestCacheMiddleware(t *testing.T) {
	// Create a test Echo instance
	e := echo.New()

	// Apply the cache middleware
	e.Use(handler.CacheMiddleware())

	// Test cases for different endpoints
	testCases := []struct {
		name           string
		path           string
		expectedMaxAge string
	}{
		{"Categories endpoint", app.APIBase + "/categories", "86400"},
		{"Platforms endpoint", app.APIBase + "/platforms", "86400"},
		{"Artifacts endpoint", app.APIBase + "/artifacts", "300"},
		{"Artifacts new endpoint", app.APIBase + "/artifacts/new", "300"},
		{"Artifact detail endpoint", app.APIBase + "/artifact/abc123", "3600"},
		{"Releaser endpoint", app.APIBase + "/releaser/test-group", "1800"},
		{"Scener endpoint", app.APIBase + "/scener/test-scener", "1800"},
		{"Groups endpoint", app.APIBase + "/groups", "3600"},
		{"Milestones endpoint", app.APIBase + "/milestones", "86400"},
		{"Area codes endpoint", app.APIBase + "/areacodes", "86400"},
		{"Websites endpoint", app.APIBase + "/websites", "86400"},
		{"Demozoo endpoint", app.APIBase + "/demozoo", "86400"},
		{"Unknown endpoint", app.APIBase + "/unknown", "300"}, // fallback
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a test request
			req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, tc.path, nil)
			rec := httptest.NewRecorder()

			// Set up Echo to handle the request properly
			e.GET(tc.path, func(c echo.Context) error {
				return c.String(http.StatusOK, "test")
			})

			// Make the request through Echo's router
			e.ServeHTTP(rec, req)

			// Check for errors
			// Check Cache-Control header
			cacheHeader := rec.Header().Get("Cache-Control")
			be.True(t, cacheHeader != "")

			// Verify the max-age value
			be.True(t, strings.Contains(cacheHeader, "public"))
			be.True(t, strings.Contains(cacheHeader, "max-age="+tc.expectedMaxAge))
		})
	}
}

// TestCacheMiddlewareOrder tests that cache middleware works with other middleware.
func TestCacheMiddlewareOrder(t *testing.T) {
	e := echo.New()

	// Add cache middleware first
	e.Use(handler.CacheMiddleware())

	// Add another middleware that modifies headers
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Response().Header().Set("X-Custom-Header", "test-value")
			return next(c)
		}
	})

	// Test that both middleware work together
	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, app.APIBase+"/categories", nil)
	rec := httptest.NewRecorder()

	e.GET(app.APIBase+"/categories", func(c echo.Context) error {
		return c.String(http.StatusOK, "test")
	})

	e.ServeHTTP(rec, req)

	// ServeHTTP doesn't return error, so we check status code instead
	be.Equal(t, rec.Code, http.StatusOK)
	be.Equal(t, rec.Header().Get("X-Custom-Header"), "test-value")
	be.True(t, strings.Contains(rec.Header().Get("Cache-Control"), "max-age=86400"))
}

// TestCacheMiddlewarePathMatching tests the path matching logic.
func TestCacheMiddlewarePathMatching(t *testing.T) {
	e := echo.New()
	e.Use(handler.CacheMiddleware())

	testCases := []struct {
		path            string
		expectedPattern string
	}{
		{app.APIBase + "/categories", "max-age=86400"},
		{app.APIBase + "/platforms", "max-age=86400"},
		{app.APIBase + "/artifacts", "max-age=300"},
		{app.APIBase + "/artifacts/new", "max-age=300"},
		{app.APIBase + "/artifact/anyhash", "max-age=3600"},
		{app.APIBase + "/artifact/another123", "max-age=3600"},
		{app.APIBase + "/releaser/group-name", "max-age=1800"},
		{app.APIBase + "/scener/person-name", "max-age=1800"},
	}

	for _, tc := range testCases {
		t.Run("path:"+tc.path, func(t *testing.T) {
			req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, tc.path, nil)
			rec := httptest.NewRecorder()

			e.GET(tc.path, func(c echo.Context) error {
				return c.String(http.StatusOK, "test")
			})

			e.ServeHTTP(rec, req)
			be.Equal(t, rec.Code, http.StatusOK)

			cacheHeader := rec.Header().Get("Cache-Control")
			be.True(t, strings.Contains(cacheHeader, tc.expectedPattern))
		})
	}
}
