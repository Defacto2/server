package handler

// Package file tmpl.go contains the custom template functions for the web framework.

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
	const layout = "layout"
	if name == "" {
		return ErrName
	}
	if w == nil {
		return fmt.Errorf("%w: %s", echo.ErrRendererNotRegistered, ErrW)
	}
	if data == nil {
		return fmt.Errorf("%w: %s", echo.ErrRendererNotRegistered, ErrData)
	}
	if c == nil {
		return fmt.Errorf("%w: %s", echo.ErrRendererNotRegistered, ErrCtx)
	}
	tmpl, ok := t.Templates[name]
	if !ok {
		return fmt.Errorf("%w: %s", ErrTmpl, name)
	}
	return tmpl.ExecuteTemplate(w, layout, data)
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
