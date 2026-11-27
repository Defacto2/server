package handler_test

import (
	"bytes"
	"crypto/sha512"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/Defacto2/server/handler"
	"github.com/Defacto2/server/handler/api"
	"github.com/Defacto2/server/handler/sess"
	"github.com/Defacto2/server/internal/config"
	"github.com/Defacto2/server/internal/testutil"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/v5/session"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	"github.com/nalgeon/be"
)

func TestSkipPaths(t *testing.T) {
	t.Parallel()

	e := echo.New()
	const moved = http.StatusMovedPermanently
	const bad = http.StatusBadRequest

	tests := []struct {
		target string
		status int
		want   bool
	}{
		{"/public/logo.png", http.StatusOK, true},
		{"/public/logo.png", moved, true},
		{"/public/logo.png", bad, false},
		{"/css/style.css", http.StatusOK, true},
		{"/", http.StatusOK, false},
	}

	for _, tt := range tests {
		c := testutil.MiddlewareStatus(t, e, tt.target, tt.status)
		got := handler.SkipPaths(c)
		be.Equal(t, got, tt.want)
	}
}

func next(c *echo.Context) error {
	return c.String(http.StatusOK, "ok")
}

func TestNoCrawl(t *testing.T) {
	t.Parallel()

	t.Run("enabled", func(t *testing.T) {
		t.Parallel()

		e := echo.New()
		h := handler.Server{Environment: config.Config{NoCrawl: true}}
		c := testutil.Middleware(t, e, "/")
		be.Err(t, h.NoCrawl(next)(c), nil)
		be.Equal(t, c.Response().Header().Get(handler.XRobotsTag), "none")
	})

	t.Run("disabled", func(t *testing.T) {
		t.Parallel()

		e := echo.New()
		h := handler.Server{Environment: config.Config{NoCrawl: false}}
		c := testutil.Middleware(t, e, "/")
		be.Err(t, h.NoCrawl(next)(c), nil)
		be.Equal(t, c.Response().Header().Get(handler.XRobotsTag), "")
	})
}

func TestReadOnlyLock(t *testing.T) {
	t.Parallel()

	t.Run("disabled", func(t *testing.T) {
		t.Parallel()

		e := echo.New()
		h := handler.Server{Environment: config.Config{ReadOnly: false}}
		called := false
		c := testutil.Middleware(t, e, "/")
		err := h.ReadOnlyLock(func(c *echo.Context) error {
			called = true
			return next(c)
		}, testLog)(c)
		be.Err(t, err, nil)
		be.True(t, called)
		be.Equal(t, c.Response().Header().Get(handler.XReadOnlyLock), "false")
	})

	t.Run("enabled", func(t *testing.T) {
		t.Parallel()

		e := echo.New()
		e.Renderer = nopRenderer{}
		h := handler.Server{Environment: config.Config{ReadOnly: true}}
		called := false
		c := testutil.Middleware(t, e, "/")
		err := h.ReadOnlyLock(func(c *echo.Context) error {
			called = true
			return next(c)
		}, testLog)(c)
		be.Err(t, err, nil)
		be.True(t, !called)
		be.Equal(t, c.Response().Header().Get(handler.XReadOnlyLock), "true")
		_, status := echo.ResolveResponseStatus(c.Response(), nil)
		be.Equal(t, status, http.StatusForbidden)
	})
}

func TestSessionLock(t *testing.T) {
	t.Parallel()

	const adminSub = "admin-google-id"
	adminHash := sha512.Sum384([]byte(adminSub))

	newSessionServer := func(t *testing.T) (*echo.Echo, *http.Cookie) {
		t.Helper()

		store := sessions.NewCookieStore([]byte("test-session-secret-key-32-bytes"))
		e := echo.New()
		e.Renderer = nopRenderer{}
		e.Use(session.Middleware(store))

		h := handler.Server{
			Environment: config.Config{
				GoogleAccounts: config.OAuth2s{adminHash},
			},
		}
		e.GET("/login", func(c *echo.Context) error {
			s, err := session.Get(sess.Name, c)
			if err != nil {
				return err
			}
			if sub := c.QueryParam("sub"); sub != "" {
				s.Values["sub"] = sub
			}
			if err := s.Save(c.Request(), c.Response()); err != nil {
				return err
			}
			return c.NoContent(http.StatusOK)
		})
		e.GET("/protected", next, func(n echo.HandlerFunc) echo.HandlerFunc {
			return h.SessionLock(n, testLog)
		})

		req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/login?sub="+adminSub, nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		be.Equal(t, rec.Code, http.StatusOK)

		cookies := rec.Result().Cookies()
		if len(cookies) == 0 {
			t.Fatal("expected session cookie")
		}
		return e, cookies[0]
	}

	t.Run("no session middleware", func(t *testing.T) {
		t.Parallel()

		e := echo.New()
		e.Renderer = nopRenderer{}
		h := handler.Server{}
		called := false
		e.GET("/protected", func(c *echo.Context) error {
			called = true
			return next(c)
		}, func(n echo.HandlerFunc) echo.HandlerFunc {
			return h.SessionLock(n, testLog)
		})

		req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/protected", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		be.Equal(t, rec.Code, http.StatusForbidden)
		be.True(t, !called)
	})

	t.Run("missing sub", func(t *testing.T) {
		t.Parallel()

		e, _ := newSessionServer(t)
		req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/protected", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		be.Equal(t, rec.Code, http.StatusForbidden)
	})

	t.Run("wrong sub", func(t *testing.T) {
		t.Parallel()

		e, _ := newSessionServer(t)
		loginReq := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/login?sub=intruder", nil)
		loginRec := httptest.NewRecorder()
		e.ServeHTTP(loginRec, loginReq)
		cookies := loginRec.Result().Cookies()
		if len(cookies) == 0 {
			t.Fatal("expected session cookie")
		}

		req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/protected", nil)
		req.AddCookie(cookies[0])
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		be.Equal(t, rec.Code, http.StatusForbidden)
	})

	t.Run("valid", func(t *testing.T) {
		t.Parallel()

		e, cookie := newSessionServer(t)
		req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/protected", nil)
		req.AddCookie(cookie)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		be.Equal(t, rec.Code, http.StatusOK)
		be.Equal(t, rec.Body.String(), "ok")
	})
}

