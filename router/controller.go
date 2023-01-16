package router

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Defacto2/server/config"
	"github.com/Defacto2/server/router/html3"
	"github.com/Defacto2/server/router/users"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
)

func Route(configs config.Config, log *zap.SugaredLogger) *echo.Echo {

	e := echo.New()
	e.HideBanner = true

	// Custom error pages
	// NOTE: this does not work with middleware loggers
	// as they will always render a 200 code
	//
	// e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
	// 	return func(c echo.Context) error {
	// 		// Extract the credentials from HTTP request header and perform a security
	// 		// check

	// 		// For invalid credentials
	// 		return echo.NewHTTPError(http.StatusUnauthorized, "Please provide valid credentials")

	// 		// For valid credentials call next
	// 		// return next(c)
	// 	}
	// })

	// HTML templates
	e.Renderer = &TemplateRegistry{
		Templates: TmplHTML3(),
	}

	// Static images
	e.File("favicon.ico", "public/images/favicon.ico")
	e.Static("/images", "public/images")

	// Middleware

	e.Use(middleware.RemoveTrailingSlashWithConfig(middleware.TrailingSlashConfig{
		RedirectCode: http.StatusMovedPermanently,
	}))
	e.Use(middleware.Recover())
	e.Use(middleware.TimeoutWithConfig(middleware.TimeoutConfig{
		Timeout: time.Duration(configs.Timeout) * time.Second,
	}))

	// HTTP status logger
	e.Use(configs.LoggerMiddleware)

	// Response headers
	if configs.NoRobots {
		e.Use(NoRobotsHeader)
	}

	// Route => handler
	e.GET("/", func(c echo.Context) error {
		const count = 999
		return c.String(http.StatusOK, fmt.Sprintf("Hello, World!\nThere are %d files\n",
			count))
	})

	// Routes => html3
	html3.Routes(html3.Root, e)

	// Routes => users
	e.GET("/users", users.GetAllUsers)
	e.POST("/users", users.CreateUser)
	e.GET("/users/:id", users.GetUser)
	e.PUT("/users/:id", users.UpdateUser)
	e.DELETE("/users/:id", users.DeleteUser)

	// Router => downloads
	e.GET("/d/:id", Download)

	e.HTTPErrorHandler = configs.CustomErrorHandler

	return e
}

// NoRobotsHeader middleware adds a `X-Robots-Tag` with noindex
// and nofollow header to the response.
// https://developers.google.com/search/docs/crawling-indexing/robots-meta-tag#xrobotstag
func NoRobotsHeader(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		const HeaderXRobotsTag = "X-Robots-Tag"
		c.Response().Header().Set(HeaderXRobotsTag, "noindex, nofollow")
		return next(c)
	}
}
