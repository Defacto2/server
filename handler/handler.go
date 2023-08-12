// Package handler provides the HTTP handlers for the Defacto2 website.
// Using the Echo Project web framework, it is the entry point for the web server.
package handler

import (
	"bufio"
	"context"
	"embed"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"time"

	"github.com/Defacto2/server/api/apiv1"
	"github.com/Defacto2/server/cmd"
	"github.com/Defacto2/server/handler/app"
	"github.com/Defacto2/server/handler/download"
	"github.com/Defacto2/server/handler/html3"
	"github.com/Defacto2/server/pkg/config"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
)

var (
	ErrLog    = errors.New("e logger instance is nil")
	ErrNoTmpl = errors.New("named template does not exist for recordsby type index")
	ErrRoutes = errors.New("e echo instance is nil")
	ErrTmpl   = errors.New("named template cannot be found")
)

const (
	ShutdownCounter = 3                             // ShutdownCounter is the number of iterations to wait before shutting down the server.
	ShutdownWait    = ShutdownCounter * time.Second // ShutdownWait is the number of seconds to wait before shutting down the server.
)

// Configuration of the handler.
type Configuration struct {
	DatbaseErr bool               // DatbaseErr is true if the database connection failed.
	Import     *config.Config     // Import configurations from the host system environment.
	ZLog       *zap.SugaredLogger // ZLog is a sugared, zap logger.
	Brand      *[]byte            // Brand points to the Defacto2 ASCII logo.
	Version    string             // Version is the results of GoReleaser build command.
	Public     embed.FS           // Public facing files.
	Views      embed.FS           // Views are Go templates.
}

// Registry returns the template renderer.
func (c Configuration) Registry() *TemplateRegistry {
	webapp := app.Configuration{
		DatbaseErr: c.DatbaseErr,
		ZLog:       c.ZLog,
		Brand:      c.Brand,
		Public:     c.Public,
		Views:      c.Views,
	}
	return &TemplateRegistry{
		Templates: Join(
			webapp.Tmpl(),
			html3.Tmpl(c.ZLog, c.Views),
		),
	}
}

// EmbedDirs serves the static files from the directories embed to the binary.
func (c Configuration) EmbedDirs(e *echo.Echo) *echo.Echo {
	e.StaticFS("/image/artpack", echo.MustSubFS(c.Public, "public/image/artpack"))
	e.GET("/image/artpack", func(ctx echo.Context) error {
		return echo.NewHTTPError(http.StatusNotFound)
	})
	e.StaticFS("/image/html3", echo.MustSubFS(c.Public, "public/image/html3"))
	e.GET("/image/html3", func(ctx echo.Context) error {
		return echo.NewHTTPError(http.StatusNotFound)
	})
	e.StaticFS("/image/layout", echo.MustSubFS(c.Public, "public/image/layout"))
	e.GET("/image/layout", func(ctx echo.Context) error {
		return echo.NewHTTPError(http.StatusNotFound)
	})
	e.StaticFS("/image/milestone", echo.MustSubFS(c.Public, "public/image/milestone"))
	e.GET("/image/milestone", func(ctx echo.Context) error {
		return echo.NewHTTPError(http.StatusNotFound)
	})
	return e
}

// Rewrites for assets.
// This is different to a redirect as it keeps the original URL in the browser.
func rewrites() map[string]string {
	return map[string]string{
		"/logo.txt": "/text/defacto2.txt",
	}
}

// Controller is the primary instance of the Echo router.
func (c Configuration) Controller() *echo.Echo {

	// TODO: handle broken DB connection

	e := echo.New()

	// Configurations
	e.HideBanner = true                              // hide the Echo banner
	e.HTTPErrorHandler = c.Import.CustomErrorHandler // custom error handler (see: pkg/config/logger.go)
	e.Renderer = c.Registry()                        // HTML templates

	// Pre configurations that are run before the router
	e.Pre(middleware.Rewrite(rewrites())) // rewrites for assets
	e.Pre(middleware.NonWWWRedirect())    // redirect www.defacto2.net requests to defacto2.net
	if c.Import.HTTPSRedirect {
		e.Pre(middleware.HTTPSRedirect()) // https redirect
	}

	// Use configurations that are run after the router
	e.Use(middleware.Secure())                                       // XSS cross-site scripting protection
	e.Use(middleware.Gzip())                                         // Gzip HTTP compression
	e.Use(c.Import.LoggerMiddleware)                                 // custom HTTP logging middleware (see: pkg/config/logger.go)
	e.Use(middleware.RemoveTrailingSlashWithConfig(c.removeSlash())) // remove trailing slashes
	e.Use(middleware.TimeoutWithConfig(c.timeout()))                 // timeout a long running operation
	e.Use(c.NoRobotsHeader)                                          // add X-Robots-Tag to all responses
	if c.Import.IsProduction {
		e.Use(middleware.Recover()) // recover from panics
	}

	// Static embedded web assets
	// These get distributed in the binary
	e = c.EmbedDirs(e)

	// Routes for the application.
	e, err := Routes(c.ZLog, e, c.Public)
	if err != nil {
		c.ZLog.Fatal(err)
	}

	// Routes for the HTML3 retro tables.
	retro := html3.Routes(c.ZLog, e)
	retro.GET("/d/:id", func(ctx echo.Context) error {
		// route for the file download handler under the html3 group
		d := download.Download{
			Path: c.Import.DownloadDir,
		}
		return d.HTTPSend(c.ZLog, ctx)
	})

	// Route for the API.
	_ = apiv1.Routes(c.ZLog, e)

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
			c.ZLog.Warnf("Could not print the brand logo: %s.", err)
		}
		w.Flush()
	}
	// Legal info
	fmt.Fprintf(w, "  %s.\n", cmd.Copyright())
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
	// Additional startup info
	if c.Import.HTTPSRedirect {
		fmt.Fprintf(w, "%sredirecting all HTTP requests to HTTPS.\n", mark)
	}
	if c.Import.NoRobots {
		fmt.Fprintf(w, "%sthe X-ROBOTS header is telling all search engines to ignore the site.\n", mark)
	}
	// Important layout assets checks
	// if _, err := os.Stat(filepath.Join(c.Public, "public/image/layout")); err != nil {
	// 	c.Log.Warnf("Could not find the layout assets directory: %s.", err)
	// }
	w.Flush()

	// Start the HTTP server
	serverAddress := fmt.Sprintf(":%d", c.Import.HTTPPort)
	if err := e.Start(serverAddress); err != nil {
		var portErr *net.OpError
		switch {
		case !c.Import.IsProduction && errors.As(err, &portErr):
			c.ZLog.Infof("air or task server could not start (this can probably be ignored): %s.", err)
		case errors.Is(err, net.ErrClosed),
			errors.Is(err, http.ErrServerClosed):
			c.ZLog.Infof("HTTP server shutdown gracefully.")
		default:
			c.ZLog.Fatalf("HTTP server could not start: %s.", err)
		}
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
		_ = c.ZLog.Sync() // do not check error as there's false positives
		dst := os.Stdout
		w := bufio.NewWriter(dst)
		fmt.Fprintf(w, "\n%s%v", alert, ShutdownWait)
		w.Flush()
		count := ShutdownCounter
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
			c.ZLog.Fatalf("Server shutdown caused an error: %w.", err)
		}
		c.ZLog.Infoln("Server shutdown complete.")
		_ = c.ZLog.Sync()
		signal.Stop(quit)
		cancel()
	}()
}
