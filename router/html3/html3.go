package html3

import (
	"context"
	"fmt"
	"math/big"
	"net/http"
	"time"

	"github.com/bengarrett/df2023/meta"
	"github.com/bengarrett/df2023/models"
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

func RedirCategories(c echo.Context) error {
	return c.Redirect(http.StatusPermanentRedirect, "/html3/categories")
}

func RedirIndex(c echo.Context) error {
	return c.Redirect(http.StatusPermanentRedirect, "/html3")
}

func Index(c echo.Context) error {
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
		"title":   title,
		"art":     art,
		"doc":     doc,
		"sw":      sw,
		"latency": fmt.Sprintf("%s.", time.Since(*start)),
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
	return c.Render(http.StatusOK, "category", map[string]interface{}{
		"title":   fmt.Sprintf("%s%s%s", title, "/category/", c.Param("id")),
		"latency": fmt.Sprintf("%s.", time.Since(*start)),
	})
}
