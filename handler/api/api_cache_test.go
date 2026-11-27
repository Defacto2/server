package api_test

import (
	"database/sql"
	"net/http"
	"testing"
	"time"

	"github.com/Defacto2/server/handler/api"
	"github.com/nalgeon/be"
)

// TestTagsCaching tests that the TagsAPI properly caches results.
func TestTagsCaching(t *testing.T) {
	t.Parallel()

	// This test requires a configured test database connection.
	db, err := sql.Open(driverName, dataSourceName)
	be.Err(t, err, nil)
	defer func() {
		if err := db.Close(); err != nil {
			t.Fatal(err)
		}
	}()

	if err := db.PingContext(t.Context()); err != nil {
		t.Skip("Database not available, skipping cache test")
	}

	const target = api.APIBase + "/categories"
	w, c := newRequest(t, target)

	// should not be cached
	err = api.CategoriesAPI(c, db)
	be.Err(t, err, nil)
	be.Equal(t, w.Code, http.StatusOK)
	firstResponse := w.Body.String()

	// should be cached and quicker
	w2, c2 := newRequest(t, target)

	start := time.Now()
	err = api.CategoriesAPI(c2, db)
	elapsed := time.Since(start)
	be.Err(t, err, nil)
	be.Equal(t, w2.Code, http.StatusOK)
	secondResponse := w2.Body.String()

	// Responses should be identical
	be.Equal(t, secondResponse, firstResponse)
	// Second call should be faster (though this is a rough check)
	be.True(t, elapsed < 10*time.Millisecond)
}
