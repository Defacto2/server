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
	"log"
	"net/http"
	"strings"

	"github.com/Defacto2/releaser"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// AboutErr renders the about file error page for the About files links.
func AboutErr(z *zap.SugaredLogger, c echo.Context, id string) error {
	const name = "status"
	if z == nil {
		return InternalErr(z, c, name, ErrZap)
	}
	if c == nil {
		return InternalErr(z, c, name, ErrCxt)
	}
	data := empty(c)
	data["title"] = fmt.Sprintf("%d error, file about page not found", http.StatusNotFound)
	data["description"] = fmt.Sprintf("HTTP status %d error", http.StatusNotFound)
	data["code"] = http.StatusNotFound
	data["logo"] = "About file not found"
	data["alert"] = fmt.Sprintf("About file %q cannot be found", strings.ToLower(id))
	data["probl"] = "The about file page does not exist, there is probably a typo with the URL."
	data["uriOkay"] = "f/"
	data["uriErr"] = id
	err := c.Render(http.StatusNotFound, name, data)
	if err != nil {
		return InternalErr(z, c, name, err)
	}
	return nil
}

// DatabaseErr is the handler for handling database connection errors.
func DatabaseErr(z *zap.SugaredLogger, c echo.Context, uri string, err error) error {
	const code = http.StatusInternalServerError
	if z == nil {
		zapNil(err)
	} else if err != nil {
		z.Errorf("%d error for %q: %s", code, uri, err)
	}
	// render the fallback, text only error page
	if c == nil {
		if z == nil {
			zapNil(fmt.Errorf("%w: databaserr", ErrCxt))
		} else {
			z.Warnf("%s: %s", ErrTmpl, ErrCxt)
		}
		return echo.NewHTTPError(http.StatusInternalServerError,
			fmt.Errorf("%w: handler app status", ErrCxt))
	}
	// render a user friendly error page
	data := empty(c)
	data["description"] = fmt.Sprintf("HTTP status %d error", code)
	data["title"] = "500 error, there is a complication"
	data["code"] = code
	data["logo"] = "Database error"
	data["alert"] = "Cannot connect to the database!"
	data["uriErr"] = ""
	data["probl"] = "This is not your fault, but the server cannot communicate with the database to display this page."
	if err := c.Render(code, "status", data); err != nil {
		if z != nil {
			z.Errorf("%s: %s", ErrTmpl, err)
		}
		return echo.NewHTTPError(http.StatusInternalServerError, ErrTmpl)
	}
	return nil
}

// DownloadErr is the handler for missing download files and database ID errors.
func DownloadErr(z *zap.SugaredLogger, c echo.Context, uri string, err error) error {
	const code = http.StatusNotFound
	id := c.Param("id")
	if z == nil {
		zapNil(err)
	} else if err != nil {
		z.Errorf("%d error for %q: %s", code, id, err)
	}
	// render the fallback, text only error page
	if c == nil {
		if z == nil {
			zapNil(fmt.Errorf("%w: downloaderr", ErrCxt))
		} else {
			z.Errorf("%s: %s", ErrTmpl, ErrCxt)
		}
		return echo.NewHTTPError(http.StatusInternalServerError,
			fmt.Errorf("%w: handler app status", ErrCxt))
	}
	// render a user friendly error page
	data := empty(c)
	data["description"] = fmt.Sprintf("HTTP status %d error", code)
	data["title"] = "404 download error"
	data["code"] = code
	data["logo"] = "Download problem"
	data["alert"] = "Cannot send you this download"
	data["probl"] = "The download you are looking for might have been removed, " +
		"had its filename changed, or is temporarily unavailable. " +
		"Is the URL correct?"
	data["uriErr"] = strings.Join([]string{uri, id}, "/")
	if err := c.Render(code, "status", data); err != nil {
		if z != nil {
			z.Errorf("%s: %s", ErrTmpl, err)
		}
		return echo.NewHTTPError(http.StatusInternalServerError, ErrTmpl)
	}
	return nil
}

