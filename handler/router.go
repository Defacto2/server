package handler

// Package file router.go contains the custom router URIs for the website.

import (
	"database/sql"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"strings"

	"github.com/Defacto2/helper"
	"github.com/Defacto2/server/handler/app"
	"github.com/Defacto2/server/handler/htmx"
	"github.com/Defacto2/server/handler/releaser"
	"github.com/Defacto2/server/handler/sitemap"
	"github.com/Defacto2/server/internal/config"
	"github.com/Defacto2/server/internal/dir"
	"github.com/Defacto2/server/internal/nils"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/v5/session"
	"github.com/labstack/echo/v5"
)

const code = http.StatusMovedPermanently

// RouteFiles defines the file locations and routes for the web server.
func (serv *Server) RouteFiles(sl *slog.Logger, e *echo.Echo, db *sql.DB, fsys fs.FS) (*echo.Echo, error) {
	const format = "route files %s: %w"
	if err := nils.Check(sl, e, db, fsys); err != nil {
		return nil, fmt.Errorf(format, "check", err)
	}

	if d, err := fs.ReadDir(fsys, "."); err != nil || len(d) == 0 {
		return nil, fmt.Errorf(format, "read dir", nils.ErrFS)
	}

	app.Caching.Records(serv.RecordCount)
	dirs := app.Dirs{ //nolint:exhaustruct
		Download:  dir.Directory(serv.Environment.AbsDownload),
		Preview:   dir.Directory(serv.Environment.AbsPreview),
		Thumbnail: dir.Directory(serv.Environment.AbsThumbnail),
		Extra:     dir.Directory(serv.Environment.AbsExtra),
		URI:       "", // URI is set later from route parameter
	}

	nonce, err := serv.nonce(e)
	if err != nil {
		return nil, fmt.Errorf(format, "nonce session key", err)
	}

	e = serv.signin(sl, e, nonce)
	e = serv.custom404(sl, e)
	e = serv.debugInfo(e)
	e = serv.static(e)
	e = serv.html(e, fsys)
	e = serv.font(e, fsys)
	e = serv.embed(e, fsys)
	e = serv.search(sl, e, db)
	e = serv.mainsite(sl, e, db, dirs)
	e = serv.apiv1(sl, e, db, fsys)
	e, err = serv.lock(sl, e, db, dirs)
	if err != nil {
		return nil, fmt.Errorf(format, "serve lock", err)
	}

	return e, nil
}

// nonce configures and returns the session key for the cookie store.
// If the read mode is enabled then an empty session key is returned.
func (serv *Server) nonce(e *echo.Echo) ([]byte, error) {
	const format = "nonce cookie store: %w"
	if err := nils.Check(e); err != nil {
		return []byte{}, fmt.Errorf(format, err)
	}
	if serv.Environment.ReadOnly {
		return []byte{}, nil
	}

	keyPairs, err := helper.CookieStore(serv.Environment.SessionKey.String())
	if err != nil {
		return []byte{}, fmt.Errorf(format, err)
	}

	e.Use(session.Middleware(sessions.NewCookieStore(keyPairs)))

	return keyPairs, nil
}

// html serves the embedded CSS, JS, WASM, and source map files for the HTML website layout.
func (serv *Server) html(e *echo.Echo, fsys fs.FS) *echo.Echo {
	const format = "html routes: %w"
	if err := nils.Check(e, fsys); err != nil {
		panic(fmt.Errorf(format, err))
	}

	paths, names := *app.Hrefs(), *app.Names()
	for key, path := range paths {
		e.FileFS(path, names[key], fsys)
	}

	// source map files
	const mapext = ".map"
	e.FileFS(paths[app.Bootstrap5]+mapext, names[app.Bootstrap5]+mapext, fsys)
	e.FileFS(paths[app.Bootstrap5JS]+mapext, names[app.Bootstrap5JS]+mapext, fsys)
	e.FileFS(paths[app.Jsdos6JS]+mapext, names[app.Jsdos6JS]+mapext, fsys)

	return e
}

