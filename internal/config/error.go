package config

// Package file error.go contains the custom error middleware for the web application.

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/Defacto2/server/handler/html3"
	"github.com/Defacto2/server/internal/logs"
	"github.com/Defacto2/server/internal/panics"
	"github.com/labstack/echo/v5"
)

const (
	PortMax = 65534 // PortMax is the highest valid port number.
	PortSys = 1024  // PortSys is the lowest valid port number that does not require system access.
)

var (
	ErrNoAccounts = errors.New("the production server has no google oauth2 user accounts to allow admin logins")
	ErrNoDir      = errors.New("directory does not exist or incorrectly typed")
	ErrNoOAuth2   = errors.New("production server requires a google, oauth2 client id to allow admin logins")
	ErrNoPort     = errors.New("server cannot start with a http or a tls port")
	ErrNoPath     = errors.New("empty path or name")
	ErrPSVersion  = errors.New("postgres did not return a version value")
	ErrTouch      = errors.New("server cannot create a file in the directory")
	ErrNotDir     = errors.New("path points to a file")
	ErrNotFile    = errors.New("path points to a directory")
)

// CustomErrorHandler handles edge case HTTP errors including
// issues such as missing template files, attempts at browsing
// restricted directories, etc.
//
// The returned result will always be a text only HTTP response,
// as there is no ability to access HTML rendered pages.
func CustomErrorHandler(sl *slog.Logger, c *echo.Context, err error) {
	const msg = "custom error handler"
	if err := panics.SC(c, sl); err != nil {
		panic(fmt.Errorf("%s: %w", msg, err))
	}
	if isV3(c.Path()) {
		if err := html3.Error(c, err); err != nil {
			logs.Fatal(sl, msg, slog.Any("html3 error", err))
		}
		return
	}

	statusCode := echo.StatusCode(err)
	if statusCode == 0 {
		statusCode = http.StatusInternalServerError
		if errors.Is(err, echo.ErrNotFound) {
			statusCode = http.StatusNotFound
		}
	}
	statusText := http.StatusText(statusCode)

	sl.Error(msg,
		slog.Any("error", err),
		slog.String("error type", fmt.Sprintf("%t", err)),
		slog.Int("code", statusCode))

	s := fmt.Sprintf("%d - %s", statusCode, statusText)
	if err1 := c.String(statusCode, s); err1 != nil {
		logs.Fatal(sl, msg, slog.Any("error", err1))
	}
}

// isV3 returns true if the path is /html3.
func isV3(path string) bool {
	return strings.Contains(path, "/html3/") || strings.HasSuffix(path, "/html3")
}
