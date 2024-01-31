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
func Routes(z *zap.SugaredLogger, e *echo.Echo) *echo.Group {
	s := Sugared{Log: z}
	g := e.Group(Prefix)
	g.GET("", s.Index)
	g.GET("/all:offset", s.All)
	g.GET("/all", s.All)
	// dynamic routes for the category tags
	// example: "/category/announcements/:offset"
	g.GET("/categories", s.Categories)
	category := g.Group("/category")
	for _, tag := range tags.List() {
		if tags.IsCategory(tag.String()) {
			category.GET(fmt.Sprintf("/%s:offset", tag), s.Category)
			category.GET(fmt.Sprintf("/%s", tag), s.Category)
		}
	}
	// dynamic routes for the platform tags
	// example: "/platform/dos/:offset"
	g.GET("/platforms", s.Platforms)
	platform := g.Group("/platform")
	for _, tag := range tags.List() {
		if tags.IsPlatform(tag.String()) {
			platform.GET(fmt.Sprintf("/%s:offset", tag), s.Platform)
			platform.GET(fmt.Sprintf("/%s", tag), s.Platform)
		}
	}
	g.GET("/groups:offset", s.Groups)
	g.GET("/groups", s.Groups)
	g.GET("/group/:id", s.Group)
	g.GET("/art:offset", s.Art)
	g.GET("/art", s.Art)
	g.GET("/documents:offset", s.Documents)
	g.GET("/documents", s.Documents)
	g.GET("/software:offset", s.Software)
	g.GET("/software", s.Software)

	// append legacy redirects
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
	// Custom 404 error, "The page cannot be found"
	g.GET("/:uri", func(x echo.Context) error {
		return echo.NewHTTPError(http.StatusNotFound,
			fmt.Sprintf("The page cannot be found: /html3/%s", x.Param("uri")))
	})
	return g
}
