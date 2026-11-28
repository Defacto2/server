//nolint:paralleltest
package handler_test

import (
	"database/sql"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/Defacto2/server/handler"
	"github.com/Defacto2/server/internal/config"
	"github.com/Defacto2/server/internal/logs"
	"github.com/Defacto2/server/internal/nils"
	"github.com/Defacto2/server/internal/testutil"
	"github.com/labstack/echo/v5"
	"github.com/nalgeon/be"
)

var testLog = logs.Discard()

type nopRenderer struct{}

func (nopRenderer) Render(_ *echo.Context, _ io.Writer, _ string, _ any) error {
	return nil
}

func setupRouter(t *testing.T, env config.Config) *echo.Echo {
	t.Helper()

	serv := handler.Server{Environment: env}
	e := echo.New()
	e.Renderer = nopRenderer{}
	e, err := serv.RouteFS(testLog, e, new(sql.DB), testutil.OpenFS(t))
	be.Err(t, err, nil)
	if e == nil {
		t.Fatal("expected non-nil echo instance")
	}
	return e
}

// TestServerRouteFS mutates package-level app.Caching via RouteFS, so it cannot run in parallel.
func TestServerRouteFS(t *testing.T) {
	t.Run("nil args", func(t *testing.T) {
		serv := handler.Server{}
		e, err := serv.RouteFS(nil, nil, nil, nil)
		be.Err(t, err)
		be.True(t, e == nil)
	})

	t.Run("empty fs", func(t *testing.T) {
		serv := handler.Server{}
		e, err := serv.RouteFS(testLog, echo.New(), new(sql.DB), fstest.MapFS{})
		be.True(t, errors.Is(err, nils.ErrFS))
		be.True(t, e == nil)
	})

	t.Run("routes", func(t *testing.T) {
		e := setupRouter(t, config.Config{})
		for _, tc := range []struct {
			path string
			want int
		}{
			{"/health-check", http.StatusOK},
			{"/robots.txt", http.StatusOK},
			{"/want-404", http.StatusNotFound},
			{"/files/for-approval", http.StatusNotFound},
			{"/files/deletions", http.StatusNotFound},
			{"/files/unwanted", http.StatusNotFound},
		} {
			req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, tc.path, nil)
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)
			be.Equal(t, rec.Code, tc.want)
		}
	})

	t.Run("query redirects", func(t *testing.T) {
		e := setupRouter(t, config.Config{})
		for _, tc := range []struct {
			path string
			loc  string
		}{
			{"/g/acid?x=1", "/g/acid"},
			{"/p/someone?x=1", "/p/someone"},
			{"/f/abc123?x=1", "/f/abc123"},
		} {
			req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, tc.path, nil)
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)
			be.Equal(t, rec.Code, http.StatusMovedPermanently)
			be.Equal(t, rec.Header().Get("Location"), tc.loc)
		}
	})

	t.Run("debug prod gate", func(t *testing.T) {
		dev := setupRouter(t, config.Config{ProdMode: false})
		req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/debug", nil)
		rec := httptest.NewRecorder()
		dev.ServeHTTP(rec, req)
		be.Equal(t, rec.Code, http.StatusOK)
		be.True(t, strings.Contains(rec.Body.String(), `"method": "GET"`))

		prod := setupRouter(t, config.Config{ProdMode: true})
		req = httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/debug", nil)
		rec = httptest.NewRecorder()
		prod.ServeHTTP(rec, req)
		be.Equal(t, rec.Code, http.StatusNotFound)
	})

	t.Run("read-only signin", func(t *testing.T) {
		ro := setupRouter(t, config.Config{ReadOnly: true})
		req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/operator/signin", nil)
		rec := httptest.NewRecorder()
		ro.ServeHTTP(rec, req)
		be.Equal(t, rec.Code, http.StatusForbidden)

		rw := setupRouter(t, config.Config{ReadOnly: false})
		req = httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/operator/signin", nil)
		rec = httptest.NewRecorder()
		rw.ServeHTTP(rec, req)
		be.Equal(t, rec.Code, http.StatusMovedPermanently)
		be.Equal(t, rec.Header().Get("Location"), "/signin")
	})
}

func TestRouteMoved(t *testing.T) {
	t.Parallel()

	e := handler.RouteMoved(echo.New())
	tests := []struct {
		path string
		loc  string
	}{
		{"/file/detail/abc", "/f/abc"},
		{"/file/download/abc", "/d/abc"},
		{"/wayback/test", "https://wayback.defacto2.net/test"},
		{"/link/list", "https://wayback.defacto2.net/"},
		{"/g/acid", "/g/acid-productions"},
		{"/g/rss", "/g/renaissance"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, tt.path, nil)
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)
			be.Equal(t, rec.Code, http.StatusMovedPermanently)
			be.Equal(t, rec.Header().Get("Location"), tt.loc)
		})
	}
}