// FilesErr renders the files error page for the Files menu and categories.
// It provides different error messages to the standard error page.
func FilesErr(z *zap.SugaredLogger, c echo.Context, uri string) error {
	const name = "status"
	if z == nil {
		return InternalErr(z, c, name, ErrZap)
	}
	if c == nil {
		return InternalErr(z, c, name, ErrCxt)
	}
	data := empty(c)
	data["title"] = fmt.Sprintf("%d error, files page not found", http.StatusNotFound)
	data["description"] = fmt.Sprintf("HTTP status %d error", http.StatusNotFound)
	data["code"] = http.StatusNotFound
	data["logo"] = "Files not found"
	data["alert"] = "Files page cannot be found"
	data["probl"] = "The files category or menu option does not exist, there is probably a typo with the URL."
	data["uriOkay"] = "files/"
	data["uriErr"] = uri
	err := c.Render(http.StatusNotFound, name, data)
	if err != nil {
		return InternalErr(z, c, name, err)
	}
	return nil
}

// PageErr renders the files page error page for the Files menu and categories.
// It provides different error messages to the standard error page.
func PageErr(z *zap.SugaredLogger, c echo.Context, uri, page string) error {
	const name = "status"
	if z == nil {
		return InternalErr(z, c, name, ErrZap)
	}
	if c == nil {
		return InternalErr(z, c, name, ErrCxt)
	}
	data := empty(c)
	data["title"] = fmt.Sprintf("%d error, files page not found", http.StatusNotFound)
	data["description"] = fmt.Sprintf("HTTP status %d error", http.StatusNotFound)
	data["code"] = http.StatusNotFound
	data["logo"] = "Page not found"
	data["alert"] = fmt.Sprintf("Files %s page does not exist", uri)
	data["probl"] = "The files page does not exist, there is probably a typo with the URL."
	data["uriOkay"] = fmt.Sprintf("files/%s/", uri)
	data["uriErr"] = page
	err := c.Render(http.StatusNotFound, name, data)
	if err != nil {
		return InternalErr(z, c, name, err)
	}
	return nil
}

// ReleaserErr renders the files error page for the Groups menu and invalid releasers.
func ReleaserErr(z *zap.SugaredLogger, c echo.Context, id string) error {
	const name = "status"
	if z == nil {
		return InternalErr(z, c, name, ErrZap)
	}
	if c == nil {
		return InternalErr(z, c, name, ErrCxt)
	}
	data := empty(c)
	data["title"] = fmt.Sprintf("%d error, releaser page not found", http.StatusNotFound)
	data["description"] = fmt.Sprintf("HTTP status %d error", http.StatusNotFound)
	data["code"] = http.StatusNotFound
	data["logo"] = "Releaser not found"
	data["alert"] = fmt.Sprintf("Releaser %q cannot be found", releaser.Humanize(id))
	data["probl"] = "The releaser page does not exist, there is probably a typo with the URL."
	data["uriOkay"] = "g/"
	data["uriErr"] = id
	err := c.Render(http.StatusNotFound, name, data)
	if err != nil {
		return InternalErr(z, c, name, err)
	}
	return nil
}

// ScenerErr renders the files error page for the People menu and invalid sceners.
func ScenerErr(z *zap.SugaredLogger, c echo.Context, id string) error {
	const name = "status"
	if z == nil {
		return InternalErr(z, c, name, ErrZap)
	}
	if c == nil {
		return InternalErr(z, c, name, ErrCxt)
	}
	data := empty(c)
	data["title"] = fmt.Sprintf("%d error, scener page not found", http.StatusNotFound)
	data["description"] = fmt.Sprintf("HTTP status %d error", http.StatusNotFound)
	data["code"] = http.StatusNotFound
	data["logo"] = "Scener not found"
	data["alert"] = fmt.Sprintf("Scener %q cannot be found", releaser.Humanize(id))
	data["probl"] = "The scener page does not exist, there is probably a typo with the URL."
	data["uriOkay"] = "p/"
	data["uriErr"] = id
	err := c.Render(http.StatusNotFound, name, data)
	if err != nil {
		return InternalErr(z, c, name, err)
	}
	return nil
}

func zapNil(err error) {
	log.Println(fmt.Errorf("cannot log the following error: %w", err))
}

