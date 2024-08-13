package handler

import (
	"database/sql"
	"fmt"

	"github.com/Defacto2/server/handler/app"
	"github.com/Defacto2/server/handler/htmx"
	"github.com/Defacto2/server/internal/command"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// Package file routerlock.go contains the custom router URIs for the website
// that are locked behind the router middleware and require a user to be logged in.

/*
	A note about the request methods in use:
	 - GET requests are used for retrieving data from the server.
	 - PATCH requests are used for updating or retrieving data on the server.
	 - PUT requests are used for creating new data on the server.
	 - POST requests are used for uploading files with or without data.
	 - DELETE requests are used for removing data from the server.
*/

func (c Configuration) lock(e *echo.Echo, db *sql.DB, logger *zap.SugaredLogger, dir app.Dirs) *echo.Echo {
	if e == nil {
		panic(fmt.Errorf("%w for lock router", ErrRoutes))
	}
	lock := e.Group("/editor")
	lock.Use(c.ReadOnlyLock, c.SessionLock)
	c.configurations(lock, db)
	creator(lock)
	date(lock)
	editor(lock, db, logger, dir)
	get(lock, db, dir)
	online(lock)
	search(lock, db, logger)
	return e
}

func (c Configuration) configurations(g *echo.Group, db *sql.DB) {
	if g == nil {
		panic(fmt.Errorf("%w for configurations router", ErrRoutes))
	}
	conf := g.Group("/configurations")
	conf.GET("", func(cx echo.Context) error {
		return app.Configurations(cx, db, c.Environment)
	})
	conf.GET("/dbconns", func(c echo.Context) error {
		return htmx.DBConnections(c, db)
	})
	conf.GET("/pings", func(cx echo.Context) error {
		proto := "http"
		port := c.Environment.HTTPPort
		if port == 0 {
			port = c.Environment.TLSPort
			proto = "https"
		}
		return htmx.Pings(cx, proto, int(port))
	})
}

func creator(g *echo.Group) {
	if g == nil {
		panic(ErrRoutes)
	}
	creator := g.Group("/creator")
	creator.PATCH("/text", htmx.RecordCreatorText)
	creator.PATCH("/ill", htmx.RecordCreatorIll)
	creator.PATCH("/prog", htmx.RecordCreatorProg)
	creator.PATCH("/audio", htmx.RecordCreatorAudio)
	creator.PATCH("/reset", htmx.RecordCreatorReset)
}

func date(g *echo.Group) {
	if g == nil {
		panic(fmt.Errorf("%w for date router", ErrRoutes))
	}
	date := g.Group("/date")
	date.PATCH("", htmx.RecordDateIssued)
	date.PATCH("/reset", func(cx echo.Context) error {
		return htmx.RecordDateIssuedReset(cx, "artifact-editor-date-resetter")
	})
	date.PATCH("/lastmod", func(cx echo.Context) error {
		return htmx.RecordDateIssuedReset(cx, "artifact-editor-date-lastmodder")
	})
}

func editor(g *echo.Group, db *sql.DB, logger *zap.SugaredLogger, dir app.Dirs) {
	if g == nil {
		panic(fmt.Errorf("%w for editor router", ErrRoutes))
	}
	g.DELETE("/delete/forever/:key", func(c echo.Context) error {
		return htmx.DeleteForever(c, logger, c.Param("key"))
	})
	g.PATCH("/16colors", htmx.Record16Colors)
	g.PATCH("/classifications", func(c echo.Context) error {
		return htmx.RecordClassification(c, db, logger)
	})
	g.PATCH("/comment", htmx.RecordComment)
	g.PATCH("/comment/reset", htmx.RecordCommentReset)
	g.PATCH("/demozoo", htmx.RecordDemozoo)
	g.PATCH("/filename", htmx.RecordFilename)
	g.PATCH("/filename/reset", htmx.RecordFilenameReset)
	g.PATCH("/github", htmx.RecordGitHub)
	g.PATCH("/links", htmx.RecordLinks)
	g.PATCH("/links/reset", htmx.RecordLinksReset)
	g.PATCH("/platform", func(c echo.Context) error {
		return app.PlatformEdit(c, db)
	})
	g.PATCH("/platform+tag", app.PlatformTagInfo)
	g.PATCH("/pouet", htmx.RecordPouet)
	g.PATCH("/relations", htmx.RecordRelations)
	g.PATCH("/releasers", htmx.RecordReleasers)
	g.PATCH("/releasers/reset", htmx.RecordReleasersReset)
	g.PATCH("/sites", htmx.RecordSites)
	g.PATCH("/tag", func(c echo.Context) error {
		return app.TagEdit(c, db)
	})
	g.PATCH("/tag/info", app.TagInfo)
	g.PATCH("/title", htmx.RecordTitle)
	g.PATCH("/title/reset", htmx.RecordTitleReset)
	g.PATCH("/virustotal", htmx.RecordVirusTotal)
	g.PATCH("/ymd", app.YMDEdit)
	g.PATCH("/youtube", htmx.RecordYouTube)

	emu := g.Group("/emulate")
	emu.PATCH("/broken/:id", htmx.RecordEmulateBroken)
	emu.PATCH("/runprogram/:id", htmx.RecordEmulateRunProgram)
	emu.PATCH("/machine/:id", htmx.RecordEmulateMachine)
	emu.PATCH("/cpu/:id", htmx.RecordEmulateCPU)
	emu.PATCH("/sfx/:id", htmx.RecordEmulateSFX)
	emu.PATCH("/umb/:id", htmx.RecordEmulateUMB)
	emu.PATCH("/ems/:id", htmx.RecordEmulateEMS)
	emu.PATCH("/xms/:id", htmx.RecordEmulateXMS)

	// these POSTs should only be used for editor, htmx file uploads,
	// and not for general file uploads or data edits.
	upload := g.Group("/upload")
	upload.POST("/file", func(c echo.Context) error {
		return htmx.UploadReplacement(c, dir.Download)
	})
	upload.POST("/preview", func(c echo.Context) error {
		return htmx.UploadPreview(c, dir.Preview, dir.Thumbnail)
	})

	dirs := command.Dirs{
		Download:  dir.Download,
		Preview:   dir.Preview,
		Thumbnail: dir.Thumbnail,
	}
	me := g.Group("/readme")
	me.PATCH("/copy/:unid/:path", func(c echo.Context) error {
		return htmx.RecordReadmeCopier(c, dir.Extra)
	})
	me.PATCH("/preview/:unid/:path", func(c echo.Context) error {
		return htmx.RecordReadmeImager(c, logger, dirs)
	})
	me.DELETE("/:unid", func(c echo.Context) error {
		return htmx.RecordReadmeDeleter(c, dir.Extra)
	})
	// readme.POST("/hide", func(cx echo.Context) error {
	// 	dir.URI = cx.Param("id")
	// 	return app.ReadmeToggle(cx)
	// })

	pre := g.Group("/preview")
	pre.PATCH("/copy/:unid/:path", func(c echo.Context) error {
		return htmx.RecordImageCopier(c, logger, dirs)
	})
	pre.PATCH("/crop11/:unid", func(c echo.Context) error {
		return htmx.RecordImageCropper(c, command.SqaureTop, dirs)
	})
	pre.PATCH("/crop43/:unid", func(c echo.Context) error {
		return htmx.RecordImageCropper(c, command.FourThree, dirs)
	})
	pre.PATCH("/crop12/:unid", func(c echo.Context) error {
		return htmx.RecordImageCropper(c, command.OneTwo, dirs)
	})
	pre.DELETE("/:unid", func(c echo.Context) error {
		return htmx.RecordImagesDeleter(c, dir.Preview)
	})

	thumb := g.Group("/thumbnail")
	thumb.PATCH("/copy/:unid/:path", func(c echo.Context) error {
		return htmx.RecordImageCopier(c, logger, dirs)
	})
	thumb.PATCH("/top/:unid", func(c echo.Context) error {
		return htmx.RecordThumbAlignment(c, command.Top, dirs)
	})
	thumb.PATCH("/middle/:unid", func(c echo.Context) error {
		return htmx.RecordThumbAlignment(c, command.Middle, dirs)
	})
	thumb.PATCH("/bottom/:unid", func(c echo.Context) error {
		return htmx.RecordThumbAlignment(c, command.Bottom, dirs)
	})
	thumb.PATCH("/left/:unid", func(c echo.Context) error {
		return htmx.RecordThumbAlignment(c, command.Left, dirs)
	})
	thumb.PATCH("/right/:unid", func(c echo.Context) error {
		return htmx.RecordThumbAlignment(c, command.Right, dirs)
	})
	thumb.PATCH("/pixel/:unid", func(c echo.Context) error {
		return htmx.RecordThumb(c, command.Pixel, dirs)
	})
	thumb.PATCH("/photo/:unid", func(c echo.Context) error {
		return htmx.RecordThumb(c, command.Photo, dirs)
	})
	thumb.DELETE("/:unid", func(c echo.Context) error {
		return htmx.RecordImagesDeleter(c, dir.Thumbnail)
	})

	imgs := g.Group("/images")
	imgs.PATCH("/pixelate/:unid", func(c echo.Context) error {
		return htmx.RecordImagePixelator(c, dir.Preview, dir.Thumbnail)
	})
	imgs.DELETE("/:unid", func(c echo.Context) error {
		return htmx.RecordImagesDeleter(c, dir.Preview, dir.Thumbnail)
	})
}

func get(g *echo.Group, db *sql.DB, dir app.Dirs) {
	if g == nil {
		panic(fmt.Errorf("%w for get router", ErrRoutes))
	}
	g.GET("/deletions",
		func(cx echo.Context) error {
			return app.Deletions(cx, db, "1")
		})
	g.GET("/get/demozoo/download/:id",
		func(cx echo.Context) error {
			return app.GetDemozooLink(cx, dir.Download)
		})
	g.GET("/for-approval",
		func(cx echo.Context) error {
			return app.ForApproval(cx, db, "1")
		})
	g.GET("/unwanted",
		func(cx echo.Context) error {
			return app.Unwanted(cx, db, "1")
		})
}

func online(g *echo.Group) {
	if g == nil {
		panic(fmt.Errorf("%w for online router", ErrRoutes))
	}
	online := g.Group("/online")
	online.PATCH("/true", func(cx echo.Context) error {
		return htmx.RecordToggle(cx, true)
	})
	online.PATCH("/false", func(cx echo.Context) error {
		return htmx.RecordToggle(cx, false)
	})
	online.GET("/true/:id", func(cx echo.Context) error {
		return htmx.RecordToggleByID(cx, cx.Param("id"), true)
	})
}

func search(g *echo.Group, db *sql.DB, logger *zap.SugaredLogger) {
	if g == nil {
		panic(fmt.Errorf("%w for search router", ErrRoutes))
	}
	search := g.Group("/search")
	search.GET("/id", app.SearchID)
	search.PATCH("/id", func(cx echo.Context) error {
		return htmx.SearchByID(cx, db, logger)
	})
}
