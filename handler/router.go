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
)

const code = http.StatusMovedPermanently

// Routes defines the routes for the web server.
func (c Configuration) Routes(e *echo.Echo, public embed.FS) (*echo.Echo, error) {
	if c.Logger == nil {
		return nil, fmt.Errorf("%w: %s", ErrZap, "handler routes")
	}
	if e == nil {
		return nil, fmt.Errorf("%w: %s", ErrRoutes, "handler routes")
	}
	if d, err := public.ReadDir("."); err != nil || len(d) == 0 {
		return nil, fmt.Errorf("%w: %s", ErrFS, "public")
	}

	app.Caching.Records(c.RecordCount)
	dir := app.Dirs{
		Download:  c.Import.DownloadDir,
		Preview:   c.Import.PreviewDir,
		Thumbnail: c.Import.ThumbnailDir,
	}

	nonce, err := c.nonce(e)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", err, "nonce")
	}
	if e, err = c.html(e, public); err != nil {
		return nil, fmt.Errorf("%w: %s", err, "html")
	}
	if e, err = c.fonts(e, public); err != nil {
		return nil, fmt.Errorf("%w: %s", err, "fonts")
	}
	if e, err = c.embedded(e, public); err != nil {
		return nil, fmt.Errorf("%w: %s", err, "embedded")
	}
	if e, err = c.static(e); err != nil {
		return nil, fmt.Errorf("%w: %s", err, "static")
	}
	if e, err = c.custom404(e); err != nil {
		return nil, fmt.Errorf("%w: %s", err, "custom404")
	}
	if e, err = c.debugInfo(e); err != nil {
		return nil, fmt.Errorf("%w: %s", err, "debugInfo")
	}
	if e, err = c.website(e, dir); err != nil {
		return nil, fmt.Errorf("%w: %s", err, "website")
	}
	if e, err = c.search(e); err != nil {
		return nil, fmt.Errorf("%w: %s", err, "search")
	}
	if e, err = c.uploader(e); err != nil {
		return nil, fmt.Errorf("%w: %s", err, "uploader")
	}
	if e, err = c.signings(e, nonce); err != nil {
		return nil, fmt.Errorf("%w: %s", err, "signings")
	}
	if e, err = c.editor(e, dir); err != nil {
		return nil, fmt.Errorf("%w: %s", err, "editor")
	}
	return e, nil
}

// nonce configures and returns the session key for the cookie store.
// If the read mode is enabled then an empty session key is returned.
func (c Configuration) nonce(e *echo.Echo) (string, error) {
	if e == nil {
		return "", ErrRoutes
	}
	if c.Import.ReadMode {
		return "", nil
	}
	b, err := helper.CookieStore(c.Import.SessionKey)
	if err != nil {
		return "", err
	}
	e.Use(session.Middleware(sessions.NewCookieStore(b)))
	return string(b), nil
}

