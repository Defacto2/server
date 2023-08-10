package handler

// Package file router.go contains the custom router URIs for the website.

import (
	"embed"
	"net/http"

	"github.com/Defacto2/server/handler/app"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// MovedPermanently redirects URL paths with a HTTP 301 Moved Permanently.
func MovedPermanently(e *echo.Echo) {
	for path, redirect := range Redirects() {
		e.GET(path, func(c echo.Context) error {
			return c.Redirect(http.StatusMovedPermanently, redirect)
		})
	}
}

// Redirects are partial URL routers that are to be redirected with a HTTP 301 Moved Permanently.
func Redirects() map[string]string {
	return map[string]string{
		"/defacto2/history":            "/history",
		"/defacto2/subculture":         "/thescene",
		"/file/index":                  "",
		"/files/json/site.webmanifest": "/site.webmanifest",
		"/link/list/:id":               "/websites",
	}
}

// Routes defines the routes for the web server.
func Routes(z *zap.SugaredLogger, e *echo.Echo, public embed.FS) (*echo.Echo, error) {
	if e == nil {
		return nil, ErrRoutes
	}
	if z == nil {
		return nil, ErrLog
	}
	// Redirects
	// these need to be before the routes and rewrites
	MovedPermanently(e)

	// Serve embedded CSS files
	e.FileFS("/css/bootstrap.min.css", "public/css/bootstrap.min.css", public)
	e.FileFS("/css/bootstrap.min.css.map", "public/css/bootstrap.min.css.map", public)
	e.FileFS("/css/layout.min.css", "public/css/layout.min.css", public)

	// Serve embedded SVG collections
	e.FileFS("/bootstrap-icons.svg", "public/image/bootstrap-icons.svg", public)

	// Serve embedded font files
	e.FileFS("/font/pxplus_ibm_vga8.woff2", "public/font/pxplus_ibm_vga8.woff2", public)
	e.FileFS("/font/pxplus_ibm_vga8.woff", "public/font/pxplus_ibm_vga8.woff", public)
	e.FileFS("/font/pxplus_ibm_vga8.ttf", "public/font/pxplus_ibm_vga8.ttf", public)

	// Serve embedded JS files
	e.FileFS("/js/bootstrap.bundle.min.js", "public/js/bootstrap.bundle.min.js", public)
	e.FileFS("/js/bootstrap.bundle.min.js.map", "public/js/bootstrap.bundle.min.js.map", public)
	e.FileFS("/js/fontawesome.min.js", "public/js/fontawesome.min.js", public)

	// Serve embedded image files
	e.FileFS("/favicon.ico", "public/image/favicon.ico", public)

	// Serve embedded text files
	e.FileFS("/osd.xml", "public/text/osd.xml", public)
	e.FileFS("/robots.txt", "public/text/robots.txt", public)
	e.FileFS("/site.webmanifest", "public/text/site.webmanifest.json", public)

	e.GET("/", func(c echo.Context) error {
		return app.Index(z, c)
	})
	e.GET("/artist", func(c echo.Context) error {
		return app.Artist(z, c)
	})
	e.GET("/bbs", func(c echo.Context) error {
		return app.BBS(z, c)
	})
	e.GET("/coder", func(c echo.Context) error {
		return app.Coder(z, c)
	})
	e.GET("/file/stats", func(c echo.Context) error {
		return app.File(z, c, true)
	})
	e.GET("/files/:id", func(c echo.Context) error {
		return app.Files(z, c, c.Param("id"))
	})
	e.GET("/file", func(c echo.Context) error {
		return app.File(z, c, false)
	})
	e.GET("/ftp", func(c echo.Context) error {
		return app.FTP(z, c)
	})
	e.GET("/g/:id", func(c echo.Context) error {
		return app.G(z, c, c.Param("id"))
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
	e.GET("/musician", func(c echo.Context) error {
		return app.Musician(z, c)
	})
	e.GET("/releaser", func(c echo.Context) error {
		return app.Releaser(z, c)
	})
	e.GET("/scener", func(c echo.Context) error {
		return app.Scener(z, c)
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

	// all other page requests return a custom 404 error page
	e.GET("/:uri", func(c echo.Context) error {
		return app.StatusErr(z, c, http.StatusNotFound, c.Param("uri"))
	})

	return e, nil
}
