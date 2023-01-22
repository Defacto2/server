// Package router provides all the functions for the Echo web framework.
// https://echo.labstack.com
package router

import (
	"embed"
	"fmt"
	"net/http"
	"time"

	"github.com/Defacto2/server/pkg/config"
	"github.com/Defacto2/server/router/html3"
	"github.com/Defacto2/server/router/users"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
)

// Router configurations.
type Router struct {
	Configs config.Config      // Configs from the enviroment.
	Log     *zap.SugaredLogger // Log is a sugared logger.
	Images  embed.FS           // Not in use.
	Views   embed.FS           // Views are Go templates.
}

// Controller is the primary instance of the Echo router.
func (r Router) Controller() *echo.Echo {

	e := echo.New()
	e.HideBanner = true

	// HTML templates
	e.Renderer = &html3.TemplateRegistry{
		Templates: html3.TmplHTML3(r.Views),
	}

	// Static embedded images
	// These get distributed in the binary
	fs1 := echo.MustSubFS(r.Images, "public/images")
	e.StaticFS("/images", fs1)
	e.File("favicon.ico", "public/images/favicon.ico") // TODO: this is not being embedded

	// Middleware
	e.Use(middleware.Gzip())
	//e.Use(middleware.Decompress())
	// remove trailing slashes
	e.Use(middleware.RemoveTrailingSlashWithConfig(middleware.TrailingSlashConfig{
		RedirectCode: http.StatusMovedPermanently,
	}))
	// www. redirect
	e.Pre(middleware.NonWWWRedirect())
	// timeout
	e.Use(middleware.TimeoutWithConfig(middleware.TimeoutConfig{
		Timeout: time.Duration(r.Configs.Timeout) * time.Second,
	}))
	if r.Configs.IsProduction {
		// recover from panics
		e.Use(middleware.Recover())
		// https redirect
		// e.Pre(middleware.HTTPSRedirect())
		// e.Pre(middleware.HTTPSNonWWWRedirect())
	}

	// HTTP status logger
	e.Use(r.Configs.LoggerMiddleware)

	// Custom response headers
	if r.Configs.NoRobots {
		e.Use(NoRobotsHeader) // TODO: only apply to HTML templates?
	}

	// Route => /
	e.GET("/", func(c echo.Context) error {
		const count = 999
		return c.String(http.StatusOK, fmt.Sprintf("Hello, World!\nThere are %d files\n",
			count))
	})
	e.GET("/file/list", func(c echo.Context) error {
		return c.String(http.StatusOK, "Coming soon!")
	})

	// Routes => /html3
	html3.Routes(e, r.Log)

	// Routes => /users
	e.GET("/users", users.GetAllUsers)
	e.POST("/users", users.CreateUser)
	e.GET("/users/:id", users.GetUser)
	e.PUT("/users/:id", users.UpdateUser)
	e.DELETE("/users/:id", users.DeleteUser)

	// Router => HTTP error handler
	e.HTTPErrorHandler = r.Configs.CustomErrorHandler

	return e
}

// NoRobotsHeader middleware adds a `X-Robots-Tag` header to the response.
// The header contains the noindex and nofollow values that tell search engine
// crawlers to not index or crawl the page or asset.
// See https://developers.google.com/search/docs/crawling-indexing/robots-meta-tag#xrobotstag
func NoRobotsHeader(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		const HeaderXRobotsTag = "X-Robots-Tag"
		c.Response().Header().Set(HeaderXRobotsTag, "noindex, nofollow")
		return next(c)
	}
}
