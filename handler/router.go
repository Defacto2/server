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
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// Routes defines the routes for the web server.
func (conf Configuration) Routes(z *zap.SugaredLogger, e *echo.Echo, public embed.FS) (*echo.Echo, error) {
	if e == nil {
		return nil, fmt.Errorf("%w: %s", ErrRoutes, "handler routes")
	}
	if z == nil {
		return nil, fmt.Errorf("%w: %s", ErrZap, "handler routes")
	}

	// Cache the database record count.
	app.Caching.RecordCount = conf.RecordCount

	// Set the application configuration for paths.
	dir := app.Dirs{
		Download:  conf.Import.DownloadDir,
		Preview:   conf.Import.PreviewDir,
		Thumbnail: conf.Import.ThumbnailDir,
	}

	// Serve embedded CSS files
	e.FileFS(app.BootCSS, app.BootCPub, public)
	e.FileFS(app.BootCSS+".map", app.BootCPub+".map", public)
	e.FileFS(app.LayoutCSS, app.LayoutPub, public)

	// Serve embedded SVG collections
	e.FileFS("/bootstrap-icons.svg", "public/image/bootstrap-icons.svg", public)

	// Serve embedded font files
	e.FileFS("/font/pxplus_ibm_vga8.woff2", "public/font/pxplus_ibm_vga8.woff2", public)
	e.FileFS("/font/pxplus_ibm_vga8.woff", "public/font/pxplus_ibm_vga8.woff", public)
	e.FileFS("/font/pxplus_ibm_vga8.ttf", "public/font/pxplus_ibm_vga8.ttf", public)
	e.FileFS("/font/topazplus_a1200.woff2", "public/font/topazplus_a1200.woff2", public)
	e.FileFS("/font/topazplus_a1200.woff", "public/font/topazplus_a1200.woff", public)
	e.FileFS("/font/topazplus_a1200.ttf", "public/font/topazplus_a1200.ttf", public)

	// Serve embedded JS files
	e.FileFS(app.BootJS, app.BootJPub, public)
	e.FileFS(app.BootJS+".map", app.BootJPub+".map", public)
	e.FileFS(app.EditorJS, app.EditorJSPub, public)
	e.FileFS(app.EditorAssetsJS, app.EditorAssetsJSPub, public)
	e.FileFS(app.EditorArchiveJS, app.EditorArchiveJSPub, public)
	e.FileFS(app.FAJS, app.FAPub, public)
	e.FileFS(app.PouetJS, app.PouetPub, public)
	e.FileFS(app.ReadmeJS, app.ReadmePub, public)
	e.FileFS(app.RestPouetJS, app.RestPouetPub, public)
	e.FileFS(app.RestZooJS, app.RestZooPub, public)
	e.FileFS(app.UploaderJS, app.UploaderPub, public)

	// Serve embedded JS DOS files
	e.FileFS(app.JSDos, app.JSDosPub, public)
	e.FileFS(app.JSWDos, app.JSWDosPub, public)
	e.FileFS("/js/wdosbox.wasm", "public/js/wdosbox.wasm", public)
	e.FileFS("/js/js-dos.js.map", "public/js/js-dos.js.map", public)

	// Serve embedded image files
	e.FileFS("/favicon.ico", "public/image/favicon.ico", public)

	// Serve embedded text files
	e.FileFS("/osd.xml", "public/text/osd.xml", public)
	e.FileFS("/robots.txt", "public/text/robots.txt", public)
	e.FileFS("/site.webmanifest", "public/text/site.webmanifest.json", public)

	// Serve asset images
	e.Static(config.StaticThumb(), conf.Import.ThumbnailDir)
	e.Static(config.StaticOriginal(), conf.Import.PreviewDir)

	e.GET("/", func(c echo.Context) error {
		return app.Index(z, c)
	})
	e.GET("/artist", func(c echo.Context) error {
		return app.Artist(z, c)
	})
	e.GET("/bbs", func(c echo.Context) error {
		return app.BBS(z, c)
	})
	e.GET("/bbs/a-z", func(c echo.Context) error {
		return app.BBSAZ(z, c)
	})
	e.GET("/coder", func(c echo.Context) error {
		return app.Coder(z, c)
	})
	e.GET(Downloader, func(c echo.Context) error {
		return app.Download(z, c, conf.Import.DownloadDir)
	})
	e.GET("/f/:id", func(c echo.Context) error {
		dir.URI = c.Param("id")
		return dir.About(z, c, conf.Import.IsReadOnly)
	})
	e.GET("/file/stats", func(c echo.Context) error {
		return app.File(z, c, true)
	})
	e.GET("/files/:id/:page", func(c echo.Context) error {
		return app.Files(z, c, c.Param("id"), c.Param("page"))
	})
	e.GET("/files/:id", func(c echo.Context) error {
		return app.Files(z, c, c.Param("id"), "1")
	})
	e.GET("/file", func(c echo.Context) error {
		return app.File(z, c, false)
	})
	e.GET("/ftp", func(c echo.Context) error {
		return app.FTP(z, c)
	})
	e.GET("/g/:id", func(c echo.Context) error {
		return app.Releasers(z, c, c.Param("id"))
	})
	e.GET("/history", func(c echo.Context) error {
		return app.History(z, c)
	})
	e.GET("/interview", func(c echo.Context) error {
		return app.Interview(z, c)
	})
	e.GET("/magazine", func(c echo.Context) error {
		return app.Magazine(z, c)
	})
	e.GET("/magazine/a-z", func(c echo.Context) error {
		return app.MagazineAZ(z, c)
	})
	e.GET("/musician", func(c echo.Context) error {
		return app.Musician(z, c)
	})
	e.GET("/p/:id", func(c echo.Context) error {
		return app.Sceners(z, c, c.Param("id"))
	})
	e.GET("/pouet/vote/:id", func(c echo.Context) error {
		return app.VotePouet(z, c, c.Param("id"))
	})
	e.GET("/pouet/prod/:id", func(c echo.Context) error {
		return app.ProdPouet(z, c, c.Param("id"))
	})
	e.GET("/zoo/prod/:id", func(c echo.Context) error {
		return app.ProdZoo(z, c, c.Param("id"))
	})
	e.GET("/r/:id", func(c echo.Context) error {
		return app.Reader(z, c, c.Param("id"))
	})
	e.GET("/releaser", func(c echo.Context) error {
		return app.Releaser(z, c)
	})
	e.GET("/releaser/a-z", func(c echo.Context) error {
		return app.ReleaserAZ(z, c)
	})
	e.GET("/scener", func(c echo.Context) error {
		return app.Scener(z, c)
	})
	e.GET("/search/file", func(c echo.Context) error {
		return app.SearchFile(z, c)
	})
	e.POST("/search/file", func(c echo.Context) error {
		return app.PostFilename(z, c)
	})
	e.GET("/search/desc", func(c echo.Context) error {
		return app.SearchDesc(z, c)
	})
	e.POST("/search/desc", func(c echo.Context) error {
		return app.PostDesc(z, c, c.FormValue("search-term-query"))
	})
	e.GET("/search/releaser", func(c echo.Context) error {
		return app.SearchReleaser(z, c)
	})
	e.POST("/search/releaser", func(c echo.Context) error {
		return app.PostReleaser(z, c)
	})
	e.GET("/search/result", func(c echo.Context) error {
		// this legacy get result should be kept for (osx.xml) opensearch compatibility
		// and to keep possible backwards compatibility with third party site links.
		terms := strings.ReplaceAll(c.QueryParam("query"), "+", " ") // AND replacement
		terms = strings.ReplaceAll(terms, "|", ",")                  // OR replacement
		return app.PostDesc(z, c, terms)
	})
	e.GET("/sum/:id", func(c echo.Context) error {
		return app.Checksum(z, c, c.Param("id"))
	})
	e.GET("/thanks", func(c echo.Context) error {
		return app.Thanks(z, c)
	})
	e.GET("/thescene", func(c echo.Context) error {
		return app.TheScene(z, c)
	})
	e.GET("/website/:id", func(c echo.Context) error {
		return app.Website(z, c, c.Param("id"))
	})
	e.GET("/website", func(c echo.Context) error {
		return app.Website(z, c, "")
	})
	e.GET("/writer", func(c echo.Context) error {
		return app.Writer(z, c)
	})
	e.POST("/uploader", func(c echo.Context) error {
		return app.PostIntro(z, c)
	})
	e.GET("/v/:id", func(c echo.Context) error {
		return app.Inline(z, c, conf.Import.DownloadDir)
	})
	// all other page requests return a custom 404 error page
	e.GET("/:uri", func(c echo.Context) error {
		return app.StatusErr(z, c, http.StatusNotFound, c.Param("uri"))
	})

	// Skip the serving of the editor pages
	if conf.Import.IsReadOnly {
		return e, nil
	}
	// TODO: Implement a middleware to check for a valid session cookie.
	// and exit here if not valid.
	e.POST("/editor/online/true", func(c echo.Context) error {
		return app.RecordToggle(z, c, true)
	})
	e.POST("/editor/online/false", func(c echo.Context) error {
		return app.RecordToggle(z, c, false)
	})

	e.POST("/editor/readme/copy", func(c echo.Context) error {
		return app.ReadmePost(z, c, dir.Download)
	})
	e.POST("/editor/readme/delete", func(c echo.Context) error {
		return app.ReadmeDel(z, c, dir.Download)
	})
	e.POST("/editor/readme/hide", func(c echo.Context) error {
		dir.URI = c.Param("id")
		return app.ReadmeToggle(z, c)
	})
	e.POST("/editor/images/copy", func(c echo.Context) error {
		return dir.PreviewPost(z, c)
	})
	e.POST("/editor/images/delete", func(c echo.Context) error {
		return dir.PreviewDel(z, c)
	})
	e.POST("/editor/ansilove/copy", func(c echo.Context) error {
		return dir.AnsiLovePost(z, c)
	})

	e.POST("/editor/title", func(c echo.Context) error {
		return app.TitleEdit(z, c)
	})
	e.POST("/editor/ymd", func(c echo.Context) error {
		return app.YMDEdit(z, c)
	})

	e.POST("/editor/platform", func(c echo.Context) error {
		return app.PlatformEdit(z, c)
	})
	e.POST("/editor/tag", func(c echo.Context) error {
		return app.TagEdit(z, c)
	})
	e.POST("/editor/platform+tag", func(c echo.Context) error {
		return app.PlatformTagInfo(z, c)
	})
	e.POST("/editor/tag/info", func(c echo.Context) error {
		return app.TagInfo(z, c)
	})

	return e, nil
}

