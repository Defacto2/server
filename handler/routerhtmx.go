package handler

import (
	"database/sql"
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
func htmxGroup(e *echo.Echo, db *sql.DB, logger *zap.SugaredLogger, downloadDir string) *echo.Echo {
	if e == nil {
		panic(fmt.Errorf("%w for htmx group router", ErrRoutes))
	}
	store := middleware.NewRateLimiterMemoryStore(rateLimit)
	g := e.Group("", middleware.RateLimiter(store))
	g.PATCH("/search/releaser", func(c echo.Context) error {
		return htmx.SearchReleaser(c, db, logger)
	})

	demozoo := g.Group("/demozoo")
	demozoo.GET("/production", func(c echo.Context) error {
		return htmx.DemozooProd(c, db)
	})
	demozoo.PUT("/production/:id", func(c echo.Context) error {
		return htmx.DemozooSubmit(c, logger)
	})
	pouet := g.Group("/pouet")
	pouet.GET("/production", func(c echo.Context) error {
		return htmx.PouetProd(c, db)
	})
	pouet.PUT("/production/:id", func(c echo.Context) error {
		return htmx.PouetSubmit(c, logger)
	})

	upload := g.Group("/uploader")
	upload.GET("/classifications", func(c echo.Context) error {
		return htmx.HumanizeCount(c, db, logger, "uploader-advanced")
	})
	upload.PATCH("/releaser/1", func(c echo.Context) error {
		return htmx.DataListReleasers(c, db, logger, releaser1(c))
	})
	upload.PATCH("/releaser/2", func(c echo.Context) error {
		return htmx.DataListReleasers(c, db, logger, releaser2(c))
	})
	upload.PATCH("/releaser/magazine", func(c echo.Context) error {
		lookup := c.FormValue("uploader-magazine-releaser1")
		return htmx.DataListMagazines(c, db, logger, lookup)
	})
	upload.PATCH("/sha384/:hash", func(c echo.Context) error {
		return htmx.LookupSHA384(c, db, logger)
	})
	upload.POST("/advanced", func(c echo.Context) error {
		return htmx.AdvancedSubmit(c, logger, downloadDir)
	})
	upload.POST("/image", func(c echo.Context) error {
		return htmx.ImageSubmit(c, logger, downloadDir)
	})
	upload.POST("/intro", func(c echo.Context) error {
		return htmx.IntroSubmit(c, logger, downloadDir)
	})
	upload.POST("/magazine", func(c echo.Context) error {
		return htmx.MagazineSubmit(c, logger, downloadDir)
	})
	upload.POST("/text", func(c echo.Context) error {
		return htmx.TextSubmit(c, logger, downloadDir)
	})
	upload.POST("/trainer", func(c echo.Context) error {
		return htmx.TrainerSubmit(c, logger, downloadDir)
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
