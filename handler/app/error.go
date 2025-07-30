package app

// Package file error.go contains the error handlers for the application.

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"syscall"

	"github.com/Defacto2/server/internal/panics"
	"github.com/Defacto2/server/internal/zaplog"
	"github.com/labstack/echo/v4"
)

var (
	ErrNoDB    = errors.New("database pointer db is nil")
	ErrNoEcho  = errors.New("echo context pointer c is nil")
	ErrNoEmbed = errors.New("embed file system instance is empty")
	ErrNoSlog  = errors.New("logger pointer sl is nil")
)

// BadRequestErr is the handler for handling Bad Request Errors, caused by invalid user input
// or a malformed client requests.
func BadRequestErr(c echo.Context, sl *slog.Logger, uri string, err error) error {
	const msg = "bad request handler"
	if err1 := panics.Slog(c, sl); err1 != nil {
		return fmt.Errorf("%s: %w", msg, err1)
	}
	const code = http.StatusBadRequest
	if err != nil {
		sl.Error(msg, slog.Int("code", code), slog.String("uri", uri), slog.String("error", err.Error()))
	}
	if nilContext := c == nil; nilContext {
		const code = http.StatusInternalServerError
		sl.Error(msg, slog.Int("code", code), slog.String("tmpl", ErrTmpl.Error()), slog.String("context", ErrCxt.Error()))
		return echo.NewHTTPError(http.StatusInternalServerError,
			fmt.Errorf("%w: handler app status", ErrCxt))
	}
	data := empty(c)
	data["description"] = fmt.Sprintf("HTTP status %d error", code)
	data["title"] = "400 error, there is a complication"
	data["code"] = code
	data["logo"] = "Client error"
	data["alert"] = "Something went wrong, " + err.Error()
	data["probl"] = "It might be a settings or configuration problem or a legacy browser issue."
	data["uriErr"] = uri
	if err := c.Render(code, "status", data); err != nil {
		sl.Error(msg, slog.Int("code", code), slog.String("uri", uri), slog.String("error", err.Error()))
		return echo.NewHTTPError(http.StatusInternalServerError, ErrTmpl)
	}
	return nil
}

// DatabaseErr is the handler for database connection issues.
// A HTTP 503 Service Unavailable error is returned, to reflect the database
// connection issue but where the server is still running and usable for the client.
func DatabaseErr(c echo.Context, uri string, err error) error {
	const unavailable = http.StatusServiceUnavailable
	logger := zaplog.Debug()
	if err != nil {
		logger.Error(fmt.Sprintf("%d database error for the URL, %q: %s", unavailable, uri, err))
	}
	if nilContext := c == nil; nilContext {
		logger.Warn(fmt.Sprintf("%s: %s", ErrTmpl, ErrCxt))
		return echo.NewHTTPError(http.StatusInternalServerError,
			fmt.Errorf("%w: handler app status", ErrCxt))
	}
	data := empty(c)
	data["description"] = fmt.Sprintf("HTTP status %d error", unavailable)
	data["title"] = fmt.Sprintf("%d error, there is a complication", unavailable)
	data["code"] = fmt.Sprintf("%d service unavailable", unavailable)
	data["logo"] = "Database error"
	data["alert"] = "Cannot connect to the database!"
	data["uriErr"] = ""
	data["probl"] = "This is not your fault, but the server cannot communicate with the database to display this page."
	if err := c.Render(unavailable, "status", data); err != nil {
		logger.Error(fmt.Sprintf("%d database render error for the URL, %q: %s", unavailable, uri, err))
		return echo.NewHTTPError(unavailable, ErrTmpl)
	}
	return nil
}

// DownloadErr is the handler for missing download files and database ID errors.
func DownloadErr(c echo.Context, uri string, err error) error {
	const code = http.StatusNotFound
	id := c.Param("id")
	logger := zaplog.Debug()
	if err != nil {
		logger.Error(fmt.Sprintf("%d download error for the URL, %q: %s", code, id, err))
	}
	if nilContext := c == nil; nilContext {
		logger.Error(fmt.Sprintf("%s: %s", ErrTmpl, ErrCxt))
		return echo.NewHTTPError(http.StatusInternalServerError,
			fmt.Errorf("%w: handler app status", ErrCxt))
	}
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
		logger.Error(fmt.Sprintf("%d download render error for the URL, %q: %s", code, id, err))
		return echo.NewHTTPError(http.StatusInternalServerError, ErrTmpl)
	}
	return nil
}

