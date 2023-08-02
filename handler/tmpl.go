package handler

import (
	"fmt"
	"html/template"
	"io"

	"github.com/labstack/echo/v4"
)

// TemplateRegistry is template registry struct.
type TemplateRegistry struct {
	Templates map[string]*template.Template
}

// Render the layout template with the core HTML, META and BODY elements.
func (t *TemplateRegistry) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	if w == nil {
		return fmt.Errorf("%w: %s", echo.ErrRendererNotRegistered, "writer is nil")
	}
	if data == nil {
		return fmt.Errorf("%w: %s", echo.ErrRendererNotRegistered, "data is nil")
	}
	if c == nil {
		return fmt.Errorf("%w: %s", echo.ErrRendererNotRegistered, "context is nil")
	}
	if name == "" {
		return ErrNoTmpl
	}
	tmpl, ok := t.Templates[name]
	if !ok {
		return fmt.Errorf("%w: %s", ErrTmpl, name)
	}
	return tmpl.ExecuteTemplate(w, "layout", data)
}

// Join multiple templates into one collection.
func Join(srcs ...map[string]*template.Template) map[string]*template.Template {
	m := make(map[string]*template.Template)
	for _, src := range srcs {
		for k, val := range src {
			m[k] = val
		}
	}
	return m
}
