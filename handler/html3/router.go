package html3

import (
	"fmt"
	"net/http"

	"github.com/Defacto2/server/internal/tags"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// Routes for the /html3 sub-route group.
// Any errors are logged and rendered to the client using HTTP codes
// and the custom /html3, group errror template.
func Routes(e *echo.Echo, logger *zap.SugaredLogger) *echo.Group {
	if e == nil {
		panic(ErrRoutes)
	}
	s := Sugared{Log: logger}
	g := e.Group(Prefix)
	g.GET("", s.Index)
	g.GET("/all:offset", s.All)
	g.GET("/all", s.All)
	g.GET("/categories", s.Categories)
	g.GET("/platforms", s.Platforms)
	g = getTags(s, g)
	g.GET("/groups:offset", s.Groups)
	g.GET("/groups", s.Groups)
	g.GET("/group/:id", s.Group)
	g.GET("/art:offset", s.Art)
	g.GET("/art", s.Art)
	g.GET("/documents:offset", s.Documents)
	g.GET("/documents", s.Documents)
	g.GET("/software:offset", s.Software)
	g.GET("/software", s.Software)
	g = moved(g)
	return custom404(g)
}

// getTags creates the get routes for the category and platform tags.
func getTags(s Sugared, g *echo.Group) *echo.Group {
	category := g.Group("/category")
	for _, tag := range tags.List() {
		if tags.IsCategory(tag.String()) {
			category.GET(fmt.Sprintf("/%s:offset", tag), s.Category)
			category.GET(fmt.Sprintf("/%s", tag), s.Category)
		}
	}
	platform := g.Group("/platform")
	for _, tag := range tags.List() {
		if tags.IsPlatform(tag.String()) {
			platform.GET(fmt.Sprintf("/%s:offset", tag), s.Platform)
			platform.GET(fmt.Sprintf("/%s", tag), s.Platform)
		}
	}
	return g
}

// custom404 is a custom 404 error handler for the website, "The page cannot be found."
func custom404(g *echo.Group) *echo.Group {
	g.GET("/:uri", func(x echo.Context) error {
		return echo.NewHTTPError(http.StatusNotFound,
			"The page cannot be found: /html3/"+x.Param("uri"))
	})
	return g
}

// moved handles the moved permanently redirects.
func moved(g *echo.Group) *echo.Group {
	const code = http.StatusMovedPermanently
	redirect := g.Group("")
	redirect.GET("/index", func(c echo.Context) error {
		return c.Redirect(code, "/html3")
	})
	redirect.GET("/categories/index", func(c echo.Context) error {
		return c.Redirect(code, "/html3/categories")
	})
	redirect.GET("/platforms/index", func(c echo.Context) error {
		return c.Redirect(code, "/html3/platforms")
	})
	return g
}
