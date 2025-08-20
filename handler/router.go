package handler

// Package file router.go contains the custom router URIs for the website.

import (
	"database/sql"
	"embed"
	"encoding/xml"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/Defacto2/helper"
	"github.com/Defacto2/releaser"
	"github.com/Defacto2/server/handler/app"
	"github.com/Defacto2/server/handler/htmx"
	"github.com/Defacto2/server/handler/sitemap"
	"github.com/Defacto2/server/internal/config"
	"github.com/Defacto2/server/internal/dir"
	"github.com/Defacto2/server/internal/panics"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
)

const code = http.StatusMovedPermanently

// FilesRoutes defines the file locations and routes for the web server.
func (c *Configuration) FilesRoutes(e *echo.Echo, db *sql.DB, sl *slog.Logger, public embed.FS,
) (*echo.Echo, error) {
	const msg = "files routes"
	if err := panics.EchoDSP(e, db, sl, public); err != nil {
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
	}
	nonce, err := c.nonce(e)
	if err != nil {
		return nil, fmt.Errorf("%s nonce session key: %w", msg, err)
	}
	e = c.signin(e, sl, nonce)
	e = c.custom404(e, sl)
	e = c.debugInfo(e)
	e = c.static(e)
	e = c.html(e, public)
	e = c.font(e, public)
	e = c.embed(e, public)
	e = c.search(e, db, sl)
	e = c.website(e, db, sl, dirs)
	e = c.lock(e, db, sl, dirs)
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
	if err := panics.EchoP(e, public); err != nil {
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
	if err := panics.EchoP(e, public); err != nil {
		panic(fmt.Errorf("%s: %w", msg, err))
	}
	e.FileFS("/favicon.ico", "public/image/favicon.ico", public)
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
func (c *Configuration) custom404(e *echo.Echo, sl *slog.Logger) *echo.Echo {
	const msg = "custom 404 error routes"
	if e == nil {
		panic(fmt.Errorf("%s: %w", msg, panics.ErrNoEchoE))
	}
	e.GET("/:uri", func(cx echo.Context) error {
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

type SitemapIndex struct {
	XMLName  xml.Name `xml:"sitemapindex"`
	XMLNS    string   `xml:"xmlns,attr"`
	Sitemaps []Locations
}

type Locations struct {
	XMLName  xml.Name `xml:"sitemap"`
	Location string   `xml:"loc"`
	LastMod  string   `xml:"lastmod"`
}

type Sitemap struct {
	XMLName xml.Name `xml:"urlset"`
	XMLNS   string   `xml:"xmlns,attr"`
	Urls    []SiteUrl
}

type SiteUrl struct {
	XMLName  xml.Name `xml:"urlset"`
	Location string   `xml:"loc"`
	LastMod  string   `xml:"lastmod"`
}

// <urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
//   <url>
//     <loc>https://www.example.com/foo.html</loc>
//     <lastmod>2022-06-04</lastmod>
//   </url>
// </urlset>

// website routes for the main site.
func (c *Configuration) website(e *echo.Echo, db *sql.DB, sl *slog.Logger, dirs app.Dirs) *echo.Echo {
	const msg = "website routes"
	if err := panics.EchoDS(e, db, sl); err != nil {
		panic(fmt.Errorf("%s: %w", msg, err))
	}
	e.GET("/health-check", func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})
	e.GET("/sitemap_index.xml", func(c echo.Context) error {
		i := sitemap.MapIndex()
		return c.XMLPretty(http.StatusOK, i, "  ")
	})
	e.GET("/sitemap.xml", func(c echo.Context) error {
		i := sitemap.MapSite(db)
		return c.XMLPretty(http.StatusOK, i, "  ")
	})
	s := e.Group("")
	s.GET("/", func(c echo.Context) error { return app.Index(c, sl) })
	s.GET("/areacodes", func(c echo.Context) error { return app.Areacodes(c, sl) })
	s.GET("/artist", func(c echo.Context) error {
		return app.Artist(c, db, sl)
	})
	s.GET("/bbs", func(c echo.Context) error {
		return app.BBS(c, db, sl)
	})
	s.GET("/bbs/a-z", func(c echo.Context) error {
		return app.BBSAZ(c, db, sl)
	})
	s.GET("/bbs/year", func(c echo.Context) error {
		return app.BBSYear(c, db, sl)
	})
	s.GET("/brokentexts", func(c echo.Context) error { return app.BrokenTexts(c, sl) })
	s.GET("/coder", func(c echo.Context) error {
		return app.Coder(c, db, sl)
	})
	s.GET(Downloader, func(cx echo.Context) error {
		return app.Download(cx, db, sl, dir.Directory(c.Environment.AbsDownload))
	})
	s.GET("/f/:id", func(cx echo.Context) error {
		uri := cx.Param("id")
		if qs := cx.QueryString(); qs != "" {
			return cx.Redirect(http.StatusMovedPermanently, "/f/"+uri)
		}
		dirs.URI = uri
		return dirs.Artifact(cx, db, sl, bool(c.Environment.ReadOnly))
	})
	s.GET("/file/stats", func(cx echo.Context) error {
		return app.Categories(cx, db, sl, true)
	})
	s.GET("/files/:id/:page", func(cx echo.Context) error {
		switch cx.Param("id") {
		case "for-approval", "deletions", "unwanted":
			return app.StatusErr(cx, sl, http.StatusNotFound, cx.Param("uri"))
		}
		return app.Artifacts(cx, db, sl, cx.Param("id"), cx.Param("page"))
	})
	s.GET("/files/:id", func(cx echo.Context) error {
		switch cx.Param("id") {
		case "for-approval", "deletions", "unwanted":
			return app.StatusErr(cx, sl, http.StatusNotFound, cx.Param("uri"))
		}
		return app.Artifacts(cx, db, sl, cx.Param("id"), "1")
	})
	s.GET("/file", func(cx echo.Context) error {
		return app.Categories(cx, db, sl, false)
	})
	s.GET("/ftp", func(c echo.Context) error {
		return app.FTP(c, db, sl)
	})
	s.GET("/g/:id", func(cx echo.Context) error {
		if qs := cx.QueryString(); qs != "" {
			return cx.Redirect(http.StatusMovedPermanently, "/g/"+cx.Param("id"))
		}
		return app.Releasers(cx, db, sl, cx.Param("id"), c.Public)
	})
	s.GET("/history", func(c echo.Context) error { return app.History(c, sl) })
	s.GET("/interview", func(c echo.Context) error { return app.Interview(c, sl) })
	s.GET("/jsdos/:id", func(cx echo.Context) error {
		return app.DownloadJsDos(cx, db, sl,
			dir.Directory(c.Environment.AbsExtra),
			dir.Directory(c.Environment.AbsDownload))
	})
	s.GET("/magazine", func(c echo.Context) error {
		return app.Magazine(c, db, sl)
	})
	s.GET("/magazine/a-z", func(c echo.Context) error {
		return app.MagazineAZ(c, db, sl)
	})
	s.GET("/new", func(c echo.Context) error { return app.New(c, sl) })
	s.GET("/musician", func(c echo.Context) error {
		return app.Musician(c, db, sl)
	})
	s.GET("/p/:id", func(cx echo.Context) error {
		if qs := cx.QueryString(); qs != "" {
			return cx.Redirect(http.StatusMovedPermanently, "/p/"+cx.Param("id"))
		}
		return app.Sceners(cx, db, sl, cx.Param("id"))
	})
	s.GET("/pouet/vote/:id", func(cx echo.Context) error {
		return app.VotePouet(cx, sl, cx.Param("id"))
	})
	s.GET("/pouet/prod/:id", func(cx echo.Context) error {
		return app.ProdPouet(cx, cx.Param("id"))
	})
	s.GET("/zoo/prod/:id", func(cx echo.Context) error {
		return app.ProdZoo(cx, cx.Param("id"))
	})
	s.GET("/releaser", func(c echo.Context) error {
		return app.Releaser(c, db, sl)
	})
	s.GET("/releaser/a-z", func(c echo.Context) error {
		return app.ReleaserAZ(c, db, sl)
	})
	s.GET("/releaser/year", func(c echo.Context) error {
		return app.ReleaserYear(c, db, sl)
	})
	s.GET("/scener", func(c echo.Context) error {
		return app.Scener(c, db, sl)
	})
	s.GET("/sum/:id", func(cx echo.Context) error {
		return app.Checksum(cx, db, sl, cx.Param("id"))
	})
	s.GET("/thanks", func(c echo.Context) error { return app.Thanks(c, sl) })
	s.GET("/thescene", func(c echo.Context) error { return app.TheScene(c, sl) })
	s.GET("/titles", func(c echo.Context) error { return app.Titles(c, sl) })
	s.GET("/website/:id", func(cx echo.Context) error {
		return app.Website(cx, sl, cx.Param("id"))
	})
	s.GET("/website", func(cx echo.Context) error {
		return app.Website(cx, sl, "")
	})
	s.GET("/writer", func(c echo.Context) error {
		return app.Writer(c, db, sl)
	})
	s.GET("/v/:id", func(cx echo.Context) error {
		return app.Inline(cx, db, sl, dir.Directory(c.Environment.AbsDownload))
	})
	return e
}

// search forms and the results for database queries.
func (c *Configuration) search(e *echo.Echo, db *sql.DB, sl *slog.Logger) *echo.Echo {
	const msg = "search routes"
	if err := panics.EchoDS(e, db, sl); err != nil {
		panic(fmt.Errorf("%s: %w", msg, err))
	}
	search := e.Group("/search")
	search.GET("/desc", func(c echo.Context) error { return app.SearchDesc(c, sl) })
	search.GET("/file", func(c echo.Context) error { return app.SearchFile(c, sl) })
	search.GET("/releaser", func(c echo.Context) error { return app.SearchReleaser(c, sl) })
	search.GET("/result", func(cx echo.Context) error {
		// this legacy get result should be kept for (osx.xml) opensearch compatibility
		// and to keep possible backwards compatibility with third party site links.
		terms := strings.ReplaceAll(cx.QueryParam("query"), "+", " ") // AND replacement
		terms = strings.ReplaceAll(terms, "|", ",")                   // OR replacement
		return app.PostDesc(cx, db, sl, terms)
	})
	search.POST("/desc", func(cx echo.Context) error {
		return app.PostDesc(cx, db, sl, cx.FormValue("search-term-query"))
	})
	search.POST("/file", func(c echo.Context) error {
		return app.PostFilename(c, db, sl)
	})
	search.POST("/releaser", func(cx echo.Context) error {
		return htmx.SearchReleaser(cx, db, sl)
	})
	return e
}

// signin for operators.
func (c *Configuration) signin(e *echo.Echo, sl *slog.Logger, nonce string) *echo.Echo {
	const msg = "signin routes"
	if err := panics.EchoS(e, sl); err != nil {
		panic(fmt.Errorf("%s: %w", msg, err))
	}
	readonlylock := func(cx echo.HandlerFunc) echo.HandlerFunc {
		return c.ReadOnlyLock(cx, sl)
	}
	signings := e.Group("")
	signings.Use(readonlylock)
	signings.GET("/signedout", func(cx echo.Context) error {
		return app.SignedOut(cx, sl)
	})
	signings.GET("/signin", func(cx echo.Context) error {
		return app.Signin(cx, sl, c.Environment.GoogleClientID.String(), nonce)
	})
	signings.GET("/operator/signin", func(cx echo.Context) error {
		return cx.Redirect(http.StatusMovedPermanently, "/signin")
	})
	google := signings.Group("/google")
	google.POST("/callback", func(cx echo.Context) error {
		return app.GoogleCallback(cx, sl,
			c.Environment.GoogleClientID.String(),
			c.Environment.SessionMaxAge.Int(),
			c.Environment.GoogleAccounts...)
	})
	return e
}

// MovedPermanently redirects are partial URL routers that are to be redirected with a HTTP 301 Moved Permanently.
func MovedPermanently(e *echo.Echo) *echo.Echo {
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
	const msg = "fixes routers"
	if e == nil {
		panic(fmt.Errorf("%s: %w", msg, panics.ErrNoEchoE))
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
	fixes.GET("/rss", func(c echo.Context) error {
		return c.Redirect(code, g+"renaissance")
	})
	fixes.GET("/trsi", func(c echo.Context) error {
		return c.Redirect(code, g+releaser.Obfuscate("TRISTAR & RED SECTOR INC"))
	})
	fixes.GET("/x-pression", func(c echo.Context) error {
		return c.Redirect(code, g+releaser.Obfuscate("X-PRESSION DESIGN"))
	})
	fixes.GET("/"+releaser.Obfuscate("DAMN EXCELLENT ANSI DESIGNERS"), func(c echo.Context) error {
		return c.Redirect(code, g+releaser.Obfuscate("DAMN EXCELLENT ANSI DESIGN"))
	})
	fixes.GET("/"+releaser.Obfuscate("pirates with attitude"), func(c echo.Context) error {
		return c.Redirect(code, g+releaser.Obfuscate("pirates with attitudes"))
	})
	fixes.GET("/"+releaser.Obfuscate("TRISTAR AND RED SECTOR INC"), func(c echo.Context) error {
		return c.Redirect(code, g+releaser.Obfuscate("TRISTAR & RED SECTOR INC"))
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
	fixes.GET("/united-software-association", func(c echo.Context) error {
		return c.Redirect(code, g+"united-software-association*fairlight")
	})
	// THESE ARE NOT WORKING, public-enemy/ and the-dream-team/ get redirected
	// fixes.GET(`/public-enemy*tristar-ampersand-red-sector-inc*the-dream-team`, func(c echo.Context) error {
	// 	return c.Redirect(code, g+"pe*trsi*tdt")
	// })
	// fixes.GET(`/the-dream-team*tristar-ampersand-red-sector-inc`, func(c echo.Context) error {
	// 	return c.Redirect(code, g+"coop")
	// })
	return e
}
