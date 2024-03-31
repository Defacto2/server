// Package br provides experimental brotli support for the Echo web framework.
package br

import (
	"bufio"
	"io"
	"net"
	"net/http"
	"strings"

	"github.com/andybalholm/brotli"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

//
// Brotli compression scheme, copied from the submission the delaneyj commented on 22 Feb 2019.
// https://github.com/labstack/echo/blob/a327810ef8a5625797ca6a106b538e5abec3917e/middleware/compress_brotli.go
//

const (
	BrotliScheme = "br" // Brotli compression header scheme.
)

var ErrHijack = echo.NewHTTPError(http.StatusInternalServerError, "response could not be hijacked")

// Brotli returns a middleware which compresses HTTP response using brotli compression
// scheme.
func Brotli() echo.MiddlewareFunc {
	return BrotliWithConfig(DefaultBrotliConfig())
}

// BrotliWithConfig returns the [Brotli] middleware with config.
func BrotliWithConfig(config BrotliConfig) echo.MiddlewareFunc {
	if config.Skipper == nil {
		config.Skipper = DefaultBrotliConfig().Skipper
	}
	if config.Level == 0 {
		config.Level = DefaultBrotliConfig().Level
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if config.Skipper(c) {
				return next(c)
			}

			resp := c.Response()
			resp.Header().Add(echo.HeaderVary, echo.HeaderAcceptEncoding)
			if strings.Contains(c.Request().Header.Get(echo.HeaderAcceptEncoding), BrotliScheme) {
				resp.Header().Set(echo.HeaderContentEncoding, BrotliScheme) // Issue #806
				rw := resp.Writer
				w := brotli.NewWriterOptions(rw, brotli.WriterOptions{Quality: config.Level})
				defer func() {
					if resp.Size == 0 {
						if resp.Header().Get(echo.HeaderContentEncoding) == BrotliScheme {
							resp.Header().Del(echo.HeaderContentEncoding)
						}
						// We have to reset response to it's pristine state when
						// nothing is written to body or error is returned.
						// See issue #424, #407.
						resp.Writer = rw
						w.Reset(io.Discard)
					}
					w.Close()
				}()
				grw := &brotliResponseWriter{Writer: w, ResponseWriter: rw}
				resp.Writer = grw
			}
			return next(c)
		}
	}
}

// DefaultBrotliConfig is the default Brotli middleware config.
func DefaultBrotliConfig() BrotliConfig {
	return BrotliConfig{
		Skipper: middleware.DefaultSkipper,
		Level:   brotli.DefaultCompression,
	}
}

type (
	// BrotliConfig defines the config for Brotli middleware.
	BrotliConfig struct {
		// Skipper defines a function to skip middleware.
		Skipper middleware.Skipper

		// Brotli compression level.
		// Optional. Default value -1.
		Level int `yaml:"level"`
	}

	brotliResponseWriter struct {
		io.Writer
		http.ResponseWriter
	}
)

func (w *brotliResponseWriter) WriteHeader(code int) {
	if code == http.StatusNoContent { // Issue #489
		w.ResponseWriter.Header().Del(echo.HeaderContentEncoding)
	}
	w.Header().Del(echo.HeaderContentLength) // Issue #444
	w.ResponseWriter.WriteHeader(code)
}

func (w *brotliResponseWriter) Write(b []byte) (int, error) {
	if w.Header().Get(echo.HeaderContentType) == "" {
		w.Header().Set(echo.HeaderContentType, http.DetectContentType(b))
	}
	return w.Writer.Write(b)
}

func (w *brotliResponseWriter) Flush() {
	if writer, ok := w.Writer.(*brotli.Writer); ok {
		writer.Flush()
	}
	if flusher, ok := w.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

func (w *brotliResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hijacker, ok := w.ResponseWriter.(http.Hijacker); ok {
		return hijacker.Hijack()
	}
	return nil, nil, ErrHijack
}
