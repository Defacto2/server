package handler

// Package file router.go contains the custom router URIs for the website.

import (
	"database/sql"
	"embed"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/Defacto2/helper"
	"github.com/Defacto2/releaser"
	"github.com/Defacto2/server/handler/app"
	"github.com/Defacto2/server/handler/htmx"
	"github.com/Defacto2/server/handler/sitemap"
	"github.com/Defacto2/server/internal/config"
	"github.com/Defacto2/server/internal/dir"
	"github.com/Defacto2/server/internal/panics"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/v5/session"
	"github.com/labstack/echo/v5"
)

const code = http.StatusMovedPermanently

// AppendFiles defines the file locations and routes for the web server.
func (c *Configuration) AppendFiles(sl *slog.Logger, e *echo.Echo, db *sql.DB, public embed.FS,
) (*echo.Echo, error) {
	const msg = "files routes"
	if err := panics.SDEP(sl, db, e, public); err != nil {
		panic(fmt.Errorf("%s: %w", msg, err))
	}
	if d, err := public.ReadDir("."); err != nil || len(d) == 0 {
		return nil, fmt.Errorf("%s: %w", msg, panics.ErrNoEmbed)
	}
	app.Caching.Records(c.RecordCount)
	dirs := app.Dirs{
		Download:  dir.Directory(c.Environment.AbsDownload),
		Preview:   dir.Directory(c.Environment.AbsPreview),
		Thumbnail: dir.Directory(c.Environment.AbsThumbnail),
		Extra:     dir.Directory(c.Environment.AbsExtra),
		URI:       "", // URI is set later from route parameter
	}
	nonce, err := c.nonce(e)
	if err != nil {
		return nil, fmt.Errorf("%s nonce session key: %w", msg, err)
	}
	e = c.signin(sl, e, nonce)
	e = c.custom404(sl, e)
	e = c.debugInfo(e)
	e = c.static(e)
	e = c.html(e, public)
	e = c.font(e, public)
	e = c.embed(e, public)
	e = c.search(sl, e, db)
	e = c.website(sl, e, db, dirs)
	e = c.api(sl, e, db, public)
	e = c.lock(sl, e, db, dirs)
	return e, nil
}

// nonce configures and returns the session key for the cookie store.
// If the read mode is enabled then an empty session key is returned.
func (c *Configuration) nonce(e *echo.Echo) (string, error) {
	const msg = "nonce cookie store"
	if e == nil {
		panic(fmt.Errorf("%s: %w", msg, panics.ErrNoEchoE))
	}
	if c.Environment.ReadOnly {
		return "", nil
	}
	b, err := helper.CookieStore(c.Environment.SessionKey.String())
	if err != nil {
		return "", fmt.Errorf("%s: %w", msg, err)
	}
	e.Use(session.Middleware(sessions.NewCookieStore(b)))
	return string(b), nil
}

