package handler

import (
	"context"
	"embed"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"time"

	"github.com/Defacto2/server/handler/download"
	"github.com/Defacto2/server/handler/html3"
	"github.com/Defacto2/server/pkg/config"
	"github.com/Defacto2/server/pkg/postgres"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
)

const ShutdownWait = 5 * time.Second

// Configuration of the handler.
type Configuration struct {
	Import *config.Config     // Import configurations from the host system environment.
	Log    *zap.SugaredLogger // Log is a sugared logger.
	Images embed.FS           // Not in use.
	Views  embed.FS           // Views are Go templates.
}

// Controller is the primary instance of the Echo router.
func (c Configuration) Controller() *echo.Echo {

	e := echo.New()
	e.HideBanner = true

	// HTML templates
	e.Renderer = &html3.TemplateRegistry{
		Templates: html3.TmplHTML3(c.Views),
	}

	// Static embedded images
	// These get distributed in the binary
	e.StaticFS("/images", echo.MustSubFS(c.Images, "public/images"))
	e.File("favicon.ico", "public/images/favicon.ico") // TODO: this is not being embedded

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
		e.Use(NoRobotsHeader) // TODO: only apply to HTML templates?
	}

	// Route => /
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Coming soon!")
	})
	e.GET("/file/list", func(c echo.Context) error {
		return c.String(http.StatusOK, "Coming soon!")
	})

	// Routes => /html3
	g := html3.Routes(e, c.Log)
	g.GET("/d/:id", func(x echo.Context) error {
		d := download.Download{
			Path: c.Import.DownloadDir,
		}
		return d.HTTPSend(c.Log, x)
	})

	// Router => HTTP error handler
	e.HTTPErrorHandler = c.Import.CustomErrorHandler

	return e
}

func (c *Configuration) StartHTTP(e *echo.Echo) {
	const mark = `⇨ `

	// Check the database connection
	var ver postgres.Version
	if err := ver.Query(); err != nil {
		c.Log.Warnln("Could not obtain the PostgreSQL server version. Is the database online?")
	} else {
		fmt.Printf("%sDefacto2 web application %s.\n", mark, ver.String())
	}

	fmt.Printf("%s%d active routines sharing %d usable threads on %d CPU cores.\n", mark,
		runtime.NumGoroutine(), runtime.GOMAXPROCS(-1), runtime.NumCPU())

	fmt.Printf("%sCompiled with Go %s.\n", mark, runtime.Version()[2:])
	if c.Import.IsProduction {
		fmt.Printf("%sserver logs are found in: %s\n", mark, c.Import.ConfigDir)
	}

	serverAddress := fmt.Sprintf(":%d", c.Import.HTTPPort)
	err := e.Start(serverAddress)
	if err != nil && err != http.ErrServerClosed {
		c.Log.Fatalf("HTTP server could not start: %s.", err)
	}
	// nothing should be placed here
}

func (c *Configuration) ShutdownHTTP(e *echo.Echo) {
	// Wait for interrupt signal to gracefully shutdown the server with a timeout of 5 seconds.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	ctx, cancel := context.WithTimeout(context.Background(), ShutdownWait)
	defer func() {
		const alert = "Detected Ctrl-C, server will shutdown in "
		if err := c.Log.Sync(); err != nil {
			c.Log.Warnf("Could not sync the log before shutdown: %s.\n", err)
		}
		fmt.Printf("\n%s%s", alert, ShutdownWait)
		count := ShutdownWait
		const pause = 1 * time.Second
		for range time.Tick(pause) {
			count--
			fmt.Printf("\r%s%ds", alert, count)
			if count <= 0 {
				fmt.Printf("\r%s%ds\n", alert, count)
				break
			}
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
		if err := c.Log.Sync(); err != nil {
			c.Log.Warnf("Could not sync the log before shutdown: %s.\n", err)
		}
		signal.Stop(quit)
		cancel()
	}()
}
