package handler

// Package file router.go contains the custom router URIs for the website.

import (
	"embed"
	"fmt"
	"net/http"
	"strings"

	"github.com/Defacto2/releaser"
	"github.com/Defacto2/server/handler/app"
	"github.com/Defacto2/server/handler/htmx"
	"github.com/Defacto2/server/internal/config"
	"github.com/Defacto2/server/internal/helper"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

const code = http.StatusMovedPermanently

// FilesRoutes defines the file locations and routes for the web server.
func (c Configuration) FilesRoutes(e *echo.Echo, logger *zap.SugaredLogger, public embed.FS) (*echo.Echo, error) {
	if e == nil {
		panic(ErrRoutes)
	}
	if logger == nil {
		return nil, fmt.Errorf("%w: %s", ErrZap, "handler routes")
	}
	if d, err := public.ReadDir("."); err != nil || len(d) == 0 {
		return nil, fmt.Errorf("%w: %s", ErrFS, "public")
	}

	app.Caching.Records(c.RecordCount)
	dir := app.Dirs{
		Download:  c.Environment.DownloadDir,
		Preview:   c.Environment.PreviewDir,
		Thumbnail: c.Environment.ThumbnailDir,
	}

	nonce, err := c.nonce(e)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", err, "nonce")
	}
	e = c.signin(e, nonce)
	e = c.custom404(e)
	e = c.debugInfo(e)
	e = c.static(e)
	e = c.uploader(e)
	e = c.html(e, public)
	e = c.font(e, public)
	e = c.embed(e, public)
	e = c.search(e, logger)
	e = c.website(e, logger, dir)
	e = c.editor(e, logger, dir)
	return e, nil
}

// nonce configures and returns the session key for the cookie store.
// If the read mode is enabled then an empty session key is returned.
func (c Configuration) nonce(e *echo.Echo) (string, error) {
	if e == nil {
		panic(ErrRoutes)
	}
	if c.Environment.ReadMode {
		return "", nil
	}
	b, err := helper.CookieStore(c.Environment.SessionKey)
	if err != nil {
		return "", err
	}
	e.Use(session.Middleware(sessions.NewCookieStore(b)))
	return string(b), nil
}

// html serves the embedded CSS, JS, WASM, and source map files for the HTML website layout.
func (c Configuration) html(e *echo.Echo, public embed.FS) *echo.Echo {
	if e == nil {
		panic(ErrRoutes)
	}
	hrefs, names := app.Hrefs(), app.Names()
	for key, href := range hrefs {
		e.FileFS(href, names[key], public)
	}
	// source map files
	const mapExt = ".map"
	e.FileFS(hrefs[app.Bootstrap5]+mapExt, names[app.Bootstrap5]+mapExt, public)
	e.FileFS(hrefs[app.Bootstrap5JS]+mapExt, names[app.Bootstrap5JS]+mapExt, public)
	e.FileFS(hrefs[app.Jsdos6JS]+mapExt, names[app.Jsdos6JS]+mapExt, public)
	return e
}

// font serves the embedded woff2, woff, and ttf font files for the website layout.
func (c Configuration) font(e *echo.Echo, public embed.FS) *echo.Echo {
	if e == nil {
		panic(ErrRoutes)
	}
	paths, names := app.FontRefs(), app.FontNames()
	font := e.Group("/font")
	for key, href := range paths {
		font.FileFS(href, names[key], public)
	}
	return e
}

// embed serves the miscellaneous embedded files for the website layout.
// This includes the favicon, robots.txt, site.webmanifest, osd.xml, and the SVG icons.
func (c Configuration) embed(e *echo.Echo, public embed.FS) *echo.Echo {
	if e == nil {
		panic(ErrRoutes)
	}
	e.FileFS("/bootstrap-icons.svg", "public/image/bootstrap-icons.svg", public)
	e.FileFS("/favicon.ico", "public/image/favicon.ico", public)
	e.FileFS("/osd.xml", "public/text/osd.xml", public)
	e.FileFS("/robots.txt", "public/text/robots.txt", public)
	e.FileFS("/site.webmanifest", "public/text/site.webmanifest.json", public)
	return e
}