func TestCacheMiddleware(t *testing.T) {
	t.Parallel()

	const (
		fiveMin = 60 * 5
		halfHr  = 60 * 30
		hour    = 60 * 60
		day     = 60 * 60 * 24
	)

	tests := []struct {
		path string
		age  int
	}{
		{"/categories", day},
		{"/milestones", day},
		{"/boards", hour},
		{"/artifact/123", hour},
		{"/releaser/acid", halfHr},
		{"/scener/someone", halfHr},
		{"/artifacts", fiveMin},
		{"/other", fiveMin},
	}

	e := echo.New()
	mw := handler.CacheMiddleware()(next)
	for _, tt := range tests {
		c := testutil.Middleware(t, e, tt.path)
		be.Err(t, mw(c), nil)
		got := c.Response().Header().Get(handler.CacheControl)
		be.Equal(t, got, "public, max-age="+strconv.Itoa(tt.age))
	}
}

func TestAPIMiddleware(t *testing.T) {
	t.Parallel()

	e := echo.New()
	e.Use(handler.APIMiddleware)
	e.GET("/api/v1/test", next)

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v1/test", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	be.Equal(t, rec.Code, http.StatusOK)
	be.Equal(t, rec.Header().Get(handler.XApiVersion), api.APIVer)
	be.True(t, strings.HasSuffix(rec.Header().Get(handler.XResponseTime), "ms"))
}

func TestRequestLogger(t *testing.T) {
	t.Parallel()

	t.Run("disabled", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer
		sl := slog.New(slog.NewTextHandler(&buf, nil))
		h := handler.Server{Environment: config.Config{LogAll: false}}

		e := echo.New()
		e.Use(middleware.RequestLoggerWithConfig(h.RequestLoggerConfig(sl)))
		e.GET("/ok", next)

		req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/ok", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		be.Equal(t, rec.Code, http.StatusOK)
		be.Equal(t, buf.String(), "")
	})

	t.Run("levels and skip", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer
		sl := slog.New(slog.NewTextHandler(&buf, nil))
		h := handler.Server{Environment: config.Config{LogAll: true}}

		e := echo.New()
		e.Use(middleware.RequestLoggerWithConfig(h.RequestLoggerConfig(sl)))
		e.GET("/css/app.css", next)
		e.GET("/ok", next)
		e.GET("/missing", func(c *echo.Context) error {
			return c.NoContent(http.StatusNotFound)
		})
		e.GET("/boom", func(c *echo.Context) error {
			return c.NoContent(http.StatusInternalServerError)
		})

		for _, path := range []string{"/css/app.css", "/ok", "/missing", "/boom"} {
			req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, path, nil)
			req.Header.Set("User-Agent", "curl/8.5.0")
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)
		}

		out := buf.String()
		be.True(t, !strings.Contains(out, "/css/app.css"))
		be.True(t, strings.Contains(out, `level=INFO`) && strings.Contains(out, `"GET 200 /ok"`))
		be.True(t, strings.Contains(out, `level=WARN`) && strings.Contains(out, `"GET 404 /missing"`))
		be.True(t, strings.Contains(out, `level=ERROR`) && strings.Contains(out, `"GET 500 /boom"`))
		be.True(t, strings.Contains(out, "agent=curl/8.5"))
	})
}

func TestShortAgent(t *testing.T) {
	t.Parallel()

	tests := []struct {
		ua   string
		want string
	}{
		{"", ""},
		{"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/158.0.0.0 Safari/537.36", "Chrome/158 (Linux)"},
		{"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:134.0) Gecko/20100101 Firefox/134.0", "Firefox/134 (Windows)"},
		{"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/18.1 Safari/605.1.15", "Safari/18 (macOS)"},
		{"Mozilla/5.0 (iPhone; CPU iPhone OS 18_1 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/18.1 Mobile/15E148 Safari/604.1", "Safari/18 (iOS)"},
		{"Mozilla/5.0 (Linux; Android 14; Pixel 8) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/130.0.6723.58 Mobile Safari/537.36", "Chrome/130 (Android)"},
		{"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/130.0.0.0 Safari/537.36 Edg/130.0.0.0", "Edge/130 (Windows)"},
		{"curl/8.5.0", "curl/8.5"},
		{"Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)", "Googlebot"},
		{"MyCustomCrawler/1.0", "bot"},
		{"Mozilla/5.0 Safari/605.1", "Safari"},
		{"VeryLongCustomUserAgentStringExceedingLimit/1.0", "VeryLongCustomUserAgentS"},
	}
	for _, tt := range tests {
		be.Equal(t, handler.ShortAgent(tt.ua), tt.want)
	}
}
