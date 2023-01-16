package config

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Defacto2/server/helpers"
	"github.com/Defacto2/server/logger"
	"github.com/Defacto2/server/router/html3"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

var ErrDirNotExist = errors.New("directory does not exist or incorrectly typed")

// https://github.com/labstack/echo/discussions/1820

// LoggerMiddleware handles the logging of HTTP servers.
func (cfg Config) LoggerMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	// Logger
	var log *zap.SugaredLogger
	switch cfg.IsProduction {
	case true:
		log = logger.Production(cfg.ConfigDir).Sugar()
		defer log.Sync()
	default:
		log = logger.Development().Sugar()
		defer log.Sync()
	}
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
			//"query":   c.Request().URL.RawQuery,
			"status": status,
		}
		switch status {
		case http.StatusOK:
			return nil
		}
		if err != nil {
			log.Debugf("HTTP %s %d: %s  info: %s", v["method"], v["status"], v["path"], err)
			// This error MUST be returned otherwise the client will always receive a 200 OK status
			return err
		}
		return nil
	}
}

// LogStorage determines the local storage path for all log files created by this web application.
func (cfg *Config) LogStorage() error {
	dir := cfg.ConfigDir
	if dir == "" {
		var err error
		dir, err = os.UserConfigDir()
		if err != nil {
			return err
		}
	}
	if ok := helpers.IsStat(dir); !ok {
		return fmt.Errorf("%w: %s", ErrDirNotExist, dir)
	}
	logs := filepath.Join(dir, "defacto2-webapp")
	if ok := helpers.IsStat(logs); !ok {
		if err := os.MkdirAll(logs, 0770); err != nil {
			return fmt.Errorf("%w: %s", err, logs)
		}
	}
	cfg.ConfigDir = logs
	return nil
	//return errors.New("hello oops")
}

// CustomErrorHandler handles customer error templates.
func (cfg Config) CustomErrorHandler(err error, c echo.Context) {
	var log *zap.SugaredLogger
	switch cfg.IsProduction {
	case true:
		log = logger.Production(cfg.ConfigDir).Sugar()
		defer log.Sync()
	default:
		log = logger.Development().Sugar()
		defer log.Sync()
	}

	switch {
	case IsHTML3(c.Path()):
		if err := html3.Error(err, c); err != nil {
			log.DPanic("Custom HTML3 response handler broke: %s", err)
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
				log.DPanic("Custom response handler broke: %s", err1)
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
