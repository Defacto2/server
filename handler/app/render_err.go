package app

// Package file render_err.go contains the custom render error pages for the website.
//
// Each customer error requires the following data keys and values to display correctly:
// Title: tab title
// Description: meta description
// Code: HTTP status code, this gets shown on the page
// Logo: the logo text
// Alert: the alert message box title
// Prob: the alert message box text
// uriOkay: the URI to the current page that is okay, this is shown on the page
// uriErr: the broken or unknown URI to the page, this is shown and gets underlined in red on the page

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// FilesErr renders the files error page for the Files menu and categories.
// It provides different error messages to the standard error page.
func FilesErr(z *zap.SugaredLogger, c echo.Context, uri string) error {
	if z == nil {
		return echo.NewHTTPError(http.StatusInternalServerError,
			fmt.Errorf("%w: handler app files", ErrLogger))
	}
	if c == nil {
		return echo.NewHTTPError(http.StatusInternalServerError,
			fmt.Errorf("%w: handler app files", ErrContext))
	}
	data := empty()
	data["title"] = fmt.Sprintf("%d error, files page not found", http.StatusNotFound)
	data["description"] = fmt.Sprintf("HTTP status %d error", http.StatusNotFound)
	data["code"] = http.StatusNotFound
	data["logo"] = "Files not found"
	data["alert"] = "Files page cannot be found"
	data["probl"] = "The files category or menu option does not exist, there is probably a typo with the URL."
	data["uriOkay"] = "files/"
	data["uriErr"] = uri

	err := c.Render(http.StatusNotFound, "status", data)
	if err != nil {
		z.Errorf("%s: %s", ErrTmpl, err)
		return echo.NewHTTPError(http.StatusInternalServerError, ErrTmpl)
	}
	return nil
}

// StatusErr is the handler for the HTTP status pages such as the 404 - not found.
func StatusErr(z *zap.SugaredLogger, c echo.Context, code int, uri string) error {
	if z == nil {
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
	default:
		s := http.StatusText(code)
		if s == "" {
			err := fmt.Errorf("%s: %d", ErrCode, code)
			z.Errorf("%s", err)
			return echo.NewHTTPError(http.StatusInternalServerError, err)
		}
		title = fmt.Sprintf("%d error, %s", code, s)
		logo = s
		alert = s
		probl = fmt.Sprintf("%d error, %s", code, s)
	}
	data["title"] = title
	data["code"] = code
	data["logo"] = logo
	data["alert"] = alert
	data["probl"] = probl
	data["uriErr"] = uri
	err := c.Render(http.StatusNotFound, "status", data)
	if err != nil {
		z.Errorf("%s: %s", ErrTmpl, err)
		return echo.NewHTTPError(http.StatusInternalServerError, ErrTmpl)
	}
	return nil
}
