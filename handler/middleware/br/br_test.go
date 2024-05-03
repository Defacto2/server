package br_test

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/Defacto2/server/handler/middleware/br"
	"github.com/andybalholm/brotli"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBrotli(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	// Skip if no Accept-Encoding header
	h := br.Brotli()(func(c echo.Context) error {
		_, _ = c.Response().Write([]byte("test")) // For Content-Type sniffing
		return nil
	})
	err := h(c)
	require.NoError(t, err)
	assert := assert.New(t)
	assert.Equal("test", rec.Body.String())
	// Brotli
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderAcceptEncoding, br.BrotliScheme)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	err = h(c)
	require.NoError(t, err)
	assert.Equal(br.BrotliScheme, rec.Header().Get(echo.HeaderContentEncoding))
	assert.Contains(rec.Header().Get(echo.HeaderContentType), echo.MIMETextPlain)
	r := brotli.NewReader(rec.Body)
	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(r)
	require.NoError(t, err)
	assert.Equal("test", buf.String())
	chunkBuf := make([]byte, 5)
	// Brotli chunked
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderAcceptEncoding, br.BrotliScheme)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	err = br.Brotli()(func(c echo.Context) error {
		c.Response().Header().Set("Content-Type", "text/event-stream")
		c.Response().Header().Set("Transfer-Encoding", "chunked")
		// Write and flush the first part of the data
		_, err = c.Response().Write([]byte("test\n"))
		require.NoError(t, err)
		c.Response().Flush()
		// Read the first part of the data
		assert.True(rec.Flushed)
		assert.Equal(br.BrotliScheme, rec.Header().Get(echo.HeaderContentEncoding))
		err := r.Reset(rec.Body)
		require.NoError(t, err)
		_, err = io.ReadFull(r, chunkBuf)
		require.NoError(t, err)
		assert.Equal("test\n", string(chunkBuf))
		// Write and flush the second part of the data
		_, err = c.Response().Write([]byte("test\n"))
		require.NoError(t, err)
		c.Response().Flush()
		_, err = io.ReadFull(r, chunkBuf)
		require.NoError(t, err)
		assert.Equal("test\n", string(chunkBuf))
		// Write the final part of the data and return
		_, err = c.Response().Write([]byte("test"))
		require.NoError(t, err)
		return nil
	})(c)
	require.NoError(t, err)
	buf = new(bytes.Buffer)
	_, err = buf.ReadFrom(r)
	require.NoError(t, err)
	assert.Equal("test", buf.String())
}

func TestBrotliNoContent(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderAcceptEncoding, br.BrotliScheme)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	h := br.Brotli()(func(c echo.Context) error {
		return c.NoContent(http.StatusNoContent)
	})
	if assert.NoError(t, h(c)) {
		assert.Empty(t, rec.Header().Get(echo.HeaderContentEncoding))
		assert.Empty(t, rec.Header().Get(echo.HeaderContentType))
		assert.Empty(t, len(rec.Body.Bytes()))
	}
}

func TestBrotliErrorReturned(t *testing.T) {
	e := echo.New()
	e.Use(br.Brotli())
	e.GET("/", func(_ echo.Context) error {
		return echo.ErrNotFound
	})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderAcceptEncoding, br.BrotliScheme)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusNotFound, rec.Code)
	assert.Empty(t, rec.Header().Get(echo.HeaderContentEncoding))
}

// Issue #806.
func TestBrotliWithStatic(t *testing.T) {
	e := echo.New()
	e.Use(br.Brotli())
	e.Static("/test", "../../../public/image/layout")
	req := httptest.NewRequest(http.MethodGet, "/test/favicon-152x152.png", nil)
	req.Header.Set(echo.HeaderAcceptEncoding, br.BrotliScheme)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
	// Data is written out in chunks when Content-Length == "", so only
	// validate the content length if it's not set.
	if cl := rec.Header().Get("Content-Length"); cl != "" {
		assert.Equal(t, cl, rec.Body.Len())
	}
	r := brotli.NewReader(rec.Body)

	want, err := os.ReadFile("../../../public/image/layout/favicon-152x152.png")
	if assert.NoError(t, err) {
		buf := new(bytes.Buffer)
		_, err = buf.ReadFrom(r)
		require.NoError(t, err)
		assert.EqualValues(t, want, buf.Bytes())
	}
}
