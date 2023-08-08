package app

// Package file render_err.go contains the custom render error pages for the website.

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// StatusErr is the handler for the HTTP status pages such as the 404 - not found.
func StatusErr(s *zap.SugaredLogger, c echo.Context, code int, uri string) error {
	if s == nil {
		return echo.NewHTTPError(http.StatusInternalServerError,
			fmt.Errorf("%w: handler app status", ErrLogger))
	}
	if c == nil {
		return echo.NewHTTPError(http.StatusInternalServerError,
			fmt.Errorf("%w: handler app status", ErrContext))
	}

	// todo: check valid status, or throw error

	data := empty()
	data["description"] = fmt.Sprintf("HTTP status %d error", code)
	title := fmt.Sprintf("%d error", code)
	alert := ""
	logo := "??!!"
	probl := "There is a complication"
	switch code {
	case http.StatusNotFound:
		title = "404 error, page not found"
		logo = "Page not found"
		alert = "The page cannot be found"
		probl = "The page you are looking for might have been removed, had its name changed, or is temporarily unavailable."
	case http.StatusForbidden:
		title = "403 error, forbidden"
		logo = "Forbidden"
		alert = "The page is locked"
		probl = "You don't have permission to access this resource."
	case http.StatusInternalServerError:
		title = "500 error, there is a complication"
		logo = "Server error"
		alert = "There is a complication"
		probl = "The server encountered an internal error or misconfiguration and was unable to complete your request."
	}
	data["title"] = title
	data["code"] = code
	data["logo"] = logo
	data["alert"] = alert
	data["probl"] = probl
	data["uri"] = uri
	err := c.Render(http.StatusNotFound, "status", data)
	if err != nil {
		s.Errorf("%s: %s", ErrTmpl, err)
		return echo.NewHTTPError(http.StatusInternalServerError, ErrTmpl)
	}
	return nil
}