// static serves the static assets for the website such as the thumbnail and preview images.
func (c Configuration) static(e *echo.Echo) *echo.Echo {
	if e == nil {
		panic(ErrRoutes)
	}
	e.Static(config.StaticThumb(), c.Environment.ThumbnailDir)
	e.Static(config.StaticOriginal(), c.Environment.PreviewDir)
	return e
}

// custom404 is a custom 404 error handler for the website,
// "The page cannot be found".
func (c Configuration) custom404(e *echo.Echo) *echo.Echo {
	if e == nil {
		panic(ErrRoutes)
	}
	e.GET("/:uri", func(cx echo.Context) error {
		return app.StatusErr(cx, http.StatusNotFound, cx.Param("uri"))
	})
	return e
}

// debugInfo returns detailed information about the HTTP request.
func (c Configuration) debugInfo(e *echo.Echo) *echo.Echo {
	if e == nil {
		panic(ErrRoutes)
	}
	if c.Environment.ProductionMode {
		return e
	}

	type debug struct {
		Protocol       string `json:"protocol"`
		Host           string `json:"host"`
		RemoteAddress  string `json:"remoteAddress"`
		Method         string `json:"method"`
		Path           string `json:"path"`
		URI            string `json:"uri"`
		Query          string `json:"query"`
		Referer        string `json:"referer"`
		UserAgent      string `json:"userAgent"`
		Accept         string `json:"accept"`
		AcceptEncoding string `json:"acceptEncoding"`
		AcceptLanguage string `json:"acceptLanguage"`
	}
	e.GET("/debug", func(cx echo.Context) error {
		req := cx.Request()
		d := debug{
			Protocol:       req.Proto,
			Host:           req.Host,
			RemoteAddress:  req.RemoteAddr,
			Method:         req.Method,
			Path:           req.URL.Path,
			URI:            req.RequestURI,
			Query:          req.URL.RawQuery,
			Referer:        req.Referer(),
			UserAgent:      req.UserAgent(),
			Accept:         req.Header.Get("Accept"),
			AcceptEncoding: req.Header.Get("Accept-Encoding"),
			AcceptLanguage: req.Header.Get("Accept-Language"),
		}
		return cx.JSONPretty(http.StatusOK, d, "  ")
	})
	return e
}

// website routes for the main site.
func (c Configuration) website(e *echo.Echo, logger *zap.SugaredLogger, dir app.Dirs) *echo.Echo {
	if e == nil {
		panic(ErrRoutes)
	}
	s := e.Group("")
	s.GET("/", app.Index)
	s.GET("/artist", app.Artist)
	s.GET("/bbs", app.BBS)
	s.GET("/bbs/a-z", app.BBSAZ)
	s.GET("/bbs/year", app.BBSYear)
	s.GET("/coder", app.Coder)
	s.GET(Downloader, func(cx echo.Context) error {
		return app.Download(cx, logger, c.Environment.DownloadDir)
	})
	s.GET("/f/:id", func(cx echo.Context) error {
		dir.URI = cx.Param("id")
		return dir.Artifact(cx, logger, c.Environment.ReadMode)
	})
	s.GET("/file/stats", func(cx echo.Context) error {
		return app.File(cx, logger, true)
	})
	s.GET("/files/:id/:page", func(cx echo.Context) error {
		switch cx.Param("id") {
		case "for-approval", "deletions", "unwanted":
			return app.StatusErr(cx, http.StatusNotFound, cx.Param("uri"))
		}
		return app.Files(cx, cx.Param("id"), cx.Param("page"))
	})
	s.GET("/files/:id", func(cx echo.Context) error {
		switch cx.Param("id") {
		case "for-approval", "deletions", "unwanted":
			return app.StatusErr(cx, http.StatusNotFound, cx.Param("uri"))
		}
		return app.Files(cx, cx.Param("id"), "1")
	})
	s.GET("/file", func(cx echo.Context) error {
		return app.File(cx, logger, false)
	})
	s.GET("/ftp", app.FTP)
	s.GET("/g/:id", func(cx echo.Context) error {
		return app.Releasers(cx, cx.Param("id"))
	})
	s.GET("/history", app.History)
	s.GET("/interview", app.Interview)
	s.GET("/magazine", app.Magazine)
	s.GET("/magazine/a-z", app.MagazineAZ)
	s.GET("/musician", app.Musician)
	s.GET("/p/:id", func(cx echo.Context) error {
		return app.Sceners(cx, cx.Param("id"))
	})
	s.GET("/pouet/vote/:id", func(cx echo.Context) error {
		return app.VotePouet(cx, logger, cx.Param("id"))
	})
	s.GET("/pouet/prod/:id", func(cx echo.Context) error {
		return app.ProdPouet(cx, cx.Param("id"))
	})
	s.GET("/zoo/prod/:id", func(cx echo.Context) error {
		return app.ProdZoo(cx, cx.Param("id"))
	})
	s.GET("/r/:id", app.Reader)
	s.GET("/releaser", app.Releaser)
	s.GET("/releaser/a-z", app.ReleaserAZ)
	s.GET("/releaser/year", app.ReleaserYear)
	s.GET("/scener", app.Scener)
	s.GET("/sum/:id", func(cx echo.Context) error {
		return app.Checksum(cx, cx.Param("id"))
	})
	s.GET("/thanks", app.Thanks)
	s.GET("/thescene", app.TheScene)
	s.GET("/website/:id", func(cx echo.Context) error {
		return app.Website(cx, cx.Param("id"))
	})
	s.GET("/website", func(cx echo.Context) error {
		return app.Website(cx, "")
	})
	s.GET("/writer", app.Writer)
	s.GET("/v/:id", func(cx echo.Context) error {
		return app.Inline(cx, logger, c.Environment.DownloadDir)
	})
	return e
}

