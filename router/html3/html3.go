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

	"github.com/Defacto2/sceners"
	"github.com/Defacto2/server/helpers"
	"github.com/Defacto2/server/models"
	"github.com/Defacto2/server/tags"
	"github.com/labstack/echo/v4"

	"github.com/Defacto2/server/postgres"
	pgm "github.com/Defacto2/server/postgres/models"
)

type TagType int

const (
	SectionTag TagType = iota
	PlatformTag
	GroupTag
)

func (t TagType) String() string {
	return [...]string{"category", "platform", "group"}[t]
}

func (t TagType) Parent() string {
	return [...]string{"categories", "platforms", "groups"}[t]
}

const (
	Root  = "/html3"
	title = "Index of " + Root
)

var Counts = models.Counts

var LegacyURLs = map[string]string{
	"/index":            "",
	"/categories/index": "/categories",
	"/platforms/index":  "/platforms",
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
	g.GET("/platforms", Platforms)
	g.GET("/platform/:id", Platform)
	g.GET("/groups", Groups)
	g.GET("/group/:id", Group)
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
	art, doc, sw, grp := 0, 0, 0, 0
	// TODO: log errors
	art, _ = models.ArtImagesCount(ctx, db)
	doc, _ = models.DocumentCount(ctx, db)
	sw, _ = models.SoftwareCount(ctx, db)
	grp, _ = models.GroupsTotalCount(ctx, db)

	return c.Render(http.StatusOK, "index", map[string]interface{}{
		"title":       title,
		"description": desc,
		"art":         art,
		"doc":         doc,
		"sw":          sw,
		"grp":         grp,
		"cat":         tags.CategoryCount,
		"plat":        tags.PlatformCount,
		"latency":     fmt.Sprintf("%s.", time.Since(*start)),
	})
}

func Categories(c echo.Context) error {
	start := latency()
	return c.Render(http.StatusOK, "metadata", map[string]interface{}{
		"title":    title + "/categories",
		"latency":  fmt.Sprintf("%s.", time.Since(*start)),
		"path":     "category",
		"tagFirst": tags.FirstCategory,
		"tagEnd":   tags.LastCategory,
		"tags":     tags.Names,
	})
}

func Platforms(c echo.Context) error {
	start := latency()
	return c.Render(http.StatusOK, "metadata", map[string]interface{}{
		"title":    title + "/platforms",
		"latency":  fmt.Sprintf("%s.", time.Since(*start)),
		"path":     "platform",
		"tagFirst": tags.FirstPlatform,
		"tagEnd":   tags.LastPlatform,
		"tags":     tags.Names,
	})
}

func Groups(c echo.Context) error {
	start := latency()
	ctx := context.Background()
	db, err := postgres.ConnectDB()
	if err != nil {
		return err
	}
	defer db.Close()
	s, err := models.Groups(ctx, db)
	if err != nil {
		return err
	}
	fmt.Println("GROUP COUNT", len(s))
	return c.Render(http.StatusOK, "group", map[string]interface{}{
		"title":   title + "/groups",
		"latency": fmt.Sprintf("%s.", time.Since(*start)),
		"path":    "group",
		"sceners": s,
	})
}

func Category(c echo.Context) error {
	return Tag(SectionTag, c)
}

func Platform(c echo.Context) error {
	return Tag(PlatformTag, c)
}

func Group(c echo.Context) error {
	return Tag(GroupTag, c)
}

func Tag(tt TagType, c echo.Context) error {
	start := latency()
	value := c.Param("id")
	ctx := context.Background()
	db, err := postgres.ConnectDB()
	if err != nil {
		return err
	}
	defer db.Close()
	var records pgm.FileSlice
	switch tt {
	case SectionTag:
		records, err = models.FilesByCategory(value, c.QueryString(), ctx, db)
	case PlatformTag:
		records, err = models.FilesByPlatform(value, c.QueryString(), ctx, db)
	case GroupTag:
		name := sceners.CleanURL(value) //CleanURL(value)
		records, err = models.FilesByGroup(name, c.QueryString(), ctx, db)
	}
	if err != nil {
		return err
	}
	count := len(records)
	fmt.Println("VALUE", value, count)
	if count == 0 {
		return echo.NewHTTPError(http.StatusNotFound,
			fmt.Sprintf("The %s %q doesn't exist", tt, value))
	}
	var sum int64
	switch tt {
	case SectionTag:
		sum, err = models.ByteCountByCategory(value, ctx, db)
	case PlatformTag:
		sum, err = models.ByteCountByPlatform(value, ctx, db)
	}
	if err != nil {
		return err
	}
	key := tags.GetURI(value)
	info := tags.Infos[key]
	name := tags.Names[key]
	desc := fmt.Sprintf("%s - %s.", name, info)
	stat := fmt.Sprintf("%d files, %s", count, helpers.ByteCountLong(sum))
	sorter := sorter(c.QueryString())
	return c.Render(http.StatusOK, tt.String(), map[string]interface{}{
		"title":       fmt.Sprintf("%s%s%s", title, fmt.Sprintf("/%s/", tt), value),
		"home":        "",
		"description": desc,
		"parent":      tt.Parent(),
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
