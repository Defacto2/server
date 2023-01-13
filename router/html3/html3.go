package html3

import (
	"context"
	"fmt"
	"math/big"
	"net/http"
	"time"

	"github.com/bengarrett/df2023/helpers"
	"github.com/bengarrett/df2023/meta"
	"github.com/bengarrett/df2023/models"
	"github.com/bengarrett/df2023/router"
	"github.com/labstack/echo/v4"

	"github.com/bengarrett/df2023/postgres"
)

const title = "Index of /html3"

var Counts = models.Counts

func latency() *time.Time {
	start := time.Now()
	r := new(big.Int)
	r.Binomial(1000, 10)
	return &start
}

func Routes(prefix string, e *echo.Echo) {
	g := e.Group(prefix)
	g.GET("", Index)
	g.GET("/index", RedirIndex) // TODO shared redirect func
	g.GET("/categories", Categories)
	g.GET("/category/:id", Category)
	g.GET("/categories/index", RedirCategories)
}

func Index(c echo.Context) error {
	const desc = "Welcome to the Firefox 2 era (October 2006) Defacto2 website, that is friendly for legacy operating systems including Windows 9x, NT-4, OS-X 10.2." // TODO: share this with html meta OR make this html templ
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
	// TODO: throw a 50x error page to user and log error
	records, _ := models.FilesByCategory(value, ctx, db)
	sum, _ := models.ByteCountByCategory(value, ctx, db)
	count := len(records)
	key := meta.GetURI(value)
	info := meta.Infos[key]
	name := meta.Names[key]
	desc := fmt.Sprintf("%s - %s.", name, info)
	stat := fmt.Sprintf("%d files, %s", count, helpers.ByteCountLong(sum))
	return c.Render(http.StatusOK, "category", map[string]interface{}{
		"title":       fmt.Sprintf("%s%s%s", title, "/category/", value),
		"home":        "",
		"description": desc,
		"stats":       stat,
		"records":     records,
		"latency":     fmt.Sprintf("%s.", time.Since(*start)),
	})
}

func DownloadX(c echo.Context) error {
	err := router.Download(c)
	if err != nil {
		return Error(err, c)
	}
	return nil
	// return c.Render(http.StatusOK, "categories", map[string]interface{}{
	// TODO: if err then render a HTML3 template error
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

// redirects

func RedirCategories(c echo.Context) error {
	return c.Redirect(http.StatusPermanentRedirect, "/html3/categories")
}

func RedirIndex(c echo.Context) error {
	fmt.Printf("\n\n%+v\n\n", c)
	fmt.Printf("%+v <<-- %s\n", c.Path(), "")
	return c.String(500, "oops")
	// return c.Redirect(http.StatusPermanentRedirect, "/html3")
}
