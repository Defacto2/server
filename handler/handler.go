package handler

import (
	"bufio"
	"context"
	"embed"
	"errors"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"time"

	"github.com/Defacto2/server/api/apiv1"
	"github.com/Defacto2/server/cmd"
	"github.com/Defacto2/server/handler/bootstrap"
	"github.com/Defacto2/server/handler/download"
	"github.com/Defacto2/server/handler/html3"
	"github.com/Defacto2/server/pkg/config"
	"github.com/Defacto2/server/pkg/postgres"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
)

var (
	ErrNoTmpl = errors.New("no template name exists for recordsby type index")
	ErrTmpl   = errors.New("named template cannot be found")
)

const (
	ShutdownCount = 3
	ShutdownWait  = ShutdownCount * time.Second
)

func Join(srcs ...map[string]*template.Template) map[string]*template.Template {
	m := make(map[string]*template.Template)
	for _, src := range srcs {
		for k, val := range src {
			m[k] = val
		}
	}
	return m
}

// TemplateRegistry is template registry struct.
type TemplateRegistry struct {
	Templates map[string]*template.Template
}

// Render the layout template with the core HTML, META and BODY elements.
func (t *TemplateRegistry) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	if name == "" {
		return ErrNoTmpl
	}
	tmpl, ok := t.Templates[name]
	if !ok {
		return fmt.Errorf("%w: %s", ErrTmpl, name)
	}
	return tmpl.ExecuteTemplate(w, "layout", data)
}

// Configuration of the handler.
type Configuration struct {
	Import  *config.Config     // Import configurations from the host system environment.
	Log     *zap.SugaredLogger // Log is a sugared logger.
	Brand   *[]byte            // Brand points to the Defacto2 ASCII logo.
	Version string             // Version is the results of GoReleaser build command.
	Public  embed.FS           // Public facing files.
	Views   embed.FS           // Views are Go templates.
}

// Controller is the primary instance of the Echo router.
func (c Configuration) Controller() *echo.Echo {
	e := echo.New()

	// Configurations
	e.HideBanner = true
	e.Use(middleware.Secure())

	// HTML templates
	e.Renderer = &TemplateRegistry{
		Templates: Join(
			html3.TmplHTML3(c.Log, c.Views),
			bootstrap.Tmpl(c.Log, c.Public, c.Views),
		),
	}

	// Static embedded web assets
	// These get distributed in the binary
	e.StaticFS("/images/html3", echo.MustSubFS(c.Public, "public/images/html3"))
	e.GET("/images/html3", func(ctx echo.Context) error {
		return echo.NewHTTPError(http.StatusNotFound)
	})
	e.StaticFS("/images/layout", echo.MustSubFS(c.Public, "public/images/layout"))
	e.GET("/images/layout", func(ctx echo.Context) error {
		return echo.NewHTTPError(http.StatusNotFound)
	})
	e.StaticFS("/images/artpack", echo.MustSubFS(c.Public, "public/images/artpack"))
	e.GET("/images/artpack", func(ctx echo.Context) error {
		return echo.NewHTTPError(http.StatusNotFound)
	})

	// Middleware
	e.Use(middleware.Gzip())
	// remove trailing slashes
	e.Use(middleware.RemoveTrailingSlashWithConfig(middleware.TrailingSlashConfig{
		RedirectCode: http.StatusMovedPermanently,
	}))
	// www. redirect
	e.Pre(middleware.NonWWWRedirect())
	// timeout
	e.Use(middleware.TimeoutWithConfig(middleware.TimeoutConfig{
		Timeout: time.Duration(c.Import.Timeout) * time.Second,
	}))

	// Redirects, these need to be before the routes and rewrites
	e.GET("/defacto2/history", func(ctx echo.Context) error {
		return ctx.Redirect(http.StatusMovedPermanently, "/history")
	})
	e.GET("/defacto2/subculture", func(ctx echo.Context) error {
		return ctx.Redirect(http.StatusMovedPermanently, "/thescene")
	})
	e.GET("/files/json/site.webmanifest", func(ctx echo.Context) error {
		return ctx.Redirect(http.StatusMovedPermanently, "/site.webmanifest")
	})

	// Rewrites for assets
	e.Pre(middleware.Rewrite(map[string]string{
		"/logo.txt": "/text/defacto2.txt",
	}))

	// Serve embeded CSS files
	e.FileFS("/css/bootstrap.min.css", "public/css/bootstrap.min.css", c.Public)
	e.FileFS("/css/bootstrap.min.css.map", "public/css/bootstrap.min.css.map", c.Public)
	e.FileFS("/css/layout.min.css", "public/css/layout.min.css", c.Public)
	// Serve embeded SVG collections
	e.FileFS("/bootstrap-icons.svg", "public/images/bootstrap-icons.svg", c.Public)
	// Serve embeded JS files
	e.FileFS("/js/bootstrap.bundle.min.js", "public/js/bootstrap.bundle.min.js", c.Public)
	e.FileFS("/js/bootstrap.bundle.min.js.map", "public/js/bootstrap.bundle.min.js.map", c.Public)
	e.FileFS("/js/fontawesome.min.js", "public/js/fontawesome.min.js", c.Public)
	// Serve embeded image files
	e.FileFS("/favicon.ico", "public/images/favicon.ico", c.Public)
	// Serve embedded text files
	e.FileFS("/osd.xml", "public/text/osd.xml", c.Public)
	e.FileFS("/robots.txt", "public/text/robots.txt", c.Public)
	e.FileFS("/site.webmanifest", "public/text/site.webmanifest.json", c.Public)

	if c.Import.IsProduction {
		// recover from panics
		e.Use(middleware.Recover())
		// https redirect
		// e.Pre(middleware.HTTPSRedirect())
		// e.Pre(middleware.HTTPSNonWWWRedirect())
	}

	// HTTP status logger
	e.Use(c.Import.LoggerMiddleware)

	// Custom response headers
	if c.Import.NoRobots {
		e.Use(NoRobotsHeader)
	}

	// Route => /
	e.GET("/", func(ctx echo.Context) error {
		return bootstrap.Index(nil, ctx)
	})
	e.GET("/file/list", func(ctx echo.Context) error {
		return ctx.String(http.StatusOK, "Coming soon!")
	})
	e.GET("/history", func(ctx echo.Context) error {
		return bootstrap.History(nil, ctx)
	})
	e.GET("/thanks", func(ctx echo.Context) error {
		return bootstrap.Thanks(nil, ctx)
	})
	e.GET("/thescene", func(ctx echo.Context) error {
		return bootstrap.TheScene(nil, ctx)
	})

	// Routes => /html3
	g := html3.Routes(e, c.Log)
	g.GET("/d/:id", func(ctx echo.Context) error {
		d := download.Download{
			Path: c.Import.DownloadDir,
		}
		return d.HTTPSend(c.Log, ctx)
	})

	// Routers => /api/v1
	_ = apiv1.Routes(e, c.Log)

	// Router => HTTP error handler
	e.HTTPErrorHandler = c.Import.CustomErrorHandler

	return e
}

