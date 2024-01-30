package handler_test

import (
	"html/template"
	"io"
	"testing"

	"github.com/Defacto2/server/handler"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestRender(t *testing.T) {
	t.Parallel()
	tr := new(handler.TemplateRegistry)
	err := tr.Render(nil, "", nil, nil)
	assert.Error(t, handler.ErrName, err)

	err = tr.Render(nil, "name", nil, nil)
	assert.Error(t, handler.ErrW, err)

	w := io.Discard
	err = tr.Render(w, "name", "data", nil)
	assert.Error(t, handler.ErrCtx, err)

	c := echo.New().NewContext(nil, nil)
	err = tr.Render(w, "name", "data", c)
	assert.Error(t, handler.ErrTmpl, err)
}

func TestJoin(t *testing.T) {
	t.Parallel()
	m := handler.Join(nil, nil)
	assert.Equal(t, 0, len(m))
	m = handler.Join(map[string]*template.Template{
		"one":   nil,
		"two":   nil,
		"three": nil,
	}, map[string]*template.Template{
		"four": nil,
		"five": nil,
		"six":  nil,
	})
	assert.Equal(t, 6, len(m))
}
