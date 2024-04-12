package handler

// Package file middleware.go contains the custom middleware functions for the Echo web framework.

// DO NOT USE THE middleware.TimeoutWithConfig().
// It is broken and causes race conditions and broken responses.
// See, https://github.com/labstack/echo/issues/2306

import (
	"crypto/sha512"
	"net/http"
	"strconv"

	"github.com/Defacto2/server/handler/app"
	"github.com/Defacto2/server/handler/sess"
	"github.com/Defacto2/server/internal/zaplog"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// NoCrawl middleware adds a `X-Robots-Tag` header to the response.
// The header contains the noindex and nofollow values that tell search engine
// crawlers to not index or crawl the page or asset.
// See https://developers.google.com/search/docs/crawling-indexing/robots-meta-tag#xrobotstag
func (c Configuration) NoCrawl(next echo.HandlerFunc) echo.HandlerFunc {
	if !c.Import.NoCrawl {
		return next
	}
	return func(e echo.Context) error {
		const HeaderXRobotsTag = "X-Robots-Tag"
		e.Response().Header().Set(HeaderXRobotsTag, "noindex, nofollow")
		return next(e)
	}
}

// ReadOnlyLock disables all POST, PUT and DELETE requests for the modification
// of the database and any related user interface.
func (c Configuration) ReadOnlyLock(next echo.HandlerFunc) echo.HandlerFunc {
	return func(e echo.Context) error {
		s := strconv.FormatBool(c.Import.ReadMode)
		e.Response().Header().Set("X-Read-Only-Lock", s)
		if c.Import.ReadMode {
			return app.StatusErr(e, http.StatusForbidden, "")
		}
		return next(e)
	}
}

// SessionLock middleware checks the session cookie for a valid signed in client.
func (c Configuration) SessionLock(next echo.HandlerFunc) echo.HandlerFunc {
	return func(e echo.Context) error {
		// https://pkg.go.dev/github.com/gorilla/sessions#Session
		sess, err := session.Get(sess.Name, e)
		if err != nil {
			return err
		}
		id, ok := sess.Values["sub"].(string)
		if !ok || id == "" {
			return app.StatusErr(e, http.StatusForbidden, "")
		}
		check := false
		for _, account := range c.Import.GoogleAccounts {
			if sum := sha512.Sum384([]byte(id)); sum == account {
				check = true
				break
			}
		}
		if !check {
			return app.StatusErr(e, http.StatusForbidden, "")
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

// configZapLogger returns the RequestLogger middleware configuration
// based on the application configuration. The logger is set to the CLI
// logger for development mode and the Production logger for production mode.
func (c Configuration) configZapLogger() middleware.RequestLoggerConfig {
	if !c.Import.LogRequests {
		return middleware.RequestLoggerConfig{
			LogValuesFunc: func(_ echo.Context, _ middleware.RequestLoggerValues) error {
				return nil
			},
		}
	}
	logger := zaplog.CLI().Sugar()
	if c.Import.ProductionMode {
		root := c.Import.LogDir
		logger = zaplog.Production(root).Sugar()
	}
	defer func() {
		_ = logger.Sync()
	}()
	return middleware.RequestLoggerConfig{
		LogURI:          true,
		LogStatus:       true,
		LogLatency:      true,
		LogResponseSize: true,
		LogValuesFunc: func(_ echo.Context, v middleware.RequestLoggerValues) error {
			const template = "HTTP %s %d: %s %s %dB"
			if v.Status > http.StatusAlreadyReported {
				logger.Warnf(template,
					v.Method, v.Status, v.URI, v.Latency, v.ResponseSize)
				return nil
			}
			logger.Infof(template,
				v.Method, v.Status, v.URI, v.Latency, v.ResponseSize)
			return nil
		},
	}
}
