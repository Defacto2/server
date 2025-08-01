// Package handler provides the HTTP handlers for the Defacto2 website.
// Using the [Echo] web framework, the handler is the entry point for the web server.
//
// [Echo]: https://echo.labstack.com/
package handler

import (
	"bufio"
	"bytes"
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"maps"
	"net"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"time"

	"github.com/Defacto2/helper"
	"github.com/Defacto2/server/flags"
	"github.com/Defacto2/server/handler/app"
	"github.com/Defacto2/server/handler/download"
	"github.com/Defacto2/server/handler/html3"
	"github.com/Defacto2/server/handler/htmx"
	"github.com/Defacto2/server/internal/config"
	"github.com/Defacto2/server/internal/dir"
	"github.com/labstack/echo-contrib/pprof"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
)

const (
	// ShutdownCounter is the number of iterations to wait before shutting down the server.
	ShutdownCounter = 3

	// ShutdownWait is the number of seconds to wait before shutting down the server.
	ShutdownWait = ShutdownCounter * time.Second

	// Downloader is the route for the file download handler.
	Downloader = "/d/:id"
)

var (
	ErrFS      = errors.New("embed file system instance is empty")
	ErrName    = errors.New("name is empty")
	ErrName404 = errors.New("named template cannot be found")
	ErrPorts   = errors.New("the server ports are not configured")
	ErrRoutes  = errors.New("echo instance is nil")
	ErrZap     = errors.New("zap logger instance is nil")
)

// Configuration of the handler.
type Configuration struct {
	Public      embed.FS      // Public facing files.
	View        embed.FS      // View contains Go templates.
	Version     string        // Version is the results of GoReleaser build command.
	Brand       []byte        // Brand contains the Defacto2 ASCII logo.
	Environment config.Config // Environment configurations from the host system.
	RecordCount int           // The total number of file records in the database.
}

// Controller is the primary instance of the Echo router.
func (c *Configuration) Controller(db *sql.DB, logger *zap.SugaredLogger) *echo.Echo {
	configs := c.Environment
	if logger == nil {
		logger, _ := zap.NewProduction()
		defer func() { _ = logger.Sync() }()
	}

	e := echo.New()
	if configs.LogAll {
		pprof.Register(e)
	}
	e.HideBanner = true
	e.HTTPErrorHandler = configs.CustomErrorHandler

	tmpl, err := c.Registry(db, logger)
	if err != nil {
		logger.Fatal(err)
	}
	e.Renderer = tmpl
	middlewares := []echo.MiddlewareFunc{
		middleware.Rewrite(rewrites()),
		middleware.NonWWWRedirect(),
	}
	e.Pre(middlewares...)

	// *************************************************
	//  NOTE: NEVER USE the middleware.Timeout()
	//   It is broken and should not be in the
	//   labstack/echo library, as it can easily crash!
	// *************************************************
	middlewares = []echo.MiddlewareFunc{
		middleware.Secure(),
		middleware.RequestLoggerWithConfig(c.configZapLogger()),
		c.NoCrawl,
		middleware.RemoveTrailingSlashWithConfig(configRTS()),
	}
	if configs.Compression {
		middlewares = append(middlewares, middleware.Gzip())
	}
	if configs.ProdMode {
		middlewares = append(middlewares, middleware.Recover())
	}
	e.Use(middlewares...)

	e = EmbedDirs(e, c.Public)
	e = MovedPermanently(e)
	e = htmxGroup(e, db, logger, configs.ProdMode, dir.Directory(c.Environment.AbsDownload))
	e, err = c.FilesRoutes(e, db, logger, c.Public)
	if err != nil {
		logger.Fatal(err)
	}
	group := html3.Routes(e, db, logger)
	group.GET(Downloader, func(cx echo.Context) error {
		return c.downloader(cx, db, logger)
	})
	return e
}

