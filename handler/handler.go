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
	"github.com/Defacto2/server/internal/config"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
)

var (
	ErrCtx    = errors.New("echo context is nil")
	ErrData   = errors.New("data interface is nil")
	ErrName   = errors.New("template name string is empty")
	ErrRoutes = errors.New("e echo instance is nil")
	ErrTmpl   = errors.New("named template cannot be found")
	ErrW      = errors.New("w io.writer instance is nil")
	ErrZap    = errors.New("zap logger instance is nil")
)

const (
	// ShutdownCounter is the number of iterations to wait before shutting down the server.
	ShutdownCounter = 3
	// ShutdownWait is the number of seconds to wait before shutting down the server.
	ShutdownWait = ShutdownCounter * time.Second
	// Downloader is the route for the file download handler.
	Downloader = "/d/:id"
)

// Configuration of the handler.
type Configuration struct {
	Import      *config.Config     // Import configurations from the host system environment.
	Logger      *zap.SugaredLogger // Logger is the zap sugared logger.
	Brand       *[]byte            // Brand points to the Defacto2 ASCII logo.
	Version     string             // Version is the results of GoReleaser build command.
	Public      embed.FS           // Public facing files.
	View        embed.FS           // View contains Go templates.
	RecordCount int                // The total number of file records in the database.
}

// Registry returns the template renderer.
func (c Configuration) Registry() (*TemplateRegistry, error) {
	webapp := app.Web{
		Import: c.Import,
		Logger: c.Logger,
		Brand:  c.Brand,
		Public: c.Public,
		View:   c.View,
	}
	webTmpl, err := webapp.Tmpl()
	if err != nil {
		return nil, err
	}
	htmTmpl := html3.Tmpl(c.Logger, c.View)
	return &TemplateRegistry{
		Templates: Join(
			webTmpl, htmTmpl,
		),
	}, nil
}

// EmbedDirs serves the static files from the directories embed to the binary.
func (c Configuration) EmbedDirs(e *echo.Echo) *echo.Echo {
	if e == nil {
		c.Logger.Fatal(ErrRoutes)
	}
	dirs := map[string]string{
		"/image/artpack":   "public/image/artpack",
		"/image/html3":     "public/image/html3",
		"/image/layout":    "public/image/layout",
		"/image/milestone": "public/image/milestone",
	}
	for path, fsRoot := range dirs {
		e.StaticFS(path, echo.MustSubFS(c.Public, fsRoot))
		e.GET(path, func(ctx echo.Context) error {
			return echo.NewHTTPError(http.StatusNotFound)
		})
	}
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
	e := echo.New()
	//
	// Configurations
	e.HideBanner = true                              // hide the Echo banner
	e.HTTPErrorHandler = c.Import.CustomErrorHandler // custom error handler (see: internal/config/logger.go)
	reg, err := c.Registry()                         // HTML templates
	if err != nil {
		c.Logger.Fatal(err)
	}
	e.Renderer = reg
	//
	// Pre configurations that are run before the router
	e.Pre(middleware.Rewrite(rewrites())) // rewrites for assets
	e.Pre(middleware.NonWWWRedirect())    // redirect www.defacto2.net requests to defacto2.net
	if c.Import.HTTPSRedirect {
		e.Pre(middleware.HTTPSRedirect()) // https redirect
	}
	//
	// Use configurations that are run after the router
	// Note: NEVER USE the middleware.Timeout() as it will cause the server to crash
	// See: https://github.com/labstack/echo/blob/v4.11.1/middleware/timeout.go
	e.Use(middleware.Secure())       // XSS cross-site scripting protection
	e.Use(middleware.Gzip())         // Gzip HTTP compression
	e.Use(c.Import.LoggerMiddleware) // custom HTTP logging middleware
	// 									(see: internal/config/logger.go)
	e.Use(middleware.RemoveTrailingSlashWithConfig(c.removeSlash())) // remove trailing slashes
	e.Use(c.NoRobotsHeader)                                          // add X-Robots-Tag to all responses
	if c.Import.IsProduction {
		e.Use(middleware.Recover()) // recover from panics
	}
	// Static embedded web assets that get distributed in the binary
	e = c.EmbedDirs(e)
	// Routes for the web application
	e, err = c.Moved(c.Logger, e)
	if err != nil {
		c.Logger.Fatal(err)
	}
	e, err = c.Routes(c.Logger, e, c.Public)
	if err != nil {
		c.Logger.Fatal(err)
	}
	// Routes for the htm retro web tables
	retro := html3.Routes(c.Logger, e)
	retro.GET(Downloader, c.downloader)
	// Route for the api
	_ = apiv1.Routes(c.Logger, e)

	return e
}

// downloader route for the file download handler under the html3 group.
func (c Configuration) downloader(ctx echo.Context) error {
	d := download.Download{
		Inline: false,
		Path:   c.Import.DownloadDir,
	}
	return d.HTTPSend(c.Logger, ctx)
}

func (c Configuration) version() string {
	if c.Version == "" {
		return "  no version info, app compiled binary directly."
	}
	return fmt.Sprintf("  %s.", cmd.Commit(c.Version))
}

// StartHTTP starts the HTTP web server.
func (c *Configuration) StartHTTP(e *echo.Echo) {
	const mark = `⇨ `
	w := bufio.NewWriter(os.Stdout)
	// Startup logo
	if logo := string(*c.Brand); len(logo) > 0 {
		w := bufio.NewWriter(os.Stdout)
		if _, err := fmt.Fprintf(w, "%s\n\n", logo); err != nil {
			c.Logger.Warnf("Could not print the brand logo: %s.", err)
		}
		w.Flush()
	}
	// Legal info
	fmt.Fprintf(w, "  %s.\n", cmd.Copyright())
	// Brief version
	fmt.Fprintf(w, "  %s\n", c.version())
	// CPU info
	fmt.Fprintf(w, "    %d active routines sharing %d usable threads on %d CPU cores.\n",
		runtime.NumGoroutine(), runtime.GOMAXPROCS(-1), runtime.NumCPU())
	// Go info
	fmt.Fprintf(w, "%scompiled with Go %s for %s on %s.\n",
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
		fmt.Fprintf(w, "%sNoRobots is on, network headers will tell web crawlers to ignore this site.\n", mark)
	}
	w.Flush()
	// Start the HTTP server
	serverAddress := fmt.Sprintf(":%d", c.Import.HTTPPort)
	if err := e.Start(serverAddress); err != nil {
		var portErr *net.OpError
		switch {
		case !c.Import.IsProduction && errors.As(err, &portErr):
			c.Logger.Infof("air or task server could not start (this can probably be ignored): %s.", err)
		case errors.Is(err, net.ErrClosed),
			errors.Is(err, http.ErrServerClosed):
			c.Logger.Infof("HTTP server shutdown gracefully.")
		default:
			c.Logger.Fatalf("HTTP server could not start: %s.", err)
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
		_ = c.Logger.Sync() // do not check error as there's false positives
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
			c.Logger.Fatalf("Server shutdown caused an error: %w.", err)
		}
		c.Logger.Infoln("Server shutdown complete.")
		_ = c.Logger.Sync()
		signal.Stop(quit)
		cancel()
	}()
}
