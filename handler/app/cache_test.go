package app_test

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Defacto2/server/handler/app"
	_ "github.com/jackc/pgx/v5"
	"github.com/labstack/echo/v5"
	"github.com/nalgeon/be"
)

// TestTagsCaching tests that the TagsAPI properly caches results.
func TestTagsCaching(t *testing.T) {
	// This test requires a configured test, database connection
	const dataSourceName = "postgres://root:example@localhost:5432/defacto2_ps?sslmode=disable" //nolint:gosec
	db, err := sql.Open("pgx", dataSourceName)
	be.Equal(t, err, nil)
	defer func() {
		if err := db.Close(); err != nil {
			t.Fatal(err)
		}
	}()

	if err := db.PingContext(t.Context()); err != nil {
		t.Skip("Database not available, skipping cache test")
	}

	const target = "/api/v0/categories"
	e := echo.New()
	r := httptest.NewRequestWithContext(t.Context(), http.MethodGet, target, nil)
	w := httptest.NewRecorder()
	c := e.NewContext(r, w)

	// should not be cached
	err = app.CategoriesAPI(t.Context(), c, db)
	be.Equal(t, err, nil)
	be.Equal(t, w.Code, http.StatusOK)
	firstResponse := w.Body.String()

	// should be cached and quicker
	r2 := httptest.NewRequestWithContext(t.Context(), http.MethodGet, target, nil)
	w2 := httptest.NewRecorder()
	c2 := e.NewContext(r2, w2)

	start := time.Now()
	err = app.CategoriesAPI(t.Context(), c2, db)
	elapsed := time.Since(start)
	be.Equal(t, err, nil)
	be.Equal(t, w2.Code, http.StatusOK)
	secondResponse := w2.Body.String()

	// Responses should be identical
	be.Equal(t, firstResponse, secondResponse)
	// Second call should be faster (though this is a rough check)
	be.True(t, elapsed < 10*time.Millisecond)
}
