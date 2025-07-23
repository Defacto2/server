package html3

// Package file router.go contains the HTML3 website route functions.

import (
	"database/sql"
	"fmt"
	"net/http"
	"slices"

	"github.com/Defacto2/server/internal/tags"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// Routes for the /html3 sub-route group.
// Any errors are logged and rendered to the client using HTTP codes
// and the custom /html3, group error template.
func Routes(e *echo.Echo, db *sql.DB, logger *zap.SugaredLogger) *echo.Group {
	if e == nil {
		panic(ErrRoutes)
	}
	s := Sugared{Log: logger}
	g := e.Group(Prefix)
	g.GET("", func(c echo.Context) error {
		return s.Index(c, db)
	})
	g.GET("/all:offset", func(c echo.Context) error {
		return s.All(c, db)
	})
	g.GET("/all", func(c echo.Context) error {
		return s.All(c, db)
	})
	g.GET("/categories", s.Categories)
	g.GET("/platforms", s.Platforms)
	g = getTags(s, db, g)
	g.GET("/groups:offset", func(c echo.Context) error {
		return s.Groups(c, db)
	})
	g.GET("/groups", func(c echo.Context) error {
		return s.Groups(c, db)
	})
	g.GET("/group/:id", func(c echo.Context) error {
		return s.Group(c, db)
	})
	g.GET("/art:offset", func(c echo.Context) error {
		return s.Art(c, db)
	})
	g.GET("/art", func(c echo.Context) error {
		return s.Art(c, db)
	})
	g.GET("/documents:offset", func(c echo.Context) error {
		return s.Documents(c, db)
	})
	g.GET("/documents", func(c echo.Context) error {
		return s.Documents(c, db)
	})
	g.GET("/software:offset", func(c echo.Context) error {
		return s.Software(c, db)
	})
	g.GET("/software", func(c echo.Context) error {
		return s.Software(c, db)
	})
	g = moved(g)
	return custom404(g)
}

// getTags creates the get routes for the category and platform tags.
func getTags(s Sugared, db *sql.DB, g *echo.Group) *echo.Group {
	category := g.Group("/category")
	for tag := range slices.Values(tags.List()) {
		if tags.IsCategory(tag.String()) {
			category.GET(fmt.Sprintf("/%s:offset", tag), func(c echo.Context) error {
				return s.Category(c, db)
			})
			category.GET(fmt.Sprintf("/%s", tag), func(c echo.Context) error {
				return s.Category(c, db)
			})
		}
	}
	platform := g.Group("/platform")
	for tag := range slices.Values(tags.List()) {
		if tags.IsPlatform(tag.String()) {
			platform.GET(fmt.Sprintf("/%s:offset", tag), func(c echo.Context) error {
				return s.Platform(c, db)
			})
			platform.GET(fmt.Sprintf("/%s", tag), func(c echo.Context) error {
				return s.Platform(c, db)
			})
		}
	}
	return g
}

// custom404 is a custom 404 error handler for the website,
// "The page cannot be found.".
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
