package handler_test

import (
	"io"
	"testing"

	"github.com/Defacto2/server/handler"
	"github.com/Defacto2/server/internal/zaplog"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegister(t *testing.T) {
	t.Parallel()
	c := handler.Configuration{}
	logger := zaplog.Status().Sugar()
	tr, err := c.Registry(logger)
	assert.Nil(t, tr)
	require.Error(t, err)
}

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
