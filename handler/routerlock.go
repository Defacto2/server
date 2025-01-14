package handler

import (
	"database/sql"
	"fmt"
	"math"

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

func (c *Configuration) lock(e *echo.Echo, db *sql.DB, logger *zap.SugaredLogger, dir app.Dirs) *echo.Echo {
	if e == nil {
		panic(fmt.Errorf("%w for lock router", ErrRoutes))
	}
	lock := e.Group("/editor")
	lock.Use(c.ReadOnlyLock, c.SessionLock)
	c.configurations(lock, db)
	creator(lock, db)
	date(lock, db)
	editor(lock, db, logger, dir)
	get(lock, db, dir)
	online(lock, db)
	search(lock, db, logger)
	return e
}

func (c *Configuration) configurations(g *echo.Group, db *sql.DB) {
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
		return htmx.Pings(cx, proto, int(math.Abs(float64(port))))
	})
}

func creator(g *echo.Group, db *sql.DB) {
	if g == nil {
		panic(ErrRoutes)
	}
	creator := g.Group("/creator")
	creator.PATCH("/text", func(c echo.Context) error {
		return htmx.RecordCreatorText(c, db)
	})
	creator.PATCH("/ill", func(c echo.Context) error {
		return htmx.RecordCreatorIll(c, db)
	})
	creator.PATCH("/prog", func(c echo.Context) error {
		return htmx.RecordCreatorProg(c, db)
	})
	creator.PATCH("/audio", func(c echo.Context) error {
		return htmx.RecordCreatorAudio(c, db)
	})
	creator.PATCH("/reset", func(c echo.Context) error {
		return htmx.RecordCreatorReset(c, db)
	})
}

func date(g *echo.Group, db *sql.DB) {
	if g == nil {
		panic(fmt.Errorf("%w for date router", ErrRoutes))
	}
	date := g.Group("/date")
	date.PATCH("", func(c echo.Context) error {
		return htmx.RecordDateIssued(c, db)
	})
	date.PATCH("/reset", func(cx echo.Context) error {
		return htmx.RecordDateIssuedReset(cx, db, "artifact-editor-date-resetter")
	})
	date.PATCH("/lastmod", func(cx echo.Context) error {
		return htmx.RecordDateIssuedReset(cx, db, "artifact-editor-date-lastmodder")
	})
}

