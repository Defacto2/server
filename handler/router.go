package handler

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
func Routes(e *echo.Echo, log *zap.SugaredLogger, public embed.FS) *echo.Echo {
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

	// TODO order alphabetically

	e.GET("/", func(c echo.Context) error {
		return app.Index(nil, c)
	})
	e.GET("/artist", func(c echo.Context) error {
		return app.Artist(nil, c)
	})
	e.GET("/bbs", func(c echo.Context) error {
		return app.BBS(nil, c)
	})
	e.GET("/coder", func(c echo.Context) error {
		return app.Coder(nil, c)
	})
	e.GET("/ftp", func(c echo.Context) error {
		return app.FTP(nil, c)
	})
	e.GET("/history", func(c echo.Context) error {
		return app.History(nil, c)
	})
	e.GET("/interview", func(c echo.Context) error {
		return app.Interview(nil, c)
	})
	e.GET("/musician", func(c echo.Context) error {
		return app.Musician(nil, c)
	})
	e.GET("/thanks", func(c echo.Context) error {
		return app.Thanks(nil, c)
	})
	e.GET("/scener", func(c echo.Context) error {
		return app.Scener(nil, c)
	})
	e.GET("/thescene", func(c echo.Context) error {
		return app.TheScene(nil, c)
	})
	// TODO: rename to singular
	e.GET("/website", func(c echo.Context) error {
		return app.Website(nil, c, "")
	})
	e.GET("/website/:id", func(c echo.Context) error {
		return app.Website(nil, c, c.Param("id"))
	})
	e.GET("/writer", func(c echo.Context) error {
		return app.Writer(nil, c)
	})
	e.GET("/file/stats", func(c echo.Context) error {
		return app.File(nil, c, true)
	})
	e.GET("/files/:id", func(c echo.Context) error {
		return app.Files(nil, c, c.Param("id"))
	})
	e.GET("/file", func(c echo.Context) error {
		return app.File(nil, c, false)
	})
	e.GET("/magazine", func(c echo.Context) error {
		return app.Magazine(nil, c)
	})
	e.GET("/releaser", func(c echo.Context) error {
		return app.Releaser(nil, c)
	})

	// all other page requests return a custom 404 error page
	e.GET("/:uri", func(c echo.Context) error {
		return app.Status(nil, c, http.StatusNotFound, c.Param("uri"))
	})

	return e
}
