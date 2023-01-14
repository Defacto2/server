// Package html3 handles the routes and views for the retro,
// mini-website that is rendered in HTML 3 syntax.
package html3

import (
	"context"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"time"

	"github.com/bengarrett/df2023/helpers"
	"github.com/bengarrett/df2023/meta"
	"github.com/bengarrett/df2023/models"
	"github.com/bengarrett/df2023/router"
	"github.com/labstack/echo/v4"

	"github.com/bengarrett/df2023/postgres"
)

const (
	Root  = "/html3"
	title = "Index of " + Root
)

var Counts = models.Counts

var LegacyURLs = map[string]string{
	"/index":            "",
	"/categories/index": "/categories",
}

var Sortings = map[string]string{
	"Name":        "A",
	"Publish":     "A",
	"Posted":      "A",
	"Size":        "A",
	"Description": "A",
}

// 	e.GET("/users/:name", func(c echo.Context) error {
// 		name := c.Param("name")
// 		return c.String(http.StatusOK, name)
//  })
//
//  curl http://localhost:1323/users/Joe

func latency() *time.Time {
	start := time.Now()
	r := new(big.Int)
	r.Binomial(1000, 10)
	return &start
}

func Routes(prefix string, e *echo.Echo) {
	g := e.Group(prefix)
	g.GET("", Index)
	g.GET("/categories", Categories)
	g.GET("/category/:id", Category)
	// append legacy redirects
	for url := range LegacyURLs {
		g.GET(url, Redirection)
	}
}

func Index(c echo.Context) error {
	const desc = "Welcome to the Firefox 2 era (October 2006) Defacto2 website, " +
		"that is friendly for legacy operating systems including Windows 9x, NT-4, OS-X 10.2." // TODO: share this with html meta OR make this html templ
	start := latency()
	ctx := context.Background()
	db, err := postgres.ConnectDB()
	if err != nil {
		return err
	}
	defer db.Close()

	// TODO: defer and cache results
	art, doc, sw := 0, 0, 0
	art, _ = models.ArtImagesCount(ctx, db)
	doc, _ = models.DocumentCount(ctx, db)
	sw, _ = models.SoftwareCount(ctx, db)

	return c.Render(http.StatusOK, "index", map[string]interface{}{
		"title":       title,
		"description": desc,
		"art":         router.LeadInt(5, art),
		"doc":         router.LeadInt(5, doc),
		"sw":          router.LeadInt(5, sw),
		"cat":         router.LeadInt(5, meta.CategoryCount),
		"latency":     fmt.Sprintf("%s.", time.Since(*start)),
	})
}

func Categories(c echo.Context) error {
	start := latency()
	return c.Render(http.StatusOK, "categories", map[string]interface{}{
		"title":   title + "/categories",
		"latency": fmt.Sprintf("%s.", time.Since(*start)),
		"tags":    meta.Names,
		"cats":    meta.Categories,
	})
}

func Category(c echo.Context) error {
	start := latency()
	value := c.Param("id")
	ctx := context.Background()
	db, err := postgres.ConnectDB()
	if err != nil {
		return err
	}
	defer db.Close()
	records, err := models.FilesByCategory(value, c.QueryString(), ctx, db)
	if err != nil {
		return err
	}
	count := len(records)
	if count == 0 {
		return echo.NewHTTPError(http.StatusNotFound,
			fmt.Sprintf("The category %q doesn't exist", value))
	}
	sum, err := models.ByteCountByCategory(value, ctx, db)
	if err != nil {
		return err
	}
	key := meta.GetURI(value)
	info := meta.Infos[key]
	name := meta.Names[key]
	desc := fmt.Sprintf("%s - %s.", name, info)
	stat := fmt.Sprintf("%d files, %s", count, helpers.ByteCountLong(sum))
	sorter := sorter(c.QueryString())
	return c.Render(http.StatusOK, "category", map[string]interface{}{
		"title":       fmt.Sprintf("%s%s%s", title, "/category/", value),
		"home":        "",
		"description": desc,
		"stats":       stat,
		"sort":        sorter,
		"records":     records,
		"latency":     fmt.Sprintf("%s.", time.Since(*start)),
	})
}

// todo: custom type
func sorter(query string) map[string]string {
	s := Sortings
	switch strings.ToUpper(query) {
	case models.NameAsc:
		s["Name"] = "D"
	case models.NameDes:
		s["Name"] = "A"
	case models.PublAsc:
		s["Publish"] = "D"
	case models.PublDes:
		s["Publish"] = "A"
	case models.PostAsc:
		s["Posted"] = "D"
	case models.PostDes:
		s["Posted"] = "A"
	case models.SizeAsc:
		s["Size"] = "D"
	case models.SizeDes:
		s["Size"] = "A"
	case models.DescAsc:
		s["Description"] = "D"
	case models.DescDes:
		s["Description"] = "A"
	}
	return Sortings
}

func Error(err error, c echo.Context) error {
	// Echo custom error handling: https://echo.labstack.com/guide/error-handling/
	start := latency()
	code := http.StatusInternalServerError
	msg := "This is a server problem"
	if he, ok := err.(*echo.HTTPError); ok {
		code = he.Code
		msg = fmt.Sprint(he.Message)
	}
	// TODO: switch codes and use a logger?
	return c.Render(code, "error", map[string]interface{}{
		"title":       fmt.Sprintf("%d error, there is a complication", code),
		"description": fmt.Sprintf("%s.", msg),
		"latency":     fmt.Sprintf("%s.", time.Since(*start)),
	})
}

func Redirection(c echo.Context) error {
	for u, redirect := range LegacyURLs {
		htm := Root + u
		if htm == c.Path() {
			return c.Redirect(http.StatusPermanentRedirect, Root+redirect)
		}
	}
	return c.String(http.StatusInternalServerError,
		fmt.Sprintf("unknown redirection, %q ", c.Path()))
}