// FileMissingErr is the handler for missing download files and database ID errors.
func FileMissingErr(c echo.Context, uri string, err error) error {
	const code = http.StatusServiceUnavailable
	id := c.Param("id")
	logger := zaplog.Debug()
	if err != nil {
		logger.Error(fmt.Sprintf("%d file missing error for the URL, %q: %s", code, id, err))
	}
	if nilContext := c == nil; nilContext {
		logger.Error(fmt.Sprintf("%s: %s", ErrTmpl, ErrCxt))
		return echo.NewHTTPError(http.StatusInternalServerError,
			fmt.Errorf("%w: handler app status", ErrCxt))
	}
	data := empty(c)
	data["description"] = fmt.Sprintf("HTTP status %d error", code)
	data["title"] = "503 download unavailable"
	data["code"] = code
	data["logo"] = "Download unavailable"
	data["alert"] = "Cannot send you this download"
	data["probl"] = "The file download needs to be added to the server; " +
		"otherwise, there may be a problem with the server configuration, or the file may be lost."
	data["uriErr"] = strings.Join([]string{uri, id}, "/")
	if err := c.Render(code, "status", data); err != nil {
		logger.Error(fmt.Sprintf("%d file missing render error for the URL, %q: %s", code, id, err))
		return echo.NewHTTPError(http.StatusInternalServerError, ErrTmpl)
	}
	return nil
}

// ForbiddenErr is the handler for handling Forbidden Errors, caused by clients requesting
// pages that they do not have permission to access.
func ForbiddenErr(c echo.Context, uri string, err error) error {
	const code = http.StatusForbidden
	logger := zaplog.Debug()
	if err != nil {
		logger.Error(fmt.Sprintf("%d forbidden error for the URL, %q: %s", code, uri, err))
	}
	if nilContext := c == nil; nilContext {
		logger.Error(fmt.Sprintf("%s: %s", ErrTmpl, ErrCxt))
		return echo.NewHTTPError(http.StatusInternalServerError,
			fmt.Errorf("%w: handler app status", ErrCxt))
	}
	data := empty(c)
	data["description"] = fmt.Sprintf("HTTP status %d error", code)
	data["title"] = "403, forbidden"
	data["code"] = code
	data["logo"] = "Forbidden"
	data["alert"] = "This page is locked"
	if err != nil {
		data["probl"] = fmt.Sprintf("This page is not intended for the general public, %s.", err.Error())
	}
	data["uriErr"] = uri
	if err := c.Render(code, "status", data); err != nil {
		logger.Info(fmt.Sprintf("%d forbidden render error for the URL, %q: %s", code, uri, err))
		return echo.NewHTTPError(http.StatusInternalServerError, ErrTmpl)
	}
	return nil
}

// InternalErr is the handler for handling Internal Server Errors, caused by programming bugs or crashes.
// The uri string is the part of the URL that caused the error.
// The optional error value is logged using the zap sugared logger.
// If the echo context is nil then a user hostile, fallback error in raw text is returned.
func InternalErr(c echo.Context, uri string, err error) error {
	if errors.Is(err, syscall.EPIPE) {
		// This is a common error when the client disconnects before the response is sent,
		// and commonly happens when using developer hot reloading.
		_, _ = fmt.Fprintf(io.Discard, "nothing to render due to the \"write: broken pipe\" error\n")
		return nil
	}
	const code = http.StatusInternalServerError
	logger := zaplog.Debug()
	if err != nil {
		logger.Error(fmt.Sprintf("%d internal error for the URL, %q: %s", code, uri, err))
	}
	if errors.Is(err, echo.ErrRendererNotRegistered) {
		return echo.NewHTTPError(code, err)
	}
	if nilContext := c == nil; nilContext {
		logger.Error(fmt.Sprintf("%s: %s", ErrTmpl, ErrCxt))
		return echo.NewHTTPError(http.StatusInternalServerError,
			fmt.Errorf("%w: handler app status", ErrCxt))
	}
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
		logger.Error(fmt.Sprintf("%d internal render error for the URL, %q: %s", code, uri, ErrTmpl))
		return echo.NewHTTPError(http.StatusInternalServerError, ErrTmpl)
	}
	return nil
}

// StatusErr is the handler for the HTTP status pages such as the 404 - not found.
// If the zap logger is nil then the error page is returned but no error is logged.
// If the echo context is nil then a user hostile, fallback error in raw text is returned.
func StatusErr(c echo.Context, code int, uri string) error {
	logger := zaplog.Debug()
	if nilContext := c == nil; nilContext {
		logger.Error(fmt.Sprintf("%s: %s", ErrTmpl, ErrCxt))
		return echo.NewHTTPError(http.StatusInternalServerError,
			fmt.Errorf("%w: handler app status", ErrCxt))
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
		return InternalErr(c, uri, nil)
	default:
		s := http.StatusText(code)
		if s == "" {
			err := fmt.Errorf("%d status error for the URL, %s: %w", code, uri, ErrCode)
			logger.Error(err.Error())
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
		logger.Error(fmt.Sprintf("%d status code render error for the URL, %s: %s", code, uri, ErrCode))
		return echo.NewHTTPError(http.StatusInternalServerError, ErrTmpl)
	}
	return nil
}
