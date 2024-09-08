package handler

// Package file router.go contains the custom router URIs for the website.

import (
	"database/sql"
	"embed"
	"fmt"
	"net/http"
	"strings"

	"github.com/Defacto2/helper"
	"github.com/Defacto2/releaser"
	"github.com/Defacto2/server/handler/app"
	"github.com/Defacto2/server/handler/htmx"
	"github.com/Defacto2/server/internal/config"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

const code = http.StatusMovedPermanently

// FilesRoutes defines the file locations and routes for the web server.
func (c Configuration) FilesRoutes(e *echo.Echo, db *sql.DB, logger *zap.SugaredLogger,
	public embed.FS,
) (*echo.Echo, error) {
	if e == nil {
		panic(fmt.Errorf("%w for files routes router", ErrRoutes))
	}
	if logger == nil {
		return nil, fmt.Errorf("%w: %s", ErrZap, "handler files routes")
	}
	if d, err := public.ReadDir("."); err != nil || len(d) == 0 {
		return nil, fmt.Errorf("%w: %s", ErrFS, "handler files routes")
	}
	app.Caching.Records(c.RecordCount)
	dir := app.Dirs{
		Download:  c.Environment.AbsDownload,
		Preview:   c.Environment.AbsPreview,
		Thumbnail: c.Environment.AbsThumbnail,
		Extra:     c.Environment.AbsExtra,
	}
	nonce, err := c.nonce(e)
	if err != nil {
		return nil, fmt.Errorf("files routes nonce session key: %w", err)
	}
	e = c.signin(e, nonce)
	e = c.custom404(e)
	e = c.debugInfo(e)
	e = c.static(e)
	e = c.html(e, public)
	e = c.font(e, public)
	e = c.embed(e, public)
	e = c.search(e, db, logger)
	e = c.website(e, db, logger, dir)
	e = c.lock(e, db, logger, dir)
	return e, nil
}

// nonce configures and returns the session key for the cookie store.
// If the read mode is enabled then an empty session key is returned.
func (c Configuration) nonce(e *echo.Echo) (string, error) {
	if e == nil {
		panic(fmt.Errorf("%w for router nonce", ErrRoutes))
	}
	if c.Environment.ReadOnly {
		return "", nil
	}
	b, err := helper.CookieStore(c.Environment.SessionKey)
	if err != nil {
		return "", fmt.Errorf("none cookie store: %w", err)
	}
	e.Use(session.Middleware(sessions.NewCookieStore(b)))
	return string(b), nil
}

// html serves the embedded CSS, JS, WASM, and source map files for the HTML website layout.
func (c Configuration) html(e *echo.Echo, public embed.FS) *echo.Echo {
	if e == nil {
		panic(fmt.Errorf("%w for html router", ErrRoutes))
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
		panic(fmt.Errorf("%w for font router", ErrRoutes))
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
		panic(fmt.Errorf("%w for embed router", ErrRoutes))
	}
	e.FileFS("/favicon.ico", "public/image/favicon.ico", public)
	e.FileFS("/osd.xml", "public/text/osd.xml", public)
	e.FileFS("/robots.txt", "public/text/robots.txt", public)
	e.FileFS("/site.webmanifest", "public/text/site.webmanifest.json", public)
	return e
}

// static serves the static assets for the website such as the thumbnail and preview images.
func (c Configuration) static(e *echo.Echo) *echo.Echo {
	if e == nil {
		panic(fmt.Errorf("%w for static router", ErrRoutes))
	}
	e.Static(config.StaticThumb(), c.Environment.AbsThumbnail)
	e.Static(config.StaticOriginal(), c.Environment.AbsPreview)
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
		panic(fmt.Errorf("%w for debug info router", ErrRoutes))
	}
	if c.Environment.ProdMode {
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
func (c Configuration) website(e *echo.Echo, db *sql.DB, logger *zap.SugaredLogger, dir app.Dirs) *echo.Echo {
	if e == nil {
		panic(fmt.Errorf("%w for website router", ErrRoutes))
	}
	e.GET("/health-check", func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})
	s := e.Group("")
	s.GET("/", app.Index)
	s.GET("/artist", func(c echo.Context) error {
		return app.Artist(c, db)
	})
	s.GET("/bbs", func(c echo.Context) error {
		return app.BBS(c, db)
	})
	s.GET("/bbs/a-z", func(c echo.Context) error {
		return app.BBSAZ(c, db)
	})
	s.GET("/bbs/year", func(c echo.Context) error {
		return app.BBSYear(c, db)
	})
	s.GET("/coder", func(c echo.Context) error {
		return app.Coder(c, db)
	})
	s.GET(Downloader, func(cx echo.Context) error {
		return app.Download(cx, db, logger, c.Environment.AbsDownload)
	})
	s.GET("/f/:id", func(cx echo.Context) error {
		dir.URI = cx.Param("id")
		return dir.Artifact(cx, db, logger, c.Environment.ReadOnly)
	})
	s.GET("/file/stats", func(cx echo.Context) error {
		return app.Categories(cx, db, logger, true)
	})
	s.GET("/files/:id/:page", func(cx echo.Context) error {
		switch cx.Param("id") {
		case "for-approval", "deletions", "unwanted":
			return app.StatusErr(cx, http.StatusNotFound, cx.Param("uri"))
		}
		return app.Artifacts(cx, db, cx.Param("id"), cx.Param("page"))
	})
	s.GET("/files/:id", func(cx echo.Context) error {
		switch cx.Param("id") {
		case "for-approval", "deletions", "unwanted":
			return app.StatusErr(cx, http.StatusNotFound, cx.Param("uri"))
		}
		return app.Artifacts(cx, db, cx.Param("id"), "1")
	})
	s.GET("/file", func(cx echo.Context) error {
		return app.Categories(cx, db, logger, false)
	})
	s.GET("/ftp", func(c echo.Context) error {
		return app.FTP(c, db)
	})
	s.GET("/g/:id", func(cx echo.Context) error {
		return app.Releasers(cx, db, logger, cx.Param("id"))
	})
	s.GET("/history", app.History)
	s.GET("/interview", app.Interview)
	s.GET("/jsdos/:id", func(cx echo.Context) error {
		return app.DownloadJsDos(cx, db,
			c.Environment.AbsExtra, c.Environment.AbsDownload)
	})
	s.GET("/magazine", func(c echo.Context) error {
		return app.Magazine(c, db)
	})
	s.GET("/magazine/a-z", func(c echo.Context) error {
		return app.MagazineAZ(c, db)
	})
	s.GET("/new", app.New)
	s.GET("/musician", func(c echo.Context) error {
		return app.Musician(c, db)
	})
	s.GET("/p/:id", func(cx echo.Context) error {
		return app.Sceners(cx, db, cx.Param("id"))
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
	s.GET("/releaser", func(c echo.Context) error {
		return app.Releaser(c, db)
	})
	s.GET("/releaser/a-z", func(c echo.Context) error {
		return app.ReleaserAZ(c, db)
	})
	s.GET("/releaser/year", func(c echo.Context) error {
		return app.ReleaserYear(c, db)
	})
	s.GET("/scener", func(c echo.Context) error {
		return app.Scener(c, db)
	})
	s.GET("/sum/:id", func(cx echo.Context) error {
		return app.Checksum(cx, db, cx.Param("id"))
	})
	s.GET("/thanks", app.Thanks)
	s.GET("/thescene", app.TheScene)
	s.GET("/website/:id", func(cx echo.Context) error {
		return app.Website(cx, cx.Param("id"))
	})
	s.GET("/website", func(cx echo.Context) error {
		return app.Website(cx, "")
	})
	s.GET("/writer", func(c echo.Context) error {
		return app.Writer(c, db)
	})
	s.GET("/v/:id", func(cx echo.Context) error {
		return app.Inline(cx, db, logger, c.Environment.AbsDownload)
	})
	return e
}

// search forms and the results for database queries.
func (c Configuration) search(e *echo.Echo, db *sql.DB, logger *zap.SugaredLogger) *echo.Echo {
	if e == nil {
		panic(fmt.Errorf("%w for search router", ErrRoutes))
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
		return app.PostDesc(cx, db, terms)
	})
	search.POST("/desc", func(cx echo.Context) error {
		return app.PostDesc(cx, db, cx.FormValue("search-term-query"))
	})
	search.POST("/file", func(c echo.Context) error {
		return app.PostFilename(c, db)
	})
	search.POST("/releaser", func(cx echo.Context) error {
		return htmx.SearchReleaser(cx, db, logger)
	})
	return e
}

// signin for operators.
func (c Configuration) signin(e *echo.Echo, nonce string) *echo.Echo {
	if e == nil {
		panic(fmt.Errorf("%w for signin router", ErrRoutes))
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

// MovedPermanently redirects are partial URL routers that are to be redirected with a HTTP 301 Moved Permanently.
func MovedPermanently(e *echo.Echo) *echo.Echo {
	if e == nil {
		panic(fmt.Errorf("%w for move permanently router", ErrRoutes))
	}
	e = nginx(e)
	e = fixes(e)
	return e
}

// nginx redirects.
func nginx(e *echo.Echo) *echo.Echo {
	if e == nil {
		panic(fmt.Errorf("%w for nginx router", ErrRoutes))
	}
	nginx := e.Group("")
	nginx.GET("/file/detail/:id", func(c echo.Context) error {
		return c.Redirect(code, "/f/"+c.Param("id"))
	})
	nginx.GET("/file/download/:id", func(c echo.Context) error {
		return c.Redirect(code, "/d/"+c.Param("id"))
	})
	nginx.GET("/file/view/:id", func(c echo.Context) error {
		return c.Redirect(code, "/v/"+c.Param("id"))
	})
	nginx.GET("/cracktros-detail.cfm/:id", func(c echo.Context) error {
		return c.Redirect(code, "/f/"+c.Param("id"))
	})
	nginx.GET("/wayback/:url", func(c echo.Context) error {
		return c.Redirect(code, "https://wayback.defacto2.net/"+c.Param("url"))
	})
	nginx.GET("/link/list", func(c echo.Context) error {
		return c.Redirect(code, "https://wayback.defacto2.net/")
	})
	return e
}

// fixes redirects repaired, releaser database entry redirects that are contained in the model fix package.
func fixes(e *echo.Echo) *echo.Echo {
	if e == nil {
		panic(fmt.Errorf("%w for fixes router", ErrRoutes))
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
