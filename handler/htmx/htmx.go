// Package htmx handles the routes and views for the AJAX responses using the htmx library.
package htmx

import (
	"context"
	"crypto/sha512"
	"embed"
	"errors"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
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
)

var (
	ErrDB    = errors.New("database connection is nil")
	ErrExist = errors.New("file already exists")
)

// Routes for the /htmx sub-route group that returns HTML fragments
// using the htmx library for AJAX responses.
func Routes(logr *zap.SugaredLogger, e *echo.Echo) *echo.Echo {
	submit := e.Group("", middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(2)))
	submit.POST("/demozoo/production", func(x echo.Context) error {
		return DemozooProd(x)
	})
	submit.POST("/demozoo/production/submit/:id", func(x echo.Context) error {
		return DemozooSubmit(logr, x)
	})
	submit.POST("/pouet/production", func(x echo.Context) error {
		return PouetProd(x)
	})
	submit.POST("/pouet/production/submit/:id", func(x echo.Context) error {
		return PouetSubmit(logr, x)
	})
	submit.POST("/search/releaser", func(x echo.Context) error {
		return SearchReleaser(logr, x)
	})
	submit.POST("/uploader/intro", func(x echo.Context) error {
		return holder(x)
	})
	return e
}

func holder(c echo.Context) error {
	// Source
	input, err := c.FormFile("uploader-intro-file")
	if err != nil {
		return err
	}
	src, err := input.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	hasher := sha512.New384()
	if _, err := io.Copy(hasher, src); err != nil {
		return err
	}
	sum := hasher.Sum(nil)
	fmt.Printf("%x; %s\n", sum, input.Filename)

	db, err := postgres.ConnectDB()
	if err != nil {
		return ErrDB
	}
	defer db.Close()
	ctx := context.Background()

	if exist, err := model.ExistsHash(ctx, db, sum); err != nil {
		return err
	} else if exist {
		return c.HTML(http.StatusOK, fmt.Sprintf("<p>File %s already exists.</p>", input.Filename))
	}

	// reopen the file
	src, err = input.Open()
	if err != nil {
		return err
	}
	defer src.Close()
	// Destination
	dst, err := os.CreateTemp("tmp", "upload-*.zip")
	//dst, err := os.Create(file.Filename)
	if err != nil {
		return err
	}
	defer dst.Close()

	// Copy
	if _, err = io.Copy(dst, src); err != nil {
		return err
	}

	return c.HTML(http.StatusOK,
		fmt.Sprintf("<p>File %s uploaded successfully with fields.</p><p>%s</p>", input.Filename, dst.Name()))

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

// Templates returns a map of the templates used by the HTML3 sub-group route.
func Templates(fs embed.FS) map[string]*template.Template {
	t := make(map[string]*template.Template)
	t["releasers"] = releasers(fs)
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
		"byteFileS": app.ByteFileS,
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
