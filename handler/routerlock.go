package handler

import (
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/Defacto2/server/handler/app"
	"github.com/Defacto2/server/handler/htmx"
	"github.com/Defacto2/server/internal/command"
	"github.com/Defacto2/server/internal/nils"
	"github.com/labstack/echo/v5"
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

func (serv *Server) lock(sl *slog.Logger, e *echo.Echo, db *sql.DB, dirs app.Dirs) (*echo.Echo, error) {
	const format = "configuration router: %w"
	if err := nils.Check(sl, e, db); err != nil {
		return nil, fmt.Errorf(format, err)
	}

	readonlylock := func(next echo.HandlerFunc) echo.HandlerFunc {
		return serv.ReadOnlyLock(next, sl)
	}

	sessionlock := func(next echo.HandlerFunc) echo.HandlerFunc {
		return serv.SessionLock(next, sl)
	}

	lock := e.Group("/editor") // lock this group route

	lock.Use(readonlylock, sessionlock)

	if err := serv.configurations(sl, lock, db); err != nil {
		return nil, fmt.Errorf(format, err)
	}
	if err := creator(lock, db); err != nil {
		return nil, fmt.Errorf(format, err)
	}
	if err := date(lock, db); err != nil {
		return nil, fmt.Errorf(format, err)
	}
	if err := editor(sl, lock, db, dirs); err != nil {
		return nil, fmt.Errorf(format, err)
	}
	if err := get(sl, lock, db, dirs); err != nil {
		return nil, fmt.Errorf(format, err)
	}
	if err := online(lock, db); err != nil {
		return nil, fmt.Errorf(format, err)
	}
	if err := search(sl, lock, db); err != nil {
		return nil, fmt.Errorf(format, err)
	}

	fixers(sl, lock, db)

	routes(sl, lock, e.Router().Routes())

	return e, nil
}

func routes(sl *slog.Logger, g *echo.Group, r echo.Routes) {
	g.GET("/routes", func(c *echo.Context) error {
		return app.Routes(sl, c, r)
	})
}

func fixers(sl *slog.Logger, g *echo.Group, db *sql.DB) {
	fixers := func(c *echo.Context) error {
		return app.Fixers(sl, c, db)
	}
	fixID := func(c *echo.Context) error {
		return app.FixNumericSuffix(sl, c, db)
	}
	g.GET("/fixers", fixers)
	g.POST("/fixers/fix/:id", fixID)
}

func (serv *Server) configurations(sl *slog.Logger, g *echo.Group, db *sql.DB) error {
	const format = "configurations group router: %w"
	if err := nils.Check(sl, g, db); err != nil {
		return fmt.Errorf(format, err)
	}

	conf := g.Group("/configurations")
	conf.GET("", func(c *echo.Context) error {
		return app.Configurations(sl, c, db, serv.Environment)
	})

	conf.GET("/dbconns", func(c *echo.Context) error {
		return htmx.DBConnections(c, db)
	})

	conf.GET("/pings", func(c *echo.Context) error {
		proto := "http"
		port := serv.Environment.HTTPPort.Value()
		if port == 0 {
			port = serv.Environment.TLSPort.Value()
			proto = "https"
		}
		return htmx.Pings(c, proto, int(port))
	})

	return nil
}

func creator(g *echo.Group, db *sql.DB) error {
	const format = "creator group router: %w"
	if err := nils.Check(g, db); err != nil {
		return fmt.Errorf(format, err)
	}

	creator := g.Group("/creator")
	creator.PATCH("/text", func(c *echo.Context) error {
		return BeginTx(c, db, func(tx *sql.Tx) error {
			return htmx.TxCreditText(c, tx)
		})
	})
	creator.PATCH("/ill", func(c *echo.Context) error {
		return BeginTx(c, db, func(tx *sql.Tx) error {
			return htmx.TxCreditIll(c, tx)
		})
	})
	creator.PATCH("/prog", func(c *echo.Context) error {
		return BeginTx(c, db, func(tx *sql.Tx) error {
			return htmx.TxCreditProg(c, tx)
		})
	})
	creator.PATCH("/audio", func(c *echo.Context) error {
		return BeginTx(c, db, func(tx *sql.Tx) error {
			return htmx.TxCreditAudio(c, tx)
		})
	})
	creator.PATCH("/reset", func(c *echo.Context) error {
		return BeginTx(c, db, func(tx *sql.Tx) error {
			return htmx.TxCreditUndo(c, tx)
		})
	})

	return nil
}

func date(g *echo.Group, db *sql.DB) error {
	if err := nils.Check(g, db); err != nil {
		return fmt.Errorf("date router: %w", err)
	}

	date := g.Group("/date")
	date.PATCH("", func(c *echo.Context) error {
		return BeginTx(c, db, func(tx *sql.Tx) error {
			return htmx.TxYMD(c, tx)
		})
	})
	date.PATCH("/reset", func(c *echo.Context) error {
		return BeginTx(c, db, func(tx *sql.Tx) error {
			return htmx.TxYMDUndo(c, tx)
		})
	})
	date.PATCH("/lastmod", func(c *echo.Context) error {
		return BeginTx(c, db, func(tx *sql.Tx) error {
			return htmx.TxLastMod(c, tx)
		})
	})
	return nil
}

func editor(sl *slog.Logger, g *echo.Group, db *sql.DB, dirs app.Dirs) error { //nolint:funlen
	if err := nils.Check(sl, g, db); err != nil {
		return fmt.Errorf("editor router: %w", err)
	}

	g.DELETE("/delete/forever/:key", func(c *echo.Context) error {
		return BeginTx(c, db, func(tx *sql.Tx) error {
			return htmx.DeleteForever(sl, c, tx, c.Param("key"))
		})
	})

	// these POSTs should only be used for editor, htmx file uploads,
	// and not for general file uploads or data edits.
	upload := g.Group("/upload")
	// /upload/file
	upload.POST("/file", func(c *echo.Context) error {
		u := htmx.Submit{
			Download:  dirs.Download,
			Extra:     dirs.Extra,
			Preview:   "",
			Thumbnail: "",
		}
		return BeginTx(c, db, func(tx *sql.Tx) error {
			return u.Replacement(sl, c, tx)
		})
	})
	// /upload/preview
	upload.POST("/preview", func(c *echo.Context) error { //nolint:contextcheck
		u := htmx.Submit{
			Download:  "",
			Extra:     "",
			Preview:   dirs.Preview,
			Thumbnail: dirs.Thumbnail,
		}
		return u.Image(sl, c)
	})

	editorPatch(sl, g, db)
	editorEmu(g, db)
	editorImgs(sl, g, dirs)

	paths := command.Dirs{
		Download:  dirs.Download,
		Preview:   dirs.Preview,
		Thumbnail: dirs.Thumbnail,
		Extra:     dirs.Extra,
	}
	editorReadme(sl, g, db, paths, dirs)
	editorPreview(sl, g, paths, dirs)
	editorThumb(sl, g, paths, dirs)

	return nil
}

func editorPatch(sl *slog.Logger, g *echo.Group, db *sql.DB) {
	g.PATCH("/16colors", func(c *echo.Context) error {
		return BeginTx(c, db, func(tx *sql.Tx) error {
			return htmx.Tx16Colors(c, tx)
		})
	})
	g.PATCH("/classifications", func(c *echo.Context) error {
		return BeginTx(c, db, func(tx *sql.Tx) error {
			return htmx.TxTags(sl, c, tx)
		})
	})
	g.PATCH("/comment", func(c *echo.Context) error {
		return BeginTx(c, db, func(tx *sql.Tx) error {
			return htmx.TxComment(c, tx)
		})
	})
	g.PATCH("/comment/reset", func(c *echo.Context) error {
		return BeginTx(c, db, func(tx *sql.Tx) error {
			return htmx.TxCommentUndo(c, tx)
		})
	})
	g.PATCH("/demozoo", func(c *echo.Context) error {
		return BeginTx(c, db, func(tx *sql.Tx) error {
			return htmx.TxDemozoo(c, tx)
		})
	})
	g.PATCH("/filename", func(c *echo.Context) error {
		return BeginTx(c, db, func(tx *sql.Tx) error {
			return htmx.TxFilename(c, tx)
		})
	})
	g.PATCH("/filename/reset", func(c *echo.Context) error {
		return BeginTx(c, db, func(tx *sql.Tx) error {
			return htmx.TxFilenameUndo(c, tx)
		})
	})
	g.PATCH("/github", func(c *echo.Context) error {
		return BeginTx(c, db, func(tx *sql.Tx) error {
			return htmx.TxGitHub(c, tx)
		})
	})
	g.PATCH("/links", func(c *echo.Context) error {
		return htmx.HTMLLinkTo(c, nil)
	})
	g.PATCH("/links/reset", func(c *echo.Context) error {
		return BeginTx(c, db, func(tx *sql.Tx) error {
			return htmx.TxLinksUndo(c, tx)
		})
	})
	g.PATCH("/platform", func(c *echo.Context) error {
		return app.PlatformEdit(sl, c, db)
	})
	g.PATCH("/platform+tag", app.PlatformTagInfo)
	g.PATCH("/pouet", func(c *echo.Context) error {
		return BeginTx(c, db, func(tx *sql.Tx) error {
			return htmx.TxPouet(c, tx)
		})
	})
	g.PATCH("/relations", func(c *echo.Context) error {
		return BeginTx(c, db, func(tx *sql.Tx) error {
			return htmx.TxRelations(c, tx)
		})
	})
	g.PATCH("/releasers", func(c *echo.Context) error {
		return BeginTx(c, db, func(tx *sql.Tx) error {
			return htmx.TxReleasers(c, tx)
		})
	})
	g.PATCH("/releasers/reset", func(c *echo.Context) error {
		return BeginTx(c, db, func(tx *sql.Tx) error {
			return htmx.TxReleasersUndo(c, tx)
		})
	})
	g.PATCH("/sites", func(c *echo.Context) error {
		return BeginTx(c, db, func(tx *sql.Tx) error {
			return htmx.TxWebsites(c, tx)
		})
	})
	g.PATCH("/tag", func(c *echo.Context) error {
		return app.TagEdit(sl, c, db)
	})
	g.PATCH("/tag/info", app.TagInfo)
	g.PATCH("/title", func(c *echo.Context) error {
		return BeginTx(c, db, func(tx *sql.Tx) error {
			return htmx.TxTitle(c, tx)
		})
	})
	g.PATCH("/title/reset", func(c *echo.Context) error {
		return BeginTx(c, db, func(tx *sql.Tx) error {
			return htmx.TxTitleUndo(c, tx)
		})
	})
	g.PATCH("/virustotal", func(c *echo.Context) error {
		return BeginTx(c, db, func(tx *sql.Tx) error {
			return htmx.TxVirusTotal(c, tx)
		})
	})
	g.PATCH("/ymd", func(c *echo.Context) error {
		return app.YMDEdit(c, db)
	})
	g.PATCH("/youtube", func(c *echo.Context) error {
		return BeginTx(c, db, func(tx *sql.Tx) error {
			return htmx.TxYouTube(c, tx)
		})
	})
}

func editorEmu(g *echo.Group, db *sql.DB) {
	emu := g.Group("/emulate")
	emu.PATCH("/broken/:id", func(c *echo.Context) error {
		return BeginTx(c, db, func(tx *sql.Tx) error {
			return htmx.TxEmulateBroken(c, tx)
		})
	})
	emu.PATCH("/runprogram/:id", func(c *echo.Context) error {
		return BeginTx(c, db, func(tx *sql.Tx) error {
			return htmx.TxEmulateRunProg(c, tx)
		})
	})
	emu.PATCH("/machine/:id", func(c *echo.Context) error {
		return BeginTx(c, db, func(tx *sql.Tx) error {
			return htmx.TxEmulateMachine(c, tx)
		})
	})
	emu.PATCH("/cpu/:id", func(c *echo.Context) error {
		return BeginTx(c, db, func(tx *sql.Tx) error {
			return htmx.TxEmulateCPU(c, tx)
		})
	})
	emu.PATCH("/sfx/:id", func(c *echo.Context) error {
		return BeginTx(c, db, func(tx *sql.Tx) error {
			return htmx.TxEmulateSFX(c, tx)
		})
	})
	emu.PATCH("/umb/:id", func(c *echo.Context) error {
		return BeginTx(c, db, func(tx *sql.Tx) error {
			return htmx.TxEmulateUMB(c, tx)
		})
	})
	emu.PATCH("/ems/:id", func(c *echo.Context) error {
		return BeginTx(c, db, func(tx *sql.Tx) error {
			return htmx.TxEmulateEMS(c, tx)
		})
	})
	emu.PATCH("/xms/:id", func(c *echo.Context) error {
		return BeginTx(c, db, func(tx *sql.Tx) error {
			return htmx.TxEmulateXMS(c, tx)
		})
	})
}

func editorReadme(sl *slog.Logger, g *echo.Group, db *sql.DB, paths command.Dirs, dirs app.Dirs) {
	diz := g.Group("/diz")
	// /editor/diz/copy
	diz.PATCH("/copy/:unid/:path", func(c *echo.Context) error {
		return htmx.FSCopyDIZ(c, paths)
	})
	diz.DELETE("/:unid", func(c *echo.Context) error {
		return htmx.FSRemoveDIZ(c, dirs.Extra)
	})
	// /editor/helper/copy
	helper := g.Group("/helper")
	helper.PATCH("/copy/:unid/:path", func(c *echo.Context) error {
		return htmx.FSCopyHelp(c, paths)
	})
	helper.DELETE("/:unid", func(c *echo.Context) error {
		return htmx.FSRemoveHelp(c, dirs.Extra)
	})

	readme := g.Group("/readme")
	readme.PATCH("/disable/:id", func(c *echo.Context) error {
		return BeginTx(c, db, func(tx *sql.Tx) error {
			return htmx.TxReadmeOff(c, tx)
		})
	})
	// /editor/readme/copy
	readme.PATCH("/copy/:unid/:path", func(c *echo.Context) error {
		return htmx.FSCopyReadme(sl, c, paths)
	})
	// /editor/readme/preview
	readme.PATCH("/preview/:unid/:path", func(c *echo.Context) error {
		return htmx.FSUseReadme(sl, c, false, paths)
	})
	// /editor/readme/preview-amiga
	readme.PATCH("/preview-amiga/:unid/:path", func(c *echo.Context) error {
		return htmx.FSUseReadme(sl, c, true, paths)
	})
	// /editor/readme/preview-binary
	readme.PATCH("/preview-binary/:unid/:path", func(c *echo.Context) error {
		return htmx.FSUseBinText(sl, c, paths)
	})
	readme.DELETE("/:unid", func(c *echo.Context) error {
		return htmx.FSRemoveReadme(c, dirs.Extra)
	})
}

func editorPreview(sl *slog.Logger, g *echo.Group, paths command.Dirs, dirs app.Dirs) {
	pre := g.Group("/preview")
	// /editor/preview/copy
	pre.PATCH("/copy/:unid/:path", func(c *echo.Context) error {
		return htmx.FSUseImage(sl, c, paths)
	})
	pre.PATCH("/crop11/:unid", func(c *echo.Context) error {
		return htmx.FSCrop(sl, c, command.SquareTop, paths)
	})
	pre.PATCH("/crop43/:unid", func(c *echo.Context) error {
		return htmx.FSCrop(sl, c, command.FourThree, paths)
	})
	pre.PATCH("/crop12/:unid", func(c *echo.Context) error {
		return htmx.FSCrop(sl, c, command.OneTwo, paths)
	})
	pre.PATCH("/remove/:unid", func(c *echo.Context) error {
		return htmx.FSRemoveImages(c, dirs.Preview)
	})
}

func editorThumb(sl *slog.Logger, g *echo.Group, paths command.Dirs, dirs app.Dirs) {
	thumb := g.Group("/thumbnail")
	thumb.PATCH("/copy/:unid/:path", func(c *echo.Context) error {
		return htmx.FSUseImage(sl, c, paths)
	})
	thumb.PATCH("/top/:unid", func(c *echo.Context) error {
		return htmx.FSAlign(sl, c, command.Top, paths)
	})
	thumb.PATCH("/middle/:unid", func(c *echo.Context) error {
		return htmx.FSAlign(sl, c, command.Middle, paths)
	})
	thumb.PATCH("/bottom/:unid", func(c *echo.Context) error {
		return htmx.FSAlign(sl, c, command.Bottom, paths)
	})
	thumb.PATCH("/left/:unid", func(c *echo.Context) error {
		return htmx.FSAlign(sl, c, command.Left, paths)
	})
	thumb.PATCH("/right/:unid", func(c *echo.Context) error {
		return htmx.FSAlign(sl, c, command.Right, paths)
	})
	thumb.PATCH("/pixel/:unid", func(c *echo.Context) error { //nolint:contextcheck
		return htmx.FSThumb(sl, c, command.Pixel, paths)
	})
	thumb.PATCH("/photo/:unid", func(c *echo.Context) error { //nolint:contextcheck
		return htmx.FSThumb(sl, c, command.Photo, paths)
	})
	thumb.PATCH("/remove/:unid", func(c *echo.Context) error {
		return htmx.FSRemoveImages(c, dirs.Thumbnail)
	})
}

func editorImgs(sl *slog.Logger, g *echo.Group, dirs app.Dirs) {
	imgs := g.Group("/images")
	imgs.PATCH("/pixelate/:unid", func(c *echo.Context) error { //nolint:contextcheck
		return htmx.FSPixelate(sl, c, dirs.Preview, dirs.Thumbnail)
	})
	imgs.PATCH("/remove/:unid", func(c *echo.Context) error {
		return htmx.FSRemoveImages(c, dirs.Preview, dirs.Thumbnail)
	})
}

func get(sl *slog.Logger, g *echo.Group, db *sql.DB, dirs app.Dirs) error {
	if err := nils.Check(sl, g, db); err != nil {
		return fmt.Errorf("get router: %w", err)
	}

	g.GET("/deletions",
		func(c *echo.Context) error {
			return app.Deletions(sl, c, db, "1")
		})
	g.GET("/get/demozoo/download/:unid/:id",
		func(c *echo.Context) error {
			return BeginTx(c, db, func(tx *sql.Tx) error {
				return app.GetDemozooParam(sl, c, tx, dirs.Download)
			})
		})
	g.GET("/for-approval",
		func(c *echo.Context) error {
			return app.ForApproval(sl, c, db, "1")
		})
	g.GET("/unwanted",
		func(c *echo.Context) error {
			return app.Unwanted(sl, c, db, "1")
		})

	return nil
}

func online(g *echo.Group, db *sql.DB) error {
	if err := nils.Check(g, db); err != nil {
		return fmt.Errorf("online router: %w", err)
	}

	online := g.Group("/online")
	online.PATCH("/true", func(c *echo.Context) error {
		return BeginTx(c, db, func(tx *sql.Tx) error {
			return htmx.TxPublic(c, tx, true)
		})
	})
	online.PATCH("/false", func(c *echo.Context) error {
		return BeginTx(c, db, func(tx *sql.Tx) error {
			return htmx.TxPublic(c, tx, false)
		})
	})
	online.GET("/true/:id", func(c *echo.Context) error {
		return BeginTx(c, db, func(tx *sql.Tx) error {
			return htmx.TxPublicByKey(c, tx, c.Param("id"), true)
		})
	})

	return nil
}

func search(sl *slog.Logger, g *echo.Group, db *sql.DB) error {
	if err := nils.Check(g, db); err != nil {
		return fmt.Errorf("search router: %w", err)
	}

	search := g.Group("/search")
	search.GET("/id", func(c *echo.Context) error {
		return app.SearchID(sl, c)
	})
	search.POST("/id", func(c *echo.Context) error {
		return htmx.SearchByID(sl, c, db)
	})

	return nil
}