func editor(g *echo.Group, db *sql.DB, logger *zap.SugaredLogger, dir app.Dirs) {
	if g == nil {
		panic(fmt.Errorf("%w for editor router", ErrRoutes))
	}
	g.DELETE("/delete/forever/:key", func(c echo.Context) error {
		return htmx.DeleteForever(c, db, logger, c.Param("key"))
	})
	g.PATCH("/16colors", func(c echo.Context) error {
		return htmx.Record16Colors(c, db)
	})
	g.PATCH("/classifications", func(c echo.Context) error {
		return htmx.RecordClassification(c, db, logger)
	})
	g.PATCH("/comment", func(c echo.Context) error {
		return htmx.RecordComment(c, db)
	})
	g.PATCH("/comment/reset", func(c echo.Context) error {
		return htmx.RecordCommentReset(c, db)
	})
	g.PATCH("/demozoo", func(c echo.Context) error {
		return htmx.RecordDemozoo(c, db)
	})
	g.PATCH("/filename", func(c echo.Context) error {
		return htmx.RecordFilename(c, db)
	})
	g.PATCH("/filename/reset", func(c echo.Context) error {
		return htmx.RecordFilenameReset(c, db)
	})
	g.PATCH("/github", func(c echo.Context) error {
		return htmx.RecordGitHub(c, db)
	})
	g.PATCH("/links", htmx.RecordLinks)
	g.PATCH("/links/reset", func(c echo.Context) error {
		return htmx.RecordLinksReset(c, db)
	})
	g.PATCH("/platform", func(c echo.Context) error {
		return app.PlatformEdit(c, db)
	})
	g.PATCH("/platform+tag", app.PlatformTagInfo)
	g.PATCH("/pouet", func(c echo.Context) error {
		return htmx.RecordPouet(c, db)
	})
	g.PATCH("/relations", func(c echo.Context) error {
		return htmx.RecordRelations(c, db)
	})
	g.PATCH("/releasers", func(c echo.Context) error {
		return htmx.RecordReleasers(c, db)
	})
	g.PATCH("/releasers/reset", func(c echo.Context) error {
		return htmx.RecordReleasersReset(c, db)
	})
	g.PATCH("/sites", func(c echo.Context) error {
		return htmx.RecordSites(c, db)
	})
	g.PATCH("/tag", func(c echo.Context) error {
		return app.TagEdit(c, db)
	})
	g.PATCH("/tag/info", app.TagInfo)
	g.PATCH("/title", func(c echo.Context) error {
		return htmx.RecordTitle(c, db)
	})
	g.PATCH("/title/reset", func(c echo.Context) error {
		return htmx.RecordTitleReset(c, db)
	})
	g.PATCH("/virustotal", func(c echo.Context) error {
		return htmx.RecordVirusTotal(c, db)
	})
	g.PATCH("/ymd", func(c echo.Context) error {
		return app.YMDEdit(c, db)
	})
	g.PATCH("/youtube", func(c echo.Context) error {
		return htmx.RecordYouTube(c, db)
	})

	emu := g.Group("/emulate")
	emu.PATCH("/broken/:id", func(c echo.Context) error {
		return htmx.RecordEmulateBroken(c, db)
	})
	emu.PATCH("/runprogram/:id", func(c echo.Context) error {
		return htmx.RecordEmulateRunProgram(c, db)
	})
	emu.PATCH("/machine/:id", func(c echo.Context) error {
		return htmx.RecordEmulateMachine(c, db)
	})
	emu.PATCH("/cpu/:id", func(c echo.Context) error {
		return htmx.RecordEmulateCPU(c, db)
	})
	emu.PATCH("/sfx/:id", func(c echo.Context) error {
		return htmx.RecordEmulateSFX(c, db)
	})
	emu.PATCH("/umb/:id", func(c echo.Context) error {
		return htmx.RecordEmulateUMB(c, db)
	})
	emu.PATCH("/ems/:id", func(c echo.Context) error {
		return htmx.RecordEmulateEMS(c, db)
	})
	emu.PATCH("/xms/:id", func(c echo.Context) error {
		return htmx.RecordEmulateXMS(c, db)
	})

	// these POSTs should only be used for editor, htmx file uploads,
	// and not for general file uploads or data edits.
	upload := g.Group("/upload")
	// /upload/file
	upload.POST("/file", func(c echo.Context) error {
		return htmx.UploadReplacement(c, db, dir.Download, dir.Extra)
	})
	// /upload/preview
	upload.POST("/preview", func(c echo.Context) error {
		return htmx.UploadPreview(c, dir.Preview, dir.Thumbnail)
	})
	dirs := command.Dirs{
		Download:  dir.Download,
		Preview:   dir.Preview,
		Thumbnail: dir.Thumbnail,
		Extra:     dir.Extra,
	}
	diz := g.Group("/diz")
	diz.PATCH("/copy/:unid/:path", func(c echo.Context) error {
		return htmx.RecordDizCopier(c, dirs)
	})
	diz.DELETE("/:unid", func(c echo.Context) error {
		return htmx.RecordDizDeleter(c, dir.Extra)
	})
	readme := g.Group("/readme")
	readme.PATCH("/disable/:id", func(c echo.Context) error {
		return htmx.RecordReadmeDisable(c, db)
	})
	readme.PATCH("/copy/:unid/:path", func(c echo.Context) error {
		return htmx.RecordReadmeCopier(c, dirs)
	})
	// /editor/readme/preview
	readme.PATCH("/preview/:unid/:path", func(c echo.Context) error {
		return htmx.RecordReadmeImager(c, logger, false, dirs)
	})
	// /editor/readme/preview-amiga
	readme.PATCH("/preview-amiga/:unid/:path", func(c echo.Context) error {
		return htmx.RecordReadmeImager(c, logger, true, dirs)
	})
	readme.DELETE("/:unid", func(c echo.Context) error {
		return htmx.RecordReadmeDeleter(c, dir.Extra)
	})
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
	g.GET("/get/demozoo/download/:unid/:id",
		func(cx echo.Context) error {
			return app.GetDemozooParam(cx, db, dir.Download)
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

func online(g *echo.Group, db *sql.DB) {
	if g == nil {
		panic(fmt.Errorf("%w for online router", ErrRoutes))
	}
	online := g.Group("/online")
	online.PATCH("/true", func(cx echo.Context) error {
		return htmx.RecordToggle(cx, db, true)
	})
	online.PATCH("/false", func(cx echo.Context) error {
		return htmx.RecordToggle(cx, db, false)
	})
	online.GET("/true/:id", func(cx echo.Context) error {
		return htmx.RecordToggleByID(cx, db, cx.Param("id"), true)
	})
}

func search(g *echo.Group, db *sql.DB, logger *zap.SugaredLogger) {
	if g == nil {
		panic(fmt.Errorf("%w for search router", ErrRoutes))
	}
	search := g.Group("/search")
	search.GET("/id", app.SearchID)
	search.POST("/id", func(cx echo.Context) error {
		return htmx.SearchByID(cx, db, logger)
	})
}
