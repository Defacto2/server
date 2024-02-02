package config

// Package file logger.go contains the logging middleware for the web application.

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Defacto2/server/handler/html3"
	"github.com/Defacto2/server/internal/helper"
	"github.com/Defacto2/server/internal/logger"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

var (
	ErrEchoNil     = errors.New("echo instance is nil")
	ErrNotDir      = errors.New("directory path points to the file")
	ErrDirNotExist = errors.New("directory does not exist or incorrectly typed")
	ErrTouch       = errors.New("the server cannot create a file in the directory")
	ErrLog         = errors.New("the server cannot log to files")
)

// https://github.com/labstack/echo/discussions/1820

// LoggerMiddleware handles the logging of HTTP servers.
func (c Config) LoggerMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	// Logger
	var z *zap.SugaredLogger
	switch c.ProductionMode {
	case true:
		z = logger.Production(c.LogDir).Sugar()
	default:
		z = logger.Development().Sugar()
	}
	defer func() {
		_ = z.Sync()
	}()
	return func(e echo.Context) error {
		timeStarted := time.Now()
		err := next(e)
		status := e.Response().Status
		httpErr := new(echo.HTTPError)
		if errors.As(err, &httpErr) {
			status = httpErr.Code
		}

		v := map[string]interface{}{
			"latency": int64(time.Since(timeStarted) / time.Millisecond),
			"method":  e.Request().Method,
			"path":    e.Request().URL.Path,
			"status":  status,
		}
		if c.LogRequests || err != nil {
			s := fmt.Sprintf("HTTP %s %d: %s", v["method"], v["status"], v["path"])
			if err != nil {
				s += fmt.Sprintf("  info: %s", err)
			}
			switch status {
			case http.StatusOK:
				z.Debug(s)
			default:
				z.Warn(s)
			}
		}
		switch status {
		case http.StatusOK:
			return nil
		default:
			if err != nil {
				// This error MUST be returned otherwise the client will always receive a 200 OK status
				return err
			}
			return nil
		}
	}
}

// LogStorage determines the local storage path for all log files created by this web application.
func (c *Config) LogStorage() error {
	const ownerGroupAll = 0o770
	logs := c.LogDir
	if logs == "" {
		dir, err := os.UserConfigDir()
		if err != nil {
			return err
		}
		logs = filepath.Join(dir, ConfigDir)
	}
	if ok := helper.IsStat(logs); !ok {
		if err := os.MkdirAll(logs, ownerGroupAll); err != nil {
			return fmt.Errorf("%w: %s", err, logs)
		}
	}
	c.LogDir = logs
	return nil
}

// CustomErrorHandler handles customer error templates.
func (c Config) CustomErrorHandler(err error, e echo.Context) {
	var z *zap.SugaredLogger
	switch c.ProductionMode {
	case true:
		z = logger.Production(c.LogDir).Sugar()
	default:
		z = logger.Development().Sugar()
	}
	defer func() {
		_ = z.Sync()
	}()
	switch {
	case IsHTML3(e.Path()):
		if err := html3.Error(e, err); err != nil {
			z.DPanic("Custom HTML3 response handler broke: %s", err)
		}
		return
	default:
		code := http.StatusInternalServerError
		var httpError *echo.HTTPError
		if errors.As(err, &httpError) {
			code = httpError.Code
		}
		errorPage := fmt.Sprintf("%d.html", code)
		if err := e.File(errorPage); err != nil {
			// fallback to a string error if templates break
			c, s, err1 := StringErr(err)
			if err1 != nil {
				z.DPanic("Custom response handler broke: %s", err1)
			}
			if err2 := e.String(c, s); err2 != nil {
				z.DPanic("Custom response handler broke: %s", err2)
			}
		}
		return
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
	// return c.String(code, fmt.Sprintf("%d - %s", code, msg))
}

// IsHTML3 returns true if the route is /html3.
func IsHTML3(path string) bool {
	splitPaths := func(r rune) bool {
		return r == '/'
	}
	rel := strings.FieldsFunc(path, splitPaths)
	return len(rel) > 0 && rel[0] == "html3"
}
