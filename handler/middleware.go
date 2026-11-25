//nolint:exhaustruct_v5
package handler

// Package file middleware.go contains the custom middleware functions for the Echo web framework.

import (
	"crypto/sha512"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Defacto2/server/handler/api"
	"github.com/Defacto2/server/handler/app"
	"github.com/Defacto2/server/handler/sess"
	"github.com/Defacto2/server/internal/nils"
	"github.com/dustin/go-humanize"
	"github.com/labstack/echo-contrib/v5/session"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

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
	return skipRequest(c.Request().RequestURI, status)
}

func skipRequest(uri string, status int) bool {
	if status == http.StatusMovedPermanently {
		return true
	}
	if status != http.StatusOK {
		return false
	}
	switch {
	case strings.HasPrefix(uri, "/public/"),
		strings.HasPrefix(uri, "/css/"),
		strings.HasPrefix(uri, "/js/"),
		strings.HasPrefix(uri, "/image/"),
		strings.HasPrefix(uri, "/svg/"),
		strings.HasPrefix(uri, "/font/"):
		return true
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
func (serv *Server) RequestLoggerConfig(sl *slog.Logger) middleware.RequestLoggerConfig {
	if !serv.Environment.LogAll {
		return middleware.RequestLoggerConfig{
			LogValuesFunc: func(_ *echo.Context, _ middleware.RequestLoggerValues) error {
				return nil
			},
			HandleError: true,
		}
	}

	const format = "request logger config handler %s: %w"
	if err := nils.Check(sl); err != nil {
		panic(fmt.Errorf(format, "check", err))
	}
	const size = 3

	logValues := func(_ *echo.Context, v middleware.RequestLoggerValues) error {
		if skipRequest(v.URI, v.Status) {
			return nil
		}

		const latencyUnit = 100 * time.Microsecond
		rsize := uint64(max(v.ResponseSize, 0))
		attrs := make([]any, 0, size)
		attrs = append(attrs,
			slog.Duration("latency", v.Latency.Round(latencyUnit)),
			slog.String("size", humanize.Bytes(rsize)),
		)
		if agent := ShortAgent(v.UserAgent); agent != "" {
			attrs = append(attrs, slog.String("agent", agent))
		}

		msg := fmt.Sprintf("%s %d %s", v.Method, v.Status, v.URI)
		switch {
		case v.Status >= http.StatusInternalServerError:
			sl.Error(msg, attrs...)
		case v.Status >= http.StatusBadRequest:
			sl.Warn(msg, attrs...)
		default:
			sl.Info(msg, attrs...)
		}
		return nil
	}

	return middleware.RequestLoggerConfig{
		LogLatency:      true,
		LogMethod:       true,
		LogURI:          true,
		LogUserAgent:    true,
		LogStatus:       true,
		LogResponseSize: true,
		LogValuesFunc:   logValues,
	}
}

// ShortAgent condenses a verbose HTTP User-Agent string into a compact
// "<client> (<os>)" label (e.g. "Chrome/158 (Linux)", "Safari/18 (iOS)", "curl/8.5").
func ShortAgent(ua string) string {
	ua = strings.TrimSpace(ua)
	if ua == "" {
		return ""
	}

	client := detectClient(ua)
	platform := detectPlatform(ua)
	if platform != "" {
		return client + " (" + platform + ")"
	}
	return client
}

func detectClient(ua string) string {
	lower := strings.ToLower(ua)
	bots := [...]struct{ needle, label string }{
		{"googlebot", "Googlebot"},
		{"bingbot", "Bingbot"},
		{"duckduckbot", "DuckDuckBot"},
		{"ahrefsbot", "AhrefsBot"},
		{"gptbot", "GPTBot"},
		{"claudebot", "ClaudeBot"},
	}
	for _, b := range bots {
		if strings.Contains(lower, b.needle) {
			return b.label
		}
	}

	cliTools := [...]struct{ prefix, label string }{
		{"curl/", "curl"},
		{"wget/", "wget"},
		{"go-http-client/", "Go-http"},
		{"python-requests/", "python-requests"},
	}
	const part2 = 2
	for _, tool := range cliTools {
		if ver := tokenVersion(ua, tool.prefix, part2); ver != "" {
			return tool.label + "/" + ver
		}
	}

	// Browser order matters: Edge/Opera include Chrome; Chrome includes Safari.
	browsers := [...]struct{ token, label string }{
		{"Edg/", "Edge"},
		{"OPR/", "Opera"},
		{"Firefox/", "Firefox"},
		{"Chrome/", "Chrome"},
	}
	for _, br := range browsers {
		if ver := tokenVersion(ua, br.token, 1); ver != "" {
			return br.label + "/" + ver
		}
	}
	if strings.Contains(ua, "Safari/") {
		if ver := tokenVersion(ua, "Version/", 1); ver != "" {
			return "Safari/" + ver
		}
		return "Safari"
	}

	if strings.Contains(lower, "bot") || strings.Contains(lower, "crawler") || strings.Contains(lower, "spider") {
		return "bot"
	}

	first, _, _ := strings.Cut(ua, " ")
	const maxLen = 24
	if len(first) > maxLen {
		return first[:maxLen]
	}
	return first
}

func detectPlatform(ua string) string {
	switch {
	case strings.Contains(ua, "iPhone") || strings.Contains(ua, "iPad") || strings.Contains(ua, "iPod"):
		return "iOS"
	case strings.Contains(ua, "Android"):
		return "Android"
	case strings.Contains(ua, "Windows"):
		return "Windows"
	case strings.Contains(ua, "Macintosh") || strings.Contains(ua, "Mac OS X"):
		return "macOS"
	case strings.Contains(ua, "CrOS"):
		return "ChromeOS"
	case strings.Contains(ua, "FreeBSD"):
		return "FreeBSD"
	case strings.Contains(ua, "Linux"):
		return "Linux"
	default:
		return ""
	}
}

func tokenVersion(ua, prefix string, parts int) string {
	idx := strings.Index(strings.ToLower(ua), strings.ToLower(prefix))
	if idx < 0 {
		return ""
	}
	rest := ua[idx+len(prefix):]
	ver, _, _ := strings.Cut(rest, " ")
	ver = strings.TrimRight(ver, ";,)")
	if ver == "" {
		return ""
	}
	segs := strings.Split(ver, ".")
	if len(segs) > parts {
		segs = segs[:parts]
	}
	return strings.Join(segs, ".")
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

		c.Response().Header().Set(XApiVersion, api.APIVer)
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
