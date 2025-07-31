package config

// Package file error.go contains the custom error middleware for the web application.

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/Defacto2/server/handler/html3"
	"github.com/Defacto2/server/internal/out"
	"github.com/Defacto2/server/internal/panics"
	"github.com/labstack/echo/v4"
)

const (
	PortMax = 65534 // PortMax is the highest valid port number.
	PortSys = 1024  // PortSys is the lowest valid port number that does not require system access.
)

var (
	ErrNoAccounts = errors.New("the production server has no google oauth2 user accounts to allow admin logins")
	ErrNoDir      = errors.New("directory does not exist or incorrectly typed")
	ErrNoOAuth2   = errors.New("production server requires a google, oauth2 client id to allow admin logins")
	ErrNoPort     = errors.New("server cannot start without a http or a tls port")
	ErrNoPath     = errors.New("empty path or name")
	ErrPSqlVer    = errors.New("postgres did not return a version value")
	ErrTouch      = errors.New("server cannot create a file in the directory")
	ErrNotDir     = errors.New("path points to a file")
	ErrNotFile    = errors.New("path points to a directory")
)

// CustomErrorHandler handles customer error templates.
func (c *Config) CustomErrorHandler(err error, ctx echo.Context, sl *slog.Logger) {
	const msg = "custom error handler"
	if err := panics.Slog(ctx, sl); err != nil {
		panic(fmt.Errorf("%s: %w", msg, err))
	}
	if IsHTML3(ctx.Path()) {
		if err := html3.Error(ctx, err); err != nil {
			out.Fatal(sl, msg, slog.Any("html3 error", err))
		}
		return
	}
	statusCode := http.StatusInternalServerError
	var httpError *echo.HTTPError
	if errors.As(err, &httpError) {
		statusCode = httpError.Code
	}
	errorPage := fmt.Sprintf("%d.html", statusCode)
	if err := ctx.File(errorPage); err != nil {
		// fallback to a string error if templates break
		code, s, err1 := StringErr(err)
		if err1 != nil {
			out.Fatal(sl, msg, slog.Any("custom response error", err))
		}
		if err2 := ctx.String(code, s); err2 != nil {
			out.Fatal(sl, msg, slog.Any("custom response error", err))
		}
	}
}

// StringErr sends the error and code as a string.
func StringErr(err error) (int, string, error) {
	if err == nil {
		return 0, "", nil
	}
	code, msg := http.StatusInternalServerError, "internal server error"
	var httpError *echo.HTTPError
	if errors.As(err, &httpError) {
		code = httpError.Code
		msg = fmt.Sprint(httpError.Message)
	}
	return code, fmt.Sprintf("%d - %s", code, msg), nil
}

// IsHTML3 returns true if the route is /html3.
func IsHTML3(path string) bool {
	splitPaths := func(r rune) bool {
		return r == '/'
	}
	rel := strings.FieldsFunc(path, splitPaths)
	return len(rel) > 0 && rel[0] == "html3"
}
