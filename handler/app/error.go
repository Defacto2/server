package app

// Package file error.go contains the error handlers for the application.

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"syscall"

	"github.com/Defacto2/server/internal/nils"
	"github.com/labstack/echo/v5"
)

const ErrTmpl = "the server could not render the html template for this page"

// BadRequestErr is the handler for handling Bad Request Errors, caused by invalid user input
// or a malformed client requests.
func BadRequestErr(sl *slog.Logger, c *echo.Context, uri string, err error) error {
	const title = "400 error, there is a complication"
	const logo = "Client error"
	const probl = "It might be a settings or configuration problem or a legacy browser issue."
	const msg = "bad request handler"
	if err1 := nils.Check(sl, c); err1 != nil {
		const format = msg + " %q: %w"
		return fmt.Errorf(format, uri, err1)
	}
	const badRequest = http.StatusBadRequest
	if err != nil {
		sl.Error(msg, slog.Int("code", badRequest), slog.String("uri", uri), slog.String("error", err.Error()))
	}
	data := empty(c)
	data["description"] = fmt.Sprintf("HTTP status %d error", badRequest)
	data["title"] = title
	data["code"] = badRequest
	data["logo"] = logo
	data["alert"] = "Something went wrong, " + err.Error()
	data["probl"] = probl
	data["uriErr"] = uri
	if err := c.Render(badRequest, "status", data); err != nil {
		sl.Error(msg, slog.Int("code", badRequest), slog.String("uri", uri), slog.String("error", err.Error()))
		return echo.NewHTTPError(http.StatusInternalServerError, ErrTmpl)
	}
	return nil
}

// DatabaseErr is the handler for database connection issues.
// A HTTP 503 Service Unavailable error is returned, to reflect the database
// connection issue but where the server is still running and usable for the client.
func DatabaseErr(sl *slog.Logger, c *echo.Context, uri string, err error) error {
	const title = "Cannot connect to the database!"
	const logo = "Database error"
	const probl = "This is not your fault, but the server cannot communicate with the database to display this page."
	const msg = "database connection handler"
	if err1 := nils.Check(sl, c); err1 != nil {
		const format = msg + " %q: %w"
		return fmt.Errorf(format, uri, err1)
	}
	const unavailable = http.StatusServiceUnavailable
	if err != nil {
		sl.Error(msg,
			slog.String("connection", "cannot connect to the database"),
			slog.Int("code", unavailable), slog.String("uri", uri),
			slog.Any("error", err))
	}
	data := empty(c)
	data["description"] = fmt.Sprintf("HTTP status %d error", unavailable)
	data["title"] = fmt.Sprintf("%d error, there is a complication", unavailable)
	data["code"] = fmt.Sprintf("%d service unavailable", unavailable)
	data["logo"] = logo
	data["alert"] = title
	data["uriErr"] = ""
	data["probl"] = probl
	if err := c.Render(unavailable, "status", data); err != nil {
		sl.Error(msg, slog.Int("code", unavailable), slog.String("uri", uri), slog.String("error", err.Error()))
		return echo.NewHTTPError(unavailable, ErrTmpl)
	}
	return nil
}

// DownloadErr is the handler for missing download files and database ID errors.
func DownloadErr(sl *slog.Logger, c *echo.Context, uri string, err error) error { //nolint:dupl
	const title = "404 download error"
	const logo = "Download problem"
	const alert = "Cannot send you this download"
	const probl = "The download you are looking for might have been removed, " +
		"had its filename changed, or is temporarily unavailable. Is the URL correct?"
	const msg = "download not found"
	if err1 := nils.Check(sl, c); err1 != nil {
		const format = msg + " %q: %w"
		return fmt.Errorf(format, uri, err1)
	}
	const notFound = http.StatusNotFound
	id := c.Param("id")
	if err != nil {
		sl.Error(msg, slog.Int("code", notFound), slog.String("id", id),
			slog.String("uri", uri), slog.Any("error", err))
	}
	data := empty(c)
	data["description"] = fmt.Sprintf("HTTP status %d error", notFound)
	data["title"] = title
	data["code"] = notFound
	data["logo"] = logo
	data["alert"] = alert
	data["probl"] = probl
	data["uriErr"] = strings.Join([]string{uri, id}, "/")
	if err := c.Render(notFound, "status", data); err != nil {
		sl.Error(msg, slog.Int("code", notFound), slog.String("uri", uri), slog.String("error", err.Error()))
		return echo.NewHTTPError(http.StatusInternalServerError, ErrTmpl)
	}
	return nil
}

// FileMissingErr is the handler for missing download files and database ID errors.
func FileMissingErr(sl *slog.Logger, c *echo.Context, uri string, err error) error { //nolint:dupl
	const title = "503 download unavailable"
	const logo = "Download unavailable"
	const alert = "Cannot send you this download"
	const probl = "The file download needs to be added to the server; " +
		"otherwise, there may be a problem with the server configuration, or the file may be lost."
	const msg = "file missing"
	if err1 := nils.Check(sl, c); err1 != nil {
		const format = msg + " %q: %w"
		return fmt.Errorf(format, uri, err1)
	}
	const serviceNA = http.StatusServiceUnavailable
	id := c.Param("id")
	if err != nil {
		sl.Error(msg, slog.Int("code", serviceNA), slog.String("id", id),
			slog.String("uri", uri), slog.Any("error", err))
	}
	data := empty(c)
	data["description"] = fmt.Sprintf("HTTP status %d error", serviceNA)
	data["title"] = title
	data["code"] = serviceNA
	data["logo"] = logo
	data["alert"] = alert
	data["probl"] = probl
	data["uriErr"] = strings.Join([]string{uri, id}, "/")
	if err := c.Render(serviceNA, "status", data); err != nil {
		sl.Error(msg, slog.Int("code", serviceNA), slog.String("uri", uri), slog.String("error", err.Error()))
		return echo.NewHTTPError(http.StatusInternalServerError, ErrTmpl)
	}
	return nil
}