// EmbedDirs serves the static files from the directories embed to the binary.
func EmbedDirs(e *echo.Echo, currentFs fs.FS) *echo.Echo {
	if e == nil {
		panic(fmt.Errorf("%w for the embed directories binary", ErrRoutes))
	}
	dirs := map[string]string{
		"/image/artpack":   "public/image/artpack",
		"/image/html3":     "public/image/html3",
		"/image/layout":    "public/image/layout",
		"/image/milestone": "public/image/milestone",
		"/image/new":       "public/image/new",
		"/svg":             "public/svg",
		"/jsdos/bin":       "public/bin/dos32",
	}
	for path, fsRoot := range dirs {
		e.StaticFS(path, echo.MustSubFS(currentFs, fsRoot))
		e.GET(path, func(_ echo.Context) error {
			return echo.NewHTTPError(http.StatusNotFound)
		})
	}
	return e
}

// Info prints the application information to the console.
func (c *Configuration) Info(logger *zap.SugaredLogger, w io.Writer) {
	if w == nil {
		w = io.Discard
	}
	nr := bytes.NewReader(c.Brand)
	if l, err := io.Copy(w, nr); err != nil {
		if logger != nil {
			logger.Warnf("Could not print the brand logo: %s.", err)
		}
	} else if l > 0 {
		_, err := fmt.Fprint(w, "\n\n")
		if err != nil {
			panic(err)
		}
	}

	_, err := fmt.Fprintf(w, "  %s.\n", flags.Copyright())
	if err != nil {
		panic(err)
	}
	_, _ = fmt.Fprintf(w, "%s\n", c.versionBrief())

	cpuInfo := fmt.Sprintf("  %d active routines sharing %d usable threads on %d CPU cores.",
		runtime.NumGoroutine(), runtime.GOMAXPROCS(-1), runtime.NumCPU())
	_, _ = fmt.Fprintln(w, cpuInfo)

	golangInfo := fmt.Sprintf("  Compiled on Go %s for %s with %s.\n",
		runtime.Version()[2:], flags.OS(), flags.Arch())
	_, _ = fmt.Fprintln(w, golangInfo)
	//
	// All additional feedback should go in internal/config/check.go (c *Config) Checks()
	//
}

// PortErr handles the error when the HTTP or HTTPS server cannot start.
func (c *Configuration) PortErr(logger *zap.SugaredLogger, port uint, err error) {
	if logger == nil {
		logger, _ := zap.NewProduction()
		defer func() { _ = logger.Sync() }()
	}
	s := "HTTP"
	if port == c.Environment.TLSPort {
		s = "TLS"
	}
	var portErr *net.OpError
	switch {
	case !c.Environment.ProdMode && errors.As(err, &portErr):
		logger.Infof("air or task server could not start (this can probably be ignored): %s.", err)
	case errors.Is(err, net.ErrClosed),
		errors.Is(err, http.ErrServerClosed):
		logger.Infof("%s server shutdown gracefully.", s)
	case errors.Is(err, os.ErrNotExist):
		logger.Fatalf("%s server on port %d could not start: %w.", s, port, err)
	default:
	}
}

// Registry returns the template renderer.
func (c *Configuration) Registry(db *sql.DB, logger *zap.SugaredLogger) (*TemplateRegistry, error) {
	webapp := app.Templ{
		Environment: c.Environment,
		Brand:       c.Brand,
		Public:      c.Public,
		RecordCount: c.RecordCount,
		Version:     c.Version,
		View:        c.View,
	}
	tmpls, err := webapp.Templates(db)
	if err != nil {
		return nil, fmt.Errorf("handler registry, %w", err)
	}
	src := html3.Templates(db, logger, c.View)
	maps.Copy(tmpls, src)
	src = htmx.Templates(c.View)
	maps.Copy(tmpls, src)
	return &TemplateRegistry{Templates: tmpls}, nil
}

