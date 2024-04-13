// Package htmx handles the routes and views for the AJAX responses using the htmx library.
package htmx

import (
	"embed"
	"errors"
	"html/template"
	"reflect"
	"strings"

	"github.com/Defacto2/releaser"
	"github.com/Defacto2/releaser/initialism"
	"github.com/Defacto2/releaser/name"
	"github.com/Defacto2/server/handler/app"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

var (
	ErrDB       = errors.New("database connection is nil")
	ErrExist    = errors.New("file already exists")
	ErrFileHead = errors.New("file header is nil")
	ErrRoutes   = errors.New("echo instance is nil")
)

const rateLimit = 2

// Routes for the /htmx sub-route group that returns HTML fragments
// using the htmx library for AJAX responses.
func Routes(e *echo.Echo, logger *zap.SugaredLogger, prod bool) *echo.Echo {
	if e == nil {
		panic(ErrRoutes)
	}
	submit := e.Group("",
		middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(rateLimit)))
	submit.POST("/demozoo/production", DemozooProd)
	submit.POST("/demozoo/production/submit/:id", func(c echo.Context) error {
		return DemozooSubmit(c, logger)
	})
	submit.POST("/pouet/production", PouetProd)
	submit.POST("/pouet/production/submit/:id", func(c echo.Context) error {
		return PouetSubmit(c, logger)
	})
	submit.POST("/search/releaser", func(c echo.Context) error {
		return SearchReleaser(c, logger)
	})
	submit.PUT("/uploader/sha384/:hash", func(c echo.Context) error {
		return LookupSHA384(c, logger)
	})
	submit.POST("/uploader/intro", func(c echo.Context) error {
		if prod {
			return transfer(c, nil, "uploader-introfile")
		}
		return transfer(c, logger, "uploader-introfile")
	})
	submit.POST("/uploader/releaser/1", func(c echo.Context) error {
		input := c.FormValue("uploader-intro-releaser1")
		return DataListReleasers(c, logger, input)
	})
	submit.POST("/uploader/releaser/2", func(c echo.Context) error {
		input := c.FormValue("uploader-intro-releaser2")
		return DataListReleasers(c, logger, input)
	})
	return e
}

// GlobTo returns the path to the template file.
func GlobTo(name string) string {
	const pathSeparator = "/"
	return strings.Join([]string{"view", "htmx", name}, pathSeparator)
}

func releasers(fs embed.FS) *template.Template {
	return template.Must(template.New("").Funcs(TemplateFuncMap()).ParseFS(fs,
		GlobTo("layout.tmpl"), GlobTo("releasers.tmpl")))
}

func datalistReleasers(fs embed.FS) *template.Template {
	return template.Must(template.New("").Funcs(TemplateFuncMap()).ParseFS(fs,
		GlobTo("layout.tmpl"), GlobTo("datalist-releasers.tmpl")))
}

// Templates returns a map of the templates used by the HTML3 sub-group route.
func Templates(fs embed.FS) map[string]*template.Template {
	t := make(map[string]*template.Template)
	t["releasers"] = releasers(fs)
	t["datalist-releasers"] = datalistReleasers(fs)
	return t
}

// TemplateFuncMap are a collection of mapped functions that can be used in a template.
func TemplateFuncMap() template.FuncMap {
	return template.FuncMap{
		"borderClass": func(name, path string) string {
			const mark = "border border-primary"
			if strings.EqualFold(name, path) {
				return mark
			}
			init := initialism.Join(initialism.Path(path))
			if strings.EqualFold(name, init) {
				return mark
			}
			return "border"
		},
		"byteFileS":  app.ByteFileS,
		"suggestion": Suggestion,
		"fmtPath": func(path string) string {
			if val := name.Path(path); val.String() != "" {
				return val.String()
			}
			return releaser.Humanize(path)
		},
		"initialisms": func(s string) string {
			return initialism.Join(initialism.Path(s))
		},
		"safeHTML": func(s string) template.HTML {
			return template.HTML(s) //nolint:gosec
		},
	}
}

// Suggestion returns a human readable string of the byte count with a named description.
func Suggestion(name, initialism string, count any) string {
	s := name
	if initialism != "" {
		s += ", " + initialism
	}
	switch val := count.(type) {
	case int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64:
		i := reflect.ValueOf(val).Int()
		p := message.NewPrinter(language.English)
		s += p.Sprintf(" (%d item", i)
		if i > 1 {
			s += "s"
		}
		s += ")"
	default:
		s += "suggestion type error: " + reflect.TypeOf(count).String()
		return s
	}
	return s
}
