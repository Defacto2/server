package handler_test

import (
	"net/http"
	"strconv"
	"testing"

	"github.com/Defacto2/server/handler"
	"github.com/Defacto2/server/internal/config"
	"github.com/Defacto2/server/internal/logs"
	"github.com/Defacto2/server/internal/testutil"
	"github.com/labstack/echo/v5"
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

	e := echo.New()

	env := config.Config{
		NoCrawl: true,
	}
	h := handler.Server{
		Environment: env,
	}

	c := testutil.Middleware(t, e, "/")
	funcHandler := h.NoCrawl(next)
	err := funcHandler(c)
	be.Err(t, err, nil)

	got := c.Response().Header().Get(handler.XRobotsTag)
	be.Equal(t, got, "none")
}

func TestReadOnlyLock(t *testing.T) {
	t.Parallel()

	e := echo.New()
	sl := logs.Discard()

	env := config.Config{
		ReadOnly: false,
	}
	h := handler.Server{
		Environment: env,
	}

	c := testutil.Middleware(t, e, "/")
	funcHandler := h.ReadOnlyLock(next, sl)
	err := funcHandler(c)
	be.Err(t, err, nil)

	got := c.Response().Header().Get(handler.XReadOnlyLock)
	be.Equal(t, got, "false")
	// don't test ReadOnly true, because it returns an status forbidden error
}

func TestCacheMiddleware(t *testing.T) {
	t.Parallel()

	e := echo.New()
	e.Use(handler.CacheMiddleware())

	c := testutil.Middleware(t, e, "/categories")
	err := handler.CacheMiddleware()(next)(c)
	be.Err(t, err, nil)
	const day = 60 * 60 * 24
	got := c.Response().Header().Get(handler.CacheControl)
	be.Equal(t, got, "public, max-age="+strconv.Itoa(day))

	c = testutil.Middleware(t, e, "/boards")
	err = handler.CacheMiddleware()(next)(c)
	be.Err(t, err, nil)
	const hour = 60 * 60
	got = c.Response().Header().Get(handler.CacheControl)
	be.Equal(t, got, "public, max-age="+strconv.Itoa(hour))
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
	}
	for _, tt := range tests {
		be.Equal(t, tt.want, handler.ShortAgent(tt.ua))
	}
}
