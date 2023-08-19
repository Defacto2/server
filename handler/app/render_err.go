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
	"os"

	"github.com/Defacto2/sceners/pkg/rename"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

func zapNil(name string) {
	fmt.Fprintln(os.Stdout,
		fmt.Errorf("%w: %w for %q", ErrTmpl, ErrLogger, name).Error())
}

// FilesErr renders the files error page for the Files menu and categories.
// It provides different error messages to the standard error page.
func FilesErr(z *zap.SugaredLogger, c echo.Context, uri string) error {
	if z == nil {
		zapNil("FilesErr")
		return echo.NewHTTPError(http.StatusInternalServerError,
			fmt.Errorf("%w: handler app files", ErrLogger))
	}
	if c == nil {
		z.Errorf("%s: %s", ErrTmpl, ErrCxt)
		return echo.NewHTTPError(http.StatusInternalServerError,
			fmt.Errorf("%w: handler app files", ErrCxt))
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

// PageErr renders the files page error page for the Files menu and categories.
// It provides different error messages to the standard error page.
func PageErr(z *zap.SugaredLogger, c echo.Context, uri, page string) error {
	if z == nil {
		z.Errorf("%s: %s", ErrTmpl, ErrLogger)
		return echo.NewHTTPError(http.StatusInternalServerError,
			fmt.Errorf("%w: handler app files", ErrLogger))
	}
	if c == nil {
		z.Errorf("%s: %s", ErrTmpl, ErrCxt)
		return echo.NewHTTPError(http.StatusInternalServerError,
			fmt.Errorf("%w: handler app files", ErrCxt))
	}
	data := empty()
	data["title"] = fmt.Sprintf("%d error, files page not found", http.StatusNotFound)
	data["description"] = fmt.Sprintf("HTTP status %d error", http.StatusNotFound)
	data["code"] = http.StatusNotFound
	data["logo"] = "Page not found"
	data["alert"] = fmt.Sprintf("Files %s page does not exist", uri)
	data["probl"] = "The files page does not exist, there is probably a typo with the URL."
	data["uriOkay"] = fmt.Sprintf("files/%s/", uri)
	data["uriErr"] = page

	err := c.Render(http.StatusNotFound, "status", data)
	if err != nil {
		z.Errorf("%s: %s", ErrTmpl, err)
		return echo.NewHTTPError(http.StatusInternalServerError, ErrTmpl)
	}
	return nil
}

// ReleaserErr renders the files error page for the Groups menu and invalid releasers.
func ReleaserErr(z *zap.SugaredLogger, c echo.Context, id string) error {
	if z == nil {
		z.Errorf("%s: %s", ErrTmpl, ErrLogger)
		return echo.NewHTTPError(http.StatusInternalServerError,
			fmt.Errorf("%w: handler app g", ErrLogger))
	}
	if c == nil {
		z.Errorf("%s: %s", ErrTmpl, ErrCxt)
		return echo.NewHTTPError(http.StatusInternalServerError,
			fmt.Errorf("%w: handler app g", ErrCxt))
	}
	data := empty()
	data["title"] = fmt.Sprintf("%d error, releaser page not found", http.StatusNotFound)
	data["description"] = fmt.Sprintf("HTTP status %d error", http.StatusNotFound)
	data["code"] = http.StatusNotFound
	data["logo"] = "Releaser not found"
	data["alert"] = fmt.Sprintf("Releaser %q cannot be found", rename.DeObfuscateURL(id))
	data["probl"] = "The releaser page does not exist, there is probably a typo with the URL."
	data["uriOkay"] = "g/"
	data["uriErr"] = id

	err := c.Render(http.StatusNotFound, "status", data)
	if err != nil {
		z.Errorf("%s: %s", ErrTmpl, err)
		return echo.NewHTTPError(http.StatusInternalServerError, ErrTmpl)
	}
	return nil
}

// ScenerErr renders the files error page for the People menu and invalid sceners.
func ScenerErr(z *zap.SugaredLogger, c echo.Context, id string) error {
	if z == nil {
		z.Errorf("%s: %s", ErrTmpl, ErrLogger)
		return echo.NewHTTPError(http.StatusInternalServerError,
			fmt.Errorf("%w: handler app p", ErrLogger))
	}
	if c == nil {
		z.Errorf("%s: %s", ErrTmpl, ErrCxt)
		return echo.NewHTTPError(http.StatusInternalServerError,
			fmt.Errorf("%w: handler app p", ErrCxt))
	}
	data := empty()
	data["title"] = fmt.Sprintf("%d error, scener page not found", http.StatusNotFound)
	data["description"] = fmt.Sprintf("HTTP status %d error", http.StatusNotFound)
	data["code"] = http.StatusNotFound
	data["logo"] = "Scener not found"
	data["alert"] = fmt.Sprintf("Scener %q cannot be found", rename.DeObfuscateURL(id))
	data["probl"] = "The scener page does not exist, there is probably a typo with the URL."
	data["uriOkay"] = "p/"
	data["uriErr"] = id

	err := c.Render(http.StatusNotFound, "status", data)
	if err != nil {
		z.Errorf("%s: %s", ErrTmpl, err)
		return echo.NewHTTPError(http.StatusInternalServerError, ErrTmpl)
	}
	return nil
}

// InternalErr is the handler for handling Internal Server Errors, caused by programming bugs or crashes.
// The uri string is the part of the URL that caused the error.
// The optional error value is logged using the zap sugared logger.
// If the zap logger is nil then the error page is returned but no error is logged.
// If the echo context is nil then a user hostile, fallback error in raw text is returned.
func InternalErr(z *zap.SugaredLogger, c echo.Context, uri string, err error) error {
	const code = http.StatusInternalServerError
	if z == nil {
		zapNil(uri) // TODO print error?
	} else if err != nil {
		z.Errorf("%d error for %q: %w", code, uri, err)
	}
	// render the fallback, text only error page
	if c == nil {
		if z != nil {
			z.Errorf("%s: %s", ErrTmpl, ErrCxt)
		}
		return echo.NewHTTPError(http.StatusInternalServerError,
			fmt.Errorf("%w: handler app status", ErrCxt))
	}
	data := empty()
	data["description"] = fmt.Sprintf("HTTP status %d error", code)
	data["title"] = "500 error, there is a complication"
	data["code"] = code
	data["logo"] = "Server error"
	data["alert"] = "Something crashed!"
	data["probl"] = "This is not your fault, but the server encountered an internal error or misconfiguration and cannot display this page."
	data["uriErr"] = uri
	if err := c.Render(code, "status", data); err != nil {
		if z != nil {
			z.Errorf("%s: %s", ErrTmpl, err)
		}
		return echo.NewHTTPError(http.StatusInternalServerError, ErrTmpl)
	}
	return nil
}

// StatusErr is the handler for the HTTP status pages such as the 404 - not found.
// If the zap logger is nil then the error page is returned but no error is logged.
// If the echo context is nil then a user hostile, fallback error in raw text is returned.
func StatusErr(z *zap.SugaredLogger, c echo.Context, code int, uri string) error {
	if z == nil {
		zapNil(uri)
	}
	// render the fallback, text only error page
	if c == nil {
		if z != nil {
			z.Errorf("%s: %s", ErrTmpl, ErrCxt)
		}
		return echo.NewHTTPError(http.StatusInternalServerError,
			fmt.Errorf("%w: handler app status", ErrCxt))
	}
	// render a user friendly error page
	data := empty()
	data["description"] = fmt.Sprintf("HTTP status %d error", code)
	title, alert, logo, probl := "", "", "", ""
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
		return InternalErr(z, c, uri, nil)
	default:
		s := http.StatusText(code)
		if s == "" {
			err := fmt.Errorf("%s: %d", ErrCode, code)
			if z != nil {
				z.Error(err)
			}
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

	if err := c.Render(code, "status", data); err != nil {
		if z != nil {
			z.Errorf("%s: %s", ErrTmpl, err)
		}
		return echo.NewHTTPError(http.StatusInternalServerError, ErrTmpl)
	}
	return nil
}
