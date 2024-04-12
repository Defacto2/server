// Package htmx handles the routes and views for the AJAX responses using the htmx library.
package htmx

import (
	"context"
	"crypto/sha512"
	"embed"
	"errors"
	"fmt"
	"html"
	"html/template"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"reflect"
	"strings"

	"github.com/Defacto2/releaser"
	"github.com/Defacto2/releaser/initialism"
	"github.com/Defacto2/releaser/name"
	"github.com/Defacto2/server/handler/app"
	"github.com/Defacto2/server/internal/postgres"
	"github.com/Defacto2/server/model"
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

// transfer is a generic file transfer handler that uploads and validates a chosen file upload.
// The provided name is that of the form input field. The logger is optional and if nil then
// the function will not log any debug information.
func transfer(c echo.Context, logger *zap.SugaredLogger, name string) error {
	file, err := c.FormFile(name)
	if err != nil {
		if logger != nil {
			logger.Error(fmt.Sprintf("The chosen file input caused an error, %s: %s", name, err))
		}
		return c.HTML(http.StatusBadRequest, "The chosen file form input caused an error.")
	}

	src, err := file.Open()
	if err != nil {
		if logger != nil {
			logger.Error(fmt.Sprintf("The chosen file input could not be opened, %s: %s", name, err))
		}
		return c.HTML(http.StatusBadRequest, "The chosen file input cannot be opened.")
	}
	defer src.Close()

	hasher := sha512.New384()
	if _, err := io.Copy(hasher, src); err != nil {
		if logger != nil {
			logger.Error(fmt.Sprintf("The chosen file input could not be hashed, %s: %s", name, err))
		}
		return c.HTML(http.StatusBadRequest, "The chosen file input cannot be hashed.")
	}
	sum := hasher.Sum(nil)

	db, err := postgres.ConnectDB()
	if err != nil {
		if logger != nil {
			logger.Error(fmt.Sprintf("%s: %s", ErrDB, err))
		}
		return c.HTML(http.StatusBadRequest, "Cannot connect to the database.")
	}
	defer db.Close()

	ctx := context.Background()
	exist, err := model.ExistsHash(ctx, db, sum)
	if err != nil {
		if logger != nil {
			logger.Error(fmt.Sprintf("%s: %s", ErrDB, err))
		}
		return c.HTML(http.StatusBadRequest, "Cannot confirm the hash with the database.")
	}
	if exist {
		return c.HTML(http.StatusOK, "<p>Thanks, but the chosen file already exists on Defacto2.</p>"+
			html.EscapeString(file.Filename))
	}
	// todo: after writer, use db to save the form data to the database.
	return writer(c, logger, file, name)
}

// writer is a generic file writer that saves the chosen file upload to a temporary file.
func writer(c echo.Context, logger *zap.SugaredLogger, file *multipart.FileHeader, name string) error {
	if file == nil {
		return ErrFileHead
	}
	src, err := file.Open()
	if err != nil {
		if logger != nil {
			logger.Error(fmt.Sprintf("The chosen file input could not be opened, %s: %s", name, err))
		}
		return c.HTML(http.StatusBadRequest, "The chosen file input cannot be opened.")
	}
	defer src.Close()

	dst, err := os.CreateTemp("tmp", "upload-*.zip")
	if err != nil {
		if logger != nil {
			logger.Error(fmt.Sprintf("Cannot create a temporary destination file, %s: %s", name, err))
		}
		return c.HTML(http.StatusBadRequest, "The temporary save cannot be created.")
	}
	defer dst.Close()

	if _, err = io.Copy(dst, src); err != nil {
		if logger != nil {
			logger.Error(fmt.Sprintf("Cannot copy to the temporary destination file, %s: %s", name, err))
		}
		return c.HTML(http.StatusBadRequest, "The temporary save cannot be written.")
	}

	html := fmt.Sprintf("<p>Thanks, the chosen file submission was a success.<br> ✓ %s</p>",
		html.EscapeString(file.Filename))

	if production := logger == nil; production {
		return c.HTML(http.StatusOK, html)
	}
	html, err = debug(c, html)
	if err != nil {
		return c.HTML(http.StatusOK, html+"<p>Could not show the the form parameters and values.</p>")
	}
	return c.HTML(http.StatusOK, html)
}

func debug(c echo.Context, html string) (string, error) {
	values, err := c.FormParams()
	if err != nil {
		return html, err
	}
	html += "<ul>"
	for k, v := range values {
		html += fmt.Sprintf("<li>%s: %s</li>", k, v)
	}
	html += "</ul>"
	return html, nil
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
