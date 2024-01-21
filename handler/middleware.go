package handler

// Package file middleware.go contains the custom middleware functions for the Echo web framework.

// DO NOT USE THE middleware.TimeoutWithConfig().
// It is broken and causes race conditions and broken responses.
// See, https://github.com/labstack/echo/issues/2306

import (
	"crypto/sha512"
	"fmt"
	"net/http"

	"github.com/Defacto2/server/handler/app"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// removeSlash return the TrailingSlash middleware configuration.
func (cfg Configuration) removeSlash() middleware.TrailingSlashConfig {
	return middleware.TrailingSlashConfig{
		RedirectCode: http.StatusMovedPermanently,
	}
}

// NoRobotsHeader middleware adds a `X-Robots-Tag` header to the response.
// The header contains the noindex and nofollow values that tell search engine
// crawlers to not index or crawl the page or asset.
// See https://developers.google.com/search/docs/crawling-indexing/robots-meta-tag#xrobotstag
func (cfg Configuration) NoRobotsHeader(next echo.HandlerFunc) echo.HandlerFunc {
	if !cfg.Import.NoRobots {
		return next
	}
	return func(c echo.Context) error {
		const HeaderXRobotsTag = "X-Robots-Tag"
		c.Response().Header().Set(HeaderXRobotsTag, "noindex, nofollow")
		return next(c)
	}
}

// ReadOnlyLock disables all POST, PUT and DELETE requests for the modification
// of the database and any related user interface.
func (cfg Configuration) ReadOnlyLock(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		c.Response().Header().Set("X-Read-Only-Lock", fmt.Sprintf("%t", cfg.Import.IsReadOnly))
		if cfg.Import.IsReadOnly {
			return app.StatusErr(cfg.Logger, c, http.StatusForbidden, "")
		}
		return next(c)
	}
}

// SessionLock middleware checks the session cookie for a valid signed in user.
func (cfg Configuration) SessionLock(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// https://pkg.go.dev/github.com/gorilla/sessions#Session
		sess, err := session.Get(app.SessionName, c)
		if err != nil {
			return err
		}
		id, ok := sess.Values["sub"].(string)
		if !ok || id == "" {
			return echo.ErrForbidden
		}
		check := false
		for _, account := range cfg.Import.GoogleAccounts {
			if sum := sha512.Sum384([]byte(id)); sum == account {
				check = true
				break
			}
		}
		if !check {
			return echo.ErrForbidden
		}
		return next(c)
	}
}
