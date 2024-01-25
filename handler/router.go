package handler

// Package file router.go contains the custom router URIs for the website.

import (
	"crypto/rand"
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
	"go.uber.org/zap"
)

// Routes defines the routes for the web server.
func (c Configuration) Routes(z *zap.SugaredLogger, e *echo.Echo, public embed.FS) (*echo.Echo, error) {
	if e == nil {
		return nil, fmt.Errorf("%w: %s", ErrRoutes, "handler routes")
	}
	if z == nil {
		return nil, fmt.Errorf("%w: %s", ErrZap, "handler routes")
	}

	// The session key, if empty a long randomized value is generated that changes on every restart.
	if !c.Import.ReadMode {
		key := []byte(c.Import.SessionKey)
		if c.Import.SessionKey == "" {
			const length = 32
			key = make([]byte, length)
			if _, err := rand.Read(key); err != nil {
				return nil, err
			}
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
	e.FileFS(hrefs[app.Bootstrap]+".map", names[app.Bootstrap]+".map", public)
	e.FileFS(hrefs[app.BootstrapJS]+".map", names[app.BootstrapJS]+".map", public)
	e.FileFS(hrefs[app.JSDosUI]+".map", names[app.JSDosUI]+".map", public)

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
		return app.StatusErr(z, x, http.StatusNotFound, x.Param("uri"))
	})

	// Use session middleware for all routes but not the embedded files.
	s := e.Group("")
	s.GET("/", func(x echo.Context) error {
		return app.Index(z, x)
	})
	s.GET("/artist", func(x echo.Context) error {
		return app.Artist(z, x)
	})
	s.GET("/bbs", func(x echo.Context) error {
		return app.BBS(z, x)
	})
	s.GET("/bbs/a-z", func(x echo.Context) error {
		return app.BBSAZ(z, x)
	})
	s.GET("/coder", func(x echo.Context) error {
		return app.Coder(z, x)
	})
	s.GET(Downloader, func(x echo.Context) error {
		return app.Download(z, x, c.Import.DownloadDir)
	})
	s.GET("/f/:id", func(x echo.Context) error {
		dir.URI = x.Param("id")
		return dir.About(z, x, c.Import.ReadMode)
	})
	s.GET("/file/stats", func(x echo.Context) error {
		return app.File(z, x, true)
	})
	s.GET("/files/:id/:page", func(x echo.Context) error {
		return app.Files(z, x, x.Param("id"), x.Param("page"))
	})
	s.GET("/files/:id", func(x echo.Context) error {
		return app.Files(z, x, x.Param("id"), "1")
	})
	s.GET("/file", func(x echo.Context) error {
		return app.File(z, x, false)
	})
	s.GET("/ftp", func(x echo.Context) error {
		return app.FTP(z, x)
	})
	s.GET("/g/:id", func(x echo.Context) error {
		return app.Releasers(z, x, x.Param("id"))
	})
	s.GET("/history", func(x echo.Context) error {
		return app.History(z, x)
	})
	s.GET("/interview", func(x echo.Context) error {
		return app.Interview(z, x)
	})
	s.GET("/magazine", func(x echo.Context) error {
		return app.Magazine(z, x)
	})
	s.GET("/magazine/a-z", func(x echo.Context) error {
		return app.MagazineAZ(z, x)
	})
	s.GET("/musician", func(x echo.Context) error {
		return app.Musician(z, x)
	})
	s.GET("/p/:id", func(x echo.Context) error {
		return app.Sceners(z, x, x.Param("id"))
	})
	s.GET("/pouet/vote/:id", func(x echo.Context) error {
		return app.VotePouet(z, x, x.Param("id"))
	})
	s.GET("/pouet/prod/:id", func(x echo.Context) error {
		return app.ProdPouet(z, x, x.Param("id"))
	})
	s.GET("/zoo/prod/:id", func(x echo.Context) error {
		return app.ProdZoo(z, x, x.Param("id"))
	})
	s.GET("/r/:id", func(x echo.Context) error {
		return app.Reader(z, x)
	})
	s.GET("/releaser", func(x echo.Context) error {
		return app.Releaser(z, x)
	})
	s.GET("/releaser/a-z", func(x echo.Context) error {
		return app.ReleaserAZ(z, x)
	})
	s.GET("/scener", func(x echo.Context) error {
		return app.Scener(z, x)
	})
	s.GET("/sum/:id", func(x echo.Context) error {
		return app.Checksum(z, x, x.Param("id"))
	})
	s.GET("/thanks", func(x echo.Context) error {
		return app.Thanks(z, x)
	})
	s.GET("/thescene", func(x echo.Context) error {
		return app.TheScene(z, x)
	})
	s.GET("/website/:id", func(x echo.Context) error {
		return app.Website(z, x, x.Param("id"))
	})
	s.GET("/website", func(x echo.Context) error {
		return app.Website(z, x, "")
	})
	s.GET("/writer", func(x echo.Context) error {
		return app.Writer(z, x)
	})
	s.GET("/v/:id", func(x echo.Context) error {
		return app.Inline(z, x, c.Import.DownloadDir)
	})

	// Search forms and results for database records.
	search := s.Group("/search")
	search.GET("/desc", func(x echo.Context) error {
		return app.SearchDesc(z, x)
	})
	search.GET("/file", func(x echo.Context) error {
		return app.SearchFile(z, x)
	})
	search.GET("/releaser", func(x echo.Context) error {
		return app.SearchReleaser(z, x)
	})
	search.GET("/result", func(x echo.Context) error {
		// this legacy get result should be kept for (osx.xml) opensearch compatibility
		// and to keep possible backwards compatibility with third party site links.
		terms := strings.ReplaceAll(x.QueryParam("query"), "+", " ") // AND replacement
		terms = strings.ReplaceAll(terms, "|", ",")                  // OR replacement
		return app.PostDesc(z, x, terms)
	})
	search.POST("/desc", func(x echo.Context) error {
		return app.PostDesc(z, x, x.FormValue("search-term-query"))
	})
	search.POST("/file", func(x echo.Context) error {
		return app.PostFilename(z, x)
	})
	search.POST("/releaser", func(x echo.Context) error {
		return app.PostReleaser(z, x)
	})

	// Uploader for anonymous client uploads
	uploader := e.Group("/uploader")
	uploader.Use(c.ReadOnlyLock)
	uploader.GET("", func(x echo.Context) error {
		return app.PostIntro(z, x)
	})

	// Sign in for operators.
	signings := e.Group("")
	signings.Use(c.ReadOnlyLock)
	signings.GET("/signedout", func(x echo.Context) error {
		return app.SignedOut(z, x)
	})
	signings.GET("/signin", func(x echo.Context) error {
		return app.Signin(z, x, c.Import.GoogleClientID)
	})
	signings.GET("/operator/signin", func(x echo.Context) error {
		return x.Redirect(http.StatusMovedPermanently, "/signin")
	})
	google := signings.Group("/google")
	google.POST("/callback", func(x echo.Context) error {
		return app.GoogleCallback(z, x,
			c.Import.GoogleClientID,
			c.Import.SessionMaxAge,
			c.Import.GoogleAccounts...)
	})

	// Editor pages to update the database records.
	editor := e.Group("/editor")
	editor.Use(c.ReadOnlyLock, c.SessionLock)
	online := editor.Group("/online")
	online.POST("/true", func(x echo.Context) error {
		return app.RecordToggle(z, x, true)
	})
	online.POST("/false", func(x echo.Context) error {
		return app.RecordToggle(z, x, false)
	})
	readme := editor.Group("/readme")
	readme.POST("/copy", func(x echo.Context) error {
		return app.ReadmePost(z, x, dir.Download)
	})
	readme.POST("/delete", func(x echo.Context) error {
		return app.ReadmeDel(z, x, dir.Download)
	})
	readme.POST("/hide", func(x echo.Context) error {
		dir.URI = x.Param("id")
		return app.ReadmeToggle(z, x)
	})
	images := editor.Group("/images")
	images.POST("/copy", func(x echo.Context) error {
		return dir.PreviewPost(z, x)
	})
	images.POST("/delete", func(x echo.Context) error {
		return dir.PreviewDel(z, x)
	})
	ansilove := editor.Group("/ansilove")
	ansilove.POST("/copy", func(x echo.Context) error {
		return dir.AnsiLovePost(z, x)
	})
	editor.POST("/title", func(x echo.Context) error {
		return app.TitleEdit(z, x)
	})
	editor.POST("/ymd", func(x echo.Context) error {
		return app.YMDEdit(z, x)
	})
	editor.POST("/platform", func(x echo.Context) error {
		return app.PlatformEdit(z, x)
	})
	editor.POST("/platform+tag", func(x echo.Context) error {
		return app.PlatformTagInfo(z, x)
	})
	tag := editor.Group("/tag")
	tag.POST("", func(x echo.Context) error {
		return app.TagEdit(z, x)
	})
	tag.POST("/info", func(x echo.Context) error {
		return app.TagInfo(z, x)
	})

	return e, nil
}

