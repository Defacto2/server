package handler_test

import (
	"net/http"
	"testing"
	"testing/fstest"

	"github.com/Defacto2/server/handler"
	"github.com/Defacto2/server/internal/logs"
	"github.com/Defacto2/server/internal/testutil"
	"github.com/labstack/echo/v5"
	"github.com/nalgeon/be"
)

func TestFiles0(t *testing.T) {
	t.Parallel()

	serv := handler.Server{}
	e, err := serv.RouteFiles(nil, nil, nil, nil)
	be.Err(t, err)
	be.Equal(t, e, nil)
}

func TestFiles1(t *testing.T) {
	t.Parallel()

	// this ends the test if there's no active database
	db := testutil.DB(t)

	sl := logs.Discard()
	e := echo.New()
	xfs := fstest.MapFS{}

	serv := handler.Server{}
	e, err := serv.RouteFiles(sl, e, db, xfs)
	be.Err(t, err)
	be.True(t, e == nil)
}

func TestFiles2(t *testing.T) {
	t.Parallel()

	serv := handler.Server{}
	db := testutil.DB(t)
	sl := logs.Discard()
	e := echo.New()
	fsys := testutil.OpenFS(t)
	e, err := serv.RouteFiles(sl, e, db, fsys)
	be.Err(t, err, nil)

	c := testutil.EchoContext(t, e, "/want-404")
	err = c.NoContent(http.StatusNotFound)
	be.Err(t, err, nil)
	res, ok := c.Response().(*echo.Response)
	be.True(t, ok)
	be.Equal(t, res.Status, http.StatusNotFound)
}
