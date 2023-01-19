// Package router provides all the functions for the Echo web framework.
// https://echo.labstack.com
package router

import (
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

// Router is the main instance of the Echo router.
func Route(configs config.Config, log *zap.SugaredLogger) *echo.Echo {

	e := echo.New()
	e.HideBanner = true

	// HTML templates
	e.Renderer = &TemplateRegistry{
		Templates: TmplHTML3(),
	}

	// Static images
	e.File("favicon.ico", "public/images/favicon.ico")
	e.Static("/images", "public/images")

	// Middleware
	// remove trailing slashes
	e.Use(middleware.RemoveTrailingSlashWithConfig(middleware.TrailingSlashConfig{
		RedirectCode: http.StatusMovedPermanently,
	}))
	// www. redirect
	e.Pre(middleware.NonWWWRedirect())
	// timeout
	e.Use(middleware.TimeoutWithConfig(middleware.TimeoutConfig{
		Timeout: time.Duration(configs.Timeout) * time.Second,
	}))
	if configs.IsProduction {
		// recover from panics
		e.Use(middleware.Recover())
		// https redirect
		// e.Pre(middleware.HTTPSRedirect())
		// e.Pre(middleware.HTTPSNonWWWRedirect())
	}

	// HTTP status logger
	e.Use(configs.LoggerMiddleware)

	// Custom response headers
	if configs.NoRobots {
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
	html3.Routes(e, log)

	// Routes => /users
	e.GET("/users", users.GetAllUsers)
	e.POST("/users", users.CreateUser)
	e.GET("/users/:id", users.GetUser)
	e.PUT("/users/:id", users.UpdateUser)
	e.DELETE("/users/:id", users.DeleteUser)

	// Router => HTTP error handler
	e.HTTPErrorHandler = configs.CustomErrorHandler

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
