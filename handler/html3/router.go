package html3

// Package file router.go contains the HTML3 website route functions.

import (
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"slices"

	"github.com/Defacto2/server/internal/panics"
	"github.com/Defacto2/server/internal/tags"
	"github.com/labstack/echo/v4"
)

// Routes for the /html3 sub-route group.
// Any errors are logged and rendered to the client using HTTP codes
// and the custom /html3, group error template.
func Routes(e *echo.Echo, db *sql.DB, sl *slog.Logger) *echo.Group {
	const msg = "htm3 routes"
	if err := panics.EchoDS(e, db, sl); err != nil {
		panic(fmt.Errorf("%s: %w", msg, err))
	}
	g := e.Group(Prefix)
	g.GET("", func(c echo.Context) error {
		return Index(c, db, sl)
	})
	g.GET("/all:offset", func(c echo.Context) error {
		return All(c, db, sl)
	})
	g.GET("/all", func(c echo.Context) error {
		return All(c, db, sl)
	})
	g.GET("/categories", func(c echo.Context) error {
		return Categories(c, sl)
	})
	g.GET("/platforms", func(c echo.Context) error {
		return Platforms(c, sl)
	})
	g = getTags(db, sl, g)
	g.GET("/groups:offset", func(c echo.Context) error {
		return Groups(c, db, sl)
	})
	g.GET("/groups", func(c echo.Context) error {
		return Groups(c, db, sl)
	})
	g.GET("/group/:id", func(c echo.Context) error {
		return Group(c, db, sl)
	})
	g.GET("/art:offset", func(c echo.Context) error {
		return Art(c, db, sl)
	})
	g.GET("/art", func(c echo.Context) error {
		return Art(c, db, sl)
	})
	g.GET("/documents:offset", func(c echo.Context) error {
		return Documents(c, db, sl)
	})
	g.GET("/documents", func(c echo.Context) error {
		return Documents(c, db, sl)
	})
	g.GET("/software:offset", func(c echo.Context) error {
		return Software(c, db, sl)
	})
	g.GET("/software", func(c echo.Context) error {
		return Software(c, db, sl)
	})
	g = moved(g)
	return custom404(g)
}

// getTags creates the get routes for the category and platform tags.
func getTags(db *sql.DB, sl *slog.Logger, g *echo.Group) *echo.Group {
	category := g.Group("/category")
	for tag := range slices.Values(tags.List()) {
		if tags.IsCategory(tag.String()) {
			category.GET(fmt.Sprintf("/%s:offset", tag), func(c echo.Context) error {
				return Category(c, db, sl)
			})
			category.GET(fmt.Sprintf("/%s", tag), func(c echo.Context) error {
				return Category(c, db, sl)
			})
		}
	}
	platform := g.Group("/platform")
	for tag := range slices.Values(tags.List()) {
		if tags.IsPlatform(tag.String()) {
			platform.GET(fmt.Sprintf("/%s:offset", tag), func(c echo.Context) error {
				return Platform(c, db, sl)
			})
			platform.GET(fmt.Sprintf("/%s", tag), func(c echo.Context) error {
				return Platform(c, db, sl)
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
