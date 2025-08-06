package handler_test

import (
	"io"
	"testing"

	"github.com/Defacto2/server/handler"
	"github.com/Defacto2/server/internal/logs"
	"github.com/labstack/echo/v4"
	"github.com/nalgeon/be"
)

func TestRender(t *testing.T) {
	t.Parallel()
	tr := new(handler.TemplateRegistry)
	err := tr.Render(nil, "", nil, nil)
	be.Err(t, err)
	err = tr.Render(nil, "name", nil, nil)
	be.Err(t, err)
	w := io.Discard
	err = tr.Render(w, "name", "data", nil)
	be.Err(t, err)
	c := echo.New().NewContext(nil, nil)
	err = tr.Render(w, "name", "data", c)
	be.Err(t, err)
}

func TestInfo(t *testing.T) {
	t.Parallel()
	c := handler.Configuration{}
	c.StartupBranding(logs.Discard(), nil)
}

func TestRegistry(t *testing.T) {
	t.Parallel()
	c := handler.Configuration{}
	x, err := c.Registry(nil, nil)
	be.Err(t, err)
	be.True(t, x == nil)
}