// html serves the embedded CSS, JS, WASM, and source map files for the HTML website layout.
func (c Configuration) html(e *echo.Echo, public embed.FS) (*echo.Echo, error) {
	if e == nil {
		return nil, ErrRoutes
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
	return e, nil
}

// fonts serves the embedded woff2, woff, and ttf font files for the website layout.
func (c Configuration) fonts(e *echo.Echo, public embed.FS) (*echo.Echo, error) {
	if e == nil {
		return nil, ErrRoutes
	}
	paths, names := app.FontRefs(), app.FontNames()
	font := e.Group("/font")
	for key, href := range paths {
		font.FileFS(href, names[key], public)
	}
	return e, nil
}

// embedded serves the miscellaneous embedded files for the website layout.
// This includes the favicon, robots.txt, site.webmanifest, osd.xml, and the SVG icons.
func (c Configuration) embedded(e *echo.Echo, public embed.FS) (*echo.Echo, error) {
	if e == nil {
		return nil, ErrRoutes
	}
	e.FileFS("/bootstrap-icons.svg", "public/image/bootstrap-icons.svg", public)
	e.FileFS("/favicon.ico", "public/image/favicon.ico", public)
	e.FileFS("/osd.xml", "public/text/osd.xml", public)
	e.FileFS("/robots.txt", "public/text/robots.txt", public)
	e.FileFS("/site.webmanifest", "public/text/site.webmanifest.json", public)
	return e, nil
}

// static serves the static assets for the website such as the thumbnail and preview images.
func (c Configuration) static(e *echo.Echo) (*echo.Echo, error) {
	if e == nil {
		return nil, ErrRoutes
	}
	e.Static(config.StaticThumb(), c.Import.ThumbnailDir)
	e.Static(config.StaticOriginal(), c.Import.PreviewDir)
	return e, nil
}

// custom404 is a custom 404 error handler for the website, "The page cannot be found."
func (c Configuration) custom404(e *echo.Echo) (*echo.Echo, error) {
	logr := c.Logger
	if logr == nil {
		return nil, ErrZap
	}
	if e == nil {
		return nil, ErrRoutes
	}
	e.GET("/:uri", func(x echo.Context) error {
		return app.StatusErr(c.Logger, x, http.StatusNotFound, x.Param("uri"))
	})
	return e, nil
}

// debugInfo returns detailed information about the HTTP request.
func (c Configuration) debugInfo(e *echo.Echo) (*echo.Echo, error) {
	if e == nil {
		return nil, ErrRoutes
	}
	if c.Import.ProductionMode {
		return e, nil
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

	e.GET("/debug", func(x echo.Context) error {
		req := x.Request()
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
		return x.JSONPretty(http.StatusOK, d, "  ")
	})
	return e, nil
}

// website routes for the main site.
func (c Configuration) website(e *echo.Echo, dir app.Dirs) (*echo.Echo, error) {
	if e == nil {
		return nil, ErrRoutes
	}
	logr := c.Logger

	s := e.Group("")
	s.GET("/", func(x echo.Context) error {
		return app.Index(logr, x)
	})
	s.GET("/artist", func(x echo.Context) error {
		return app.Artist(logr, x)
	})
	s.GET("/bbs", func(x echo.Context) error {
		return app.BBS(logr, x)
	})
	s.GET("/bbs/a-z", func(x echo.Context) error {
		return app.BBSAZ(logr, x)
	})
	s.GET("/bbs/year", func(x echo.Context) error {
		return app.BBSYear(logr, x)
	})
	s.GET("/coder", func(x echo.Context) error {
		return app.Coder(logr, x)
	})
	s.GET(Downloader, func(x echo.Context) error {
		return app.Download(logr, x, c.Import.DownloadDir)
	})
	s.GET("/f/:id", func(x echo.Context) error {
		dir.URI = x.Param("id")
		return dir.Artifact(logr, x, c.Import.ReadMode)
	})
	s.GET("/file/stats", func(x echo.Context) error {
		return app.File(logr, x, true)
	})
	s.GET("/files/:id/:page", func(x echo.Context) error {
		switch x.Param("id") {
		case "for-approval", "deletions", "unwanted":
			return app.StatusErr(logr, x, http.StatusNotFound, x.Param("uri"))
		}
		return app.Files(logr, x, x.Param("id"), x.Param("page"))
	})
	s.GET("/files/:id", func(x echo.Context) error {
		switch x.Param("id") {
		case "for-approval", "deletions", "unwanted":
			return app.StatusErr(logr, x, http.StatusNotFound, x.Param("uri"))
		}
		return app.Files(logr, x, x.Param("id"), "1")
	})
	s.GET("/file", func(x echo.Context) error {
		return app.File(logr, x, false)
	})
	s.GET("/ftp", func(x echo.Context) error {
		return app.FTP(logr, x)
	})
	s.GET("/g/:id", func(x echo.Context) error {
		return app.Releasers(logr, x, x.Param("id"))
	})
	s.GET("/history", func(x echo.Context) error {
		return app.History(logr, x)
	})
	s.GET("/interview", func(x echo.Context) error {
		return app.Interview(logr, x)
	})
	s.GET("/magazine", func(x echo.Context) error {
		return app.Magazine(logr, x)
	})
	s.GET("/magazine/a-z", func(x echo.Context) error {
		return app.MagazineAZ(logr, x)
	})
	s.GET("/musician", func(x echo.Context) error {
		return app.Musician(logr, x)
	})
	s.GET("/p/:id", func(x echo.Context) error {
		return app.Sceners(logr, x, x.Param("id"))
	})
	s.GET("/pouet/vote/:id", func(x echo.Context) error {
		return app.VotePouet(logr, x, x.Param("id"))
	})
	s.GET("/pouet/prod/:id", func(x echo.Context) error {
		return app.ProdPouet(logr, x, x.Param("id"))
	})
	s.GET("/zoo/prod/:id", func(x echo.Context) error {
		return app.ProdZoo(logr, x, x.Param("id"))
	})
	s.GET("/r/:id", func(x echo.Context) error {
		return app.Reader(logr, x)
	})
	s.GET("/releaser", func(x echo.Context) error {
		return app.Releaser(logr, x)
	})
	s.GET("/releaser/a-z", func(x echo.Context) error {
		return app.ReleaserAZ(logr, x)
	})
	s.GET("/releaser/year", func(x echo.Context) error {
		return app.ReleaserYear(logr, x)
	})
	s.GET("/scener", func(x echo.Context) error {
		return app.Scener(logr, x)
	})
	s.GET("/sum/:id", func(x echo.Context) error {
		return app.Checksum(logr, x, x.Param("id"))
	})
	s.GET("/thanks", func(x echo.Context) error {
		return app.Thanks(logr, x)
	})
	s.GET("/thescene", func(x echo.Context) error {
		return app.TheScene(logr, x)
	})
	s.GET("/website/:id", func(x echo.Context) error {
		return app.Website(logr, x, x.Param("id"))
	})
	s.GET("/website", func(x echo.Context) error {
		return app.Website(logr, x, "")
	})
	s.GET("/writer", func(x echo.Context) error {
		return app.Writer(logr, x)
	})
	s.GET("/v/:id", func(x echo.Context) error {
		return app.Inline(logr, x, c.Import.DownloadDir)
	})
	return e, nil
}

// search forms and the results for database queries.
func (c Configuration) search(e *echo.Echo) (*echo.Echo, error) {
	if e == nil {
		return nil, ErrRoutes
	}
	logr := c.Logger
	search := e.Group("/search")
	search.GET("/desc", func(x echo.Context) error {
		return app.SearchDesc(logr, x)
	})
	search.GET("/file", func(x echo.Context) error {
		return app.SearchFile(logr, x)
	})
	search.GET("/releaser", func(x echo.Context) error {
		return app.SearchReleaser(logr, x)
	})
	search.GET("/result", func(x echo.Context) error {
		// this legacy get result should be kept for (osx.xml) opensearch compatibility
		// and to keep possible backwards compatibility with third party site links.
		terms := strings.ReplaceAll(x.QueryParam("query"), "+", " ") // AND replacement
		terms = strings.ReplaceAll(terms, "|", ",")                  // OR replacement
		return app.PostDesc(logr, x, terms)
	})
	search.POST("/desc", func(x echo.Context) error {
		return app.PostDesc(logr, x, x.FormValue("search-term-query"))
	})
	search.POST("/file", func(x echo.Context) error {
		return app.PostFilename(logr, x)
	})
	search.POST("/releaser", func(x echo.Context) error {
		return htmx.SearchReleaser(logr, x)
	})
	return e, nil
}

// uploader for anonymous client uploads
func (c Configuration) uploader(e *echo.Echo) (*echo.Echo, error) {
	if e == nil {
		return nil, ErrRoutes
	}
	logr := c.Logger
	uploader := e.Group("/uploader")
	uploader.Use(c.ReadOnlyLock)
	uploader.GET("", func(x echo.Context) error {
		return app.PostIntro(logr, x)
	})
	return e, nil
}

// signins for operators.
func (c Configuration) signings(e *echo.Echo, nonce string) (*echo.Echo, error) {
	if e == nil {
		return nil, ErrRoutes
	}
	logr := c.Logger
	signings := e.Group("")
	signings.Use(c.ReadOnlyLock)
	signings.GET("/signedout", func(x echo.Context) error {
		return app.SignedOut(logr, x)
	})
	signings.GET("/signin", func(x echo.Context) error {
		return app.Signin(logr, x, c.Import.GoogleClientID, nonce)
	})
	signings.GET("/operator/signin", func(x echo.Context) error {
		return x.Redirect(http.StatusMovedPermanently, "/signin")
	})
	google := signings.Group("/google")
	google.POST("/callback", func(x echo.Context) error {
		return app.GoogleCallback(logr, x,
			c.Import.GoogleClientID,
			c.Import.SessionMaxAge,
			c.Import.GoogleAccounts...)
	})
	return e, nil
}

// editor pages to update the database records.
func (c Configuration) editor(e *echo.Echo, dir app.Dirs) (*echo.Echo, error) {
	if e == nil {
		return nil, ErrRoutes
	}
	logr := c.Logger
	editor := e.Group("/editor")
	editor.Use(c.ReadOnlyLock, c.SessionLock)
	editor.GET("/get/demozoo/download/:id",
		func(x echo.Context) error {
			return app.GetDemozooLink(logr, x, dir.Download)
		})
	editor.GET("/for-approval",
		func(x echo.Context) error {
			return app.FilesWaiting(logr, x, "1")
		})
	editor.GET("/deletions",
		func(x echo.Context) error {
			return app.FilesDeletions(logr, x, "1")
		})
	editor.GET("/unwanted",
		func(x echo.Context) error {
			return app.FilesUnwanted(logr, x, "1")
		})
	online := editor.Group("/online")
	online.POST("/true", func(x echo.Context) error {
		return app.RecordToggle(logr, x, true)
	})
	online.POST("/false", func(x echo.Context) error {
		return app.RecordToggle(logr, x, false)
	})
	readme := editor.Group("/readme")
	readme.POST("/copy", func(x echo.Context) error {
		return app.ReadmePost(logr, x, dir.Download)
	})
	readme.POST("/delete", func(x echo.Context) error {
		return app.ReadmeDel(logr, x, dir.Download)
	})
	readme.POST("/hide", func(x echo.Context) error {
		dir.URI = x.Param("id")
		return app.ReadmeToggle(logr, x)
	})
	images := editor.Group("/images")
	images.POST("/copy", func(x echo.Context) error {
		return dir.PreviewPost(logr, x)
	})
	images.POST("/delete", func(x echo.Context) error {
		return dir.PreviewDel(logr, x)
	})
	ansilove := editor.Group("/ansilove")
	ansilove.POST("/copy", func(x echo.Context) error {
		return dir.AnsiLovePost(logr, x)
	})
	editor.POST("/releasers", func(x echo.Context) error {
		return app.ReleaserEdit(logr, x)
	})
	editor.POST("/title", func(x echo.Context) error {
		return app.TitleEdit(logr, x)
	})
	editor.POST("/ymd", func(x echo.Context) error {
		return app.YMDEdit(logr, x)
	})
	editor.POST("/platform", func(x echo.Context) error {
		return app.PlatformEdit(logr, x)
	})
	editor.POST("/platform+tag", func(x echo.Context) error {
		return app.PlatformTagInfo(logr, x)
	})
	tag := editor.Group("/tag")
	tag.POST("", func(x echo.Context) error {
		return app.TagEdit(logr, x)
	})
	tag.POST("/info", func(x echo.Context) error {
		return app.TagInfo(logr, x)
	})
	return e, nil
}

// Moved redirects are partial URL routers that are to be redirected with a HTTP 301 Moved Permanently.
func (c Configuration) Moved(e *echo.Echo) (*echo.Echo, error) {
	if e == nil {
		return nil, ErrRoutes
	}
	e, err := nginx(e)
	if err != nil {
		return nil, err
	}
	e, err = retired(e)
	if err != nil {
		return nil, err
	}
	e, err = wayback(e)
	if err != nil {
		return nil, err
	}
	e, err = fixes(e)
	if err != nil {
		return nil, err
	}
	return e, nil
}

// nginx redirects
func nginx(e *echo.Echo) (*echo.Echo, error) {
	if e == nil {
		return nil, ErrRoutes
	}
	nginx := e.Group("")
	nginx.GET("/welcome", func(x echo.Context) error {
		return x.Redirect(code, "/")
	})
	nginx.GET("/file/download/:id", func(x echo.Context) error {
		return x.Redirect(code, "/d/"+x.Param("id"))
	})
	nginx.GET("/file/view/:id", func(x echo.Context) error {
		return x.Redirect(code, "/v/"+x.Param("id"))
	})
	nginx.GET("/apollo-x/fc.htm", func(x echo.Context) error {
		return x.Redirect(code, "/wayback/apollo-x-demo-resources-1999-december-17/fc.htm")
	})
	nginx.GET("/bbs.cfm", func(x echo.Context) error {
		return x.Redirect(code, "/bbs")
	})
	nginx.GET("/contact.cfm", func(x echo.Context) error {
		return x.Redirect(code, "/") // there's no dedicated contact page
	})
	nginx.GET("/cracktros.cfm", func(x echo.Context) error {
		return x.Redirect(code, "/files/intro")
	})
	nginx.GET("/cracktros-detail.cfm:/:id", func(x echo.Context) error {
		return x.Redirect(code, "/f/"+x.Param("id"))
	})
	nginx.GET("/documents.cfm", func(x echo.Context) error {
		return x.Redirect(code, "/files/text")
	})
	nginx.GET("/index.cfm", func(x echo.Context) error {
		return x.Redirect(code, "/")
	})
	nginx.GET("/index.cfm/:uri", func(x echo.Context) error {
		return x.Redirect(code, "/")
	})
	nginx.GET("/index.cfml/:uri", func(x echo.Context) error {
		return x.Redirect(code, "/")
	})
	nginx.GET("/groups.cfm", func(x echo.Context) error {
		return x.Redirect(code, "/releaser")
	})
	nginx.GET("/magazines.cfm", func(x echo.Context) error {
		return x.Redirect(code, "/magazine")
	})
	nginx.GET("/nfo-files.cfm", func(x echo.Context) error {
		return x.Redirect(code, "/files/nfo")
	})
	nginx.GET("/portal.cfm", func(x echo.Context) error {
		return x.Redirect(code, "/website")
	})
	nginx.GET("/rewrite.cfm", func(x echo.Context) error {
		return x.Redirect(code, "/")
	})
	nginx.GET("/site-info.cfm", func(x echo.Context) error {
		return x.Redirect(code, "/") // there's no dedicated about site page
	})
	return e, nil
}

// retired, redirects from the 2020 edition of the website
func retired(e *echo.Echo) (*echo.Echo, error) {
	if e == nil {
		return nil, ErrRoutes
	}
	retired := e.Group("")
	retired.GET("/code", func(x echo.Context) error {
		return x.Redirect(code, "https://github.com/Defacto2/server")
	})
	retired.GET("/commercial", func(x echo.Context) error {
		return x.Redirect(code, "/")
	})
	retired.GET("/defacto", func(x echo.Context) error {
		return x.Redirect(code, "/history")
	})
	retired.GET("/defacto2/donate", func(x echo.Context) error {
		return x.Redirect(code, "/thanks")
	})
	retired.GET("/defacto2/history", func(x echo.Context) error {
		return x.Redirect(code, "/history")
	})
	retired.GET("/defacto2/subculture", func(x echo.Context) error {
		return x.Redirect(code, "/thescene")
	})
	retired.GET("/file/detail/:id", func(x echo.Context) error {
		return x.Redirect(code, "/f/"+x.Param("id"))
	})
	retired.GET("/file/list/waitingapproval", func(x echo.Context) error {
		return x.Redirect(code, "/files/for-approval")
	})
	retired.GET("/file/index", func(x echo.Context) error {
		return x.Redirect(code, "/file")
	})
	retired.GET("/file/list/:uri", func(x echo.Context) error {
		return x.Redirect(code, "/files/new-uploads")
	})
	retired.GET("/files/json/site.webmanifest", func(x echo.Context) error {
		return x.Redirect(code, "/site.webmanifest")
	})
	retired.GET("/help/cc", func(x echo.Context) error {
		return x.Redirect(code, "/") // there's no dedicated contact page
	})
	retired.GET("/help/privacy", func(x echo.Context) error {
		return x.Redirect(code, "/") // there's no dedicated privacy page
	})
	retired.GET("/help/viruses", func(x echo.Context) error {
		return x.Redirect(code, "/") // there's no dedicated virus page
	})
	retired.GET("/home", func(x echo.Context) error {
		return x.Redirect(code, "/")
	})
	retired.GET("/link/list", func(x echo.Context) error {
		return x.Redirect(code, "/website")
	})
	retired.GET("/link/list/:id", func(x echo.Context) error {
		return x.Redirect(code, "/website")
	})
	//nolint:misspell
	retired.GET("/organisation/list/bbs", func(x echo.Context) error {
		return x.Redirect(code, "/bbs")
	})
	//nolint:misspell
	retired.GET("/organisation/list/group", func(x echo.Context) error {
		return x.Redirect(code, "/releaser")
	})
	//nolint:misspell
	retired.GET("/organisation/list/ftp", func(x echo.Context) error {
		return x.Redirect(code, "/ftp")
	})
	//nolint:misspell
	retired.GET("/organisation/list/magazine", func(x echo.Context) error {
		return x.Redirect(code, "/magazine")
	})
	retired.GET("/person/list", func(x echo.Context) error {
		return x.Redirect(code, "/scener")
	})
	retired.GET("/person/list/artists", func(x echo.Context) error {
		return x.Redirect(code, "/artist")
	})
	retired.GET("/person/list/coders", func(x echo.Context) error {
		return x.Redirect(code, "/coder")
	})
	retired.GET("/person/list/musicians", func(x echo.Context) error {
		return x.Redirect(code, "/musician")
	})
	retired.GET("/person/list/writers", func(x echo.Context) error {
		return x.Redirect(code, "/writer")
	})
	retired.GET("/upload", func(x echo.Context) error {
		return x.Redirect(code, "/")
	})
	retired.GET("/upload/file", func(x echo.Context) error {
		return x.Redirect(code, "/")
	})
	retired.GET("/upload/external", func(x echo.Context) error {
		return x.Redirect(code, "/")
	})
	retired.GET("/upload/intro", func(x echo.Context) error {
		return x.Redirect(code, "/")
	})
	retired.GET("/upload/site", func(x echo.Context) error {
		return x.Redirect(code, "/")
	})
	retired.GET("/upload/document", func(x echo.Context) error {
		return x.Redirect(code, "/")
	})
	retired.GET("/upload/magazine", func(x echo.Context) error {
		return x.Redirect(code, "/")
	})
	retired.GET("/upload/art", func(x echo.Context) error {
		return x.Redirect(code, "/")
	})
	retired.GET("/upload/other", func(x echo.Context) error {
		return x.Redirect(code, "/")
	})
	return e, nil
}

// wayback redirects
func wayback(e *echo.Echo) (*echo.Echo, error) {
	if e == nil {
		return nil, ErrRoutes
	}
	wayback := e.Group("")
	wayback.GET("/scene-archive/:uri", func(x echo.Context) error {
		return x.Redirect(code, "/")
	})
	wayback.GET("/includes/documentsweb/df2web99/scene-archive/history.html", func(x echo.Context) error {
		return x.Redirect(code, "/wayback/defacto2-from-1999-september-26/scene-archive/history.html")
	})
	wayback.GET("/includes/documentsweb/tKC_history.html", func(x echo.Context) error {
		return x.Redirect(code, "/wayback/the-life-and-legend-of-tkc-2000-october-10/index.html")
	})
	wayback.GET("/legacy/apollo-x/:uri", func(x echo.Context) error {
		return x.Redirect(code, "/wayback/apollo-x-demo-resources-1999-december-17/:uri")
	})
	wayback.GET("/web/20120827022026/http:/www.defacto2.net:80/file/list/nfotool", func(x echo.Context) error {
		return x.Redirect(code, "/files/nfo-tool")
	})
	wayback.GET("/web.pages/warez_world-1.htm", func(x echo.Context) error {
		return x.Redirect(code, "/wayback/warez-world-from-2001-july-26/index.html")
	})
	return e, nil
}

// fixes redirects repaired, releaser database entry redirects that are contained in the model fix package.
func fixes(e *echo.Echo) (*echo.Echo, error) {
	if e == nil {
		return nil, ErrRoutes
	}
	fixes := e.Group("/g")
	const g = "/g/"
	fixes.GET("/acid", func(x echo.Context) error {
		return x.Redirect(code, g+releaser.Obfuscate("ACID PRODUCTIONS"))
	})
	fixes.GET("/ansi-creators-in-demand", func(x echo.Context) error {
		return x.Redirect(code, g+releaser.Obfuscate("ACID PRODUCTIONS"))
	})
	fixes.GET("/ice", func(x echo.Context) error {
		return x.Redirect(code, g+releaser.Obfuscate("INSANE CREATORS ENTERPRISE"))
	})
	fixes.GET("/"+releaser.Obfuscate("pirates with attitude"), func(x echo.Context) error {
		return x.Redirect(code, g+releaser.Obfuscate("pirates with attitudes"))
	})
	fixes.GET("/"+releaser.Obfuscate("TRISTAR AND RED SECTOR INC"), func(x echo.Context) error {
		return x.Redirect(code, g+releaser.Obfuscate("TRISTAR & RED SECTOR INC"))
	})
	fixes.GET("/x-pression", func(x echo.Context) error {
		return x.Redirect(code, g+releaser.Obfuscate("X-PRESSION DESIGN"))
	})
	fixes.GET("/"+releaser.Obfuscate("DAMN EXCELLENT ANSI DESIGNERS"), func(x echo.Context) error {
		return x.Redirect(code, g+releaser.Obfuscate("DAMN EXCELLENT ANSI DESIGN"))
	})
	fixes.GET("/"+releaser.Obfuscate("THE ORIGINAL FUNNY GUYS"), func(x echo.Context) error {
		return x.Redirect(code, g+releaser.Obfuscate("ORIGINALLY FUNNY GUYS"))
	})
	fixes.GET("/"+releaser.Obfuscate("ORIGINAL FUNNY GUYS"), func(x echo.Context) error {
		return x.Redirect(code, g+releaser.Obfuscate("ORIGINALLY FUNNY GUYS"))
	})
	fixes.GET("/"+releaser.Obfuscate("DARKSIDE INC"), func(x echo.Context) error {
		return x.Redirect(code, g+releaser.Obfuscate("DARKSIDE INCORPORATED"))
	})
	fixes.GET("/"+releaser.Obfuscate("RSS"), func(x echo.Context) error {
		return x.Redirect(code, g+releaser.Obfuscate("renaissance"))
	})
	return e, nil
}
