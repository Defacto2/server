package handler_test

import (
	"io"
	"testing"

	"github.com/Defacto2/server/handler"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegister(t *testing.T) {
	t.Parallel()
	c := handler.Configuration{}
	tr, err := c.Registry()
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

func TestCookieStore(t *testing.T) {
	t.Parallel()
	b, err := handler.CookieStore("")
	require.NoError(t, err)
	assert.Len(t, b, 32)

	const key = "my-secret-key"
	b, err = handler.CookieStore(key)
	require.NoError(t, err)
	assert.Len(t, b, len(key))
}