// ForbiddenErr is the handler for handling Forbidden Errors, caused by clients requesting
// pages that they do not have permission to access.
func ForbiddenErr(sl *slog.Logger, c *echo.Context, uri string, err error) error {
	const title = "403, forbidden"
	const logo = "Forbidden"
	const alert = "This page is locked"
	const msg = "forbidden access"
	if err1 := nils.Check(sl, c); err1 != nil {
		const format = msg + " %q: %w"
		return fmt.Errorf(format, uri, err1)
	}
	const forbidden = http.StatusForbidden
	if err != nil {
		sl.Error(msg, slog.Int("code", forbidden),
			slog.String("uri", uri), slog.Any("error", err))
	}
	data := empty(c)
	data["description"] = fmt.Sprintf("HTTP status %d error", forbidden)
	data["title"] = title
	data["code"] = forbidden
	data["logo"] = logo
	data["alert"] = alert
	if err != nil {
		data["probl"] = fmt.Sprintf("This page is not intended for the general public, %s.", err.Error())
	}
	data["uriErr"] = uri
	if err := c.Render(forbidden, "status", data); err != nil {
		sl.Error(msg, slog.Int("code", forbidden), slog.String("uri", uri), slog.String("error", err.Error()))
		return echo.NewHTTPError(http.StatusInternalServerError, ErrTmpl)
	}
	return nil
}

// InternalErr is the handler for handling Internal Server Errors, caused by programming bugs or crashes.
// The uri string is the part of the URL that caused the error.
// The optional error value is logged using the logger.
// If the echo context is nil then a user hostile, fallback error in raw text is returned.
func InternalErr(sl *slog.Logger, c *echo.Context, uri string, err error) error {
	const title = "500 error, there is a complication"
	const logo = "Server error"
	const alert = "Something crashed!"
	const probl = "This is not your fault," +
		" but the server encountered an internal error or misconfiguration and cannot display this page."
	const msg = "internal server error"
	if err1 := nils.Check(sl, c); err1 != nil {
		const format = msg + " %q: %w"
		return fmt.Errorf(format, uri, err1)
	}
	const internalError = http.StatusInternalServerError
	if errors.Is(err, syscall.EPIPE) {
		// This is a common error when the client disconnects before the response is sent,
		// and commonly happens when using developer hot reloading.
		discard(err)
		return nil
	}
	if err != nil {
		sl.Error(msg, slog.Int("code", internalError),
			slog.String("uri", uri), slog.Any("error", err))
	}
	if errors.Is(err, echo.ErrRendererNotRegistered) {
		message := fmt.Sprintf("%s", err)
		return echo.NewHTTPError(internalError, message)
	}
	data := empty(c)
	data["description"] = fmt.Sprintf("HTTP status %d error", internalError)
	data["title"] = title
	data["code"] = internalError
	data["logo"] = logo
	data["alert"] = alert
	data["probl"] = probl
	data["uriErr"] = uri
	if err := c.Render(internalError, "status", data); err != nil {
		sl.Error(msg, slog.Int("code", internalError), slog.String("uri", uri), slog.String("error", err.Error()))
		return echo.NewHTTPError(http.StatusInternalServerError, ErrTmpl)
	}
	return nil
}

// StatusErr is the handler for the HTTP status pages such as the 404 - not found.
// If the logger is nil then the error page is returned but no error is logged.
// If the echo context is nil then a user hostile, fallback error in raw text is returned.
func StatusErr(sl *slog.Logger, c *echo.Context, code int, uri string) error {
	const msg = "http status"
	if err1 := nils.Check(sl, c); err1 != nil {
		const format = msg + " %q: %w"
		return fmt.Errorf(format, uri, err1)
	}
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
		return InternalErr(sl, c, uri, nil)
	default:
		s := http.StatusText(code)
		if s == "" {
			err := fmt.Sprintf("%d status error for the URL, %s: %s", code, uri, ErrStatus)
			sl.Error(msg, slog.String("status", "unknown and unsupported status code"),
				slog.Int("code", code), slog.String("uri", uri))
			return echo.NewHTTPError(http.StatusInternalServerError, err)
		}
		const fmtcode = "%d error, %s"
		title = fmt.Sprintf(fmtcode, code, s)
		logo = s
		alert = s
		probl = fmt.Sprintf(fmtcode, code, s)
	}
	data["title"] = title
	data["code"] = code
	data["logo"] = logo
	data["alert"] = alert
	data["probl"] = probl
	data["uriErr"] = uri
	if err := c.Render(code, "status", data); err != nil {
		sl.Error(msg, slog.Int("code", code), slog.String("uri", uri), slog.String("error", err.Error()))
		return echo.NewHTTPError(http.StatusInternalServerError, ErrTmpl)
	}
	return nil
}