// search forms and the results for database queries.
func (c Configuration) search(e *echo.Echo, logger *zap.SugaredLogger) *echo.Echo {
	if e == nil {
		panic(ErrRoutes)
	}
	search := e.Group("/search")
	search.GET("/desc", app.SearchDesc)
	search.GET("/file", app.SearchFile)
	search.GET("/releaser", app.SearchReleaser)
	search.GET("/result", func(cx echo.Context) error {
		// this legacy get result should be kept for (osx.xml) opensearch compatibility
		// and to keep possible backwards compatibility with third party site links.
		terms := strings.ReplaceAll(cx.QueryParam("query"), "+", " ") // AND replacement
		terms = strings.ReplaceAll(terms, "|", ",")                   // OR replacement
		return app.PostDesc(cx, terms)
	})
	search.POST("/desc", func(cx echo.Context) error {
		return app.PostDesc(cx, cx.FormValue("search-term-query"))
	})
	search.POST("/file", app.PostFilename)
	search.POST("/releaser", func(cx echo.Context) error {
		return htmx.SearchReleaser(cx, logger)
	})
	return e
}

// uploader for anonymous client uploads.
func (c Configuration) uploader(e *echo.Echo) *echo.Echo {
	if e == nil {
		panic(ErrRoutes)
	}
	uploader := e.Group("/uploader")
	uploader.Use(c.ReadOnlyLock)
	uploader.GET("", app.PostIntro)
	return e
}

// signin for operators.
func (c Configuration) signin(e *echo.Echo, nonce string) *echo.Echo {
	if e == nil {
		panic(ErrRoutes)
	}
	signings := e.Group("")
	signings.Use(c.ReadOnlyLock)
	signings.GET("/signedout", app.SignedOut)
	signings.GET("/signin", func(cx echo.Context) error {
		return app.Signin(cx, c.Environment.GoogleClientID, nonce)
	})
	signings.GET("/operator/signin", func(cx echo.Context) error {
		return cx.Redirect(http.StatusMovedPermanently, "/signin")
	})
	google := signings.Group("/google")
	google.POST("/callback", func(cx echo.Context) error {
		return app.GoogleCallback(cx,
			c.Environment.GoogleClientID,
			c.Environment.SessionMaxAge,
			c.Environment.GoogleAccounts...)
	})
	return e
}

