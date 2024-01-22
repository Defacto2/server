package config

// Package file logger.go contains the logging middleware for the web application.

import (
	"errors"
	"fmt"
	"log"
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
	ErrNotDir      = errors.New("not a directory")
	ErrDirNotExist = errors.New("directory does not exist or incorrectly typed")
)

// https://github.com/labstack/echo/discussions/1820

// LoggerMiddleware handles the logging of HTTP servers.
func (cfg Config) LoggerMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	// Logger
	var z *zap.SugaredLogger
	switch cfg.IsProduction {
	case true:
		z = logger.Production(cfg.LogDir).Sugar()
	default:
		z = logger.Development().Sugar()
	}
	defer func() {
		if err := z.Sync(); err != nil {
			log.Printf("zap logger sync error: %s", err)
		}
	}()
	return func(c echo.Context) error {
		timeStarted := time.Now()
		err := next(c)
		status := c.Response().Status
		httpErr := new(echo.HTTPError)
		if errors.As(err, &httpErr) {
			status = httpErr.Code
		}

		v := map[string]interface{}{
			"latency": int64(time.Since(timeStarted) / time.Millisecond),
			"method":  c.Request().Method,
			"path":    c.Request().URL.Path,
			"status":  status,
		}
		if cfg.LogRequests || err != nil {
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
func (cfg *Config) LogStorage() error {
	const ownerGroupAll = 0o770
	logs := cfg.LogDir
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
	cfg.LogDir = logs
	return nil
}

// CustomErrorHandler handles customer error templates.
func (cfg Config) CustomErrorHandler(err error, c echo.Context) {
	var z *zap.SugaredLogger
	switch cfg.IsProduction {
	case true:
		z = logger.Production(cfg.LogDir).Sugar()
	default:
		z = logger.Development().Sugar()
	}
	defer func() {
		if err := z.Sync(); err != nil {
			log.Printf("zap logger sync error: %s", err)
		}
	}()
	switch {
	case IsHTML3(c.Path()):
		if err := html3.Error(c, err); err != nil {
			z.DPanic("Custom HTML3 response handler broke: %s", err)
		}
		return
	default:
		code := http.StatusInternalServerError
		if he, ok := err.(*echo.HTTPError); ok {
			code = he.Code
		}
		errorPage := fmt.Sprintf("%d.html", code)
		if err := c.File(errorPage); err != nil {
			// fallback to a string error if templates break
			if err1 := StringErr(err, c); err1 != nil {
				z.DPanic("Custom response handler broke: %s", err1)
			}
		}
		return
	}
}

// StringErr sends the error and code as a string.
func StringErr(err error, c echo.Context) error {
	code, msg := http.StatusInternalServerError, "internal server error"
	if he, ok := err.(*echo.HTTPError); ok {
		code = he.Code
		msg = fmt.Sprint(he.Message)
	}
	return c.String(code, fmt.Sprintf("%d - %s", code, msg))
}

// IsHTML3 returns true if the route is /html3.
func IsHTML3(path string) bool {
	splitPaths := func(r rune) bool {
		return r == '/'
	}
	rel := strings.FieldsFunc(path, splitPaths)
	return len(rel) > 0 && rel[0] == "html3"
}