// InternalErr is the handler for handling Internal Server Errors, caused by programming bugs or crashes.
// The uri string is the part of the URL that caused the error.
// The optional error value is logged using the zap sugared logger.
// If the zap logger is nil then the error page is returned but no error is logged.
// If the echo context is nil then a user hostile, fallback error in raw text is returned.
func InternalErr(z *zap.SugaredLogger, c echo.Context, uri string, err error) error {
	const code = http.StatusInternalServerError
	if z == nil {
		zapNil(err)
	} else if err != nil {
		z.Errorf("%d error for %q: %s", code, uri, err)
	}
	// render the fallback, text only error page
	if c == nil {
		if z == nil {
			zapNil(fmt.Errorf("%w: internalerr", ErrCxt))
		} else {
			z.Errorf("%s: %s", ErrTmpl, ErrCxt)
		}
		return echo.NewHTTPError(http.StatusInternalServerError,
			fmt.Errorf("%w: handler app status", ErrCxt))
	}
	// render a user friendly error page
	data := empty(c)
	data["description"] = fmt.Sprintf("HTTP status %d error", code)
	data["title"] = "500 error, there is a complication"
	data["code"] = code
	data["logo"] = "Server error"
	data["alert"] = "Something crashed!"
	data["probl"] = "This is not your fault," +
		" but the server encountered an internal error or misconfiguration and cannot display this page."
	data["uriErr"] = uri
	if err := c.Render(code, "status", data); err != nil {
		if z != nil {
			z.Errorf("%s: %s", ErrTmpl, err)
		}
		return echo.NewHTTPError(http.StatusInternalServerError, ErrTmpl)
	}
	return nil
}

// BadRequestErr is the handler for handling Bad Request Errors, caused by invalid user input
// or a malformed client requests.
func BadRequestErr(z *zap.SugaredLogger, c echo.Context, uri string, err error) error {
	const code = http.StatusBadRequest
	if z == nil {
		zapNil(err)
	} else if err != nil {
		z.Errorf("%d error for %q: %s", code, uri, err)
	}
	// render the fallback, text only error page
	if c == nil {
		if z == nil {
			zapNil(fmt.Errorf("%w: internalerr", ErrCxt))
		} else {
			z.Errorf("%s: %s", ErrTmpl, ErrCxt)
		}
		return echo.NewHTTPError(http.StatusInternalServerError,
			fmt.Errorf("%w: handler app status", ErrCxt))
	}
	// render a user friendly error page
	data := empty(c)
	data["description"] = fmt.Sprintf("HTTP status %d error", code)
	data["title"] = "400 error, there is a complication"
	data["code"] = code
	data["logo"] = "Client error"
	data["alert"] = "Something went wrong, " + err.Error()
	data["probl"] = "It might be a settings or configuration problem or a legacy browser issue."
	data["uriErr"] = uri
	if err := c.Render(code, "status", data); err != nil {
		if z != nil {
			z.Errorf("%s: %s", ErrTmpl, err)
		}
		return echo.NewHTTPError(http.StatusInternalServerError, ErrTmpl)
	}
	return nil
}

// ForbiddenErr is the handler for handling Forbidden Errors, caused by clients requesting
// pages that they do not have permission to access.
func ForbiddenErr(z *zap.SugaredLogger, c echo.Context, uri string, err error) error {
	const code = http.StatusForbidden
	if z == nil {
		zapNil(err)
	} else if err != nil {
		z.Errorf("%d error for %q: %s", code, uri, err)
	}
	// render the fallback, text only error page
	if c == nil {
		if z == nil {
			zapNil(fmt.Errorf("%w: internalerr", ErrCxt))
		} else {
			z.Errorf("%s: %s", ErrTmpl, ErrCxt)
		}
		return echo.NewHTTPError(http.StatusInternalServerError,
			fmt.Errorf("%w: handler app status", ErrCxt))
	}
	// render a user friendly error page
	data := empty(c)
	data["description"] = fmt.Sprintf("HTTP status %d error", code)
	data["title"] = "403, forbidden"
	data["code"] = code
	data["logo"] = "Forbidden"
	data["alert"] = "This page is locked"
	data["probl"] = fmt.Sprintf("This page is not intended for the general public, %s.", err.Error())
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
		zapNil(fmt.Errorf("%w: statuserr", ErrZap))
	}
	// render the fallback, text only error page
	if c == nil {
		if z == nil {
			zapNil(fmt.Errorf("%w: statuserr", ErrCxt))
		} else {
			z.Errorf("%s: %s", ErrTmpl, ErrCxt)
		}
		return echo.NewHTTPError(http.StatusInternalServerError,
			fmt.Errorf("%w: handler app status", ErrCxt))
	}
	// render a user friendly error page
	data := empty(c)
	data["description"] = fmt.Sprintf("HTTP status %d error", code)
	var title, alert, logo, probl string
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
			err := fmt.Errorf("%w: %d", ErrCode, code)
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