// editor pages to update the database records.
func (c Configuration) editor(e *echo.Echo, logger *zap.SugaredLogger, dir app.Dirs) *echo.Echo {
	if e == nil {
		panic(ErrRoutes)
	}
	editor := e.Group("/editor")
	editor.Use(c.ReadOnlyLock, c.SessionLock)
	editor.GET("/get/demozoo/download/:id",
		func(cx echo.Context) error {
			return app.GetDemozooLink(cx, dir.Download)
		})
	editor.GET("/for-approval",
		func(cx echo.Context) error {
			return app.FilesWaiting(cx, "1")
		})
	editor.GET("/deletions",
		func(cx echo.Context) error {
			return app.FilesDeletions(cx, "1")
		})
	editor.GET("/unwanted",
		func(cx echo.Context) error {
			return app.FilesUnwanted(cx, "1")
		})
	online := editor.Group("/online")
	online.POST("/true", func(cx echo.Context) error {
		return htmx.RecordToggle(cx, true)
	})
	online.POST("/false", func(cx echo.Context) error {
		return htmx.RecordToggle(cx, false)
	})
	readme := editor.Group("/readme")
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
	images := editor.Group("/images")
	images.POST("/copy", func(c echo.Context) error {
		return dir.PreviewPost(c, logger)
	})
	images.POST("/delete", dir.PreviewDel)
	ansilove := editor.Group("/ansilove")
	ansilove.POST("/copy", func(c echo.Context) error {
		return dir.AnsiLovePost(c, logger)
	})
	editor.POST("/releasers", app.ReleaserEdit)
	editor.POST("/title", app.TitleEdit)
	editor.POST("/ymd", app.YMDEdit)
	editor.POST("/platform", app.PlatformEdit)
	editor.POST("/platform+tag", app.PlatformTagInfo)
	tag := editor.Group("/tag")
	tag.POST("", app.TagEdit)
	tag.POST("/info", app.TagInfo)
	return e
}

// MovedPermanently redirects are partial URL routers that are to be redirected with a HTTP 301 Moved Permanently.
func MovedPermanently(e *echo.Echo) *echo.Echo {
	if e == nil {
		panic(ErrRoutes)
	}
	e = nginx(e)
	e = retired(e)
	e = wayback(e)
	e = fixes(e)
	return e
}

// nginx redirects.
func nginx(e *echo.Echo) *echo.Echo {
	if e == nil {
		panic(ErrRoutes)
	}
	nginx := e.Group("")
	nginx.GET("/welcome", func(c echo.Context) error {
		return c.Redirect(code, "/")
	})
	nginx.GET("/file/download/:id", func(c echo.Context) error {
		return c.Redirect(code, "/d/"+c.Param("id"))
	})
	nginx.GET("/file/view/:id", func(c echo.Context) error {
		return c.Redirect(code, "/v/"+c.Param("id"))
	})
	nginx.GET("/apollo-x/fc.htm", func(c echo.Context) error {
		return c.Redirect(code, "/wayback/apollo-x-demo-resources-1999-december-17/fc.htm")
	})
	nginx.GET("/bbs.cfm", func(c echo.Context) error {
		return c.Redirect(code, "/bbs")
	})
	nginx.GET("/contact.cfm", func(c echo.Context) error {
		return c.Redirect(code, "/") // there's no dedicated contact page
	})
	nginx.GET("/cracktros.cfm", func(c echo.Context) error {
		return c.Redirect(code, "/files/intro")
	})
	nginx.GET("/cracktros-detail.cfm:/:id", func(c echo.Context) error {
		return c.Redirect(code, "/f/"+c.Param("id"))
	})
	nginx.GET("/documents.cfm", func(c echo.Context) error {
		return c.Redirect(code, "/files/text")
	})
	nginx.GET("/index.cfm", func(c echo.Context) error {
		return c.Redirect(code, "/")
	})
	nginx.GET("/index.cfm/:uri", func(c echo.Context) error {
		return c.Redirect(code, "/")
	})
	nginx.GET("/index.cfml/:uri", func(c echo.Context) error {
		return c.Redirect(code, "/")
	})
	nginx.GET("/groups.cfm", func(c echo.Context) error {
		return c.Redirect(code, "/releaser")
	})
	nginx.GET("/magazines.cfm", func(c echo.Context) error {
		return c.Redirect(code, "/magazine")
	})
	nginx.GET("/nfo-files.cfm", func(c echo.Context) error {
		return c.Redirect(code, "/files/nfo")
	})
	nginx.GET("/portal.cfm", func(c echo.Context) error {
		return c.Redirect(code, "/website")
	})
	nginx.GET("/rewrite.cfm", func(c echo.Context) error {
		return c.Redirect(code, "/")
	})
	nginx.GET("/site-info.cfm", func(c echo.Context) error {
		return c.Redirect(code, "/") // there's no dedicated about site page
	})
	return e
}