// StartHTTP starts the HTTP web server.
func (c *Configuration) StartHTTP(e *echo.Echo) {
	const mark = `⇨ `
	w := bufio.NewWriter(os.Stdout)
	// Startup logo
	if logo := string(*c.Brand); len(logo) > 0 {
		w := bufio.NewWriter(os.Stdout)
		if _, err := fmt.Fprintf(w, "%s\n\n", logo); err != nil {
			c.Log.Warnf("Could not print the brand logo: %s.", err)
		}
		w.Flush()
	}
	// Legal info
	fmt.Fprintf(w, "  %s.\n", cmd.Copyright())
	// Check the database connection
	var psql postgres.Version
	if err := psql.Query(); err != nil {
		c.Log.Warnln("Could not obtain the PostgreSQL server version. Is the database online?")
	} else {
		fmt.Fprintf(w, "%sDefacto2 web application %s %s.\n", mark, cmd.Commit(c.Version), psql.String())
	}
	// CPU info
	fmt.Fprintf(w, "%s%d active routines sharing %d usable threads on %d CPU cores.\n", mark,
		runtime.NumGoroutine(), runtime.GOMAXPROCS(-1), runtime.NumCPU())
	// Go info
	fmt.Fprintf(w, "%sCompiled with Go %s for %s, %s.\n",
		mark, runtime.Version()[2:], cmd.OS(), cmd.Arch())
	// Log location info
	if c.Import.IsProduction {
		fmt.Fprintf(w, "%sserver logs are found in: %s\n", mark, c.Import.LogDir)
	}
	w.Flush()

	serverAddress := fmt.Sprintf(":%d", c.Import.HTTPPort)
	err := e.Start(serverAddress)
	if err != nil && err != http.ErrServerClosed {
		c.Log.Fatalf("HTTP server could not start: %s.", err)
	}
	// nothing should be placed here
}

// ShutdownHTTP waits for a Ctrl-C keyboard press to initiate a graceful shutdown of the HTTP web server.
// The shutdown procedure occurs a few seconds after the key press.
func (c *Configuration) ShutdownHTTP(e *echo.Echo) {
	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	ctx, cancel := context.WithTimeout(context.Background(), ShutdownWait)
	defer func() {
		const alert = "Detected Ctrl-C, server will shutdown in "
		_ = c.Log.Sync() // do not check error as there's false positives
		dst := os.Stdout
		w := bufio.NewWriter(dst)
		fmt.Fprintf(w, "\n%s%v", alert, ShutdownWait)
		w.Flush()
		count := ShutdownCount
		pause := time.NewTicker(1 * time.Second)
		for range pause.C {
			count--
			w := bufio.NewWriter(dst)
			if count <= 0 {
				fmt.Fprintf(w, "\r%s%s\n", alert, "now")
				w.Flush()
				break
			}
			fmt.Fprintf(w, "\r%s%ds", alert, count)
			w.Flush()
		}
		select {
		case <-quit:
			cancel()
		case <-ctx.Done():
		}
		if err := e.Shutdown(ctx); err != nil {
			c.Log.Fatalf("Server shutdown caused an error: %w.", err)
		}
		c.Log.Infoln("Server shutdown complete.")
		_ = c.Log.Sync()
		signal.Stop(quit)
		cancel()
	}()
}
