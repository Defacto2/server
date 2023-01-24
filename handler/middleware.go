package handler

import "github.com/labstack/echo/v4"

// NoRobotsHeader middleware adds a `X-Robots-Tag` header to the response.
// The header contains the noindex and nofollow values that tell search engine
// crawlers to not index or crawl the page or asset.
// See https://developers.google.com/search/docs/crawling-indexing/robots-meta-tag#xrobotstag
func NoRobotsHeader(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		const HeaderXRobotsTag = "X-Robots-Tag"
		c.Response().Header().Set(HeaderXRobotsTag, "noindex, nofollow")
		return next(c)
	}
}
