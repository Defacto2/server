// Package html3 handles renders the /html3 sub-route of the website.
// This generates pages for the website for browsing of the file database using HTML3 styled tables.
package html3

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Defacto2/server/internal/helper"
	"github.com/labstack/echo/v4"
)

var (
	ErrDB = fmt.Errorf("database value is nil")
)

const (
	errConn = "Sorry, at the moment the server cannot connect to the database"
	errSQL  = "Database connection problem or a SQL error"
	errTag  = "No database query was created for the tag"
	errTmpl = "The server could not render the HTML template for this page"
)

// Error renders a custom HTTP error page for the HTML3 sub-group.
func Error(c echo.Context, err error) error {
	// Echo custom error handling: https://echo.labstack.com/guide/error-handling/
	start := helper.Latency()
	code := http.StatusInternalServerError
	msg := "This is a server problem"
	var httpError *echo.HTTPError
	if errors.As(err, &httpError) {
		code = httpError.Code
		msg = fmt.Sprint(httpError.Message)
	}
	return c.Render(code, "html3_error", map[string]interface{}{
		"title":       fmt.Sprintf("%d error, there is a complication", code),
		"description": fmt.Sprintf("%s.", msg),
		"latency":     fmt.Sprintf("%s.", time.Since(*start)),
	})
}

// ID returns the ID from the URL path.
// This is only used for the category and platform routes.
func ID(c echo.Context) string {
	x := strings.TrimSuffix(c.Path(), ":offset")
	s := strings.Split(x, "/")
	if len(s) != 4 {
		return ""
	}
	return s[3]
}