// ShutdownHTTP waits for a Ctrl-C keyboard press to initiate a graceful shutdown of the HTTP web server.
// The shutdown procedure occurs a few seconds after the key press.
func (c *Configuration) ShutdownHTTP(e *echo.Echo, logger *zap.SugaredLogger) {
	if e == nil {
		panic(fmt.Errorf("%w for the HTTP shutdown", ErrRoutes))
	}
	if logger == nil {
		logger, _ := zap.NewProduction()
		defer func() { _ = logger.Sync() }()
	}
	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	waitDuration := ShutdownWait
	waitCount := ShutdownCounter
	ticker := 1 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), waitDuration)
	defer func() {
		const alert = "Detected Ctrl + C, server will shutdown"
		_ = logger.Sync() // do not check Sync errors as there can be false positives
		out := os.Stdout
		buf := bufio.NewWriter(out)
		_, err := fmt.Fprintf(buf, "\n%s in %v ", alert, waitDuration)
		if err != nil {
			panic(err)
		}
		err = buf.Flush()
		if err != nil {
			panic(err)
		}
		count := waitCount
		pause := time.NewTicker(ticker)
		for range pause.C {
			count--
			w := bufio.NewWriter(out)
			if count <= 0 {
				_, err := fmt.Fprintf(w, "\r%s %s\n", alert, "now     ")
				if err != nil {
					panic(err)
				}
				err = w.Flush()
				if err != nil {
					panic(err)
				}
				break
			}
			_, err = fmt.Fprintf(w, "\r%s in %ds ", alert, count)
			if err != nil {
				panic(err)
			}
			err = w.Flush()
			if err != nil {
				panic(err)
			}
		}
		select {
		case <-quit:
			cancel()
		case <-ctx.Done():
		}
		if err := e.Shutdown(ctx); err != nil {
			logger.Fatalf("Server shutdown caused an error: %w.", err)
		}
		logger.Infoln("Server shutdown complete.")
		_ = logger.Sync()
		signal.Stop(quit)
		cancel()
	}()
}

// Start the HTTP, and-or the TLS servers.
func (c *Configuration) Start(e *echo.Echo, logger *zap.SugaredLogger, configs config.Config) error {
	if e == nil {
		panic(fmt.Errorf("%w for the web application startup", ErrRoutes))
	}
	if logger == nil {
		logger, _ := zap.NewProduction()
		defer func() { _ = logger.Sync() }()
	}
	switch {
	case configs.UseTLS() && configs.UseHTTP():
		go func() {
			e2 := e // we need a new echo instance, otherwise the server may use the wrong port
			c.StartHTTP(e2, logger)
		}()
		go c.StartTLS(e, logger)
	case configs.UseTLSLocal() && configs.UseHTTP():
		go func() {
			e2 := e // we need a new echo instance, otherwise the server may use the wrong port
			c.StartHTTP(e2, logger)
		}()
		go c.StartTLSLocal(e, logger)
	case configs.UseTLS():
		go c.StartTLS(e, logger)
	case configs.UseHTTP():
		go c.StartHTTP(e, logger)
	case configs.UseTLSLocal():
		go c.StartTLSLocal(e, logger)
	default:
		return ErrPorts
	}
	return nil
}

// StartHTTP starts the insecure HTTP web server.
func (c *Configuration) StartHTTP(e *echo.Echo, logger *zap.SugaredLogger) {
	if e == nil {
		panic(fmt.Errorf("%w for the HTTP startup", ErrRoutes))
	}
	port := c.Environment.HTTPPort
	address := c.address(port)
	if address == "" {
		return
	}
	if err := e.Start(address); err != nil {
		c.PortErr(logger, port, err)
	}
}

