package handler

// Package file router.go contains the custom router URIs for the website.

import (
	"embed"
	"fmt"
	"net/http"
	"strings"

	"github.com/Defacto2/releaser"
	"github.com/Defacto2/server/handler/app"
	"github.com/Defacto2/server/internal/config"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
)

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

// Routes defines the routes for the web server.
func (c Configuration) Routes(e *echo.Echo, public embed.FS) (*echo.Echo, error) {
	if e == nil {
		return nil, fmt.Errorf("%w: %s", ErrRoutes, "handler routes")
	}
	logr := c.Logger

	const mapExt = ".map"

	// Cookie session key for the session store.
	if !c.Import.ReadMode {
		key, err := CookieStore(c.Import.SessionKey)
		if err != nil {
			return nil, err
		}
		e.Use(session.Middleware(sessions.NewCookieStore(key)))
	}

	// Cache the database record count.
	app.Caching.RecordCount = c.RecordCount

	// Set the application configuration for paths.
	dir := app.Dirs{
		Download:  c.Import.DownloadDir,
		Preview:   c.Import.PreviewDir,
		Thumbnail: c.Import.ThumbnailDir,
	}

	// Serve embedded CSS, JS and WASM files
	hrefs, names := app.Hrefs(), app.Names()
	for key, href := range hrefs {
		e.FileFS(href, names[key], public)
	}
	// Serve embedded CSS and JS map files
	e.FileFS(hrefs[app.Bootstrap]+mapExt, names[app.Bootstrap]+mapExt, public)
	e.FileFS(hrefs[app.BootstrapJS]+mapExt, names[app.BootstrapJS]+mapExt, public)
	e.FileFS(hrefs[app.JSDosUI]+mapExt, names[app.JSDosUI]+mapExt, public)

	// Serve embedded SVG collections
	e.FileFS("/bootstrap-icons.svg", "public/image/bootstrap-icons.svg", public)

	// Serve embedded font files
	fonts, fnames := app.FontRefs(), app.FontNames()
	font := e.Group("/font")
	for key, href := range fonts {
		font.FileFS(href, fnames[key], public)
	}

	// Serve embedded image files
	e.FileFS("/favicon.ico", "public/image/favicon.ico", public)

	// Serve embedded text files
	e.FileFS("/osd.xml", "public/text/osd.xml", public)
	e.FileFS("/robots.txt", "public/text/robots.txt", public)
	e.FileFS("/site.webmanifest", "public/text/site.webmanifest.json", public)

	// Serve asset images
	e.Static(config.StaticThumb(), c.Import.ThumbnailDir)
	e.Static(config.StaticOriginal(), c.Import.PreviewDir)

	// Custom 404 error, "The page cannot be found"
	e.GET("/:uri", func(x echo.Context) error {
		return app.StatusErr(logr, x, http.StatusNotFound, x.Param("uri"))
	})

	// Request debug information
	if !c.Import.ProductionMode {
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
	}

	// Use session middleware for all routes but not the embedded files.
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
	s.GET("/coder", func(x echo.Context) error {
		return app.Coder(logr, x)
	})
	s.GET(Downloader, func(x echo.Context) error {
		return app.Download(logr, x, c.Import.DownloadDir)
	})
	s.GET("/f/:id", func(x echo.Context) error {
		dir.URI = x.Param("id")
		return dir.About(logr, x, c.Import.ReadMode)
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

	// Search forms and results for database records.
	search := s.Group("/search")
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
		return app.PostReleaser(logr, x)
	})

	// Uploader for anonymous client uploads
	uploader := e.Group("/uploader")
	uploader.Use(c.ReadOnlyLock)
	uploader.GET("", func(x echo.Context) error {
		return app.PostIntro(logr, x)
	})

	// Sign in for operators.
	signings := e.Group("")
	signings.Use(c.ReadOnlyLock)
	signings.GET("/signedout", func(x echo.Context) error {
		return app.SignedOut(logr, x)
	})
	signings.GET("/signin", func(x echo.Context) error {
		return app.Signin(logr, x, c.Import.GoogleClientID)
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

	// Editor pages to update the database records.
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
		return nil, fmt.Errorf("%w: %s", ErrRoutes, "handler routes")
	}
	const code = http.StatusMovedPermanently
	// nginx redirects
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
		return x.Redirect(code, "/") // there's no dedicated about page
	})
	// 2020 website redirects
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
	// wayback redirects
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
	// repaired, releaser database entry redirects
	fixes := e.Group("/g")
	const g = "/g/"
	fixes.GET("/acid", func(x echo.Context) error {
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
	return e, nil
}