// retired, redirects from the 2020 edition of the website.
func retired(e *echo.Echo) *echo.Echo {
	if e == nil {
		panic(ErrRoutes)
	}
	retired := e.Group("")
	retired.GET("/code", func(c echo.Context) error {
		return c.Redirect(code, "https://github.com/Defacto2/server")
	})
	retired.GET("/commercial", func(c echo.Context) error {
		return c.Redirect(code, "/")
	})
	retired.GET("/defacto", func(c echo.Context) error {
		return c.Redirect(code, "/history")
	})
	retired.GET("/defacto2/donate", func(c echo.Context) error {
		return c.Redirect(code, "/thanks")
	})
	retired.GET("/defacto2/history", func(c echo.Context) error {
		return c.Redirect(code, "/history")
	})
	retired.GET("/defacto2/subculture", func(c echo.Context) error {
		return c.Redirect(code, "/thescene")
	})
	retired.GET("/file/detail/:id", func(c echo.Context) error {
		return c.Redirect(code, "/f/"+c.Param("id"))
	})
	retired.GET("/file/list/waitingapproval", func(c echo.Context) error {
		return c.Redirect(code, "/files/for-approval")
	})
	retired.GET("/file/index", func(c echo.Context) error {
		return c.Redirect(code, "/file")
	})
	retired.GET("/file/list/:uri", func(c echo.Context) error {
		return c.Redirect(code, "/files/new-uploads")
	})
	retired.GET("/files/json/site.webmanifest", func(c echo.Context) error {
		return c.Redirect(code, "/site.webmanifest")
	})
	retired.GET("/help/cc", func(c echo.Context) error {
		return c.Redirect(code, "/") // there's no dedicated contact page
	})
	retired.GET("/help/privacy", func(c echo.Context) error {
		return c.Redirect(code, "/") // there's no dedicated privacy page
	})
	retired.GET("/help/viruses", func(c echo.Context) error {
		return c.Redirect(code, "/") // there's no dedicated virus page
	})
	retired.GET("/home", func(c echo.Context) error {
		return c.Redirect(code, "/")
	})
	retired.GET("/link/list", func(c echo.Context) error {
		return c.Redirect(code, "/website")
	})
	retired.GET("/link/list/:id", func(c echo.Context) error {
		return c.Redirect(code, "/website")
	})
	//nolint:misspell
	retired.GET("/organisation/list/bbs", func(c echo.Context) error {
		return c.Redirect(code, "/bbs")
	})
	//nolint:misspell
	retired.GET("/organisation/list/group", func(c echo.Context) error {
		return c.Redirect(code, "/releaser")
	})
	//nolint:misspell
	retired.GET("/organisation/list/ftp", func(c echo.Context) error {
		return c.Redirect(code, "/ftp")
	})
	//nolint:misspell
	retired.GET("/organisation/list/magazine", func(c echo.Context) error {
		return c.Redirect(code, "/magazine")
	})
	retired.GET("/person/list", func(c echo.Context) error {
		return c.Redirect(code, "/scener")
	})
	retired.GET("/person/list/artists", func(c echo.Context) error {
		return c.Redirect(code, "/artist")
	})
	retired.GET("/person/list/coders", func(c echo.Context) error {
		return c.Redirect(code, "/coder")
	})
	retired.GET("/person/list/musicians", func(c echo.Context) error {
		return c.Redirect(code, "/musician")
	})
	retired.GET("/person/list/writers", func(c echo.Context) error {
		return c.Redirect(code, "/writer")
	})
	retired.GET("/upload", func(c echo.Context) error {
		return c.Redirect(code, "/")
	})
	retired.GET("/upload/file", func(c echo.Context) error {
		return c.Redirect(code, "/")
	})
	retired.GET("/upload/external", func(c echo.Context) error {
		return c.Redirect(code, "/")
	})
	retired.GET("/upload/intro", func(c echo.Context) error {
		return c.Redirect(code, "/")
	})
	retired.GET("/upload/site", func(c echo.Context) error {
		return c.Redirect(code, "/")
	})
	retired.GET("/upload/document", func(c echo.Context) error {
		return c.Redirect(code, "/")
	})
	retired.GET("/upload/magazine", func(c echo.Context) error {
		return c.Redirect(code, "/")
	})
	retired.GET("/upload/art", func(c echo.Context) error {
		return c.Redirect(code, "/")
	})
	retired.GET("/upload/other", func(c echo.Context) error {
		return c.Redirect(code, "/")
	})
	return e
}