// Moved redirects are partial URL routers that are to be redirected with a HTTP 301 Moved Permanently.
func (c Configuration) Moved(z *zap.SugaredLogger, e *echo.Echo) (*echo.Echo, error) {
	if e == nil {
		return nil, fmt.Errorf("%w: %s", ErrRoutes, "handler routes")
	}
	if z == nil {
		return nil, fmt.Errorf("%w: %s", ErrZap, "handler routes")
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
	fixes.GET("/acid", func(x echo.Context) error {
		return x.Redirect(code, "/g/"+releaser.Obfuscate("ACID PRODUCTIONS"))
	})
	fixes.GET("/ice", func(x echo.Context) error {
		return x.Redirect(code, "/g/"+releaser.Obfuscate("INSANE CREATORS ENTERPRISE"))
	})
	fixes.GET("/"+releaser.Obfuscate("pirates with attitude"), func(x echo.Context) error {
		return x.Redirect(code, "/g/"+releaser.Obfuscate("pirates with attitudes"))
	})
	fixes.GET("/"+releaser.Obfuscate("TRISTAR AND RED SECTOR INC"), func(x echo.Context) error {
		return x.Redirect(code, "/g/"+releaser.Obfuscate("TRISTAR & RED SECTOR INC"))
	})
	return e, nil
}
