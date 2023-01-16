// Package html3 handles the routes and views for the retro,
// mini-website that is rendered in HTML 3 syntax.
package html3

import (
	"context"
	"errors"
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

var ErrByTag = errors.New("unknown bytag record group")

// GroupBy are the record groupings.
type GroupBy int

const (
	BySection  GroupBy = iota // BySection groups records by the section file table column.
	ByPlatform                // BySection groups records by the platform file table column.
	ByGroup                   // ByGroup groups the records by the distinct, group_brand_for file table column.

	asc  = "A" // asc is order by ascending.
	desc = "D" // desc is order by descending.
)

func (t GroupBy) String() string {
	return [...]string{"category", "platform", "group"}[t]
}

func (t GroupBy) Parent() string {
	return [...]string{"categories", "platforms", "groups"}[t]
}

const (
	Root  = "/html3" // Root path of the HTML3 router group.
	title = "Index of " + Root
)

var Counts = models.Counts

// Routes for the /html3 sub-route group.
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

// LegacyURLs are partial URL routers that are to be redirected with a HTTP 308
// permanent redirect status code. These are for retired URL syntaxes that are still
// found on websites online, so their links to Defacto2 do not break with 404, not found errors.
var LegacyURLs = map[string]string{
	"/index":            "",
	"/categories/index": "/categories",
	"/platforms/index":  "/platforms",
}

// Sort is the display name of column that can be used to sort and order the records.
type Sort string

const (
	Name    Sort = "Name"        // Sort records by the filename.
	Publish Sort = "Publish"     // Sort records by the published year, month and day.
	Posted  Sort = "Posted"      // Sort records by the record creation dated.
	Size    Sort = "Size"        // Sort records by the file size in byte units.
	Desc    Sort = "Description" // Sort the records by the title.
)

// Sortings are the name and order of columns that the records can be ordered by.
var Sortings = map[Sort]string{
	Name:    asc,
	Publish: asc,
	Posted:  asc,
	Size:    asc,
	Desc:    asc,
}

// Index is the homepage of the /html3 sub-route.
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

// Categories lists the names, descriptions and sums of the category (section) tags.
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

// Platforms lists the names, descriptions and sums of the platform tags.
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

// Groups lists the names and sums of all the distinct scene groups.
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

// Category lists the file records associated with the category tag that is provided by the ID param in the URL.
func Category(c echo.Context) error {
	return Tag(BySection, c)
}

// Platform lists the file records associated with the platform tag that is provided by the ID param in the URL.
func Platform(c echo.Context) error {
	return Tag(ByPlatform, c)
}

// Group lists the file records associated with the group that is provided by the ID param in the URL.
func Group(c echo.Context) error {
	return Tag(ByGroup, c)
}

// Tag fetches all the records associated with the GroupBy grouping.
func Tag(tt GroupBy, c echo.Context) error {
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
	case BySection:
		records, err = models.FilesByCategory(value, c.QueryString(), ctx, db)
	case ByPlatform:
		records, err = models.FilesByPlatform(value, c.QueryString(), ctx, db)
	case ByGroup:
		name := sceners.CleanURL(value)
		records, err = models.FilesByGroup(name, c.QueryString(), ctx, db)
	default:
		return ErrByTag
	}
	if err != nil {
		return err
	}
	count := len(records)
	if count == 0 {
		return echo.NewHTTPError(http.StatusNotFound,
			fmt.Sprintf("The %s %q doesn't exist", tt, value))
	}
	var byteSum int64
	switch tt {
	case BySection:
		byteSum, err = models.ByteCountByCategory(value, ctx, db)
	case ByPlatform:
		byteSum, err = models.ByteCountByPlatform(value, ctx, db)
	}
	if err != nil {
		return err
	}
	key := tags.TagByURI(value)
	info := tags.Infos[key]
	name := tags.Names[key]
	desc := fmt.Sprintf("%s - %s.", name, info)
	stat := fmt.Sprintf("%d files, %s", count, helpers.ByteCountLong(byteSum))
	sorter := sorter(c.QueryString())
	fmt.Printf("%+v\n", sorter)
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

// Error renders a custom HTTP error page for the HTML3 sub-group.
func Error(err error, c echo.Context) error {
	// Echo custom error handling: https://echo.labstack.com/guide/error-handling/
	start := latency()
	code := http.StatusInternalServerError
	msg := "This is a server problem"
	if he, ok := err.(*echo.HTTPError); ok {
		code = he.Code
		msg = fmt.Sprint(he.Message)
	}
	return c.Render(code, "error", map[string]interface{}{
		"title":       fmt.Sprintf("%d error, there is a complication", code),
		"description": fmt.Sprintf("%s.", msg),
		"latency":     fmt.Sprintf("%s.", time.Since(*start)),
	})
}

// Redirection redirects any legacy URL matches.
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

// 	e.GET("/users/:name", func(c echo.Context) error {
// 		name := c.Param("name")
// 		return c.String(http.StatusOK, name)
//  })
//
//  curl http://localhost:1323/users/Joe

func latency() *time.Time {
	start := time.Now()
	r := new(big.Int)
	const n, k = 1000, 10
	r.Binomial(n, k)
	return &start
}

func sorter(query string) map[string]string {
	s := Sortings
	switch strings.ToUpper(query) {
	case models.NameAsc:
		s[Name] = desc
	case models.NameDes:
		s[Name] = asc
	case models.PublAsc:
		s[Publish] = desc
	case models.PublDes:
		s[Publish] = asc
	case models.PostAsc:
		s[Posted] = desc
	case models.PostDes:
		s[Posted] = asc
	case models.SizeAsc:
		s[Size] = desc
	case models.SizeDes:
		s[Size] = asc
	case models.DescAsc:
		s[Desc] = desc
	case models.DescDes:
		s[Desc] = asc
	}
	// to be usable in the template, convert the map keys into strings
	tmplSorts := make(map[string]string, len(s))
	for key, value := range Sortings {
		tmplSorts[string(key)] = value
	}
	return tmplSorts
}
