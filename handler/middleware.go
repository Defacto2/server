package handler

// Package file middleware.go contains the custom middleware functions for the Echo web framework.

// WARN: DO NOT USE THE middleware.TimeoutWithConfig().
// It is broken by causing race conditions and broken responses.
// See, https://github.com/labstack/echo/issues/2306

import (
	"crypto/sha512"
	"fmt"
	"log/slog"
	"math"
	"net/http"
	"runtime"
	"strconv"
	"strings"

	"github.com/Defacto2/server/handler/app"
	"github.com/Defacto2/server/handler/sess"
	"github.com/Defacto2/server/internal/panics"
	"github.com/dustin/go-humanize"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// SkipPaths are parent route paths that should not be logged,
// to reduce the logging output. Otherwise every image
// or required resource for every page request would be returned.
func skipPaths(e echo.Context) bool {
	if redirect := e.Response().Status == http.StatusMovedPermanently; redirect {
		return true
	}
	uri := e.Request().RequestURI
	statusOk := e.Response().Status == http.StatusOK
	switch {
	case strings.HasPrefix(uri, "/public/"),
		strings.HasPrefix(uri, "/css/"),
		strings.HasPrefix(uri, "/js/"),
		strings.HasPrefix(uri, "/image/"),
		strings.HasPrefix(uri, "/svg/"),
		strings.HasPrefix(uri, "/font/"):
		if statusOk {
			return true
		}
	}
	return false
}

// NoCrawl middleware adds a `X-Robots-Tag` header to the response.
// The header contains the noindex and nofollow values that tell search engine
// crawlers to not index or crawl the page or asset.
// See https://developers.google.com/search/docs/crawling-indexing/robots-meta-tag#xrobotstag
func (c *Configuration) NoCrawl(next echo.HandlerFunc) echo.HandlerFunc {
	if !c.Environment.NoCrawl {
		return next
	}
	return func(e echo.Context) error {
		const xrobotstag = "X-Robots-Tag"
		e.Response().Header().Set(xrobotstag, "none")
		return next(e)
	}
}

// ReadOnlyLock disables all PATCH, POST, PUT and DELETE requests for the modification
// of the database and any related user interface.
func (c *Configuration) ReadOnlyLock(next echo.HandlerFunc, sl *slog.Logger) echo.HandlerFunc {
	const msg = "middleware read only lock"
	if sl == nil {
		panic(fmt.Errorf("%s: %w", msg, panics.ErrNoSlog))
	}
	return func(e echo.Context) error {
		const xreadonlylock = "X-Read-Only-Lock"
		s := strconv.FormatBool(bool(c.Environment.ReadOnly))
		e.Response().Header().Set(xreadonlylock, s)
		if c.Environment.ReadOnly {
			if err := app.StatusErr(e, sl, http.StatusForbidden, ""); err != nil {
				return fmt.Errorf("%s status: %w", msg, err)
			}
			// do not run next(e)
			return nil
		}
		return next(e)
	}
}

// SessionLock middleware checks the session cookie for a valid signed in client.
func (c *Configuration) SessionLock(next echo.HandlerFunc, sl *slog.Logger) echo.HandlerFunc {
	const msg = "middleware session lock"
	if sl == nil {
		panic(fmt.Errorf("%s: %w", msg, panics.ErrNoSlog))
	}
	return func(e echo.Context) error {
		// Help, https://pkg.go.dev/github.com/gorilla/sessions#Session
		sess, err := session.Get(sess.Name, e)
		if err != nil {
			return fmt.Errorf("%s get: %w", msg, err)
		}
		id, subExists := sess.Values["sub"].(string)
		if !subExists || id == "" {
			if err := app.StatusErr(e, sl, http.StatusForbidden, ""); err != nil {
				return fmt.Errorf("%s subexists forbid: %w", msg, err)
			}
			return nil
		}
		check := false
		for _, account := range c.Environment.GoogleAccounts {
			if sum := sha512.Sum384([]byte(id)); sum == account {
				check = true
				break
			}
		}
		if !check {
			if err := app.StatusErr(e, sl, http.StatusForbidden, ""); err != nil {
				return fmt.Errorf("%s check forbid: %w", msg, err)
			}
			return nil
		}
		return next(e)
	}
}

// trailSlash return the TrailingSlash middleware configuration.
func trailSlash() middleware.TrailingSlashConfig {
	return middleware.TrailingSlashConfig{
		RedirectCode: http.StatusMovedPermanently,
	}
}

// RequestLoggerConfig handles logging for HTTP page requests.
// A slog Logger is required otherwise it will panic.
//
// If Configuration.LogAll is false then this returns a nil.
// Otherwise it logs all web server HTTP requests to info logs.
func (c *Configuration) RequestLoggerConfig(sl *slog.Logger) middleware.RequestLoggerConfig {
	if logall := c.Environment.LogAll; !logall {
		exitRequest := func(_ echo.Context, _ middleware.RequestLoggerValues) error {
			return nil
		}
		return middleware.RequestLoggerConfig{LogValuesFunc: exitRequest}
	}
	const msg = "request logger config handler"
	if sl == nil {
		panic(fmt.Errorf("%s: %w", msg, panics.ErrNoSlog))
	}
	// logValues is used by the returned middleware.RequestLoggerConfig().LogValuesFunc
	logValues := func(_ echo.Context, v middleware.RequestLoggerValues) error {
		// memory usage
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		rsize := uint64(math.Abs(float64(v.ResponseSize)))
		alloc := humanize.Bytes(m.Alloc)
		// use funcs to maintain the readability of the nested slog arguments
		response := func() slog.Attr {
			return slog.Group("response",
				slog.Int64("size", v.ResponseSize),
				slog.String("humanize", humanize.Bytes(rsize)))
		}
		cpuinfo := func() slog.Attr {
			return slog.Group("cpu",
				slog.Int("cores", runtime.NumCPU()),
				slog.Int("go_routines", runtime.NumGoroutine()))
		}
		requests := func() slog.Attr {
			return slog.Group("request",
				slog.String("agent", v.UserAgent), // browser agent used for debugging
				slog.String("path", v.URIPath),    // uri path without any params
				slog.String("route", v.RoutePath), // internal route path with values
				slog.String("uri", v.URI),         // complete url request
			)
		}
		sl.Info(fmt.Sprintf("HTTP %s %d", v.Method, v.Status),
			slog.Duration("latency", v.Latency),
			slog.String("uri", v.URIPath),
			response(), cpuinfo(),
			slog.String("allocation", alloc),
			// slog.Any("request", v), // uncomment for verbose & debugging
			requests())
		return nil
	}
	return middleware.RequestLoggerConfig{
		Skipper:          skipPaths,
		LogLatency:       true,
		LogProtocol:      false,
		LogRemoteIP:      false,
		LogHost:          false,
		LogMethod:        true,
		LogURI:           true,
		LogURIPath:       true,
		LogRoutePath:     true,
		LogRequestID:     false,
		LogReferer:       false,
		LogUserAgent:     true,
		LogStatus:        true,
		LogError:         false,
		LogContentLength: false,
		LogResponseSize:  true,
		LogHeaders:       nil,
		LogQueryParams:   nil,
		LogFormValues:    nil,
		LogValuesFunc:    logValues,
	}
}
