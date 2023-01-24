// Package html3 handles the routes and views for the retro,
// mini-website that is rendered in HTML 3 syntax.
package html3

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
)

// HTTP status codes in Go
// https://go.dev/src/net/http/status.go

// LegacyURLs are partial URL routers that are to be redirected with a HTTP 308
// permanent redirect status code. These are for retired URL syntaxes that are still
// found on websites online, so their links to Defacto2 do not break with 404, not found errors.
var LegacyURLs = map[string]string{
	"/index":            "",
	"/categories/index": "/categories",
	"/platforms/index":  "/platforms",
}

// GroupCache is a cached collection of important, expensive group data.
// The Mu mutex must always be locked before writing this varable.
var IndexCache IndexSums

// GroupCol is a cached collection of important, expensive group data.
// The Mu mutex must always be locked when writing to the Groups map.
type IndexSums struct {
	Mu   sync.Mutex
	Sums map[int]int
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