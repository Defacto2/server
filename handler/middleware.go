package handler

// Package file middleware.go contains the custom middleware functions for the Echo web framework.

// WARN: DO NOT USE THE middleware.TimeoutWithConfig().
// It is broken by causing race conditions and broken responses.
// See, https://github.com/labstack/echo/issues/2306

import (
	"crypto/sha512"
	"fmt"
	"math"
	"net/http"
	"runtime"
	"strconv"
	"strings"

	"github.com/Defacto2/server/handler/app"
	"github.com/Defacto2/server/handler/sess"
	"github.com/Defacto2/server/internal/zaplog"
	"github.com/dustin/go-humanize"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// NoCrawl middleware adds a `X-Robots-Tag` header to the response.
// The header contains the noindex and nofollow values that tell search engine
// crawlers to not index or crawl the page or asset.
// See https://developers.google.com/search/docs/crawling-indexing/robots-meta-tag#xrobotstag
func (c *Configuration) NoCrawl(next echo.HandlerFunc) echo.HandlerFunc {
	if !c.Environment.NoCrawl {
		return next
	}
	return func(e echo.Context) error {
		const HeaderXRobotsTag = "X-Robots-Tag"
		e.Response().Header().Set(HeaderXRobotsTag, "none")
		return next(e)
	}
}

// ReadOnlyLock disables all PATCH, POST, PUT and DELETE requests for the modification
// of the database and any related user interface.
func (c *Configuration) ReadOnlyLock(next echo.HandlerFunc) echo.HandlerFunc {
	const msg = "middleware read only lock"
	return func(e echo.Context) error {
		s := strconv.FormatBool(bool(c.Environment.ReadOnly))
		e.Response().Header().Set("X-Read-Only-Lock", s)
		if c.Environment.ReadOnly {
			if err := app.StatusErr(e, http.StatusForbidden, ""); err != nil {
				return fmt.Errorf("%s status: %w", msg, err)
			}
			return nil
		}
		return next(e)
	}
}

// SessionLock middleware checks the session cookie for a valid signed in client.
func (c *Configuration) SessionLock(next echo.HandlerFunc) echo.HandlerFunc {
	const msg = "middleware session lock"
	return func(e echo.Context) error {
		// Help, https://pkg.go.dev/github.com/gorilla/sessions#Session
		sess, err := session.Get(sess.Name, e)
		if err != nil {
			return fmt.Errorf("%s get: %w", msg, err)
		}
		id, subExists := sess.Values["sub"].(string)
		if !subExists || id == "" {
			if err := app.StatusErr(e, http.StatusForbidden, ""); err != nil {
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
			if err := app.StatusErr(e, http.StatusForbidden, ""); err != nil {
				return fmt.Errorf("%s check forbid: %w", msg, err)
			}
			return nil
		}
		return next(e)
	}
}

// configRTS return the TrailingSlash middleware configuration.
func configRTS() middleware.TrailingSlashConfig {
	return middleware.TrailingSlashConfig{
		RedirectCode: http.StatusMovedPermanently,
	}
}

// TODO: remove or repurpose to use slog?
// configZapLogger returns the RequestLogger middleware configuration
// based on the application configuration. The logger is set to the CLI
// logger for development mode and the Production logger for production mode.
func (c *Configuration) configZapLogger() middleware.RequestLoggerConfig {
	noLogging := func(_ echo.Context, _ middleware.RequestLoggerValues) error {
		return nil
	}
	if logAllRequests := c.Environment.LogAll; !logAllRequests {
		return middleware.RequestLoggerConfig{LogValuesFunc: noLogging}
	}

	logger := zaplog.Status().Sugar()
	if c.Environment.ProdMode {
		logPath := c.Environment.AbsLog
		logger = zaplog.Store(zaplog.Text(), string(logPath)).Sugar()
	}
	defer func() {
		_ = logger.Sync()
	}()

	logValues := func(_ echo.Context, v middleware.RequestLoggerValues) error {
		// memory usage
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		rsize := uint64(math.Abs(float64(v.ResponseSize)))
		alloc := humanize.Bytes(m.Alloc)
		// cpu usage
		numCPU := runtime.NumCPU()
		numGoroutine := runtime.NumGoroutine()
		// log template
		const template = "%d %s %s > %s [%s][%s][%s][CPU %d of %d] %s"
		if v.Status > http.StatusAlreadyReported {
			logger.Warnf(template, v.Status, v.Method, v.URI,
				v.RoutePath, v.Latency, humanize.Bytes(rsize), alloc, numGoroutine, numCPU, v.UserAgent)
			return nil
		}
		logger.Infof(template, v.Status, v.Method, v.URIPath,
			v.RoutePath, v.Latency, humanize.Bytes(rsize), alloc, numGoroutine, numCPU, v.UserAgent)
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
		strings.HasPrefix(uri, "/svg/"):
		if statusOk {
			return true
		}
	}
	return false
}