// font serves the embedded woff2, woff, and ttf font files for the website layout.
func (serv *Server) font(e *echo.Echo, fsys fs.FS) *echo.Echo {
	const format = "font routes: %w"
	if err := nils.Check(e, fsys); err != nil {
		panic(fmt.Errorf(format, err))
	}

	paths, names := *app.FontRefs(), *app.FontNames()
	font := e.Group("/font")
	for key, path := range paths {
		font.FileFS(path, names[key], fsys)
	}

	return e
}

// embed serves the miscellaneous embedded files for the website layout.
// This includes the favicon, robots.txt, osd.xml, and the SVG icons.
func (serv *Server) embed(e *echo.Echo, fsys fs.FS) *echo.Echo {
	const format = "embed routes: %w"
	if err := nils.Check(e, fsys); err != nil {
		panic(fmt.Errorf(format, err))
	}

	e.FileFS("/favicon.ico", "public/image/favicon.ico", fsys)
	e.FileFS("/license.xml", "public/text/license.xml", fsys)
	e.FileFS("/osd.xml", "public/text/osd.xml", fsys)
	e.FileFS("/robots.txt", "public/text/robots.txt", fsys)
	// wdosbox is required by `js-dos.js`
	e.FileFS("/js/wdosbox.wasm.js", "public/js/wdosbox.wasm", fsys)

	return e
}

// static serves the static assets for the website such as the thumbnail and preview images.
func (serv *Server) static(e *echo.Echo) *echo.Echo {
	const format = "static routes: %w"
	if err := nils.Check(e); err != nil {
		panic(fmt.Errorf(format, err))
	}

	e.Static(config.StaticThumb(), serv.Environment.AbsThumbnail.String())
	e.Static(config.StaticOriginal(), serv.Environment.AbsPreview.String())

	return e
}

// custom404 is a custom 404 error handler for the website,
// "The page cannot be found".
func (serv *Server) custom404(sl *slog.Logger, e *echo.Echo) *echo.Echo {
	const format = "custom 404 error routes: %w"
	if err := nils.Check(sl, e); err != nil {
		panic(fmt.Errorf(format, err))
	}

	e.GET("/:uri", func(c *echo.Context) error {
		return app.StatusErr(sl, c, http.StatusNotFound, c.Param("uri"))
	})

	return e
}