// html serves the embedded CSS, JS, WASM, and source map files for the HTML website layout.
func (c *Configuration) html(e *echo.Echo, public embed.FS) *echo.Echo {
	const msg = "html routes"
	if e == nil {
		panic(fmt.Errorf("%s: %w", msg, panics.ErrNoEchoE))
	}
	hrefs, names := *app.Hrefs(), *app.Names()
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
func (c *Configuration) font(e *echo.Echo, public embed.FS) *echo.Echo {
	const msg = "font routes"
	if err := panics.EP(e, public); err != nil {
		panic(fmt.Errorf("%s: %w", msg, err))
	}
	paths, names := *app.FontRefs(), *app.FontNames()
	font := e.Group("/font")
	for key, href := range paths {
		font.FileFS(href, names[key], public)
	}
	return e
}

// embed serves the miscellaneous embedded files for the website layout.
// This includes the favicon, robots.txt, osd.xml, and the SVG icons.
func (c *Configuration) embed(e *echo.Echo, public embed.FS) *echo.Echo {
	const msg = "embed routes"
	if err := panics.EP(e, public); err != nil {
		panic(fmt.Errorf("%s: %w", msg, err))
	}
	e.FileFS("/favicon.ico", "public/image/favicon.ico", public)
	e.FileFS("/license.xml", "public/text/license.xml", public)
	e.FileFS("/osd.xml", "public/text/osd.xml", public)
	e.FileFS("/robots.txt", "public/text/robots.txt", public)
	e.FileFS("/js/wdosbox.wasm.js", "public/js/wdosbox.wasm", public) // this is required by `js-dos.js`
	return e
}

// static serves the static assets for the website such as the thumbnail and preview images.
func (c *Configuration) static(e *echo.Echo) *echo.Echo {
	const msg = "static routes"
	if e == nil {
		panic(fmt.Errorf("%s: %w", msg, panics.ErrNoEchoE))
	}
	e.Static(config.StaticThumb(), c.Environment.AbsThumbnail.String())
	e.Static(config.StaticOriginal(), c.Environment.AbsPreview.String())
	return e
}

// custom404 is a custom 404 error handler for the website,
// "The page cannot be found".
func (c *Configuration) custom404(sl *slog.Logger, e *echo.Echo) *echo.Echo {
	const msg = "custom 404 error routes"
	if e == nil {
		panic(fmt.Errorf("%s: %w", msg, panics.ErrNoEchoE))
	}
	e.GET("/:uri", func(cx *echo.Context) error {
		return app.StatusErr(cx, sl, http.StatusNotFound, cx.Param("uri"))
	})
	return e
}

// debugInfo returns detailed information about the HTTP request.
func (c *Configuration) debugInfo(e *echo.Echo) *echo.Echo {
	const msg = "debug info routes"
	if e == nil {
		panic(fmt.Errorf("%s: %w", msg, panics.ErrNoEchoE))
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
	e.GET("/debug", func(cx *echo.Context) error {
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

// api routes for the public API endpoints.
func (c *Configuration) api(sl *slog.Logger, e *echo.Echo, db *sql.DB, public embed.FS) *echo.Echo {
	const msg = "api routes"
	if err := panics.SDEP(sl, db, e, public); err != nil {
		panic(fmt.Errorf("%s: %w", msg, err))
	}
	e.FileFS("/openapi.json", "public/json/openapi.json", public)
	e.GET("/api", func(c *echo.Context) error { return app.APIInfo(sl, c) })
	// register API routes as a group to use a custom HTTP header
	apiGroup := e.Group(app.APIBase)
	apiGroup.Use(CacheMiddleware())
	apiGroup.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			const thousand = 1000.0
			start := time.Now()
			c.Response().Header().Set("X-Api-Version", app.APIVer)
			// use a custom response writer to capture the timing
			resp, err := echo.UnwrapResponse(c.Response())
			if err != nil {
				return err
			}
			resp.Before(func() {
				end := time.Since(start)
				ms := float64(end.Microseconds()) / thousand
				value := fmt.Sprintf("%.3fms", ms)
				resp.Header().Set("X-Response-Time", value)
			})
			return next(c)
		}
	})
	apiGroup.GET("/categories", func(c *echo.Context) error { return app.CategoriesAPI(c, db) })
	apiGroup.GET("/category/:category", func(c *echo.Context) error { return app.CategoryAPI(sl, c, db) })
	apiGroup.GET("/platforms", func(c *echo.Context) error { return app.PlatformsAPI(c, db) })
	apiGroup.GET("/platform/:platform", func(c *echo.Context) error { return app.PlatformAPI(sl, c, db) })
	apiGroup.GET("/milestones", app.MilestonesAPI)
	apiGroup.GET("/milestones/highlights", app.MilestoneHighlightsAPI)
	apiGroup.GET("/milestones/year/:year", app.MilestoneYearAPI)
	apiGroup.GET("/milestones/years/:range", app.MilestoneYearsAPI)
	apiGroup.GET("/milestones/decade/:decade", app.MilestoneDecadeAPI)
	apiGroup.GET("/areacodes", app.AreacodesAPI)
	apiGroup.GET("/areacode/:code", app.AreaCodeAPI)
	apiGroup.GET("/areacodes/search/:query", app.AreacodeSearchAPI)
	apiGroup.GET("/areacodes/regions", app.RegionsAPI)
	apiGroup.GET("/areacodes/region/:abbr", app.RegionAPI)
	apiGroup.GET("/websites", app.WebsitesAPI)
	apiGroup.GET("/demozoo", app.DemozooAPI)
	apiGroup.GET("/groups", func(c *echo.Context) error { return app.GroupsAPI(sl, c, db) })
	apiGroup.GET("/sites", func(c *echo.Context) error { return app.SitesAPI(sl, c, db) })
	apiGroup.GET("/boards", func(c *echo.Context) error { return app.BoardsAPI(sl, c, db) })
	apiGroup.GET("/magazines", func(c *echo.Context) error { return app.MagazinesAPI(sl, c, db) })
	apiGroup.GET("/releaser/:name", func(c *echo.Context) error { return app.ReleaserAPI(sl, c, db) })
	apiGroup.GET("/artifacts", func(c *echo.Context) error { return app.ArtifactsAPI(sl, c, db) })
	apiGroup.GET("/artifacts/new", func(c *echo.Context) error { return app.ArtifactsNewAPI(sl, c, db) })
	apiGroup.GET("/artifact/:id", func(c *echo.Context) error { return app.FileAPI(sl, c, db) })
	apiGroup.GET("/sceners", func(c *echo.Context) error { return app.ScenersAPI(sl, c, db) })
	apiGroup.GET("/sceners/artist", func(c *echo.Context) error { return app.ArtistsAPI(sl, c, db) })
	apiGroup.GET("/sceners/coder", func(c *echo.Context) error { return app.CodersAPI(sl, c, db) })
	apiGroup.GET("/sceners/musician", func(c *echo.Context) error { return app.MusiciansAPI(sl, c, db) })
	apiGroup.GET("/sceners/writer", func(c *echo.Context) error { return app.WritersAPI(sl, c, db) })
	apiGroup.GET("/scener/:name", func(c *echo.Context) error { return app.ScenerAPI(sl, c, db) })

	return e
}

// website routes for the main site.
func (c *Configuration) website(sl *slog.Logger, e *echo.Echo, db *sql.DB, dirs app.Dirs) *echo.Echo { //nolint:funlen
	const msg = "website routes"
	if err := panics.SDE(sl, db, e); err != nil {
		panic(fmt.Errorf("%s: %w", msg, err))
	}
	e.GET("/health-check", func(c *echo.Context) error {
		return c.NoContent(http.StatusOK)
	})
	e.GET("/sitemaps.xml", func(c *echo.Context) error {
		i := sitemap.MapIndex()
		return c.XMLPretty(http.StatusOK, i, "  ")
	})
	e.GET("/"+sitemap.Website, func(c *echo.Context) error {
		i := sitemap.MapSite(db, sl)
		return c.XMLPretty(http.StatusOK, i, "  ")
	})
	e.GET("/"+sitemap.Releaser, func(c *echo.Context) error {
		i := sitemap.MapReleaser(db, sl)
		return c.XMLPretty(http.StatusOK, i, "  ")
	})
	e.GET("/"+sitemap.Magazine, func(c *echo.Context) error {
		i := sitemap.MapMagazine(db, sl)
		return c.XMLPretty(http.StatusOK, i, "  ")
	})
	e.GET("/"+sitemap.BBS, func(c *echo.Context) error {
		i := sitemap.MapBBS(db, sl)
		return c.XMLPretty(http.StatusOK, i, "  ")
	})
	e.GET("/"+sitemap.FTP, func(c *echo.Context) error {
		i := sitemap.MapFTP(db, sl)
		return c.XMLPretty(http.StatusOK, i, "  ")
	})
	s := e.Group("")
	s.GET("/", func(c *echo.Context) error { return app.Index(sl, c) })
	s.GET("/apps", func(c *echo.Context) error { return app.Apps(sl, c) })
	s.GET("/areacodes", func(c *echo.Context) error { return app.Areacodes(sl, c) })
	s.GET("/artist", func(c *echo.Context) error {
		return app.Artist(sl, c, db)
	})
	s.GET("/bbs", func(c *echo.Context) error {
		return app.BBS(sl, c, db)
	})
	s.GET("/bbs/a-z", func(c *echo.Context) error {
		return app.BBSAZ(sl, c, db)
	})
	s.GET("/bbs/year", func(c *echo.Context) error {
		return app.BBSYear(sl, c, db)
	})
	s.GET("/brokentexts", func(c *echo.Context) error { return app.BrokenTexts(sl, c) })
	s.GET("/coder", func(c *echo.Context) error {
		return app.Coder(sl, c, db)
	})
	s.GET("/compression", func(c *echo.Context) error { return app.Compression(sl, c) })
	s.GET(Downloader, func(cx *echo.Context) error {
		return app.Download(sl, cx, db, dir.Directory(c.Environment.AbsDownload))
	})
	s.GET("/f/:id", func(cx *echo.Context) error {
		uri := cx.Param("id")
		if qs := cx.QueryString(); qs != "" {
			return cx.Redirect(http.StatusMovedPermanently, "/f/"+uri)
		}
		dirs.URI = uri
		return dirs.Artifact(sl, cx, db, bool(c.Environment.ReadOnly))
	})
	s.GET("/file/stats", func(cx *echo.Context) error {
		return app.Categories(sl, cx, db, true)
	})
	s.GET("/files/:id/:page", func(cx *echo.Context) error {
		switch cx.Param("id") {
		case "for-approval", "deletions", "unwanted":
			return app.StatusErr(cx, sl, http.StatusNotFound, cx.Param("id"))
		}
		return app.Artifacts(sl, cx, db, cx.Param("id"), cx.Param("page"))
	})
	s.GET("/files/:id", func(cx *echo.Context) error {
		switch cx.Param("id") {
		case "for-approval", "deletions", "unwanted":
			return app.StatusErr(cx, sl, http.StatusNotFound, cx.Param("id"))
		}
		return app.Artifacts(sl, cx, db, cx.Param("id"), "1")
	})
	s.GET("/file", func(cx *echo.Context) error {
		return app.Categories(sl, cx, db, false)
	})
	s.GET("/fixes", func(c *echo.Context) error { return app.Fixes(sl, c) })
	s.GET("/ftp", func(c *echo.Context) error {
		return app.FTP(sl, c, db)
	})
	s.GET("/g/:id", func(cx *echo.Context) error {
		if qs := cx.QueryString(); qs != "" {
			return cx.Redirect(http.StatusMovedPermanently, "/g/"+cx.Param("id"))
		}
		return app.Releasers(sl, cx, db, cx.Param("id"), c.Public)
	})
	s.GET("/history", func(c *echo.Context) error { return app.History(sl, c) })
	s.GET("/interview", func(c *echo.Context) error { return app.Interview(sl, c) })
	s.GET("/jsdos/:id", func(cx *echo.Context) error {
		return app.DownloadJsDos(sl, cx, db,
			dir.Directory(c.Environment.AbsExtra),
			dir.Directory(c.Environment.AbsDownload))
	})
	s.GET("/magazine", func(c *echo.Context) error {
		return app.Magazine(sl, c, db)
	})
	s.GET("/magazine/a-z", func(c *echo.Context) error {
		return app.MagazineAZ(sl, c, db)
	})
	s.GET("/new", func(c *echo.Context) error { return app.New(sl, c) })
	s.GET("/musician", func(c *echo.Context) error {
		return app.Musician(sl, c, db)
	})
	s.GET("/p/:id", func(cx *echo.Context) error {
		if qs := cx.QueryString(); qs != "" {
			return cx.Redirect(http.StatusMovedPermanently, "/p/"+cx.Param("id"))
		}
		return app.Sceners(sl, cx, db, cx.Param("id"))
	})
	s.GET("/pouet/vote/:id", func(cx *echo.Context) error {
		return app.VotePouet(sl, cx, cx.Param("id"))
	})
	s.GET("/pouet/prod/:id", func(cx *echo.Context) error {
		return app.ProdPouet(cx, cx.Param("id"))
	})
	s.GET("/zoo/prod/:id", func(cx *echo.Context) error {
		return app.ProdZoo(cx, cx.Param("id"))
	})
	s.GET("/releaser", func(c *echo.Context) error {
		return app.Releaser(sl, c, db)
	})
	s.GET("/releaser/a-z", func(c *echo.Context) error {
		return app.ReleaserAZ(sl, c, db)
	})
	s.GET("/releaser/year", func(c *echo.Context) error {
		return app.ReleaserYear(sl, c, db)
	})
	s.GET("/scener", func(c *echo.Context) error {
		return app.Scener(sl, c, db)
	})
	s.GET("/sum/:id", func(cx *echo.Context) error {
		return app.Checksum(sl, cx, db, cx.Param("id"))
	})
	s.GET("/terms", func(c *echo.Context) error { return app.Terms(sl, c) })
	s.GET("/thanks", func(c *echo.Context) error { return app.Thanks(sl, c) })
	s.GET("/thescene", func(c *echo.Context) error { return app.TheScene(sl, c) })
	s.GET("/titles", func(c *echo.Context) error { return app.Titles(sl, c) })
	s.GET("/website/:id", func(cx *echo.Context) error {
		return app.Website(sl, cx, cx.Param("id"))
	})
	s.GET("/website", func(cx *echo.Context) error {
		return app.Website(sl, cx, "")
	})
	s.GET("/writer", func(c *echo.Context) error {
		return app.Writer(sl, c, db)
	})
	s.GET("/v/:id", func(cx *echo.Context) error {
		return app.Inline(sl, cx, db, dir.Directory(c.Environment.AbsDownload))
	})
	return e
}

// search forms and the results for database queries.
func (c *Configuration) search(sl *slog.Logger, e *echo.Echo, db *sql.DB) *echo.Echo {
	const msg = "search routes"
	if err := panics.SDE(sl, db, e); err != nil {
		panic(fmt.Errorf("%s: %w", msg, err))
	}
	search := e.Group("/search")
	search.GET("/desc", func(c *echo.Context) error { return app.SearchDesc(sl, c) })
	search.GET("/file", func(c *echo.Context) error { return app.SearchFile(sl, c) })
	search.GET("/releaser", func(c *echo.Context) error { return app.SearchReleaser(sl, c) })
	search.GET("/result", func(c *echo.Context) error {
		// this legacy get result should be kept for (osx.xml) opensearch compatibility
		// and to keep possible backwards compatibility with third party site links.
		terms := strings.ReplaceAll(c.QueryParam("query"), "+", " ") // AND replacement
		terms = strings.ReplaceAll(terms, "|", ",")                  // OR replacement
		return app.PostDesc(sl, c, db, terms)
	})
	search.POST("/desc", func(c *echo.Context) error {
		return app.PostDesc(sl, c, db, c.FormValue("search-term-query"))
	})
	search.POST("/file", func(c *echo.Context) error {
		return app.PostFilename(sl, c, db)
	})
	search.POST("/releaser", func(cx *echo.Context) error {
		return htmx.SearchReleaser(sl, cx, db, &c.TidbitIndex)
	})
	return e
}

// signin for operators.
func (c *Configuration) signin(sl *slog.Logger, e *echo.Echo, nonce string) *echo.Echo {
	const msg = "signin routes"
	if err := panics.SE(sl, e); err != nil {
		panic(fmt.Errorf("%s: %w", msg, err))
	}
	readonlylock := func(cx echo.HandlerFunc) echo.HandlerFunc {
		return c.ReadOnlyLock(cx, sl)
	}
	signings := e.Group("")
	signings.Use(readonlylock)
	signings.GET("/signedout", func(cx *echo.Context) error {
		return app.SignedOut(sl, cx)
	})
	signings.GET("/signin", func(cx *echo.Context) error {
		return app.Signin(sl, cx, c.Environment.GoogleClientID.String(), nonce)
	})
	signings.GET("/operator/signin", func(cx *echo.Context) error {
		return cx.Redirect(http.StatusMovedPermanently, "/signin")
	})
	google := signings.Group("/google")
	google.POST("/callback", func(cx *echo.Context) error {
		return app.GoogleCallback(sl, cx,
			c.Environment.GoogleClientID.String(),
			c.Environment.SessionMaxAge.Int(),
			c.Environment.GoogleAccounts...)
	})
	return e
}

// AppendMoved redirects are partial URL routers that are to be redirected with a HTTP 301 Moved Permanently.
func AppendMoved(e *echo.Echo) *echo.Echo {
	const msg = "moved permanently routes"
	if e == nil {
		panic(fmt.Errorf("%s: %w", msg, panics.ErrNoEchoE))
	}
	e = nginx(e)
	e = fixes(e)
	return e
}

// nginx redirects.
func nginx(e *echo.Echo) *echo.Echo {
	const msg = "nginx redirects"
	if e == nil {
		panic(fmt.Errorf("%s: %w", msg, panics.ErrNoEchoE))
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
	const msg = "fixes routers"
	if e == nil {
		panic(fmt.Errorf("%s: %w", msg, panics.ErrNoEchoE))
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
	// THESE ARE NOT WORKING, public-enemy/ and the-dream-team/ get redirected
	// fixes.GET(`/public-enemy*tristar-ampersand-red-sector-inc*the-dream-team`, func(c *echo.Context) error {
	// 	return c.Redirect(code, g+"pe*trsi*tdt")
	// })
	// fixes.GET(`/the-dream-team*tristar-ampersand-red-sector-inc`, func(c *echo.Context) error {
	// 	return c.Redirect(code, g+"coop")
	// })
	return e
}