// Moved redirects are partial URL routers that are to be redirected with a HTTP 301 Moved Permanently.
func (c Configuration) Moved(z *zap.SugaredLogger, e *echo.Echo) (*echo.Echo, error) {
	const code = http.StatusMovedPermanently
	// nginx redirects
	e.GET("/welcome", func(c echo.Context) error {
		return c.Redirect(code, "/")
	})
	e.GET("/file/download/:id", func(c echo.Context) error {
		return c.Redirect(code, "/d/"+c.Param("id"))
	})
	e.GET("/file/view/:id", func(c echo.Context) error {
		return c.Redirect(code, "/v/"+c.Param("id"))
	})
	e.GET("/apollo-x/fc.htm", func(c echo.Context) error {
		return c.Redirect(code, "/wayback/apollo-x-demo-resources-1999-december-17/fc.htm")
	})
	e.GET("/bbs.cfm", func(c echo.Context) error {
		return c.Redirect(code, "/bbs")
	})
	e.GET("/contact.cfm", func(c echo.Context) error {
		return c.Redirect(code, "/") // there's no dedicated contact page
	})
	e.GET("/cracktros.cfm", func(c echo.Context) error {
		return c.Redirect(code, "/files/intro")
	})
	e.GET("/cracktros-detail.cfm:/:id", func(c echo.Context) error {
		return c.Redirect(code, "/f/"+c.Param("id"))
	})
	e.GET("/documents.cfm", func(c echo.Context) error {
		return c.Redirect(code, "/files/text")
	})
	e.GET("/index.cfm", func(c echo.Context) error {
		return c.Redirect(code, "/")
	})
	e.GET("/index.cfm/:uri", func(c echo.Context) error {
		return c.Redirect(code, "/")
	})
	e.GET("/index.cfml/:uri", func(c echo.Context) error {
		return c.Redirect(code, "/")
	})
	e.GET("/groups.cfm", func(c echo.Context) error {
		return c.Redirect(code, "/releaser")
	})
	e.GET("/magazines.cfm", func(c echo.Context) error {
		return c.Redirect(code, "/magazine")
	})
	e.GET("/nfo-files.cfm", func(c echo.Context) error {
		return c.Redirect(code, "/files/nfo")
	})
	e.GET("/portal.cfm", func(c echo.Context) error {
		return c.Redirect(code, "/website")
	})
	e.GET("/rewrite.cfm", func(c echo.Context) error {
		return c.Redirect(code, "/")
	})
	e.GET("/site-info.cfm", func(c echo.Context) error {
		return c.Redirect(code, "/") // there's no dedicated about page
	})
	// 2020 website redirects
	e.GET("/code", func(c echo.Context) error {
		return c.Redirect(code, "https://github.com/Defacto2/server")
	})
	e.GET("/commercial", func(c echo.Context) error {
		return c.Redirect(code, "/")
	})
	e.GET("/defacto", func(c echo.Context) error {
		return c.Redirect(code, "/history")
	})
	e.GET("/defacto2/donate", func(c echo.Context) error {
		return c.Redirect(code, "/thanks")
	})
	e.GET("/defacto2/history", func(c echo.Context) error {
		return c.Redirect(code, "/history")
	})
	e.GET("/defacto2/subculture", func(c echo.Context) error {
		return c.Redirect(code, "/thescene")
	})
	e.GET("/file/detail/:id", func(c echo.Context) error {
		return c.Redirect(code, "/f/"+c.Param("id"))
	})
	e.GET("/file/index", func(c echo.Context) error {
		return c.Redirect(code, "/file")
	})
	e.GET("/file/list/:uri", func(c echo.Context) error {
		return c.Redirect(code, "/files/new-uploads")
	})
	e.GET("/files/json/site.webmanifest", func(c echo.Context) error {
		return c.Redirect(code, "/site.webmanifest")
	})
	e.GET("/help/cc", func(c echo.Context) error {
		return c.Redirect(code, "/") // there's no dedicated contact page
	})
	e.GET("/help/privacy", func(c echo.Context) error {
		return c.Redirect(code, "/") // there's no dedicated privacy page
	})
	e.GET("/help/viruses", func(c echo.Context) error {
		return c.Redirect(code, "/") // there's no dedicated virus page
	})
	e.GET("/home", func(c echo.Context) error {
		return c.Redirect(code, "/")
	})
	e.GET("/link/list", func(c echo.Context) error {
		return c.Redirect(code, "/website")
	})
	e.GET("/link/list/:id", func(c echo.Context) error {
		return c.Redirect(code, "/website")
	})
	//nolint:misspell
	e.GET("/organisation/list/bbs", func(c echo.Context) error {
		return c.Redirect(code, "/bbs")
	})
	//nolint:misspell
	e.GET("/organisation/list/group", func(c echo.Context) error {
		return c.Redirect(code, "/releaser")
	})
	//nolint:misspell
	e.GET("/organisation/list/ftp", func(c echo.Context) error {
		return c.Redirect(code, "/ftp")
	})
	//nolint:misspell
	e.GET("/organisation/list/magazine", func(c echo.Context) error {
		return c.Redirect(code, "/magazine")
	})
	e.GET("/person/list", func(c echo.Context) error {
		return c.Redirect(code, "/scener")
	})
	e.GET("/person/list/artists", func(c echo.Context) error {
		return c.Redirect(code, "/artist")
	})
	e.GET("/person/list/coders", func(c echo.Context) error {
		return c.Redirect(code, "/coder")
	})
	e.GET("/person/list/musicians", func(c echo.Context) error {
		return c.Redirect(code, "/musician")
	})
	e.GET("/person/list/writers", func(c echo.Context) error {
		return c.Redirect(code, "/writer")
	})
	e.GET("/upload", func(c echo.Context) error {
		return c.Redirect(code, "/")
	})
	e.GET("/upload/file", func(c echo.Context) error {
		return c.Redirect(code, "/")
	})
	e.GET("/upload/external", func(c echo.Context) error {
		return c.Redirect(code, "/")
	})
	e.GET("/upload/intro", func(c echo.Context) error {
		return c.Redirect(code, "/")
	})
	e.GET("/upload/site", func(c echo.Context) error {
		return c.Redirect(code, "/")
	})
	e.GET("/upload/document", func(c echo.Context) error {
		return c.Redirect(code, "/")
	})
	e.GET("/upload/magazine", func(c echo.Context) error {
		return c.Redirect(code, "/")
	})
	e.GET("/upload/art", func(c echo.Context) error {
		return c.Redirect(code, "/")
	})
	e.GET("/upload/other", func(c echo.Context) error {
		return c.Redirect(code, "/")
	})
	// wayback redirects
	e.GET("/scene-archive/:uri", func(c echo.Context) error {
		return c.Redirect(code, "/")
	})
	e.GET("/includes/documentsweb/df2web99/scene-archive/history.html", func(c echo.Context) error {
		return c.Redirect(code, "/wayback/defacto2-from-1999-september-26/scene-archive/history.html")
	})
	e.GET("/includes/documentsweb/tKC_history.html", func(c echo.Context) error {
		return c.Redirect(code, "/wayback/the-life-and-legend-of-tkc-2000-october-10/index.html")
	})
	e.GET("/legacy/apollo-x/:uri", func(c echo.Context) error {
		return c.Redirect(code, "/wayback/apollo-x-demo-resources-1999-december-17/:uri")
	})
	e.GET("/web/20120827022026/http:/www.defacto2.net:80/file/list/nfotool", func(c echo.Context) error {
		return c.Redirect(code, "/files/nfo-tool")
	})
	e.GET("/web.pages/warez_world-1.htm", func(c echo.Context) error {
		return c.Redirect(code, "/wayback/warez-world-from-2001-july-26/index.html")
	})
	// repaired releaser database entry redirects
	e.GET("/g/acid", func(c echo.Context) error {
		return c.Redirect(code, "/g/"+releaser.Obfuscate("ACID PRODUCTIONS"))
	})
	e.GET("/g/ice", func(c echo.Context) error {
		return c.Redirect(code, "/g/"+releaser.Obfuscate("INSANE CREATORS ENTERPRISE"))
	})
	e.GET("/g/"+releaser.Obfuscate("pirates with attitude"), func(c echo.Context) error {
		return c.Redirect(code, "/g/"+releaser.Obfuscate("pirates with attitudes"))
	})
	e.GET("/g/"+releaser.Obfuscate("TRISTAR AND RED SECTOR INC"), func(c echo.Context) error {
		return c.Redirect(code, "/g/"+releaser.Obfuscate("TRISTAR & RED SECTOR INC"))
	})
	return e, nil
}
