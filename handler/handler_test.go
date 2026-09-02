package handler_test

import (
	"bytes"
	"context"
	"io"
	"reflect"
	"testing"

	"github.com/Defacto2/server/handler"
	"github.com/Defacto2/server/handler/app"
	"github.com/Defacto2/server/internal/config"
	"github.com/Defacto2/server/internal/logs"
	"github.com/Defacto2/server/internal/testutil"
	"github.com/labstack/echo/v5"
	"github.com/nalgeon/be"
)

// checked in Sep 26, test coverage with an active database was fine at around 50%+
// otherwise it drops to 9%

func TestHandler(t *testing.T) {
	t.Parallel()

	db := testutil.DB(t)
	sl := logs.Discard()
	root := testutil.OpenRoot(t)

	c := config.Config{
		Compression: true,
		ProdMode:    true,
		Quiet:       true,
	}
	h := handler.Server{
		Environment: c,
		Public:      root.FS(),
		View:        root.FS(),
	}

	_ = h.Handler(t.Context(), sl, db)
}

func TestStart(t *testing.T) {
	t.Parallel()

	sl := logs.Discard()
	h := handler.Server{}
	c := config.Config{}
	got := h.Start(t.Context(), sl, nil, c)
	be.Err(t, got)

	// do not test a working configured startup as it is too messy,
	// both due to echo's handling of shutdown and the use of goroutines
}

func TestTempReg(t *testing.T) {
	t.Parallel()

	tmpl := handler.TemplateRegistry{}
	got := tmpl.Render(nil, nil, "", nil)
	be.Err(t, got)

	e := echo.New()
	c := e.NewContext(nil, nil)
	w := io.Discard
	got = tmpl.Render(c, w, "invalid-template-name", nil)
	be.Err(t, got)

	// this cannot be tested due to the lack of fs.FS in Render
	const thanks = "thanks.tmpl"
	got = tmpl.Render(c, w, app.GlobTo(thanks), "template placeholder")
	be.Err(t, got)
}

func TestCerts(t *testing.T) {
	t.Parallel()

	root := testutil.OpenRoot(t)

	sl := logs.Discard()
	c := config.Config{
		TLSPort: 1234, // must provide a tls port, otherwise the func will exit
	}
	h := handler.Server{
		Public:      root.FS(), // must provide a fs to serve the local file certificates
		Environment: c,
	}
	_, _, _, err := h.Local(t.Context(), nil)
	be.Err(t, err)

	x, cert, key, err := h.Local(t.Context(), sl)
	be.Err(t, err, nil)
	be.True(t, !reflect.ValueOf(x).IsZero())

	const wantC = 1090
	be.True(t, len(cert) == wantC)

	const wantK = 1704
	be.True(t, len(key) == wantK)
}

func TestRender(t *testing.T) {
	t.Parallel()
	tr := new(handler.TemplateRegistry)
	err := tr.Render(nil, nil, "", nil)
	be.Err(t, err)
	err = tr.Render(nil, nil, "name", nil)
	be.Err(t, err)
	w := io.Discard
	err = tr.Render(nil, w, "name", "data")
	be.Err(t, err)
	c := echo.New().NewContext(nil, nil)
	err = tr.Render(c, w, "name", "data")
	be.Err(t, err)
}

func TestInfo(t *testing.T) {
	t.Parallel()
	c := handler.Server{}
	err := c.Print(nil)
	be.Err(t, err, nil)

	var w bytes.Buffer
	err = c.Print(&w)
	be.Err(t, err, nil)
	be.True(t, len(w.String()) > 100)
}

func TestRegistry(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	c := handler.Server{}
	x, err := c.TemplRegistry(ctx, nil, nil)
	be.Err(t, err)
	be.True(t, x == nil)
}