// debugInfo returns detailed information about the HTTP request.
func (serv *Server) debugInfo(e *echo.Echo) *echo.Echo {
	const format = "debug info routes: %w"
	if err := nils.Check(e); err != nil {
		panic(fmt.Errorf(format, err))
	}
	if serv.Environment.ProdMode {
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
	e.GET("/debug", func(c *echo.Context) error {
		req := c.Request()
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
		return c.JSONPretty(http.StatusOK, d, "  ")
	})

	return e
}

// apiv1 routes for the public API endpoints.
func (serv *Server) apiv1(sl *slog.Logger, e *echo.Echo, db *sql.DB, fsys fs.FS,
) *echo.Echo {
	const format = "api routes %s: %w"
	if err := nils.Check(sl, e, db, fsys); err != nil {
		panic(fmt.Errorf(format, "check", err))
	}

	e.FileFS("/openapi.json", "public/json/openapi.json", fsys)
	// register API routes as a group to use a custom HTTP header
	e.GET("/api", func(c *echo.Context) error { return app.APIInfo(sl, c) })
	v1 := e.Group(app.APIBase)
	v1.Use(CacheMiddleware())
	v1.Use(APIMiddleware)
	v1.GET("/categories", func(c *echo.Context) error { return app.CategoriesAPI(c, db) })
	v1.GET("/category/:category", func(c *echo.Context) error { return app.CategoryAPI(sl, c, db) })
	v1.GET("/platforms", func(c *echo.Context) error { return app.PlatformsAPI(c, db) })
	v1.GET("/platform/:platform", func(c *echo.Context) error { return app.PlatformAPI(sl, c, db) })
	v1.GET("/milestones", app.MilestonesAPI)
	v1.GET("/milestones/highlights", app.MilestoneHighlightsAPI)
	v1.GET("/milestones/year/:year", app.MilestoneYearAPI)
	v1.GET("/milestones/years/:range", app.MilestoneYearsAPI)
	v1.GET("/milestones/decade/:decade", app.MilestoneDecadeAPI)
	v1.GET("/areacodes", app.AreacodesAPI)
	v1.GET("/areacode/:code", app.AreaCodeAPI)
	v1.GET("/areacodes/search/:query", app.AreacodeSearchAPI)
	v1.GET("/areacodes/regions", app.RegionsAPI)
	v1.GET("/areacodes/region/:abbr", app.RegionAPI)
	v1.GET("/websites", app.WebsitesAPI)
	v1.GET("/demozoo", app.DemozooAPI)
	v1.GET("/groups", func(c *echo.Context) error { return app.GroupsAPI(sl, c, db) })
	v1.GET("/sites", func(c *echo.Context) error { return app.SitesAPI(sl, c, db) })
	v1.GET("/boards", func(c *echo.Context) error { return app.BoardsAPI(sl, c, db) })
	v1.GET("/magazines", func(c *echo.Context) error { return app.MagazinesAPI(sl, c, db) })
	v1.GET("/releaser/:name", func(c *echo.Context) error { return app.ReleaserAPI(sl, c, db) })
	v1.GET("/artifacts", func(c *echo.Context) error { return app.ArtifactsAPI(sl, c, db) })
	v1.GET("/artifacts/new", func(c *echo.Context) error { return app.ArtifactsNewAPI(sl, c, db) })
	v1.GET("/artifact/:id", func(c *echo.Context) error { return app.FileAPI(sl, c, db) })
	v1.GET("/sceners", func(c *echo.Context) error { return app.ScenersAPI(sl, c, db) })
	v1.GET("/sceners/artist", func(c *echo.Context) error { return app.ArtistsAPI(sl, c, db) })
	v1.GET("/sceners/coder", func(c *echo.Context) error { return app.CodersAPI(sl, c, db) })
	v1.GET("/sceners/musician", func(c *echo.Context) error { return app.MusiciansAPI(sl, c, db) })
	v1.GET("/sceners/writer", func(c *echo.Context) error { return app.WritersAPI(sl, c, db) })
	v1.GET("/scener/:name", func(c *echo.Context) error { return app.ScenerAPI(sl, c, db) })

	return e
}

// mainsite routes for the main site.
func (serv *Server) mainsite(sl *slog.Logger, e *echo.Echo, db *sql.DB, dirs app.Dirs) *echo.Echo {
	const format = "mainsite routes %s: %w"
	if err := nils.Check(sl, db, e); err != nil {
		panic(fmt.Errorf(format, "check", err))
	}

	e = health(e)
	e = sitemaps(sl, e, db)
	s := e.Group("")
	s = sites(sl, s, db)
	s = serv.files(sl, s, db, dirs)
	s = serv.sceners(sl, s, db)
	s = serv.releasers(sl, s, db)

	s.GET("/", func(c *echo.Context) error { return app.Index(sl, c) })
	s.GET("/apps", func(c *echo.Context) error { return app.Apps(sl, c) })
	s.GET("/areacodes", func(c *echo.Context) error { return app.Areacodes(sl, c) })
	s.GET("/brokentexts", func(c *echo.Context) error { return app.BrokenTexts(sl, c) })
	s.GET("/compression", func(c *echo.Context) error { return app.Compression(sl, c) })
	s.GET("/fixes", func(c *echo.Context) error { return app.Fixes(sl, c) })
	s.GET("/history", func(c *echo.Context) error { return app.History(sl, c) })
	s.GET("/interview", func(c *echo.Context) error { return app.Interview(sl, c) })
	s.GET("/new", func(c *echo.Context) error { return app.New(sl, c) })
	s.GET("/terms", func(c *echo.Context) error { return app.Terms(sl, c) })
	s.GET("/thanks", func(c *echo.Context) error { return app.Thanks(sl, c) })
	s.GET("/thescene", func(c *echo.Context) error { return app.TheScene(sl, c) })
	s.GET("/titles", func(c *echo.Context) error { return app.Titles(sl, c) })

	return e
}

func (serv *Server) releasers(sl *slog.Logger, s *echo.Group, db *sql.DB) *echo.Group {
	const moved = http.StatusMovedPermanently

	releaser := func(c *echo.Context) error {
		uri := c.Param("id")
		if unwanted := c.QueryString(); unwanted != "" {
			return c.Redirect(moved, "/g/"+uri)
		}
		return app.Releasers(sl, c, db, uri, serv.Public)
	}

	s.GET("/g/:id", releaser)

	s.GET("/releaser", func(c *echo.Context) error {
		return app.Releaser(sl, c, db)
	})
	s.GET("/releaser/a-z", func(c *echo.Context) error {
		return app.ReleaserAZ(sl, c, db)
	})
	s.GET("/releaser/year", func(c *echo.Context) error {
		return app.ReleaserYear(sl, c, db)
	})

	s.GET("/magazine", func(c *echo.Context) error {
		return app.Magazine(sl, c, db)
	})
	s.GET("/magazine/a-z", func(c *echo.Context) error {
		return app.MagazineAZ(sl, c, db)
	})

	return s
}

func (serv *Server) sceners(sl *slog.Logger, s *echo.Group, db *sql.DB) *echo.Group {
	const moved = http.StatusMovedPermanently

	scener := func(c *echo.Context) error {
		uri := c.Param("id")
		if unwanted := c.QueryString(); unwanted != "" {
			return c.Redirect(moved, "/p/"+uri)
		}
		return app.Sceners(sl, c, db, uri)
	}

	s.GET("/p/:id", scener)

	s.GET("/scener", func(c *echo.Context) error {
		return app.Scener(sl, c, db)
	})
	s.GET("/artist", func(c *echo.Context) error {
		return app.Artist(sl, c, db)
	})
	s.GET("/coder", func(c *echo.Context) error {
		return app.Coder(sl, c, db)
	})
	s.GET("/musician", func(c *echo.Context) error {
		return app.Musician(sl, c, db)
	})
	s.GET("/writer", func(c *echo.Context) error {
		return app.Writer(sl, c, db)
	})

	return s
}

func (serv *Server) files(sl *slog.Logger, s *echo.Group, db *sql.DB, dirs app.Dirs) *echo.Group {
	const moved = http.StatusMovedPermanently

	artifact := func(c *echo.Context) error {
		uri := c.Param("id")
		if unwanted := c.QueryString(); unwanted != "" {
			return c.Redirect(moved, "/f/"+uri)
		}
		dirs.URI = uri
		dirs.ReadOnly = bool(serv.Environment.ReadOnly)
		return dirs.Artifact(sl, c, db)
	}

	s.GET(Downloader, func(c *echo.Context) error {
		return app.Download(sl, c, db, dir.Directory(serv.Environment.AbsDownload))
	})
	s.GET("/f/:id", artifact)
	s.GET("/file/stats", func(c *echo.Context) error {
		return app.Categories(sl, c, db, true)
	})
	s.GET("/files/:id/:page", func(c *echo.Context) error {
		switch c.Param("id") {
		case
			"for-approval", "deletions", "unwanted":
			return app.StatusErr(sl, c, http.StatusNotFound, c.Param("id"))
		}
		return app.Artifacts(sl, c, db, c.Param("id"), c.Param("page"))
	})
	s.GET("/files/:id", func(c *echo.Context) error {
		switch c.Param("id") {
		case
			"for-approval", "deletions", "unwanted":
			return app.StatusErr(sl, c, http.StatusNotFound, c.Param("id"))
		}
		return app.Artifacts(sl, c, db, c.Param("id"), "1")
	})
	s.GET("/file", func(c *echo.Context) error {
		return app.Categories(sl, c, db, false)
	})
	s.GET("/jsdos/:id", func(c *echo.Context) error {
		return app.DownloadJsDos(sl, c, db,
			dir.Directory(serv.Environment.AbsExtra),
			dir.Directory(serv.Environment.AbsDownload))
	})
	s.GET("/sum/:id", func(c *echo.Context) error {
		return app.Checksum(sl, c, db, c.Param("id"))
	})
	s.GET("/v/:id", func(c *echo.Context) error {
		return app.Inline(sl, c, db, dir.Directory(serv.Environment.AbsDownload))
	})

	return s
}

func sitemaps(sl *slog.Logger, e *echo.Echo, db *sql.DB) *echo.Echo {
	e.GET("/sitemaps.xml", func(c *echo.Context) error {
		i := sitemap.MapIndex()
		return c.XMLPretty(http.StatusOK, i, "  ")
	})
	e.GET("/"+sitemap.Website, func(c *echo.Context) error {
		ctx := c.Request().Context()
		i := sitemap.MapSite(ctx, sl, db)
		return c.XMLPretty(http.StatusOK, i, "  ")
	})
	e.GET("/"+sitemap.Releaser, func(c *echo.Context) error {
		ctx := c.Request().Context()
		i := sitemap.MapReleaser(ctx, sl, db)
		return c.XMLPretty(http.StatusOK, i, "  ")
	})
	e.GET("/"+sitemap.Magazine, func(c *echo.Context) error {
		ctx := c.Request().Context()
		i := sitemap.MapMagazine(ctx, sl, db)
		return c.XMLPretty(http.StatusOK, i, "  ")
	})
	e.GET("/"+sitemap.BBS, func(c *echo.Context) error {
		ctx := c.Request().Context()
		i := sitemap.MapBBS(ctx, sl, db)
		return c.XMLPretty(http.StatusOK, i, "  ")
	})
	e.GET("/"+sitemap.FTP, func(c *echo.Context) error {
		ctx := c.Request().Context()
		i := sitemap.MapFTP(ctx, sl, db)
		return c.XMLPretty(http.StatusOK, i, "  ")
	})

	return e
}

func sites(sl *slog.Logger, s *echo.Group, db *sql.DB) *echo.Group {
	s.GET("/bbs", func(c *echo.Context) error {
		return app.BBS(sl, c, db)
	})
	s.GET("/bbs/a-z", func(c *echo.Context) error {
		return app.BBSAZ(sl, c, db)
	})
	s.GET("/bbs/year", func(c *echo.Context) error {
		return app.BBSYear(sl, c, db)
	})
	s.GET("/ftp", func(c *echo.Context) error {
		return app.FTP(sl, c, db)
	})
	s.GET("/website/:id", func(c *echo.Context) error {
		return app.Website(sl, c, c.Param("id"))
	})
	s.GET("/website", func(c *echo.Context) error {
		return app.Website(sl, c, "")
	})
	s.GET("/pouet/vote/:id", func(c *echo.Context) error {
		return app.VotePouet(sl, c, c.Param("id"))
	})
	s.GET("/pouet/prod/:id", func(c *echo.Context) error {
		return app.ProdPouet(c, c.Param("id"))
	})
	s.GET("/zoo/prod/:id", func(c *echo.Context) error {
		return app.ProdZoo(c, c.Param("id"))
	})

	return s
}

func health(e *echo.Echo) *echo.Echo {
	e.GET("/health-check", func(c *echo.Context) error {
		return c.NoContent(http.StatusOK)
	})
	return e
}

// search forms and the results for database queries.
func (serv *Server) search(sl *slog.Logger, e *echo.Echo, db *sql.DB) *echo.Echo {
	const format = "search routes: %w"
	if err := nils.Check(sl, e, db); err != nil {
		panic(fmt.Errorf(format, err))
	}

	// this legacy get result should be kept for (osx.xml) opensearch compatibility
	// and to keep possible backwards compatibility with third party site links.
	opensearch := func(c *echo.Context) error {
		terms := strings.ReplaceAll(c.QueryParam("query"), "+", " ") // AND replacement
		terms = strings.ReplaceAll(terms, "|", ",")                  // OR replacement
		return app.PostDesc(sl, c, db, terms)
	}

	search := e.Group("/search")
	search.GET("/desc", func(c *echo.Context) error { return app.SearchDesc(sl, c) })
	search.GET("/file", func(c *echo.Context) error { return app.SearchFile(sl, c) })
	search.GET("/releaser", func(c *echo.Context) error { return app.SearchReleaser(sl, c) })
	search.GET("/result", opensearch)
	search.POST("/desc", func(c *echo.Context) error {
		return app.PostDesc(sl, c, db, c.FormValue("search-term-query"))
	})
	search.POST("/file", func(c *echo.Context) error {
		return app.PostFilename(sl, c, db)
	})
	search.POST("/releaser", func(c *echo.Context) error {
		return htmx.SearchReleaser(sl, c, db, &serv.TidbitIndex)
	})

	return e
}

// signin for operators.
func (serv *Server) signin(sl *slog.Logger, e *echo.Echo, nonce []byte) *echo.Echo {
	const format = "signin routes: %w"
	if err := nils.Check(sl, e); err != nil {
		panic(fmt.Errorf(format, err))
	}

	readonlylock := func(c echo.HandlerFunc) echo.HandlerFunc {
		return serv.ReadOnlyLock(c, sl)
	}

	signings := e.Group("")
	signings.Use(readonlylock)
	signings.GET("/signedout", func(c *echo.Context) error {
		return app.SignedOut(sl, c)
	})
	signings.GET("/signin", func(c *echo.Context) error {
		return app.Signin(sl, c, serv.Environment.GoogleClientID.String(), nonce)
	})
	signings.GET("/operator/signin", func(c *echo.Context) error {
		return c.Redirect(http.StatusMovedPermanently, "/signin")
	})
	google := signings.Group("/google")
	google.POST("/callback", func(c *echo.Context) error {
		return app.GoogleCallback(sl, c,
			serv.Environment.GoogleClientID.String(),
			serv.Environment.SessionMaxAge.Int(),
			serv.Environment.GoogleAccounts...)
	})

	return e
}

// AppendMoved redirects are partial URL routers that are to be redirected with a HTTP 301 Moved Permanently.
func AppendMoved(e *echo.Echo) *echo.Echo {
	const msg = "moved permanently routes"
	if err := nils.Check(e); err != nil {
		panic(fmt.Errorf("%s: %w", msg, err))
	}

	e = nginx(e)
	e = fixes(e)
	return e
}

// nginx (legacy tool) redirects.
func nginx(e *echo.Echo) *echo.Echo {
	const format = "nginx redirects: %w"
	if err := nils.Check(e); err != nil {
		panic(fmt.Errorf(format, err))
	}

	nginx := e.Group("")
	nginx.GET("/file/detail/:id", func(c *echo.Context) error {
		return c.Redirect(code, "/f/"+c.Param("id"))
	})
	nginx.GET("/file/download/:id", func(c *echo.Context) error {
		return c.Redirect(code, "/d/"+c.Param("id"))
	})
	nginx.GET("/file/view/:id", func(c *echo.Context) error {
		return c.Redirect(code, "/v/"+c.Param("id"))
	})
	nginx.GET("/cracktros-detail.cfm/:id", func(c *echo.Context) error {
		return c.Redirect(code, "/f/"+c.Param("id"))
	})
	nginx.GET("/wayback/:url", func(c *echo.Context) error {
		return c.Redirect(code, "https://wayback.defacto2.net/"+c.Param("url"))
	})
	nginx.GET("/link/list", func(c *echo.Context) error {
		return c.Redirect(code, "https://wayback.defacto2.net/")
	})

	return e
}

// fixes redirects repaired, releaser database entry redirects that are contained in the model fix package.
func fixes(e *echo.Echo) *echo.Echo {
	const format = "fixes routers: %w"
	if err := nils.Check(e); err != nil {
		panic(fmt.Errorf(format, err))
	}

	fixes := e.Group("/g")
	const g = "/g/"
	fixes.GET("/acid", func(c *echo.Context) error {
		return c.Redirect(code, g+releaser.Obfuscate("ACID PRODUCTIONS"))
	})
	fixes.GET("/ansi-creators-in-demand", func(c *echo.Context) error {
		return c.Redirect(code, g+releaser.Obfuscate("ACID PRODUCTIONS"))
	})
	fixes.GET("/ice", func(c *echo.Context) error {
		return c.Redirect(code, g+releaser.Obfuscate("INSANE CREATORS ENTERPRISE"))
	})
	fixes.GET("/rss", func(c *echo.Context) error {
		return c.Redirect(code, g+"renaissance")
	})
	fixes.GET("/trsi", func(c *echo.Context) error {
		return c.Redirect(code, g+releaser.Obfuscate("TRISTAR & RED SECTOR INC"))
	})
	fixes.GET("/x-pression", func(c *echo.Context) error {
		return c.Redirect(code, g+releaser.Obfuscate("X-PRESSION DESIGN"))
	})
	fixes.GET("/"+releaser.Obfuscate("DAMN EXCELLENT ANSI DESIGNERS"), func(c *echo.Context) error {
		return c.Redirect(code, g+releaser.Obfuscate("DAMN EXCELLENT ANSI DESIGN"))
	})
	fixes.GET("/"+releaser.Obfuscate("pirates with attitude"), func(c *echo.Context) error {
		return c.Redirect(code, g+releaser.Obfuscate("pirates with attitudes"))
	})
	fixes.GET("/"+releaser.Obfuscate("TRISTAR AND RED SECTOR INC"), func(c *echo.Context) error {
		return c.Redirect(code, g+releaser.Obfuscate("TRISTAR & RED SECTOR INC"))
	})
	fixes.GET("/"+releaser.Obfuscate("THE ORIGINAL FUNNY GUYS"), func(c *echo.Context) error {
		return c.Redirect(code, g+releaser.Obfuscate("ORIGINALLY FUNNY GUYS"))
	})
	fixes.GET("/"+releaser.Obfuscate("ORIGINAL FUNNY GUYS"), func(c *echo.Context) error {
		return c.Redirect(code, g+releaser.Obfuscate("ORIGINALLY FUNNY GUYS"))
	})
	fixes.GET("/"+releaser.Obfuscate("DARKSIDE INC"), func(c *echo.Context) error {
		return c.Redirect(code, g+releaser.Obfuscate("DARKSIDE INCORPORATED"))
	})
	fixes.GET("/united-software-association", func(c *echo.Context) error {
		return c.Redirect(code, g+"united-software-association*fairlight")
	})

	return e
	// THESE ARE NOT WORKING, public-enemy/ and the-dream-team/ get redirected
	// fixes.GET(`/public-enemy*tristar-ampersand-red-sector-inc*the-dream-team`, func(c *echo.Context) error {
	// 	return c.Redirect(code, g+"pe*trsi*tdt")
	// })
	// fixes.GET(`/the-dream-team*tristar-ampersand-red-sector-inc`, func(c *echo.Context) error {
	// 	return c.Redirect(code, g+"coop")
	// })
}
