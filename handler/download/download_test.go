package download_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Defacto2/server/handler/download"
	"github.com/labstack/echo/v5"
	"github.com/nalgeon/be"
)

func newContext() *echo.Context {
	e := echo.New()
	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/", strings.NewReader("{}"))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	return e.NewContext(req, rec)
}

func TestChecksum(t *testing.T) {
	t.Parallel()
	err := download.Checksum(newContext(), nil, "")
	be.Err(t, err)
}

func TestHTTPSend(t *testing.T) {
	t.Parallel()
	ctx := context.TODO()
	d := download.Download{}
	err := d.HTTPSend(ctx, nil, newContext(), nil)
	be.Err(t, err)
}

func TestEZHTTPSend(t *testing.T) {
	t.Parallel()
	ez := download.ExtraZip{}
	err := ez.HTTPSend(newContext(), nil)
	be.Err(t, err)
}
