package router

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/bengarrett/df2023/config"
	"github.com/bengarrett/df2023/router/html3"
	"github.com/bengarrett/df2023/router/users"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func Route(cfg config.Config) *echo.Echo {

	e := echo.New()
	e.HideBanner = true

	// Custom error pages
	e.HTTPErrorHandler = CustomErrorHandler

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
	if cfg.LogRequests {
		e.Use(middleware.Logger())
	}
	e.Use(middleware.Recover())
	e.Use(middleware.TimeoutWithConfig(middleware.TimeoutConfig{
		Timeout: config.Timeout,
	}))

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

	return e
}

func CustomErrorHandler(err error, c echo.Context) {
	splitPaths := func(r rune) bool {
		return r == '/'
	}
	rel := strings.FieldsFunc(c.Path(), splitPaths)
	html3Route := len(rel) > 0 && rel[0] == "html3"
	if html3Route {
		if err := html3.Error(err, c); err != nil {
			panic(err) // TODO: logger?
		}
		return
	}
	code := http.StatusInternalServerError
	msg := "internal server error"
	if he, ok := err.(*echo.HTTPError); ok {
		code = he.Code
		msg = fmt.Sprint(he.Message)
	}
	c.Logger().Error(err)
	c.String(code, fmt.Sprintf("%d - %s", code, msg))
	// errorPage := fmt.Sprintf("%d.html", code)
	// if err := c.File(errorPage); err != nil {
	// 	c.Logger().Error(err)
	// }
}
