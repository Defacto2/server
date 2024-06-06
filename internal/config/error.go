package config

// Package file error.go contains the custom error middleware for the web application.

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/Defacto2/server/handler/html3"
	"github.com/Defacto2/server/internal/zaplog"
	"github.com/labstack/echo/v4"
)

var (
	ErrDirNotExist = errors.New("directory does not exist or incorrectly typed")
	ErrEchoNil     = errors.New("echo instance is nil")
	ErrLog         = errors.New("the server cannot log to files")
	ErrNotDir      = errors.New("directory path points to the file")
	ErrTouch       = errors.New("the server cannot create a file in the directory")
)

// CustomErrorHandler handles customer error templates.
func (c Config) CustomErrorHandler(err error, ctx echo.Context) {
	logger := zaplog.Development().Sugar()
	if c.ProdMode {
		root := c.AbsLog
		logger = zaplog.Production(root).Sugar()
	}
	defer func() {
		_ = logger.Sync()
	}()
	if IsHTML3(ctx.Path()) {
		if err := html3.Error(ctx, err); err != nil {
			logger.DPanic("Custom HTML3 response handler broke: %s", err)
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
			logger.DPanic("Custom response handler broke: %s", err1)
		}
		if err2 := ctx.String(code, s); err2 != nil {
			logger.DPanic("Custom response handler broke: %s", err2)
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
