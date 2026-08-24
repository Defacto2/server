package config

// Package file error.go contains the custom error middleware for the web application.

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/Defacto2/server/handler/html3"
	"github.com/Defacto2/server/internal/nils"
	"github.com/labstack/echo/v5"
)

const (
	PortMax = 65534 // PortMax is the highest valid port number.
	PortSys = 1024  // PortSys is the lowest valid port number that does not require system access.
)

var (
	ErrNoAccounts = errors.New("config: production server has no google oauth2 user accounts to allow admin logins")
	ErrNoOAuth2   = errors.New("config: production server requires a google, oauth2 client id to allow admin logins")
	ErrSession    = errors.New("config: production server has user sessions that never expire")
	ErrNoDir      = errors.New("config: directory does not exist on the host file system")
	ErrTouch      = errors.New("config: directory does not permit file writing")
	ErrNotDir     = errors.New("config: directory points to a file")
	ErrNotFile    = errors.New("config: file path points to a directory")
	ErrNoPort     = errors.New("config: no port, server cannot start without a configured http or a tls port")
	ErrNoPath     = errors.New("config: no path, path or name cannot be empty")
	ErrFormat     = errors.New("config: unsupported format")
	ErrPSVersion  = errors.New("postgres did not return a version value")
)

// CustomErrorHandler handles edge case HTTP errors including
// issues such as missing template files, attempts at browsing
// restricted directories, etc.
//
// The returned result will always be a text only HTTP response,
// as there is no ability to access HTML rendered pages.
func CustomErrorHandler(ctx context.Context, sl *slog.Logger, c *echo.Context, err error) {
	const msg = "custom error handler"
	if err := nils.Check(ctx, sl, c); err != nil {
		panic(fmt.Errorf("%s: %w", msg, err))
	}

	path := c.Path()
	if strings.HasPrefix(path, "/html3") {
		if err := html3.Error(c, err); err != nil {
			sl.Error(msg, slog.String("context", "html3 error delivery failed"), slog.Any("error", err))
		}
		return
	}

	code := echo.StatusCode(err)
	if code == 0 {
		code = http.StatusInternalServerError
		if errors.Is(err, echo.ErrNotFound) {
			code = http.StatusNotFound
		}
	}
	statusText := http.StatusText(code)

	sl.Error(msg, slog.Any("error", err), slog.String("type", fmt.Sprintf("%T", err)),
		slog.Int("status_code", code))

	s := fmt.Sprintf("%d - %s", code, statusText)
	if sErr := c.String(code, s); sErr != nil {
		sl.Error(msg, slog.String("context", "failed to write error response"), slog.Any("error", err))
	}
}
