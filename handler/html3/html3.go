// Package html3 handles the routes and views for the retro,
// Defacto2 mini-website that is rendered in HTML 3 syntax.
package html3

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/Defacto2/server/internal/helper"
	"github.com/labstack/echo/v4"
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
