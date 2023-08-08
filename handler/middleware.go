package handler

// Package file middleware.go contains the custom middleware functions for the Echo web framework.

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// removeSlash return the TrailingSlash middleware configuration.
func (c Configuration) removeSlash() middleware.TrailingSlashConfig {
	return middleware.TrailingSlashConfig{
		RedirectCode: http.StatusMovedPermanently,
	}
}

// Timeout returns the timeout middleware configuration.
func (c Configuration) timeout() middleware.TimeoutConfig {
	return middleware.TimeoutConfig{
		Timeout: time.Duration(c.Import.Timeout) * time.Second,
	}
}

// NoRobotsHeader middleware adds a `X-Robots-Tag` header to the response.
// The header contains the noindex and nofollow values that tell search engine
// crawlers to not index or crawl the page or asset.
// See https://developers.google.com/search/docs/crawling-indexing/robots-meta-tag#xrobotstag
func (c Configuration) NoRobotsHeader(next echo.HandlerFunc) echo.HandlerFunc {
	if !c.Import.NoRobots {
		return next
	}
	return func(c echo.Context) error {
		const HeaderXRobotsTag = "X-Robots-Tag"
		c.Response().Header().Set(HeaderXRobotsTag, "noindex, nofollow")
		return next(c)
	}
}
