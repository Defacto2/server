package config

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Defacto2/server/logger"
	"github.com/Defacto2/server/router/html3"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
)

// https://echo.labstack.com/middleware/logger/#configuration

// NOTE: these current already return a 200 status code.
// https://github.com/labstack/echo/issues/2310

// LoggerDeveloper prints both requests and request errors to the console.
var LoggerDeveloper = middleware.RequestLoggerConfig{
	LogStatus:   true,
	LogURI:      true,
	LogError:    true,
	HandleError: true, // forwards error to the global error handler, so it can decide appropriate status code
	LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
		if v.Error == nil {
			fmt.Printf("REQUEST: uri: %v, status: %v\n", v.URI, v.Status)
		} else {
			fmt.Printf("REQUEST_ERROR: uri: %v, status: %v, err: %v\n", v.URI, v.Status, v.Error)
		}
		return nil
	},
}

func LoggerProduction(log *zap.SugaredLogger) middleware.RequestLoggerConfig {
	return middleware.RequestLoggerConfig{
		LogURI:      true,
		LogStatus:   true,
		LogError:    true,
		HandleError: true, // forwards error to the global error handler, so it can decide appropriate status code
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			if v.Error == nil {
				log.Info("request",
					zap.String("URI", v.URI),
					zap.Int("status", v.Status),
				)
			} else {
				log.Error("request error",
					zap.String("URI", v.URI),
					zap.Int("status", v.Status),
					zap.Error(v.Error),
				)
			}
			return nil
		}}
}

// https://github.com/labstack/echo/discussions/1820

var CustomMiddleware = func(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		timeStarted := time.Now()
		err := next(c)

		status := c.Response().Status

		httpErr := new(echo.HTTPError)
		if errors.As(err, &httpErr) {
			status = httpErr.Code
		}

		fields := map[string]interface{}{
			"latency": int64(time.Since(timeStarted) / time.Millisecond),
			"method":  c.Request().Method,
			"path":    c.Request().URL.Path,
			"query":   c.Request().URL.RawQuery,
			"status":  status,
		}

		if err != nil {
			fmt.Printf("on error: %v\n", fields)
			//s.logger.SendErrorWithFields(err, fields)
			return err
		}
		fmt.Printf("fields: %v\n", fields)
		//s.logger.SendWithFields(fields)
		return nil
	}
}

func (c Config) LoggerMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	// Logger
	var log *zap.SugaredLogger
	switch c.IsProduction {
	case true:
		log = logger.Production().Sugar()
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
			"query":   c.Request().URL.RawQuery,
			"status":  status,
		}

		if err != nil {
			fmt.Printf("REQUEST_ERROR: uri: %v, status: %v, err: %v\n", v["path"], v["status"], err.Error())
			return err // THIS MUST BE returned otherwise 200 will always be sent to the client
		}
		fmt.Printf("REQUEST: uri: %v, status: %v\n", v["path"], v["status"])
		return nil
	}
}

// TODO:
func (x Config) CustomErrorHandler(err error, c echo.Context) {
	//c.String(500, "oh heck! "+err.Error())
	code := http.StatusInternalServerError
	if he, ok := err.(*echo.HTTPError); ok {
		code = he.Code
	}
	//c.Logger().Error(err) // TODO: use zap or custom logger
	errorPage := fmt.Sprintf("%d.html", code)
	if err := c.File(errorPage); err != nil {
		//c.Logger().Error(err)
		fmt.Println(err)
	}
	splitPaths := func(r rune) bool {
		return r == '/'
	}
	rel := strings.FieldsFunc(c.Path(), splitPaths)
	html3Route := len(rel) > 0 && rel[0] == "html3"
	switch {
	case html3Route:
		if err := html3.Error(err, c); err != nil {
			panic(err) // TODO: logger?
		}
		return
	default:
		if err := X(err, c); err != nil {
			panic(err) // TODO: logger?
		}
		return
	}

	//c.Logger().Error(err)

	// errorPage := fmt.Sprintf("%d.html", code)
	// if err := c.File(errorPage); err != nil {
	// 	c.Logger().Error(err)
	// }
}

func X(err error, c echo.Context) error {
	code, msg := http.StatusInternalServerError, "internal server error"
	if he, ok := err.(*echo.HTTPError); ok {
		code = he.Code
		msg = fmt.Sprint(he.Message)
	}
	return c.String(code, fmt.Sprintf("%d - %s!", code, msg))
}
