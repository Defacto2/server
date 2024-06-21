package handler

import (
	"fmt"

	"github.com/Defacto2/server/handler/htmx"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
)

// Package file routerhtmx.go contains the custom router URIs for the website
// that use the htmx ajax library and require a rate limiter middleware.

const rateLimit = 2

// htmxGroup is the /htmx sub-route group that returns HTML fragments
// using the htmx library for AJAX responses.
func htmxGroup(e *echo.Echo, logger *zap.SugaredLogger, prod bool, downloadDir string) *echo.Echo {
	if e == nil {
		panic(fmt.Errorf("%w for htmx group router", ErrRoutes))
	}
	store := middleware.NewRateLimiterMemoryStore(rateLimit)
	g := e.Group("", middleware.RateLimiter(store))
	g.PUT("/uploader/sha384/:hash", func(c echo.Context) error {
		return htmx.LookupSHA384(c, logger)
	})
	g.POST("/demozoo/production", htmx.DemozooProd)
	g.POST("/demozoo/production/submit/:id", func(c echo.Context) error {
		return htmx.DemozooSubmit(c, logger)
	})
	g.POST("/pouet/production", htmx.PouetProd)
	g.POST("/pouet/production/submit/:id", func(c echo.Context) error {
		return htmx.PouetSubmit(c, logger)
	})
	g.POST("/search/releaser", func(c echo.Context) error {
		return htmx.SearchReleaser(c, logger)
	})
	g.POST("/uploader/advanced", func(c echo.Context) error {
		return htmx.AdvancedSubmit(c, logger, prod, downloadDir)
	})
	g.POST("/uploader/classifications", func(c echo.Context) error {
		return htmx.HumanizeAndCount(c, logger, "uploader-advanced")
	})
	g.POST("/uploader/image", func(c echo.Context) error {
		return htmx.ImageSubmit(c, logger, prod, downloadDir)
	})
	g.POST("/uploader/intro", func(c echo.Context) error {
		return htmx.IntroSubmit(c, logger, prod, downloadDir)
	})
	g.POST("/uploader/magazine", func(c echo.Context) error {
		return htmx.MagazineSubmit(c, logger, prod, downloadDir)
	})
	g.POST("/uploader/releaser/1", func(c echo.Context) error {
		return htmx.DataListReleasers(c, logger, releaser1(c))
	})
	g.POST("/uploader/releaser/2", func(c echo.Context) error {
		return htmx.DataListReleasers(c, logger, releaser2(c))
	})
	g.POST("/uploader/releaser/magazine", func(c echo.Context) error {
		lookup := c.FormValue("uploader-magazine-releaser1")
		return htmx.DataListMagazines(c, logger, lookup)
	})
	g.POST("/uploader/text", func(c echo.Context) error {
		return htmx.TextSubmit(c, logger, prod, downloadDir)
	})
	g.POST("/uploader/trainer", func(c echo.Context) error {
		return htmx.TrainerSubmit(c, logger, prod, downloadDir)
	})
	return e
}

func releaser1(c echo.Context) string {
	lookups := []string{
		"artifact-editor-releaser1",
		"uploader-intro-releaser1",
		"uploader-trainer-releaser1",
		"uploader-text-releaser1",
		"uploader-image-releaser1",
		"uploader-advanced-releaser1",
	}
	for _, lookup := range lookups {
		if val := c.FormValue(lookup); val != "" {
			return val
		}
	}
	return ""
}

func releaser2(c echo.Context) string {
	lookups := []string{
		"artifact-editor-releaser2",
		"uploader-intro-releaser2",
		"uploader-trainer-releaser2",
		"uploader-text-releaser2",
		"uploader-image-releaser2",
		"uploader-advanced-releaser2",
	}
	for _, lookup := range lookups {
		if val := c.FormValue(lookup); val != "" {
			return val
		}
	}
	return ""
}