// wayback redirects.
func wayback(e *echo.Echo) *echo.Echo {
	if e == nil {
		panic(ErrRoutes)
	}
	wayback := e.Group("")
	wayback.GET("/scene-archive/:uri", func(c echo.Context) error {
		return c.Redirect(code, "/")
	})
	wayback.GET("/includes/documentsweb/df2web99/scene-archive/history.html", func(c echo.Context) error {
		return c.Redirect(code, "/wayback/defacto2-from-1999-september-26/scene-archive/history.html")
	})
	wayback.GET("/includes/documentsweb/tKC_history.html", func(c echo.Context) error {
		return c.Redirect(code, "/wayback/the-life-and-legend-of-tkc-2000-october-10/index.html")
	})
	wayback.GET("/legacy/apollo-x/:uri", func(c echo.Context) error {
		return c.Redirect(code, "/wayback/apollo-x-demo-resources-1999-december-17/:uri")
	})
	wayback.GET("/web/20120827022026/http:/www.defacto2.net:80/file/list/nfotool", func(c echo.Context) error {
		return c.Redirect(code, "/files/nfo-tool")
	})
	wayback.GET("/web.pages/warez_world-1.htm", func(c echo.Context) error {
		return c.Redirect(code, "/wayback/warez-world-from-2001-july-26/index.html")
	})
	return e
}

// fixes redirects repaired, releaser database entry redirects that are contained in the model fix package.
func fixes(e *echo.Echo) *echo.Echo {
	if e == nil {
		panic(ErrRoutes)
	}
	fixes := e.Group("/g")
	const g = "/g/"
	fixes.GET("/acid", func(c echo.Context) error {
		return c.Redirect(code, g+releaser.Obfuscate("ACID PRODUCTIONS"))
	})
	fixes.GET("/ansi-creators-in-demand", func(c echo.Context) error {
		return c.Redirect(code, g+releaser.Obfuscate("ACID PRODUCTIONS"))
	})
	fixes.GET("/ice", func(c echo.Context) error {
		return c.Redirect(code, g+releaser.Obfuscate("INSANE CREATORS ENTERPRISE"))
	})
	fixes.GET("/"+releaser.Obfuscate("pirates with attitude"), func(c echo.Context) error {
		return c.Redirect(code, g+releaser.Obfuscate("pirates with attitudes"))
	})
	fixes.GET("/"+releaser.Obfuscate("TRISTAR AND RED SECTOR INC"), func(c echo.Context) error {
		return c.Redirect(code, g+releaser.Obfuscate("TRISTAR & RED SECTOR INC"))
	})
	fixes.GET("/x-pression", func(c echo.Context) error {
		return c.Redirect(code, g+releaser.Obfuscate("X-PRESSION DESIGN"))
	})
	fixes.GET("/"+releaser.Obfuscate("DAMN EXCELLENT ANSI DESIGNERS"), func(c echo.Context) error {
		return c.Redirect(code, g+releaser.Obfuscate("DAMN EXCELLENT ANSI DESIGN"))
	})
	fixes.GET("/"+releaser.Obfuscate("THE ORIGINAL FUNNY GUYS"), func(c echo.Context) error {
		return c.Redirect(code, g+releaser.Obfuscate("ORIGINALLY FUNNY GUYS"))
	})
	fixes.GET("/"+releaser.Obfuscate("ORIGINAL FUNNY GUYS"), func(c echo.Context) error {
		return c.Redirect(code, g+releaser.Obfuscate("ORIGINALLY FUNNY GUYS"))
	})
	fixes.GET("/"+releaser.Obfuscate("DARKSIDE INC"), func(c echo.Context) error {
		return c.Redirect(code, g+releaser.Obfuscate("DARKSIDE INCORPORATED"))
	})
	fixes.GET("/"+releaser.Obfuscate("RSS"), func(c echo.Context) error {
		return c.Redirect(code, g+releaser.Obfuscate("renaissance"))
	})
	return e
}
