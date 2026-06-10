package html3

// Package file router.go contains the HTML3 website route functions.

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"slices"

	"github.com/Defacto2/server/internal/panics"
	"github.com/Defacto2/server/internal/tags"
	"github.com/labstack/echo/v5"
)

// Routes for the /html3 sub-route group.
// Any errors are logged and rendered to the client using HTTP codes
// and the custom /html3, group error template.
func Routes(ctx context.Context, sl *slog.Logger, e *echo.Echo, db *sql.DB) *echo.Group {
	const msg = "htm3 routes"
	if err := panics.SDE(sl, db, e); err != nil {
		panic(fmt.Errorf("%s: %w", msg, err))
	}
	g := e.Group(Prefix)
	g.GET("", func(c *echo.Context) error {
		return Index(ctx, sl, c, db)
	})
	g.GET("/all:offset", func(c *echo.Context) error {
		return All(ctx, sl, c, db)
	})
	g.GET("/all", func(c *echo.Context) error {
		return All(ctx, sl, c, db)
	})
	g.GET("/categories", func(c *echo.Context) error {
		return Categories(sl, c)
	})
	g.GET("/platforms", func(c *echo.Context) error {
		return Platforms(sl, c)
	})
	g = getTags(ctx, sl, db, g)
	g.GET("/groups:offset", func(c *echo.Context) error {
		return Groups(ctx, sl, c, db)
	})
	g.GET("/groups", func(c *echo.Context) error {
		return Groups(ctx, sl, c, db)
	})
	g.GET("/group/:id", func(c *echo.Context) error {
		return Group(ctx, sl, c, db)
	})
	g.GET("/art:offset", func(c *echo.Context) error {
		return Art(ctx, sl, c, db)
	})
	g.GET("/art", func(c *echo.Context) error {
		return Art(ctx, sl, c, db)
	})
	g.GET("/documents:offset", func(c *echo.Context) error {
		return Documents(ctx, sl, c, db)
	})
	g.GET("/documents", func(c *echo.Context) error {
		return Documents(ctx, sl, c, db)
	})
	g.GET("/software:offset", func(c *echo.Context) error {
		return Software(ctx, c, db, sl)
	})
	g.GET("/software", func(c *echo.Context) error {
		return Software(ctx, c, db, sl)
	})
	g = moved(g)
	return custom404(g)
}

// getTags creates the get routes for the category and platform tags.
func getTags(ctx context.Context, sl *slog.Logger, db *sql.DB, g *echo.Group) *echo.Group {
	category := g.Group("/category")
	hCategory := func(c *echo.Context) error {
		return Category(ctx, sl, c, db)
	}
	for tag := range slices.Values(tags.List()) {
		if tags.IsCategory(tag.String()) {
			category.GET(fmt.Sprintf("/%s:offset", tag), hCategory)
			category.GET(fmt.Sprintf("/%s", tag), hCategory)
		}
	}
	platform := g.Group("/platform")
	hPlatform := func(c *echo.Context) error {
		return Platform(ctx, sl, c, db)
	}
	for tag := range slices.Values(tags.List()) {
		if tags.IsPlatform(tag.String()) {
			platform.GET(fmt.Sprintf("/%s:offset", tag), hPlatform)
			platform.GET(fmt.Sprintf("/%s", tag), hPlatform)
		}
	}
	return g
}

// custom404 is a custom 404 error handler for the website,
// "The page cannot be found.".
func custom404(g *echo.Group) *echo.Group {
	g.GET("/:uri", func(x *echo.Context) error {
		s := "The page cannot be found: /html3/" + x.Param("uri")
		return echo.NewHTTPError(http.StatusNotFound, s)
	})
	return g
}

// moved handles the moved permanently redirects.
func moved(g *echo.Group) *echo.Group {
	const code = http.StatusMovedPermanently
	redirect := g.Group("")
	redirect.GET("/index", func(c *echo.Context) error {
		return c.Redirect(code, "/html3")
	})
	redirect.GET("/categories/index", func(c *echo.Context) error {
		return c.Redirect(code, "/html3/categories")
	})
	redirect.GET("/platforms/index", func(c *echo.Context) error {
		return c.Redirect(code, "/html3/platforms")
	})
	return g
}
