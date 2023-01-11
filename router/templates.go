package router

import (
	"errors"
	"html/template"
	"io"

	"github.com/labstack/echo/v4"
)

// Define the template registry struct
type TemplateRegistry struct {
	Templates map[string]*template.Template
}

// Implement e.Renderer interface
func (t *TemplateRegistry) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	tmpl, ok := t.Templates[name]
	if !ok {
		err := errors.New("Template not found -> " + name)
		return err
	}
	return tmpl.ExecuteTemplate(w, "layout", data)
}

// support multiple template directories
// https://stackoverflow.com/questions/38686583/golang-parse-all-templates-in-directory-and-subdirectories
//t := template.Must(template.ParseGlob("public/views/*.html"))
//template.Must(t.ParseGlob("template/layout/*.tmpl"))

// Instantiate a template registry with an array of template set
// Ref: https://gist.github.com/rand99/808e6e9702c00ce64803d94abff65678

func TmplHTML3() map[string]*template.Template {
	templates := make(map[string]*template.Template)
	templates["index"] = template.Must(template.New("").Funcs(TemplateFuncMap).ParseFiles(
		"public/views/html3/layout.html", "public/views/html3/index.html"))
	templates["categories"] = template.Must(template.New("").Funcs(TemplateFuncMap).ParseFiles(
		"public/views/html3/layout.html", "public/views/html3/categories.html"))
	templates["category"] = template.Must(template.New("").Funcs(TemplateFuncMap).ParseFiles(
		"public/views/html3/layout.html", "public/views/html3/files.html"))
	return templates
}

// ----- old

// TemplateRenderer is a custom html/template renderer for Echo framework
type TemplateRenderer struct {
	templates *template.Template
}

// Render renders a template document
func (t *TemplateRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {

	// Add global methods if data is a map
	if viewContext, isMap := data.(map[string]interface{}); isMap {
		viewContext["reverse"] = c.Echo().Reverse
	}

	return t.templates.ExecuteTemplate(w, name, data)
}
