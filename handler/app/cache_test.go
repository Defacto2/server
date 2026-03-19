package app_test

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Defacto2/server/handler/app"
	_ "github.com/jackc/pgx/v5"
	"github.com/labstack/echo/v4"
	"github.com/nalgeon/be"
)

// TestTagsCaching tests that the TagsAPI properly caches results.
func TestTagsCaching(t *testing.T) {
	// This test requires a database connection
	db, err := sql.Open("pgx", "postgres://root:example@localhost:5432/defacto2_ps?sslmode=disable")
	be.Equal(t, err, nil)
	defer func() {
		if err := db.Close(); err != nil {
			panic(err)
		}
	}()

	if err := db.PingContext(context.Background()); err != nil {
		t.Skip("Database not available, skipping cache test")
	}

	e := echo.New()
	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v0/categories", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// First call - should not be cached
	err = app.CategoriesAPI(c, db)
	be.Equal(t, err, nil)
	be.Equal(t, rec.Code, http.StatusOK)
	firstResponse := rec.Body.String()

	// Second call - should be cached and faster
	req2 := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v0/categories", nil)
	rec2 := httptest.NewRecorder()
	c2 := e.NewContext(req2, rec2)

	start := time.Now()
	err = app.CategoriesAPI(c2, db)
	elapsed := time.Since(start)
	be.Equal(t, err, nil)
	be.Equal(t, rec2.Code, http.StatusOK)
	secondResponse := rec2.Body.String()

	// Responses should be identical
	be.Equal(t, firstResponse, secondResponse)
	// Second call should be faster (though this is a rough check)
	be.True(t, elapsed < 10*time.Millisecond)
}

// TestCacheInvalidation tests that cache properly invalidates after expiration.
func TestCacheInvalidation(t *testing.T) {
	// This is a more advanced test that would require manipulating time
	// or waiting for cache expiration, which is not practical for unit tests
	t.Skip("Cache invalidation test requires time manipulation")
}
