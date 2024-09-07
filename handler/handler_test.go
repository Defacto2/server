package handler_test

import (
	"io"
	"testing"

	"github.com/Defacto2/server/handler"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
)

func TestRender(t *testing.T) {
	t.Parallel()
	tr := new(handler.TemplateRegistry)
	err := tr.Render(nil, "", nil, nil)
	require.Error(t, err)

	err = tr.Render(nil, "name", nil, nil)
	require.Error(t, err)

	w := io.Discard
	err = tr.Render(w, "name", "data", nil)
	require.Error(t, err)

	c := echo.New().NewContext(nil, nil)
	err = tr.Render(w, "name", "data", c)
	require.Error(t, err)
}

func TestInfo(t *testing.T) {
	t.Parallel()
	c := handler.Configuration{}
	c.Info(nil, nil)
}

func TestRegistry(t *testing.T) {
	t.Parallel()
	c := handler.Configuration{}
	x, err := c.Registry(nil, nil)
	require.Error(t, err)
	require.Nil(t, x)
}
