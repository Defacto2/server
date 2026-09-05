//nolint:gochecknoglobals,exhaustruct_v5
package handler

// Package file middleware.go contains the custom middleware functions for the Echo web framework.

import (
	"crypto/sha512"
	"fmt"
	"log/slog"
	"net/http"
	"runtime"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/Defacto2/server/handler/app"
	"github.com/Defacto2/server/handler/sess"
	"github.com/Defacto2/server/internal/nils"
	"github.com/dustin/go-humanize"
	"github.com/labstack/echo-contrib/v5/session"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

var requestCounter atomic.Int64

const (
	CacheControl  = "Cache-Control"
	XApiVersion   = "X-Api-Version"
	XRobotsTag    = "X-Robots-Tag"
	XReadOnlyLock = "X-Read-Only-Lock"
	XResponseTime = "X-Response-Time"
)

// SkipPaths are parent route paths that should not be logged,
// to reduce the logging output. Otherwise every image
// or required resource for every page request would be returned.
func SkipPaths(c *echo.Context) bool {
	_, status := echo.ResolveResponseStatus(c.Response(), nil)

	if redirect := status == http.StatusMovedPermanently; redirect {
		return true
	}

	uri := c.Request().RequestURI
	statusOk := status == http.StatusOK
	switch {
	case strings.HasPrefix(uri, "/public/"),
		strings.HasPrefix(uri, "/css/"),
		strings.HasPrefix(uri, "/js/"),
		strings.HasPrefix(uri, "/image/"),
		strings.HasPrefix(uri, "/svg/"),
		strings.HasPrefix(uri, "/font/"):
		if statusOk {
			return true
		}
	}

	return false
}

// NoCrawl middleware adds a `X-Robots-Tag` header to the response.
// The header contains the noindex and nofollow values that tell search engine
// crawlers to not index or crawl the page or asset.
// See https://developers.google.com/search/docs/crawling-indexing/robots-meta-tag#xrobotstag
func (serv *Server) NoCrawl(next echo.HandlerFunc) echo.HandlerFunc {
	if !serv.Environment.NoCrawl {
		return next
	}
	return func(c *echo.Context) error {
		c.Response().Header().Set(XRobotsTag, "none")
		return next(c)
	}
}

// ReadOnlyLock disables all PATCH, POST, PUT and DELETE requests for the modification
// of the database and any related user interface.
func (serv *Server) ReadOnlyLock(next echo.HandlerFunc, sl *slog.Logger) echo.HandlerFunc {
	const format = "middleware read only lock %s: %w"
	if err := nils.Check(next, sl); err != nil {
		panic(fmt.Errorf(format, "check", err))
	}
	return func(c *echo.Context) error {
		value := strconv.FormatBool(bool(serv.Environment.ReadOnly))
		c.Response().Header().Set(XReadOnlyLock, value)

		if serv.Environment.ReadOnly {
			if err := app.StatusErr(sl, c, http.StatusForbidden, ""); err != nil {
				return fmt.Errorf(format, "status", err)
			}
			return nil // do not run next(e)
		}

		return next(c)
	}
}

// SessionLock middleware checks the session cookie for a valid signed in client.
func (serv *Server) SessionLock(next echo.HandlerFunc, sl *slog.Logger) echo.HandlerFunc {
	const format = "middleware session lock %s: %w"
	if err := nils.Check(next, sl); err != nil {
		panic(fmt.Errorf(format, "check", err))
	}

	return func(c *echo.Context) error {
		const code = http.StatusForbidden

		// help: https://pkg.go.dev/github.com/gorilla/sessions#Session
		sess, err := session.Get(sess.Name, c)
		if err != nil {
			sl.Warn("get session lock", slog.Any("error", err))
			if err := app.StatusErr(sl, c, code, ""); err != nil {
				return fmt.Errorf(format, "get", err)
			}
			return nil
		}

		id, ok := sess.Values["sub"].(string)
		if !ok || id == "" {
			if err := app.StatusErr(sl, c, code, ""); err != nil {
				return fmt.Errorf(format, "subexists forbid", err)
			}
			return nil
		}

		for _, account := range serv.Environment.GoogleAccounts {
			sum := sha512.Sum384([]byte(id))
			check := sum == account
			if check {
				return next(c)
			}
		}

		if err := app.StatusErr(sl, c, code, ""); err != nil {
			return fmt.Errorf(format, "check forbid", err)
		}
		return nil
	}
}

// configTrailSlash return the TrailingSlash middleware configuration.
func configTrailSlash() middleware.RemoveTrailingSlashConfig {
	return middleware.RemoveTrailingSlashConfig{
		RedirectCode: http.StatusMovedPermanently,
		Skipper:      nil,
	}
}

// RequestLoggerConfig handles logging for HTTP page requests.
// A slog Logger is required otherwise it will panic.
//
// If Configuration.LogAll is false then this returns a nil.
// Otherwise it logs all web server HTTP requests to info logs.
func (serv *Server) RequestLoggerConfig(sl *slog.Logger) middleware.RequestLoggerConfig {
	if !serv.Environment.LogAll {
		exitRequest := func(_ *echo.Context, _ middleware.RequestLoggerValues) error {
			return nil
		}
		return middleware.RequestLoggerConfig{
			LogValuesFunc:  exitRequest,
			Skipper:        nil,
			BeforeNextFunc: nil,
			HandleError:    true,
		}
	}

	const format = "request logger config handler %s: %w"
	if err := nils.Check(sl); err != nil {
		panic(fmt.Errorf(format, "check", err))
	}

	// logValues is used by the returned middleware.RequestLoggerConfig().LogValuesFunc
	logValues := func(_ *echo.Context, v middleware.RequestLoggerValues) error {
		// memory usage - but only sample every 10th request
		var alloc string
		count := requestCounter.Add(1)
		if count%10 == 0 {
			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			alloc = humanize.Bytes(m.Alloc)
		}

		memory := func() slog.Attr {
			if alloc == "" {
				return slog.Attr{}
			}
			return slog.String("allocation", alloc)
		}

		rsize := uint64(v.ResponseSize) //nolint:gosec // G115: ResponseSize is always non-negative (content length)
		response := func() slog.Attr {
			return slog.Group("response", slog.Int64("size", v.ResponseSize),
				slog.String("humanize", humanize.Bytes(rsize)))
		}

		cpuinfo := func() slog.Attr {
			return slog.Group("cpu", slog.Int("cores", runtime.NumCPU()),
				slog.Int("go_routines", runtime.NumGoroutine()))
		}

		latency := func() slog.Attr {
			return slog.Duration("latency", v.Latency)
		}

		// NOTE: Using using any new v middleware.RequestLoggerValues will require an update
		// to middleware.RequestLoggerConfig values that are returned by this func.
		requests := func() slog.Attr {
			return slog.Group("request",
				slog.String("agent", v.UserAgent), // browser agent used for debugging
				slog.String("path", v.URIPath),    // uri path without any params
				slog.String("route", v.RoutePath), // internal route path with values
				slog.String("uri", v.URI),         // complete url request
			)
		}

		// slog.Any("request", v), // add for verbose & debugging
		msg := "HTTP(S) " + v.Method + " " + strconv.Itoa(v.Status)
		sl.Info(msg, latency(), response(), cpuinfo(), memory(), requests())
		return nil
	}

	return middleware.RequestLoggerConfig{
		Skipper:          SkipPaths,
		LogLatency:       true,
		LogProtocol:      false,
		LogRemoteIP:      false,
		LogHost:          false,
		LogMethod:        true,
		LogURI:           true,
		LogURIPath:       true,
		LogRoutePath:     true,
		LogRequestID:     false,
		LogReferer:       false,
		LogUserAgent:     true,
		LogStatus:        true,
		LogContentLength: false,
		LogResponseSize:  true,
		LogHeaders:       nil,
		LogQueryParams:   nil,
		LogFormValues:    nil,
		LogValuesFunc:    logValues,
		// LogError:         false,
	}
}

// CacheMiddleware sets appropriate Cache-Control headers for API responses.
func CacheMiddleware() echo.MiddlewareFunc {
	const (
		age5min    = "300"
		age30min   = "1800"
		age1hour   = "3600"
		age24hours = "86400"
	)
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			// set Cache-Control header
			const maxAge = "public, max-age="
			path := c.Request().URL.Path
			switch {
			case
				strings.Contains(path, "/categories"),
				strings.Contains(path, "/platforms"):
				c.Response().Header().Set(CacheControl, maxAge+age24hours)
			case
				strings.Contains(path, "/artifacts"),
				strings.Contains(path, "/artifacts/new"):
				c.Response().Header().Set(CacheControl, maxAge+age5min)
			case
				strings.Contains(path, "/artifact/"):
				c.Response().Header().Set(CacheControl, maxAge+age1hour)
			case
				strings.Contains(path, "/releaser/"),
				strings.Contains(path, "/scener/"):
				c.Response().Header().Set(CacheControl, maxAge+age30min)
			case
				strings.Contains(path, "/groups"),
				strings.Contains(path, "/magazines"),
				strings.Contains(path, "/boards"),
				strings.Contains(path, "/sites"):
				c.Response().Header().Set(CacheControl, maxAge+age1hour)
			case
				strings.Contains(path, "/milestones"),
				strings.Contains(path, "/areacodes"),
				strings.Contains(path, "/websites"),
				strings.Contains(path, "/demozoo"):
				c.Response().Header().Set(CacheControl, maxAge+age24hours)
			default:
				c.Response().Header().Set(CacheControl, maxAge+age5min)
			}

			return next(c)
		}
	}
}

func APIMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c *echo.Context) error {
		start := time.Now()

		c.Response().Header().Set(XApiVersion, app.APIVer)
		// use a custom response writer to capture the timing
		resp, err := echo.UnwrapResponse(c.Response())
		if err != nil {
			const format = "api unwrap response: %w"
			return fmt.Errorf(format, err)
		}

		resp.Before(func() {
			const thousand = 1000.0
			end := time.Since(start)
			ms := float64(end.Microseconds()) / thousand
			value := fmt.Sprintf("%.3fms", ms)
			resp.Header().Set(XResponseTime, value)
		})

		return next(c)
	}
}
