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
		c := testutil.EchoContextS(t, e, tt.target, tt.status)
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

	c := testutil.EchoContext(t, e, "/")
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

	c := testutil.EchoContext(t, e, "/")
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

	c := testutil.EchoContext(t, e, "/categories")
	err := handler.CacheMiddleware()(next)(c)
	be.Err(t, err, nil)
	const day = 60 * 60 * 24
	got := c.Response().Header().Get(handler.CacheControl)
	be.Equal(t, got, "public, max-age="+strconv.Itoa(day))

	c = testutil.EchoContext(t, e, "/boards")
	err = handler.CacheMiddleware()(next)(c)
	be.Err(t, err, nil)
	const hour = 60 * 60
	got = c.Response().Header().Get(handler.CacheControl)
	be.Equal(t, got, "public, max-age="+strconv.Itoa(hour))
}