// StartTLS starts the encrypted TLS web server.
func (c *Configuration) StartTLS(e *echo.Echo, logger *zap.SugaredLogger) {
	if e == nil {
		panic(fmt.Errorf("%w for the TLS startup", ErrRoutes))
	}
	if logger == nil {
		logger, _ := zap.NewProduction()
		defer func() { _ = logger.Sync() }()
	}
	port := c.Environment.TLSPort
	address := c.address(port)
	if address == "" {
		return
	}
	certFile := c.Environment.TLSCert
	keyFile := c.Environment.TLSKey
	const failure = "Could not start the TLS server"
	if certFile == "" || keyFile == "" {
		logger.Fatalf("%s, missing certificate or key file.", failure)
	}
	if !helper.File(certFile) {
		logger.Fatalf("%s, certificate file does not exist: %s.", failure, certFile)
	}
	if !helper.File(keyFile) {
		logger.Fatalf("%s, key file does not exist: %s.", failure, keyFile)
	}
	if err := e.StartTLS(address, certFile, keyFile); err != nil {
		c.PortErr(logger, port, err)
	}
}

// StartTLSLocal starts the localhost, encrypted TLS web server.
// This should only be triggered when the server is running in local mode.
func (c *Configuration) StartTLSLocal(e *echo.Echo, logger *zap.SugaredLogger) {
	if e == nil {
		panic(fmt.Errorf("%w for the TLS local mode startup", ErrRoutes))
	}
	if logger == nil {
		logger, _ := zap.NewProduction()
		defer func() { _ = logger.Sync() }()
	}
	port := c.Environment.TLSPort
	address := c.address(port)
	if address == "" {
		return
	}
	const cert, key = "public/certs/cert.pem", "public/certs/key.pem"
	const failure = "Could not read the internal localhost"
	certB, err := c.Public.ReadFile(cert)
	if err != nil {
		logger.Fatalf("%s, TLS certificate: %s.", failure, err)
	}
	keyB, err := c.Public.ReadFile(key)
	if err != nil {
		logger.Fatalf("%s, TLS key: %s.", failure, err)
	}
	if err := e.StartTLS(address, certB, keyB); err != nil {
		c.PortErr(logger, port, err)
	}
}

func (c *Configuration) address(port uint) string {
	if port == 0 {
		return ""
	}
	address := fmt.Sprintf(":%d", port)
	if c.Environment.MatchHost != "" {
		address = fmt.Sprintf("%s:%d", c.Environment.MatchHost, port)
	}
	return address
}

// downloader route for the file download handler under the html3 group.
func (c *Configuration) downloader(cx echo.Context, db *sql.DB, logger *zap.SugaredLogger) error {
	d := download.Download{
		Inline: false,
		Dir:    dir.Directory(c.Environment.AbsDownload),
	}
	if err := d.HTTPSend(cx, db, logger); err != nil {
		return fmt.Errorf("d.HTTPSend: %w", err)
	}
	return nil
}

// versionBrief returns the application version string.
func (c *Configuration) versionBrief() string {
	if c.Version == "" {
		return "  no version info, app compiled binary directly."
	}
	return fmt.Sprintf("  %s.", flags.Commit(c.Version))
}

// Rewrites for assets.
// This is different to a redirect as it keeps the original URL in the browser.
func rewrites() map[string]string {
	return map[string]string{
		"/logo.txt": "/text/defacto2.txt",
	}
}

// TemplateRegistry is template registry struct.
type TemplateRegistry struct {
	Templates map[string]*template.Template
}

// Render the layout template with the core HTML, META and BODY elements.
func (t *TemplateRegistry) Render(w io.Writer, name string, data any, c echo.Context) error {
	const layout, info = "layout", "template registry render"
	if name == "" {
		return fmt.Errorf("%s layout: %w", info, ErrName)
	}
	if w == nil {
		return fmt.Errorf("%s io.writer is nil: %w", info, echo.ErrRendererNotRegistered)
	}
	if data == nil {
		return fmt.Errorf("%s data interface is nil: %w", info, echo.ErrRendererNotRegistered)
	}
	if c == nil {
		return fmt.Errorf("%s echo context is nil: %w", info, echo.ErrRendererNotRegistered)
	}
	tmpl, exists := t.Templates[name]
	if !exists {
		return fmt.Errorf("registry render %w: %q", ErrName404, name)
	}
	if err := tmpl.ExecuteTemplate(w, layout, data); err != nil {
		return fmt.Errorf("%s execute: %w", info, err)
	}
	return nil
}
