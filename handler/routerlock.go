package handler

import (
	"fmt"

	"github.com/Defacto2/server/handler/app"
	"github.com/Defacto2/server/handler/htmx"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// Package file routerlock.go contains the custom router URIs for the website
// that are locked behind the router middleware and require a user to be logged in.

func (c Configuration) lock(e *echo.Echo, logger *zap.SugaredLogger, dir app.Dirs) *echo.Echo {
	if e == nil {
		panic(fmt.Errorf("%w for lock router", ErrRoutes))
	}
	lock := e.Group("/editor")
	lock.Use(c.ReadOnlyLock, c.SessionLock)
	lock.GET("/configs", func(cx echo.Context) error {
		return app.Configurations(cx, c.Environment)
	})
	lock.GET("/configs/pings", func(cx echo.Context) error {
		proto := "http"
		port := c.Environment.HTTPPort
		if port == 0 {
			port = c.Environment.TLSPort
			proto = "https"
		}
		return htmx.Pings(cx, proto, int(port))
	})
	creator(lock)
	date(lock)
	editor(lock, logger, dir)
	get(lock, dir)
	images(lock, logger, dir)
	online(lock)
	readme(lock, logger, dir)
	search(lock, logger)
	return e
}

func creator(g *echo.Group) {
	if g == nil {
		panic(ErrRoutes)
	}
	creator := g.Group("/creator")
	creator.POST("/text", htmx.RecordCreatorText)
	creator.POST("/ill", htmx.RecordCreatorIll)
	creator.POST("/prog", htmx.RecordCreatorProg)
	creator.POST("/audio", htmx.RecordCreatorAudio)
	creator.POST("/reset", htmx.RecordCreatorReset)
}

func date(g *echo.Group) {
	if g == nil {
		panic(fmt.Errorf("%w for date router", ErrRoutes))
	}
	date := g.Group("/date")
	date.POST("", htmx.RecordDateIssued)
	date.POST("/reset", func(cx echo.Context) error {
		return htmx.RecordDateIssuedReset(cx, "artifact-editor-date-resetter")
	})
	date.POST("/lastmod", func(cx echo.Context) error {
		return htmx.RecordDateIssuedReset(cx, "artifact-editor-date-lastmodder")
	})
}

func editor(g *echo.Group, logger *zap.SugaredLogger, dir app.Dirs) {
	if g == nil {
		panic(fmt.Errorf("%w for editor router", ErrRoutes))
	}
	g.POST("/16colors", htmx.Record16Colors)
	g.POST("/ansilove/copy", func(c echo.Context) error {
		return dir.AnsiLovePost(c, logger)
	})
	g.POST("/classifications", func(c echo.Context) error {
		return htmx.RecordClassification(c, logger)
	})
	g.POST("/comment", htmx.RecordComment)
	g.POST("/comment/reset", htmx.RecordCommentReset)
	g.POST("/delete/permenantly", func(c echo.Context) error {
		return htmx.PermenantDelete(c, logger)
	})
	g.POST("/demozoo", htmx.RecordDemozoo)
	g.POST("/filename", htmx.RecordFilename)
	g.POST("/filename/reset", htmx.RecordFilenameReset)
	g.POST("/github", htmx.RecordGitHub)
	g.POST("/links", htmx.RecordLinks)
	g.POST("/links/reset", htmx.RecordLinksReset)
	g.POST("/platform", app.PlatformEdit)
	g.POST("/platform+tag", app.PlatformTagInfo)
	g.POST("/pouet", htmx.RecordPouet)
	g.POST("/relations", htmx.RecordRelations)
	g.POST("/releasers", htmx.RecordReleasers)
	g.POST("/releasers/reset", htmx.RecordReleasersReset)
	g.POST("/sites", htmx.RecordSites)
	g.POST("/tag", app.TagEdit)
	g.POST("/tag/info", app.TagInfo)
	g.POST("/title", htmx.RecordTitle)
	g.POST("/title/reset", htmx.RecordTitleReset)
	g.POST("/virustotal", htmx.RecordVirusTotal)
	g.POST("/ymd", app.YMDEdit)
	g.POST("/youtube", htmx.RecordYouTube)
}

func get(g *echo.Group, dir app.Dirs) {
	if g == nil {
		panic(fmt.Errorf("%w for get router", ErrRoutes))
	}
	g.GET("/deletions",
		func(cx echo.Context) error {
			return app.Deletions(cx, "1")
		})
	g.GET("/get/demozoo/download/:id",
		func(cx echo.Context) error {
			return app.GetDemozooLink(cx, dir.Download)
		})
	g.GET("/for-approval",
		func(cx echo.Context) error {
			return app.ForApproval(cx, "1")
		})
	g.GET("/unwanted",
		func(cx echo.Context) error {
			return app.Unwanted(cx, "1")
		})
}

func images(g *echo.Group, logger *zap.SugaredLogger, dir app.Dirs) {
	if g == nil {
		panic(fmt.Errorf("%w for images router", ErrRoutes))
	}
	images := g.Group("/images")
	images.POST("/copy", func(c echo.Context) error {
		return dir.PreviewPost(c, logger)
	})
	images.POST("/delete", dir.PreviewDel)
}

func online(g *echo.Group) {
	if g == nil {
		panic(fmt.Errorf("%w for online router", ErrRoutes))
	}
	online := g.Group("/online")
	online.POST("/true", func(cx echo.Context) error {
		return htmx.RecordToggle(cx, true)
	})
	online.POST("/false", func(cx echo.Context) error {
		return htmx.RecordToggle(cx, false)
	})
}

func search(g *echo.Group, logger *zap.SugaredLogger) {
	if g == nil {
		panic(fmt.Errorf("%w for search router", ErrRoutes))
	}
	search := g.Group("/search")
	search.GET("/id", app.SearchID)
	search.POST("/id", func(cx echo.Context) error {
		return htmx.SearchByID(cx, logger)
	})
	// search.POST("/reset", func(cx echo.Context) error {
	// 	return htmx.RecordSearchReset(cx)
	// })
}

func readme(g *echo.Group, logger *zap.SugaredLogger, dir app.Dirs) {
	if g == nil {
		panic(fmt.Errorf("%w for readme router", ErrRoutes))
	}
	readme := g.Group("/readme")
	readme.POST("/copy", func(cx echo.Context) error {
		return app.ReadmePost(cx, logger, dir.Download)
	})
	readme.POST("/delete", func(cx echo.Context) error {
		return app.ReadmeDel(cx, dir.Download)
	})
	readme.POST("/hide", func(cx echo.Context) error {
		dir.URI = cx.Param("id")
		return app.ReadmeToggle(cx)
	})
}
